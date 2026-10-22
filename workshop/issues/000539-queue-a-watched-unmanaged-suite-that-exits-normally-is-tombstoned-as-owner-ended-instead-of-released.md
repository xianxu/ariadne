---
id: 000539
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c243a82174a1ca4fb0a675412247430509077b4b' # card fields mirrored from issue-cards; edit via sdlc
---

# queue: a watched unmanaged suite that exits normally is tombstoned as 'owner ended' instead of released

## Problem

## Spec

The machine queue tombstoned a direct (unmanaged) suite that **finished normally** as "watched owner ended;
descendants require operator reconciliation", so it held suite capacity until the :0 steward reconciled it.
Evidence (echo:1, 10-21 PDT): op f70ce4520db34fdaa8208eb03bf778c0, watched pid 88676 started 18:31:04.407, queue
record 18:31:04.542; the command finished 19:24:43.770 with exit 0 after acquiring at 3214 s, both PASS lines;
session kept open throughout. A Couch restart happened at 18:59:22 (possibly related; unverified). Reconciled
19:50:08 (reconcile.jsonl line 3). An earlier echo:1 op 65dbdc31 (started 16:56) had the same shape.

Fix: a normal finish of a watched unmanaged run releases its record without a reconcile; record the pid's exit time
(and status when knowable) so a steward can tell a normal end from an orphan. Find why the watch saw "owner
ended, phase unknown" for a clean exit (e.g. the watcher losing its handle across a Couch restart).

## Done when

- A watched unmanaged suite that exits 0 releases its queue record, with no tombstone (test with a real child process).
- Same across a simulated watcher restart mid-run (test), or the tombstone records the observed exit time and is
  auto-released when the pid's exit is confirmed.

## Plan

- [ ]

## Log

### 2026-10-21
