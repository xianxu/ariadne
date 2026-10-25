---
collaboration-mode: discussion
---

# Measured actuals never block close: implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `sdlc close` (and `milestone-close`) never refuses because measurement failed. Each close records what it measured
together with a confidence: `exact`, `partial` (with the attributed share) or `unmeasured` (with the reason). Calibration then
weights each row by that confidence.

**Architecture:** A pure classifier turns the measurement outcome into an `actualConfidence`. The measurement seams that refuse
today stop returning errors when the minutes in question have a known extent: unknown-branch intervals, a source whose branch
attribution failed, a snapshot over its bound, and queue attempts whose provenance is unknown. Those seams record the minutes as
unattributed instead. Close adopts every outcome. The confidence goes into the close evidence and into two appended
calibration-ledger columns, and the calibration readers weight each row by its share.

**Tech stack:** Go (`cmd/sdlc`, `internal/activetime`, `internal/gatestate`, `internal/estimate`) and CUE vocabulary.

---

## Problem recap (from the issue and the code survey)

Close adopts hours only when the status is `actualMeasured` (`resolveOmittedActual`, close.go:1765). Its caller exits 1 on every
other status (close.go:459), and the advice it prints ends in `--no-actual`, which records N/A and drops all of the hours. The
refusing shapes, and the data still available when each one refuses:

