---
id: 000518
status: done
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '4bd94891f24647b1a0d8de54689c70ded78f97d9' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T20:55:45-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: operator, spec: "b36308fd", done: "e981f94b"}
actual_hours: 0.10
---

# tl-lead: a SHAPING section (product definition before issue SDLC: steps, drivers, mock conventions) and project datatype sections for users, flows, states, mock, components

## Problem

Projects have gone from a paragraph straight to guessed slices, with UIs one-shot in code. Project product-shaping
(ops `teams/devinfra/projects/product-shaping/`, operator-approved 10-20) adds a shaping stage before issue-level SDLC,
run by the TL in the team's operations repo at :0, with artifacts committed straight to main there.

## Spec

- `tl-lead --help` gains a "SHAPING" section: the six steps and who drives each (1 users and moments: operator;
  2 flows: operator, TL spots gaps; 3 clickable mock: TL builds, operator sketches, numbered versions; 4 visual design:
  operator's style references, TL's two or three directions on the same wireframe, a tokens.css; 5 components: TL
  proposes, operator smell-checks; 6 slices: TL drives, operator reviews), what's reviewed when, the mock conventions
  (`couch mock serve`, a state list, one fixtures.json, empty/error/long-list states), and the mock → real gradient.
- The project datatype (`construct/datatype/project.md`) gains optional sections: Users, Flows, States, Mock,
  Components, so a shaped project has a known form.

## Done when

- The help section and the datatype sections exist, with tests pinning the help text; the atlas points to them.

## Plan

- [x] Pin the six shaping steps, drivers/reviews, mock conventions and development-only gradient in rendered-help contract tests; verify red before editing help.
- [x] Add SHAPING and connect the execution playbook; add optional project body sections; synchronize README and atlas.
- [x] Run targeted help/package checks, inspect compiled help, and queue the required recorded suite.

## Log

### 2026-10-20
- 2026-10-20: closed — Complete tl-lead package passes (21.024s); staged-file Go lint catalog passes (0.845s); diff check clean. New rendered-help tests first failed on missing SHAPING and pass at --help and no-args entry points; mutations swapping the Flows driver and permitting mock-backed production screens both fail. Compiled bin/tl-lead help inspected. Scratch project with Users, Flows, States, Mock and Components passes sdlc project validate. README/atlas updated. Recorded suite queued on pushed implementation head; required final-head pass remains before landing. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

- Dispatch authenticated as submitted from ops:0 (delivery 3e7eb79e-7308-49a7-ab4b-9c89946e1db1). Read the operator-settled ops product-shaping/design.md and the issue's six-step Spec. The dispatch authorizes quick-flow execution; no new product design is being proposed.
- Quick plan posted to the hub, proceeding unless objected to. The canonical TL procedure stays in embedded help (ARCH-DRY); the prototype only describes optional project shape. Optional sections change no required fields, lifecycle or parser behavior (ARCH-EVOLVE). No runtime effects, durable state machine or external dependency is introduced; the server and cast-sketch remain pair-owned work.
- Selected files: cmd/tl-lead/help.md and shaping_help_test.go, construct/datatype/project.md, README.md and atlas/workflow/tl-skills.md. Keep shipped help additions within the 100-line quick-flow shell. The existing slot worktree is isolated on the fresh issue branch prepared by start-plan, based on main after #485 landed.

- Implemented 53 added shipped-help lines (within the quick shell), both rendered-help tests, optional project summaries and authoring guidance, README and atlas pointers. The help remains the canonical procedure; no worker rules or runtime mock/server implementation were added.
- Red: both new tests failed because SHAPING was absent. Mutation checks then swapped the Flows driver and allowed mock-backed production screens; each failed at both help entry points. Restored text passes TestLeadHelp/TestRoot and the complete tl-lead package (21.024 s). Staged-file Go lint catalog passes (0.845 s); git diff --check is clean.
- Built bin/tl-lead and inspected its rendered SHAPING section. A scratch defined project with all five optional sections conforms under sdlc project validate; no new schema fields or required sections were introduced. This deliverable changes shipped instructions, with no new live-environment operation to stage.

- Published PR #359 and queued the required recorded full suite on the pushed implementation head. Close review may overlap this queued suite under tl-worker; a recorded pass on the final landing head remains mandatory.

## Revisions

- 2026-10-20 — Kept the implementation checklist separate from the later close/landing gates: its verification row ends at queuing the required suite. Publication remains mandatory and is recorded in Log when complete, avoiding a circular close prerequisite without waiving verification.
