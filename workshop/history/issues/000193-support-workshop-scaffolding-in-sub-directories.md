---
id: 000193
status: punt
deps: []
github_issue:
created: 2026-08-16
updated: 2026-08-20
estimate_hours:
---

# support workshop scaffolding in sub directories

## Problem

some repos contains related but largely independent parts, for example, the kbench repo which contains competition code for various competitions in kaggle. each competitions are totally separate from each other. conceptually, we can make a git repo for each competition, but for whatever reasons I made a single repo for it. this at least makes some aspect easy, and such bigger multi-purpose repo seems have some utility anyway. 

one issue with current sdlc process in ariadne is it is repo global. this task is about systematic change to ariadne's sdlc process, basically artifacts inside workshop/ folder, so that we can create instances of workshop for subdirectories. some thoughts:

1. ariadne should be aware of these different workshops. 
2. which shop is the "default" workshop, basically depend on where coding agent is started.
3. we should have a skill to bootstrap workshop structure inside a subdirectory. 
4. issues can be address like kbench/competition/arc-agi-3#3, to note it's in kbench repo, and competition directory. do an audit what other places we need to extend. 

I'm not quite sure if we should go down this path; or when we are dealing with kbench style repo, have some lighter weight process. the trigger for me to make this task, is when working in kbench/competition/arc-agi-3, there's some amount of code, and test, but no defined sdlc processes there, thus there's no planning stage, or various quality gates. on the other hand, full sdlc might be too cumbersome for it and complicate a lot of things, e.g. the point 4 above feels quite a deep hole.

## Spec

## Done when

-

## Plan

- [ ]

## Log

### 2026-08-20

- Punted. Per-subdirectory workshops add too much complexity for the payoff:
  point 4 (hierarchical refs like `kbench/competition/arc-agi-3#3`) alone
  ripples through id allocation, resolution, deps, project/roadmap
  center-of-gravity, and the close gate's fleet-wide project sweep — a deep
  hole, as the Problem section already suspected. Standing decision: **one
  repo, one workshop.** A kbench-style multi-purpose repo keeps a single
  repo-global `workshop/`; if the lighter-weight-process itch returns, that is
  a separate issue about process weight, not about workshop topology.
- Reopen (`sdlc issue set-status --issue 193 working`) if the one-workshop
  constraint starts actually hurting.

### 2026-08-16
