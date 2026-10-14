# Changelog

## 0.2.0

V0.2 moves the harness from a local proof-of-pattern to a production-cloud portability test.

V0.1 used a deterministic mock provider and Ollama. That proved the router/provider abstraction but did not demonstrate that the same caller contract could sit above enterprise managed AI platforms.

V0.2 adds:

- Microsoft Foundry / Azure OpenAI via the v1 Chat Completions API;
- Amazon Bedrock via the Converse API;
- Google Vertex AI via the `generateContent` API;
- environment-based credentials for managed cloud providers;
- contract tests for endpoint construction, authentication, payloads, response parsing, token usage and HTTP failures;
- updated PRD, EDD, architecture, examples and generated design cross-check.

The routing contract remains unchanged:

> The caller classifies. The harness routes.

No automatic classifier, security gateway, provider failover or autonomous model selection was added.

## 0.1.0

Initial working edition:

- explicit simple / standard / complex lanes;
- mock provider;
- Ollama provider;
- CLI;
- structured trace;
- unit tests;
- PRD and hand-written EDD;
- `codebase-to-edd` generated EDD and design comparison.
