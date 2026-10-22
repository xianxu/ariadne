---
id: 000539
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours: 2.13
collaboration-mode: collaboration
card_mirror: 'f31bcb7f607c9da5036ff0e3e7bd70b8b69331c3' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T21:43:50-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
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

- On Darwin, a watched direct suite's exact make incarnation exits with positively observed status 0 and its private process group is confirmed empty: the queue operation retires with no unresolved tombstone, and a compact audit records its exit observation time. Prove with real make and real child processes.
- Survivors, nonzero or unobserved exit, cancellation, failed/shared group observation, or watcher loss retain uncertainty with available exit timestamp/status and survivor PIDs. Historical records without exit evidence remain unresolved.
- A simulated supervisor exit while the independent watcher survives does not prevent normal completion. Losing the watcher itself never authorizes retirement from PID disappearance; fresh-reader crash recovery distinguishes committed finished phase from an unproved running lifetime.

## Plan

- [x] Implement and verify [the completion-evidence plan](../plans/000539-direct-suite-completion-plan.md), after plan approval and change-code.

## Log

- 10-21 21:2x (ops:0): third occurrence, no Couch restart involved this time: echo:3 op b54088501a804c8eab8d7bab2de27e79,
  watched pid 97474, started 21:03:05 (04:03:05Z), an in-slot `make test` that finished with exit 0 at about 21:11 per
  ops-echo:0, session open. Reconciled ~21:25. So the restart isn't the trigger; a normal finish of a watched
  unmanaged run never releases.

### 2026-10-21
- 2026-10-21: closed — Recorded suite on ce987fd9: all 8 shards, 2079 SDLC tests and 75 other packages passed, zero failures (771.7s); real-make red-green, three original captures, exit/survivor/cancel/crash/handshake/audit controls, old-reader evidence preservation and current-provider priority harness pass. Atlas updated; no live mutation or stage required by the approved private-process proof.; review verdict: SHIP

- Claimed and started planning in ariadne:5. Root cause: `scripts/test-lock.py:QueueGate.retire` unconditionally creates an unresolved tombstone for every admitted unmanaged operation. Its `owner_ended` argument is unused. `watch_queue_owner` observes PID existence only, never an exit status or completion acknowledgment. A Couch restart is unnecessary; the third incident follows directly from this code.
- Distinguished physical capacity from queue uncertainty: watcher exit closes the actual suite slot flocks; the unresolved record retains a logical unknown lifetime and blocks a clean drain. Do not describe the tombstone alone as proof that the physical slot remains locked.
- Read the three original records from the machine queue's reconciliation audit, without changing live state. SHA256 digests: `65dbdc31d6444c45a8692456d1e0a595` = `1ca996749fca4c752f74e1802b465ccbc11ff79567a04cbbc5f96c32782250c6`; `f70ce4520db34fdaa8208eb03bf778c0` = `7dbdcd3470524c1e2150fc99d31898c0ce597c64ad81ca26b44593fda40f82b7`; `b54088501a804c8eab8d7bab2de27e79` = `e61db6374cc1a79edcafa53eb1b6d8bd127afaf1b28f5669c01c656e25f0923a`. Exact audit reconciliation times are respectively 2026-10-22T01:25:41.720585Z, 02:50:08.407156Z and 04:27:31.974581Z. These are reconciliation times, not PID exit times. No historical record contains an exit timestamp/status. The reported second finish remains 19:24:43.770 PDT; the third is approximate, about 21:11 PDT.
- TL decision (authenticated Couch receipt `298daa02-f072-4415-a649-99a5c8cc6d5b`): take the narrow fix. Exact make incarnation's positively observed exit 0 plus an empty process group permits normal retirement. A surviving group member, nonzero/unobserved exit, cancellation or watcher crash retains uncertainty. Record `exit_observed_at` and survivor PIDs; preserve all three real tombstones as fixtures. The managed `sdlc local-check` path remains the whole-command owner.
- Private native feasibility probe: a sibling observer registered Darwin kqueue NOTE_EXIT plus NOTE_EXITSTATUS before releasing the target's input barrier. It received the exact target PID and wait status 0 despite not being its parent. Initial probe accidentally retained the target's input writer in the observer; closing that inherited writer fixed the experiment. No source code or tests changed. kqueue registration must happen after fork and be acknowledged before acquire returns; an inherited kqueue is insufficient.

## Revisions

- 2026-10-21: TL narrowed the completion contract to observed exit 0 plus verified empty, private make process group. The original restart alternative must not turn PID disappearance into completion authority: watcher loss keeps an unknown tombstone; restarting Couch while the independent watcher survives must not change completion. Unsupported exit-status observation and shared/unknown process groups remain conservative. The implementation plan carries these controls explicitly (ARCH-ORDER, ARCH-SECURE).

- 2026-10-21: Concrete plan approved by TL receipt `c86c6fec-3912-4277-9ea1-f519c2aa7f25`, with the diagnostic receipt-directory proposal cut in favor of one compact audit line beside reconciliation. Actual initial plan commit: `35fb28672` (corrects the accidental placeholder in the outgoing checkpoint message, already corrected through Couch). Enter change-code next; implementation remains unchanged.

### Implementation checkpoint

