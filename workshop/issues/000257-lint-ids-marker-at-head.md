---
id: 000257
status: working
deps: [000255]
github_issue:
created: 2026-09-27
updated: 2026-09-27
estimate_hours:
card_mirror: 'f1ea34457673ed8ee3f29a6c2e96db1b454961eb' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-09-27T23:33:11-07:00
flow: {kind: quick, provenance: inferred, spec: "3b3861c1", done: "e78591ed"}
---

# lint-ids reads the cutover marker from the checked commit, not the checkout

## Problem

`sdlc issue lint-ids --base B --head H` (merge-check `40-duplicate-issue-id.sh`)
decides the cutover state from the **checkout's** `workshop/issue-tracker.json`,
not from `H`. In a pre-push hook the pushed commit is not checked out, so the
two disagree. Found cutting over you-decide (#255, 2026-09-27): `migrate --apply`
had bootstrapped the tracker and was pushing the commit that adds the marker;
the hook's lint saw "tracker exists, checkout lacks the marker", exited 2, and
the publish gate blocked the push. The apply finished only with the hook
skipped once (operator-authorized). Any repository with the pre-push publish
gate hits this at cutover, and the lint is wrong in general: it must judge the
commit under check.

## Spec

lint-ids reads the cutover marker (and so the tracked/legacy decision) from
the tree of its head (`--head`, default `HEAD`), never the checkout's files. The
cutover mismatch guard still refuses a real mismatch at `H`.

## Done when

- A pre-push of the migration commit from a checkout without the marker passes
  the id lint (e2e: bootstrap tracker, lint `--head` the marker commit).
- A head whose marker names another tracker root still refuses.
- Reading a commit's marker treats only a missing path as absent: an
  unresolvable head is an error, never "not cut over"; lint-ids and the
  migration's already-migrated check share that one reader.

## Plan

- [x] Read the marker at `--head` in lint-ids' mode decision
- [x] Tests for both outcomes

## Log

### 2026-09-27
- 2026-09-27: closed — Round 3: migratedAlready returns the root it read; reconcile drops its re-read (and the discarded parse error). IssueMigrate/Leftover/LintIDs green. Earlier evidence stands (you-decide reproduced without the fix; commit guard unit test).; review verdict: SHIP
- 2026-09-27: closed — Round 2: four close-review Minors fixed (careful ReadCutoverMarkerAt: only a missing path is absent; one commit-marker reader for lint-ids and migratedAlready; Spec wording; TestGuardCutoverAtJudgesTheCommit, incl. bad commit errors). Round 1 evidence stands: TestIssueLintIDsJudgesTheMarkerAtHead reproduces you-decide without the fix. Tracker package and LintIDs/Leftover/TrackedLegacy/Legacy/IssueMigrate/Offline green.; review verdict: SHIP
- 2026-09-27: closed — Guard reads its marker via one Repository.readMarker (checkout, or a commit via GuardCutoverAt), used by checkCutover, the absent-tracker check and Presence; lint-ids guards at --head. TestIssueLintIDsJudgesTheMarkerAtHead (subprocess) passes, and without the fix reproduces you-decide exact error; the forged-root head refuses. Tracker package and LintIDs/Leftover/TrackedLegacy/Legacy/IssueMigrate/Offline tests green. Atlas notes GuardCutoverAt.; review verdict: SHIP
- Fix: the guard reads its marker through one `Repository.readMarker`
  (checkout by default; `GuardCutoverAt(root, commit)` reads a commit's tree),
  used by `checkCutover`, the absent-tracker check and `Presence`; lint-ids
  guards at `--head`. `TestIssueLintIDsJudgesTheMarkerAtHead` (subprocess, as
  lint-ids exits itself) reproduces you-decide's exact error without the fix.
- Close round 1 SHIP; its four advisory Minors fixed after the evidence
  commit: `ReadCutoverMarkerAt` treats only a missing path as absent (an
  unresolvable commit or git failure errors); lint-ids and `migratedAlready`
  now read through it (one commit-tree marker reader); Spec wording (lint-ids
  always has a head); `TestGuardCutoverAtJudgesTheCommit` pins the commit
  guard (unmarked commit beside a marked checkout refuses, committed match
  passes, foreign root refuses, bad commit errors).
- Close round 2 SHIP; its Minor fixed: `migratedAlready` returns the root it
  read, so reconcile no longer re-reads main's marker (and no longer drops a
  parse error). One commit-marker read per decision.

## Revisions

### 2026-09-27 — Done when: the negative case

Reason: "a head without the marker in a tracked repository still refuses" was
wrong — lint-ids deliberately treats a pre-cutover head as legacy (that is
what a branch awaiting `--reconcile` looks like). The real negative case is a
head whose marker names another tracker root; Done when now says so.
