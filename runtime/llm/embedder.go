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

type OllamaEmbedder struct {
	baseURL    string
	model      string
	httpClient *http.Client
	circuit    *CircuitBreaker
}

type embedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embedResponse struct {
	Embedding []float32 `json:"embedding"`
}

func NewOllamaEmbedder(baseURL, model string) *OllamaEmbedder {
	return &OllamaEmbedder{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		circuit: DefaultCircuitBreaker(),
	}
}

// Embed vectorise text via Ollama /api/embeddings.
// Retourne ErrCircuitOpen si le circuit est ouvert.
func (e *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if err := e.circuit.Allow(); err != nil {
		return nil, err
	}

	body, err := json.Marshal(embedRequest{Model: e.model, Prompt: text})
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/api/embeddings", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		e.circuit.RecordFailure()
		return nil, fmt.Errorf("embed request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		e.circuit.RecordFailure()
		return nil, fmt.Errorf("ollama embed status: %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		e.circuit.RecordFailure()
		return nil, fmt.Errorf("read embed response: %w", err)
	}

	var out embedResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		e.circuit.RecordFailure()
		return nil, fmt.Errorf("unmarshal embed response: %w", err)
	}

	if len(out.Embedding) == 0 {
		e.circuit.RecordFailure()
		return nil, fmt.Errorf("empty embedding returned by model")
	}

	e.circuit.RecordSuccess()
	return out.Embedding, nil
}
