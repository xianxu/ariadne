---
id: 000440
status: done
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
started: 2026-10-18T05:43:21-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 2.32
tracker:
    version: 1
    completion:
        token: close-e1e96551f22e
        repository: github.com/xianxu/ariadne
        reviewed_head: a35fbd0b408119d01476b230a632d2fff0219fde
        evidence_commit: 18e5ccc97f4796f7666a37a359101b96f764182b
        landed_commit: 846e3aba1e0e5f7b1161ae81c40c03b3cf2c3be6
---

# lint: a wire-message annotation and a rule against strict decoding into wire types

## Problem

Strict JSON decoding of cross-version peer or wire messages (`DisallowUnknownFields`) breaks mixed-version fleets (pair#594), but the same call is right for local files. A syntactic rule can't tell them apart, so #407 deferred it.
