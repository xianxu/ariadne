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
- tl-lead's restart guidance lists the ariadne:0 rebuild first.

## Plan

Durable design: [fleet-resume plan](../plans/000564-tl-lead-fleet-resume-plan.md). Operator review required before code.

- [ ] Review/approve the options and recovery state table through the hub.
- [ ] Implement the explicit-asleep planner and bounded receipt-driven scheduler with adversarial tests.
- [ ] Update restart guidance; qualify privately, run recorded pre-close, close and land.

## Log

### 2026-10-22


### 2026-10-24 — recovery and design checkpoint

#580 landed before starting this issue (PR 425); no retained suite/merge remains here. Claimed #564, ran start-plan, and read current fleet/resume/receipt code and recovery guidance. #533 remains parked.

pair:6 confirmed pair#830 is pre-close, not landed. Its schema-1 report carries the optional RFC3339 marker at rows[].agent.rows[].asleep_at; omission cannot distinguish an old producer from an awake thread and authorizes no wake. Normal resume accepts a marked worker without a claim and clears the marker. The current hub-only wake verb does not wake workers. Peer source and tests were read without modifying Pair. Pair's native full-fleet test remains skipped pending pair#863; no native wake or live recovery was run.

The durable plan proposes default concurrency 3 per invocation, pending-phase completion before idle wake, and stopping new launches on the second matching stable receipt failure (immediately for uncertain/unclassified failures), joining calls already started. Shared-tool rebuild remains a steward procedure before the fleet notice. ARCH-DRY keeps one recovery snapshot/classifier; ARCH-RESOURCE keeps a permit for the whole call. Await operator review via the hub before change-code.

## Revisions

- 2026-10-24: Corrected the planning interface from the older top-level-marker/no-wake description to pair:6's verified nested asleep marker and hub-only wake semantics. Added explicit options, partial-failure/restart states, counted tests and private-conformance limits in the durable plan. This records a proposed design, not permission to execute fleet recovery.
