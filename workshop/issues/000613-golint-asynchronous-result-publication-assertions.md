---
id: 000613
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: 'cdde5705a7d4b6709b6d69ea3a4ea66fae5ff8e5' # card fields mirrored from issue-cards; edit via sdlc
---

# golint: asynchronous result publication assertions

## Problem

echo#36 found a test that assumes a caller's result-channel send is ready immediately after Stop joins an internal operation. RestoreFork releases its supervisor-owned WaitGroup during its return defer; the caller goroutine sends the result only after that return. The supervisor can correctly join hydration while the test's nonblocking select/default still reports failure.

Concrete case: echo supervisor/replay_test.go, TestRestoreForkShutdownCancelsAndJoinsHydration, assertion "Stop returned before hydration owner". Reported on echo#27 and echo#30 CI integrations. echo#36 fixes the test with a deterministic delayed-publication fixture, direct hydration-reader exit evidence, and a bounded result receive.

## Spec

Design and add a defensible Go lint-catalog guard for positive completion assertions that conflate an internal Stop/Close join with a caller goroutine's subsequent result publication. Cite echo#36 in the rule and fixtures. Trace the actual producer/result ordering within the supported AST shape; do not infer ownership solely from a variable named done.

Distinguish legitimate nonblocking negative assertions (a gated operation must not have completed), synchronously published results (for example a synchronous router delivery), and unsupported/unknown producer relationships. Do not broadly ban select/default or require a delay in every channel test. If a sound narrow detector cannot be implemented, record the boundary with fixtures and a concrete alternative for review rather than shipping false positives.

This is an independent Ariadne lint design; echo#36 remains a bounded test-only fix and does not depend on an unlanded lint implementation. Machine-load policy applies: targeted verification only within current limits, full suites through managed local-check.

## Done when

- The supported asynchronous wrapper/result pattern is detected by a registered pkg/golint catalog rule, with the echo#36 shape as a regression fixture.
- Negative nonblocking assertions and synchronously published results are accepted by contrasting fixtures; the detector's supported and unsupported dataflow is documented.
- The bounded receive correction is accepted, and catalog/schema/documentation tests plus managed verification pass.

## Plan

- [ ]

## Log

### 2026-10-24
