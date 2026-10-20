---
id: 000484
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '660ce95ac980fa801ae8c49fd1c92c9936d6eee1' # card fields mirrored from issue-cards; edit via sdlc
---

# Suite oversubscribes the machine: 12 shards alone push load to 150+ (51% kernel); measure 12/8/6 and pick the default

## Problem

10-19 22:50: one ariadne full suite (12 parallel `sdlc.test` shards) alone drove the machine's load average to
150+ (1-minute 157, 15-minute 99) with 51% of CPU in the kernel and 0% idle; no other test ran. Other slots' work
stalled behind it, and a load-sensitive test (ariadne#480) failed twice under such load. The machine has 12 cores;
12 git- and process-heavy shards plus their children oversubscribe it badly, mostly in syscalls (process spawn,
filesystem).

## Spec

Measure, then size: run the suite at 12, 8 and 6 shards (3 runs each, quiet machine) and record wall time, the peak
1-minute load and kernel CPU share. Choose the shard count that keeps wall time within ~10% of the best while keeping
peak load under about 2× the core count. Make it the default (keep it overridable), and set `WF_TEST_WEIGHT` to match.

## Done when

- The measurements are in the issue Log (a table), and the default shard count follows them.
- A note in the test-economy project with the before/after peak load.

## Plan

- [ ]

## Log

### 2026-10-19
