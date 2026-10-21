---
id: 000463
status: working
deps: [ariadne#524]
github_issue:
created: 2026-10-19
updated: 2026-10-21
estimate_hours: 3
collaboration-mode: collaboration
card_mirror: '8d9eff729ef8fcd0c0065dcbdff6706533838a4a' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T04:30:26-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# local-check: resume recorded known-flake failures with targeted tests

## Problem

ariadne#461's recorded suite attempt fa09ccf7eb6fd0862fadf1d94e865944 failed
only TestSuiteQueuePriorityRealHelper when its private provider launch returned
Operation not permitted. The exact same commit passed that targeted test without
changes, and the test was registered on ariadne#462. Existing local-check run
can invoke the known-flake step only while processing a fresh full suite;
republish reposts an already-saved failure unchanged. There is no supported
way to finish this failed attempt with just the now-registered targeted tests,
so the hub approved another full suite for #461. This is a #394 follow-up.

## Spec

Add an SDLC verb to resume an already-failed recorded suite attempt whose named
failures are all currently registered known flakes. Run only those tests on
the original exact head through the existing known-flake verification path,
then post the suite success with its flake issue links and both attempt records.
Reuse the existing registry, classifier, rerun checks and evidence model rather
than introducing a second acceptance policy. Do not treat manual target output
or a republished failure as a recorded passing suite.

## Done when

- A saved failed suite whose only failures are registered known flakes can be
  resumed with named-test reruns only, without executing the full command again.
- The rerun uses the original exact commit and verifies the saved log/digest,
  repository identity, attempt completion and current registry/policy first.
- Success links the original failure, targeted rerun and open flake issues;
  every required test has an actual PASS. Other tests retain their original
  full-suite evidence instead of being silently marked as newly run.
- Unknown/unnamed failures, timeouts, missing or changed evidence, disabled
  policy, and closed/ambiguous flake registrations cannot become success.
  Revalidate registration and policy before posting.
- Cancellation and concurrent/repeated resume are handled with existing run
  ownership/evidence rules; no duplicate or unconfirmed run becomes a pass.
- Stateful regression tests prove targeted-only execution, exact-head binding,
  refusal paths and linked commit status/reporting.
- Operator option A: cleanup runs once in the original live owner. Any cleanup
  failure, crash or first-pass inconsistency becomes needs-reconcile; no retry
  infers completion from absent files or metadata. Reconciliation publishes or
  returns only terminal interrupted proof and performs zero destructive actions.
- Pending/interrupted cleanup bundles and all terminal/orphan receipts stay
  pinned regardless of current Git-lock visibility. Completed-child bundles
  whose live-owner cleanup finished keep their existing GC policy. Surviving
  input files/locks are retained; already completed unlinks are not restored.

## Plan

- [x] M1 — Immutable resume link, pure proof/reconciliation model, parent-lifetime lock and bundle retention; reuse #524's evidence snapshot and exact-package batch.
M2 — Explicit resume/reconcile commands and delivery — **superseded, never closed**, by the operator's option A decision (TL receipt 390faec8-dc49-498e-9ad5-2c1a718c2c37). Its REWORK evidence and open BR-3 remain historical review evidence; no SHIP or completed milestone is claimed.

- [ ] M3 — Option A: one live-owner cleanup pass; needs-reconcile preserves remaining inputs; reconcile only terminalizes and retains; interrupted/receipt GC is permanently pinned; reverify CLI/readers and deliver through fresh gates.

Durable proposal: [resume known flakes plan](../plans/000463-resume-known-flakes-plan.md). Await hub approval before change-code or implementation; close/PR/landing remain mandatory after the implementation boundaries.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. Calibration is marked stale by estimate-source, so this is provisional. Scope includes original issue/spec shaping, phase-2 core/store, CLI/readers, private conformance, delivered docs and three review boundaries. #524's completed implementation is excluded.

The two greenfield concerns use existing standard JSON/hash and native flock/publication primitives; that reuse halves the selected 1h design anchor, then the approved detailed design applies the 0.2 spec discount (0.1h each). Existing CLI, projection and fixture extensions use the smaller-module/refactor anchors with the same discount. Implementation uses the upper bounded v2 module ranges at 40%, with familiarity 1.0 after #524. Three reviews combine 3 × (0.1h design × 0.2 discount, 0.4h impl × 0.4 scale). The current-stack primitives already exist; no novel external API/library discovery is needed. Design subtotal 0.97h × 1.15 plus scaled implementation subtotal 1.88h = 2.9955h, rounded to 3.0h.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.5 impl=0.08
item: greenfield-go-module design=0.1 impl=0.32
item: greenfield-go-module design=0.1 impl=0.32
item: smaller-go-module design=0.06 impl=0.2
item: cross-cutting-refactor design=0.04 impl=0.2
item: smaller-go-module design=0.06 impl=0.2
item: atlas-docs design=0.05 impl=0.08
item: milestone-review design=0.06 impl=0.48
design-buffer: 0.15
total: 3.0
```

## Log

Filed unclaimed at the hub's explicit request, discovered during ariadne#461.
The current workflow required one extra full recorded suite despite established
same-exact-head fail/pass evidence. Original full log is retained under
ariadne/.git/sdlc/test-stats/failed/fa09ccf7eb6fd0862fadf1d94e865944.log.
The classified intermittent provider failure itself belongs to ariadne#462;
this issue owns only recovery of already-recorded failed attempts.

### 2026-10-21 — Resume planning after #522
- 2026-10-21: closed M1 — BR-1/BR-2 reproduced then fixed: actual receipt publication preserves raw unknown JSON and duplicate refuses; killed publisher staging retained while active and reclaimed after death. Broader M1+legacy tests 1.872s; ResumeStore race 5.410s; Go lint .709s and shell lint .335s; diff clean, atlas updated.; review verdict: SHIP

Authenticated dispatch dff45850-dd79-4197-8aa4-5089515c4de7 prioritizes this recovery after #522, with the plan going to the hub before code. Claimed and start-plan created this independent branch from fresh main. #520 remains held on its separate branch. Read the current knownFlakeStep, saved-attempt/republish/pass guard, immutable attempt writer, bounded store, queue/cancellation and flakes classifier/evidence contracts. No implementation or tests run.

Hub receipt116faa29-e008-4b37-8750-df1f1d747b73 settles the selection question: preserve #466's whole-top-level test (siblings included), but remove the cross-package name cross-product. Shared linked-record resolution is approved as the direction; the linked durable plan proposes the exact CLI, managed batch, lifetime parent lock and crash fence for approval. ARCH-DRY keeps one acceptance classifier/proof resolver; ARCH-ORDER separates reserved/unknown from completed child evidence; ARCH-EVOLVE explicitly names why old readers must refuse the new resume-link variant. Completed full-suite samples stay original, not fabricated from targeted duration. Pending proposal details are not approval.

## Revisions

### 2026-10-21 — Proposed recovery design and exact package boundary

The plan replaces the empty planning row with two genuine review boundaries. It retains the existing Spec/Done-when and adds a proposed immutable resume-link variant plus a parent-lifetime lock to protect concurrent/crashed execution. Version refusal is proposed because ignoring unresolved or failed resume proof would weaken old success guards. The hub must approve these mechanisms before implementation; no new acceptance policy or retry-until-green is proposed.

### 2026-10-21 04:40 PDT — Reader-first prerequisite approved

Hub approval eb6d1928-c198-41cd-bde5-fe8c292a4601 accepted the core plan, conditioned on scoped old-reader behavior and one real recovery command for an uncertain resume. Reported the current old reader's whole-store warning/refusal, which a new record cannot change retroactively. Authenticated917bf022-032a-4d84-ac6b-39075bbca2b5 selects a separate reader/batch prerequisite landing, then the hub upgrades every slot and this worker verifies actual binary revisions, then #463 activates resume in a second landing from main. The durable plan's Revisions records this superseding order. No link writer or resume command may ship before that barrier, and links are never hidden from old readers. Preserve this branch's design while working the prerequisite from main; phase-2 reconciliation must repair the link/proof and name one real command.

The prerequisite is filed/claimed as #524, scoped future evidence and exact-package reruns. #463 depends on that landing plus the hub's verified fleet rollout; its design branch remains preserved while #524 starts independently from main.

### 2026-10-21 10:40 PDT — Phase 2 restored; reconcile proposal before code

Authenticated TL dispatch 367355ec-38a3-427e-87c9-a02c69a27af7 accepts the rollout audit with an explicit exclusion: ariadne-slot1..5/ariadne/bin/sdlc are branch development builds, absent from runtime PATH, and integrate main through their own issue work. Shared bare sdlc is observed at clean c22771a77d40531c2debbf96f09018149e42a6f2; the independent 30-path audit found the other 25 discovered paths at that build and containing #524. TL owns rollout/runtime policy; no peer slots were messaged or modified. The dispatch authorizes #463 phase 2, with any plan delta first. Restored this issue's branch, integrated current origin/main including #524, and reran start-plan. No implementation has started.

The same dispatch relays the operator's decision to drop #520: “let's just not do this.” Ran sdlc abandon --issue 520 --as wontfix with that reason. Its Log and plan are archived, its work retained at refs/ariadne/abandoned/000520, and its parked local/remote branch deleted through that verb. #520 contributes no code to this issue.

Read-only bounded recovery survey confirms queue reconcile checks owner PID incarnation and records an operator attestation of no descendants; it neither tracks spawned PGIDs nor repairs resume proof. The plan revision proposes a narrow resume reconciliation command under the same authority and attestation boundary, which only terminalizes interrupted proof and can never authorize another execution or a pass. Explicitly awaiting the hub's approval of that recovery tradeoff before change-code.

### 2026-10-21 10:40 PDT — Active milestone scope after #524

The two active Plan rows now match the approved reader-first split. The superseded M1 combined exact-package execution with the link model; exact-package execution and scoped unknown-record readers landed in #524, while this issue retains two actual boundaries: link/proof/retention, then CLI/readers/delivery. The durable plan's appended phase-2 revision is the current implementation contract; earlier task lists are preserved as design history. Done when remains unchanged: interrupted or repeated resume cannot become success.

### 2026-10-21 10:44 PDT — Phase-2 delta approved

Authenticated receipt 928fcd1f-3d96-4c04-8231-025c9adaf064 approves the exact reconcile command, operator/:0 attestation boundary, owner-incarnation check under the parent lock, and immutable terminal interrupted receipt. No stage: private production-CLI conformance must prove real registered suite failure → resume → post → guard accepts, and crash mid-resume → reconcile. The TL will arrange the first natural post-landing live resume; this worker must not repair production evidence pre-land. Proceed to change-code. No automatic descendant discovery, successful repair, link deletion or invented attempt is authorized.

### 2026-10-21 10:48 PDT — Plan review refinements

First change-code review refused with PQ-1 (retirement monotonicity) and PQ-2 (named risky function/test strategies). Confirmed the existing pruner removes members individually. The revised contract durably removes the selectable parent first, preserving its fence until a successful directory sync, then cleans child/receipt/log members and link last; retry never restores execution eligibility. Replaced prose case inventories with named function-level adversarial strategies. No operator contract or mechanism changed; rerun the plan gate before implementation. These semantic crash-order and plan-quality classes are enforced by deterministic pruner fault tests and the existing gate, not a source-name lint rule.

### 2026-10-21 11:10 PDT — Retirement ordering refined during M1

TL approval aad50f3d-c5e6-4b21-9cd1-ff95d3fb46d6 accepts parent-first durable removal, logs/inflight, link, then terminal child/receipt witness last. The final witness includes the child raw log. This allows interrupted parent-missing cleanup to prove terminality again while preserving the no-new-execution invariant; no additional retirement-record family or authority change is introduced. The plan Revisions records the superseding order. Core decoder/proof/transition and parent-lock/publication tests have observed RED then GREEN; production-pruner pending-bundle regression observed the old independent pruning deleting the parent before integration.

### 2026-10-21 11:18 PDT — Stop retirement integration for full locking refinement

M1 targeted core/store/pruner checks passed (1.518s); pure fuzz completed 803 executions. Self-review plus bounded independent retention analysis found a stale protected-snapshot race with legacy pruning. Stop this integration before M1 close: parent locks alone do not protect publication completed after the snapshot, and rescanning per parent is quadratic. The plan Revisions proposes one existing-directory inode mutex around the snapshot/mutation boundary, publisher parent→directory and pruner directory→nonblocking-parent, with deterministic process barriers. Asked the hub to approve this revised locking contract; no CLI activation or production writes. The typed core, immutable publication and bundle pruning are checkpointed as in-progress, not milestone-complete.

### 2026-10-21 11:25 PDT — Directory-lock revision approved and implemented

Authenticated TL f21f7022-6260-451b-86ae-a0a56158abac approves the directory-flock contract, emphasizing the strictly nonblocking pruner as the condition preventing an inverse-order deadlock. Code comments and atlas state that invariant. Basic directory-lock tests observed RED before implementation, then GREEN (0.416s). Three real subprocess interleavings initially passed (1.331s); a deliberate legacy-pruner blocking-flock mutation deadlocked the controlled old-snapshot/publisher sequence and failed its one-minute negative-control timeout (60.364s). Restored the strictly nonblocking acquisition immediately; the two private mutation-test helper process groups were explicitly terminated and verified gone. No recorded suite, production process or evidence was touched. This is the second retirement design change; a third requires stopping to re-plan retirement as a whole, per TL direction.

### 2026-10-21 — M1 verification before boundary review

Focused core/store/snapshot/retention and legacy reader/writer/pruner tests passed in 2.203s; all Resume tests under race passed in 5.506s. Pure decoder fuzz passed 803 executions in 4.822s; Go lint passed 0.823s and shell lint 0.295s; diff check clean. The old malformed “old” JSON expiry fixture now writes a valid SuiteAttempt, since unknown malformed evidence conservatively stops pruning rather than deleting potential authority. The linked-pruner tests cover IO failure at every deletion/sync effect and actual subprocess lock lifetimes; the model tests independently reject duplicate execution and unproved success. M1 remains pending its mandatory fresh review; no resume/reconcile CLI or automatic link writer is present yet.

### 2026-10-21 — M1 round-1 REWORK and repairs

Mandatory review found BR-1 Critical: the publisher's broad proofError rejected the very supported link required for receipt publication. BR-2 Important: `.resume-*` staging files relied solely on defer and escaped crash cleanup/capacity. Both were reproduced through production seams: valid receipt publication and abandoned staging tests failed in 0.514s; a real killed-publisher regression failed in 0.576s. Fixed the publisher to refuse unsupported authority then validate its supported graph explicitly, and reclaim staging under the existing directory mutex in the next publication/prune. No new retirement lock/order/mechanism is introduced. Append-only raw parent/link/receipt and unknown-field preservation remain intact. The plan Revisions records staging creation, ordinary removal and crash reclamation. These are semantic variant/lifetime rules enforced by the real publisher and killed-owner tests; a lexical source-name lint cannot prove them.

Clock correction: earlier manually supplied hour/minute headings in this session's notes are inaccurate. Git commit timestamps, Couch receipts and gate sidecar timestamps carry the actual chronology; later notes use date-only headings.

BR-1/BR-2 repair verification: receipt/staging/crashed-publisher tests passed 0.773s; broader M1 plus legacy compatibility tests passed 1.872s; ResumeStore race passed 5.410s. Raw future JSON values in the published receipt remain byte-identical, and duplicate receipt publication refuses. Diff check clean. Pending mandatory M1 round-2 disposition before starting M2.


### 2026-10-21 — M2 checkpoint: commands, contextual proof and private acceptance

M1 closed SHIP at the mandatory boundary (evidence commit recorded by Git).
M2 now wires explicit suite resume and operator/:0 terminal reconciliation via
the pure transition model. The parent inode spans all effects, and managed
execution uses #524's exact-package batch with suite capacity and existing strong
group cleanup. Contextual saved reuse/republish and each status retry resolve all
same-head links against current policy/registry and actual raw logs; no raw
parent is rewritten. Registry-change, stale posted child identity and aggregate
raw-pass bypass regressions observed RED before their reader repairs.

Private production-CLI acceptance passed 38.698s (completed 14.35s; owner crash
and reconciliation 23.78s), output /tmp/ariadne-463-private-resume-conformance.log.
The first run exposed repeat-resume's same-state repost transition; fixed and
reran the full fixture. It proves failure before registration → exact-package
resume → post → pass guard, immutable original bytes and no repeat execution;
then owner-only crash, explicit fixture descendant stop, terminal receipt,
repeat reconciliation and refusal of a later real green suite. No stage or
production evidence mutation. Broad current/legacy reader checks passed 60.461s;
CLI validation passed 0.322s; Go lint passed 1.028s. Environment lifetime tests
passed 2.864s before an additional changed-owner-marker regression observed RED
0.538s; its direct revalidation fix awaits final focused verification. Further
counted N/2N checks, broader/race verification, recorded suite and M2/issue gates
remain pending. This commit is an in-progress recovery checkpoint.


### 2026-10-21 — M2 verification before recorded suite and boundary review

CLI/authority/live-owner/foreign-host/wrong-digest/committed-child refusal and
private environment retirement checks passed 3.475s, including the final
changed-inflight-owner revalidation guard. All Resume tests under race passed
15.754s. Broader existing saved-proof/run/republish/known-flake/scoped-reader
checks passed 23.114s. Counted real-store N/2N checks passed 0.719s: snapshot
opens exactly 3N+1 paths once each, contextual projection/lookup/guard reads
exactly 2N actual logs, republish 4N across two decisions, and policy/registry
reads once per decision at N=4/8. No per-parent snapshot rescan. Go lint 0.634s,
shell lint 0.307s, diff check clean. The missing-duration retention validation
now refuses incomplete parents/children while accepting measured zero; it adds
no retirement mechanism. Docs and help cover the exact commands, terminal
recovery, explicit descendant attestation, unknown queue independence and
unchanged original measurement samples. Checkpoint/push before the recorded
suite; M2 remains pending the mandatory review and recorded pass.

### 2026-10-21 — M2 REWORK BR-3; stop for whole retirement re-plan

Mandatory review found BR-3 Important (durable-staging-cleanup): completed-child
publication removes its inflight owner binding before deferred Git checkout
retirement. A real CLI kill during status posting reproduced a permanently
locked checkout after repeated resume and an aged sweep. Verified the source
ordering; no dispute. The finding generalizes to pre-link, pending, committed
and reconciled protected-input lifetimes, not just one post call.

Stopped the obsolete recorded suite through `sdlc local-check cancel --run
lease-3748438976`; it was still queued, never acquired capacity, and completed
with context-canceled refusal. The third-change rule requires a whole-retirement
re-plan: appended the complete lifecycle, invariant, alternatives and named crash
strategies to the durable plan. Proposed fence-before-protection, durable marker
before lock, confirmed environment retirement before child commit, and bundle
GC preserving protected/unreadable cleanup authority. No implementation of this
revision before hub approval. M2 and issue remain open; no PR or stage.


### 2026-10-21 — Approved whole-lifecycle repair and final private verification

TL receipt 91596b33-cfe5-4d0d-bfff-1cc377ee1656 approves the whole retirement
re-plan and requires stopping for operator review if the next review finds
another retirement/lifecycle gap. The fence now precedes protection; marker,
child-directory and inflight-parent durability precede Git locking. Managed
scratch cleanup proves/syncs absence; exact checkout removal precedes Git
registration removal, both confirmed/synced before child publication. Marker
release comes last. Receipt cleanup shares exact binding/removal logic while
retaining its own absent-owner/attestation authority. Prepared unlocked inputs
remain ordinary collector work, without weakening deletion authority. Bundle
and orphan-receipt GC preserve locked or unreadable owner bindings.

All eight declared crash windows passed in the final combined private run:
95.599s, output /tmp/ariadne-463-final-private-conformance.log. The status-post
crash test now proves no prepared checkout/registration remains before or after
repeat/republish, immutable parent/child/link bytes and no duplicate execution.
Environment/retention targeted group passed 5.270s; all Resume race checks passed
13.378s; scratch/CLI/owner/authority checks passed 0.583s; Go lint 0.604s and shell
lint 0.167s; diff clean. Silent scratch deletion failure, incomplete ownership,
unrelated dangling scratch markers and unreadable/dangling GC marker controls
observed RED before the class repair. No production state, stage or PR. Push
this final code, queue a fresh recorded suite and rerun mandatory M2 review.

### 2026-10-21 — STOP: M2 re-review retains BR-3; operator decision required

The mandatory M2 re-review returned REWORK. It confirms the original
completed-child/status-post leak is fixed and its regression fails under a
negative-control defer mutation. BR-3 remains not-addressed in the same
`durable-staging-cleanup` family: retry can observe missing Git registration
and release the owner without retrying a failed or never-acknowledged sync of
the registration namespace; similarly, a missing-marker return bypasses a
previously failed owner-release directory sync. The reviewer’s disposable
production-cleanup fault tests failed for lost removal reply, registration-sync
failure and owner-release-sync failure. Source inspection confirms those early
returns. Visible absence alone does not discharge durable deletion obligations.

Per authenticated TL 91596b33-cfe5-4d0d-bfff-1cc377ee1656, **stop further lifecycle
patching and return the finding to the TL for operator review**. No additional
code fix, dispute, review rerun, close, PR, stage or landing is authorized by
this checkpoint. The recorded suite was still fifth in the shared queue;
canceled only our exact lease-2271555155 through local-check cancel. It never
acquired capacity, so no full-suite pass is claimed. Existing private/race and
reviewer checks remain evidence for delivered behavior, not completion of the
lifecycle invariant. M1 remains closed; M2 and issue remain working/open.

The re-review sidecar contains the exact failure cases and locations. Next
worker/TL must read it and await the operator’s decision rather than deriving a
new incremental patch from this Log. All work and review evidence are pushed;
the checkout is left on the issue branch.


### 2026-10-21 — Operator option A; M2 superseded by fresh M3 (plan review pending)

Authenticated TL 390faec8-dc49-498e-9ad5-2c1a718c2c37 relays the operator's
choice of A exactly as the read-only deletion map: one live-owner cleanup pass;
all first-pass inconsistencies become needs-reconcile; reconcile terminalizes
and retains with zero destructive effects; no reconcile-initiated cleanup and
no new one-shot boundary. Interrupted cleanup bundles/terminal or orphan
receipts are never automatic-GC candidates, independent of lock visibility.
Completed successful-cleanup GC stays unchanged. Updated active Done when and
Plan accordingly; M2 is superseded without ticking or closing it, and M1 stays
closed. Its review/ledger evidence remains intact, including open BR-3.

Appended a fresh M3 design/tasks/test and visible-state contract to the durable
plan. No code changes or tests started. Commit/push the design and send it to
the TL before code. After TL approval, rerun change-code for fresh plan-quality
admission, follow the gates' exact freshness remedy if required, then execute
M3. Final issue close must dispose of the carried M2 BR-3 under the new operator
contract through the real review, not by manually editing its ledger.

### 2026-10-21 — Fresh M3 accepted; change-code next

TL acceptance f1791b10-0ac2-4140-bde3-ae229e144241 approves M3's one original-owner
pass, needs-reconcile on every inconsistency, evidence-only interrupted audit,
pinned pending/interrupted/receipt graph and unchanged completed-cleanup GC.
Unused preparation that never launched retains the accepted ordinary collector
boundary. At whole-issue review, dispose carried M2 BR-3 as superseded by the
operator's option A, citing decision 390faec8-dc49-498e-9ad5-2c1a718c2c37; do
not clear its ledger. Record acceptance before the fresh change-code plan gate.


### 2026-10-21 — M3 plan-quality admitted; deletion work begins

Fresh change-code passed plan-quality round 4 with no open findings and accepted
the M3 design on the issue branch. Estimate-quality returned INFO only: the
original 3.0h baseline predates the superseded M2 re-reviews/operator re-plan;
those unplanned costs were not individually allocated in its generic rows. It
remains the original forecast for calibration, not a claim that known actual
rework fits that allowance or a fabricated updated remaining forecast. Actuals
will be measured by the gates. No retrospective estimate change is made.

M3 evidence-only reconciliation observed RED 16.341s: the old first reconcile
deleted the retained Git lock. Removed both cleanup calls. Strict scratch
uncertainty tests observed RED 0.455s for cancellation, timeout, launch failure,
signal exit and unconfirmed groups; new forward-only handling passed the
focused scratch/CLI/authority checks 0.642s. Environment/retention and full CLI
rewrites are in progress under the accepted plan; no completion is claimed.

### 2026-10-21 — M3 final verification before recorded suite and fresh boundary

M3 removes receipt cleanup/retry adapters, missing-state successes and cleanup
reconstruction. Reconcile first/repeat has zero destructive effects, retaining
the interrupted audit and all surviving inputs. The original owner has one
forward cleanup pass; uncertain managed results and every initial inconsistency
stop, with needs-reconcile and retained paths. Pending/interrupted/receipt and
orphan graphs remain pinned without relying on Git-lock/marker visibility.
Completed-child GC, original raw evidence and all success-proof rules remain.

Final private production-CLI acceptance passed 91.155s, including the seven
crash windows, additional cleanup-first-pass failure, completed/running flow,
first/repeated audit-only reconcile, aged prune twice, exact remaining-byte/
path/mode/registration/lock preservation, no extra launches/posts, and later
actual green refusal. Output /tmp/ariadne-463-m3-private-conformance.log.
Focused receipt/fault behaviors passed earlier but that package invocation
reported a source-checkout guard failure when the docs baseline changed; the
final combined run passed that guard too. Environment/retention checks passed
4.983s, with N8/N16 receipt/orphan graphs × three prunes showing zero removals,
and an added 100-day pending-graph preservation test. The prior RED1.151s
independently showed receipt/orphan member removals and partial-state successes.

All Resume race checks passed 12.688s; broader legacy saved-proof/run/republish/
known-flake/scoped-reader checks passed 19.428s; scratch/CLI/authority/current
saved-proof checks passed 2.504s; Go lint 0.632s and shell lint 0.274s. Diff clean.
Checkpoint/push final code, queue the fresh recorded suite and close M3 through
its own boundary. The original M2 remains superseded, not closed; the whole-
issue gate must dispose of BR-3 as superseded by operator option A, decision
390faec8-dc49-498e-9ad5-2c1a718c2c37, with ledger history preserved. No stage.

### 2026-10-21 — Required noon freshness integration before M3 review

M3 milestone-close refused at the integrate gate before any review dispatch:
main code landed by noon was absent. Followed its exact git-merge remedy rather
than bypassing it. Canceled only still-queued lease-956462783 on the previous
head through local-check cancel; it never acquired suite capacity. Integrated
origin/main without conflicts (including landed #514/#527 and the named noon
changes), preserving both atlas/lessons edits, and stamped the merge trailers.
This is the gate-required freshness integration, not repeated manual chasing
of newer main. Reverify relevant targets/lint, push the integrated code, queue
one fresh recorded suite and rerun the same M3 gate. No full suite pass or M3
review verdict is claimed yet.

Required integration verification: all Resume race tests passed 17.464s on the
integrated tree; Go lint 1.244s, shell lint 0.377s and diff check clean. The
private M3 acceptance evidence remains 91.155s from the unchanged feature code;
the fresh recorded suite will validate the full integrated tree. Checkpoint,
push and requeue once before rerunning the M3 boundary.
