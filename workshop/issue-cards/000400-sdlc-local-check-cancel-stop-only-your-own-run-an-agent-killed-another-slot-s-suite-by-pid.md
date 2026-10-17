---
id: 000400
status: working
created: 2026-10-15
updated: 2026-10-17
estimate_hours:
github_issue:
started: 2026-10-17T15:01:16-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# sdlc local-check cancel: stop only your own run (an agent killed another slot's suite by PID)

## Problem

2026-10-15: pair:5 wanted to stop its own `sdlc local-check run --name suite` and picked a PID with
`ps -axo pid,ppid,command`, taking the first global match. That PID was parley.nvim:2's suite (parley#337's final
run on ff50e564). `kill -INT` ran before the ownership check (`lsof -d cwd`) was read, and it cancelled another
slot's full suite (exit 130), costing a requeue behind three other suites.

Nothing today lets an agent stop *its own* local check without hand-picking PIDs, and the global process table holds
every slot's identical-looking `sdlc local-check run --name suite` commands.
