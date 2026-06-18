---
id: 000117
status: working
deps: []
github_issue:
created: 2026-06-17
updated: 2026-06-17
estimate_hours: 3.4
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

## Spec

You can't make the forecast *value* deterministic (irreducible judgment; truth
only known at close). So don't harden the value — harden the **accounting around
it**, at three bite points that map onto machinery sdlc already has. The shell is
**model-agnostic**: it enforces that whatever model the provenance line names is
actually applied — today that's estimate-logic-v2; if #112 ever lands, the same
shell carries it. (Operator chose all three: 1 + 2 + 3.)

**(1) Itemized reconciliation — "no unitemized estimate." [form: hard, fast]**
Replace the bare `estimate_hours` field with a machine-parseable `## Estimate`
block: line items (the **current estimate-logic-v2 primitives** → hours, split
design/impl per v2's two columns), the spec-quality / familiarity factors v2
already defines, and a provenance line naming the model version (`estimate-logic
-v2`). `sdlc change-code` parses it and deterministically enforces:
`estimate_hours == Σ(items)` (within rounding), provenance present + model version
recognized, item types drawn from the closed v2-primitive vocabulary. Per-item
numbers stay judgment, but a free-floating guess is structurally impossible and
the breakdown is diffable + reviewable + scoreable. Mirrors `close`'s
required-evidence flags. Pure parser + checker (ARCH-PURE); a new guard composed
into the existing `change-code` gate, not a new verb (ARCH-DRY).

**(2) Estimate-quality judge. [essence: soft, fast]** A fresh-context judge at
`change-code` (sibling of the plan-quality judge; reuse `internal/judge` +
`architecture.md` harness) reads spec + `## Estimate` breakdown + the model doc
and returns a verdict: was the model actually applied? are the per-primitive
hours plausible for this spec? Catches "itemized but fabricated." Lands as a
verdict trailer; gated behind the existing `--no-judge` escape hatch.

**(3) Auto-calibration at close — close the loop. [feedback: hard, slow]**
`sdlc close` already computes the actual. Have it automatically append a row to a
calibration ledger (home: alongside the model under
`brain/data/life/42shots/velocity/`, machine-readable — supersedes the hand-kept
validation table): `(issue, estimate, est-design, est-impl, actual, ratio,
model-version, supervised|delegated, window-trusted?)` and flag systematic drift
(>2× same-direction miss over the last N trusted rows → warn / require a
`Model-Revision:` note). Makes every estimate falsifiable and produces the v2
dataset.

  **Window-trust flag (the #116 coupling, made safe).** Each row records whether
  its actual came from a `started:`-windowed measurement (#116) or the legacy
  first-commit-parent window. Pre-#116 rows are stamped `window-trusted: no` and
  excluded from drift stats; trusted rows accrue after #116. This is the #68
  posture applied to calibration — a known-truncated actual must not masquerade as
  a clean data point. It lets this issue land *before* #116 without logging
  garbage. **Backfill** the handful of past points we have both sides for (#110,
  #111, charon #13) as `window-trusted: no` historical rows.

## Done when

- `## Estimate` block format defined (closed v2-primitive vocab) + `change-code`
  reconciles `estimate_hours == Σ(items)`, provenance, vocab — deterministically,
  unit-tested, with a precise error message (next-action spec) when it fails.
- Estimate-quality judge wired into `change-code` behind `--no-judge`; fresh
  context; verdict trailer; tested against a faithful and a fabricated breakdown.
- `sdlc close` appends a window-trust-flagged row to the calibration ledger +
  drift-flags on trusted rows; tested. Ledger format documented next to the model.
- Past points (#110/#111/charon#13) backfilled as untrusted-window rows.
- The `change-code` / `issue` / `close` help text (helptext/*.md) documents the
  `## Estimate` contract and the close-the-loop behavior.
- Atlas reconciled (estimate-shell surface + the form/essence/feedback split).

## Scope / non-goals

- Validates the **existing estimate-logic-v2** model; does not author a new model
  (#112 is parked pending this issue's data).
- The `## Estimate` *content* model is whatever the provenance line names; this
  issue is the **shell** (parse, gate, judge, score) around it.
- The `sdlc estimate` arithmetic *engine* (prose→tool, mirroring active-time's
  evolution) is deferred — build it only after the ledger shows the model has
  stabilized. Mechanism (1) provides the structured inputs an engine would compute
  from.
- #116 is **not a start-blocker** (trust-flagging handles its absence); it is the
  upgrade that flips new rows to `window-trusted: yes`. Sequence: ship #117, then
  #116, then trusted data accrues.

## Estimate

The first dogfood of the contract this issue builds. Estimated via the *current*
convention (estimate-logic-v2 build-effort) — deliberately, so #117 becomes
calibration data point #1: a v2 build-effort estimate (3.4h) that the ledger will
later score against the measured operator-attention actual (expected lower —
that's the v2 gap we're instrumenting).

```estimate
model: estimate-logic-v2
familiarity: 1.0
item: greenfield-go-module   design=0.3 impl=0.6
item: smaller-go-module      design=0.2 impl=0.6
item: smaller-go-module      design=0.2 impl=0.5
item: atlas-docs             design=0.0 impl=0.2
item: milestone-review       design=0.0 impl=0.6
design-buffer: 0.30
total: 3.4
```

design 0.7 ×1.30 buffer = 0.91 · impl 2.5 ×1.0 familiarity · total ≈ 3.4

## Plan

- [x] M1 — pure `internal/estimate` package (grammar, parse, check, vocab, ledger-row, drift)
- [x] M2 — change-code enforcement: reconciliation guard + estimate-quality judge + helptext
- [ ] M3 — close-the-loop: ledger append + drift at close, backfill past points, helptext/close + atlas

Detailed plan: `workshop/plans/000117-estimate-shell-reconcile-judge-calibrate-plan.md`

## Log

### 2026-06-17
- 2026-06-17: closed M2 — go build/test/vet ./cmd/sdlc/... green — full suite incl. new estimate-section/judge/changecode/helptext-drift tests. Live `change-code --issue 117 --dry-run`: reconciliation gate passes on the reconciling block, both plan-quality + estimate-quality judges dry-run, flow reaches branching. TestEstimateReconRefusal pins pass/mismatch/no-block/--no-estimate-recon; estimate-quality kept out of AllCategories (TestEstimateQuality_NotInBulkDispatch → no merge-time bulk dispatch); helptext↔vocab drift guard green. ARCH-PURE (estimateReconRefusal pure), ARCH-DRY (reused change-code gate + judge harness + section regex). actual 0.75h is the cumulative issue-window measure (M1 0.43h + M2 increment).; review verdict: FIX-THEN-SHIP
- 2026-06-17: closed M1 — go test ./cmd/sdlc/internal/estimate/... + go vet green; go build ./... OK. Pure parse/check/vocab/ledger/drift table-tested: canonical 3.4 green fixture reconciles; total≠recomputed + estimate_hours≠total + unknown primitive/model all fail with next-action messages; ledger round-trips (10 stable cols); DriftVerdict excludes untrusted rows + flags >2x same-direction. Pure core, zero IO (ARCH-PURE). --no-atlas: M1 is the internal estimate pkg only, no user-facing surface yet; atlas reconciliation lands in M3 (Task 11) when the change-code gate + close ledger make it user-facing.; review verdict: SHIP
Created during the #112 brainstorm once the operator relocated the root cause
from "wrong unit" to "no deterministic shell," then chose to **park #112 and
validate v2 first** (measure before rebuild). Repointed from #112's model onto the
existing estimate-logic-v2; dropped the #112/#116 hard deps (model-agnostic shell;
#116 handled via the window-trust flag). Operator chose shell depth 1+2+3.
Connects to the deterministic-shell / form-vs-essence / minimum-mechanism
principles. Work order this session: #117 → #116.

`sdlc change-code` plan-quality judge: **INFO** (high confidence, plan approved to
start). 4 advisory findings folded into the plan before implementing: (1) Important
— close-time ledger append must degrade gracefully (skip+warn) when no sibling
`brain/` exists, since sdlc is base-layer and propagates downstream; (2) fixed the
Core-concepts example block to reconcile (canonical 3.4 green fixture); (3) keep
`EstimateQuality` out of `AllCategories()`/bulk-dispatch (change-code-time only);
(4) added a vocab↔helptext drift-guard test. Branch `000117-estimate-shell-
reconcile-judge-calibrate` created in-place.

**Decisions (this session):**
- `## Estimate` format = **fenced ```estimate block** (key:value + `item:` lines).
  Deterministic recompute `total = Σdesign×(1+design-buffer) + Σimpl×familiarity`,
  reconciled against `total:` and frontmatter `estimate_hours`. Per-primitive
  `item:` slugs from a closed v2 vocabulary; `model:` from a recognized set. The
  concrete grammar + reconciliation rules live in the durable plan.
- Integration seams (mapped, for ARCH-DRY reuse — anchors in the plan): new pure
  pkg `internal/estimate`; guard slots at `changecode.go` after the estimate gate
  (~:144) and the judge at the `!NoJudge` block (~:147) mirroring
  `runPlanQualityJudge`/`internal/judge`; ledger append in `close.go` near the
  actual (~:519/574); reuse `issue.Parse/GetField/SetField` + the `plan.go`
  `## Section` regex; helptext via `helptext.MustGet`.
