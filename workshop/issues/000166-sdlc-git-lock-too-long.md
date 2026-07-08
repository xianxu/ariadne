---
id: 000166
status: codecomplete
deps: []
github_issue:
created: 2026-07-07
updated: 2026-07-07
estimate_hours: 2.05
started: 2026-07-07T16:43:47-07:00
actual_hours: 1.04
---

# sdlc git lock is too long

## Problem

When `sdlc` runs a long review action, it holds `.git/sdlc.lock` for the entire
duration. That blocks unrelated `sdlc` commands that need git state, even while
the long-running step is waiting on external review/model work rather than
mutating local repo state.

## Spec

Minimize the duration of the SDLC repo transaction lock for review-bearing close
commands. The lock should cover local repo mutations and coherent reads that must
be serialized, but it should not wrap long external review work when that work can
safely run outside the critical section.

Design:

- Treat `sdlc close` and `sdlc milestone-close` as manually locked commands
  rather than command-wrapper locked commands. They still need `.git/sdlc.lock`,
  but not for the whole `RunE`.
- Run `computeClose` and review-window resolution inside a locked critical
  section. This protects the issue/project/git reads that form the review input.
- Release the lock while `dispatchBoundaryReview` invokes the external reviewer.
  This is the long wait that should not block unrelated `sdlc` work.
- Before finalizing after a finalizing verdict, reacquire the lock and verify the
  repo state that was reviewed is still current. If HEAD or the issue file changed
  during the unlocked review, halt without writing the close so the operator can
  rerun against the new state.
- Keep other mutating commands automatically wrapped by the existing centralized
  lock wrapper (ARCH-DRY). The new path should reuse the same acquire/release
  primitive rather than creating a second lock implementation (ARCH-PURE).

## Done when

- Long-running review actions no longer hold `.git/sdlc.lock` for their full
  runtime.
- The mutation/read sections that still require serialization remain protected.
- Tests prove a review action releases the lock while external review work is in
  progress, without allowing unsafe concurrent repo mutation.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec             design=0.3 impl=0.1
item: cross-cutting-refactor design=0.4 impl=0.2
item: smaller-go-module      design=0.2 impl=0.2
item: atlas-docs             design=0.1 impl=0.1
item: milestone-review       design=0.0 impl=0.3
design-buffer: 0.15
total: 2.05
```

## Plan

- [x] Locate the SDLC transaction-lock call sites around close/milestone/review
      flows and identify which steps truly need serialization.
- [x] Add a regression test with a controllable slow review seam that observes
      the lock is released during the slow external work.
- [x] Refactor the flow to use narrower lock scopes while preserving locked
      repo mutations and final state commits.
- [x] Run targeted tests plus the relevant `sdlc` package suite.

## Log

### 2026-07-07
- 2026-07-07: closed — go test ./cmd/sdlc -count=1; go test ./...; git diff --check; review verdict: SHIP

- Moved from pair#109 to ariadne#166 because the `sdlc` binary lives here.
- Plan: narrow the lock to close compute/finalize critical sections while running
  the external boundary review unlocked; re-check reviewed HEAD/issue state before
  finalization (ARCH-DRY, ARCH-PURE, ARCH-PURPOSE).
- Implemented manual repo-lock mode for close/milestone-close and regression tests
  for unlocked review dispatch plus stale issue/HEAD refusal.
- Verification: `go test ./cmd/sdlc -count=1` passed.
- Verification: `go test ./...` passed.
- Verification: `git diff --check` passed.
- Close review returned REWORK; fixing stale project-file validation and aligning
  the implemented lock mode with the plan's `RepoLockMode` entity.
- Review-fix verification: focused stale-guard/repolock tests, `go test
  ./cmd/sdlc -count=1`, `go test ./...`, and `git diff --check` passed.
- Second close review returned REWORK on artifact/doc hygiene: moved
  `CloseReviewSnapshot` to integration concepts, updated stale-check docs to
  include project files, and normalized the generated review sidecar whitespace.
