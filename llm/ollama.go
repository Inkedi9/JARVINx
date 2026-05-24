package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OllamaClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaResponse struct {
	Message ollamaMessage `json:"message"`
}

// RetryConfig configure le comportement des retries
type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		Delay:       2 * time.Second,
	}
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

// Think envoie un prompt et retourne la réponse brute
func (c *OllamaClient) Think(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	payload := ollamaRequest{
		Model:  c.model,
		Stream: false,
		Messages: []ollamaMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	// Crée la requête HTTP avec le context
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/api/chat",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama status: %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(raw, &ollamaResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	return ollamaResp.Message.Content, nil
}

// ThinkWithDecision combine Think + ParseDecision avec retries
func (c *OllamaClient) ThinkWithDecision(
	ctx context.Context,
	systemPrompt, userPrompt string,
	retry RetryConfig,
) (Decision, int, error) {

	var lastErr error

	for attempt := 1; attempt <= retry.MaxAttempts; attempt++ {
		// Vérifie si le context est annulé avant chaque tentative
		select {
		case <-ctx.Done():
			return fallbackDecision("context cancelled"), attempt, ctx.Err()
		default:
		}

		raw, err := c.Think(ctx, systemPrompt, userPrompt)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: %w", attempt, err)
			fmt.Printf("[ LLM ] Tentative %d/%d échouée : %v\n",
				attempt, retry.MaxAttempts, err)

			if attempt < retry.MaxAttempts {
				time.Sleep(retry.Delay)
			}
			continue
		}

		decision, err := ParseDecision(raw)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d parse: %w", attempt, err)
			fmt.Printf("[ LLM ] Tentative %d/%d — JSON invalide, on retente...\n",
				attempt, retry.MaxAttempts)

			if attempt < retry.MaxAttempts {
				time.Sleep(retry.Delay)
			}
			continue
		}

		// Succès
		if attempt > 1 {
			fmt.Printf("[ LLM ] Succès à la tentative %d\n", attempt)
		}
		return decision, attempt, nil
	}

	// Toutes les tentatives ont échoué — fallback
	fmt.Printf("[ LLM ] Toutes les tentatives échouées — fallback décision\n")
	return fallbackDecision("all attempts failed"), retry.MaxAttempts, lastErr
}
