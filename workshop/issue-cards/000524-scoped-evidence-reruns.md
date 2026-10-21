---
id: 000524
status: codecomplete
created: 2026-10-21
updated: 2026-10-21
estimate_hours: 2.82
github_issue:
started: 2026-10-21T04:42:56-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.79
tracker:
    version: 1
    completion:
        token: close-f093925b6750
        repository: github.com/xianxu/ariadne
        reviewed_head: 30b4964360436b238aadaed6ce5a7f261622952e
        evidence_commit: aa9e33e0c1b9afefc480805d2524c30dd929a598
---

# Scoped future evidence and exact-package flake reruns

## Problem

Prerequisite for ariadne#463, approved by hub receipt917bf022-032a-4d84-ac6b-39075bbca2b5. Today's attempt reader treats an unsupported record as a global warning; guardSuitePass then refuses unrelated successes and saved-pass loading marks the whole store incomplete. A future ResumeLink cannot retrofit stale binaries. Existing known-flake reruns also use one name-pattern cross-product across selected packages.
