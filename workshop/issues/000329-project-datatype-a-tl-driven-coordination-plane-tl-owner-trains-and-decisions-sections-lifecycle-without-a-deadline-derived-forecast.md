---
id: 000329
status: working
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-12
estimate_hours: 7.71
card_mirror: '618c7f6b4d9d0e2d3c36674983df4b898a347446' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T16:34:43-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# Project datatype: a TL-driven coordination plane (tl: owner, Trains and Decisions sections, lifecycle without a deadline, derived forecast)

## Problem

The `project` datatype was built as a shipping plan (scope, estimate, deadline, a status tracking the schedule). That
mode was never exercised. Run by a TL (ops' ariadne-robustness-1, 10-10), a project is instead a higher plane for
cross-issue problems, sequencing and driving execution, and everything the TL used lived outside the lifecycle.
The operator approved pivoting the datatype (no compatibility to keep). Design note, approved at the batch-2
recap: `ops/workshop/pensive/2026-10-10-01-pensive-project-datatype-tl-centric.md`.

## Spec

Per the design note:
- **Layers:**
  - Charter: problem, cross-issue requirements, `done_when`, scope.
  - Plan: trains, cross-train edges (⇒ hard, ~ overlap), allocation per train.
  - Decisions: cross-issue only, dated, who decided, one line of why; append-only at the end of the file.
  - Live state stays in the TL's status file, outside the datatype.
  - Forecast: derived, never typed.
- **Decision placement:**
  - issue-scoped decisions go in the issue's `## Revisions`;
  - cross-issue decisions go in the project;
  - process rules go in the TL skill.
  - The Breakdown's per-issue scope-event blocks shrink to pointers.
- **Ownership:** `tl:` frontmatter (one TL or a list, each owning trains). The close gate side is ariadne#326.
- **Lifecycle:** `defined → executing → closing → closed`, with no deadline and no planned-finish fields. A date,
  when needed, is set against the forecast and never drives status.
- **Tooling:**
  - `sdlc project find` and `resolve` report an issue's train and car;
  - a read-only `sdlc project forecast` (remaining estimates ÷ calibrated build rate × (1 + converge k from
    ariadne#234), across N parallel trains, respecting ⇒ edges) gives a range.

Open (from the note): one file or several as a project grows; where cross-operator allocation is negotiated; whether
`mvp_scope` becomes "cars up to a marked point".

## Done when

- [ ] `construct/datatype/project.md` and the vocabulary describe the new shape and lifecycle; the deadline fields are gone.
- [ ] `sdlc project find` / `resolve` name the train and car (test).
- [ ] `sdlc project forecast` prints a range for a fixture project (test).
- [ ] ops' ariadne-robustness-1 conforms to the new shape.

## Plan

Durable plan: `workshop/plans/000329-project-datatype-tl-plan.md` (decisions D1–D10).

- [ ] M1 — Vocabulary and lifecycle: cue statuses/guards/field ban, set-status and close rework, trains parser, `sdlc project forecast` range (replacing the calendar forecast and fog ledger), datatype prose, help and atlas
- [ ] M2 — find/resolve name the train and car; ops conformance patch for the TL

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec                 design=1.0 impl=0.1
item: typed-data-prototype       design=0.8 impl=0.15
item: cross-cutting-refactor     design=0.5 impl=0.2
item: cross-cutting-refactor     design=0.5 impl=0.2
item: greenfield-go-module       design=0.5 impl=0.3
item: greenfield-go-module       design=0.8 impl=0.3
item: smaller-go-module          design=0.2 impl=0.15
item: smaller-go-module          design=0.2 impl=0.15
item: atlas-docs                 design=0.15 impl=0.08
item: milestone-review           design=0.0 impl=0.2
item: milestone-review           design=0.0 impl=0.2
item: cross-repo-refactor-small  design=0.2 impl=0.1
design-buffer: 0.15
total: 7.71
```

Items, in order:
1. Spec and plan (done before the estimate).
2. The vocabulary and datatype prose.
3. Two refactors: the set-status/close rework, and the deletion sweep of the calendar forecast, fog/calendar ledgers and throughput across about 20 files and tests.
4. Two greenfield modules: the trains parser, and the scheduler/forecast.
5. Two small modules: the guards, and find/resolve train/car.
6. Docs and atlas.
7. Two milestone reviews.
8. The ops conformance patch.

## Log

### 2026-10-10

### 2026-10-12
- Claimed at ariadne:3. The hub confirmed the lifecycle (defined → executing → closing → closed, plus dropped) and that the ops item ships as a patch the TL applies (workers don't edit ops project files).
- Fleet migration list for the TL: `ideation` in pair couch-slots-v2, parley.nvim parley-packaging and parley-v1-release; deadline fields in metis arena3-rogii-wellbore and ml-research-workflow; no `paused` anywhere.
- M1 built (e873b94b). The trains parser and forecast moved into M1; see the plan's Revisions. Targeted runs are green: pkg/vocab, internal/project (with fuzz seeds and DAG invariants), internal/estimate, processmanual, cmd/vocabulary (the end-to-end deadline ban), cue vet_test.sh, and the cmd/sdlc project/forecast/retro/close/gate-catalog tests.
- `atlas/process-manual.md` is generated and refreshed periodically. It still describes the old project help and is left for the next refresh.
- M1 review round 1, FIX-THEN-SHIP, every finding fixed:
  - BR-1: the cycle error names a car on the cycle, with a test where the downstream car comes first.
  - BR-2: the archive-scope test.
  - Minors:
    - close's calibration-row path uses `calibrationLedgerPath` (one rule);
    - `projectIssueTerminalFn` resolves the checkout once, through the shared `projectCheckout`;
    - `CalibrateForecast` excludes negative estimate and converge values;
    - `runProjectForecast` takes an injected stderr;
    - the D6 drift is recorded in the plan's Revisions.
