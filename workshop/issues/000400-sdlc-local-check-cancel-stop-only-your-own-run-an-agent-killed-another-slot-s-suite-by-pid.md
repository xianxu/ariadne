---
id: 000400
status: working
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-17
estimate_hours:
card_mirror: '8a602b1bfc996cb83e467bd68082d2a088aee09a' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T15:01:16-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# sdlc local-check cancel: stop only your own run (an agent killed another slot's suite by PID)

## Problem

2026-10-15: pair:5 wanted to stop its own `sdlc local-check run --name suite` and picked a PID with
`ps -axo pid,ppid,command`, taking the first global match. That PID was parley.nvim:2's suite (parley#337's final
run on ff50e564). `kill -INT` ran before the ownership check (`lsof -d cwd`) was read, and it cancelled another
slot's full suite (exit 130), costing a requeue behind three other suites.

Nothing today lets an agent stop *its own* local check without hand-picking PIDs, and the global process table holds
every slot's identical-looking `sdlc local-check run --name suite` commands.

## Spec

- `sdlc local-check cancel [--name suite]` stops only the run started from the current checkout (or the current
  slot's worktree): it finds the run through its own record (the in-flight file under `.git/sdlc/`, with the
  runner's PID and process-group ID, written at start), checks it's alive and still that run (start time and
  command match), and signals its process group. Unknown, stale or foreign → refuse, naming what it found.
- The runner records PID, PGID and start time in its in-flight record when it starts (if it doesn't already).
- tl-worker's rules gain one line: stop your own checks with `sdlc local-check cancel`, never `kill` a PID found by
  `ps` (phrase-guarded in the help tests).

### Design (2026-10-17)

Ownership is held, not observed (tl-worker "held for the owner's lifetime"): each run holds a **lease file** for
its whole life, and cancel acts only on a lease that is (a) held and (b) owned by this checkout.

- **Lease** — `runLocalCheck` (the one path `local-check run` and `sdlc merge` share) creates
  `<git common dir>/sdlc/local-checks/runs/<random>.json` before it queues, takes an exclusive `flock` on it
  (`internal/flock`, already used by landingline/repolock), writes `{root, check, sha, pid, started}` and keeps the
  descriptor open until the run returns. Go opens files close-on-exec, so no child inherits it: the lock is held
  exactly while that sdlc process lives, which makes "alive and still that run" one non-blocking lock attempt
  instead of a start-time/command comparison against a PID that may have been reused (ARCH-ORDER: a lock the
  kernel releases on any death, so a stale record can't be acted on).
- **Cancel** — `sdlc local-check cancel [--name N]` resolves this checkout's top level, scans `runs/`, and for each
  record: lock free → stale, delete it (the sweep); lock held and `root` ≠ this checkout → foreign, never touched
  (listed in the output only as "another checkout's run"); held and ours (and `--name` matching) → `SIGTERM` the
  recorded sdlc PID. The runner already turns SIGTERM into a clean stop (#416's path: terminates the check's
  process group or the queued helper, releases the slot, removes the worktree, records nothing). Cancel then
  waits (bounded, 2 min) for the lease to be released and reports each run it stopped; no own run → says so, exit 0.
- **Pure core** — `classifyRunLease(rec, held, root, name) → stale | foreign | other-check | target` is a pure
  function with table tests; the scan/lock/kill shell stays thin (ARCH-PURE).
- **Removal** (ARCH-FUNERAL) — the run deletes its lease on return; a crashed run leaves an unlocked file that the
  next `cancel` or `run` sweeps. Bound: one small file per live run.
- **Attribution** — the root cause was that nothing named the owner: the queue's "held by" lines print the run's
  temp worktree (`sdlc-local-xxxx (/tmp/…)`), not the slot. The run passes `WF_TEST_LOCK_OWNER=<checkout root>`
  to `scripts/test-lock.py`, whose record uses it for `repo`/`worktree` when set, so the waiting lines name the
  owning slot.
- **Rule** — tl-worker help gains: stop your own checks with `sdlc local-check cancel`, never `kill` a PID found by
  `ps` (a `rules.tsv` phrase row). local-check help documents the verb.
- Residual: between the held-lock check and `kill`, the run could exit and its PID be reused within microseconds;
  accepted (macOS has no pidfd), and the window is far narrower than today's hand-picked PID.
- ARCH-SECURE: the lease is read from a directory shared by all slots of the clone; a malformed record is skipped
  (reported), and only a record whose lock is held *and* whose root equals ours is ever signalled.
  ARCH-CONSTRAINTS: N/A beyond the bounded 2-minute wait — a handful of files per scan. ARCH-MOCK: real processes
  and real `flock` in the tests (no external service).

## Done when

- A process test: two in-flight runs from two checkouts of one clone coexist; `local-check cancel` from one stops
  only its own run (exit, slot released, lease gone) and the other keeps running; a stale lease (dead holder) is
  swept, never signalled.
- `classifyRunLease` table test covers stale, foreign, other-check, target.
- The test-lock record names the owning checkout when `WF_TEST_LOCK_OWNER` is set (scripts/test/test-lock.test.sh).
- tl-worker --help states the rule (phrase-guarded); `sdlc local-check --help` documents cancel.

## Plan

- [ ] Lease: `localcheck_lease.go` (create/hold/remove, scan, `classifyRunLease`) + unit tests
- [ ] Hook the lease into `runLocalCheck`; pass `WF_TEST_LOCK_OWNER` to the helper; test-lock.py honours it (+ shell test)
- [ ] `local-check cancel` verb + two-checkout process test (cancel one, the other keeps running; stale lease swept)
- [ ] Help: local-check.md cancel section; tl-worker help line + rules.tsv row
- [ ] Atlas: sdlc-binary local-check entry

## Log

### 2026-10-15

### 2026-10-17

- Claimed by ariadne:2; design above sent to the hub before code (a new mechanism: a run lease and a verb).
