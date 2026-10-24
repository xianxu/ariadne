---
id: 000582
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
card_mirror: 'e34b39d11c43c34ff7d4922e26b0a1447d76bc51' # card fields mirrored from issue-cards; edit via sdlc
---

# Test-output parser misses '--- FAIL' lines prefixed by terminal escape bytes, so a named failure becomes unattributable

## Problem

## Spec

**Test-output parser misses a failure when terminal escape bytes precede `--- FAIL` (found by pair:9 on pair#841,
10-23).** A pair affected-check attempt (e5a7db4c973ffa5a5ff24dae9232b02b on d30f7c51) failed only in
`couchcore.TestLeaveParkRetriesProvenRollbackWithinOneQuit/15`, but OSC/title and setup-failure bytes written by the
test sat immediately before the `--- FAIL` line, so sdlc's parser recorded no `failed_tests` and classified the attempt
"unattributable-failure". That blocks the normal known-flake resume and forces a new head and full rerun even when the
failure is a single named test.

Fix: strip or tolerate terminal control sequences (OSC, CSI, title sets) and partial-line prefixes when matching Go
test result lines (`--- FAIL`, `--- PASS`, `FAIL\t`, `ok\t`), without loosening what counts as a failure. Add the real
captured line as a fixture.

## Done when

- The captured pair#841 log parses to `failed_tests: [couchcore.TestLeaveParkRetriesProvenRollbackWithinOneQuit/15]`
  (fixture test); control sequences elsewhere don't create or hide results (tests).

## Plan

- [ ]

## Log

### 2026-10-23
