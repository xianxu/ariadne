---
id: 000387
status: open
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
started: 2026-10-15T11:01:43-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
---

# Agent: trailer says 'human' (not 'sdlc') for a commit made outside an agent session

## Problem

ariadne#367's `Agent:` trailer says `Agent: sdlc` when `sdlc` runs in a plain shell with no agent session, i.e. when
the operator makes the commit by hand. Velocity and usage numbers then credit the operator's hand work to "sdlc". The
operator (2026-10-15): "say human in that case".
