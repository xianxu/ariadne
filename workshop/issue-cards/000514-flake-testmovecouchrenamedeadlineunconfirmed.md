---
id: 000514
status: codecomplete
created: 2026-10-20
updated: 2026-10-21
estimate_hours:
github_issue:
flake_tests: [github.com/xianxu/ariadne/cmd/tl-lead.TestMoveCouchRenameDeadlineUnconfirmed, github.com/xianxu/ariadne/cmd/tl-lead.TestMoveCouchKnownRefusals, github.com/xianxu/ariadne/cmd/tl-lead.TestMoveCouchPermissionFailsClearly]
started: 2026-10-20T18:24:57-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 2.27
tracker:
    version: 1
    completion:
        token: close-aa327dab6873
        repository: github.com/xianxu/ariadne
        reviewed_head: ed14a3fc97dc91e10b245f00516fce367dbc308a
        evidence_commit: ff28d6a6125253b3c7a54bb479f80e4cf2d871da
---

# flake: TestMoveCouchRenameDeadlineUnconfirmed

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/tl-lead.TestMoveCouchRenameDeadlineUnconfirmed: 2026-10-20 ariadne:2 integration of origin/main 94bf14d9262f4ced7db0ebdb43b44583d7d8e2d5, HEAD 4cf622a73b25b158202e54a200f94c8cc591f4fa, exact unchanged source/index tree c7acf7c0b00664949af14a68bd6d27b446cd8328. TL_LEAD_RESUME_CONFORMANCE=1 go test -race ./cmd/tl-lead -count=1 failed only TestMoveCouchRenameDeadlineUnconfirmed at status_move_timeout_test.go:88: rename never published, effect missing after 300ms. Then go test -race ./cmd/tl-lead -run ^TestMoveCouchRenameDeadlineUnconfirmed$ -count=10 -v passed all ten with no source changes. The test starts a real race-instrumented helper and assumes it reaches WriteFile before the 300ms subprocess deadline; startup under load can consume that budget. Fix should synchronize helper readiness/effect before exercising cancellation rather than assume process-start timing.
