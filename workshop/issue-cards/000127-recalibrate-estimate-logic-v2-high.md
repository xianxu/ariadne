---
id: '000127'
status: done
started: 2026-06-29T11:15:03-07:00
created: 2026-06-25
updated: 2026-06-29
estimate_hours: 2.55
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
