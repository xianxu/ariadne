---
id: 000458
status: working
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
started: 2026-10-19T04:32:38-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# flake list: drop a same-head flake once a closed issue's fix merged after that head

## Problem

`sdlc flake list` keeps showing a same-head fail→pass flake under "not registered (file or register today)" after an issue has fixed it. The 2026-10-19 pilot in pair hit this: five of the eight listed tests had been fixed and closed already (pair#585, #589, #609, #659), and each failing head predates its fix merge. The daily view asks for a filing that would duplicate a closed issue, and the real work gets lost among them. The same happened in ariadne with TestConstructScripts on a2b56335, which predates the #448 fix (6d623285).
