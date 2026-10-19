---
id: 000456
status: open
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '2831154ed39a0c69a0571931c61dec7e5cb0f0c9' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-18
