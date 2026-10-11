---
gate: boundary-review
issue: 234
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-10T18:36:44-07:00"
      dispatched: "2026-10-10T18:33:33-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: computeActualFor reads the gate ledger via env plansDir(), ignoring close/milestone-close --plans-dir
          detail: issueBoundaryLedger (actualsplit.go:28) uses global plansDir() while persistBoundaryRound and closeMetrics/rowSplitMeta use f.plansDir(); with --plans-dir the split sees no rounds (no-review, all build) but the row records source dispatch and the round count, and the E1 increment drops converge. Thread the plans dir through computeActualFor and the computeActualForCloseFn seam.
          family: plans-dir-single-source
          round: 1
        - id: BR-2
          severity: Minor
          title: dispatchNow var sits between dispatchBoundaryReview's doc comment and the func, stealing its godoc
          family: doc-comment-attachment
          round: 1
        - id: BR-3
          severity: Minor
          title: splitRounds drops a round with an unparseable Timestamp but rowSplitMeta still counts it as a stamped round
          detail: A hand-edited ledger can yield a row labelled dispatch whose split never saw that round; apply one parse rule in both or downgrade the source to finalize.
          family: split-meta-parse-parity
          round: 1
        - id: BR-4
          severity: Minor
          title: Plan claims a property test for no-converge-without-post-dispatch-commit; only a table case covers it
          family: test-claims-match-tests
          round: 1
        - id: BR-5
          severity: Minor
          title: The computeActualFor integration fixture never produces converge greater than 0 (no REWORK round)
          family: integration-fixture-coverage
          round: 1
        - id: BR-6
          severity: Minor
          title: Calibration rows grow by 5 cells per close with no stated growth bound (ARCH-FUNERAL note)
          family: ledger-growth-bound
          round: 1
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 2
      timestamp: "2026-10-10T18:40:12-07:00"
      dispatched: "2026-10-10T18:38:50-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: plans threaded through computeActualFor, issueBoundaryLedger and the close seam (close.go:444, close.go:1656); TestComputeActualReadsTheGivenPlansDir goes red without it. Close-site stubs ignore the arg, so the call-site threading is pinned only at the engine level.
          round: 2
        - id: BR-2
          disposition: addressed
          note: 'milestoneclose.go: dispatchNow precedes the doc comment, which now sits directly on dispatchBoundaryReview.'
          round: 2
        - id: BR-3
          disposition: addressed
          note: cutTimes is the shared rule (actualsplit.go:67); the new "an unparseable round" table case expects 1 round and dispatch, and is red without the fix.
          round: 2
        - id: BR-4
          disposition: addressed
          note: split_test.go:156 asserts each non-open converge ends at an issue commit, which implies no converge without a post-dispatch commit.
          round: 2
        - id: BR-5
          disposition: addressed
          note: TestComputeActualCountsConvergeAfterRework yields build 20m and converge 30m from an open REWORK round.
          round: 2
        - id: BR-6
          disposition: addressed
          note: ledger.go:110 states about 150 B per close, no removal (calibration history), linear reader, year-split remedy.
          round: 2
      findings:
        - id: BR-7
          severity: Minor
          title: groupBoundaries skips BoundaryAll rounds but rowSplitMeta counts them, a second usability filter outside cutTimes
          detail: 'This is the 2nd finding in family split-meta-parse-parity. The rule: every predicate that decides whether a round can cut the split lives in cutTimes, and nothing downstream filters again. Fix it by moving the Boundary == "*" exclusion into cutTimes (or splitRounds) and dropping it from groupBoundaries. Prevalence: there is one other filter site. Production is unaffected today because the only BoundaryAll producer (the plan-gate seed, boundaryledger.go:124) is NoCap.'
          family: split-meta-parse-parity
          round: 2
      boundary: M1
      recipe: milestone-review
      reviewed: a7b840c0a5fed76374beec640a8babbf542874f3
      blocked: false
---

# Gate ledger — ariadne#234 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-10T18:36:44-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `plans-dir-single-source` computeActualFor reads the gate ledger via env plansDir(), ignoring close/milestone-close --plans-dir
  issueBoundaryLedger (actualsplit.go:28) uses global plansDir() while persistBoundaryRound and closeMetrics/rowSplitMeta use f.plansDir(); with --plans-dir the split sees no rounds (no-review, all build) but the row records source dispatch and the round count, and the E1 increment drops converge. Thread the plans dir through computeActualFor and the computeActualForCloseFn seam.
- **BR-2** [Minor] `doc-comment-attachment` dispatchNow var sits between dispatchBoundaryReview's doc comment and the func, stealing its godoc
- **BR-3** [Minor] `split-meta-parse-parity` splitRounds drops a round with an unparseable Timestamp but rowSplitMeta still counts it as a stamped round
  A hand-edited ledger can yield a row labelled dispatch whose split never saw that round; apply one parse rule in both or downgrade the source to finalize.
- **BR-4** [Minor] `test-claims-match-tests` Plan claims a property test for no-converge-without-post-dispatch-commit; only a table case covers it
- **BR-5** [Minor] `integration-fixture-coverage` The computeActualFor integration fixture never produces converge greater than 0 (no REWORK round)
- **BR-6** [Minor] `ledger-growth-bound` Calibration rows grow by 5 cells per close with no stated growth bound (ARCH-FUNERAL note)

## Round 2 — 2026-10-10T18:40:12-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — plans threaded through computeActualFor, issueBoundaryLedger and the close seam (close.go:444, close.go:1656); TestComputeActualReadsTheGivenPlansDir goes red without it. Close-site stubs ignore the arg, so the call-site threading is pinned only at the engine level.
- BR-2 — addressed — milestoneclose.go: dispatchNow precedes the doc comment, which now sits directly on dispatchBoundaryReview.
- BR-3 — addressed — cutTimes is the shared rule (actualsplit.go:67); the new "an unparseable round" table case expects 1 round and dispatch, and is red without the fix.
- BR-4 — addressed — split_test.go:156 asserts each non-open converge ends at an issue commit, which implies no converge without a post-dispatch commit.
- BR-5 — addressed — TestComputeActualCountsConvergeAfterRework yields build 20m and converge 30m from an open REWORK round.
- BR-6 — addressed — ledger.go:110 states about 150 B per close, no removal (calibration history), linear reader, year-split remedy.

### Raised

- **BR-7** [Minor] `split-meta-parse-parity` groupBoundaries skips BoundaryAll rounds but rowSplitMeta counts them, a second usability filter outside cutTimes
  This is the 2nd finding in family split-meta-parse-parity. The rule: every predicate that decides whether a round can cut the split lives in cutTimes, and nothing downstream filters again. Fix it by moving the Boundary == "*" exclusion into cutTimes (or splitRounds) and dropping it from groupBoundaries. Prevalence: there is one other filter site. Production is unaffected today because the only BoundaryAll producer (the plan-gate seed, boundaryledger.go:124) is NoCap.

## Open findings

- **BR-7** [Minor] `split-meta-parse-parity` groupBoundaries skips BoundaryAll rounds but rowSplitMeta counts them, a second usability filter outside cutTimes
