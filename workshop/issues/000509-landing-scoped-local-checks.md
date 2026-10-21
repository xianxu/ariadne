---
id: 000509
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours: 3.41
collaboration-mode: collaboration
card_mirror: '1c3301ffed1847860b79f84bad78663b4b64eff0' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T22:27:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# Landing scoped local checks

## Problem

pair#743 moves native conformance from capacity-constrained GitHub macOS runners to a machine-wide Tart runner. The operator requires one conditional conformance run per landing train, after its suite, with nightly burn-in before activation. Current required local checks also gate each car's readiness; declaring conformance required would run it per car. The schema has no scope/path selection and every local run acquires a full-suite slot.

## Spec

- Add a narrow opt-in landing scope and conditional path-list reference to declared local checks, preserving existing per-head defaults and strict config validation. Read authoritative config/path selection from the train's configured main base.
- Exclude landing-scoped evidence from individual-car readiness. Run selected checks once on the exact integration head after a passing suite and before merge, including train-of-one/single-landing fallback. Failure blocks landing without suite bisection or automatic retries.
- Select from the union of each included car's touched paths, including deleted/renamed paths; a later car reverting an earlier touch must not hide eligibility. Recompute for the prefix actually landing and after reintegration. Unavailable path evidence refuses conservatively.
- Reuse existing local-check execution, retained logs and exact-head status/provenance. Code changes invalidate evidence. Explicitly define any docs-only reuse through the existing equivalence proof.
- Permit a VM-backed local check to use its own runner-owned resource instead of consuming/waiting for a suite slot. Do not reinterpret weight (suite slots) as VM capacity or add a second ticket queue. pair owns the two fixed VM positions.
- Leave landing activation off during nightly-only burn-in; an optional failed train status is still red CI, so burn-in must not run on train heads. External scheduling invokes the named check on main; no scheduler is required in SDLC.
- Pair consumes the capability only after this lands; no shared-contract implementation in pair.

## Done when

- Stateful train tests prove post-suite order, exactly one eligible execution per final integration head, no execution for unrelated paths, and no per-car execution.
- Cases cover any-car union/later revert, excluded car/good prefix, train-of-one, failure without bisection, code reintegration, and stale/absent evidence.
- Runner tests prove resource distinction, cancellation/timeout cleanup, retained failure logs, and compatible existing configuration.
- Operator verifies local-gates changes and applies the label; workers never self-label.

## Plan

Durable design: [landing checks plan](../plans/000509-landing-checks-plan.md). Await operator review before implementation and the plan-quality gate.

- [x] M1 — Versioned declaration, pure selection/proof and separate runner-resource execution; adversarial controls and milestone review.
- [x] M2 — Shared post-suite train/single-landing gate, full consumer sweep, stateful/private conformance and milestone review.
- [x] Recorded full suite, accepted private executable proof and final delivery preparation; close and central landing follow through their gates.

## Log

### 2026-10-20

Filed from pair#743 read-only audit. Relevant seams: cmd/sdlc/localchecks.go, localcheck_run.go, landingcar.go productionCarGates, landingtrain.go conductor.build/runSuite/waitCI, cigate.go, internal/landingline/trainstate.go, internal/affected/glob.go and internal/gitx/window.go. Tests: localchecks_test.go, landingtrain_test.go, landingcar_affected_test.go, landingtrain_provenance_test.go, localcheck_signal_unix_test.go. ARCH-DRY: extend the existing post-suite check owner and exact-head runner, rather than wrapping or bypassing landing in pair.


### 2026-10-20 22:45 PDT — Design checkpoint

Claim/start-plan completed on this issue's branch from main. Wrote the durable proposal after auditing local-check codecs, saved evidence, subprocess cancellation, CI/car consumers, pure train transitions and both single-landing shells. ARCH-DRY: one shared post-suite gate and existing execution/log shell; runner-owned lifetime admission remains external, with zero suite allocation or VM lease probing.

