---
id: 000635
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: 'd89a71e2d1a6d34702008b0026eda6e9d9f37f00' # card fields mirrored from issue-cards; edit via sdlc
---

# Turn on landing trains for ariadne (.sdlc/landing.json, 60-minute interval to start)

## Problem

## Spec

**Turn on landing trains for ariadne (operator, 10-25).** ariadne lands one PR per full suite (`trains: off`, no
`.sdlc/landing.json`), and an ariadne suite takes both machine test slots, so each landing blocks the whole machine for
16–21 minutes and ariadne landings queue behind each other. pair's trains (`train_interval_minutes: 60`, up to 8 cars,
failing trains bisect) have proven out: train 391 landed pair#866 and #862 with one suite on 10-25.

Add `.sdlc/landing.json` to ariadne with a train interval, starting at 60 minutes like pair (then tune from
test-stats). It's under `.sdlc/`, on the local-gates guard list, so the operator verifies the PR and applies
`local-gates-verified`.

Before enabling, confirm ariadne-specific train behaviour holds: a known-flake rerun and the new incomplete-timeout
handling (ariadne#630) inside a train, and that a fleet-blocking grant still lands alone.

## Done when

- ariadne's landings batch into trains (first multi-car train recorded in the Log with its suite time).
- The config PR is operator-verified; train behaviour with grants, flakes and incomplete runs checked in tests or the
  Log.

## Plan

- [ ]

## Log

### 2026-10-25
