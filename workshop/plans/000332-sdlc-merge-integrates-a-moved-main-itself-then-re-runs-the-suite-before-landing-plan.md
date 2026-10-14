# sdlc merge: Integrate a Moved Main, Enforce the Suite at Landing — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** One `sdlc merge --yes` lands a reviewed branch even when main moved. It merges main into the branch when that is clean, never lands a head without a passing suite record, refuses with the paths on conflicts, labels suite failures as introduced or pre-existing, and re-integrates (bounded) when main moves again.

**Architecture:** The suite becomes a REQUIRED recorded local check (`sdlc-local/suite`, #345's engine). It is declared in `.sdlc/local-checks.json` or, by default, `make test` with the Makefile's own weight. The CI gate demands it for the head it merges, and accepts a record on an ancestor whose difference from the head is docs-only, because close's own commits are docs. `runDurableMerge` gains an integration loop with a pure decision function: merge main into the issue branch, pre-check the publish gate's replay rule (#304) on the merge commit, push with a lease, let the suite record be produced, wait for CI (bounded), re-check main, then land or loop (at most 3). A failing suite triggers one baseline run on main's sha to classify failures.

**Tech Stack:** Go (`cmd/sdlc`), git (merge, merge-tree via existing helpers), `gh` (statuses via the #345 seam), `make`, `scripts/test-lock.py` (#319/#341/#346).

**Approved:** hub (ops:0), 2026-10-13. The answers:
- (a) the suite is the declaration, defaulting to `make test` with the Makefile's weight;
- (b) every landing needs `sdlc-local/suite` on its head, and `sdlc local-check run --name suite` is THE close-suite command, so merge reruns only when the head changed;
- (c) the legacy path refuses when main moved;
- (d) a bounded 20-minute CI wait.
- The baseline runs on failure only.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `localCheck.Weight` | `cmd/sdlc/localchecks.go` | modified |
| `defaultSuiteCheck` / `effectiveLocalChecks` | `cmd/sdlc/localchecks.go` | new |
| `ciInputs.RecordedAncestors` (docs-only-distance rule) | `cmd/sdlc/cigate.go` | modified |
| `failingTests(log) []string` | `cmd/sdlc/suitefailures.go` | new |
| `classifyFailures(merged, baseline) (introduced, preexisting []string)` | `cmd/sdlc/suitefailures.go` | new |
| `integrationObs` / `integrationStep` / `nextIntegrationStep` | `cmd/sdlc/landingintegrate.go` | new |
| `integrationTrailer` (`Integrated-Main: <sha>`) | `cmd/sdlc/landingintegrate.go` | new |

- **localCheck.Weight**: an optional `weight` in the declaration (1..`WF_TEST_SUITES`, default 1), passed to `test-lock.py acquire --slots`. This fixes the found bug: `acquireSuiteSlot` took weight 1, and since it exports `WF_TEST_LOCK_HELD_BY`, `make test`'s own weighted acquire went reentrant, so ariadne's sharded suite held one slot instead of two.
- **defaultSuiteCheck(weight int)** returns `{name: "suite", command: ["make","test"], required: true, timeout_minutes: 60, weight}`. **effectiveLocalChecks(decl, hasMakeTest, weight)** is the declaration, plus the default suite when the declaration has no `suite` and the repo has a `test` target. A declared `suite` always wins (the operator's override). Pure, so the shell supplies `hasMakeTest` and `weight`.
- **RecordedAncestors**: the CI gate's required-check rule becomes "a `pass` status for `sdlc-local/<name>` on the head, or on an ancestor A where every path in `diff(A, head)` is docs-only". It reuses `publishGateHasCodeSurface` (ARCH-DRY: the same docs/code line as the publish gate). This is why the suite an agent records before close stays valid through close's own commits (Log line, card mirror, archive). The pure input is `map[name]bool` of names satisfied by a qualifying ancestor; the shell computes it.
- **failingTests(log)**: one parser for go-test failure lines, `--- FAIL: TestX` (go test) and `--- FAIL <pkg> <TestX>` (test-shard's summary), normalized to `pkg.TestX` when a package is known, else `TestX`, de-duplicated and sorted.
- **classifyFailures**: `introduced = merged − baseline`, `preexisting = merged ∩ baseline`. A baseline that failed with no parseable tests is "main is red (unparsed)".
- **nextIntegrationStep(obs)**: the integration state machine (ARCH-ORDER). It takes observations and returns the next step. See ARCH-ORDER below.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `makeTestFacts(root) (has bool, weight int)` | `cmd/sdlc/localchecks_make.go` | new | `make -s -p -f Makefile -f <probe.mk> __sdlc_weight` |
| `acquireSuiteSlot` (weight) | `cmd/sdlc/localcheck_run.go` | modified | `test-lock.py acquire --slots N` |
| `runLocalCheck` (`record` option) | `cmd/sdlc/localcheck_run.go` | modified | — (baseline runs are not recorded) |
| `suiteAncestorRecords` | `cmd/sdlc/cigate.go` | new | `git rev-list --first-parent`, `git diff --name-only`, `ghCaller.CommitStatuses` |
| `ghCaller.CommitStatuses(ctx, repo, sha)` | `ghclient.go`, `cigate.go` (realGH), `cigate_fake_test.go` | new | `gh api repos/R/commits/SHA/statuses` |
| `integrateMain` (shell for the state machine) | `cmd/sdlc/landingintegrate.go` | new | `git merge`, `git merge --abort`, `git push --force-with-lease`, `git reset --keep` |
| `waitCIGate` | `cmd/sdlc/cigate.go` | new | `runCIGate` polled (30s, 20m cap) |
| legacy refusal | `cmd/sdlc/merge.go` | modified | `git merge-base --is-ancestor` |

- **makeTestFacts**: the Makefile stays the single source of its weight (ARCH-DRY). sdlc runs make with a probe makefile (`__sdlc_weight: ; @echo $(or $(WF_TEST_WEIGHT),1)`), using `-p` to print the rule database. `test` is never a goal there, so Makefile.workflow's parse-time lock is NOT taken. `has` is true iff the database lists a `test:` rule. Seam: `makeTestFactsFn`.
- **suiteAncestorRecords**: walks the head's first-parent ancestors back to the merge base with main, at most 30. For each one it checks the docs-only distance first (a git diff, cheap), and only then asks GitHub for that commit's statuses. It stops at the first qualifying success. Bounded: ≤30 diffs and ≤30 status calls, usually 1–3.
- **integrateMain**: runs in `t.Root` on the issue branch. `start-plan` from rest is already refused at `landing.go:236`, and the tree is clean via `landingCheckoutReady`. The merge commit message is `Merge <main> into <branch> (sdlc merge, #332)` plus the trailer `Integrated-Main: <main sha>`.
- **waitCIGate**: polls only pending and not-registered verdicts. Red, unknown and missing still refuse immediately. It prints `waiting for CI on PR #N (Xm of 20m): <pending names>` each poll.

### Operating envelope (ARCH-CONSTRAINTS)

| Constraint | Budget / behavior | Basis |
|---|---|---|
| Workload | Batch, operator/agent initiated; one landing at a time per slot (the repo lock) | existing |
| Suite runtime | ariadne's sharded suite is about 6–13 min, plus queueing behind the machine lock. The check timeout defaults to 60 min for the default suite, and a timeout is a recorded failure. | measured today |
| Machine load | The suite holds the repo's declared weight (ariadne 2 = alone). The baseline runs only after a failure, under the same lock, sequentially. | #346, hub (b) |
| CI wait | Polls every 30s, capped at 20 min, then refuses naming the pending checks | hub (d) |
| Integration retries | At most 3 per run, then refuse | issue Spec |
| GitHub calls | Ancestor statuses: ≤30 per landing (usually 1–3); CI polls ≤40 | design |

### ARCH-SECURE

- **Inputs:**
  - The declaration is read from main (#345), unchanged.
  - Makefile facts come from the repo's own Makefile (trusted repo code, same as `make test` itself).
  - Statuses come from GitHub. Anyone with the repo token can post `sdlc-local/suite` (#345's accepted trust model: local runs on the operator's machine are trusted; CI guards configuration).
- **Ancestor rule:** the docs-only test uses git on our own objects, so a doc-only change cannot hide code, because `publishGateHasCodeSurface` treats embedded docs (`cmd/**/*.md`) as code.
- **Merge commit:** created locally from fetched refs, then pushed with `--force-with-lease=<branch>:<old head>`, so a concurrent push is never overwritten.

### ARCH-ORDER

The integration loop is the component with external events. `nextIntegrationStep(obs integrationObs) integrationStep`:

| Observation | Step |
|---|---|
| main ⊆ head, suite recorded (head or docs-only ancestor), CI green | `land` |
| main ⊆ head, suite not recorded | `runSuite` (CI gate runs the required check inline) |
| main ⊆ head, CI pending | `waitCI` |
| main ⊄ head, attempts < 3 | `merge` |
| main ⊄ head, attempts = 3 | `refuse(moved-too-often)` |
| merge result: conflict | `abort` → `refuse(conflict, paths)` (branch and remote unchanged) |
| merge result: clean, publish replay collides on code | `reset` → `refuse(fixed-then-ship)` (nothing pushed) |
| merge result: clean, replay ok | `push` |
| push rejected by lease | `refuse(concurrent push)` (the branch moved under us; nothing overwritten) |
| local = remote + one commit carrying `Integrated-Main:` whose first parent is the remote head | `push` (resume an interrupted integration) |
| suite failed | `baseline` → `refuse(introduced/preexisting)` |

Events the caller cannot block, and their handling:
- **Process death after the merge commit, before the push:** the next run hits the resume row. The marker is the trailer, not a timestamp.
- **Death mid-suite:** no status was posted (statuses are written after the run), so the next run runs it again.
- **Death after the push, before landing:** the next run sees main ⊆ head and the recorded suite (if it finished), and lands.
- **A second actor pushing the branch:** the lease refuses.
- **main moving during the suite:** the post-suite re-fetch loops back to `merge`.
- **CI's later checks arriving out of order:** the CI gate pins the head before and after (#309 BR-1).

Most likely to be mishandled: main moving after the suite passes, but before `gh pr merge`. Hence the re-fetch is the last step before `LandingMerge`, and the merge stays pinned with `--match-head-commit`.

**Extent:** everything runs in the foreground of `sdlc merge`, and the suite subprocess is bounded by its timeout. Nothing outlives the command except the suite's machine-lock watcher, which exits with the sdlc process (#319).

**Sequence tests** drive the real `runDurableMerge` against `landingFixture`. They push to the bare remote from a hook (the injected suite runner) to reproduce "main moved during the suite" deterministically.

### ARCH-FUNERAL

- **The merge commit** is part of the branch and lands or dies with it.
- **Suite logs** reuse #345's `<git common dir>/sdlc/local-checks/<sha>/`, which keeps the newest 20.
- **Baseline runs** are not recorded on GitHub (`record: false`), and their log goes to the same capped directory.
- **Statuses** are bounded by GitHub per commit.
- **Creates nothing else durable.**

### ARCH-MOCK

Stateful fakes:
- git: real git against a bare remote (the established fixture);
- GitHub: `landingFakeGH` + `ciFakeGH`, which gains `CommitStatuses` reading its existing status map;
- the suite: the declared check's command (a script that passes or fails, and can push to the remote) or `runSuiteFn`'s seam for the baseline;
- make: `makeTestFactsFn`.

Live conformance: run `makeTestFacts` on ariadne's real Makefile, expecting has=true and weight=2, as a test that skips outside the repo.

### ARCH-PURPOSE

The point is that landings stop turning main red: the tested tree is the landed tree. Delivered in this repo:
- the enforcement at every landing (b);
- integration;
- conflict refusal;
- failure labelling;
- bounded retry;
- the legacy refusal;
- the docs that make `sdlc local-check run --name suite` the close suite.

Nothing deferred.

---

## Plan milestones

### M1 — the suite as a recorded, enforced check

**Files:** `localchecks.go` (+test), `localchecks_make.go` (+test), `localcheck_run.go` (+test), `localcheck_cmd.go` (+test), `cigate.go` (+`cigate_test.go`, `cigate_fake_test.go`), `ghclient.go`, `fetch_test.go` (stubGH), help (`local-check.md`, `merge.md`, `close.md`), `closefixed.go:34`, `publishgate.go:299`, `close.go:1852` (prose: `sdlc local-check run --name suite`), `AGENTS.base.md` §5 (+ regenerated `AGENTS.md`), atlas `ci-merge-check.md`.

- [ ] Tests:
  - `TestParseLocalChecks` for `weight`: valid; 0 and > max refused.
  - `TestEffectiveLocalChecks`: the default is added only without a declared `suite` and with a test target; a declared one wins.
  - `TestMakeTestFacts`: probe a fixture Makefile with and without a test target and with `WF_TEST_WEIGHT ?= 2`; assert the lock helper is never invoked (a fake `test-lock.py` that fails loudly).
  - `TestAcquireSuiteSlotWeight`: the fake helper records argv and sees `--slots 2`.
  - `TestDecideCIGateSuiteAncestor`: an ancestor record counts only through a docs-only distance; a code change invalidates it.
  - `TestRunCIGate_SuiteRecordedBeforeClose`: through `ciFakeGH`, a record on the pre-close head plus close's doc commits passes with no rerun; a code commit after the record triggers the inline run.
  - `TestLocalCheckCmd_DefaultSuite`: `sdlc local-check run --name suite` works with no declaration.
  - Live `TestMakeTestFactsAriadne`.
- [ ] Implement it, update the docs, and commit. Then **M1 boundary: `sdlc milestone-close --issue 332 --milestone M1`**, and report to the hub.

### M2 — integration loop in the durable path; legacy refusal

**Files:** `landingintegrate.go` (+test), `landing.go` (the integration loop around the gates), `cigate.go` (`waitCIGate`), `merge.go` (legacy refusal), `internal/recovery/catalog.go` (merge contract), `helptext/merge.md`, atlas.

- [ ] Tests:
  - `TestNextIntegrationStep`: the table above, every row, plus the attempt bound.
  - `TestLandingIntegratesAMovedMain`: main moved, no conflict; one merge produces the merge commit (trailer), push, suite recorded on the new head, and lands. Done-when row 1.
  - `TestLandingConflictRefusesAndLeavesTheBranch`: paths named, next action printed, local and remote heads unchanged. Done-when row 2.
  - `TestLandingIntegrationPublishCollisionResets`: main rewrote reviewed code lines with a clean git merge but a replay conflict; refused with fixed-then-ship, branch reset, nothing pushed.
  - `TestLandingMainMovesDuringSuite`: the suite hook pushes to main once; integrates twice and lands. With 3 pushes it refuses with moved-too-often.
  - `TestLandingResumesAnInterruptedIntegration`: a local trailer commit ahead of the remote is pushed and landed.
  - `TestLandingLeaseRefusesConcurrentPush`.
  - `TestWaitCIGate`: pending turns green within the bound; past the bound it refuses, naming the pending checks; red refuses at once.
  - `TestRunMerge_LegacyRefusesMovedMain`.
  - `TestRecoveryContractsAreProven` passes with the new proofs.
- [ ] Implement and commit. **M2 boundary: `sdlc milestone-close --issue 332 --milestone M2`**, and report.

### M3 — failure labelling with a baseline run

**Files:** `suitefailures.go` (+test, fuzz), `landingintegrate.go` / `cigate.go` (on a suite failure: the baseline, then the refusal), `localcheck_run.go` (`record:false`).

- [ ] Tests:
  - `TestFailingTests`: go test and test-shard summary lines, subtests ignored, deduplicated; `FuzzFailingTests` never panics and returns sorted unique names.
  - `TestClassifyFailures`.
  - `TestLandingSuiteFailureLabels`: the merged suite fails A and B, main's baseline fails B; the refusal reports A introduced and B pre-existing, and the baseline posts no status. Done-when row 3.
  - Main red with an unparsed failure gets the right message.
- [ ] Implement and commit. Then full `make -k test`, `sdlc close --issue 332`, PR, `sdlc merge --yes`, and report.
