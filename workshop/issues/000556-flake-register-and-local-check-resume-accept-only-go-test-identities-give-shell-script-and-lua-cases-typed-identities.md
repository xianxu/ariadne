---
id: 000556
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours: 6.18
collaboration-mode: collaboration
card_mirror: '5012db0f94a1090d2588e0ce77aa69ce1b1254d1' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T13:11:41-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
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
- Peer adoption follows ariadne#558, which separately distributes the Plenary producer through the operator-verified base-manifest row. #556 ships the implementation/source and private adoption conformance without editing local-gates configuration.

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

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. The calibration source is flagged stale by `sdlc estimate-source`; values are provisional. Scope is one Ariadne change plus private native conformance and adoption documentation; occupied peer branches and live gate declarations are outside this implementation.

Decomposition: issue/spec shaping; typed identities; runner declaration/wrapper; unit evidence; log projection; attempt schema; saved-proof consumer integration; focused Lua producer; private CLI/provider acceptance; lint guard; docs; one close review; native runner discovery. Code design values use the thorough-spec ×0.2 discount, implementation values are already scaled to 40% of the v2 ranges, familiarity is 1.0, and design buffer is 15%. Library availability: the existing managed-process/strict cleanup helper halves runner design from 0.20 to 0.10; JSON/CUE/tracker and provider seams are reused. No existing library supplies our bound unit-proof protocol. Plenary's existing completed result table removes the need for a new assertion runner. The integration and Lua rows use the upper implementation range for real native and process-death conformance.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.80 impl=0.08
item: smaller-go-module design=0.05 impl=0.20
item: greenfield-go-module design=0.10 impl=0.32
item: greenfield-go-module design=0.30 impl=0.32
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: cross-cutting-refactor design=0.20 impl=0.20
item: lua-neovim design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: smaller-go-module design=0.06 impl=0.12
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.00 impl=0.20
item: real-api-discovery design=0.00 impl=0.24
design-buffer: 0.15
total: 6.18
```

## Log

### 2026-10-22

Claimed and entered planning on the issue branch. Read tl-worker/SDLC contracts, brainstorming and writing-plans skills, the issue vocabulary, registry/classifier/rerun and saved-resume consumers. Read-only delegated surveys recovered the actual peer evidence and the persisted-format consumer inventory.

Pair#773's retained suite log records `bash tests/nvim-route-poke-admission-test.sh` followed by `run_headless: TIMEOUT after 30s` and make error 124; this identifies the script, not a Lua assertion. Its failed attempt was unattributable under the old Go-only mechanism. Parley's #410/#412 evidence names `session-sync native edits retries refused admission with native o entry (#408 retry)` in `tests/integration/session_sync_native_spec.lua`. The copied/edited isolated pass is not same-head evidence. Raw historical records will remain unchanged.

ARCH-DRY/PURE: one typed identity, coverage and unit-proof model must feed registration, live reruns and all saved readers. ARCH-EVOLVE: attempt v3 is already reserved for audited could-not-run corrections, so the draft uses v4 for typed-capable checks and a declaration version barrier. Existing resume links, parent locks, immutable interrupted receipts and option-A cleanup stay unchanged. No tests or production code edits have run for #556.

Checkpoint: durable plan drafted; next step is ops:0 review of the explicit runner/producer contract, peer adoption and private conformance gate. Wait for approval before `sdlc change-code`.

### 2026-10-22 — TL approval

ops:0 approved the full plan: unchanged Go identities, script path and Lua spec/case identities, whole-spec Lua reruns requiring both case and file proof, wrapper-owned script exit, Plenary's separate result channel, v4 attempts, immutable old records, and private production CLI plus real shell/Neovim/Plenary conformance with no stage (Couch delivery `995542a8-742e-4df9-9be0-28058c64cf55`). Proceeding to the full-flow plan-quality gate before production edits.

### 2026-10-22 — plan-quality round 1

Baseline `go test ./cmd/sdlc/internal/flakes ./cmd/sdlc/internal/issue -count=1` passed. Preserved the pre-change binary at `/Volumes/TestScratch/tmp/ariadne-slot3/ariadne-556-prechange.DRxOiQ/sdlc`; `go version -m` verifies revision `a3f81c32ae131786fb4aedc9eb3c080dc0a6330f` and `vcs.modified=false`.

The full-flow gate blocked on PQ-1 (nested process ownership) and PQ-2 (function-level test strategies), with PQ-3 (native conformance cadence) Minor. Verified PQ-1 against `runManagedCommandBatch` and `processgroup.Configure`: nested reuse would indeed create a separate group. Revised the plan so only the outer helper creates/joins the group; the unit child inherits it, and typed-capable checks use the existing strict group-cleanup path. The wrapper requires managed invocation, with a normal direct-runner recipe fallback outside local-checks. Added synchronized wrapper-death coverage and a planned lint rule. Function-named strategies and native conformance repeat ownership address PQ-2/PQ-3. No production code edited; ownership delta sent to TL before rerunning admission.

ops:0 approved the ownership clarification, existing StrictGroupCleanup, wrapper-death test and lint rule (delivery `dee4f241-4334-488e-bbed-ab6a0232dc2b`); rerunning change-code.

Round 2 plan-quality returned CLEAN, disposing PQ-1/PQ-2/PQ-3 as addressed, with no new findings. The remaining admission gate requires the derived estimate above and card setter; no production edits yet.

### 2026-10-22 — implementation admitted; bounded cores and adapter checkpoint

`sdlc change-code --issue 556 --flow full` completed successfully after the estimate setter. Plan-quality round 3 remained CLEAN; estimate-quality INFO accepted the derivation and noted that compatibility/conformance and recorded suite/landing time may exceed the primitive allowances. The issue/spec row includes whole-issue shaping; schema, log and acceptance rows absorb their own tests, while the consumer row covers shared wiring. This is a provisional estimate, not a cap or grounds to omit any gate.

Implemented and verified the immutable typed identity model, exact-only non-Go coverage, literal runner argv projection with opaque JSON preservation, strict unit/summary decoding and the sticky unit lifecycle. Unit-evidence fuzzing and 8/16 growth tests defend malformed frames, ordering and lookup work. Existing Go coverage and emitted-name rules remain intact. The Plenary producer preserves native formatting/hooks/assertions and writes a bounded private channel; native pass/fail and setup-error behavior was compared with the installed library.

The v3 declaration parser and `flake run --unit` wrapper pass targeted integration tests, including a literal script path with shell metacharacters, observed exit 124, stdout impostors, a required private Lua summary, contract mismatch, replaced result-channel symlink and typed display. Red failures were observed before each implementation slice. Verified commands: `go test ./cmd/sdlc/internal/flakes ./cmd/sdlc/internal/issue -count=1`; native producer `SDLC_PLENARY_DIR=/Users/xianxu/.local/share/nvim/lazy/plenary.nvim go test ./scripts/test -run PlenaryEvents -count=1`; root targeted `go test ./cmd/sdlc -run 'TestFlakeRun|TestFlakeUnitSummary|TestLocalChecksV3|TestParseLocalChecks' -count=1 -timeout 10m`.

Outstanding: bind the outer managed run's trusted contract/environment and strict lifetime, add the wrapper-death/lint guards, connect typed log facts and v4 records to live/resumed/saved proof and registration, then private full CLI acceptance and recorded verification/close/landing. This is not yet a usable end-to-end typed recorded check. The planned producer distribution row is pending: `construct/base.manifest` is local-gates protected; ops:0 was notified before editing it (receipt `47731f88-5148-4a8d-b064-ff83aeab8371`), with a choice to split rollout if desired. No live `.sdlc` or seeded workflow was edited.

ops:0 directed the split (delivery `8556cb2f-b354-48e4-ad56-4a99963da64e`). Filed unclaimed ariadne#558 on origin, depending on #556, for the one `scripts/sdlc-plenary-events.lua` symlink row and operator-held local-gates verification. #556's Done-when now records that peer adoption follows #558; this branch leaves `construct/base.manifest` unchanged.

Checkpoint verification also includes the pure mixed-kind rerun planner/proof dispatcher and the curated `unit-child-process-group` lint rule, including single-command wrapper aliases. `go test ./pkg/golint -count=1` passed after atlas catalog documentation; fresh pure flakes/issue and explicit installed-Plenary tests passed. No full/recorded suite or close review has run yet. All delegated bounded files are stable, and no own test process remains running at this checkpoint.
