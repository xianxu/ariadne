---
id: 000580
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '2fd4c20cced97b57d570664f7bab660421fd703f' # card fields mirrored from issue-cards; edit via sdlc
---

# Active-time attribution cross-counts concurrent same-slot claims instead of active branches

## Problem

On2026-10-23, parley.nvim:3 held concurrent claims for parley.nvim#418 and parley.nvim#322. Actual work stayed on #418's issue branch until its landing at14:29 and hub dispatch to #322 at14:30. A fresh main-based #322 branch began then. At14:42, sdlc close --issue322 proposed4.59h measured actual, including a dominant203.5m/74% segment11:18–14:42. That segment includes #418's targeted flake classification, queued verification and landing, not #322 implementation.

Observed measurement text: window2026-10-23T00:22:39-07:00 through2026-10-23T14:42:03-07:00; provenance active-time-v3-engagement-v1;1 ownership interval;5 admitted sessions; dominant #322203.5m/74%11:18–14:42. Code candidate19dda8d9339507f2cfd8b368a8505a25d133e023; tracker5dcff7790ce1157ff18229a5c7d019573a3befcd. The reporter stopped the close review before finalization; tracker remained working with no completion and no actual adopted. TL35643b0f directed --no-actual, no typed hours, and this unclaimed filing.

The printed exclusions already removed some segments adopted by #418's earlier closes. Thus this particular capture directly proves wrong-issue attribution of #418's post-close tail, not that the same exact interval was persisted into both final cards. Concurrent claims must not cross-count one slot's work or assign an unadopted tail to the other claim merely because its ownership interval remains open.

## Spec

Attribute each interval to the issue whose branch the slot was actually working on, while retaining verified ownership/session provenance and disjoint adoption. A second held claim is a reservation, not evidence of active work. Preserve unknown branch evidence explicitly rather than choosing a concurrent claim. Branch transitions, post-close work, queued tests and the final landing tail must remain associated with the actual owning issue until that work ends. This is an attribution repair in Ariadne, separate from Parley's product fix.

## Done when

A deterministic fixture with two simultaneous same-slot claims, activity on A through close/test/landing, then a main-based branch for B, attributes the intervals exactly once to their witnessed active issue. It reproduces the observed A-tail-to-B failure without the correction. Close/actual/active-time share the corrected classifier; previously adopted intervals remain excluded and ambiguous branch evidence is reported rather than guessed. No manually typed time totals are used to hide the defect.

## Plan

Investigate branch/engagement evidence in active-time-v3 and build the same-slot two-claim regression before changing classification.

## Log

Filed unclaimed from parley.nvim:3 at TL request. Product candidate is parley.nvim PR344; #418 landed as PR338. Relevant hub dispatch1341251d at14:30:52 authorizes starting #322 after #418. Clean #322 implementation began14:30; #418 tail of concern is11:18–14:30. No measurement-source repair was attempted.
