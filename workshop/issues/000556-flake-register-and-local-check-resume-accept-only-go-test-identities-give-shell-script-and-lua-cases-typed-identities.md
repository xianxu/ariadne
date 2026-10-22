---
id: 000556
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '22cbaefd627f7eb539d377bc10dfd7ec62823c10' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T13:11:41-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# flake register and local-check resume accept only Go test identities; give shell-script and Lua cases typed identities

## Problem

## Spec

`sdlc flake register` accepts only Go `PKG.TestName` identities, so failures in non-Go suites can't be registered or
resumed:
- pair#773 (10-22): `tests/nvim-route-poke-admission-test.sh` timed out at 30s; the suite classified it
  "unattributable-failure" (no Go test identity), so `local-check resume` couldn't apply and the slot had to rerun the
  whole suite. The script is tracked as pair#685.
- parley.nvim#410 (10-22): a #408 native Lua case (o-retry, initial `modifiable=false` assertion) failed once and passed
  on an isolated rerun; it can't be registered at all.

Fix: a typed test identity for non-Go suites: shell scripts by path (`script:tests/x.sh`) and Lua or busted cases by
spec file plus case name (`lua:<spec>#<case>`), produced by the existing suite runners from their own output, so a
script timeout or a named Lua case failure gets an identity the flake registry and `local-check resume` can use.
Same evidence rules as Go flakes: a saved same-head fail-then-pass.

## Done when

- A script timeout in a recorded suite is reported with a `script:` identity (not unattributable), and can be
  registered and resumed (test).
- A failing busted case reports a `lua:` identity that registration accepts (test).

## Plan

Durable design: [typed identities plan](../plans/000556-typed-flake-identities-plan.md).

- [ ] Freeze the identity, runner/producer and compatibility contract with ops:0 before code.
- [ ] Implement one type-aware identity and coverage model across registry and evidence consumers.
- [ ] Add trusted unit runner declarations and positive script/Lua completion evidence.
- [ ] Connect recorded failures, exact-unit retries, saved resume, republish and guards; preserve legacy Go proof.
- [ ] Prove both motivating shapes with private production CLI, real shell and Plenary conformance; document adoption.
- [ ] Recorded verification, issue-close review, land and report to hub.

## Revisions

### 2026-10-22 — planning proposal, awaiting TL approval

The Lua case is Plenary and the canonical harness selects whole spec files, so rerun the owning file and require the exact named case plus whole-file success. Script identity covers the script only. Neither case permits manufacturing a Go name or treating diagnostic text/exit zero as named-case proof. The proposed runner/producer contract and compatibility barriers are detailed in the durable plan; no implementation is authorized yet.

## Log

### 2026-10-22

Claimed and entered planning on the issue branch. Read tl-worker/SDLC contracts, brainstorming and writing-plans skills, the issue vocabulary, registry/classifier/rerun and saved-resume consumers. Read-only delegated surveys recovered the actual peer evidence and the persisted-format consumer inventory.

Pair#773's retained suite log records `bash tests/nvim-route-poke-admission-test.sh` followed by `run_headless: TIMEOUT after 30s` and make error 124; this identifies the script, not a Lua assertion. Its failed attempt was unattributable under the old Go-only mechanism. Parley's #410/#412 evidence names `session-sync native edits retries refused admission with native o entry (#408 retry)` in `tests/integration/session_sync_native_spec.lua`. The copied/edited isolated pass is not same-head evidence. Raw historical records will remain unchanged.

ARCH-DRY/PURE: one typed identity, coverage and unit-proof model must feed registration, live reruns and all saved readers. ARCH-EVOLVE: attempt v3 is already reserved for audited could-not-run corrections, so the draft uses v4 for typed-capable checks and a declaration version barrier. Existing resume links, parent locks, immutable interrupted receipts and option-A cleanup stay unchanged. No tests or production code edits have run for #556.

Checkpoint: durable plan drafted; next step is ops:0 review of the explicit runner/producer contract, peer adoption and private conformance gate. Wait for approval before `sdlc change-code`.
