---
id: 000441
status: codecomplete
created: 2026-10-18
updated: 2026-10-18
estimate_hours: 0.735
github_issue:
started: 2026-10-18T09:10:46-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.89
tracker:
    version: 1
    completion:
        token: close-437fda1661c0
        repository: github.com/xianxu/ariadne
        reviewed_head: bfe6b7b53ab4e5b9bbf0b8b600105bdfa03fbda7
        evidence_commit: 84f1c69b848722d3d49d784bc8b75e3f724bc775
---

# doc-review: choose its reviewer through the #409 family-aware resolver

## Problem

`cmd/doc-review` hard-codes its reviewer (`DefaultAgent = AgentCodex`, "a different vendor from the usual co-authoring agent"). #409 gives sdlc one family-aware reviewer resolver (cross-family by default, visible fallback); doc-review does not use it, so a codex-authored document is reviewed by codex, and an unavailable codex has no visible fallback.
