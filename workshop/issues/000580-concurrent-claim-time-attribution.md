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
flow: {kind: quick, provenance: inferred, spec: "16d489ad", done: "ec9e255c"}
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

2026-10-23 19:32 PDT: Authenticated TL approved the branch/sole-claim/unknown-concurrency policy and the shared captured-history selector with additive provenance and immutable replay. Approval recorded in plan Revisions; entering change-code before test/implementation work.

2026-10-23: change-code entered (quick inferred from design/no milestones so far; full close review will apply if the implementation exceeds the shell). Added TestOwnedActualConcurrentClaimDoesNotCreditForeignBranchTail through the actual production resolver with real tracker claims and a private worktree. RED in 2.830s: B receives its whole 00:02–00:31 claim window although A remains the active branch through code, close and unadopted verification/landing tail; B's main-based branch starts at 00:20. This is a labelled reconstruction of the issue report, not a raw terminal replay. No production code changed. Next: native HEAD-history grammar/continuity qualification, then pure selector and bounded IO integration.

Pure selector controls now pass (0.150s): exact concurrent A/B partition, no credit without ownership, sole-claim planning labels, ambiguity beginning at the second claim, preserved engagement origin, invalid branch identifiers, and counted 8/16 transition sweeps with disjoint/exhaustive selection. A native private Git probe confirmed worktree-local HEAD logs and rename/checkout records; initial worktree creation does not itself name the branch. Vocabulary inspection resolved the terminal-owner distinction: terminal cards keep attribution identity but are not reservations. A new terminal-owner control first failed, then passed after separating reservation cardinality from historical owner evidence. Plan Revisions records these findings. Capture/parser, tracker IO, snapshot provenance and production wiring remain unimplemented; the production incident fixture is intentionally still RED. No suite, review, private runner or other background work is active.

Pure HEAD-history decoder plus native worktree controls are green (0.721s). Native RED identified missing-tab creation and paired rename records, now normalized only in those exact forms. Decoder also refuses partial records, object-chain gaps, reversed timestamps and contradictory branch/tip witnesses, and preserves detached/unsupported intervals as unknown. Found and corrected a plan assumption: active-time currently has no ownership route. Authenticated TL approved the default tracked-repo semantic change (explicit dirs/time bounds only narrow eligibility; multi-issue intervals partition before allocation; no-tracker behavior unchanged), now recorded in Revisions before wiring. Production incident remains RED; no production resolver integration or full suite has run.

### 2026-10-23 20:03 PDT — production resolver branch selection

The incident resolver regression is GREEN (3.738 s), retaining A’s unadopted tail exclusion and B’s switch boundary. Added expired-log refusal and private native capture controls for separate worktree logs, truncated/malformed/oversized history and cancellation; combined targeted controls pass (5.441 s). Immutable tracker event reads now retain lifecycle status, with the existing engagement derivation unchanged (tracker targeted tests 2.371 s). Source selection is wired before measurement persistence; additive audit provenance, standalone command wiring and the rest of the verification matrix remain unfinished. No suite, close or PR yet.
