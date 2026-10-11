# Boundary Review — ariadne#234 (milestone M1)

| field | value |
|-------|-------|
| issue | 234 — Split actuals at the first boundary review: build time vs converge time |
| repo | ariadne |
| issue file | workshop/issues/000234-build-converge-split.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 9769e3c79f91c3b761e43176375699b03dd94ecb..2e8d15bcdc73ace5ba112d566530f55c2fa79860 |
| command | sdlc milestone-close --issue 234 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-10T18:36:44-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Base-diff read; now checking the gate-ledger round semantics (Boundary labels, never-ran rounds, `Reviewed`) that the split depends on.

Targeted tests pass; reviewing the split test table and one edge case (multi-boundary ordering) before writing the verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 delivers what the plan promised. Each boundary round now records when its review was dispatched. A pure timeline and split (`activetime.BuildTimeline`/`Split`) encode the D1 cut rule with its stated precedence. `computeActualFor` adopts build + converge. `milestone-close` now measures and adopts its own increment (E1), and the deviation check compares a typed milestone value against that increment. The five new ledger columns follow the append-only width rule, and rows with no estimate no longer count as calibration points (E3). Targeted tests pass: `activetime`, `estimate`, `gatestate`, and `cmd/sdlc -run 'Split|Dispatch|AdoptOmitted|ActualDeviation|CalibrationLine|AppendCalibration|PrintActual|FormatAdopt'`.

One Important finding stands in the way of a clean SHIP. The measurement reads the gate ledger from a different directory than the one `close` / `milestone-close` write it to, so `--plans-dir` produces a split and a row that disagree. The fix is cheap.

