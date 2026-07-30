---
id: 000192
status: working
deps: []
github_issue:
created: 2026-07-29
updated: 2026-07-29
estimate_hours: 2.33
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

- [x] Failing test: `SpanThroughput` over a fixture with a 3-times-closed issue (growing
      actuals, the real `ariadne#167` shape) equals the computation over that issue's last row
      alone
- [x] `NewestPerIssue` in `internal/estimate` per the contract above, with the blank-issue
      positional fallback absorbed from `driftSample`
- [x] Route `driftSample` through it — **`drift_test.go` must pass unmodified** — so `grep` finds
      one implementation
- [x] Fix `SpanThroughput` to use it; mutation-verify (removing the dedupe must fail the test)
- [x] `SpanMeasure`: `Issues` (summed) + `RowsScanned` (seen). **`UntrustedRows` must be counted
      over the DEDUPED set**, or `projectthroughput.go:103` prints "12 of 8 rows" — mixed
      denominators. Update `projectthroughput.go:94,101,103` and print both counts, labelled
- [x] Known-answer check: the 2026-06-22..07-19 span recomputes to **~80 h/wk**, not 110.60
- [x] Re-bless the baseline from the corrected measure; record the prior inflation as a `#`
      comment in `throughput-baseline.tsv`
- [x] Docs — **three** surfaces, not one:
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
`baseline-v3.1.md`. Method A only. Revised once on the estimate-quality review (1.94 → 2.33) —
see the deltas below.*

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec                design=0.30 impl=0.10
item: smaller-go-module         design=0.15 impl=0.15
item: smaller-go-module         design=0.10 impl=0.15
item: cross-cutting-refactor    design=0.00 impl=0.08
item: smaller-go-module         design=0.00 impl=0.10
item: atlas-docs                design=0.05 impl=0.08
item: atlas-docs                design=0.00 impl=0.06
item: cross-repo-refactor-small design=0.10 impl=0.10
item: milestone-review          design=0.15 impl=0.00
item: milestone-review          design=0.00 impl=0.20
item: milestone-review          design=0.00 impl=0.20
design-buffer: 0.30
total: 2.33
```

**Verified by recomputation:** Σdesign 0.85 × 1.30 = 1.1050, + Σimpl 1.22 = **2.3250** vs stated
2.33 (δ 0.0050, tolerance 0.117). No item above its v3.1 ceiling.

**What the estimate-quality review changed, and why each was right:**

- **`design-buffer` 0.15 → 0.30.** v3.1 step 4 gives +15% only "when the issue has a thorough
  plan doc"; this issue's Plan opens by stating it has **none** (§1 simple work). I had copied
  #190's buffer without the condition that licensed it there — #190 does have a durable plan doc.
  This alone is 1.94 → 2.04.
- **Docs are three instances, not one at ceiling.** `design=0.05` is the atlas-docs **floor**, not
  its ceiling, and the v2 calibration anchor priced charon's "docs ×3" as three instances. Split
  into two `atlas-docs` rows plus a `cross-repo-refactor-small` for the two **brain** writes —
  which are a genuine second-repo coordination cost, since the close gate never auto-commits into
  a brain, so those land as a manual commit there.
- **The known-answer check had no item at all.** "the span recomputes to ~80 h/wk, not 110.60" is
  the issue's headline verification and means building the binary, running it against a live
  318-row ledger in a peer repo, and reconciling to a hand-computed figure. Now its own row.
- **`issue-spec design` 0.35 → 0.30, onto the model's lattice.** Step 3 offers ×0.2 / ×0.5 / ×1.0;
  0.35 corresponded to ~×0.7 and was derived from nothing. 0.30 is the ×0.2 band top — and the
  review correctly noted my stated reason (the investigation predates the claim commit, so it is
  outside the measured window) argues for the ×0.2 band, not above it.
- **Gate rounds moved from `impl` to `design`.** v3.1 keeps design hours unscaled precisely
  because "the design conversation does not compress like implementation." Booking gate rounds as
  `impl` pre-scaled them by 0.40, which understates them.
- **`projectthroughput.go` re-slugged** from `smaller-go-module` to `cross-cutting-refactor`:
  three call sites is not a Go module.

**A disagreement recorded rather than split.** The review's strongest non-blocking point is that
`milestone-review` ×3 — now 26% of the total — is where this estimate will most likely over-fire,
citing #190, which booked 0.60h of review overhead against a **0.86h measured actual for the whole
issue**. I have kept the three rows, because each names work that will actually happen (a review
dispatch, fixing its findings, and the gate rounds already spent). If it over-fires, that is
evidence about the **`milestone-review` primitive's hours** — a table-calibration finding — not a
reason to shave the estimate now. Shaving it to hedge against a predicted miss is the same
dishonesty as #187's warning about re-deriving to make a ratio look better, just in the other
direction. The ledger adjudicates.

**Correction to an earlier claim in this issue.** An earlier draft argued the post-2026-07-20
regression was "substantially explained" by the #190 attribution bug, citing #71 re-measuring
1.97× → 1.48×. **That 1.48 is not in the ledger** — it was an ad-hoc `sdlc actual` run, and the
newest recorded `ariadne#71` row still reads 1.97×. Meanwhile all three ariadne rows dated
2026-07-29 sit at 3.6–4.4× over, **including #190 itself, which was measured with the fix in
place**. So the honest statement is: the bug was *one* proven contributor (#71's own re-measurement
moved inside 1.5×), and whether it explains the rest is **open**. #192 is among the first issues
measured with correct attribution throughout, so its ratio is a data point on that question — which
is why the number above must not be tuned toward a hoped-for answer.

## Log

### 2026-07-29
