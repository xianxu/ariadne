---
collaboration-mode: discussion
---

# Weekly resource trend implementation plan

> **For agentic workers:** Consult AGENTS.md §3 for execution approach. Check each step in place and commit plan changes before long waits.

**Goal:** Add a comparable, read-only resource section to #392's dated weekly brain snapshot, including data-family sizes, large Couch ledgers, suite-duration deltas, Go cache size, AgentScratch free space and weekly device gauges.

**Architecture:** #392 owns the weekly clock, snapshot schema, previous-week selection and immutable write. This issue adds an inventory-backed observation adapter and a pure delta renderer; the adapter reads metadata only. It attaches observations to #392's snapshot after that issue lands, using its actual extension seam and format rather than creating a second scheduler or writer.

**Tech stack:** Go, pair's JSON source inventory (read-only), `os.ReadDir`/`os.Lstat` and `Statfs`, #392's suite statistics and brain snapshot writer.

## Dependency and decisions

#392 is open and unclaimed as of 2026-10-17; it has no implementation on main. No #402 code or live snapshot can be completed until its schema and writer land. The first task after #392 is to replace the tentative file paths below with its actual extension seam and record that binding in a Revision. Pair's checked-in `cmd/internal/historyarchive/sources.json` already lists the backup source families; read it by configurable absolute path rather than copying selectors into ariadne. Its source list is the sole declaration for data families (ARCH-DRY). Go cache and AgentScratch are machine gauges, not backup sources.

The hub replaced the ops project's proposed daily points with weekly read-only device gauges: `ioreg` `Bytes (Write)` and SSD percentage used if `smartctl` can read it without root. No daily producer or inferred daily series is part of #402. The `ioreg` counter may reset at boot; identify its device and counter epoch and suppress a delta across resets rather than presenting a negative or false TB/week value.

## Core concepts

| Name | Kind | Lives in | Status |
|---|---|---|---|
| `ResourceObservation` | PURE | `cmd/sdlc/weeklyresource.go` | new |
| `ResourceDelta` | PURE | `cmd/sdlc/weeklyresource.go` | new |
| `RenderResourceSection` | PURE | `cmd/sdlc/weeklyresource.go` | new |
| `ResourceCollector` | INTEGRATION | `cmd/sdlc/weeklyresource_collect.go` | new |
| `WeeklySnapshotAttachment` | INTEGRATION | #392's landed producer file, to be named in a Revision | modified |

`ResourceObservation` holds source-kind, byte count, file count, collection status and provenance; physical free/total bytes, device write counter and SSD wear are separate machine gauges. It owns no IO. `ResourceDelta` compares the same kind/path between two complete weekly snapshots, reporting signed bytes and a percentage only when the prior value is positive. First observation is a baseline, not zero growth. A missing, unreadable, or changed-definition source is unknown; it cannot trigger or clear the >25% growth flag. A previous zero with a new positive value is marked **new**, not an infinite percentage. Device-write deltas require matching device and counter epoch and a nondecreasing counter; SSD wear delta requires matching device. `RenderResourceSection` sorts stable rows, chooses the ten largest Couch ledger paths, displays current and prior sizes with deltas, and shows per-repo full-suite p50/p90 and sample counts from #392's already-computed metrics. Its pure tests need no filesystem.

`ResourceCollector` reads pair's inventory JSON and resolves the declared root tokens using explicit configuration. It uses a single metadata traversal per root, assigns each regular file to at most one declared source kind, skips symlinks, and records Lstat sizes without reading contents. Unsupported patterns, overlapping assignments, missing roots and permission errors mark affected families incomplete; it never silently substitutes zero. Go cache resolves from `go env GOCACHE` or the producer's explicit environment; AgentScratch is an explicit mount path whose `Statfs` supplies total and available bytes. A missing volume is unknown, not a full or empty volume. A bounded `ioreg` call and optional unprivileged `smartctl` call supply the device gauges; missing binaries, denied access, ambiguous device and unparsable output are unknown. A fixture temp tree and fake capacity/device samples exercise the adapter; the live run uses the actual roots. Expected scan work is O(number of visited filesystem entries + inventory matches); test counters at N and 2N reject repeated whole-tree scans (ARCH-CONSTRAINTS).

`WeeklySnapshotAttachment` consumes the collected observation and the prior immutable snapshot within #392's own window. It adds the resource section to that one writer; stable reruns of the same week must preserve the immutable result, and a failed collection must remain visible as partial coverage. No second output file or scheduling loop is introduced (ARCH-FUNERAL).

## Visible states and recovery

Actions: fix source configuration/permissions, retry the same week before publication, or collect the next week. A published immutable snapshot is not silently rewritten.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Complete observation, comparable prior | complete | collect next week | none |
| Complete observation, no prior | baseline | collect next week | first-week delta unavailable |
| Missing inventory/root, unreadable path, missing mount, or malformed prior | partial/unknown with named source | fix source and retry before publication | that week's comparison if already published |
| Writer fails before #392 publication | no snapshot, explicit error | retry same week | none if source data remains |

The TL should tell the operator that a published partial week stays partial and the following week may lack a valid delta for that family (ARCH-STATE).

## Tasks

### 1. Bind to #392 and pin semantics

- [ ] Verify #392 is landed; inspect its schema, previous-week lookup, immutable rerun behavior, and brain path. Record the actual files and any interface change in a dated Revision here before code.
- [ ] Write a failing fixture test for baseline, positive/negative deltas, prior zero, changed source definition, >25% growth, unknown coverage, ledger top ten, suite p50/p90 deltas, counter resets and unreadable wear. Run its targeted `go test` and observe the expected failures.
- [ ] Implement the pure model/renderer and pass the targeted test. Commit.

### 2. Collect metadata from declared sources

- [ ] Write a failing temp-tree test for pair inventory patterns, symlink avoidance, duplicate assignments, missing/unreadable paths, Go cache sizing, injected AgentScratch capacity and bounded device-tool fixtures. Add N/2N visit counters.
- [ ] Implement one bounded metadata traversal and explicit root/mount resolution; keep per-family errors rather than discarding the whole section. Pass targeted tests and commit.

### 3. Attach, verify, and publish

- [ ] Add a failing #392 producer fixture proving one weekly snapshot contains both test-economy and resource sections and that a same-week rerun is stable. Implement the attachment and pass targeted tests.
- [ ] Update atlas with the source inventory and snapshot flow, then run `make -k build`, `git diff --check`, and the relevant focused packages. Commit and push.
- [ ] Run `sdlc local-check run --name suite` on the pushed head. Run the producer live once and inspect the brain snapshot's source coverage, largest ledgers, cache bytes, AgentScratch free/total, device gauges, suite durations and comparison status. Do not fabricate prior-week deltas for the first run.

## Revisions

### 2026-10-17 — hub's weekly device decision

Replace the unassigned daily device-write concept with `ioreg` `Bytes (Write)` and optional unprivileged SSD percentage used in the same weekly snapshot. Preserve a counter epoch so reboots or device changes cannot produce a false weekly write delta. #392 remains the prerequisite and is this slot's next issue.
- [ ] Record the live path and evidence in the issue Log; close with the required review, then PR/merge through the landing line. Report the result to the hub; the TL updates the ops project.
