---
id: 000518
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'dbe91c1e842f3ce5e9a7668f6f1850c55e5d3a2c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T20:55:45-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ] Pin the six shaping steps, drivers/reviews, mock conventions and development-only gradient in rendered-help contract tests; verify red before editing help.
- [ ] Add SHAPING and connect the execution playbook; add optional project body sections; synchronize README and atlas.
- [ ] Run targeted help/package checks, inspect rendered help, and queue the recorded suite; close and land through SDLC.

## Log

### 2026-10-20

- Dispatch authenticated as submitted from ops:0 (delivery 3e7eb79e-7308-49a7-ab4b-9c89946e1db1). Read the operator-settled ops product-shaping/design.md and the issue's six-step Spec. The dispatch authorizes quick-flow execution; no new product design is being proposed.
- Quick plan posted to the hub, proceeding unless objected to. The canonical TL procedure stays in embedded help (ARCH-DRY); the prototype only describes optional project shape. Optional sections change no required fields, lifecycle or parser behavior (ARCH-EVOLVE). No runtime effects, durable state machine or external dependency is introduced; the server and cast-sketch remain pair-owned work.
- Selected files: cmd/tl-lead/help.md and shaping_help_test.go, construct/datatype/project.md, README.md and atlas/workflow/tl-skills.md. Keep shipped help additions within the 100-line quick-flow shell. The existing slot worktree is isolated on the fresh issue branch prepared by start-plan, based on main after #485 landed.
