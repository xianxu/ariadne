---
id: 000509
status: codecomplete
created: 2026-10-20
updated: 2026-10-21
estimate_hours: 3.41
github_issue:
started: 2026-10-20T22:27:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 2.30
tracker:
    version: 1
    completion:
        token: close-38c2b5125ef6
        repository: github.com/xianxu/ariadne
        reviewed_head: 8ed79fac33b27fd91dda3182514bf06e9fc54d1d
        evidence_commit: 031a357dd41679805234f21db4d197f108265da0
---

# Landing scoped local checks

## Problem

pair#743 moves native conformance from capacity-constrained GitHub macOS runners to a machine-wide Tart runner. The operator requires one conditional conformance run per landing train, after its suite, with nightly burn-in before activation. Current required local checks also gate each car's readiness; declaring conformance required would run it per car. The schema has no scope/path selection and every local run acquires a full-suite slot.
