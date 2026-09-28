---
id: '000256'
status: done
started: 2026-09-27T23:03:37-07:00
created: 2026-09-27
updated: 2026-09-27
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
