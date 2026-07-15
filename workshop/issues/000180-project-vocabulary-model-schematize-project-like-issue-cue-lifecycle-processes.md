---
id: 000180
status: working
deps: []
github_issue:
created: 2026-07-15
updated: 2026-07-15
estimate_hours:
started: 2026-07-15T15:12:34-07:00
---

# project vocabulary model: schematize project like issue (cue + lifecycle + processes)

## Problem

`project` is a datatype in prose only (`construct/datatype/project.md`:
frontmatter table, `status: active|paused|done|dropped`, done_when, MVP
scope, single-operator discipline) — it is NOT a vocabulary model.
`construct/vocabulary/` holds issue.cue / pensive.cue / verdict.cue; issue's
status enum + lifecycle are formally modeled and ENFORCED (sdlc reads them
via `pkg/vocab`: set-status transition guards, the compiled done-guard,
discovery dirs feeding resolve). Project has none of that:

- `sdlc close`'s project gate parses project files by convention (task tick
  + detail-block upsert against an unchecked shape);
- no gate validates a project instance's conformance (issues get the
  instance-conformance gate at merge; projects get nothing);
- the status enum exists only as a markdown table nothing reads — a
  hand-maintained restatement, not an enforced source (the exact
  ARCH-PURPOSE gap the issue vocabulary work closed for issues).

This matters more now: #171 lifts project management into the sdlc spine
(project files move into coding repos; the close gate resolves them across
peers; project verbs join sdlc). A lift onto an unschematized noun bakes the
prose-only shape into code.

## Spec

*(Brainstorm converged 2026-07-15, operator + agent. The organizing insight:
the project lifecycle is the issue lifecycle one level up — ideation→
brainstorm, PRD→Spec, project estimation→## Estimate, breakdown→Plan,
execution→implementation, retro→Log, close→measured close. The artifact
answer follows fractally: one file growing sections through gated stages,
binary-owned gates, calibrated estimation, derived views.)*

**Taxonomy** (the three nouns): *issue* = a concrete thing in a single repo;
*pensive* = unstructured musing/realization; *project* = a structured,
TIME-BOUND push for a major change in a product, across dependencies/repos.
A project is not merely a container of issues — it carries a deadline.

