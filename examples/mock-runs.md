# Verified mock-provider runs

These commands were executed against V0.1 after `go test ./...` passed.

The mock provider is deliberately deterministic. Timestamps and request IDs vary by execution.

## Simple

```bash
go run ./cmd/harness \
  -complexity simple \
  -prompt "Extract the invoice number: INV-20491"
```

```json
{
  "output": "mock[mock-small]: Extract the invoice number: INV-20491",
  "trace": {
    "complexity": "simple",
    "sensitive": false,
    "provider": "mock",
    "model": "mock-small",
    "duration_ms": 0,
    "success": true
  }
}
```

## Standard

```bash
go run ./cmd/harness \
  -complexity standard \
  -prompt "Summarise these paragraphs and identify the main risks."
```

```json
{
  "output": "mock[mock-standard]: Summarise these paragraphs and identify the main risks.",
  "trace": {
    "complexity": "standard",
    "sensitive": false,
    "provider": "mock",
    "model": "mock-standard",
    "duration_ms": 0,
    "success": true
  }
}
```

## Complex

```bash
go run ./cmd/harness \
  -complexity complex \
  -prompt "Compare these two architecture proposals and explain the trade-offs."
```

```json
{
  "output": "mock[mock-strong]: Compare these two architecture proposals and explain the trade-offs.",
  "trace": {
    "complexity": "complex",
    "sensitive": false,
    "provider": "mock",
    "model": "mock-strong",
    "duration_ms": 0,
    "success": true
  }
}
```

## What this proves

The examples prove the mechanical contract only:

- the caller supplies the complexity class;
- each class maps to the configured lane;
- provider/model identity is returned in the trace;
- all three lanes run through the same caller contract.

They do not prove that a particular model belongs in a particular lane. That is deliberately outside the harness.
