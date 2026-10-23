---
id: 000567
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '3fa5582e5a694d43c547f9e1cfb8b202e0025d3b' # card fields mirrored from issue-cards; edit via sdlc
---

# flake: TestOperatingCancellationStopsGitHookDescendants

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/tl-lead.TestOperatingCancellationStopsGitHookDescendants: Recorded #563 suite attempt 2cb913014a018f68422fdf9ef52180cf on Git-verified head 6a0b35faea2b351b5b1ec19a310f6ea1d19f1948 failed at doc_operating_test.go:419: Git descendant remains runnable: ?E (0.87s). Unchanged exact detached checkout targeted command go test ./cmd/tl-lead -run ^TestOperatingCancellationStopsGitHookDescendants$ -count=1 -timeout 2m returned exit0 PASS in0.986s. Captured original JSON events and targeted result are retained in common test-stats/evidence/000563/cancellation-failure.jsonl and cancellation-targeted-pass.txt. Full original suite is still running; no full retry or source edit. Investigate whether transient Darwin exiting state E is misclassified as runnable, without weakening cleanup proof.

## Spec

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [ ]

## Log

### 2026-10-22
