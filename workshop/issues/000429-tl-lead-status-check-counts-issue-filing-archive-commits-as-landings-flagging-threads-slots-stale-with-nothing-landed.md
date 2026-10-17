---
id: 000429
status: working
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: '3a97ea3b66cace39b614fbd5477d416c5dbde65c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T12:28:10-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# tl-lead status check counts issue-filing/archive commits as landings, flagging Threads/Slots stale with nothing landed

## Problem

## Spec

`tl-lead status check` marks Threads and Slots stale when they're older than "the fleet's newest landing". It treats any commit on a fleet repo's origin/main as a landing. That includes `#N: issue: new` filings, issue publishes, "archive completed issues to history" and project ticks. On 2026-10-17 the check and pair#630's reminder flagged the session file stale several times an hour after issue filings alone, for example parley.nvim 12:17 "#378: issue: new" and pair 11:46 "#652: issue: new". That makes the hub re-stamp sections with nothing to report and dulls the signal.

A landing is a merged PR: a merge commit from `sdlc merge`, or a landing ledger entry. Prefer the landing ledger or merge-commit shape over commit-message heuristics if one is available.

## Done when

- Issue-filing, publish, archive and tick commits don't count as landings. A merged PR does. Tests cover both.
- The stale message names the landing it compared against: repo, PR or merge commit, and time.

## Plan

- [ ] Reproduce direct-main false landings with real Git; cover first-parent merges and identity in stale output.
- [ ] Select the latest first-parent merge, render its hash, and synchronize help/atlas.
- [ ] Run focused checks and the recorded suite, close, and land on SHIP.

## Log

### 2026-10-17

- Root cause: gitNewestLanding reads the origin/main tip timestamp without filtering. Use first-parent merge shape, which sdlc merge emits via gh pr merge --merge; no subject heuristics. The landing-line ledger is local, has 30-day retention, and records completion rather than the immutable merge commit. Existing Git seam and stale renderer remain the owners (ARCH-DRY, ARCH-PURE). No new durable state or lifecycle.
