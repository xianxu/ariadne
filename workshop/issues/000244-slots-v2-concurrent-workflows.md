---
id: 000244
status: working
deps: [ariadne#242, ariadne#243]
github_issue:
created: 2026-09-22
updated: 2026-09-23
estimate_hours:
started: 2026-09-23T11:35:53-07:00
flow: {kind: full, provenance: inferred}
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

### Simplified publication scope — 2026-09-23 (current)

This agreement supersedes ownership-token and receipt-based engineering proposals. Origin/main's issue status is the reservation authority: a fresh open→working conditional publication has one winner; any already-working claim refuses, including a repeat from the original worker. No ownership identity, private publication receipt store or automatic takeover is required.

For issue-body and related design updates, the agent/operator selects a coherent local documentation commit. It may contain the issue file, separate plan and deliberately selected same-repository project/design records; SDLC does not infer semantic relatedness. Publish that commit's change onto fresh origin/main using Git three-way merge and conditional remote publication. Preserve unrelated updates, expose conflicts, and preserve the source commit and caller state. Compatible same-record changes may merge. Repeated publication must not undo later changes.

Apply the same mechanism in :0, numbered slots, ordinary feature worktrees and private dependency clones, without a special whole-local-main push path. Local issue sync remains issue-file-only and network-free; including a separate plan is an explicit commit-selection decision. Repository mutation locks remain short; external reviews run unlocked and validate all relevant inputs before recording results.

## Done when

- A real two-worktree race to claim an open remote issue has one winner and an explicit already-working loser; repeated claims do not imply ownership.
- Concurrent allocation preserves unique IDs. Explicitly selected documentation commits preserve unrelated changes; compatible edits merge and conflicting edits are visible.
- A controlled slow review permits unrelated issue operations to finish; changed review inputs prevent stale finalization.
- Automated interleaving tests cover the production publication/lock boundaries and retain existing SDLC gates.

- Independent clones targeting the same remote have tested allocation/claim/body-update conflict handling, as well as the existing shared-worktree cases.
- A parent-slot-driven dependency issue can be created, updated and explicitly published using normal SDLC without losing local issue-body edits or assuming a shared local lock.

- A selected commit can intentionally group an issue and its separate plan/project records; no implicit artifact discovery or unrelated code publication occurs.
- Primary :0 uses the same selected-commit publication behavior and preserves unrelated local commits and dirty files.

## Estimate

The previous 21.28h estimate applied to the removed design; it remains in Git history at `bbd3122`. Derive a replacement after the revised plan clears plan-quality; no new estimate is asserted by this scope update.

## Plan

Current durable plan: [concurrent workflows](../plans/000244-slots-v2-concurrent-workflows-plan.md). The operator approved the simplified scope. Prior gate clearance applied to v1; implementation is paused pending revised gate evidence.

- [ ] M1 — Fresh status claims and explicit commit publication across all slots/clones.
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

### 2026-09-23 — Entered implementation

Plan-quality has no open findings; estimate-quality returned INFO with a provisional 21.28h concern-by-concern derivation. `sdlc change-code` created the issue branch in place. M1 divides into evidenced Git transactions, pure publication/receipt state, and command integration. Root owns verification and boundary gates. No production change is yet claimed.

### 2026-09-23 — Simplified scope approved

Operator clarified that working on origin/main is sufficient reservation state and rejected complexity introduced for identifying claim retries. Agreed fresh conditional status claims, explicit agent-selected documentation commits (issue plus plan/project when deliberately included), Git-based merging/conflict handling, and identical publication in :0 and other slots/clones. Retained unlocked external reviews with stale-input validation. Preserved v1 design/estimate separately; current estimate is unset pending the revised design gate. Partial v1 code/docs remain uncommitted and unfinished; this turn changes only design artifacts.

### 2026-09-23 — Simplified design review

Fresh-context review approved the revised issue/plan with no blocking inconsistencies. The review confirms status-only claims, explicit commit selection, identical publication across checkout kinds, and unlocked reviews with validation before persistence. This is design review only; revised implementation gate/estimate evidence is still required before code resumes.

### 2026-09-23 — Resume simplified implementation

Operator authorized implementation and removal of the obsolete design artifact. Retired the uncommitted receipt draft outside the repository and restored its tracked source/docs to HEAD. Re-entering the design/estimate gate for the current scope; earlier gate records are historical evidence only.
