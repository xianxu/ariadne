---
id: 000330
status: working
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: '2908a6d17a9d5b0bdae3d95ba0d05e3c931982a8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-10T23:14:05-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
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

- [x] Baseline: `make test` ×2 on the issue branch (= main + this body), plus the Go packages `make test` never runs.
- [ ] Sandbox failure: replace the three `ps -o stat=` probes (processgroup, judge ×2) with one exec-free helper in `testfix`.
- [ ] Load-sensitive deadline failures: point at #325 (add the new instance to its Log), no duplicate fix here.
- [ ] Coverage gap: `make test` runs only `cmd/sdlc/...`; decide whether it runs `./...`.
- [ ] CI decision recorded; job added or filed.
- [ ] `make test` green ×2 with wall time + shard count in the Log.

## Log

### 2026-10-10
- Claimed from ariadne:3 on the TL's dispatch (ops:0). Load before the run: 7–8 (pair:4 suite holding the machine lock).
- Run A (`make test`, 12 shards over 1047 cmd/sdlc tests + 20 other packages): waited 177s for the lock behind pair:4's
  suite, then wall 661.8s (real 839.6s with the wait). Load peaked ~58 (my own concurrent targeted run of the non-sdlc
  packages added to it). 3 failures, none a regression:
  - `processgroup.TestCancellationKillsDescendants`: `fork/exec /bin/ps: operation not permitted`. Claude Code's
    Seatbelt sandbox denies exec of `ps`, so it fails in every sandboxed agent slot. The two judge
    cancellation tests use the same `ps -o stat=` probe but ignore its error, so under the sandbox their
    "descendant still running" check silently passes. Root fix below (exec-free kernel query).
  - `TestPlanningReviewConcurrencySchedules` (plan + estimate: "unrelated mutation could not finish ... signal:
    killed") — exactly #325.
  - `TestReviewConcurrencySchedules/milestone-close/branch`: "reviewer process did not start" after 22s — same
    load-sensitive deadline family; recorded on #325 rather than fixed here.
- `make test` covers only `go list ./cmd/sdlc/...`. cmd/weave (14 pkgs), pkg/* (6), cmd/datatype, cmd/doc-review,
  cmd/vocabulary have no suite entry point; run separately here.

