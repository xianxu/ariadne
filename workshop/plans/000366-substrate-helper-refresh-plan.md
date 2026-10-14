# Substrate helper compatibility and issue-branch refresh implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy). Delegate bounded implementation only after approval; keep integration in the main session. Steps use checkboxes.

**Goal:** Refuse incompatible substrate helpers with a usable repair command, and let an issue branch refresh its substrate without moving its own Git branch.

**Architecture:** Probe the capabilities needed by the existing suite-lock caller before acquisition. Extend the existing refresh transaction to distinguish the observed graph from repositories allowed to advance; retain the consumer as a pinned, read-only Git subject. No automatic refresh, new installation policy, or separate refresh implementation.

**Tech stack:** Go/Cobra, existing Git adapters and stateful Git fixtures, Python argparse help, existing setup lease and layer graph.

## Approved problem and proposed contract

- `sdlc local-check` resolves `scripts/test-lock.py` beside the real `Makefile.workflow`. Probe `python3 <helper> acquire --help` for the actual option declarations `--watch` and `--slots`, then acquire exactly as today. No commit-age comparison: an older compatible helper remains usable.
- For a dependency-supplied helper, an unsupported helper refuses before queueing, running the check or posting any GitHub status. The single-line diagnostic identifies the resolved helper, its supplying checkout and exact HEAD, the missing capabilities, and the shell-quoted command `cd <consumer> && weave refresh --substrate`, followed by retrying the original command. No argparse usage dump and no automatic refresh.
- Failed, malformed, cancelled, oversized or timed-out probes fail closed. Probe execution uses the existing process-group cancellation pattern, a five-second bound and 64 KiB combined-output cap. These are read-only help budgets; tests exercise the bounds without waiting five seconds. Errors distinguish unverifiable capability from proven missing options.
- `weave refresh --substrate` observes the complete current graph but never fetches or advances the consuming root, including its origin/main ref. It refreshes only the existing declared substrate subjects ordinary refresh would select. Numbered environments continue to enforce independent private checkouts through the existing acquisition policy; no other slot or data mount becomes a subject.
- The consumer can be on a clean issue branch ahead of or divergent from main. Preserve its branch, HEAD, tracked source and index; root identity, declarations and cleanliness remain revalidated before dependency effects and compilation. Generated composition may change as an explicit consequence of compile and is reported normally.
- Retain all existing dependency safeguards: clean checkouts, no active Git operation, source identity, unchanged target declarations, ignored-file collisions, captured targets, optimistic revalidation and partial-progress reporting. Never stash or reset.
- Reject `--substrate --rebase` before setup effects. Substrate subjects fast-forward only. Ordinary refresh and its explicit rebase mode retain their existing behavior.
- Proposed source-owned exception for hub review: when the stale helper is tracked in the consuming repository itself, substrate-only refresh cannot update that source without violating branch preservation. Name integration of current main into the issue branch as the required source repair; do not claim substrate-only refresh repairs it.
- With no substrate dependencies (this Ariadne source checkout), compile the current branch once and report that no substrate Git checkout advanced. Updating source-owned helper code still requires integrating main separately; this mode never rewrites the source branch to manufacture an update.

## Core concepts

| PURE entity | Lives in | Status |
|---|---|---|
| Required helper option parsing (`missingHelperOptions`) | `cmd/sdlc/helpercapability.go` | new |
| Refresh mode (`Options`) | `cmd/weave/internal/refresh/model.go` | new |

`missingHelperOptions` accepts help bytes and required option names, returning the missing names. Match option declarations, not an incidental mention in prose or a substring such as `--slots-extra`. `Options` is a value selecting the allowed effect set; no new independent state machine is introduced.

| INTEGRATION | Lives in | Status | Wraps |
|---|---|---|---|
| Helper capability probe and refusal renderer | `cmd/sdlc/helpercapability.go` | new | bounded Python process and supplying-checkout Git evidence |
| Suite-lock acquisition (`acquireSuiteSlot`) | `cmd/sdlc/localcheck_run.go` | modified | resolved helper capability and existing machine lock |
| Refresh transaction (`RunWithOptions`, existing `Run` wrapper) | `cmd/weave/internal/refresh/refresh.go` | modified | acquisition graph, Git observations/effects, compile callback |
| Refresh CLI (`buildRefresh`) | `cmd/weave/refresh.go` | modified | mutually exclusive mode flags and existing setup lease |

The helper and Git probes use production process seams with real portable temporary repositories. Existing Python lock doubles gain explicit side-effect-free help; acquisition markers distinguish a probe from joining the queue. Refresh tests reuse the stateful real-Git model and numbered-environment fixture rather than inventing a second graph or Git implementation.

## Consumer inventory

- `acquireSuiteSlot` is the only direct SDLC substrate-script consumer found. Both explicit local-check and merge's automatic local-check share it.
- Makefile.workflow also invokes test-lock.py, but does not use the installed SDLC binary. This issue changes no protected workflow/lock configuration.
- `probeMakeTestFactsContext` runs Make under a non-test goal with `WF_TEST_LOCK=off`; it does not acquire the lock.
- Installed weave/vocabulary/agent commands are not substrate-relative scripts with a newly added helper argument. Do not impose a speculative repository-wide minimum version.

## Ordering, trust, bounds and lifetime

