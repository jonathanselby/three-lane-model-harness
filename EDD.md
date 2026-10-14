# Three-Lane Model Harness — Engineering Design Document

**Version:** 0.2

## 1. Purpose

The Three-Lane Model Harness is a small Go program that routes an AI task to one of three configured model lanes based on a complexity value supplied by the caller.

The core design rule remains:

> **The caller classifies. The harness routes.**

V0.1 proved the pattern with `mock` and `ollama`. V0.2 extends the same provider contract to three managed enterprise AI platforms: Microsoft Foundry / Azure OpenAI, Amazon Bedrock and Google Vertex AI.

The V0.2 engineering question is not whether Go can call three more APIs. It is whether materially different cloud AI platforms can sit behind one unchanged caller/router contract.

## 2. Non-goals

The harness does not:

- decide whether work is simple, standard or complex;
- infer risk or sensitivity;
- enforce security policy;
- acquire or refresh cloud credentials;
- implement workload identity;
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
        +----+----+    +----+----+    +----+----+
             |              |              |
             +--------------+--------------+
                            |
                            v
                  +-------------------+
                  | Provider Registry |
                  +---------+---------+
                            |
       +---------+----------+----------+---------+
       |         |          |          |         |
       v         v          v          v         v
     MOCK      OLLAMA     AZURE     BEDROCK    VERTEX
       |         |          |          |         |
     local     local/     Microsoft   Amazon    Google
              remote      managed     managed   managed
```

### Components

| Component | Responsibility |
|---|---|
| CLI (`cmd/harness`) | Parse request/lane configuration, load non-secret cloud settings, read credentials from environment, create providers, set timeout, print JSON result |
| Router (`internal/harness`) | Validate request, select lane, resolve provider, invoke provider, produce trace |
| Provider contract (`internal/provider`) | Stable interface between router and provider-specific transports |
| Mock provider | Deterministic local implementation for tests and examples |
| Ollama provider | HTTP implementation for an Ollama-compatible `/api/generate` endpoint |
| Azure provider | Microsoft Foundry / Azure OpenAI v1 Chat Completions adapter |
| Bedrock provider | Amazon Bedrock Runtime Converse adapter |
| Vertex provider | Google Vertex AI `generateContent` adapter |

## 4. Request model

```go
type Request struct {
    Task       string
    Complexity Complexity
    Sensitive  bool
}
```

`Complexity` must be one of `simple`, `standard`, or `complex`.

`Sensitive` is caller metadata. It is written to the trace but has **no routing or enforcement effect**.

## 5. Routing behaviour

Configuration remains exactly one lane per complexity class:

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

Routing remains a direct lookup:

```text
simple   -> Config.Simple
standard -> Config.Standard
complex  -> Config.Complex
```

No V0.2 provider adds classification logic. An unknown complexity is rejected before provider invocation.

## 6. Provider abstraction

All providers implement the unchanged interface:

```go
type Provider interface {
    Name() string
    Generate(ctx context.Context, model, prompt string) (Result, error)
}
```

The router maintains a registry keyed by provider name. Lane configuration identifies the provider and model. The caller does not need provider-specific request types.

V0.2 providers:

| Provider name | Intended role | Transport |
|---|---|---|
| `mock` | deterministic tests/examples | in-process |
| `ollama` | local/self-hosted development | HTTP `/api/generate` |
| `azure` | Microsoft managed enterprise AI | HTTPS v1 Chat Completions |
| `bedrock` | AWS managed enterprise AI | HTTPS Bedrock Converse |
| `vertex` | Google Cloud managed enterprise AI | HTTPS Vertex `generateContent` |

## 7. Enterprise provider design

### Microsoft Foundry / Azure OpenAI

The adapter calls the v1-compatible Chat Completions endpoint and supports either:

- `AZURE_OPENAI_API_KEY`; or
- `AZURE_OPENAI_BEARER_TOKEN` for a short-lived Microsoft Entra token.

The non-secret endpoint is supplied by `-azure-endpoint` or `AZURE_OPENAI_ENDPOINT`.

The adapter sends the configured lane model/deployment name in the request body and returns the first text completion plus prompt/completion token counts.

### Amazon Bedrock

The adapter uses the Bedrock Runtime Converse API and a bearer token supplied through `AWS_BEARER_TOKEN_BEDROCK`.

The region is supplied by `-bedrock-region`, `AWS_REGION` or `AWS_DEFAULT_REGION`. An optional `-bedrock-url` exists for testing or explicit endpoint override.

Credential acquisition is outside the harness. V0.2 deliberately does not embed AWS credential resolution or SigV4 signing logic into the router.

### Google Vertex AI

The adapter uses the v1 publisher-model `generateContent` API.

Configuration:

- project: `-vertex-project` or `GOOGLE_CLOUD_PROJECT`;
- location: `-vertex-location` or `VERTEX_LOCATION`, default `global`;
- short-lived bearer token: `GOOGLE_OAUTH_ACCESS_TOKEN`.

Token acquisition through ADC, workload identity or `gcloud` remains a deployment concern.

## 8. Credential boundary

V0.2 separates routing configuration from secrets.

Provider/model names, project/region/location and endpoints can be flags. Credentials are environment variables only.

```text
DEPLOYMENT IDENTITY / SECRET STORE
              |
              | short-lived token / API key
              v
          environment
              |
              v
      provider adapter only
              |
              v
       managed endpoint
