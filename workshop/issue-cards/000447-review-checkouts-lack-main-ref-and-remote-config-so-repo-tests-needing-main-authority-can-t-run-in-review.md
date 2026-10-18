---
id: 000447
status: codecomplete
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
started: 2026-10-18T13:57:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.06
tracker:
    version: 1
    completion:
        token: close-735e9ea9ead1
        repository: github.com/xianxu/ariadne
        reviewed_head: 126eb88b52d3a8b4b671895272ba6b2214d418e1
        evidence_commit: 93d0ea60d3865c499041e4fa2645ae3d61b38dc5
---

# review checkouts lack main ref and remote config, so repo tests needing main authority can't run in review

## Problem

The boundary review runs in a disposable checkout that lacks the main branch ref, its upstream and remote configuration. Repository tests that need main authority (pair's Go lint catalog finds tracked files and the merge base against main) correctly refuse there, so the reviewer cannot run them and raises an "unverifiable" finding. It hit pair#665 M1 (BR-4) and pair#666 (BR-1) on 2026-10-18; each slot had to capture host evidence and dispute it.
