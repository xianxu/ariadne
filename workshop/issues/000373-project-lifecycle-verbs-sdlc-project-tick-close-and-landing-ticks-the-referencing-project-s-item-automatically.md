---
id: 000373
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-15
estimate_hours:
card_mirror: 'ebcf05fad7b51c7bd8cf07eebe661a1f4a05bf69' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T01:00:50-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "eadd258f", done: "8004846b"}
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

Durable plan: `workshop/plans/000373-project-lifecycle-verbs-sdlc-project-tick-close-and-landing-ticks-the-referencing-project-s-item-automatically-plan.md`.

- [x] Pure core: `TickLandedIssue`, `MVPProgress` (mvp-landed guard on it)
- [x] `tickLanded` + `sdlc project tick` (origin publish via TrunkFile, TL-owned allowed)
- [x] merge post-archive hook
- [x] `project status` MVP + done_when
- [x] `project close` --verified, MVP re-check, --dry-run, symlinked record
- [x] docs: helptext, project datatype ownership, atlas
- [x] verify on ariadne-robustness-1 (dry-run), full suite

## Log

### 2026-10-14

### 2026-10-15
- TL approved points 2–4 of the plan; point 1 (landing ticks TL-owned projects) is the operator's call, so TL-owned projects are opt-in via `landing_tick: true` (default `project.TLLandingTickDefault = false`) and merge prints `sdlc project tick --issue R --pr N` for the TL instead. Flipping the default is one line.
- Landed issue comes from the branch's id prefix (#272: one issue, one branch), so a resumed landing still ticks; the tick is idempotent.
- Real-file dry runs: `project status` on ariadne-robustness-1 prints `mvp: 44/44 landed`; `project close --dry-run` refuses on status executing and the missing retro and would close the real teams/ file in place; `project tick --issue ariadne#240 --pr 241 --dry-run` → no change (row names PR #241); `--issue ariadne#376` → would publish to ops's origin.
- Bug found and fixed: `project close` on a linked ops record renamed the link and orphaned teams/…/project.md.
- Mutation checks: disabling the landing_tick policy, the durable-merge hook call, and the linked-close branch each fail their tests.

