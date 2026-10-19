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
actual_hours: 0.14
tracker:
    version: 1
    completion:
        token: close-958814f322ad
        repository: github.com/xianxu/ariadne
        reviewed_head: c3fca0ece7d29f05d3b014a3030590748cd7de60
        evidence_commit: a6082d5cb19c98b6860d56004bdbe1a41b39deae
---

# weave: honor a repo's CI-workflow opt-out (ops re-seeded after opting out)

## Problem

ops' AGENTS.local.md says "ops opts out of ariadne's seeded CI workflows", yet a weave refresh in ops modified the tracked .github/workflows/merge-check.yml and added .github/workflows/local-gates-guard.yml. On 2026-10-19 the operator removed ops' CI workflow; the next weave refresh would seed it again.
