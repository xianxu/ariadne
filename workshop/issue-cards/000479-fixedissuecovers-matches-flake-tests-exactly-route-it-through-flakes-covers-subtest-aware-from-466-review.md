---
id: 000479
status: codecomplete
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
started: 2026-10-19T20:21:13-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.08
tracker:
    version: 1
    completion:
        token: close-015ae1d47d5d
        repository: github.com/xianxu/ariadne
        reviewed_head: 04bf0c5e603d97658e0e3827ef8364aa49664498
        evidence_commit: d59fbc351cabaa0d65f50bca7aa0e17bcb115ef1
---

# fixedIssueCovers matches flake_tests exactly; route it through flakes.Covers (subtest-aware, from #466 review)

## Problem

From ariadne#466's close review (Minor, fails safe): `fixedIssueCovers` (cmd/sdlc/flake_fixed.go:104, from #458)
matches a fixed issue's `flake_tests` entries against a failure row by exact string. Since #466 lets `flake_tests`
name exact subtests, a fixed issue registered on a subtest never hides or labels a later failure reported on its
top-level test (or the reverse), so a fixed flake can resurface as an unlabeled failure.
