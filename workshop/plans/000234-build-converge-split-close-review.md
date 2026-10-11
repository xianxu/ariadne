# Boundary Review — ariadne#234 (whole-issue close)

| field | value |
|-------|-------|
| issue | 234 — Split actuals at the first boundary review: build time vs converge time |
| repo | ariadne |
| issue file | workshop/issues/000234-build-converge-split.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9769e3c79f91c3b761e43176375699b03dd94ecb..9f5d9e7b53585e4e6106e7ebd1eddcae6e26ed06 |
| command | sdlc close --issue 234 |
| reviewer | claude |
| timestamp | 2026-10-10T19:02:52-07:00 |
| verdict | SHIP |

## Review

Reviewing the M2 diff now; checking ledger parsing of the split_source column and README coverage.

```verdict
verdict: SHIP
confidence: high
```

I recommend SHIP. BR-7 is fixed at the rule level, as the family escalation asked. `cutTimes` (`cmd/sdlc/actualsplit.go:66-83`) now treats `gatestate.BoundaryAll` rounds as unusable, and `groupBoundaries` (`cmd/sdlc/internal/activetime/split.go:145`) no longer filters anything itself; its doc comment says the caller passes only rounds that can cut. That leaves one round filter, shared by `splitRounds` and `rowSplitMeta`. A test pins it: the new `"a round for every boundary"` case in `TestRowSplitMetaAndSource` expects `rowSplitMeta` to count 1 round. Before the fix it counted 2, so the case would fail without the fix.

M2 delivers what the Plan's M2 row promises:
- Drift now compares the estimate with build (`BuildRatio`), and `calibratable` only accepts rows that have a split with positive build.
- `estimate.ConvergeFactors` and `sdlc converge-report` report k per (flow, model, recipe).
- The brain's `velocity/SKILL.md` documents the new columns; I checked it on disk.
- The atlas has the new verb row, the file-map entries and a drift/k note.

The targeted tests (`internal/estimate`, `internal/activetime`, and the converge-report, split-meta, compute-actual, calibration-row and drift tests in `cmd/sdlc`) all pass. Nothing blocks.

1. **Strengths**
   - `cutTimes` is now the only rule deciding which rounds can cut, which is the class-level fix the escalation asked for. `lessons.md` records the rule.
   - `TestDrift_EstimateAgainstBuild` (`drift_test.go`) uses a fixture where est/actual ≈ 1 but est/build = 3. Drift must warn, so the test only passes if drift really reads build. It also checks that legacy and `manual` rows don't count toward n.
   - `ConvergeFactors` filters rows before deduping with `NewestPerIssue`, as that function's contract requires. Its test covers a re-closed issue (the newer row wins), legacy rows, typed-actual rows and quick rows.
   - `HasSplit` is the one predicate that both drift and the k report read, so the two can't disagree about which rows count.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `convergereport.go:52` `dashIfEmpty` duplicates the unexported `estimate.dash` (`ledger.go:262`). This is too small to need a fix.
   - The edited `DriftVerdict` doc comment in `drift.go` has one line that wasn't re-wrapped. Cosmetic only.
   - Rows with `split_source` `no-review` have an empty recipe, so they form their own `-` recipe group with k = 0. That keeps them from diluting real recipe groups, which is right, but the help text could say so explicitly.

5. **Test coverage:** this round adds unit tests for `ConvergeFactors`, drift against build, and the est-0 exclusion from M1. `TestConvergeReport` covers the report's IO shell against a `$WF_CALIB_LEDGER` fixture, both the empty case and the case with a split row. The `BoundaryAll` case has a regression test that would fail without the fix.

6. **Architecture**
   - ARCH-DRY: pass, apart from the `dashIfEmpty` Minor.
   - ARCH-PURE: pass. `ConvergeFactors` and `DriftVerdict` are pure; `converge-report` is a thin read-and-render wrapper.
   - ARCH-PURPOSE: pass. Every consumer of the calibration data (drift, the k report, the row's metadata) reads the split through `HasSplit`. The backfill was moved to ariadne#323 by a recorded scope revision; that's a separate piece of work, not the issue's core purpose being deferred.
   - ARCH-MOCK: N/A. The new code calls no external binary or service.
   - ARCH-CONSTRAINTS: pass. The work is a one-shot read of a ledger with a few hundred rows.
   - ARCH-SECURE: pass. The ledger is parsed into typed rows, and `-` and empty cells parse back to empty.
   - ARCH-ORDER: N/A for M2, which is stateless reads over the ledger.
   - ARCH-FUNERAL: M2 creates nothing durable (the report only prints). Row growth was settled under BR-6.

7. **Plan revisions:** none needed. The issue's Revisions section already records the backfill move, E1/E3 and the `review_recipe` column.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      cutTimes (actualsplit.go:72) now excludes gatestate.BoundaryAll; groupBoundaries no longer filters; TestRowSplitMetaAndSource "a round for every boundary" case fails without the fix (counted 2 rounds before, 1 now).
findings:
  - id: new
    severity: Minor
    family: dash-helper-reuse
    title: |
      convergereport.go dashIfEmpty duplicates estimate.dash (ledger.go:262)
    detail: |
      Trivial duplication; export one helper or accept the two-line copy.
```
