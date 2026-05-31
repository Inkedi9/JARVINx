package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Inkedi9/jarvinx/jxlog"
)

type OllamaClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
	circuit    *CircuitBreaker
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
		circuit: DefaultCircuitBreaker(),
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

	// Vérifie le circuit breaker avant tout
	if err := c.circuit.Allow(); err != nil {
		jxlog.Warn("LLM", fmt.Sprintf("Circuit breaker %s — appel bloqué", c.circuit.State()))
		return fallbackDecision("circuit breaker open"), 0, err
	}

	var lastErr error

	for attempt := 1; attempt <= retry.MaxAttempts; attempt++ {
		// Vérifie si le context est annulé avant chaque tentative
		select {
		case <-ctx.Done():
			c.circuit.RecordFailure()
			return fallbackDecision("context cancelled"), attempt, ctx.Err()
		default:
		}

		raw, err := c.Think(ctx, systemPrompt, userPrompt)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: %w", attempt, err)
			c.circuit.RecordFailure()
			jxlog.Warn("LLM", fmt.Sprintf("Tentative %d/%d échouée : %v",
				attempt, retry.MaxAttempts, err))

			if attempt < retry.MaxAttempts {
				time.Sleep(retry.Delay)
			}
			continue
		}

		decision, err := ParseDecision(raw)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d parse: %w", attempt, err)
			jxlog.Warn("LLM", fmt.Sprintf("Tentative %d/%d — JSON invalide, on retente...",
				attempt, retry.MaxAttempts))

			if attempt < retry.MaxAttempts {
				time.Sleep(retry.Delay)
			}
			continue
		}

		// Succès
		c.circuit.RecordSuccess()
		if attempt > 1 {
			jxlog.Info("LLM", fmt.Sprintf("Succès à la tentative %d", attempt))
		}
		return decision, attempt, nil
	}

	// Toutes les tentatives ont échoué — fallback
	c.circuit.RecordFailure()
	jxlog.Warn("LLM", "Toutes les tentatives échouées — fallback décision")
	return fallbackDecision("all attempts failed"), retry.MaxAttempts, lastErr
}

// CircuitStats expose les stats du circuit breaker
func (c *OllamaClient) CircuitStats() CircuitStats {
	return c.circuit.Stats()
}
