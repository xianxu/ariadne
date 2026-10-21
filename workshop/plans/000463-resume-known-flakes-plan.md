---
collaboration-mode: discussion
---

# Resume recorded known-flake failures implementation plan

> **For agentic workers:** Follow AGENTS.md Section 3. Execute the warm evidence/queue integration in this session; milestone-close owns each fresh-context review.

**Goal:** Finish a saved failed full suite using one recorded rerun of its registered failures on the original exact head, without running or counting another full suite.

**Architecture:** Reuse the existing flake classifier, whole-top-level-test rule, managed suite execution and immutable attempt/log records. Add an explicit resume command, exact-package rerun batches and one shared proof resolver. A lifetime lock serializes recovery of one parent; an immutable resume link published before execution keeps a crash or incomplete child from becoming absence of an attempt.

**Tech stack:** Go, existing real Git/worktree adapter, stateful fake GitHub/registry seams, native advisory locks and existing managed queue/process ownership.

## Core concepts

| Name | Lives in | Status |
| --- | --- | --- |
| Exact-package rerun plan | cmd/sdlc/internal/flakes/rerun.go | modified |
| ResumeLink | cmd/sdlc/localcheck_resume_evidence.go | new |
| ResumeProof | cmd/sdlc/localcheck_resume_evidence.go | new |
| Evidence snapshot | cmd/sdlc/teststats_store.go | modified |

- **Exact-package rerun plan:** validated failing top-level names grouped by their own package. One anchored argv per package (packages may share one command only when their selected name sets are identical). Subtests still promote to the whole owning test, including siblings, per #466. No cross-product of every test name with every package. Deterministic order; at most the existing flake limit, no shell-generated test names.
- **ResumeLink:** one immutable, typed link from a saved parent and its exact record/log digests to a reserved child attempt ID. It claims no successful execution. One link per parent; repeated requests find it, not a new execution. The parent and raw full-suite log never change.
- **ResumeProof:** a pure result: eligible, active, resolved-pass, resolved-failure or uncertain/refused, with parent/child identity and reason. The resolver consumes a complete evidence snapshot and verified log facts; it does not read files, call GitHub or launch commands. Existing flakes.Classify/Revalidate remain the acceptance policy.
- **Evidence snapshot:** one bounded read indexes attempts and resume links by ID and parent. Existing timing/actual/statistics consumers still receive only real SuiteAttempt rows; links never count as test runs. The saved-status/guard/resume consumers use the same proof projection.

| Integration | Lives in | Status | Wraps |
| --- | --- | --- | --- |
| Resume CLI | cmd/sdlc/localcheck_cmd.go and localcheck_resume.go | new |
| Managed command batch | cmd/sdlc/localcheck_run.go | modified | existing queue, slot, cancellation, process groups and scratch |
| Linked evidence store/lock | cmd/sdlc/localcheck_resume_evidence.go and teststats_store.go | new/modified | existing portable test-stats root and atomic publication |
| Saved proof readers | localcheck_republish.go and localcheck_flakes.go | modified | saved pass, republish, automatic flake step and success guard |

All storage resolves through the existing Git-common-dir/local-check log root. Tests use private roots; no production evidence or queue is mutated. Reuse ghFake and the existing fake registry, plus real private Git/worktree fixtures and the real lock helper.

## Command and eligibility

Propose `sdlc local-check resume --name suite --attempt ID [--rev HEAD]`. Require an explicit parent ID; no automatic selection of an older convenient failure. Resolve --rev once (default HEAD), require it equal the parent's full SHA and remain a pushed commit; execution uses a fresh detached checkout of that SHA. The caller's newer checkout never supplies test contents. Initially suite only: affected/runner/baseline/legacy/manual output are refused, without changing their existing behavior.

Before effects, verify repository identity, complete bounded store read, parent completion, nonzero ordinary exit, no timeout, supported record type, original raw-record/log digests and parsed failures. Compare parsed facts to the stored failure list; missing, changed, truncated, oversize or unattributed evidence refuses. Only outcomes the existing flakes.Explainable rule permits can be resumed; a previously rejected twice-failed/inconclusive/disabled-policy attempt stays blocked.

Read the declared suite/policy from one freshly fetched main pin and the current open-issue registry. Every failing leaf must be covered without ambiguous registrations, within the existing limit. Policy off refuses. For policy accept, this explicit command still requires a real targeted rerun: evaluate it through the existing rerun policy, never synthesize accepted-without-run proof. Fetch/read the current policy and registry again before posting; off, unreadable or dropped registration refuses.

## Coordination and publication protocol

1. Acquire an exclusive native lock on the selected immutable parent record inode, held through verification, queue wait, child execution, record publication and final status handling. Recheck the path/inode and digests after acquiring it. A second caller waits/cancels, then rereads the same link/proof. No machine-wide repo lock is held across a suite wait.
2. Reserve the existing child attempt writer first. Exclusively publish the ResumeLink before launching anything, under the parent lock. Its exact child ID removes any “choose the newest passing rerun” ambiguity. A crash after link publication leaves an unfinished proof; it cannot cause an automatic duplicate execution or a pass.
3. Run the exact-package batch through the existing managed suite path: one weight-1 capacity lease and one total existing targeted-rerun timeout (15m, capped by the suite timeout), not one budget per package. Reuse queue admission, grants, private scratch, owned signals and local-check cancel. Never invoke c.Command/full make test. Run package commands sequentially with literal argv; stop on nonzero/cancel/timeout and classify missing required PASS evidence as inconclusive. Runner-resource behavior remains single-command and unchanged.
4. knownFlakeStep shares the new batch builder/executor with resume. Keep one `suite-rerun` child, linked by existing RerunOf, with actual argv/log/queue/duration/slot/agent evidence. Require the actual PASS for every original failing registered name and its owning package; a PASS printed by another package cannot count. Preserve original full-suite coverage and durations.
5. Publish the completed child before any success post. Resolve the parent/link/child through the shared resolver, with verified logs and classifier outcome. A failed/inconclusive child is terminal for that link and blocks a repeated resume and the pass guard; it is never an invitation to retry until green.
6. Revalidate fresh policy/registry and the complete current evidence set, including any unrelated unexplained same-head failure, immediately before posting `sdlc-local/suite`. Reuse postLocalCheckRecord/guardSuitePass; describe “known flake rerun” and link both attempt IDs and issue numbers. The status is composite evidence, not a claim of a newly run full suite.
7. A repeated request with a verified completed passing link revalidates/reposts that proof without running tests. Lost GitHub response uses the existing republish path, now driven by this same resolver. A missing/unconfirmed child or interrupted attempt names its retained paths and requests inspection; never auto-rerun it. The existing queue reconciliation handles uncertain processes before an operator can authorize evidence repair.

