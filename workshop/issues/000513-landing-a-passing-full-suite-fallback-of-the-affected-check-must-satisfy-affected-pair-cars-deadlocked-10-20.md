---
id: 000513
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '09290463546918e0fb586bc34abdf22777872065' # card fields mirrored from issue-cards; edit via sdlc
---

# landing: a passing full-suite fallback of the affected check must satisfy 'affected' (pair cars deadlocked 10-20)

## Problem

On 10-20 pair's affected check fell back to the full suite on fresh checkouts (the runtime bundle isn't generated,
so the Go graph can't load; pair#738), and the fallback **passed** but posted only `sdlc-local/suite`. `sdlc merge`
then refuses readiness forever: "PR #489 needs recorded local checks on its head: affected". pair#732 (and pair#706)
deadlocked with all checks green; no pair landing for over two hours.

## Spec

- When the affected check falls back to the full suite and that suite passes on the same head, record the affected
  check as satisfied (with "fallback: full suite" in its evidence): the full suite is a superset of any affected set.
- A failing fallback records affected as failed, as now.
- `sdlc landing status` names the fallback when it applies.

## Done when

- Tests pin: fallback pass satisfies the affected requirement on that head; fallback fail does not; a suite on a
  different head does not.
- pair#732-style car (fresh checkout, fallback) becomes ready without manual recovery.

## Plan

- [ ]

## Log

### 2026-10-20
