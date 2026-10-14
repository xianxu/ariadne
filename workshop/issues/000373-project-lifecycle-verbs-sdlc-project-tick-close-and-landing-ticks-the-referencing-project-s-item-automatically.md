---
id: 000373
status: open
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: 'a49b3038e86955994f27e6414d7d6367d99bdcd3' # card fields mirrored from issue-cards; edit via sdlc
---

# Project lifecycle verbs: sdlc project tick/close, and landing ticks the referencing project's item automatically

## Problem

Projects are hand-maintained files (`workshop/projects/…`, ops `teams/*/projects/*/project.md`). Nothing ticks a project's task when its issue lands, and there's no verb to close a project: on 2026-10-14 ariadne-robustness-1 showed 8 landed MVP items unticked. The close gate finds referencing projects but ticking/closing stays manual. Retro item (d); pairs with the operator's rule that the TL status file is mutual focus and its projects must stay current.

## Spec

`sdlc project tick --issue N` (idempotent; finds every referencing project fleet-wide, as the close gate already does) and run it automatically at landing (`sdlc merge`, after archive) for the landed issue. `sdlc project status PROJECT` prints done/open against mvp_scope and done_when. `sdlc project close PROJECT --verified` checks every MVP item is done and records the close (archive per the project datatype). Writes to another repo's project file go through that repo's normal publish path.

## Done when

- Landing an issue ticks it in every referencing project (test across two repos).
- `project status` and `project close` work on ariadne-robustness-1's real file (dry-run).

## Plan

- [ ]

## Log

### 2026-10-14
