---
id: 000379
status: codecomplete
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
severity: medium
card_mirror: 'cf42f3bcf00ae6dfdd6f5adb342876bcb97f8bdb' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T17:08:54-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "ee32aab6", done: "4e0a0b91"}
actual_hours: 0.35
---

# TestPreparationOutputAndCancellation requires output before a 200ms startup deadline; use readiness-based cancellation

## Problem

The recorded #240 suite on 6433ffd3ecd69e3e3b1a0591a4c5002f6f8eaba8 failed only TestPreparationOutputAndCancellation (cmd/sdlc/internal/judge/preparation_test.go:79): output bound: len=0 truncated=false. All 12 root SDLC shards passed; the judge package failed. The test starts a 200ms context before Dispatch performs confinement, tracked-state Git inspection and process startup, then requires output truncation from a yes producer. The deadline can legitimately expire before the producer captures any output.

Focused count=3 reproduced the assertion once; a diagnostic-only Go overlay reproduced it twice in five attempts with dispatch_err and preparation_err both review preparation attempt 1: context deadline exceeded. Main 74e803ecd896400bbe2e7f4b957c69383e9f1397 has identical judge/reviewprepare/processgroup sources, and three isolated main retries passed. This is a scheduling-dependent test expectation, not a #240 publication-code change.

## Spec

Prove output retention and cancellation without requiring OS/Git/process startup to finish within 200ms. Use observable producer readiness before cancellation, or separate finite oversized-output coverage from cancellation coverage. Keep bounded shutdown and descendant retirement checks. An expiry before producer startup is valid cancellation and must not be required to have truncated output. Reuse the readiness approach documented with #369 rather than extending arbitrary timing margins.

## Done when

- Oversized preparation output is captured and truncated to the declared budget through the production dispatch path.
- Cancellation is exercised after observable preparation readiness, remains bounded, and never dispatches the reviewer.
- A deterministic delayed-start or early-expiry control distinguishes valid pre-start cancellation from output truncation.
- Focused preparation tests and the recorded full suite pass under the machine gate.

## Plan

- [x] Make the preparation subprocess write beyond `OutputLimit` and signal readiness before test cancellation; bound only the wait for readiness. Assert the cancellation error, no reviewer call, and capped/truncated evidence.
- [x] Add an already-canceled preparation control that accepts empty, untruncated output and still refuses to dispatch the reviewer.
- [x] Run focused repetitions and the recorded suite, then check the patch before close.

## Log

### 2026-10-14
- 2026-10-14: closed — The preparation test waits for a producer marker after writing 64x OutputLimit before canceling; an already-canceled control accepts empty evidence. Focused pair passed 20 repetitions, whole judge package passed, make -k tools passed, and recorded sdlc-local/suite passed on 6bb8b055 (12 shards, 1353 sdlc tests, 59 packages). No atlas surface: test-only change. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

Filed unassigned from the primary checkout while #240 remains owned by ariadne:5. Evidence: /tmp/ariadne240-preparation-targeted.log, /tmp/ariadne240-preparation-diagnostic.log and /tmp/ariadne240-preparation-main.log. Full recorded suite failed after 756.0s (1366 SDLC tests, 59 other packages); log /Users/xianxu/workspace/ariadne/.git/sdlc/local-checks/6433ffd3ecd69e3e3b1a0591a4c5002f6f8eaba8/suite.log, SHA-256 prefix d75cb1db7b6a. Independent repair should land from main before #240 reruns its acceptance suite.

Claimed in ariadne:1. Design keeps production untouched (ARCH-PURE, ARCH-PURPOSE): an explicit signal from the oversized producer, rather than elapsed wall time, establishes that truncation has become observable before cancellation. A pre-canceled control documents that early expiry need not retain output.

The focused pair passed 20 repetitions; the whole `judge` package passed once. The new producer writes 64 times the output cap before marking readiness and waits to be canceled. The parent uses a bounded marker wait only as a setup failure guard; cancellation is event-driven. `git diff --check` passed.

`make -k tools` passed. The recorded `sdlc local-check run --name suite` passed on `6bb8b055773d6ae49e4078d04a05bf3bb4751e8a`: all 12 shards (1,353 `cmd/sdlc` tests) and 59 other packages, 0 failures. The next delivery steps are the close review, PR, landing, and notification to ariadne:5 and the hub.

## Revisions

- 2026-10-14 — Split delivery from the final Plan checkbox: the close gate requires all implementation and verification items checked before it runs, while close/land/notification necessarily follow that gate. Scope and acceptance are unchanged.