**Lock-shaped invariant:** at most one resume executor per parent is the one parent inode lock held for its full lifetime. The immutable link is the crash fence: loss of the lock after process death never proves the effect did not occur. Do not unlink or replace a locked parent to release it. Pruning/replaced-path races refuse by identity/digest rather than authorizing a new run.

## Evidence schema and compatibility decision for TL approval

The existing parent and child SuiteAttempt schemas remain unchanged. ResumeLink is a new version-3 record variant in the same bounded `test-stats/attempts/` namespace, with a distinct kind and random ID, repository, full SHA, parent ID/digests and reserved child ID. Decode as an explicit typed union; known types/required fields/duplicate keys remain strict, unknown members are tolerated. The link is never rewritten, so unknown bytes are retained unchanged.

This version boundary is intentional (ARCH-EVOLVE): an older reader that ignores a pending resume or a failed linked child could accept a later blind pass while the resume is unknown or twice-failed. Existing readers reject an unsupported record and mark the store incomplete, refusing green. New readers recognize links without counting them as SuiteAttempt samples; all authority readers use the shared resolver. Test the actual old-reader/new-record refusal. New versions are written only when resume is invoked; ordinary automatic reruns keep their existing schema.

This is coordination evidence, not a second flake acceptance policy. If the TL prefers parent mutation instead, stop and revise before implementation: do not quietly add a rewrite or an unrecorded retry. A parent rewrite would need raw-field preservation and crash-safe adoption; an in-memory-only resume cannot distinguish crash from absence.

## Visible states and recovery

Actions: repair evidence, wait/cancel, inspect/reconcile, fix code, resume once, republish saved proof.

| Internal condition | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Complete eligible parent, no link | ready to resume | resume once | none; full-suite evidence retained |
| Owner holds parent lock; queued/running child | resume active | wait/cancel | time already spent; no pass before completion |
| Explicit cancel or child missing after owner loss | interrupted/uncertain | inspect/reconcile | partial targeted output, no passing authority |
| Completed child fails, times out, omits PASS or fails elsewhere | still failed | fix code | original and targeted failure evidence retained |
| Registry/policy/evidence becomes unreadable or changes | cannot verify | repair evidence | no new success post |
| Child passes and full proof still validates | verified known-flake pass | republish saved proof if needed | no full-suite rerun; status may be temporarily unavailable |

Confirmed cancellation retires managed processes through the existing path; the original failure/link and interruption evidence remain. This first version does not automatically recover a child whose completion was never durably recorded. That conservative boundary is explicit, not absence inferred from PID death. Operator repair must prove managed descendants gone and retain the original failure; no force-to-success flag is proposed.

## Review boundaries and implementation tasks

The original M1 exact-package work was delivered by #524. The phase-2
boundaries and function-level test strategies are in the current revision below;
this section is retained as the original boundary heading, not an active task list.

## Cost, retention and scope

ARCH-CONSTRAINTS: manual recovery path. One bounded evidence snapshot per decision/revalidation, one registry read before execution and one before post; never rescan the whole store per package or test. Existing caps: 10,000 records/64MiB snapshot, 1MiB record, existing bounded log read, at most five registered failures, one weight-1 suite lease and a single 15m targeted budget. Count store scans, log reads and command launches for N and 2N synthetic parents/links; index once and require linear work. Capacity/FIFO/grants stay real in production.

ARCH-FUNERAL: one immutable link per resumed parent, one normal child attempt and existing inflight/interruption files. Completed bundles share the existing 30-day attempt retention and are pruned as a unit so a retained pass never loses its required parent/log first. Pending/unknown links retain evidence until operator reconciliation; the normal bounded reader refuses when the store cannot be read completely. Do not add expiry that turns uncertainty into absence. Native lock handles die on close; no permanent separate lock-file family.

ARCH-SECURE/PURE/MOCK: raw record and log bytes are untrusted; typed decode and verified hashes precede effect authority. Build literal argv from validated names. Use fake GH/registry plus private real subprocess/Git/lock fixtures; never patch the live queue. ARCH-DRY/PURPOSE: one classifier and one linked-proof resolver defend every passing reader. Scope excludes affected-suite resume, arbitrary named checks, manual-output adoption, retry-until-green, flake registration or changing the operator's policy.

## Approval state

Draft for the hub. Authenticated receipt 116faa29-e008-4b37-8750-df1f1d747b73 approves retaining #466 whole-parent semantics and removing cross-package over-selection, and asks for this plan. CLI shape, immutable-link crash fence/version boundary and uncertain-resume recovery are proposals awaiting approval. No implementation has started.

## Revisions

### 2026-10-21 04:40 PDT — Approved design, reader-first rollout required

Authenticated approval eb6d1928-c198-41cd-bde5-fe8c292a4601 approves the core design, with scoped compatibility and one named crash-recovery command as conditions. The current pre-change reader cannot satisfy scoped compatibility: an unsupported record becomes a global read warning, and guardSuitePass refuses any passing status before repository/head/check filtering. loadSavedAttempt also marks the entire snapshot incomplete. No v3 payload can retrofit already-built readers, and hiding links would let them ignore unresolved or failed recovery authority.

Authenticated follow-up917bf022-032a-4d84-ac6b-39075bbca2b5 chooses two separate landings, superseding the original M1/M2 publication grouping:

1. A prerequisite issue, based on main, lands reader-only compatibility plus exact-package managed rerun batches. Supported older/v3 envelopes are scoped to the affected attempt/proof with a “needs a newer sdlc” refusal; unrelated checks/records remain usable. Malformed input whose identity cannot be proved remains unknown rather than silently discarded. No link writer or resume activation ships in this landing. The atlas documents the pre-change reader limitation and activation barrier.
2. The hub rolls **every slot's actual sdlc build** forward. Verify bare sdlc --version or executable build revisions across the fleet; do not assume :0 aliases or sibling ownership without observing actual paths. Old running writers must finish or be upgraded through the hub's rollout. Record the observed revisions, gaps and completion evidence; absence of a response is not upgrade proof.
3. Only after the first landing and verified rollout does #463 resume from main containing that prerequisite and implement link writing/resume activation in a separate landing. No stacked branch and no reused merged branch name. Its own M1/M2 boundaries become link/proof/reconciliation, then CLI/readers/delivery; exact-package execution is reused from the prerequisite.

