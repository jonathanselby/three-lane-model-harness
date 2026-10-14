package harness

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jonathanselby/three-lane-model-harness/internal/provider"
)

var requestCounter uint64

// Router maps an explicit complexity class to one configured provider/model lane.
type Router struct {
	config    Config
	providers map[string]provider.Provider
	now       func() time.Time
}

func New(config Config, providers ...provider.Provider) (*Router, error) {
	registry := make(map[string]provider.Provider, len(providers))
	for _, p := range providers {
		if p == nil {
			return nil, errors.New("provider cannot be nil")
		}
		name := p.Name()
		if name == "" {
			return nil, errors.New("provider name cannot be empty")
		}
		if _, exists := registry[name]; exists {
			return nil, fmt.Errorf("duplicate provider %q", name)
		}
		registry[name] = p
	}
	return &Router{config: config, providers: registry, now: time.Now}, nil
}

func (r *Router) laneFor(c Complexity) (Lane, error) {
	switch c {
	case Simple:
		return r.config.Simple, nil
	case Standard:
		return r.config.Standard, nil
	case Complex:
		return r.config.Complex, nil
	default:
		return Lane{}, fmt.Errorf("invalid complexity %q", c)
	}
}

func (r *Router) Run(ctx context.Context, req Request) (Result, error) {
	started := r.now()
	trace := Trace{
		Timestamp:  started.UTC(),
		RequestID:  fmt.Sprintf("req-%d-%d", started.UnixNano(), atomic.AddUint64(&requestCounter, 1)),
		Complexity: req.Complexity,
		Sensitive:  req.Sensitive,
	}

	if req.Task == "" {
		err := errors.New("task cannot be empty")
		trace.Error = err.Error()
		trace.DurationMS = r.now().Sub(started).Milliseconds()
		return Result{Trace: trace}, err
	}

	lane, err := r.laneFor(req.Complexity)
	if err != nil {
		trace.Error = err.Error()
		trace.DurationMS = r.now().Sub(started).Milliseconds()
		return Result{Trace: trace}, err
	}
	trace.Provider = lane.Provider
	trace.Model = lane.Model

	if lane.Provider == "" || lane.Model == "" {
		err := fmt.Errorf("lane %q is not fully configured", req.Complexity)
		trace.Error = err.Error()
		trace.DurationMS = r.now().Sub(started).Milliseconds()
		return Result{Trace: trace}, err
	}

	p, ok := r.providers[lane.Provider]
	if !ok {
		err := fmt.Errorf("provider %q is not registered", lane.Provider)
		trace.Error = err.Error()
		trace.DurationMS = r.now().Sub(started).Milliseconds()
		return Result{Trace: trace}, err
	}

	providerResult, err := p.Generate(ctx, lane.Model, req.Task)
	trace.DurationMS = r.now().Sub(started).Milliseconds()
	trace.InputTokens = providerResult.InputTokens
	trace.OutputTokens = providerResult.OutputTokens
	if err != nil {
		trace.Error = err.Error()
		return Result{Trace: trace}, err
	}
	trace.Success = true
	return Result{Output: providerResult.Output, Trace: trace}, nil
}
