package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jonathanselby/three-lane-model-harness/internal/provider"
)

// Provider calls a local or remote Ollama-compatible /api/generate endpoint.
type Provider struct {
	BaseURL string
	Client  *http.Client
}

func (p Provider) Name() string { return "ollama" }

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type generateResponse struct {
	Response        string `json:"response"`
	PromptEvalCount int    `json:"prompt_eval_count"`
	EvalCount       int    `json:"eval_count"`
}

func (p Provider) Generate(ctx context.Context, model, prompt string) (provider.Result, error) {
	base := strings.TrimRight(p.BaseURL, "/")
	if base == "" {
		base = "http://localhost:11434"
	}
	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}

	payload, err := json.Marshal(generateRequest{Model: model, Prompt: prompt, Stream: false})
	if err != nil {
		return provider.Result{}, fmt.Errorf("encode ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/generate", bytes.NewReader(payload))
	if err != nil {
		return provider.Result{}, fmt.Errorf("build ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return provider.Result{}, fmt.Errorf("call ollama: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return provider.Result{}, fmt.Errorf("read ollama response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return provider.Result{}, fmt.Errorf("ollama returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var decoded generateResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return provider.Result{}, fmt.Errorf("decode ollama response: %w", err)
	}

	return provider.Result{
		Output:       decoded.Response,
		InputTokens:  decoded.PromptEvalCount,
		OutputTokens: decoded.EvalCount,
	}, nil
}