The crashed-resume refusal must print one executable recovery command. Existing queue reconcile is sufficient only if it actually reconciles the resume link/proof as well as process ownership; otherwise add a narrow resume-specific reconcile under the parent lock with exact link digest, operator note, proven stopped descendants and retained audit. Do not ship a generic “inspect” instruction or claim queue tombstone removal repairs a resume link. That command's precise shape will be finalized before phase-2 code, not guessed in phase 1.

### 2026-10-21 10:40 PDT — Phase-2 execution and terminal reconciliation proposal

**Reason/delta:** #524 is landed; its managed exact-package batch and scoped unsupported-evidence snapshot replace the corresponding original M1 tasks. Authenticated TL dispatch 367355ec-38a3-427e-87c9-a02c69a27af7 accepts the 25 current runtime-path builds and excludes the five slot-local branch dev binaries by policy, then authorizes phase 2 subject to a plan delta first. Bare sdlc --version observes clean c22771a77d40531c2debbf96f09018149e42a6f2. This appendix supersedes the original task grouping, pending approval state and generic crash-recovery prose. It is a proposal awaiting hub approval; no code is authorized by this appendix itself.

**Simplest alternative checked:** existing queue reconcile alone cannot settle a ResumeLink. Queue Operation/runLease contain the SDLC owner PID, not a durable list of descendant process groups. Automatically claiming descendant absence after owner death would require a new launch-barrier/ownership subsystem to cover process start before PGID publication. That is unnecessary for a conservative terminal repair: reuse queue recovery's explicit operator-attestation boundary, and never let repair grant a pass or another execution. The operator/:0 steward must verify stopped descendants externally and attest it. The program proves owner-incarnation absence, not descendant absence. This limitation requires explicit hub acceptance before code.

#### Current entities and seams

| Pure entity | Lives in | Status |
| --- | --- | --- |
| ResumeLink | cmd/sdlc/localcheck_resume_evidence.go | new |
| ResumeReconciliation | cmd/sdlc/localcheck_resume_evidence.go | new |
| ResumeProof / ResumeTransition | cmd/sdlc/localcheck_resume_evidence.go | new |
| SuiteEvidenceSnapshot | cmd/sdlc/teststats_evidence.go | modified |

ResumeLink uses the v3 identity envelope already scoped by #524: id, repo, sha, check=suite and rerun_of=parent ID, with distinct kind, exact parent-record/log digests, reserved child ID, owner PID/start-time/boot/host identity and creation time. Strict typed decoding rejects duplicate/null/invalid known fields and inconsistent identities; unknown fields stay in immutable raw bytes. Each parent has one link, published exclusively under its inode lock before execution; the bounded snapshot indexes parent/link/child once. Only actual SuiteAttempt rows feed measurement consumers.

ResumeReconciliation is a separate immutable v3 variant in that same namespace, scoped through rerun_of to its parent. It binds the exact link ID/digest and child ID, actor, note and time, with terminal interrupted/inconclusive disposition. It claims no test run, exit status, PASS, or duration. Raw parent/link/log bytes are retained. A valid completed child is resolved normally and cannot be replaced by reconciliation. Conflicting receipts or child/receipt combinations refuse. Repair never deletes or rewrites a link and never returns the parent to eligibility.

| Integration | Lives in | Status | Wraps |
| --- | --- | --- | --- |
| Parent lock and exclusive proof publication | cmd/sdlc/localcheck_resume_store.go | new | native file lock, typed/raw evidence and atomic publication |
| Resume/reconcile CLI | cmd/sdlc/localcheck_resume.go, localcheck_cmd.go | new/modified | Git, current declarations/registry, parent lock and managed batch |
| Shared linked saved proof | localcheck_republish.go, localcheck_flakes.go | modified | snapshot/resolver and existing post guard |
| Bundle pruning | teststats_store.go | modified | snapshot, parent lock, retained raw logs/inflight state |

Reuse #524's readSuiteEvidence, knownFlakeRerunCommands and runUnderSlotBatch; do not add a second batch engine. Reuse queueOwnerAbsent/processgroup.StartTime for exact owner-incarnation checks and landingline.Authorize for operator/:0 authority. ghFake and the tracker fixture remain the stateful external seams; subprocess/Git/lock tests use private roots only. Link-derived paths must stay under the validated store, and symlinks/replaced parent paths refuse lock/identity validation.

#### One executable uncertain-resume recovery command

`sdlc local-check reconcile --attempt PARENT_ID --link-digest SHA256 --note 'operator verified no descendants remain'`

The refusal prints the exact parent ID and raw-link digest in that command, plus retained paths. It does not silently invoke the command. Only an operator outside Couch or its :0 steward can execute it, following the same authority rule as queue reconcile. Notes are bounded, nonempty data. Require same-host/boot identity or an explicit safely absent old boot; an uncheckable owner refuses. The command takes the parent inode lock, then rereads/revalidates the exact parent, link, child, digest and existing receipt. A live or unknown owner refuses. Operator attestation supplies the descendant-stop fact; a missing PID or an empty queue never supplies it.

With no committed child, reconciliation publishes one exclusive durable terminal receipt before any retirement. Repeated identical requests return that same receipt without another mutation. Crash after publication is resolved by rereading it; a changed digest or competing receipt refuses. Any existing queue unknown record remains handled by the existing exact queue reconciliation command; this command does not remove unrelated queue state or imply admission has resumed. Inflight/raw partial logs stay pinned until the receipt and operator attestation are durable. No success status is posted, no child completion is invented, and no full-suite timing sample is added. Subsequent resume/republish/pass guards report terminal interruption; the supported way forward is to fix/advance the code and run a suite on a new head. This completes proof repair rather than merely clearing a queue tombstone.

#### Authoritative state and sequence tests

