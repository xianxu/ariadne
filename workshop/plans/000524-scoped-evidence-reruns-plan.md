---
collaboration-mode: discussion
---

# Reader-first evidence compatibility and exact reruns plan

> **For agentic workers:** Apply AGENTS.md Section 3; keep the shared evidence/queue integration in this session. The two real review boundaries are scoped readers, then managed rerun batches.

**Goal:** Land the safe reader/execution prerequisite for #463 without writing links or enabling resume.

**Architecture:** A typed store snapshot separates supported attempts, valid but unsupported proof identities, and genuinely unscopable read failures. Authority readers refuse only related unsupported proof. Known-flake execution uses package-specific argv batches inside the existing one-lease managed runner, with package-bound PASS evidence.

**Tech stack:** Go, strict JSON helpers, real private Git/process/lock fixtures, existing stateful ghFake and flake registry seam.

## Core concepts

| Name | Kind | Lives in | Status / boundary |
| --- | --- | --- | --- |
| UnsupportedProof | PURE | cmd/sdlc/teststats_evidence.go | new / M1 |
| SuiteEvidenceSnapshot | PURE projection; INTEGRATION bounded IO shell | cmd/sdlc/teststats_evidence.go and teststats_store.go | new/modified / M1 |
| ExactRerunBatch | PURE | cmd/sdlc/internal/flakes/rerun.go | planned / M2 |
| Package-bound rerun PASS facts | PURE | cmd/sdlc/internal/flakes/rerun.go | planned / M2 |

- **UnsupportedProof:** read-only identity, not an accepted attempt. Decode a duplicate-free bounded object with a positive unsupported version, exact ID matching its filename, valid repo/full SHA/check and optional valid rerun_of. Never infer a pass, duration, failures or absence from opaque newer content. A future ResumeLink must use the same envelope, check naming the original proof and rerun_of naming its parent. Preserve raw bytes; no rewrite or writer is added.
- **SuiteEvidenceSnapshot:** one bounded scan yields Attempts, Unsupported and unscopable Warnings. Map by repo/SHA/check/ID/parent so each authority decision uses one consistent snapshot. Ordinary statistics/actual/registration consume only supported real attempts, with unsupported coverage displayed separately. Unreadable files, invalid/ambiguous identity, duplicate keys or a reached read budget stay incomplete; do not turn them into an empty store.
- **ExactRerunBatch:** deterministic literal argv per package, selected names anchored and regex-escaped by the existing validated-name functions. Promote failed subtests to whole top-level tests per #466. Never apply one package's name set to another package. One batch remains one suite-rerun attempt; no full-suite sample is invented.
- **Package-bound PASS facts:** newly executed batches use go test -json. Require structured pass events for every needed package/test/subtest, rather than a bare string match that a same-name test or quoted output can satisfy. Existing saved legacy proof is not rewritten.

| Integration | Lives in | Status | Wraps |
| --- | --- | --- | --- |
| Evidence IO | cmd/sdlc/teststats_store.go | modified | existing portable bounded store read |
| Proof consumers | localcheck_republish.go, teststats_cmd.go, flake evidence readers | modified | saved pass, republish, guard and coverage views |
| Batch execution | cmd/sdlc/localcheck_run.go and localcheck_flakes.go | modified | existing queue/slot/helper/process/scratch/attempt seams |

No external dependency is added. Tests reuse stateful fake GitHub/registry and real private filesystem/Git/native lock/process fixtures. No production evidence or machine setting is touched.

## Risky functions and adversarial strategies

