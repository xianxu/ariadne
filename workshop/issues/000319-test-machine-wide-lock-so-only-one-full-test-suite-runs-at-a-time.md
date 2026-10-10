---
id: 000319
status: open
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: 'd20f0b6304c1b8ca6579324efe9f1b128f25b80c' # card fields mirrored from issue-cards; edit via sdlc
---

# test: machine-wide lock so only one full test suite runs at a time

## Problem

Under parallel slots, full test suites collide. Each repo's full `make test` uses every core by design (ariadne's `test-shard.py` defaults to 12 shards; parley.nvim's integration suite and pair's suites are heavy too). On 10-09, ariadne:4's sharded run plus parley.nvim:1's suite pushed the load average to 85 on 12 cores. Ghostty froze for about 5s and Couch dropped its connection (ops learnings for ariadne-robustness-1, kinks 22 and 24). The interim convention, "targeted tests while iterating, and the TL staggers full runs", depends on a human or TL remembering.

## Spec

**Operator decision (10-10):** a machine-wide lock. Each repo keeps its own test contract and uses all local resources for its full run; the lock serializes *who* runs a full suite. Fewer shards under load was the rejected alternative.

- **Scope:** full suites only (`make test` and its equivalents in woven repos). Targeted runs (`go test -run …`, a single spec) stay unlocked.
- **Shared, via the base layer:** one lock helper that every woven repo's full-test target goes through, so ariadne, pair and parley.nvim all share it. The lock lives in a per-user, machine-wide location, not in any repo or worktree.
- **macOS:** there's no `flock(1)`. Use a kernel-released lock (e.g. `fcntl` via Python, which `test-shard.py` already needs, or a Go helper), so a crashed or killed run never leaves a stale lock.
- **Waiting is visible:** while waiting, print who holds the lock (repo, slot or worktree, pid, start time) and keep waiting. An escape hatch (an env var) skips the lock, logged loudly.
- **Gates:** `sdlc close` and `milestone-close` run the full suite inside agent turns. A long wait must not trip their timeouts, or the wait must be reported as such.

## Done when

- Two full `make test` runs started from different slots (or repos) run one after the other; the second prints the holder while waiting (a test drives two concurrent runs of the helper).
- Killing the holder releases the lock immediately.
- A targeted `go test -run` doesn't take the lock.
- The helper ships in the base layer, and at least ariadne's and pair's full-test targets use it after `weave refresh`.
- The atlas describes the lock and the escape hatch.

## Plan

- [ ]

## Log

### 2026-10-10
