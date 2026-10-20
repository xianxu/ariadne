---
id: 000494
status: done
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
started: 2026-10-20T06:34:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 5.43
tracker:
    version: 1
    completion:
        token: close-e3fb6653eeb0
        repository: github.com/xianxu/ariadne
        reviewed_head: 058f4f15587adf7659fed2095b9b8f9750a1de02
        evidence_commit: 194143620069696390ae4efbfa60028e1b267c8e
        landed_commit: 60aabea5db6aeeb3eb5f5f4997a9f41ce9939726
---

# Train car's own merge warns 'active issue missing without confirmed archive' after the train already archived it

## Problem

In pair's first multi-car train (176, 10-20 05:48), car pair#717's own `sdlc merge` (PR 483) ended with "active issue
missing without confirmed archive", although the train had already archived the issue (it's under history on main
and its card is done). Harmless, but it reads as an error at the end of a successful landing, and a worker can't tell.
