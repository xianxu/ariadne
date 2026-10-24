---
id: 000587
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
card_mirror: '97ba78adc33785a6cabc013e8dcf2a967ba534e1' # card fields mirrored from issue-cards; edit via sdlc
---

# Lint: flag tests that keep a real production deadline around a synchronous fake

## Problem

## Spec

**Catch tests that keep a real production deadline around a synchronous fake (pair#852's root cause, 10-23).** pair#827
landed (07:16) with a test that kept the real 1-second completion timeout while its fake completed synchronously; the
test passed under the morning's load, then failed deterministically under the evening's (finalization took about
1.1 s), turning pair's main red for about two hours and failing every couchcore suite and train. ariadne's
`pkg/golint` already has `test-shrunk-deadline`, but it only flags sub-second literal constants, not a test that
leaves a production timeout in place around fake work.

Fix: extend the lint (or add a sibling rule) to flag a test that exercises code with a real wall-clock budget
(production timeout constants, or `context.WithTimeout` built from them) where the test's own seam for that budget
exists but isn't used; and a short allowlist mechanism for genuine end-to-end timing tests. Distribute it to pair
through the usual lint catalog.

## Done when

- The lint flags pair#827's original test shape (fixture) and passes pair#852's corrected one (test).
- No false positives on the current pair and ariadne trees, or each is allowlisted with a reason (run recorded).

## Plan

- [ ]

## Log

### 2026-10-23
