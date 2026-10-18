---
id: 000443
status: open
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
started: 2026-10-18T10:57:47-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# sdlc actual: a parked claim double-counts time the slot spent on other issues

## Problem

A slot that holds a claim on one issue while it works others gets the same time counted twice. pair:6 held pair#636 from 2026-10-17 10:14 to 2026-10-18 08:41, parked behind pair#647, and meanwhile worked pair#649 and pair#656, which each adopted their own measured actuals. `sdlc actual --issue 636` then reported 9.28 h over 11 admitted sessions, mostly time already counted toward #649 and #656. The slot had to close with --no-actual.
