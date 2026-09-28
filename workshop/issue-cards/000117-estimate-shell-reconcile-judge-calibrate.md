---
id: '000117'
status: done
created: 2026-06-17
updated: 2026-06-17
estimate_hours: 3.4
actual_hours: 0.93
---

# Deterministic shell for estimate_hours: reconcile, judge, and close the loop against active-time

## Problem

The root cause of the estimate↔actual incoherence is not the *unit* (#112,
parked) — it's that **estimation has no deterministic shell.** The ACTUAL side has
one: `internal/activetime` measures from ground truth with the #68 guards so a
missing measurement can't read as "0 hours." The ESTIMATE side has a prose doc
and a single presence check (`change-code` requires `estimate_hours > 0`). The
number has no provenance, no derivation, no consequence — so a faithfully-derived
`0.9` and a made-up `5` are indistinguishable to the system.

Observed live: #110 (`estimate 5 / actual 0.89`) and #111 (`estimate 7 / actual
0.35`) had **no `## Estimate` section and no estimate-logic provenance** — the
numbers were gut guesses; the model was never run. By-the-book v2 would have
landed near the actuals (#110 ~0.4–1.2h, #111 ~1.4–4.4h). So the headline gap is
overwhelmingly a *compliance* failure: the documented model isn't being applied.

**Goal: measure before rebuild.** This shell forces the *existing* estimate-logic
-v2 model to actually be applied (reconcile), checks the application is faithful
(judge), and scores it against an accurate actual (close-the-loop). That produces
the first real estimate↔actual dataset on v2 — which tells us whether v2 is fine
once followed, or genuinely needs the #112 rebuild. We don't speculate; we
instrument.

The estimate is the one forecast in the system with a **deterministic ground-truth
measurement waiting for it** (active-time-v3). The "estimate-side counterpart to
active-time-v3" is therefore not a parallel measurement but the **feedback
coupling** between the forecast and that measurement. Today the loop is open: the
validation log in `estimate-logic-v2.md` literally says *"none yet."*
