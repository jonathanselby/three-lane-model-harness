package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jonathanselby/three-lane-model-harness/internal/provider"
)

// Provider calls the Microsoft Foundry / Azure OpenAI v1 Chat Completions API.
// Endpoint should be the Foundry/Azure resource base URL. Authentication can use
// either an API key or a short-lived Microsoft Entra bearer token.
type Provider struct {
	Endpoint    string
	APIKey      string
	BearerToken string
	Client      *http.Client
}

func (p Provider) Name() string { return "azure" }

type requestBody struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseBody struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func (p Provider) Generate(ctx context.Context, model, prompt string) (provider.Result, error) {
	endpoint := strings.TrimRight(strings.TrimSpace(p.Endpoint), "/")
	if endpoint == "" {
		return provider.Result{}, errors.New("azure endpoint is required")
	}
	if p.APIKey == "" && p.BearerToken == "" {
		return provider.Result{}, errors.New("azure authentication is required")
	}
	if !strings.HasSuffix(endpoint, "/openai/v1") {
		endpoint += "/openai/v1"
	}

	payload, err := json.Marshal(requestBody{
		Model:    model,
		Messages: []message{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return provider.Result{}, fmt.Errorf("encode azure request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return provider.Result{}, fmt.Errorf("build azure request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if p.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.BearerToken)
	} else {
		req.Header.Set("api-key", p.APIKey)
	}

	resp, err := p.client().Do(req)
	if err != nil {
		return provider.Result{}, fmt.Errorf("call azure: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return provider.Result{}, fmt.Errorf("read azure response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return provider.Result{}, fmt.Errorf("azure returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var decoded responseBody
	if err := json.Unmarshal(body, &decoded); err != nil {
		return provider.Result{}, fmt.Errorf("decode azure response: %w", err)
	}
	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return provider.Result{}, errors.New("azure response contained no text output")
	}
	return provider.Result{
		Output:       decoded.Choices[0].Message.Content,
		InputTokens:  decoded.Usage.PromptTokens,
		OutputTokens: decoded.Usage.CompletionTokens,
	}, nil
}

func (p Provider) client() *http.Client {
	if p.Client != nil {
		return p.Client
	}
	return http.DefaultClient
}
