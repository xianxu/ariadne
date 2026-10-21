---
id: 000463
status: done
created: 2026-10-19
updated: 2026-10-21
estimate_hours: 3
github_issue:
started: 2026-10-21T04:30:26-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 7.69
tracker:
    version: 1
    completion:
        token: close-7adac0c194ca
        repository: github.com/xianxu/ariadne
        reviewed_head: 31d5e5b22a8629a9aa774c2e3ff924b9f334b24b
        evidence_commit: f0a2ff98887a6a2ee1fdd9d6c54341f358250163
        landed_commit: b1a4c7d5b471363705d45af4a4004d47267467a2
---

# local-check: resume recorded known-flake failures with targeted tests

## Problem

ariadne#461's recorded suite attempt fa09ccf7eb6fd0862fadf1d94e865944 failed
only TestSuiteQueuePriorityRealHelper when its private provider launch returned
Operation not permitted. The exact same commit passed that targeted test without
changes, and the test was registered on ariadne#462. Existing local-check run
can invoke the known-flake step only while processing a fresh full suite;
republish reposts an already-saved failure unchanged. There is no supported
way to finish this failed attempt with just the now-registered targeted tests,
so the hub approved another full suite for #461. This is a #394 follow-up.
