---
collaboration-mode: discussion
---

# Suite startup infrastructure evidence implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy).
> Execute with superpowers-executing-plans in the current session; use bounded
> read-only delegation where it saves context. Steps use checkbox syntax.

**Goal:** A recognized infrastructure failure before tests start is could-not-run,
and an explicitly audited correction can remove an already misclassified run's
failure authority without manufacturing a pass.

**Architecture:** Extend the existing strict startup-log classifier and
suiteEnvironmentError path. Add one exact-attempt recovery command, backed by
the existing local evidence store and an immutable original-record audit.
Publish recovered authority atomically in a versioned record; every proof
consumer derives from that authority.

**Tech stack:** Go, Cobra, local JSON evidence, real temporary Git repositories,
the existing fake GitHub client, recorded process-output fixtures.

## Scope and alternatives

This is full-flow gate-evidence work, one atomic close boundary. No implementation
before the hub approves this design and change-code admits the plan. There are no
Mx tags because there is one final review boundary.

1. **Recommended:** reuse strict startup recognition, then explicitly reclassify
   one verified historical attempt with an audit and a v3 identity envelope.
2. Interpret old logs automatically on every read: less command code, but silently
   changes recorded evidence, offers no operator-visible correction and increases
   proof-path IO. Rejected.
3. Delete a failed record or add an ignored v2 flag/sidecar: smaller change, but
   legacy logs can recreate it and old readers keep the v2 failure authoritative.
   Rejected.

The correction remains narrow: ENOSPC/EDQUOT from Go cache, temporary work or build
output, and a recognized missing-toolchain startup diagnostic. No generic retry,
arbitrary error-text waiver, flake registration or cache cleanup is introduced.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| diagnoseSuiteStartup / diagnoseSuiteStartupLines | cmd/sdlc/suiteenvironment.go | new, generalizes existing suiteEnvironmentLines |
| planSuiteReclassification | cmd/sdlc/suitereclassification.go | new |
| decodeSuiteAttempt / encodeReclassifiedAttempt | cmd/sdlc/teststats_store.go; teststats_reclassify.go | modified / new |
| validateReclassificationProof | cmd/sdlc/suitereclassification.go | new |
| flakeAttempt / isFlakeEvidence / flakes.Unexplained | cmd/sdlc/teststats.go; internal/flakes/evidence.go | modified |

Startup diagnosis returns a typed resource/cause/path, rather than only a bool.
It scans the complete bounded log once, accepting only enumerated startup lines
and complete known wrapper grammars. It rejects any test activity, unrelated
compiler error, unknown output, truncation or incomplete traceback. Empty failure
names alone never prove that no test ran. Existing TMPDIR recognition remains a
member of the same classifier. No map of observed booleans becomes authoritative.

The eligibility decision takes an exact attempt, verified record/log identity,
caller identity, diagnosis and requested reason. It returns either an explicit
refusal or the single permitted transition to could-not-run. The result cannot
pass, explain a code failure, confirm a flake, or become a baseline verdict.

One record owns one original-record audit. All proof consumers share the same
projection; guardSuitePass, saved-result reuse, flake registration/reporting,
stats and baseline/train handling must not implement independent overrides.

### Function test contracts

Individual input cases belong in executable tests. These are the required
adversarial strategies and independent mechanical guards for the risky functions:

