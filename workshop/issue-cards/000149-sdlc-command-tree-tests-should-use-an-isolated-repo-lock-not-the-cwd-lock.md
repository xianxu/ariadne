---
id: '000149'
status: done
started: 2026-07-05T21:50:49-07:00
created: 2026-07-01
updated: 2026-07-05
estimate_hours: 0.6
actual_hours: 0.35
---

# sdlc command-tree tests should use an isolated repo lock, not the cwd lock

## Problem

Command-tree tests that drive `buildRoot().Execute()` (e.g.
`TestSetStatusAlias_BothPathsMutate` and peers) acquire the **real cwd-based
repo transaction lock** (`.git/sdlc.lock`) rather than an isolated, per-test
lock. Consequences:

- `go test ./cmd/sdlc/...` **hangs to timeout** whenever a live `sdlc` command
  holds the lock concurrently — surfaced during the #140 close boundary review:
  the suite blocked on the lock held by `pid …: sdlc close --issue 140`.
- The affected tests are **non-hermetic and un-parallelizable across processes**:
  they contend on one machine-global lock keyed off the checkout's cwd.

This is pre-existing (outside the #140 window) — flagged by the #140 close
boundary review under "test-hygiene backlog".
