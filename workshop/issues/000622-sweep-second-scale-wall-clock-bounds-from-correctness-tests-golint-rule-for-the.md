---
id: 000622
status: working
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: 'ef95d347a68f5e50ff1951bfd2d0c72785e82e66' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-25T04:05:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

Durable plan: `workshop/plans/000622-sweep-second-scale-wall-clock-bounds-from-correctness-tests-golint-rule-for-the-plan.md`.

- [ ] M1 — the widened `test-tight-wallclock` rule (a transitional per-file pending list), group A polled deadlines, and #625
- [ ] M2 — groups B, C, D; delete the pending list so the rule enforces; close #625 with this issue

## Log

### 2026-10-25

### 2026-10-25

- TL dispatch: sweep plus rule, fold in #625. The widened rule flags 35 sites (6 named). TL scope call: all four groups here, full flow, M1 = rule + group A + #625, M2 = B/C/D + enforce; proofs per package `-count=10`, one at a time.
- Rule shape: deadlines flagged only when polled (`time.Now().After/Before(v)`, `time.Until(v)`), which cut 73 raw hits (fixture timestamps) to 35.
