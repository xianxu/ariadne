---
id: 000319
status: working
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: '5cc22288d011054cf025a7fb97bab167de265327' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-10T11:23:39-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
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

Design: the lock is taken at **parse time** by `Makefile.workflow` whenever `test` is
among `MAKECMDGOALS` — `$(shell scripts/test-lock.py acquire --watch $$PPID)`, where
`$$PPID` is make itself. A prerequisite can't do it: GNU make 3.81 (macOS) runs the
recipe-rule's own prerequisites first, so pair's prerequisite-heavy `test` would run
unlocked. Parse time precedes every recipe and needs no per-repo edits, so every woven
repo's `make test` takes it after `weave refresh` (ARCH-DRY: one helper, one wiring).

- Helper `scripts/test-lock.py` (base layer, symlinked): `flock` on
  `${XDG_STATE_HOME:-~/.local/state}/ariadne/test-suite.lock`; once held it forks a
  watcher that inherits the descriptor, closes stdio (so `$(shell)` returns), and exits
  when make exits (kqueue NOTE_EXIT, poll fallback). Kernel release on any death.
- Holder record (JSON in the lock file): repo, worktree, make pid, holder pid, start
  time. Waiters print it at once and every 30s with elapsed time.
- Re-entrant: if the holder's make pid is the caller's make or one of its ancestors
  (nested `$(MAKE) test`, make's restart after remaking includes), pass through.
- `WF_TEST_LOCK=off` skips, loudly on stderr. `WF_TEST_LOCK_TIMEOUT=<s>` bounds the
  wait; expiry fails make with "lock wait, not a test failure" so a gate run reports
  the wait as such (default: wait forever; agents run `make test` in the background).
- Durable artifact: one lock file per user, rewritten per run, never grows; removed
  with `~/.local/state` (ARCH lifecycle).
- [x] `scripts/test-lock.py` + `Makefile.workflow` wiring + manifest symlink row
- [x] `scripts/test-lock.test.sh`: serialization with holder shown, kill releases,
      re-entrant, off, timeout, and `make test` wiring via `make -n`
- [x] atlas: `atlas/workflow/sdlc-binary.md` test-suite section (lock + hatch)
- [x] pair wiring: simulated (pair's Makefile + Makefile.local, this branch's Makefile.workflow + helper): `make -n test` blocks before any prerequisite. pair's `Makefile.workflow` symlinks `../ariadne/Makefile.workflow`, so it picks the wiring up when ariadne:0 fast-forwards; `weave refresh` adds the `scripts/test-lock.py` symlink (until then the `wildcard` guard skips the lock).

## Log

### 2026-10-10
- 2026-10-10: closed — make test under the new lock: 12 shards / 1020 cmd/sdlc tests green (fix round touched only Makefile.workflow/test-lock.py/manifest, no Go); only failure processgroup TestCancellationKillsDescendants = sandbox denies /bin/ps, passes unsandboxed. Lock file named this worktree during the run and was free after. scripts/test-lock.test.sh 14/14 (serialize+holder shown, kill releases, SIGINT/HUP survive, reentrant, off, timeout, foreign record, make wiring). Pair Makefile simulated with only the Makefile.workflow symlink: make -n test blocks before prerequisites.; review verdict: SHIP
- 2026-10-10: flow upgraded quick → full — 160 added lines in code files (limit 100); an earlier round of this close already ran the full review

- Re-entrancy first walked `ps` ancestry; `ps` is denied in the agent sandbox, so
  Makefile.workflow exports `WF_TEST_LOCK_HELD_BY` (the holder's make pid) instead.
- `scripts/test-lock.test.sh`: 12/12 pass (serialize + holder shown, kill -9 of the
  watcher releases, re-entrant, off, timeout, make wiring).
- Close review round 1 (FIX-THEN-SHIP): [BR-1] Important, a valid-JSON record of the
  wrong shape crashed the waiter → `holder()` validates shape, reads as unidentified.
  Minors fixed in round: the helper now resolves beside `Makefile.workflow`'s real path
  (no manifest symlink, no weave-refresh dependency; a missing helper warns instead of
  silently skipping), and the watcher ignores SIGINT/SIGHUP (make's exit releases) and
  its poll fallback treats PermissionError as alive. Regression tests added; 14/14 pass.
  Pair re-simulated with only the `Makefile.workflow` symlink: `make -n test` blocks.
