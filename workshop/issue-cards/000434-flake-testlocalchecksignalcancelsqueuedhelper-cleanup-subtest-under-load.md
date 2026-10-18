---
id: 000434
status: open
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
github_issue:
---

# flake: TestLocalCheckSignalCancelsQueuedHelper cleanup subtest under load

## Problem

`TestLocalCheckSignalCancelsQueuedHelper` (cmd/sdlc/localcheck_signal_unix_test.go) fails intermittently under load and passes alone:

- 2026-10-17, ariadne:2 on the #431 branch (a broad `go test ./cmd/sdlc/ -run 'LocalCheck|CIGate|Suite|Landing|Merge'` run alongside other slots' suites): subtest `terminated/cleanup` failed at localcheck_signal_unix_test.go:171, "cleanup never reached its blocked Git seam". Rerun alone: 3/3 pass.
- Earlier (during #424, ariadne:2): the same test failed once under load, passed 3/3 alone and in the full suite.

The cleanup phase waits for the fixture's blocked `git worktree remove --force` to write its pid; under load the run's 50 ms `localCheckCleanupTimeout` (set by SDLC_BLOCK_PHASE=cleanup) likely expires before the shimmed git starts, so the seam is never reached.
