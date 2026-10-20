---
id: 000481
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-20
estimate_hours: 3.03
collaboration-mode: collaboration
card_mirror: '901aae431b29df187fe687d32879d38a40db52ca' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T01:19:56-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# sdlc queue pause/resume (+ --cancel-running) and the tl-worker drain rule, for couch --drain

## Problem

A fleet drain before a machine or Couch restart (pair counterpart: `couch --drain [--fast]`) needs the machine test
queue and the landing lines to stop admitting work, and workers need a rule for the drain notice. Operator item 15
(10-19), discussion: ops `teams/devinfra/tl-status-xian-xu.side/2026-10-19-15-fleet-drain.md`.

## Spec

- `sdlc queue pause [--reason TEXT] [--cancel-running]` / `sdlc queue resume` (operator shell or a :0 hub only, the
  same rule as grants): while paused, the machine test queue and every landing line admit nothing new; waiters stay
  queued and print the pause and its reason. Running suites finish, or, with `--cancel-running` (used by
  `couch --drain --fast`), are cancelled and recorded as interrupted, rerunnable on the same head, never as failures.
- A merge interrupted by the pause resumes on `sdlc merge --yes` after resume, with no new suite needed when its
  head is unchanged and the suite record exists.
- tl-worker rule: on a drain notice, graceful means finish the current atomic step, push, write a Log line saying
  where you stopped, report "drained" to the hub and stay idle; fast means stop now, write the Log line, report.


## Contract (frozen by pair#722's approved plan, 10-19; copied verbatim from pair `workshop/plans/000722-fleet-drain-plan.md` at ca8135bb)

This supersedes the Spec's first bullet where they differ: pause and resume take `--owner TOKEN` (owner-scoped), and `status --json` follows the schema below. Build against it; conformance with pair#722's fake executable adapter is the acceptance.

### Notice text (v1, owned by pair#722's formatter)

The dynamic ID is a canonical UUID; render optional reason as quoted data on its own labelled line. No reason text becomes an argv or an instruction. The formatter is the single production text owner; skill/help describe the contract rather than duplicating the full template.

Graceful:

> Couch drain v1 {ID}: checkpoint and stop. Finish the current atomic step; commit and push the checkpoint; write where you stopped in the issue Log. Start nothing new. Run couch --drain --ack {ID} after the checkpoint is durable, then end your turn and stay idle. Do not send a separate drained report.

Fast:

> Couch drain v1 {ID} (fast): stop now and checkpoint. Do not finish the current step. Write where you stopped in the issue Log; commit and push the checkpoint without discarding unfinished working-tree edits. Start nothing new. Run couch --drain --ack {ID} after the checkpoint is durable, then end your turn and stay idle. Do not send a separate drained report.

These are supervisor-origin notices through Broker.SendSystem with authenticated receipt lookup. A delivery receipt is not an ack. Old or unsupported wrappers remain stuck with a relaunch-needed reason; drain does not relaunch them itself. Exactly one notice is admitted per drain/current incarnation; an uncertain transport result is inspected, not blindly pasted again.

### The sdlc seam (frozen)

Run argv arrays, never a shell. The queue adapter runs from the verified steward workspace and does not manufacture another slot's environment/authority. Owner token is `couch-drain:{ID}`; create and persist it before the first external pause attempt.

```
["sdlc", "queue", "pause", "--reason", "Couch drain {ID}: {reason}", "--owner", "couch-drain:{ID}"]
# fast appends one argument:
"--cancel-running"
["sdlc", "queue", "resume", "--owner", "couch-drain:{ID}"]
["sdlc", "queue", "status", "--json"]
```

Pause/resume need only conventional exit status; every successful or uncertain mutation is reconciled through status. Pause idempotently adds that owner; resume idempotently removes that owner and no other. Multiple owners compose: admission stays paused while any owner remains. Fast cancellation records interrupted work, not failed work; owner removal never restarts it. There is no TTL that silently resumes a live pause. The #481 worker rule uses the notice text and ack action above.

Exact required status schema, lower_snake_case JSON keys:

