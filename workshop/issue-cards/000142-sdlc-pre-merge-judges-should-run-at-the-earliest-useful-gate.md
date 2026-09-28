---
id: '000142'
status: wontfix
started: 2026-07-01T23:05:44-07:00
created: 2026-06-29
updated: 2026-07-02
---

# sdlc pre-merge judges should run at the earliest useful gate

## Problem

`sdlc merge` runs pre-merge judges after the PR branch is pushed:

- `plan` — issue plan completeness;
- `specs` — atlas/README sync;
- `lessons` — whether new lessons should be captured.

In pair#84, the `specs` judge found an optional README keybinding gap only at
merge time, after the issue had already been closed with a `SHIP` boundary
review and a PR had been opened. The fix was small, but it forced another
commit/push/merge loop.

Some of these checks may be more useful earlier:

- plan completeness belongs near `sdlc close`, before status becomes done;
- atlas/README sync may belong near close because it is part of "is the issue
  complete?";
- lessons might still be appropriate at merge, because it considers the whole
  branch/session and is explicitly a pre-ship reflection.

The current all-at-merge placement maximizes late discovery.
