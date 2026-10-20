---
id: 000484
status: wontfix
created: 2026-10-19
updated: 2026-10-20
estimate_hours:
github_issue:
---

# Suite oversubscribes the machine: 12 shards alone push load to 150+ (51% kernel); measure 12/8/6 and pick the default

## Problem

10-19 22:50: one ariadne full suite (12 parallel `sdlc.test` shards) alone drove the machine's load average to
150+ (1-minute 157, 15-minute 99) with 51% of CPU in the kernel and 0% idle; no other test ran. Other slots' work
stalled behind it, and a load-sensitive test (ariadne#480) failed twice under such load. The machine has 12 cores;
12 git- and process-heavy shards plus their children oversubscribe it badly, mostly in syscalls (process spawn,
filesystem).
