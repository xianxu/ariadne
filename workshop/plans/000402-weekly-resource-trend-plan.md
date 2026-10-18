---
collaboration-mode: discussion
---

# Weekly resource trend implementation plan

> **For agentic workers:** Consult AGENTS.md §3 for execution approach. Check each step in place and commit plan changes before long waits.

**Goal:** Add a comparable, read-only resource section to #392's dated weekly brain snapshot, including data-family sizes, large Couch ledgers, suite-duration deltas, Go cache size, AgentScratch free space and weekly device gauges.

**Architecture:** #392 owns the weekly clock, versioned snapshot, fleet discovery, suite statistics and immutable write. This issue adds an inventory-backed metadata collector, an exact previous-week reader and a pure comparison renderer. It attaches one resource section to #392's snapshot command; there is no second scheduler or writer.

**Tech stack:** Go, pair's JSON source inventory (read-only), `os.ReadDir`/`os.Lstat` and `Statfs`, #392's suite statistics and brain snapshot writer.

## Dependency and decisions

#392 landed on 2026-10-18 and its code is merged into this branch. `WeeklySnapshot` in `cmd/sdlc/weeklystats.go` is schema v1 (`window`, `repos`); `runWeeklySnapshot` in `cmd/sdlc/weeklystats_cmd.go` builds the whole fleet, and `publishBrainSnapshot` in `cmd/sdlc/weeklystats_write.go` atomically refuses a different value at an existing dated path. #392 does **not** read a previous snapshot. #402 adds that read from exactly `window.End - 7 days`, validates its version and window, and extends the schema to v2 with a top-level resource section. A v1 prior snapshot has no resource baseline; it is not treated as zero. Pair's checked-in `cmd/internal/historyarchive/sources.json` lists the backup source families; read it by configurable absolute path rather than copying selectors into ariadne (ARCH-DRY). Go cache and AgentScratch are machine gauges, not backup sources.

The already published `2026-10-18.json` is immutable and remains v1. A pre-close run on this machine writes only to a temp brain: on 2026-10-18, #392's default completed Monday boundary is 2026-10-12, but resource gauges are measured on 2026-10-18, so publishing them as an Oct 12 observation would be misleading. The resource section records its actual UTC `observed_at`. The first real #402 publication uses the next completed Monday boundary, 2026-10-19; the TL runs `sdlc test-stats snapshot` by hand after 2026-10-19T00:00:00Z from the landed binary. There is no weekly scheduler yet.

The hub replaced the ops project's proposed daily points with weekly read-only device gauges: `ioreg` `Bytes (Write)` and SSD percentage used if `smartctl` can read it without root. If `smartctl` is missing or denied, record wear as unknown with the exact reason; do not install it or invoke sudo. No daily producer or inferred daily series is part of #402. The `ioreg` counter may reset at boot; identify its device and counter epoch and suppress a delta across resets rather than presenting a negative or false TB/week value.

## Core concepts

| Name | Kind | Lives in | Status |
|---|---|---|---|
| `ResourceObservation` | PURE | `cmd/sdlc/weeklyresource.go` | new |
| `ResourceDelta` | PURE | `cmd/sdlc/weeklyresource.go` | new |
| `RenderResourceSection` | PURE | `cmd/sdlc/weeklyresource.go` | new |
| `ResourceCollector` | INTEGRATION | `cmd/sdlc/weeklyresource_collect.go` | new |
| `PreviousResourceReader` | INTEGRATION | `cmd/sdlc/weeklyresource_previous.go` | new |
| `WeeklySnapshotAttachment` | INTEGRATION | `cmd/sdlc/weeklystats_cmd.go`, `cmd/sdlc/weeklystats.go` | modified |

`ResourceObservation` holds source-kind, byte count, file count, collection status and provenance; physical free/total bytes, device write counter and SSD wear are separate machine gauges. It owns no IO. `ResourceDelta` compares the same kind/path between two complete weekly snapshots, reporting signed bytes and a percentage only when the prior value is positive **and** the actual `observed_at` instants are 6–8 days apart. Outside that interval the point-in-time delta is unknown with the measured gap in the reason. First observation is a baseline, not zero growth. A missing, unreadable, or changed-definition source is unknown; it cannot trigger or clear the >25% growth flag. A previous zero with a new positive value is marked **new**, not an infinite percentage. Device-write deltas require matching device and counter epoch and a nondecreasing counter; SSD wear delta requires matching device. `RenderResourceSection` sorts stable rows, chooses the ten largest Couch ledger paths, displays current and prior sizes with deltas, and shows per-repo full-suite p50/p90 and sample counts from #392's already-computed metrics. Suite duration deltas compare fixed weekly windows and are not suppressed by a point-observation timing gap. Its pure tests need no filesystem.

