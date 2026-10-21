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
