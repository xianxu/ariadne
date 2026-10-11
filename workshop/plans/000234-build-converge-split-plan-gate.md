---
gate: plan-quality
issue: 234
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-10T17:49:47-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: Row split metadata (boundary_rounds, review_recipe, split_source) taken from the pre-dispatch measurement misses the finalizing round
          detail: adoptOmittedActual runs at close.go:420, before the dispatch. appendCalibrationRow runs in applyClose, after persistBoundaryRound (close.go:1404, then 1429). A first-round pass would write no-review, 0 rounds and an empty recipe. Keep the minutes from the measurement, and read rounds, recipe and source from the persisted ledger when the row is written.
          family: measure-before-dispatch-snapshot
          round: 1
        - id: PQ-2
          severity: Important
          title: D1 phases overlap when tailStart (last commit at or before the final dispatch) precedes the first dispatch
          detail: Every single-round boundary, and every REWORK re-run with no fix commits, gives build ending at the first dispatch and a tail starting earlier, at the last commit. The phases then overlap. State the precedence (for example, build ends at min(first dispatch, tailStart), and converge is empty when tailStart <= first dispatch), and test a single-round M1 followed by M2.
          family: cut-rule-precedence-undefined
          round: 1
        - id: PQ-3
          severity: Minor
          title: nowRFC3339 is a plain function (reviewsidecar.go:130), not a fakeable seam; persistBoundaryRound takes its timestamp as a parameter
          family: unbacked-existing-code-claim
          round: 1
        - id: PQ-4
          severity: Minor
          title: 'Stamping Dispatched on a failed dispatch has no effect: DispatchError returns before persistBoundaryRound, and the res() paths build a fresh reviewResult'
          family: unbacked-existing-code-claim
          round: 1
        - id: PQ-5
          severity: Minor
          title: 'BoundaryRound.Finalizing has no source field; map it from Round.Reviewed and take the LAST finalizing round per boundary (re-close, #183)'
          family: unbacked-existing-code-claim
          round: 1
        - id: PQ-6
          severity: Minor
          title: Plan does not say how a malformed or hand-edited dispatched timestamp is handled (fall back to the finalize time and the finalize source)
          family: untrusted-persisted-input
          round: 1
        - id: PQ-7
          severity: Minor
          title: Tasks 2, 4, 5 and 7 list test cases in prose; compress to one strategy line per risky function (property-test the Split invariant over random rounds and segments)
          family: plan-test-prose-enumeration
          round: 1
        - id: PQ-8
          severity: Minor
          title: 'Writing the ratio column as ''-'' on est-0 rows: confirm no brain velocity script parses that column as a number'
          family: ledger-consumer-sweep
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-10-10T17:51:10-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: rowSplitMeta reads the persisted ledger in applyClose (close.go:1429 after persist at 1404); minutes stay from the measurement.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: D1 precedence stated (build ends at min(firstDispatch, tailStart)); the table and property test cover single-round M1 then M2.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: dispatchNow package var is the seam.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Plan states a failed dispatch returns before persist; no stamp.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: Finalizing maps from Reviewed; the last finalizing round per boundary counts.
          round: 2
        - id: PQ-6
          disposition: addressed
          note: Malformed stamp, or dispatch after finalize, falls back to finalize with source finalize.
          round: 2
        - id: PQ-7
          disposition: addressed
          note: Strategy lines plus a property test over the Split invariant.
          round: 2
        - id: PQ-8
          disposition: addressed
          note: 'Consumer sweep stated: only ParseRows reads the ratio, and it recomputes it.'
          round: 2
      blocked: false
    - "n": 3
      timestamp: "2026-10-10T17:52:12-07:00"
      agent: claude
      blocked: false
      protocol_error: no valid findings block
content_hash: 42f4ea2c078b71b4ca613512f999f1d9145a931feff2148d53b5122933b4309e
---

# Gate ledger — ariadne#234 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-10T17:49:47-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `measure-before-dispatch-snapshot` Row split metadata (boundary_rounds, review_recipe, split_source) taken from the pre-dispatch measurement misses the finalizing round
  adoptOmittedActual runs at close.go:420, before the dispatch. appendCalibrationRow runs in applyClose, after persistBoundaryRound (close.go:1404, then 1429). A first-round pass would write no-review, 0 rounds and an empty recipe. Keep the minutes from the measurement, and read rounds, recipe and source from the persisted ledger when the row is written.
- **PQ-2** [Important] `cut-rule-precedence-undefined` D1 phases overlap when tailStart (last commit at or before the final dispatch) precedes the first dispatch
  Every single-round boundary, and every REWORK re-run with no fix commits, gives build ending at the first dispatch and a tail starting earlier, at the last commit. The phases then overlap. State the precedence (for example, build ends at min(first dispatch, tailStart), and converge is empty when tailStart <= first dispatch), and test a single-round M1 followed by M2.
- **PQ-3** [Minor] `unbacked-existing-code-claim` nowRFC3339 is a plain function (reviewsidecar.go:130), not a fakeable seam; persistBoundaryRound takes its timestamp as a parameter
- **PQ-4** [Minor] `unbacked-existing-code-claim` Stamping Dispatched on a failed dispatch has no effect: DispatchError returns before persistBoundaryRound, and the res() paths build a fresh reviewResult
- **PQ-5** [Minor] `unbacked-existing-code-claim` BoundaryRound.Finalizing has no source field; map it from Round.Reviewed and take the LAST finalizing round per boundary (re-close, #183)
- **PQ-6** [Minor] `untrusted-persisted-input` Plan does not say how a malformed or hand-edited dispatched timestamp is handled (fall back to the finalize time and the finalize source)
- **PQ-7** [Minor] `plan-test-prose-enumeration` Tasks 2, 4, 5 and 7 list test cases in prose; compress to one strategy line per risky function (property-test the Split invariant over random rounds and segments)
- **PQ-8** [Minor] `ledger-consumer-sweep` Writing the ratio column as '-' on est-0 rows: confirm no brain velocity script parses that column as a number

## Round 2 — 2026-10-10T17:51:10-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — rowSplitMeta reads the persisted ledger in applyClose (close.go:1429 after persist at 1404); minutes stay from the measurement.
- PQ-2 — addressed — D1 precedence stated (build ends at min(firstDispatch, tailStart)); the table and property test cover single-round M1 then M2.
- PQ-3 — addressed — dispatchNow package var is the seam.
- PQ-4 — addressed — Plan states a failed dispatch returns before persist; no stamp.
- PQ-5 — addressed — Finalizing maps from Reviewed; the last finalizing round per boundary counts.
- PQ-6 — addressed — Malformed stamp, or dispatch after finalize, falls back to finalize with source finalize.
- PQ-7 — addressed — Strategy lines plus a property test over the Split invariant.
- PQ-8 — addressed — Consumer sweep stated: only ParseRows reads the ratio, and it recomputes it.

## Round 3 — 2026-10-10T17:52:12-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Open findings

(none — every finding has been disposed)
