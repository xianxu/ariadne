---
id: 000338
status: working
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-12
estimate_hours:
card_mirror: '54c9dd6aaddbc31d05ec7e47985cb6e45787e288' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T12:00:04-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# main is red: TestPlanningReviewCompetingResults fails on clean main (planning review stale: issue file changed)

## Problem

Reported by ariadne:2 (10-12 ~12:0x): `TestPlanningReviewCompetingResults` in cmd/sdlc fails on a clean origin/main worktree (c1afb049), reproducibly: `plan-quality dispatch failed: planning review stale: …000206-issue-sync-verb.md changed; rerun change-code`. `make test` is therefore red on main for every ariadne slot, and every close cites it as pre-existing. The last changes to the test came from #325 (9933e8db, condition waits replacing deadlines) and #330 (9616217b).

## Spec

Reproduce on clean main, find the cause (likely #325's change to how the test waits interacts with the planning review's staleness check: the test's fixture issue file changes between the review's snapshot and its check, or a wait now observes a state the old deadline skipped), and fix the root cause in the test or the product. Check whether #308's issue-new change (details published to main) also moved the fixture. No quarantine.

## Done when

- The test passes on main repeatedly (10 runs) and under make test; the cause is in the Log.

## Plan

- [x] Reproduce: the single test fails about 1 in 5 runs on main (not every run; it raced).
- [x] Root cause: confirmed with the competing CLI's output captured on failure.
- [x] Fix: one `planningAgentEnv` helper builds every planning CLI's environment; the competing CLI's leftover fixed deadlines use `testfix.HangCeiling`.
- [x] Verify: 10 runs of the `TestPlanningReview` family, then `make test`.

## Log

### 2026-10-12
- Cause (#325, 9933e8db): #325 moved the fake reviewer's spin bound out of the script into a `PLANNING_SPINS` env var, but set it only in `startPlanningReviewCLI`. The competing CLI in `TestPlanningReviewCompetingResults` builds its own env and never got it, so its reviewer's `[ "$n" -lt "" ]` errored (`integer expression expected`) and the reviewer exited at once instead of waiting at the barrier. Under the test's `--force`, that CLI finished change-code (a no-verdict round, gate bypassed, estimate clean) and wrote the issue file while the first CLI was still paused. When the first CLI revalidated, it saw the issue had changed and failed as stale. That only happens when the second CLI wins the race, so the failure was intermittent. #330 and #308 are not involved.
- Fix: one helper (`planningAgentEnv`) builds the env for both CLIs, so a CLI can't be launched with only part of the set (ARCH-DRY). The test's remaining 25s/10s/5s deadlines now use `testfix.HangCeiling`, following #325's own rule.
