# Three-Lane Model Harness — PRD

## Problem

Teams often choose a model first and define workload fit later. That couples application logic to provider/model choices and makes cost, latency and capability decisions harder to change.

## Product objective

Provide a deliberately small executable harness that routes **explicitly classified** AI work to one of three configured model lanes: simple, standard and complex.

The caller decides the complexity class. The harness routes. It does not hide classification logic inside the runtime.

## Users

Engineers and architects experimenting with model-provider abstraction, workload routing and simple model operating patterns.

## In scope

- explicit `simple` / `standard` / `complex` classification;
- configured provider and model per lane;
- structured request and result;
- command-line execution;
- provider abstraction;
- deterministic mock provider;
- Ollama-compatible provider;
- minimal execution trace;
- timeout propagation;
- tests for all three lanes and material failure paths.

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
- cost optimisation engine.

## Acceptance criteria

- caller can submit a request with an explicit complexity class;
- harness sends it to the configured provider/model for that lane;
- response identifies provider and model;
- execution emits a trace with request ID, timing and success/failure;
- provider implementation can be swapped without changing caller behaviour;
- mock mode runs without credentials or external services;
- Ollama mode uses the same provider contract;
- tests demonstrate routing for all three lanes;
- tests demonstrate invalid classification, missing provider and provider failure behaviour;
- `Sensitive` is recorded but does not alter routing or imply security enforcement.
