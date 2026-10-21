---
id: 000510
status: done
created: 2026-10-20
updated: 2026-10-21
estimate_hours:
github_issue:
started: 2026-10-21T02:17:18-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.55
tracker:
    version: 1
    completion:
        token: close-69cafd5c6dd4
        repository: github.com/xianxu/ariadne
        reviewed_head: dcc50b776875f0dd40a0d2639a5fed3a79063d5b
        evidence_commit: 7ec9141d37ca0c169f8683c0ce306ce7d4266e58
        landed_commit: a6a76cf067e021b620f09b8a8f1b0775e101dd52
---

# sdlc local-check cancel: select one run by id (not only --name); notice when queuing a second run of the same check on a different head

## Problem

`sdlc local-check cancel` selects runs only by `--name`. On 10-20, pair:6 had two queued `suite` runs from one
checkout (an original on an older head, and an affected-check fallback to the full suite on the current head, the
pair#738 case) and could not cancel just one: `--name suite` cancels both and drops the original's FIFO place, so it
had to cancel both and requeue.
