---
id: 000506
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '8d533c91d7bb2f4d2eb543aa69efc6c191be812b' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-worker draining rule and tl-lead drain-the-fleet procedure (fleet-drain; recovery only on the operator's word)

## Problem

Project fleet-drain (ops `teams/devinfra/projects/fleet-drain/`, operator-designed 10-20) puts the drain's judgment
in the agents and the steward; Couch only flags slots and delivers notices (pair#722 re-scoped). The agents and the
steward need the rules.

## Spec

- **tl-worker, "draining":** on a drain notice, graceful: let your background work (suites, merges) finish, then
  write state (commit and push the branch, Log the exact next step) and stay idle. Fast: stop your own background
  work through its supported controls, preserve unfinished edits, write state, stay idle. While drained, reply to new
  messages "acknowledge message received, will resume processing after drain is lifted" and don't act on them;
  don't acknowledge acknowledgements. An optional "drained" reply to the hub is fine but nobody relies on it. On
  "drain lifted" or a restart notice, resume from the Log.
- **tl-lead, "drain the fleet":** `sdlc queue pause`; `couch --drain [--fast]`; watch idleness (Couch's idle notices,
  multi-slot `couch --peek`, `couch --drain --status`); when all are idle, notify the operator that it's safe to
  restart. After a restart, start in safe mode and **wait for the operator's "start recovery"**; then resume slots
  one at a time. To call it off: `couch --undrain [--message]` and `sdlc queue resume`.

## Done when

- `tl-worker --help` and `tl-lead --help` carry both, with tests pinning the sections.

## Plan

- [ ]

## Log

### 2026-10-20
