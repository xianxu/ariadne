---
id: 000528
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c3a9a1c4ab2c14579572c6dd9879cab4231031da' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead: hub takeover procedure after a forced agent switch (session file, Threads stamp, transcript for the gap)

## Problem

## Spec

The operator decided (ops session.side/2026-10-21-09-hub-standby.md, 10-21): no hot-standby hub; instead an emergency
**force switch** of the hub's coding agent (pair#728), after which the new agent must take over as TL. tl-lead needs
an explicit takeover procedure for that moment (also for any fresh hub session):

1. Run `tl-lead --help` and follow it.
2. Read the session file (and `session.side/`). The **Threads section's `_updated_` stamp** says how current it is.
   The operator uses the session file as an ephemeral scratch pad (deletes items once handled), so it is not a
   complete history; tolerate some loss.
3. If the gap since that stamp matters, read the previous agent's **terminal transcript from that time on**
   (`couch --peek <hub> --previous --all`, or the agent's own transcript on disk), only for the gap.
4. Re-establish: peek working slots, check `sdlc landing status` / `queue status`, open project files; then update
   Threads with a fresh stamp and tell the operator in one line that the takeover is done and what (if anything) may
   have been lost.

The same procedure applies whether the new agent is in the same slot (force switch) or a fresh session.

## Done when

- tl-lead's help has a concise TAKEOVER section with the steps above; rendered-help phrase guards cover it.

## Plan

- [ ]

## Log

### 2026-10-21