| Function | File / colocated test | Strategy and mechanical guard |
| --- | --- | --- |
| decodeUnsupportedProof(data, filename) | teststats_evidence.go / TestDecodeUnsupportedProof | Perturb each identity boundary in a valid opaque newer record; strict duplicate/type/name validation must return either a validated identity or unscopable error, never accepted attempt facts. Raw input remains byte-identical. |
| unsupportedProofBlocks(proof, subject) | teststats_evidence.go / TestUnsupportedProofBlocks | Generate subject/proof identity pairs varying one authority dimension and parent relation; an independent expected-relation table guards locality and forbids proof fallback. |
| readSuiteEvidenceBudget(root, limits) | teststats_store.go / TestReadSuiteEvidenceBudget | Interleave supported, opaque and unreadable files under explicit read budgets; counted IO/visits at N and2N proves one scan and supported measurements survive only when their authority scope is complete. |
| loadSavedAttempt and guardSuitePass | localcheck_republish.go / TestScopedSavedProofConsumers | Stateful store/registry/GH fixtures drive unsupported evidence through actual post/reuse paths; count status effects and assert scoped newer-reader errors preserve bytes and never authorize related uncertainty. |
| RerunCommands(tests) | internal/flakes/rerun.go / TestRerunCommandsExactPackages | Generate named tests over overlapping packages; independently reconstruct package/name selections from emitted literal argv, require exact set equality after top-level promotion, and count groups rather than timing. |
| RerunMissingJSON(log, tests) | internal/flakes/rerun.go / TestRerunMissingJSON | Mutate structured event provenance and event completeness around valid pass events; require matching package/test authority, treating quoted Output or non-events as data and missing proof as inconclusive. |
| runUnderSlotBatch and runManagedCommandBatch | localcheck_run.go / TestManagedRerunBatch | Order real subprocess publication points with pipes/markers, then cancel or fail before the next launch; independent child/lease counters prove one admission, bounded aggregate lifetime, joined descendants and zero launches after cancellation. |
| knownFlakeStep | localcheck_flakes.go / TestKnownFlakeExactPackageGate | Real two-package decoy fixture behind stateful registry/GH proves the production path runs only selected packages, saves one child and never counts it as another suite. |

Pure decoder/relation/batch/PASS tests use supplied values only. IO/process tests exercise the same production entrypoints with private real resources or the existing stateful fake. The table defines the function-level strategy; executable tests own individual case inventories.

## Scoped compatibility rule

The parser recognizes supported v1/v2 attempts unchanged. For a valid future/v3 identity envelope, store it as UnsupportedProof rather than a global warning. Selecting that attempt or a linked parent returns a typed error with “needs a newer sdlc”, record ID and path. Do not silently fall back to an older passing proof of the same subject.

Unrelated repository/SHA/check proof remains available and the whole snapshot remains readable. The existing suite/affected success guard deliberately shares same-head failed-test evidence: an unsupported test-run record in that existing authority scope is related uncertainty, not an unrelated check. Ordinary named checks elsewhere are unaffected. An unsupported link naming a parent applies to that parent/proof; its opaque body cannot create authority over a different entity.

Use one shared predicate/projection at loadSavedAttempt, republish, republishSavedPass and guardSuitePass, and audit all readSuiteAttempts consumers. Do not filter unsupported objects out at the generic reader and leave a raw authority consumer unaware. Stats/actual may retain valid measurements without treating unsupported rows as runs or silently claiming full coverage. Malformed/unscopable files still fail authority closed because their relation cannot be proved; budget exhaustion remains incomplete because unseen identities are unknown.

## Exact batch execution

Replace the cross-product RerunCommand use in knownFlakeStep with a pure per-package batch. Retain an adapter only if a real caller needs it; no unsafe old cross-product caller remains. Use `go test -json -count=1 -run <anchored names> <one package>` per group. Package grouping may be combined only for identical name sets; a straightforward one-package command is preferred.

Generalize the suite-owned internal command execution seam to accept a nonempty batch; ordinary checks pass one argv. The outer queue admission, one weight-1 suite slot, run lease, private scratch, one total targeted timeout and raw log span the whole batch. Each package process uses the existing process-group/cancellation/pipe-drain behavior and is joined before the next starts. Stop on nonzero, timeout or cancellation; unexecuted required PASS events make the rerun inconclusive. No new process supervisor or runner-resource cleanup policy.

Keep runner-resource checks on their single-command path, including #509's30s grace and retention authority. Release suite capacity after the final owned process is stopped, before record/status IO. Existing environment/interrupted classification remains outside ordinary passing SuiteAttempt evidence. Log the exact argv for each group and record one linked suite-rerun result with real duration/slot/agent/queue data.