```json
{
  "schema_version": 1,
  "state": "paused",
  "complete": true,
  "pause_owners": [
    {"owner": "couch-drain:00000000-0000-4000-8000-000000000001", "reason": "Couch drain ...", "cancel_running": false}
  ],
  "active": [
    {"id": "opaque-stable-operation-id", "workspace": "pair:5", "kind": "suite", "state": "running", "head": "optional-git-sha"}
  ],
  "detail": ""
}
```

`state` is `running`, `paused` or `unknown`; `kind` is `suite`, `merge` or `landing`; active entry `state` is `running`, `interrupting` or `unknown`. All fields shown are required except `head`. `complete:false` or `state:unknown` means absence cannot be inferred. `detail` explains incomplete/unknown observations. `running` requires zero pause owners; `paused` requires at least one. Entries with unknown workspace use `workspace:""` and block global quiescence. Waiters held before admission are not active. An operation may own several distinct kinds, each represented; do not identify work by PID alone. A complete empty `active` list is authoritative quiescence only at the captured observation; Couch refreshes it while the board is active.

Decoder rejects duplicate keys, malformed required types, unsupported versions, invalid enum combinations and oversized/truncated output. Tolerate unknown additive fields read-only. Limits proposed: 1MiB output, 128 pause owners, 4096 active entries, bounded text fields; overflow becomes unknown/incomplete, never truncated success. Couch performs no PID killing or registry-file edits.

`FakeDrainQueue` is a stateful implementation of the same pause/resume/status interface: owner map, FIFO waiters and exact operation IDs with workspace/kind/head/state; controllable finish, interruption and observation failure. Pause blocks admission, graceful retains running work, fast marks interrupting then explicitly interrupted/removed, resume releases only one owner's hold. Lost replies can occur after an effect; status still sees the applied state. Include a fake executable adapter test that validates exact argv/JSON, in addition to in-process model tests. Live conformance runs in #481's isolated portable queue state root, never against the operator's machine queue during tests.

## Done when

- Tests: pause blocks admission in both the queue and the landing line; resume releases in FIFO order;
  `--cancel-running` records interrupted (not failed) and a same-head rerun is accepted.
- `tl-worker --help` carries the drain rule.
- The frozen owner-scoped argv/status schema conforms with pair#722's actual executable adapter and strict decoder using private portable queue/lock/fleet state, never the live machine queue.
- Status covers suite, merge and landing lifetimes; waiting operations are excluded, and legacy, unreadable or incomplete evidence cannot become authoritative quiescence.
- A graceful admitted merge encountering a paused new-suite boundary exits interrupted; existing successful exact-head suite evidence remains reusable after matching-owner resume.

## Plan

- [x] M1 — Pure owner/operation transitions, portable gate and lifetime leases, bounded legacy observation and frozen CLI protocol; targeted verification and milestone review.
- [x] M2 — Wire every suite/landing/merge admission, retain interruption evidence, ship worker drain rule and observed pair#722 conformance; recorded full suite, review and landing.

Durable plan: [queue drain implementation](../plans/000481-queue-drain-plan.md).

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.* Calibration source is marked stale by estimate-source; values are provisional.