Authenticated pair:3 receipt 74652a7c-7ea4-4228-9888-41cf23280c2d confirms make conformance-vm, the direct tests/conformance-paths.json reference, 20m runner budget plus up to10s cleanup, 30m named timeout and at least10s cancellation grace. The proposal uses15s runner grace. No activation or peer workflow changes. A follow-up send was refused with “session listing observations disagree” and no receipt; it was not blindly resent. Read-only inspection confirmed the current path syntax. The contract establishes the intended seam, not completed VM qualification.

Operator tradeoffs are explicit in the plan: v2 unsafe-to-ignore semantics, manual-only disabled landing checks, owner-local evidence rather than unbound saved-status reuse, cleanup grace, terminal check refusal without bisection, and rollout proof/staging. Main-only burn-in excludes --all; selected landing checks still owe real suite proof under --no-ci. No source changes, change-code, test runs or live VM actions during planning. Previous TTY capture remains unavailable; recovery uses the branch, sdlc state and issue evidence.

### 2026-10-20 23:06 PDT — M1 implementation and controls

change-code passed plan-quality (PQ-1 addressed by function-level strategies) and estimate gates. Existing libraries cover matching, Git, groups and retained evidence. The estimate's pure selection/proof items describe deterministic policy, while the seven smaller modules cover codec/IO/wiring; cross-cutting extraction covers resource lifetime separation. Private executable delivery checks are included in resource/IO implementation and final review allowances; suite execution is included in final delivery allowance, provisional under the stale calibration.

Observed RED then GREEN: v2 declaration acceptance/round-trip, path selection, pinned/cache reader, bound proof, zero-helper runner resource, and leader-exits-before-descendant-cleanup. The latter reproduces Pair's qualified public Make boundary (authenticated receipt f1ee074c-7076-498f-9452-d9777c0249de); group cleanup retains30s even with closed output pipes. Broader targeted parser/selection/runner/signal controls passed (19.697s); affected package passed. No VM or production resource was driven. M2 still owns automatic gate/CLI main-only burn-in integration.

M1 fuzzing found a real empty-array round-trip defect: v2 checks [] became null on encode. Kept the minimized seed and initialized the decoded empty slice; rechecking before review. The fuzz failure wrote its seed under the real testdata directory and the hermeticity guard correctly called out that new file; this is intentional retained regression evidence, not a test fixture writing production state.

M1 additional evidence: fuzz round-trip passed32,233 executions/3s after the empty-array fix; narrow race controls passed9.163s. The disposable mutation ignoring all earlier cars failed TestLandingCheckSelection as intended, then was restored. Distinct8/16-car reads and8/16-path matching have counted linear guards; repeated polling adds zero Git reads. Runner cleanup control now waits6s after leader exit (past the old5s bound), and controlled unresponsive-child escalation joins.

### 2026-10-20 23:17 PDT — M1 round1 fixes

BR-1: observed tracked-blank regression for both declaration and path blobs, then rejected present blank policy while preserving true absence. BR-2: nil force signal with still-live group refuses; separately bounded join must confirm completion. Unknown cleanup leaves the existing queue record, raw log and Git-locked checkout; lock spans runner lifetime, and every stale candidate's own metadata is checked across repositories. The cross-repo stale-sweep control reproduced deletion before its fix. BR-3: README now documents delivered M1 fields/selection/resource behavior, explicitly leaving automatic gates pending M2. Added the prevention rule to lessons and appended plan Revisions. Targeted correction controls passed8.900s, broad controls32.410s and narrow race2.100s before strengthening the cross-repository sweep test; rechecking that final change. Pair:3's authenticated c35246b6-e0e0-41a6-b9be-37b05a1d1494 requests landed/schema notification; activation stays held, report owed on landing.

### 2026-10-20 23:21 PDT — Stopped for cleanup re-plan

M1 review round2 disposed BR-1/2/3 but raised Critical BR-4: unconditional scratch cleanup and later owner-PID/age collectors can delete inputs still used by unknown descendants. This is the third cleanup design delta; stopped before source changes. The durable plan now enumerates checkout, scratch, raw log and leases under one lifetime rule, recommending references to the existing Git worktree lock and one shared protection predicate, including log collectors. Runner launch without an attempt writer would refuse, preserving suite legacy fallback. Simplest instance fix and bundle/supervisor alternatives are explained. Approval requested from hub; M1 remains open and M2 is not started. No background work remains.

