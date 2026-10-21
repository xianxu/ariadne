---
id: 000514
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '8212ed4a6ec579495595c5e8e53fafd330c237bf' # card fields mirrored from issue-cards; edit via sdlc
---

# flake: TestMoveCouchRenameDeadlineUnconfirmed

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/tl-lead.TestMoveCouchRenameDeadlineUnconfirmed: 2026-10-20 ariadne:2 integration of origin/main 94bf14d9262f4ced7db0ebdb43b44583d7d8e2d5, HEAD 4cf622a73b25b158202e54a200f94c8cc591f4fa, exact unchanged source/index tree c7acf7c0b00664949af14a68bd6d27b446cd8328. TL_LEAD_RESUME_CONFORMANCE=1 go test -race ./cmd/tl-lead -count=1 failed only TestMoveCouchRenameDeadlineUnconfirmed at status_move_timeout_test.go:88: rename never published, effect missing after 300ms. Then go test -race ./cmd/tl-lead -run ^TestMoveCouchRenameDeadlineUnconfirmed$ -count=10 -v passed all ten with no source changes. The test starts a real race-instrumented helper and assumes it reaches WriteFile before the 300ms subprocess deadline; startup under load can consume that budget. Fix should synchronize helper readiness/effect before exercising cancellation rather than assume process-start timing.

## Spec

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [ ]

## Log

### 2026-10-20
