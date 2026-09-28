# Boundary Review — ariadne#255 (whole-issue close)

| field | value |
|-------|-------|
| issue | 255 — Fleet cutover to the issue tracker; delete the legacy writers |
| repo | ariadne |
| issue file | workshop/issues/000255-fleet-tracker-cutover.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8b19a7afd70579fb4c71ea5dfe713379f1ccf1c3..07873a74cf1f9af0f87fadb30e823f89af93153f |
| command | sdlc close --issue 255 |
| reviewer | claude |
| timestamp | 2026-09-28T09:16:13-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The window has 209 files, but nearly all of it belongs to other issues. It starts before #252 landed, so it includes the code from #252 (tracker, migration, legacy mode), #256 (archived-on-branch refusal) and #257 (lint-ids reads the marker at `--head`). Each of those was reviewed at its own close. #255's own 13 commits only change `workshop/issues/000255-fleet-tracker-cutover.md` and `workshop/plans/000255-fleet-tracker-cutover-checklist.md`. That fits the Spec: the work was operational, carried out in the peer repos. I checked the main claim directly. All 15 listed work repos have `workshop/issue-tracker.json` on `origin/main` and an `origin/issue-tracker` branch: 42shots, ariadne, astro, ducks, kaggle, kbench, metis, nous, pair, parley.nvim, parli, tools, xianxu.dev, you-decide and robotics. The three brains have neither, which the charter requires. Phase D was spun off properly to #258, with the full removal list and a `## Revisions` entry. Nothing blocks this close. Three cheap record fixes remain:
- Two ariadne-layer directories in `~/workspace` are neither cut over nor listed as excluded.
- The Spec asked for a per-repo record in the Log, and most of it is missing.
- The durable checklist was left almost entirely unticked, so it contradicts the issue's Plan.

1. **Strengths**
   - I confirmed the fleet state independently from each peer's `origin/main` and `origin/issue-tracker`; it matches the Log's sweep line.
   - The Phase D deferral is done right. The Done-when bullet is struck and pointed at #258, the Revisions entry gives a reason and the change, and #258 carries the complete removal list plus a "confirm nothing takes a legacy path" check before deleting.
   - Bugs found during the cutover went into real fixes with tests rather than workarounds: the archived-on-branch gap became #256, and you-decide's lint-ids hook became #257. The one-time hook skip is logged as operator-authorized.
   - Each repo's blockers are recorded with how they were cleared: the `origin/feature` deletion, nous #48 landing, and tools publishing #82 before migrating.

2. **Critical:** none.

3. **Important**
   - **Done-when 1 is not fully met as written.** `~/workspace/metis.bak` and `~/workspace/test-repo` are both ariadne-layered (they have `construct/`, `workshop/` and `Makefile.workflow`). They have no remote, no issues and no marker, and the Log doesn't list them as excluded. The Done-when requires "cut over … or recorded here as deliberately excluded". Fix: add one Log line excluding them (a backup and a scratch repo), or delete them.
   - **The durable checklist contradicts the issue's Plan.** Nearly every per-repo row in `workshop/plans/000255-fleet-tracker-cutover-checklist.md` is still `[ ]` with blank digests: C-sections at lines 242–305, C12 (astro/parli, "decide") at line 308, and Phase D at lines 315–319. The issue ticks every row. This file gets archived with the issue as the record. Fix: tick what was done, say rows are "superseded — see #255 Log", and mark Phase D "→ #258".

4. **Minor**
   - **The Spec's per-repo record is only partly delivered.** The Spec asks for dry-run counts, the apply digest and verification for each repo. Only ariadne has the full set (digest `f810548c…`, cards, inferences, duplicates). Batch 1 and batch 2 give tracker roots and "0 refusals", with no digests and no card or inference counts (pair's card counts are the exception). If the digests still exist in each migration commit or terminal history, add them; otherwise note that roots were recorded instead.
   - **The sweep line lists `sdlc-smoke` as on the tracker**, but a later entry says it was deleted, and it's gone from disk. Say "(since deleted)".
   - **The Spec section still describes Phase D** as in scope. The Revisions entry covers it, so this is acceptable, but a one-word "(→ #258)" in the Spec would stop anyone misreading it.
   - **#119's plan was deleted outright**, not archived. It happened in the archive commit e23d2b74 when #119 was abandoned (wontfix). I only saw that commit's subject and didn't check whether a copy was kept in `workshop/history/`.

5. **Test coverage notes**
   - There is no new executable surface attributable to #255. The bugs the cutover exposed each got a regression test in their own issue (#256, #257), which is the right place.
   - A fleet-wide "already migrated" dry-run sweep would be the strongest proof, but it would need `sdlc` run in each peer. My marker-and-branch check is a close enough stand-in.

6. **Architectural notes**
   - **ARCH-DRY: pass.** No code.
   - **ARCH-PURE: pass.** No code.
   - **ARCH-PURPOSE: pass.** The purpose is the fleet cutover, and it's done and verified. Deferring Phase D is an operator decision to keep a fallback, not dropping the point of the issue, and #258 owns it explicitly.
   - **ARCH-MOCK: N/A.** There is no new external seam.
   - **ARCH-CONSTRAINTS: N/A.** This is an operational migration.
   - **ARCH-SECURE: pass.** New private remote `xianxu/robotics`; the hook bypass was one-time and logged.
   - **ARCH-ORDER: pass.** The freeze → dry run → apply → verify → unfreeze order was followed. The one interrupted apply (you-decide) was recovered by re-running the same idempotent `--apply`.
   - **ARCH-FUNERAL: flag (Minor).** Legacy mode is now residue with a named end (#258). But #258 is `punt` with no trigger or date, and "a while" never becomes a date. Give #258 a concrete trigger, e.g. "after N weeks with no legacy-path hits".
   - Leftover pre-cutover branches (kbench `000028-turn-level-conductor`, pair #292) must run `--reconcile` before reuse. That's handled by tooling, and the Log records it.

7. **Plan revision recommendations**
   - Add a Revisions or Log entry to the checklist plan: "per-repo rows superseded by #255's Log; C12 decided: cut over (operator); Phase D → #258".
   - Add a Log line in #255: "`metis.bak` and `test-repo` excluded (a backup and a scratch repo, no remote, no issues)", or delete the directories.

```findings
findings:
  - id: new
    severity: Important
    family: done-when-enumeration-incomplete
    title: |
      metis.bak and test-repo are ariadne-layered, not cut over, and not recorded as excluded
    detail: |
      Done-when 1 requires every ariadne-layer repo in ~/workspace to be cut over or recorded as deliberately excluded. Both have construct/ and workshop/ but no marker and no remote. Add a Log exclusion line or delete them.
  - id: new
    severity: Important
    family: durable-artifact-drift
    title: |
      Cutover checklist plan left unticked; contradicts the issue Plan it is archived with
    detail: |
      workshop/plans/000255-fleet-tracker-cutover-checklist.md lines 242-319 still show every C-row, C12 decide and Phase D as open with blank digests. Tick them, mark them superseded by the Log, and point Phase D at 258.
  - id: new
    severity: Minor
    family: spec-record-incomplete
    title: |
      Per-repo Log lacks the dry-run counts and apply digests the Spec asked for (only ariadne has them)
  - id: new
    severity: Minor
    family: durable-artifact-drift
    title: |
      Sweep line lists sdlc-smoke as on the tracker although it was later deleted
  - id: new
    severity: Minor
    family: lifecycle-end-unnamed
    title: |
      Legacy-mode removal in 258 is punted with no trigger or date for "a while"
```
