---
id: 000329
status: working
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-12
estimate_hours:
card_mirror: '5b6d8a9ed85ed2370a6fc1ca258e17a72d36d2f0' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T16:34:43-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ]

## Log

### 2026-10-10
