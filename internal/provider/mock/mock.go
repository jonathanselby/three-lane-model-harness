package mock

import (
	"context"
	"fmt"

	"github.com/jonathanselby/three-lane-model-harness/internal/provider"
)

// Provider is a deterministic provider for local runs and tests.
type Provider struct{}

func (Provider) Name() string { return "mock" }

func (Provider) Generate(_ context.Context, model, prompt string) (provider.Result, error) {
	return provider.Result{
		Output:       fmt.Sprintf("mock[%s]: %s", model, prompt),
		InputTokens:  len([]rune(prompt)),
		OutputTokens: len([]rune(prompt)) + len([]rune(model)) + 8,
	}, nil
}
