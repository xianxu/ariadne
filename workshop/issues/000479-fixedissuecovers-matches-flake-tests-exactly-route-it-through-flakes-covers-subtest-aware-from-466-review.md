---
id: 000479
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c0caccb105d270bf9fbc8cf7c1eef272a498b3fa' # card fields mirrored from issue-cards; edit via sdlc
---

# fixedIssueCovers matches flake_tests exactly; route it through flakes.Covers (subtest-aware, from #466 review)

## Problem

From ariadne#466's close review (Minor, fails safe): `fixedIssueCovers` (cmd/sdlc/flake_fixed.go:104, from #458)
matches a fixed issue's `flake_tests` entries against a failure row by exact string. Since #466 lets `flake_tests`
name exact subtests, a fixed issue registered on a subtest never hides or labels a later failure reported on its
top-level test (or the reverse), so a fixed flake can resurface as an unlabeled failure.

## Spec

Route `fixedIssueCovers` through `flakes.Covers`, the same subtest-aware matcher #466 introduced for known flakes, so
fixed and known flakes use one coverage rule (ARCH-DRY).

## Done when

- A fixed issue registered on `TestX/sub` covers a failure row naming `TestX/sub` and is handled by the same rule as
  a known flake for `TestX` rows (table test mirroring #466's Covers tests).
- No other caller keeps an exact-string match for flake coverage (grep check in the test or a lint).

## Plan

- [ ]

## Log

### 2026-10-19
