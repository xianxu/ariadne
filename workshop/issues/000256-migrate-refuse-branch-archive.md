---
id: 000256
status: working
deps: [000255]
github_issue:
created: 2026-09-27
updated: 2026-09-27
estimate_hours:
started: 2026-09-27T23:03:37-07:00
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

- [ ] Refusal in `PlanTrackerMigration` for a deleting branch edit when main has the issue active
- [ ] Unit test in `migration_test.go`, plus a dry-run e2e in `issuemigrate_test.go`
- [ ] Record in #255
