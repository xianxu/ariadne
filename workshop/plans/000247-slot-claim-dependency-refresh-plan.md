# Explicit Weave Refresh Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach. Use superpowers-executing-plans for the shared discovery/compile integration; bounded independent test work may use subagents. Steps use checkbox syntax for tracking.

**Goal:** Provide explicit `weave refresh [--rebase]` that checks the host and declared substrate repositories together, advances their current branches to captured remote-main commits, then compiles.

**Architecture:** Reuse acquisition discovery and numbered-environment validation. A small refresh package captures immutable per-repository observations, validates the whole set, then applies Git operations sequentially with revalidation. Compile shares the existing prepared setup context and lease; no claim hook, journal, inventory or automatic rollback.

**Tech Stack:** Go, Cobra, existing Weave acquisition/layergraph/setup code, Git CLI, disposable local bare Git repositories for integration tests.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|---|---|---|
| Snapshot | cmd/weave/internal/refresh/model.go | new |
| Prepared | cmd/weave/internal/refresh/model.go | new |
| Phase / advance | cmd/weave/internal/refresh/model.go | new |

- `Snapshot` holds canonical checkout/common-dir identity, symbolic branch, initial HEAD, effective origin identity, captured target SHA, and parsed dependency declarations. One per deduplicated substrate checkout, including the host. It is invocation-local, not a receipt.
- `Prepared` owns an ordered snapshot set only after every preflight check succeeds. Its fields are private to refresh; the IO shell cannot apply an unchecked list. Pure validation selects fast-forward-only eligibility or explicit rebase and aggregates per-repository blockers. Equality is current; ahead and divergent HEADs refuse by default.
- `Phase` enumerates inspecting, ready, applying, compiling, complete, stopped. `advance` validates events and returns the permitted next operation. Snapshots and phase die at return. This makes the all-checks-before-first-update invariant explicit (ARCH-PURE, ARCH-ORDER).

### Integration points

| Name | Lives in | Status | Wraps |
|---|---|---|---|
| acquire.Client.Restore | cmd/weave/internal/acquire/acquire.go | reused | declaration parser, graph resolution, source and private-checkout validation |
| refresh Git probes | cmd/weave/internal/refresh/git.go | new | Git observations and update arguments through acquire.GitRunner |
| acquire.ExecGit | cmd/weave/internal/acquire/git.go | modified | shared bounded cancellable Git process adapter; preserve existing callers |
| buildRefresh / runRefresh | cmd/weave/refresh.go | new | Cobra, setup context, refresh engine, compile |
| compilePrepared | cmd/weave/main.go | new extraction | existing compile body after prepareSetup |
| prepareSetup | cmd/weave/environment.go | reused | verified environment policy and inherited setup lease |
| Git model | cmd/weave/internal/refresh/refresh_test.go | new test fixture | mutable refs, graph, dirty/operation state, failures and ordering |

`discoverRefresh` wraps `Restore(ctx, root, true)` for read-only discovery/validation. It refuses either a returned error or nonempty `Result.Missing` before any fetch/update, then consumes `Result.Layers`, already deduplicated and foundation-first. Existing Restore already returns an incomplete-graph error after traversal when Missing is nonempty (`cmd/weave/internal/acquire/acquire.go:370–372`); the wrapper makes refresh's strict contract explicit without relying solely on that current implementation detail or changing ordinary setup behavior. It does not clone during preflight. Missing checkouts receive guidance to run ordinary setup first. Source-less substrate declarations retain existing behavior: the existing checkout must still have a usable origin/main for refresh.

Only the host and transitive `substrate` dependencies are Git refresh subjects. `data` mounts keep their existing compile/acquisition behavior; they are not Ariadne layers and are not rebased by refresh. Their declarations remain part of the topology check described below. No new declaration parser or path inventory (ARCH-DRY).

## Behavior and safety boundaries

### Discovery and preflight

