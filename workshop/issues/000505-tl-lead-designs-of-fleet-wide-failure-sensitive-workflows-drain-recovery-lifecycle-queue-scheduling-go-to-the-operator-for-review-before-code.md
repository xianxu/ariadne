---
id: 000505
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '484d72be83ed0880b5afd8d8187f620e41747b7c' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead: designs of fleet-wide, failure-sensitive workflows (drain, recovery, lifecycle, queue, scheduling) go to the operator for review before code

## Problem

The TL approved pair#722's fleet drain through several revisions (acks, a queue barrier, a message hold, persisted
state). Its first live run on 10-20 wedged the whole fleet: refused acks, all messages held, a cancel that never
confirmed, state that survived a reboot. The operator's own model, written in ten lines afterwards, was far simpler
and robust (steward-driven, an in-memory flag, agents keep accepting messages). Operator, 10-20: "this also shed
light of where operator needs to pay more attention: those type product/system design that needing assess risks of
various event and design workflow that's robust ... those are the design that should be reviewed by the operator, a
lot of different ways things can be done, and AI tends to both over engineer and under engineer."

## Spec

Add a rule to `tl-lead --help` (the TL's design-approval section):
- **Operator-reviewed designs:** a design for a **workflow that must stay robust across failures** goes to the
  operator for review before code, as a side doc with the options, not just the TL's approval. Recognize it by its
  shape: it coordinates many slots or processes, holds or pauses something fleet-wide, persists state across
  restarts, or defines a recovery path. Examples: drain and shutdown, restart recovery, slot lifecycle and admission,
  the landing queue and trains, scheduling.
- The side doc states the internal states, the few states the user sees, and one recovery action for each (the
  existing "agent designs: over- and under-engineered" guidance), plus what happens when each step fails or a
  restart lands in the middle.
- The TL may still approve ordinary designs alone. When unsure, ask.

## Done when

- `tl-lead --help` carries the rule with the recognition shape and examples; a test pins the section.

## Plan

- [ ]

## Log

### 2026-10-20
