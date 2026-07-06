---
id: 000149
status: working
deps: []
github_issue:
created: 2026-07-01
updated: 2026-07-05
estimate_hours:
started: 2026-07-05T21:50:49-07:00
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

## Spec

Make command-tree tests hermetic w.r.t. the repo lock: a test invoking
`buildRoot().Execute()` should serialize (if at all) on a lock rooted in the
test's own temp dir, not the developer's real checkout. Two candidate shapes:

- inject the lock directory/root through the command wiring so tests point it at
  `t.TempDir()`; or
- have the lock key off the resolved repo root the command operates on (already
  temp in these tests) rather than process cwd.

Prefer the option that keeps production behavior identical (the lock must still
serialize real mutating verbs on the real checkout) while letting tests isolate.

## Done when

- [ ] `go test ./cmd/sdlc/...` runs green while an unrelated live `sdlc`
      mutating command holds `.git/sdlc.lock` (no hang).
- [ ] The affected command-tree tests acquire a lock under a test-local temp
      dir, not the real checkout's `.git/sdlc.lock`.
- [ ] Production lock behavior unchanged — real mutating verbs still serialize on
      the common-dir lock (`internal/repolock` semantics intact).

## Plan

- [ ] Locate the lock-root resolution in `internal/repolock` + how
      `buildRoot().Execute()` reaches it; find the seam to redirect the root.
- [ ] Redirect command-tree tests' lock root to `t.TempDir()` (injection or
      repo-root-derived key).
- [ ] Add a regression test: two `Execute()` runs against distinct temp roots do
      not contend; a concurrent real-lock holder does not block a temp-rooted test.

## Log

### 2026-07-01

- Filed from the #140 close boundary review (SHIP), "test-hygiene backlog" note:
  `buildRoot().Execute()` tests grab the real cwd repo lock, so the suite hangs
  when a live `sdlc` (here `sdlc close --issue 140`, pid 82204) holds it.
