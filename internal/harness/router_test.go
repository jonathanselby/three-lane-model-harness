package harness

import (
	"context"
	"errors"
	"testing"

	"github.com/jonathanselby/three-lane-model-harness/internal/provider"
	"github.com/jonathanselby/three-lane-model-harness/internal/provider/mock"
)

type failingProvider struct{}

func (failingProvider) Name() string { return "failing" }
func (failingProvider) Generate(context.Context, string, string) (provider.Result, error) {
	return provider.Result{}, errors.New("provider failed")
}

func testConfig() Config {
	return Config{
		Simple:   Lane{Provider: "mock", Model: "small"},
		Standard: Lane{Provider: "mock", Model: "normal"},
		Complex:  Lane{Provider: "mock", Model: "strong"},
	}
}

func TestRoutesAllThreeLanes(t *testing.T) {
	r, err := New(testConfig(), mock.Provider{})
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		complexity Complexity
		model      string
	}{
		{Simple, "small"},
		{Standard, "normal"},
		{Complex, "strong"},
	}

	for _, tc := range cases {
		t.Run(string(tc.complexity), func(t *testing.T) {
			got, err := r.Run(context.Background(), Request{Task: "hello", Complexity: tc.complexity})
			if err != nil {
				t.Fatal(err)
			}
			if got.Trace.Provider != "mock" || got.Trace.Model != tc.model {
				t.Fatalf("got provider/model %s/%s", got.Trace.Provider, got.Trace.Model)
			}
			if !got.Trace.Success {
				t.Fatal("expected successful trace")
			}
		})
	}
}

func TestSensitiveDoesNotChangeRoute(t *testing.T) {
	r, _ := New(testConfig(), mock.Provider{})
	got, err := r.Run(context.Background(), Request{Task: "hello", Complexity: Simple, Sensitive: true})
	if err != nil {
		t.Fatal(err)
	}
	if got.Trace.Model != "small" || !got.Trace.Sensitive {
		t.Fatalf("unexpected trace: %+v", got.Trace)
	}
}

func TestInvalidComplexityFails(t *testing.T) {
	r, _ := New(testConfig(), mock.Provider{})
	got, err := r.Run(context.Background(), Request{Task: "hello", Complexity: Complexity("mystery")})
	if err == nil {
		t.Fatal("expected error")
	}
	if got.Trace.Success {
		t.Fatal("failed execution must not be marked successful")
	}
}

func TestMissingProviderFails(t *testing.T) {
	cfg := testConfig()
	cfg.Complex.Provider = "missing"
	r, _ := New(cfg, mock.Provider{})
	got, err := r.Run(context.Background(), Request{Task: "hello", Complexity: Complex})
	if err == nil {
		t.Fatal("expected error")
	}
	if got.Trace.Provider != "missing" {
		t.Fatalf("expected configured provider in trace, got %q", got.Trace.Provider)
	}
}

func TestProviderFailureIsTraced(t *testing.T) {
	cfg := testConfig()
	cfg.Standard.Provider = "failing"
	r, _ := New(cfg, mock.Provider{}, failingProvider{})
	got, err := r.Run(context.Background(), Request{Task: "hello", Complexity: Standard})
	if err == nil {
		t.Fatal("expected error")
	}
	if got.Trace.Error != "provider failed" || got.Trace.Success {
		t.Fatalf("unexpected trace: %+v", got.Trace)
	}
}

func TestEmptyTaskFailsBeforeProviderCall(t *testing.T) {
	r, _ := New(testConfig(), mock.Provider{})
	got, err := r.Run(context.Background(), Request{Complexity: Simple})
	if err == nil {
		t.Fatal("expected error")
	}
	if got.Trace.Provider != "" {
		t.Fatal("provider should not be selected for empty task")
	}
}

func TestDuplicateProviderRejected(t *testing.T) {
	_, err := New(testConfig(), mock.Provider{}, mock.Provider{})
	if err == nil {
		t.Fatal("expected duplicate provider error")
	}
}
