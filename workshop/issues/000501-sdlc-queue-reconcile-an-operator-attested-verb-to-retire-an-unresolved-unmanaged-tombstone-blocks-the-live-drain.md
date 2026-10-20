---
id: 000501
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'e96ae5aa9af25d299b5a5c2191fb35fa54003709' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-20
