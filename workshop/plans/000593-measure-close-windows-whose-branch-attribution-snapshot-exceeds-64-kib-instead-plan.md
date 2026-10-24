---
collaboration-mode: discussion
---

# Measure big close windows: a compact v4 branch proof Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A close whose measurement window is within the 128-source bound measures and retains its snapshot, rather than failing late inside MarshalJSON at 64 KiB.

**Architecture:** Measurement inputs version 4 keeps the in-memory `MeasurementInputs` identical, but encodes branch evidence compactly. An interval names its source by index into `sources`, and inherits its record's branch. The snapshot's byte bound becomes version-aware: 64 KiB through v3, 128 KiB for v4. It is enforced with a counted error in `Validate`. The codec's structural guard stays as a separate, larger cap.

**Tech Stack:** Go, `cmd/sdlc/internal/gatestate` (measurement codec), `cmd/sdlc/actualinputs.go` (the writer).

---

## Background (measured on #586's window, 2026-10-24)

- The snapshot is retained as `actual_inputs` in the issue's gate ledger (YAML in a committed file). Every reader checks the JSON byte bound on decode (`checkMeasurementJSON`), and `Validate` checks it again.
- #586's snapshot was 67,203 bytes against `SnapshotLimitBytes` = 65,536:
  - `sources`: 75 entries, 35,090 bytes.
  - `branches`: 1 record with 150 intervals (75 `branch`, 75 `ownership-only`), 30,643 bytes. Each interval repeats `claude/<uuid>` (43 characters), two RFC 3339 times, and the 112-character issue branch name.
- Before #580 the same window measured 1.25h. #580 roughly doubled the cost per source, so the 64 KiB byte bound now binds at about 75 sources, well before `SnapshotLimitSources` (128).
- The failure is late. `encodeMeasurementObject` runs `checkMeasurementJSON` inside `MarshalJSON`, so `Validate`'s `json.Marshal(m)` fails with "does not serialize: … exceeds 65536 bytes" before its own counted check is reached.

## Decisions (TL-approved 2026-10-24)

- **D1: compact branch proof, measurement version 4.**
  - Writers set `Version: 4` for every branch method. The four branch method names are unchanged: version names the encoding, and method names the interpretation. `Validate` accepts branch methods at version 3 or 4, and v3 snapshots replay unchanged.
  - On the v4 wire, an interval carries either `source_index` (its position in `sources`, by `harness/id`) or, when that source is not in the list, an explicit `source`.
  - An interval omits `branch` when it equals its record's `branch`. A present `branch` (even `""`) is literal.
  - Decoding expands each interval to exactly the in-memory interval the writer had, so the engine, `validateOwnedBranchProof` and the warnings see identical evidence.
  - Unknown members of the record and the interval survive decode and rewrite (ARCH-EVOLVE).
- **D2: a version-aware byte bound.**
  - `SnapshotLimitBytes` stays at 64 KiB for v1 to v3. The new `SnapshotLimitBytesV4` is 128 KiB, so the 128-source bound is the one that binds.
  - The codec's structural guard (`checkMeasurementJSON`, the YAML budget) moves to `measurementStructuralLimit` = 512 KiB. It is a parse-cost bound, not the snapshot bound.
  - The snapshot bound is enforced on the whole `MeasurementInputs`, in `Validate` and at the end of its decode, against its version's limit. A v3 snapshot over 64 KiB still refuses to decode, as today.
- **D3: a counted refusal.** `Validate` reports an oversized snapshot as "measurement snapshot is N bytes, over the version V bound of B (S sources, R branch records, I branch intervals)". It never reports the generic MarshalJSON text. Close's existing recovery line (`--no-actual` with the evidence) follows it unchanged.

Rejected alternatives:
- Raising the bound alone: committed ledgers would carry about 110 KB of repeated strings per close.
- Compaction alone: still fails near 128 sources.
- Dropping or capping branch intervals: the measurement would no longer be replay-exact.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `MeasurementInputs` (version 4) | `cmd/sdlc/internal/gatestate/measurement.go` | modified |
| `measurementBranchEvidenceV4` / `measurementBranchIntervalV4` (wire) | `cmd/sdlc/internal/gatestate/measurement_branch_v4.go` | new |
| `compactBranches` / `expandBranches` | `cmd/sdlc/internal/gatestate/measurement_branch_v4.go` | new |
| `measurementSnapshotLimit(version)` | `cmd/sdlc/internal/gatestate/measurement.go` | new |

- **v4 wire types.** These exist only between `MeasurementInputs`' codec and the bytes. Each carries its own `extra` and has JSON and YAML methods that reuse `decodeMeasurementObject` / `encodeMeasurementObject` / `decodeMeasurementYAML` / `encodeMeasurementYAML`.
  - **Relationships:** one v4 record per in-memory `MeasurementBranchEvidence`, and one v4 interval per interval (1:1, so no merging).
  - **DRY rationale:** they reuse the generic codec helpers. There is no second copy of `MeasurementInputs`' 20 fields, because only the `branches` member is swapped (see the integration point below).
  - **Future extensions:** another interned field (a time base) widens the same wire types.
- **`compactBranches(sources, branches)` / `expandBranches(sources, wire)`.** These are pure, mutually inverse functions.
  - `expand(compact(x)) == x` for every valid x, including unknown members. A test covers this.
  - `expand` refuses an out-of-range index, an interval with both or neither of `source` and `source_index`, and a negative index.
