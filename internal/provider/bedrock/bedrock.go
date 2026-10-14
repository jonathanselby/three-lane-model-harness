package bedrock

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/jonathanselby/three-lane-model-harness/internal/provider"
)

// Provider calls Amazon Bedrock Runtime's Converse API using a Bedrock bearer token.
// Short-term Bedrock API keys inherit IAM permissions and are the intended production
// credential for this small HTTP adapter. SigV4 credential acquisition is deliberately
// outside the harness.
type Provider struct {
	Region      string
	BaseURL     string
	BearerToken string
	Client      *http.Client
}

func (p Provider) Name() string { return "bedrock" }

type requestBody struct {
	Messages []message `json:"messages"`
}

type message struct {
	Role    string         `json:"role"`
	Content []contentBlock `json:"content"`
}

type contentBlock struct {
	Text string `json:"text"`
}

type responseBody struct {
	Output struct {
		Message message `json:"message"`
	} `json:"output"`
	Usage struct {
		InputTokens  int `json:"inputTokens"`
		OutputTokens int `json:"outputTokens"`
	} `json:"usage"`
}

func (p Provider) Generate(ctx context.Context, model, prompt string) (provider.Result, error) {
	if strings.TrimSpace(p.BearerToken) == "" {
		return provider.Result{}, errors.New("bedrock bearer token is required")
	}
	base := strings.TrimRight(strings.TrimSpace(p.BaseURL), "/")
	if base == "" {
		region := strings.TrimSpace(p.Region)
		if region == "" {
			return provider.Result{}, errors.New("bedrock region is required")
		}
		base = "https://bedrock-runtime." + region + ".amazonaws.com"
	}

	payload, err := json.Marshal(requestBody{Messages: []message{{
		Role:    "user",
		Content: []contentBlock{{Text: prompt}},
	}}})
	if err != nil {
		return provider.Result{}, fmt.Errorf("encode bedrock request: %w", err)
	}
	endpoint := base + "/model/" + url.PathEscape(model) + "/converse"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return provider.Result{}, fmt.Errorf("build bedrock request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.BearerToken)

	resp, err := p.client().Do(req)
	if err != nil {
		return provider.Result{}, fmt.Errorf("call bedrock: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return provider.Result{}, fmt.Errorf("read bedrock response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return provider.Result{}, fmt.Errorf("bedrock returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var decoded responseBody
	if err := json.Unmarshal(body, &decoded); err != nil {
		return provider.Result{}, fmt.Errorf("decode bedrock response: %w", err)
	}
	var output string
	for _, block := range decoded.Output.Message.Content {
		if strings.TrimSpace(block.Text) != "" {
			output = block.Text
			break
		}
	}
	if output == "" {
		return provider.Result{}, errors.New("bedrock response contained no text output")
	}
	return provider.Result{Output: output, InputTokens: decoded.Usage.InputTokens, OutputTokens: decoded.Usage.OutputTokens}, nil
}

func (p Provider) client() *http.Client {
	if p.Client != nil {
		return p.Client
	}
	return http.DefaultClient
}
