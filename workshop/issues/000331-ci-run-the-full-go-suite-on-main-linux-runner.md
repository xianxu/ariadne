---
id: 000331
status: open
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: '181d82c8c51370901254737e2c0dec608f6d682c' # card fields mirrored from issue-cards; edit via sdlc
---

# CI: run the full Go suite on main (Linux runner)

## Problem

ariadne has no test CI: `merge-check.yml` runs only `scripts/merge-checks.d/` (weave drift, duplicate issue IDs) on
PRs, so nothing shows whether main's full suite passes. #330 made `make test` cover the whole module and green
locally, and decided main should get a CI test job. It was filed rather than added there because the suite has never
run on Linux, and its cost on a hosted runner is unknown.

## Spec

- A GitHub Actions job runs `make test` on pushes to main (and optionally on PRs), on `ubuntu-latest`.
- First make the suite pass on Linux: the darwin-specific paths (`testfix.PreferRealGit`'s xcrun-shim workaround,
  `testfix.AwaitExit`'s `/proc` branch, Seatbelt-shaped assumptions) have only been exercised on macOS.
- Size it: #330 measured 12 shards on the local Mac (wall in #330's Log). A 4-core hosted runner gets ~4 shards; if
  it runs much past 30 minutes, decide between a larger runner, a nightly schedule, or a lighter subset on PRs.
- The machine-wide lock (#319) and test gate (#324) are no-ops on a fresh runner; confirm they don't warn or wait.

## Done when

- `make test` passes on a Linux runner, with its wall time recorded.
- main's CI runs it on every push, and a red run is visible (badge or required check).

## Plan

- [ ]

## Log

### 2026-10-10
