---
collaboration-mode: discussion
---

# Managed bootstrap suite implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 for execution strategy. Execute this atomic policy change in the current session using superpowers-executing-plans. SDLC owns the fresh-context reviews; there is one close boundary, no Mx subdivision.

**Goal:** A pushed first-suite PR can run and land with managed default `make test` evidence when main has neither a declaration nor a test target.

**Architecture:** Share one subject-aware extension of main-authoritative suite derivation. Only proven declaration absence and a successful negative main probe permit a default derived from the exact subject. Main's CI requirements and existing declared/default suites remain authoritative.

**Tech Stack:** Go, existing Git runner, detached-worktree/make probes, managed queue/lease/record machinery, stateful fake GitHub and private native Git repositories.

## Core concepts

No new pure business entity is needed. Existing constructors remain the required/default policy source (ARCH-DRY, ARCH-PURE); their status below is reused unchanged rather than a fictitious modification.

| Pure entity | Lives in | Status |
| --- | --- | --- |
| `defaultSuiteCheck`, `effectiveLocalChecks` | `cmd/sdlc/localchecks.go` | Reused unchanged |
| `decideCIGate` | `cmd/sdlc/cigate.go` | Reused unchanged; no new CI stand-ins |

| Integration point | Lives in | Status | Wraps |
| --- | --- | --- | --- |
| `withDefaultSuiteContext`, `makeTestFactsContext` | `cmd/sdlc/localchecks_make.go` | Reused unchanged | Main's suite and context-bound native make probe |
| `withSubjectDefaultSuiteContext` | `cmd/sdlc/localchecks_bootstrap.go` | New | Main-first derivation, exact-head fallback and per-call bootstrap diagnostic flag |
| `runLocalCheckCmd` | `cmd/sdlc/localcheck_cmd.go` | Modified | Pin/fetch main, verify pushed subject before fallback, select and execute managed suite |
| `prepareCIGateFactsContext`, `ciGateFacts` | `cmd/sdlc/cigate_wait.go` | Modified | Same subject policy inside the existing CI wait/read budget |
| `runPreparedCIGate` | `cmd/sdlc/cigate.go` | Modified | Visible bootstrap diagnostic and ordinary exact-head evidence acceptance |
| `declaredCheck`, resume policy readers | `cmd/sdlc/localcheck_cmd.go`, `localcheck_resume*.go` | Modified | Current pinned main plus saved/original subject, shared derivation |
| `diagnoseSuiteFailure` | `cmd/sdlc/suitebaseline.go` | Modified | Ephemeral bootstrap failure provenance, no nonexistent main baseline |
| Conductor `build`, `deriveSuite`, `runSuite`, `reintegrate` | `cmd/sdlc/landingtrain.go` | Modified | Same policy for each tested prefix, including split and changed base |

`withSubjectDefaultSuiteContext` accepts parsed main declaration and pinned base/head. First call existing main derivation. If the original declaration is present (including valid empty), or main provides a suite, return that result and bootstrap=false. Only original Version=0 plus successful no-suite derivation allows the existing head make probe/default injection; bootstrap=true only if the head has a test target. Propagate failed/unknown/interrupted probes rather than treating them as absence. Never read the head declaration.

The default is exactly `make test`, required, with existing default timeout and the selected Makefile's weight. Main's existing suite/weight cannot be lowered. Head-only custom names, substitutions and declarations do not activate until on main. A subject without a test target receives a no-suite refusal when a run is requested. No policy is published as a side effect.

### Risky functions and independent guards

| Boundary | Adversarial strategy | Mechanical guard |
| --- | --- | --- |
| `withSubjectDefaultSuiteContext` | Adversarial absence/unknown/authority combinations | Native Git/make fixtures assert selected command, weight and required policy |
| `runLocalCheckCmd` | Subject identity and undeployed head policy attempts | Private bare origin, native command markers and exact-head GH status assertions |
| `prepareCIGateFactsContext`, `runPreparedCIGate` | Evidence from another subject or code state, red server checks | Existing stateful `ciFakeGH` and real ordinary/immutable landing fixtures |
| Conductor `build`, `deriveSuite`, `runSuite`, `reintegrate` | Prefix and base transitions that invalidate suite eligibility | Native execution records exact head/command/weight; absent-target prefixes cannot virtual-pass |
| `declaredCheck`, `resumeInputs`, `resolvedResumeRecords` | Saved proof recovery from an unrelated checkout with no target | Native known-flake republish/resume and current-main policy refusal |
| `diagnoseSuiteFailure` | Bootstrap failure without any runnable main baseline | Real CI gate failure retains original error and unverified diagnosis |
| `makeTestFactsContext` via subject fallback | Cancellation and failed probe must not establish absence | Existing process/budget controls plus fallback cancellation regression; no later launch |

These are semantic policy/provenance invariants, defended by executable tests rather than a helper-name lint. Existing parsers and protected-path guards keep schema/config ownership.

## Visible states and recovery

Actions: run suite, push subject, fix target/declaration, repair probe environment, cancel exact run ID, retry exact head.

