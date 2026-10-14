---
id: 000353
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: 'c010a9b6508124c5f39181996857ee702e25f5c1' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T17:18:22-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# bash 3.2 empty-array expansion under set -u: shelllint rule, sweep, and parallel-checks regression (follow-up to #351)

## Problem

macOS `/bin/bash` is 3.2, where expanding an empty array as `"${a[@]}"` under
`set -u` is an "unbound variable" error. #351 fixed one instance in
`scripts/parallel-checks.sh` (line 108, `pids=("${alive[@]}")`, which killed the
audit whenever every running check finished between polls) with
`${a[@]+"${a[@]}"}`. Its close review found more of the class:

- `construct/local/introspect/scripts/introspect-extract.sh:102`
  `SEGMENTS=("${TRIMMED[@]}")` dies when `--limit 0` leaves `TRIMMED` empty.
- `scripts/parallel-checks.sh:105`: with `MAX_PARALLEL_CHECKS=0` the wait loop spins
  and expands an empty `"${pids[@]}"` (pre-existing).
- The line 108 fix has no automated regression test (verified by a manual stub run).

## Spec

Add a `pkg/shelllint` rule for the class: in a script that runs with `set -u`
(`set -u`, `set -eu…`, `set -o nounset`), an array expansion `"${name[@]}"` that is not
the guarded form `${name[@]+"${name[@]}"}` is a finding. Apply it to tracked scripts
like the mktemp rule, sweep the findings, and decide whether `MAX_PARALLEL_CHECKS<=0`
should be refused or clamped. `scripts/parallel-checks.sh` is a local-gates path: its
part needs the operator's `local-gates-verified` label (split it out if the rest should
land first).

## Done when

- The rule fails on an unguarded `"${a[@]}"` under `set -u` and passes the guarded form.
- No tracked script has a finding (or each has a named, issue-linked exemption).
- `MAX_PARALLEL_CHECKS<=0` has a defined behavior.

## Plan

- [ ]

## Log

### 2026-10-13

Filed by ariadne:3 from #351's close review (two advisory Minors, one class).
