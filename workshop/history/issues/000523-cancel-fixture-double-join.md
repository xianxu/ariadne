---
id: 000523
status: done
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '20a1669722ac6d4e2731fd6fce915416acc082a8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T06:19:03-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "5e2de842", done: "9b4ae3c5"}
actual_hours: 0.34
---

# Native cancellation fixture waits four minutes after successful joins

## Problem

TestLocalCheckCancelOnlyOwnRun spends about four minutes waiting in successful cleanup after both cancellation assertions already joined their native children. This inflates the machine-wide suite wall time without exercising more behavior.

## Spec

Make the fixture child completion join idempotent while retaining actual native-process cancellation, foreign-checkout refusal, PID reuse fencing, bounded teardown and child cleanup on failed assertions. Do not reduce or bump a timeout to hide the wait.

The sole Wait goroutine publishes the existing buffered error result, then closes its owned completion channel. The main assertion still receives the real error; teardown may observe closure after that value was consumed. Keep child termination, PID/lease fencing and cancellation assertions intact. Add a wall-time assertion registered before the child cleanups, so it runs last and includes their entire cost. Its ceiling is one existing testfix.HangCeiling, capped by half the remaining Go deadline when present; no new guessed duration. Old code waits two HangCeilings and must fail the new bound; fixed code must pass the same native fixture promptly.

ARCH-DRY/ORDER: reuse the completion channel as the lifetime signal; no second consumed-result flag or join cache. ARCH-PURE: the only IO change is closing the producer-owned test channel; budget selection is simple duration arithmetic. ARCH-PURPOSE/ERROR: measure through cleanup, preserve the first Wait error, and retain failure teardown and all safety assertions. ARCH-MOCK: existing real native subprocess and stateful Git/lease fixture is the conformance. ARCH-CONSTRAINTS: O(1) added operations per child, no hot-path or timeout increase; derive diagnostic ceiling from the existing fallback and actual supervisor. ARCH-FUNERAL/EVOLVE: no product artifact, new resource, state, format, or operator surface. Channel multi-consumer ordering is semantic; the actual fixture wall-time regression enforces it rather than a lexical rule that would demand unsafe closes on unrelated channels.

## Done when

- Successful test completion does not wait either two-minute HangCeiling after both children were joined.
- Foreign checkout stays alive; owning checkout cancellation stops its managed run and check; failure teardown still joins/reaps children.
- A controlled regression demonstrates repeated cleanup/join observation completes without waiting for the fallback ceiling.

## Plan

- [x] Add the end-of-cleanup wall-time guard and observe it reject the unchanged slow fixture.
- [x] Close producer-owned completion channels after publishing Wait results; verify native owner/foreign cancellation, failure teardown and adjacent lease controls.
- [x] Run recorded full suite and prepare close verification.

## Log


- 2026-10-21: closed — Real native cancellation fixture RED 240.53s above 2m whole-cleanup bound, GREEN 393.97ms with identical safety assertions; all seven targeted lease/cancel controls and pkg/golint passed. Recorded suite e3515f11 passed eight shards,1931 SDLC tests and75 packages,zero failures,707.5s; persistent evidence/000523/branch-suite.log. Test-only channel completion fix has no new architectural surface; hub owns project updates.; review verdict: SHIP
- Discovered during independent #521 cancellation validation. Current cmd/sdlc/localcheck_lease_test.go TestLocalCheckCancelOnlyOwnRun start helper sends one result into buffered c.done but never closes it. The main assertion consumes a.done and b.done; each registered cleanup then selects c.done again versus time.After(testfix.HangCeiling). Both results have already been consumed, so each cleanup waits two minutes.
- Targeted current-branch run passed the native control in240.56s (whole held/cancel command256.650s), while no child processes remained during its long final wait. This is deterministic cleanup overhead, not a registered flake or a #521 runtime defect. Native source unchanged by #521. Main verification summary retained at common test-stats/evidence/000521/main-verification-summary.txt; raw pass timing also appears in the worker tool transcript. No nonexistent raw log is cited.

- 2026-10-21: Picked up from fresh main after #521 landed PR #366. Quick plan sent to hub, proceeding unless it objects (679cb81b-dbbf-41e6-bf4d-b709add00734 submitted). No product surface/live stage. Current main still has the unclosed producer. Bounded sweep: newer run-ID helper in localcheck_runid_unix_test.go tracks a finished flag before teardown, so it does not have this double-consumption delay; its other helper closes done. No second affected cancellation fixture identified.

- 2026-10-21 TDD: unchanged producer plus new whole-cleanup guard failed as intended: native fixture 4m0.531748083s versus 2m budget (package 240.943s). After close-after-send, the same fixture and unchanged safety assertions completed in393.971959ms. All seven targeted lease/cancel controls passed (package3.999s): native foreign/own cancellation, classification, held lease, stale sweep, PID reuse, lease lifetime, and queued sibling run-ID cancellation. Persistent raw logs: /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/evidence/000523/native-cancel-red.log and native-cancel-green-controls.log. No timeout changed; git diff --check passed.

- 2026-10-21: Recorded suite on Git-verified e3515f11af7b9f219f620ff9adaed48e1a0317e6 passed all eight shards (1931 cmd/sdlc tests) and75 other packages, zero failures,707.5s wall after1690s queue. Persistent branch-suite.log in evidence/000523; SHA256 471d75149153c9074ef0e7c8b0724bf9f063cc22340249129d8f06453a0a6ba5. Independent pkg/golint passed0.623s. This removes the measured native fixture overhead; total suite wall is subject to machine load and is not used as a speedup claim.

- 2026-10-21 final: close review SHIP/high with no findings; independently passed all seven controls, confirmed no-close overlay RED240.39s, and forced failures before either join/after first join promptly reaped both runs and checks. Deadline-capped and absent-deadline budgets exercised. Landed PR #369, Git-verified merge2be41ca0ec221ae5bb100e25c06ac3b858ce92ab. Authoritative integrated code b6647051a973388facf94b593d54c912df222305 passed all eight shards (1931 SDLC tests) and75 other packages, zero failures/reruns,661.7s after1079s queue. Persistent integrated-suite.log SHA256 a26a5f64d13ac44363975a1b2763faed70c1169ff295b0a2763d6b20f4f0bbdc. Final issue-record integration changed no code; final CI3green, runner exited0. Actual0.34h measured at close; no quick-flow estimate/calibration point. No stage needed for test-only fixture.

## Revisions

- 2026-10-21: Specify the chosen close-after-send correction and whole-cleanup wall-time bound before implementation; original acceptance preserved.
