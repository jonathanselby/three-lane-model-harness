package provider

import "context"

// Result is the provider-level result returned to the harness.
type Result struct {
	Output       string
	InputTokens  int
	OutputTokens int
}

// Provider executes a prompt against a named model.
// Provider implementations own transport-specific behaviour; routing stays in the harness.
type Provider interface {
	Name() string
	Generate(ctx context.Context, model, prompt string) (Result, error)
}
