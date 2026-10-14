# Three-Lane Model Harness — Engineering Design Document

## 1. Purpose

The Three-Lane Model Harness is a small Go program that routes an AI task to one of three configured model lanes based on a complexity value supplied by the caller.

The core design rule is deliberately narrow:

> **The caller classifies. The harness routes.**

This repository demonstrates explicit workload classification, provider abstraction, structured requests/results, configuration, traceability and failure handling without embedding proprietary model-selection or security logic.

## 2. Non-goals

The harness does not:

- decide whether work is simple, standard or complex;
- infer risk or sensitivity;
- enforce security policy;
- retry or fail over between providers;
- orchestrate agents;
- persist requests or traces;
- optimise cost automatically;
- choose providers autonomously.

## 3. Component model

```text
                         CALLER
                           |
                           | Request{task, complexity, sensitive}
                           v
                 +--------------------+
                 |       ROUTER       |
                 | laneFor(complexity)|
                 +---------+----------+
                           |
             +-------------+-------------+
             |             |             |
             v             v             v
        +---------+    +---------+    +---------+
        | SIMPLE  |    |STANDARD |    | COMPLEX |
        |  Lane   |    |  Lane   |    |  Lane   |
        +----+----+    +----+----+    +----+----+
             |              |              |
             +--------------+--------------+
                            |
                            v
                  +-------------------+
                  | Provider Registry |
                  +---------+---------+
                            |
                 +----------+----------+
                 |                     |
                 v                     v
          +-------------+       +-------------+
          |    Mock     |       |   Ollama    |
          |  Provider   |       |  Provider   |
          +-------------+       +-------------+
```

### Components

| Component | Responsibility |
|---|---|
| CLI (`cmd/harness`) | Parse user input and lane configuration; set timeout; print JSON result |
| Router (`internal/harness`) | Validate request, select lane, resolve provider, invoke provider, produce trace |
| Provider contract (`internal/provider`) | Stable interface between router and provider-specific transports |
| Mock provider | Deterministic local implementation for tests and examples |
| Ollama provider | HTTP implementation for an Ollama-compatible `/api/generate` endpoint |

## 4. Request model

```go
type Request struct {
    Task       string
    Complexity Complexity
    Sensitive  bool
}
```

`Complexity` must be one of `simple`, `standard`, or `complex`.

`Sensitive` is metadata supplied by the caller. It is written to the trace but has **no routing or enforcement effect**.

## 5. Routing behaviour

Configuration contains exactly one lane per complexity class:

```go
type Config struct {
    Simple   Lane
    Standard Lane
    Complex  Lane
}

type Lane struct {
    Provider string
    Model    string
}
```

Routing is a direct lookup by explicit complexity. There is no scoring, model evaluation or hidden classifier.

```text
simple   -> Config.Simple
standard -> Config.Standard
complex  -> Config.Complex
```

An unknown complexity is rejected before provider invocation.

## 6. Provider abstraction

Providers implement:

```go
type Provider interface {
    Name() string
    Generate(ctx context.Context, model, prompt string) (Result, error)
}
```

The router maintains a registry keyed by provider name. Lane configuration identifies the provider and model. Callers do not need provider-specific request types.

V0.1 includes:

- `mock` — deterministic, local, no credentials;
- `ollama` — JSON/HTTP call to `/api/generate`.

## 7. Runtime flow

```text
CLI INPUT
   |
   v
PARSE FLAGS
   |
   v
BUILD THREE-LANE CONFIG
   |
   v
CREATE PROVIDER REGISTRY
   |
   v
CREATE REQUEST + TIMEOUT CONTEXT
   |
   v
ROUTER.RUN
   |
   +-- empty task ----------> FAIL + TRACE
   |
   +-- invalid complexity --> FAIL + TRACE
   |
   v
SELECT CONFIGURED LANE
   |
   +-- incomplete lane -----> FAIL + TRACE
   |
   v
RESOLVE PROVIDER
   |
   +-- missing provider ----> FAIL + TRACE
   |
   v
PROVIDER.GENERATE
   |
   +-- provider error ------> FAIL + TRACE
   |
   v
SUCCESS RESULT + TRACE
```

## 8. Configuration

V0.1 uses CLI flags rather than a configuration file.

Each lane has two settings:

```text
--simple-provider
--simple-model
--standard-provider
--standard-model
--complex-provider
--complex-model
```

Additional runtime settings include:

```text
--ollama-url
--timeout
--sensitive
```

Defaults point all three lanes to the mock provider so the program runs immediately after checkout.

## 9. Logging and trace

Every attempted request returns a `Trace` object containing:

```text
timestamp
request_id
complexity
sensitive
provider
model
duration_ms
success / failure
error (when present)
input_tokens (when provider supplies them)
output_tokens (when provider supplies them)
```

The trace is part of the returned result rather than a separate observability system.

No prompts or outputs are independently persisted by the harness.

## 10. Failure handling

| Failure | Behaviour |
|---|---|
| empty task | reject before lane/provider selection; return failed trace |
| invalid complexity | reject; return failed trace |
| empty provider/model in selected lane | reject; return failed trace |
| provider not registered | reject; return failed trace with configured provider/model |
| provider error | return error and failed trace |
| context timeout/cancellation | provider receives context; resulting error is traced |
| Ollama non-2xx | provider returns explicit error including status/body excerpt |
| malformed Ollama response | provider returns decode error |

V0.1 deliberately does not retry or fail over. Failure is explicit rather than hidden.

## 11. Test strategy

Unit tests cover:

- simple lane routing;
- standard lane routing;
- complex lane routing;
- sensitive metadata without routing change;
- invalid complexity;
- missing provider;
- provider failure;
- empty task;
- duplicate provider registration;
- Ollama-compatible successful response;
- Ollama non-2xx response.

The repository is also run with the mock provider for all three demonstration tasks.

## 12. Security boundary

This is a model-routing harness, not a security control.

Important boundaries:

- user-provided prompts may be sent to whichever provider the selected lane names;
- `Sensitive=true` is **only metadata** and does not block external providers;
- Ollama URL is caller-configurable and therefore defines an outbound trust boundary;
- credentials are not implemented in V0.1;
- no prompt filtering, injection defence, policy enforcement or data-loss prevention is present;
- traces contain metadata and provider token counts but the harness does not separately persist prompt/output content.

The absence of security enforcement is deliberate and should not be interpreted as a security guarantee.

## 13. Known limitations

- classification quality is wholly the caller's responsibility;
- lane configuration is supplied at process startup, not dynamically reloaded;
- no retries, failover or circuit breaker;
- no pricing/cost calculation;
- no persistent trace store;
- no streaming provider contract;
- mock token counts are deterministic character counts, not model tokenisation;
- Ollama is the only real provider implementation in V0.1;
- request IDs are process-local timestamp/counter identifiers, not globally coordinated IDs.
