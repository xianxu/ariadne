---
id: 000542
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'ca3d82bd6d3ecc1b950780f59200ec3b700cedf2' # card fields mirrored from issue-cards; edit via sdlc
---

# issue new in a legacy (no tracker) repo leaves a local issue-sync commit on the current branch, contrary to help

## Problem

## Spec

From ops-echo:0 (10-21): in echo, a legacy repository (no `workshop/issue-tracker.json`, no issue-tracker branch),
`sdlc issue new` publishes the issue to main **and** leaves a local issue-sync commit on the current branch, though
`sdlc issue --help` says creation writes no checkout (#308). Seen twice: filing echo#2 from echo:0 on main, and
echo:0 filing echo#9 while on its echo#3 issue branch, which then had to rebase the stray commit off (a stacked
commit on an issue branch, against #272).

Fix: on the legacy path, creation behaves like the tracker path, publishing to origin main without committing in the
checkout; or, if legacy creation genuinely needs a checkout commit, the help says so and it refuses on an issue
branch instead of committing there. Decide which, and consider migrating echo to the tracker.

## Done when

- In a legacy-fixture repo, `sdlc issue new` from an issue branch leaves that branch's HEAD unchanged and the issue on
  origin main (test).
- Help and behaviour agree for both repository kinds.

## Plan

- [ ]

## Log

### 2026-10-21
