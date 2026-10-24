---
collaboration-mode: discussion
---

# Measured actuals never block close: implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `sdlc close` (and `milestone-close`) never refuses because measurement failed. It records what it measured together with a
confidence (`exact`, `partial` with the attributed share, or `unmeasured` with the reason), and calibration weights rows by that
confidence.

**Architecture:** A pure classifier turns the measurement outcome (`actualResult` plus the notes about excluded intervals) into an
`actualConfidence`. The places that refuse today stop returning errors for the shapes whose minutes can still be counted: unknown
or detached branch intervals, a branch-attribution failure on a single source, and a snapshot over its bound. Instead they report
those minutes as unattributed. Close adopts every outcome. The confidence goes to the close evidence (a trailer, or frontmatter in
legacy repositories) and to two appended calibration-ledger columns. The calibration readers weight each row by its share.

**Tech stack:** Go (`cmd/sdlc`, `internal/activetime`, `internal/gatestate`, `internal/estimate`) and CUE vocabulary.

---

## Problem recap (from the issue and the code survey)

Close adopts hours only when the status is `actualMeasured` (`resolveOmittedActual`, close.go:1765). Every other outcome exits 1
and tells the user to use `--no-actual`, which writes N/A and drops all of the hours. The refusing shapes, and the data that is
still on hand at the point of refusal:

