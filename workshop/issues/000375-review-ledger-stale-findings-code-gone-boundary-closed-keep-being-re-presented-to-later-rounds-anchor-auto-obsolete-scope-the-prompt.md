---
id: 000375
status: open
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: 'e8bcd7262174f0bb4abd4c18c85852e6d1997755' # card fields mirrored from issue-cards; edit via sdlc
---

# Review ledger: stale findings (code gone, boundary closed) keep being re-presented to later rounds; anchor, auto-obsolete, scope the prompt

## Problem

Each issue's boundary review ledger (`cmd/sdlc/boundaryledger.go`, #187/#194) gives findings stable ids (BR-n) that later rounds must dispose of before raising new ones. The ledger is issue-wide and only grows: when a rework, a merge with main or a milestone change makes a finding moot (the code it pointed at is gone), the row is still re-presented to every later round. Long issues this week (pair#528: 10 rounds; ariadne#332: 8; ariadne#350) spent reviewer time and prompt space re-disposing rows that could no longer apply. Retro item (c); the operator: "a no-brainer".

## Spec

1. **Measure first (one step, sets the baseline):** for this week's long ledgers (pair#528, ariadne#332, #350, #364), count rows re-presented after their anchored code had changed or their boundary had closed; record the share in the Log.
2. **Anchor** each finding when raised: file plus line range plus a content hash at the reviewed commit.
3. **Auto-obsolete:** at the next round, a finding whose anchored content no longer exists at the head under review is marked `obsolete (code changed)` with the diff as evidence, shown collapsed, not re-argued. The reviewer may reopen it with a reason.
4. **Scope the prompt:** only the current boundary's open findings go into the reviewer's prompt; closed boundaries' rows are archived (extend `gatestate.FilterBoundary`).
5. **Manual verb:** `sdlc finding obsolete --issue N --id BR-n --reason …`, contestable by the next round, like `finding dispute`.

## Done when

- A rework that deletes a finding's code makes that finding obsolete automatically at the next round (test); a closed milestone's findings don't reach the next boundary's prompt (test).
- The Log has the before measurement and the prompt-size or row-count reduction on a replayed ledger.

## Plan

- [ ]

## Log

### 2026-10-14
