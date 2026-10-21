---
id: 000526
status: codecomplete
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours: 2.79
collaboration-mode: collaboration
card_mirror: '7b6a3d715baa8ab5582e90aa8ed1b06dd3f5b846' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T08:01:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 2.78
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

- [x] Implement and verify the generation-pinned drop protocol and atomic merge fence.
- [x] Wire authorized command, interruptible conductor and survivor re-formation; verify controlled race/crash and real three-car cases.
- [ ] Update help/README/atlas; run recorded full suite and prepare close evidence.

Durable design: [implementation plan](../plans/000526-landing-drop-plan.md).

## Log

### 2026-10-21
- 2026-10-21: closed — BR-1 completion/drop precedence and BR-2 nested unknown-field preservation corrected within approved design. BR-2 request regression failed before fix; full landingline package including counted growth guard passes 1.689s, race 2.803s, tracked golint 0.841s, go vet exit0. BR-1 deterministic order controls, shell outcome audit, mutations and both real three-car survivor trains passed as logged. Operator-approved real-git conformance substitutes for staging. Fresh recorded suite on pushed 2eb193c29 is queued/running, not acceptance evidence yet; tl-worker explicitly permits review overlap. --no-plan-check acknowledges pending final suite item remains unchecked; final-code recorded pass is mandatory before landing. Hub owns project.; review verdict: SHIP

- Plan committed and submitted to hub; receipt2475e436-1c76-4c59-a899-06585679a21b submitted. Hub relayed operator review pending (5eb93475-33d7-49f8-9070-cc4584ea31b5): hold code until explicit answer, recommended merge-start boundary and deterministic real-Git fixture conformance. No change-code, implementation, PR, local-check lease or merge runner was started.
- After Couch restart, reread Log and verified clean retained issue branch/design. Queue running complete=true, no ariadne:1 operation; repository line is held by #524/ariadne:3 and trains remain off. GitHub PR lookup for this branch returned no PR (service degradation caveat retained). No runner/check to recover and nothing restarted; operator-approval hold continues.

- 2026-10-21 implementation checkpoint: pure/storage protocol TDD and strict codec, atomic drop/merge and apply-batch tests implemented; CLI/conductor/cancellation/positive-cleanup wiring in progress. Main CLI RED unknown command (0.414s), then real three-runner RED no control record (23.763s). First integrated fixture: dropped conductor PASS (obsolete PR CLOSED, survivors one new train, dropped head absent); dropped middle fails recomposition invariant: old conductor releases head before its fresh retry ticket/vouch exists, so another survivor departs alone and the conductor subsequently lands serially. Both surviving heads landed and dropped head stayed out; this is a production handoff ordering defect in the planned retry wrapper, not flaky timing. Raw evidence common test-stats/evidence/000526/shell-green-attempt1.log (package63.394s). Stop dependent handoff implementation, send plan refinement to hub; no timeout/sleep workaround.

### 2026-10-21 — Handoff and cleanup controls verified

- Approved handoff implementation now retains the old head through old-attempt queue retirement, enqueues/vouches the fresh ticket synchronously, then releases before ordinary Wait. An initial correction omitted that enqueue callback: shell-green-attempt2.log records the resulting five-minute timeout, with handoff-stall.sample.txt. The explicit callback correction passes both real three-car variants in shell-green-attempt3.log (35.230s), without weakening the one-train survivor oracle. Scoped retry resource tests separately reproduced retained-head leaks before extraction, then exercise early error, cancellation, resume-before-enqueue and the vouched handoff against real locks.
- Stateful negative controls exposed two cleanup boundaries: the after-push/before-Create interleaving attempted a train PR despite an accepted drop (controls-red-publication.log); a remote alias retargeted to a different real bare repository with the same ref/SHA could close/delete the wrong destination (retarget-red-publication-green.log). Publication-started now records attempted PR creation, with another fence immediately before Create; cleanup revalidates effective fetch and push destinations against the pinned repository under the deletion lock. shell-controls-green.log passes all command, real train, recovery, destination and blocked-merge API controls (33.022s).
- Broad landing-controls-green.log passed exit0 in487.849s: train/drop/grant/local-check cancellation regressions, including held-conductor/middle-car real Git cases, native cancellation, positive cleanup after lost replies, dead-conductor serial recovery, irreversible merge-start refusal and wrapper lifetime tests. Existing train/archive/clock/issue-only/reintegration controls also passed. These are targeted checks, not the required recorded full suite.
- Counted production execution profiles reject the old repeated membership scan: cost-old-scan work4074/7114/13290 at4/8/16cars, growth6176 > linear bound6080. Linear implementation cost-linear work4072/7076/13084, growth6008 = bound6008. Coverage sums all landingline production blocks, including callbacks; one instrumented fixture process runs at a time. Driver deadline handling is being finished before final package/lint checks.
- Evidence is persisted on SSD under the repository common Git directory, test-stats/evidence/000526 (including failed attempts and profiles); it does not rely on the scratch RAM disk. No stage, recorded suite, PR or merge runner has been started for this item yet. Operator-approved real Git conformance replaces staging.

