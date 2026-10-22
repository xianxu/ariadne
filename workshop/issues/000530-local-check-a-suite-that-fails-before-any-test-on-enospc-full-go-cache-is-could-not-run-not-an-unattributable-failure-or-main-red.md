---
id: 000530
status: codecomplete
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours: 3.7
collaboration-mode: collaboration
card_mirror: 'f67cf59356db18144ec1dfeda56f00c7636f43d0' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T14:22:07-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 1.63
---

# local-check: a suite that fails before any test on ENOSPC (full Go cache) is could-not-run, not an unattributable failure or main-red

## Problem

## Spec

10-21 ~12:40 the shared Go build cache `/Volumes/AgentScratch/go-build` (182 GB) filled its volume (186 GB, 100 MiB
free). Every recorded local-check suite on the machine then failed before running a single test, writing the cache
(`no space left on device`, typically in pair's runtimebundle generation): pair#768 (train 238 baseline), #771, #774
(integration + diagnostic main), #775, ariadne#526, #528, among others. sdlc classified these as **unattributable
failures** on the exact head, not as the narrow "could not run" (infra) class it uses for TMPDIR problems. Slots now
expect the same-head unexplained-failure guard to resist a plain rerun once space is restored, and train 238's
baseline turned a main-red verdict out of it.

Fix:
1. Classify a suite that fails **before any test ran** with a recognised infrastructure error (ENOSPC/EDQUOT on the
   Go cache, TMPDIR or the build output; also missing toolchain) as **could-not-run**, same as the TMPDIR class: no
   failure evidence, no main-red verdict, no same-head guard; the message names the full volume and path.
2. A main-baseline diagnostic that could not run must not mark main red or hold/split a train.
3. A supported recovery for runs already recorded as unattributable under this signature (one explicit
   `local-check` reclassification by the owner or :0, logged), never a flake registration.

### Proposed design (awaiting hub approval)

Extend the existing strict `suiteEnvironmentLog`/`suiteEnvironmentError` path,
not the flake classifier. A complete bounded recognized startup log produces a
typed cause/resource/path; any test activity, unrelated diagnostic, truncation,
timeout or unknown evidence remains blocking. New infrastructure failures retain
their diagnostic without publishing code failure evidence. Direct execution,
affected, known-flake rerun, main diagnosis and train consumers share this path.

For historical runs, propose `sdlc local-check reclassify --attempt ID --reason
TEXT`: one exact attempt, immutable log/hash proof, same startup classifier,
originating recorded slot or verified repository :0. No force flag, automatic
bulk sweep, flake registration, status rewrite or old-pass republication.
Preserve original record bytes and raw log identity in an immutable audit before
atomically publishing a nonpassing v3 could-not-run disposition. Existing v1/v2
readers remain supported; older binaries encountering v3 refuse safely. The
shared evidence projection prevents legacy resurrection, false flake proof and
saved-pass promotion. Interrupted correction stays blocking until the same
command reconciles it. Audit pruning follows the attempt's existing lifecycle.

Alternatives rejected: silently interpreting old logs changes evidence without
an audit; deleting records or adding an ignored v2 flag leaves old readers or
legacy logs with conflicting authority. The durable plan contains the full
state/action/recovery table, consumer inventory and counted scan/read checks.

## Done when

- A test drives a fake ENOSPC in the Go cache before any test output and gets could-not-run, rerun allowed,
  no main-red; a real test failure after tests start is still a failure.
- The reclassification path exists and is logged; tests cover refusal when tests had already run.
- Recorded live startup output replays in the normal suite for candidate and
  diagnostic main; no false main-red, train hold/split/blame or passing status.
- Only an exact eligible attempt's originating slot or verified repository :0
  can reclassify; mixed/unknown/incomplete output and changed/missing log proof
  refuse. Existing real failures on the head remain blocking.
- Original record/log evidence and actor/time/reason audit survive interruption
  and duplicate recovery. No legacy resurrection, flake proof or old-pass
  promotion; old-reader behavior and unknown-field preservation are exercised.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only; the source is marked stale, so provisional.
Derived after plan-quality cleared. Familiar Go/evidence territory (1.0).

Decomposition: issue/spec; existing startup classifier; new pure recovery
transition; audited store; existing schema extension; cross-cutting evidence
projection; CLI wiring; pinned-reader fixture; docs; one close review. Existing
code mirrors receive the 0.2 spec discount. Recovery transition is repo-specific
authority, with no library substitute. Audited storage reuses queue reconciliation
and standard atomic-file primitives, halving its base design before the 0.2
discount. Implementation values are already the v2 table's 40% values.
Design subtotal 1.48h, implementation subtotal 2.00h; thorough-plan buffer 15%:
1.48 * 1.15 + 2.00 = 3.702h, rounded to 3.70h.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.75 impl=0.08
item: smaller-go-module design=0.06 impl=0.20
item: greenfield-go-module design=0.20 impl=0.32
item: greenfield-go-module design=0.15 impl=0.32
item: smaller-go-module design=0.06 impl=0.20
item: cross-cutting-refactor design=0.10 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: atlas-docs design=0.04 impl=0.08
item: milestone-review design=0.00 impl=0.20
design-buffer: 0.15
total: 3.70
```

## Plan

Full flow, one atomic close boundary, plan approval required before code.
Durable plan: [suite startup evidence](../plans/000530-suite-startup-evidence-plan.md).

- [x] Replay the recorded startup failure; extend the strict live execution path
  with positive/negative controls and candidate/main/train coverage.
- [x] Implement exact-attempt, owner/:0, audited recovery and versioned evidence
  projection; exercise interruptions, consumers, old readers and retention.
- [x] Document the contract, exercise focused verification and queue the recorded
  full suite on the pushed implementation head; prepare the close evidence.

The recorded suite pass, close review, landing and hub outcome report remain
mandatory subsequent workflow gates; they are tracked in the Log below.

## Log

### 2026-10-21
- 2026-10-21: closed — BR-1 actual plan locations cross-checked and corrected. BR-2 executed-head distinct-from-caller regressions red then PASS across suite/affected, rerun, baseline and missing-Go; startup/environment plus actual recovery-catalog guard PASS 9.529s. Built private CLI, recorded Go-cache capture, audited nonpassing v3 recovery, original-digest/interruption/retention and real-code-failure controls previously PASS; actual #524 reader compatibility and fuzz PASS. First recorded suite completed all 8 shards/1982 SDLC tests/75 packages and exposed the missing recovery catalog contract; that branch defect is fixed, not relabelled. New recorded suite queued on pushed 99c89cf3f2181f380fafd754afd28eb65e10674d; final pass required before landing under authorized review overlap. TL decided no :0 stage.; review verdict: SHIP

- Claimed after #508 landed. TL explicitly requested plan-first gate-evidence
  work. No production changes or tests have been written for #530.
- Re-read the actual #508 retained ENOSPC log: Go list failed while writing
  /Volumes/AgentScratch/go-build, followed by the complete Python go-list
  traceback and make exit. The existing classifier accepts only a work-dir
  creation diagnostic and an older go-test-compile traceback, so this output
  falls through to unattributable failure. Baseline consumers then mistake
  an unparsed failed run for main-red. This establishes the live root cause.
- Read-only delegated mapping confirmed that same-head suite/affected guards,
  saved pass reuse, flake/stats projections and legacy per-head logs all consume
  the persisted attempt. A v2 flag would be ignored by older readers. The #524
  future-version identity envelope and queue reconciliation audit pattern are
  the existing seams to reuse (ARCH-DRY, ARCH-EVOLVE, ARCH-ORDER).
- Drafted the durable plan with explicit state/recovery and compatibility
  coverage. No estimate before plan-quality, no change-code before hub approval.
- TL approved the committed design, command/ownership, audited nonpassing v3
  disposition, genuine rerun and strict refusal controls. Explicit added proof:
  a real #524-era reader must scope the unsupported v3 record to its subject,
  rather than poisoning unrelated evidence; fleet shared sdlc includes #524.
  Proceeding to the full-flow change-code plan gate before implementation.
- Plan gate round 1 raised PQ-1 (Important): name risky functions and replace
  test-case inventories with adversarial strategy/mechanical-guard contracts.
  Added those function contracts and compressed the task inventories; no design
  mechanism or scope changed. Re-running plan-quality before implementation.
- Plan-quality cleared after the function-contract correction. Change-code
  admitted full-flow implementation after the derived 3.70h estimate. Its INFO
  advises tight verification allowances: parser/run-path tests belong to the
  startup module, schema/proof preservation to the decoder module, interruption
  and pruning to the store, and real old-reader/full-run checks to the compatibility
  fixture allocation; the one review allowance is not a second review boundary.
  These allocations do not waive any required verification. Calibration remains
  provisional; measured actuals will record the real cost.
- Task 1 red: the verbatim #508 candidate log is now a replay fixture. New
  startup grammar, native recorded-run and baseline/train controls fail on the
  old implementation (targeted package 2.878s): the capture persists/posts
  unattributable failure, and its main replay reports main-red. No volume was
  filled; fixture bytes match the retained raw capture.
- Task 1 replay and existing environment controls pass after the strict grammar
  extension (5.664s). A native missing-Go fixture then exposed a separate launch
  boundary: Exit=-1 with an empty log is still recorded as unattributable (red,
  targeted package 1.513s). Proposed a positive-proof first-Go-start exception
  in the plan and asked the TL before implementing it. No empty historical log
  will be reclassified by inference.

- TL approved the narrow missing-Go delta (receipt
  54589b6d-ecf9-455d-87e9-af3ee82a75e7). The first Go command now carries typed
  missing-executable/interpreter proof only when Process is nil, through normal
  queue and cleanup completion. Empty raw output is retained; no status or failed
  attempt is published. Earlier-command, unrelated executable, permission and
  started/exit-127 controls preserve ordinary failure. Native missing-Go and
  batch controls were red (0.665s) before implementation. Startup/environment
  and managed-batch controls now pass (19.611s), including lease/checkout cleanup.
  No full recorded suite has run yet. Next: exact-attempt audited v3 recovery,
  consumer sequence controls and actual pinned #524 reader compatibility.

- Recovery core checkpoint: exact-attempt eligibility red (0.379s), audited
  publication red (0.376s), neutral flake-evidence rule red (0.128s), old-pass
  promotion red (0.381s), and CLI lock metadata red (0.437s) preceded their
  implementations. The v3 correction preserves original JSON facts/unknown
  values and immutable raw log, commits an immutable audit first, and validates
  that proof in the existing scoped snapshot. Ordinary writers remain v2.
  Verified audit time is projected into that snapshot to reject every saved
  pass that started before correction; only a genuine later run is reusable.
  No GitHub effect exists in reclassification. Atomic-publication interruption,
  audit/log/record mutation and exact 8/16 read-count controls pass (0.726s).
  Combined decoder/scoped-proof/neutral-rule/lock controls pass (0.867s + 0.059s).
  Remaining: stronger CLI/native ownership and consumer sequence coverage,
  pinned real #524-era compatibility, pruning/unknown-audit/schema-fuzz controls,
  docs, full recorded suite and the one close boundary. No completion claim.

- Final focused controls pass (11.052s), including recorded cache replay on
  suite/affected, known-flake rerun, baseline/train, existing code-failure guards,
  owner/:0 CLI, exact-head recovery advice and saved-proof consumers. Broader
  local-check/suite regression passed (72.859s); race-checked recovery and
  pruning passed (3.792s). Internal flakes and Go/shell lint packages pass.
- The built real CLI passes the private repository fixture (1.490s): correction
  changes no GitHub status and makes no GitHub call; genuine repaired execution
  can publish, while a separate code failure still blocks. Later advice control
  now names the exact recorded check/head. No live fleet/store was modified.
- Actual encoder bytes passed the pinned landed #524 reader control (0.549s):
  related implicit/explicit saved proof and post guards refuse with upgrade
  guidance; unrelated proof posts successfully; opaque v3 bytes remain unchanged.
  Initial compatibility fixture setup used an unknown fake-GitHub SHA and was
  corrected to the private repository's real head; this was fixture setup, not
  product evidence. The earlier pre-524 control remains separate.
- Startup fuzz passed (4,416 executions); schema/opaque-field encoder fuzz
  passed (125,968 executions). Iterator and proof-open count guards remain exact
  at N=8 and 16. No lexical lint was added: startup grammar and recovery authority
  are semantic/sequence properties, defended by these independent controls.
- Two extra red invariants were fixed: a conflicting corrected head now makes
  the read incomplete rather than concealing original authority (red 0.482s);
  interrupted retry requires the exact retained raw original digest, including
  whitespace (red 0.449s). No approval mechanism changed.
- Bounded lifecycle exploration found audit residue after interrupted pruning
  or publication racing an expired original. Red tests (0.761s) preceded the
  existing-pruner extension: failed authority unlink retains proof; canonical
  orphan audits get the existing 30-day committed-audit grace and bounded scan.
  Private partial writes still use 24-hour dead-owner cleanup. Recovery/pruning
  controls pass (1.576s). Full recorded suite, smoke decision and close remain.

- TL smoke decision confirmed (receipt 7bd0bb95-e1ec-4c45-af27-c6267a41b649):
  no :0 stage. Private built-CLI, captured consumer replay and #524 compatibility
  are sufficient; live proof is the first real full-disk/missing-Go incident
  after landing. Recorded suite is retained on pushed implementation head
  f56c83f606694c08b1ec78e68185cc1ae822a0dd, lease-1942744462, terminal 13109;
  initial normal FIFO position 10. No duplicate runner or per-run cache override.
  Boundary review may overlap the queued run under the current tl-worker rule;
  passing final-head evidence still precedes landing. No suite-pass claim yet.

- Close round 1 returned REWORK: BR-1 Critical is the plan's wrong location for
  validateReclassificationProof; BR-2 Important is live recovery advice omitting
  the executed revision. Corrected the entity/task paths and cross-checked the
  class. New caller-HEAD-advanced regressions were red (2.431s), then pass across
  direct suite/affected, both parent-check reruns, baseline and unstarted Go.
  Retention and typed launch errors carry the executed head; rerun advice names
  the originating suite/affected check, never an internal record name.
- The first recorded suite finished before the cancellation request, so the
  exact-run cancel safely refused as unavailable; no runner remained or was
  duplicated. It ran all eight shards over 1,982 SDLC tests and 75 other packages
  in 758.5s after 3,909s FIFO wait. The sole named failure was the new command's
  missing recovery catalog entry, TestRecoveryContractsAreProven (plus its package
  failure aggregate). This is a branch defect, not infrastructure or a flake;
  no resume, blind retry or evidence reclassification was used. Added the actual
  convergent-retry contract and proof bindings, plus README discovery of the
  pinned-reader script and its invoker-owned scratch lifecycle.
- Startup/environment and recovery-catalog regression now passes (9.529s).
  Review lessons record exact-subject propagation, plan-location cross-checks and
  the existing mutating-verb catalog tripwire. Fresh recorded suite and full
  close round 2 are still required for these code changes; smoke decision remains
  no :0 stage. The earlier run is not final passing evidence.

- Close round 2: SHIP, zero new findings, BR-1 and BR-2 disposed, zero open
  blocking findings. Card is codecomplete, bound to evidence commit
  e75169873c630fd95313fe81565e488e3d836e1d; mirrored close head was
  c1ae80512572291e5f660dce237a3cb1694f5c68. PR is
  https://github.com/xianxu/ariadne/pull/379 and its server local-gates guard passed.
- Retained final branch suite: implementation head
  99c89cf3f2181f380fafd754afd28eb65e10674d, lease-1870387321, terminal 71348,
  attempt 596fc2146a83d0444c641809908f9f7c. Its inflight log is under the common
  Git directory's sdlc/test-stats/inflight/596fc2146a83d0444c641809908f9f7c/run.log.
  It remains normal FIFO (last observed position 5), not passing evidence yet.
  The difference from implementation head to close head contains only issue and
  review records. Preserve this run; wait for its pass before starting merge so
  pending close-only equivalence cannot create a duplicate suite (#502 lesson).
  No stage, private cache or completed private CLI/compatibility scratch remains.

- Final branch suite passed on 99c89cf3f2181f380fafd754afd28eb65e10674d:
  attempt 596fc2146a83d0444c641809908f9f7c, all eight shards, 1,982 SDLC tests
  and 75 other packages, zero failures, wall 724.6s after 3,465s FIFO wait.
  Raw log is sdlc/local-checks/99c89cf3f2181f380fafd754afd28eb65e10674d/suite.log
  in the common Git directory; reported digest prefix 277132873062. The retained
  runner survived the Couch restart; no duplicate run was started. Recorded as
  sdlc-local/suite on that implementation head. Close round 2 is already SHIP;
  enter merge now and let its turn own current-main integration and verification.

- Landing reached its turn, then refused code conflicts against main
  660157583b200d7b488ace2fd5a6a21d0c2249dc in localcheck_cmd.go,
  localcheck_republish.go and teststats_store.go. Followed its prescribed manual
  integration; no gate waiver. Preserve newly landed #463 resume/reconcile
  commands, composite validation, raw-record digest indexing and directory /
  nonblocking-parent pruning locks alongside reclassification and checked audit
  deletion. Both deterministic pruning seams delegate to one implementation.
- Integration regressions: a completed resume composite cannot predate a
  separate same-head corrected startup failure, either through implicit/explicit
  reuse, cached status posting or resume's completed-proof repost. A fresh full
  suite remains reusable, while a pending resume fence still refuses it.
  Publication also shares the existing attempts-directory lock: without it,
  pruning between the exact final reread and rename could resurrect v3 authority
  after deleting its proof. Eligible startup corrections and valid named-failure
  resume parents are disjoint, so no additional parent lock is needed here
  (ARCH-DRY, one existing directory mutation boundary).
- Both integration regressions were red (0.547s) before correction; the focused
  resume/reclassification/process-lock/recovery controls then passed (7.016s).
  Startup/scoped-proof controls and race checks also passed. The original branch
  suite remains valid historical evidence, but this code integration requires a
  new recorded suite and the merge-prescribed narrow re-close with
  --fixed-then-ship and --full-suite before landing.

- Retained integrated-code suite: pushed head
  0181ef043a90f5807118d663c26e437bb46cf687, attempt
  767c1bb9ca999987696c7fbbda27762a, lease-225718307, terminal 97786,
  detached checkout /Volumes/TestScratch/local-checks/sdlc-local-603910783.
  Inflight raw log is under the common Git directory at
  sdlc/test-stats/inflight/767c1bb9ca999987696c7fbbda27762a/run.log.
  Last observed FIFO position 4. Keep this runner; no live merge runner remains
  after the conflict refusal. Wait for its pass, then narrow re-close with the
  actual --full-suite proof, push the new close generation and re-enter merge.

## Revisions

- 2026-10-21 — Added the proposed design and explicit acceptance controls after
  tracing the outage through existing consumers. Original requirements remain
  above; ownership, atomic audit and compatibility details await hub approval.
- 2026-10-21 — Clarified the last procedural checkbox for pre-close readiness.
  Its original sequence was "Document, verify the full recorded suite, close
  review and land through sdlc; report the outcome to the hub." Those gates remain
  mandatory after implementation readiness, with pending/result evidence in Log;
  no requirement or Done-when criterion changed and no gate is waived.
- 2026-10-21 — Integration with landed #463 preserves the same genuine-rerun
  contract across its composite reuse/repost paths, and uses its existing
  attempts-directory mutation lock for audit publication and retention. No new
  store or authority state; the required integration suite and narrow re-close
  replace the superseded pre-integration landing attempt.
