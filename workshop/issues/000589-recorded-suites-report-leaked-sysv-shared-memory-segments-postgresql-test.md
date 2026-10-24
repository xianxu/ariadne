---
id: 000589
status: working
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-24
estimate_hours:
card_mirror: 'c4f647edebc1cf26fdf1a071b3aab01aa9793348' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T03:44:14-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# Recorded suites report leaked SysV shared-memory segments; PostgreSQL test harnesses must not leak them

## Problem

## Spec

**Leaked System V shared-memory segments fill the machine's limit and break PostgreSQL tests (echo:0, 10-23).** macOS
allows 32 SysV segments (`kern.sysv.shmmni`). On 10-23 31 were in use, every one with 0 attached processes and no
PostgreSQL running: leftovers from test PostgreSQL clusters killed without a clean shutdown (likely suites cancelled,
timed out, or killed by teardown). echo#26's `TestDaemonPostgresLeaseLossExits` then failed with
`shmget ... No space left on device`.

Fix, in two parts:
1. **The recorded suite notices leaks:** `sdlc local-check` records the slot owner's SysV segment count before and
   after a run, and a run that leaves new segments with no attached process reports them in its result (warning, with
   the ids), so the leak is pinned to the run that caused it.
2. **Guidance for harnesses that start PostgreSQL:** stop the cluster (or use a shared-memory-free configuration, such
   as `shared_memory_type=mmap` / `dynamic_shared_memory_type=posix`) in every exit path, including cancellation; the
   repos that start PostgreSQL (echo first) get their own issue.

## Done when

- A suite that leaks a segment gets a warning naming it in its recorded result (test with a fixture that creates and
  abandons a segment).
- The guidance is in the workbench docs, and echo has a filed issue for its harness.

## Plan

Durable design: [suite SysV warning plan](../plans/000589-suite-sysv-warning-plan.md). Await TL approval before code.

- [ ] Approve the warning, attribution, persistence and qualification boundaries.
- [ ] Implement pure observations/delta and stateful-fake managed lifecycle tests.
- [ ] Qualify native single-owned-segment behavior after operator-cleared capacity, document harness guidance and existing echo#32, then recorded pre-close/close/land.

## Log

### 2026-10-23

- 2026-10-23 (ops:0, from echo#32's investigation): PostgreSQL always allocates one 56-byte SysV interlock segment,
  even with `shared_memory_type=mmap`; its id is on `postmaster.pid` line 7. A clean or SIGQUIT stop removes it; a
  SIGKILLed postmaster leaks it (that was the 31 of 32). A leak report can name the culprit: `ipcs -m -p` gives each
  segment's creator PID (CPID), the postmaster, so record the process name and the run with the ids. echo#32 stops
  clusters with SIGQUIT first, SIGKILL only after 5 s, and tests that each cluster's own segment is gone.


### 2026-10-24 — claimed and design checkpoint

Ran tl-worker help, claimed and started #589 from main after #591 landed. No own check/merge is running; #564 stays held, #590 is next after this issue. Read the managed resource/batch cleanup, attempt publication and saved-pass proof seams. Read echo local rules (no MEMORY.md found) and inspected echo#32: owned by echo:2, codecomplete/SHIP, not landed; no duplicate issue or peer change.

Read-only native ipcs shows32 current owner segments,56 bytes each, NATTCH0; most CTIME displays00:43:21–00:43:47. No segment was created, attached, removed or signalled. All32 slots are occupied, so no native allocation test runs now. Ops:0 approved honest not-attributed wording for concurrent observations unless the creator is witnessed in the suite's own descendants/group (receiptc3fee6b9-4430-4614-9f8c-9e5bb9dc2b57), and directed fakes first, native qualification after operator cleanup.

The proposed plan uses two admitted-run boundary snapshots, at most one optional terminal group witness, and a diagnostic footer in the existing digest-verified log. No new attempt schema, cleanup mechanism or failure state. ARCH-PURE separates parsing/delta from host IO; ARCH-DRY shares warning rendering between fresh and saved-result publication. The native monitor must never infer that every same-UID allocation came from the running suite.

## Revisions

- 2026-10-24: Corrected the planning claim that a before/after count pins the culprit: parallel suites share the UID, and exited creators cannot be named reliably. Keep observed/unattributed warnings explicit; name known creator PIDs/process witnesses only. Existing echo#32 satisfies the requested harness issue; mmap does not remove PostgreSQL's56-byte interlock.
