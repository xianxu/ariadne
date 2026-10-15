---
id: 000387
status: codecomplete
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
actual_hours: 1.11
tracker:
    version: 1
    completion:
        token: close-d4f1e2adbaf6
        repository: github.com/xianxu/ariadne
        reviewed_head: 870f183161f3b9a02677cb355caeaa3afe7f99a2
        evidence_commit: f1a111faf6491fa81bb8bb81e22234e6fcea12a8
---

# Agent: trailer says 'human' (not 'sdlc') for a commit made outside an agent session

## Problem

ariadne#367's `Agent:` trailer says `Agent: sdlc` when `sdlc` runs in a plain shell with no agent session, i.e. when
the operator makes the commit by hand. Velocity and usage numbers then credit the operator's hand work to "sdlc". The
operator (2026-10-15): "say human in that case".
