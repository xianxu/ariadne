---
id: 000259
status: working
deps: [000255]
github_issue:
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: '76225029040b45dd2203a24aaa2f30ab7050b3ab' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-09-28T09:25:50-07:00
flow: {kind: quick, provenance: inferred, spec: "976ea22b", done: "fa4e239f"}
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
- Migrate names a dirty issue path containing a space exactly; no status
  reader parses status text by hand.
- `ParseStatusZ` has unit tests (modified-first, untracked, rename, spaces,
  malformed); fleet's count uses it.

## Plan

- [x] `gitx.ParseStatusZ` + tests; fleet's count reuses it
- [x] `trackerEnv.statusEntries`; `closeEvidence` and `dirtyIssuePaths` use it
- [x] Regression tests (second-round close ledger; spaced dirty path)
- [x] Lesson: never slice status output at fixed columns from a trimming runner

## Log

### 2026-09-28
- 2026-09-28: closed — gitx.ParseStatusZ is the one byte-exact status reader (fleet count reuses it); trackerEnv.statusEntries reads untrimmed; closeEvidence and migrate dirtyIssuePaths use it. TestTrackerReCloseCommitsTheModifiedGateLedger and TestIssueMigrateNamesADirtyPathWithASpace fail without the fix; TestParseStatusZ covers modified-first/untracked/rename/spaces/trimmed-refused. gitx/fleet/tracker + Close/Tracker/IssueMigrate/Merge/StartPlan/Leftover/Legacy tests green (only main baseline #210 fails). --no-atlas: internal parser consolidation, no new surface.; review verdict: SHIP
- `gitx.ParseStatusZ` (moved from fleet's validating counter, which now
  returns its length; `ValidStatusCode` exported for fleet's fake git) and
  `trackerEnv.statusEntries` (untrimmed). `closeEvidence` and migrate's
  `dirtyIssuePaths` use it. Evidence commits confirmed the diagnosis: #257
  118c7039 (round 1, ledger new) carried the gate ledger; aad120e7, 6bde769c
  and #255 3092bf0c (later rounds, ledger modified) did not.
- Tests: `TestParseStatusZ` (modified-first, untracked, rename, spaces;
  trimmed-first refused), `TestTrackerReCloseCommitsTheModifiedGateLedger` and
  `TestIssueMigrateNamesADirtyPathWithASpace` — both fail without the fix.
  gitx, fleet, tracker and the Close/Tracker/IssueMigrate/Merge/StartPlan/
  Leftover/Legacy command tests pass (the only failure is main's baseline #210).
- Close round 1 SHIP; its three advisory Minors fixed: every status reader now
  goes through `gitx.ParseStatusZ` — merge's dirty check and its pre-merge
  re-check, start-plan's dirty count, push's interrupted-archive recovery
  (`porcelainPaths` and `parsePorcelainStatus` deleted; `assessDirty` and
  `preparedArchiveMoves` take entries; tests feed real `-z` bytes via
  `statusZ`, plus a spaced-path case); `trackerEnv.gitRaw` is the one exec path
  (`gitEnv` trims it, `statusEntries` parses it); gitx imports grouped with the
  project's. Wider command tests green (786 s).
