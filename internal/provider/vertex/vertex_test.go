package vertex

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
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"vertex ok"}]}}],"usageMetadata":{"promptTokenCount":17,"candidatesTokenCount":6}}`))
	}))
	defer ts.Close()

	p := Provider{ProjectID: "proj", Location: "australia-southeast1", BaseURL: ts.URL, AccessToken: "g-token", Client: ts.Client()}
	got, err := p.Generate(context.Background(), "gemini-2.5-flash", "hello")
	if err != nil { t.Fatal(err) }
	wantPath := "/v1/projects/proj/locations/australia-southeast1/publishers/google/models/gemini-2.5-flash:generateContent"
	if gotPath != wantPath { t.Fatalf("path=%q want=%q", gotPath, wantPath) }
	if gotAuth != "Bearer g-token" { t.Fatalf("auth=%q", gotAuth) }
	if !strings.Contains(gotBody, `"role":"user"`) || !strings.Contains(gotBody, `"text":"hello"`) { t.Fatalf("body=%s", gotBody) }
	if got.Output != "vertex ok" || got.InputTokens != 17 || got.OutputTokens != 6 { t.Fatalf("result=%+v", got) }
}

func TestGenerateRequiresProject(t *testing.T) {
	_, err := (Provider{AccessToken: "x"}).Generate(context.Background(), "m", "p")
	if err == nil || !strings.Contains(err.Error(), "project") { t.Fatalf("err=%v", err) }
}
