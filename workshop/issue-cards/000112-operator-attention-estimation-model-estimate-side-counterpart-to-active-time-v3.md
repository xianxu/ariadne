---
id: '000112'
status: punt
created: 2026-06-17
updated: 2026-06-17
---

# Operator-attention estimation model (estimate-side counterpart to active-time-v3)

## Problem

`estimate_hours` has no shared definition, so estimate↔actual calibration is
incoherent. Observed live (#110, #111): the agent set `estimate_hours` as **total
build-effort** ("hours a competent engineer would spend designing + implementing +
testing"), while `sdlc actual` (active-time-v3) measures **operator-attention**
(gap-truncated operator session time). Different units → ~20× estimate/actual gaps
that are noise, not forecasting error. We have a measured ACTUAL model
(active-time-v3, `cmd/sdlc/internal/activetime`); we lack a matching ESTIMATION
model.

The scarce resource is **operator attention/time**, not AI execution — top
providers offer flat subsidized rates, so AI compute cost is immaterial for now
(enterprise/product software ~always; personal software *maybe* at the optimization
tail — deferred until it's a real constraint). So estimation should be anchored to
operator attention. But raw attention-minutes is too crude — two refinements below
make it a real model.
