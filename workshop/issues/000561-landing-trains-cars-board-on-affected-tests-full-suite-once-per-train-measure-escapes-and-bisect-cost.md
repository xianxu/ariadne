---
id: 000561
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'eb4937f47be65fa28366037eb99f9bc3290b9e83' # card fields mirrored from issue-cards; edit via sdlc
---

# Landing trains: cars board on affected tests; full suite once per train; measure escapes and bisect cost

## Problem

## Spec

Landing trains (ariadne#449, #475) rarely batch: since 10-21 21:11 every pair train carried one car (trains 271, 277,
278). The car readiness gate requires a recorded **full suite** on each car's head before it boards ("needs recorded
local checks on its head: affected or its full-suite fallback", and in practice affected escalates to the full suite),
so each car pays a full suite, cars are rarely ready together with the queue an hour deep, and the train's single
suite saves little. pair ready→landed p50 52 min / p90 2.4 h (n=138), unchanged from the pre-train baseline.

Operator decision (10-22): **a car boards on its affected tests only** (pair declares them, pair#725; `make
test-affected`), and the train runs the **full suite once per train**, as originally designed. "Please monitor closely
the train test failure and the cost to bisect. All depends on whether we have a high-quality filter of affected tests
to catch issues early."

Scope:
- Readiness: a recorded **affected** check on the car's head suffices to board; no per-car full-suite fallback
  escalation for boarding (escalation stays for close/merge outside trains).
- The train's full suite stays the landing gate; a failing train splits (existing behaviour).
- **Measure the filter's quality** (the operator's condition): per train, record each failure the full suite found
  that the failing car's affected run passed ("escape"), the escape rate per repo over time, and the bisect cost
  (extra suite runs and wall time spent splitting). Surface it in `sdlc landing status` and a weekly line, so the TL
  can watch it and turn per-car full suites back on for a repo whose escape rate climbs.

- **Start small (operator, 10-22):** begin with trains of at most **2 cars** per repo, so a failing train splits once
  and an escape is cheap to attribute; raise the cap only after the escape rate stays low for about a week. The cap is a
  per-repo setting the TL adjusts, shown in `sdlc landing status`.

## Done when

- A car with a passing affected record boards without a full suite (test); the train's full suite still gates.
- A train failure attributable to a car whose affected run passed is recorded as an escape, with the bisect cost
  (test with a fake suite).
- `sdlc landing status` shows the escape rate and bisect cost for the repo.

## Plan

- [ ]

## Log

### 2026-10-22
