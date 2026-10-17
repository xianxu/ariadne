---
id: 000429
status: open
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'f621a3a1d5f0521b9bcf708373e24f9117e4c6f6' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-17
