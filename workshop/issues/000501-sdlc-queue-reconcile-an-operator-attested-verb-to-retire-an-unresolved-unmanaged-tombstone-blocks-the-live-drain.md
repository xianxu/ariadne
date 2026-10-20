---
id: 000501
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'b8ef9a6171e5dbc51f17e9f3ead6a538ae3c62b8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T10:56:04-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
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
- `sdlc queue status` on the live machine becomes complete once the 10-20 tombstone is reconciled.

## Plan

Durable design: [queue reconciliation plan](../plans/000501-queue-reconcile-plan.md). Full flow, one atomic close boundary.

- [ ] Pure request, audit receipt and effect/replay model with strict additive codecs.
- [ ] Gate-serialized durable audit-before-unlink transaction, failure/replay/identity/retention controls.
- [ ] Steward CLI and real private executable conformance; recorded suite, close and landing.


## Log

### 2026-10-20

### 2026-10-20 — claimed and durable design

Started from main after #492 landed; no stacked code. The existing unknown tombstone deliberately lacks a live descendant lease, so a new supervisor cannot retroactively prove absence. Reuse control.lock for concurrency and require the explicit operator note for descendant absence; kernel evidence only proves the original incarnation ended. Plan preserves original raw bytes in an immutable audit bundle before deletion and names finite retention/replay semantics. Full design awaits the hub before change-code; no implementation or production mutation.

## Revisions

### 2026-10-20 — post-reboot live acceptance (proposed, awaiting hub)

The original 10-20 tombstone no longer appears in current live status after the OS reboot. Preserve the original Done-when as history; proposed replacement evidence is same-boot private executable conformance against that unmanaged record shape, then read-only live observation/operator acceptance. Unrelated running and legacy-overlay entries can still keep live complete=false; do not fabricate a production tombstone or stop other work to make this issue pass. Also make audit retention explicit:30 days/128 completed bundles with partial transactions pinned and replay refusal after evidence expiry. See durable plan for alternatives, exact states, failure tests and all ARCH principles.

### 2026-10-20 — approved simplification

Authenticated hub receipt 9a890439-8430-48c5-b14b-97462e76a4bc approves exact digest/lease/process/note checks under control.lock, then one append-only base64 audit line, fsync, unlink. It explicitly withdraws the proposed model, bundle and retention/expiry machinery, and approves private executable plus read-only live acceptance after reboot. Flow remains inferred. The durable plan's latest Revisions section is operative; prior proposal/checklists are superseded, retained as history. Proceed to change-code and TDD, no scope extension.
