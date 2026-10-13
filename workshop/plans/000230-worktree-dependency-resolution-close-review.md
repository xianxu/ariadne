# Boundary Review — ariadne#230 (whole-issue close)

| field | value |
|-------|-------|
| issue | 230 — Make Weave dependency resolution worktree-aware |
| repo | ariadne |
| issue file | workshop/issues/000230-worktree-dependency-resolution.md |
| boundary | whole-issue close |
| milestone | — |
| window | 08afcec554b1d5cb4ff03b84f57352274df5db1d..d9d786c27c5498644377d8fe47727ce7127130d0 |
| command | sdlc close --issue 230 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T13:25:50-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking for existing Git-reader implementations and env handling to compare against the new `ExecGit`.

Running targeted tests; two failures look environmental — checking them against the base commit.

Found a likely spec-5 gap: `pathExists` uses `Stat`, so a dangling symlink beside the worktree would count as Absent. Verifying with a scratch test.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** The core design holds up and is correctly wired. There is one pure chooser (`ChooseSubstrate`), one Git boundary (`workspace.DependencyAnchor`), and one `layergraph.Resolver` that composes them. All three earlier resolvers now go through it: the layergraph walk (and through it vocabulary, datatype and weave's read-only verbs), `acquire.Restore`, and sdlc's `substrateWalk`. I ran a mutation check: making `GitAnchors.PrimaryCheckout` always fail turns 8 layergraph tests red, plus every new acquire, weave and sdlc worktree test. So the tests really do exercise the fallback. Three things should be fixed before SHIP, all cheap:
- **Dangling local symlink is bypassed.** A dangling symlink beside the worktree is treated as absent, and resolution silently switches to the primary's sibling. That breaks spec item 5. I reproduced it.
- **No-anchor reasons are dropped.** They are computed, then thrown away in compile errors, although the plan (Decision 5) and the issue's ticked Task 4 promise them.
- **README is now wrong.** It does not describe the new linked-worktree behaviour, and one sentence in it ("record a source to restore a missing checkout") is no longer true inside a linked worktree.

Targeted suites pass. `pkg/workspace` (`TestEnvironmentLegacyAndArchive`) and `cmd/datatype` (`TestFindRepoRoot_GitFallback`) fail because this review checkout's TMPDIR sits inside a Git repo. The workspace failure reproduces at base `08afcec5`; the datatype failure is the same environmental shape (its temp dir is under the checkout's repo), though I didn't re-run that one at base. Neither looks caused by this diff.

