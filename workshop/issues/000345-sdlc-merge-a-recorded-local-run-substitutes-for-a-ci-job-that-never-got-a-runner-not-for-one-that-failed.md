---
id: 000345
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '5acde55f59388e488f947e803ad994efe3cbb71b' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc merge: a recorded local run substitutes for a CI job that never got a runner (not for one that failed)

## Problem

`sdlc merge` (#309) refuses while a required CI check is red or pending. On 2026-10-13 pair#523 sat for over an hour because GitHub's macOS arm64 runners were short: pair's `couch-zellij-conformance` job was 'not acquired by a hosted runner' and cancelled after 15 minutes with zero steps run. That job is the only routine run of pair's live Couch/Zellij tests (`make -f Makefile.local test-couch-zellij-live`, gated by PAIR_LIVE_COUCH=1, and `test-native-terminal-ci`): local `make test` skips them. Every slot runs on the same Mac, so the same tests can run locally.

## Spec

The operator (2026-10-13): a CI job that **could not get a runner** (no steps executed; cancelled/queued past a timeout) may be substituted by a **recorded local run** of the repo-declared equivalent target; a job that **ran and failed** still blocks.
1. A repo declares, per CI job, its local equivalent (e.g. pair: `couch-zellij-conformance` → `make -f Makefile.local test-couch-zellij-live test-native-terminal-ci`).
2. `sdlc merge` distinguishes 'no runner / zero steps' from 'failed' using the check run's data; for the former it offers the substitute: run the local target, record its exit and log as evidence on the PR (a comment or check), then proceed.
3. Pending-but-running jobs still wait; failed jobs still refuse; `--no-validate` stays forbidden.
4. pair's docs (AGENTS/README) say which local target substitutes for the macOS job, and that it is also worth running before close when a change touches Couch's session/Zellij code.

## Done when

- A merge whose required macOS job never got a runner proceeds after a recorded passing local run of the declared target; a failed job still blocks; tests cover both; pair declares its mapping.

## Plan

- [ ]

## Log

### 2026-10-13

## Revisions

- 2026-10-13 — the operator's stance broadens this: for a small company, **local runs are the primary verification**
  (faster, more controlled iteration on the operator's Mac); **CI mainly guards that what agents aren't supposed to change
  hasn't changed** (#309's local-gates-guard, merge-check's integrity checks). So design the recorded local run as first-class
  evidence for heavy runtime suites, not only as a fallback, and keep CI as the tamper guard. Which jobs stay required on CI
  is the operator's call per repo.
