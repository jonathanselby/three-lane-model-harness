package bedrock

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	var gotPath, gotAuth, gotBody string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output":{"message":{"role":"assistant","content":[{"text":"bedrock ok"}]}},"usage":{"inputTokens":13,"outputTokens":5}}`))
	}))
	defer ts.Close()

	p := Provider{BaseURL: ts.URL, BearerToken: "br-token", Client: ts.Client()}
	got, err := p.Generate(context.Background(), "us.anthropic.claude-sonnet-4-6", "hello")
	if err != nil { t.Fatal(err) }
	if gotPath != "/model/us.anthropic.claude-sonnet-4-6/converse" { t.Fatalf("path=%q", gotPath) }
	if gotAuth != "Bearer br-token" { t.Fatalf("auth=%q", gotAuth) }
	if !strings.Contains(gotBody, `"role":"user"`) || !strings.Contains(gotBody, `"text":"hello"`) { t.Fatalf("body=%s", gotBody) }
	if got.Output != "bedrock ok" || got.InputTokens != 13 || got.OutputTokens != 5 { t.Fatalf("result=%+v", got) }
}

func TestGenerateRequiresToken(t *testing.T) {
	_, err := (Provider{Region: "us-east-1"}).Generate(context.Background(), "m", "p")
	if err == nil || !strings.Contains(err.Error(), "token") { t.Fatalf("err=%v", err) }
}
