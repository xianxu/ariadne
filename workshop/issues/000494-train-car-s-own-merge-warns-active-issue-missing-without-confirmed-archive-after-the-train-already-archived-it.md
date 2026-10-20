---
id: 000494
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c12c8f52b22f7094cfeff79316d8ec78f48d375e' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T06:34:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# Train car's own merge warns 'active issue missing without confirmed archive' after the train already archived it

## Problem

In pair's first multi-car train (176, 10-20 05:48), car pair#717's own `sdlc merge` (PR 483) ended with "active issue
missing without confirmed archive", although the train had already archived the issue (it's under history on main
and its card is done). Harmless, but it reads as an error at the end of a successful landing, and a worker can't tell.

## Spec

When a car's merge finishes after the train landed it, treat "archived by the train" as the confirmed outcome (the
train's `outcomes/<seq>` record says so), and print it as such, not as a missing-issue warning.

## Done when

- A 2-car train fixture: the second car's merge ends with a confirmed train-archive message, no warning (test).

## Plan

- [ ] Reproduce the tracked two-car cleanup ordering and confirm archive provenance.
- [ ] Retain the train outcome and recognize its already archived exact completion without weakening serial archive checks.
- [ ] Verify focused controls, recorded suite, fresh close review and landing.

## Log

### 2026-10-20

- Investigation: pair landing events record train 176 landing tickets 176/178 (pair#724/#717) with train PR #484. Git history identifies #717 archive commit 7d4e0c3ab as the generic “archive after a merge outside sdlc” sweep, without Landing-PR/Landing-Head trailers. The conductor completes all reachable cards, archives its own PR, then finishLandedLeftovers archives sibling done cards. The waiting merge discards its landed outcome, and confirmLandingArchive requires its own PR trailers; missing active details then fail planLandingArchive. Existing outcome confirms landed only, not archive completion, and precedes cleanup. Preserve this distinction (ARCH-ORDER/SECURE); no production edits yet. A tracked two-real-car fixture is required: legacy train fixtures do not run the done-card sweep.
