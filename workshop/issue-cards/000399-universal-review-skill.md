---
id: 000399
status: codecomplete
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
actual_hours: 0.51
tracker:
    version: 1
    completion:
        token: close-c7b2ec125264
        repository: github.com/xianxu/ariadne
        reviewed_head: c33651b0a66ff6ad77fd4892325b1c139de64c66
        evidence_commit: 4365719264d16b2fc04bc0cf51d0695c18c9aedf
---

# Universal review skill and canonical round export

## Problem

Universal-agent-input-interface S2a (parley.nvim#319) should send a short review
request rather than repeating the review protocol every round. Agents need one
shared skill, and standalone Parley still needs an inline fallback derived from
the same protocol. S1 is parley.nvim#332 plus pair#573.
