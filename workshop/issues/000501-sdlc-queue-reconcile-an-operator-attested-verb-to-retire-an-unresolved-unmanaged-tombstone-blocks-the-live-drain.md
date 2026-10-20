---
id: 000501
status: codecomplete
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'a313bccd7905add7473cacde69d329052f06fdca' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T10:56:04-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.48
---

# sdlc queue reconcile: an operator-attested verb to retire an unresolved unmanaged tombstone (blocks the live drain)

## Problem

ariadne#481's queue keeps an **unresolved tombstone** when an unmanaged direct `make test` owner ends without proven
descendant cleanup (by design: no ps-based absence inference). Only a verified OS reboot retires it today, and
`sdlc queue` has no verb for the operator to reconcile one on the same boot. Live on 10-20:
`~/.local/state/ariadne/queue/unresolved/2b6ca56a5f4147569665ee2f073ec211.json` (managed=false, phase=unknown,
original PID 91801 / start 1792501451.112364, now absent; raw sha256 50d51e03…) keeps `sdlc queue status` at
complete=false forever, so pair#722's live drain can never reach "drained" and its acceptance is blocked.

## Spec

(Design from ariadne:3, #481's author.) `sdlc queue reconcile --id ID --digest SHA256 --note TEXT` (the steward/:0 authority, like pause and grants):
- under control.lock, require the same unknown unresolved record bytes (matching digest), no operation lease or record
  still active for it, and the original PID incarnation absent;
- preserve the original record, digest, actor, note and time in durable reconciliation evidence, then retire only that
  tombstone;
- repeating with the same evidence is idempotent; a changed record, a live owner or an uncertain check refuses;
- keep interruption evidence and pause owners; never resume or restart work; no automatic ps-based inference: the
  note is the operator's attestation that the operation left no descendants.

## Done when

- Tests: matching digest plus absent owner retires the tombstone and writes evidence; wrong digest, live owner or an
  active lease refuses; a repeat is idempotent; pause owners are untouched.
- Private real-executable recovery of the 10-20 unmanaged record shape preserves the exact original bytes in durable audit, repeats without another audit entry and ends with running/complete/empty private status. Crash and append/sync failures preserve evidence and recover without false confirmation.
- Read-only live acceptance after the reboot confirms the original tombstone is absent. Unrelated running work or legacy observations may keep the live snapshot incomplete; do not fabricate production residue or clear others' work for this acceptance.

## Plan

Durable design: [queue reconciliation plan](../plans/000501-queue-reconcile-plan.md). Full flow, one atomic close boundary.

- [x] Gate-serialized exact-record repair: operator note, digest, operation/lease/process exclusion, one immutable audit append/fsync before tombstone unlink; no model/bundle/retention machinery.
- [x] Incident-shape recovery, refusal, real crash/retry, future-field and pause/interruption preservation controls; real kernel lifecycle checks.
- [x] Steward CLI/help and atlas, tracked Go lint, private real-executable conformance and read-only live observation.

Delivery gates remain the recorded suite, close review and central landing; they are outside this implementation checklist. A passing suite precedes merge.


## Log

### 2026-10-20
- 2026-10-20: closed — Resolved the #489 integration conflict by retaining main jsonstrict Scan and UseNumber unchanged, preserving both branches tests. Focused jsonstrict, queuegate reconciliation and real CLI controls pass on the frozen integrated head; full recorded suite passes. Prior private executable recovery, exact audit/replay and read-only live acceptance remain unchanged. (fixed-then-ship; full suite: Recorded sdlc-local/suite PASS on 904e2a9ffb758a43ee764c12b8dad823a8c351b6: 789.0s, 8 shards, 1875 cmd/sdlc tests and 75 other packages, zero failures.); review verdict: SHIP
- 2026-10-20: closed — Rebuilt modified=false executable passes seven private incident-recovery controls, final running/complete/empty with exact original audit bytes and idempotent replay. Nine before/partial/after append and log/root/post-unlink sync failure controls pass; failed pre-unlink effects preserve the tombstone, post-unlink uncertainty retries without duplicate audit. Real crash and kernel owner conformance pass. All TestQueue controls, queuegate/jsonstrict packages and Go lint pass. README/help/atlas and active Done-when now match the approved post-reboot scope; old acceptance is preserved in Revisions. Replacement recorded suite is queued and must pass before merge.; review verdict: SHIP
- 2026-10-20: flow upgraded quick → full — 372 added lines in code files (limit 100); an earlier round of this close already ran the full review

### 2026-10-20 — claimed and durable design

Started from main after #492 landed; no stacked code. The existing unknown tombstone deliberately lacks a live descendant lease, so a new supervisor cannot retroactively prove absence. Reuse control.lock for concurrency and require the explicit operator note for descendant absence; kernel evidence only proves the original incarnation ended. Plan preserves original raw bytes in an immutable audit bundle before deletion and names finite retention/replay semantics. Full design awaits the hub before change-code; no implementation or production mutation.

## Revisions

### 2026-10-20 — post-reboot live acceptance (proposed, awaiting hub)

The original 10-20 tombstone no longer appears in current live status after the OS reboot. Preserve the original Done-when as history; proposed replacement evidence is same-boot private executable conformance against that unmanaged record shape, then read-only live observation/operator acceptance. Unrelated running and legacy-overlay entries can still keep live complete=false; do not fabricate a production tombstone or stop other work to make this issue pass. Also make audit retention explicit:30 days/128 completed bundles with partial transactions pinned and replay refusal after evidence expiry. See durable plan for alternatives, exact states, failure tests and all ARCH principles.

### 2026-10-20 — approved simplification

Authenticated hub receipt 9a890439-8430-48c5-b14b-97462e76a4bc approves exact digest/lease/process/note checks under control.lock, then one append-only base64 audit line, fsync, unlink. It explicitly withdraws the proposed model, bundle and retention/expiry machinery, and approves private executable plus read-only live acceptance after reboot. Flow remains inferred. The durable plan's latest Revisions section is operative; prior proposal/checklists are superseded, retained as history. Proceed to change-code and TDD, no scope extension.

### 2026-10-20 — TDD and targeted implementation evidence

Observed RED through the real Cobra command (unknown --id before implementation). The new store and command now pass exact incident-shape recovery; base64 original preserves opaque large numeric fields and raw bytes. Controls refuse wrong digest, missing note, held tombstone, any operation file/lease, live or unknown owner, unsafe/symlink/wrong-phase/incomplete records, changed post-audit evidence and corrupt/torn audit. A helper process exits71 after actual log fsync and before unlink; its parent confirms durable audit plus surviving tombstone, then idempotent retry removes only that tombstone without another line. Pause/interruption bytes are unchanged; pending cancellation retains interruption evidence and does not become orphan unknown. Real kernel helper controls pass live/different-incarnation/waited-for-dead observations.

The audit's shared duplicate-key checker initially rejected an unknown future JSON number (1e9999) by converting it to float64: new queue-codec and jsonstrict controls both failed, then one UseNumber line fixes that semantic gap without weakening strict known-field/duplicate checks. Queuegate full package PASS6.405s; jsonstrict PASS0.068s; all TestQueue controls PASS23.904s; tracked Go lint catalog PASS0.846s. No live mutation. Private executable conformance is next, before the recorded suite/close/landing.

After the Couch relaunch, sdlc agent-id reports Agent: codex/-/-; copy that available identity rather than invent model/effort metadata for new commits.

### 2026-10-20 — private executable and implementation checkpoint

Built this slot's source527f7aa5d04aadcc3ac6a9e75f27ce1fbdad3ec4; go version -m independently confirms that revision, trimpath=true and modified=false. Private fixture /Volumes/TestScratch/tmp/ariadne-slot3/ariadne-501-conformance.gaD8A0 retains probe.py and transcript.log. Seven real-executable controls PASS: missing note, wrong digest, real held operation lease and actual numbered-worker authority refuse without state/audit changes; operator-context private recovery saves exact base64 original/actor/note/time; replay leaves exactly one unchanged audit line; final private status is running/complete/empty. All queue/lock/discovery overrides are private; only the isolated operator-context subprocess omits COUCH_TREE, and worker refusal is separately checked. No production pause/tombstone was changed. Remove the owned private fixture after landing.

Read-only central live status contains neither the original10-20 unmanaged tombstone nor the pre-reboot pair:3 tombstone. It remains complete=false because of named active work and two unbound legacy pending-admission tickets, unrelated to #501; no absence claim about that work and no attempt to clear it. The hub approved this post-reboot acceptance boundary.

Current implementation checkboxes now reflect the approved simplification. Preserved original proposal checklist (withdrawn by hub9a890439):

> - [ ] Pure request, audit receipt and effect/replay model with strict additive codecs.
> - [ ] Gate-serialized durable audit-before-unlink transaction, failure/replay/identity/retention controls.
> - [ ] Steward CLI and real private executable conformance; recorded suite, close and landing.

### 2026-10-20 — first close findings (BR-1/BR-2/BR-3)

Round1 found no Critical correctness issue but raised three Important gaps; close did not finalize. README now documents the invocation/authority/attestation/audit/retry contract alongside help and atlas. Active Done-when now reflects the hub-approved post-reboot private recovery plus read-only live acceptance. Original superseded criterion is preserved verbatim:

> - `sdlc queue status` on the live machine becomes complete once the 10-20 tombstone is reconciled.

The previously checked-off append/sync test claim was premature: the process crash proved successful-fsync ordering only. Added TestReconcileIOFailures with nine before/partial/after-effect controls for append, log fsync, root-directory fsync and post-unlink directory fsync. All nine were RED when the production IO had no fault seam, then GREEN after routing the existing writes/syncs through two optional per-store IO callbacks. This adds no state, transaction model or retry policy. Pre-unlink errors keep the original tombstone; a torn append stays refused without truncation; an uncertain post-unlink sync retries against the same audit without a duplicate. Cancelled only this checkout's old queued suite before the source change, leaving others untouched; a new recorded suite will run on the final pushed head.

Post-fix controls: all nine IO failures PASS (package0.153s), full queuegate PASS0.954s, jsonstrict PASS0.121s, all TestQueue controls PASS14.067s, tracked Go lint PASS0.621s. The IO callbacks change testability only; default append/fsync/unlink order is unchanged. README/active acceptance now match approved scope, with superseded criteria preserved above. Re-run central close to dispose BR-1/2/3; no finalized close generation exists yet.

### 2026-10-20 — integration with landed #489

Round2 closed SHIP with no open findings, and recorded suite ebe98520937f7852044ae44a3e40d04a passed on fe03d734f03caabf9bd30cf238d24e9dcc8f8008 (688.3s, eight shards, 1871 command tests and 75 other packages). The landing line then admitted #501 after #489 landed at main24dc067b4af6884677ba3dcaf74e266b7ac16021. Automatic integration refused one content conflict in jsonstrict.go and restored the branch.

Manual integration keeps #489's shared Scan entry point and its UseNumber setting, which supplies exactly the opaque-number behavior added by #501. Both branches' regression controls remain. No reconciliation behavior or acceptance scope changes. Verify the merged controls, run a fresh recorded full suite on the integrated head, then re-close with --fixed-then-ship and --full-suite before re-entering the central landing line.

The merged jsonstrict controls passed0.145s and queuegate controls passed0.487s. The first command-control run printed PASS but its hermeticity guard refused because this agent edited the issue log during that run; this is not a passing run or a product regression. Commit the resolution first, then rerun focused controls with every tracked file frozen.
