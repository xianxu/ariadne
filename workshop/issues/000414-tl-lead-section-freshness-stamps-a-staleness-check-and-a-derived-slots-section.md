---
id: 000414
status: open
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: '1c724ffb9a87dde44a8cc50635390eb659b2cc2b' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead: section freshness stamps, a staleness check, and a derived Slots section

## Problem

The TL session file (`teams/<team>/tl-status-<operator>.md` in ops, written only through `tl-lead status`) goes stale:
on 2026-10-16 its Threads section was ~20 hours old while Needs you stayed current. Nothing tells the TL a section is
stale: updating Threads is a synthesis no single event asks for, and writes blocked by the operator's lock get queued
and forgotten. Operator, 10-16: move this into the hard shell; long-term, only Operator instructions (chat) and Needs
you need to be prose in the doc; the rest becomes command output, a "flexible cockpit", some of it eventually a pure
operations dashboard.

## Spec

1. **Freshness stamp inside each section (operator's design: no scattered state).** Every `tl-lead status` write of a
   section puts one line right after its `## ` heading, e.g. `_updated 2026-10-16 17:05_`. Tolerant: a section whose
   stamp line the operator deleted is fine (treated as unknown age) and gets the stamp back on the next write. The
   stamp never counts as operator content (the operator-section guard ignores it).
2. **Staleness check.** `tl-lead status check` (light) reports sections older than their budget (default: Threads and
   Slots older than 2 h, or older than the newest landing the TL recorded); the session-sync ping to the TL carries the
   stale list so it survives compactions.
3. **Derived sections.** `tl-lead status slots --refresh` writes Slots from `couch --list` (slot, agent kind, live or
   not) plus each slot's claimed issue and branch from `sdlc` (title, status); no TL prose. Design it as the first of
   several derived sections (a "landed since the last update" per project from the project ticks is the next).
4. TL prose keeps only judgement lines; (2) flags them when stale.

## Done when

- Tests: a write stamps its section; a deleted stamp is restored on the next write and doesn't trip the operator guard;
  `status check` flags a stale section and not a fresh one; `slots --refresh` renders a fixture fleet exactly.
- Live: the ops session file shows stamps, `slots --refresh` produces today's Slots, and a stale Threads shows up in
  the TL's ping.

## Plan

- [ ]

## Log

### 2026-10-16
