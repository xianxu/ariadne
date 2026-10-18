---
id: 000441
status: working
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
started: 2026-10-18T09:10:46-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# doc-review: choose its reviewer through the #409 family-aware resolver

## Problem

`cmd/doc-review` hard-codes its reviewer (`DefaultAgent = AgentCodex`, "a different vendor from the usual co-authoring agent"). #409 gives sdlc one family-aware reviewer resolver (cross-family by default, visible fallback); doc-review does not use it, so a codex-authored document is reviewed by codex, and an unavailable codex has no visible fallback.