**Lifecycle enum (start with this set):**
`ideation → defined → committed → executing → done | dropped` (+ `paused`).
The transitions are the gates:
- **define** (ideation→defined): a PRD exists in the project file. PRD
  authoring is itself tracked as a normal issue; ideation lives in parley
  (linked via `sources:`). Candidate: a fresh-eyes PRD review (plan-quality
  judge's sibling).
- **commit** (defined→committed): the Phase-A estimate exists + the reality
  check passes (fits the roadmap month's capacity, given calibrated velocity
  and competing priorities). Commit SETS THE BASELINE: `deadline:` (the
  time-bound attribute) + planned finish + parallelism intent.
- **breakdown** (committed→executing): PRD converted to issues across repos
  (`deps:` qualified refs). Candidate judge: does the issue set COVER the
  PRD, and does it include the infra/maintenance work PRDs ignore?
- **close** (executing→done|dropped): mandatory PROJECT RETRO entry + the
  fog-factor calibration row (below).

**Time-bound attribute:** frontmatter gains `deadline:` (and the committed
baseline: planned finish, set at commit). Distinguishes project from a mere
issue container; feeds on-track computation.

**Two-phase estimation, one calibration process, different vocabularies:**
- *Phase A (at commit, from the PRD):* a project-level estimate with its OWN
  vocabulary and axes — PRD-level primitives + an explicit uncertainty
  multiplier — because the issue primitive table needs specificity that
  doesn't exist yet. Recorded in the project file's `## Estimate`.
- *Phase B (at breakdown):* the existing per-issue machinery
  (estimate-logic-v3.1), unchanged.
- *Calibration bridge (the fog factor):* at project close, roll up the
  `mvp_scope` issues' measured actuals from the calibration ledger and
  record Phase-A-estimate vs Σ-actuals — a project-level ledger row. Over
  projects this calibrates the PRD-stage multiplier the same way closes
  calibrated issue estimation.

**Kanban: baseline STORED, progression DERIVED, re-forecasts LOGGED** (three
aspects, kept separate — operator decision):
- *Baseline* (stored at commit/breakdown in `## Breakdown`): the
  non-derivable intent — deadline, planned finish, thread assignments,
  sequencing decisions and why. Never overwritten.
- *Progression* (derived, never stored): `sdlc project status` computes the
  board from live issue frontmatter across repos via the resolve machinery —
  dependency frontier (what's unblocked), Σ remaining estimates vs deadline,
  parallel threads as independent dep-subgraphs. A hand-maintained board is
  the drifting-copy lesson waiting to bite.
- *Re-forecasts* (logged): retro entries append "where we are + new
  forecast" to `## Log`; the baseline stays intact so slippage is visible.
- Parallelism ceiling is OPERATOR ATTENTION (~2 concurrent sessions per the
  #117 measurements), not agent count — the constant the estimate unit-note
  said lives "one level up"; this is that level.

**Retro mechanism, not mandate:** `sdlc project retro` verb computes the
on-track summary and prompts the Log entry; the issue-close project gate
(which already touches the project file) nudges when the last retro entry on
an executing project is stale (>1 week); project close REQUIRES a retro
entry (gate, like issue close requires --verified).

**Schema/machinery candidates (as filed, still current):**
- `construct/vocabulary/project.cue` — fields incl. `deadline:`, the
  lifecycle above, transition rules, and a `discovery:` block: home =
  **`workshop/project/`** (operator, 2026-07-15 — project files live under
  workshop/ like every SDLC artifact, one dir per repo; cross-repo
  resolution globs it across peers per #171). Settle at design: singular
  vs plural (sibling dirs are plural — issues/plans/targets), and whether
  done projects archive to workshop/history/ like issues or stay in place
  as records (the datatype prose says "the file becomes a record").
- `pkg/vocab` accessor (`vocab.Project()`) mirroring `Issue()`; no consumer
  hardcodes the enum.
- Project verbs on the spine (new/list/show/set-status/status/retro; tick
  semantics at issue close), transition guards, instance conformance (the
  gate class issues get at merge).
- Prose doc derives, not duplicates: `construct/datatype/project.md` cites
  the cue as schema authority (procedure refers, registry defines), drift
  test binding the two.

**Dogfood:** the first project file under the new model should be this very
effort — the "project-management lift" (this issue + #171's gate/navigation
half + the verbs) PRD'd, estimated, and broken down by its own machinery as
it comes online.

Out of scope (own tickets later): `product` and `roadmap` deserve the same
lift; project first — it is the one the sdlc spine touches.

Related: #171 (residency/navigation/close-gate half — consumes this model;
soft ordering: model first or together).

## Done when

- `construct/vocabulary/project.cue` models the project noun (fields,
  status enum, lifecycle) and `pkg/vocab` exposes it; no consumer hardcodes
  the enum.
- `sdlc close`'s project-file update parses/validates against the model
  (typed records, not substring convention — lessons.md #167).
- A project instance failing conformance is caught by a gate (which gate —
  merge instance-conformance vs close — is a design decision).
- `construct/datatype/project.md` cites the cue as schema authority; a
  drift test binds prose table ↔ model.
- xx-vocabulary skill's claim ("the system's nouns are formally modeled in
  construct/vocabulary/*.cue") becomes true for project.
- The lifecycle funnel (`ideation → defined → committed → executing → done |
  dropped`, + `paused`) and the `deadline:` attribute are in the model, with
  transition gates at define / commit / breakdown / close (close requires a
  retro entry).
- Two-phase estimation is designed: a Phase-A (PRD-level) vocabulary + a
  project-level ledger row at close (Phase-A estimate vs Σ mvp_scope
  actuals — the fog factor).
- Kanban split holds in the tooling: baseline stored, progression derived
  (`sdlc project status` over live cross-repo issue state), re-forecasts
  logged — no hand-maintained board anywhere.

## Plan

- [x] brainstorm: taxonomy, lifecycle funnel, two-phase estimation,
      kanban baseline/derived/logged split, retro mechanism (Spec)
- [ ] design at start-plan: cue shape (esp. cross-repo discovery),
      transition guard mechanics, Phase-A estimate vocabulary, which gate
      owns conformance, verb set, ordering vs #171 — likely structured AS
      the dogfood project (multi-boundary)

## Log

### 2026-07-15

Filed from the #171 thread (operator): "is project a datatype? we should
lift it to be properly schematized just like issue and think about
processes around it." Current state verified: datatype prose exists,
vocabulary model does not; sdlc's project gate parses by convention.

### 2026-07-15 — brainstorm converged (operator + agent)

Operator refinements folded into the Spec: (1) two-phase estimation follows
the SAME calibration process with different vocabularies/axes per phase;
(2) kanban has two distinct aspects kept separate — the committed baseline
("parallel threads → 2 weeks") vs live progression + re-forecast — mapped to
stored-baseline / derived-progression / logged-re-forecasts; (3) lifecycle
enum adopted as proposed, PLUS projects are time-bound, not issue
containers → `deadline:` attribute set at commit; (4) project close runs a
mandatory project retro. Organizing insight recorded: the project lifecycle
is the issue lifecycle one level up, so the artifacts follow fractally (one
file growing gated sections; derived views; calibrated loops). Dogfood: the
project-management lift itself becomes the first project file.

### 2026-07-15 — residency dir: workshop/project/

Operator: project files live in `workshop/project/` (per coding repo) — the
workshop/ family, alongside issues/plans/targets, not the brain-era
`data/project/` path. Folded into the cue discovery candidate; singular-vs-
plural naming + archive-on-done left as design details.
