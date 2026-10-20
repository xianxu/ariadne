---
id: 000491
status: codecomplete
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
severity: low
started: 2026-10-20T06:03:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.22
tracker:
    version: 1
    completion:
        token: close-0c2e89e5571f
        repository: github.com/xianxu/ariadne
        reviewed_head: 0b60e8b7e74e9ce5fff56a4d8b5963515dd6f228
        evidence_commit: b173aac4db037f435d4f4ec81a6f4f7baadbf72d
---

# Suite queue position becomes fictitious on allocation contention

## Problem

Observed while #481's recorded suite waited under real allocation contention: reports alternate between the actual position 3 (ahead contains the train and another waiter) and position 1 with the same next-in-line holder. No suite admission occurred. The new nonblocking secondary-lock callback returns its default empty ahead list when allocation cannot be inspected; acquire interprets len(ahead)+1 as an observed position. This produces false position updates and log churn.