The pure model owns the transition table and produces permitted effects. The IO shell feeds confirmed publication/launch/completion/reconciliation outcomes back to it; it cannot turn an error into an inferred absence. The exclusive parent lock is held through queue, execution, child commit and post; reread after waiting. The immutable link is the fence if that lock's process dies. Existing automatic knownFlakeStep keeps its unchanged v2 record semantics and is not retrofitted with resume links.

| State / event | Next state and permitted effects |
| --- | --- |
| Eligible / resume requested | Reserve child and publish exactly one link; no launch before durable link |
| Link publication fails or is unconfirmed | Refused/unknown; reread exact bytes, no launch or fresh link |
| Linked / launch and managed completion confirmed | Commit exact reserved child; validate actual JSON PASS facts and current registry/policy |
| Linked / owner death, cancel or child publication unconfirmed | Interrupted/unknown; retain bundle, no duplicate launch or success |
| Child committed and verified / post reply lost or repeated resume | Revalidate and repost same composite proof; no execution |
| Child fails, times out or omits required PASS | Terminal failed; no pass and no second child |
| Unknown / authorized exact reconciliation with absent owner and operator stop attestation | Commit terminal receipt; retain evidence, no pass and no second child |
| Reconciled / repeated identical reconciliation | Return same terminal receipt; no new mutation |
| Any / conflicting identity, digest, owner, receipt or unreadable store | Refuse; preserve bytes and existing authority |

Actions are resume once, wait/cancel, reconcile interruption, fix/advance code, and republish verified proof. Covered: owner death, concurrent callers, changed evidence/policy, repeated requests and lost status replies. Not covered automatically: descendant discovery after a killed owner, repairing corrupt evidence, or retrying an interrupted parent. Those cases cannot gain successful authority. The visible-state table above remains applicable, with interrupted/uncertain now naming reconcile interruption; after reconciliation it is terminal failed and names fix/advance code. Possible loss remains uncommitted targeted completion/PASS evidence, never original full-suite history.

#### Current review boundaries

**M1 — Link/proof/reconciliation/retention core.**

- [x] TDD `decodeResumeEvidence`, `resolveResumeProof`, `resumeTransition`: strict byte-boundary fuzz/property tests and generated event sequences; independently assert one child identity and that uncertainty/failure never authorizes pass or execution.
- [x] TDD `lockResumeParent` and `publishResumeEvidence`: real-process contention with deterministic publication/rename fault barriers; assert exact inode/raw-byte identity, exclusive publication and no effects before the lock/fence.
- [x] TDD `retireResumeBundle` through `pruneSuiteAttempts`: deterministic crashes/deletion/fsync faults at every retirement step; assert no retained selectable parent ever loses its execution fence. Pending evidence is pinned, terminal bundles age together, and linear-work counters enforce the snapshot envelope. Update atlas and close M1 through its gate.

**M2 — Resume/reconcile CLI and every success reader.**

- [ ] TDD `resumeSavedSuite`: stateful registry/GH plus private real Git/process fixtures perturb IO and process order; independently count one exact-package rerun, verify original SHA/bytes/full duration, and reject duplicate execution or insufficient PASS evidence.
- [ ] TDD `resolvedSavedRecord`, `loadSavedAttempt`, `republishSavedPass`, `guardSuitePass`: inject changed or conflicting linked evidence and policy around reads/posting; all consumers must derive the same result and preserve unrelated failure guards.
- [ ] TDD `reconcileResume`: real owner-death barriers and operator/:0 authority fixtures; exact attestation and durable receipt yield terminal interruption only, and repeated/lost-publication recovery cannot alter proof or run work.
- [ ] Private production-main CLI conformance drives a real registered suite failure through resume, status publication and the actual pass guard, plus crash/reconcile through the same commands. Use private ghFake/tracker/helper/lock/scratch, counting test sentinels and retained immutable bytes as independent oracles. Never repair live evidence pre-land.
- [ ] Update delivered help/README/atlas and eligible failure recovery output. Push final head, record the required affected/full check through sdlc local-check, close M2/issue through their reviews, PR/merge. TL-approved acceptance is no stage; private production-CLI proof is the gate, and the TL arranges the first natural post-landing live use.

ARCH-CONSTRAINTS/DRY: one bounded snapshot per decision and final revalidation, indexed O(N) over attempts/links/receipts, no per-package rescans. Count index/projection visits and command launches at N/2N and assert linear work. Known-failure count, record/log byte caps, one weight-1 lease and aggregate targeted timeout stay existing limits. Parent/receipt writer checks reserve room for its bundle; incomplete budget refuses safely. ARCH-FUNERAL: no new permanent lock-file family; open parent handles die with their owner, unknown bundles remain bounded by writer refusal until explicit terminal repair, terminal bundles share the existing retention/pruner. No credential surface is added. This is the first phase-2 recovery finalization, not an implementation bypass; approval and change-code plan-quality still precede code.

### 2026-10-21 10:48 PDT — Plan-quality PQ-1/PQ-2 refinements

PQ-1 (durable-retirement-monotonicity): `retireResumeBundle` acquires the same
parent inode lock and validates an expired, terminal bundle from a complete
snapshot. Remove the parent's selectable attempts/ID.json first and fsync the
attempts directory successfully before touching its link or child. Failure or
uncertain sync stops retirement, retaining the fence. Only after that durable
parent removal may child/receipt/log members be removed; the link is removed
last. A crash at any step therefore leaves either a selectable parent plus its
fence, or an absent parent that cannot execute. On retry an orphan terminal link
can finish that same parent-first deletion (including a fresh directory sync);
a pending/unknown orphan remains retained. Any cleanup error keeps remaining
members and does not restore authority. All members use terminal bundle age,
never an independently aged parent. Production prune bypasses its legacy
individual-file deletion for every indexed bundle member. Deterministic IO
barriers exercise the full deletion-order/failure class, including repeated
retirement, rather than testing only a successful cleanup.

PQ-2 (named-risk-function-test-contract): replaced the original and current
prose case inventories with the function-level strategies above. The risky
functions are the decoder, pure resolver/transition, lock, publisher, retirement,
resume runner, shared saved-record projection and reconcile. Individual cases
belong in their executable tests. No scope or operator-facing contract changes.
The TL's authenticated 928fcd1f-3d96-4c04-8231-025c9adaf064 also approves the
phase-2 delta and no-stage private acceptance before this gate rerun.

