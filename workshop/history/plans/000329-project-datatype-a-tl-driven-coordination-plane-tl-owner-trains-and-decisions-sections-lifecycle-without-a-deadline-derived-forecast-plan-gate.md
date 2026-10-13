---
gate: plan-quality
issue: 329
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-12T16:42:03-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: D1 silently renames terminal status done to closed; archived done records and hardcoded "done" sites are unaddressed
          detail: 'project.cue:24/38/124 and projectsetstatus.go:140 use "done"; archived projects (workshop/history/projects, peers) carry status: done and may carry deadline/planned_finish, which the D2 _|_ ban would now reject. State the rename explicitly and say how archived records are migrated or exempted, or keep "done".'
          family: unstated-vocabulary-rename
          round: 1
        - id: PQ-2
          severity: Important
          title: D9 keeps project throughput --bless, whose only reader is the calendar forecast being deleted
          detail: 'projectthroughput.go:1-6 says the blessed baseline "is read by the #182 calendar forecast", which M1 deletes. That leaves a brain file writer with no reader (ARCH-FUNERAL), which contradicts the plan''s "ends two writers" claim. Remove --bless and the baseline file, or name D6''s forecast as the new reader.'
          family: orphaned-writer
          round: 1
        - id: PQ-3
          severity: Minor
          title: The cross-repo issue-card lookup behind IssueTerminal and forecast car hours is not named, nor its fake
          detail: mvp_scope and train refs are repo#N across repos; in tracker repos the cards live on the issue-tracker branch. Name the existing resolver function and the fake that the guard and forecast tests use (ARCH-MOCK).
          family: external-seam-unnamed
          round: 1
        - id: PQ-4
          severity: Minor
          title: M2 enumerates test cases in prose instead of one strategy line per risky function
          detail: 'Compress to: the trains parser over hand-authored markdown, table-driven plus fuzz seeded with malformed rows (never panics, always a typed error); the scheduler, checked with DAG invariants (critical path is at least each train''s sum, and a cycle is always reported).'
          family: test-prose-enumeration
          round: 1
        - id: PQ-5
          severity: Minor
          title: No stated non-goals; the Spec's open questions are not declared out of scope
          family: missing-non-goals
          round: 1
        - id: PQ-6
          severity: Minor
          title: show/status gain a best-effort forecast that does per-car card lookups; no latency bound or fallback is stated
          family: operating-envelope-unstated
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-10-12T16:44:08-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: not-addressed
          note: 'D1 never mentions done; only Revisions claims it does. Eight archived records across the fleet carry status: done, and two of them (ariadne project-management-primitive, metis arena2-playground-s6e7) also carry deadline fields that D2''s ban rejects. find/resolve read history/projects. State the rename in D1 and the exemption or migration for archived records, or keep done.'
          round: 2
        - id: PQ-2
          disposition: addressed
          note: D9 deletes project throughput, SpanThroughput, ThroughputBaseline and ParseBaselineTSV; the tsv is left as history.
          round: 2
        - id: PQ-3
          disposition: not-addressed
          note: D8 says "injected from the issue lookup" but still does not name the resolver function (tracker-branch cards, cross-repo) or the fake. Minor, carried forward to close.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: M2 now gives one strategy line for the parser (table-driven plus fuzz) and one for the scheduler (DAG invariants).
          round: 2
        - id: PQ-5
          disposition: addressed
          round: 2
        - id: PQ-6
          disposition: addressed
          note: 'Envelope paragraph: about 50 cars, under a second, a lookup error becomes a note and never fails the verb.'
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-10-12T16:44:47-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: D1 states the done→closed rename, why archived records are exempt (with a pinning test), and the hardcoded sites; projectsetstatus.go:140 confirmed.
          round: 3
        - id: PQ-3
          disposition: addressed
          note: D8 names projectIssueLookupFn (projectstatus.go:24) and the map-backed fake at projectclose_test.go:56; the forecast reuses it.
          round: 3
      blocked: false
    - "n": 4
      timestamp: "2026-10-12T16:45:35-07:00"
      agent: claude
      blocked: false
      protocol_error: no valid findings block
content_hash: c4bf241bd2e99548e8aea5ac90d6b43aff8a8a3b99d15103c71d536bb59f0001
---

# Gate ledger — ariadne#329 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T16:42:03-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `unstated-vocabulary-rename` D1 silently renames terminal status done to closed; archived done records and hardcoded "done" sites are unaddressed
  project.cue:24/38/124 and projectsetstatus.go:140 use "done"; archived projects (workshop/history/projects, peers) carry status: done and may carry deadline/planned_finish, which the D2 _|_ ban would now reject. State the rename explicitly and say how archived records are migrated or exempted, or keep "done".
- **PQ-2** [Important] `orphaned-writer` D9 keeps project throughput --bless, whose only reader is the calendar forecast being deleted
  projectthroughput.go:1-6 says the blessed baseline "is read by the #182 calendar forecast", which M1 deletes. That leaves a brain file writer with no reader (ARCH-FUNERAL), which contradicts the plan's "ends two writers" claim. Remove --bless and the baseline file, or name D6's forecast as the new reader.
- **PQ-3** [Minor] `external-seam-unnamed` The cross-repo issue-card lookup behind IssueTerminal and forecast car hours is not named, nor its fake
  mvp_scope and train refs are repo#N across repos; in tracker repos the cards live on the issue-tracker branch. Name the existing resolver function and the fake that the guard and forecast tests use (ARCH-MOCK).
- **PQ-4** [Minor] `test-prose-enumeration` M2 enumerates test cases in prose instead of one strategy line per risky function
  Compress to: the trains parser over hand-authored markdown, table-driven plus fuzz seeded with malformed rows (never panics, always a typed error); the scheduler, checked with DAG invariants (critical path is at least each train's sum, and a cycle is always reported).
- **PQ-5** [Minor] `missing-non-goals` No stated non-goals; the Spec's open questions are not declared out of scope
- **PQ-6** [Minor] `operating-envelope-unstated` show/status gain a best-effort forecast that does per-car card lookups; no latency bound or fallback is stated

## Round 2 — 2026-10-12T16:44:08-07:00 (claude) — BLOCKED

### Disposed

- PQ-1 — not-addressed — D1 never mentions done; only Revisions claims it does. Eight archived records across the fleet carry status: done, and two of them (ariadne project-management-primitive, metis arena2-playground-s6e7) also carry deadline fields that D2's ban rejects. find/resolve read history/projects. State the rename in D1 and the exemption or migration for archived records, or keep done.
- PQ-2 — addressed — D9 deletes project throughput, SpanThroughput, ThroughputBaseline and ParseBaselineTSV; the tsv is left as history.
- PQ-3 — not-addressed — D8 says "injected from the issue lookup" but still does not name the resolver function (tracker-branch cards, cross-repo) or the fake. Minor, carried forward to close.
- PQ-4 — addressed — M2 now gives one strategy line for the parser (table-driven plus fuzz) and one for the scheduler (DAG invariants).
- PQ-5 — addressed
- PQ-6 — addressed — Envelope paragraph: about 50 cars, under a second, a lookup error becomes a note and never fails the verb.

## Round 3 — 2026-10-12T16:44:47-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — D1 states the done→closed rename, why archived records are exempt (with a pinning test), and the hardcoded sites; projectsetstatus.go:140 confirmed.
- PQ-3 — addressed — D8 names projectIssueLookupFn (projectstatus.go:24) and the map-backed fake at projectclose_test.go:56; the forecast reuses it.

## Round 4 — 2026-10-12T16:45:35-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Open findings

(none — every finding has been disposed)
