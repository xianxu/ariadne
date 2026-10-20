---
id: 000485
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'd747157879b47d4556649214bac4c28b1567f284' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead in a multi-hub setup: scope status slots and fleet resume to the hub's own slots; couch calls fail under the sandbox (actors '?')

## Problem

From introduce-team's live proof (10-19, two hubs: `couch ops ops-echo`):
1. **`tl-lead status slots` isn't scoped to the hub.** In ops-echo:0 it rendered the whole machine's fleet into echo's
   status file, not just the slots its `.hub` claims (echo).
2. **tl-lead's couch calls fail inside the agent sandbox.** `tl-lead status slots` shows `?` in every actor column,
   and `tl-lead fleet resume` lists every slot as "unknown: couch --actors unreadable … dial unix
   /tmp/pair-message-501/….sock: operation not permitted". The sandbox exemption covers a bare `couch` call only, so
   a couch call made from inside tl-lead is blocked. Run outside the sandbox, `fleet resume` classed correctly (one
   slot to resume, the rest running or idle).
3. `fleet resume`'s dry run prints dozens of "ambiguous … unclaimed" rows for dormant repos (robotics, zellij,
   xianxu.dev, …), which buries the one actionable row.

## Spec

- `status slots` and `fleet resume` scope to the hub's own slots by default (from `.hub` and Couch's ownership,
  hub_ownership.go rules), with `--all` for the machine view.
- tl-lead reaches Couch the way agents can: either via the exempt bare `couch` invocation (spawned so the sandbox
  rule matches), or by reading the state Couch publishes for this (whichever pair/ariadne already supports); if
  Couch can't be read, say so once at the top, not on every row.
- `fleet resume` hides unclaimed dormant repos unless `--all`, and prints a one-line count instead.

## Done when

- In a two-hub fixture, each hub's `status slots` lists only its own slots; `--all` lists every slot.
- A run where Couch is unreachable prints one note, not N per-row errors (test).
- The dry run on the live machine shows only owned slots plus a count of unclaimed ones.

## Plan

- [ ]

## Log

### 2026-10-19
