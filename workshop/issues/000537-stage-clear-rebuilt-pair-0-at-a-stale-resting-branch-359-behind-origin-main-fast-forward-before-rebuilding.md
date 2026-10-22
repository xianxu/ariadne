---
id: 000537
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '00af734b305145f0a8a0894ecb7778e377bd9fd0' # card fields mirrored from issue-cards; edit via sdlc
---

# stage --clear rebuilt pair:0 at a stale resting branch (359 behind origin/main); fast-forward before rebuilding

## Problem

## Spec

10-21 18:40, pair: after #752 landed, the stage holder ran `sdlc stage --clear` on pair:0 (the operator's live
checkout, resting branch `main`). It returned pair:0 to its local `main`, which was 359 commits behind
`origin/main` (last fast-forwarded 02:13), and rebuilt and verified the Go binaries **at that stale commit**. If the
operator had restarted Couch then, they would have run a day-old build, losing everything that landed today. The
TL caught it by checking `git status` and fast-forwarded plus rebuilt by hand.

Fix: `stage --clear` fetches and fast-forwards the resting branch to `origin/main` before rebuilding (it's a resting
branch: fast-forward only, never committed on). If it can't fast-forward (diverged or dirty), refuse and name the
state instead of building a stale tree.

## Done when

- `stage --clear` on a target whose resting branch is behind origin/main ends on origin/main's tip, with binaries
  built from it (test with a fake remote).
- A diverged resting branch makes `--clear` refuse before building (test).

## Plan

- [ ]

## Log

### 2026-10-21