- Final targeted verification: final-shell-green.log PASS71.797s, including drop-self/drop-middle conformance (43.47s), native old-suite cancellation, merge API boundary, serial crash recovery, retry lifetime paths, ordinary-branch refusal before GH effects, and completed receipt rendering. The latter two controls reproduced failures first (final-controls-red.log and completed-receipt-red.log); the local-file independent-clone receipt fixture supplies its common repository identity through the existing seam.
- final-core-lint.log: landingline PASS4.848s, tracked pkg/golint PASS1.290s; final-core-race.log PASS3.248s; go vet ./cmd/sdlc ./cmd/sdlc/internal/landingline exit0. Seven deterministic Python deadline/process-ownership controls pass0.002s. The count guard inherits t.Deadline and reserves shutdown time; no fixed child completion timeout resets. Cost-budget red/green logs retain the driver correction proof. All source/staging/HEAD remained fixed throughout each main-package test.

- Recorded branch suite on Git-verified957b61af534af2fc2b3fb4ac47a73439f0b41ca5 PASS: all8shards,1940cmd/sdlc tests and75other packages, zero failures, wall793.3s. Machine queue waited1650s before admission; run lease-4251668674 completed exit0 and posted sdlc-local/suite. Original log common Git sdlc/local-checks/957b61af534af2fc2b3fb4ac47a73439f0b41ca5/suite.log; SSD evidence copy test-stats/evidence/000526/branch-suite.log. PR374 is open; server local-gates guard passed and confirms no protected paths touched. Ready for close review; not landed yet. Hub owns project (--no-project); deterministic fixture is the approved live conformance.

- The recorded suite's queue wait crossed noon. Close refused before dispatching review because the branch lacks the main code required by .sdlc/freshness.json (main cutoff510aab6ff). Followed the named git merge origin/main path rather than bypassing. The only conflict was adjacent atlas additions: retained both #526 drop documentation and #524 reader-first attempt compatibility. No code conflict or branch implementation change; fresh recorded integrated suite is required before retrying close. Original passing branch suite retained.

- Fresh integrated suite on Git-verified88ad817e553c996e0f856d2ff17c2dc626993bfd failed before tests/setup: open /Volumes/AgentScratch/go-build/...: no space left on device. Run lease-2135974495 waited2054s, then exited2 in0s; recorded unattributable-failure. Original failed log test-stats/failed/96a1c200914cff8507ec4063a674e08f.log, SSD copy evidence/000526/integrated-suite-enospc.log, sha25696f6e398106a18399fa108eddb644f2f477336579e9ecf11e403a6a362b5b59f. No blind retry or test flake registration: no tests ran.
- Read-only diagnosis: /Volumes/AgentScratch is186Gi used186Gi,100Mi available (100%); configured GOCACHE is /Volumes/AgentScratch/go-build and du reports182G. TestScratch has11Gi available; normal Data volume279Gi available. Shared cache/machine changes require hub coordination per tl-worker; no shared deletion or setting change performed. Waiting for hub to coordinate reclamation or authorize an isolated per-invocation SSD GOCACHE and classification/retry path. PR374 remains open; close review has not started.

- Infrastructure recovered by hub (authenticated receipt26411032-120c-4c00-9464-6a3540cd3410): machine Go configuration no longer points GOCACHE at full AgentScratch. Independently verified go env GOCACHE resolves /Users/xianxu/Library/Caches/go-build. No private cache, per-run override, shared deletion, test flake registration or attempt-store edit was made. ENOSPC remains infrastructure evidence only, not test evidence.
- Explicit recorded rerun on the existing pushed infrastructure-log head1c7f5f11ac7f5b5309ee856b6ca95f9f06cd2bca PASS: all8shards,1959cmd/sdlc tests and75other packages, zero failures, wall707.9s. Run lease-7655348 waited2305s, then exited0 and posted sdlc-local/suite; no same-head refusal occurred. Original log common Git sdlc/local-checks/1c7f5f11ac7f5b5309ee856b6ca95f9f06cd2bca/suite.log; SSD copy evidence/000526/integrated-suite-green.log, sha256a5e7d83b6418 (full checksums in evidence/000526/sha256.json). Suite reports zero attributed SSD temp leakage, ten unattributed folders from other processes. Ready to retry close after the required fresh integrated suite; not landed yet.

