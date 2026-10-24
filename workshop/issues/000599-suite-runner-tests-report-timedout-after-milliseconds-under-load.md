---
id: 000599
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: 'ef0d89618caa6be1330d0b7b52047cc8aeb2c41f' # card fields mirrored from issue-cards; edit via sdlc
---

# Suite-runner tests report TimedOut after milliseconds under load

## Problem

Under machine load, fake suites in the local-check and suite-runner tests
report `TimedOut` after milliseconds on a 1-minute or 30-minute timeout. Seen
on ariadne#593's head b8b2e7f9 in two full suites (2026-10-24):

- Attempt b60aa6b12878755b709c1f56faaa8c93: `TestHelperRefreshPrecedesSuitePreflight/recorded`
  reported TimedOut after 0.66 ms (`command: true`, TimeoutMinutes 1).
- Attempt 049e81aac7f5dc327189075c6bb967c0 (wall 1761s against a usual 1081s):
  nine failures, several the same shape (TimedOut after 19 to 33 ms):
  `TestRunLocalCheckLightOverrunRetriesInNormalLine`,
  `TestRunLocalCheckTerminatesDescendants/timeout`,
  `TestSuiteCapabilityResumeRetryAfterRepair`,
  `TestSuiteEnvironmentBaselineIsNotRedMain`,
  `TestSuiteEnvironmentRetentionFailureKeepsRawEvidence/false` and `/true`,
  `TestSuiteInfrastructureBaselineCannotBlameCars`,
  `TestSuiteInfrastructureFailureHasNoCodeAuthority/affected`,
  `TestSuiteInfrastructureMissingGoNeverStarts`.

Run in isolation, eight of the nine pass.
`TestRunLocalCheckLightOverrunRetriesInNormalLine` fails deterministically
outside the suite, identically on main e5ba9973f ("the retry does not carry the
overrun"), though main's baseline inside `make test` passed it earlier the same
day.

A TimedOut after under 1 ms on a 1-minute timeout is not plain load; it looks
like a deadline computed from a stale or wrong start time (TL observation).
If production suite runs share that computation, a real suite could be killed
early, which is serious. `TestHelperRefreshPrecedesSuitePreflight` is
registered under #596 meanwhile; the others are deliberately not registered as
flakes, since they are one bug.

## Spec

## Done when

- The root cause of the early TimedOut is found and fixed, and a test pins it
  without depending on machine load.
- It is stated, with evidence, whether production suite runs were exposed.
- `TestRunLocalCheckLightOverrunRetriesInNormalLine` passes deterministically
  in and outside `make test`.
- The listed tests pass in a full suite.

## Plan

- [ ]

## Log

### 2026-10-24
