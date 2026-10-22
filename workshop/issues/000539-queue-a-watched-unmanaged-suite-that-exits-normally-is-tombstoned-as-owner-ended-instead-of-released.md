---
id: 000539
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '8a8cd38bd31cdd6cc5a9391e1bb99988c700a0fa' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T21:43:50-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ] Implement and verify [the completion-evidence plan](../plans/000539-direct-suite-completion-plan.md), after plan approval and change-code.

## Log

- 10-21 21:2x (ops:0): third occurrence, no Couch restart involved this time: echo:3 op b54088501a804c8eab8d7bab2de27e79,
  watched pid 97474, started 21:03:05 (04:03:05Z), an in-slot `make test` that finished with exit 0 at about 21:11 per
  ops-echo:0, session open. Reconciled ~21:25. So the restart isn't the trigger; a normal finish of a watched
  unmanaged run never releases.

### 2026-10-21

- Claimed and started planning in ariadne:5. Root cause: `scripts/test-lock.py:QueueGate.retire` unconditionally creates an unresolved tombstone for every admitted unmanaged operation. Its `owner_ended` argument is unused. `watch_queue_owner` observes PID existence only, never an exit status or completion acknowledgment. A Couch restart is unnecessary; the third incident follows directly from this code.
- Distinguished physical capacity from queue uncertainty: watcher exit closes the actual suite slot flocks; the unresolved record retains a logical unknown lifetime and blocks a clean drain. Do not describe the tombstone alone as proof that the physical slot remains locked.
- Read the three original records from the machine queue's reconciliation audit, without changing live state. SHA256 digests: `65dbdc31d6444c45a8692456d1e0a595` = `1ca996749fca4c752f74e1802b465ccbc11ff79567a04cbbc5f96c32782250c6`; `f70ce4520db34fdaa8208eb03bf778c0` = `7dbdcd3470524c1e2150fc99d31898c0ce597c64ad81ca26b44593fda40f82b7`; `b54088501a804c8eab8d7bab2de27e79` = `e61db6374cc1a79edcafa53eb1b6d8bd127afaf1b28f5669c01c656e25f0923a`. Exact audit reconciliation times are respectively 2026-10-22T01:25:41.720585Z, 02:50:08.407156Z and 04:27:31.974581Z. These are reconciliation times, not PID exit times. No historical record contains an exit timestamp/status. The reported second finish remains 19:24:43.770 PDT; the third is approximate, about 21:11 PDT.
- TL decision (authenticated Couch receipt `298daa02-f072-4415-a649-99a5c8cc6d5b`): take the narrow fix. Exact make incarnation's positively observed exit 0 plus an empty process group permits normal retirement. A surviving group member, nonzero/unobserved exit, cancellation or watcher crash retains uncertainty. Record `exit_observed_at` and survivor PIDs; preserve all three real tombstones as fixtures. The managed `sdlc local-check` path remains the whole-command owner.
- Private native feasibility probe: a sibling observer registered Darwin kqueue NOTE_EXIT plus NOTE_EXITSTATUS before releasing the target's input barrier. It received the exact target PID and wait status 0 despite not being its parent. Initial probe accidentally retained the target's input writer in the observer; closing that inherited writer fixed the experiment. No source code or tests changed. kqueue registration must happen after fork and be acknowledged before acquire returns; an inherited kqueue is insufficient.

## Revisions

- 2026-10-21: TL narrowed the completion contract to observed exit 0 plus verified empty, private make process group. The original restart alternative must not turn PID disappearance into completion authority: watcher loss keeps an unknown tombstone; restarting Couch while the independent watcher survives must not change completion. Unsupported exit-status observation and shared/unknown process groups remain conservative. The implementation plan carries these controls explicitly (ARCH-ORDER, ARCH-SECURE).

- 2026-10-21: Concrete plan approved by TL receipt `c86c6fec-3912-4277-9ea1-f519c2aa7f25`, with the diagnostic receipt-directory proposal cut in favor of one compact audit line beside reconciliation. Actual initial plan commit: `35fb28672` (corrects the accidental placeholder in the outgoing checkpoint message, already corrected through Couch). Enter change-code next; implementation remains unchanged.
