---
id: 000400
status: open
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '31e78e2ae6bded9a6f14f6d193c8ecd849f22411' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc local-check cancel: stop only your own run (an agent killed another slot's suite by PID)

## Problem

2026-10-15: pair:5 wanted to stop its own `sdlc local-check run --name suite` and picked a PID with
`ps -axo pid,ppid,command`, taking the first global match. That PID was parley.nvim:2's suite (parley#337's final
run on ff50e564). `kill -INT` ran before the ownership check (`lsof -d cwd`) was read, and it cancelled another
slot's full suite (exit 130), costing a requeue behind three other suites.

Nothing today lets an agent stop *its own* local check without hand-picking PIDs, and the global process table holds
every slot's identical-looking `sdlc local-check run --name suite` commands.

## Spec

- `sdlc local-check cancel [--name suite]` stops only the run started from the current checkout (or the current
  slot's worktree): it finds the run through its own record (the in-flight file under `.git/sdlc/`, with the
  runner's PID and process-group ID, written at start), checks it's alive and still that run (start time and
  command match), and signals its process group. Unknown, stale or foreign → refuse, naming what it found.
- The runner records PID, PGID and start time in its in-flight record when it starts (if it doesn't already).
- tl-worker's rules gain one line: stop your own checks with `sdlc local-check cancel`, never `kill` a PID found by
  `ps` (phrase-guarded in the help tests).

## Done when

- A fake-backed test: two in-flight runs from two checkouts; cancel from one stops only its own process group and
  leaves the other running; a stale record (dead or reused PID) is refused.
- tl-worker --help states the rule.

## Plan

- [ ]

## Log

### 2026-10-15
