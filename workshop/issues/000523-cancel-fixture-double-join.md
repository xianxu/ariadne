---
id: 000523
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'ab75d9b8dcd0f0ff81ca9b5a9d9ea5325d8ec1be' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T06:19:03-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ] Add the end-of-cleanup wall-time guard and observe it reject the unchanged slow fixture.
- [ ] Close producer-owned completion channels after publishing Wait results; verify native owner/foreign cancellation, failure teardown and adjacent lease controls.
- [ ] Run recorded full suite and prepare close verification.

## Log

- Discovered during independent #521 cancellation validation. Current cmd/sdlc/localcheck_lease_test.go TestLocalCheckCancelOnlyOwnRun start helper sends one result into buffered c.done but never closes it. The main assertion consumes a.done and b.done; each registered cleanup then selects c.done again versus time.After(testfix.HangCeiling). Both results have already been consumed, so each cleanup waits two minutes.
- Targeted current-branch run passed the native control in240.56s (whole held/cancel command256.650s), while no child processes remained during its long final wait. This is deterministic cleanup overhead, not a registered flake or a #521 runtime defect. Native source unchanged by #521. Main verification summary retained at common test-stats/evidence/000521/main-verification-summary.txt; raw pass timing also appears in the worker tool transcript. No nonexistent raw log is cited.

- 2026-10-21: Picked up from fresh main after #521 landed PR #366. Quick plan sent to hub, proceeding unless it objects (679cb81b-dbbf-41e6-bf4d-b709add00734 submitted). No product surface/live stage. Current main still has the unclosed producer. Bounded sweep: newer run-ID helper in localcheck_runid_unix_test.go tracks a finished flag before teardown, so it does not have this double-consumption delay; its other helper closes done. No second affected cancellation fixture identified.

## Revisions

- 2026-10-21: Specify the chosen close-after-send correction and whole-cleanup wall-time bound before implementation; original acceptance preserved.