| Shape | Where it refuses | Minutes on hand when it refuses |
|---|---|---|
| Resting branch (#591) and stage branch (#594) | not a refusal: excluded with the notes `resting-excluded` and `stage-excluded` | yes; already shown as excluded minutes |
| Unknown, detached or ambiguous branch under concurrent claims | `SelectBranchActivity` error (activetime/branch.go:168) | yes; the windows before it, plus the interval's bounds |
| Branch attribution for one source (no root lineage, no pinned owner, ambiguous historical owner) | actualbranch.go:233-252 returns `nil, err` for the whole resolve | yes; that source's activity windows, with no branch for them |
| More than 128 sources at the **branch proof** (pair#838, pair#851: "measurement branch proof: measurement sources exceed snapshot limit 128") | `validateOwnedBranchProof` calls `m.Validate()` on the fresh inputs (actualbranch.go:313), *before* the sources are trimmed to the ones that contributed | all of it; the trimmed snapshot may well fit |
| Snapshot over 128 KiB (v4) or 64 KiB (v1-3), or over 128 sources after trimming | `m.Validate()` at actualmeasurement.go:174 | all of it; the hours, split and phases are already computed |
| Queue evidence incomplete | four `queueEvidenceError` sites (below) | yes; the windows. The queue time of the unknown attempts cannot be subtracted |
| Everything else: resolver input errors, telemetry gap, empty window, no window, plan unavailable, exclusion replay, lineage cycle | various | no measured number |

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `actualConfidence` | `cmd/sdlc/actualconfidence.go` | new |
| `classifyActual` | `cmd/sdlc/actualconfidence.go` | new |
| `MeasurementUnattributed` (the `unattributed` member of `MeasurementInputs`) | `cmd/sdlc/internal/gatestate/measurement_unattributed.go` | new |
| `MeasurementSummary` | `cmd/sdlc/internal/gatestate/measurement_summary.go` | new |
| `MeasurementInputs.Validate` split into `ValidateShape` and `ValidateBounds` | `cmd/sdlc/internal/gatestate/measurement.go` | modified |
| `BranchSelection.Unknown` (unknown-branch intervals returned, not raised) | `cmd/sdlc/internal/activetime/branch.go` | modified |
| `resolveOmittedActual` | `cmd/sdlc/close.go` | modified |
| `LedgerRow.Confidence`, `LedgerRow.AttributedShare` and `CalibrationWeight()` | `cmd/sdlc/internal/estimate/ledger.go` | modified |
| `CalibrateForecast`, `ConvergeFactors`, `DriftVerdict` | `cmd/sdlc/internal/estimate/*.go` | modified |

- **actualConfidence**: `{Level: exact|partial|unmeasured; AttributedMinutes, UnattributedMinutes float64; Bounded bool; Reasons []string}`.
  It is always derived, never typed by hand. `Share()` is attributed ÷ (attributed + unattributed), and 0 when `Bounded` is false.
  - **One meaning of partial hours:** the hours exclude every minute whose attribution is unproven but whose extent is known, so
    they are a lower bound. The single exception is an unknown that has no extent (an unbounded queue attempt or an unreadable
    attempt record). That record has `Bounded=false` and share 0, and shows its hours as an upper bound. Calibration weight 0
    keeps it out of calibration (D4).
  - Rendering: `exact`, `partial 0.83 (12m unattributed: unknown branch during …)`, or
    `partial, upper bound (queue attempt X has no endpoints)`.
  - **Relationships:** each close or milestone-close result has exactly one. The ledger row and the evidence trailer each copy it.
  - **DRY rationale:** one classifier and one renderer replace the status logic now split across `resolveOmittedActual`,
    `explainActual` and the `--no-actual` advice in `printActual` (actual.go:443-461). `sdlc actual`, close and milestone-close
    all print the same line.
  - **Future extensions:** a confidence per segment, if calibration ever wants a weight per phase.
- **classifyActual(res actualResult) actualConfidence** is a pure mapping:
  - `actualMeasured` with no unattributed record gives `exact`. Resting and stage exclusions keep a close `exact` (D1); they are
    shown as excluded minutes, as #591 does.
  - `actualMeasured` with `Inputs.Unattributed` entries gives `partial`. Its minutes come from those entries, clipped to the
    measurement window.
  - Any other status gives `unmeasured`, with `Detail` (plus `Recovery`, when present) as the reason.
- **MeasurementUnattributed**: `{Source, Start, End, Kind: unknown-branch|source-attribution|queue-unknown|summary-widened, Bounded bool, Reason}`.
  It is the one record of minutes that were seen but not counted.
  - **Why it is not a new branch-interval kind** (plan-quality PQ-1): `MeasurementBranchInterval.Kind` is a strict enum
    (measurement_branch.go:97), and `validateOwnedBranchProof` treats any kind other than an exclusion as a selected witness
    (actualbranch.go:321-329). A source whose attribution failed also has no branch capture to hang a note on.
  - Instead it is a new top-level member of `MeasurementInputs`. The sources' `Windows` stay as they are (they hold only the
    counted, selected time), and a source with no attribution at all is not retained in `Sources`.
  - That makes the member **safe to ignore**, the ARCH-EVOLVE condition. An older reader preserves it through
    `decodeMeasurementObject`'s extra fields, ignores it, and replays exactly the adopted hours, because the counted windows never
    included these minutes. The branch-interval enum and the proof validator stay unchanged.
- **MeasurementSummary**: the retained form of a snapshot that is still over its bound after trimming.
  - Fields: `{Version, Since, Until, Repository, Sources, Bytes int, Hours float64, Digest string, Credited []adoptedCredit (Issue, Slot, Start, End)}`.
  - It is stored *instead of* the full inputs, as a new `MeasurementInputs` method `summary-v1`. The hours stay `exact` (D2).
  - **Credited intervals keep exclusion working** (PQ-2). A later issue's `loadAdoptedExclusions` needs only the prior issue's
    credited intervals; today it gets them by replaying the prior issue (actualexclusion.go:387-412). With a summary it reads
    `Credited` directly and checks Σcredited against the adopted hours within the same 0.005h tolerance. So a summarised close
    never pushes the issues that overlap it to `unmeasured`.
  - **Its size is bounded** by `MaxSummaryCredits` (2048 intervals, about 160 KiB). Intervals in each slot are merged at their
    joins. If that is still too many, the closest neighbours are coalesced until it fits. Coalescing over-protects, so the later
    issue's hours are undercounted, which is the safe direction; that later issue records the widened minutes as `Unattributed`
    entries of Kind `summary-widened`, so its own confidence shows it.
  - Only `sdlc actual --replay` of the summarised issue itself refuses, with "not replayable: summarised measurement".
  - An older reader that meets method `summary-v1` rejects the unknown method, sees no trustworthy saved window, and refuses as it
    does today. That is the safe direction. It is transitional, since the fleet runs `:0`'s binaries.
- **ValidateShape and ValidateBounds**: `Validate` becomes the pair of them. `validateOwnedBranchProof` and the decoding of fresh
  inputs call `ValidateShape` only. The bounds are checked once, at retention (actualmeasurement.go:174), after trimming. Over
  the bound there means a `MeasurementSummary`, not a refusal. This alone fixes the pair#838 and pair#851 shape whenever the
  trimmed snapshot fits.
- **BranchSelection.Unknown**: the default case at branch.go:168 appends `{Start, End}` and continues, instead of returning an
  error. The resolver turns those entries into `Unattributed` records of Kind `unknown-branch`. The "no pinned claim" error at
  :146 stays an error, because it means the inputs disagree and there is no interval to count; it classifies as `unmeasured`.
- **LedgerRow** gains two appended columns, `confidence` and `attributed_share`; the header contract is append-only
  (ledger.go:127). A row without them reads as `exact` with share 1, which is true because close enforced exact measurement when
  that row was written. `CalibrationWeight()` is 1 for `exact`, the share for `partial` (0 when not bounded), and 0 for
  `unmeasured`. `unmeasured` rows are never written in any case, since a row needs a numeric actual.
- **Calibration readers**: the sums become weighted, as Σw·build, Σw·converge and Σw·estimate. The percentile inputs take only
  rows with w ≥ 0.5 (D3). The reported N becomes "N rows, M partial".

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `resolveOwnedBranchSources` | `cmd/sdlc/actualbranch.go` | modified | per-source HEAD-log capture (git) |
| `computeOwnedActualFor` | `cmd/sdlc/actualmeasurement.go` | modified | snapshot validation and retention |
| `attachOwnedQueueEvidence`, `filterIssueQueueAttempts` | `cmd/sdlc/actualinputs.go`, `actualqueue_filter.go` | modified | suite-attempt store, git head membership |
| `loadAdoptedExclusions` | `cmd/sdlc/actualexclusion.go` | modified | tracker cards plus the saved inputs of earlier issues |
| `adoptOmittedActual`, close.go:459, and the close writers | `cmd/sdlc/close.go`, `closetracker.go` | modified | issue file, tracker evidence commit, calibration ledger |

- **resolveOwnedBranchSources**: when one source fails attribution (no root lineage, no pinned owner, ambiguous historical owner),
  that source is dropped from `Sources`. Its activity windows become `Unattributed` records of Kind `source-attribution`, carrying
  the error text, and the other sources keep their attribution. A lineage *cycle* stays an error and classifies as `unmeasured`,
  because it corrupts the lineage of every source.
- **computeOwnedActualFor**: when `ValidateBounds` fails (a typed `oversizeError` carrying the counts), it builds the
  `MeasurementSummary` and keeps the status measured. It warns that the snapshot is summarised and cannot be replayed.
  `ValidateShape` failures stay errors.
- **Queue evidence, one rule per site.** The rule: the attempt's endpoints are known means a bounded `queue-unknown` record (its
  overlap with the counted windows is subtracted as if it were queue time and recorded as unattributed, so the hours remain a
  lower bound); its endpoints are not known means an unbounded record.

  | Site | Cause | Outcome |
  |---|---|---|
  | actualinputs.go:437 | unreadable or opaque attempt records in the store | unbounded: `partial`, upper bound (no endpoints are known) |
  | actualinputs.go:472 | the head-membership read failed for every attempt | each attempt has endpoints, so each is bounded: `partial` s |
  | actualqueue_filter.go:55 | one attempt has unknown head attribution | bounded: `partial` s |
  | actualinputs.go:483, from `selectOwnedQueueAttempts` | a legacy attempt with no endpoints (actualqueue.go:73) | unbounded: `partial`, upper bound |
  | the same site | unknown queue or slot provenance, with endpoints (actualqueue.go:46, :93) | bounded: `partial` s |

  The hard errors stay errors and classify as `unmeasured`: a duplicate attempt, several matching slot engagements
  (actualqueue.go:116), and `potentialIssueQueueAttempts` failing.
