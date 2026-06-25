---
id: 000123
status: open
deps: []
created: 2026-06-24
updated: 2026-06-24
---

# sdlc state can track all ariadne changes

right now, `sdlc state` operate one repo at a time, which is by design. but sometimes I want to make sure all "dependency (ancestor) repos are clean" before I start some big change. so let's add a mode `sdlc state --full` to check also state of dependencies declared of the current repo. 

and one small thing, to display in green when branch's at main and same as origin. yellow if it's on main but not aligned with origin and red if it's on a different branch. 

## Done when

-

## Spec


## Plan

- [ ]

## Log

### 2026-06-24

