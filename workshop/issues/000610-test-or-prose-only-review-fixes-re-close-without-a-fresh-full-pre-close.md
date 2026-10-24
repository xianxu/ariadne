---
id: 000610
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '91caab1c8b982b048155c3d7c339d1936f809139' # card fields mirrored from issue-cards; edit via sdlc
---

# Test- or prose-only review fixes re-close without a fresh full pre-close

## Problem

## Spec

**A review fix that touches only tests or prose doesn't need a fresh full pre-close (operator approved, 10-24).** Each
close round that changes anything currently requires a fresh full pre-close before re-close; pair#866 queued five full
suites largely this way. Source: ops session.side/2026-10-24-09-test-inventory-across-lifecycle.md (operator approved 10-24).

Change: if a close round's delta since the last passing pre-close touches only test files, docs or issue records
(classified the same way affected selection classifies paths), the re-close accepts targeted proof of the changed
tests plus the earlier pass; production-code deltas still need a fresh pre-close. The integrated suite at merge
covers the rest.

## Done when

- A test-only or prose-only round re-closes on targeted proof plus the earlier pass; a production-code round still
  requires a fresh pre-close; both tested.

## Plan

- [ ]

## Log

### 2026-10-24
