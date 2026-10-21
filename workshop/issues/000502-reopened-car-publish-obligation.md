---
id: 000502
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '4fd7ba1906a3fcaa3baeb7638a93cf0e8b2a664d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T06:49:06-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "7bdb4ff9", done: "665cf118"}
---

# Reopened issue can silently remove a waiting car publish obligation

## Problem

A landing-train car can remain ready after its issue owner reopens it, even with fresh tracker reads. selectTrackedLandingIssues in cmd/sdlc/landingarchive.go delegates to ownedCompletions, which skips working cards. validatePublishAnchors in cmd/sdlc/publishgate.go accepts an empty selection, so the reopened card silently removes the review obligation.

Observed while fixing #499. A temporary extension to TestCarReadinessObservesExternalReclose used a real close-ready repo, SHIP close, correction, corrected SHIP close and productionCarGates with NoCI true. It then externally changed the card from codecomplete to working, preserving the existing completion binding, and called carReadiness again with the fresh per-pass records scope. Expected not-ready; got Ready:true PR:499. Command: go test ./cmd/sdlc -run '^TestCarReadinessObservesExternalReclose$' -count=1. Exit 1 in 13.650s; assertion “reopened issue authorized landing”. Host evidence: /tmp/ariadne499-reopen-red.log and /tmp/ariadne499-reopen-probe.go in ariadne:5; probe was removed from the passing cache fix after capture.

## Spec

Policy proposal for TL triage: an externally reopened card refuses its car with a named reason, since landing work its owner reopened is surprising. Decide how to establish branch ownership after the reopen and how truly issue-free branches behave; do not turn all empty publish selections into refusals accidentally.

### Proposed design — 2026-10-21

The current completion binding remains the authority for a branch's publish obligation even when its card is reopened. Match its repository and the existing evidence-in-head/not-on-base or exact current Close-Token predicate before applying status eligibility. A carried close whose current card is neither codecomplete nor done refuses publication, naming the issue, current status and re-close action. Apply the same guard to ordinary publish and immutable landing publication, so every waiting-car readiness pass sees the current tracker generation.

Do not globally widen ownedCompletions: settlement and archive retain their current eligible-status selection. No carried completion means the existing empty-selection behavior remains valid. One eligible issue's covering review never hides another carried issue's reopened status. Current-token supersession and rewritten evidence retain their existing semantics. The detailed plan is workshop/plans/000502-reopened-car-publish-obligation-plan.md; this proposal awaits hub approval before code.

## Done when

- A deterministic production-gate regression starts with a ready closed car, externally reopens its card, and observes not-ready with a named reason on the next readiness pass.
- Controls preserve legitimate issue-free branches and current completion generation rules.

- Both ordinary publish and production landing readiness refuse the carried reopened close with issue/status/recovery named.
- Rebase/current-token ownership and independently published details remain correct; unrelated working cards and truly issue-free branches do not acquire an obligation.
- Multiple carried issues are checked before selecting a covering review; settlement does not complete reopened work.

## Plan

- [x] Add a deterministic ready → external reopen → not-ready → current re-close → ready production-gate regression, plus ownership and empty-selection controls.
- [x] Reuse status-independent completion ownership in one publish-eligibility guard shared by ordinary publish and landing; preserve settlement and archive selection.
- [x] Update the atlas and verify targeted ownership/readiness controls and the isolated production-path proof after approval. The required recorded suite, close review and landing are tracked by their gate records below.

## Log


- 2026-10-21: closed — Production readiness regression failed Ready:true before fix (8.706s), passed after fix (24.524s), expanded reopen/re-close/push/rebase/independent-details and obligation controls passed 74.975s; broader readiness/landing-publication/generation/archive group passed 223.673s. Lint passed 0.919s and card-model tests 0.474s. Atlas updated. Clean own binary 9f2afa63f780c2660367c9f3743ae6ed01ff5fdd built; hub accepted isolated proof and no separate live check (6d2201e5). Recorded full suite lease-53849329 queued on pushed c0565fcf; passing final-head evidence remains required before landing. Review overlaps the queued suite under tl-worker permission; session remains alive. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
2026-10-20: TL ops:0 requested this separate unclaimed filing; #499 remains scoped to cache freshness/status visibility. Policy not implemented or presumed approved.