- ARCH-DRY / ARCH-PURE: capability parsing is byte-in/decision-out; acquisition retains one caller and refresh retains one transaction. Existing `Run` delegates with default options so current callers/tests stay compatible.
- ARCH-PURPOSE: repair includes the issue-branch substrate refresh, not just a better error naming the previously unusable command.
- ARCH-MOCK: portable Git repositories own refs, files and operation markers. Existing stateful adapters inject fetch-time mutations and compile failure. Run the real current Python helper against an isolated lock path to check the capability advertisement and weighted admission seam.
- ARCH-CONSTRAINTS: helper help is bounded by time and output; graph retains the existing 128-layer and 1 MiB declaration limits. Full suites use the weighted machine gate.
- ARCH-SECURE: preserve the actual supplying checkout identity through symlinks; use quoted paths, omit credential-bearing Git stderr, and refresh only policy-admitted subjects.
- ARCH-ORDER: unsupported helper is a no-effect refusal. Refresh observes/preflights all subjects before any branch update, revalidates the unchanged root around IO, and confirms each dependency update before compile. Cancellation never becomes success; compilation failure retains confirmed dependency updates and retry recompiles.
- ARCH-FUNERAL: no capability cache or durable version registry. Probe output dies with the call; process descendants are reaped. Existing setup leases, detached check worktrees and local logs retain their existing cleanup policies.

## Task 1 — capability refusal at the real suite-lock caller

**Files:** `cmd/sdlc/helpercapability.go`, colocated tests, `localcheck_run.go`, `localcheck_run_test.go`, `localcheck_cmd_test.go`, and existing reentry fixtures only where their fake help must be updated.

- [x] Add pure parser cases for declared required options, missing options, misleading prose/substrings and malformed output.
- [x] Add a stale helper fixture with real supplying Git HEAD and workflow symlink. Through the local-check command/run path assert the helper acquisition marker, check marker and GH status/comment collections remain untouched; the diagnostic names the supplier path/HEAD and runnable repair. Verify RED against current code.
- [x] Add bounded probe cancellation/hang/output/error cases; current helper and weight-two fixtures; paths containing spaces; retained reentry and diagnostic baseline behavior.
- [x] Implement the bounded probe and caller wiring. Use `gitx.RunGitInContext` for supplier top-level/HEAD, never the consumer-bound `d.Git` for that evidence. Require normal helper file resolution; unreadable evidence refuses with a concrete inspection action rather than inventing a hash.
- [x] Update existing acquisition doubles to handle `acquire --help` without acquisition effects. Run targeted local-check/helper/reentry tests and verify GREEN, including the real helper fixture.

Command: `go test ./cmd/sdlc -run '^(TestHelperCapability|TestRunLocalCheck|TestLocalCheck|TestAcquireSuiteSlot)' -count=1 -timeout 15m` (adjust the explicit test-name list to the existing named reentry tests discovered during implementation).

## Task 2 — substrate-only mode within the existing refresh transaction

**Files:** `cmd/weave/internal/refresh/model.go`, `refresh.go`, existing refresh integration tests, `cmd/weave/refresh.go`, `cmd/weave/refresh_test.go`.

- [x] Write a real numbered-slot regression: consumer issue branch ahead/divergent, its remote main independently advanced, private dependency behind remote. Assert dependency reaches captured main; consumer HEAD/branch/index and origin/main remain unchanged; compile sees updated substrate and runs once; primary and another slot remain untouched. Verify RED.
- [x] Add no-dependency source-branch and `--substrate --rebase` cases, with no Git updates and explicit compile/no-effect expectations respectively.
- [x] Add stale root evidence injected during dependency fetch/pre-apply/pre-compile, dirty or divergent dependency, target declaration change, missing checkout, private-ownership escape and compile-failure/retry cases. Reuse existing cases where they already prove the shared path; add mode-specific controls for the excluded root and selected subjects.
- [x] Implement `RunWithOptions` and have `Run` delegate with ordinary mode. Keep root in the observation and final-validation set, set its preserved target to its original HEAD, skip root fetch/target checks/update, and use the original phase transitions for all effects. Add CLI mode and mutual-exclusion validation before setup.
- [x] Run all refresh package and CLI regressions. Verify setup lease is held through compile and current ordinary refresh remains unchanged.

Commands: `go test ./cmd/weave/internal/refresh -count=1`; `go test ./cmd/weave -run '^TestRefresh' -count=1 -timeout 10m`.

## Task 3 — executable recovery documentation and acceptance

**Files:** `atlas/workflow/weave.md`, `atlas/workflow/workspace-branching.md`, `atlas/workflow/ci-merge-check.md`, issue details. Reuse the existing atlas pages; their index entries already exist.

- [x] Document the capability refusal, supplying HEAD, `weave refresh --substrate`, clean issue-branch requirement, source-only edge and retained-update retry semantics. Update CLI help consistently.
- [x] Check that removing capability wiring makes the production regression fail and removing root exclusion makes the issue-branch refresh regression fail. Record actual RED/GREEN evidence, not wording tests.
- [ ] Run all weave packages and appropriate local-check targets; commit code/docs, merge current main, then run full `make -k test` through the machine gate with the checkout frozen.
- [ ] Record local acceptance and enter binary-owned close review. PR CI and landing are subsequent gates; land autonomously after SHIP with current main integrated and `sdlc merge --yes`, never validation/CI bypasses.

## Revisions

- 2026-10-14 — Planning incorporates the issue's 08:20 revision. The diagnostic repair is substrate-only because ordinary refresh attempts to advance an issue branch. The complete observed graph remains an invariant even when the consumer is excluded from the effect set. No implementation has started; hub plan approval is pending.
