---
id: 000442
status: done
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
started: 2026-10-18T08:30:28-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.04
tracker:
    version: 1
    completion:
        token: close-4ea3bb8eabb3
        repository: github.com/xianxu/ariadne
        reviewed_head: f4dee0dcde01fb7d590b79c65ce9c2c641176f2c
        evidence_commit: 8b53118fb0bea36d724331015a6586f09e3cce92
        landed_commit: 1b24cffbc8bdfa8a1b4e3ad59de895d751804f0b
---

# tl-worker: a review may overlap a queued recorded suite without hub approval

## Problem

Slots ask the hub whether a milestone or close review may run while their recorded suite waits in the machine queue (pair:1 asked twice on 2026-10-18; queues ran 20+ minutes). The answer is always yes under two conditions, but the rule "run the full suite before a close" reads as forbidding it, so careful slots wait or ask.
