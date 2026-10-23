---
id: 000574
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '468d8c0873073d2cc9338673bec96716c3038ab4' # card fields mirrored from issue-cards; edit via sdlc
---

# Keep slot-sibling ariadne checkouts current so queue-helper rollouts finish

## Problem

## Spec

**Slot-sibling ariadne checkouts drift, so queue-helper rollouts never finish (TL, 10-23).** Each non-ariadne slot has
a sibling `worktree/<repo>-slotN/ariadne` checkout on `main` that supplies its queue helper (`test-lock.py`, kept
checkout-supplied by ariadne#563). Nothing fast-forwards these: on 10-23 they ranged from 10-21 to 10-23. After
ariadne#570 landed, every waiter from a stale sibling counted as a legacy-policy waiter, so `sdlc landing grant` kept
refusing (correctly) with no end in sight; the TL fast-forwarded all 21 clean siblings by hand at 08:2x.

Fix: keep sibling ariadne checkouts current at a natural boundary (for example `sdlc claim`/`start-plan`, or
local-check before taking a ticket, fast-forwards a clean sibling on `main` to `origin/main`, never touching a dirty or
non-main one, and says so), and have queue status name the stale sibling behind a legacy ticket.

## Done when

- A clean sibling ariadne on `main` behind `origin/main` is fast-forwarded at the chosen boundary; dirty or off-main
  siblings are left alone with one line naming them (test).
- Queue status's mixed-policy line names the supplying checkout's revision for each legacy ticket.

## Plan

- [ ]

## Log

### 2026-10-23
