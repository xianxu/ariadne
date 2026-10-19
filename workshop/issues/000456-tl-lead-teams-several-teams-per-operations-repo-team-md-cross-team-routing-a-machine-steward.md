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

- tl-lead help and skill describe teams, team.md, routing and the steward role; `tl-lead team new <name>` scaffolds a team folder.
- A second team in ops (echo) runs with its own session file, and cross-team routing works in a live exchange.

## Plan

Durable plan: [workshop/plans/000456-tl-lead-teams-plan.md](../plans/000456-tl-lead-teams-plan.md) (design: ops `teams/devinfra/projects/introduce-team/design.md` §1, §2, §4).

- [ ] Team model and `--team` resolution (`team:<name>`, `.hub` claim grammar mirroring pair's, `team.md`)
- [ ] `tl-lead team new` (scaffold and adopt, resumable) and `team list [REPO]`
- [ ] Session FILE defaults to the team's `tl-status-*.md`
- [ ] `tl-lead team commit` (path-limited, pull --rebase, push, resumable)
- [ ] `tl-lead team issues` (open, unclaimed, in owned repos; `sdlc state --json`)
- [ ] Help: tl-lead TEAMS, tl-worker direct filing, atlas
- [ ] Full suite; read-only manual check; hand the live proof to the hub

## Log

### 2026-10-18

### 2026-10-19
- Claimed and started from ariadne:4 after #455. The hub clarified `.hub` (bare word = repo claim, `repo:N` = slot claim, `key=value` directives are pair's; tl-lead writes only claims). The operator decided on `team:<name>` addressing and direct cross-team filing (no hub relay) plus an unclaimed-issue view over `owns:`.
- Durable plan drafted; decisions D1–D5 are in it for hub/operator review (the operator reviews on the morning of 10-19).
- The hub approved the plan and confirmed D1-D5 as proposed. change-code is held until the operator's morning review of the introduce-team designs; the hub will relay the answer.
