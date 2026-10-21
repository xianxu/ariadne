---
id: 000463
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '317790ff8bc57d30944392613709328ff76c559a' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T04:30:26-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# local-check: resume recorded known-flake failures with targeted tests

## Problem

ariadne#461's recorded suite attempt fa09ccf7eb6fd0862fadf1d94e865944 failed
only TestSuiteQueuePriorityRealHelper when its private provider launch returned
Operation not permitted. The exact same commit passed that targeted test without
changes, and the test was registered on ariadne#462. Existing local-check run
can invoke the known-flake step only while processing a fresh full suite;
republish reposts an already-saved failure unchanged. There is no supported
way to finish this failed attempt with just the now-registered targeted tests,
so the hub approved another full suite for #461. This is a #394 follow-up.

## Spec

Add an SDLC verb to resume an already-failed recorded suite attempt whose named
failures are all currently registered known flakes. Run only those tests on
the original exact head through the existing known-flake verification path,
then post the suite success with its flake issue links and both attempt records.
Reuse the existing registry, classifier, rerun checks and evidence model rather
than introducing a second acceptance policy. Do not treat manual target output
or a republished failure as a recorded passing suite.

## Done when

- A saved failed suite whose only failures are registered known flakes can be
  resumed with named-test reruns only, without executing the full command again.
- The rerun uses the original exact commit and verifies the saved log/digest,
  repository identity, attempt completion and current registry/policy first.
- Success links the original failure, targeted rerun and open flake issues;
  every required test has an actual PASS. Other tests retain their original
  full-suite evidence instead of being silently marked as newly run.
- Unknown/unnamed failures, timeouts, missing or changed evidence, disabled
  policy, and closed/ambiguous flake registrations cannot become success.
  Revalidate registration and policy before posting.
- Cancellation and concurrent/repeated resume are handled with existing run
  ownership/evidence rules; no duplicate or unconfirmed run becomes a pass.
- Stateful regression tests prove targeted-only execution, exact-head binding,
  refusal paths and linked commit status/reporting.

## Plan

- [ ] M1 — Exact-package rerun batches, immutable linked-proof model and shared evidence snapshot, with focused adversarial tests and atlas update.
- [ ] M2 — Explicit resume command, lifetime lock, saved-proof/republish/pass-guard integration, private executable proof and recorded-suite delivery preparation.

Durable proposal: [resume known flakes plan](../plans/000463-resume-known-flakes-plan.md). Await hub approval before change-code or implementation; close/PR/landing remain mandatory after the implementation boundaries.

## Log

Filed unclaimed at the hub's explicit request, discovered during ariadne#461.
The current workflow required one extra full recorded suite despite established
same-exact-head fail/pass evidence. Original full log is retained under
ariadne/.git/sdlc/test-stats/failed/fa09ccf7eb6fd0862fadf1d94e865944.log.
The classified intermittent provider failure itself belongs to ariadne#462;
this issue owns only recovery of already-recorded failed attempts.

### 2026-10-21 — Resume planning after #522

Authenticated dispatch dff45850-dd79-4197-8aa4-5089515c4de7 prioritizes this recovery after #522, with the plan going to the hub before code. Claimed and start-plan created this independent branch from fresh main. #520 remains held on its separate branch. Read the current knownFlakeStep, saved-attempt/republish/pass guard, immutable attempt writer, bounded store, queue/cancellation and flakes classifier/evidence contracts. No implementation or tests run.

Hub receipt116faa29-e008-4b37-8750-df1f1d747b73 settles the selection question: preserve #466's whole-top-level test (siblings included), but remove the cross-package name cross-product. Shared linked-record resolution is approved as the direction; the linked durable plan proposes the exact CLI, managed batch, lifetime parent lock and crash fence for approval. ARCH-DRY keeps one acceptance classifier/proof resolver; ARCH-ORDER separates reserved/unknown from completed child evidence; ARCH-EVOLVE explicitly names why old readers must refuse the new resume-link variant. Completed full-suite samples stay original, not fabricated from targeted duration. Pending proposal details are not approval.

## Revisions

### 2026-10-21 — Proposed recovery design and exact package boundary

The plan replaces the empty planning row with two genuine review boundaries. It retains the existing Spec/Done-when and adds a proposed immutable resume-link variant plus a parent-lifetime lock to protect concurrent/crashed execution. Version refusal is proposed because ignoring unresolved or failed resume proof would weaken old success guards. The hub must approve these mechanisms before implementation; no new acceptance policy or retry-until-green is proposed.
