---
id: 000589
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
card_mirror: '4e37c856c63444a5896a9e698f4c5a51a17f7c5f' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-23

## Log

- 2026-10-23 (ops:0, from echo#32's investigation): PostgreSQL always allocates one 56-byte SysV interlock segment,
  even with `shared_memory_type=mmap`; its id is on `postmaster.pid` line 7. A clean or SIGQUIT stop removes it; a
  SIGKILLed postmaster leaks it (that was the 31 of 32). A leak report can name the culprit: `ipcs -m -p` gives each
  segment's creator PID (CPID), the postmaster, so record the process name and the run with the ids. echo#32 stops
  clusters with SIGQUIT first, SIGKILL only after 5 s, and tests that each cluster's own segment is gone.