Decomposition: pure queue model, portable lease store and bounded legacy observer are three greenfield single-concern modules. Existing flock/processgroup/x/sys/JSON adapters satisfy the library check: raw 1.0h design is halved for available integration support, then multiplied by 0.2 for the approved detailed plan (0.10h each); upper-table 0.8h implementation becomes 0.32h each at v3.1's 40%. CLI, Python queue adapter and interrupted-evidence integration extend existing smaller-module patterns (0.16/0.20/0.20h implementation). The cross-cutting Go admission integration, worker dispatcher, atlas documentation and one small peer executable conformance hand-off are separate primitives. Three review chunks count the two genuine milestone boundaries plus final issue review. Familiarity is 1.0 in the recently modified queue/runner code; design buffer is 15% for the accepted plan. No vendor-mode propagation surcharge, no public API discovery and no machine queue-wait budget is added.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=0.10 impl=0.32
item: greenfield-go-module design=0.10 impl=0.32
item: greenfield-go-module design=0.10 impl=0.32
item: smaller-go-module design=0.03 impl=0.16
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: cross-cutting-refactor design=0.12 impl=0.20
item: skill-or-dispatcher design=0.06 impl=0.12
item: atlas-docs design=0.02 impl=0.06
item: milestone-review design=0.00 impl=0.08
item: milestone-review design=0.00 impl=0.08
item: milestone-review design=0.00 impl=0.08
item: cross-repo-refactor-small design=0.04 impl=0.12
design-buffer: 0.15
total: 3.03
```

Design subtotal 0.67h × 1.15 + implementation subtotal 2.26h = 3.0305h, rounded to 3.03h. The estimate is derived after plan-quality round 2 passed; actual remains automatically measured at close.

## Log

### 2026-10-19

### 2026-10-20 — claim and design
- 2026-10-20: closed M2 — Managed interruption and private real adapter conformance passed; queue and real priority race checks passed 79.121s, core/line/lint/worker checks passed. BR-5 real-lock regressions prove secondary contention releases the global gate and retains FIFO; approved legacy-pending regressions and counted 8/16 status plus zero-scan admission controls pass. Final recorded suite queued on pushed code head.; review verdict: SHIP
- 2026-10-20: closed M1 — Core queuegate race tests pass (5.945s); CLI and legacy TestQueue race tests pass (1.934s); helptext and golint pass; diff check clean. Deterministic RED then GREEN covers intent-before-phase, dead interrupted owner, malformed intent identity, cancellation after mutex acquisition and applied-owner partial publication. Orphan intent needs matching interruption proof; strict codec controls preserve opaque additions. A 4096-operation registry scan completed in 374ms with exactly 4096 reads. No live queue or primary checkout touched; admission integration and executable conformance belong to M2.; review verdict: SHIP

Claimed in ariadne:3 and started from fresh main after #487 landed. Loaded current TL-worker/SDLC contracts. The frozen pair#722 contract is the authority; pair:5 confirmed the argv adapter, strict decoder, 5-second call ceiling and inherited-environment behavior. The portable state override will be supplied for isolated executable conformance.

Existing run interruption currently returns without saving an attempt, so this item must retain a distinct interrupted record without admitting it to failure/main/flake readers. Existing ordinary-worktree and already-merged cleanup paths bypass the landing line; they are included in the shared admission design. Full fleet inventory is unsuitable for 5-second status because it can fetch tracker state; legacy discovery will be bounded and local, reporting unknown when incomplete.

Hub clarified that graceful pause interrupts an admitted merge at a new-suite boundary, and legacy activity remains unknown until old runners finish. ARCH-ORDER/ARCH-DRY shaped one global admission mutex plus lifetime leases; ARCH-PURPOSE requires all merge paths; ARCH-EVOLVE keeps additive fields across rewrites. Durable plan committed before requesting implementation approval.

Plan-quality round 1 found a real direct-make ownership gap: Python's current watcher proves only make-PID lifetime; Go's managed process group owns descendant cancellation. Hub approved the honest-unknown boundary: managed sdlc suites cancel fully; direct make is stopped only with proved identity, never via a shared group, and stays unknown until reconciled. Do not build a new direct-make supervisor now; record it as a possible later follow-up if this visible stuck state causes an incident. Direct full suites should use recorded local-check; targeted tests use go test -run.

### 2026-10-20 — M1 implementation and targeted proof

Implemented the pure owner/operation state model, strict additive codecs, portable serialized gate/lifetime leases, owner-scoped CLI/status framing, bounded read-only legacy overlay and embedded help. Interrupted records are a separate family. Unmanaged owner death retains named unknown evidence; prior-boot reconciliation uses semantic OS boot identity, not raw struct bytes. Registry and legacy binding scans have counted 8/16 guards; CLI mutation/admission-side setup performs no legacy scan.

TDD observed missing feature/API RED, concrete malformed owner/UTF-8 and held-waiter binding RED, a duplicate retirement observation RED, unsafe filename hiding an unknown holder RED, and unpublished-enrollment residue RED before fixes. Final queuegate race run passed 1.395s; combined TestQueue CLI/legacy + help/catalog race checks passed 2.771s; helptext passed 0.133s; golint passed 0.457s; diff check clean. No full suite yet: the recorded final suite and live peer executable conformance belong to M2 after all admission consumers are wired. No test changed the live queue, :0 or another checkout.

### 2026-10-20 — M2 retained implementation checkpoint

M1 closed SHIP after round 2; all four findings disposed. M2 now wires normal/resume/ordinary merges and real weighted suite admission to the shared gate, plus managed cancellation and the worker drain rule. RED tests showed missing waiting leases, a landing binding rooted at the wrong directory, the watcher's shared mutex descriptor after fork, mismatched Python ID grammar, and an unverified workspace guess; corresponding fixes pass. Go combined race checks passed 22.306s; broader local-check compatibility passed 280.288s; existing priority tests (16), shell weighted/FIFO/watch tests and TL-worker/lead tests pass. New full-flow interruption controls prove retained logs, no SuiteAttempt/GitHub/baseline/flake poisoning and same-head retry. Counted 8/16 runtime admissions perform zero process/topology commands. Python private tests pass (12); real peer executable conformance and final recorded suite/review remain pending. Both bounded subagents hit model capacity; continuing their retained work in the main thread per hub direction.

### 2026-10-20 — isolated executable conformance and verification scope

pair:5 ran its production `sdlcDrainQueue` argv adapter and `decodeQueueSnapshot` against this slot's clean build (build revision b39a97b35386e2567c05862bc0b961f6e60bb8f5), passing in 15.128s. Verified owner composition/literal reason/replay/resume, worker refusal, held suite+merge+landing kinds/workspace/head, waiter exclusion, graceful/fast outcomes, interrupted-not-success retirement, and retained unknown/corrupt evidence across resume. Its probe/helper remain at `/Volumes/TestScratch/tmp/pair-slot5/pair722-realqueue.Fgg78v`; receipt 85ed0bd0-5fd1-4e0b-8b47-72526fd5fb75 is submitted and authenticated. That probe drives intent/retirement, not real descendants; our managed descendant test separately proves release of a real child's lifetime flock before return. Final targeted queue race run passed 17.590s, core/line races 4.190s/1.707s, and worker/lead/helptext/golint pass. A counted 8/16-field Python raw-member test reproduced suffix copying (1,872 copied bytes for 120 input bytes at 8 fields); a linear whitespace scan now passes all 13 private helper tests. Changes after the conformance build affect that Python scan and test coverage; Go production code is unchanged.

Hub approved the normal central sdlc for the recorded suite and merge with only `WF_SDLC_QUEUE_ROOT=/Volumes/TestScratch/tmp/ariadne-slot3/ariadne-481-verification-queue.tHUG4N` set and `WF_TEST_LOCK_FILE` unset. The real machine capacity, FIFO and grants stay in force. This avoids leaving unmanaged unknown residue in production when the pre-rollout central binary invokes the new helper. The temporary launcher is `/tmp/ariadne-481-run-central.py`, which execs `/Users/xianxu/workspace/ariadne/bin/sdlc` with exactly that environment change. Clean the private queue root and this slot's conformance fixture after landing; the peer's sibling unknown fixture belongs to pair:5. Hub rebuilds central sdlc after landing to activate managed leases fleet-wide.

### 2026-10-20 — M2 review BR-5 correction

M2 round 1 returned REWORK for blocking secondary mutexes inside the global gate. Real Go and Python contention fixtures reproduced the defect before the fix: pause expired against the held line lock, and Python admission could not return until allocation was released. Both consumers now try the nested lock without blocking, release the global gate on contention and retain their FIFO ticket. Go also proves cancellation while allocation stays held and readmission of the same ticket after resume; Python proves cancellation of its private waiter. Core/line races pass 3.743s/1.585s; combined TestQueue and real-provider priority race checks pass 73.050s; 15 private Python controls pass; golint passes. Cancelled the obsolete recorded suite with `sdlc local-check cancel --name suite` before it started; it exited interrupted, with no suite attempt. The final-code suite and M2 review will be requeued.

Additional pair:5 lost-reply conformance passed 4.863s: successful pause/resume effects survived proxy exit 29, and status reconciled them without removing the second owner. Receipt 77228edf-6f1c-4545-9444-5cc060ab3d8a was authenticated/submitted; primary private fixture ended running/complete/empty.

### 2026-10-20 — legacy pending-admission status correction

Verified the hub's approved boundary in canonical submitted receipt 278d62f1-1ea2-4cb2-bd64-b29f7cb5fa91. Both legacy token/ticket regressions failed before implementation: status had no activity and no incompleteness detail. The status-only probe now reports named pending uncertainty without active entries. Verified new waiting bindings remain complete; recycled identity does not. Unlocked residue/allocation mutex controls and unreadable-directory control pass. Queue legacy and both counted zero-scan admission controls pass under the race detector (1.770s); 8/16 pending fixtures require exactly one index/visit/lookup per waiter. This preserves the frozen schema and adds no admission scan.

### 2026-10-20 — M2 SHIP and final consumer conformance

M2 round 2 SHIP, no new findings; BR-5 independently verified by reverting both fixes in disposable copies and observing their regressions fail. The reviewer ran queuegate/landingline and TestQueue normally and under race, 15 Python pause controls, 16 real-provider priority controls, saved-pass/repeated-run compatibility and worker/help checks. Final branch queue/priority race run passed 79.121s; core/line/lint/worker checks passed.

Authenticated submitted pair:5 receipt e3701d33-6b0b-4b49-b962-ef6e33381e43 confirms repeat production adapter/strict decoder conformance against afb94a8e2be9d2477fbbe7b53d7b8eb7f52152c7, independently verified modified=false: core 19.66s, lost-response 4.06s, combined 24.049s. Real consumer legacy check passed 3.814s: unbound held token/ticket produced complete=false, named path/detail and active=[], unchanged bytes; release restored complete. Verified waiting bindings stayed complete, pause denied admission, resume admitted exactly one suite and retirement restored complete/empty. An initial probe cleanup ordering failure was retained and fixed only in the temporary probe before the passing rerun; no product change. Source/logs remain under pair:5's original private fixture consumer-repeat.2TGKnT; its diagnostic unknown sibling remains peer-owned. No live pause. The final recorded full suite is queued; server CI is green.

## Revisions

### 2026-10-20 — M1 review corrections

Round 1 requested cancellation crash recovery, in-transaction deadlines, actual-symbol traceability and one shared strict JSON boundary. Tests reproduced ignored durable intents, lost interruption diagnostics and scans continuing after cancellation before fixes. Owners/status now consume exact-ID intents; immutable diagnostic envelopes coexist with unresolved descendant evidence; orphan cleanup requires proof. Unknown additive fields remain preserved. Plan Revisions record recovery and budget semantics. Core race checks passed 5.945s, CLI/legacy race checks passed 1.934s, helptext and golint passed. Maximum-capacity pause read all 4,096 records once in 374 ms; deterministic cancellation tests guard partial publication. M1 review will rerun; M2 and the recorded full suite remain pending.

### 2026-10-20 — make frozen contract and clarified acceptance explicit

The original Spec and Done-when remain above for context. Added explicit frozen-schema conformance, complete lifetime coverage/legacy uncertainty, and graceful merge interruption with successful-head reuse, reflecting the frozen Contract and hub clarifications. No change to pair#722's external protocol. Implementation is full flow with two actual review boundaries: backend/CLI, then admission/interruption/conformance.

### 2026-10-20 — full plan approved

Hub approved both boundaries and requested a counted zero-process/zero-topology-scan guard on admission. Added the named test to the durable plan; legacy observation remains status-only. Proceeding to change-code's plan-quality gate, with no implementation edits yet.

### 2026-10-20 — PQ-1/PQ-2 approved refinement

Kept the frozen external protocol unchanged. Added an explicit unmanaged direct-make cancellation/unknown boundary and retained the supervisor as a later follow-up only if needed, per hub approval. Compressed test inventories to production-function adversarial strategies. Managed suite cancellation remains complete and interruptions remain outside SuiteAttempt/flake classification.
