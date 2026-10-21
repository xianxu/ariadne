---
id: 000527
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '1ccff69f006aeb403f7fe88d6dd656238ce2466d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T08:36:43-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

## Plan

- [ ] Add deterministic pure-state and production-train regressions for train 234's external registration sequence and independent-package failure, with positive culprit and gate-safety controls.
- [ ] Add the shared attribution decision/effect before suite Holds, preserve exact prefix evidence, and connect fresh registry and dependency-aware impact facts.
- [ ] Document the state/event/effect and verify bounded scan/IO work, current flake semantics and the isolated production path before the binary-owned suite, close and landing gates.

## Log

### 2026-10-21

2026-10-21: Claimed and start-plan from current main after #502/PR370 landed. Root-cause trace: runSuite feeds only booleans into Train.Results; Bisect finds adjacent pass/fail bounds and split immediately emits Hold, without re-reading historical failure facts or current registrations. A bounded independent read confirmed that tracker.Fresh alone can reuse the command scope and that directory-only equality misses dependency/test/embed inputs (ARCH-DRY, ARCH-ORDER). Existing Release clears its outcome and waits again automatically, so unconditional release could form an unbounded retrain loop. Baseline pure packages passed: landingline 7.021s, flakes 0.140s, affected 0.396s. No new regression, production code or gate configuration changed yet.

2026-10-21 08:45: Authenticated Couch restart notices. Re-read this Log; clean issue branch and pushed plan survived. sdlc queue status reported complete=true with only peer operations; landing status showed #524/ariadne:3, no owned runner. GitHub still confirms #502 PR370 merged at 15:35:16Z; #527 has no PR or recorded check yet. All own earlier checks/runners had completed before the restart; restarted/requeued nothing. Clarified exact new regression command names in the draft plan before sending it for approval.

2026-10-21: TL approved the full design and explicit stop policy (Couch 538eeecd-6365-4b44-8729-b5a2d048e632, submitted). Unresolved cars receive non-blaming unattributed refusal with exact retry, only verified prefix lands, fresh registry/dependency-aware impact and explicit reducer phase; train234 registered broadcast/scribe-only sequence is mandatory. Entering full-flow change-code; no implementation yet.

## Revisions

2026-10-21: Refined the original behavior into a proposed attribution boundary and explicit safe stop policy. Registration/untouchedness prevent blame but do not waive passing evidence. The plan proposes final unattributed refusal for unresolved cars rather than immediate automatic retrain; TL approval is required before implementation.

2026-10-21: Recorded TL approval of the proposed stop policy; no design delta.