| Function | Adversarial strategy and mechanical guard |
|---|---|
| diagnoseSuiteStartup / diagnoseSuiteStartupLines | Fuzz arbitrary bytes and wrapper mutations seeded with the recorded outage; appending any independently encoded test-start event must reject classification. Count exact iterator visits at N/2N. |
| planSuiteReclassification | Table/sequence-test the state/action relation below with actor/subject/proof mutations; only an eligible original failure can transition, and no transition returns passing evidence. |
| decodeSuiteAttempt / encodeReclassifiedAttempt | Fuzz schema/version/type/duplicate mutations and round-trip unknown fields; independently compare preserved original bytes and nested unknown values, and run the pinned old-reader control. |
| validateReclassificationProof | Mutate each bound record/audit/log identity independently and truncate at every publication boundary; any mismatch refuses, while a valid proof still cannot satisfy Lands. |
| flakeAttempt / isFlakeEvidence / flakes.Unexplained | Sequence-test real failure, neutral correction and genuine rerun through the shared pure rules; a neutral record never confirms intermittency or explains a different failure. |
| publishSuiteReclassification / readSuiteEvidenceWithOpen | Inject write/rename failures and interleave complete reads at each commit boundary; readers see the old failure, verified correction or scoped refusal, never absence/success. Count bounded record/log reads at N/2N. |
| guardSuitePass / savedAttemptRecord / aggregateCheckAttempts | Drive actual consumers from the same persisted fixtures; compare outcomes with an independent ledger of genuine test runs to reject legacy resurrection and old-pass promotion. |
| diagnoseSuiteFailure / conductor.runSuite / conductor.mainIsRed | Replay the captured startup output through the production runner seam; assert no main-red/attribution transition and preserve an independently introduced genuine code failure. |

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| Live startup handling | cmd/sdlc/localcheck_run.go; localcheck_flakes.go | modified | complete process log and environment retention |
| Exact-attempt recovery | cmd/sdlc/localcheck_cmd.go; localcheck_reclassify.go | new | CLI, workspace identity, repo transaction lock |
| Reclassification store | cmd/sdlc/teststats_reclassify.go | new | original bytes, audit, atomic record publication |
| Evidence snapshot consumers | cmd/sdlc/teststats_evidence.go; localcheck_republish.go | modified | scoped proof reads and immutable log verification |

Reuse localRunRepo, the real private Git checkout and the stateful fake GitHub
client. No live fleet, shared cache mutation, private Couch harness or new external
service is needed. The reclassification command performs local evidence effects
only; it does not overwrite GitHub status or automatically republish an old pass.
The normal exact-head rerun owns any later status publication.

## Execution and classification contract

- Classify only failed, completed suite/affected executions (and the existing
  known-flake rerun path), with a complete readable bounded log. A timeout,
  cancellation, unknown exit or successful command cannot become infrastructure
  evidence merely because its output quotes a diagnostic.
- Recognized startup forms must cover the recorded #508 go-list traceback,
  direct Go cache writes, Go work-directory creation, build-output storage and
  missing Go toolchain. ENOSPC and EDQUOT are distinct recognized causes.
  Each accepted form has a complete positive fixture and adversarial controls.
- A test RUN/PASS/FAIL/SKIP record, JSON test event, package test result, mixed
  ordinary compiler failure, quoted test output, unknown wrapper or truncated
  line disqualifies startup-only recognition. Unknown evidence stays blocking.
- The message says could-not-run, identifies the resource and reported path,
  and says repair the environment then rerun the exact check/head. It does not
  claim knowledge of a mount point the diagnostic does not expose.
- Newly recognized startup failures follow existing retainSuiteEnvironment:
  retain the raw diagnostic, return suiteEnvironmentError, no failed SuiteAttempt,
  no failing local status, no flake classification and no legacy suite.log.
- Direct startup errors do not trigger a diagnostic main run. A baseline startup
  error propagates as could-not-run, without SuiteMainRed, hold, split or blame.
  Genuine candidate failures remain blocking while that baseline is unknown.

## Recovery command and evidence authority

Proposed syntax:

```text
sdlc local-check reclassify --attempt ID --reason TEXT
```

ID is an exact published attempt ID, not a head selector or an abbreviation.
The command accepts only a failed, non-timeout suite/affected attempt whose
immutable raw log hash matches and whose complete output satisfies the same
startup classifier. Refuse pass records, unknown subjects, missing/changed logs,
tests that started, mixed failures, unsupported evidence and arbitrary reason-only
overrides. Existing real failures on the same head are unaffected.

The originating recorded slot may correct its own attempt in this repository;
the repository's verified :0 may correct another slot's attempt. Resolve the
caller through existing workspace identity. Never take actor/slot authority from
a user flag. Missing or ambiguous original ownership requires :0. A diagnostic
baseline is eligible under these same checks, but never becomes landing proof.

Before replacing authority, retain the original JSON bytes and log identity in an
immutable audit, bound to repository, attempt, head, check, old outcome and record
digest. Record actor, UTC time, required reason, recognized cause/path and the
classifier revision. Reuse the queue reconciliation pattern for fsync-before-
publication, digest checking and idempotent interrupted-operation recovery.
Do not move or edit the raw failure log.

Publish a v3 SuiteAttempt identity envelope atomically at the same attempt path,
with a nonpassing could-not-run disposition linked to that audit. Preserve every
unknown field from the original record; retain unknown audit fields on any
recovery rewrite. The original exit, timeout, timestamps and hashes remain facts.
Readers verify the audit and log proof before excluding its failure authority.
Missing, mismatched or incomplete proof remains blocking, never absent.

