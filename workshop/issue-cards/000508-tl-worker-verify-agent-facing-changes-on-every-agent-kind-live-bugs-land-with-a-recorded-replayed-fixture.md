---
id: 000508
status: codecomplete
created: 2026-10-20
updated: 2026-10-21
estimate_hours:
github_issue:
started: 2026-10-21T01:53:24-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 3.16
tracker:
    version: 1
    completion:
        token: close-51d6a22dd31d
        repository: github.com/xianxu/ariadne
        reviewed_head: 42e0402d5f674bcf2b8f9142278dd29bf3e13098
        evidence_commit: 23f3760f1f6e4d22a9e85481f670254aa57ab310
---

# tl-worker: verify agent-facing changes on every agent kind; live bugs land with a recorded replayed fixture

## Problem

Live-only bugs on 10-20 (pair#727, #735, #742, the drain) were each found by hand, and pair#717's live check passed
on the agent it was built for and failed on Codex. Operator decisions (ops side doc
`2026-10-20-19-conformance-structure.md`): every agent kind, every time; a live bug's fix lands with its recorded
capture as a replayed fixture.
