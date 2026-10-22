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

- [x] Freeze the identity, runner/producer and compatibility contract with ops:0 before code.
- [x] Implement one type-aware identity and coverage model across registry and evidence consumers.
- [x] Add trusted unit runner declarations and positive script/Lua completion evidence.
- [x] Connect recorded failures, exact-unit retries, saved resume, republish and guards; preserve legacy Go proof.
- [x] Prove both motivating shapes with private production CLI, real shell and Plenary conformance; document adoption.
- [x] Prepare and queue recorded verification; enter issue-close review with the implementation and private acceptance evidence complete.

After the close verdict, the recorded suite must pass before landing; any review code fix requires fresh verification. Land through sdlc merge and report the PR/outcome to ops:0.

## Revisions

### 2026-10-22 — planning proposal, awaiting TL approval

The Lua case is Plenary and the canonical harness selects whole spec files, so rerun the owning file and require the exact named case plus whole-file success. Script identity covers the script only. Neither case permits manufacturing a Go name or treating diagnostic text/exit zero as named-case proof. The proposed runner/producer contract and compatibility barriers are detailed in the durable plan; no implementation is authorized yet.

### 2026-10-22 — review corrections and adoption bookkeeping

Retain the approved complete-suite requirement for typed registration; BR-1 corrects its lost targeted-check scope without changing Go confirmation. BR-2 reconciles the durable Core concepts inventory with delivered symbols and their actual PURE/INTEGRATION boundaries; BR-3 adds README user-surface and adoption guidance. The final issue Plan row records preparation/queueing and entry to review; its post-verdict recorded-pass/landing/report obligations remain explicit. The TL-directed #558 distribution split is unchanged; unclaimed pair#816 and parley.nvim#415 now capture the separate adoptions, each depending on #556 and #558.

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

### 2026-10-22 — typed live and saved proof integrated; native conformance passed

V4 reads/writes now bind typed-capable runs to their verified runner declaration; legacy Go checks keep v2. The outer runner strips stale unit authority, verifies the exact subject declaration, shares strict descendant cleanup/lifetime and protects ordinary/automatic-child inputs with existing checkout markers. The synchronized wrapper-death test proves one inherited group and no retirement before join. Typed log facts, exact-unit batches, registration/list/stats hydration, manual resume, automatic-child saved proof, republish and status guards now share the typed model. Archive retention checks compatible v4 parent/child contracts; no retirement ordering, locks, links, receipt authority or automatic retry policy changed. Train package-impact attribution stays unverified for typed failures, while baseline diagnosis uses shared facts.

Observed RED before fixes: saved loaders trusted malformed/zero-exit typed passes; automatic uncertain cleanup deleted child inflight output; mismatched typed parent/child contracts could retire evidence; a rejected outer-zero attempt's log could be overwritten by a later pass. Their focused regressions now pass. Native production CLI acceptance also exposed ordinary marker release incorrectly running for manual resume; only the ordinary caller now grants that release, leaving manual resume's existing owner marker for its own terminal boundary. This is an ownership-wiring correction, not a change to option A. Ordinary markers are released only after group/scratch absence is confirmed; subsequent canceled ordinary residue follows the existing temporary/inflight collector boundary. Linked/interrupted resume graphs retain the existing indefinite pinning policy.

Fresh pure flakes/issue tests passed. Root combined verification passed the typed parser, identity evidence, binding, schema, saved proof, resume core/legacy controls, wrapper death and future-scoping tests; native/frozen follow-up passed in 100.413s with SDLC_PRIVATE_TYPED_CONFORMANCE=1, the installed SDLC_PLENARY_DIR and SDLC_PRE_TYPED_BIN pointing to the previously verified frozen binary. Native variants execute mixed Go, a real script timeout (observed 124) and native Plenary hooks/assertions; register all three from exact-head fail/pass; resume only the selected package/files; reject an omitted Lua case; repeat without launching; and crash after the existing lock effect, then reconcile without destructive effects or success. Old production binary refuses the related v4 suite without a status, permits unrelated lint, and preserves raw bytes. The first old-reader fixture attempts lacked the CLI's repository identity/duration; those fixture errors were corrected before the passing measurement.

