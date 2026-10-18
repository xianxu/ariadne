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

- A fixture-backed test produces the section and its deltas; a live run on this machine writes one snapshot.

Depends on ariadne#392.

## Plan

- [x] After #392 lands, bind its weekly window, immutable writer, previous-snapshot reader and brain destination to the resource section in `workshop/plans/000402-weekly-resource-trend-plan.md`.
- [ ] Test the inventory-backed, read-only collector and pure delta/alert renderer with complete, missing and first-week fixtures; include Go cache, AgentScratch capacity and largest ledgers.
- [ ] Integrate the section into #392's snapshot, run focused tests and the recorded suite, then write and inspect one live snapshot before close.

## Log


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
