---
id: 000526
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours: 2.79
collaboration-mode: collaboration
card_mirror: 'c15f690588c13a400c4c972c09b0aa22ad988697' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T08:01:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
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

- 2026-10-21 implementation checkpoint: pure/storage protocol TDD and strict codec, atomic drop/merge and apply-batch tests implemented; CLI/conductor/cancellation/positive-cleanup wiring in progress. Main CLI RED unknown command (0.414s), then real three-runner RED no control record (23.763s). First integrated fixture: dropped conductor PASS (obsolete PR CLOSED, survivors one new train, dropped head absent); dropped middle fails recomposition invariant: old conductor releases head before its fresh retry ticket/vouch exists, so another survivor departs alone and the conductor subsequently lands serially. Both surviving heads landed and dropped head stayed out; this is a production handoff ordering defect in the planned retry wrapper, not flaky timing. Raw evidence common test-stats/evidence/000526/shell-green-attempt1.log (package63.394s). Stop dependent handoff implementation, send plan refinement to hub; no timeout/sleep workaround.

## Revisions

- 2026-10-21: Full-flow plan gate CLEAN after two rounds; PQ-1/PQ-2 addressed. Derive provisional2.79h estimate only after plan acceptance, from the named v3.1 calibration.

- 2026-10-21: Hub clarified “returns to waiting” means not lost, not automatically ready: a dropped car stays out until explicit owner retry; its runner exits with reason/recovery like #521 (654ff9bf-3a5d-4ddf-ab4c-b831827f1f83). Done-when wording updated accordingly. Hub endorsed persistent generation-pinned acceptance under line.lock and one atomic drop-versus-merge fence; detailed plan awaiting approval.

- 2026-10-21: Operator approved the detailed plan through hub relay2364fcf8-71f4-477d-804e-c5fa4d1a4d45: atomic admission of the Merge effect is the merge-start boundary; deterministic real-Git fixture is the live conformance, with no staged Pair train. Proceed to change-code and its plan-quality gate.

- 2026-10-21: First full-flow plan gate found PQ-1 (existing close/reclaim is best-effort, insufficient for promised cleanup barrier) and PQ-2 (compress test inventory into named risky functions/strategies). Plan now names durable cleanup obligation, positive closure proof, successor guard covering serial/train attempts and retirement ordering; test strategies compressed. No implementation yet.

## Estimate

Produced via brain/data/life/42shots/velocity/estimate-logic-v3.1.md against baseline-v3.1.md. Method A only; calibration source is marked stale, so provisional. Familiar landing/lease code, familiarity1.0. Thorough design uses15% design buffer. Existing flock, JSON codec, workspace and GH/process seams short-circuit library work; pure control design1.0h is halved for reuse and discounted0.2 to0.10h. Other implementation entries are40% of v2.1 primitive hours, applied once. Issue-spec retains0.60h design (this session's shaping); remaining well-specified primitive design uses0.2 discount. Eight rows map to shaping, pure control, storage, CLI authority, cleanup confirmation, conductor/retry refactor, docs and the one close review respectively.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.60 impl=0.08
item: greenfield-go-module design=0.10 impl=0.24
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.04 impl=0.16
item: api-integration design=0.20 impl=0.40
item: cross-cutting-refactor design=0.10 impl=0.20
item: atlas-docs design=0.02 impl=0.06
item: milestone-review design=0.00 impl=0.16
design-buffer: 0.15
total: 2.79
```

- 2026-10-21 proposed handoff revision (awaiting hub): retain the old head.lock capability in the outer retry wrapper across old-attempt deferred queue cleanup. During the next ordinary attempt, enqueue a fresh ticket and synchronously refresh/vouch it before releasing the old head; then normal Wait/admission/formation proceeds. A scoped fallback releases the retained old turn on every error/cancellation/return before this handoff. This guarantees both ready survivors exist at formation without changing readiness authority or inventing a timer; no suite/CI runs under a non-riding car. Replacing the conductor generation in place was considered but would duplicate the normal admission path and approved fresh-ticket design.

- 2026-10-21: Hub approved proposed handoff revision1139e7ac-1690-4829-bbaf-df866fed2fea; implement without weakening either two-survivor train oracle. Core package/race/fuzz checks passed; retained dropped tickets now excluded from both admission and Board candidate Status, and accepted current drops survive waiter ticket consumption until atomic ApplyTrainDrops seals all requests.
