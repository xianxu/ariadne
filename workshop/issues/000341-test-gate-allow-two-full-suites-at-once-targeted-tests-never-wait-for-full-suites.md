---
id: 000341
status: codecomplete
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-12
estimate_hours:
card_mirror: '17e0af42215719e759820bed8982a68064eafb91' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T16:13:39-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "9096fb8a", done: "021e0f35"}
actual_hours: 0.05
---

# test gate: allow two full suites at once; targeted tests never wait for full suites

## Problem

10-12: the machine-wide test gate serializes everything. One full suite at a time (scripts/test-lock.py, flock on ~/.local/state/ariadne/test-suite.lock, shared by every repo), and targeted test runs (pkg/testgate, test-gate.<k>.lock slots) also wait while any full suite runs. With 4–5 slots each needing a full suite to close, work went single-file (5–10 min per suite; slots waited half an hour) while the machine sat at load ~2. The strict rule was right at load ~200 this morning; with slots' spawned work moving to nice 10 (pair#502) and agent churn on a no-journal scratch volume (pair#496) it wastes the machine.

## Spec

Decided by the operator, 10-12:
1. **Full suites: up to 2 concurrently** machine-wide (configurable, e.g. `WF_TEST_SUITES=2`), still flock-based so a dead holder releases automatically; the waiting message names both holders.
2. **Targeted tests never wait for full suites:** they keep only their own small concurrency cap (the existing test-gate slots, configurable), independent of the suite lock.
3. The atlas states the rule (what takes which lock, the defaults, the env vars), and the agent-facing docs say targeted runs are always fine.
Keep `WF_TEST_LOCK=off` as the logged emergency hatch.

## Done when

- Two full suites run at once and a third waits; a targeted test run starts immediately while two suites run; tests for both (with fakes for the lock files, no real load generated).
- The atlas updated.

## Plan

Design (ARCH-DRY: reuse the one flock mechanism, N files instead of one;
ARCH-CONSTRAINTS: the cap is an operator choice, default 2, measured basis in ## Problem;
ARCH-FUNERAL: the slot files are fixed, N of them, rewritten per run, removed by nothing,
like today's one file; ARCH-ORDER: flock is the state, released by the kernel on exit).

- [x] `scripts/test-lock.py`: `WF_TEST_SUITES` slots (default 2). Slot 0 stays
  `test-suite.lock` (an older test-lock.py in a not-yet-refreshed checkout contends
  for it, so the cap holds across versions); slot k is `test-suite.lock.<k>`.
  One scan per tick: try each slot non-blocking, keep the first free one, read
  the records of the busy ones. A busy slot held by this make or `WF_TEST_LOCK_HELD_BY`
  means re-entrant (drop any free slot taken). None free: wait, naming every holder.
  The record gains `slot`.
- [x] `pkg/testgate`: drop the full-suite probe; a targeted run takes only its own
  `test-gate.<k>.lock` slot. Package doc and messages follow.
- [x] Tests, fakes only (private lock files, `sleep` holders):
  `scripts/test-lock.test.sh` two suites acquire, a third waits naming both and takes
  the freed slot; re-entrancy with the holder on slot 1; `WF_TEST_SUITES=1` keeps the
  single-suite cases. `pkg/testgate`: both suite slots held exclusively, `enter`
  takes a gate slot without waiting.
- [x] Atlas (`atlas/workflow/sdlc-binary.md`), `Makefile.workflow` comment, and
  AGENTS.base.md: targeted runs are always fine; at most 2 full suites.

## Log

### 2026-10-12
- 2026-10-12: closed — scripts/test-lock.test.sh 23/23 (two suites acquire, third waits naming both and takes the freed slot, re-entrant with a slot free and on slot 1, bad WF_TEST_SUITES); go test ./pkg/testgate pass incl. TestEnterIgnoresFullSuites (both suite files held exclusively, any wait fails). make test: all shards green except TestVerbContextsReachTrackerReads, pre-existing red on main from #302, filed as #340. Live: my make test took slot 1 while a pair suite under the old script held slot 0, both running.; review verdict: SHIP
- Implemented. `scripts/test-lock.test.sh`: 23/23 pass, including two suites acquiring,
  a third waiting with both holders named and taking the freed slot, re-entrancy with
  a slot free and on slot 1. `go test ./pkg/testgate`: pass; the new
  `TestEnterIgnoresFullSuites` holds both suite files exclusively and fails on any wait,
  so restoring the old suite probe fails it. Holders are `sleep` processes; no load.
