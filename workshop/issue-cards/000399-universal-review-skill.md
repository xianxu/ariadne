---
id: 000399
status: done
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
started: 2026-10-15T16:24:58-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.63
tracker:
    version: 1
    completion:
        token: close-eb263b3a1904
        repository: github.com/xianxu/ariadne
        reviewed_head: 95dced13d8b1d1e29eb533b7d608733c8cc96d01
        evidence_commit: 003f1dbf1c67dd15f4f64c9c7912815f405a4908
        landed_commit: f19623c46566c209a1b90c6d95a4714ce61212a9
---

# Universal review skill and canonical round export

## Problem

Universal-agent-input-interface S2a (parley.nvim#319) should send a short review
request rather than repeating the review protocol every round. Agents need one
shared skill, and standalone Parley still needs an inline fallback derived from
the same protocol. S1 is parley.nvim#332 plus pair#573.
