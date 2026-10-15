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
        token: close-3fddd758849e
        repository: github.com/xianxu/ariadne
        reviewed_head: 83cf89adaae2526d9c5d35ae228dda063a84f75a
        evidence_commit: 060c365e739e6420b14cfdd131181b5f52f062f6
---

# Agent: trailer says 'human' (not 'sdlc') for a commit made outside an agent session

## Problem

ariadne#367's `Agent:` trailer says `Agent: sdlc` when `sdlc` runs in a plain shell with no agent session, i.e. when
the operator makes the commit by hand. Velocity and usage numbers then credit the operator's hand work to "sdlc". The
operator (2026-10-15): "say human in that case".
