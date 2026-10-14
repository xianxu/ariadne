---
id: 000351
status: codecomplete
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '8ce15f448832635ca014da0fe5f658f8287bff30' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T15:49:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.45
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

- [x] Line 144: `mktemp -d "${TMPDIR:-/tmp}/parallel-checks.XXXXXX"` with an explicit
  failure path; the cleanup trap expands `"${OUTDIR:?}"`.
- [x] Remove the `scripts/parallel-checks.sh` exemption from `pkg/shelllint` (the map
  stays, empty, for the next rule that lands before a local-gates fix).
- [x] Folded in (same file, same operator verification): line 108's
  `pids=("${alive[@]}")` dies under bash 3.2's `set -u` when every running check
  finishes between polls; now `${alive[@]+"${alive[@]}"}`.

## Log

### 2026-10-13
- 2026-10-13: closed — Full make test green (12 shards, 1136 cmd/sdlc tests, 47 packages, 0 failed). pkg/shelllint green with an empty exemption map (scripts/parallel-checks.sh now scanned). Failure path: TMPDIR=/nonexistent-342 parallel-checks.sh --audit prints Cannot make a temp dir and exits 1, writing nothing. Success path in a scratch repo with a stub pre-merge-checks.sh (real checks launch LLM agents): first died at line 108 alive[@] unbound variable (macOS bash 3.2.57, set -u), fixed with ${alive[@]+...}; then completes, exit 0, no parallel-checks.* left in TMPDIR. No atlas surface: sandbox.md already describes the exemption mechanism, which stays (empty).; review verdict: SHIP
- 2026-10-13: unclaimed: filed from #342; local-gates change, operator verifies; unclaimed for whoever the hub assigns

Filed by ariadne:3 while working #342, which exempts this one path until this lands.

### 2026-10-13 — ariadne:3

Verified: `go test ./pkg/shelllint` green with no exemption. `TMPDIR=/nonexistent-342
scripts/parallel-checks.sh --audit` prints "Cannot make a temp dir under
/nonexistent-342" and exits 1, writing nothing. The success path ran in a scratch repo
with a stub `pre-merge-checks.sh` (the real checks launch LLM agents, `specs` with write
tools): it first stopped at line 108 with `alive[@]: unbound variable` (macOS /bin/bash
3.2.57, empty array under `set -u`, a pre-existing bug the instant stubs make certain);
fixed, it runs to completion, exit 0, and leaves no `parallel-checks.*` dir in TMPDIR.

Landing waits on the operator's `local-gates-verified` label on the PR (this is a
local-gates path); that is the merge gate, not a Plan step.

Close review (SHIP, round 1) left two advisory Minors, one class: bash 3.2 empty-array
expansion under `set -u` (a sibling at `introspect-extract.sh:102`; no automated
regression for the line 108 fix; `MAX_PARALLEL_CHECKS=0` at line 105). Filed as #353
(unclaimed): a `pkg/shelllint` rule for the class plus its sweep. That is new design
beyond this local-gates PR, whose scope the operator verifies.
