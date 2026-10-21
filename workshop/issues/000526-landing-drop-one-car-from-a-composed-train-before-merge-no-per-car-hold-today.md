---
id: 000526
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'b30650d5cbad1c1caf75c854d548315c4a0f97ca' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T08:01:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# landing: drop one car from a composed train before merge (no per-car hold today)

## Problem

## Spec

There is no supported way to drop one car from a composed landing train. 10-21 07:15, pair train 234 (#768, #767,
#765): pair:2 found a late test-only defect in #768 and asked to hold its car; pair:9 (the conductor) found that cars
are pinned at build and merge rechecks only main and the train head, so pushing a fix or drafting the car's PR doesn't
reliably exclude its old commit. The only supported options were abandoning the whole train (requeueing the clean
cars) or closing the train PR mid-flight. The TL let the train land with #768 as is (the defect was test-portability
only), but a real "this car must not land" case would cost the whole train.

Add `sdlc landing drop --issue N` (the car's owner or :0): before the train's merge, remove that car, recompose the
train head from the remaining cars (reusing #521's re-formation path), and rerun only what the recomposition
invalidates. After merge has started, refuse with the exact state.

## Done when

- Dropping one car of three before merge lands the other two with a recomposed head; the dropped car stays out, its runner exits with the reason and exact recovery, and only an explicit owner retry queues it again.
- After merge start, drop refuses and names what happened.

## Plan

- [ ] Implement and verify the generation-pinned drop protocol and atomic merge fence.
- [ ] Wire authorized command, interruptible conductor and survivor re-formation; verify controlled race/crash and real three-car cases.
- [ ] Update help/README/atlas; run recorded full suite and prepare close evidence.

Durable design: [implementation plan](../plans/000526-landing-drop-plan.md).

## Log

### 2026-10-21

- Plan committed and submitted to hub; receipt2475e436-1c76-4c59-a899-06585679a21b submitted. Hub relayed operator review pending (5eb93475-33d7-49f8-9070-cc4584ea31b5): hold code until explicit answer, recommended merge-start boundary and deterministic real-Git fixture conformance. No change-code, implementation, PR, local-check lease or merge runner was started.
- After Couch restart, reread Log and verified clean retained issue branch/design. Queue running complete=true, no ariadne:1 operation; repository line is held by #524/ariadne:3 and trains remain off. GitHub PR lookup for this branch returned no PR (service degradation caveat retained). No runner/check to recover and nothing restarted; operator-approval hold continues.

## Revisions

- 2026-10-21: Hub clarified “returns to waiting” means not lost, not automatically ready: a dropped car stays out until explicit owner retry; its runner exits with reason/recovery like #521 (654ff9bf-3a5d-4ddf-ab4c-b831827f1f83). Done-when wording updated accordingly. Hub endorsed persistent generation-pinned acceptance under line.lock and one atomic drop-versus-merge fence; detailed plan awaiting approval.

- 2026-10-21: Operator approved the detailed plan through hub relay2364fcf8-71f4-477d-804e-c5fa4d1a4d45: atomic admission of the Merge effect is the merge-start boundary; deterministic real-Git fixture is the live conformance, with no staged Pair train. Proceed to change-code and its plan-quality gate.
