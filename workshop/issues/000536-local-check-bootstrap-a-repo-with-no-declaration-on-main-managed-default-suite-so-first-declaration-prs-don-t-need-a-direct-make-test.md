---
id: 000536
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-22
estimate_hours: 1.27
collaboration-mode: collaboration
card_mirror: '97c37eebd3541dcba599e990a4475d535a4f1c09' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T01:19:13-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# local-check: bootstrap a repo with no declaration on main (managed default suite), so first-declaration PRs don't need a direct make test

## Problem

## Spec

From ops-echo:0 (10-21): `sdlc local-check run` refuses in echo because origin/main has no `.sdlc/local-checks.json`
yet, and echo#1 is the PR that adds echo's first declaration (a bootstrap case). echo:1 therefore ran `make test`
directly; it was tracked as an unmanaged suite, and when its owner exited it left a tombstone that held machine-wide
suite capacity until the :0 steward reconciled it (op 65dbdc31, 18:30).

Fix: a supported bootstrap path. When main has **no** declaration, `local-check run` uses the built-in default suite
(`make test`, the same default `effectiveLocalChecks` already injects) under a normal managed lease and records it,
or, for a PR that adds the first declaration, uses that PR's declaration for its own evidence only, clearly labelled
as bootstrap. Never weaker than the default; refuse only when there is no Makefile `test` target.

## Done when

- In a repo with no declaration on main, `sdlc local-check run --name suite` runs `make test` under a managed lease
  and records the result (test).
- The first-declaration PR can produce recorded evidence that its own merge accepts (test), and no unmanaged run is
  needed.
- Bootstrap is allowed only when main has neither a declaration nor a test target and the exact pushed subject has a test target. It is visibly labelled and runs the required default `make test`; head declarations cannot replace/lighten it or supply CI substitutes/custom checks.
- Existing main declarations (including valid empty), default command/weight, server CI and unknown/invalid probe results remain authoritative and strict. All landing subjects use the same policy; a prefix without a test target cannot virtual-pass.
- A committed regression replays echo#1's verified 2026-10-21 main/head Makefile and declaration state, reproduces the refusal before the fix and passes through managed execution/admission afterwards.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. Calibration is marked stale by estimate-source, so its rates are provisional.

Decomposition: shared subject-policy helper and per-prefix train selection are two smaller Go extensions; CLI/CI subject/policy wiring is one cross-cutting refactor; issue shaping, documentation and the one close review are separate primitives. Existing constructors, bounded make/Git probes, managed runners and native/stateful fixtures cover the stack: no greenfield library/API discovery. Thorough approved plan applies design ×0.2 and +15% buffer; familiar stack applies impl ×1.0. Impl values below already apply v3.1's 40% scale to the v2 table, exactly once.

