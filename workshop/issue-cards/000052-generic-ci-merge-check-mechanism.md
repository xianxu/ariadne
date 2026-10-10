---
id: 000052
status: working
created: 2026-05-31
updated: 2026-10-09
estimate_hours: 3
started: 2026-10-09T18:42:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# Generic CI merge-check mechanism (pluggable publish gate for derivatives)

## Problem

you-decide needs a server-side gate: a PR to `main` must not merge unless its
shared-substrate files are `review: passed` (you-decide#4 M3). But the *mechanism*
— "run a repo-defined gate over the PR's range in CI" — is generic: every ariadne
derivative wants a place to plug its own merge-time checks (a license check, `go
test`, a lint, a substrate gate, or nothing). Build the generic mechanism in the
base layer first; you-decide then plugs in its `review-gate.sh`.