### 2026-10-21 00:03 PDT — BR-4 protection and confirmed fixture incident
- 2026-10-21: closed M2 — Broad frozen-source CI/train/CLI/legacy/issue-only regressions pass286.594s; production train/single/legacy/manual cases prove suite-first exact-head execution and terminal VM refusal. Corrected good-prefix and canceled-owner controls pass11.428s; gate race4.022s, model0.491s, golint1.012s; gate-bypass mutation fails then restored. Private real binary completes one exact-head runner check while both private suite slots remain held; no live VM/queue/staging or consumer activation. Delivery refresh and recorded full suite remain.; review verdict: SHIP
- 2026-10-21: closed M1 — Corrected-source runner/signal/scratch/attempt controls pass32.030s, narrow race4.359s, golint0.674s. Private collector mutation observed deletion failure then restored; all fake-dead sweep roots derive from t.TempDir and lint enforces provenance. Unknown completion and panic retain checkout/scratch/log/queue through shared Git lock; unavailable attempt writer refuses before launch. Confirmed incident and unsafe execution windows recorded in Log. M2 automatic gate remains pending.; review verdict: SHIP

Implemented the operator-approved shared Git-lock authority for checkout, scratch, raw log and temporary attempt collectors. Runner launch refuses without an attempt writer; a shared retirement proof stays false from launch intent until confirmed process-group completion, including panic unwinding. Unknown endings retain all artifacts for operator reconciliation. Independent-owner and counted owner-read controls cover isolation.

Incident: the original new TestRunnerArtifactsRetainScratchAndLogUntilReconciled called sweepStaleTemp(filepath.Dir(scratch), ..., fake-dead), which targeted shared /Volumes/TestScratch/tmp. Authenticated hub receipt 75817674-ced7-4303-bcf2-213a9888a0d7 confirms impact on pair#743 owner56219: its live run root disappeared, causing missing importcfg, binaries and fixtures. This was my unsafe fixture, not a production flake. Initial disclosure receipt84ca93c4-65d6-498a-850c-15fb1a8e5cad preceded confirmation. Pair:3 has now received submitted receipt a4d2f16d-0ccf-47d2-9cdf-363b6646f391 confirming only the isolated fixture will run; its invalidated suite may be rerun.

Recovered from this conversation's timestamped tool transcript, conservative affected execution windows on 2026-10-20 PDT (launch through observed completion, not exact sweep instants):

- 23:38:28–23:38:47: initial RED; failed at the scratch-existence assertion before either sweep, so no fake-dead sweep reached in this invocation.
- 23:40:00–23:40:47: focused corrected implementation, unsafe fixture reached.
- 23:42:01–23:43:43: broad controls and narrow race run both included the unsafe fixture.
- 23:45:21–23:46:06: focused runner controls included the unsafe fixture.
- 23:46:59–23:47:47: broad controls and narrow race run both included the unsafe fixture.

The first source isolation was applied at23:49:04 PDT, before further execution. Exact kernel sweep timestamps are unavailable; the windows intentionally include compilation and delayed result polling. Other missing-temp failures overlapping these windows should be classified against this incident. No peer process or lock was changed.

The fixture now sets WF_TEST_SCRATCH=t.TempDir(), supplies a private fallback, asserts containment and sweeps only its explicit private child. Per authenticated TL receipt c2678fdc-28f8-4cbc-b547-2ef8932ac278, M1 adds pkg/golint test-sweep-private-root: destructive test sweep roots require visible TempDir provenance through safe literal joins; runtime roots, parent traversal, uncertain aliases/reassignments and address escapes fail. Added lessons.md prevention rule. Private fixture passed0.567s; safe private collector plus integration controls passed0.642s; all golint tests passed0.510s. Earlier broad controls passed26.711s and narrow race2.541s, but used the unsafe fixture: these do not establish safe isolation. Final corrected-source controls and safe mutation remain before M1 review. M2 has not started.

### 2026-10-21 00:05 PDT — Corrected-source M1 verification

