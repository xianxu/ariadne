---
gate: boundary-review
issue: 593
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 21405f87-d955-4351-af0a-6b110574a6a3
            repository: github.com/xianxu/ariadne
            issue_id: "000593"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/a338cd41baef65e5
            reviewed_head: ca9d0d20966837e3ad342f30490ade836cf3cd05
            slot: ariadne:2
            requested_at: "2026-10-24T13:50:55.796062Z"
          phase: finished
          pid: 54168
          started_at: "2026-10-24T13:50:55.968315Z"
          finished_at: "2026-10-24T14:14:35.518827Z"
          session_id: 09c8a668-7ea3-43c1-af26-36bfcf51b435
          outcome: returned
      "n": 1
      timestamp: "2026-10-24T07:14:36-07:00"
      dispatched: "2026-10-24T06:50:55-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: A snapshot over the 512 KiB structural cap still fails late inside MarshalJSON instead of the counted refusal
          detail: 'Validate (measurement.go:315) calls json.Marshal first, and encodeMeasurementObject''s checkMeasurementJSON refuses at 512 KiB with the generic "does not serialize ... error calling MarshalJSON" text, before the counted check is reached. Branch intervals and windows per source have no count bound, so a window within SnapshotLimitSources can reach this path, which is exactly what Done-when 2 rules out. No test covers it (the 128-source fixture uses 2 intervals per source). Fix: map the structural-cap error to the counted refusal (size or bound, version, counts) and add a many-interval fixture over 512 KiB that asserts the error contains no "does not serialize".'
          family: snapshot-bound-refusal-not-total
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: The edited paragraph in helptext/actual.md has an over-long line and broken wrapping
          detail: 'Line 66: "new method or version rather than silently drop a crossing execution span. Existing adopted v1/v2" runs past the paragraph''s column width.'
          family: helptext-wrap-width
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: The plan file's Plan checklist is still unchecked though the issue marks every item done
          detail: workshop/plans/000593-...-plan.md lists every step as "- [ ]" while the issue's Plan section shows "- [x]".
          family: plan-checklist-drift
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: c31854dd-be4c-469b-83ae-248068cc6ed6
            repository: github.com/xianxu/ariadne
            issue_id: "000593"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/d7e4613561548b1b
            reviewed_head: 2f61a831440f7c9175898603abe6a0337e44fe4a
            slot: ariadne:2
            requested_at: "2026-10-24T14:19:30.535958Z"
          phase: finished
          pid: 51973
          started_at: "2026-10-24T14:19:30.715974Z"
          finished_at: "2026-10-24T14:32:02.055332Z"
          session_id: 08073e54-9a00-4458-bfcf-d815a556f592
          outcome: returned
      "n": 2
      timestamp: "2026-10-24T07:32:03-07:00"
      dispatched: "2026-10-24T07:19:30-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Validate maps measurementSizeError to the counted refusal (measurement.go:320); reverting it makes TestMeasurementPastStructuralCapCountedRefusal fail with "does not serialize ... exceeds 524288 bytes".
          round: 2
        - id: BR-2
          disposition: addressed
          note: helptext/actual.md BRANCH ATTRIBUTION paragraph rewrapped; all lines now 76-78 columns, consistent with the surrounding text.
          round: 2
        - id: BR-3
          disposition: addressed
          note: Plan file checklist now shows every step as [x], matching the issue's Plan section, plus a Revisions entry for BR-1.
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: 2f61a831440f7c9175898603abe6a0337e44fe4a
      blocked: false
---

# Gate ledger — ariadne#593 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-24T07:14:36-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `snapshot-bound-refusal-not-total` (kind: defect) A snapshot over the 512 KiB structural cap still fails late inside MarshalJSON instead of the counted refusal
  Validate (measurement.go:315) calls json.Marshal first, and encodeMeasurementObject's checkMeasurementJSON refuses at 512 KiB with the generic "does not serialize ... error calling MarshalJSON" text, before the counted check is reached. Branch intervals and windows per source have no count bound, so a window within SnapshotLimitSources can reach this path, which is exactly what Done-when 2 rules out. No test covers it (the 128-source fixture uses 2 intervals per source). Fix: map the structural-cap error to the counted refusal (size or bound, version, counts) and add a many-interval fixture over 512 KiB that asserts the error contains no "does not serialize".
- **BR-2** [Minor] `helptext-wrap-width` (kind: defect) The edited paragraph in helptext/actual.md has an over-long line and broken wrapping
  Line 66: "new method or version rather than silently drop a crossing execution span. Existing adopted v1/v2" runs past the paragraph's column width.
- **BR-3** [Minor] `plan-checklist-drift` (kind: defect) The plan file's Plan checklist is still unchecked though the issue marks every item done
  workshop/plans/000593-...-plan.md lists every step as "- [ ]" while the issue's Plan section shows "- [x]".

## Round 2 — 2026-10-24T07:32:03-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Validate maps measurementSizeError to the counted refusal (measurement.go:320); reverting it makes TestMeasurementPastStructuralCapCountedRefusal fail with "does not serialize ... exceeds 524288 bytes".
- BR-2 — addressed — helptext/actual.md BRANCH ATTRIBUTION paragraph rewrapped; all lines now 76-78 columns, consistent with the surrounding text.
- BR-3 — addressed — Plan file checklist now shows every step as [x], matching the issue's Plan section, plus a Revisions entry for BR-1.

## Open findings

(none — every finding has been disposed)
