---
id: 000565
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '7af06a10c69a15393562f4b3cec441abf2ca6ebe' # card fields mirrored from issue-cards; edit via sdlc
---

# Merge intents survive a full shutdown: record them and give one exact re-entry step

## Problem

## Spec

From the 10-22 full-shutdown recovery: **merge runners don't survive a full Couch shutdown,** so every landing in
flight had to be re-entered by hand afterwards (each slot then also had to decide whether the shutdown freeze still
applied). Make re-entry automatic: on restart, a slot whose issue has a durable merge intent (an admitted merge or a
landing ticket recorded before the shutdown, and no live runner) gets a single clear next step from `sdlc` (or the
worker's restart prompt) to re-enter `sdlc merge` on the same head, without duplicating suites that already passed.
No automatic merge without the slot re-entering it; just a durable record and an exact recovery command.

## Done when

- After a simulated full shutdown with an admitted merge, `sdlc landing status`/the worker restart summary names the
  orphaned merge intent and the exact re-entry command (test).
- Re-entering reuses a passed recorded suite on an unchanged head (test).

## Plan

- [ ]

## Log

### 2026-10-22
