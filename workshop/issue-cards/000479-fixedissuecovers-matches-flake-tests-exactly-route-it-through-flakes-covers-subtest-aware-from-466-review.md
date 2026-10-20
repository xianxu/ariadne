---
id: 000479
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
---

# fixedIssueCovers matches flake_tests exactly; route it through flakes.Covers (subtest-aware, from #466 review)

## Problem

From ariadne#466's close review (Minor, fails safe): `fixedIssueCovers` (cmd/sdlc/flake_fixed.go:104, from #458)
matches a fixed issue's `flake_tests` entries against a failure row by exact string. Since #466 lets `flake_tests`
name exact subtests, a fixed issue registered on a subtest never hides or labels a later failure reported on its
top-level test (or the reverse), so a fixed flake can resurface as an unlabeled failure.
