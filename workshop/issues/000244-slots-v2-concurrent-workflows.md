---
id: 000244
status: open
deps: [ariadne#242]
github_issue:
created: 2026-09-22
updated: 2026-09-22
estimate_hours:
---

# Slots v2: concurrent issue workflows

## Problem

Two slots running SDLC concurrently must not claim the same issue silently, overwrite newer issue records, or block unrelated work throughout a long review.

## Spec

Project: `pair/workshop/projects/couch-slots-v2.md`. Fresh task derived from the current v2 contract; historical task bodies are not prerequisites or implementation plans.

Audit and fix current SDLC behavior with two worktrees of one repository. Authoritative reservation checks for issue creation and claiming must be fresh at publication; a loser receives an actionable conflict. Publishing an issue-body update must preserve concurrent unrelated records and refuse conflicting edits rather than replacing them from a stale checkout.

Keep repository mutation serialization where needed, while external review waits should not monopolize the shared repository lock. Revalidate the exact reviewed/prepared state after reacquiring authority. No automatic claim transfer is introduced by branching from another workspace. ARCH-SECURE and ARCH-DRY: use existing ownership/publication machinery, with fresh evidence at mutation boundaries. This is a new v2 acceptance contract; audit live code rather than inheriting old issue conclusions.

## Done when

- A real two-worktree race to claim one issue has one winner and an explicit loser, with no overwritten ownership.
- Concurrent issue allocation and independent issue updates preserve unique IDs and both records; conflicting same-record updates are visible.
- A controlled slow review permits unrelated issue operations to finish; changed review inputs prevent stale finalization.
- Automated interleaving tests cover the production publication/lock boundaries and retain existing SDLC gates.

## Plan

Task outline only; settle implementation design through start-plan before change-code.

- [ ] Map current claim/allocation/body-publication transactions and long lock scopes.
- [ ] Build failing concurrency fixtures for unmet acceptance cases and fix those boundaries.
- [ ] Verify two independent slot workflows and document contention/retry behavior.

## Log

### 2026-09-22 — fresh v2 task

Created from the agreed workspace/UI contract and the request for a clean task breakdown. Implementation has not started; estimates follow design approval.
