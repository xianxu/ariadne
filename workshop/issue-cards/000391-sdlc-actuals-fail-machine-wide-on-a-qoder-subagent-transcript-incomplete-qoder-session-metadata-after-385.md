---
id: 000391
status: codecomplete
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
started: 2026-10-15T13:31:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: N/A
tracker:
    version: 1
    completion:
        token: close-3d0396d67887
        repository: github.com/xianxu/ariadne
        reviewed_head: 6bdd2b646b30f7c880c369c876a6eebb2483f615
        evidence_commit: 2b3b2208dca68dbec5e4e42f3ab794227fbf9f0e
---

# sdlc actuals fail machine-wide on a Qoder subagent transcript ('incomplete qoder session metadata') after #385

## Problem

After ariadne#385 (Qoder transcripts in sdlc) landed, `sdlc close` refuses to compute actuals on this machine:
"can't auto-compute, source ~/.qoder/projects/-Users-xianxu-workspace-pair/6d3388c7-…/subagents/agent-aExplore-….jsonl:
incomplete qoder session metadata" (ariadne:4 closing #387, 2026-10-15). A Qoder **subagent** transcript (under a
session's `subagents/` directory) lacks the session metadata the new reader requires, and one unreadable file makes
the whole measurement fail. Since every close/actual scans the transcript roots, this likely blocks **every**
`sdlc close` and `sdlc actual` on any machine with Qoder subagent files.
