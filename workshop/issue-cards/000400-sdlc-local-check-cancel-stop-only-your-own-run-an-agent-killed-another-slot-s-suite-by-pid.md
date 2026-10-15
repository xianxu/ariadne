---
id: 000400
status: open
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
---

# sdlc local-check cancel: stop only your own run (an agent killed another slot's suite by PID)

## Problem

2026-10-15: pair:5 wanted to stop its own `sdlc local-check run --name suite` and picked a PID with
`ps -axo pid,ppid,command`, taking the first global match. That PID was parley.nvim:2's suite (parley#337's final
run on ff50e564). `kill -INT` ran before the ownership check (`lsof -d cwd`) was read, and it cancelled another
slot's full suite (exit 130), costing a requeue behind three other suites.

Nothing today lets an agent stop *its own* local check without hand-picking PIDs, and the global process table holds
every slot's identical-looking `sdlc local-check run --name suite` commands.