Why a new version: v2 readers ignore additive fields. A v2 reclassification flag
would let them continue treating the old exit as code failure. Existing #524
readers recognize a future-version identity and refuse only its subject/linked
parent with upgrade guidance; older readers may refuse more broadly. Verify both
behaviors against pinned old code. No old reader may accept it as a pass.

The recovered head stays represented in knownSuiteHeads so legacy per-head logs
cannot resurrect the failure. Exclude its neutral disposition from fail/pass
intermittency and known-flake denominators. Saved-pass reuse must still require
its existing newest/identity/hash rules; reclassification does not promote an
older success. A fresh genuine pass may post after all real failures are explained.

## State design

Actions: inspect the retained log; repair the environment; rerun the exact check;
reclassify one eligible attempt; repeat the same recovery command; upgrade sdlc.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Live startup diagnosis confirmed | could-not-run with resource/path and raw log | repair environment, rerun exact check | execution time; no code evidence exists |
| Tests started or log is mixed/unknown | ordinary blocking failure | inspect retained log | none; failure history retained |
| Eligible old failure, no correction | blocking historical attempt | reclassify exact attempt | none; original bytes preserved |
| Foreign or unprovable owner | refused, names required authority | inspect as owner or repository :0 | none |
| Audit saved, authority not published; interruption | correction incomplete, old failure remains authoritative | repeat same recovery command | no original evidence lost |
| Audit and v3 disposition confirmed | could-not-run, no acceptance evidence | repair environment, rerun exact check | none; old failure remains in audit |
| Repeated completed correction | already reclassified, same audit | rerun exact check | no duplicate audit or status |
| Audit/log missing or changed | unverified evidence, blocking | inspect retained log | unavailable evidence cannot be inferred |
| Old binary sees v3 | needs newer sdlc | upgrade sdlc | no permissive reinterpretation |
| Baseline cannot run | baseline unknown, no car blame | repair environment, rerun exact check | no code attribution is available |

The repository transaction lock owns mutation admission; an observation-only
no-other-writer check is insufficient. Readers consume atomically published
records. Any interrupted file effect is reconciled by exact digest/identity;
unknown completion never deletes the old attempt or clears failure authority.
No background work outlives command return.

Covered: live execution, candidate/main/train/rerun consumers, ownership, schema
compatibility, interruption and duplicate recovery. Not covered: generic OOM,
network/toolchain download outages, arbitrary shell failures, cache management,
automatic bulk reclassification, old legacy logs without a recorded attempt,
or changing flake/landing admission policy. These lack the requested narrow proof.

## Architecture and operating envelope

- **ARCH-DRY/PURE:** one diagnosis and one recovery transition, existing execution
  error path and evidence snapshot projections. No second cache or shadow flag.
- **ARCH-PURPOSE:** enumerate direct suite, affected, flake-rerun, baseline, train,
  same-head guard, saved result, legacy and stats readers; test each authority.
- **ARCH-MOCK:** private real Git/process fixtures plus existing fake GitHub state;
  recorded outage output replays in the normal suite. No machine-wide ENOSPC.
- **ARCH-CONSTRAINTS:** command/batch path, not interactive UI. Keep existing
  complete-log/line/record/store byte bounds. One O(L) diagnosis per log and
  O(R + total verified recovery bytes) per evidence read; fixed work per byte.
  Count visits/read calls at N=8 and 16. Exceeding bounds refuses classification.
- **ARCH-SECURE:** stored JSON, logs and CLI input are untrusted. Exact canonical
  subject, hash and ownership checks precede authority mutation. No credentials.
- **ARCH-ORDER:** explicit transitions above; durable audit before atomic authority
  replacement, exact retry and refusal for unknown completion.
- **ARCH-FUNERAL:** audit lifetime follows its attempt's existing 30-day retention;
  prune both together. Incomplete artifacts use the existing 24-hour stale-owner
  model and cannot be collected while a live recovery owns them. No forever
  side ledger. Raw log retention remains the store's existing lifecycle.
- **ARCH-EVOLVE:** v1/v2 remain readable; future semantics refuse on old readers.
  Preserve original unknown JSON and audit bytes. Test actual old readers and
  unrelated-subject behavior, rather than only a new decoder simulation.

## Implementation tasks (after approval and change-code)

### Task 1: Recorded startup diagnostics and live execution