- Plan-quality CLEAN after PQ-1, estimate recorded after acceptance; change-code full flow passed. Implemented initial observer/retirement path after red evidence. Pure finite-product decision and malformed-boundary tests went red first; a subsequent wrong-incarnation and noncanonical-field regression also went red before their guards.
- Real red: `python3 scripts/test/test-lock-exit.test.py -v`, clean make PID 98439 returned 0 and physical slot became free, but unresolved operation `4ec001462e77407b9c488b8836088f9c` remained unknown (1 test, 0.508 s, expected failure). Copied the three original audited JSON payloads unchanged, with hash-checked provenance.
- Initial green: ten private real-process tests, including background survivor, signal/nonzero, cancellation, watcher crash, simulated supervisor exit and five rapid completions; eight decision/snapshot/audit tests; all 18 existing pause/helper controls and the full shell helper regression pass. Fresh Go observer/codec controls preserve additive evidence and never promote historical tombstones.
- Standalone priority Python tests selected the checkout's stale Oct 15 bin/sdlc and failed with provider-unavailable/FIFO diagnostics. Rerun through `TestSuiteQueuePriorityRealHelper`, whose harness supplies the current production provider; do not classify the stale-provider run as feature evidence or change production to accommodate it.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only; source tagged stale by estimate-source, so hours are provisional. Python/make tooling maps to the documented smaller-module primitive. Native observer integration uses existing Python kqueue and libproc (halve API design for library availability), with the thorough approved spec's 0.2 design factor and 15% design buffer. Implementation values are 40% of v2 ranges; familiarity 1.5 reflects bounded novel native exit observation. Includes the real-process discovery, actual make fixtures, one close review, and atlas work; no peer adoption or deployment.

```estimate
model: estimate-logic-v3.1
familiarity: 1.5
item: issue-spec design=0.10 impl=0.04
item: api-integration design=0.10 impl=0.24
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: atlas-docs design=0.02 impl=0.04
item: milestone-review design=0.02 impl=0.20
item: real-api-discovery design=0.00 impl=0.24
design-buffer: 0.15
total: 2.13
```

- 2026-10-21: Updated Done when to the authenticated TL-approved narrow contract (receipts `298daa02-f072-4415-a649-99a5c8cc6d5b` and `c86c6fec-3912-4277-9ea1-f519c2aa7f25`). This explicitly replaces the original unsafe watcher-restart/PID-exit auto-release alternative; supervisor restart with a surviving independent watcher is the real-process control.

- Final targeted checkpoint: `go test ./cmd/sdlc -run '^(TestQueueRealHelper|TestSuiteQueuePriorityRealHelper)' -count=1 -timeout 10m` PASS (33.678 s), `go test ./cmd/sdlc/internal/queuegate -count=1` PASS (1.264 s); Python decision/observation/audit controls 11/11 PASS (0.129 s). The current provider harness resolves the standalone stale-bin failure. Full shell helper script previously PASS. Readiness failure/timeout and production terminal-publication crash cuts are now covered; no live queue mutation or :0 stage performed.

- Durable wait checkpoint: PR https://github.com/xianxu/ariadne/pull/384, exact tested/pushed code head `f81beb07b`; recorded suite retained as `lease-2723494484` (local terminal session 54063), queue position 6 at admission. Detached check tree `/Volumes/TestScratch/local-checks/sdlc-local-1638512170`; immutable inflight log `/Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/inflight/e867597985e7f3e4e0210c63ede06a4a/run.log`. Retain this runner; do not duplicate or reorder. No :0 stage, live queue repair or private cache. Remaining: recorded suite, mandatory close review, merge. Code is committed; later bookkeeping commits are docs-only.

- Audit failure-path correction: inspection found the initial implementation changed confirmed `group_observation=empty` into unknown after an audit write failure. A new assertion reproduced that loss red, then the handler was changed to preserve all observed facts and name `completion audit failed` in detail. Pure tests 11/11 PASS (0.148 s); current-provider helper integration PASS (38.935 s). Cancelled only our queued old-head `lease-2723494484` via local-check cancel before starting any replacement; it never executed a suite and recorded no test result. Terminal 54063 finished interrupted. New recorded run will target the corrected head; do not resume the cancelled lease.

- Replacement durable wait checkpoint: PR #384 now pushed at corrected code head `ce987fd9ca68`. Single retained recorded suite is `lease-1229221159` (terminal 77638), queue position 5 initially. Worktree `/Volumes/TestScratch/local-checks/sdlc-local-4137269143`, log `/Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/inflight/43ec553f60c0828c46d74f65896002b8/run.log`. The old lease is cancelled, not retained; no duplicate suite exists. Continue this run, then close review and merge.

- Recorded full suite PASS on `ce987fd9ca68ebe377c8d80e5750e5e37122d0f7`: all 8 shards, 2079 cmd/sdlc tests, 75 other packages, zero failures, wall 771.7 s after 2742 s queue wait. Status `sdlc-local/suite`; log `/Users/xianxu/workspace/ariadne/.git/sdlc/local-checks/ce987fd9ca68ebe377c8d80e5750e5e37122d0f7/suite.log`, SHA256 prefix `1022973b3872`. Server merge-check and local-gates-guard passed; no protected gate paths or operator label required. Close review and merge remain.
