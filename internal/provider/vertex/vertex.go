package vertex

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

// Provider calls Google Vertex AI's v1 generateContent API for Gemini models.
// The caller supplies a short-lived Google Cloud access token; token acquisition
// through ADC / workload identity belongs to the deployment environment.
type Provider struct {
	ProjectID   string
	Location    string
	BaseURL     string
	AccessToken string
	Client      *http.Client
}

func (p Provider) Name() string { return "vertex" }

type requestBody struct {
	Contents []content `json:"contents"`
}

type content struct {
	Role  string `json:"role"`
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type responseBody struct {
	Candidates []struct {
		Content content `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
}

func (p Provider) Generate(ctx context.Context, model, prompt string) (provider.Result, error) {
	project := strings.TrimSpace(p.ProjectID)
	location := strings.TrimSpace(p.Location)
	if project == "" {
		return provider.Result{}, errors.New("vertex project ID is required")
	}
	if location == "" {
		location = "global"
	}
	if strings.TrimSpace(p.AccessToken) == "" {
		return provider.Result{}, errors.New("vertex access token is required")
	}
	base := strings.TrimRight(strings.TrimSpace(p.BaseURL), "/")
	if base == "" {
		if location == "global" {
			base = "https://aiplatform.googleapis.com"
		} else {
			base = "https://" + location + "-aiplatform.googleapis.com"
		}
	}

	payload, err := json.Marshal(requestBody{Contents: []content{{
		Role:  "user",
		Parts: []part{{Text: prompt}},
	}}})
	if err != nil {
		return provider.Result{}, fmt.Errorf("encode vertex request: %w", err)
	}
	endpoint := fmt.Sprintf("%s/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
		base, url.PathEscape(project), url.PathEscape(location), url.PathEscape(model))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return provider.Result{}, fmt.Errorf("build vertex request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.AccessToken)

	resp, err := p.client().Do(req)
	if err != nil {
		return provider.Result{}, fmt.Errorf("call vertex: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return provider.Result{}, fmt.Errorf("read vertex response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return provider.Result{}, fmt.Errorf("vertex returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var decoded responseBody
	if err := json.Unmarshal(body, &decoded); err != nil {
		return provider.Result{}, fmt.Errorf("decode vertex response: %w", err)
	}
	for _, candidate := range decoded.Candidates {
		for _, part := range candidate.Content.Parts {
			if strings.TrimSpace(part.Text) != "" {
				return provider.Result{Output: part.Text, InputTokens: decoded.UsageMetadata.PromptTokenCount, OutputTokens: decoded.UsageMetadata.CandidatesTokenCount}, nil
			}
		}
	}
	return provider.Result{}, errors.New("vertex response contained no text output")
}

func (p Provider) client() *http.Client {
	if p.Client != nil {
		return p.Client
	}
	return http.DefaultClient
}
