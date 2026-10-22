---
id: 000538
status: codecomplete
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-22
estimate_hours: 1.43
collaboration-mode: collaboration
card_mirror: 'b52ee9fc894e720ca62bc6b882f42c7d39119896' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T02:51:52-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 1.36
---

# stage --clear recognizes its landed holder after branch deletion

## Problem

## Spec

After pair#752 landed, landing deleted its issue branch and returned pair:9 to its resting branch. Its subsequent plain `sdlc stage --clear` refused the stage lease because the holder identity is currently derived from the caller branch; the holder had to use `--clear --abandoned 752` even though landing is the normal completion, not abandonment. Reported by ops:0 while dispatching ariadne#537.

Allow the original holder slot to clear its landed stage after normal branch deletion using confirmed durable landing evidence. Preserve refusal for other slots/issues, unlanded or unknown outcomes, replaced leases and explicit abandoned-holder acknowledgement. Do not treat a missing branch alone as landing proof.

## Done when

- A regression stages from a worker issue branch, completes a durable landing and branch cleanup, then clears from that original slot on its rest without --abandoned.
- Another slot, unfinished/unknown holder and replaced lease still refuse without clearing or rebuilding.

## Plan

- [x] Add a proof-backed original-slot/resting-caller route to clear a landed issue's retained stage, sharing the existing main pin and archive authority.
- [x] Replay the branch-deleted landing regression plus independent identity/generation/lease/cancellation refusals; preserve ordinary and explicit abandoned clears.
- [x] Document, verify, close and land. Detailed design: [implementation plan](../plans/000538-stage-clear-recognizes-its-landed-holder-after-branch-deletion-plan.md).

## Log

2026-10-21: Split from #537 per hub permission if not small. Fresh-main restoration is independent of the holder-authentication policy. Existing checkClearLease only sees branch-derived holder strings; safe normal-completion recovery needs durable landed evidence and original-slot binding, rather than a branch-absence or slot-only bypass.


### 2026-10-22
- 2026-10-22: closed — Recorded suite PASS on pushed f04efa370aa0c308faddb46ad594e5b41624464c, attempt 2378cbb55929bb96ab54f3583083144b: 2154 cmd/sdlc tests plus 76 other packages, zero failures, 986.1s execution; log SHA256 50c13d0a1cc8c98c4ef2525a0dd4dd39cb8fba8f1ac70c9b83e8e80891100d99. Native production-path stage then real completion/archive/branch deletion then plain clear verifies exact fresh-main rebuild; original holder refusal, replaced-token acceptance and target movement each observed RED before fixes. Adversarial generation/identity/lease/cancellation/order controls, merge strategies/train-car/generic archive, ordinary/abandoned/freshness/move controls, recovery catalog, lint and compiled help pass. Hub approved isolated proof; no live stages touched. README/help/atlas/recovery updated.; review verdict: SHIP

- Claimed and started from fresh main after #536 landed. Hub highlighted four upcoming operator stages and requested claim/plan, not live :0 mutation.
- Exploration confirms runUnstage currently derives holder only from the caller branch. The target already has a lifetime stage lock and records canonical original Slot, Branch and staged Commit, so no new lease format is needed. Reuse #537's main pin and #512's durable archive proof (ARCH-DRY).
- Proof trap found before code: landingArchiveComplete treats an empty issue selection as success. Recovery must explicitly select the held issue's current done completion/generation before using archive confirmation. Missing/reopened/replaced bindings cannot become issue-free success. Durable plan prepared for hub review; no implementation or tests changed.

- 2026-10-22 03:00 PDT — Hub approved the committed full-flow plan and isolated native production-path proof (receipt 6ad82555-6756-417d-9954-ffba37f12d76). Confirmed original canonical resting slot, current done binding/merged PR/staged commit and explicit nonempty archive selection under existing stage lock/main pin, with final caller/lease/generation re-observation. Do not touch the live pair:0 stages. The existing explicit abandoned acknowledgment remains the interim operator path.

