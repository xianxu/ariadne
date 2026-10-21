---
id: 000520
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '85ddedbbc5084943138799b9dd4e438f224fb584' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead: one confirmation round-trip before acting on an irreversible request submitted remotely

## Problem

remote-session (ops project, operator-approved 10-20): a remote instruction can make the TL do anything it could do
locally; least capability limits the API, not the agents.

## Spec

`tl-lead --help` gains: a submit tagged "submitted remotely" that asks for something irreversible or outside the
normal flow (machine changes, credentials, deleting data, force-pushes, turning gates off, publishing outside the
team) gets **one confirmation round-trip** to the operator (on the phone) before the TL acts.

## Done when

- The rule is in `tl-lead --help`, with a test pinning it.

## Plan

- [ ]

## Log

### 2026-10-20
