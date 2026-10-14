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

## Revisions

### R1 — 2026-10-13, plan-quality round 1

**PQ-1 (Important), the class "enforcement configuration read from the branch under test".** Every input that decides *whether* and *how heavily* the suite is required must come from main. The *test subject* is the head. The enumeration:

| Input | Source before R1 | Source now |
|---|---|---|
| Declaration (`.sdlc/local-checks.json`) | main's tree (#345) | unchanged |
| Default-suite facts (`test` target exists, `WF_TEST_WEIGHT`) | branch checkout ✗ | **main**: `makeTestFacts` runs the probe in a detached worktree at the gate's base (main's sha), removed after (the #345 worktree helper, ARCH-DRY) |
| Lock helper location (`Makefile.workflow` realpath) | the slot root | unchanged: the machine lock is slot infrastructure, not branch content; a branch cannot raise or lower the cap through it |
| Suite command execution | head (detached worktree at head) | unchanged: the head is what is under test |

So a branch that deletes `test:` or lowers `WF_TEST_WEIGHT` still owes the suite at main's weight. Test: `TestDefaultSuiteFactsComeFromMain`, where the branch deletes `test:` and lowers the weight, yet the gate still requires `sdlc-local/suite` and acquires main's weight.

**PQ-2 (Minor), PR head sync.** After the lease push, `integrateMain` re-reads the PR (`gh.LandingPRs` → `selectLandingPR`) until `pr.HeadOID` equals the pushed merge commit. It polls every 5s, bounded at 2 min, then refuses with "GitHub has not registered the pushed head; re-run sdlc merge". Only then do the publish gate and the head-match checks run, against the re-read `pr` (HeadOID and BaseOID). This is covered in `TestLandingIntegratesAMovedMain` with a fake that lags one read.

**PQ-3 (Minor), worst-case wall time.**
- **One integration costs** the suite (about 6–13 min plus lock queueing), the CI wait (≤20 min) and the PR-head sync (≤2 min).
- **The worst case** is 3 × (13 + 20 + 2) ≈ 105 min, plus one baseline (≈13 min) on failure, so about 2 h in the foreground.
- **This is acceptable:** merge is batch work that agents already run in the background, and each bound is printed while it applies. A run that exceeds them stops with a named next action rather than hanging.

**PQ-4 (Minor), non-goals.**
- No rebasing; integration is a merge commit.
- No automatic conflict resolution.
- No re-review after a clean integration, beyond the publish gate's existing replay rule (#304).
- No change to the server-side merge method (`gh pr merge --merge --match-head-commit`).
- No integration on the legacy (non-slot) path; it refuses.
- No suite for repositories with neither a declaration nor a `make test` target; the requirement then does not apply.

### R2 — 2026-10-13, M1 boundary review

- **BR-1 (Important), the class "path lists fed to the docs/code classifier must name every changed path".** With git's default rename detection, `git diff --name-only` lists a rename only by its new name, so code renamed to `.md` read as docs. That would keep an old suite record valid, and also let the publish gate pass such a rename. The feeders, all now built by one `gitx.ChangedPathArgs` (`--no-renames -z`):
  - the ancestor walk;
  - `gitx.DiffNames` (publish gate, review anchor, close window, review window plan);
  - freshness's diff;
  - the tracker migration's anchor classification.

  Tests: `TestDiffNamesListsBothSidesOfARename` and the rename case in `TestLocalRecordsOnAncestors`, both of which fail with the flag removed.
- **BR-2 (Important):** direct tests for `parseCommitStatuses` (newest first) and `localRecordsOnAncestors`: merge-base bound, early stop, cap, root commit.
- **Minors:**
  - a `rev-list` error is returned, with the root commit detected explicitly;
  - make's "Not a target" entries are skipped;
  - the probe has a 2-minute timeout;
  - one shared `withDefaultSuite` serves the gate and the verb.
- **Names as delivered** (the plan table said otherwise): `localRecordsOnAncestors` (was `suiteAncestorRecords`); `makeTestFacts(git, rev)` probing a worktree at main (was `makeTestFacts(root)`); a declared `weight: 0` means the default 1, and only negative or >8 is refused.

### R3 — 2026-10-13 23:08 PDT, M1 boundary review round 2

- **BR-8 (Critical):** execution overrides cannot disable the mandatory suite.
  `effectiveLocalChecks` preserves a declared suite's command, timeout and weight
  while making it required. `withDefaultSuite` applies that same derivation to
  both declaration paths. Tests cover omitted, false and true `required`, and
  exercise the real CI gate with omitted/false declarations (refusal without a
  runner; a recorded suite run before admission). The bypass tests failed before
  the fix. This follows ARCH-PURPOSE and ARCH-DRY.
- **BR-3:** positive parent evidence comes from `rev-list --parents --max-count=1`.
  Read failures propagate instead of implying a root commit; injected parent and
  ancestor-walk failures are regression cases, alongside the real root fixture.
- **BR-5:** the make probe uses the existing process-group helper to terminate
  descendants, with a one-second pipe-drain bound after its two-minute deadline
  (or parent exit). Tests cover a hanging make shell and a parent that exits while
  its child retains pipes, and assert the child exits. Makefile stat errors other
  than absence also propagate. Both process cases and the stat case failed before
  correction (ARCH-CONSTRAINTS, ARCH-ORDER).
  The same lifetime defect existed in `runLocalCheck`; its timeout and
  parent-exit paths now terminate the group before removing the worktree too.
  Both runner descendant tests failed without this correction. Its existing
  five-second pipe-drain limit remains.
- **BR-9:** README joins help and atlas in documenting the close-suite invocation,
  automatic default, mandatory declared suite, ancestor reuse and weight setting.
  M1 documentation includes `README.md`; process tests live in
  `localchecks_make_process_test.go`, enforcement/error tests in
  `suite_enforcement_test.go`.

### R4 — 2026-10-13 23:26 PDT, M1 boundary review round 3

- **BR-10:** all four descendant tests (probe/runner × timeout/parent exit)
  must bound their exit assertion below the child's 60-second natural lifetime.
  The runner already used one second; the probe now does too. Removing probe
  process-group configuration and both termination calls makes both cases fail
  on a surviving child (4.867s total), instead of waiting for natural exit.
  Runner mutations were already proven red in R3. ARCH-ORDER: a cleanup test
  must distinguish termination from eventual natural completion.
- **BR-11:** add opt-in read-only `TestLiveCommitStatusesConformance` through
  `realGH.CommitStatuses`, against ariadne PR #229's immutable head
  `38ca18d3d9e95151f2ad3659e278ca45fe954b07`. Its guard context has known pending,
  failure, and success records, so the live history and expected success test the
  endpoint and newest-first reduction. Document invocation and cadence in atlas:
  on changes to the adapter/parser/fake and before landing dependent behavior.

### R5 — 2026-10-14, M2 implementation details

- The integration transition model delegates suite/CI substeps to the existing
  CI gate under `integrationGates`; `waitCIGate` owns pending/unregistered polling.
  `ciGateError` preserves refusal reasons, and `ciGateFacts`/`prepareCIGateFacts`
  load immutable main/head configuration once per wait (ARCH-DRY). The wait
  excludes local-run duration from its 20-minute budget. These live in
  `cigate_wait.go` (+ tests), with `runPreparedCIGate` in `cigate.go`.
- Integration helpers as delivered: `integrateLandingMain`, `integrationResume`,
  `selectIntegrationPR`, `syncIntegrationPR`, `integrationAncestor`, plus
  `nextIntegrationStep`, all in `landingintegrate.go`. Resume requires an exact
  two-parent clean merge tree, not just a trailer. Its replay is checked before
  push too, including interruption before the original replay check.
- The legacy refusal uses `requireLegacyMainIntegrated` in `merge_main.go`,
  before and after checks. `merge_main_test.go` verifies both points leave local
  and remote issue heads unchanged. Existing codecomplete/archive setup now
  integrates main first, so it continues testing archive behavior past this new
  admission requirement.
- Main's final fetch narrows a remaining external race: `--match-head-commit`
  pins the PR head, not main atomically. The approved server merge method is
  unchanged; atlas states this limit explicitly.

### R6 — 2026-10-14, M2 boundary review corrections

- **BR-12 / ARCH-CONSTRAINTS:** the 20-minute CI budget now covers every
  polling read: PR head/checks, Actions job, ancestor statuses, and Git facts
  or ancestry reads. Each call receives the remaining deadline; expiry is
  latched and checked before accepting any result, including green. The local
  runner keeps its parent context and its entire elapsed duration (including
  lock queueing) is excluded. Worktree remove/prune cleanup uses fresh bounded
  IO after cancellation, separately from the admission budget.
- **BR-13 / ARCH-CONSTRAINTS, ARCH-ORDER:** enumerate the integration Git graph:
  target resolution, remote/ref reads, ancestor and resume probes, fetch, merge,
  replay reads and commit-tree, lease push, abort/reset, worktree and checkout
  cleanup. Durable landing commands use a contextual 2-minute command bound;
  abort/reset and worktree cleanup use fresh 30-second commands. Shared gitx
  defaults, environment commands, contextual input commands and history streams
  also have a 2-minute ceiling, covering publish/ownership paths that do not use
  the landing runner. Each subprocess owns its process group, bounds pipe drain
  to one second, and reclaims descendants after cancellation or parent exit.
  Cancellation during push is uncertain: retain the integration, then re-read
  Git/GitHub and validate its existing resume evidence on retry. No rollback
  follows a push whose publication may already have occurred.
- **BR-14 / ARCH-ORDER:** `integrationState` and `transitionIntegration` in
  `integrationstate.go` supersede the action-only observation helper. The core
  owns head, previous head (lease/rollback base), main, PR, attempts, resume and
  phase. Typed observations/outcomes produce the next state and effect: merge,
  replay, abort/reset, push, gates, land or refusal. Production effects report
  their results through this transition function, including failed push,
  recovery completion, and main moving after gates. Invalid phase/event pairs
  refuse. The shell only observes IO and executes the selected effects.
- Regression evidence: deadline tests reproduced slow reads without deadlines
  and accepted late green/Git results before correction. Removing retry counting
  fails state/bound tests; removing post-exit group cleanup fails landing and
  shared gitx descendant tests well before their 60-second natural lifetime.
  Real post-receive cancellation tests establish that the remote already
  published, then preserve both heads and recover landing without a duplicate
  integration.

### R7 — 2026-10-14, M2 boundary review round 7

The previous R6 enumeration was incomplete. BR-12's remaining-budget wrapper
covered the Git and GitHub read callbacks, but the default suite's Makefile
probe opened an independent background context. A 40 ms remaining budget let
that probe continue for 1.07 seconds. BR-13's bounded landing runner covered
its direct Git calls, but publish replay used gitx helpers with background
contexts and archive ownership reads used the separate global archive runner.
The reviewer reproduced a cancelled publish replay continuing for 1.06 seconds.
BR-14's owned state transitions were accepted.

Complete the same deadline through each actual production edge: CI facts
preparation, default-suite derivation and Makefile probe (with fresh bounded
worktree cleanup); publish replay's Git reads and environment commands; archive
ownership reads; direct landing transport/probes; and cleanup after a cancelled
operation. Preserve the existing two-minute per-command ceiling and the
20-minute CI wait, with suite runtime excluded. Add real slow-command tests
that fail on the previous background-context paths, then verify the existing
landing, publish and archive behaviors. ARCH-CONSTRAINTS, ARCH-ORDER and
ARCH-PURPOSE require this class sweep.

### R8 — 2026-10-14, R7 implementation detail and verification scope

The cancellation graph now has explicit context-bearing edges:

- **CI facts:** `prepareCIGateFactsContext` → `withDefaultSuiteContext` →
  `makeTestFactsContext` → `probeMakeTestFactsContext` uses the wait's
  remaining deadline; detached-worktree cleanup receives separate fresh
  bounded authority. The existing wrappers serve non-wait callers.
- **Publish replay:** `runLandingPublishGate` and its ownership/anchor reads
  call contextual gitx primitives through main-base resolution, replay
  merge-tree and commit-tree construction, diff, tracker repository lookup,
  close-token checks, and archive ownership snapshots. On cancellation, no
  empty-capture fallback can authorize publication. The quick-flow final
  diff uses a contextual whole-issue base and numstat read.
- **Transport and recovery:** direct durable landing Git remains under the
  two-minute per-command bound. Archive/tracker readers use their caller's
  context. Abort/reset and explicit worktree cleanup retain the independent
  30-second recovery bound.

Regression tests run real sleeping Makefile and Git commands and assert prompt
termination. Before the change, the Makefile probe took about 1.6s on both
budget expiry and cancellation, publish and archive ownership reads exceeded
the cancellation bound, quick-flow numstat continued after cancellation, and
contextual replay APIs were absent. Two additional tracker reader modes were
added after a call-graph sweep; both also failed before their correction.
The previous R6 sentence claiming complete coverage is superseded by this
enumeration and R7's reproduced counterexamples.

### R9 — 2026-10-14, M3 failure evidence and diagnostic baseline

The CI gate now returns a typed failed-local-check result with its check and
log record, so the landing shell can identify a failed suite without parsing
refusal prose. It reads the merged-head log before a diagnostic run can prune
older logs, then runs the same main-selected suite on main with status/comment
recording disabled. Pure `failingTests` and `classifyFailures` functions label
introduced and pre-existing tests; an unparsed failed baseline or merged run
is explicitly unknown rather than attributed to either tree. This preserves
ARCH-PURE and ARCH-ORDER: the classifier has no IO, and the baseline cannot
turn a failed gate into a landing. Real local Git/GitHub fake tests cover both
labels, unparsed results, and the absence of a main status.

### R10 — 2026-10-14, full-suite integration corrections

The first full run exposed a source guard that targeted M2's new
context-free compatibility wrappers, plus older end-to-end fixtures whose
static PR head and moved-main setup no longer represented the intended state.
Keep the contextual APIs as the production path; move fixture-only aliases
into tests and centralize the remaining legacy nil-context fallback at the
bounded Git execution seam. Preserve cancellation-independent recovery with
the caller's context values. Let landing fakes observe pushed heads, and bring
the legacy asset fixture's feature branch current before testing archive
materialization. These are compatibility and proof corrections, not a change
to M3's classification rule (ARCH-ORDER, ARCH-CONSTRAINTS).

### R11 — 2026-10-14, M3 boundary BR-15 lock-owner reentry

The M3 review reproduced a real machine-lock self-wait: after one local suite
run acquired the slots, the next run's helper returned `reentrant`, but the Go
runner returned an empty holder and omitted `WF_TEST_LOCK_HELD_BY` from the
nested `make test`. The suite then waited on slots held until `sdlc` exited.
This applies to the diagnostic baseline, integration retries, and consecutive
named local checks. Keep the actual owner per lock scope when the helper first
acquires it; on reentry, return that owner, or the inherited parent holder if
the current process nested under another make. `skipped` remains unowned.
Real-helper tests use an isolated lock file and a nested make for the baseline
and repeated check sequence. Both tests failed before the correction and
passed after it (ARCH-PURPOSE, ARCH-MOCK, ARCH-ORDER).
