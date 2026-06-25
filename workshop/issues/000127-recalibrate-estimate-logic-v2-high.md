---
id: 000127
status: open
deps: []
github_issue:
created: 2026-06-25
updated: 2026-06-25
estimate_hours:
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

## Plan

- [ ] Diagnose: pull the post-#118 trusted ledger rows; confirm the 2.3× is a unit/scale offset vs. structurally-high primitives. Reconcile what `estimate-logic-v2.1` (parley.nvim#134) already changed.
- [ ] Choose path (a/b/c) and apply to `cmd/sdlc/internal/estimate/vocab.go`; bump `model:` version.
- [ ] Update the grammar doc (`helptext/estimate.md`) + any `estimate-logic-v2.md` narrative to derive from the new vocab; the drift test in estimate package stays the guard.
- [ ] Backtest against the last ~8 ledger issues; show ratios re-center near 1.0.
- [ ] Atlas: update the ledger-landscape page if the model version/derivation changes.

## Log

### 2026-06-25

- Filed from the #126 close, where `DriftVerdict` flagged "last 5 trusted estimates all >2× OVER
  actual — estimates running high; add a Model-Revision note." Rather than back-fit a Model-Revision
  note inline at the #126 close (which would settle for the easy win), this is broken out as its own
  issue: recalibrating the estimate model is a **separable** cross-cutting decision affecting all
  future estimates — genuinely follow-up, not #126's deferred purpose (ARCH-PURPOSE applied: the
  separable extension belongs in its own issue, the principle itself does not). Evidence: the ledger
  table above + `cmd/sdlc/internal/estimate/drift.go`. Related: #117 (calibration ledger), #118 (the
  ship-wall-clock actual redefinition that is the leading suspected cause).
