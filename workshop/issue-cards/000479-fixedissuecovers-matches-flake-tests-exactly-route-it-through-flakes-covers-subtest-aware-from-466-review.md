---
id: 000479
status: done
created: 2026-10-19
updated: 2026-10-20
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
actual_hours: 0.13
tracker:
    version: 1
    completion:
        token: close-596b2a557246
        repository: github.com/xianxu/ariadne
        reviewed_head: 3bea1fc6f744779005297def5e253cff45e17aa2
        evidence_commit: 11caf9dee5c1b26f5c018d833539bb3348f52523
        landed_commit: 6c4646af19d210a474e70625b3c6e4f0d4f20f9c
---

# fixedIssueCovers matches flake_tests exactly; route it through flakes.Covers (subtest-aware, from #466 review)

## Problem

From ariadne#466's close review (Minor, fails safe): `fixedIssueCovers` (cmd/sdlc/flake_fixed.go:104, from #458)
matches a fixed issue's `flake_tests` entries against a failure row by exact string. Since #466 lets `flake_tests`
name exact subtests, a fixed issue registered on a subtest never hides or labels a later failure reported on its
top-level test (or the reverse), so a fixed flake can resurface as an unlabeled failure.
