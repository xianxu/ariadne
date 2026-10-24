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
flow: {kind: quick, provenance: inferred, spec: "0a3b6838", done: "bf85bccb"}
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

Durable design: [suite SysV warning plan](../plans/000589-suite-sysv-warning-plan.md). Approved by ops:0 on 2026-10-24; implementation gate completed.

- [x] Approve the warning, attribution, persistence and qualification boundaries.
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

### 2026-10-24 — resumed implementation

Retained change-code completed successfully and committed accepted design (Git state reports 9078882e). Previous recovery screen is unavailable; recovery used the issue, durable plan, Git state and retained process exit. Full plan approval receipt: 38e93c9b-c77c-42ff-9f67-fedbfc22b3af. No native IPC allocation or cleanup performed. Independently filed hub-requested staging-attribution follow-up #594 unclaimed; it does not change this scope.

### 2026-10-24 — first implementation checkpoint

Observed red then green for pure delta/native table parsing and bounded-reader fake controls. The managed fake abandoned-segment regression is green: exactly two snapshots and one terminal witness hook, warning in stderr and digest-verified saved log, successful status unchanged. Parent final summary overrides earlier child-looking summaries. Production local-check, resume, single merge, train and baseline constructors explicitly inject native observation; unconfigured unit-test dependencies do not probe live machine state. Native creator names/group/incarnations are read from one kernel record after the bounded terminal PID/group snapshot; exited or moved rows remain unknown. No real IPC allocated or removed. Native qualification request sent to hub, receipt 1f398835-ab92-43f2-a13c-9893367f0ce9 submitted. More adversarial lifecycle, saved-replay, limits, docs and native qualification remain before full suite/close.