Raw table picks: issue-spec D0.50/I0.10; each smaller-go-module D0.20/I0.50; cross-cutting-refactor D0.50/I0.50; atlas-docs D0.10/I0.15; milestone-review D0.10/I0.50. Discounted design subtotal 0.32, scaled implementation subtotal 0.90: 0.32 ×1.15 +0.90 =1.268, rounded 1.27h. Suite capacity waits remain excluded by measured actuals.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.10 impl=0.04
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: cross-cutting-refactor design=0.10 impl=0.20
item: atlas-docs design=0.02 impl=0.06
item: milestone-review design=0.02 impl=0.20
design-buffer: 0.15
total: 1.27
```

## Plan

- [x] Reproduce the incident and share subject-aware main-first policy between CLI and CI facts, preserving exact pushed-head and cancellation controls.
- [x] Cover ordinary/immutable landing and train prefixes/reintegration with the same derivation, no weakened gate or fabricated baseline evidence.
- [ ] Document the contract, complete focused/recorded verification and prepare close. Detailed steps: [bootstrap suite plan](../plans/000536-bootstrap-suite-plan.md).

## Log

### 2026-10-21

### 2026-10-22

- Claimed and started from fresh main after #545 landed/reported and hub rebuilt ariadne:0. Existing no-declaration/main-test case is already supported by TestLocalCheckCmd_DefaultSuite; the missing bootstrap case is no main test target. Hub approved exact-subject default only, no head custom declarations or CI substitutions, and requested the real echo regression.
- Verified echo#1 archived Log and native Git objects read-only: aeee435d3953c60cdd1ab4350e206d0e162646f7 and final pre-merge main 77a0fd6320ccc5ff603c9cfefdb8bb4789486556 share Makefile blob 0e078508dd6066bb3ac059e6a5b649e26a568694 and lack tracked Makefile.local, Makefile.workflow and local-check declaration. Tested head e3965c2c6d79ece162e704d07335b5b883fe9668 adds the test target and first make-test declaration. No echo checkout mutation.
- Planning through the writing-plans skill because policy must cover CLI selection, CI admission and per-prefix train execution. New helper is integration glue, not a fake pure entity; reuse effectiveLocalChecks/defaultSuiteCheck and existing bounded probes (ARCH-DRY, ARCH-PURE). No implementation code changed yet; durable plan awaits hub review before change-code.
- Native read-only incident probe: materialized the exact historical Makefile bytes in private scratch, used the production make database/weight goal with WF_TEST_LOCK=off (never a test goal). Planning-era main exited 0 with has_test=false, weight=1; the tested echo head exited 0 with has_test=true, weight=1. This confirms the incident is the approved absent-main-target case, not merely a missing-declaration assumption.
- Hub approved the committed full-flow plan and isolated real CLI/landing fixtures as sufficient live proof, no :0 touch. Initial unpinned change-code inferred quick from design size; corrected with the hub-approved --flow full before implementation. Plan-quality accepted with no blocking findings in rounds 1 and 2; PQ-1 Minor requested function-named test strategies and no duplicate case inventories. Compressed those mechanically under the plan's Revisions; no scope/behavior delta. Estimate 1.27h derived only after plan-quality acceptance, set via the card verb and accepted by estimate-quality with advisory optimism. Module allowances include writing/debugging regressions and active verification; queue waits alone are excluded. Full flow is now recorded and implementation admitted.

- 2026-10-22 01:49 PDT — TDD: captured byte-preserving echo#1 inputs; native CLI regression failed with the original origin/main-declares-no-local-checks refusal. Native CI controls failed on missing bootstrap default and ignored broken head probe. Train regressions failed on empty policy and attempted empty main baseline. Shared derivation, pushed-subject-first CLI, CI facts and per-prefix train selection implemented. Focused bootstrap/default/probe controls passed (cmd/sdlc, 16.015s), including native managed execution/status, fallback cancellation and unverified bootstrap baseline. Broader verification follows. ARCH-DRY keeps constructors/probes/record machinery shared; no head custom policy or server-CI waiver.

- Implementation verification: broad landing/train/CI controls ran to completion (457.446s); only immutable-probe count expectations and a canceled-fallback seam failed. Updated the count to one main plus one subject probe cached across polling, and added the between-probe context guard. Native immutable failure reproduced false main-red diagnosis before the fix; now it preserves the original failure and reports unavailable/unverified baseline. Recovery red confirmed main-only suite refusal for native known-flake republish and original-SHA resume; fixed every current-policy reader through one shared derivation. The fixture was corrected to emit structured Go evidence, preserving resume's completeness guard.
- Final targeted bootstrap/default/train/probe/baseline/resume controls passed (cmd/sdlc, 35.246s); golint passed (0.894s), shelllint passed (0.259s). Compiled local-check help inspected. Isolated native CLI, exact-head status, train and recovery fixtures satisfy the hub-approved live gate, with no :0 touch. Full recorded suite follows on the pushed implementation head.

## Revisions

- 2026-10-22 01:21 PDT — Hub selected the default-suite alternative in the original Spec: only absent declaration plus absent main test permits exact pushed-head default bootstrap. Head-declared custom checks wait until that declaration is on main; no head policy replaces/lightens the default or stands in for CI. Added explicit acceptance controls and the requested echo incident replay.

- 2026-10-22 01:49 PDT — Hub confirmed existing republish/resume readers belong to this same mechanism: use saved/original SHA with shared main-first policy, never current HEAD/head declarations. No new state, command or authority. Integration review also found immutable baseline diagnosis must retain bootstrap provenance rather than call absent-target main red.
