---
id: 000528
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'a86e7fb068e7767662b233feb00a174cce5096ee' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T12:23:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "fd446026", done: "65a2416d"}
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

- [ ] Add rendered-help guards for both entry points, with each recovery step checked in order and its instructions confined to that step; observe the missing-section failure.
- [ ] Add TAKEOVER before ROLES: read the contract, session and Threads stamp, transcript only for a material gap, re-establish live work and refresh Threads, then give a one-line notice including possible loss. Preserve operator-started fleet recovery.
- [ ] Verify the focused tests and own built CLI help, run the recorded suite, close and land.

Scope: quick-flow embedded help and phrase guards; no runtime changes. ARCH-DRY keeps the procedure in the existing embedded help, with its status writer owning the fresh stamp. No new state or hot path.

## Log

### 2026-10-21

- Recovery: claim/start-plan survived; no #528 recorded check or merge runner exists. The operator-approved procedure is settled; implement as quick-flow help text.
