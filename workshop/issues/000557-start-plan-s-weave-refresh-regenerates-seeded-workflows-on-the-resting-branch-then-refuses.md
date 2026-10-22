---
id: 000557
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'dbfe3801379f7d47bbd3b1d356759d10cab32f96' # card fields mirrored from issue-cards; edit via sdlc
---

# start-plan's weave refresh regenerates seeded workflows on the resting branch, then refuses

## Problem

## Spec

`sdlc start-plan`'s catch-up bumps a peer repo's ariadne pin and runs `weave refresh`, which regenerates seeded files
(here `.github/workflows/merge-check.yml`: ariadne's SIGPIPE/drain fix in the compiler-provenance step, two
`grep -Eq` lines becoming `grep -E ... >/dev/null` plus a drain comment) **on the slot's resting branch**. start-plan
then refuses, because the resting branch has uncommitted tracked changes beyond the issue's details. Every slot that
refreshes against the new ariadne hits this:
- pair:8 (pair#809, 10-22 13:1x): bump ariadne 2dc46703 → 5785330c; saved the patch, restored the file, re-ran.
- parley.nvim:3 (parley.nvim#409, 10-22): "weave-generated merge-check workflow drift" kept in a named stash.

Resting branches only fast-forward (#284), so the refresh must not leave tracked changes there, and seeded workflow
files are local-gates configuration (#309): their update has to reach main as its own change the operator verifies,
not ride along on an issue branch.

Fix: start-plan's catch-up never leaves regenerated tracked files on a resting branch. When a refresh would change a
tracked seeded file, it restores the file, reports "seeded file X is behind ariadne; file or land the weave refresh
as its own change", and continues (or the refresh runs where it can't dirty the resting branch). Plus a way for the
pending regeneration to reach each consumer's main as one local-gates change.

## Done when

- A start-plan whose catch-up would regenerate a tracked seeded workflow leaves the resting branch clean, proceeds,
  and names the pending seeded-file update (test).
- The pending update can be produced as one reviewable change for the consumer repo (documented, tested).

## Plan

- [ ]

## Log

### 2026-10-22
