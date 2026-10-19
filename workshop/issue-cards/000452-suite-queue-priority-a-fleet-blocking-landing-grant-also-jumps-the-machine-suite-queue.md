---
id: 000452
status: codecomplete
created: 2026-10-18
updated: 2026-10-18
estimate_hours: 1.63
github_issue:
started: 2026-10-18T18:04:52-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.05
tracker:
    version: 1
    completion:
        token: close-cb9f91795cc6
        repository: github.com/xianxu/ariadne
        reviewed_head: 0e36a07b0b8ff23f565224549d685ef68a7c4e36
        evidence_commit: 8156ff5f41b631813a166c68af31115b94e39932
---

# suite queue priority: a fleet-blocking landing grant also jumps the machine suite queue

## Problem

A fleet-blocking landing grant (sdlc landing grant) orders the merge line but not the machine-wide suite queue (WF_TEST_SUITES, two slots). On 2026-10-18 the urgent pair#672 and pair#676 each waited behind ordinary suites for a machine slot even with the grant; the weekly snapshot shows a median queue wait of 11 to 20 minutes and a 90th percentile over 30 minutes.
