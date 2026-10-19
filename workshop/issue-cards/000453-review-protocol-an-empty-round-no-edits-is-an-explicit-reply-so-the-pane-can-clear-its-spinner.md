---
id: 000453
status: working
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
started: 2026-10-18T22:24:44-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# review protocol: an empty round (no edits) is an explicit reply, so the pane can clear its spinner

## Problem

Since ariadne#437 / pair#664 retired the legacy review handoff, an agent that answers a review round with no edits writes no round artifact. In pair's review pane the awaiting spinner and the Alt+c pending block then stay up until the next send. The legacy path cleared them on an empty handoff; #437 defines no empty-round reply. Found by pair:2 closing pair#664 (2026-10-18).
