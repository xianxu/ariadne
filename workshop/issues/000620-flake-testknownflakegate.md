---
id: 000620
status: working
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-25
estimate_hours:
card_mirror: '204c5c3950f79b4d99a40758b49310cfdb386bcf' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-25T00:29:23-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flake_tests: [github.com/xianxu/ariadne/cmd/sdlc.TestKnownFlakeGate]
flow: {kind: quick, provenance: inferred, spec: "ccc7e7b8", done: "a0bbe98b"}
---

# flake: TestKnownFlakeGate

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestKnownFlakeGate: failed in attempt 2a3963a2f77b656a3bec32d24a99ca80, passed on the same head in attempt 0c0df556fe88758c8a8562c1ee90f7ba

## Spec

The BR-16 subtest ("the rerun is bounded by the check's timeout") proved the bound by asserting the whole `runLocalCheck` finished in under 8 s of wall clock. That wall time covers the fixture repo, worktree, preflight, the failing suite command, the rerun and record saving. Under fleet suite load it ran 9.7 to 11.2 s and failed. It failed twice on #603's integration head and on main ce04fa47, so it is a persistent main failure.

Root cause: wall clock as the oracle. The runner already records the deadline event: the rerun's saved attempt carries `TimedOut`. An unbounded rerun would finish its sleep and pass, so the outcome plus `TimedOut` prove the bound without timing.

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [x] Replace the 8 s elapsed assertion with the rerun attempt's `TimedOut` record (the runner's deadline event).
- [x] Give the failing parent command margin: a 5 s test minute, and a 120 s rerun sleep that an unbounded rerun would complete and pass.
- [x] Mutation check: drop the check-timeout bound from `rerunEnvelope`; the test fails (the rerun passes).
- [x] Under load: -count=10 at load 7 to 20, and -count=10 with `nice -n 19` and `-cpu 1`, all green.

## Log

### 2026-10-24

### 2026-10-25
- 2026-10-25: closed — BR-16 now asserts the saved rerun attempt TimedOut instead of an 8s wall bound. Mutation (rerunEnvelope without the check-timeout bound) fails the test. Targeted -count=10 green twice, once under nice -n 19 and -cpu 1 at load 7-20. Full pre-close suite green on 71f220cc (sdlc-local/suite, 19m47s, 0 failed). Test-only fix, no new surface.; review verdict: SHIP

- TL dispatched (main red on BR-16, blocking #603's landing). Fix: assert the runner's `TimedOut` record instead of elapsed time.
- The `test-tight-wallclock` golint rule flags only sub-second bounds; this 8 s bound was a hang-sized bound used as the oracle, which lint cannot tell from a real hang bound. No new rule. Sibling second-scale bounds exist (`ghlanding_test.go:153`, `helpercapability_test.go:167`, `landingattribution_impact_test.go:126`, `tl-lead`, `weave`); reported to the hub, not swept here.
- Verified: mutation caught; -count=10 twice (once niced, -cpu 1) green; full `TestKnownFlakeGate` green; `pkg/golint` green.
