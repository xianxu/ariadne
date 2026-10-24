---
id: 000590
status: done
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
github_issue:
started: 2026-10-24T04:36:31-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.20
tracker:
    version: 1
    completion:
        token: close-b461858a9399
        repository: github.com/xianxu/ariadne
        reviewed_head: aa3c9a767075f7d4a6a3a290c9988116bdafa6fc
        evidence_commit: d49451e9d4084e6d0071aa49e7a68a7e2ab9b169
        landed_commit: fb48cafb2bf86942fc31a4266624c47633bbe3e5
---

# Qualify recursive JSON lint against independent oracles and DOM materializers

## Problem

Ariadne pkg/golint json-recursive-scanner reports proven false positives in Pair while adopting 5a17f789619d for pair#838. The matcher treats recursive Token + More + any string-key map as duplicate-validator authority. It flags both an independent differential-test oracle and a standalone DOM materializer. Its advice names cmd/sdlc/internal/jsonstrict, which downstream Go modules cannot import.
