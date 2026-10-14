package harness

import "time"

// Complexity is an explicit workload classification supplied by the caller.
type Complexity string

const (
	Simple   Complexity = "simple"
	Standard Complexity = "standard"
	Complex  Complexity = "complex"
)

func (c Complexity) Valid() bool {
	switch c {
	case Simple, Standard, Complex:
		return true
	default:
		return false
	}
}

// Request is the work submitted to the harness.
type Request struct {
	Task       string
	Complexity Complexity
	Sensitive  bool
}

// Lane identifies the configured provider and model for a complexity class.
type Lane struct {
	Provider string
	Model    string
}

// Config contains the three explicit routing lanes.
type Config struct {
	Simple   Lane
	Standard Lane
	Complex  Lane
}

// Trace is the minimal execution record emitted for every attempted request.
type Trace struct {
	Timestamp    time.Time  `json:"timestamp"`
	RequestID    string     `json:"request_id"`
	Complexity   Complexity `json:"complexity"`
	Sensitive    bool       `json:"sensitive"`
	Provider     string     `json:"provider"`
	Model        string     `json:"model"`
	DurationMS   int64      `json:"duration_ms"`
	Success      bool       `json:"success"`
	Error        string     `json:"error,omitempty"`
	InputTokens  int        `json:"input_tokens,omitempty"`
	OutputTokens int        `json:"output_tokens,omitempty"`
}

// Result combines model output with the authoritative execution trace.
type Result struct {
	Output string `json:"output"`
	Trace  Trace  `json:"trace"`
}
