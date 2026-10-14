---
id: 000353
status: done
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '25db9d575b518b8be465474578ebf7c58e96403f' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T17:18:22-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 1.16
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

- [x] `pkg/shelllint.ArrayFindings`: in a bash script that enables `set -u` (`set -u`,
  `set -euo …`, `set -o nounset`), every `${name[@]}` / `${name[*]}` expansion is a
  finding unless it is a count (`${#a[@]}`), sits inside a guard (`${a[@]+…}`,
  `${a[@]:+…}`), or supplies a default (`-`, `:-`, `=`, `:=`, `?`, `:?`). Default-deny
  with provable exceptions only (#342's lesson): no "is it non-empty" analysis; guarding a
  literal array is free. Double-quoted text is scanned; single-quoted text and comments are
  not (the shared lexer). Table tests for each form.
- [x] Repo test applies it to tracked bash scripts (`.sh`/`.bash`, bash/sh shebangs; not
  zsh, not Makefiles), with the existing exemption map.
- [x] Sweep with `${a[@]+"${a[@]}"}`: introspect-extract.sh (5), docflow.sh (4),
  list-peers.sh and .tart/tart-list-peers.sh (2 each), apply-gitignore-entries.sh,
  rollback.sh, .openshell/overlay/bootstrap.sh, and the local-gates paths
  parallel-checks.sh (4), run-merge-checks.sh (1), pre-merge-checks.sh (1).
- [x] `MAX_PARALLEL_CHECKS` must be a positive integer: refuse otherwise (fail loudly
  rather than clamp a misconfiguration).
- [x] Run each swept script that can run here; `bash -n` the rest.

## Log

### 2026-10-13
- 2026-10-13: closed — Round 1 fixes: scripts/test/parallel-checks.test.sh (run by the 20-ci-workflow-tests merge check) 6/6 green: MAX 3 and 1 complete with no unbound variable (1 drives the #351 pids rebuild), 0/abc/-2 refuse exit 1, unwritable TMPDIR refuses; perl-alarm bounded. Mutations: validation removed fails 3/6 (0 and -2 hit the alarm, abc exits 0); #351 unguarded rebuild restored fails 2/6. Shebang -u detection added with tests; isShell reuses shebangNames; plan Revisions records the measured operator set. go test ./pkg/shelllint green incl. repo scan. Earlier evidence unchanged: rule measured on bash 3.2.57, 21 sites guarded, swept scripts run; full suite x3, each with one unrelated load/env flake filed as #358 (MAKEFLAGS=k, deterministic), #360, #361, each passing alone. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
- 2026-10-13: flow upgraded quick → full — 149 added lines in code files (limit 100); an earlier round of this close already ran the full review
- 2026-10-13: unclaimed: filed from #351's close review; unclaimed for whoever the hub assigns

Filed by ariadne:3 from #351's close review (two advisory Minors, one class).

### 2026-10-13 — survey (ariadne:3)

30 `${name[@]}`/`${name[*]}` occurrences in 13 tracked `set -u` bash scripts; 4 already
use the guarded form (list-peers x2, clone-data-deps, .openshell/sandbox.sh) and 2 are in
comments (run-merge-checks.sh:31, .colima/colima.sh:62). Three files are local-gates
paths, so the PR needs the operator's label.

### 2026-10-13 — implemented (ariadne:3)

Measured on macOS /bin/bash 3.2.57 with an empty array under `set -u`: only the bare
`${a[@]}`/`${a[*]}` and the bash-4 case operators (`^` `,`) fail; `:1`, `:0:1`, `#p`,
`##p`, `%p`, `%%p`, `/a/b`, `//a/b`, `-`, `:-`, `+`, `:+` all succeed. The first cut
flagged every non-default operator; narrowed to the measured set (the plan said "a
default"; the measurement is the contract). `ArrayFindings` reported 21 sites in 9 files
(`.tart/scripts/tart-list-peers.sh` is a symlink to list-peers.sh, scanned once); all
guarded with `${a[@]+"${a[@]}"}` (`[*]` with `:-`), the list-peers substring edit reverted
as unneeded. Guarded forms checked on 3.2: empty gives 0 elements, words with spaces stay
whole, substring keeps its meaning. `MAX_PARALLEL_CHECKS` 0 and `abc` refuse (exit 1);
default and 1 (which drives the wait loop's rebuild) complete, in the stub harness.
Ran: docflow.test.sh ALL PASS, list-peers.sh, `pre-merge-checks.sh nosuchcheck` (the
changed Available line), run-merge-checks.sh over this branch (3/3); `bash -n` for
rollback.sh, apply-gitignore-entries.sh, .openshell/overlay/bootstrap.sh and
introspect-extract.sh (state-changing or need a run dir/LLM).

Close review round 1 (FIX-THEN-SHIP): one Important, three Minors, all fixed.
- Important: the `MAX_PARALLEL_CHECKS` refusal had only a manual check. Added
  `scripts/test/parallel-checks.test.sh` (run in CI by the `20-ci-workflow-tests`
  merge check): the real script against a stub runner; 3 and 1 complete with no
  "unbound variable" (1 drives #351's pids rebuild), 0, abc and -2 refuse, an
  unwritable TMPDIR refuses (#351). Each run is bounded by a perl alarm, so a spinning
  regression fails instead of hanging. Mutations: dropping the validation fails 3 of 6
  (0 and -2 spin to the alarm, abc exits 0); restoring #351's unguarded rebuild fails 2.
- Minor: `isShell` reuses `shebangNames`.
- Minor: nounset is also read from shebang flags (`#!/bin/bash -eu`); a caller's
  `bash -u script` cannot be seen statically.
- Minor: the plan's default-deny wording is superseded; see Revisions.

## Revisions

- 2026-10-13 — Plan row 1 described the rule as "default-deny except count, guard or
  default". Measured on /bin/bash 3.2.57, only the bare `${a[@]}`/`${a[*]}` and the
  bash-4 case operators (`^` `,`) fail for an empty array under `set -u`. The rule flags
  exactly those; substring, pattern-removal and replacement forms are not findings.
