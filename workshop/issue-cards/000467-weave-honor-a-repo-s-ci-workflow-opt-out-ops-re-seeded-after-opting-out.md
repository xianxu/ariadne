---
id: 000467
status: codecomplete
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
started: 2026-10-19T10:00:05-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.13
tracker:
    version: 1
    completion:
        token: close-1ecff54e66ab
        repository: github.com/xianxu/ariadne
        reviewed_head: e4ee6277a482202da97576459c5847c1de49eaa3
        evidence_commit: 58e6f42f3987bda12b040ad6a3435b23e2f96ffc
---

# weave: honor a repo's CI-workflow opt-out (ops re-seeded after opting out)

## Problem

ops' AGENTS.local.md says "ops opts out of ariadne's seeded CI workflows", yet a weave refresh in ops modified the tracked .github/workflows/merge-check.yml and added .github/workflows/local-gates-guard.yml. On 2026-10-19 the operator removed ops' CI workflow; the next weave refresh would seed it again.
