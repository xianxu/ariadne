---
id: 000404
status: codecomplete
created: 2026-10-15
updated: 2026-10-16
estimate_hours:
github_issue:
started: 2026-10-16T09:16:33-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.44
tracker:
    version: 1
    completion:
        token: close-1662eeb1b218
        repository: github.com/xianxu/ariadne
        reviewed_head: f845245242b0895b87cdc1e688db15fb3f1750bc
        evidence_commit: ce7b5b2edc0a22a0c6fcc37468f52f92c7f1a119
---

# tl-lead/tl-worker: a third design delta on one mechanism triggers a re-plan; invariants weigh a lock-shaped design first

## Problem

2026-10-15: the TL approved eight design deltas to pair#584, one at a time, each reasonable alone (argv for every OS
process, legacy rows, unbound launches, ledger caps, codec speed, concurrency), without asking why one guard needed
eight. The guard (pair#556) proves "no live writer" by forensics over historical state, about 1,700 lines with ~45
refusal reasons; a wrapper-held flock that dies with its process does the same in one step with two outcomes
(pair#595). The operator: "the same old over engineering." #556's plan never weighed a lock.