- **loadAdoptedExclusions**: when the saved method is `summary-v1`, it uses `Summary.Credited` where it would otherwise replay. It
  still refuses any other "no trustworthy saved window" case, as it does today. That error, raised in the resolver, now
  classifies the *current* close as `unmeasured` instead of refusing it.
- **Close writers**: close.go:459 stops calling `exitWithCode(1)` for measurement outcomes; `explainActual` itself never exited.
  `adoptOmittedActual` adopts every confidence:
  - `exact` and `partial` write the hours and the confidence.
  - `unmeasured` writes `N/A` and the reason. It prints the recovery as a warning and names `sdlc issue correct-actual` for filling
    the value in later.
  - Tracker repos get an evidence-commit trailer, `Close-Actual-Confidence: partial 0.83; <reasons>`, beside `Close-Actual`.
  - Legacy repos get frontmatter `actual_confidence:` beside `actual_hours`, as an optional cue field.
  - The calibration row carries both new columns.
  - The `--no-actual` warning text stays unchanged (gatesig.go:86 pins it).

## Decisions (the TL ruled on 2026-10-24: the D1 alternative; D2-D5 as recommended)

- **D1:** resting and stage exclusions keep a close `exact`, and the excluded minutes are still shown. A slot that switches
  branches for a few seconds is not doing unattributed issue work.
