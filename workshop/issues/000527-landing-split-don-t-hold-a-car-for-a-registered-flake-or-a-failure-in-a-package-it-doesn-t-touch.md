---
id: 000527
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours: 1.72
collaboration-mode: collaboration
card_mirror: '2ff0fe8a35ee49dc03856be87f56d5255fc2455b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T08:36:43-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# landing split: don't hold a car for a registered flake or a failure in a package it doesn't touch

## Problem

## Spec

Pair train 234 (10-21): the combined-head suite failed only `broadcast.TestHubRandomInterleavings` (/30). The test
was then registered as a known flake (pair#771) after same-head targeted reruns passed. The train's automatic split
still held car 2 (#768) at 08:05:30 "based on the earlier coarsened broadcast failure", even though broadcast files are
identical across the prefix and the failing head and the parent test passed on the same head. #768 (scribe-only) lost
its landing to an unrelated, registered flake.

When a split or bisection assigns blame for a failure, it should re-classify that failure against the **current**
flake registry, and check whether the failing package's sources differ between the passing prefix and the failing
head. A failure in a registered flake, or in a package the suspect car didn't touch, must not by itself hold a car;
report "unattributed: registered flake / untouched package" and continue (or rerun just that test on the suspect
prefix).

### Proposed design — approved 2026-10-21

Before the pure train machine emits a suite-culprit Hold, request one attribution assessment of the exact adjacent passing/failing prefixes. Retain the failed prefix's SHA and local-run record/log, refresh the flake registry in a new issue-records scope, and reuse exact leaf/parent coverage from internal/flakes. Compare the candidate's actual prefix delta using both endpoints' package ownership/dependency information, including test imports, embeds and testdata. Directory equality alone does not establish an untouched package.

A currently registered failure or a provably untouched package is reported as unattributed and never alone holds that car. Neither fact authorizes landing a failed prefix. Existing policy-valid known-flake outcomes remain the sole failure-to-pass authority. In this proposal, only the already passing prefix lands; unresolved cars receive a final non-blaming refusal with the unattributed reason and retry action, rather than automatic Release/retrain. This explicit stop bounds retries: current Release automatically waits again and could loop forever on an unchanged failure. Unknown evidence also stops without assigning a culprit. A genuine unregistered failure the candidate can affect retains ordinary culprit holding.

The plan is workshop/plans/000527-landing-split-attribution-plan.md. Full-flow design approved by TL, including the non-blaming stop instead of automatic retrain.

## Done when

- A train whose only failure is a test registered as flaky (registered before or during the split) doesn't hold a car
  for it; a regression reproduces train 234's sequence.
- A failure in a package the blamed car doesn't change is reported as unattributed rather than pinned on the car.

- No failed or unverified prefix becomes passing from registration or package equality; verified-prefix landing and genuine-culprit holding remain covered.
- Unattributed outcomes name the test/issue or package and a concrete recovery action, with no unbounded automatic retry.
- After a code-main reintegration, a failed multi-car train never splits on old-base smaller prefix heads or lands an untested integration of one.

## Plan

- [x] Add deterministic pure-state and production-train regressions for train 234's external registration sequence and independent-package failure, with positive culprit and gate-safety controls.
- [x] Add the shared attribution decision/effect before suite Holds, preserve exact prefix evidence, and connect fresh registry and dependency-aware impact facts.
- [x] Document the state/event/effect and verify bounded scan/IO work, current flake semantics and the isolated production path before the binary-owned suite, close and landing gates.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. Source is tagged stale by sdlc estimate-source, so calibration is provisional. Derived only after plan-quality round2 cleared both findings.

Familiarity 1.0: existing train, flake, affected and Git seams. Thorough accepted plan gives x0.2 design discount and 15% design buffer. Implementation values below already include the v3.1 40% scale, applied once. Library check: internal/flakes and affected supply matching/ownership/graph behavior; no external API discovery or new dependency. For the new pure classifier, base design 0.5h is halved for those reusable helpers, then discounted to 0.05h. Other design values come from smaller-module base 0.2/0.3h, atlas 0.1h, review 0.05h, and spec 0.5h, each x0.2. Implementation bases are 0.5h per focused classifier/state/IO/impact/regression unit, atlas 0.1h, one review 0.4h, spec 0.2h, each x0.4. Queue idle time is excluded from ship-time; required gate work is represented by the review primitive.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=0.05 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: atlas-docs design=0.02 impl=0.04
item: milestone-review design=0.01 impl=0.16
item: issue-spec design=0.10 impl=0.08
design-buffer: 0.15
total: 1.72
```

Design 0.38 x 1.15 + implementation 1.28 = 1.717h, rounded to 1.72h.

## Log

### 2026-10-21
- 2026-10-21: closed — Real concurrent-train regression refreshes independent tracker registration during split; registered broadcast and scribe-only untouched cases stop without blame and print exact retry while only verified prefix lands. Genuine affected culprit still held. Old-decision mutation reproduces false Hold (14.208s); mixed-base reintegration RED11.430s to GREEN5.34s prevents stale-prefix landing. Root train/split/evidence group passed361.530s; latest reintegration/docs/deadline/8-16 counted IO group passed21.952s; pure classifier/reducer/flakes/affected and Go lint pass. Recorded suite on pushed f2ad2f5c25200df5e05e4791b172466bfa7fc1ca is queued as lease-315431783, not yet passing; completed exact-head suite remains required before landing. Own binary built clean; hub live-gate decision pending, no :0 touched.; review verdict: SHIP

2026-10-21: Claimed and start-plan from current main after #502/PR370 landed. Root-cause trace: runSuite feeds only booleans into Train.Results; Bisect finds adjacent pass/fail bounds and split immediately emits Hold, without re-reading historical failure facts or current registrations. A bounded independent read confirmed that tracker.Fresh alone can reuse the command scope and that directory-only equality misses dependency/test/embed inputs (ARCH-DRY, ARCH-ORDER). Existing Release clears its outcome and waits again automatically, so unconditional release could form an unbounded retrain loop. Baseline pure packages passed: landingline 7.021s, flakes 0.140s, affected 0.396s. No new regression, production code or gate configuration changed yet.

2026-10-21 08:45: Authenticated Couch restart notices. Re-read this Log; clean issue branch and pushed plan survived. sdlc queue status reported complete=true with only peer operations; landing status showed #524/ariadne:3, no owned runner. GitHub still confirms #502 PR370 merged at 15:35:16Z; #527 has no PR or recorded check yet. All own earlier checks/runners had completed before the restart; restarted/requeued nothing. Clarified exact new regression command names in the draft plan before sending it for approval.

2026-10-21: TL approved the full design and explicit stop policy (Couch 538eeecd-6365-4b44-8729-b5a2d048e632, submitted). Unresolved cars receive non-blaming unattributed refusal with exact retry, only verified prefix lands, fresh registry/dependency-aware impact and explicit reducer phase; train234 registered broadcast/scribe-only sequence is mandatory. Entering full-flow change-code; no implementation yet.

2026-10-21: Plan-quality round1 raised PQ-1 (named-function test strategies rather than prose case inventory) and PQ-2 (whole-assessment IO deadline). Revised the plan across both classes: ClassifySplit/Transition/splitImpact/readSplitEvidence/conductor.assessSplit strategies, and one 2-minute outer assessment context with 5-second bounded cleanup, unverified expiry and stalled-dependency proof. This refines the approved bounded effect without changing stop/landing authority. Re-running plan-quality before code.

2026-10-21: Plan-quality round2 CLEAN, both PQ findings addressed, no new findings. Then read shared estimate grammar and resolved v3.1 calibration, v2/v2.1 procedure and baseline-v3.1. Derived 1.72h from itemized Method A components (source stale/provisional), to be set through the card setter and reconciled by change-code.

- 2026-10-21 implementation checkpoint: full-flow change-code admitted implementation after clean plan-quality review and INFO-only estimate review. The estimate's four smaller-module rows map to reducer, exact evidence/fresh registry shell, endpoint-impact integration, and regressions; milestone-review covers execution/review allowance (queue idle is excluded). Captured behavioral RED before reducer changes: TestTrainSplitRequiresAttributionBeforeHold reported “split held car 2 without fresh attribution”. Added exact-pair attribution phase, final non-blaming remainder refusal, unchanged failed Results and explicit retry diagnostics; pure landingline package now passes (0.955s). Classifier bounded helper work is complete; endpoint graph/lifetime and stateful train234 regression are being implemented in separate files. No recorded suite or merge runner is owned yet; root-package testing waits until all parallel file edits freeze because TestMain snapshots this checkout. ARCH-PURE/ARCH-ORDER preserve one attribution boundary before suite Hold, and no registration/impact fact creates passing evidence.


2026-10-21: Integrated exact two-endpoint dependency impact, fresh per-decision registry scope and named evidence checks. Initial production run caught invalid tracker fixture schema and missing tracker in the older Hold control; corrected fixtures use canonical card scaffold plus real tracker bootstrap. New production regressions and genuine-culprit control pass (57.259s: registered sequence18.68s, untouched17.80s, genuine Hold20.27s). Controlled mutation restoring old adjacency-only split reproduced the production false Hold: registered replay failed “held: suite adjacency without attribution, want non-blaming refused” (14.208s). Restored the attribution boundary immediately after that completed test. Internal landingline/flakes/affected passed (0.842/0.073/0.290s), staged-file Go lint passed2.538s. Earlier focused root run passed dependency, resource, evidence-corruption and 8/16 constant-IO controls; no passing authority was changed. Stalled-preparation diagnostic deadline test allows5s for startup under machine load, with the same bounded cleanup contract.

2026-10-21: Broader root train/split/evidence tests passed361.530s. Safety review then reproduced a pre-existing mixed-base reintegration hole: final head was rebuilt on code-main, lower heads were not; production regression RED11.430s observed first car landing despite unverified attribution. Implemented the original approved unverified-evidence policy with existing PhaseTesting/Reintegrated and Refuse machinery: a failed multi-car code reintegration stops all cars before stale-prefix bisection; lower heads/evidence are invalidated. No new phase, outcome type, automatic retry or passing authority. This narrowed implementation needs no train recomposition. New regression GREEN5.34s; docs-only reintegration9.35s preserved; focused group21.952s. Latest pure landingline1.943s and lint1.849s pass. Counted8/16 requests each use one path diff and two endpoint graphs; measured810.066459ms/676.920958ms in real Git/Go fixture. Stalled preparation expired and reclaimed worktrees/child in5.19s. Hub isolated-proof report53572738-c5c4-4648-bc63-fa95f8413ff7 submitted; reintegration heads-up f0e38585-7149-406c-abe5-ec83a7e5fe39 remains queued (hub quota screen), not claimed approved separately. The guard implements the already-approved global unverified stop, using existing state/refusal rather than the additional mechanism considered in that heads-up.

## Revisions

2026-10-21: Refined the original behavior into a proposed attribution boundary and explicit safe stop policy. Registration/untouchedness prevent blame but do not waive passing evidence. The plan proposes final unattributed refusal for unresolved cars rather than immediate automatic retrain; TL approval is required before implementation.

2026-10-21: Recorded TL approval of the proposed stop policy; no design delta.

2026-10-21: Plan-quality refinements name testable functions and enforce the previously promised bounded attribution IO; no change to the approved refusal or verified-prefix-only policy.

2026-10-21: Recorded the code-main stale-prefix regression and conservative existing-state unverified stop; this enforces the approved no-unverified-prefix-landing contract without recomposing trains or adding state/outcome types.