```

This does not make the harness a secret-management system. It simply avoids introducing credential flags into shell history and process arguments.

## 9. Runtime flow

```text
CLI INPUT
   |
   v
PARSE REQUEST + LANE CONFIG
   |
   v
READ PROVIDER CREDENTIALS FROM ENVIRONMENT
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
   +-- config/auth error ---> FAIL + TRACE
   +-- HTTP error ----------> FAIL + TRACE
   +-- malformed response --> FAIL + TRACE
   +-- timeout/cancel ------> FAIL + TRACE
   |
   v
SUCCESS RESULT + TRACE
```

## 10. Configuration

Each lane still has only provider and model:

```text
--simple-provider
--simple-model
--standard-provider
--standard-model
--complex-provider
--complex-model
```

Provider-specific non-secret settings:

```text
--ollama-url
--azure-endpoint
--bedrock-region
--bedrock-url
--vertex-project
--vertex-location
--vertex-url
--timeout
--sensitive
```

Managed-provider credentials:

```text
AZURE_OPENAI_API_KEY
AZURE_OPENAI_BEARER_TOKEN
AWS_BEARER_TOKEN_BEDROCK
GOOGLE_OAUTH_ACCESS_TOKEN
```

Defaults still point all three lanes to `mock`, so checkout/test use does not require credentials.

## 11. Logging and trace

Every attempted request returns a `Trace` containing:

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

The trace remains part of the returned result. Prompts and outputs are not independently persisted by the harness.

## 12. Failure handling

| Failure | Behaviour |
|---|---|
| empty task | reject before provider selection; failed trace |
| invalid complexity | reject; failed trace |
| incomplete lane | reject; failed trace |
| provider not registered | reject; failed trace |
| missing provider configuration/credential | provider returns explicit configuration error |
| provider HTTP non-2xx | provider returns status and bounded response body |
| malformed response | provider returns decode/no-output error |
| timeout/cancellation | context is propagated to provider HTTP request |

V0.2 still does not retry or fail over. Failure remains visible rather than hidden.

## 13. Test strategy

The original router and Ollama tests remain.

V0.2 adds provider contract tests using local HTTP test servers. They verify:

- endpoint/path construction;
- authentication headers;
- request payload structure;
- model propagation;
- text output parsing;
- token usage parsing;
- missing configuration/credential errors;
- HTTP failure behaviour.

These are deterministic contract tests. They do **not** claim that the repository has been run against live customer cloud credentials.

GitHub Actions runs `go test ./...` on pushes and pull requests.

## 14. Security boundary

This remains a model-routing harness, not a security control.

Important boundaries:

- caller prompts are sent to the configured provider for the selected lane;
- `Sensitive=true` is metadata only and does not block managed cloud providers;
- credentials enter provider adapters from environment variables;
- credential acquisition/refresh and secret storage remain external responsibilities;
- no prompt filtering, injection defence, policy enforcement or DLP is present;
- no provider is silently selected or failed over.

## 15. Known limitations

- classification quality is wholly the caller's responsibility;
- lane configuration is process-startup configuration;
- no retries, failover or circuit breaker;
- no pricing/cost calculation;
- no persistent trace store;
- no streaming provider contract;
- no managed credential refresh;
- enterprise adapters currently target one stable text-generation API shape per platform rather than every model-specific capability;
- provider contract tests use local HTTP servers rather than live enterprise accounts;
- request IDs are process-local timestamp/counter identifiers.

## 16. V0.2 design conclusion

The key result is that the routing layer did not change when three managed enterprise platforms were added.

The extension occurred entirely behind the provider boundary plus CLI construction/configuration. That is the intended architecture: provider growth should widen the edge of the system, not increase the complexity of the routing core.
