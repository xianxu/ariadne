---
id: 000362
status: working
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: 'a09f6a5fc9fddd0efdc9214ee24f4ee14d67a2ac' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T21:52:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "6069416a", done: "4e15c9e1"}
---

# make test doesn't run construct/scripts/test/*.test.sh (dev-aliases, others), so their regressions are only caught by hand

## Problem

Shell tests under `construct/scripts/test/*.test.sh` (e.g. `dev-aliases.test.sh`, 48 cases added by #355) aren't run by `make test` or CI, so a regression in them is caught only if someone runs them by hand. #357 fixed the same gap for `test-lock.test.sh` by moving it under `scripts/test/`, which CI runs.

## Spec

Run every `construct/scripts/test/*.test.sh` from `make test` (and the CI merge check), the same way `scripts/test/` is run, under bash 3.2-compatible settings. If any are slow or need a live environment, mark them and leave those out explicitly rather than silently.

## Done when

- `make test` runs all `construct/scripts/test/*.test.sh` (or names the excluded ones and why); a deliberately broken test fails `make test`.

## Plan

- [x] One runner, `scripts/test/construct-scripts.test.sh`: runs every
  `construct/scripts/test/*.test.sh` under the bash on PATH, each in its own process,
  names each failure and fails the run; its `EXCLUDED` list (empty today, each entry
  with a reason) is the one place to leave a test out. Living in `scripts/test/`, the
  CI merge check (`20-ci-workflow-tests`) runs it with no change to a local-gates file.
- [x] `make test` (Go only, via test-shard.py over `go list ./...`) runs it through a
  test-only Go package beside it, `scripts/test/scripts_test.go`
  (`TestConstructScripts`, under `testgate.Enter`).
- [x] Proof: a deliberately failing assertion in `dev-aliases.test.sh` fails `make test`.
- [x] Atlas: ci-merge-check.md and base-layer.md say where these tests run.

## Log

### 2026-10-13

### 2026-10-13 — ariadne:3

All six construct tests pass locally in about 13s (dev-aliases 8s, clone-data-deps 3s,
the rest about 1s), so nothing is excluded; CI's Linux runner is the remaining check.
Proof: with `ko "deliberately broken for the #362 proof"` appended to
`dev-aliases.test.sh` (working tree only, restored after), the full `make test` failed
with `--- FAIL scripts/test TestConstructScripts` naming `dev-aliases.test.sh` and the
assertion (`construct-scripts: 5 passed, 1 failed`). That run also hit the #361 weave
load flake, unrelated. The runner itself guards its array expansion per #353's rule
(the shelllint scan reads the index, so it saw the file once staged).
