package azure

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateWithAPIKey(t *testing.T) {
	var gotPath, gotKey, gotBody string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.Header.Get("api-key")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"azure ok"}}],"usage":{"prompt_tokens":11,"completion_tokens":7}}`))
	}))
	defer ts.Close()

	p := Provider{Endpoint: ts.URL, APIKey: "secret", Client: ts.Client()}
	got, err := p.Generate(context.Background(), "deployment-a", "hello")
	if err != nil { t.Fatal(err) }
	if gotPath != "/openai/v1/chat/completions" { t.Fatalf("path=%q", gotPath) }
	if gotKey != "secret" { t.Fatalf("api-key=%q", gotKey) }
	if !strings.Contains(gotBody, `"model":"deployment-a"`) || !strings.Contains(gotBody, `"content":"hello"`) { t.Fatalf("body=%s", gotBody) }
	if got.Output != "azure ok" || got.InputTokens != 11 || got.OutputTokens != 7 { t.Fatalf("result=%+v", got) }
}

func TestGeneratePrefersBearerToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token" { t.Fatalf("authorization=%q", got) }
		if got := r.Header.Get("api-key"); got != "" { t.Fatalf("unexpected api-key=%q", got) }
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer ts.Close()
	_, err := (Provider{Endpoint: ts.URL, APIKey: "key", BearerToken: "token", Client: ts.Client()}).Generate(context.Background(), "m", "p")
	if err != nil { t.Fatal(err) }
}

func TestGenerateRejectsHTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "denied", http.StatusUnauthorized) }))
	defer ts.Close()
	_, err := (Provider{Endpoint: ts.URL, APIKey: "key", Client: ts.Client()}).Generate(context.Background(), "m", "p")
	if err == nil || !strings.Contains(err.Error(), "401") { t.Fatalf("err=%v", err) }
}
