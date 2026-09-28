---
id: '000192'
status: done
started: 2026-07-29T22:32:28-07:00
created: 2026-07-29
updated: 2026-07-31
estimate_hours: 2.33
actual_hours: 0.55
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