| Shape | Where it refuses | Minutes on hand when it refuses |
|---|---|---|
| Resting branch (#591) and stage branch (#594) | already excluded with notes (`resting-excluded`, `stage-excluded`), not refused | yes; the excluded minutes are known but recorded nowhere |
| Unknown, detached or ambiguous branch with concurrent claims | `SelectBranchActivity` error (activetime/branch.go:168) | yes; the earlier windows plus this interval's bounds |
| Branch attribution for one source (lineage cycle, no root lineage, no pinned owner, ambiguous historical owner) | actualbranch.go:228-252 returns `nil, err` for the whole resolve | yes; that source's activity windows, with no branch for them |
| Snapshot over 128 KiB (v4) or over 64 KiB (v1-3) | `m.Validate()` at actualmeasurement.go:174 | all of it; hours, split and phases are already computed |
| More than 128 sources | same call site | all of it |
| Queue evidence incomplete (legacy or unknown attempt provenance) | `queueEvidenceError` at the end of the resolver (actualinputs.go:324) | windows yes; queue time cannot be subtracted for the unknown attempts |
| Everything else: resolver input errors, telemetry gap, empty window, no window, plan unavailable, exclusion replay | various | no measured number |

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `actualConfidence` | `cmd/sdlc/actualconfidence.go` | new |
| `classifyActual` | `cmd/sdlc/actualconfidence.go` | new |
| `BranchNote` kind `unknown-excluded` | `cmd/sdlc/internal/activetime/branch.go` | modified |
| `MeasurementSummary` | `cmd/sdlc/internal/gatestate/measurement_summary.go` | new |
| `resolveOmittedActual` | `cmd/sdlc/close.go` | modified |
| `LedgerRow.Confidence` and `LedgerRow.AttributedShare`, with `CalibrationWeight()` | `cmd/sdlc/internal/estimate/ledger.go` | modified |
| `CalibrateForecast`, `ConvergeFactors`, `DriftVerdict` | `cmd/sdlc/internal/estimate/*.go` | modified |

- **actualConfidence**: `{Level: exact|partial|unmeasured; AttributedMinutes, UnattributedMinutes float64; Reasons []string}`.
  It is derived and never typed by hand.
  - `Share()` is attributed / (attributed + unattributed).
  - `exact` means there are no unattributed minutes. `unmeasured` means there is no number.
  - Rendering: `exact`, `partial 0.83 (12m unattributed: unknown branch during …)`, or `unmeasured (telemetry gap: …)`.
  - **Relationships:** each close or milestone-close result has one. The ledger row and the evidence trailer each hold a copy.
  - **DRY rationale:** one classifier replaces the status-switch logic now spread across `resolveOmittedActual`, `explainActual`
    and the `--no-actual` advice in `printActual` (actual.go:443-461). `sdlc actual`, close and milestone-close all print the
    same rendering.
  - **Future extensions:** a confidence per segment, if calibration ever wants a weight per phase.
- **classifyActual(res actualResult) actualConfidence**: a pure mapping.
  - `actualMeasured` with no excluded or unattributed notes gives `exact`.
  - `actualMeasured` with notes of kind `unknown-excluded`, `source-unattributed` or `resting-excluded`/`stage-excluded` gives
    `partial`. Resting and stage minutes count as unattributed, so the share is honest; see Decision D1.
  - Any other status gives `unmeasured`, with `Detail` (and `Recovery`, when present) as the reason.
- **unknown-excluded note**: the default case at branch.go:168 appends a note and continues instead of returning an error. The
  "no pinned claim" error at :146 stays an error, because it is an input inconsistency with no interval to count, so it classifies
  as `unmeasured`.
- **MeasurementSummary**: a retained form for a snapshot that is over its bound or over the source cap.
  - Fields: `{Version, Sources int, Bytes int, Minutes float64, Digest string}`. The digest is over the ordered source digests and
    windows.
  - Stored in place of the full `MeasurementInputs`, so a replay of that close is impossible; it identifies the evidence but cannot
    re-verify it.
  - The hours stay `exact`. Confidence describes attribution, not whether the evidence can be replayed (D2).
  - Consumers that need replay (overlap exclusion in `loadAdoptedExclusions`, correction) already have a "no trustworthy saved
    window" path. An issue that later overlaps a summarised one degrades to `unmeasured` through that path instead of refusing.
- **LedgerRow**: two columns are appended, `confidence` and `attributed_share`. The header contract is append-only
  (ledger.go:127).
  - An old row without the columns reads as `exact`, share 1. Close enforced exact measurement then, so that reading is true.
  - `CalibrationWeight()` is 1 for `exact`, the share for `partial`, and 0 for `unmeasured`. `unmeasured` rows are never written
    anyway, because the row requires a numeric actual.
- **Calibration readers**: each sum becomes weighted (Σ w·build, Σ w·converge, Σ w·estimate). Percentile inputs use rows with
  w ≥ 0.5; see D3. The `N` they report becomes "N rows, M partial".

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `resolveOwnedBranchSources` | `cmd/sdlc/actualbranch.go` | modified | per-source HEAD-log capture (git) |
| `computeOwnedActualFor` | `cmd/sdlc/actualmeasurement.go` | modified | snapshot validation and retention |
| `attachOwnedQueueEvidence` | `cmd/sdlc/actualinputs.go` | modified | suite-attempt store |
| `adoptOmittedActual` and the close writers | `cmd/sdlc/close.go`, `closetracker.go` | modified | issue file, tracker evidence commit, calibration ledger |

- **resolveOwnedBranchSources**: a single source's attribution error no longer fails the resolve. That source's windows become
  `source-unattributed` notes carrying the error text, and the other sources still count. A lineage *cycle* is different: it
  corrupts every source's lineage, so it stays an error and classifies as `unmeasured`.
  - **Injected into:** the existing fakes, the recorded HEAD-log fixtures and the capture seam. No new external surface.
- **computeOwnedActualFor**: when `Validate` fails only on size or source count (a typed `oversizeError` carrying the counts),
  it retains a `MeasurementSummary` and keeps the status `actualMeasured`, with a warning that the evidence is not replayable.
  Any other `Validate` failure stays an error.
- **attachOwnedQueueEvidence**: today, incomplete queue evidence is an error. Instead the engine runs without subtracting queue
  time for the unknown attempts. The overlap of those attempts' known bounds with the counted windows becomes unattributed minutes,
  which makes the result `partial`. An attempt with no bounds at all (a legacy attempt with no endpoints) makes the whole result
  `partial` with share 0: the hours are recorded as an upper bound, but calibration gives the row weight 0 (D4).
  The hard errors (duplicate attempt, several matching slot engagements) stay errors and classify as `unmeasured`.
- **Close writers**: every outcome is adopted.
  - `exact` and `partial` write the hours (attributed minutes only, never extrapolated) and the confidence.
  - `unmeasured` writes `N/A` plus the reason. It prints the recovery as a warning, not a refusal, and names
    `sdlc issue correct-actual` for filling the value in later.
  - Tracker repos: the evidence commit gains a trailer, `Close-Actual-Confidence: partial 0.83; <reasons>`, beside
    `Close-Actual`.
  - Legacy repos: frontmatter `actual_confidence:` beside `actual_hours`.
  - The calibration row carries both new columns.
  - The `--no-actual` warning text is unchanged, because gatesig.go:86 pins it.

## Decisions for the TL

- **D1: resting and stage minutes count as unattributed.** They are real work time that no branch owns. Counting them makes the
  share honest. The cost is that more closes read `partial` than read `exact`. *Alternative:* treat them as correctly excluded
  (not this issue's time), which keeps those closes `exact`. **I recommend counting them as unattributed**, because #591 and #594
  dropped those minutes without leaving any record.
- **D2: an oversized snapshot or over-cap source count is `exact` with a summary.** It is not `partial`. The issue's item 2 asks
  to merge or summarise sources; the summary does that. Every hour was attributed; only replay is lost. *Alternative:* keep the
  128 sources that contributed most and count the rest as unattributed. That preserves replay for the retained part, but it
  turns full measurements into partial ones and needs v4 branch-interval reindexing. I recommend the summary.
- **D3: weighting.**
  - The sums are weighted by share.
  - A row enters the percentile range (Low/High) only when its share is at least 0.5, so heavily partial rows can't widen the
    range.
  - Partial hours are attributed minutes only, a lower bound that is never extrapolated, so a partial row biases its own ratio
    low. Weighting by share bounds that bias.
- **D4: a queue attempt with no bounds gives `partial` with share 0.** The hours are kept as an upper bound and shown, but they
  never reach calibration. *Alternative:* `unmeasured` with N/A, which loses the hours again; that is what the issue is against.
- **D5: the confidence is stored in the trailer and the ledger, not on the card.** Adding it to the card would change the
  `construct/vocabulary/issue.cue` schema and the card codec for a field nothing on the card reads.
  - Legacy repos have no trailer, so they get the frontmatter field. That needs the cue field `actual_confidence` (an optional
    string), an additive change.
  - Alternative: put it on the card too, for `sdlc project status`. I'd defer that until something reads it.
- **`--no-actual` stays** for genuinely inapplicable windows (item 4). Its help text stops presenting it as the fix for a
  measurement failure.

## State design table

Actions: close (or milestone-close), `sdlc actual` (preview), `sdlc issue correct-actual`, close with `--no-actual`.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| All windows attributed, snapshot within bounds | `exact` | none | none |
| Snapshot or source count over the bound | `exact`, with a "not replayable" warning | none; a later overlapping issue degrades to `unmeasured` | the ability to replay this close |
| Unknown or detached branch intervals, or one source's attribution failed | `partial s` with the reasons | `correct-actual` if the operator knows the true value | unattributed minutes (recorded, not counted) |
| Resting or stage minutes (D1) | `partial s` | none | the same |
| Queue attempts with unknown provenance but known bounds | `partial s` | none | queue time cannot be subtracted, so it is overcounted inside the unattributed share |
| Legacy attempt with no bounds | `partial 0`: hours shown as an upper bound, calibration weight 0 | `correct-actual` | the row's calibration value |
| Input or replay error, telemetry gap, empty window, no window, lineage cycle, duplicate attempt | `unmeasured (<reason>)`, N/A, the recovery printed as a warning | recover the evidence, then `correct-actual` | the hours, until corrected |
| Measurement panics or the process dies mid-close | unchanged: nothing is written before the evidence commit | rerun close | none |

Covered: every refusal shape in the issue plus the residual errors. Not covered: changing how the measurement itself attributes
time. This plan only stops discarding what has already been attributed.

## Envelope (ARCH-*)

- **ARCH-PURE:** `classifyActual`, `CalibrationWeight` and the weighted calibration are pure. The IO changes are limited to
  turning returned errors into notes at three existing seams.
- **ARCH-EVOLVE:** the ledger columns are appended, and an old row reads as exact. The trailer is new and additive.
  `MeasurementSummary` is a new optional member of the boundary-ledger round `actual_inputs`; one of `inputs` or `summary` is
  present. The round decoder already preserves unknown members, which a test verifies. An older reader that sees a summary and no
  inputs treats the round as having no trustworthy saved window, a safe fallback. The frontmatter field is an optional addition
  to the cue.
- **ARCH-CONSTRAINTS:** close runs in batch. The classification adds O(notes) work. Building the summary is one pass over the
  sources, O(S). No new IO.
- **ARCH-FUNERAL:** no new artifact family. The summary is no larger than the snapshot it replaces, and the bound is now on the
  writer.
- **ARCH-SECURE:** the reasons are error strings we produce ourselves, and they reach a trailer. They must not contain transcript
  content; the error texts carry IDs and timestamps only, which a test asserts.
- **ARCH-ORDER:** no state persists between events because close is a single pass. The evidence commit is still the one write
  that makes the result durable.
- **ARCH-MOCK:** the existing fakes and fixtures are reused, with no new external surface.
- **ARCH-PURPOSE:** every refusal shape in Done-when is covered in this one issue. Calibration and docs ship in M2, not as a
  follow-up.
- **ARCH-DRY:** one classifier and one renderer are shared by close, milestone-close and `sdlc actual`.

## Replay fixtures (Done-when)

| Window | Shape | Fixture | Expected |
|---|---|---|---|
| pair#856, #812, #863 | resting-branch claim transition | `activetime/testdata/pair856`, `pair812`, `pair863` (exist) | `partial`, with resting minutes as the unattributed share (D1) |
| pair#830 | stage branch | `branch_staging_test.go` (in code) | `partial` |
| pair#851 | more than 128 sources | new: a recorded source manifest (IDs, sizes and windows with identity scrubbed) extracted from the refusal, or failing that a generator shaped from its printed counts | `exact`, with a summary |
| pair#838 | the 64 KiB proof (from lessons) | new: a fixture shaped from the refusal's printed counts | `exact`, with a summary |

Each fixture runs end to end, through the real resolver and engine and then `classifyActual`, using the close seam with a fake
transcript tree. #851 and #838 have no captures in the repo. I'll get their printed refusal text from pair's issue Logs. If the
raw sources can't be recovered, the fixture is synthetic and shaped from the printed counts, and its README says so.

## Plan

### M1: close never refuses on measurement

- [ ] **Task 1: `actualConfidence` and `classifyActual`**, test-first in `actualconfidence_test.go`. Table-driven: every status,
  every note kind, share arithmetic, rendering, and a test that the reason text carries no transcript content.
- [ ] **Task 2: the `unknown-excluded` note.** `SelectBranchActivity` notes the interval instead of returning an error. Update
  `branch_*_test.go`, plus a new test: concurrent claims with a detached interval give the windows before and after it, plus the
  note.
- [ ] **Task 3: a per-source attribution failure becomes `source-unattributed`.** In `actualbranch.go`, a lineage cycle stays an
  error. Tests cover a missing owner and an ambiguous owner, each giving `partial`.
- [ ] **Task 4: oversize gives `MeasurementSummary`.** Type the `oversize()` error, build the summary, and retain it in the
  ledger round. Tests: the v4 cap at 129 sources and at 128 KiB + 1 gives `exact` with a summary;
  `TestMeasurementOversizeCountedRefusal` is replaced by a test of the summary; an older round-reader preserves the summary
  member; an overlapping later issue gives `unmeasured` (not a refusal).
- [ ] **Task 5: queue evidence degrades** (D4). Unknown attempts with bounds give `partial`; with no bounds, `partial 0`. Hard
  errors stay `unmeasured`. Reuse `testdata/actualqueue`.
- [ ] **Task 6: close adopts every outcome.** `resolveOmittedActual` takes `actualConfidence`; the info line renders it; the
  trailer, frontmatter field and cue field are added; `explainActual` stops calling `exitWithCode(1)`. Tests: close end to end
  for each level, in tracker and legacy repos. Run `TestAdoptLineNoGatesigCollision` with the new renderings. `--no-actual` is
  unchanged.
- [ ] **Task 7: replay fixtures** for pair#856, #812, #863, #830, #851 and #838 (table above), each through the close seam.
- [ ] **M1**: milestone-close.

### M2: calibration reads confidence; docs

- [ ] **Task 8: the ledger columns** `confidence` and `attributed_share`. Old rows read as exact. Test the round trip, including
  an old header.
- [ ] **Task 9: weighted calibration** in `CalibrateForecast`, `ConvergeFactors` and `DriftVerdict` (D3). Tests: an exact row and
  a partial 0.5 row give the weighted point; a `partial 0` row has no effect; only rows with share ≥ 0.5 enter the quartiles; the
  reported N counts the partial rows.
- [ ] **Task 10: help and atlas.** Update `helptext/close.md`, `actual.md` and `milestone-close.md` (the confidence levels, and
  `--no-actual` only for inapplicable windows). In the atlas, update `sdlc-binary.md` (branch selection and the v4 bound),
  `issue-lifecycle.md` and `ledger-landscape.md`.
- [ ] **M2**: close.
