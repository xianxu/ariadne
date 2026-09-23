---
id: 000244
status: working
deps: [ariadne#242, ariadne#243]
github_issue:
created: 2026-09-22
updated: 2026-09-23
estimate_hours: 6.88
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

The calibration source is tagged stale by `sdlc estimate-source` (#127), so these are provisional focused ship-hours, not an elapsed-time promise. Primitive decomposition includes design, integration, adversarial tests, documentation and three actual review boundaries (M1, M2, issue close). Familiarity is 1.0 for the existing Go/Git stack. The approved detailed plan earns the 0.2 spec discount and 15% design buffer. Implementation entries are already scaled to 40% of the v2 table; no second scale is applied.

| Primitive | Scope | Design derivation | Implementation derivation |
|---|---|---|---|
| issue-spec | audit, approval and design | 1.0, no spec discount on authoring itself | 0.3 × 0.4 = 0.12 |
| greenfield-go-module | publication decision model | 1.5 × 0.2 = 0.30; no library supplies our workflow semantics | 0.8 × 0.4 = 0.32 |
| greenfield-go-module | receipt persistence | 1.5 × 0.5 library × 0.2 spec = 0.15; standard JSON/filesystem primitives | 0.8 × 0.4 = 0.32 |
| api-integration | evidenced Git transaction | 2.0 × 0.5 reuse × 0.2 spec = 0.20; existing TrunkFile plumbing | 1.5 × 0.4 = 0.60 |
| api-integration | controlled Git/reviewer conformance fixtures | 2.0 × 0.2 = 0.40; domain-specific schedules | 1.5 × 0.4 = 0.60 |
| cross-cutting-refactor | all issue publication callers | 1.0 × 0.2 = 0.20 | 0.5 × 0.4 = 0.20 |
| greenfield-go-module | prepared review transitions | 1.5 × 0.2 = 0.30; existing snapshot supplies IO, domain decisions remain | 0.8 × 0.4 = 0.32 |
| cross-cutting-refactor | review caller integration | 1.0 × 0.2 = 0.20 | 0.5 × 0.4 = 0.20 |
| smaller-go-module | vocabulary and compatibility | 0.2 × 0.2 = 0.04 | 0.5 × 0.4 = 0.20 |
| atlas-docs | help, atlas, README and project | 0.10 | 0.2 × 0.4 = 0.08 |
| milestone-review × 3 | two milestones and final close | 0.0 each | 0.5 × 0.4 = 0.20 each |

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=1.0 impl=0.12
item: greenfield-go-module design=0.30 impl=0.32
item: greenfield-go-module design=0.15 impl=0.32
item: api-integration design=0.20 impl=0.60
item: api-integration design=0.40 impl=0.60
item: cross-cutting-refactor design=0.20 impl=0.20
item: greenfield-go-module design=0.30 impl=0.32
item: cross-cutting-refactor design=0.20 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: atlas-docs design=0.10 impl=0.08
item: milestone-review design=0.0 impl=0.20
item: milestone-review design=0.0 impl=0.20
item: milestone-review design=0.0 impl=0.20
design-buffer: 0.15
total: 6.88
```

Design subtotal 2.89 × 1.15 = 3.3235; implementation subtotal 3.56; total 6.8835, rounded to 6.88.

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