### 2026-10-21 11:10 PDT — Terminal witness outlives fence retirement

Authenticated TL approval aad50f3d-c5e6-4b21-9cd1-ff95d3fb46d6 refines the deletion order without changing authority: parent unlink and successful attempts-directory fsync first, then ordinary logs/inflight, then link removal plus directory fsync, then terminal child/receipt witness last. A child's raw log is part of that witness and stays until link removal is durable. Strictly deleting the link last could lose the only evidence of terminality before a crash, making its retained orphan indistinguishable from a pending resume forever. The parent-first durability invariant still prevents execution in every intermediate state; keeping the witness allows bounded cleanup retries. Shared logs/members are retained rather than deleting another bundle's evidence. Pending or unverifiable evidence is never garbage-collection authority.

### 2026-10-21 11:18 PDT — Re-plan snapshot/publication/retirement locking

Self-review found a race spanning the old pruner and new publisher: a new link can appear after the pruner's protected-member snapshot, then its legacy deletion pass can delete the original parent using that stale snapshot. Parent inode locks alone cannot fix it: publication may complete and release the parent before the pruner later takes it. Per-parent rescans would duplicate bounded-store IO quadratically. This section re-plans that entire lifetime boundary; implementation of this part waits for hub approval.

Use one native flock on the **existing attempts directory inode** as the snapshot/publication/retirement mutex, with path/inode revalidation. It introduces no permanent lock-file family. Link/receipt publication takes parent lock → directory lock, rereads one snapshot/capacity, revalidates exact parent bytes and publishes/syncs. Full pruning takes directory lock before its one snapshot and holds it through both bundle and legacy cleanup. Every parent acquisition by the pruner is strictly nonblocking; a busy parent is skipped, never waited on while holding the directory lock. Thus the opposite acquisition orders cannot form a wait cycle. The parent lock still exclusively owns one executor for its entire runtime; directory locking covers only bounded store mutation/GC, never queue or test execution. Existing ordinary attempt completion need not take this mutex: a visible pending link pins its reserved child ID, and terminal bundle retirement already holds its parent lock.

`lockResumeDirectory` and a nonblocking `tryLockResumeParent` share inode/type validation with the existing lock helper. `pruneSuiteAttempts` consumes an already-locked snapshot through the bundle index and legacy pass. Tests use real process/order barriers at snapshot/publication boundaries to prove that a waiting publisher's locked parent survives legacy cleanup, that a pruner sees a completed fresh bundle despite old parent mtime, and that prune-first parent removal causes publication refusal. Kernel release on owner death and two-pruner serialization are tested through the same seam. A counted scan guard still requires linear snapshot/index work. The approved terminal-witness-last deletion order remains unchanged. This resource closes the whole race class rather than adding observations to individual call sites.

Approved by authenticated TL f21f7022-6260-451b-86ae-a0a56158abac. The inverse order is safe only because every pruner parent acquisition is strictly nonblocking; code comments, atlas and a real-process negative-control test defend that invariant. This is the second retirement change (deletion order, then locking); a third stops implementation for a whole-retirement re-plan.

### 2026-10-21 — M1 review: supported receipt publication and staging lifecycle

BR-1: the publication seam must reject unsupported proof, then validate the supported parent/link/receipt graph explicitly. The general passing-reader refusal cannot be used before publishing a receipt, because its existing link is required input to that operation. Successful receipt publication through the real seam joins decoder-only tests, with duplicate/unknown-byte preservation controls.

BR-2: a publisher creates `.resume-*` temporary files while holding the existing directory mutex; normal completion removes them. Process death releases that mutex and leaves uncommitted residue. The next publication or production pruning pass, under that same mutex, streams/removes leftover regular files in this reserved staging namespace and syncs the directory. No age/PID inference, new artifact family, lock order or bundle-retirement design is added. A killed-publisher subprocess proves that active staging survives competing pruning, then crash residue is reclaimed. Existing JSON byte/record capacity remains authoritative for committed records; per-publication staging cleanup prevents repeated pre-commit crashes accumulating outside those bounds.

### 2026-10-21 — Whole retirement re-plan after M2 BR-3 (approval pending)

M2 review reproduced a class failure: cleanup authority must outlive every
protected artifact, but `publishSuiteAttempt` removes the inflight owner marker
before deferred worktree cleanup. A crash during status publication strands the
locked checkout; repeat resume only reposts and reconciliation refuses a
committed child. This is the third retirement design change, so stop
implementation and re-plan retirement as a whole rather than patching only the
reported window. Existing parent-first/witness-last evidence retirement and
parent→directory / directory→strict-TRY-parent locking remain unchanged.

| Integration entity | Lives in | Status | Wraps |
| --- | --- | --- | --- |
| Protected resume environment | localcheck_resume_environment.go, localcheck_resume.go | modified | Existing inflight owner marker, Git lock/registration, managed group cleanup, scratch and child publication |
| Bundle retirement eligibility | localcheck_resume_retention.go | modified | Existing protected-artifact predicate plus terminal bundle proof and directory/parent locks |

**Single lifecycle invariant:** a protected checkout/scratch always has a
durable, discoverable owner binding until its retirement is confirmed. A
committed child implies its execution inputs already retired; test outcome
proof and cleanup uncertainty never substitute for one another.

Proposed order: hold the immutable parent inode for all actions; reserve child
and create an initially unlocked detached checkout; publish/sync the immutable
link before any protection or execution; write/sync the existing owner marker
and its inflight directory before taking the exact child Git worktree lock;
then launch through the existing managed suite batch. On ordinary completion,
strong process-group cleanup and scratch retirement must be confirmed first.
Validate the same reverse pointers/child reason/SHA and retire the checkout
with confirmed Git registration/path absence **before** publishing the child.
Only then may existing child publication remove the inflight marker. A lost or
failed cleanup response leaves the link pending, raw log and marker retained;
no child completion or passing status is published. Recovery remains the exact
operator/:0 reconciliation with owner absence and descendant-stop attestation,
which durably publishes terminal interruption before attempting cleanup.

