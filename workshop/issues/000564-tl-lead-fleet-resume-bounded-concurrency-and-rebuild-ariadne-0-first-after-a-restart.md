---
id: 000564
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '1462a05b7a9ff73e15eeb0959f0e267df6bdddf1' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-22
