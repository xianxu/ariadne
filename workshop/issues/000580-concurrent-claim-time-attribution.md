---
id: 000580
status: working
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'cc5b1d7045eb6609727bb76b7740d6abf87b2254' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T19:27:24-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
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

Durable proposal: [concurrent claim attribution plan](../plans/000580-concurrent-claim-time-attribution-plan.md). Await TL approval before code.

## Log

Filed unclaimed from parley.nvim:3 at TL request. Product candidate is parley.nvim PR344; #418 landed as PR338. Relevant hub dispatch1341251d at14:30:52 authorizes starting #322 after #418. Clean #322 implementation began14:30; #418 tail of concern is11:18–14:30. No measurement-source repair was attempted.

2026-10-23: Claimed and started in ariadne:3 after #574 landed. Confirmed the structural cause: EligibleSources admits a persistent root session over every matching owned worktree window, with no branch discriminator; resolveOwnedActualInputsFor retains those windows and the engine's issue-scoped commits can assign A's unadopted tail to B. Existing adopted-credit exclusions cannot prove active ownership of that tail. Own native HEAD reflog confirms per-worktree checkout transitions exist, including issue-to-resting and fresh issue-branch starts; full operation grammar and continuity qualification are planned, not claimed proven. The issue contains a reported incident summary, not raw captured terminal bytes. Proposed a pure shared branch/claim interval selector with stable private HEAD-history capture, saved-window replay and an explicit sole-claim planning compatibility policy. No code, fixtures or production measurement changes yet; plan goes to the hub.
