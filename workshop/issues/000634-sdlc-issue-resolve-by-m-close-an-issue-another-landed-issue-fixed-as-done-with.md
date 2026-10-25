---
id: 000634
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: 'c0fad8ef277f2165b70b68271779b2e78087fdbc' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc issue resolve --by M: close an issue another landed issue fixed, as done with provenance

## Problem

## Spec

**No honest way to close an issue that another issue fixed (ariadne:2, 10-25).** ariadne#625 (a flake) was folded into
ariadne#622 by TL decision and is fixed by #622's merge (4c164029). Closing #625 now means either a full
start-plan/close/review/merge walk over a docs-only window (and, without an affected declaration, a full suite;
see ariadne#632) for no code, or `abandon --as wontfix`, which misrecords a fixed defect as rejected. `sdlc issue
set-status` refuses codecomplete and done by design (#160).

Add a supported terminal transition for "resolved by another issue": e.g. `sdlc issue resolve --issue N --by M`
(M must be landed: its merge commit is verified on main), recording status done with a `resolved_by: M` field and a
Log line citing M's merge, no actual hours, and (for a flake issue) unregistering its tests exactly as a normal close
does. Refuses if M isn't landed or N has unlanded commits of its own.

## Done when

- `resolve --by` closes an issue as done with the provenance recorded; refuses an unlanded M or an N with its own
  unlanded work; a flake issue's tests are unregistered. Tests for each.
- The vocabulary model (construct/vocabulary/issue.cue) carries the new transition and field.

## Plan

- [ ]

## Log

### 2026-10-25
