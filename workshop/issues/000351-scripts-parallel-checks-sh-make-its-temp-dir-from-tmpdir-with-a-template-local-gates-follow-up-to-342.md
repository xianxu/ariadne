---
id: 000351
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '75f3e2cda2983b430fd3e4d277b2e60a19876e87' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T15:49:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# scripts/parallel-checks.sh: make its temp dir from TMPDIR with a template (local-gates follow-up to #342)

## Problem

`scripts/parallel-checks.sh:144` makes its output dir with a template-less
`OUTDIR=$(mktemp -d)`. On macOS that ignores `TMPDIR` and creates under
`/var/folders/.../T`, which the Claude Code sandbox denies (#342). The script is a
local-gates path (`local-gates-guard` protects `scripts/parallel-checks.sh`), so #342
could not change it: a local-gates change is its own issue, verified by the operator.

## Spec

Change line 144 to `OUTDIR=$(mktemp -d "${TMPDIR:-/tmp}/parallel-checks.XXXXXX") || exit 1`
(or the script's existing failure path), and remove the `scripts/parallel-checks.sh`
exemption from #342's template-less `mktemp` check.

## Done when

- `scripts/parallel-checks.sh` uses a TMPDIR-rooted template, and #342's check has no
  exemption left.
- The operator has verified the local-gates change (`local-gates-verified`).

## Plan

- [ ] Edit line 144; delete the exemption row in the #342 check; run the check.

## Log

### 2026-10-13

Filed by ariadne:3 while working #342, which exempts this one path until this lands.
