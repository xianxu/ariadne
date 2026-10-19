---
id: 000456
status: working
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'd9045422ccd3b1624701f9634bb934e7f86ad371' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T01:26:56-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
---

# tl-lead teams: several teams per operations repo, team.md, cross-team routing, a machine steward

## Problem

tl-lead assumes one hub per operator, the slot Couch started from, with its state in `teams/<team>/` of the ops repo. The operator (2026-10-18, project introduce-team) wants teams as first-class: an operations repo holds several teams, each with its own TL (hub), named after the team, under a different agent contract than workers (tl-lead vs tl-worker), and aware of each other.

## Spec

tl-lead gains the team contract: a team folder `teams/<team>/` holds `team.md` (hub name, repos owned, what it asks other teams for, machine steward or not) and the committed `.hub` claims (pair#694's format); the session file and its side docs; and projects. tl-lead commands take a team (default: the folder they run in). Cross-team routing rule: a hub that hits an issue in a repo it doesn't own files it there and messages the owning team's hub, found through the teams' `team.md` files. One machine steward per machine (devinfra today) owns the shared test queue, quota and Couch incidents. tl-worker's help: report to the hub that claims your slot. Hub commits are path-limited to its team folder. The tl-lead help and the TL skill say all of this.

## Done when

The scope was revised to one team per repo (2026-10-19). The authoritative list is the plan's "Done when (revised)". In short:
- `tl-lead team new PATH`, `team list [REPO]` (one ownership resolver), `team issues`, and `fleet resume [--apply]`, which resumes only the slots this hub owns alone, one at a time, and treats a failed read as unknown. All four are tested against fakes, with live decodes behind `TL_LEAD_LIVE`.
- The tl-lead and tl-worker help state the hub contract, the steward (the first hub on Couch's command line), routing by `.hub`, and direct filing.
- Live: read-only dry runs from ops. The `ops-echo` live proof is the operator's and the hub's, after this lands.

## Revisions

### 2026-10-19: one team per repo
The operator decided one team per ops repo (design §7–§9). The scope shrinks to the hub contract, `.hub` as the only ownership list, a `team.md` pointer, routing, the steward as the first hub, `team new` creating or adopting an ops repo, and a "resume the fleet" verb. The full delta is in the durable plan's Revisions.

## Plan

Durable plan: [workshop/plans/000456-tl-lead-teams-plan.md](../plans/000456-tl-lead-teams-plan.md). The revised scope (one team per repo) is in its Revisions section.

- [ ] R1 `.hub` claim grammar (mirrors pair's), a `.hub` reader and `ownerOf` (Couch's rule)
- [ ] R2 `tl-lead team new PATH` (create or adopt an ops repo; resumable; refuses on any differing file)
- [ ] R3 `tl-lead team list [REPO]` (sibling repos with `.hub`; the owning hub by Couch's rule)
- [ ] R4 `tl-lead team issues` (open issues nobody has claimed, in repos `.hub` claims)
- [ ] R5 `tl-lead fleet resume [--apply]` (targets from `ownerOf`; join couch resume candidates with sdlc claims; sequential resumes)
- [ ] R6 Help: tl-lead TEAMS and hub contract; tl-worker direct filing; atlas
- [ ] R7 Full suite; read-only live dry runs from ops; hand `team new` and the live proof to the hub

## Log

### 2026-10-18

### 2026-10-19
- Claimed and started from ariadne:4 after #455. The hub clarified `.hub` (bare word = repo claim, `repo:N` = slot claim, `key=value` directives are pair's; tl-lead writes only claims). The operator decided on `team:<name>` addressing and direct cross-team filing (no hub relay) plus an unclaimed-issue view over `owns:`.
- Durable plan drafted; decisions D1–D5 are in it for hub/operator review (the operator reviews on the morning of 10-19).
- The hub approved the plan and confirmed D1-D5 as proposed. change-code is held until the operator's morning review of the introduce-team designs; the hub will relay the answer.
