# Three-Lane Model Harness — PRD

**Version:** 0.2

## Why V0.2 exists

V0.1 proved the routing pattern with a deterministic mock provider and Ollama. That was enough to demonstrate that the caller could classify work and the harness could route it without embedding provider choice in application logic.

It was not enough to demonstrate a production enterprise operating model.

Ollama is valuable for local development and self-hosted experiments, but enterprise AI estates commonly run through managed cloud platforms with cloud identity, managed endpoints, audit/governance controls and production service boundaries.

V0.2 therefore adds three production-oriented provider adapters:

- Microsoft Foundry / Azure OpenAI;
- Amazon Bedrock;
- Google Vertex AI.

The product test for V0.2 is stronger than "can we call three more APIs?" The important question is:

> Can the same three-lane caller and router contract sit above materially different enterprise AI platforms without changing routing logic?

## Problem

Teams often choose a model first and define workload fit later. That couples application logic to provider/model choices and makes cost, latency and capability decisions harder to change.

Enterprise teams have an additional problem: the same application may need to operate across local development, self-hosted models and one or more managed cloud AI platforms without spreading platform-specific code through the application.

## Product objective

Provide a deliberately small executable harness that routes **explicitly classified** AI work to one of three configured model lanes: simple, standard and complex.

The caller decides the complexity class. The harness routes. Provider-specific transport and authentication remain behind a stable provider contract.

## Users

Engineers and architects experimenting with model-provider abstraction, workload routing and simple production operating patterns across local and managed enterprise AI platforms.

## In scope

- explicit `simple` / `standard` / `complex` classification;
- configured provider and model per lane;
- structured request and result;
- command-line execution;
- provider abstraction;
- deterministic mock provider;
- Ollama-compatible provider for local/self-hosted use;
- Microsoft Foundry / Azure OpenAI provider;
- Amazon Bedrock provider;
- Google Vertex AI provider;
- environment-based credential injection;
- minimal execution trace;
- timeout propagation;
- tests for all three lanes, provider wire contracts and material failure paths.

## Out of scope

- automatic workload classification;
- business-policy enforcement;
- security controls;
- proprietary routing logic;
- autonomous model selection;
- agent orchestration;
- provider failover;
- retry policy;
- persistent telemetry platform;
- cost optimisation engine;
- cloud credential acquisition or refresh;
- workload identity implementation;
- production secret storage;
- streaming provider contract.

Cloud identity is deliberately a deployment responsibility. V0.2 accepts short-lived credentials supplied by the environment rather than building credential lifecycle management into the router.

## Acceptance criteria

- caller can submit a request with an explicit complexity class;
- harness sends it to the configured provider/model for that lane;
- response identifies provider and model;
- execution emits a trace with request ID, timing and success/failure;
- provider implementation can be swapped without changing caller behaviour;
- mock mode runs without credentials or external services;
- Ollama mode uses the same provider contract;
- Azure, Bedrock and Vertex use the same provider contract and router;
- enterprise provider credentials are not accepted as command-line flags;
- provider contract tests verify URL construction, authentication headers, request payload, response parsing, token usage and HTTP failure behaviour;
- tests demonstrate routing for all three lanes;
- tests demonstrate invalid classification, missing provider and provider failure behaviour;
- `Sensitive` is recorded but does not alter routing or imply security enforcement.

## Definition of done for V0.2

V0.2 is complete when the enterprise providers compile and pass contract tests, documentation explains their production boundary, the generated EDD is refreshed from the implementation, and the hand-written/generated design comparison shows that provider growth did not change the caller or routing contract.
