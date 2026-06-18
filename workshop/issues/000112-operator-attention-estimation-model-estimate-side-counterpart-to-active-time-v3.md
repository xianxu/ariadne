---
id: 000112
status: punt
deps: []
github_issue:
created: 2026-06-17
updated: 2026-06-17
estimate_hours:
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

## Spec

Develop a **prose** estimation model — the forward-looking counterpart to
active-time-v3 — anchored to operator attention and sharing its unit (so estimate
and actual are comparable). Prose/heuristic for now (a framework agents apply when
setting `estimate_hours`), NOT a binary — mirroring how active-time started as a
procedure before `cmd/sdlc/internal/activetime`. Doc home alongside the actual
method: `brain/data/life/42shots/velocity/` (next to `baseline-v3.md`).

The model needs at least three factors:

1. **Base operator-attention.** Expected focused human time steering / deciding /
   reviewing the issue — the same unit `sdlc actual` measures. NOT total
   build-effort, NOT AI execution time. (This also fixes the `estimate_hours`
   definition: it is *expected operator-attention hours*, not work-size. That
   definition should land in the issue contract — `sdlc issue --help` — and the
   `sdlc claim` estimate-guard prose.)

2. **Fragmentation / context-switching cost.** Attention is NOT fungible across
   time. N scattered re-engagements cost more than the same total minutes in one
   block — each return reloads context. Work that needs the operator only at the
   *start* ≪ work that needs them at *many* points. The model must weight the
   number + spacing of distinct attention touchpoints, not just sum minutes. (This
   is also why a heavily-delegated issue can be cheap on summed-attention yet still
   expensive if it kept yanking the operator back.)

3. **Delegation-induced attention (not free, not full).** AI-delegated work is not
   counted as competent-engineer-hours — but it is NOT discounted to 0 either. The
   longer / deeper a task runs in autonomous mode, the higher the probability it
   escalates back to the operator (a bug, a wrong turn, an ambiguity, a failed
   gate). Model an **expected-escalation term that grows with time-in-autonomous-
   mode** — the tail-risk that AI execution converts into operator attention.
   (Pairs with #2: an escalation IS a context switch, so its cost is amplified by
   the fragmentation factor.)

### Companion (actual-side) — referenced, possibly its own issue

For the estimate to calibrate against the actual, both must be anchored
consistently. active-time-v3's window starts at the first `#N`-subject commit's
parent, so operator-attention-heavy DESIGN time (brainstorm / spec / plan /
reviews) before the first code commit is excluded. Fix: stamp a precise
`started:` (ISO) timestamp at `sdlc claim` and window active-time from it — with
the **claim-at-start-of-engagement** convention (claim → brainstorm → plan → code)
so the window stays tight (bounding cross-issue bleed, cf. #092) yet complete.
Gap-truncation makes an earlier anchor safe from dormant-time inflation. This may
be split into its own actual-side issue.

## Done when

- A prose operator-attention estimation model exists (doc under
  `brain/data/life/42shots/velocity/`), covering the three factors (base
  attention, fragmentation/context-switch, delegation-induced escalation) and
  sharing active-time-v3's unit.
- `estimate_hours` is defined as *expected operator-attention hours* in the issue
  contract (`sdlc issue --help`) + the `sdlc claim` estimate guard, so the agent
  and operator mean the same thing at estimate time.
- The model is legible enough that an agent can apply it to produce an
  `estimate_hours` that is comparable to the `sdlc actual` it will later be scored
  against.

## Scope / non-goals

- **Prose-based for now** — a heuristic/framework, not a binary. A tool can follow
  once the model stabilizes (as active-time did).
- **AI-cost (token/compute) modeling is out of scope** — deferred until it becomes
  a binding constraint (flat subsidized rates make it a non-issue now).
- Pairs with: active-time-v3 / `cmd/sdlc/internal/activetime` (the actual model),
  `baseline-v3.md` (method doc), #092 (segment over-attribution — related window
  concern).

## Plan

- [ ]

## Log

### 2026-06-17 — PARKED (punt) pending calibration data
Brainstormed (see git history of this session). Operator relocated the root cause
from "the model is the wrong unit" to "**the model isn't consistently applied, and
the actual isn't accurately measured**" — so a new model is premature. Decision:
**park #112; build the apparatus first** — #117 (deterministic shell: force *any*
model to be applied + judged + scored) and #116 (accurate `started:`-windowed
actuals). Those produce real estimate↔actual data on the *existing* estimate-logic
-v2 model. **Revisit #112 only if that data shows v2 is structurally inadequate**
(e.g. systematic misses that track supervision-mode / delegation-depth — the
variables v2 is blind to). The brainstormed attention model (touchpoint
accounting; supervision-mode as the headline variable; escalation tail;
fragmentation = reload-tax; DRY-reuse of v2's columns) is preserved in this
session's transcript and summarized above as the candidate if a rebuild is needed.
Carved out: #116 (window fix) + #117 (shell). The original "## Spec" stands as the
candidate model sketch, not an approved design.