- **`measurementSnapshotLimit(version)`.** It returns 128 KiB for v4 and 64 KiB otherwise. `Validate` and the top-level decode both use it.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `MeasurementInputs` JSON/YAML codec | `cmd/sdlc/internal/gatestate/measurement_codec.go` | modified | the ledger's persisted bytes |
| snapshot writer | `cmd/sdlc/actualinputs.go` | modified | measurement construction |

- **Codec.** For v4, `MarshalJSON` encodes `measurementInputsWire` with `Branches` set to nil, then adds a `branches` member holding the compact form. `MarshalYAML` does the same through the YAML node. `UnmarshalJSON` (which YAML decoding reaches through `decodeMeasurementYAML`) checks the version first.
  - If the version is 4, it lifts the raw `branches` member out before the ordinary decode, decodes it as v4 wire, and expands it.
  - Otherwise the path is unchanged.
  - **Injected into:** nothing new. Every caller keeps using `MeasurementInputs`.
- **Writer.** `actualinputs.go:254-257` writes `MeasurementCompactBranchInputsVersion` (4) for the branch and branch-hybrid methods. The queue-method rewrite at line 497 keeps the version.

## Operating envelope (ARCH-CONSTRAINTS)

- **Workload:** one encode and a few decodes per close and per `sdlc actual`. This is batch work, not on a hot path.
- **Size:** a v4 snapshot is at most 128 KiB, enforced. The basis is measured: #586 is about 55 KB in v4, against 67 KB in v3. At 128 sources with #586's shape (two intervals per source), it is about 60 + 31 = 91 KB, under the bound.
- **Parse cost:** the structural cap is 512 KiB, a fixed multiple that bounds a hostile or hand-edited ledger. Decode is linear in its size.
- **Over the bound:** the counted refusal (D3). Close keeps its `--no-actual` recovery.

## Other principles

- **ARCH-EVOLVE:** v4 changes interpretation, because an old reader would read `source_index` as an unknown member and find no `source`. So it is a version bump, and readers before #593 refuse version 4 by the existing version check (`measurement.go:184`). Unknown members survive on the inputs, record and interval at both versions. A test rewrites a v4 snapshot that carries unknown members at all three levels and gets them back byte-equal.
- **ARCH-SECURE:** a ledger is untrusted input, since it can be hand-edited or truncated. `expandBranches` turns indices into typed values and refuses anything out of range. Expanded evidence then passes the unchanged `validateBranchEvidence`.
- **ARCH-PURE:** compaction and expansion are pure. The codec is the thin seam.
- **ARCH-ORDER:** the snapshot holds no state between events. It is an immutable value written once per close generation.
- **ARCH-FUNERAL:** creates nothing new and durable. The ledger already exists, and each v4 snapshot is smaller than the v3 one it replaces.
- **ARCH-MOCK:** N/A. No external dependency is involved; the tests use in-memory snapshots and the existing frozen v1 reader in `testdata/measurementv1`.
- **ARCH-DRY:** the generic codec helpers are reused, and there is a single limit function.
- **ARCH-PURPOSE:** the bound agreement is enforced as well (D2), so the source bound, not bytes, decides. This is the issue's point, not only the #586 shape.

## Plan

- [ ] Write failing tests in `cmd/sdlc/internal/gatestate/measurement_branch_v4_test.go`:
  - `TestCompactBranchesRoundTrip`: compact and then expand returns identical evidence, including interval branches that differ from the record, an empty-branch `ownership-only` interval, a source missing from `sources`, and unknown members on the record and interval.
  - `TestExpandBranchesRefusesBadIndex`.
  - `TestMeasurementV4Issue586Shape`: 75 sources, 1 record, 150 intervals, a 112-character branch. It validates, the JSON is under 64 KiB, and it round-trips through JSON and YAML to an equal value.
  - `TestMeasurementV4At128Sources`: 128 sources of the same shape validate under 128 KiB.
  - `TestMeasurementOversizeCountedRefusal`: the error names the bytes, the bound, the version and the counts, and does not contain "does not serialize".
  - `TestMeasurementV3ReplaysUnchanged`: a v3 golden snapshot decodes and re-encodes byte-identically, and a v3 snapshot over 64 KiB still refuses.
  - `TestMeasurementV4RefusedByOlderReader`: the frozen `measurementv1` reader, and the version check as a pre-#593 reader runs it, refuse version 4.
  - `TestMeasurementV4PreservesUnknownMembers`.
- [ ] Implement the wire types and `compactBranches` / `expandBranches` in `measurement_branch_v4.go`.
- [ ] Add the version-aware limit and the structural cap, plus `Validate`'s counted refusal and the version gate (branch methods accept 3 or 4) in `measurement.go`. Codec: `checkMeasurementJSON` and the YAML budget use the structural cap, and the v4 `branches` member is swapped in the MarshalJSON, MarshalYAML and UnmarshalJSON paths in `measurement_codec.go`.
- [ ] Writer: `actualinputs.go` writes version 4.
- [ ] Run `go test ./cmd/sdlc/internal/gatestate/` and `go test ./cmd/sdlc -run 'Actual|Measurement|Branch|Close'`.
- [ ] Docs: the `sdlc actual` help (`cmd/sdlc/helptext/actual.md`, the BRANCH ATTRIBUTION section), covering the v4 compact proof, the 128 KiB bound, and the counted refusal. Atlas: the measurement entry, if one names the 64 KiB bound.
- [ ] Verify by re-measuring #586 with the built binary: `sdlc actual --issue 586` gives a number, not a refusal.

## Done when (from the issue)

- A window shaped like #586 measures and is retained under the bound, and a regression test pins that shape.
- At `SnapshotLimitSources`, a window either fits or refuses up front with the counted error, never late inside MarshalJSON.
- Replay of an existing v3 snapshot is unchanged.
