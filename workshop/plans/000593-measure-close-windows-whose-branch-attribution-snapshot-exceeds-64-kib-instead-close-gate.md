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

## Open findings

- **BR-1** [Important] `snapshot-bound-refusal-not-total` (kind: defect) A snapshot over the 512 KiB structural cap still fails late inside MarshalJSON instead of the counted refusal
- **BR-2** [Minor] `helptext-wrap-width` (kind: defect) The edited paragraph in helptext/actual.md has an over-long line and broken wrapping
- **BR-3** [Minor] `plan-checklist-drift` (kind: defect) The plan file's Plan checklist is still unchecked though the issue marks every item done
