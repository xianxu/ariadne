---
id: 000434
status: codecomplete
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
github_issue:
started: 2026-10-17T21:15:30-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.79
tracker:
    version: 1
    completion:
        token: close-ada9030e6d8f
        repository: github.com/xianxu/ariadne
        reviewed_head: 41f0e729593a39a5a1aa2635aa622646dbc5a04e
        evidence_commit: 30829fb6a048de58d5564741887c67832b86bce5
---

# flake: TestLocalCheckSignalCancelsQueuedHelper cleanup subtest under load

## Problem

`TestLocalCheckSignalCancelsQueuedHelper` (cmd/sdlc/localcheck_signal_unix_test.go) fails intermittently under load and passes alone:

- 2026-10-17, ariadne:2 on the #431 branch (a broad `go test ./cmd/sdlc/ -run 'LocalCheck|CIGate|Suite|Landing|Merge'` run alongside other slots' suites): subtest `terminated/cleanup` failed at localcheck_signal_unix_test.go:171, "cleanup never reached its blocked Git seam". Rerun alone: 3/3 pass.
- Earlier (during #424, ariadne:2): the same test failed once under load, passed 3/3 alone and in the full suite.

The cleanup phase waits for the fixture's blocked `git worktree remove --force` to write its pid; under load the run's 50 ms `localCheckCleanupTimeout` (set by SDLC_BLOCK_PHASE=cleanup) likely expires before the shimmed git starts, so the seam is never reached.
