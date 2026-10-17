---
id: 000411
status: done
created: 2026-10-16
updated: 2026-10-16
estimate_hours: 1.13
github_issue:
started: 2026-10-16T15:06:50-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.94
tracker:
    version: 1
    completion:
        token: close-e9b630d6188c
        repository: github.com/xianxu/ariadne
        reviewed_head: 1813e25dbdb94a2e86e7f500c1950e6d89cf559a
        evidence_commit: 2b7613cb16b69a9f4e285fb1892522c282f4ce3a
        landed_commit: 51400e9ec9a06dce27a3d339c68b9eb93ae42f2d
---

# start-plan's weave refresh fails the first run in a Claude sandbox (settings.json is write-denied); a stale settings file goes unnoticed

## Problem

In a Claude slot (parley.nvim:4, 2026-10-16), every first `sdlc start-plan` fails: its weave refresh tries to write
`.claude/settings.json`, which Claude Code's sandbox write-denies (`operation not permitted`; settings files are in the
sandbox's deny-within-allow list by design, so an agent can't rewrite its own permissions). Re-running start-plan
then goes through, so the slot's `.claude/settings.json` may lag the woven one without anyone noticing.
