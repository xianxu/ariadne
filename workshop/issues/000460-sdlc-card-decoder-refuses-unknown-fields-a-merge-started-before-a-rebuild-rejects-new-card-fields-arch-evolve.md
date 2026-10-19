---
id: 000460
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours: 2.84
collaboration-mode: collaboration
card_mirror: '9dbc03ba20ef82acef0fd88a707bbb1f4c59b3b1' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T00:51:23-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# sdlc card decoder refuses unknown fields: a merge started before a rebuild rejects new card fields (ARCH-EVOLVE)

## Problem

ariadne#394 added `flake_tests` to the tracker card. sdlc's card decoder refuses unknown fields, so any sdlc process started before ariadne:0 was rebuilt rejects every card that carries the new field. On 2026-10-19 pair:3's long-running `sdlc merge` for pair#687 (started on the old binary) refused the card after the server-side merge and needed manual recovery; other merges in flight at the time are exposed the same way. This is exactly what ARCH-EVOLVE (ariadne#454, landed the same night) forbids: an older reader must tolerate and preserve unknown fields.

## Spec

Make the tracker-card decoder (and any card rewrite path) tolerate and preserve unknown members, per ARCH-EVOLVE: capture on decode, write back unchanged; known fields, types and required fields stay strict. Audit sdlc's other persisted records that a long-lived process reads (landing events, test-stats records, readiness sidecars from #449) for the same strictness, and fix or list them. A regression: an old-shape decoder reads a card with an extra field, rewrites it, and the field survives.

## Done when

- A card with an unknown field round-trips through every card read-modify-write path; the regression proves it.
- The audit list is in the issue Log, each item fixed or filed.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. The source is marked stale against the current calibration ledger, so these values are provisional. The nested YAML-preservation helper is the one new Go concern; the card and attempt decoders extend existing modules. The plan resolves the design choices, so the 15% design buffer applies. Implementation hours use v3.1's 40% scale against the v2 primitive ranges.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.5 impl=0.08
item: greenfield-go-module design=0.55 impl=0.32
item: smaller-go-module design=0.25 impl=0.2
item: smaller-go-module design=0.2 impl=0.2
item: milestone-review design=0.1 impl=0.2
design-buffer: 0.15
total: 2.84
```

## Plan

- [x] Make old-shape card parsing accept unknown top-level and nested members while retaining strict known fields, required fields, versions, and duplicate-key checks.
- [x] Preserve unknown card members through scalar, claimant, and tracker-envelope setters, then prove a newer card survives a full old-reader rewrite chain.
- [x] Audit landing events, test-stats attempts, and #449 readiness sidecars; fix compatible readers here or file a scoped follow-up for any unsafe or unlanded format.
- [ ] Run targeted tests, `make -k`, the recorded suite on the pushed head, and the close review; land on SHIP.

## Log

### 2026-10-19

- Planning audit: `ParseCard` rejects an unrecognized top-level key. `parseClaimant` also rejects an unrecognized nested key; `updateEnvelope` decodes and re-marshals known nested records, dropping unknown members below `tracker` even though its root `Extra` map exists. The scalar setters and mirror operate on source spans, so they can preserve unknown top-level bytes once parsing admits them.
- Landing events are decoded with ordinary `json.Unmarshal` and retention rewrites their original JSON lines, preserving unknown members. Test-stats v2 currently rejects unknown top-level keys and uses `DisallowUnknownFields`; it is a separate read-only attempt ledger. #449 readiness sidecars are still on ariadne:2's unlanded branch: its reader uses ordinary `json.Unmarshal`, and the owning car writes an entirely new vouch rather than modifying a prior one. The audit and tests will be finalized against its landed version.
- Hub approved the plan and directed #449's sidecar fix into its owning unlanded branch. Sent ariadne:2 the unknown-field preservation and strict-validation contract; Couch receipt `ca2dc1a7-b35a-48ba-a77c-e84f02be3afa` is submitted. #460 will audit the #449 result, not duplicate its code.
- Plan-quality round 1 raised PQ-1 (nested known tracker validation did not yet reach `ParseCard`) and PQ-2 (no named adversarial strategy for the pure merge helper or JSON canonical keys), plus a minor scope-exclusions suggestion. The durable plan now names the boundary call, helper/property and alias checks at every object depth.
- Plan-quality round 2 accepted PQ-1 and the scope exclusions. PQ-2 stayed open because the task prose duplicated the new adversarial strategy table; compressed the task steps to refer to that table.
- Plan-quality round 3 disposed PQ-2 as addressed; no blocking findings remain. Estimate derived only after the plan cleared, as required by the gate.
- Implementation audit: card scalar/title changes retain source spans; claimant replacement and all tracker-envelope setters merge unknown YAML descendants at retained mappings; clearing a claimant deliberately removes that record. A Git-backed CAS regression checks that a setter retains a newer card field. `ParseCard` now validates nested known tracker records before returning. Tests cover extra top-level, claimant, tracker, completion, and release-by members; missing or malformed known fields still refuse.
- Persisted-record audit: landing events already decode additively and the sweep copies original JSON lines; an unknown-member retention regression passes. Test-stats v2 attempt records were strict, so the read-only decoder now accepts additive members at root and in unit/known-flake objects while rejecting duplicate keys at every depth, known-key aliases, malformed known values, and unsupported versions. #449 readiness sidecars are owned by ariadne:2 and remain unlanded; its commit `5918c6fdcc5888303161edcc0f9d0bdc8fe59c36` adds strict-known/tolerant-unknown decoding and preserves unknown vouch members on rewrite, with `TestDecodeReadinessStrictKnownTolerantUnknown`, `TestVouchRewriteKeepsUnknownMembers`, and `TestTrainEventFieldsSurviveTheSweep`. Inspected its code and tests read-only; no competing edit is needed here.
- Pre-suite review found that tracker accessors call `cardEnvelope` directly. A red regression proved they admitted an unsupported version; `cardEnvelope` now checks the version itself so every accessor and envelope setter refuses it. Cancelled the first suite on the old pushed head through `sdlc local-check cancel --name suite`; its worktree was discarded without a recorded result, and the amended head will receive the suite.

## Revisions

- 2026-10-19: #449's sidecar preservation belongs on its still-unlanded branch, per hub approval; coordinate directly and verify it rather than defer until after landing.
- 2026-10-19: PQ-1/PQ-2: enforce nested known validation at `ParseCard` and require direct generated-tree and canonical-JSON checks in the plan before coding.
- 2026-10-19: PQ-2: keep one named adversarial test strategy table and make task steps cite it instead of repeating fixture inventories.