Safe mutation: replacing only runnerArtifactProtected with false failed TestRunnerCollectorsRespectPrivateAuthority on deletion of its private scratch, then source restored in finally. No runtime-selected collector is reachable from that mutation test. Corrected-source broad runner/signal/scratch/attempt controls passed32.030s; narrow runner race controls passed4.359s; complete golint package passed0.674s. Git diff is clean after mutation restoration. M1 is ready for fresh boundary review; automatic enforcement remains pending M2.

### 2026-10-21 00:20 PDT — M1 SHIP and M2 gate implementation checkpoint

M1 round3 SHIP, zero new findings, all four previous blockers disposed; milestone evidence committed/pushed by sdlc. Began M2 with observed RED: suite jumped directly to landing, the conductor rejected the new RunLandingChecks effect, and pure CI demanded VM on car heads. Added typed current-prefix checking completion, shared owner-bound selection/evidence/attempt gate, head-only CI projection, main-only explicit manual burn-in, conductor and both single shells. Passing owner proof verifies retained log hash; docs reuse also refuses watched paths. Issue-only fast landing applies that same movement proof before bypassing integration. Selected checks require a real suite even with --no-ci; VM failures never enter suite diagnosis/bisection.

Early production train+single controls passed14.095s, shared gate controls0.918s, reintegration/selection controls1.256s, and watched-main movement control2.212s. Added broader production cases (disabled/unrelated/failure/no-ci/trains-disabled/good-prefix), legacy shell and manual CLI controls; final verification pending. Two broader runs completed all test bodies successfully but were invalidated by the hermeticity guard because I edited staged source while they ran (127.476s and208.479s). Do not treat these as passing verification; source will remain frozen for corrected-source checks. The obsolete first private synthetic-waiter fixture was canceled via its exact owned Go/test processes after correcting test admission order; no live queue or peer process was involved. No VM or consumer activation.

### 2026-10-21 00:33 PDT — M2 controls and private executable proof

Frozen-source broad CI/train/CLI/legacy/issue-only regressions passed286.594s. New production cases passed for train, single, terminal VM failure, disabled/unrelated policy, --no-ci, trains disabled, legacy normal/no-ci/failure and manual main-only CLI. The initial good-prefix oracle checked second.go while its fixture created second.go.go; corrected the pushCar basename, then observed the intended full failure/baseline/good-prefix suite sequence and exactly one final-prefix VM (10.73s). Canceled-owner RED showed the suite seam was called before checking cancellation; early context checks now prevent all launches, GREEN with good-prefix control11.428s. Final narrow gate race passed4.022s, model0.491s and golint1.012s. Pure gate-bypass mutation failed the ordering oracle, then source restored.

Private real-executable conformance at /Volumes/TestScratch/tmp/ariadne-slot3/ariadne509-conformance.xkckge used this slot's clean revision9f023487f binary, an isolated Git repository/fake GitHub executable, private queue/scratch/lock roots and external runner-owned flock. The real lock helper held both private suite slots while the real sdlc named runner check completed and posted one exact-head success; inherited WF_TEST_LOCK_HELD_BY was absent inside the runner. Both private suite locks released after the owned watcher ended. Source and raw logs retained there. This precedes the small context-guard correction; refresh the binary for final delivery conformance. No live VM, production queue or :0 staging was touched. M2 ready for fresh review; delivery and recorded full suite remain.

### 2026-10-21 00:38 PDT — M2 SHIP and accepted delivery proof

M2 first review SHIP, zero findings, milestone evidence committed/pushed. Refreshed private binary at clean Git revision d0e43f42fa53f1bf7d94137baf3ba18988103526 (version reports modified=false) passed the native private two-slot-holder resource conformance again. Final probe scripts/logs live at /Volumes/TestScratch/tmp/ariadne-slot3/ariadne509-final-conformance.qPmW6a. Cooperative SIGTERM joined the external runner's cleanup, returned nonzero without posting a code-result status, retained one interrupted record, retired its checkout and released the runner-owned lock; isolated cancel queue status is running/complete/empty.