## Visible states and recovery

Actions: use supported proof, upgrade reader, repair evidence, wait/cancel, fix code.

| Condition | Visible state | Recovery | Possible loss |
| --- | --- | --- | --- |
| Supported validated record | ordinary proof | use supported proof | unchanged |
| Valid unsupported record related to requested proof | needs a newer sdlc | upgrade reader | delay; original bytes untouched |
| Unsupported record unrelated to requested proof | ordinary unrelated proof; scoped coverage note | use supported proof | unsupported measurement unavailable |
| Malformed/unreadable/unscopable evidence or exhausted budget | evidence incomplete | repair evidence | no fabricated absence/pass |
| Batch queued/running | targeted rerun active | wait/cancel | time spent; no pass until all required events |
| Any batch failure/timeout/missing PASS | rerun failed/inconclusive | fix code | both failure logs retained |

This landing creates no cross-event resume state, lock or durable coordination record. The existing managed run/queue lifetime owns active work. The new read-only values die with the snapshot; no new artifact family or persisted version is written.

## M1 — Scoped evidence reader and all authority consumers

**Files:** new teststats_evidence.go/test; teststats_store.go/test; localcheck_republish.go/test; teststats_cmd.go/test and existing flake/actual readers as the caller audit requires; atlas/workflow/sdlc-binary.md.

- [x] Write RED mixed-store tests through actual consumers: one valid future/v3 envelope, a supported parent and unrelated supported records. Explicit unsupported selection refuses newer-reader; unrelated repo/head/named-check reuse and republish succeed. Same-head unknown test evidence cannot silently authorize suite/affected success.
- [x] Implement decodeUnsupportedProof, unsupportedProofBlocks and readSuiteEvidenceBudget using the named adversarial strategies above; retain strict supported-record validation and untouched raw unknown bytes.
- [x] Route every authority consumer through the scoped view. Test actual/statistics/registration projection against unrelated valid rows, with explicit unsupported coverage and no fake samples. Count scans/record visits at N and2N; reject per-record whole-store rescans.
- [x] Retain a pre-change decoder/guard control proving a v3 record currently poisons the whole snapshot; test the upgraded compatibility reader against the same store and assert scoped refusal only. Do not label the already-built pre-change binary compatible.
- [x] Documented the two-landing rollout barrier in atlas; close M1 using the binary review. No v3 writer is present.

## M2 — Exact-package managed reruns and delivery preparation

**Files:** internal/flakes/rerun.go and flakes_test.go; localcheck_flakes.go/tests; localcheck_run.go plus process/signal tests; localcheck_affected_flakes_test.go; helptext/local-check.md, README.md and atlas/workflow/sdlc-binary.md.

- [ ] RED real two-package fixture: A fails TestA, B fails TestB; A.TestB and B.TestA write decoy sentinels if invoked. Existing cross-product execution must fail this control. Keep real parent/subtest/sibling coverage and named-test absence controls.
- [ ] Implement RerunCommands and RerunMissingJSON with the pure function strategies above; preserve the existing five-flake limit and revalidation policy.
- [ ] Implement runUnderSlotBatch/runManagedCommandBatch with the controlled-order strategy above. Fail-fast cannot become a pass when a later required group did not run.
- [ ] Route automatic suite/affected known-flake reruns through it, preserving one immutable linked child and original suite samples. Run focused normal/race regressions for flake gate, republish/saved pass, queue signals and single-command runner cleanup.
- [ ] Private executable proof with fake GH/registry/private Git, queue, scratch and lock paths: unrelated proof survives one synthetic future record, affected proof names newer sdlc; exact-package decoys remain absent and one real linked attempt is saved. No real v3 production record or live queue pause.
- [ ] Push, record the central full suite on final head and complete delivery preparation. Then M2 close, whole close, PR and normal central merge follow through their gates. Report the landing and fleet rollout instructions; do not activate #463.

## Operating envelope and lifecycle

