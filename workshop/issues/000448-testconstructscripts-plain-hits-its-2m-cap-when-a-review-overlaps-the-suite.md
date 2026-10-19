---
id: 000448
status: working
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
severity: low
collaboration-mode: collaboration
card_mirror: 'c74bdbb2023227a5d8a344795ee4acc9f655b4ac' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T19:54:59-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "19ed9822", done: "828e3cbf"}
---

# TestConstructScripts/plain hits its 2m cap when a review overlaps the suite

## Problem

`TestConstructScripts/plain` (`scripts/test/scripts_test.go`) runs all of `construct/scripts/test` under one `runnerDeadline` of 2m per environment. Under load the plain subtest hits that cap and fails, unrelated to the change being tested.

Evidence (ariadne:2, #441, 2026-10-18):
- The recorded suite on 208ed53a (`sdlc local-check run --name suite`) failed with `construct/scripts/test did not finish within 2m0s`. The last script reported ok was `bootstrap-transitive.test.sh`. The whole run took 14m7s against a usual 8-12m, because a `sdlc milestone-close` boundary review ran alongside it, which the #442 overlap rule permits. The failed log is kept at `/Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/dd86d2996958c8e19e84dd8b9c49ee61.log`.
- The branch touched nothing under `construct/scripts` or `scripts/test` (only prose in `construct/local/fix/SKILL.md`).
- On the same tree, `go test ./scripts/test/ -run TestConstructScripts -count=1` alone passed in 230.9s for three environments, about 77s each against the 120s cap, so the headroom is only about 1.5x.
- The next two recorded suites (b7c033ab, 389adfdc), run with nothing overlapping, passed.

## Spec

The canonical shell runner owns selection and exclusions. Add a NUL-delimited
`--list` mode that lists selected paths without executing them. The Go harness
uses that selection and runs each script as a named subtest, creating a fresh
unchanged `testgate.HangCeiling` (two minutes) for each command. Preserve process
group cancellation and pipe-drain bounds; preserve the default CI batch runner.
No timeout increases and no changes to the scripts under test.

ARCH-DRY: selection remains in one runner. ARCH-PURE: selection is distinct from
execution. ARCH-PURPOSE: a completed script no longer spends the next script's
hang budget. ARCH-CONSTRAINTS: six sequential scripts across three environments,
with shared-machine load; bound a hung script, not cumulative healthy progress.
ARCH-MOCK/SECURE: private fixture scripts, no external service or credentials.
ARCH-ORDER: no durable state between invocations; cancellation owns one process
group until it exits. ARCH-FUNERAL: fixture files use test-owned temporary dirs;
children are killed on cancellation and output draining stays bounded.

Alternatives: increasing the aggregate cap violates the dispatch and merely
moves the load cliff; implementing a shell watchdog duplicates Go's existing
process-group cancellation. Use the existing Go command seam per selected path.

## Done when

- Each selected script in each environment receives its own unchanged two-minute
  hang bound, preserving cancellation of its process group. A recorded suite
  under the machine's normal concurrent workload passes and records timings.
- Go and CI share the canonical selection and exclusions. Fixture regressions
  cover listing without execution, unusual path names, exclusions, empty input,
  and default runner success/failure behavior.
- `pkg/golint` rejects Go tests executing the known construct batch runner except
  for listing, citing this issue. Tests demonstrate the original call is caught.

## Plan

- [ ] Add regressions for canonical listing and the original aggregate call;
  run red before implementation.
- [ ] Add listing and per-script Go subtests; retain the existing per-command
  ceiling and process-group cleanup. Run targeted regressions and the catalog.
- [ ] Run a recorded full suite, close, open PR and land through the normal gates.

## Log

### 2026-10-18

- Claimed and started in ariadne:3. Hub dispatch: diagnose first, no timeout bumps.
- Read both retained failure logs: the first reports bootstrap and clone passing
  before timing out; the #452 ancestor reports bootstrap, clone, dev-aliases and
  portable passing. The original Problem's first last-success summary was
  incomplete. No one script consistently hangs.
- Unchanged individual plain-environment measurements (seconds): bootstrap
  0.746; clone 8.129; dev-aliases 19.205; portable 0.758; sandbox 0.814; vm-log
  0.042. All passed. These are a fresh baseline, not a loaded-suite guarantee.
- Posted the quick-flow plan to the hub, proceeding unless it objects. A bounded
  independent read-only diagnosis confirmed cumulative scope and recommended
  listing plus per-script execution. The narrow lint protects this known batch
  entrypoint; it does not claim to infer arbitrary shell loops from Go syntax.

## Revisions

### 2026-10-18 — dispatch constraint and diagnosed scope

Removed the cap-increase option from Done when: the hub explicitly requires no
timeout bumps. Preserve two minutes for each independent script, and verify
canonical selection plus guard the original aggregate invocation.
