# Three-Lane Model Harness

A deliberately small Go harness for routing explicitly classified AI work across simple, standard and complex model lanes.

The core rule is:

> **Do not start by choosing a model. Start by classifying the work.**

And the implementation boundary is equally important:

> **The caller classifies. The harness routes.**

There is no hidden complexity classifier in this repository. A caller submits a task with `simple`, `standard` or `complex`; the harness routes it to the provider/model configured for that lane and returns the output with a minimal execution trace.

## Why

Model choice is often allowed to leak through application code. That makes provider changes, cost decisions and capability upgrades harder than they need to be.

This harness keeps three concerns separate:

```text
WORKLOAD DECISION          ROUTING                 MODEL EXECUTION
(caller responsibility)    (this harness)          (provider responsibility)

      simple  -----------+      +-------------> configured provider/model
      standard ----------+----->|
      complex -----------+      +-------------> configured provider/model
```

The design is deliberately boring. That is the point.

## What it demonstrates

- explicit workload classification;
- Go interfaces and small data structures;
- provider abstraction;
- configuration by lane;
- structured request/result contracts;
- failure-first behaviour;
- minimal traceability;
- a local deterministic provider;
- an Ollama-compatible provider;
- product intent (`PRD.md`) -> design (`EDD.md`) -> implementation -> tests;
- architecture recovery by `codebase-to-edd` and comparison with the intended design.

## What it does not do

This is **not** an automatic model selector, security gateway or agent framework.

It does not contain:

- automatic classification;
- risk scoring;
- policy enforcement;
- prompt-injection controls;
- autonomous provider choice;
- retries/failover;
- proprietary routing logic.

`Sensitive` is recorded as caller-supplied metadata. It does **not** change routing or enforce a policy.

## Architecture

```text
                           REQUEST
                              |
                 caller supplies complexity
                              |
                              v
                     +----------------+
                     |     ROUTER     |
                     +-------+--------+
                             |
              +--------------+--------------+
              |              |              |
              v              v              v
         +---------+     +---------+     +---------+
         | SIMPLE  |     |STANDARD |     | COMPLEX |
         +----+----+     +----+----+     +----+----+
              |               |               |
              +---------------+---------------+
                              |
                              v
                    +-------------------+
                    | PROVIDER REGISTRY |
                    +---------+---------+
                              |
                    +---------+---------+
                    |                   |
                    v                   v
                 +------+            +--------+
                 | MOCK |            | OLLAMA |
                 +------+            +--------+
```

See [`EDD.md`](EDD.md) for the intended design and [`architecture.txt`](architecture.txt) for the standalone plain-text diagrams.

## Run it

Requires Go 1.23+.

The defaults use the mock provider, so no API key or model server is required.

```bash
go run ./cmd/harness \
  -complexity simple \
  -prompt "Extract the invoice number: INV-20491"
```

Representative result:

```json
{
  "output": "mock[mock-small]: Extract the invoice number: INV-20491",
  "trace": {
    "complexity": "simple",
    "provider": "mock",
    "model": "mock-small",
    "success": true
  }
}
```

### Standard lane

```bash
go run ./cmd/harness \
  -complexity standard \
  -prompt "Summarise these paragraphs and identify the main risks."
```

### Complex lane

```bash
go run ./cmd/harness \
  -complexity complex \
  -prompt "Compare these two architecture proposals and explain the trade-offs."
```

Verified mock runs for all three lanes are captured in [`examples/mock-runs.md`](examples/mock-runs.md).

## Use Ollama

Any lane can be pointed at the Ollama provider without changing caller behaviour.

```bash
go run ./cmd/harness \
  -complexity simple \
  -simple-provider ollama \
  -simple-model qwen3:8b \
  -prompt "Extract the invoice number: INV-20491"
```

The default endpoint is `http://localhost:11434`; override it with `-ollama-url`.

## Configure different models per lane

```bash
go run ./cmd/harness \
  -complexity complex \
  -simple-provider ollama \
  -simple-model qwen3:8b \
  -standard-provider ollama \
  -standard-model qwen3:14b \
  -complex-provider ollama \
  -complex-model qwen3:30b \
  -prompt "Compare these two architecture proposals and explain the trade-offs."
```

The model names above are examples. The harness does not decide whether those models are appropriate.

## Test it

```bash
go test ./...
```

The tests cover all three lanes plus invalid classification, missing providers, provider failures, empty input, duplicate registration and Ollama HTTP behaviour. GitHub Actions runs the same test command on pushes and pull requests to `main`.

## Trace

Each execution records:

```text
timestamp
request_id
complexity
sensitive
provider
model
duration_ms
success / failure
error (when applicable)
input_tokens / output_tokens (when supplied by the provider)
```

This is deliberately enough to inspect model routing without building an observability platform.

## Repository map

```text
three-lane-model-harness/
├── README.md
├── PRD.md
├── EDD.md
├── architecture.txt
├── LICENSE
├── go.mod
├── .github/
│   └── workflows/
│       └── test.yml
├── cmd/
│   └── harness/
│       └── main.go
├── internal/
│   ├── harness/
│   │   ├── types.go
│   │   ├── router.go
│   │   └── router_test.go
│   └── provider/
│       ├── provider.go
│       ├── mock/
│       │   └── mock.go
│       └── ollama/
│           ├── ollama.go
│           └── ollama_test.go
├── examples/
│   └── mock-runs.md
└── generated/
    ├── EDD.md
    └── COMPARISON.md
```

## Design cross-check

This repository is also a reference test for [`codebase-to-edd`](https://github.com/jonathanselby/codebase-to-edd).

[`EDD.md`](EDD.md) is the hand-written intended design. [`generated/EDD.md`](generated/EDD.md) was produced by inspecting the implementation as an unfamiliar codebase while deliberately excluding the hand-written EDD and PRD from its evidence basis.

[`generated/COMPARISON.md`](generated/COMPARISON.md) compares them. The runtime architecture is recovered cleanly; the material gaps are design rationale and future intent, which are exactly the information that should remain in the PRD/EDD rather than be invented from code.

## Licence

MIT. See [`LICENSE`](LICENSE).