| Lifecycle / crash window | Visible state | Recovery | Possible loss |
| --- | --- | --- | --- |
| Before durable link | no resume executed; unlocked temporary checkout/inflight only | existing dead-owner/age collectors | unused preparation only |
| Link before marker | pending, no launch; checkout unlocked | operator reconcile; ordinary age collector may retire unlocked checkout | no test execution |
| Durable marker before Git lock | pending, no launch; discoverable unlocked inputs | operator reconcile; no lock to strand | unused preparation |
| Git lock through execution or uncertain group cleanup | pending/unknown with owner binding and protected inputs | wait/cancel, then exact attested reconcile | uncommitted targeted completion |
| Group/scratch confirmed; checkout retirement or reply uncertain | pending; owner binding retained until absence confirmed | exact attested reconcile, retry converges if Git already removed | actual completion may remain uncommitted |
| Checkout retired, before child commit | pending, no protected inputs | exact reconcile; raw partial/complete log retained | uncommitted targeted PASS |
| Child committed, including status-post crash | terminal child; no locked inputs | resume/republish validates and reposts; no execution | status reply only |
| Terminal receipt before/after cleanup, including lost reply | interrupted; binding retained until inputs gone | repeat same reconcile | no original suite history or receipt loss |

Bundle pruning holds the existing attempts-directory mutex across one snapshot
and all cleanup, takes only nonblocking parent locks and keeps the approved
parent-first durable deletion / terminal-witness-last order. Before retiring
any bundle or orphan terminal receipt, it must retain the entire bundle when
an existing inflight owner marker still protects a Git-locked checkout or its
authority is unreadable. A terminal receipt proves interruption, not cleanup;
age alone cannot authorize deleting that marker. Once exact cleanup confirms
absence, normal terminal retention applies. This uses no new artifact family,
lock files, marker schema, observer-based descendant discovery or retry grant.

**Simplest alternatives assessed:** deferred cleanup alone demonstrably loses
on process death. Keeping the checkout unlocked throughout permits stale
collection while unknown descendants still consume it. Keeping markers after
child commit can work, but adds a completed-child cleanup recovery path and
changes the generic publisher; retiring confirmed inputs before committing the
child avoids that extra state. PID absence cannot replace explicit descendant
attestation in an uncertain execution. The conservative loss of uncommitted
PASS on cleanup uncertainty already matches the approved recovery contract.

**Function-level test strategy:** extend production-CLI conformance with real
process barriers at pre-link preparation, durable link/before lock, running,
checkout removal/before child commit, committed child/status post, and terminal
receipt cleanup. Kill only fixture owners and explicitly stop fixture
subprocess groups before attestation. Independently assert no duplicate launch,
unchanged parent bytes, owner-marker discoverability whenever Git is locked,
and no protected checkout after a committed child. Private Git environment
integration tests inject marker/directory sync failure and lost remove replies;
retry confirms absence without widening authority. Production-pruner tests age
terminal receipt/child bundles with protected and unreadable owner markers:
assert byte-for-byte preservation until cleanup, then bounded retirement. Keep
existing deletion/fsync faults, real-process lock inversion negative control,
linear N/2N snapshot/index/log reads and exact-package counters. Run targeted
and race verification, full private CLI acceptance and a fresh recorded suite
on the eventual pushed final head, then rerun the mandatory M2 gate.

Read-only lifecycle critique clarifies the same proposed invariant: sync the
marker, child inflight directory **and its inflight parent directory** before
Git lock/launch (the latter makes the child directory name durable). Existing
scratch cleanup ignores removal errors; resume must explicitly confirm scratch
path absence and propagate uncertainty before checkout removal or child commit.
Confirm both the checkout path and Git worktree registration absent after Git
removal, rather than treating an exit status alone as retirement proof. Factor
the exact-binding/removal core behind two distinct authorities: live parent-lock
owner with confirmed managed-group cleanup, or durable interrupted receipt plus
the approved absent-owner/operator attestation. The receipt wrapper retains all
existing authorization checks; normal retirement cannot forge a receipt or call
its destructive path without its own lifetime precondition. These complete the
whole-lifecycle proposal; approval remains pending.


Whole-lifecycle re-plan approved by TL receipt
91596b33-cfe5-4d0d-bfff-1cc377ee1656. The next review finding of another
retirement/lifecycle gap means stop and return to the TL for operator review,
not another autonomous patch. Implementation keeps the same marker schema and
locks. Cleanup removes/confirms/syncs the checkout path while locked Git
registration still preserves its reverse pointer; only then does it remove and
confirm/sync registration absence, and finally release/sync the owner marker.
Scratch removal likewise confirms absence and syncs its parent before ownership
can disappear. A receipt for prepared-but-unlocked inputs is terminal proof
repair; it leaves unprotected unused preparation for the ordinary age collector
rather than weakening exact reverse-child deletion authority.

Crash-window test mapping (one named test per row): `TestResumeCrashPreLink`,
`TestResumeCrashLinkBeforeMarker`, `TestResumeCrashMarkerBeforeLock`,
`TestLocalCheckResumeConformance/running`, `TestResumeScratchRetirement`,
`TestResumeCrashCheckoutRetiredBeforeChildCommit`,
`TestResumeCrashCommittedChildStatusPost`, and
`TestResumeCrashTerminalReceiptBeforeCleanup`. Integration tests additionally
fault marker/directory sync, retain protected/unreadable GC authority, mutate
owner bindings and retry a removed checkout with extant locked registration.

### 2026-10-21 — Stop after whole-lifecycle re-review (operator escalation)

M2 re-review leaves BR-3 open: cleanup retry skips prior or unacknowledged
registration-directory and owner-release sync obligations on already-absent
entries. This meets the approved stop condition. Preserve the current code and
review evidence; do not implement another lifecycle refinement or seek another
incremental TL approval. The TL takes the failure to the operator. The plan is
not complete and the lifecycle invariant is not claimed satisfied.


### 2026-10-21 — Fresh M3: operator option A supersedes M2 (TL plan review pending)

**Decision and scope.** TL receipt
390faec8-dc49-498e-9ad5-2c1a718c2c37 relays the operator's explicit selection
of A. This revision supersedes M2 and its retry-capable environment cleanup;
M2 never closed and its review/BR-3 evidence remains intact. M1 remains closed.
M3 is a fresh mandatory milestone boundary covering the delivered CLI/readers
and the simplified lifecycle together, not a claim that the old review passed.
Implementation waits for TL review of this plan, then fresh `change-code`
plan-quality admission. The final issue gate must explicitly dispose of the
carried M2 BR-3 against the operator's new policy; no manual ledger reset or
review bypass is planned.

