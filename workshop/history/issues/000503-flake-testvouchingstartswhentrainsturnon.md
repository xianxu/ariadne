---
id: 000503
status: done
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '3d8e5e62f14fb54a7c5cb7df360f166532d1a99c' # card fields mirrored from issue-cards; edit via sdlc
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
actual_hours: 0.05
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

## Plan

- [x] Force the callback/publication gap and record red evidence.
- [x] Wait for persisted publication and preserve early-cancellation behavior.
- [x] Run focused and race controls and prepare the recorded-suite and close requests; require their passing results before landing.

## Log

### 2026-10-20
- 2026-10-20: closed — Deterministic red: callback held after its entry counter, old observer cancelled and left empty readiness. Fixture now waits for the existing post-Vouch event and verifies persisted readiness before cancellation; separate in-flight cancellation control preserves no publication. Focused vouch/wait-loop/schedule checks count=25 passed (11.757s); fixture count=100 under -race passed (4.807s). No production or timeout change, so no new atlas surface. No second callback-counter/cancel instance found; TL requested no lint for one fixture. Recorded pushed-head suite requested under authorized overlap; pass required before landing.; review verdict: SHIP

Claimed and started in ariadne:5. Confirmed the readiness callback increments gateReads before returning; vouchWhileWaiting checks ctx.Err before Vouch. Quick plan posted to hub, proceeding unless objected. ARCH-DRY: reuse watchWriter's completed-write signal and Line.Readiness rather than add a production hook. No new state, format, hot path or live-environment behavior; the fixture owns and joins its worker.

Red: held readiness on a channel after incrementing gateReads, retained the old observer's cancel, then released the callback. TestVouchingStartsWhenTrainsTurnOn failed deterministically with empty readiness (0.01s, package 0.421s). Green: fixture now holds the callback, verifies no publication in flight, releases it, waits for the existing post-Vouch writer event, verifies the persisted readiness, then cancels and joins. A second ordering cancels during the callback and verifies no vouch. Focused vouch/wait-loop/shared-schedule checks passed at count=25 (11.757s); repaired fixture passed count=100 under -race (4.807s). No production change or timeout increase. Preparing pushed-head recorded suite and binary-owned close review; recorded suite must pass before landing.

## Revisions

### 2026-10-20T13:12-07:00 — lint only for a repeated pattern

TL approved the fixture/control plan and required another instance before adding lint (Couch c529c41d-b5ad-4a34-a254-54002f8d7baf). Repository-wide Load-in-loop/cancel searches found only this test; inspection of all ten test files with .Add(1) found WaitGroups, joined build-count assertions and lock controls using separate events, with no second callback-entry/cancel pattern. Removed the original planned lint acceptance and step; deterministic regression is sufficient for this single fixture. No lint machinery or timeout change.

The verification checklist names preparing the suite/review requests rather than promising to complete close/merge before the close gate. Their results remain required before landing, under tl-worker's authorized review/suite overlap.
