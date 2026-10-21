---
id: 000524
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours: 2.82
collaboration-mode: collaboration
card_mirror: '20febc2959000ea7c3eeed24e9e4034844429ec9' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T04:42:56-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# Scoped future evidence and exact-package flake reruns

## Problem

Prerequisite for ariadne#463, approved by hub receipt917bf022-032a-4d84-ac6b-39075bbca2b5. Today's attempt reader treats an unsupported record as a global warning; guardSuitePass then refuses unrelated successes and saved-pass loading marks the whole store incomplete. A future ResumeLink cannot retrofit stale binaries. Existing known-flake reruns also use one name-pattern cross-product across selected packages.

## Spec

Reader-first landing only: make valid unsupported/v3 attempt evidence refuse its own proof as “needs a newer sdlc”, without global read warnings or whole-snapshot incompleteness for unrelated records/checks. Scope using a strictly validated identity envelope; malformed or ambiguous identity remains unknown, never silently ignored. Preserve existing failure guards for affected evidence.

Build exact-package known-flake rerun batches, retaining #466's whole-top-level-test rule and siblings but never running a failed name in another package where it did not fail. Reuse one managed weight-1 lease, existing aggregate targeted budget, ownership/cancellation/process cleanup and attempt linkage. The automatic path uses the shared batch.

No ResumeLink writer or resume activation. Atlas must name the old-reader limitation and the mandatory rollout barrier: after this issue lands, the hub upgrades every slot's actual sdlc; this worker verifies paths/revisions before ariadne#463 phase2 may write links. Never hide links from older readers.

## Done when

- A mixed private store with one valid future/v3 envelope and unrelated ordinary attempts permits unrelated proof; selecting the affected proof refuses “needs a newer sdlc”. Malformed/ambiguous identity never becomes passing evidence.
- Shared store consumers preserve scoped uncertainty without discarding supported records or weakening same-head failure guards. The real pre-change reader control documents why fleet rollout is required.
- A two-package same-name-decoy fixture proves no cross-package over-selection; whole-parent subtest behavior stays intact. Exactly one managed lease and aggregate timeout/cancel behavior are verified.
- No production v3/link writer or resume command is enabled. Atlas records separate landings and fleet binary verification before activation.
- Targeted tests, private executable proof, central recorded suite, close review and normal landing pass; hub receives rollout instructions.

## Estimate

Produced via brain/data/life/42shots/velocity/estimate-logic-v3.1.md against baseline-v3.1.md. Method A only; the source is flagged stale, so provisional. Standard encoding/json/jsonstrict, existing attempt/Git/GH/queue/process seams and stateful fakes cover library needs; no novel stack or new API discovery.

Six smaller modules cover envelope decoding, bounded snapshot/index, authority-consumer integration, exact-package planning, structured PASS collection and flake-step integration. Each uses v2 design0.2 × thorough-spec0.2 =0.04 and impl0.5 × v3.1 scaling0.4 =0.20. The managed-run refactor uses base design0.4 ×0.2 =0.08 and impl0.5 ×0.4 =0.20; private executable conformance is one smaller module at0.04/0.20. Docs use base0.1/0.2 →0.02/0.08; issue/spec0.5/0.1 →0.10/0.04. Two milestone reviews plus final close use base0.1/0.4 →0.02/0.16 each. Familiarity1.0 from recent queue/process/evidence work. Design0.54 ×1.15 buffer + impl2.20 =2.821, rounded2.82h.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.10 impl=0.04
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: cross-cutting-refactor design=0.08 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.02 impl=0.16
item: milestone-review design=0.02 impl=0.16
item: milestone-review design=0.02 impl=0.16
design-buffer: 0.15
total: 2.82
```

## Plan

- [ ] M1 — Typed scoped unsupported-evidence reader and all proof-consumer projections, old-reader controls, atlas rollout boundary.
- [ ] M2 — Exact-package managed rerun batch, automatic integration, process/adversarial tests and final delivery preparation.

## Log

### 2026-10-21 — Approved reader-first prerequisite

Claimed at filing and started independently from fresh main. The approved design is recorded in [the durable plan](../plans/000524-scoped-evidence-reruns-plan.md): one typed scoped read view plus exact-package managed batches. Hub receipt917bf022-032a-4d84-ac6b-39075bbca2b5 owns the split and fleet rollout barrier. Phase1 does not create links or enable resume; #463 and #520 branches are preserved. Package-bound JSON PASS events prevent a same-name or quoted output from satisfying another package's test. Unscopable malformed data remains unknown; valid future identities are localized so unrelated records/checks stay usable.

### 2026-10-21 — Plan-quality round1

PQ-1 (function-test-strategy-contract) asks for named risky functions and adversarial strategies rather than case inventories. The plan now names decoder, authority relation, bounded snapshot, saved/post consumers, batch builder, structured PASS collector, managed batch and production flake entrypoint with one strategy/mechanical guard per function. This is test-contract refinement only; no implementation, mechanism delta or scope change.

Filed as the separately landable phase1 prerequisite from the approved ariadne#463 plan. The original #463 branch holds only committed design, with no implementation. #520 remains held for operator approval on its own branch.

### 2026-10-21 — M1 RED and implementation admission

change-code passed after plan-quality disposed PQ-1 and estimate-quality returned INFO. The derived2.82h is provisional; each component includes its adversarial tests, and review/private-conformance allocations include verification execution and gate overhead, with integration/remediation a stated uncertainty rather than excluded unattended work. M1 TestScopedSavedProofConsumers failed as intended through real saved-proof/post guards: a single valid future record made unrelated repo/head/check proofs incomplete and refused status, while related records lacked a scoped upgrade diagnostic. The existing supported-attempt decoder still rejects opaque v3 data. All fixture state is private; no production evidence was written.
