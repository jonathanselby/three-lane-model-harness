# Hand-written vs generated EDD

This comparison checks whether the intended design in [`../EDD.md`](../EDD.md) can be recovered from the implementation alone.

The generated EDD analysed commit `4fd03c3697703b7b0dff94d6f1c1cd99be5f186d` and deliberately excluded the hand-written EDD and PRD from its evidence basis.

## Summary

The generated EDD recovered the material runtime architecture correctly:

- caller-supplied complexity;
- direct three-lane routing;
- provider registry and interface;
- mock and Ollama providers;
- CLI configuration;
- request/trace contracts;
- timeout propagation;
- explicit failure paths;
- no persistence;
- no retry/failover;
- `Sensitive` as metadata rather than enforcement.

The cross-check therefore supports the main claim of the repository: the implementation reflects the written design closely enough that an unfamiliar-code analysis can reconstruct the important engineering model.

## Comparison

| Design area | Hand-written EDD | Generated EDD | Assessment |
|---|---|---|---|
| core boundary | caller classifies; harness routes | recovered directly from request type and `laneFor` | MATCH |
| lane model | simple / standard / complex | recovered exactly | MATCH |
| provider abstraction | common `Provider` interface | recovered exactly | MATCH |
| providers | mock + Ollama | recovered exactly | MATCH |
| configuration | CLI flags per lane | recovered exactly | MATCH |
| trace model | request/provider/model/timing/success/tokens | recovered exactly | MATCH |
| sensitivity | metadata only; no enforcement | recovered from types, router and tests | MATCH |
| persistence | none | no durable store found | MATCH |
| retries/failover | deliberately absent | none found | MATCH |
| security boundary | routing harness, not security control | absence of filtering/policy and configurable outbound endpoint recovered | MATCH |
| deployment intent | small executable harness | inferred from CLI lifecycle | PARTIAL — intent inferred, not encoded as a formal constraint |
| business classification rules | out of scope | cannot be recovered because they are absent | MATCH / EXPECTED UNKNOWN |
| future provider/auth strategy | not defined | open question | EXPECTED UNKNOWN |

## What the generated EDD cannot recover reliably

Some design intent exists only in product/design documentation and should remain there:

1. **Why three lanes were chosen.** Code proves three lanes exist; it cannot establish the organisational reasoning for choosing exactly three.
2. **Why automatic classification is excluded.** Its absence is observable, but the implementation alone cannot prove whether that is a deliberate product boundary or unfinished work.
3. **Why retries and failover are excluded.** Again, absence is factual; intent requires the hand-written design.
4. **Portfolio/product purpose.** Code can show a small CLI harness but cannot establish why it exists as an engineering demonstration.

That is a useful result rather than a defect. The implementation should reveal runtime truth; the PRD/EDD should preserve intent and boundaries that code alone cannot prove.

## Cross-test conclusion

**PASS.**

The generated EDD recovered the material implementation architecture without access to the intended design. The main differences are correctly limited to design rationale and future intent — information that should not be invented from code.

This is the desired relationship:

```text
PRD
 |
 v
HAND-WRITTEN EDD ---- intent / boundaries
 |
 v
IMPLEMENTATION ------ runtime truth
 |
 v
CODEBASE-TO-EDD ----- recovered runtime model
 |
 v
COMPARE ------------- drift / missing intent / hidden behaviour
```
