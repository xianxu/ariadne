---
id: 000338
status: open
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-12
estimate_hours:
card_mirror: 'bf34be971abd5f4d6644642e564beb6b104d27aa' # card fields mirrored from issue-cards; edit via sdlc
---

# main is red: TestPlanningReviewCompetingResults fails on clean main (planning review stale: issue file changed)

## Problem

Reported by ariadne:2 (10-12 ~12:0x): `TestPlanningReviewCompetingResults` in cmd/sdlc fails on a clean origin/main worktree (c1afb049), reproducibly: `plan-quality dispatch failed: planning review stale: …000206-issue-sync-verb.md changed; rerun change-code`. `make test` is therefore red on main for every ariadne slot, and every close cites it as pre-existing. The last changes to the test came from #325 (9933e8db, condition waits replacing deadlines) and #330 (9616217b).

## Spec

Reproduce on clean main, find the cause (likely #325's change to how the test waits interacts with the planning review's staleness check: the test's fixture issue file changes between the review's snapshot and its check, or a wait now observes a state the old deadline skipped), and fix the root cause in the test or the product. Check whether #308's issue-new change (details published to main) also moved the fixture. No quarantine.

## Done when

- The test passes on main repeatedly (10 runs) and under make test; the cause is in the Log.

## Plan

- [ ]

## Log

### 2026-10-12
