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

**Revised 10-25 (operator decision): registered flakes don't block landings, and aren't rerun.** Original problem:
a known-flake failure in ariadne#598's merge queued a one-package rerun behind full suites for 30+ minutes. The
operator's model replaces the rerun: a registered flaky test still runs, but its failure is recorded and never
blocks a landing (no automatic rerun). Every registered flake already has an open issue and daily TL triage; a flake
that fails on most runs (e.g. 3 of its last 5 recorded runs) is escalated to the hub as likely broken, not flaky.

## Done when
- A suite whose only failures are registered flakes passes the gate without a rerun; the failures are recorded in the
  attempt and counted in the flake view; tests for flake-only, flake-plus-real (blocks on the real one), and an
  unregistered failure (blocks as today).
- A registered flake failing on most recent runs raises one escalation to the hub.
- The known-flake rerun path is removed (or left unused) with its tests updated.

## Plan

- [ ]

## Log

### 2026-10-24
