---
id: 000467
status: done
created: 2026-10-19
updated: 2026-10-20
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
actual_hours: 0.35
tracker:
    version: 1
    completion:
        token: close-f8c98aa65d3b
        repository: github.com/xianxu/ariadne
        reviewed_head: d6b87e80bf5e891c05351d9670782417935a1eef
        evidence_commit: ce1f09980ecacc5fd645cf30b06a0bf435b4615f
        landed_commit: c80dc2ed89c654387a96ad649711262a421ff809
---

# weave: honor a repo's CI-workflow opt-out (ops re-seeded after opting out)

## Problem

ops' AGENTS.local.md says "ops opts out of ariadne's seeded CI workflows", yet a weave refresh in ops modified the tracked .github/workflows/merge-check.yml and added .github/workflows/local-gates-guard.yml. On 2026-10-19 the operator removed ops' CI workflow; the next weave refresh would seed it again.
