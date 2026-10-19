---
id: 000462
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'cf03b1ce96314a827d14766b7ca9e99ad90c3c54' # card fields mirrored from issue-cards; edit via sdlc
---

# flake: TestSuiteQueuePriorityRealHelper

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestSuiteQueuePriorityRealHelper: Observed same-exact-head fail then pass at Git commit 70a7180bbdfc215ef63e281a128256d8a19c38d1. Full recorded suite attempt fa09ccf7eb6fd0862fadf1d94e865944 failed only TestSuiteQueuePriorityRealHelper: private priority provider Popen returned Operation not permitted, falling back to normal FIFO. Log /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/fa09ccf7eb6fd0862fadf1d94e865944.log. Unchanged detached worktree /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne461-rerun.iGuqYv/check at that exact Git commit ran go test ./cmd/sdlc -run ^TestSuiteQueuePriorityRealHelper$ -count=1 -v; PASS TestSuiteQueuePriorityRealHelper32.92s, package33.878s, exit0. No source or environment fix applied; worktree clean. Earlier deterministic FleetEndToEnd fixture failure was on a different commit and was fixed separately.

## Spec

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [ ]

## Log

### 2026-10-19
