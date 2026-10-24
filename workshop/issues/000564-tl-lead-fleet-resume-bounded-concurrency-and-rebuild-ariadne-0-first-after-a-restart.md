---
id: 000564
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-24
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'd59fb8ad95be0ee2162932218ad6961dc2095420' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T01:28:24-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# tl-lead fleet resume: bounded concurrency, and rebuild ariadne:0 first after a restart

## Problem

## Spec

From the 10-22 full-shutdown recovery (ops review: `session.side/2026-10-22-30-recovery-review.md`). Two tl-lead
procedure gaps:

1. **`tl-lead fleet resume --apply` is strictly serial,** waiting for each slot's terminal receipt (~1.5 min) before
   the next: 15 slots took ~25 minutes. Allow a small bounded concurrency (e.g. 3 at a time, configurable), keeping
   the per-slot terminal-outcome reporting and stopping on a failure class that suggests a shared cause.
2. **The post-restart procedure doesn't rebuild ariadne:0's tools first.** After a toolchain change (10-22: Go 1.27.1 →
   1.27.3) every slot's `sdlc` is still ariadne:0's old binary until the hub rebuilds it. Add "rebuild ariadne:0
   (`make tools`, verify `go version -m bin/sdlc`)" as the first post-restart step in tl-lead's restart/fleet-resume
   guidance, before the fleet notice.


**Operator refinement (10-22, night), with pair#830:** after a full shutdown the operator wakes only the asleep hubs;
each hub's fleet resume then wakes its own asleep workers **pending-work first, then the remaining asleep slots**
(using pair#830's asleep flag once it lands).


**Interface (TL, 10-23, from pair#830):** the asleep set comes from the existing `couch --recover-plan-from-sdlc` report:
each per-slot row gains `asleep_at` (omitempty), next to its sdlc claim, issue branch, unlanded and dirty evidence, so
one store read gives both the asleep set and the pending-work order. There is no `couch --wake` CLI (after a full
shutdown no live slot exists to call it); the operator's switcher key wakes the asleep hubs, and each woken hub runs
this resume for its own asleep workers.

## Done when

- Fleet resume with concurrency N resumes N slots at once, reports each outcome, and stops on repeated identical
  failures (test with a fake resume verb).
- tl-lead's restart guidance lists the ariadne:0 rebuild first, and startup is a hubs-only quiet hold until explicit operator-requested recovery.
- Only explicit valid asleep marks authorize worker resume; pending-work wave 1 finishes before remaining-asleep wave 2.
- Every hub reports its worker-wave outcomes and completion handoff. The machine steward alone resumes scheduled tasks after both waves complete on every hub, or explicit operator go; worker resume never enables automation/scheduling itself. The external scheduler part is linked as pair#874.

## Plan

Durable design: [fleet-resume plan](../plans/000564-tl-lead-fleet-resume-bounded-concurrency-and-rebuild-ariadne-0-first-after-a-restart-plan.md). Operator approval is recorded in Revisions; change-code remains the implementation gate.

- [x] Review/approve the options and recovery state table through the hub; operator approval and third-wave clarification recorded below.
- [ ] Implement the explicit-asleep planner and bounded receipt-driven scheduler with adversarial tests.
- [ ] Update restart guidance; qualify privately, run recorded pre-close, close and land.

## Log

### 2026-10-22


### 2026-10-24 — recovery and design checkpoint

#580 landed before starting this issue (PR 425); no retained suite/merge remains here. Claimed #564, ran start-plan, and read current fleet/resume/receipt code and recovery guidance. #533 remains parked.

pair:6 confirmed pair#830 is pre-close, not landed. Its schema-1 report carries the optional RFC3339 marker at rows[].agent.rows[].asleep_at; omission cannot distinguish an old producer from an awake thread and authorizes no wake. Normal resume accepts a marked worker without a claim and clears the marker. The current hub-only wake verb does not wake workers. Peer source and tests were read without modifying Pair. Pair's native full-fleet test remains skipped pending pair#863; no native wake or live recovery was run.

The durable plan proposes default concurrency 3 per invocation, pending-phase completion before idle wake, and stopping new launches on the second matching stable receipt failure (immediately for uncertain/unclassified failures), joining calls already started. Shared-tool rebuild remains a steward procedure before the fleet notice. ARCH-DRY keeps one recovery snapshot/classifier; ARCH-RESOURCE keeps a permit for the whole call. Await operator review via the hub before change-code.

### 2026-10-24 — operator approval and scheduler ownership

ops:0 relayed operator approval (submitted receipt e44fc3b4-971d-4160-9bb4-269af41d0540): after a full shutdown only hubs start, automation off and workers asleep; hold until explicit operator-requested recovery. Wave 1 wakes pending-work asleep workers, wave 2 remaining asleep workers, wave 3 resumes scheduled tasks. Clarification 25d26f1d-9e80-483b-829f-7499e93e9f04 makes wave 3 machine-wide: each hub reports both worker waves complete; the machine steward acts only after all hubs report or the operator says go. No hub toggles it alone. Current help's first-operator-message automation toggle must move to that steward step.

Read Pair AGENTS.local.md (root MEMORY.md is absent), current automation owner and scheduler issue evidence without changing source. pair#737 is working/not landed; current Couch exposes per-hub automation controls but tl-lead owns no scheduler. Filed unclaimed pair#874 for the missing scheduler-side machine-wide third-wave contract, with dependencies pair#737 and ariadne#564; report receipt 2dc4144a-4d41-4b42-9715-9314f2cf60cb is submitted. Primary Pair was detached and creation refused before effects; filed from clean resting pair:1 with no checkout writes, status clean afterward. No live automation, scheduling, wake or recovery call was made.

The #564 revision is isolated at /Volumes/TestScratch/tmp/ariadne-slot3/ariadne564-plan.PDsTx9/repo. #594's retained merge (terminal 56512) still owns the durable slot checkout; it is first waiter behind #592. Do not switch that checkout or change its head. Once it is released, detach this private worktree before checking #564 out in the slot, then run change-code and its applicable design gates before any implementation. #589 native qualification remains held.


## Revisions

- 2026-10-24: Corrected the planning interface from the older top-level-marker/no-wake description to pair:6's verified nested asleep marker and hub-only wake semantics. Added explicit options, partial-failure/restart states, counted tests and private-conformance limits in the durable plan. This records a proposed design, not permission to execute fleet recovery.

2026-10-24 — Operator-approved quiet startup and global third wave replace the initial two-wave-only proposal. Keep worker execution scoped to the owning hub; add an explicit handoff and steward-only third wave, with Couch-owned missing part tracked as pair#874. Durable plan renamed canonically and pure/IO classifications and function-level adversarial strategies made explicit. No permission to run a live fleet recovery is implied.

### 2026-10-24 — read-only producer qualification preparation

While #589's fresh suite and mandatory close are retained, inspected the Pair producer without mutation or running a native fleet. Pair origin/main is Git-verified503ba31b9fadffa4e965105f805391eaccdc0af8. RecoverPlan schema1 exposes rows[].agent.rows[].asleep_at exactly as planned; TestRecoverPlanCarriesTheAsleepMark covers positive mark and absent plain-parked omission. However TestDisposableCouchFleetWake on that origin/main still unconditionally skips as blocked by pair#863, even though its landing was reported. This is an explicit native qualification gap, not a passing acceptance result; revisit the owning shared harness before this issue closes, or obtain the hub's approval of the private executable limit as planned. No live worker, automation or scheduler control was called. Existing old-implementation Resume/Candidates baseline passed5.888s before implementation; it is not proof of the new behavior. #589 must land before this issue's code starts.