**Goal:** retain the targeted-resume feature while removing cleanup recovery as
an automatic capability. Exactly one cleanup pass belongs to the original live
owner. Any failure, crash or inconsistent first-pass observation stops further
cleanup, leaves surviving files and persistent Git locks as observed, and
reports needs-reconcile with the existing exact command. Reconcile writes or
returns immutable terminal interruption only. It never deletes, unlocks,
releases markers, retries cleanup, grants execution or posts a pass. Already
completed deletions cannot be restored; retaining means no subsequent mutation
of the surviving inputs. No new persistent cleanup state, one-shot marker,
record variant, authority or manual cleanup command is introduced.

#### Entities and integration changes

| Pure entity | Lives in | Status |
| --- | --- | --- |
| ResumeProof / transition presentation | cmd/sdlc/localcheck_resume_evidence.go | modified: existing pending/unknown maps to needs-reconcile; execution and pass rules unchanged |
| Resume retention eligibility/index | cmd/sdlc/localcheck_resume_retention.go | modified: pending/interrupted/receipt graph is permanently protected; completed-child retention unchanged |

| Integration | Lives in | Status | Wraps |
| --- | --- | --- | --- |
| Original-owner cleanup pass | cmd/sdlc/localcheck_resume_environment.go, localcheck_resume.go | modified | existing parent lock, current owner incarnation, exact link/marker/reverse Git bindings and one forward deletion sequence |
| Reconcile evidence only | cmd/sdlc/localcheck_resume.go | modified | existing operator/:0 authority, owner absence, descendant-stop attestation and immutable receipt publisher |
| Receipt-authorized cleanup and replay | cmd/sdlc/localcheck_resume_environment.go | deleted | remove receipt adapters, retry classification/reconstruction and replay-specific control flow |
| Managed strict scratch retirement | cmd/sdlc/localcheck_run.go, localcheck_resume.go | modified | retain scratch on managed errors/uncertainty; one confirmed cleanup pass on ordinary completion |
| Retained-bundle GC guard | cmd/sdlc/localcheck_resume_retention.go | modified | existing one snapshot, directory mutex, nonblocking parent locks and protected-ID projection |

ARCH-DRY/PURE: preserve the one classifier, immutable proof resolver, exact
package batch, queue/lease, evidence publisher and existing transition model.
Do not add another cleanup model, durable retry journal or file-absence success
classifier. The single parent lock plus immutable link already ensures the
original owner is the only cleanup caller: a later resume of pending proof
refuses, and completed proof only revalidates/reposts. Reconcile has no cleanup
call at all. This is why no extra one-shot boundary is needed.

#### Forward-only first-pass algorithm and deletions

Keep the approved preparation order: parent lock → reserve child/unlocked
checkout → durable link → durable existing owner marker and child/inflight
parent directories → exact child Git lock → managed execution. Run one weight-1
batch and one aggregate timeout as before. Ordinary completed test failure is
still a real terminal failed child; it never grants another execution or pass.
Cancellation, timeout, launch failure, unconfirmed group completion or IO
inconsistency reports needs-reconcile and retains surviving inputs.

For an ordinary managed completion, confirm group shutdown and execute one
scratch cleanup attempt. Confirm and sync absence once, then perform one exact
checkout removal while locked registration still preserves its reverse pointer;
confirm/sync checkout absence, remove Git registration once, confirm/sync both
namespaces, then release/sync the existing marker once. Check cancellation
before each subsequent effect. Any error stops immediately; do not continue
later deletions, retry an effect, reinterpret missing metadata as completion or
invoke deferred cleanup after the link. Child publication follows only a
completely successful first pass. If publication later fails/crashes, the link
still fences execution; a committed visible child can be revalidated/reposted,
otherwise the operator can only terminalize it as interrupted.

Delete these paths rather than repairing their durability retries:

- Both reconcile calls to `retireResumeEnvironment`; receipt-exists returns its
  already validated interrupted audit result, new receipt returns after durable
  publication. Neither path calls Git removal/unlock, scratch removal or marker
  release. Owner/authority/digest and committed-child refusal checks remain.
- Receipt cleanup adapters, `resumeEnvironmentReceipt`, `receiptCleanup` mode,
  and receipt-only prepared-unlocked cleanup/no-op classification. Terminal
  proof resolution already validates receipt/link identity through the shared
  reader; it does not need a destructive adapter.
- Missing inflight/marker/registration success paths, empty-binding release,
  missing-checkout reconstruction, interrupted-registration replay, and missing
  marker release success. First-pass preconditions require present canonical
  directories, exact marker and lock, both Git reverse pointers and exact SHA.
  Once this invocation removes the checkout, its already validated local binding
  supplies the next registration step; never rerun a general inspector that
  reconstructs a previous invocation's partial cleanup.
- Retry-specific scanned-root resynchronization and convergence loops. Keep
  syncs for this invocation's actual deletions; a lost reply/error is a stop,
  not permission to infer absence and continue.
- Success-on-retry tests and documentation. Replace them with independent
  retained-byte/lock and zero-effect assertions.

The strict managed scratch defer must not remove scratch when the managed
result is canceled, timed out, did not launch or has unconfirmed cleanup. It
must not run a second cleanup after the first cleanup returned an error.
Preserve actual logs and exact needs-reconcile diagnostics on all such returns.
Persistent Git locks on protected inputs remain; kernel flock handles naturally
close on process death. Unused pre-link/unlocked preparation has not launched a
child and retains the existing ordinary temporary-collector policy; the new
resume/reconcile code never tries to reclaim it after an inconsistent outcome.
This distinction is explicit for TL review, not an inferred deletion authority.

#### Evidence retention policy and cost

Pending links remain pinned as today. Any terminal receipt, including an orphan
receipt with missing parent/link, permanently pins all identified records,
canonical raw logs and inflight members regardless of marker presence, lock
visibility, age, or a partly completed cleanup. Remove terminal-receipt bundle
retirement and the orphan-receipt retirement loop. Keep their protected IDs in
the returned projection so legacy individual-file/temp cleanup cannot delete
members. Unsupported/unreadable/ambiguous evidence continues to refuse GC.

