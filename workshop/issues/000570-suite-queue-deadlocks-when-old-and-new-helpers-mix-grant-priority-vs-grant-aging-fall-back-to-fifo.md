---
id: 000570
status: working
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '41a6b9aa2719ce1cf06a853a71fc35f81a47160e' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T02:36:18-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "c7a2b022", done: "4bdabeb6"}
---

# Suite queue deadlocks when old and new helpers mix (grant priority vs grant aging); fall back to FIFO

## Problem

## Spec

**Suite-queue deadlock from mixed helper versions after ariadne#560 (grant aging), observed 10-23 ~01:30.** No suite
started for minutes with both machine slots free. Queue status (ariadne:0's sdlc) showed protected FIFO tickets 32–36
(all "age unknown", from older helpers) ahead of granted ticket 37 (pair#832, "aging overrode grant priority"). The
granted runner (new helper) waited for ticket 32's holder; ticket 33's runner (an older helper, from its checkout's
pinned ariadne) said the next was #832. Each side deferred to the other: older helpers apply plain grant priority, the
newer one applies aging. Revoking the grant restored a single order and suites started at once.

Root cause: admission order is computed by each waiter's own `test-lock.py`, and ariadne#563 deliberately keeps the
checkout-supplied helper, so during a rollout, helpers with different ordering rules share one lock root.

Fix options to weigh in the plan: (a) record the ordering-policy version in each ticket and have a newer helper defer
to the oldest policy present (or refuse grants) while mixed; (b) a single shared ordering authority (the newest
helper on the machine) that every waiter consults; (c) a guard: `sdlc landing grant` refuses (or warns) while any
waiting ticket comes from a helper older than the grant-aging policy. Must never deadlock: when waiters disagree,
fall back to plain FIFO.

## Done when

- A mixed-version queue (old-policy waiters plus a new-policy granted waiter) always admits someone (test with both
  helper versions against one lock root).
- Status names a mixed-policy queue explicitly.

The original universal mixed-version progress bullet above is superseded by the
approved guarded-rollout contract in Revisions. Current acceptance: new grants
refuse legacy/unknown held waiters or unavailable observation; status names an
existing granted-waiter/legacy coexistence risk with explicit revoke recovery;
compatible queues retain FIFO and aging. Old helpers and later joins remain a
manual rollout responsibility, without automatic cancellation or policy changes.

## Plan

- [x] Claim and start-plan from fresh main; reproduce both historical and current admission against one private native lock root.
- [x] Trace the grant provider/cache and distinguish a new-helper decision from agreement by already-running old code.
- [ ] Resolve the rollout choice with the hub before implementation (proposal below).
- [ ] After agreement, write final function-level adversarial plan, enter change-code, observe regression RED, implement, and run isolated mixed-version conformance plus normal recorded gates.

### Proposed rollout boundary — awaits hub

A FIFO change in only the new helper cannot make existing old helpers use FIFO. The unchanged old helper ignores future ticket policy fields and still sees the active grant; its cache is keyed to grants.jsonl identity/mtime/size and grant temporal boundaries, not queue membership. Status and admission share the new pure projection, but that does not make old admission agree.

Options:

1. Ticket policy negotiation alone: useful for future versions that already implement negotiation, insufficient for unmodified historical waiters. Even selecting the oldest policy in the newest helper leaves pre-fix aging helpers applying their own order in a three-version queue.
2. Central shared helper authority: future consumers could agree, but this changes #563's deliberately checkout-supplied helper ownership and does not replace code already executing. It is a larger rollout change.
3. Recommend a guarded rollout: refuse new grants while any live waiting helper cannot prove the agreed ordering contract (status explicitly names mixed/unknown policy); the TL explicitly revokes any existing conflicting grants and allows old waiters to drain unchanged. Do not automatically revoke, kill, unlock, rewrite another waiter's identity or alter supplier pins. Once all live waiters support negotiation, mixed supported policy versions use FIFO. This requires the hub to accept that historical active-grant mixes need an explicit rollout operation; no new code can retroactively change their decisions.

Proposed invariant authority: use the existing allocation mutex plus each waiter's lifetime ticket flock to read exact live membership and additive ordering-policy metadata. No process scan or separate daemon. A decision made only at grant creation can race a legacy waiter joining later, so a creation-only guard cannot justify a permanent never-deadlock claim; approval must settle whether incompatible later joins disable priority via a shared supported contract, or grants remain operationally off until helper rollout is complete. Do not implement a nominal FIFO fallback while leaving this contradiction unresolved.

Actions: wait for drain; request a grant after compatible rollout; TL revokes a conflicting existing grant.

| State / failure | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Proven compatible waiting helpers | Normal negotiated queue order | Continue waiting | None |
| Legacy/mixed/unknown helper with requested new grant | Priority unavailable, named incompatible ticket | Wait for drain/rollout, then request grant | Priority only, never queued work |
| Existing conflicting grant plus immutable old waiter | Mixed-policy queue; disagreement possible | TL revokes the conflicting grant and lets waiters drain | Priority only; no work cancellation |
| Ticket observation unavailable | Compatibility unknown; no new grant authorized | Retry after named observation fault resolves | No false agreement |

ARCH-DRY/PURPOSE: one shared projection for updated admission/status is necessary but not authority over old code. Preserve old-reader behavior explicitly. Expected scans are O(N) live tickets per observation and O(N) policy classification; no per-ticket subprocess/provider calls. Count N/2N live-record visits and provider invocations. Any final guard must share the existing mutex rather than infer liveness from stale filenames, and must state its joining-waiter race boundary.

## Log

### 2026-10-23

### 2026-10-23 — claim, root cause and isolated native reproduction

Authenticated dispatch42339e6f-6456-4657-82ed-5c2b018c1a0a assigns claim and plan, not implementation. tl-worker rules loaded; start-plan created this issue branch from main. No new helper/Go implementation or live queue/grant mutation.

Loaded the exact pre-aging helper from Git32e6177751fb5936be76207b1888372325536207 and current source independently. On identical ticket identities32 and37 with a grant for issue832, the old pure order is37,32 while current order is32,37. A private native-flock fixture then ran both actual Tickets.admission implementations against one fresh lock root with two free slots: old_taken=[], old_head=false; aging_taken=[], aging_head=false. Removing grant evidence admitted the old waiter into both slots, then the newer waiter after the old ticket retired. No watched processes, live queue root or production grant were driven. Evidence retained in common sdlc/test-stats/evidence/000570/mixed-helper-diagnosis.json. The old source is immutable Git evidence, not fresh instructions.

Tickets.elect is the decision boundary. GrantEvidence caches per grant-file signature and temporal refresh, so a provider-only membership change would not immediately invalidate all old helper evidence. A new additive ordering-policy field alone is not sufficient: old admission never reads it. This is a rollout/authority decision, not a simple missing conditional. Send the concrete alternatives and recommended guarded rollout to the hub before code; await a clarified compatibility contract. No full suite is needed for the read-only diagnosis and none was started.

### 2026-10-23 — guarded rollout selected; final plan awaiting review

Hub receipt37cca118-4502-47a9-9f23-1f971793731c submitted the diagnosis and rollout choice. Authenticated reply0d18f5fd-7569-444d-a57a-100cee559713 selects guarded rollout, explicitly rejecting old grant-first compatibility: new landing grants refuse on legacy/unknown live waiters, text queue status warns about an existing grant/legacy coexistence and tells the TL to revoke, and current FIFO/aging remains once legacy waiters are gone. Record the rollout in atlas and tl-lead guidance, never claim the new code prevents every historical mixed-helper deadlock.

Final executable plan: workshop/plans/000570-suite-queue-rollout-plan.md. It adds one exact additive ticket contract marker and one shared Go compatibility assessment over existing raw inspector ticket records; no scheduler rewrite, new schema version, frozen drain JSON change, grant-log cache/format change, automatic revoke or shared-helper substitution. Fresh grant preflight is not a lock over future legacy joins; status and manual revoke cover that accepted risk. Revoke never depends on inspection. Private conformance is proposed instead of :0 staging or live grants. This likely leaves the quick shell, so run inferred full-flow change-code gates after plan approval. No implementation or new suite started.

### 2026-10-23 — final plan approved

Authenticated hub deliveryfbba1f6b-67d8-452d-89ba-6c8a1271f693 approves workshop/plans/000570-suite-queue-rollout-plan.md as written, including private conformance plus post-land hub rebuild and no live grants. Proceed to inferred change-code gates. No permanent implementation has begun; preserve the read-only RED diagnosis and add behavior-specific RED tests after entry.

### 2026-10-23 — implementation and focused verification checkpoint

sdlc change-code --issue570 admitted inferred quick flow (within design limits); close may upgrade the final code diff to full review, as the binary specifies. No flow pin or bypass. Implementation stamps fifo-grant-aging-v1 under the ticket allocation mutex and preserves existing admission. One pure iterator-based assessment supplies grant refusal and text risk/revoke hints. A fresh three-second supplied-inspector preflight runs after caller/argument authorization and before grant append; cancellation after a successful observation still prevents mutation. Revoke never inspects. Frozen queue JSON is unchanged.

Observed RED: Python marker test retained caller-forgery rather than the exact marker; Go risk text omitted warnings/hints and the public grant command accepted a legacy snapshot. The minimal implementation made these pass (Go0.702s). Additional duplicate-policy controls observed RED in both Go and native Python before routing each to the existing strict duplicate-key parser. TL grant guidance phrase guards were observed RED before updating help. Captured historical definitions and the minimized incident fixture are committed test data; the native two-version test runs actual Tickets.admission plus public Go grant/revoke commands against only private lock/grant roots. It reproduces disagreement, refusal and restored progress after revoke, without changing live queues.

The broader focused run passed cmd/sdlc48.597s (including the real helper harness), tl-lead0.240s, Go lint0.903s; full shell lint passed0.283s. Race controls for mixed-policy text, counted N/2N iterator visits, grant/cancel/revoke lifecycle, native supplied snapshot, duplicate authority, frozen JSON and decoder/cancellation passed1.739s. Checkout stayed frozen during cmd/sdlc commands. No full suite has been started. README, atlas, landing help, TL phrase guards and lessons now describe the guarded rollout and its remaining risk.

Next: commit clean implementation, rebuild own slot tools once with Go1.27.3, verify private CLI build metadata, then private executable grant/status/revoke proof. Checkpoint that proof, push and queue the central recorded branch suite; fresh close and normal central merge on SHIP. No :0 staging or live grants; hub owns post-land rebuild.

## Revisions

### 2026-10-23 — hub-approved acceptance replaces universal historical fallback

Delivery0d18f5fd-7569-444d-a57a-100cee559713 supersedes the original first Done-when bullet and initial option analysis: unchanged old helpers cannot be made to elect FIFO by a new field or new-helper conditional. The acceptance contract is now: grant refuses and names legacy/unknown held waiters; status names mixed/unknown ordering and a live granted waiter plus legacy waiter as deadlock risk with manual TL revoke instruction; compatible queues keep the current FIFO-plus-aging rule. Existing grants and later legacy joins are explicitly retained rollout risks. Revoke/expiry/drain restore the existing common order without killing or unlocking waiters. The original universal mixed-version admission assertion is not claimed. This policy is approved; the concrete function-level plan still goes to the hub before code.

### 2026-10-23 — implementation approval

Deliveryfbba1f6b-67d8-452d-89ba-6c8a1271f693 approves the concrete final plan. The earlier proposed alternatives are retained history; only the durable guarded-rollout plan is executable scope. Revoke and suite admission remain independent of compatibility inspection.

### 2026-10-23 — implementation details within approved contract

Use a counted iterator for the shared pure assessment so N/2N tests prove one record visit, and reuse existing duplicate-key parsers at both inspection boundaries; conflicting policy members cannot create compatibility by last-value-wins decoding. No additional rollout state or authority is introduced. The current marker is stamped by the supplying helper, never accepted from its caller. Native private conformance also verifies holder publication retains the marker and unrelated nested metadata.
