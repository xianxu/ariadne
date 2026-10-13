---
id: 000346
status: working
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '19ad9672eaa649c50329c0ffaa6e28b5b63692b2' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T13:40:11-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "3686a4f9", done: "8536d9ff"}
---

# test gate: a sharded full suite counts as both suite slots (two ariadne suites at once pushed load to 185)

## Problem

#341 lets two full suites run at once machine-wide (WF_TEST_SUITES=2). That suits pair's suite, but ariadne's sdlc suite runs ~12 parallel shard processes; on 2026-10-13 two of them at once (ariadne:2 #305, ariadne:4 #282) pushed load to 185 and starved other slots' work (pair's timing check failed at 12 s vs 6 s).

## Spec

A suite declares its weight against the machine-wide budget: a sharded suite takes both slots (runs alone); a light suite takes one, as today. Simplest form: `scripts/test-lock.py` accepts a weight (e.g. `--slots 2` from the sharded runner, or a per-repo declaration read by Makefile.workflow), acquires that many of the WF_TEST_SUITES slots atomically (no partial holds, no deadlock between two weight-2 waiters), and the waiting message names the holders. Targeted runs stay unaffected (#341).

## Done when

- Two ariadne sharded suites never run at once; a sharded suite waits for a running pair suite and vice versa per the weights; tests with fake lock files cover atomic multi-slot acquisition and the no-deadlock case; never generate real load.

## Plan

Design. The weight is a per-repo declaration: `WF_TEST_WEIGHT` in the root
Makefile, set before `Makefile.workflow` is included (ariadne: 2; default 1).
`Makefile.workflow` passes it to `test-lock.py acquire --slots N`, clamped to
`WF_TEST_SUITES`.
- **Atomic:** one scan takes all N free slots or none, so a waiter never holds a
  partial set.
- **Queue:** while waiting, a run holds the queue file (`<lock>.queue`, flock).
  Only the head of the queue may take slots, so light suites can't starve a
  heavy waiter.
- **No deadlock:** the queue head holds no slot, and slot holders wait on
  nothing.
- **Records:** each slot file gets the holder's record, `slots` listing all of
  them. The wait message names each holder once, plus who is next in line.

Lifecycle: no new durable artifact beyond one queue lock file per machine,
beside the slot files, rewritten in place and never growing.

- [x] `test-lock.py --slots N`: atomic multi-slot take, the queue, deduped holders.
- [x] `Makefile.workflow` passes `WF_TEST_WEIGHT`; ariadne's Makefile declares 2.
- [x] `scripts/test-lock.test.sh`, with fake lock files and `sleep` holders
      only. Cases:
  - a weight-2 suite takes both slots;
  - a weight-1 suite waits behind it, and it waits for a weight-1 suite;
  - it holds no slot while waiting;
  - a light newcomer queues behind a heavy waiter;
  - two weight-2 waiters both finish (no deadlock);
  - the clamp;
  - the make wiring.
- [x] Atlas: the sdlc-binary "Two full suites" section.

## Log

### 2026-10-13
- 2026-10-13: closed — make test green at 32a52802 under the new weighted lock (held slots [0,1]): 12 shards, 1104 cmd/sdlc tests plus 46 packages, 0 failed. bash scripts/test-lock.test.sh all pass, with private mktemp lock files and sleep holders only. New cases: a weight-2 take of both slots; a light suite waits behind it and a heavy suite waits for a light one; no partial hold while waiting; a light newcomer queues behind a heavy waiter; two heavy waiters both finish (no deadlock); the clamp; the make wiring with ariadne weight 2 and a WF_TEST_WEIGHT=1 override. Mutation-checked the queue head rule.; review verdict: FIX-THEN-SHIP

- Dispatched by the TL (ops:0); claimed in ariadne:4.
- `WF_TEST_WEIGHT ?= 2`, not `:=`: an operator's env override (e.g.
  `WF_TEST_WEIGHT=1`) must win; the make-wiring test caught `:=` beating it.
- `bash scripts/test-lock.test.sh`: all pass (fake lock files, `sleep`
  holders, no load). Mutation-checked: letting a non-head waiter take slots
  fails "a light newcomer queues behind the heavy waiter" and "the heavy
  waiter takes both slots once free".
- Close review round 1 (FIX-THEN-SHIP), 4 Minors, all fixed in round:
  - the `?=` form in the Makefile.workflow comment;
  - the AGENTS.base.md machine budget names weighted suites;
  - the "cannot starve" claim narrowed to "never overtake the head waiter",
    since the queue is a polled flock, not FIFO;
  - test-lock.test.sh unsets WF_TEST_WEIGHT.

  The lock tests pass again.