1. Resolve the invocation repository root, rejecting detached/unborn HEAD and invalid Git topology. Obtain the existing setup context with `prepareSetup`; preserve lexical numbered-environment evidence until validated.
2. Discover existing substrate checkouts using read-only Restore. Gather baseline observations for **all** subjects before fetching. Clean means tracked index/worktree and nonignored untracked files are clean, including dirty submodules. Refuse active merge, cherry-pick, revert, rebase, sequencer and bisect states using Git-resolved paths; fail closed on read errors.
3. Fetch each origin with explicit main refspec and no recursive submodule fetch. Immediately resolve the fetched remote-tracking main to a full commit SHA once; this becomes the preflight target. Use `fetch --no-recurse-submodules --no-tags origin +refs/heads/main:refs/remotes/origin/main` and validate that captured commit before checking fast-forward eligibility. All subsequent checks and updates use the captured SHA; no later ref resolution supplies an update target. The implementation tests must cover a remote-tracking ref moving after capture. Record full validated object IDs, not abbreviated output.
4. Check reachability of captured HEAD to captured target for default mode. `--rebase` permits local commits; both modes still require clean/no-operation checkouts. Aggregate identified blockers; if any fetch/probe/check fails, no working branch may advance.
5. Require parsed `construct/deps` rows at each captured target to equal the checkout's captured rows. Ignore comment/format-only differences through `layergraph.ParseRows`; treat missing file as empty, but reject nonordinary blobs and read errors. A changed dependency path/source/mount stops preflight with the affected repository and instructions to reconcile that declaration change separately. This initial boundary prevents compile from discovering unchecked repositories after refresh. It is intentionally simpler than adding a second graph planner over prospective Git trees.

Preflight does not install, clone, build, publish issues or rewrite local branches. Fetch can change remote-tracking refs and download objects. The captured snapshot is per repository, not one global point in time.

### Application

Before the first update, revalidate the whole snapshot set; before each individual update, repeat checkout identity, symbolic branch, initial HEAD, effective origin, cleanliness and no-operation checks for that repository. Refuse a redirected path, changed remote, branch switch or new local commit. Use the **captured SHA** even if origin/main has subsequently moved.

- Default: `git -c submodule.recurse=false -c merge.autoStash=false merge --ff-only --no-edit --no-overwrite-ignore TARGET_SHA`.
- Opt-in rebase: `git -c submodule.recurse=false -c rebase.autoStash=false -c rebase.updateRefs=false -c rebase.autoSquash=false rebase --no-fork-point --no-autostash TARGET_SHA`. No auto-stash, update of other branches or pull policy inference. Document ordinary Git rebase semantics for local commits.
- Equal initial/target SHAs require no Git update but still require readiness checks.
- Verify the resulting current branch and target reachability after each operation; recheck cleanliness/no-operation and declarations before compile. Report path, captured starting/target SHAs and confirmed progress. An error or cancellation is not proof that an update did nothing.

Run updates foundation-first using the existing resolved order; the host is last. A failure stops subsequent mutations and compile. Earlier successful updates remain. Rebase conflicts remain for the operator to resolve or abort explicitly; rerun starts a new preflight after the active operation is finished. Never auto-abort or roll back.

Rechecks and Git's native safety checks do not make arbitrary editor/Git activity atomic. Numbered refresh participates in the existing environment setup lease. Primary/shared siblings retain the existing cooperative setup limitations; operators must avoid concurrent Git writers. This change adds no fleet-wide lock or second owner registry.

### Compile and retry

Extract `compilePrepared` from `runCompile` so ordinary compile and refresh share one implementation without reacquiring the setup lease. `runCompile` remains the public wrapper that prepares/defer-closes setup; `runRefresh` prepares once, holds the lease across refresh and compile, and calls the helper with the same acquisition client and runner. Git children inherit lease FDs just as existing setup children do.

Revalidate the complete final subject set and confirmed post-update HEADs before compilation. Run read-only Restore again and require the same ordered layers, parsed declaration rows and mounts as preflight; a changed rebase result or path binding refuses before acquisition. Preserve declaration row order because it affects graph/data placement. Always compile after successful refresh, including when all targets were already current. Failed compile reports that Git updates remain; ordinary rerun fetches new targets and compiles again. Compilation retains existing Brewfile, owner-tool, generation, and machine-wide installation policy. No command is added to `sdlc claim`.

### States and events

| State/event | Next state and permitted effects |
|---|---|
| inspecting / all checks pass | ready; retain immutable snapshots |
| inspecting / any blocker or cancellation | stopped; report blockers, no branch update |
| ready / full revalidation passes | applying; first update allowed |
| applying / next repository still matches | applying; update captured target, verify and record progress |
| applying / failure, conflict, stale state or uncertain result | stopped; no later update or compile; preserve progress |
| applying / all updates confirmed and final revalidation passes | compiling; invoke shared compile once |
| compiling / success | complete |
| compiling / failure or cancellation | stopped; Git progress remains; ordinary rerun allowed |

No persisted refresh phase. Retry derives truth from Git and declarations. No concurrent/background refresh tasks outlive the command; cancellation waits for owned child shutdown. Direct external writers are outside the setup lease (ARCH-ORDER).

