---
id: 000402
status: open
deps: [ariadne#392]
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '399e8a1c8c845361b79deb6eaebf153d5f585891' # card fields mirrored from issue-cards; edit via sdlc
---

# Weekly resource trend: sizes per data family and suite durations, with deltas, in #392's snapshot

## Problem

Slow growth never feels urgent day to day (21 MB Couch ledgers, 64 GB of captures, 54.8 GB of event logs). Ops project
resource-observability wants a weekly trend so growth is visible as a delta.

## Spec

- Extend the weekly snapshot that ariadne#392 builds for test stats (same schedule, same brain destination, same
  format conventions) with a resource section: sizes per directory family (pair `repos/*` by artifact family,
  `couch/`, agent transcript dirs), the largest Couch ledgers, and full-suite durations per repo, each with its delta
  from the previous week.
- Read-only and cheap: `stat`-based sizes, no file contents. Sources come from one declared list (shared with ops
  project history-backup's source inventory once that exists).
- Flag any family that grew by more than a configurable share (default 25%) week on week.

## Done when

- A fixture-backed test produces the section and its deltas; a live run on this machine writes one snapshot.

Depends on ariadne#392.

## Plan

- [ ]

## Log

### 2026-10-15
