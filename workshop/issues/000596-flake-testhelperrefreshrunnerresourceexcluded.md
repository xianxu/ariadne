---
id: 000596
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '1b752786bba2c7ae14952171244821739c9bf8e5' # card fields mirrored from issue-cards; edit via sdlc
flake_tests: [github.com/xianxu/ariadne/cmd/sdlc.TestHelperRefreshRunnerResourceExcluded, github.com/xianxu/ariadne/cmd/sdlc.TestHelperRefreshPrecedesSuitePreflight]
---

# flake: TestHelperRefreshRunnerResourceExcluded

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestHelperRefreshRunnerResourceExcluded: Failed in pre-close attempt d5d9d108b935de10548b6452dd5c2396 on 6eb2f26b (ariadne#592, a test-fixture-only diff): runLocalCheck reported TimedOut after 338us with TimeoutMinutes 1, shard6 under full-suite load. Passed on 468ee2e4 an hour earlier, and passes 5/5 alone on 6eb2f26b.

## Spec

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [ ]

## Log

### 2026-10-24

- 2026-10-24 (ariadne:2): registered
  `TestHelperRefreshPrecedesSuitePreflight` here with --evidence: same head
  b8b2e7f9 (ariadne#593) failed in suite attempt
  b60aa6b12878755b709c1f56faaa8c93 (the recorded subtest reported TimedOut
  after 0.66 ms on a 1-minute timeout) and passed in attempt
  049e81aac7f5dc327189075c6bb967c0; 5 of 5 in isolation. ariadne#599 explains
  this shape: a local check's deadline started before scratch setup and was
  read after slot and queue bookkeeping, so under load setup spent a test's
  1-second budget. Once #599 lands, this issue's tests should no longer fail
  this way; check before closing it.