Files: suiteenvironment.go, suiteenvironment_test.go, localcheck_run.go,
localcheck_flakes.go; recorded fixtures under cmd/sdlc/testdata/.

- [ ] Capture today's retained #508 candidate and baseline logs into replay
  fixtures, preserving diagnostic structure and documenting any path sanitation.
- [ ] Implement the red diagnosis contract above in colocated tests before code.
- [ ] Run focused tests and retain red evidence. Extend the typed classifier and
  existing error/retention path only after the regression fails.
- [ ] Exercise runLocalCheck on a private repo: no failure status/attempt/flake
  evidence, raw log retained, same-head genuine repair passes. Keep a prior real
  failure blocking. Cover affected and a known-flake rerun's startup failure.
- [ ] Verify the baseline/train function contract through the same replay seam.
- [ ] Count scanner work at 8/16 units. Commit the coherent verified change.

Target: go test ./cmd/sdlc -run 'TestSuite(Environment|Infrastructure)' -count=1

### Task 2: Audited exact-attempt reclassification

Files: suitereclassification.go and tests, teststats_reclassify.go and tests,
teststats_store.go, teststats_evidence.go, teststats.go, localcheck_cmd.go,
localcheck_reclassify.go and tests, localcheck_republish.go.

- [ ] Write the red transition, proof, encoding and consumer contracts above
  using storedAttemptFixture; add actual CLI tests through localRunRepo.
- [ ] Implement the exact command, audit-first publication and explicit v3
  validation/projection. Reuse existing lock/recovery patterns and scope reads.
- [ ] Exercise the publication/read ordering contract with interruption injection,
  then verify the shared consumer contracts against the independent run ledger.
- [ ] Test v1/v2 reads, unknown-field preservation and real pinned old-reader
  behavior using the #524 compatibility fixture approach. In particular, the
  #524-era reader must scope a v3 disposition to its subject and allow unrelated
  proof. The TL confirmed #524 is deployed fleet-wide. Test joint pruning.
- [ ] Run focused controls, race-check the bounded mutation/read interleavings,
  and commit only with passing evidence.

Target: go test ./cmd/sdlc -run 'Test(SuiteReclass|LocalCheckReclass|ScopedSavedProof|TestRunGuard)' -count=1

### Task 3: Discovery, complete verification and landing

Files: cmd/sdlc/helptext/local-check.md, README.md,
atlas/workflow/ci-merge-check.md, workshop issue/lessons.

- [ ] Document recognition limits, exact command/owner/:0 rules, audit location,
  partial-operation recovery, no GitHub status rewrite and required real rerun.
- [ ] Evaluate the Go lint rule obligation: this is semantic evidence grammar
  and authority, defended by replay/transition/consumer tests; add a lint only
  if investigation reveals an enumerable mechanical misuse class.
- [ ] Ask the hub for the live-smoke decision with the concrete reviewed result;
  never fill a volume or modify the live fleet. Prefer a bounded real CLI replay
  against a private local evidence store.
- [ ] Push, run sdlc local-check run --name suite, wait for final pass, then close
  with behavior evidence. Fix all blocking review findings and record lessons.
- [ ] Enter sdlc merge --yes promptly; retain the waiter. Report the verified
  merged PR, tracker done/archive and clean resting branch to the hub.

No estimate until the plan-quality gate passes. No gate bypass or flake label is
part of this design. Any third design delta on the mechanism triggers re-planning.

## Revisions

- 2026-10-21 — TL approved this design and explicitly required the #524-era
  reader compatibility control; added it to Task 2. No mechanism change.
- 2026-10-21 — PQ-1: named the risky functions and their adversarial strategies
  and mechanical guards; compressed procedural test-case inventories into
  executable test contracts. Scope, evidence authority and recovery stay intact.
- 2026-10-21 — Runtime discovery, awaiting TL approval: a direct missing Go
  executable produces Exit=-1 and an empty log, which the current runner also
  records as unattributable. Proposed narrow positive-proof exception to the
  unknown-exit rule: the first command is Go, Cmd.Process is nil, and its typed
  start error proves the executable/interpreter is missing. Carry a typed startup
  error through the existing cleanup/queue lifecycle into environment retention;
  do not synthesize process-output bytes or add passing authority. Reject the
  exception after any earlier batch command or for an unrelated missing tool.
  Historical empty logs still refuse correction without positive proof. Native
  first-command/later-command controls will pin the distinction. This changes
  runtime failure provenance only, not the v3/audit mechanism.
