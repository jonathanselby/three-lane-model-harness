# Hand-Written vs Generated EDD — V0.2

## Purpose

This comparison tests whether the V0.2 implementation still communicates the architecture intended by the hand-written `EDD.md` after adding Microsoft Foundry / Azure OpenAI, Amazon Bedrock and Google Vertex AI.

The generated EDD was produced from implementation/tests/runtime material without using the hand-written `EDD.md` or `PRD.md` as architecture evidence.

## Result

**Pass.**

The material runtime architecture is recoverable from code and remains aligned with the intended design.

Most importantly, V0.2 did not require changes to the routing core. The enterprise platforms were added behind the existing provider contract.

## Comparison

| Design area | Hand-written EDD | Generated EDD | Assessment |
|---|---|---|---|
| caller owns classification | explicit design rule | recovered directly from `Request.Complexity` and lane lookup | MATCH |
| three fixed lanes | explicit | recovered from `Config` and `laneFor` | MATCH |
| provider abstraction | explicit | recovered from `Provider` interface and registry | MATCH |
| provider-neutral router | explicit | no provider-specific branching found in `Router.Run` | MATCH |
| mock provider | intentional local/test implementation | recovered | MATCH |
| Ollama provider | local/self-hosted adapter | recovered | MATCH |
| Azure provider | managed Microsoft adapter | recovered from HTTP path/auth/payload | MATCH |
| Bedrock provider | managed AWS adapter | recovered from Converse path/auth/payload | MATCH |
| Vertex provider | managed Google adapter | recovered from generateContent path/auth/payload | MATCH |
| environment credential boundary | explicit V0.2 design | recovered from CLI/provider construction | MATCH |
| credentials not CLI flags | explicit | recovered from flag definitions and `os.Getenv` | MATCH |
| timeout propagation | explicit | recovered from CLI context and provider HTTP requests | MATCH |
| failure is explicit | explicit | recovered across router/provider error paths | MATCH |
| retries/failover absent | deliberate non-goal | no implementation found | MATCH |
| `Sensitive` is metadata only | explicit | no branch uses it for routing/enforcement | MATCH |
| persistent telemetry absent | explicit | no store/sink implementation found | MATCH |

## What code communicates well

V0.2 makes several important design properties observable without reading the design document:

1. **Cloud growth does not alter routing.** `Router.Run` still knows only provider names/models and the common interface.
2. **Platform-specific complexity stays at the edge.** URLs, auth headers and JSON shapes are confined to provider packages.
3. **The caller remains authoritative for workload class.** No provider or router attempts to infer complexity.
4. **Credentials are deployment inputs.** Managed provider adapters receive resolved credentials; the core does not acquire or refresh them.
5. **Failure is visible.** Missing credentials, non-2xx responses and malformed output are returned rather than silently retried or failed over.

## What implementation cannot explain by itself

The generated EDD correctly marks several things as unknown/inferred rather than inventing rationale:

- why there are exactly three lanes;
- why the V0.2 production platforms are Azure, Bedrock and Vertex rather than direct OpenAI/Anthropic APIs;
- why standard-library HTTP was preferred to cloud SDKs;
- whether mixed-cloud lane configuration is expected in production or merely possible;
- future policy for retries, streaming, failover and credential providers.

Those belong in the PRD/hand-written EDD because they are product and architecture intent rather than executable behaviour.

## V0.2 architectural finding

The strongest result from this cross-check is that V0.2 expanded **horizontally at the provider edge** rather than **vertically through the core**.

```text
V0.1
CALLER -> ROUTER -> Provider -> mock / ollama

V0.2
CALLER -> ROUTER -> Provider -> mock / ollama / azure / bedrock / vertex
```

The left side of the architecture did not change.

That is the behaviour the repository was intended to demonstrate.
