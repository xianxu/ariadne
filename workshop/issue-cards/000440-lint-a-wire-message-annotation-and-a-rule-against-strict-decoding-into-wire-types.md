---
id: 000440
status: working
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
---

# lint: a wire-message annotation and a rule against strict decoding into wire types

## Problem

Strict JSON decoding of cross-version peer or wire messages (`DisallowUnknownFields`) breaks mixed-version fleets (pair#594), but the same call is right for local files. A syntactic rule can't tell them apart, so #407 deferred it.
