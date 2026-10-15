# Three-Lane Model Harness

[![test](https://github.com/jonathanselby/three-lane-model-harness/actions/workflows/test.yml/badge.svg)](https://github.com/jonathanselby/three-lane-model-harness/actions/workflows/test.yml)

**Current version: 0.2.0**

A deliberately small Go harness for routing explicitly classified AI work across simple, standard and complex model lanes.

The core rule is:

> **Do not start by choosing a model. Start by classifying the work.**

And the implementation boundary is equally important:

> **The caller classifies. The harness routes.**

There is no hidden complexity classifier in this repository. A caller submits a task with `simple`, `standard` or `complex`; the harness routes it to the provider/model configured for that lane and returns the output with a minimal execution trace.

## Why V0.2

V0.1 proved the pattern with a deterministic mock provider and Ollama. That was useful, but it proved local portability rather than an enterprise production model.

Ollama remains useful for development and self-hosted experiments. V0.2 adds the managed AI platforms of the three largest global cloud providers:

| Cloud | V0.2 provider | Role |
|---|---|---|
| Microsoft Azure | `azure` | Microsoft Foundry / Azure OpenAI |
| Amazon Web Services | `bedrock` | Amazon Bedrock Runtime |
| Google Cloud | `vertex` | Vertex AI / Gemini |

The provider choice follows the current global cloud market rather than personal preference: AWS, Microsoft and Google are the clear top three cloud infrastructure providers. The important V0.2 test is therefore:

> Can one unchanged caller/router contract sit above local models and the three dominant enterprise cloud AI platforms?

The answer in V0.2 is yes. Provider growth widened the edge of the application; it did not change the routing core.

See [`CHANGELOG.md`](CHANGELOG.md) and [`PRD.md`](PRD.md) for the version rationale.

## Architecture

```text
                         CALLER
                           |
                  explicit complexity
                           |
                           v
                 +--------------------+
                 |       ROUTER       |
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
      +---------+-----------+-----------+---------+
      |         |           |           |         |
      v         v           v           v         v
    MOCK      OLLAMA      AZURE       BEDROCK   VERTEX
```

See [`architecture.txt`](architecture.txt) and [`EDD.md`](EDD.md).

## Provider support

| Provider | Intended use | Authentication in V0.2 | API surface |
|---|---|---|---|
| `mock` | deterministic tests/examples | none | in-process |
| `ollama` | local/self-hosted | none in harness | `/api/generate` |
| `azure` | managed Microsoft enterprise AI | API key or Entra bearer token | Foundry/Azure OpenAI v1 Chat Completions |
| `bedrock` | managed AWS enterprise AI | Bedrock bearer token | Bedrock Runtime Converse |
| `vertex` | managed Google Cloud enterprise AI | Google OAuth access token | Vertex AI `generateContent` |

The managed-cloud adapters use standard-library HTTP rather than pulling cloud SDKs into the routing core. Credential acquisition and refresh remain deployment responsibilities.

## Run locally

Requires Go 1.26+.

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

## Ollama

```bash
go run ./cmd/harness \
  -complexity simple \
  -simple-provider ollama \
  -simple-model qwen3:8b \
  -prompt "Extract the invoice number: INV-20491"
```

Default endpoint: `http://localhost:11434`. Override with `-ollama-url`.

## Microsoft Foundry / Azure OpenAI

Credentials are environment variables, not command-line flags.

```bash
export AZURE_OPENAI_ENDPOINT="https://YOUR-RESOURCE.openai.azure.com"
export AZURE_OPENAI_API_KEY="..."

go run ./cmd/harness \
  -complexity standard \
  -standard-provider azure \
  -standard-model YOUR-DEPLOYMENT \
  -prompt "Summarise these paragraphs and identify the main risks."
```

For keyless production identity, supply a short-lived Entra token in `AZURE_OPENAI_BEARER_TOKEN`; it is preferred when both Azure credentials are present.

## Amazon Bedrock

```bash
export AWS_REGION="ap-southeast-2"
export AWS_BEARER_TOKEN_BEDROCK="..."

go run ./cmd/harness \
  -complexity complex \
  -complex-provider bedrock \
  -complex-model YOUR-BEDROCK-MODEL-ID \
  -prompt "Compare these two architecture proposals and explain the trade-offs."
```

V0.2 uses the Bedrock Converse API. Short-lived Bedrock API keys are the intended bearer credential for this lightweight HTTP adapter.

## Google Vertex AI

```bash
export GOOGLE_CLOUD_PROJECT="your-project"
export VERTEX_LOCATION="australia-southeast1"
export GOOGLE_OAUTH_ACCESS_TOKEN="..."

go run ./cmd/harness \
  -complexity simple \
  -simple-provider vertex \
  -simple-model YOUR-GEMINI-MODEL \
  -prompt "Extract the invoice number from this text."
```

Token acquisition through ADC, workload identity or `gcloud` stays outside the harness.

More examples: [`examples/enterprise-providers.md`](examples/enterprise-providers.md).

## Mixed-cloud lanes

The three lanes can point at different clouds. The router does not care.

```text
SIMPLE   -> Azure deployment
STANDARD -> Vertex model
COMPLEX  -> Bedrock model
```

Only the lane selected by the caller's explicit complexity is invoked.

## Credential boundary

V0.2 deliberately separates non-secret configuration from credentials.

Non-secret settings can be flags:

```text
-azure-endpoint
-bedrock-region
-bedrock-url
-vertex-project
-vertex-location
-vertex-url
```

Credentials are environment-only:

```text
AZURE_OPENAI_API_KEY
AZURE_OPENAI_BEARER_TOKEN
AWS_BEARER_TOKEN_BEDROCK
GOOGLE_OAUTH_ACCESS_TOKEN
```

This is not a secret-management system. Production deployments should source those values from their normal workload identity or secret-management mechanism.

## Test it

```bash
go test ./...
```

Tests cover:

- all three routing lanes;
- invalid classification and empty work;
- missing and failing providers;
- Ollama HTTP behaviour;
- Azure endpoint, auth, payload, response and errors;
- Bedrock endpoint, bearer auth, Converse payload and usage parsing;
- Vertex endpoint, bearer auth, `generateContent` payload and usage parsing.

Provider tests use local HTTP servers, so they verify the wire contract without requiring real cloud credentials. They do not claim live account integration testing.

GitHub Actions runs the full Go test suite on pushes and pull requests.

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

This is deliberately enough to inspect routing without building an observability platform.

## What this does not do

This is **not** an automatic model selector, security gateway or agent framework.

It does not contain:

- automatic classification;
- risk scoring;
- policy enforcement;
- prompt-injection controls;
- autonomous provider choice;
- retries/failover;
- credential lifecycle management;
- proprietary routing logic.

`Sensitive` is caller-supplied metadata. It does **not** change routing or enforce a policy.

## Repository map

```text
three-lane-model-harness/
├── README.md
├── VERSION
├── CHANGELOG.md
├── PRD.md
├── EDD.md
├── architecture.txt
├── LICENSE
├── go.mod
├── cmd/harness/main.go
├── internal/
│   ├── harness/
│   └── provider/
│       ├── mock/
│       ├── ollama/
│       ├── azure/
│       ├── bedrock/
│       └── vertex/
├── examples/
│   ├── mock-runs.md
│   └── enterprise-providers.md
└── generated/
    ├── EDD.md
    └── COMPARISON.md
```

## Design cross-check

This repository is also a reference test for [`codebase-to-edd`](https://github.com/jonathanselby/codebase-to-edd).

`EDD.md` records intended design. `generated/EDD.md` records what can be recovered from the implementation without treating the hand-written EDD as evidence. `generated/COMPARISON.md` compares the two.

## References for provider interfaces

- Microsoft Foundry/Azure OpenAI v1 API: https://learn.microsoft.com/en-us/azure/foundry/openai/latest
- Amazon Bedrock API keys and Converse: https://docs.aws.amazon.com/bedrock/latest/userguide/api-keys.html
- Vertex AI Gemini REST quickstart: https://cloud.google.com/vertex-ai/generative-ai/docs/start/quickstart

## Licence

MIT. See [`LICENSE`](LICENSE).
