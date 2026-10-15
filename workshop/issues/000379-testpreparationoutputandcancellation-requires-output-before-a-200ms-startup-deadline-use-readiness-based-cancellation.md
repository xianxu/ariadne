---
id: 000379
status: open
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
severity: medium
card_mirror: 'c8a282924e55950d33335cd835294c3526f8c0ec' # card fields mirrored from issue-cards; edit via sdlc
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

## Log

### 2026-10-14

Filed unassigned from the primary checkout while #240 remains owned by ariadne:5. Evidence: /tmp/ariadne240-preparation-targeted.log, /tmp/ariadne240-preparation-diagnostic.log and /tmp/ariadne240-preparation-main.log. Full recorded suite failed after 756.0s (1366 SDLC tests, 59 other packages); log /Users/xianxu/workspace/ariadne/.git/sdlc/local-checks/6433ffd3ecd69e3e3b1a0591a4c5002f6f8eaba8/suite.log, SHA-256 prefix d75cb1db7b6a. Independent repair should land from main before #240 reruns its acceptance suite.
