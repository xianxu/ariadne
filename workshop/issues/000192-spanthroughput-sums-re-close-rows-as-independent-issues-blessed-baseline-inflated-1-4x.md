---
id: 000192
status: working
deps: []
github_issue:
created: 2026-07-29
updated: 2026-07-29
estimate_hours: 1.94
started: 2026-07-29T22:32:28-07:00
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

Simple work (§1): no durable plan doc — these rows are the plan. Single-pass, plain checkboxes,
one `sdlc close`.

**The shared helper's contract** (tightened on plan-quality round 1 — the first draft would have
broken `drift_test.go`):

```go
// NewestPerIssue returns, for each distinct issue in an ALREADY-FILTERED slice, the LAST row
// for that issue — the newest measurement, the ledger being append-only.
func NewestPerIssue(rows []LedgerRow) []LedgerRow
```

- **It takes a PRE-FILTERED slice, and that ordering is load-bearing.** `driftSample` filters
  (trusted, `Actual > 0`, same model) *before* its `seen` check. Deduping first would let an
  untrusted newest row mask an older trusted one and drop that issue from the sample entirely —
  a silent change to drift semantics.
- **A blank `Issue` is keyed POSITIONALLY (`@row:N`), not collapsed.** Rows with no issue id
  cannot be known to be the same issue, and `driftSample` already does this
  (`drift.go:50-53`). This is not defensive: `drift_test.go`'s `trustedRow(est, act)` fixtures
  leave `Issue` empty, so `TestDrift_AllOver`/`AllUnder`/`TrustedZeroActualExcluded` pass three
  *distinct* observations. Keyed plainly on `Issue` they collapse to one, the sample falls below
  `n`, and all three tests fail.
- **`drift_test.go` passing UNMODIFIED is the guard** that the extraction changed no drift
  behavior. A test needing an edit means the semantics moved — a finding, not a fixup.
- Direction-agnostic: `driftSample` walks backward taking first-seen, `SpanThroughput` forward
  taking last-seen. Same rule, opposite traversal, so the helper must not bake in an order.

**The baseline TSV schema does NOT change.** `ParseBaselineTSV` reads 6 columns positionally into
an append-only file; renaming the `rows` column would leave the existing header saying one thing
and new rows meaning another. `ThroughputBaseline.Rows` keeps its name, is fed the deduped count,
and its doc records that pre-#192 rows counted raw ledger lines and are not comparable.

- [ ] Failing test: `SpanThroughput` over a fixture with a 3-times-closed issue (growing
      actuals, the real `ariadne#167` shape) equals the computation over that issue's last row
      alone
- [ ] `NewestPerIssue` in `internal/estimate` per the contract above, with the blank-issue
      positional fallback absorbed from `driftSample`
- [ ] Route `driftSample` through it — **`drift_test.go` must pass unmodified** — so `grep` finds
      one implementation
- [ ] Fix `SpanThroughput` to use it; mutation-verify (removing the dedupe must fail the test)
- [ ] `SpanMeasure`: `Issues` (summed) + `RowsScanned` (seen). **`UntrustedRows` must be counted
      over the DEDUPED set**, or `projectthroughput.go:103` prints "12 of 8 rows" — mixed
      denominators. Update `projectthroughput.go:94,101,103` and print both counts, labelled
- [ ] Known-answer check: the 2026-06-22..07-19 span recomputes to **~80 h/wk**, not 110.60
- [ ] Re-bless the baseline from the corrected measure; record the prior inflation as a `#`
      comment in `throughput-baseline.tsv`
- [ ] Docs — **three** surfaces, not one:
      `atlas/workflow/ledger-landscape.md` (the ledger is per-CLOSE; readers must dedupe),
      `atlas/workflow/sdlc-binary.md:210-219` (documents the sum-over-rows semantics and the
      `rows` provenance field), and
      `brain/data/life/42shots/velocity/SKILL.md:93` — which states "one row per closed issue",
      *the exact wrong fact this defect grew from*, in the doc the recalibration loop reads
      (`calibration-findings.md:30` analyses those rows)
- [ ] `sdlc close --issue 192`

**Known limit, named not fixed:** keeping the newest in-span row means an issue whose closes
straddle `from` contributes hours accrued *before* the span. Acceptable — the alternative is
apportioning a cumulative measurement across span boundaries, which needs per-close deltas the
ledger does not record — but stated beside the append-only non-goal so it is a choice, not an
oversight.

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec            design=0.35 impl=0.10
item: smaller-go-module     design=0.15 impl=0.15
item: smaller-go-module     design=0.10 impl=0.15
item: smaller-go-module     design=0.00 impl=0.10
item: atlas-docs            design=0.05 impl=0.08
item: atlas-docs            design=0.00 impl=0.06
item: milestone-review      design=0.00 impl=0.20
item: milestone-review      design=0.00 impl=0.20
item: milestone-review      design=0.00 impl=0.15
design-buffer: 0.15
total: 1.94
```

**Verified by recomputation:** Σdesign 0.65 × 1.15 = 0.7475, + Σimpl 1.19 = **1.9375** vs stated
1.94 (δ 0.0025, tolerance 0.097). No item above its v3.1 ceiling.

| item | what | why this value |
|---|---|---|
| `issue-spec` | the Spec + plan rows | **design 0.35, under the 0.5 table floor, by a NAMED discount:** the investigation is already complete and it happened **before the claim commit** — quantifying the 31% duplication, tracing both consumers, and finding the 1.41× blessed-baseline inflation were all done pre-filing. `started:` anchors the window at the claim, so that time is *outside* the measured actual and must not be estimated as if it will be spent again. (Precisely the inverse of #190, where I declined the ×0.2 discount *because* the design work fell inside the window.) |
| `smaller-go-module` | `NewestPerIssue` + routing `driftSample` through it | the load-bearing chunk: pre-filtered contract, positional blank-issue key, `drift_test.go` unmodified as the guard |
| `smaller-go-module` | `SpanThroughput` + the `SpanMeasure` field split | includes the `UntrustedRows` denominator fix |
| `smaller-go-module` | `projectthroughput.go` call sites + labelled output | design 0.00 — mechanical; the decision is priced above |
| `atlas-docs` | **three** doc surfaces | at ceiling: `ledger-landscape.md`, `sdlc-binary.md:210-219`, and brain `SKILL.md:93` |
| `atlas-docs` | re-bless + the `#` provenance comment | 0.06 under the 0.08 ceiling — one command and one line |
| `milestone-review` ×3 | close review, fixing its findings, the gate rounds | the pattern #190 established: review and fix cannot share one row, and `started:` puts the gate rounds inside the window |

**Design share is 33.5%**, below the 41–61% peer band — expected and not padded: this is a bugfix
whose analysis was finished before filing, so there is genuinely less design left than in a
feature. Stating it because a share outside the band is normally worth a second look.

**Calibration context.** ariadne on v3.1 measured **geo-mean 1.02× over 41 issues before
2026-07-20**, then four rows at 3.73× median — a regression substantially explained by the
attribution bug #190 just fixed (re-measuring #71 moved it 1.97× → 1.48×). #192 is among the
first issues measured with correct attribution, so its ratio is a **data point about whether that
regression is closed**, not just about this estimate. Recorded so the close reads it that way.

## Log

### 2026-07-29
