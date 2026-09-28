---
id: 000259
status: open
deps: [000255]
github_issue:
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: '4e102fcd8913cf6021bdd7d56e7216c96744cc5b' # card fields mirrored from issue-cards; edit via sdlc
---

# close evidence and other status readers drop the first modified path

## Problem

`trackerEnv.git` trims its output, and `closeEvidence` slices each `git status
--porcelain -z` entry at a fixed column (`entry[3:]`). A modified file's entry
starts with a space (` M path`); when it is the first entry, the trim eats that
space and the path loses its first character, so the evidence commit records a
no-op deletion of `orkshop/plans/…` and leaves the real file uncommitted. Seen on
every close round after the first: the boundary gate ledger (`-close-gate.md`,
modified, sorts first) was missing from #257's rounds 2–3 and #255's round 2
evidence commits (round 1, where the ledger is new `??`, was fine). The same
trim broke migrate's dirty-issue list (#255, ducks), patched there with a
whitespace split that still mis-parses paths containing spaces.

## Spec

One `git status --porcelain=v1 -z` parser in `gitx` (`ParseStatusZ`): validates
the XY code, returns each entry's path and, for renames/copies, its source,
byte-exact. `trackerEnv` reads status untrimmed through it (`statusPaths`);
`closeEvidence` and migrate's dirty-issue list use it; fleet's status count
reuses it. No reader of status output slices fixed columns from trimmed text.

## Done when

- A close whose gate ledger is modified (not new) commits it in the evidence
  commit (regression test fails today).
- Migrate names a dirty issue path containing a space exactly.
- `ParseStatusZ` has unit tests (modified-first, untracked, rename, spaces,
  malformed); fleet's count uses it.

## Plan

- [ ] `gitx.ParseStatusZ` + tests; fleet's count reuses it
- [ ] `trackerEnv.statusPaths`; `closeEvidence` and `dirtyIssuePaths` use it
- [ ] Regression tests (second-round close ledger; spaced dirty path)
- [ ] Lesson: never slice status output at fixed columns from a trimming runner

## Log

### 2026-09-28