Harness corrections were private only: the first status decoder mistakenly mixed the expected private-binary stderr warning into stdout; the initial queue also retained the standalone native holder's honest unknown tombstone and discovered live legacy fleet processes. The corrected cancellation probe used a fresh private cancel queue, explicit empty private fleet discovery, and separate stdout/stderr. It passed; original diagnostic residue remains private. No live queue or process was changed.

Authenticated hub receipt e29d3ea3-075f-4463-88e0-718c8a6ab581 accepts private executable proof plus the hub's normal post-land :0 rebuild, with no staging/live VM. Capability only; no consumer declaration or hosted workflow changes. Pair#743 M3 owns the later manual-only declaration and named run on Pair main. Direct landed/schema notice to pair:3 is owed after main contains this change. Proceed to the central recorded full suite, whole close and landing.

### 2026-10-21 01:28 PDT — Recorded full suite passed

Central sdlc local-check run --name suite passed on4400533860ab73eae6ff12dccfd063ceaa496431 after2279s queued: wall649.7s, eight shards covering1908 cmd/sdlc tests and75 other packages, all exit0 with zero failures/reruns. Status sdlc-local/suite recorded on that head; log /Users/xianxu/workspace/ariadne/.git/sdlc/local-checks/4400533860ab73eae6ff12dccfd063ceaa496431/suite.log, sha256 prefix4a9192a35a78. The retained runner survived the transient model-capacity interruption; no replacement suite was started. No local-gates configuration path is changed. M1/M2 SHIP and accepted private conformance complete delivery preparation; full close, PR and central merge remain next.

## Revisions

### 2026-10-20 22:45 PDT — Expand the initial design task into reviewable boundaries

The original single planning row (“Design the narrow schema and train/single-landing selection together, then obtain full-flow approval”) is superseded by the linked durable proposal and two genuine implementation review boundaries. Scope and Done-when remain unchanged; operator approval is pending.

### 2026-10-20 22:51 PDT — Operator approval

Authenticated hub receipt 15f0281c-9941-4e15-aee3-03d074fdd9d0 records operator approval of all seven choices, with external-runner cancellation grace30s instead of15s. Proceed to change-code and M1/M2. Rollout remains capability only; no Pair activation or hosted workflow changes. This supersedes the pending approval state above.

### 2026-10-20 23:36 PDT — Operator approved re-plan

Authenticated receipt df094eb2-20d4-4e7d-a541-b8f3dab0494f approves both retention choices and continuing M1 with shared lock-based protection. Keep all artifacts until operator reconciliation after an unknown ending; refuse a runner launch without its attempt writer. Revisit retention only if unknown endings are common in practice. The approved durable plan Revision is the execution contract.

### 2026-10-21 01:28 PDT — Delivery checklist describes pre-close evidence

Narrowed the final checklist row to recorded suite/private proof and delivery preparation, now complete. Close and merge own their subsequent effects and remain mandatory; the tick does not claim a future landing. This removes the self-referential requirement to complete close/landing before entering close, with no scope or gate waiver.

## Estimate

Produced via brain/data/life/42shots/velocity/estimate-logic-v3.1.md against baseline-v3.1.md. Method A only; calibration is flagged stale, so provisional. Existing matcher, Git adapters, process-group shell and attempt store supply the libraries; no novel stack. Thorough approved plan discounts design by0.2 and uses15% buffer; impl uses40% of v2 values, familiarity1.0 from recent queue/train work.

Seven smaller modules (codec, selection, Git inputs, resource runner, proof gate, train wiring, single/CI consumers) each use v2 design0.2 ×0.2 =0.04 and impl0.5 ×0.4 =0.20. Two new pure selection/proof concerns use design0.5 ×0.2 =0.10 and impl0.5 ×0.4 =0.20. Cross-cutting runner/consumer extraction uses0.08/0.16; docs0.02/0.08; three reviews0.02/0.16 each; issue/spec0.10/0.04.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.10 impl=0.04
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: greenfield-go-module design=0.10 impl=0.20
item: greenfield-go-module design=0.10 impl=0.20
item: cross-cutting-refactor design=0.08 impl=0.16
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.02 impl=0.16
item: milestone-review design=0.02 impl=0.16
item: milestone-review design=0.02 impl=0.16
design-buffer: 0.15
total: 3.41
```