2026-10-21: Claimed and started from main after #525/PR368 landed. Read-only tracing confirms statusDecision preserves completion metadata while ownedCompletions filters status before branch ownership. A bounded fresh-context read agrees: reuse binding/repository/current-token ancestry, keep publish eligibility separate from settlement, and check all carried issues before review-anchor ranking (ARCH-DRY, ARCH-ORDER). Existing TestCarReadinessObservesExternalReclose, TestOlderCloseTokenNeverClaimsANewerClose, TestRebasedCloseIsStillOwnedAndGated and TestSquashedCloseRefusesLoudly passed together, 42.504s. No implementation or new regression written yet; reported #499 red evidence remains in Problem. A draft durable plan is ready for hub review. Current Couch has no --description option; no slot label was changed through that rejected command.

2026-10-21: Hub approved the plan (Couch dead3f53-cbb8-4293-9b51-b71b122a51b0, submitted): shared publish-only guard, every carried issue, exact re-close command, unchanged ownedCompletions eligibility; missing/replaced bindings stay outside scope. change-code admitted the quick flow. Implementing with regression-first verification.

2026-10-21: TDD reproduction TestCarReadinessRefusesExternallyReopenedClose failed as expected in 8.706s: external working card still returned Ready:true. Added the shared publish-only guard and factored the existing carried-binding predicate without changing ownedCompletions status selection. The original regression then passed in 24.524s, including named action, ordinary publish refusal and re-close/push recovery. Expanded production cases cover original/rebased/independently published details, unchanged reopened card/branch and settlement selection; pure status coverage, issue-free/foreign/not-carried/on-base controls, every-carried-issue refusal and 8/16 real-Git ownership counts also pass together in 74.975s. One intermediate expanded-fixture failure was the existing transferred-details guard rejecting a re-close edit not based on independently published main details; integrating that version before re-close fixed the fixture, with no production-policy widening. This semantic ownership/eligibility bug has no mechanically detectable API misuse for a new Go lint rule; production behavior regressions defend it.

2026-10-21: Broader readiness, landing publish, rebase/squash/current-token, publication completion and archive regression group passed in 223.673s. pkg/golint passed in 0.919s and internal/issue passed in 0.474s. git diff --check clean. Production diff is 91 added lines (including the factored existing ownership predicate); missing/replaced binding policy and persisted formats remain unchanged. Next: committed binary, pushed PR, recorded full suite and mandatory close review.

2026-10-21: Clean own bin/sdlc built from Git-verified 9f2afa63f780c2660367c9f3743ae6ed01ff5fdd; go version -m reports that vcs.revision and vcs.modified=false. PR370 opened. Hub accepted the isolated production-path proof and explicitly waived any separate live check or :0 stage (Couch 6d2201e5-ce53-40ee-8b28-e32b30dcb59b, submitted). Recorded suite and close remain required before landing.

2026-10-21: First close attempt stopped before review on the remaining checkbox, which combined completed implementation checks with the close review itself. Clarified that row to record completed atlas/targeted/production-proof work; recorded suite, close review and landing remain required through the binary-owned gates. The original recorded suite run lease-53849329 remains queued; no cancellation or restart.

## Revisions

2026-10-21: Refined the original policy proposal into a reviewable publish-only design. The completion binding survives reopen; its ownership predicate can be reused without changing settlement, archive, tracker schema or issue-free publication. Added explicit multiple-issue, status, generation and recovery controls. Awaiting hub approval of this new refusal.

2026-10-21: Clarified the final Plan checkbox to avoid requiring the close review before admitting that same close review. This is bookkeeping only: verification and landing requirements remain in the durable plan and enforced gates; no code or policy change.
