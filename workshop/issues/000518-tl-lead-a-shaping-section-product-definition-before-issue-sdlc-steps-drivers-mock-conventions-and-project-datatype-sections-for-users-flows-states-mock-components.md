---
id: 000518
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'b34f3128bbf396137c6c8f3aebf5e4db434b3f0d' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-20