Manual/batch verification path, no keystroke hot path. Existing bounded store caps remain10,000 records/64MiB,1MiB per record and bounded log reads. Snapshot/index cost is O(records+bytes), pure batch cost O(names log names) for deterministic order, with at most five registered failures in production. Count scans, visits and package launches for N/2N controls; timing alone is not the defense. Each batch takes one weight-1 lease and one existing15m targeted timeout capped by the check declaration, not one budget per package.

No new production files/rows/process families beyond existing attempts/logs/managed execution. The opt-in compatibility control creates a private verification clone, retained through inspection and removed by its invoker using the exact printed parent directory; it is not production state or indefinite evidence retention. Existing30-day attempt retention remains unchanged because no resume link is written. ARCH-DRY keeps one scoped authority projection and one batch runner; ARCH-PURE separates identity/selection from IO; ARCH-MOCK reuses stateful seams; ARCH-SECURE verifies identity before scoping; ARCH-ORDER preserves managed cancellation/extent; ARCH-CONSTRAINTS bounds reads and batch work; ARCH-FUNERAL adds no residue; ARCH-EVOLVE reads unknown versions as scoped unavailable without rewriting bytes. ARCH-PURPOSE forbids activating the later writer before fleet rollout.

## Rollout and approval

Hub receipt917bf022-032a-4d84-ac6b-39075bbca2b5 approves reader-first plus exact-package batches as a separate landing. This issue is that phase1. After landing, the hub upgrades every slot's actual sdlc and this worker records observed paths/revisions; any stale, unknown or still-running old writer keeps #463 link activation held. The pre-change whole-store limitation cannot be fixed by a payload. Never hide links from older readers. Phase2's resume-specific recovery command remains #463 work.

## Revisions

### 2026-10-21 — PQ-1 function-level test contracts

Plan review requested named risky functions and compact adversarial strategies instead of prose case inventories. Added one table covering the pure decoder/relation/batch/PASS functions, bounded reader, actual proof consumers, managed execution and flake integration; milestone steps now reference it. No design, scope, state or rollout policy changed.

### 2026-10-21 — M1 BR-1/BR-2 authority and reproducible compatibility evidence

Linked-parent authority must survive implicit as well as explicit selection even when a future record's check differs from its supported parent's check. The bounded read resolves parent edges into the same subject index once; all load/reuse/post consumers then share that authority. Production consumer regressions cover both selectors, merge reuse and zero success publication, with the exact cross-check fixture observed RED before correction. This completes the existing linked-proof rule, not a new state or rollout policy.

Retain the pre-change control as a committed fixture and opt-in scripts/verify-pre524-reader.sh. It executes the immutable pre-change source in a private shared clone, checking the real old read warning, saved-proof incompleteness, guard and publication refusal. Compare with the equivalent unrelated-store case in TestScopedSavedProofConsumers. Decoder rejection alone is not the old compatibility proof. Core concepts now distinguish PURE projections/INTEGRATION IO and current M1/planned M2 entities.

### 2026-10-21 — M1 review2 documentation and inverse parent scope

BR-3 adds README discovery for the opt-in compatibility control and fleet rollout barrier. BR-4 defines the private clone lifecycle: the invoker owns it, retains it through inspection, then removes only its printed parent directory. The committed fixture/script and issue Log remain the durable proof; no collector, retention setting or production artifact family is introduced.

The BR-1 class sweep also tests the inverse inconsistent-check case: a link to a supported VM parent must not gain authority over an unrelated suite from its own check field. The fixture reproduced that false refusal, then the shared index was narrowed to the resolved parent's check. Unresolved parents retain their declared scope as uncertainty. Explicit and implicit consumers use the same completed snapshot; the scan remains linear.


### 2026-10-21 — M1 finalized; M2 execution contracts

M1 round3 returned SHIP with BR-1 through BR-4 addressed and no new findings;
the binary finalized and pushed its boundary. M2's real two-package decoy
fixture failed against the original cross-product runner, then passed with
per-package argv. Command announcements use quoted argv: a multiline argv
must remain data rather than injecting synthetic JSON events into the raw log.
The real argv-injection control failed before that correction. This refines
the approved structured-proof boundary; no persisted state or policy is added.
