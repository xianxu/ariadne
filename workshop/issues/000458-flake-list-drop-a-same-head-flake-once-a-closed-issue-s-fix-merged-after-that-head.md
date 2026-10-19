---
id: 000458
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'eb4c3c5fada3733ddf99ae5372d5435e4fdd1427' # card fields mirrored from issue-cards; edit via sdlc
---

# flake list: drop a same-head flake once a closed issue's fix merged after that head

## Problem

`sdlc flake list` keeps showing a same-head fail→pass flake under "not registered (file or register today)" after an issue has fixed it. The 2026-10-19 pilot in pair hit this: five of the eight listed tests had been fixed and closed already (pair#585, #589, #609, #659), and each failing head predates its fix merge. The daily view asks for a filing that would duplicate a closed issue, and the real work gets lost among them. The same happened in ariadne with TestConstructScripts on a2b56335, which predates the #448 fix (6d623285).

## Spec

- A same-head flake counts as fixed when a closed (terminal) issue covers its test and that issue's landing merge is not contained in the failing head. To cover a test, the issue lists it in `flake_tests`, or names it as an exact test in its details. Containment is `git merge-base --is-ancestor`.
- Default view: fixed flakes are dropped. A verbose view (`flake list -v`) shows each one as `fixed by #N (merge <sha>)`.
- A failure on a head that already contains the fix merge is a recurrence. It stays listed, with `recurred after #N`, because that needs a new issue.
- An unreadable tracker, a missing merge commit, or ambiguous coverage (two closed issues) fails toward showing the flake, never toward hiding it.

## Done when

- Tests with fake cards and a real git history: a pre-fix flake is hidden by default and shown as fixed by #N with -v. A post-fix recurrence stays listed, marked as recurred. The ambiguous and unreadable cases stay listed.
- Run against pair's store, `flake list` no longer lists the five pilot tests fixed by pair#585, #589, #609 and #659.

## Plan

- [ ]

## Log

### 2026-10-19

Filed from ariadne:2 after the #394 pilot, on the hub's approval of the proposal.
