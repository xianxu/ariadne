---
id: 000456
status: codecomplete
created: 2026-10-18
updated: 2026-10-19
estimate_hours: 2.54
github_issue:
started: 2026-10-19T01:26:56-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 2.32
tracker:
    version: 1
    completion:
        token: close-2abd9be620fa
        repository: github.com/xianxu/ariadne
        reviewed_head: 5dbc3fc3c36276e7c15858a9de82eac60063cc88
        evidence_commit: 5cb60152912dc51e8b1683f29cde69a5c0523265
---

# tl-lead teams: several teams per operations repo, team.md, cross-team routing, a machine steward

## Problem

tl-lead assumes one hub per operator, the slot Couch started from, with its state in `teams/<team>/` of the ops repo. The operator (2026-10-18, project introduce-team) wants teams as first-class: an operations repo holds several teams, each with its own TL (hub), named after the team, under a different agent contract than workers (tl-lead vs tl-worker), and aware of each other.
