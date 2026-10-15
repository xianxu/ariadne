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

Durable plan: `workshop/plans/000373-project-lifecycle-verbs-plan.md`.

- [ ] Pure core: `TickLandedIssue`, `MVPProgress` (mvp-landed guard on it)
- [ ] `tickLanded` + `sdlc project tick` (origin publish via TrunkFile, TL-owned allowed)
- [ ] merge post-archive hook
- [ ] `project status` MVP + done_when
- [ ] `project close` --verified, MVP re-check, --dry-run, symlinked record
- [ ] docs: helptext, project datatype ownership, atlas
- [ ] verify on ariadne-robustness-1 (dry-run), full suite

## Log

### 2026-10-14
