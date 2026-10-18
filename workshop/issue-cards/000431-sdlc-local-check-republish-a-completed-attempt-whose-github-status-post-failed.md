---
id: 000431
status: codecomplete
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
github_issue:
started: 2026-10-17T18:23:15-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.89
tracker:
    version: 1
    completion:
        token: close-25cd82393a69
        repository: github.com/xianxu/ariadne
        reviewed_head: 391c17e163125b715265c3952ba0c1e1c14722c4
        evidence_commit: d5a0c0b7c98fc9d7ea834e2ae4aae9d398bb9a1c
---

# sdlc local-check: republish a completed attempt whose GitHub status post failed

## Problem

A completed `sdlc local-check run --name suite` whose GitHub status POST fails (network error to api.github.com after the run) leaves the attempt record and log intact locally, but no `sdlc-local/suite` status on the commit. There is no verb that publishes only that saved result, so the slot must rerun the whole suite (12-15 min, and a machine suite slot) or hand-post the status with `gh api`, which bypasses the gate. Hit twice on 2026-10-17 within minutes: pair:7 (pair#621, attempt a1f1f773, integrated head 055cb82b) and pair:5 (pair#651, head 80063e6d).
