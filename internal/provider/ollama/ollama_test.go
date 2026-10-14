package ollama

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":"hello","prompt_eval_count":3,"eval_count":2}`))
	}))
	defer server.Close()

	p := Provider{BaseURL: server.URL, Client: server.Client()}
	got, err := p.Generate(context.Background(), "model-a", "prompt")
	if err != nil {
		t.Fatal(err)
	}
	if got.Output != "hello" || got.InputTokens != 3 || got.OutputTokens != 2 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestGenerateNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model missing", http.StatusNotFound)
	}))
	defer server.Close()

	p := Provider{BaseURL: server.URL, Client: server.Client()}
	_, err := p.Generate(context.Background(), "missing", "prompt")
	if err == nil {
		t.Fatal("expected error")
	}
}
