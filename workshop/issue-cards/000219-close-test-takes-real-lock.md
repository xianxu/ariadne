---
id: 000219
status: wontfix
created: 2026-09-09
updated: 2026-10-09
estimate_hours:
github_issue:
started: 2026-10-09T18:40:42-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
tracker:
    version: 1
    abandoned: {}
---

# close_test flag assertion takes the real repo lock, so the suite cannot run inside an sdlc transaction

## Problem

`TestCloseCmd_MilestoneFlagHidden` (`cmd/sdlc/close_test.go:131`) executes a real
`NewCloseCmd()` to assert that `--milestone` is hidden and that `close` redirects
to `milestone-close`. That is a **flag-visibility assertion**, but executing the
command takes the production repo transaction lock at `.git/sdlc.lock`.

**Consequence: `go test ./cmd/sdlc/` cannot report green from inside any sdlc
transaction.** Every mutating verb holds that lock, so a suite run launched from
within `change-code`, `close`, or `merge` blocks in `repolock.Acquire`
(`repolock.go:210`) until `DefaultWaitTimeout` — 30 minutes (`repolock.go:19`).

Two ways this bites, and the second is worse:

1. **The gates cannot verify what they ask for.** A close gate that demands
   `go test ./cmd/sdlc/...` as evidence runs it inside the transaction that
   forbids it succeeding.
2. **It presents as a hang, not a lock wait.** Found during ariadne#218's
   plan-quality review: the reviewer saw `go test` blow its 600s default with a
   goroutine parked in syscall for nine minutes and reasonably read it as a hung
   child process. From a plain shell the same suite passes in 115-136s, so the
   two observations look irreconcilable until the lock is suspected. A defect that
   masquerades as a different defect costs more than its size.

Note `-timeout 1800s` does not rescue it: 30m equals `DefaultWaitTimeout`
exactly, so the two expire together.
