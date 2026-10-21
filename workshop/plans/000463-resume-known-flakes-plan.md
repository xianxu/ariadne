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

### M1 — Exact-package reruns and reusable proof model

**Files:** internal/flakes/rerun.go, rerun/flakes tests; localcheck_run.go; localcheck_flakes.go and tests; new localcheck_resume_evidence.go/tests; teststats_store.go/tests; atlas/workflow/sdlc-binary.md.

- [ ] RED: a two-package fixture has TestA fail only in package A and TestB only in B, with same-named decoys that write sentinels if over-selected. Keep subtest/sibling controls from #466. Observe the current cross-package command fail the sentinels.
- [ ] Add deterministic per-package command planning and the bounded sequential managed batch. Ordinary checks wrap a single argv; only the known-flake path uses batches. Exercise fail-fast, aggregate timeout, cancellation between commands, inherited-pipe/process cleanup, one capacity acquisition, and runner single-command regressions.
- [ ] Add typed ResumeLink decoding/publication and a pure resolver over supplied records/log facts. Test orphan/wrong-parent/wrong-repo/wrong-SHA/digest/child-link combinations; pending link can never resolve to pass. Keep an older decoder fixture proving unknown-version refusal.
- [ ] Route automatic reruns through the exact-package batch; target existing flake, affected, queue/signal and runner regressions. Update the atlas for the new proof seam. Close M1 through sdlc milestone-close.

### M2 — Explicit resume, all proof readers and delivery

**Files:** new localcheck_resume.go/tests and process/CLI tests; localcheck_cmd.go; localcheck_republish.go/tests; localcheck_flakes.go/tests; teststats_store.go/tests; teststats.go only for shared projection if needed; helptext/local-check.md, README.md, atlas/workflow/sdlc-binary.md.

- [ ] RED: create a complete saved failure through the real attempt writer, register it in a stateful registry, invoke the resume seam and assert no full command, only exact-package argv, one linked child, the original parent/log bytes unchanged and a known-flake status. Assert no new full-suite timing/run sample.
- [ ] Implement strict CLI selection, digest/identity verification and lifetime parent lock. Reserve child/write link before launch, then use the shared classifier/batch/resolver. Carry real slot/agent identity and registry issue links.
- [ ] Route saved pass selection, republish, guardSuitePass and repeated resume through one evidence snapshot/resolver. A failed child or unresolved link blocks a later otherwise-green full attempt. Do not make unrelated readers duplicate classification.
- [ ] Adversarial stateful tests interleave two real subprocess callers under the same parent lock; count one execution, then idempotent repost. Kill the first owner after link publication/before child publication and assert the second refuses unknown rather than duplicating. Cancel a managed child, change registry/policy midrun, corrupt either log, add a later unexplained failure, and lose the GH post reply. Every refusal preserves bytes and posts no success.
- [ ] Exercise actual exact-SHA detached Git contents while caller HEAD differs, missing/unpushed refs, baselines, legacy records, disabled/accept policies and malformed/partial store reads. Reuse current fake GH/registry and real private Git/helper, not production services.
- [ ] Update delivered help/README, print the recovery command for eligible saved failures, and document retained uncertain links and existing operator reconciliation. Record private executable conformance against a tiny fixture repo with fake GH, private queue/lock/scratch and actual log/record inspection; no live failed suite is repaired pre-land.
- [ ] Push final head, central recorded suite, M2 close, whole-issue close and normal PR/merge. The hub's :0 rebuild precedes first production resume; report the exact supported verb/evidence semantics.

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

- [ ] RED pure resolver and transition tests: exact link identity/digests; missing child; same-parent conflicting links; malformed/unknown records; unsupported v3 kind; terminal failure/interruption cannot become eligible or pass. Unknown fields/large opaque numbers remain preserved in raw bytes.
- [ ] Implement typed union/snapshot indexes, exclusive link/receipt publication and cancellable parent-inode lifetime lock. Real subprocess contenders prove one link/child reservation, and replaced-path/digest races refuse.
- [ ] Add link-aware bundle retention before enabling a writer. Pending bundles pin parent/link/inflight/partial logs; terminal bundles retain all members for 30 days from their latest completion/receipt, then prune under the same parent lock. Incomplete/unreadable evidence is retained. Preserve byte/record bounds and refuse new links at capacity rather than dropping unknown evidence. Exercise aged parent versus fresh child, live lock, owner death, partial reads and repeated prune. Update atlas; close M1 through its gate.

**M2 — Resume/reconcile CLI and every success reader.**

- [ ] RED production-path resume tests: actual saved failed parent, current stateful registry, only exact-package argv on original pushed SHA, one child, original bytes unchanged, no full command or timing sample. Reuse existing real-Go PASS controls from #524.
- [ ] Implement explicit suite-only selection, declared policy/current main pin, real detached worktree, reserved child/link lifecycle, managed batch and before-post fresh policy/registry revalidation. Policy accept still requires actual rerun. Keep failure guard active for unrelated same-head failures.
- [ ] Route loadSavedAttempt, republishSavedPass, republishRevalidates, guardSuitePass and repeated resume through the same linked resolver. A completed valid parent/child supplies composite status, original full duration and issue links; child or receipt never becomes an independent suite. Unsupported/incomplete/failed proof blocks scoped acceptance. Test every reader.
- [ ] Implement the exact reconcile command, authority/owner/digest/attestation checks and terminal receipt. Real subprocess sequence barriers kill owner after link/before launch and after child start/before child commit; another caller refuses execution. Operator fixture proves dead owner plus surviving child is not automatic stopped proof; only explicit authorized attestation terminalizes. Tests retain logs and audit bytes on cancellation/error, publication failure and repair retry; no success after repair.
- [ ] Private production-main CLI conformance covers resume twice, cross-package decoys, caller HEAD drift, registry/policy changes, timeout/cancel, exact child commit then lost post reply, linked failure beside later green attempt, and reconcile after an injected crash. ghFake/tracker/private helper/lock/scratch only; no live failure repair before land. Compare CLI/dependency fake semantics through the same production seam.
- [ ] Update delivered help/README/atlas and eligible failure recovery output. Push final head, record the required affected/full check through sdlc local-check, close M2/issue through their reviews, PR/merge. Ask the TL for its phase-2 live acceptance/staging decision; prior #524 no-stage applied only to reader/batch work. First production resume waits for the hub's upgraded phase-2 runtime and acceptance choice.

ARCH-CONSTRAINTS/DRY: one bounded snapshot per decision and final revalidation, indexed O(N) over attempts/links/receipts, no per-package rescans. Count index/projection visits and command launches at N/2N and assert linear work. Known-failure count, record/log byte caps, one weight-1 lease and aggregate targeted timeout stay existing limits. Parent/receipt writer checks reserve room for its bundle; incomplete budget refuses safely. ARCH-FUNERAL: no new permanent lock-file family; open parent handles die with their owner, unknown bundles remain bounded by writer refusal until explicit terminal repair, terminal bundles share the existing retention/pruner. No credential surface is added. This is the first phase-2 recovery finalization, not an implementation bypass; approval and change-code plan-quality still precede code.
