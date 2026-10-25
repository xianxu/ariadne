---
id: 000622
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '5647c87391c0bab0c4d5f4687134ef4d52456aff' # card fields mirrored from issue-cards; edit via sdlc
---

# Sweep second-scale wall-clock bounds from correctness tests; golint rule for the shape

## Problem

## Spec

**Second-scale wall-clock bounds in correctness tests (found by ariadne:2 fixing ariadne#620, 10-25).** ariadne#620's
BR-16 subtest asserted that a run finished within 8 s of wall time; under fleet load it failed on main and blocked
every ariadne landing. ariadne:2 fixed it by asserting the outcome (the rerun's saved attempt timed out) instead of
the elapsed time. The same shape remains elsewhere: `cmd/sdlc/ghlanding_test.go:153`,
`helpercapability_test.go:167`, `landingattribution_impact_test.go:126`, and tests in tl-lead and weave.

Test-economy's rule (project file, 10-15): no wall-clock assertions in correctness tests; wait on signals, inject
clocks, assert outcomes. Sweep these to outcome or injected-clock assertions, and add a `pkg/golint` rule that flags
a duration comparison against `time.Since`/`time.Now` deltas in `_test.go` files (with a reviewed escape hatch for
true performance tests), so new ones are caught at lint time.

## Done when

- Each named site asserts an outcome or uses an injected clock; each change proven with -count=10 under load.
- The golint rule flags the old shape (a fixture showing it red on the pre-fix BR-16) and passes the swept tree.

## Plan

- [ ]

## Log

### 2026-10-25
