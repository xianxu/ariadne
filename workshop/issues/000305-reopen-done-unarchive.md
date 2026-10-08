---
id: 000305
status: open
deps: []
github_issue:
created: 2026-10-08
updated: 2026-10-08
estimate_hours:
card_mirror: '55aaea43941f0448726ec92e166b8aa8cfd3981f' # card fields mirrored from issue-cards; edit via sdlc
---

# Reopening a done issue does not un-archive its details

## Problem

Filed from #286. A merged (`done`) issue's details are archived to `workshop/history/issues/` on main. Reopening it (`set-status working`) leaves them there, and no issue branch is restored. Work resumed on a fresh branch either re-creates `workshop/issues/X` beside the archived copy, or merges main and has git follow the archive rename, which puts the live details back in `history/`.

#286 solved this for abandoned issues (`restoreAbandoned`: merge main, then move the archived details and plans back out in one commit). A `done` issue has no kept branch, but the un-archive half applies as is.

## Done when

- Reopening a `done` issue whose details are archived on main moves them (and its plan artifacts) back to `workshop/issues/` and `workshop/plans/` on the reopened branch, so its next landing un-archives them cleanly through the transfer guard.

## Plan

- [ ]

## Log

### 2026-10-08
