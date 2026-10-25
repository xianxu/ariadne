---
id: 000621
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: 'f3f2122088d855e4f9be3661a2bc657a23e7b7d2' # card fields mirrored from issue-cards; edit via sdlc
---

# Known-flake exact-package rerun queues behind full suites; use the targeted-run budget

## Problem

## Spec

**A known-flake package rerun waits for a full-suite slot (TL, 10-24).** ariadne#598's merge (pid 74001) finished its
integrated suite at about 22:20 with one failure, the registered known flake `TestKnownFlakeGate` (BR-16, 8.46 s
against its 8 s bound, ariadne#620's registry). The merge then queued the exact-package rerun in the machine-wide
full-suite queue (lease-993501919) at position 3, behind two pair full suites, and held ariadne's landing line (and
the fleet-blocking ariadne#603 grant) for 30+ minutes more, for a rerun of one package that takes seconds.

An exact-package rerun is a targeted run, which the machine budget always allows (its own small slot cap, never a
full-suite slot; constitution §5). It should take a targeted slot (or the light lane), not a full-suite ticket.

## Done when

- A known-flake exact-package rerun inside `sdlc merge` (and close's pre-close, if it shares the path) runs under the
  targeted-run budget, not the full-suite queue, with a test that a rerun while both full-suite slots are held starts
  without waiting for them.
- The merge output says which budget the rerun used.

## Plan

- [ ]

## Log

### 2026-10-24