- Close round1 REWORK, BR-1 Critical (drop-preempts-completion): review's controlled overlay demonstrated that an accepted drop could lose precedence to non-merge membership/terminal completion. UpdateTrain had no pending-drop guard; only merge/publication had one. Also swept deferred Aborted cleanup, which previously settled cars/closed PR before its final control update. This is an implementation gap in the approved drop-precedence contract, not a new protocol or scope.
- Core BR1-core-red.log reproduces bisect membership, terminal refusal and ordinary completion overwriting accepted drops; one common guard now refuses under line.lock without changing bytes. Both orderings (six scenarios) pass, full core1.887s, race2.856s, vet clean. The conductor keeps its pre-completion train until the guarded commit; every pending result routes through Apply, including deferred abort. Abort effects run only after the fence, and the redundant deferred final update is removed because each transition already committed or Apply sealed it.
- Shell BR1-shell-red.log reproduced all four missing pending routes plus uncommitted abort state. The first green attempt exposed an incorrect new assertion (terminal abort clears Cars and records Released); corrected the assertion, preserving the two-survivor oracle. BR1-shell-green2.log PASS86.475s includes actual outcome audit (dropped held, both survivors released exactly once), four completion paths, real drop-self/drop-middle landings, blocked merge, conflict/red-CI and legacy abort controls. BR1-abort-order-mutation.log then moved abort effects before the fence: the audit correctly failed because all three cars were released (2.104s). Safe ordering restored; final focused/lint/vet and a fresh recorded suite still required before re-close.

- BR-1 final safe source verification: focused completion/abort controls PASS22.720s (BR1-shell-final.log), tracked pkg/golint PASS1.580s, go vet exit0 and staged diff whitespace clean. Earlier broader BR1-shell-green2.log PASS86.475s and full core/race remain valid for this restored source. Fresh recorded suite and the full second close review remain pending; BR-1 is not marked disposed by the implementer.

- Close round2 explicitly disposed BR-1 as addressed after storage/routing mutations and real train controls; ledger correctly refused finalization on BR-2 Important (preserve-unknown-fields-on-rewrite). A real RequestTrainDrop regression reproduced cleanup.publication-only metadata disappearing on the second accepted drop (BR2-red.log). The same-generation rewrite now merges existing cleanup snapshot unknown fields while refreshing known publication values (ARCH-EVOLVE). Swept control rewrites: publication updates already merge Extra, completion/apply/finish mutate retained decoded records; BeginTrain intentionally replaces retired generation-specific nested state only after cleanup. The regression also seals both drops and finishes cleanup, asserting the metadata survives to disk. Full core package including counted growth guard PASS1.689s; fresh recorded suite and third full close review remain pending. Cancelled only our obsolete queued e48 suite by its lease after the code change became required; its terminal confirms context cancellation before any suite execution.

- Final reviewed-code suite on Git-verified2eb193c29df4550c63266ec065030a071a234d4b PASS: all8shards,1961command tests and75other packages, zero failures, wall623.1s. Run lease-2978758371 waited4547s, exited0 and recorded sdlc-local/suite. SSD copy evidence/000526/final-reviewed-suite-green.log, sha256d322bc007f493e5e201d7a7dae39bb593befbd562ea7676cb12abfffd01916c1. Close round3 returned SHIP, disposing both BR-1 and BR-2 with independent mutation verification and no new findings. Close finalized codecomplete; all changes since the reviewed source are documentation/evidence only. Final plan item now completed. Enter the normal merge gate; no stage/private cache exists to clear.

- Landing runner admitted after #463, then refused the newer-main atlas conflict and exited without landing. Followed its named manual integration path. Resolved only atlas event/effect/phase documentation by retaining main’s attribution and landing-check entries plus CarsDropped. Automatic Go integration brought main’s #527 attribution flow. Existing completion-first regression reproduced our strict control decoder rejecting its new attributing phase (integration-attribution-red.log); expanded two-order tests reproduce it directly (integration-attribution-red2.log). Recognize that existing main phase and move the membership-reduction oracle to SplitAssessed, with a separate SuiteDone-to-attributing case so the original precedence contract remains tested. No new protocol or merge boundary. Integrated core/count guard PASS4.207s, race5.854s, lint1.255s, vet exit0; command controls including both real three-car re-formation cases and updated ordering PASS47.355s. Fresh integrated recorded suite and requested narrow re-close pending; final plan item reopened until that pass.

## Revisions

- 2026-10-21 main-integration correction: accept the landed attribution phase in the strict control codec; retain membership/drop precedence tests at the new SplitAssessed reduction boundary and test entry to attribution separately. Preserve both atlas additions. This is compatibility with main, not new scope.

- 2026-10-21 BR-2: preserve cleanup publication unknown metadata during every same-generation request rewrite, including multiple accepted drops; no schema or behavior change beyond the approved forward-compatible persistence contract.

- 2026-10-21 BR-1: make the already-approved completion-versus-drop rule explicit for every pre-merge completion, including bisect membership reductions, terminal refusals and deferred aborts. Commit under line.lock before replacing the in-memory train or executing completion effects. A drop that wins uses the original riding set; deterministic tests cover both lock orders and outcome identity. No new persisted state, recovery action or merge boundary.

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
