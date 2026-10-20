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
4. **`tl-lead status check` staleness isn't scoped either** (from ops-echo:0, 10-19 23:40): echo's Threads and Slots
   were marked stale by a parley.nvim merge (ccbc2dd5), a repo echo's `.hub` doesn't claim. Staleness should come only
   from the repositories the hub's `.hub` claims, so devinfra landings don't ping an idle team.

## Spec

- `status slots` and `fleet resume` scope to the hub's own slots by default (from `.hub` and Couch's ownership,
  hub_ownership.go rules), with `--all` for the machine view.
- tl-lead reaches Couch the way agents can: either via the exempt bare `couch` invocation (spawned so the sandbox
  rule matches), or by reading the state Couch publishes for this (whichever pair/ariadne already supports); if
  Couch can't be read, say so once at the top, not on every row.
- `fleet resume` hides unclaimed dormant repos unless `--all`, and prints a one-line count instead.

## Done when

- `status check` in a two-hub fixture marks a hub stale only for merges in repos its `.hub` claims (test).

- In a two-hub fixture, each hub's `status slots` lists only its own slots; `--all` lists every slot.
- A run where Couch is unreachable prints one note, not N per-row errors (test).
- The dry run on the live machine shows only owned slots plus a count of unclaimed ones.

## Plan

- [ ]

## Log

### 2026-10-19
- 10-20 11:00 (TL): add to scope. After the 10-20 reboot, `tl-lead fleet resume --apply` fired `couch --resume` for
  14 slots back to back: 5 came up, 9 ended "context deadline exceeded … outcome uncertain" (Couch's 2 s admission,
  pair#726), and the operator resumed the rest by hand from the switcher. fleet resume should (1) resume one slot,
  then poll `couch --operation-status` for that request until it settles before the next; (2) on an uncertain
  outcome, check the slot's state and stop the batch, naming the remaining slots and the manual path, instead of
  firing the rest into the same timeout. pair#726 fixes the admission itself; this keeps tl-lead honest either way.
- 10-20 11:15 (TL): test spec for the serial resume above (operator: their manual pair:0–8 resume, one at a time,
  tens of seconds each, all succeeded; that's the behavior to match): unit test against a fake Couch whose resume
  takes 10–30 s, asserting serial order, a wait for each slot to settle, and stop-and-name on the first uncertain or
  failed outcome.
