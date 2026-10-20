---
id: 000483
status: working
created: 2026-10-19
updated: 2026-10-20
estimate_hours:
github_issue:
started: 2026-10-20T04:06:52-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# Continuing an issue whose branch already landed: refuse with 'close it and file a follow-up' (pair#622 case)

## Problem

pair#622 (10-19): the issue's code landed (PR 406), but its live acceptance stayed open and needed a little more
work. `sdlc change-code --issue 622` insists on the canonical branch, which already merged; `--name` and `--issue`
are mutually exclusive; so there is no supported way to continue an issue whose branch has landed. The worker
improvised a fresh branch outside the gates.
