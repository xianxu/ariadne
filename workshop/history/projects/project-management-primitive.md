---
type: project
name: project-management-primitive
goal: Lift project management into the sdlc spine with the same rigor issues got — formal model, binary-owned gates, calibrated estimation, derived views.
done_when: This file itself reaches `done` via `sdlc project close` — retro entry + fog-factor ledger row recorded — having been driven through every funnel transition (define → commit → breakdown → close) by the machinery it tracked the building of.
status: done
deadline: 2026-07-19
mvp_scope: [ariadne#180, ariadne#171, ariadne#182]
explicitly_out: [ariadne#15]
created: 2026-07-16
updated: 2026-07-19
sources: [workshop/issues/000180-project-vocabulary-model-schematize-project-like-issue-cue-lifecycle-processes.md, workshop/issues/000171-the-tension-between-brain-and-other-repos.md, workshop/issues/000182-project-calendar-estimator.md, workshop/plans/000180-project-vocabulary-model-plan.md]
planned_finish: 2026-07-19
---

# project-management-primitive

The guinea pig: the project-management lift, managed as the first project file
under its own emerging machinery. Hand-authored to the #180 model shape at
ideation (the schema does not exist yet — this file is the live test subject
for the conformance gate, the verbs, and every funnel transition as they come
online). Headline omission: the `product`/`roadmap` noun lifts (ariadne#15 —
deliberately out; project first, it is the noun the sdlc spine touches). The
effort→calendar estimator (ariadne#182) is IN the MVP — the timeline aspect
is the key differentiator between a project and an issue (managing something
higher-level and longer-running), so the lift isn't delivered while the
deadline math is still attestation.

## PRD

**Goal.** Lift project management into the `sdlc` spine with the same rigor
issues already have: a formal model, binary-owned lifecycle gates, calibrated
estimation, and issue-derived views. Before this work a "project" was a
hand-maintained markdown file — no schema, no guarded transitions, timeline as
free-text attestation — while an issue had `construct/vocabulary/issue.cue`,
`sdlc` gates at every boundary, a Phase-A/actuals calibration loop, and derived
state. The organizing insight (#180 Spec): **the project lifecycle is the issue
lifecycle one level up** — a time-bound, cross-repo push — and it should be held
to the same standard.

**Requirements.**

1. **Formal model (R1).** A `project` vocabulary in
   `construct/vocabulary/project.cue` as the single source for the status set,
   the funnel (ideation → defined → committed → executing → done | dropped, plus
   paused), the named guards per edge, the data shape (`#Project`), and the
   discovery home. sdlc reads the exported model via `pkg/vocab`; the prose
   companion (`construct/datatype/project.md`) cites the cue as schema
   authority, bound by a drift test. No hardcoded enums. [ariadne#180]

2. **Binary-owned gates (R2).** Every funnel transition runs through
   `sdlc project set-status` (never hand-edited frontmatter), each edge enforcing
   its model-named guards: `prd-present` (define); `phase-a-estimate` +
   `baseline-set` + `reality-check` (commit); `issues-cover-prd` (breakdown).
   Terminal completion is a dedicated `sdlc project close` owning the
   `retro-recorded` + `fog-factor-recorded` gates and the archive. [ariadne#180]

3. **Calibrated estimation + the effort→calendar bridge (R3).** A PRD-stage
   Phase-A estimate (`base × fog`) recorded at commit, and a project-close loop
   that rolls the complete MVP issue actuals into a fog-factor ledger — the
   calendar analogue of the issue estimate loop. **The timeline aspect is the
   differentiator between a project and an issue** (managing something
   higher-level and longer-running), so the effort→calendar bridge is IN scope,
   not deferred: `sdlc project throughput` measures focused h/wk from an
   operator-blessed span, and a pure forecast core projects a finish date
   (throughput ÷ contention) surfaced at commit / show / status / close. It
   **informs, never blocks** — estimation is a funny business; always track and
   inform, never gate. [ariadne#182]

4. **Cross-repo residency (R4).** Projects live in their center-of-gravity
   coding repo under `workshop/projects/`, never in brain (brain is
   capture/measurement only — the spine refuses there). Every reference is a
   qualified `repo#id`, so a record reads identically from any vantage and
   tooling discovers it fleet-wide: close ticks the referencing project wherever
   it lives; `sdlc project find` / `resolve --kind project` navigate to it. Four
   legacy records migrated out of brain. [ariadne#171]

5. **Derived views (R5).** `sdlc project status` / `show` render an
   issue-derived progress board and the live forecast-vs-deadline drift, so
   project state is computed from the referenced issues rather than
   hand-maintained. [ariadne#171, ariadne#182]

**Acceptance boundary.**

- **In MVP:** `ariadne#180` (vocabulary model + gates + verbs), `ariadne#171`
  (residency + close-gate lift), `ariadne#182` (calendar estimator). This file
  itself is the guinea pig — authored at ideation to the #180 shape, then driven
  through every funnel transition by the machinery it tracked the building of.
  `done_when` is met when it reaches `done` via `sdlc project close`.
- **Explicitly out:** `ariadne#15` (the `product` / `roadmap` noun lifts).
  Project is the noun the sdlc spine touches first; the durable-charter and
  month-aggregate nouns are a separate, later lift (follow-up `ariadne#185`
  lifts roadmap out of brain).

## Estimate

Phase-A (PRD-stage), per
`brain/data/life/42shots/velocity/estimate-logic-project-v1.md`:

**phase-a:** 72h
**fog:** 1.5
**basis:** 3 workstreams sized at PRD time, before breakdown — formal model +
gates + verbs [ariadne#180] (L / 20h); residency + close-gate lift + fleet
discovery + migration [ariadne#171] (L / 20h); calendar estimator [ariadne#182]
(M / 8h). base = 48h; fog = 1.5 (the seed default while the Fog ledger holds <3
project rows); phase-a = 48 × 1.5 = 72h.

Phase-B was the per-issue estimation done when each workstream became an issue
(issue `estimate_hours` — #180 8.1h, #171 4.1h, #182 2.65h; Σ 14.85h). Measured
actuals — #180 11.93h, #171 10.02h, #182 3.21h; **Σ 25.16h**. `sdlc project
close` records the first Fog-ledger row from these (observed fog ≈ 25.16 / 72 ≈
0.35) — the calibration signal that the Phase-A seed midpoints overestimate
ariadne-scale focused work, the exact datum the ledger exists to accumulate.

## Breakdown

- [x] project vocabulary model: cue + lifecycle + gates + verbs [ariadne#180]
- [x] brain-vs-repos residency + close-gate lift (consumes the model) [ariadne#171]
- [x] calendar estimator: effort→calendar bridge, computed reality-check [ariadne#182]

## Log

### 2026-07-18

Ticked [ariadne#180] retroactively (done 2026-07-16, actual 11.93h, PR #100
— archived in ariadne/workshop/history/issues/). Its close predated #171 M2's
fleet-wide discovery, so the then-current `--brain-dir` lookup never found
this local project file to tick the row. [ariadne#171] closed 2026-07-17
(actual 10.02h, PR #101) and was ticked by the new close gate itself.

### 2026-07-16

Created at ideation (operator: "use the creation of project management in
ariadne as a project to guinea pig the project management improvement
itself"). This reverses the 2026-07-15 dogfood-deferral decision — the file
now exists BEFORE the model, deliberately: each #180 milestone should
exercise its machinery against this instance (M2's conformance gate, M3's
set-status/guards, M4's board/retro/close). Task rows seeded in
`## Breakdown` ahead of the formal breakdown transition — bootstrapping
license: the issues predate the machinery; the committed baseline (deadline,
planned finish, threads) still lands at the commit transition, not before.
Note the arc is self-referential on purpose: `define` will gate on a PRD this
file doesn't have yet, and Phase-A estimation (`## Estimate`) can only be
filled once #180 M4.4 defines the method.

**Scope event 2026-07-16**: #182 moved from explicitly_out INTO mvp_scope
(operator): "that is the key differentiator between a project and an issue,
the timeline aspect, of managing something higher level, longer running."
A project lift whose deadline feasibility is still free-text attestation
hasn't delivered the differentiator. explicitly_out now carries ariadne#15
(the product/roadmap noun lifts).

### 2026-07-19 — transition evidence

- reality-check: All three MVP issues (#180/#171/#182) shipped and merged before commit — board remaining is 0, so the #182 forecast reports no remaining hours to project (informational, per informs-never-blocks). planned_finish=2026-07-19: the only work left is this tracking-project funnel bookkeeping, committed and closing same-day.

### 2026-07-19 — planned_finish

- planned_finish set manually: 2026-07-19

### 2026-07-19 — transition evidence

- issues-cover-prd: MVP scope covers all five PRD requirements: R1 formal model + R2 gates/verbs → #180; R3 calibrated estimation + effort→calendar bridge → #182; R4 cross-repo residency → #171; R5 derived views → #171 (board) + #182 (forecast line). All three issues done+merged; #15 (product/roadmap noun) explicitly out. No PRD requirement is uncovered.

### 2026-07-19 — retro

**board:** 3/3 done · Σ remaining ≈ 0h · deadline 2026-07-19 (0 days) · frontier: -

The dogfood arc closed. All three MVP issues shipped and merged — #180
(model + gates + verbs), #171 (residency + close-gate lift), #182 (calendar
estimator) — and this file was driven through its own funnel by exactly that
machinery: define gated on the PRD, commit on the Phase-A estimate + baseline,
breakdown on issue-coverage, each guard the work itself introduced. The
self-referential loop held at commit: #182's forecast ran against this very
project and correctly reported zero remaining hours (all issues done),
informing without blocking — the informs-never-blocks posture working as
designed rather than fabricating a date. Calibration read: Phase-A sized the
three workstreams at 72h (L/L/M, base 48 × fog 1.5); measured actuals totalled
25.16h, so the first Fog-ledger row lands at ≈0.35 — the seed midpoints
overestimate ariadne-scale focused work, precisely the datum the ledger exists
to accumulate toward a calibrated multiplier (fog stays the 1.5 default until
≥3 rows). Follow-ups filed: #185 (lift roadmap out of brain) and #186 (shared
git-fixture test package). Next: `sdlc project close` records the fog +
calendar ledger rows and archives to `workshop/history/projects/`.

### 2026-07-19 — close

- phase-a: 72h
- actuals: 25.16h
- fog: 0.35
