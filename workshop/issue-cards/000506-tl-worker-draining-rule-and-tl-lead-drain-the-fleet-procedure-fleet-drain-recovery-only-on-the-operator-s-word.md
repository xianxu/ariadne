---
id: 000506
status: working
created: 2026-10-20
updated: 2026-10-22
estimate_hours:
github_issue:
started: 2026-10-22T02:11:50-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# tl-worker draining rule and tl-lead drain-the-fleet procedure (fleet-drain; recovery only on the operator's word)

## Problem

Project fleet-drain (ops `teams/devinfra/projects/fleet-drain/`, operator-designed 10-20) puts the drain's judgment
in the agents and the steward; Couch only flags slots and delivers notices (pair#722 re-scoped). The agents and the
steward need the rules.