## Operating envelope and trust

Interactive batch operation over a small developer layer graph. Sequential work; initial limit 128 unique layers, chosen well above the current few-layer environments. Enforce the discovery limit before further queue expansion (optional acquisition discovery limit, zero preserves other callers), not after a complete unbounded walk. Explain the limit in errors; no truncation. Cap dependency document reads at 1 MiB (including acquisition discovery: bounded reader before allocation/parsing, not only Git target blobs) and Git captured output at 4 MiB per command; reject overflow. Initial per-Git-command timeout 2 minutes with bounded child/pipe shutdown; compilation retains its existing signal-aware lifecycle rather than a new arbitrary overall deadline. Reassess limits using the chain fixture and representative clone timings (ARCH-CONSTRAINTS).

Treat Git stdout, paths, remote config and declaration blobs as parsed external evidence. Validate full SHAs, symbolic refs and exact checkout roots; use argv arrays, never shell interpolation. Preserve path bytes through Git adapters rather than generic TrimSpace. Do not print credentials from origin URLs in diagnostics. Fake remotes use temporary folders only; no real-network writes in tests (ARCH-SECURE).

No new durable receipt, index or inventory. Reuse existing setup lock and compilation ownership/staging lifecycle. Git retains fetched objects and reflogs under its normal maintenance. Interrupted rebases retain standard Git state until explicitly resolved or aborted (ARCH-FUNERAL).

## Test strategy and conformance

| Function | Strategy and mechanical guard |
|---|---|
| `eligibility` | Pure relation/mode table: only equality/ancestry authorize default mode; mutation of this predicate must fail the table. |
| `advance` | Exhaustive state/event matrix and generated event sequences: no apply before ready and no compile before all confirmed updates. |
| `parseOID`, `parseBranch`, `parseRecord` | Fuzz malformed/truncated/framing inputs; reject ambiguity and preserve path bytes. |
| `observe`, `readDeclarations` | Real Git and bounded file fixtures: errors/active operations/dirty state cannot become readiness, nonordinary or excessive input cannot become declarations. |
| `discoverRefresh` | Missing/incomplete results must refuse before fetch; `TestRefreshMissingCheckout` verifies no mutations, while existing `TestRestoreDryRunMissingAndLocalCycle` and `TestCompileDryRunDoesNotMutate` defend ordinary setup semantics. |
| `Run` preflight | Mutable stateful backend: any repository's blocker prevents every branch update; prove by independent snapshots of refs/index/files. |
| `revalidate`, `Run` application | Inject external changes at each observation/effect boundary; only the captured SHA is a destination and stale starting evidence prevents the next effect. |
| `runRefresh`, `compilePrepared` | Real command/compile with portable tool fixtures: compile happens once after confirmed updates, retry runs compile even for current refs, and one lease spans both. |
| `acquire.Client.Restore` | Adversarial graph/document input with a counting backend: reject at configured limits before unbounded work/allocation; existing unlimited callers remain compatible. |

`TestGitConformance` exercises the same GitRunner contract directly and through
a stateful repository-backed event adapter. Temporary real Git repositories own
the graph, refs, index, files and operation state; the adapter injects external
changes and uncertain responses at deterministic IO boundaries. This reuses Git's
actual semantics rather than maintaining a second implementation of Git. Additional
sequence tests drive the same adapter through rebase, conflict and interruption.
Conformance runs on every package test invocation, with no network or credentials.
Compilation tests reuse real owned Weave code and temporary tool payloads.

## Chunk 1: Implementation and acceptance

One atomic delivery and one SDLC close review; no Mx boundary tags.

### Task 1: Pure refresh contract and Git probes

**Files:** create `cmd/weave/internal/refresh/model.go`, `model_test.go`, `git.go`, `refresh_test.go`; extend `cmd/weave/internal/acquire/git.go` without changing existing caller semantics.

- [ ] Write failing tests for `eligibility`, `advance`, parsers and Git probes using the strategy table.
- [ ] Run `go test ./cmd/weave/internal/refresh -count=1`; confirm intended failures, then implement typed snapshots/prepared state and rules.
- [ ] Reuse acquire.GitRunner; add opt-in raw/bounded execution to ExecGit, with cancellation and inherited setup descriptors. Implement probes and the stateful model through the shared seam.
- [ ] Run pure, probe and `TestGitConformance` tests; commit the verified unit.

### Task 2: Discovery, preflight and update orchestration

