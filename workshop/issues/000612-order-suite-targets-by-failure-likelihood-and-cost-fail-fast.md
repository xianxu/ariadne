---
id: 000612
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '3356d2da3e889de5028383e20de544e6fae8b795' # card fields mirrored from issue-cards; edit via sdlc
---

# Order suite targets by failure likelihood and cost (fail fast)

## Problem

## Spec

**Order suite targets by failure likelihood and cost, so a doomed suite fails in a minute (operator approved, 10-24).**
pair#866 lost three full suites, each to a single cheap target that failed after 15+ minutes of other work. Source:
ops session.side/2026-10-24-09-test-inventory-across-lifecycle.md (operator approved 10-24).

Change: a repo's suite declares (or sdlc infers from test-stats history) target order: cheap and often-failing targets
first (inventory, lint, catalog, script and Lua targets, recently failing packages), expensive and stable ones last,
stopping at the first failure only when configured to fail fast.

## Done when

- A suite can declare or derive a target order; the runner honours it; a test pins ordering from history.
- pair's suite uses it; the measured time to first failure on a known-bad head drops accordingly.

## Plan

- [ ]

## Log

### 2026-10-24
