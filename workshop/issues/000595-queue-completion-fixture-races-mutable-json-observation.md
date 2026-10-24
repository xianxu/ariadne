---
id: 000595
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '20254cc23e836a805fb95448912f61c20744b737' # card fields mirrored from issue-cards; edit via sdlc
---

# Queue completion fixture races mutable JSON observation

## Problem

Recorded ariadne#589 suite hit JSONDecodeError in the queue-helper completion fixture while polling mutable operation bytes. The same top-level test then passed on its exact failed head with no source changes.

## Spec

Fix the fixture's record-observation race at scripts/test/test-lock-exit.test.py:269 without weakening production lifetime/absence proof, replacing a lock-owned inode blindly, or increasing deadlines. A transient unreadable/partial record is not a successful interrupt acknowledgement. Preserve the original evidence and prove the observation protocol deterministically.

## Done when

- A deterministic fixture forces an observation during an incomplete update and waits for valid interruption evidence, rather than throwing JSONDecodeError or treating uncertainty as absence.
- Existing completion, cancellation and lifetime controls pass with the real lock helper.
- The registered flake is unregistered on close.

## Plan

- [ ]

## Log

2026-10-24 — Exact-head fail/pass classification by ariadne:3 during #589. Source d4575ac91559a6a8d41d9f781f3e52dfb8465fe1, original recorded attempt543d32f001f5bf36a8cb7d885ca80d4d. Suite failure: TestQueueRealHelperCompletion/test-lock-exit.test.py, DirectSuiteExit.test_cancellation_wins_before_recipe_completion, JSONDecodeError at scripts/test/test-lock-exit.test.py:269 (json.loads(path.read_bytes())). Separate TestVerbContextsReachTrackerReads is #589's own deterministic code failure, not a flake.

Private detached exact-head Go rerun of TestQueueRealHelperCompletion passed4.504s (both Python helper subtests passed). Git HEAD before/after unchanged and both status checks clean. Evidence: shared Git directory sdlc/test-stats/evidence/000589/queue-completion-classification.md and original-full-suite.log; private checkout /Volumes/TestScratch/tmp/ariadne-slot3/ariadne589-classify.3b7IeV/repo. No manual process/token/queue changes. Filed unclaimed under the self-applied flake rule; this is independent of the SysV feature.
