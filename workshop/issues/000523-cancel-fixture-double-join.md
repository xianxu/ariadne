---
id: 000523
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '13f56da2204150e4678f3a8a8a6e5e057a6d1163' # card fields mirrored from issue-cards; edit via sdlc
---

# Native cancellation fixture waits four minutes after successful joins

## Problem

TestLocalCheckCancelOnlyOwnRun spends about four minutes waiting in successful cleanup after both cancellation assertions already joined their native children. This inflates the machine-wide suite wall time without exercising more behavior.

## Spec

Make the fixture child completion join idempotent while retaining actual native-process cancellation, foreign-checkout refusal, PID reuse fencing, bounded teardown and child cleanup on failed assertions. Do not reduce or bump a timeout to hide the wait.

## Done when

- Successful test completion does not wait either two-minute HangCeiling after both children were joined.
- Foreign checkout stays alive; owning checkout cancellation stops its managed run and check; failure teardown still joins/reaps children.
- A controlled regression demonstrates repeated cleanup/join observation completes without waiting for the fallback ceiling.

## Plan

- [ ]

## Log

- Discovered during independent #521 cancellation validation. Current cmd/sdlc/localcheck_lease_test.go TestLocalCheckCancelOnlyOwnRun start helper sends one result into buffered c.done but never closes it. The main assertion consumes a.done and b.done; each registered cleanup then selects c.done again versus time.After(testfix.HangCeiling). Both results have already been consumed, so each cleanup waits two minutes.
- Targeted current-branch run passed the native control in240.56s (whole held/cancel command256.650s), while no child processes remained during its long final wait. This is deterministic cleanup overhead, not a registered flake or a #521 runtime defect. Native source unchanged by #521. Main verification summary retained at common test-stats/evidence/000521/main-verification-summary.txt; raw pass timing also appears in the worker tool transcript. No nonexistent raw log is cited.
