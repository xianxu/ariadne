---
id: 000117
status: open
deps: [ariadne#112, ariadne#116]
github_issue:
created: 2026-06-17
updated: 2026-06-17
estimate_hours:
---

# Deterministic shell for estimate_hours: reconcile, judge, and close the loop against active-time

## Problem

The root cause of the estimate↔actual incoherence is not the *unit* (that's
#112) — it's that **estimation has no deterministic shell.** The ACTUAL side has
one: `internal/activetime` measures from ground truth with the #68 guards so a
missing measurement can't read as "0 hours." The ESTIMATE side has a prose doc
and a single presence check (`change-code` requires `estimate_hours > 0`). The
number has no provenance, no derivation, no consequence — so a faithfully-derived
`0.9` and a made-up `5` are indistinguishable to the system.

Observed live: #110 (`estimate 5 / actual 0.89`) and #111 (`estimate 7 / actual
0.35`) had **no `## Estimate` section and no estimate-logic provenance** — the
numbers were gut guesses; the model was never run. By-the-book v2 would have
landed near the actuals (#110 ~0.4–1.2h, #111 ~1.4–4.4h). So the headline gap is
largely a *compliance* failure the shell must close, on top of #112's unit fix.

The estimate is the one forecast in the system with a **deterministic ground-truth
measurement waiting for it** (active-time-v3). The "estimate-side counterpart to
active-time-v3" is therefore not a parallel measurement but the **feedback
coupling** between the forecast and that measurement. Today the loop is open: the
validation log in `estimate-logic-v2.md` literally says *"none yet."*

## Spec

You can't make the forecast *value* deterministic (irreducible judgment; truth
only known at close). So don't harden the value — harden the **accounting around
it**, at three bite points that map onto machinery sdlc already has. (Operator
chose all three: 1 + 2 + 3.)

**(1) Itemized reconciliation — "no unitemized estimate." [form: hard, fast]**
Replace the bare `estimate_hours` field with a machine-parseable `## Estimate`
block: line items (touchpoint/primitive → expected operator-attention minutes,
per #112's model), a familiarity/spec-quality tier, and a provenance line naming
the model version. `sdlc change-code` parses it and deterministically enforces:
`estimate_hours == Σ(items)` (P50, within rounding), provenance present + model
version recognized, item types drawn from a closed vocabulary (the #112
touchpoint set). Per-item numbers stay judgment, but a free-floating guess is
structurally impossible and the breakdown is diffable + reviewable + scoreable.
Mirrors `close`'s required-evidence flags. Pure parser + checker (ARCH-PURE),
new guard composed into the existing `change-code` gate (ARCH-DRY — reuse the
gate, don't add a verb).

**(2) Estimate-quality judge. [essence: soft, fast]** A fresh-context judge at
`change-code` (sibling of the plan-quality judge; reuse `internal/judge` +
`architecture.md` harness) reads spec + `## Estimate` breakdown + the #112 model
doc and returns a verdict: was the model actually applied? are the touchpoint
counts/modes plausible for this spec? does it ignore obvious
delegation/fragmentation? Catches "itemized but fabricated." Lands as a verdict
trailer like the other judges. Gated behind `--no-judge` (the existing
change-code flag) for the escape hatch.

**(3) Auto-calibration at close — close the loop. [feedback: hard, slow]**
`sdlc close` already computes the actual. Have it automatically append
`(issue, estimate-P50, estimate-range, actual, ratio, model-version,
supervised|delegated)` to a calibration ledger (home: alongside the model under
`brain/data/life/42shots/velocity/`, machine-readable — supersedes the hand-kept
validation table) and flag systematic drift (>2× same-direction miss over the
last N closes → warn / require a `Model-Revision:` note). Makes every estimate
falsifiable and feeds #112's model the calibration data it's starving for.
**Integrity depends on #116** — a truncated actual would log garbage ratios.

## Done when

- `## Estimate` block format defined (closed vocab from #112) + `change-code`
  reconciles `estimate_hours == Σ(items)`, provenance, vocab — deterministically,
  unit-tested, with a precise error message (next-action spec) when it fails.
- Estimate-quality judge wired into `change-code` behind `--no-judge`; fresh
  context; verdict trailer; tested against a faithful and a fabricated breakdown.
- `sdlc close` appends to the calibration ledger + drift-flags; tested. Ledger
  format documented next to the #112 model.
- The `change-code` / `issue` / `close` help text (helptext/*.md) documents the
  `## Estimate` contract and the close-the-loop behavior.
- Atlas reconciled (estimate-shell surface + the form/essence/feedback split).

## Scope / non-goals

- The `## Estimate` *content* model (touchpoints, escalation, fragmentation) is
  #112's; this issue is the **shell** (parse, gate, judge, score) around it.
- The `sdlc estimate` arithmetic *engine* (prose→tool, mirroring active-time's
  evolution) is deferred — build it only after the auto-calibration ledger (3)
  shows the #112 model has stabilized. Mechanism (1) provides the structured
  inputs an engine would later compute from.

## Plan

- [ ]

## Log

### 2026-06-17
Created during the #112 brainstorm once the operator relocated the root cause
from "wrong unit" to "no deterministic shell." Operator chose shell depth 1+2+3
and to keep #112 as the prose model + spin this shell issue separately. Deps:
#112 (the `## Estimate` content vocabulary) + #116 (trustworthy actual for the
loop). Connects to the deterministic-shell / form-vs-essence / minimum-mechanism
principles.
