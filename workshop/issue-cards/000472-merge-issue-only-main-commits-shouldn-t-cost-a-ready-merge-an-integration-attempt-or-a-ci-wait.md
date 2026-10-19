---
id: 000472
status: codecomplete
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
started: 2026-10-19T15:37:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.97
tracker:
    version: 1
    completion:
        token: close-39585fd46b21
        repository: github.com/xianxu/ariadne
        reviewed_head: 2b74d047b6d241597e30adde4819a5a29f50f704
        evidence_commit: 06fadf2e8c5d423531ed94e95e5a6742ed28a79a
---

# merge: issue-only main commits shouldn't cost a ready merge an integration attempt or a CI wait

## Problem

2026-10-19, pair#705 (pair:7): its integrated suite passed and CI passed, then two **issue-only** commits reached main (a
`sdlc issue new` publish and an `abandon` archive, touching only `workshop/`). Each forced `sdlc merge` to integrate main
again and wait for fresh CI, using up attempt 3 of 3. On a busy repo, issue filing alone can starve a ready merge, and the
hub had to ask every slot to stop publishing.
