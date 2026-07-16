---
id: 000127
status: done
deps: []
github_issue:
created: 2026-06-25
updated: 2026-06-29
estimate_hours: 2.55
started: 2026-06-29T11:15:03-07:00
actual_hours: 0.77
---

# recalibrate estimate-logic-v2 — last 5 trusted closes all ran >2x OVER (model running high)

## Problem

`sdlc close`'s calibration drift guard (`cmd/sdlc/internal/estimate/drift.go`,
`DriftVerdict`) fired at the #126 close: **the last 5 trusted-window closes all came in
>2× OVER estimate.** From the calibration ledger
(`brain/data/life/42shots/velocity/calibration-ledger.tsv`):

| issue | est | actual | ratio |
|---|---|---|---|
| pair#66 | 30.00 | 12.87 | 2.33× |
| pair#67 | 1.80 | 0.72 | 2.50× |
| pair#67 | 1.80 | 0.81 | 2.22× |
| ariadne#122 | 10.00 | 4.67 | 2.14× |
| ariadne#126 | 1.36 | 0.54 | 2.52× |

This is a **systematic** miss in one direction (~2.3× high), not noise — exactly the signal
the #117 calibration ledger was built to surface. estimate-logic-v2 produces estimates that
are consistently ~2× the measured ship wall-clock.

Likely root cause (to confirm, per [[measure-before-rebuild]]): #118 redefined `sdlc actual`
as **idle-removed ship wall-clock with subagent-execution spans kept** — a much smaller
number than the build-effort the v2 primitive table was authored against. So the model isn't
necessarily "wrong" about effort; its **unit may have drifted out of sync with the actual it's
calibrated against.** The fix could be a uniform scale-down of the impl primitives, OR a
re-derivation of the primitive table against the post-#118 actual definition — that's the
investigation this issue owns. (Note: `parley.nvim#134` already stamped `estimate-logic-v2.1`,
so a version bump is in flight somewhere — reconcile with it; don't fork a parallel model.)

## Spec

Recalibrate the estimate model so trusted-window ratios center on ~1.0 again. Decide between:
- **(a) uniform impl scale-down** — multiply the impl hours of the primitive table by ~1/2.3,
  leaving the structure intact (cheapest; treats the miss as a unit-scale offset);
- **(b) re-derive the primitive table** against a sample of post-#118 closed issues (most
  faithful; more work);
- **(c) reconcile/adopt `estimate-logic-v2.1`** if it already encodes part of this fix.

Whichever path: the canonical source is `cmd/sdlc/internal/estimate/vocab.go` (the grammar doc
`cmd/sdlc/helptext/estimate.md` + any `estimate-logic-v2.md` narrative mirror it — single source
⇒ consumers derive, **ARCH-PURPOSE**: don't leave a hand-maintained restatement of the table out
of sync). Bump the `model:` version string so old ledger rows stay attributable to the model that
produced them (the ledger already carries the model column for exactly this).

## Done when

- The systematic 2× bias is diagnosed (confirm it's the #118 unit-redefinition vs. genuinely
  high primitives) and the chosen recalibration (a/b/c) is applied to the single source.
- The `model:` version is bumped and the grammar doc + any narrative mirror derive from the
  canonical vocab (no stale restatement — ARCH-PURPOSE shadow-sweep).
- A backtest: re-running the new model against the last ~8 ledger issues' scopes yields ratios
  centered near 1.0 (not a fresh systematic miss in either direction).
- The drift guard no longer fires on a freshly-closed issue estimated with the new model.

## Estimate