- **D2:** a snapshot still oversized after trimming is `exact` plus a summary. It is not partial, and the sources are not cut to
  the top 128.
- **D3:** sums are weighted by share, and only rows with share ≥ 0.5 enter the percentile range.
- **D4:** an unknown with no extent is `partial` with share 0, shown as an upper bound, with calibration weight 0.
- **D5:** the confidence goes in the trailer and the ledger (plus frontmatter in legacy repos), not in a new card field.
- **`--no-actual` stays** for genuinely inapplicable windows (item 4). Its help text stops presenting it as the fix for a
  measurement failure.

## State design table

Actions: close (or milestone-close), `sdlc actual` (preview), `sdlc issue correct-actual`, close with `--no-actual`.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| All windows attributed, snapshot within bounds after trimming | `exact` | none | none |
| Resting or stage minutes | `exact`, with the excluded minutes shown | none | none (not this issue's work) |
| Snapshot still oversized after trimming | `exact`, with a "summarised, not replayable" warning | none; exclusion uses the summary's credited intervals | replay of this close |
| Unknown-branch intervals, or a source whose attribution failed | `partial s`, with the reasons | `correct-actual` if the true value is known | the unattributed minutes (recorded, not counted) |
| Queue attempt with unknown provenance or membership, but with endpoints | `partial s` | none | the overlap of the attempt with the counted windows |
| Unbounded queue attempt, or unreadable attempt records | `partial`, upper bound, weight 0 | `correct-actual` | this row's calibration value |
| An earlier overlapping issue had its credits coalesced | `partial s`, with `summary-widened` minutes | none | the widened minutes |
| Input or replay error, telemetry gap, empty window, no window, lineage cycle, duplicate attempt, no trustworthy saved window | `unmeasured (<reason>)`, N/A, the recovery printed as a warning | recover the evidence, then `correct-actual` | the hours, until corrected |
| The process dies mid-close | unchanged: nothing is written before the evidence commit | rerun close | none |

Covered: every refusal shape in the issue, plus the residual errors. Not covered: changing how measurement attributes time. This
plan only stops discarding what has already been seen.

## Envelope (ARCH-*)

- **ARCH-EVOLVE:**
  - `unattributed` is a new member that is safe to ignore; the argument is above. The test is an older-shaped reader rewriting a
    newer record and keeping the member.
  - `summary-v1` is a new method behind the existing strict method check. An older reader refuses it, which is the safe
    direction: ignoring a missing source list would be unsafe, so a refusal is warranted.
  - The ledger columns are appended, and an old row reads as exact.
  - The trailer and the cue field are additive.
  - The branch-interval kind enum is unchanged.
- **ARCH-PURE:** `classifyActual`, `CalibrationWeight`, building the summary, coalescing credits and the weighted calibration are
  all pure. The IO changes only turn errors into records, at existing seams.
- **ARCH-CONSTRAINTS:** close runs as a batch. Classifying is O(records). Building the summary is one pass over the sources,
  O(S), and coalescing is O(C log C) over at most a few thousand credits. No new IO is added.
- **ARCH-FUNERAL:** there is no new artifact family. `unattributed` grows by one record per uncounted interval, clipped to the
  window. The summary is bounded by `MaxSummaryCredits`, and the bound is on the writer.
- **ARCH-SECURE:** the reasons are error strings that sdlc itself produces (IDs, SHAs and timestamps). They reach a trailer and the
  ledger. A test asserts that no transcript content appears in them.
- **ARCH-ORDER:** close holds no state between events, because it is a single pass. The evidence commit is still the one write
  that makes it durable. Exclusion reads an earlier issue's summary as an observation pinned to that issue's completion evidence,
  the same generation rule as the replay it replaces.
- **ARCH-MOCK:** the existing fakes, recorded fixtures and capture seam are reused, with no new external surface.
- **ARCH-PURPOSE:** every refusal shape in Done-when is covered in this one issue. Calibration and docs are M2 of this issue, not
  a follow-up.
- **ARCH-DRY:** one classifier and one renderer for close, milestone-close and `sdlc actual`. One `Unattributed` record serves all
  three kinds of uncounted minutes.

## Replay fixtures (Done-when)

| Window | Shape (from pair's Log) | Fixture | Expected |
|---|---|---|---|
| pair#856, #812, #863 | resting-branch claim transition | `activetime/testdata/pair856`, `pair812`, `pair863` (exist) | `exact`, with the resting minutes excluded and shown |
| pair#830 | stage branch | `branch_staging_test.go` (in code) | `exact`, with the stage minutes excluded and shown |
| pair#838 | "measurement branch proof: measurement sources exceed snapshot limit 128" (pair 09f4314e9) | new: generated from the refusal's shape, more than 128 fresh sources and at most 128 that contribute | `exact`, with the **full** replayable snapshot (the proof no longer checks bounds) |
| pair#851 | the same refusal, window 2026-10-23T20:43:25-07:00 → 2026-10-24T10:40:39-07:00 (pair Log) | new: the same generator with more than 128 contributing sources | `exact`, with a `MeasurementSummary`; a later overlapping issue still measures |

pair#838 and #851 exercise the source-cap shape at the branch proof. That refusal still happens on main after #593, because #593
changed the byte bound, not the timing of the proof check. Neither has raw captures in this repository, so both fixtures are
synthetic. They are shaped from the printed refusal and window, and the README says so. Each fixture runs end to end through the
real resolver, the engine and `classifyActual`, using the close seam with a fake transcript tree.

## Plan

### M1: close never refuses on measurement

- [x] **Task 1: `actualConfidence` and `classifyActual`**, test-first in `actualconfidence_test.go`. Strategy: a table test over the
  closed (status × unattributed kind × bounded) domain.
- [x] **Task 2: `MeasurementUnattributed`.** Add it in gatestate with strict known fields. Tests cover the JSON and YAML round trip
  and an older-shaped reader rewriting a record without dropping the member.
- [x] **Task 3: `ValidateShape` and `ValidateBounds`.** The proof validator calls only `ValidateShape`. Test: more than 128 fresh
  sources with at most 128 contributing gives `exact` with the full snapshot (the pair#838 shape).
- [x] **Task 4: `BranchSelection.Unknown`.** `SelectBranchActivity` returns the unknown intervals and the resolver records them.
  Update `branch_*_test.go`. Test: concurrent claims around a detached interval keep the windows before and after it, and the
  interval is recorded.
- [x] **Task 5: per-source attribution failure.** The source is dropped and recorded as `source-attribution`; a lineage cycle
  stays an error. Tests cover a missing owner and an ambiguous owner.
- [x] **Task 6: `MeasurementSummary`.** Add the `summary-v1` method, the credited intervals and coalescing. Strategy:
  the coalescer property test below; each bound (sources, bytes) crossed by one gives a summary.
- [x] **Task 7: exclusion against a summary.** `loadAdoptedExclusions` uses `Credited`. Tests: an overlapping later issue is
  measured, Σcredited must match the adopted hours, and widened minutes make the later issue `partial`.
- [x] **Task 8: queue degradation**, one test per site in the table. Reuse `testdata/actualqueue`. The hard errors stay
  `unmeasured`.
- [x] **Task 9: close adopts every outcome.** close.go:459 no longer exits on measurement. `resolveOmittedActual` takes an
  `actualConfidence`, the info line renders it, and the trailer, frontmatter field and cue field are added. Tests: close end to
  end for each level, in tracker and legacy repos, plus `TestAdoptLineNoGatesigCollision` over the new renderings. `--no-actual`
  is unchanged.
- [x] **Task 10: replay fixtures** for pair#856, #812, #863, #830, #838 and #851 (table above), each through the close seam.
- **Test strategy for the risky functions (plan-quality advisory):** the credit coalescer gets a property test over random
  credit sets: the output has at most `MaxSummaryCredits` intervals, its cover is a superset of the input cover, and the widened
  minutes equal cover(output) − cover(input). `classifyActual` stays a table test, because its domain is a closed enumeration.
- [x] **M1**: milestone-close.

### M2: calibration reads confidence; docs

- [x] **Task 11: ledger columns** `confidence` and `attributed_share`. Old rows read as exact. Test the round trip, including an old
  header.
- [x] **Task 12: weighted calibration** in `CalibrateForecast`, `ConvergeFactors` and `DriftVerdict` (D3). Strategy: the property test below.
- **Test strategy:** weighted calibration gets a property test: adding a row with weight 0 leaves every output unchanged, and an
  all-exact ledger gives exactly today's numbers.
- [x] **Task 13: help and atlas.**
  - Help: update `helptext/close.md`, `actual.md` and `milestone-close.md` for the confidence levels, and to limit `--no-actual`
    to windows where measurement does not apply.
  - Atlas: update `sdlc-binary.md` (branch selection, the v4 bound and the summary), `issue-lifecycle.md`, `ledger-landscape.md`
    and `gate-state.md`.
- [x] **M2**: close.

## Revisions

- 2026-10-24, TL approval: D1 was reversed to the alternative (resting and stage exclusions keep a close `exact`, and the minutes
  are shown). D2-D5 were approved as written.
- 2026-10-24, plan-quality round 1:
  - **PQ-1:** the uncounted minutes move to a new top-level `unattributed` member that is safe to ignore. The branch-interval
    enum and the proof validator are untouched.
  - **PQ-2:** `MeasurementSummary` keeps the credited intervals (bounded, with coalescing), so exclusion needs no replay, and
    `Validate` is split so the proof checks shape only.
  - **Minors:** `partial` hours have one meaning (a lower bound, except an unbounded unknown, which has weight 0). All four
    queue sites are mapped. The exit is at close.go:459. pair#838 and #851 are confirmed as the source-cap-at-proof shape, which
    #593 did not fix.
- 2026-10-24, Task 8 (TL approved; this issue's third design delta): every unknown queue attempt, bounded or not, takes the
  D4 path. Its wait is not subtracted, the hours are an upper bound, partial with share 0; a bounded attempt also records its
  interval and the counted time it overlaps, so the output says how much the hours may overstate. The plan's bounded
  subtraction is dropped: the engine binds queue waits to sources by slot, which these attempts lack, and replaying the
  subtraction would make the records unignorable by older readers.
- 2026-10-24, implementation: the summary also keeps the coalesced gaps (`widened`, at most `MaxSummaryCredits`), so a later
  issue records the over-protected minutes; the summary byte bound is 448 KiB. Unattributed minutes subtract only this
  issue's counted cover.
- 2026-10-24, sequencing: Tasks 11-13 (ledger columns, weighted calibration, help and atlas) landed in the same commits
  as M1, so the M1 milestone review covers them; M2's close reviews what follows. The summary branch of exclusion is a pure
  `summaryProtection`, unit-tested, rather than an end-to-end overlapping-card fixture.
- 2026-10-24, M1 boundary review round 1 (all findings fixed in the round):
  - **Frontmatter:** `actual_confidence` persists the bare `Token()` (exact, partial 0.83, partial upper-bound,
    unmeasured), never the reason text, which made invalid YAML; the cue field matches exactly that set.
  - **Unmeasured recovery:** an N/A actual cannot be corrected later (`correct-actual` needs a numeric actual and a ledger
    row), so close no longer advertises it; it says to restore the evidence before the close. Supporting N/A-to-number
    corrections would be a fourth design delta, which the TL ruled is a scope cut. State table: the unmeasured row's
    recovery is "restore the evidence before the close"; the lost hours stay lost.
  - **Corrections** require an exact measurement, since a correction row carries no confidence columns.
  - **Ambiguous owner:** the claim stays filed under its real owner key; only the shared root is marked ambiguous.
  - **Widened gaps** count only where they meet the candidate's slot window.
  - **Reasons** in a rendering are bounded (240 bytes each) and home-relative.
  - Exclusion checks a summary's recorded hours (as the replay path checks replayed hours), not Σcredited.
  - Task 13: `atlas/workflow/gate-state.md` holds no measurement content and was not changed; `README.md` was.
- 2026-10-24, M1 boundary review round 2 (FIX-THEN-SHIP; fixed in the round): writer bounds now match the validator's.
  `CoalesceCredits` bounds `widened` to `MaxSummaryCredits` too (coalescing gaps can only overstate what a later issue
  loses); every unattributed reason passes one UTF-8-safe `gatestate.BoundReason`/`BoundLine`, which the rendering bound
  shares; the summary's `Bytes` comes from the `OversizeError` the encoder measured. Task 10 as delivered: the pair#856,
  #812, #863 and #830 fixtures are selection-level `SelectBranchActivity` tests asserting no unattributed activity (the
  close seam adds nothing to them); pair#838, #851 and #866 run through the real resolver and engine.
