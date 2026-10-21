---
id: 000505
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'a33ae2634183930e0f548914a3f00863d25bf225' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T17:05:04-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "2f75c9c9", done: "fe0d87c7"}
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

- [x] Pin the operator-review rule and remove the contradictory taste/smoke-only summary.
- [x] Verify rendered help, update the atlas map, and prepare recorded-suite/close requests; require passing results before landing.

## Log

### 2026-10-20

Claimed and started in ariadne:5 after #503 landed. Implementing the operator-reviewed Spec from dispatch. ARCH-DRY: keep the rule in the lead’s embedded help, beside existing design approval; tests inspect the rendered help and the atlas points there. Quick plan posted to hub, proceeding unless objected. No runtime state, formats, or new workflow mechanism.

Red: rendered-help regression failed on every missing operator-review clause and on the taste/smoke-only playbook sentence (package 0.202s). Added the rule beside design approval, updated the summary, and mapped it in atlas/workflow/tl-skills.md. Green: full targeted tl-lead, tl-worker and pkg/tlrules package tests passed (7.809s / 0.202s / 0.197s); git diff --check clean. Preparing pushed-head recorded suite and binary-owned close review; both must pass before landing.
