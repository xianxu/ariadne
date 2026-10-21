---
id: 000524
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '7c8c7934b4da94209e6653f06adea3e246c74873' # card fields mirrored from issue-cards; edit via sdlc
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

## Plan

- [ ] M1 — Typed scoped unsupported-evidence reader and all proof-consumer projections, old-reader controls, atlas rollout boundary.
- [ ] M2 — Exact-package managed rerun batch, automatic integration, process/adversarial tests and final delivery preparation.

## Log

Filed as the separately landable phase1 prerequisite from the approved ariadne#463 plan. The original #463 branch holds only committed design, with no implementation. #520 remains held for operator approval on its own branch.
