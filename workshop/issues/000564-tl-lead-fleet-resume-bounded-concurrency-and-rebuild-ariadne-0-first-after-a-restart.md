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

## Done when

- Fleet resume with concurrency N resumes N slots at once, reports each outcome, and stops on repeated identical
  failures (test with a fake resume verb).
- tl-lead's restart guidance lists the ariadne:0 rebuild first.

## Plan

- [ ]

## Log

### 2026-10-22
