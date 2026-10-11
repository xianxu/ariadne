---
id: 000330
status: open
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: '1d64dd301d479d42f22f41298e21610b32832ade' # card fields mirrored from issue-cards; edit via sdlc
---

# Main is green: the full suite passes on ariadne main

## Problem

ariadne has no CI on main, so nothing shows whether the full suite passes there. During batch 2 (10-10) workers judged
failures by "it also fails on main", and full suites flaked under load (ops kink 53). The operator (batch-2 recap,
10-10): "let's ensure ariadne and pair are green on main".

## Spec

- Run the full suite (`make test`) on a clean main; list every failure, every flake (run it twice), and the slowest
  packages.
- Each failure or flake: fix it at the root, or quarantine it by name with an owning issue (the exception, named in
  the Log). Overlaps the test-infra train (ariadne#325 deadline tests, #302 tempdir cleanup race): fold or point,
  don't duplicate.
- Record the suite's wall time and shard count in the Log (input for the shard-count question from the recap).
- Decide whether ariadne gets a CI job on main; if yes, add it (or file it).

## Done when

- [ ] `make test` passes on main from a clean slot, twice in a row, with wall time recorded.
- [ ] Every quarantined test has an owning issue.
- [ ] The CI decision is recorded, and the job exists or is filed.

## Plan

- [ ]

## Log

### 2026-10-10
