---
id: 000418
status: open
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: 'e8f33a8d24c82699c479aec3a3d849ec77d2c9b1' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc merge: a reused branch name matches the earlier milestone's MERGED PR too and refuses ('found 2')

## Problem

## Spec

Reported by pair:2 (pair#618 M2), 2026-10-16. `sdlc merge` refuses with: "need one exact matching PR for 000618-slots-...; found 2".

How it happened:
1. M1 landed from the issue branch as PR #380, now MERGED. `sdlc merge` then deleted the branch.
2. `start-plan` recreated the same branch name for M2, and M2's PR #387 is OPEN at 9917e65f.
3. PR matching by branch name finds both PRs and refuses.

Every multi-milestone issue that lands milestone by milestone hits this.

Rule: match only PRs whose head is this branch's current head, or that are OPEN. A MERGED or CLOSED PR whose head commit is not the branch head is history, not a candidate. Two OPEN PRs for one branch still refuse.

## Done when

- With a fake GitHub holding one MERGED PR (an older head) and one OPEN PR (the current head) for the same branch, `sdlc merge` picks the OPEN one; a test covers it.
- Two OPEN PRs for the branch still refuse, naming both.
- pair#618 M2's PR #387 lands through `sdlc merge` with no manual gh merge.

## Plan

- [ ]

## Log

### 2026-10-16