`ResourceCollector` reads pair's inventory JSON and resolves its declared root tokens using explicit configuration: the single pair-data/couch-store/agent-home roots and each applicable slot-store/git-common root are separate instances. Pair defines slash-separated `path.Match` selectors; a matched directory contributes its regular-file descendants. The collector walks each resolved root once, assigns each regular file to at most one declared kind, skips symlinks, and records Lstat sizes without reading contents. Unsupported patterns, overlapping assignments, missing roots and permission errors mark affected families incomplete; it never silently substitutes zero. Go cache resolves from `go env GOCACHE` or the producer's explicit environment; AgentScratch is an explicit mount path whose `Statfs` supplies total and available bytes. A missing volume is unknown, not a full or empty volume. A bounded `ioreg` call and optional unprivileged `smartctl` call supply the device gauges; missing binaries, denied access, ambiguous device and unparsable output are unknown. The fake seam models command outputs and errors for successive calls; a live conformance probe checks parsing on this machine (ARCH-MOCK). A fixture temp tree and fake capacity/device samples exercise the adapter; the live run uses the actual roots. Expected scan work is O(number of visited filesystem entries × matching rules), not one traversal per pattern; test counters at N and 2N reject repeated whole-tree scans (ARCH-CONSTRAINTS).

`PreviousResourceReader` reads only the exact prior date under #392's brain weekly directory. It checks schema and `window.End` before accepting a prior resource section; missing, v1 or incomplete resource data creates named baseline/unknown states. A malformed file is unknown, never a zero. It is read-only and bounded in bytes; #392's immutable writer remains the only publication authority.

`WeeklySnapshotAttachment` consumes the collected observation and the prior immutable snapshot within #392's own window. It reuses `WeeklyRepoStats.Suite.Wall` for current per-repo p50/p90 and sample coverage; the resource rows add prior values and deltas, without re-aggregating attempts. It adds the resource section to #392's one writer; stable reruns of the same week must preserve the immutable result, and a failed collection must remain visible as partial coverage. No second output file or scheduling loop is introduced (ARCH-FUNERAL). The added bytes live only in the same weekly brain artifact, whose retention and backup are the brain measurement charter's responsibility.

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

- [x] Verify #392 is landed; inspect its schema, lack of previous-week lookup, immutable rerun behavior, and brain path. Record the actual files and interface change in a dated Revision here before code.
- [x] Write failing fixture tests for baseline, positive/negative deltas, prior zero, changed source definition, >25% growth, unknown coverage, ledger ordering, suite p50/p90 deltas, counter resets and unreadable wear. Run targeted `go test` and observe expected failures.
- [x] Implement the pure model/renderer and pass the targeted tests. Commit.

### 2. Collect metadata from declared sources

- [x] Write a failing temp-tree test for pair inventory patterns, symlink avoidance, duplicate assignments, missing/unreadable paths, Go cache sizing, injected AgentScratch capacity and bounded device-tool fixtures. Add N/2N visit counters.
- [x] Implement one metadata traversal per root and explicit root/mount resolution; keep per-family errors rather than discarding the whole section. Pass targeted tests and commit.

### 3. Attach, verify, and publish

- [x] Add a failing #392 producer fixture proving one weekly snapshot contains both test-economy and resource sections and that a same-week rerun is stable. Implement the attachment and pass targeted tests.
- [ ] Update atlas with the source inventory and snapshot flow, then run `make -k build`, `git diff --check`, and the relevant focused packages. Commit and push.
- [ ] Run `sdlc local-check run --name suite` on the pushed head. Run the producer into a temp brain and inspect its source coverage, largest ledgers, cache bytes, AgentScratch free/total, device gauges, `observed_at`, suite durations and comparison status. Give the exact pilot output/hash to the close reviewer; do not fabricate prior-week deltas for the first run.
- [ ] After SHIP review, open the PR and merge through the landing line. Report the result to the hub with the temp pilot path/hash and the manual first-real-publication command for the TL after 2026-10-19T00:00:00Z. The TL updates the ops project.

## Revisions

### 2026-10-17 — hub's weekly device decision

Replace the unassigned daily device-write concept with `ioreg` `Bytes (Write)` and optional unprivileged SSD percentage used in the same weekly snapshot. Preserve a counter epoch so reboots or device changes cannot produce a false weekly write delta. #392 remains the prerequisite and is this slot's next issue.

### 2026-10-18 — bind to landed #392

#392 landed at PR #297; its v1 model, fleet producer, suite wall quantiles and immutable writer are now in `weeklystats.go`, `weeklystats_cmd.go` and `weeklystats_write.go`. There is no previous-week reader, contrary to the provisional plan; #402 adds an exact-date, version/window-validating reader. The new top-level resource section requires schema v2. The published 2026-10-18 v1 pilot cannot be changed. The first real resource snapshot will use an unused completed Monday boundary after SHIP, with baseline rather than an invented delta if its exact predecessor is absent or v1. Current suite duration samples come from #392's `WeeklyRepoStats.Suite.Wall`; #402 computes only the comparison. These changes preserve one producer and one immutable writer (ARCH-DRY, ARCH-FUNERAL).

### 2026-10-18 — hub approval and device access

The hub approved the landed-schema delta. An absent or root-only `smartctl` is a named unknown wear measurement; neither a tool install nor sudo is part of this issue. `ioreg` supplies the separate write counter without that dependency.

### 2026-10-18 — point-observation time after live pilot

The temp-brain pilot showed the resource point measurements were taken on Oct 18 while #392's default completed week ended Oct 12. The hub approved an explicit `observed_at`, a 6–8-day interval guard for numeric point-resource deltas (actual gap printed on refusal), and no backdated real publication. The Oct 12 pilot remains temporary. The first real v2 resource snapshot is the completed 2026-10-19 week, run manually by the TL after 00:00 UTC; the next completed weekly snapshot supplies the first possible seven-day resource delta. The already published Oct 18 v1 artifact remains immutable.
