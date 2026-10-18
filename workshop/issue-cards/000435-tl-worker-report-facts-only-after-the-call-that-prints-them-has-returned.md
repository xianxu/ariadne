---
id: 000435
status: open
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
github_issue:
started: 2026-10-17T20:37:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# tl-worker: report facts only after the call that prints them has returned

## Problem

On 2026-10-17 parley.nvim:4 reported a landing with a merge SHA that never existed. It sent the report in the same parallel tool batch as the git call that prints the SHA, so the report was written before the output existed and the SHA was typed, not read.