Done-when: current — the Revisions entries (absent-everywhere refusal; no fallback across a numbered environment) are covered by "Missing dependencies or unavailable fallback anchors produce useful errors". The diff builds nothing outside the existing bullets. The "unavailable fallback anchors" half is only partly delivered (Important #2).

### 1. Strengths
- **Pure chooser, properly tested.** `pkg/layergraph/substrate.go:28` is pure. Its table test (`substrate_test.go`) checks the result and also the exact probe sequence, so "local layer wins without probing the candidate" is pinned.
- **No Git run outside a checkout root.** `DependencyAnchor` checks `dir/.git` with `Lstat` before running Git (`pkg/workspace/anchor.go:33`), so a non-checkout layer never borrows an enclosing repo's primary. A counting reader proves zero Git calls (`anchor_test.go:56`).
- **Numbered-environment policy keeps its own refusal.** The `c.Policy == nil` guard (`acquire.go:326`) leaves the environment's refusal and message in place. The existing `TestDependencyFeatureCompileCannotFallBackToGenericAcquisition` catches a regression there.
- **Real Git fixtures throughout.** The fixtures use real `git worktree add` and cover spaces, symlink vantage, transitive deps, absolute rows, an invalid fallback, and the parley.nvim#254 vocabulary regression (`worktree_test.go:191`). `TestCompileLinkedWorktreeMissingDependencyPreservesOutput` snapshots the whole worktree, so "output preserved on failure" is pinned.
- **Lookups are lazy and memoized.** Git runs only when a relative target is absent, once per owner per walk (`anchor.go:59,82`).

### 2. Critical findings
None.

### 3. Important findings
1. **Dangling or unreadable local target is silently bypassed (spec item 5; Done-when bullet 2).**
   - **Where:** `pkg/layergraph/anchor.go:51-58` classifies the local target with `hasManifest` and `pathExists`. Both use `FS.Stat`, which follows symlinks.
   - **What happens:** a dangling `fleet/worktree/leaf/base -> fleet/gone` probes as `Absent`, and resolution picks `fleet/base` via the primary. Verified with a scratch test: `Walk = [fleet/base, br]`, `err = nil`. Any other `Stat` error (EACCES, for example) is also treated as absent. `Restore` inherits the same behaviour. Before this change, `Restore` would `Lstat` the dangling link and fail.
   - **Why it matters:** the issue's own manual workaround is exactly this kind of sibling symlink.
   - **Fix sketch:** classify "present" with `Lstat` (add `Lstat` to `layergraph.FS`, or treat any `Stat` error other than `IsNotExist` as `NotLayer`). Add a regression row in `worktree_test.go`.
2. **No-anchor reasons are computed, then dropped (plan Decision 5; issue Task 4 "anchor reason in missing errors"; Done-when bullet 4).**
   - In `acquire.go:320-333`, any `reason` that is not `FallbackAbsentError` (bare-primary worktree, Git failure, unlisted or unavailable primary) is discarded.
   - Concretely, for a worktree whose anchor lookup fails:
     - A sourceless row still fails with the old "record its source" text, with no hint of why the fallback didn't apply.
     - A sourced row silently clones a duplicate beside the worktree, which is the exact outcome the issue wanted to stop.
   - `DeclaredSubstrate.Via` and `.Unanchored` have zero production consumers. `DeclaredSubstrates` is called only from tests, so the provenance never reaches a user.
   - **Fix sketch:** add `; worktree fallback: <reason>` to the two missing-substrate errors (the `recoverSource` failure and `Ensure`'s "missing repository") when `res.Probe == Absent && reason != nil`. Add a test with a bare-primary or failing-anchor fake through `Client.Anchors`. Otherwise, record a Revision that withdraws the claim (ARCH-PURPOSE).
3. **README not updated for linked-worktree resolution, and now contradicts behaviour.**
   - `README.md:103-110` explains source recovery in primaries and Couch slots, and ends "Without either, record a source to restore a missing checkout."
   - In a linked feature worktree, a sourced row that is absent beside both the worktree and the primary now refuses instead of restoring.
   - The new fallback, its `weave: substrate … resolved via primary checkout …` notice, and the refusal are only in atlas.
   - **Fix:** add a short paragraph next to the Couch-slot text.

### 4. Minor findings
- **Fake seam declared but unused (ARCH-MOCK).** `NewResolver` falls back to `GitAnchors` for any FS, including in-memory fakes passed through `walk.Walk`. The plan says fakes stay hermetic by implementing `Anchors`, but no fake does, so a fake FS consults real disk and Git.
- **Duplicate Git reader (ARCH-DRY, ARCH-CONSTRAINTS).** `workspace.ExecGit` (`pkg/workspace/anchor.go:14`) nearly duplicates `cmd/weave/environment.go:18` `setupGitReader`, minus the context. The vocabulary, datatype and sdlc walks therefore run Git with no cancellation or timeout. Consolidate on a context-taking `workspace.ExecGit`.
- **Atlas overstates the refusal.** `setup-and-replication.md` says the refusal happens "before compile writes anything". Rows dequeued earlier may already have been acquired (`Ensure`). Compile output is untouched, but the wording claims more than that.
- **Misleading field name.** `Unanchored` also carries `FallbackAbsentError`, which is the case where an anchor *was* found.
- **Redundant probing.** `Resolver.Substrate` probes the local target twice (the pre-check, then again inside `ChooseSubstrate`). This is negligible.

### 5. Test coverage notes
- **Pinned:** fallback, local-wins, invalid local, invalid fallback, transitive, absolute, spaces, symlink vantage, the no-anchor present-skip, vocabulary regression, compile notice, output preservation, refusal in both real and dry runs, and the in-environment feature worktree.
- **Missing:** a dangling or unreadable local target (Important #1), and a failing-anchor path through `Client.Anchors` in `Restore` (Important #2). No acquire or weave test injects `Client.Anchors`, so that seam is never exercised.

### 6. Architectural notes
| Principle | Result | Notes |
|---|---|---|
| ARCH-DRY | flag (Minor) | The resolution rule exists once; only the `ExecGit` duplicate is flagged. |
| ARCH-PURE | pass | The chooser is pure and tested without IO; Git sits at one boundary. |
| ARCH-PURPOSE | flag | All three consumers derive from the Resolver, but read-only consumers never surface `Via` or `Unanchored` (Important #2). |
| ARCH-MOCK | flag (Minor) | Production goes through the `Anchors` seam, but no fake implements it; the anchor is tested with real Git plus an existing workspace conformance fixture. |
| ARCH-CONSTRAINTS | pass with note | Lookups are lazy and memoized, bounded by layer count; no Git timeout. |
| ARCH-SECURE | pass | Git output parses through the existing `loadTopology` parser; no secrets involved. |
| ARCH-ORDER | pass | The only state is a per-walk memo; Restore's acquisition state machine is untouched. |
| ARCH-FUNERAL | pass | The change creates nothing durable and removes the duplicate-clone artifact. |

### 7. Plan revision recommendations
- **Decision 5 and Task 4 Step 3:** the "`; worktree fallback: <reason>` wrapping" either gets implemented (Important #2) or needs a Revision saying only `FallbackAbsentError` is surfaced, as a refusal. The test name `TestRestoreMissingEverywhereNamesAnchorReason` became `TestRestoreLinkedWorktreeMissingEverywhere`.
- **Core concepts / Resolver:** "A test FS may implement `Anchors` itself … so in-memory fakes stay hermetic" — either wire a fake, or revise to say the default for every FS is `GitAnchors`.
- **Decision 2:** state that "present" means `Lstat`-present (a dangling link counts as present-invalid) once Important #1 is fixed.

```findings
findings:
  - id: new
    severity: Important
    family: existing-target-not-bypassed
    title: |
      Dangling or unstat-able local substrate target is classified Absent and silently bypassed for the primary fallback
    detail: |
      pkg/layergraph/anchor.go:51-58 probes with FS.Stat (follows symlinks); a dangling fleet/worktree/leaf/base symlink yields Walk=[fleet/base, br] with no error (reproduced). Spec item 5 requires an existing invalid local target to be an error. Probe presence with Lstat (or treat non-IsNotExist Stat errors as NotLayer) and add a regression row.
  - id: new
    severity: Important
    family: computed-diagnostic-not-surfaced
    title: |
      No-anchor reasons are computed but dropped; compile missing-substrate errors never name why the fallback did not apply
    detail: |
      acquire.go:320-333 discards every reason except FallbackAbsentError (bare primary, Git failure, unavailable primary), contradicting plan Decision 5, the ticked issue Task 4 'anchor reason in missing errors', and Done-when bullet 4; a sourced row then silently clones a duplicate beside the worktree. DeclaredSubstrate.Via/Unanchored have zero production consumers. Wrap the missing-substrate errors with the reason and test via Client.Anchors, or revise the plan.
  - id: new
    severity: Important
    family: readme-tracks-user-surface
    title: |
      README not updated for linked-worktree resolution; its 'record a source to restore a missing checkout' is now false in linked worktrees
    detail: |
      README.md:103-110 covers primary and Couch-slot behavior only; the new fallback, its compile notice, and the absent-everywhere refusal are documented only in atlas.
  - id: new
    severity: Minor
    family: fake-seam-unwired
    title: |
      NewResolver defaults every FS (including in-memory fakes) to GitAnchors; no fake implements Anchors despite the plan's hermeticity claim
  - id: new
    severity: Minor
    family: shared-git-reader
    title: |
      workspace.ExecGit duplicates cmd/weave setupGitReader and lacks a context/timeout
  - id: new
    severity: Minor
    family: doc-overclaims-behavior
    title: |
      Atlas says the absent-everywhere refusal happens before compile writes anything; earlier rows may already have been acquired
  - id: new
    severity: Minor
    family: field-name-matches-meaning
    title: |
      DeclaredSubstrate.Unanchored also carries FallbackAbsentError, where an anchor did exist
```

---

## Re-review — 2026-10-13T13:50:30-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 230 — Make Weave dependency resolution worktree-aware |
| repo | ariadne |
| issue file | workshop/issues/000230-worktree-dependency-resolution.md |
| boundary | whole-issue close |
| milestone | — |
| window | 08afcec554b1d5cb4ff03b84f57352274df5db1d..960169b8519fd92ea55a738e7a914649aa0f7770 |
| command | sdlc close --issue 230 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T13:50:30-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Let me see the latest fix commit first (it addresses prior findings), then the full diff.

Verifying the round-1 fixes; next I'll run the test suites for the touched packages and do revert checks.

Both failures reproduce at the base commit (environment-specific to this review checkout), so they are not regressions. Now revert-checking the BR-1 and BR-2 fixes.

Both regression tests go red when the fixes are reverted. Next I'm checking the remaining minors and sweeping for sibling instances.

Nearly done. I'm checking whether `weave refresh` is a fourth consumer of substrate resolution that bypasses the shared resolver.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The round-1 fixes hold up. BR-1 (Lstat-based presence) and BR-2 (typed `NoAnchorError`, plus the reason appended to compile's missing-substrate error) both have regression tests, and I confirmed each test fails when its fix is reverted. The pure chooser, the Git boundary, and the walk/acquire/sdlc wiring are all sound. Two Important items stand between this and SHIP. First, the consumer sweep finds a fourth resolver that the plan never listed: `weave refresh` re-resolves rows beside the worktree and will fail from a linked worktree. Second, the plan file was not updated for the round-1 contract changes and still describes the old behaviour.

Two test failures appear in this checkout: `TestEnvironmentLegacyAndArchive` and `TestCompileWritesNothingInMountlessDependency`. Both fail the same way at the base commit (08afcec5), so they come from this review environment and are not regressions. The tests in `pkg/layergraph`, `cmd/weave/internal/acquire` and the sdlc `Substrate` tests pass.

Done-when: current. Bullet 4 ("Missing dependencies or unavailable fallback anchors produce useful errors without deleting previously generated output") covers both the absent-everywhere refusal from Revisions and the new fail-closed handling of anchor errors. Bullets 1–3 and 5 match what the diff builds.

**1. Strengths**
- `pkg/layergraph/substrate.go:31` `ChooseSubstrate` is the whole selection rule, it is pure, and table tests cover it with probe stubs (ARCH-PURE).
- `pkg/layergraph/anchor.go:79-98`: errors are now split cleanly. An expected absence (`NoAnchorError`) is carried in `Resolution.NoFallback`. A failure to find out is a hard error, so no clone happens on a guess.
- `pkg/workspace/anchor.go:46-63`: a `.git` directory answers "primary checkout" without running Git, and a missing `.git` never lets an enclosing repository lend its primary. Both are pinned by `TestDependencyAnchorSkipsGitOutsideCheckouts`.
- Memoized, lazy anchor lookup: Git runs only when a relative row is absent, at most once per owner per walk (ARCH-CONSTRAINTS).
- The real-Git worktree fixtures in `pkg/layergraph/worktree_test.go` and `cmd/weave/worktree_test.go` exercise the actual boundary, not mocks (ARCH-MOCK).

**2. Critical findings**
None.

**3. Important findings**
- **`cmd/weave/internal/refresh/refresh.go:35-40`, `declaredOrigins`** (ARCH-PURPOSE / ARCH-DRY):
  - **The bug:** it joins `s.path + row.Path` itself and canonicalizes with `EvalSymlinks` (`refresh/git.go:20`). `discoverRefresh` → `Restore` resolves a sourced `../base` from a linked worktree to the primary's sibling. `declaredOrigins` then re-resolves it beside the worktree, gets ENOENT, and `weave refresh` fails.
  - **The spec:** it requires all consumers to share the rule. The plan listed three resolvers; this is a fourth.
  - **Fix sketch:** compare against the `Result.Layers`/edges `Restore` already produced, or resolve through `layergraph.Resolver`. Add a linked-worktree refresh test.
- **Stale plan (`doc-overclaims-behavior`).** This is the 2nd finding in that family. The rule is: a review fix that changes a contract updates every artifact restating it, in the same commit. To enumerate, grep the old identifier and the old semantic phrase across `workshop/plans`, `atlas/`, `README.md` and doc comments. Measured here:
  - plan:57-62: Decision 5 calls "Git failure" a no-anchor value.
  - plan:80: the table names `Unanchored`.
  - plan:96-97: `Resolution` is listed without `NoFallback`.
  - plan:106-108: says the error is "never fatal by itself".
  - plan:273 and plan:341: still say `Unanchored`.

**4. Minor findings**
- BR-5 is only half-done. `GitAnchors` still builds `workspace.ExecGit{}` without a context, so `Restore`, which has a `ctx`, runs uncancellable Git anchor calls (`pkg/layergraph/anchor.go:24`).
- `cmd/sdlc/startplan.go:464-475`: the fail-closed branch has no test and no seam to inject `Anchors`. It also re-derives the local path that `ChooseSubstrate` already computes, and skips the `seen` dedupe.

**5. Test coverage notes**
- I reverted each fix and re-ran its test:
  - BR-1: disabling the Lstat path makes `TestWalkLinkedWorktreeDanglingLocalNotBypassed` fail.
  - BR-2: dropping the reason append makes `TestRestoreMissingNamesNoFallbackReason` fail.
- `TestRestoreAnchorFailureFailsClosed` and `TestWalkAnchorFailureIsAnError` pin the fail-closed behaviour.
- There is no refresh coverage for the linked-worktree case.

**6. Architectural notes**
- ARCH-DRY: flagged. `refresh` re-derives substrate resolution (above). There are four other `GitInDir` readers in `cmd/sdlc`, but they predate this issue and have different stdout semantics, so they are not swept here.
- ARCH-PURE: pass.
- ARCH-PURPOSE: flagged (the refresh consumer).
- ARCH-MOCK: pass. The `Anchors` seam is real and has both FS-implemented and `Client.Anchors` fakes.
- ARCH-CONSTRAINTS: pass, apart from the context gap in BR-5.
- ARCH-SECURE: pass. It reuses the bounded declaration reader and the existing worktree-list parsing.
- ARCH-ORDER: pass. It holds no state between events except a per-walk memo that dies with the `Resolver`.
- ARCH-FUNERAL: pass. It creates nothing durable, and the refusal prevents duplicate clones; already-acquired dependencies staying is documented.
- One behaviour change applies to all layers: a dangling local symlink is now a `NotLayerError`, not a silent present-skip. That matches spec 5 and the atlas, but downstream repos will see it.

**7. Plan revision recommendations**
- Add `## Revisions` to the plan: "2026-10-13 — close review: `Unanchored`→`NoFallback`; `Resolution.NoFallback` added. `Resolver.Substrate`'s error is now a failure to find the anchor and is fatal to walk/Restore; sdlc reports the row absent. Expected absences are `workspace.NoAnchorError` (incl. `ErrBarePrimary`). Probe uses Lstat. Decision 5's 'Git failure' moves out of the no-anchor reasons."
- Add a row or task for the `refresh` consumer.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Resolver.probe uses Lstat (anchor.go:61); walk uses res.Probe; TestWalkLinkedWorktreeDanglingLocalNotBypassed goes red with the Lstat path disabled.
  - id: BR-2
    disposition: addressed
    note: |
      NoAnchorError typed; Restore appends "no worktree fallback: <reason>"; TestRestoreMissingNamesNoFallbackReason goes red when the append is reverted; anchor failures fail closed (tested).
  - id: BR-3
    disposition: addressed
    note: |
      README.md:112-119 describes the linked-worktree fallback, compile notice and the absent-everywhere refusal; matches acquire.go:322-331 behavior.
  - id: BR-4
    disposition: addressed
    note: |
      TestWalkUsesTheFSAnchors wires an FS implementing Anchors with no Git; in-memory fakes without it hit Lstat-not-exist so no Git runs.
  - id: BR-5
    disposition: not-addressed
    note: |
      Duplicate removed and ExecGit gained Context, but GitAnchors (anchor.go:24) still builds ExecGit{} without one, so Restore's anchor Git calls ignore its ctx.
  - id: BR-6
    disposition: addressed
    note: |
      setup-and-replication.md now says composition writes nothing and already-acquired dependencies stay, matching Restore's BFS order.
  - id: BR-7
    disposition: addressed
    note: |
      Renamed to NoFallback with doc covering both NoAnchorError and FallbackAbsentError (walk.go:47-49).
findings:
  - id: new
    severity: Important
    family: single-resolver-all-consumers
    title: |
      weave refresh declaredOrigins re-resolves substrate rows beside the worktree, bypassing layergraph.Resolver; fails from a linked worktree
    detail: |
      refresh.go:35-40 joins s.path+row.Path and EvalSymlinks it, while snapshots come from Restore's fallback-resolved Layers; a sourced ../base absent beside the worktree errors with ENOENT. Use Restore's resolved edges or the Resolver, and add a linked-worktree refresh test.
  - id: new
    severity: Important
    family: doc-overclaims-behavior
    title: |
      Plan still states the pre-fix contract: Git failure as a no-anchor value, "never fatal", Unanchored, Resolution without NoFallback
    detail: |
      2nd finding in this family. Rule: a review fix that changes a contract updates every artifact restating it in the same commit; enumerate by grepping the old identifier and semantic phrase across workshop/plans, atlas, README and doc comments. Measured stale sites: plan:57-62, 80, 96-97, 106-108, 273, 341. Add a plan Revisions entry.
  - id: new
    severity: Minor
    family: behavior-branch-needs-regression
    title: |
      sdlc substrateWalk's fail-closed anchor-error branch is untested, has no Anchors seam, and re-derives the local path
    detail: |
      startplan.go:464-475 builds NewResolver(OSFS) inline, so the branch is unreachable from tests; it also skips the seen-dedupe.
```

---

## Re-review — 2026-10-13T14:51:32-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 230 — Make Weave dependency resolution worktree-aware |
| repo | ariadne |
| issue file | workshop/issues/000230-worktree-dependency-resolution.md |
| boundary | whole-issue close |
| milestone | — |
| window | c40e1b7f5ff15d1d81f5311815e8123101104046..69aa17f639a9ef681c2ddf8396cdbc3d9eadb6ea |
| command | sdlc close --issue 230 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T14:51:32-07:00 |
| verdict | SHIP |

## Review

Disposing the four open findings. Next I'll check that the BR-8 regression test fails when the fix is reverted.

Now revert refresh.go's fix to confirm the test goes red.

BR-8's test goes red without the fix, so it is a real regression test. Next: the sdlc walk (BR-10) and the stale plan sites (BR-9).

BR-9's stale plan sites are marked superseded and a Revisions entry exists. Checking the walk's Git-failure path and running the affected packages.

Four `cmd/weave` tests fail with "not a git repository". Checking whether this diff causes them or the test environment does.

The same four tests also fail at the base commit, so they come from the review environment, not this diff. Removing the temporary worktree.

```verdict
verdict: SHIP
confidence: high
```

All four open findings are addressed and checked against the code, not the commit messages. BR-8 is fixed: weave refresh's `declaredOrigins` now resolves rows through `layergraph.Resolver`, the same resolver discovery uses. Its new linked-worktree test fails when the fix is reverted (ENOENT on `worktree/host/dep`) and passes with it. BR-10 is fixed: sdlc's walk has a `substrateWalkWith` test seam, and a test drives the branch where Git cannot answer, including de-duplication of the absent row. BR-9 is fixed: every stale plan site carries a *Superseded* marker and points to a 2026-10-13 Revisions entry that states the new contract accurately. BR-5 is fixed: `setupGitReader` is gone, `workspace.ExecGit` replaces it and takes a context. The only new finding is Minor: two places that list the resolver's consumers still omit refresh.

Done-when: current — the five bullets cover what the diff builds: the real-worktree fixture, local-wins-even-when-invalid, absolute/symlink/transitive/spaces cases, useful errors for a missing or unavailable anchor (the revised "Git failure is fatal" contract makes this stronger), and weave and vocabulary resolving the same graph.

1. **Strengths**
   - `ChooseSubstrate` (`pkg/layergraph/substrate.go`) is a pure selection rule tested with probe stubs. `Resolver` keeps the IO behind its `FS` and `Anchors` fields, and memoizes the anchor per owner.
   - The no-anchor type is clean (`pkg/workspace/anchor.go:38`). An expected absence is a `*NoAnchorError`; any other error means Git failed, and `Resolver.Substrate` never confuses the two (`pkg/layergraph/anchor.go:79-84`).
   - The regression tests are real. BR-8's test goes red on revert, and `TestSubstrateWalkAnchorFailureReportsAbsent` pins both the local spelling and the de-duplication.
   - The Revisions entry lists every superseded site, and it explains why `migrateSourceless` and fleet membership correctly keep the plain owner-relative join (`cmd/sdlc/internal/fleet/membership.go:64`).

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - The `Resolver` doc comment (`pkg/layergraph/anchor.go:28-30`) and `atlas/workflow/weave.md:158-159` list its consumers as the walk, acquisition and sdlc's walk. Refresh's `declaredOrigins` (added for BR-8) is missing from both lists. Details in the findings block.
   - `GitAnchors` calls `workspace.ExecGit{}` with no context (`pkg/layergraph/anchor.go:24`), so anchor lookups on the compile path have no timeout. Leftover from BR-5; note only.

5. **Test coverage notes**
   - These pass: `pkg/layergraph`, `pkg/workspace`, all `cmd/weave/internal/...` packages including refresh and acquire, `cmd/sdlc/internal/fleet`, and `cmd/sdlc -run TestSubstrateWalk`.
   - Four `cmd/weave` tests fail with "not a git repository": `TestCompileRefusesDataMountAtGeneratedPath`, `TestCompileRefusesDataMountAboveGeneratedOutput`, `TestCompileMovesPathFromDataMountToGeneratedOutput` and `TestCompileMovesDataMountToGeneratedParent`. They fail the same way at the base commit `c40e1b7f` in a separate worktree, so the cause is this review checkout's `TMPDIR`, not this diff.

6. **Architectural notes** (one line per principle)
   - **ARCH-DRY:** pass. One resolver handles every row reader; the git reader is consolidated.
   - **ARCH-PURE:** pass. The chooser is pure; anchors are injected.
   - **ARCH-PURPOSE:** pass. All consumers now go through the resolver, refresh included.
   - **ARCH-MOCK:** pass. Fake `Anchors` implementations are used in the layergraph, acquire and sdlc tests, and the anchor tests also run against real Git.
   - **ARCH-CONSTRAINTS:** pass, apart from the unbounded `GitAnchors` call noted above.
   - **ARCH-SECURE:** pass. Presence is checked with `Lstat`, and an invalid local target is reported rather than bypassed.
   - **ARCH-ORDER:** passes on the basis that the resolver holds only a per-walk memo and no state between events.
   - **ARCH-FUNERAL:** passes on the basis that nothing durable is created: the change refuses to clone a duplicate beside the worktree.

7. **Plan revision recommendations:** none beyond the doc line in the Minor finding.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      setupGitReader removed; workspace.ExecGit{Context} used in environment.go:21 and main.go:636. Residual: GitAnchors still uses ExecGit{} with no context (anchor.go:24), note only.
  - id: BR-8
    disposition: addressed
    note: |
      refresh.go:29-47 resolves via layergraph.Resolver with c.Anchors; TestRefreshLinkedWorktreeSubstrate fails with ENOENT on worktree/host/dep when the fix is reverted, passes at head.
  - id: BR-9
    disposition: addressed
    note: |
      Plan sites 57-62, 80, 97, 106-114, 273, 341 are marked Superseded and point to Revisions 2026-10-13, which states the typed no-anchor, the fatal Git failure and the NoFallback contract.
  - id: BR-10
    disposition: addressed
    note: |
      substrateWalkWith seam (startplan.go:446); TestSubstrateWalkAnchorFailureReportsAbsent covers the Git-failure branch and absent dedupe; the path comes from res.Path.
findings:
  - id: new
    severity: Minor
    family: doc-overclaims-behavior
    title: |
      Resolver doc comment and atlas weave.md list its consumers without refresh's declaredOrigins
    detail: |
      This is the 3rd finding in family doc-overclaims-behavior. Rule: any prose that lists a component's consumers or contract must be regenerated from a grep of the component's call sites whenever a fix adds or removes one. Here, grep NewResolver/Resolver.Substrate callers and update both lists in pkg/layergraph/anchor.go:28-30 and atlas/workflow/weave.md:158-159, which still omit cmd/weave/internal/refresh after BR-8.
```
