---
id: 000434
status: open
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: '99cd98bb965ecc57f3c7085beeafae9a862fd4d3' # card fields mirrored from issue-cards; edit via sdlc
---

# flake: TestLocalCheckSignalCancelsQueuedHelper cleanup subtest under load

## Problem

`TestLocalCheckSignalCancelsQueuedHelper` (cmd/sdlc/localcheck_signal_unix_test.go) fails intermittently under load and passes alone:

- 2026-10-17, ariadne:2 on the #431 branch (a broad `go test ./cmd/sdlc/ -run 'LocalCheck|CIGate|Suite|Landing|Merge'` run alongside other slots' suites): subtest `terminated/cleanup` failed at localcheck_signal_unix_test.go:171, "cleanup never reached its blocked Git seam". Rerun alone: 3/3 pass.
- Earlier (during #424, ariadne:2): the same test failed once under load, passed 3/3 alone and in the full suite.

The cleanup phase waits for the fixture's blocked `git worktree remove --force` to write its pid; under load the run's 50 ms `localCheckCleanupTimeout` (set by SDLC_BLOCK_PHASE=cleanup) likely expires before the shimmed git starts, so the seam is never reached.

## Spec

## Done when

- The cleanup subtest doesn't depend on process-start latency under load (e.g. the fixture waits for the shim's readiness, or the assertion accepts a deadline that fired before the seam), and a loaded run (`-count=20 -parallel` alongside a suite) passes.

## Plan

- [ ]

## Log

### 2026-10-17
