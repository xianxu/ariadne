---
id: 000192
status: open
deps: []
github_issue:
created: 2026-07-29
updated: 2026-07-29
estimate_hours:
---

# SpanThroughput sums re-close rows as independent issues — blessed baseline inflated 1.4x

## Problem

`estimate.SpanThroughput` (`cmd/sdlc/internal/estimate/throughput.go:52-66`) sums `r.Actual`
over every ledger row in the date span with no per-issue dedupe. But
`appendCalibrationRow` writes a row **per close invocation**, not per issue — and re-closing an
already-done issue is legal (`--no-reclose-guard` exists for it). So repeat closes appear as
independent observations.

**They are not copies — they are partial sums.** Each re-close measured a LONGER cumulative
window of the same work:

```
ariadne#167   1.43 → 1.74 → 1.90 → 2.10 → 2.25 → 2.44 → 2.71
```

That issue took 2.71h. Summed across its 7 rows it contributes **14.57h — 5.4×**.

**Measured scope (2026-07-29):** 318 data rows for 219 distinct issues — **31% of the ledger is
duplicate observations**. 54 issues carry >1 row; worst offenders `parley.nvim#147` and
`ariadne#167` at 7 rows each, `metis#8` at 6.

**The wrong number is in production.** `SpanThroughput` feeds
`sdlc project throughput --bless`, which wrote the current blessed baseline in
`brain/data/life/42shots/velocity/throughput-baseline.tsv`:

| | h/wk | rows |
|---|---|---|
| blessed 2026-07-19 (span 06-22..07-19) | **110.60** | 280 |
| deduped recount of that span | **80.31** | 191 issues |

So the blessed throughput baseline is inflated **~1.41×** — roughly **33 h/wk of capacity that
does not exist** — and every roadmap or project forecast derived from it is over-optimistic by
that factor.

**Why this shape of bug survives review.** Every individual row is honest: each was a real
measurement at a real close. Nothing in the file is corrupt, so no validation of the data can
find it. The error is entirely in the READ — one consumer treats "rows" as "issues".

**And the two consumers disagree, with only one right.** `DriftVerdict`'s `driftSample`
(`drift.go:37-61`) ALREADY dedupes by issue (`seen[key]`, walking backward so the newest row per
issue wins), so drift detection is correct today. `SpanThroughput` does not. That divergence —
two readers of one file with different notions of an observation — is the actual defect to fix,
not just the arithmetic.

Found while investigating whether the estimation model had regressed (it had not; that was a
separate pooling error on my part). Related: this inflation and the estimate↔actual ratios are
both inputs to "is the model calibrated", so this should land before any estimation-model change.

## Spec

- **Dedupe at READ, in one shared place.** `SpanThroughput` counts the LAST row per issue within
  the span — the newest measurement of that issue's work, matching what `driftSample` already
  does. Extract the rule as one helper both call (`ARCH-DRY`); a second private copy of the
  dedupe is how these two drifted apart in the first place.
- **`SpanMeasure.Rows` must stop meaning "lines read".** It currently reports 280 where there
  were 191 issues, which is what made the inflation invisible in the output. Report issues, or
  report both and label them.
- **Keep the ledger append-only.** Do NOT fix this by rewriting or replacing rows on re-close:
  #117 chose append-only deliberately and `atlas/workflow/ledger-landscape.md` documents it as a
  design principle. The write path is correct; the read is what is wrong.
- **The blessed baseline must be re-blessed after the fix**, since the stored value was computed
  the wrong way. That is an operator action, not something the fix should do silently — but the
  issue is not done until it has happened, because the inflated number is what production reads.
- **A guard that would have caught this.** A test asserting `SpanThroughput` over a fixture
  containing a re-closed issue equals the same computation over that issue's last row alone.
  Mutation-verified: removing the dedupe must fail it.

## Done when

- `SpanThroughput` over a span containing an N-times-closed issue counts that issue's hours
  ONCE, at its newest measurement — pinned by a test with a re-close fixture, and
  mutation-verified (deleting the dedupe fails it).
- The dedupe rule exists in exactly ONE place, consumed by both `SpanThroughput` and
  `driftSample`; `grep` shows no second implementation.
- `SpanMeasure` no longer reports a row count that reads as an issue count.
- Recomputing the 2026-06-22..07-19 span yields **~80 h/wk**, not 110.60 — the concrete
  known-answer check, since both figures are already computed in the Problem section.
- The blessed baseline in `brain/.../throughput-baseline.tsv` has been re-blessed from the
  corrected measure, with a row recording that the prior one was inflated.
- `atlas/workflow/ledger-landscape.md` states that the ledger is per-CLOSE not per-issue, and
  that every reader must dedupe — the fact that would have prevented this.

## Plan

- [ ] Failing test: `SpanThroughput` over a fixture with a 3-times-closed issue (growing
      actuals) equals the computation over its last row alone
- [ ] Extract the per-issue dedupe as one shared helper; route `driftSample` through it so there
      is a single implementation
- [ ] Fix `SpanThroughput` to use it; mutation-verify by removing the dedupe
- [ ] Make `SpanMeasure.Rows` honest (issues, or both counts labeled)
- [ ] Known-answer check: the 06-22..07-19 span recomputes to ~80 h/wk, not 110.60
- [ ] Re-bless the baseline from the corrected measure; record the prior inflation in the file
- [ ] `atlas/workflow/ledger-landscape.md`: the ledger is per-close; readers must dedupe
- [ ] `sdlc close --issue 192`

## Log

### 2026-07-29
