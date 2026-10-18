---
id: 000402
status: working
deps: [ariadne#392]
github_issue:
created: 2026-10-15
updated: 2026-10-18
estimate_hours:
card_mirror: '0b20b0481a39833a19da352924f28cbfb5881d85' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T21:36:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "5a306d96", done: "d65ad378"}
---

# Weekly resource trend: sizes per data family and suite durations, with deltas, in #392's snapshot

## Problem

Slow growth never feels urgent day to day (21 MB Couch ledgers, 64 GB of captures, 54.8 GB of event logs). Ops project
resource-observability wants a weekly trend so growth is visible as a delta.

## Spec

- Extend the weekly snapshot that ariadne#392 builds for test stats (same schedule, same brain destination, same
  format conventions) with a resource section: sizes per directory family (pair `repos/*` by artifact family,
  `couch/`, agent transcript dirs), the largest Couch ledgers, and full-suite durations per repo, each with its delta
  from the previous week.
- Read-only and cheap: `stat`-based sizes, no file contents. Sources come from one declared list (shared with ops
  project history-backup's source inventory once that exists).
- Flag any family that grew by more than a configurable share (default 25%) week on week.

## Done when

- A fixture-backed test produces the section and its deltas; a live run on this machine writes and inspects one temp-brain snapshot. The first permanent resource snapshot is the completed 2026-10-19 week, run by the TL by hand after 00:00 UTC from the landed binary.

Depends on ariadne#392.

## Plan

- [x] After #392 lands, bind its weekly window, immutable writer, previous-snapshot reader and brain destination to the resource section in `workshop/plans/000402-weekly-resource-trend-plan.md`.
- [x] Test the inventory-backed, read-only collector and pure delta/alert renderer with complete, missing and first-week fixtures; include Go cache, AgentScratch capacity and largest ledgers.
- [ ] Integrate the section into #392's snapshot, run focused tests and the recorded suite, then write and inspect one live temp-brain snapshot before close.

## Log

- 2026-10-18: Pre-suite review found that two root tokens resolving to one physical file could leave a prior family counted when a later token matched two kinds. A red fixture reproduced it; the collector now removes the earlier assignment and marks all claimants unknown. I canceled only this checkout's queued suite on the superseded head and will rerun it on the corrected commit.
- 2026-10-18: Post-fix live temp-brain snapshot at `/var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne402-observed.yXKoFD/memory/work/devinfra/test-economy/weekly/2026-10-12.json` is schema v2, 41,887 bytes, sha256 `b4d1c143405bb2f8d549b9114173dd23c0f1135b6b91b4eb19ffa53f930b1681`. The suite window is Oct 5–12, while `resources.observed_at` is 2026-10-18T09:34:28.073975Z. All 17 families have complete metadata; ten ledger rows and 16 repo suite rows appear; Go cache is 155,134,309,459 bytes, AgentScratch free/total is 39,597,510,656/200,000,000,000 bytes, disk0 writes 35,779,148,959,744 bytes and wear 8%. Both device deltas correctly say baseline. `make -k build`, `make -k sdlc-build`, focused tests and diff check passed. The pilot is temporary; the TL runs the first permanent resource snapshot after Oct 19 00:00 UTC.
- 2026-10-18: A red test showed the historical pilot lacked a point-observation time. Schema v2 now records `observed_at` and suppresses family, ledger and machine deltas outside a measured 6–8-day observation gap; suite-duration comparisons remain tied to exact weekly windows. Focused temporal tests pass, including both interval boundaries and a predecessor missing `observed_at`.
- 2026-10-18: Live temp-brain pilot at `/var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne402-pilot.TIwyUv/memory/work/devinfra/test-economy/weekly/2026-10-12.json` produced schema v2 (41,837 bytes; sha256 `d2b9fca6dcad837c71c76075ac2c9ce39c3f7f2fee527fb3c947db60baa8fb5a`). All 17 declared families were complete; the largest ten ledgers, 154,897,415,411-byte Go cache, 39,613,689,856-byte AgentScratch free gauge, disk0 write counter and 8% wear were present. The date is the historical suite window; the resource gauge was measured today, so the pilot is not published to the real brain.
- 2026-10-18: #392 producer now attaches schema-v2 resources through its one immutable writer. Fake-backed machine probing, prior-week reader, root resolution (global Pair store despite slot-specific PAIR_DATA_DIR), and command fixture passed focused tests. A v2 predecessor with a missing machine section is unknown, not an invented baseline. Build, live temp-brain check, suite and close remain.
- 2026-10-18: TDD slices for exact prior-week reading, byte comparison, declared-source collection, ledger ranking, suite duration comparison and device parsing are green in focused cmd/sdlc tests. A live read-only probe found ioreg Bytes (Write) for internal disk0; smartctl prints Percentage Used without root but exits 4 after an unrelated error-log query, so usable wear output must be parsed independently of that exit code. The bounded ioreg form is `-r -c IOBlockStorageDriver -w 0 -l -d 2`.

- 2026-10-17: unclaimed: Plan approved by hub, including weekly device gauges. Waiting for prerequisite ariadne#392 to land; ariadne:1 takes #392 first, then resumes this branch and binds the landed snapshot interface before code.
### 2026-10-15

2026-10-17: Claimed and entered planning. #392 is still open and unclaimed, with no snapshot producer, schema or weekly writer on main; #402 cannot safely bind to it yet. Pair's history-backup source inventory now exists at `cmd/internal/historyarchive/sources.json`; use its declared data families instead of duplicating source patterns (ARCH-DRY). The hub additionally requested Go build cache size and AgentScratch free space. Those are machine gauges, not backup source families.

## Revisions

### 2026-10-17 — hub dispatch and dependency survey

Add Go build cache size and AgentScratch free/total space to the planned weekly resource section. Treat #392's landed snapshot interface as a prerequisite for implementation and a live write. Keep family sizes sourced from pair's existing backup inventory; machine gauges have separate explicit paths.

### 2026-10-17 — weekly device gauges replace daily points

The hub approved the plan, directed this slot to land #392 first, and replaced the ops project's unassigned daily device-write points with two weekly gauges: the `ioreg` `Bytes (Write)` counter and SSD percentage used when readable without root. Compare weekly observations only when the device and counter epoch remain comparable; a reset yields an unknown delta. The TL owns the resource-observability project Log and will record the daily-to-weekly replacement there.

### 2026-10-18 — re-plan on #392's landed schema

#392's v1 snapshot, producer and immutable writer are now on this branch. It has no previous-week reader, so #402 must add the exact-date, bounded reader itself and issue schema v2. The already published 2026-10-18 v1 pilot remains untouched; the first real resource snapshot uses an unused completed Monday boundary after SHIP review. Current suite wall statistics come from #392's existing per-repo model, while #402 computes the previous-week comparison. The durable plan names the actual files and failure states.

### 2026-10-18 — hub approval

The hub approved the schema-binding revision. If `smartctl` is absent or requires root, SSD wear is unknown with the reason; no install or sudo. `ioreg` writes remain a separate unprivileged gauge.

### 2026-10-18 — point-observation time and first real run

The hub approved recording `observed_at` and allowing a numeric point-resource delta only for observations 6–8 days apart, with the actual gap in an unknown reason otherwise. The Oct 12 run remains a temp-brain probe; the first real resource snapshot is the completed 2026-10-19 week. The TL will run `sdlc test-stats snapshot` by hand from the landed binary after 2026-10-19T00:00:00Z, since no weekly scheduler exists. Suite-duration deltas still use the exact historical windows.
