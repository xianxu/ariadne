---
id: 000503
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '5e9db8b2ca171dd8683136a88750559d37029eb3' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T13:10:13-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flake_tests: [github.com/xianxu/ariadne/cmd/sdlc.TestVouchingStartsWhenTrainsTurnOn]
flow: {kind: quick, provenance: inferred, spec: "3da5ff93", done: "f99b6282"}
---

# flake: TestVouchingStartsWhenTrainsTurnOn

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestVouchingStartsWhenTrainsTurnOn: Recorded suite attempt bf7ab6a535a92236feeee763b72f9002 failed this test on exact head 12aff78f6515173afc4aabd5f28ee7f47b4306e4 (0.01s, no vouch after trains turned on); all other shards and packages passed. Targeted go test ./cmd/sdlc -run ^TestVouchingStartsWhenTrainsTurnOn$ -count=1 -v in a detached worktree at that same exact head passed (test 0.01s, package 0.482s). Fixture landingcar_test.go waits for gateReads, incremented inside readiness before the callback returns; it cancels immediately after seeing that counter. vouchWhileWaiting checks ctx.Err after readiness and before line.Vouch, so cancellation can win the publication race. No source change or deadline bump. Fix the fixture to wait for the persisted vouch, with adversarial scheduling coverage.

## Spec

The fixture must cancel only after the waiting car's vouch has been persisted. A readiness callback entry counter is not publication evidence (ARCH-STATE); keep production cancellation before publication intact. Force the callback-in-flight ordering with channels rather than sleeps or longer deadlines. Preserve the trains-off gate-read check and join the vouching goroutine on every exit.

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.
- A controlled in-flight callback reproduces the old missing vouch; the corrected fixture observes persisted readiness before cancellation, and early cancellation publishes nothing.
- The Go lint catalog rejects the callback-counter polling followed immediately by context cancellation, with valid-effect and unrelated-counter controls.

## Plan

- [ ] Force the callback/publication gap and record red evidence.
- [ ] Wait for persisted publication, preserve early-cancellation behavior, and add the counter-cancellation lint rule with fixtures.
- [ ] Run focused and race controls, recorded suite, close review and merge.

## Log

### 2026-10-20

Claimed and started in ariadne:5. Confirmed the readiness callback increments gateReads before returning; vouchWhileWaiting checks ctx.Err before Vouch. Quick plan posted to hub, proceeding unless objected. ARCH-DRY: reuse watchWriter's completed-write signal and Line.Readiness rather than add a production hook. No new state, format, hot path or live-environment behavior; the fixture owns and joins its worker.
