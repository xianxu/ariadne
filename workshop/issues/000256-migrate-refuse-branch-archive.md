---
id: 000256
status: codecomplete
deps: [000255]
github_issue:
created: 2026-09-27
updated: 2026-09-27
estimate_hours:
started: 2026-09-27T23:03:37-07:00
flow: {kind: quick, provenance: inferred, spec: "dbdf0577", done: "24e209b3"}
actual_hours: 0.15
---

# issue migrate: refuse a branch that archives an issue main still has active

## Problem

`PlanTrackerMigration` skips any branch edit that deletes an issue file, as "a
branch's own archive move". But a branch can archive an issue that never landed:
nous's `000048-…` branch (June) closed #48 and archived it as `done` while nous's
main still said `working`, and the dry run reported 0 refusals. Had nous cut over
then, the card would have stayed `working` forever, and landing the branch would
have modify/delete-conflicted with the converted details. (#48 has since landed;
the gap is in the tool.)

## Spec

A branch that deletes an issue's details file while main still has that issue
active is refused, like the other branch refusals: "a branch archives an issue
main still has active", next: "land or drop the branch before cutover". A branch
deleting a file main has already archived (its own landed archive move) stays
silent, as today. Grouped per (path, reason) across carrying branches, like the
existing refusals.

## Done when

- The dry run refuses an unlanded branch that archives an issue main still has
  active, naming the branch; a landed archive move is still not refused.
- A regression test covers both cases and fails without the fix.

## Plan

- [x] Refusal in `PlanTrackerMigration` for a deleting branch edit when main has the issue active
- [x] Unit test in `migration_test.go`, plus a dry-run e2e in `issuemigrate_test.go`
- [x] Record in #255
- The existing unit test encoded the gap (`feat-e`: "a deletion: the branch's
  own archive move" of an issue main still had active). Before cutover a legacy
  archive happens on main at merge, so such a deletion is an unlanded close.
  Kept silent: a rename (the branch keeps other active details for the ID) and
  a deletion main made too. Both tests fail without the fix.
- Real-repo dry runs with the fix: pair, kaggle, kbench 0 refusals; ariadne's
  2 are pre-existing (`origin/feature` edits archived #148, #159).
- Close review round 1 (FIX-THEN-SHIP), BR-1 Important: the post-cutover
  `--reconcile` path made the same "own archive move" assumption, so a branch
  the dry run never saw could archive an open issue silently. One shared rule
  now (`tracker.RemovalArchivesActive` + `ArchivesActiveReason`): the dry run
  asks whether main still has the issue active, reconcile whether its card is
  terminal (`vocab.Issue().IsTerminal`); both keep renames silent.
  `TestIssueMigrateReconcileRefusesAnUnlandedArchive` fails without the fix.

## Log

- 2026-09-27: closed — Refusal for a branch that archives an issue main still has active, one shared rule (tracker.RemovalArchivesActive) used by the dry run and --reconcile; renames and issues main closed stay silent. TestPlanTrackerMigrationRefusesLegacyDivergence, TestIssueMigrateRefusesABranchThatArchivesAnActiveIssue and TestIssueMigrateReconcileRefusesAnUnlandedArchive each fail without the fix; tracker package and all TestIssueMigrate* green; real dry runs: pair/kaggle/kbench 0 refusals, ariadne 2 pre-existing (origin/feature). Atlas migration step 4 names the refusal.; review verdict: SHIP
- Close round 2 SHIP; its two advisory Minors fixed in the close commit:
  reconcile's closed-card silent path now has a test
  (`TestIssueMigrateReconcileAllowsAnArchiveMainClosedToo`, fails when the
  terminal check is forced false), and an unreadable card refuses with its
  parse error instead of claiming the issue is still open.
