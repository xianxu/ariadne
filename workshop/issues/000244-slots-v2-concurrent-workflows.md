---
id: 000244
status: working
deps: [ariadne#242, ariadne#243]
github_issue:
created: 2026-09-22
updated: 2026-09-23
estimate_hours:
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