**Files:** create `cmd/weave/internal/refresh/refresh.go`, `refresh_test.go`; modify `cmd/weave/internal/acquire/acquire.go`, `git.go` for bounded discovery options; add colocated acquisition regression tests.

- [ ] Write failing `Run`, `revalidate` and bounded-discovery tests per the strategy table, using real Git checkouts plus deterministic external-event injection.
- [ ] Implement read-only Restore integration, sequential preflight, immutable targets, graph comparison and transition-controlled updates with accurate partial-progress reporting.
- [ ] Implement apply-time/final revalidation and failure recovery behavior from the contract above; no automatic rollback or acquisition of unchecked repositories.
- [ ] Run `go test ./cmd/weave/internal/refresh ./cmd/weave/internal/acquire -count=1`, then commit.

### Task 3: CLI and shared compile integration

**Files:** create `cmd/weave/refresh.go`, `refresh_test.go`; modify `cmd/weave/main.go`, `environment.go` only as needed, `environment_test.go` and relevant startup tests.

- [ ] Write failing command and compile-orchestration tests per the strategy table.
- [ ] Register `buildRefresh`; extract `compilePrepared` without changing ordinary compile behavior. Wire refresh through one prepareSetup/close lifetime and always compile successful passes.
- [ ] Verify lease lifetime and independent-environment preservation through existing real setup fixtures; verify cancellation leaves no command-owned child running.
- [ ] Run `go test ./cmd/weave/... ./pkg/layergraph/... ./pkg/workspace/... -count=1`, then commit.

### Task 4: Documentation, verification and publication

**Files:** modify `README.md`, `atlas/workflow/weave.md`, `atlas/workflow/workspace-branching.md`; update issue and this plan. No implementation step edits Pair: leave referencing project discovery/ticking to the existing SDLC close gate. Any final peer publication must select only this issue's documentation delta through `sdlc issue publish`, preserving concurrent peer work; it is not a runtime or blocking dependency.

- [ ] Document the contract above in README, atlas and CLI help. Confirm existing atlas index links remain valid; clarify ordinary compile preserves existing revisions.
- [ ] Run full Weave/layergraph/workspace tests, `go test -race ./cmd/weave/internal/refresh`, `go vet ./cmd/weave/...`, build `bin/weave` in this checkout, and smoke-test help in a temporary environment. No live refresh of this working checkout.
- [ ] Mutation-check ancestry refusal, captured-SHA application and all-repository preflight; each disabled guard must fail its test, then restore and verify. Record runtimes and evidence.
- [ ] Update Log/lessons and commit. Close through `sdlc close --issue 247 --verified 'actual evidence'`, address review findings, open PR through SDLC. Merge only on operator instruction.

## Review and approval

Implementation is authorized in this checkout. Full-flow plan-quality accepted after three rounds; estimate-quality accepted the derived provisional 4.24h with advisory optimism notes. The topology-change refusal and substrate-only refresh boundary remain the agreed implementation constraints.

## Revisions

### 2026-09-23 — Fresh planning review

The independent spec/plan reviewer approved with no blocking findings. Clarified
that the dependency-document read limit applies inside reused discovery before
allocation, and that raw Git output must preserve acquisition caller semantics.
Engineering-plan approval remains the next step; no implementation has started.

### 2026-09-23 — Implementation authorized in the primary checkout

Operator requested continuing here because local PATH uses this checkout's bin/.
This supersedes the new-slot handoff and pending approval text above. Plan gate
PQ-1/2/3 refinements name test functions/strategies, define shared real-Git/model
conformance on every package run, and leave peer-project ticking to SDLC close.
The implementation contract is unchanged. No runtime code has changed yet.

### 2026-09-23 — Strict discovery contract clarified

PQ-4: refresh uses discoverRefresh to reject both discovery errors and nonempty
Missing results before fetch/update. This explicitly guards the boundary while
preserving Restore and ordinary compile/dependencies semantics; named regression
coverage distinguishes these paths. The current Restore already returns an
incomplete-graph error at the end of traversal, which the reviewer had missed.

### 2026-09-23 — Implementation structure and regression discoveries

The engine keeps one Run orchestration with a private Prepared snapshot set,
rather than exporting separate preparation/application entrypoints. Its stateful
Git fixture uses real temporary repositories plus deterministic event injection;
there is no second Git implementation to maintain. Acquisition now exposes one
bounded ReadDeclarations helper reused by refresh. Tests exposed predicate-exit
error ambiguity after cancellation/overflow and ignored-path collisions in
intermediate rebase commits; both classes were corrected with regressions.
