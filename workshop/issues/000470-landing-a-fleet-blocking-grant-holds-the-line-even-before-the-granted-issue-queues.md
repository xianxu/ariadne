---
id: 000470
status: codecomplete
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '4a4d90949127c7278078dd11a4291adb8bca0607' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T18:22:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "087e030e", done: "cac5de36"}
actual_hours: 0.47
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
- Existing grant FIFO, normal FIFO after the last grant ends, and running-head non-preemption guarantees remain covered; fixtures reflect the revised grant-hold policy.

## Plan

- [x] Add deterministic line fixtures for absent grant, granted arrival, hand-off, expiry/revoke, multiple grants, and cancellation/output; prove red.
- [x] Reuse the grant authority in pure admission and status/wait reporting; require bounded expiry on grant writes.
- [x] Update help/README/atlas and verify with targeted tests and the recorded full suite.

Close via the mandatory boundary review, then merge on SHIP and report to the hub.

Actions: enqueue granted issue; revoke grant; wait for expiry; cancel waiter.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Running head plus new grant | Existing landing continues | Wait for head | None |
| Valid grant, no granted ticket | Held by issue, reason, expiry, revoke command | Enqueue granted issue or revoke grant | None |
| Several valid grants | Granted tickets FIFO; normals held | Enqueue granted issue | None |
| Granted waiter cancels/dies or fails | Grant still holds normals until expiry | Revoke grant | None |
| Last grant revoked/expires | Normal FIFO resumes | Wait for admission | None |
| Malformed/unbounded grant | No authority (existing read policy) | Reissue bounded grant | None |

ARCH-DRY/PURE: one pure grant-hold rule uses existing ValidGrants; no additional lease or persisted state. The valid grant itself is the resource whose lifetime holds normal admissions. ARCH-ORDER: admission evaluates current grants under line.lock after acquiring head.lock; running heads are not preempted. Existing landing completion revokes before releasing head. ARCH-SECURE/EVOLVE/FUNERAL: retain existing authority and log formats/retention; reject invalid grant expiry at the writer and reader. ARCH-MOCK: real temp-dir flocks and persisted grant log cover the production seam. ARCH-CONSTRAINTS: each 2s poll adds O(G log G) sorted grant formatting only when held; no new subprocess/IO per ticket. Existing queue scan/sort costs remain. ARCH-PURPOSE: cover both the absent initial issue and between-granted-issues hand-off; no train redesign (#473/#475) or baseline classification (#477).

## Log

### 2026-10-19
- 2026-10-19: closed — Recorded suite passed on 8530569a (12 shards, 1758 cmd/sdlc tests, 72 other packages, zero failures; sdlc-local/suite success). Real-flock regressions cover absent grant, FIFO, expiry, revoke, cancellation and hand-off; line race tests and focused serial/train output/status/readiness checks pass. TL live smoke verified grant hold, bounded hours and revoke recovery; final fix preserves the existing train output prefix. Stage cleared; atlas/README/help updated. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

## Log

- 2026-10-19 ~15:00 (hub): a second case. ariadne:3 held the turn with #459 (granted) and would switch the same checkout to
  #465 (granted, fixes the flake hitting every ariadne suite). Between #459 exiting and #465's ticket existing, the normal
  waiters #466 and #467 could be admitted first. A grant also needs to cover "my next granted issue in this checkout", or a
  short hand-off hold. Same fix shape: a fleet-blocking grant holds the line until its issue queues or the grant expires.

## Revisions

- 2026-10-19 (BR-1, close SHIP): the original Done-when promised “Existing grant-ordering tests unchanged.” The absent-grant ordering expectation must change for this feature, and the completed-grant fixture now models production revocation. Replace that literal test-text promise with preservation of FIFO and non-preemption guarantees under the revised hold policy. This records the already-approved behavior without changing scope.

- 2026-10-19: TL approved clarified policy: while any valid grant exists, normals cannot acquire a free head, even when a different granted issue has queued. Granted tickets retain FIFO; active heads continue. Require expiry on grant writes and exact revoke recovery in output.

- Root cause: NextHead selects Order first without considering grants whose tickets are absent; tryAdmit uses that answer after taking head.lock. Existing grants are all fleet-blocking (no other grant class). TL approval received via Couch.

- Verification: new absent-issue, multi-grant, waiter-output and unbounded-expiry fixtures all failed on the original code; line package passes after the fix, including race detection. CLI status regression failed before the fix; serial/train production waiter-output test passes after it. Broader landing/train checks in progress. No landingtrain.go change.

- TL live smoke passed on staged 2619d603 in tools :0: binary revision matched and vcs.modified=false; grant status named hold/expiry/revoke; hours 0 and 30 refused; revoke removed grant/hold. No real waiter available; automated serial/train waiter tests cover its output.
- Broader targeted command TestLanding|TestLine|TestTrain reached PASS, but the overall run failed its real-repository HEAD guard because implementation commits occurred while it ran. This is not passing verification evidence. Recorded suite remains queued on immutable pushed 2619d603 in a detached checkout; use that result for final verification. No code changes followed the staged smoke.

- Recorded suite on 2619d603 ran all shards: all passed except TestEnterLandingLineVouchesWhileWaiting, which caught a changed existing train hold prefix. Preserve that prefix by choosing line/train once when opening the line; strengthened the new waiter-output test to assert both prefixes. This is a reproducible branch regression, not a flake. Full suite must rerun on the fix.

- Final recorded suite passed on 8530569a334f37306ea516c8897769045a3f0fbc: all 12 shards over 1758 cmd/sdlc tests and 72 other packages, zero failures, wall 720.1s. sdlc-local/suite success; log sha256 ecf031f0292e. Focused prefix/readiness/grant status/output checks also passed before the final push. The final code change preserves the existing train prefix; grant/status/revoke behavior is unchanged from the TL's passed live check.

- Close review returned SHIP, with no Critical/Important findings. Advisory BR-1 addressed by aligning the current Done-when wording and preserving the original promise/reason in Revisions. Reviewer independently passed line/race and focused command checks; its broader command group hit the default 10-minute timeout, while the recorded sharded full suite already passed on the final code.

## Log — integration verification

- 2026-10-19: integrated main at 184cd3a42a421c9eea1309871204b7d05aeeb890. The first recorded suite failed only TestLandingTrainLandsEveryCarOnOneSuite/head (fixed two-minute completion ceiling); the complete train test then passed on that exact unchanged head. Registered the exact /head subtest as flake #480 with both results. TL confirmed the load-coupled failure and dispatched its root-cause investigation after this landing. No gates waived.
- Fresh recorded suite passed on the same 184cd3a42a421c9eea1309871204b7d05aeeb890: 12 shards over 1780 cmd/sdlc tests and 72 other packages, zero failures, wall 614.6s; sdlc-local/suite success, log sha256 a609f6f5980d. All train variants passed (/head 12.60s, parent 40.78s). Shared suite queue wait was 1896s. Retrying merge with this passing integration evidence.
