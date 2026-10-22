---
id: 000536
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'bafb101e25b0cf8ddf9037ba3b68d5f134ffa633' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T01:19:13-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

## Plan

- [ ] Reproduce the incident and share subject-aware main-first policy between CLI and CI facts, preserving exact pushed-head and cancellation controls.
- [ ] Cover ordinary/immutable landing and train prefixes/reintegration with the same derivation, no weakened gate or fabricated baseline evidence.
- [ ] Document the contract, complete focused/recorded verification and prepare close. Detailed steps: [bootstrap suite plan](../plans/000536-bootstrap-suite-plan.md).

## Log

### 2026-10-21

### 2026-10-22

- Claimed and started from fresh main after #545 landed/reported and hub rebuilt ariadne:0. Existing no-declaration/main-test case is already supported by TestLocalCheckCmd_DefaultSuite; the missing bootstrap case is no main test target. Hub approved exact-subject default only, no head custom declarations or CI substitutions, and requested the real echo regression.
- Verified echo#1 archived Log and native Git objects read-only: aeee435d3953c60cdd1ab4350e206d0e162646f7 and final pre-merge main 77a0fd6320ccc5ff603c9cfefdb8bb4789486556 share Makefile blob 0e078508dd6066bb3ac059e6a5b649e26a568694 and lack tracked Makefile.local, Makefile.workflow and local-check declaration. Tested head e3965c2c6d79ece162e704d07335b5b883fe9668 adds the test target and first make-test declaration. No echo checkout mutation.
- Planning through the writing-plans skill because policy must cover CLI selection, CI admission and per-prefix train execution. New helper is integration glue, not a fake pure entity; reuse effectiveLocalChecks/defaultSuiteCheck and existing bounded probes (ARCH-DRY, ARCH-PURE). No implementation code changed yet; durable plan awaits hub review before change-code.

## Revisions

- 2026-10-22 01:21 PDT — Hub selected the default-suite alternative in the original Spec: only absent declaration plus absent main test permits exact pushed-head default bootstrap. Head-declared custom checks wait until that declaration is on main; no head policy replaces/lightens the default or stands in for CI. Added explicit acceptance controls and the requested echo incident replay.
