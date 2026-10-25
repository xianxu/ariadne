---
id: 000626
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '8e0bddf20c9e9e8effe42eb85b32158aa5475b5f' # card fields mirrored from issue-cards; edit via sdlc
---

# Sweep select time.After fail-timeouts below the hang ceiling; extend test-tight-wallclock

## Problem

ariadne#622's plan-quality review (PQ-1) found a second-scale timeout shape that #622's widened `test-tight-wallclock` rule does not see: a `select` whose `case <-time.After(N * time.Second):` arm fails the test. A grep found about 33 such sites in `_test.go` files under `cmd/` and `pkg/`, for example `cmd/tl-lead/doc_operating_test.go:398` and `cmd/tl-lead/doc_repository_test.go:289,350`. It is the same load flake as ariadne#625's 30 s poll: under fleet load the awaited event arrives late, and the timeout reads as a failure.

## Spec

Extend `tightWallclock` (`pkg/golint/rules.go`) to flag a `time.After(d)` channel received in a `select` case whose body fails the test (`t.Fatal*`, `t.Error*`, `panic`), unless `hangBound(d)` holds (an expression naming `HangCeiling`, or a constant of at least `testgate.HangCeiling`). Reuse #622's `hangBound`; do not add a sibling rule. Sweep the flagged sites: the awaited event stays the oracle, and `HangCeiling` becomes the only hang guard. Where a site proves that something did not happen within a window, use an event if one exists, else `//lint:allow test-tight-wallclock — <reason>`.

Non-goal, same as #622: `context.WithTimeout` deadlines, which usually bound the code under test rather than the test's own wait (`test-shrunk-deadline` owns the sub-second ones).

## Done when

- The rule flags the `select`/`time.After` fail-timeout shape below the hang ceiling, with rule tests (red below the ceiling, green at `HangCeiling`).
- Every flagged site is swept or carries a reasoned allow; `TestTrackedGoFilesPassTheCatalog` passes.
- Each changed package's tests pass `-count=10`, one run at a time.

## Plan

- [ ]

## Log

### 2026-10-25
