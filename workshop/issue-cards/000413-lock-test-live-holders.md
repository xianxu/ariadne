---
id: 000413
status: open
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
github_issue:
severity: medium
started: 2026-10-16T17:13:13-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# Concurrent dead-holder lock test marks its live contenders dead and flakes with missing metadata

## Problem

During ariadne#411 recorded suite on 986653326c2242c56ef7a095a3ebab498f98918f, cmd/sdlc/internal/repolock failed TestConcurrentAcquireReclaimsDeadHolderOnce (0.03s): repolock_test.go:415 reported write sdlc repo lock metadata: open .../.git/sdlc.lock/meta.json: no such file or directory. The package is unchanged from current main. The full suite is still collecting other results; run ID 5f54a534ecf74f0c3b0c7af2873769a8.

The fixture supplies ProcessAlive: func(int) bool { return false } for the original dead PID 555 and both newly acquiring holders 111/222. This makes the live holders appear dead as well. That is a likely invalid liveness model; this observation alone does not establish a production lock bug.