`go test ./pkg/golint ./pkg/vocab ./scripts/test -count=1` passed; scripts/test completed in 43.385s. Actual motivating diagnostic excerpts are retained under cmd/sdlc/testdata/typed-flakes; a normal replay test ensures historical text cannot manufacture typed authority. Help/atlas document explicit declaration/producer adoption and the #558 boundary. No construct/base.manifest, live .sdlc or seeded workflow changes. #558 report receipt b82dadae-4705-42b3-a799-ea1f05ab28fa was verified submitted; the later Couch restart requires no resend.

Checkpoint next step: run the existing private legacy-Go resume conformance on this new reader, then push and run the declared recorded affected/suite check. No own recorded check is queued/running yet. Complete the issue-close review, fix/dispute blockers, then PR/land and report the outcome to ops:0. All bounded agents are stable; no own test process remains running.

### 2026-10-22 — legacy compatibility and recorded queue

`SDLC_PRIVATE_RESUME_CONFORMANCE=1 go test ./cmd/sdlc -run '^TestLocalCheckResumeConformance$' -count=1 -timeout 10m` passed in 84.477s: the new reader still resumes legacy Go evidence, reposts without re-execution and keeps crash/reconcile authority intact. Implementation checkpoint 9b546c7c9 is pushed.

The required recorded suite is queued through `sdlc local-check run --name suite`, run lease-641560774, raw attempt f0be49df0fbad21919b67ad5f615f2e9 on that checkpoint. Last observed admission is normal FIFO position 19, behind the running Ariadne suite. It has not passed. The tl-worker contract explicitly permits the close review to overlap a queued recorded check, provided a final code pass precedes landing and review code changes get fresh verification. Entering that review now; retain this runner and queue position. No continuation/restart while it is queued/running.

### 2026-10-22 — close checklist clarification

The final issue Plan row records preparation/queueing and entry to the close boundary; the existing post-verdict suite/landing/report work is stated immediately afterward. This avoids claiming a future landing or verdict was completed merely to satisfy the pre-close unchecked-item gate. No acceptance, gate or scope changed.

### 2026-10-22 — close round 1 and bounded corrections

The boundary review returned REWORK with BR-1 (Critical, typed targeted-child confirmation), BR-2 (Critical, inaccurate Core concepts inventory) and BR-3 (Important, README coverage). Verified BR-1 through the production SuiteAttempt projection: script and Lua suite-rerun children incorrectly confirmed registration, whole-attempt history and aggregate stats. The new reader-only Targeted projection now reaches the single typed confirmation predicate; affected passes remain excluded and existing Go targeted confirmation remains unchanged. Pure RED/GREEN and root `go test ./cmd/sdlc -run 'TestTypedConfirmation|TestTypedFactsHydration|TestFlake|TestSuiteStats' -count=1 -timeout 10m` passed (33.762s). BR-2/BR-3 corrections are documented in the plan Revisions and README; existing unmodified helpers are no longer claimed as source changes. No finding is disputed or self-cleared; the next boundary review owns disposition.

Canceled queued suite lease-641560774 through the supported exact-run command because BR-1 changes code and needs a new checked head. The queued attempt never ran a suite; its controller ended with context-canceled admission. There is no duplicate recorded runner. Next: native complete-suite registration/resume smoke on this correction, push the verified checkpoint, requeue the required suite and repeat the full close gate.

At ops:0's request, filed pair#816 and parley.nvim#415 unclaimed on origin with both dependencies, exact adoption steps and the new typed-capable same-head fail/pass requirement. Both peer checkouts were left unchanged. Pair :0 was detached for operator staging, so filing used its existing :7 branch; ops:0 subsequently instructed all future cross-repo filings to use a temporary worktree off origin/main. Confirmed :7's status before/after filing matched its three pre-existing edits, then a fresh read showed it clean; confirmation receipt 5d336684-0694-411a-ab89-ef1c78b7daf6 is submitted. No activation or local-gates label was applied.