- 2026-10-22 03:08 PDT — change-code admitted full flow after clean plan-quality round 2 and INFO estimate-quality. Estimate allowances separate production wiring/tests from adapting the existing durable-slot landing fixture and cross-boundary ordering controls; no new API discovery. Compiled help, native proof, recorded suite execution and landing observation are included in active verification; idle queue time is excluded. Fixture estimate remains optimistic and provisional, as the reviewer noted.

- 2026-10-22 03:23 PDT — TDD: original durable-landing regression RED on the existing holder refusal, then GREEN with a real make/build stamp at fresh main (8.579s). Broader first matrix found a genuine replaced-token false acceptance through an unmirrored legacy archive; added stage-only requirement that the archive's done-card mirror pins the full current completion, leaving shared legacy archive behavior unchanged. Corrected two adversarial fixture setters to corrupt raw input rather than fail during setup. Read-only sanity review also found target movement during proof: native target-hook regression RED (16.042s), then added final target fact re-observation under the existing lock. The focused target/post-archive replacement and replaced-token/wrong-repository/unreachable/unmirrored controls now PASS (50.467s). Original slot, caller, lease, target and fresh generation are all checked before effects; unknown raw PR merge objects use existing head reachability fallback. No live :0 stage touched.

- 2026-10-22 — Targeted regressions PASS (77.474s): original-slot clear with exact-pin rebuild, squash/rebase, missing/unrelated/unknown raw merge objects, generic archive, pure candidate table, existing stage leases, freshness and move controls, and recovery proof catalog. Focused refusal/order checks PASS (50.467s). Go/shell lint packages PASS. Newly compiled stage help displays the new route; its non-fleet-binary warning is expected for go run and no tool slot was modified. Added the existing recovery catalog's new authority/proof binding and final target observation uses the existing bounded Git seam via interface-typed readers. Full recorded suite remains pending on pushed branch.

- 2026-10-22 — Required recorded suite PASS on pushed Git head f04efa370aa0c308faddb46ad594e5b41624464c: attempt 2378cbb55929bb96ab54f3583083144b, retained run lease-113685271, 2154 cmd/sdlc tests across eight shards and 76 other packages, zero failures/reruns/waivers. Execution wall 986.1s (16m26s), queue wait 1708s; saved log SHA256 50c13d0a1cc8c98c4ef2525a0dd4dd39cb8fba8f1ac70c9b83e8e80891100d99. CI checks passed. PR #391: https://github.com/xianxu/ariadne/pull/391. Isolated native production-path proof is the approved live gate; no live stage touched. Implementation and verification complete; proceeding to close review and then landing, with no manual main chase.

## Revisions

- 2026-10-22 02:59 PDT — Proposed full-flow design narrows the new recovery route to the original canonical slot on its own rest with current tracked done completion, matching landed PR and confirmed archive. Existing same-holder/legacy and exact --abandoned paths remain; new recovery grants no staging or move authority. Propose isolated native proof to preserve the four live operator stages; awaiting hub review.

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. Calibration source is stale, so values remain provisional.*

Reuse existing completion/archive/GitHub integrations; no novel stack or external library is required. Raw v2 picks: two smaller Go modules (predicate/proof and locked clear integration), each design 0.30 / implementation 0.50; one cross-cutting refactor/native fixture and adversarial proof coverage, design 1.00 / implementation 0.50; issue spec 0.50 / 0.10; docs 0.10 / 0.15; one review boundary 0.10 / 0.50. Thorough approved plan applies design ×0.2 and +15% buffer; familiar implementation ×1.0 after v3.1 ×0.4 scaling. Tests, debugging and active verification are included in implementation, queue-only time excluded. Design subtotal 0.46; implementation subtotal 0.90; total 0.46 ×1.15 +0.90 =1.429, rounded 1.43.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: cross-cutting-refactor design=0.20 impl=0.20
item: issue-spec design=0.10 impl=0.04
item: atlas-docs design=0.02 impl=0.06
item: milestone-review design=0.02 impl=0.20
design-buffer: 0.15
total: 1.429
```
