---
id: 000280
status: working
deps: [ariadne#277, ariadne#278, ariadne#279]
github_issue:
created: 2026-10-01
updated: 2026-10-02
estimate_hours: 3.51
card_mirror: 'a4818566365763c2342564af1874f7cd54a1e9f2' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-02T10:21:50-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# Publish tested operation recovery contracts

## Problem

A lossy multi-step request channel is useful if agents can verify effects and safely retry operations. Today those guarantees are not a discoverable, tested SDLC contract.

## Spec

Project: pair/workshop/projects/cross-slot-work-scheduling.md. Captured for operator review; no implementation is authorized by this issue creation.

Publish operation-level recovery contracts alongside the binaries/subcommands that own them. For claim, reclaim, planning/implementation entry, milestone/close and publication, enumerate observable effects, repetition behavior, scope/preconditions, lost-response handling and when guarantees cease. Do not label every command idempotent: distinguish read-only, duplicate-safe refusal, convergent retry and non-repeatable/uncertain effects.

Existing claim CAS and TestClaimRaceHasExactlyOneWinner are initial evidence. Repeated claim is safe against an already claimed issue, but later reopening/reclaiming changes the context. Arbitrary subsequent code edits are not made idempotent by the claim. Git transaction receipts and revisions should be reused where they already establish outcomes.

Put claim-before-work and owner-checked continuation in SDLC once, not in every scheduling prompt. Agent guidance should verify effects via read-only commands, treat a roughly 30-second interval as a revisit heuristic rather than proof of loss, and retry only under the operation contract. Unknown observations are not negative evidence.

## Done when

- Agent-discoverable help/docs define recovery behavior and observable evidence for the MVP workflow commands, linked to supporting implementations/tests.
- Race, duplicate, lost-ack and reopened/reclaimed-generation cases validate the declared guarantees.
- A scheduling example verifies claim, branch/activity and gate progress without restating defensive claim logic in every message.
- Unproven guarantees are explicitly unknown; documentation does not claim exactly-once execution of arbitrary agent instructions.

## Plan

Durable plan: `workshop/plans/000280-operation-recovery-contracts-plan.md`.

- [x] M1 — the `internal/recovery` registry (class, effects, evidence,
      preconditions, repeat, lost-response, ends, and proofs naming tests),
      appended to each verb's help by `attachRecoveryContracts`;
      `TestRecoveryContractsAreProven`; uncertain-publication guidance plus
      lost-ack tests for claim and set-status.
- [x] M2 — the `sdlc help recovery` topic (classes, agent guidance, table,
      example; cross-links `issue recovery`); the executable scheduling
      example (`TestSchedulingExampleRuns` runs the rendered steps); docs.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec             design=0.5 impl=0.05
item: greenfield-go-module   design=0.4 impl=0.22
item: smaller-go-module      design=0.2 impl=0.14
item: smaller-go-module      design=0.1 impl=0.14
item: smaller-go-module      design=0.1 impl=0.14
item: smaller-go-module      design=0.1 impl=0.14
item: smaller-go-module      design=0.2 impl=0.14
item: atlas-docs             design=0.2 impl=0.08
item: atlas-docs             design=0.05 impl=0.05
item: milestone-review       design=0.0 impl=0.14
item: milestone-review       design=0.0 impl=0.14
design-buffer: 0.15
total: 3.51
```

Items, in order:
- issue-spec: decisions plus two plan-quality rounds.
- greenfield-go-module: `internal/recovery` (types, validation, rendering).
- smaller-go-module ×5:
  1. the contract test (derived verb set, AST test-name scan);
  2. the `cardPublish` seam, uncertain helper and lost-ack tests;
  3. the two generation tests;
  4. the help topic and the per-verb help sections;
  5. the executable scheduling example.
- atlas-docs ×2: the catalog's contract text, rechecked against code; then
  the atlas, README and manual.
- Two milestone reviews.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.* The calibration source is flagged stale (#127). Recent rows: #277 came in at 2.41h actual against 4.09h estimated, and #279 at 1.59h against 3.31h.

## Log

### 2026-10-01

Captured from the performance → messaging guarantees → SDLC ownership/observability → recovery discussion. No implementation started.

### 2026-10-02 (implementation session)
- 2026-10-02: closed — recovery contracts: internal/recovery catalog (16 contracts over the 26 commands derived from the tree, 8 reasoned exemptions, scope = tracker repositories) appended to every workflow verb's help; sdlc help recovery topic; TestRecoveryContractsAreProven (derived coverage, stale exemptions, help sections, every named proof a declared test; mutation-checked); cardPublish seam + uncertainCardWrite with caller guard; new proofs: claim/set-status lost-ack reruns, close re-run is a new generation, landing leaves a reopened card alone; TestSchedulingExampleRuns executes the documented example across two checkouts with duplicate delivery (mutation-checked); full sharded suite green (processgroup ps test needs the sandbox off; green outside it); --no-project: pair's project tracks ariadne#280 at issue granularity; review verdict: SHIP
- 2026-10-02: closed M2 — sdlc help recovery topic rendered from internal/recovery; TestSchedulingExampleRuns executes recovery.Example across recipient and coordinator checkouts, double-delivering convergent steps (longest-prefix contract lookup), restoring the tracker right after the offline step, failing unknown actors; mutation-checked (flow, verdict, stale, bad command red); plan tables swept against the tree; issue recovery --help links the topic; README + atlas; full sharded suite green (processgroup ps test green outside the sandbox); actual = sdlc actual cumulative 3.29 minus M1's 3.08; --no-project: pair's project tracks ariadne#280 at issue granularity; review verdict: SHIP
- 2026-10-02: closed M1 — recovery catalog (16 contracts over 26 derived commands, 8 reasoned exemptions) in every verb's help with a scope line (tracker repos; legacy unknown); TestRecoveryContractsAreProven (derived set, stale exemptions, help sections, declared tests; mutation-checked); cardPublish seam with caller guard over all initializers + uncertainCardWrite; proofs: claim/set-status lost-ack reruns, close re-run is a new generation, landing leaves a reopened card alone; plan Revisions match the code; full suite green; actual = sdlc actual (first milestone); --no-project: pair's project tracks ariadne#280 at issue granularity; review verdict: SHIP

- Operator authorized the work ("work on #280"). Claimed in ariadne:1 (the
  claimant is recorded) and ran start-plan; the branch sits at main, which
  includes #277, #278 and #279. Mapping where each MVP command's recovery
  behavior lives today (CAS, receipts, recovery reconcile, observe) before
  designing the contract surface.

- Operator decisions:
  1. **Contract home:** one Go registry, like the gate catalog. Each verb's
     entry gives its class (read-only, duplicate-safe refusal, convergent
     retry, or uncertain), effects, evidence query (`issue show --json`
     fields), preconditions, lost-response action, when guarantees end, and
     the names of the tests proving each claim. It renders into each verb's
     `--help` through a placeholder and into one `sdlc help recovery` page.
     A contract test fails if a named test is missing or a verb lacks an
     entry.
  2. **Scheduling example:** executable. A worked example in the recovery
     page (send work, verify claim, check branch and activity, check gate
     progress, and handle unknown or stale) that a test runs step by step
     against a real-git fixture.
- An explorer is mapping each verb's actual recovery behavior and the tests
  behind it.
- Explorer map, per verb:
  - **claim, adopt, relocation, set-status, reclaim:** a single card CAS
    with no receipt, decided by the card on rerun. Convergent, except that
    reclaim alone tells the caller to "rerun".
  - **start-plan:** convergent (creates or switches the branch). It takes no
    repo lock.
  - **change-code:** convergent for git and ledger state. Estimate-quality
    re-dispatches every run.
  - **milestone-close:** a new review round each run (non-repeatable); no
    commit and no card write.
  - **close:** receipt-driven (evidence commit, then codecomplete, then the
    mirror commit); `issue recovery reconcile` finishes an interrupted
    close; a re-run after SHIP is a new generation (non-repeatable).
  - **pr:** the durable path updates the open PR (convergent); the legacy
    path fails on a re-run (gh: PR exists).
  - **merge:** the legacy path refuses a duplicate run; the durable path
    prints a `--branch` recovery command and observes the landing state.
  - **move:** a re-run is refused (duplicate-safe).
  - Gaps: claim, adopt, relocation and set-status return a raw
    `ErrPublicationUncertain` with no guidance; there are no lost-ack tests
    for claim or set-status.
  - Precedent: the gate catalog (`GateCatalog`, rendered via
    `{{GATE_FLAGS}}`, guarded by registered-flag and help tests).
- Wrote the durable plan (M1/M2). Awaiting operator approval.
- change-code plan-quality round 1: the required set is now derived from the command tree; the surface is `sdlc help recovery`; Done-when proofs are named per class, with two new generation tests. The plan is revised.
- Plan-quality cleared (round 2). Estimate: 3.51h (v3.1).
- change-code passed (estimate-quality: expect well under 3.51h, given the calibration). Starting M1: the `internal/recovery` registry.
- M1 steps done:
  - `internal/recovery` holds the types, validation, wrapping and rendering,
    plus the catalog: 16 contracts covering the 26 derived commands, and 8
    exemptions with reasons.
  - Every contracted verb's help carries its section (one pass in
    `buildRoot`).
  - `TestRecoveryContractsAreProven` is mutation-checked: a missing contract
    and a contract for a non-command both go red.
  - Gap fixes: `cardPublish` (one seam, caller-guarded) plus
    `uncertainCardWrite`, used by claim, adopt, relocation, setters and
    reclaim.
  - New proofs: `TestClaimRerunSettlesALostResponse`,
    `TestSetStatusRerunSettlesALostResponse`,
    `TestCloseRerunAfterShipStartsANewGeneration` and
    `TestLandingLeavesAReopenedCardAlone`.
  - Atlas: added `workflow/recovery-contracts.md` and its index link.
- M1 review round 1: FIX-THEN-SHIP. Fixed BR-1 (the scope line on every section: tracker repositories; legacy is unknown) and BR-2 (plan-table Revisions), plus the Minors (stale doc, the caller guard over all initializers, the gitx lost-ack citation). The `sdlc help recovery` pointer resolves in M2.
- M1 closed (SHIP; 3.08h measured, which includes this session's design time). Advisory BR-7 fixed first in M2: the atlas page now carries the scope line, and the M2 page renders `recovery.Scope`.
- M2 steps done:
  - `sdlc help recovery` is a cobra help topic: `helptext/recovery.md` plus
    `{{RECOVERY_SCOPE/CLASSES/TABLE/EXAMPLE}}`, rendered from
    `recovery/page.go`.
  - `TestSchedulingExampleRuns` executes `recovery.Example` across two
    checkouts and double-delivers convergent steps. Four mutations were
    checked, and all go red: flow, verdict, stale, and a bad command. The
    stale mutation first passed because the test derived its setup from the
    expectation; the condition is now an explicit `TrackerUnreachable`
    field.
  - Docs: README, plus the atlas page (topic and example).

- M2 review round 1: FIX-THEN-SHIP. BR-8 (the plan tables were rewritten to current reality after a full row sweep, and a lesson was added) and five Minors are fixed: the harness-flag description, longest-prefix class lookup, the immediate tracker restore, the unknown-actor default, the `issue recovery --help` link, and the 30 s guidance.
- M2 closed (SHIP; 0.21h increment). Advisory BR-9 fixed: a sweep of the backticked identifiers in the plan's Core concepts and the Plan rows (Verb → Verbs, no JSON, `attachRecoveryContracts`). The plan steps are ticked. The process manual picks up `helptext/recovery.md` automatically, which `sdlc process-manual` confirms.