P50 2.55h. Revised after creating `estimate-logic-v3.1`: this is a narrow `sdlc` calibration/code-path change plus analytical brain-doc authoring and a close-boundary review.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec          design=0.4 impl=0.1
item: pensive             design=0.8 impl=0.1
item: smaller-go-module   design=0.3 impl=0.2
item: atlas-docs          design=0.1 impl=0.1
item: milestone-review    design=0.0 impl=0.2
design-buffer: 0.15
total: 2.55
```

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Plan

- [x] Diagnose: pull the post-#118 trusted ledger rows; separate clean calibration rows from actual-side artifacts. Reconcile `estimate-logic-v2.1` and `estimate-logic-v3`.
- [x] Apply the numerical recalibration in the brain velocity docs as `estimate-logic-v3.1` / `baseline-v3.1`: keep design hours, keep v2.1/v3's 15% thorough-plan buffer, and scale AI-paired impl hours to 40% for `sdlc actual` ship-wall-clock.
- [x] Apply the shared-method changes in `cmd/sdlc/internal/estimate/`: recognize `estimate-logic-v3.1`, make it the current model, and keep drift windows per-model + unique issue.
- [x] Update the grammar/source docs and any atlas surfaces that currently nudge agents toward stale `estimate-logic-v2`.
- [x] Backtest against the recent ledger sample; show clean rows re-center near 1.0 and artifact rows stay called out as measurement problems.
- [x] Atlas: update the ledger-landscape page if the model version/derivation changes.

## Revisions

### 2026-06-29 — plan-quality rescope to real recalibration

Reason: `sdlc change-code` plan-quality found that a provenance-only v3 adoption would not change any estimate numbers and would only reset the drift guard.

Delta: chose path (a), a uniform implementation-hours scale-down, realized as `estimate-logic-v3.1` / `baseline-v3.1` in the brain velocity docs. The ariadne code scope now recognizes v3.1, makes it the current model, and makes drift windows per-model and unique per issue. The estimate block was restamped from v3 to v3.1 after the new calibration source existed.

## Log

### 2026-06-29
- 2026-06-29: closed — go test ./cmd/sdlc/... -count=1 passed; git diff --check passed; go run ./cmd/sdlc estimate-source --brain-dir ../brain reports estimate-logic-v3.1 [ok]; brain commit 5df9efc adds estimate-logic-v3.1/baseline-v3.1 docs; review verdict: FIX-THEN-SHIP

- `sdlc start-plan --issue 127` delivered the ARCH-* plan lens. `sdlc change-code --issue 127`
  first returned `VERDICT: FAILURE` because the initial plan only adopted v3 provenance and drift
  hygiene; v3 is a pure shim over v2.1 and does not numerically recalibrate. Revised the plan to
  include brain-side calibration docs, then reran `sdlc change-code`; plan-quality returned
  `VERDICT: INFO`, estimate-quality returned `VERDICT: INFO`, and the branch
  `000127-recalibrate-estimate-logic-v2-high` was created.
- Diagnosis/backtest: latest unique trusted v2 rows with actual >= 0.5h show the clean model-bias
  rows moving from roughly 2.0-2.5x old ratios to roughly 0.6-1.35x under
  `design*1.15 + impl*0.40`. Residual 3-6x rows remain classified as actual-quality/scope-compression
  outliers rather than evidence for further global shrinking. This is an in-sample fit, not held-out
  validation; the next v3.1 closes are the real validation set.
- Created the brain calibration pair `estimate-logic-v3.1.md` / `baseline-v3.1.md` and updated the
  brain velocity skill to make v3.1 current. Existing unrelated brain WIP was present before this
  issue and was left untouched.

### 2026-06-25

- Filed from the #126 close, where `DriftVerdict` flagged "last 5 trusted estimates all >2× OVER
  actual — estimates running high; add a Model-Revision note." Rather than back-fit a Model-Revision
  note inline at the #126 close (which would settle for the easy win), this is broken out as its own
  issue: recalibrating the estimate model is a **separable** cross-cutting decision affecting all
  future estimates — genuinely follow-up, not #126's deferred purpose (ARCH-PURPOSE applied: the
  separable extension belongs in its own issue, the principle itself does not). Evidence: the ledger
  table above + `cmd/sdlc/internal/estimate/drift.go`. Related: #117 (calibration ledger), #118 (the
  ship-wall-clock actual redefinition that is the leading suspected cause).
- **Data-quality caveat surfaced at the #128 close (2026-06-25):** ariadne#128 closed with est 1.66 /
  actual **0.13** → ratio **12.8×**, and the ledger marked it **trusted-window**. That 0.13h is a
  *degenerate-window artifact*, NOT real over-estimation: active-time-v3 infers active spans from
  commit-to-commit author-date gaps, and #128 landed design+impl in a *single* commit, so almost all
  the real work-time is uncaptured. The diagnosis here MUST separate two distinct failures before
  rescaling primitives: (1) the genuine ~2.3× model-high bias (the unit drift from #118), vs.
  (2) **measurement** artifacts where a low-commit-count issue understates actual and inflates the
  ratio. Rescaling the estimate model to "fix" inflated ratios that are really actual-side
  measurement noise would mis-calibrate. Action for the backtest: filter or down-weight degenerate
  windows (e.g. single-commit, or actual below an active-time floor), and consider whether
  `WindowTrusted` should exclude single-commit windows at all (an active-time-v3 engine fix, possibly
  its own issue). The 12.8× row is a known outlier — don't treat it as a 12.8× model miss.