**1. Strengths**
- `cmd/sdlc/internal/activetime/split.go:93-147`: D1/PQ-2 lives in one pure function. `emit` clamps each phase to the end of the previous one, so phases never overlap. A 500-case property test (`split_test.go:122`) checks that phases are ordered, contiguous and cover the window, and that every minute is accounted for.
- `split.go:79-83` and `:141-144`: the "past milestone measured later" bug found during implementation is fixed in the right place. Trailing build goes to `current` only while `current` is still open. `TestSplitForBoundary` pins this (M1 measured after M2 does not absorb M2's work).
- `actualsplit.go:59-61`: a rejected stamp falls back to the finalize time and says so in the source. A dispatch stamp later than its finalize is distrusted (`dispatchOf`). Never-ran rounds are `NoCap` (`boundaryledger.go:166`), so they correctly carry no stamp and no cut. This is good ARCH-SECURE handling of a hand-editable artifact.
- `close.go:1655-1661`: removing the milestone refusal is backed by a real increment (`Timeline.For`), not just by deleting the guard. Both the adopt test and the deviation test assert the seam is asked for `M2`.
- `estimate/ledger.go`: the widths are named constants (`flowCols`, `splitCols`). Every historical width (10/20/23) is round-trip tested, and the 23→28 header upgrade is tested too.

**2. Critical findings**
None.

**3. Important findings**
- **The measurement ignores `--plans-dir`** (`actualsplit.go:28` uses the global `plansDir()`, which only reads the environment).
  - `close` and `milestone-close` expose `--plans-dir` (`close.go:180`, `milestoneclose.go:123`).
  - `persistBoundaryRound` and `closeMetrics` (→ `rowSplitMeta`) read `f.plansDir()`. `computeActualForCloseFn` → `issueBoundaryLedger` does not.
  - **What goes wrong:** with `--plans-dir X`, the measurement finds no rounds, labels the split `no-review` and counts everything as build. The row's metadata comes from X, so `rowSplitSource` writes `dispatch` with converge 0. The row is mislabelled, and the E1 increment ignores real converge.
  - **Fix:** thread the plans dir through `computeActualFor` (or `issueBoundaryLedger`) and the `computeActualForCloseFn` seam, pass `f.plansDir()` from close, and keep `plansDir()` for `sdlc actual`.
  - ARCH-DRY: one source for "which ledger is this issue's".

**4. Minor findings**
- `milestoneclose.go:620-627`: `var dispatchNow` sits between `dispatchBoundaryReview`'s doc comment and the function, so the whole comment now documents the var and the function has none. Move the var above the comment.
- `splitRounds` (`actualsplit.go:55`) silently drops a round whose finalize `Timestamp` won't parse, but `rowSplitMeta` still counts it and treats it as stamped. With a hand-edited ledger, the row can claim `dispatch` while the split saw no such round. Either have `rowSplitMeta` apply the same parse rule, or downgrade the source to `finalize`.
- The plan (Task 2) says the property test asserts "no converge when there is no commit after the first dispatch". The property test doesn't check this; only the table case "REWORK without fix commits" does. Add the property or correct the plan prose.
- The integration fixture (`actualsplit_test.go`) only has a first-round pass, so converge is 0 in every case. A non-finalizing REWORK round in the fixture would exercise the real `Reviewed`→`Finalizing` mapping in `splitRounds` with converge > 0.
- ARCH-FUNERAL: each calibration row grows by 5 cells (about 30 bytes per close) in a ledger that is never pruned. This is negligible, but the bound isn't stated anywhere. A one-line note in the atlas ledger paragraph would cover it.

**5. Test coverage notes**
- The tests check the rule itself, not mocks. The pure `BuildTimeline`/`Split` table covers every D1 shape: no rounds, first-round pass, REWORK with fixes, REWORK without fixes, M1 then M2, open REWORK, a missing stamp, and a stamp after its finalize. Properties on top: conservation and contiguity.
- The integration path uses real git with a transcript fixture (`TestComputeActualSplitsAtBoundaries`).
- `IssueCommitTimes` is tested with tracker refs and citations, and tracker card commits for the issue are correctly included.
- Gaps: the `--plans-dir` divergence above, and converge > 0 through the IO shell.

**6. Architectural notes for upcoming work (M2)**
- **ARCH-DRY:** pass, apart from the plans-dir divergence. The single `Compute` result feeds both the actual and the split.
- **ARCH-PURE:** pass. Timeline, split, `rowSplitMeta` and `rowSplitSource` are pure. The IO shell is `issueBoundaryLedger` plus `computeActualFor`.
- **ARCH-PURPOSE:** pass for M1's scope. M2 must make drift read only `HasSplit()` rows, est against build. Until then, drift still mixes old-rule and new-rule actuals for multi-milestone issues closed after this lands. That is acceptable at the milestone boundary, but it is the point of M2.
- **ARCH-MOCK:** pass. Real git plus fixture transcripts; the `dispatchNow` seam is a clock, not a stateful dependency.
- **ARCH-CONSTRAINTS:** pass. The work is O(segments × phases) per close, with phases bounded by the number of rounds.
- **ARCH-SECURE:** pass with the one Minor above. A malformed stamp is handled visibly; a malformed finalize time is skipped quietly.
- **ARCH-ORDER:** pass. The timeline is derived from the event ledger on every measurement and holds no state between events. The persist order is explicit: the row metadata is read after the round is persisted (PQ-1).
- **ARCH-FUNERAL:** Minor note only. `dispatched` adds one field per round, and the ledger row grows by 5 cells.
- For `converge-report`: rows mark mixed recipes with `+`. Decide whether k groups by the joined string or by each recipe in it.

**7. Plan revision recommendations**
- Task 2: either change "property test … no converge when there is no commit after the first dispatch" to say the table case covers it, or add that property.
- Task 3: the test lives in `trackerref_test.go`, not `commit_test.go`. This is cosmetic, so only update the plan if a Revisions entry is being written anyway.

```findings
findings:
  - id: new
    severity: Important
    family: plans-dir-single-source
    title: |
      computeActualFor reads the gate ledger via env plansDir(), ignoring close/milestone-close --plans-dir
    detail: |
      issueBoundaryLedger (actualsplit.go:28) uses global plansDir() while persistBoundaryRound and closeMetrics/rowSplitMeta use f.plansDir(); with --plans-dir the split sees no rounds (no-review, all build) but the row records source dispatch and the round count, and the E1 increment drops converge. Thread the plans dir through computeActualFor and the computeActualForCloseFn seam.
  - id: new
    severity: Minor
    family: doc-comment-attachment
    title: |
      dispatchNow var sits between dispatchBoundaryReview's doc comment and the func, stealing its godoc
  - id: new
    severity: Minor
    family: split-meta-parse-parity
    title: |
      splitRounds drops a round with an unparseable Timestamp but rowSplitMeta still counts it as a stamped round
    detail: |
      A hand-edited ledger can yield a row labelled dispatch whose split never saw that round; apply one parse rule in both or downgrade the source to finalize.
  - id: new
    severity: Minor
    family: test-claims-match-tests
    title: |
      Plan claims a property test for no-converge-without-post-dispatch-commit; only a table case covers it
  - id: new
    severity: Minor
    family: integration-fixture-coverage
    title: |
      The computeActualFor integration fixture never produces converge greater than 0 (no REWORK round)
  - id: new
    severity: Minor
    family: ledger-growth-bound
    title: |
      Calibration rows grow by 5 cells per close with no stated growth bound (ARCH-FUNERAL note)
```

---

## Re-review — 2026-10-10T18:40:12-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 234 — Split actuals at the first boundary review: build time vs converge time |
| repo | ariadne |
| issue file | workshop/issues/000234-build-converge-split.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 9769e3c79f91c3b761e43176375699b03dd94ecb..a7b840c0a5fed76374beec640a8babbf542874f3 |
| command | sdlc milestone-close --issue 234 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-10T18:40:12-07:00 |
| verdict | SHIP |

## Review

Focus on the round-1 fix commit first, then the core files.

Round-1 fixes check out and targeted tests pass. One last check: whether `rowSplitMeta` and the split agree on `"*"` (all-boundary) rounds.

```verdict
verdict: SHIP
confidence: high
```

All six round-1 findings are fixed, and each fix is backed by evidence. BR-1 now passes `plans` through `computeActualFor`, `issueBoundaryLedger` and the `computeActualForCloseFn` seam. A new integration test, `TestComputeActualReadsTheGivenPlansDir`, would fail without that change: the ledger would not be found, so the source would be `no-review` instead of `dispatch`. BR-3 is fixed by `cutTimes`, one usability rule now shared by `splitRounds` and `rowSplitMeta`, and the new "an unparseable round" table case pins it. BR-5 now has a REWORK fixture that yields 30m of converge. BR-4, BR-2 and BR-6 are handled in the property test, the doc-comment placement and the ledger growth note. The targeted tests pass (`activetime`, `estimate`, `gatestate`, and the `cmd/sdlc` split, adopt and deviation tests). Only one Minor remains: a sibling in the parse-parity family, which does not block.

1. **Strengths**
   - `cutTimes` (`cmd/sdlc/actualsplit.go:67`) gives the split and the row's metadata a single source for which rounds count (ARCH-DRY), and its doc comment says why.
   - `splitFixture` (`actualsplit_test.go`) now takes a plans dir and a rounds builder, so one git and ledger fixture drives three integration scenarios without copy-paste.
   - The new property (`split_test.go:156`) is stronger than the plan's wording. A finalized converge ending at an issue commit later than the cut implies "no converge without a post-dispatch commit".
   - The `BuildTimeline` doc comment matches the code: the cut precedence, the tail rule, and a past milestone's trailing build tagged `""`.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `groupBoundaries` (`split.go`) also skips `Boundary == "*"` rounds, but `rowSplitMeta` does not. So "which rounds count" still has a second filter outside `cutTimes`. In production the only `BoundaryAll` round is the plan-gate seed, which is `NoCap` (`boundaryledger.go:124`), so the two agree today. A hand-edited ledger could still make them diverge, which is the BR-3 class.

5. **Test coverage notes**
   - Regression tests for BR-1, BR-3 and BR-5 all fail if their fixes are reverted (by inspection).
   - Close-path stubs of `computeActualForCloseFn` discard the new `plans` argument (`_`). A close that passed `""` would therefore still pass the tests. The engine-level test does pin the actual defect, so I am not raising this as a finding. A one-line assertion in `TestAdoptOmittedActualMilestoneAdoptsItsIncrement` that the seam receives `f.plansDir()` would close the gap.

6. **Architectural notes**
   - **ARCH-DRY:** pass. `cutTimes` is now the single rule; the Minor above is the one leftover.
   - **ARCH-PURE:** pass. `BuildTimeline` and `Split` are pure and tested without IO; `computeActualFor` is the thin shell around them.
   - **ARCH-PURPOSE:** pass for M1's scope. Drift and k are M2.
   - **ARCH-MOCK:** pass. The integration test uses a real git fixture, and the seam stubs are limited to close-decision tests.
   - **ARCH-CONSTRAINTS:** pass. The ledger is parsed linearly once per close, and the note states its size.
   - **ARCH-SECURE:** pass. A hand-edited ledger with a bad timestamp is skipped, and the row now agrees with that. A dispatch stamp later than the finalize falls back to `finalize` and is labelled as such, rather than inventing a value.
   - **ARCH-ORDER:** pass. Gate rounds are appended at finalize under the lock, and the timeline is a pure fold over them.
   - **ARCH-FUNERAL:** pass. The growth bound and its remedy are stated at `ledger.go:110`.
   - For M2: drift should read only rows whose `split_source` is `dispatch` or `finalize`, and k should come from the same `rowSplitMeta` recipe string. Neither should re-derive from the ledger.

7. **Plan revision recommendations:** none. The plan's property-test claim (BR-4) is now met by the test that exists.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      plans threaded through computeActualFor, issueBoundaryLedger and the close seam (close.go:444, close.go:1656); TestComputeActualReadsTheGivenPlansDir goes red without it. Close-site stubs ignore the arg, so the call-site threading is pinned only at the engine level.
  - id: BR-2
    disposition: addressed
    note: |
      milestoneclose.go: dispatchNow precedes the doc comment, which now sits directly on dispatchBoundaryReview.
  - id: BR-3
    disposition: addressed
    note: |
      cutTimes is the shared rule (actualsplit.go:67); the new "an unparseable round" table case expects 1 round and dispatch, and is red without the fix.
  - id: BR-4
    disposition: addressed
    note: |
      split_test.go:156 asserts each non-open converge ends at an issue commit, which implies no converge without a post-dispatch commit.
  - id: BR-5
    disposition: addressed
    note: |
      TestComputeActualCountsConvergeAfterRework yields build 20m and converge 30m from an open REWORK round.
  - id: BR-6
    disposition: addressed
    note: |
      ledger.go:110 states about 150 B per close, no removal (calibration history), linear reader, year-split remedy.
findings:
  - id: new
    severity: Minor
    family: split-meta-parse-parity
    title: |
      groupBoundaries skips BoundaryAll rounds but rowSplitMeta counts them, a second usability filter outside cutTimes
    detail: |
      This is the 2nd finding in family split-meta-parse-parity. The rule: every predicate that decides whether a round can cut the split lives in cutTimes, and nothing downstream filters again. Fix it by moving the Boundary == "*" exclusion into cutTimes (or splitRounds) and dropping it from groupBoundaries. Prevalence: there is one other filter site. Production is unaffected today because the only BoundaryAll producer (the plan-gate seed, boundaryledger.go:124) is NoCap.
```
