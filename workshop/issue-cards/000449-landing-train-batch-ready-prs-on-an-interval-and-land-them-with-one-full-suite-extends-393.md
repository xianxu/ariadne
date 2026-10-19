---
id: 000449
status: done
created: 2026-10-18
updated: 2026-10-19
estimate_hours: 5.47
github_issue:
started: 2026-10-19T00:24:20-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 3.92
tracker:
    version: 1
    completion:
        token: close-ab17b31bcf66
        repository: github.com/xianxu/ariadne
        reviewed_head: 74171bfb3611d7c69ee0def1dbb914d7c2b33136
        evidence_commit: 0c7f26b19a42f1c4821c78d500eb8978b9bed649
        landed_commit: 88e38943c26cce06e8177df5e97d2f03a0d703a5
---

# landing train: batch ready PRs on an interval and land them with one full suite (extends #393)

## Problem

The landing line (#393) lands one PR at a time, each with its own full suite on the combined code. On a busy day that is one full suite per landing, and suites queued 20-50 minutes for the two machine slots (2026-10-18). The operator wants the suite frequency to be a knob.