Completed-child bundles with successful live-owner cleanup retain the existing
30-day GC, parent-first durable deletion, terminal-child/log witness last and
existing directory→strict-TRY-parent locking. A failed test child with fully
successful cleanup is still a completed child for this retention rule; test
result acceptance is independent. Do not introduce an automatic cleanup retry
for retained input artifacts or change unrelated suite/runner cleanup policy.

ARCH-FUNERAL/CONSTRAINTS: intentional indefinite retention is the operator's
tradeoff. Each uncertain resume can retain one reserved child inflight/raw log,
one detached checkout, its existing scratch and Git registration/lock, one link
and at most one receipt, plus original failure evidence. Repeated resume or
reconcile adds no children, cleanup attempts or receipts. Input bytes follow
the existing suite's actual output/build size, so no new byte cap is claimed.
Existing bounded evidence publication/read budgets still refuse further proof
rather than evict retained authority; no emergency automatic pruning is added.
Diagnostics name retained paths so the operator can inspect actual disk use.
Removing retained inputs later is a separate operator decision, not this CLI.

The scan remains O(A+L+R) for attempts, links and receipts, one indexed snapshot
per decision; receipt pinning adds constant indexed work per receipt and skips
its deletion. Existing N/2N tests stay; add counted N/2N retained receipts with
zero removal calls, including orphan cases. Normal first-pass input cleanup
scans scratch roots once and checks each candidate once; no retry scan remains.

#### Visible states, actions and loss

Actions: resume once, wait/cancel, reconcile interruption, republish completed
proof, fix/advance the code. There is no cleanup-retry action.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Eligible saved failure | ready to resume | resume once | none |
| Active owner with durable fence | running or queued | wait/cancel | uncommitted child on cancellation |
| Any crash, managed uncertainty, cleanup failure or first-pass inconsistency | needs reconcile, exact parent/digest command and retained paths | reconcile interruption after existing owner-absence/descendant attestation | uncommitted targeted PASS; already completed unlink effects are not reversed |
| Interrupted receipt, including repeated reconcile | interrupted; retained, no cleanup attempted | fix/advance code | retained disk cost, no original history loss |
| Child committed and valid | pass or terminal test failure with linked evidence | republish completed proof, or fix/advance code for failure | status reply only |
| Corrupt/changed/unreadable authority | refused, evidence retained | operator inspection | no automatic repair claim |

Covered: one execution and cleanup caller, no cleanup on reconcile (first or
repeat), raw original evidence, proof invalidation, all crash/error windows and
permanent interrupted/receipt retention. Not covered: automatic descendant
inference, cleanup repair after uncertainty, storage reclamation for pinned
bundles, restoring deleted files, or corrupt-evidence repair. These omissions
are explicit operator-selected boundaries, not future automatic recovery.

#### M3 tasks and verification (single fresh boundary)

1. **TDD evidence-only reconciliation.** In localcheck_resume_test.go,
   localcheck_resume_environment_test.go and localcheck_resume_conformance_test.go,
   use real private Git/tracker/GH fixtures to create a pending link with locked
   inputs. Snapshot every surviving byte, path/lock and destructive-call count.
   First and repeated reconcile must publish/return the same receipt and perform
   zero removals, unlocks, marker releases or launches. Run the focused test and
   observe RED against current receipt cleanup, then remove the adapters/calls
   and rerun GREEN. Preserve live-owner/foreign-host/digest/authority/committed
   child refusal and raw unknown-field preservation tests. Commit this deletion.
2. **TDD one-pass failure behavior.** Replace lost-reply and partial-registration
   convergence tests with a deterministic fault matrix over each forward remove
   and sync effect (scratch, checkout, registration, marker). After the first
   injected failure, assert no later mutation; after owner exit, resume refuses,
   repeated reconcile has zero destructive effects, all surviving bytes/locks
   remain and exact needs-reconcile command is printed. Kill fixture owners at
   each existing crash-window barrier and explicitly stop fixture descendants
   before attestation. Observe RED, simplify inspect/core and strict scratch
   lifecycle, rerun GREEN and commit. No new persistent state or retry loop.
3. **TDD permanent interrupted retention.** In retention tests call the real
   production pruner with aged pending, interrupted and orphan receipt bundles;
   vary missing/unlocked/changed/unreadable markers and missing parent/link.
   Assert every remaining member/log is byte-identical and removal count zero,
   including repeated pruning. Add N/2N count checks. Keep the completed-child
   deletion-order/fsync failures, shared-member protections and real-process
   nonblocking-lock negative control unchanged. Observe RED, remove receipt
   retirement eligibility/loop and rerun GREEN. Commit.
4. **Delivered contract.** Update README, embedded local-check help, atlas and
   eligible refusal/reconcile output: one pass, needs reconcile, evidence-only
   reconciliation, retained paths/disk cost, no cleanup retry, no inferred
   completion and no automatic receipt GC. Keep original execution/proof rules.
   Review all references to retireResumeEnvironment/receiptCleanup/retry for
   dead code/tests and contradictory documentation; add lessons for the class.
5. **Final acceptance and gate.** Run `go test ./cmd/sdlc -run '^TestResume'
   -count=1 -race -timeout=5m`, focused legacy saved-proof/republish/known-flake
   checks, Go/shell lint and `git diff --check`. Run the private production-main
   conformance with `SDLC_PRIVATE_RESUME_CONFORMANCE=1 go test ./cmd/sdlc -run
   '^(TestResumeCrash|TestLocalCheckResumeConformance)' -count=1 -v`: actual
   pre-registration failure → exact package resume → pass/post/guard, repeat
   without execution, crash → attested receipt without cleanup, repeated
   reconcile/prune preservation, later actual green refused. Record exact
   evidence and all eight named windows' new oracles. Push final code, queue
   `sdlc local-check run --name suite`; follow required freshness/integration
   gate remedies rather than waive them. Close M3 with its fresh review, then
   whole issue with carried finding disposition, PR/merge and report to TL.
   No stage; TL's private acceptance/natural post-land live-use policy remains.

No code is admitted by this plan write. Send its committed/pushed revision to
TL, wait for plan acceptance, then run the fresh change-code plan-quality gate.
M2 sidecars and the issue-wide ledger are preserved as review history.
