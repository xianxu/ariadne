---
id: 000166
status: open
deps: []
github_issue:
created: 2026-07-07
updated: 2026-07-07
estimate_hours:
---

# sdlc git lock is too long

## Problem

When `sdlc` runs a long review action, it holds `.git/sdlc.lock` for the entire
duration. That blocks unrelated `sdlc` commands that need git state, even while
the long-running step is waiting on external review/model work rather than
mutating local repo state.

## Spec

Minimize the duration of the SDLC repo transaction lock for review-bearing
commands. The lock should cover local repo mutations and coherent reads that
must be serialized, but it should not wrap long external review work when that
work can safely run outside the critical section.

## Done when

- Long-running review actions no longer hold `.git/sdlc.lock` for their full
  runtime.
- The mutation/read sections that still require serialization remain protected.
- Tests prove a review action releases the lock while external review work is in
  progress, without allowing unsafe concurrent repo mutation.

## Plan

- [ ] Locate the SDLC transaction-lock call sites around close/milestone/review
      flows and identify which steps truly need serialization.
- [ ] Add a regression test with a controllable slow review seam that observes
      the lock is released during the slow external work.
- [ ] Refactor the flow to use narrower lock scopes while preserving locked
      repo mutations and final state commits.
- [ ] Run targeted tests plus the relevant `sdlc` package suite.

## Log

### 2026-07-07

- Moved from pair#109 to ariadne#166 because the `sdlc` binary lives here.
