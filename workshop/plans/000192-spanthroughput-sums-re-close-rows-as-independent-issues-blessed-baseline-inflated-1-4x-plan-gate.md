---
gate: plan-quality
issue: 192
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-07-29T22:37:23-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: Shared dedupe helper's contract is under-specified — routing driftSample through it as described breaks existing tests and changes drift semantics
          detail: |-
            driftSample keys blank-issue rows positionally (@row:N, drift.go:51-53) and filters
            trusted/model BEFORE the seen-check (drift.go:47-49). A helper keyed plainly on
            LedgerRow.Issue collapses the empty-Issue rows that drift_test.go:13,43,57 depend on,
            and deduping before filtering silently lets an untrusted newest row mask an older
            trusted one. State that the helper takes an already-filtered slice, absorbs the
            blank-issue fallback, and that drift_test.go passing unmodified is the guard.
          round: 1
        - id: PQ-2
          severity: Important
          title: Plan corrects the ariadne atlas but leaves the same wrong per-issue claim standing in brain SKILL.md
          detail: |-
            brain data/life/42shots/velocity/SKILL.md:93 states the ledger holds "one row per closed
            issue" — the exact fact this defect grew from, in the doc the recalibration loop reads
            (calibration-findings.md:30 analyses those rows). The plan already writes to brain for
            the re-bless, so add that line to the checklist (ARCH-PURPOSE).
          round: 1
        - id: PQ-3
          severity: Minor
          title: SpanMeasure.UntrustedRows denominator not addressed — projectthroughput.go:103 can print "12 of 8 rows"
          detail: |-
            The field list names Issues and RowsScanned but not UntrustedRows; if it stays
            raw-counted while the denominator becomes issues, the untrusted warning mixes
            denominators.
          round: 1
        - id: PQ-4
          severity: Minor
          title: atlas/workflow/sdlc-binary.md:210-219 also documents the sum-over-rows semantics and is not in the plan
          round: 1
        - id: PQ-5
          severity: Minor
          title: Span-boundary attribution unstated — a re-close inside the span carries cumulative hours worked before it
          detail: |-
            Keeping the last in-span row means pre-span hours count toward the span rate when an
            issue's closes straddle `from`. Acceptable, but worth naming as a known limit beside the
            append-only non-goal.
          round: 1
      blocked: true
---

# Gate ledger — ariadne#192 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-07-29T22:37:23-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] Shared dedupe helper's contract is under-specified — routing driftSample through it as described breaks existing tests and changes drift semantics
  driftSample keys blank-issue rows positionally (@row:N, drift.go:51-53) and filters
  trusted/model BEFORE the seen-check (drift.go:47-49). A helper keyed plainly on
  LedgerRow.Issue collapses the empty-Issue rows that drift_test.go:13,43,57 depend on,
  and deduping before filtering silently lets an untrusted newest row mask an older
  trusted one. State that the helper takes an already-filtered slice, absorbs the
  blank-issue fallback, and that drift_test.go passing unmodified is the guard.
- **PQ-2** [Important] Plan corrects the ariadne atlas but leaves the same wrong per-issue claim standing in brain SKILL.md
  brain data/life/42shots/velocity/SKILL.md:93 states the ledger holds "one row per closed
  issue" — the exact fact this defect grew from, in the doc the recalibration loop reads
  (calibration-findings.md:30 analyses those rows). The plan already writes to brain for
  the re-bless, so add that line to the checklist (ARCH-PURPOSE).
- **PQ-3** [Minor] SpanMeasure.UntrustedRows denominator not addressed — projectthroughput.go:103 can print "12 of 8 rows"
  The field list names Issues and RowsScanned but not UntrustedRows; if it stays
  raw-counted while the denominator becomes issues, the untrusted warning mixes
  denominators.
- **PQ-4** [Minor] atlas/workflow/sdlc-binary.md:210-219 also documents the sum-over-rows semantics and is not in the plan
- **PQ-5** [Minor] Span-boundary attribution unstated — a re-close inside the span carries cumulative hours worked before it
  Keeping the last in-span row means pre-span hours count toward the span rate when an
  issue's closes straddle `from`. Acceptable, but worth naming as a known limit beside the
  append-only non-goal.

## Open findings

- **PQ-1** [Important] Shared dedupe helper's contract is under-specified — routing driftSample through it as described breaks existing tests and changes drift semantics
- **PQ-2** [Important] Plan corrects the ariadne atlas but leaves the same wrong per-issue claim standing in brain SKILL.md
- **PQ-3** [Minor] SpanMeasure.UntrustedRows denominator not addressed — projectthroughput.go:103 can print "12 of 8 rows"
- **PQ-4** [Minor] atlas/workflow/sdlc-binary.md:210-219 also documents the sum-over-rows semantics and is not in the plan
- **PQ-5** [Minor] Span-boundary attribution unstated — a re-close inside the span carries cumulative hours worked before it
