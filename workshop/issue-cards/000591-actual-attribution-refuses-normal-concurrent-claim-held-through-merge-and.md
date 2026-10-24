---
id: 000591
status: codecomplete
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
github_issue:
started: 2026-10-24T01:44:54-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: N/A
tracker:
    version: 1
    completion:
        token: close-7610f07cd9c1
        repository: github.com/xianxu/ariadne
        reviewed_head: c701ea2fe03b33401640ba1e203a3e5ab38d26c1
        evidence_commit: a5e86fc38d108c5bff58386f09eb6218ab8c014f
---

# Actual attribution refuses normal concurrent claim held through merge and resting train checkout

## Problem

Follow-up to ariadne#580, reported by ops:0 from pair:4 closing pair#856. Actual now refuses session 062e2184-2117-4c26-b929-ceec3b5a8030 in pair-slot4: concurrent claims have unknown, resting or ambiguous active branch during 2026-10-23T21:35:40-07:00 to 2026-10-24T04:36:48Z. The slot legitimately held pair#850 while merging (its branch, then resting checkout during the train) and pair#856. Earlier close rounds measured 2.72 h and 3.33 h. The hub authorized --no-actual for that close meanwhile.
