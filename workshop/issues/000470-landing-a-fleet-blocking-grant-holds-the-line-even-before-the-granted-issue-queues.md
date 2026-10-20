---
id: 000470
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '280556c2c6e6a625b7688e48f61d528e10cb5f50' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T18:22:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "087e030e", done: "cac5de36"}
---

# landing: a fleet-blocking grant holds the line even before the granted issue queues

## Problem

2026-10-19 ~12:20: ariadne main was red (#469 fixes it) and #469 held a fleet-blocking landing grant, but #469 had not yet
entered `sdlc merge`. The landing line admitted #465's waiting merge anyway: it integrated main and pushed, about to spend a
two-slot recorded suite on a known-broken main. A fleet-blocking grant only reorders merges already waiting; while the granted
issue is still being worked, everyone else keeps landing (and failing).

## Spec

While an unexpired fleet-blocking grant exists for an issue that is not yet waiting in the line, the line admits no other
merge in that repo (they wait, with a message naming the grant, its reason and expiry). Grants that are not fleet-blocking
keep today's reordering. Expiry or revocation releases the hold. `sdlc landing status` shows the hold. State table in the plan.

## Done when

- A fixture: grant for #A (not queued) + waiting #B → #B is not admitted and reports the grant; #A enters → #A admitted first;
  grant expires with #A never queued → #B admitted.
- Multiple grants admit granted tickets FIFO while holding normals; cancellation/hand-off retains the hold until revoke/expiry.
- Hold output/status names issue, reason, expiry and exact revoke command; unbounded grant writes are refused.
- Existing grant-ordering tests unchanged.

## Plan

- [ ] Add deterministic line fixtures for absent grant, granted arrival, hand-off, expiry/revoke, multiple grants, and cancellation/output; prove red.
- [ ] Reuse the grant authority in pure admission and status/wait reporting; require bounded expiry on grant writes.
- [ ] Update help/README/atlas, run targeted tests and recorded suite, then close review and report for SHIP.

Actions: enqueue granted issue; revoke grant; wait for expiry; cancel waiter.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Running head plus new grant | Existing landing continues | Wait for head | None |
| Valid grant, no granted ticket | Held by issue, reason, expiry, revoke command | Enqueue granted issue or revoke grant | None |
| Several valid grants | Granted tickets FIFO; normals held | Enqueue granted issue | None |
| Granted waiter cancels/dies or fails | Grant still holds normals until expiry | Revoke grant | None |
| Last grant revoked/expires | Normal FIFO resumes | Wait for admission | None |
| Malformed/unbounded grant | No authority (existing read policy) | Reissue bounded grant | None |

ARCH-DRY/PURE: one pure grant-hold rule uses existing ValidGrants; no additional lease or persisted state. The valid grant itself is the resource whose lifetime holds normal admissions. ARCH-ORDER: admission evaluates current grants under line.lock after acquiring head.lock; running heads are not preempted. Existing landing completion revokes before releasing head. ARCH-SECURE/EVOLVE/FUNERAL: retain existing authority and log formats/retention; reject invalid grant expiry at the writer and reader. ARCH-MOCK: real temp-dir flocks and persisted grant log cover the production seam. ARCH-CONSTRAINTS: each 2s poll adds O(G) grant formatting only when held; no new subprocess/IO per ticket. Existing queue scan/sort costs remain. ARCH-PURPOSE: cover both the absent initial issue and between-granted-issues hand-off; no train redesign (#473/#475) or baseline classification (#477).

## Log

### 2026-10-19

## Log

- 2026-10-19 ~15:00 (hub): a second case. ariadne:3 held the turn with #459 (granted) and would switch the same checkout to
  #465 (granted, fixes the flake hitting every ariadne suite). Between #459 exiting and #465's ticket existing, the normal
  waiters #466 and #467 could be admitted first. A grant also needs to cover "my next granted issue in this checkout", or a
  short hand-off hold. Same fix shape: a fleet-blocking grant holds the line until its issue queues or the grant expires.

## Revisions

- 2026-10-19: TL approved clarified policy: while any valid grant exists, normals cannot acquire a free head, even when a different granted issue has queued. Granted tickets retain FIFO; active heads continue. Require expiry on grant writes and exact revoke recovery in output.

- Root cause: NextHead selects Order first without considering grants whose tickets are absent; tryAdmit uses that answer after taking head.lock. Existing grants are all fleet-blocking (no other grant class). TL approval received via Couch.
