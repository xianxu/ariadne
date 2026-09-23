---
id: 000244
status: working
deps: [ariadne#242, ariadne#243]
github_issue:
created: 2026-09-22
updated: 2026-09-23
estimate_hours: 21.28
started: 2026-09-23T11:35:53-07:00
---

# Slots v2: concurrent issue workflows

## Problem

Two slots running SDLC concurrently must not claim the same issue silently, overwrite newer issue records, or block unrelated work throughout a long review.

## Spec

Project: `pair/workshop/projects/couch-slots-v2.md`. Fresh task derived from the current v2 contract; historical task bodies are not prerequisites or implementation plans.

Audit and fix current SDLC behavior with two worktrees of one repository. Authoritative reservation checks for issue creation and claiming must be fresh at publication; a loser receives an actionable conflict. Publishing an issue-body update must preserve concurrent unrelated records and refuse conflicting edits rather than replacing them from a stale checkout.

Keep repository mutation serialization where needed, while external review waits should not monopolize the shared repository lock. Revalidate the exact reviewed/prepared state after reacquiring authority. No automatic claim transfer is introduced by branching from another workspace. ARCH-SECURE and ARCH-DRY: use existing ownership/publication machinery, with fresh evidence at mutation boundaries. This is a new v2 acceptance contract; audit live code rather than inheriting old issue conclusions.

### Agreed scope — 2026-09-23

This section takes precedence over earlier conflicting layout or policy text.

Apply the publication/concurrency audit to both numbered main-repository worktrees and ordinary dependency clones inside their environments. Independent clones share a remote but not a local Git lock, so allocation, reservation and conflicting issue-body publication must rely on fresh remote evidence. A thread rooted in Pair may create, update, review and publish an Ariadne issue by targeting that environment's Ariadne checkout through existing SDLC commands. Creating/publishing a reservation does not imply later local issue-body edits are published; preserve and expose those edits through the existing explicit sync/publication workflow. No automatic recursive dependency publication or claim transfer is introduced.

## Done when

- A real two-worktree race to claim one issue has one winner and an explicit loser, with no overwritten ownership.
- Concurrent issue allocation and independent issue updates preserve unique IDs and both records; conflicting same-record updates are visible.
- A controlled slow review permits unrelated issue operations to finish; changed review inputs prevent stale finalization.
- Automated interleaving tests cover the production publication/lock boundaries and retain existing SDLC gates.

- Independent clones targeting the same remote have tested allocation/claim/body-update conflict handling, as well as the existing shared-worktree cases.
- A parent-slot-driven dependency issue can be created, updated and explicitly published using normal SDLC without losing local issue-body edits or assuming a shared local lock.

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

Revised after estimate-quality feedback: separate concrete concerns instead of treating all receipt behavior, all callers or all fixtures as one module. Each implementation primitive includes its own unit tests; the separately itemized integration harnesses exercise cross-component processes and protocol conformance, not those unit tests. Gate reviews are three real boundaries. The calibration source is marked stale (#127), so the result is provisional focused ship-hours.

Derivation: greenfield concerns use base design 1.5 × 0.2 detailed-spec discount = 0.30 and base impl 0.8 × 0.4 AI-paired scale = 0.32. Standard-library JSON/filesystem primitives halve receipt IO/decoding design to 0.15; domain ownership/recovery/lifecycle have no library substitute. Each API integration uses base design 2.0 × 0.2 = 0.40 and base impl 1.5 × 0.4 = 0.60. These are behavioral integrations, not mechanical refactors. Smaller module uses 0.3 × 0.2 design and 0.5 × 0.4 impl. Each docs group uses 0.10 design and 0.20 × 0.4 impl. Reviews use 0 design and 0.5 × 0.4 impl. Issue authoring uses 1.0 design without double-discounting the authoring work itself, and 0.3 × 0.4 impl. Familiarity remains 1.0 (existing Go/Git stack); thorough-plan design buffer is 15%.

| Primitive | Concrete concern | Design | AI-paired implementation |
|---|---|---|---|
| issue-spec | Audit and durable design | 1.00 | 0.12 |
| greenfield-go-module | Publication transition model | 0.30 | 0.32 |
| greenfield-go-module | Ownership provenance and metadata | 0.30 | 0.32 |
| greenfield-go-module | Record preconditions and reconciliation policy | 0.30 | 0.32 |
| greenfield-go-module | Atomic receipt IO | 0.15 | 0.32 |
| greenfield-go-module | Strict receipt schema and decoding | 0.15 | 0.32 |
| greenfield-go-module | Crash recovery and predecessor lineage | 0.30 | 0.32 |
| greenfield-go-module | Receipt bounds and retirement | 0.30 | 0.32 |
| api-integration | Git evidence hooks and conditional push | 0.40 | 0.60 |
| api-integration | Unknown push outcome reconciliation | 0.40 | 0.60 |
| api-integration | Claim caller integration | 0.40 | 0.60 |
| api-integration | Create and reallocation integration | 0.40 | 0.60 |
| api-integration | Sync baseline and conflict integration | 0.40 | 0.60 |
| greenfield-go-module | Prepared review transition model | 0.30 | 0.32 |
| api-integration | Plan review prepare/finalize integration | 0.40 | 0.60 |
| api-integration | Estimate review prepare/finalize integration | 0.40 | 0.60 |
| api-integration | Close and milestone review record integration | 0.40 | 0.60 |
| greenfield-go-module | Reviewer timeout/process lifetime shell | 0.30 | 0.32 |
| api-integration | Stateful Git protocol fake and conformance | 0.40 | 0.60 |
| api-integration | Independent-clone and worktree publication harness | 0.40 | 0.60 |
| api-integration | Reviewer subprocess barrier harness | 0.40 | 0.60 |
| api-integration | Nested dependency CLI conformance | 0.40 | 0.60 |
| smaller-go-module | Vocabulary and generated contract compatibility | 0.06 | 0.20 |
| atlas-docs | Claim and issue command help | 0.10 | 0.08 |
| atlas-docs | Issue sync and lifecycle atlas | 0.10 | 0.08 |
| atlas-docs | Review/gate atlas and command help | 0.10 | 0.08 |
| atlas-docs | README and project checkpoint | 0.10 | 0.08 |
| milestone-review | M1 boundary | 0.00 | 0.20 |
| milestone-review | M2 boundary | 0.00 | 0.20 |
| milestone-review | Issue close boundary | 0.00 | 0.20 |

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=1.00 impl=0.12
item: greenfield-go-module design=0.30 impl=0.32
item: greenfield-go-module design=0.30 impl=0.32
item: greenfield-go-module design=0.30 impl=0.32
item: greenfield-go-module design=0.15 impl=0.32
item: greenfield-go-module design=0.15 impl=0.32
item: greenfield-go-module design=0.30 impl=0.32
item: greenfield-go-module design=0.30 impl=0.32
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: greenfield-go-module design=0.30 impl=0.32
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: greenfield-go-module design=0.30 impl=0.32
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: smaller-go-module design=0.06 impl=0.20
item: atlas-docs design=0.10 impl=0.08
item: atlas-docs design=0.10 impl=0.08
item: atlas-docs design=0.10 impl=0.08
item: atlas-docs design=0.10 impl=0.08
item: milestone-review design=0.00 impl=0.20
item: milestone-review design=0.00 impl=0.20
item: milestone-review design=0.00 impl=0.20
design-buffer: 0.15
total: 21.28
```

Design 8.66 × 1.15 + implementation 11.32 = 21.28 hours (rounded).

## Plan

Proposed durable plan: [concurrent workflows](../plans/000244-slots-v2-concurrent-workflows-plan.md). Implementation awaits operator approval and the change-code gate.

- [ ] M1 — Guard issue reservation and publication across worktrees and clones.
- [ ] M2 — Release external-review locks, reject stale results, and verify dependency workflows.

## Log

### 2026-09-22 — fresh v2 task

Created from the agreed workspace/UI contract and the request for a clean task breakdown. Implementation has not started; estimates follow design approval.

## Revisions

### 2026-09-23 — Independent dependency repositories use normal SDLC

Reason: operator agreed nested environments, ordinary remote dependency clones and existing per-repository publication. Delta: added the authoritative scope clarification and acceptance criteria above; original task context remains as provenance. Added #243 as a prerequisite for the nested identity contract. No implementation or lifecycle-status change is claimed by this revision.

### 2026-09-23 — Live audit and proposed engineering design

Claimed and entered start-plan. Read-only audits confirmed same-record last-writer-wins publication, missing claim identity, same-slug allocation ambiguity, main-path publication divergence, and planning-review lock contention. Close already unlocks but persists review records before checking freshness. Proposed one guarded trunk publication path with private durable intents and exact read-set validation before review persistence (ARCH-DRY, ARCH-ORDER, ARCH-SECURE). Durable plan records two review boundaries and deterministic fake/real-Git tests; implementation has not started. This revision replaces the preliminary task outline, preserving its scope.

### 2026-09-23 — Design review corrections

Fresh-eyes review identified missing rejected-intent recovery and an invalid shared-worktree test barrier. The plan now distinguishes confirmed rejection from unknown outcomes, allows resolved body publication after explicit Git integration, and tests clone CAS contention separately from common-directory serialization. No code changes or implementation verification are claimed.

### 2026-09-23 — Proposed plan review passed

Fresh-context re-review approved plan commit `7f0850d` with no remaining blocking findings. The operator asked why locks remain with separate worktrees: linked worktrees isolate files but share Git repository state; short mutation locks remain, external reviews run unlocked, and independent clones require remote publication preconditions. Awaiting implementation approval; no production code changed.

### 2026-09-23 — Implementation authorized; plan-quality refinements

Operator approved implementation. Baseline targeted Git/publication/lock tests passed. The first plan-quality gate raised PQ-1–PQ-3; the durable plan now specifies persisted ownership/receipt formats, named adversarial test strategies, and review interruption transitions. Implementation still waits for the gate; no code changed.

### 2026-09-23 — Estimate revised after gate feedback

The first 6.88h derivation under-itemized independent receipt concerns, production integrations and conformance harnesses. Replaced it with concrete concern rows from the same calibration method; no implementation scope added. The original estimate-quality refusal remains in `/tmp/ariadne-244-change-code-4.log`.
