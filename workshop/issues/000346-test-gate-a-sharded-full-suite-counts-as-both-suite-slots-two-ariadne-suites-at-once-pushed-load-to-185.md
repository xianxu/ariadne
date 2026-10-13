---
id: 000346
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '2a6942e2a1173f793ff39710f3a2904af530acb3' # card fields mirrored from issue-cards; edit via sdlc
---

# test gate: a sharded full suite counts as both suite slots (two ariadne suites at once pushed load to 185)

## Problem

#341 lets two full suites run at once machine-wide (WF_TEST_SUITES=2). That suits pair's suite, but ariadne's sdlc suite runs ~12 parallel shard processes; on 2026-10-13 two of them at once (ariadne:2 #305, ariadne:4 #282) pushed load to 185 and starved other slots' work (pair's timing check failed at 12 s vs 6 s).

## Spec

A suite declares its weight against the machine-wide budget: a sharded suite takes both slots (runs alone); a light suite takes one, as today. Simplest form: `scripts/test-lock.py` accepts a weight (e.g. `--slots 2` from the sharded runner, or a per-repo declaration read by Makefile.workflow), acquires that many of the WF_TEST_SUITES slots atomically (no partial holds, no deadlock between two weight-2 waiters), and the waiting message names the holders. Targeted runs stay unaffected (#341).

## Done when

- Two ariadne sharded suites never run at once; a sharded suite waits for a running pair suite and vice versa per the weights; tests with fake lock files cover atomic multi-slot acquisition and the no-deadlock case; never generate real load.

## Plan

- [ ]

## Log

### 2026-10-13