| State / failure mode | Visible state | Recovery | Possible loss |
| --- | --- | --- | --- |
| Main has declaration or default suite | Ordinary main-authoritative managed run | Run suite | None |
| Main lacks both; pushed subject has test | Bootstrap diagnostic and ordinary required `make test` record on that head | Run suite | None; future config not activated |
| Neither has target, or requested custom name is head-only | Actionable no-suite/undeclared refusal | Fix target or select suite, push, retry | None |
| Main declaration/probe unreadable, invalid, interrupted or unknown | Preparation/infrastructure refusal, never fabricated absence | Repair named input and retry | None |
| Run fails/cancels or status post uncertain | Existing failure/cancel/republish state and exact ID | Existing retry/cancel/republish command | Existing retained-log rules |

Republish/resume current-policy readers pass the saved/original SHA to the same helper. Immutable suite failure retains ephemeral bootstrap provenance so its absent main baseline is reported unverified rather than executed.

No new persistent state, formats, resources or recovery verbs (ARCH-EVOLVE, ARCH-FUNERAL). Queue, lease, process group and detached checkout retain ownership of execution. Selection carries no state between external events: recompute for each train prefix/changed base; CI wait caches immutable facts only for its invocation (ARCH-ORDER).

Operator-started batch: at most one base and one fallback head make probe per prepared subject, under the existing 2-minute/context/wait bounds. No polling probes, extra concurrency or automatic retry (ARCH-CONSTRAINTS). Repository code remains at the existing trust boundary; no credentials or shell interpolation added (ARCH-SECURE). Reuse native private Git repositories, existing queue seams and stateful GH fake; no live fleet mutation (ARCH-MOCK).

## Incident fixture provenance

echo#1's archived Log records the 2026-10-21 refusal and steward-authorized direct suite. Read-only Git inspection verified:

- Planning-era main `aeee435d3953c60cdd1ab4350e206d0e162646f7` and final pre-merge main `77a0fd6320ccc5ff603c9cfefdb8bb4789486556` have identical Makefile blob `0e078508dd6066bb3ac059e6a5b649e26a568694`.
- Neither tracks `Makefile.local`, `Makefile.workflow` or `.sdlc/local-checks.json`. Root Makefile only optionally includes the absent files and defines help; a detached probe has no test target.
- Tested head `e3965c2c6d79ece162e704d07335b5b883fe9668` adds `.PHONY: test`, `test:` and `bash tests/ci-workflow.test.sh` in Makefile.local; its first declaration requires `make test`.

Commit minimal byte-preserving Makefile/declaration inputs and a provenance README. Supply a private marker implementation of the referenced harness: this proves bootstrap dispatch/management, not echo's separately owned assertions. Tests never read a live sibling.

## Task 1: Reproduce and share policy

**Files:** Create `cmd/sdlc/localchecks_bootstrap.go`, `cmd/sdlc/localchecks_bootstrap_test.go`, `cmd/sdlc/testdata/bootstrap-suite/echo-1/`; modify CLI and CI integration points above.

- [x] Capture verified incident inputs/provenance. Write a native CLI regression and confirm old code refuses before the fixture harness runs.
- [x] Implement the executable adversarial strategies for the shared derivation and CLI rows above; retain their meaningful red results.
- [x] Implement shared main-first derivation and pushed-subject validation before fallback. Label bootstrap explicitly; route CI facts through the same function.
- [x] Run focused native tests plus existing default/declared-suite, CI wait/probe and ancestor-record controls. Assert actual command, weight, required and recorded SHA, not only text. Commit code and Log evidence.

## Task 2: Cover every landing subject

**Files:** `cmd/sdlc/landingtrain.go`; extend existing native train and single/immutable landing tests.

- [x] Implement the executable CI/admission and conductor strategies above; retain meaningful red results before wiring those paths.
- [x] Derive the suite per actual prefix through the same policy before execution; retain normal declared-policy paths and failure attribution. An unavailable main baseline stays unverified, never invented main-health proof. Do not let an empty/virtual runner satisfy a bootstrap suite.
- [x] Re-run native train, suite enforcement, main advance and CI budget/cancellation controls through existing resource seams. Commit code and Log evidence.

## Task 3: Document and verify

**Files:** `cmd/sdlc/helptext/local-check.md`, `README.md`, existing CI/sdlc atlas pages, issue Log.

- [x] Document exact eligibility, managed command, head-only evidence, delayed custom-policy activation, protected-label ownership and refusal recovery.
- [x] Run focused packages and compiled help; verify `git diff --check`. Ask hub whether isolated production CLI/landing proof suffices for the live gate; no :0 staging without approval.
- [x] Commit/push final implementation, retain `sdlc local-check run --name suite` on its exact head through completion, and record the full result before close. No manual main chase.

Then `sdlc close` owns the one boundary review. On SHIP enter `sdlc merge --yes`; verify merged PR, archive and clean rest, and report to hub. No new estimate until change-code's plan-quality gate accepts this full-flow plan.

## Revisions

- 2026-10-22 01:35 PDT — Address plan-quality advisory PQ-1 without changing behavior or scope: name actual risky functions in the strategy table and remove repeated task-level case inventories. Executable regressions own the individual cases; task checkboxes reference the strategies once.

- 2026-10-22 01:49 PDT — Hub confirmed that existing known-flake republish/resume consumers must use the same main-first helper with the saved/original SHA. Add native recovery regressions from an unrelated HEAD; no head declaration becomes policy. Carry ephemeral bootstrap provenance through immutable failure diagnosis to avoid invented main health. Same mechanism and approved contract, no new persistent state or commands.
