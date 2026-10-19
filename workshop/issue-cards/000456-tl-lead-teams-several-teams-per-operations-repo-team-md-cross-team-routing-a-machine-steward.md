---
id: 000456
status: open
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
---

# tl-lead teams: several teams per operations repo, team.md, cross-team routing, a machine steward

## Problem

tl-lead assumes one hub per operator, the slot Couch started from, with its state in `teams/<team>/` of the ops repo. The operator (2026-10-18, project introduce-team) wants teams as first-class: an operations repo holds several teams, each with its own TL (hub), named after the team, under a different agent contract than workers (tl-lead vs tl-worker), and aware of each other.
