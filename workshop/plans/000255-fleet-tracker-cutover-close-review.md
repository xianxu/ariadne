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

---

## Re-review — 2026-09-28T09:18:30-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 255 — Fleet cutover to the issue tracker; delete the legacy writers |
| repo | ariadne |
| issue file | workshop/issues/000255-fleet-tracker-cutover.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8b19a7afd70579fb4c71ea5dfe713379f1ccf1c3..3ce173c5d70da69010df6304187bc43085945252 |
| command | sdlc close --issue 255 |
| reviewer | claude |
| timestamp | 2026-09-28T09:18:30-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This close boundary for #255 is operational work plus its record. The large `cmd/sdlc` part of the window is #252's tracker code, which reached SHIP through its own 13 review rounds. #255 adds cutover records (Log, checklist, #258 spin-off) on top of it. I checked the Log against `~/workspace` itself, not the prose. Every directory with `construct/` or `workshop/` is in one of three states:
- **On the tracker** (`workshop/issue-tracker.json` present): 42shots, ariadne, astro, ducks, kaggle, kbench, metis, nous, pair, parley.nvim, parli, robotics, tools, xianxu.dev, you-decide.
- **A brain** (`.brain/` present): brain, brain-family, brain-private.
- **Recorded as excluded**: `metis.bak` and `test-repo`, each a single commit with no remote and no issues.

`sdlc-smoke` is gone from disk, and `sdlc issue migrate` in robotics (the latest cutover) reports *already migrated*. All five prior findings are fixed and nothing blocks the close.

1. **Strengths**
   - The per-repository results table in the #255 Log gives dry-run counts, the apply digest and the tracker root for all 15 repositories, which is the record the Spec asked for.
   - The checklist now has a **Result:** line per C-row with the same digest and root as the Log table, so the plan and the issue agree.
   - Moving Phase D to #258 is logged as a `## Revisions` entry that says why and what changed, and Done-when 3 is struck through with a pointer to #258 rather than deleted.
   - The one live check that wasn't run (a tracker-mode `milestone-close`) is marked `[skipped]` and says what covers it, rather than ticked.
   - The exclusions give a real reason ("a tracker needs a publication remote"), which matches what I found on disk.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - `workshop/plans/000255-fleet-tracker-cutover-checklist.md` C11: step 3 is ticked while its digest is still `______`. The **Result:** line says the row was superseded by A4, so this is cosmetic. Striking the step, or writing "n/a (A4)", would read better.
   - Same file, C12: its **Result:** line sits after the `---` separator, so it looks detached from its row. Formatting only.
   - I raised neither as a finding. Neither belongs to an open family and neither is worth a follow-up.

5. **Test coverage notes**
   - #255 adds no code. The tooling bugs the cutover found were each fixed with regression tests in their own issues:
     - #256: refuse a branch that archives an issue main still has active.
     - #257: lint-ids judges the cutover marker at `--head`, not the checkout.
   - The live tracker-mode `milestone-close` path is covered only by #252's e2e tests, not by a real-repository run. That is acceptable, and it is recorded.

6. **Architectural notes** (on #255's delta)
   - **ARCH-DRY:** pass. The per-repository numbers appear in both the Log table and the checklist Result lines. That is deliberate: one is the record, the other is the plan's closure. The values match.
   - **ARCH-PURE:** pass. #255 adds no code.
   - **ARCH-PURPOSE:** pass.
     - Shadow-sweep: every ariadne-layer repository in the workspace is on the tracker, a brain, or recorded as excluded, and I confirmed this on disk.
     - Moving Phase D out is not the easy-subset failure. The issue's purpose (the fleet cutover) is delivered, and the operator made keeping legacy mode as a fallback an explicit decision.
   - **ARCH-MOCK:** N/A. There are no new external calls; the live GitHub checks were A1 and A2 in #252.
   - **ARCH-CONSTRAINTS:** N/A. No runtime change.
   - **ARCH-SECURE:** pass.
     - robotics was created as a private repository.
     - Brains stay on their encrypted remotes and were not touched.
     - The one hooks-skipped `--apply` was operator-authorized and fixed properly in #257.
   - **ARCH-ORDER:** pass. Leftover pre-cutover branches (kbench `000028`, pair #292) are sent to `--reconcile` on next use rather than left in an undefined state.
   - **ARCH-FUNERAL:** pass.
     - Legacy mode's end is now named: #258, revisit by 2026-11-30 or sooner if an sdlc change touches a legacy path.
     - The `sdlc-smoke` sandbox was removed.
     - robotics' brain scaffolding was dropped.
   - For #258: before deleting the legacy writers, confirm the three brains never go down a legacy writer path. Brain repositories are refused by the spine guard, but read paths such as `estimate-source` still run there.

7. **Plan revision recommendations:** none. The Revisions entry already records the Phase D move.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Log "Excluded" bullet and checklist Excluded section name metis.bak and test-repo; verified on disk (1 commit, no remote, 0 issues each).
  - id: BR-2
    disposition: addressed
    note: |
      Checklist C1-C13 rows ticked with digest/root Result lines, C11 marked superseded by A4, C12 decided, Phase D rows point at 258.
  - id: BR-3
    disposition: addressed
    note: |
      Per-repository results table in the 255 Log carries counts, digest and tracker root for all 15 repos.
  - id: BR-4
    disposition: addressed
    note: |
      Appended Excluded bullet corrects the sweep line; sdlc-smoke confirmed deleted from the workspace.
  - id: BR-5
    disposition: addressed
    note: |
      258 Log now says revisit by 2026-11-30 or when an sdlc change next touches a legacy path.
```
