# Boundary Review — ariadne#265 (whole-issue close)

| field | value |
|-------|-------|
| issue | 265 — weave robustness: atomic compile, symlinked Makefile, legacy include, dependency writes |
| repo | ariadne |
| issue file | workshop/issues/000265-weave-robustness.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5655861efb765f10a5e5e9bd514dc3743c23fc56..4e2fbb31019f4a8bebbed683df4cc281c5081d86 |
| command | sdlc close --issue 265 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T18:34:54-07:00 |
| verdict | REWORK |

## Review

Checking whether `prepareManaged` writes anything, then testing a suspected false refusal when a path moves from a data mount to a generated output.

Prepare is write-free, but I confirmed a false refusal: a path that moves from a data mount to a generated output is now refused on every compile. Running the package tests next.

These failures come from the environment: TMPDIR points inside the review checkout. Re-running the new compile tests with a normal TMPDIR.

The harness blocks `git init` outside its own directories, so the compile-level tests can't run here. Last check: does `ProducedPathSet` see generated (staged) outputs in the collision check?

```verdict
verdict: REWORK
confidence: medium
```

**Verdict: REWORK.** Items 3 and 4 are correct, and the prepare/commit split of `ApplyManaged` is clean. But the new check-everything-first order in `compilePrepared` checks the artifacts pass against the state *before* the data passes run. When a path moves from a data mount to a generated output, that check refuses every compile, even though the old order handled the move without error. Nothing inside weave can clear it, because the data pass that would retire the old mount never runs. A related problem: the collision guard only compares exact paths, so a mount over a *parent* directory of a generated output still produces a partial write before the artifacts pass refuses. That breaks the "refused compile writes nothing" promise. Both come from one cause: the artifacts check doesn't account for what the data pass is about to change.

**Inspection limits:** I could not run the compile-level tests in `cmd/weave/startup_test.go` in this checkout. The review harness refuses `git init` outside its own directories, and with TMPDIR inside the checkout, `rev-parse` fails. The exception is `TestCompileWritesNothingInMountlessDependency`, which ran and passes on the head and on the reverted `main.go`, as the plan expects for something #264 already fixed. So I could not confirm by reverting that the item-1 test fails without the fix. The plan-level evidence below comes from a scratch test I ran and then deleted. All `internal/...` packages pass.

1. **Strengths**
   - `cmd/weave/internal/plan/ownership.go:155-272`: `prepareManaged` only reads; every write is in `commitManaged`, and `ApplyManaged` and `CheckManaged` share one inspection routine (ARCH-DRY pass). I checked that `materializeManaged` doesn't write either.
   - `cmd/weave/internal/startup/tools.go:17-27`: switching to `Lstat` with a refusal that names the link is the right small fix, and the test confirms the linked Makefile's recipe never runs.
   - `cmd/weave/internal/plan/apply.go:126-130`: `TrimPrefix("-")` plus a switch covers both `include` and `-include` for both forms, and the test covers indented lines too.
   - The atlas was updated for the new compile order and the symlink refusal.

2. **Critical**
   - **`cmd/weave/main.go:567-585` — false refusal that weave cannot recover from (ARCH-PURPOSE).**
     - `CheckManaged(root, actions, ScopeArtifacts)` reads the inventory before the root data pass runs. Any wanted artifact path still recorded under `data` fails with "output X belongs to data" (`ownership.go:212-214`).
     - In the old order, the data pass retired that entry first, so the artifacts pass succeeded.
     - **Reproduced at plan level:** after `ApplyManaged(data, [Symlink x.md])`, `CheckManaged(artifacts, [WriteFile x.md])` returns `output x.md belongs to data`. The old sequence, `ApplyManaged(data, nil)` then `ApplyManaged(artifacts, …)`, returns nil.
     - Every re-run fails the same way. Recovering means editing `ownership.json` by hand.
     - **Fix sketch:** prepare the artifacts pass against the root data pass's projected state. For example, give `prepareManaged` an optional projected inventory (the data pass's `other+wanted`) and the data pass's produced paths. That also makes the ad hoc collision loop unnecessary (same family as the next finding).

3. **Important**
   - **`cmd/weave/main.go:574-579` — the collision guard only compares exact paths, so parent/child overlaps still write before refusing.**
     - Example: a data mount at `construct/generated` and a generated output at `construct/generated/x/SKILL.md`. The artifacts check passes, because no symlink parent exists yet.
     - The data apply then commits its symlink, inventory and `.gitignore`, and the artifacts apply fails in `safeParents`. That is the non-atomic refusal this issue set out to remove.
     - **Same family:** the false refusal above. Both are the artifacts check ignoring the data pass's effects, and the projected-state fix covers both.
   - **No test for the transition case.** Add a compile-level test where a data mount is first applied at path X, then the deps drop it and a generator produces X. The compile should succeed.

4. **Minor**
   - `main.go:573`: `ProducedPathSet(actions)` is recomputed inside the loop. It only runs for root, but it could be hoisted.
   - `CheckManaged` skips `ReclaimPublications`, while `ApplyManaged` reclaims before preparing. This is harmless today, but the doc comment says it runs "every refusal ApplyManaged would", which isn't quite true.

5. **Test coverage**
   - Each Done-when clause has a test: atomic refusal, mount-less dependency, symlink refusal, and indirect include.
   - Both states are covered for items 3 and 4. Regular Makefiles are covered by existing tests, and so is the direct include.
   - What's missing is the case where the check order changes the result (the cross-pass transition above).

6. **Architecture**
   - **ARCH-DRY: pass.** One shared prepare routine. The collision loop is a second, partial copy of overlap logic that `prepareManaged`/`safeParents` already have; it should go away once the projected-state fix lands.
   - **ARCH-PURE: pass.** Prepare is a read-only inspection and commit holds all the writes. The orchestration in `main.go` is sensible.
   - **ARCH-PURPOSE: flag.** "Check everything before writing" is only atomic if the check models what earlier passes will change. Right now it models each pass separately, which gives both false refusals and missed refusals.

7. **Plan revisions**
   - Add a `## Revisions` entry: "The artifacts check is prepared against the root data pass's projected inventory and produced paths, not the current state on disk. This replaces the exact-path collision check and covers data→artifact path transitions."
   - The Log's statement that "each pass is inspected twice" should also say that the second inspection is the authoritative one, and that the first must not be stricter than it.

```findings
findings:
  - id: new
    severity: Critical
    family: cross-pass-check-ignores-prior-pass-effects
    title: |
      Up-front artifacts check refuses a data-to-artifact path transition the old order converged (permanent, unrecoverable refusal)
    detail: |
      main.go:581 CheckManaged(root, artifacts) reads the pre-data inventory; a wanted path still recorded under scope data fails "output X belongs to data" (ownership.go:212), though the root data pass would retire it first. Reproduced at plan level: CheckManaged errors, while ApplyManaged(data,nil) then ApplyManaged(artifacts) succeeds. Since the data pass never runs, every re-run refuses. Fix: prepare artifacts against the data pass's projected inventory/produced paths; add a compile-level transition test.
  - id: new
    severity: Important
    family: cross-pass-check-ignores-prior-pass-effects
    title: |
      Data-vs-artifact collision guard is exact-path only; parent/child overlaps still commit the data pass before the artifacts refusal
    detail: |
      main.go:574-579 compares ProducedPathSet keys only. A data mount at construct/generated with a generated construct/generated/x/SKILL.md passes both checks (no symlink parent exists yet), then data apply writes its mount, inventory and .gitignore, and artifacts apply fails in safeParents, so the refusal is not atomic. Family: the same root cause as the transition false refusal; the projected-state prepare covers both and replaces this ad hoc loop (ARCH-DRY).
  - id: new
    severity: Minor
    family: doc-claim-overstates-behavior
    title: |
      CheckManaged doc says it runs every refusal ApplyManaged would, but it skips ReclaimPublications
    detail: |
      ownership.go:131 vs ApplyManaged:144. Harmless today; reword or note the omission.
  - id: new
    severity: Minor
    family: loop-invariant-recompute
    title: |
      ProducedPathSet(actions) recomputed inside the per-owner loop in compilePrepared
    detail: |
      main.go:573; hoist above the loop (moot if the collision loop is replaced).
```

---

## Re-review — 2026-10-12T18:53:56-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 265 — weave robustness: atomic compile, symlinked Makefile, legacy include, dependency writes |
| repo | ariadne |
| issue file | workshop/issues/000265-weave-robustness.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5655861efb765f10a5e5e9bd514dc3743c23fc56..f41a8cafc8f9ef080ee9fa5a996dd192782fb04e |
| command | sdlc close --issue 265 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T18:53:56-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing the BR-1/BR-2 fixes in `prepareManaged`; next I'll run the suite and revert-test the fixes.

Running the compile tests to separate environment failures from real ones.

BR-1 and BR-2 fixes both go red when reverted. Now checking the issue's plan items against the diff.

Writing a throwaway test to check whether the untested "retired" branch is needed for generated (strict) outputs.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

All four Plan items and Done-when criteria are delivered, and both prior blocking findings are fixed with tests that go red when the fix is reverted. Splitting the old apply into a no-write prepare step plus a commit step means the check and the real apply run the same refusals. The new "every pass is checked before any pass writes" ordering keeps a refused compile from touching the checkout. One gap remains, and it is cheap to close. The BR-1 fix has two parts:

1. **Projected inventory:** the earlier pass's data records replace the current ones before the artifacts check runs.
2. **Retired-path skip:** a path the data pass will retire is treated as already empty.

The committed BR-1 test only needs one of the two to pass. With the retired-path skip disabled, every committed test stays green, yet that skip is what lets a path move from a data mount to a *generated* output. I confirmed this with a throwaway test: it passes at head, and with the skip disabled it fails with a permanent "non-directory parent" refusal. That scenario needs a committed test.

**How the tests were run.** The plain run of `go test ./...` failed all 31 `TestCompile*` tests. That is caused by this review environment, not the code: `GIT_CEILING_DIRECTORIES` covers `.sdlc-review/tmp`, so git can't find a repo above the temp dirs. With TMPDIR moved under the checkout and the ceiling relaxed, the new compile tests pass. All internal packages pass as-is. I made each revert check in the disposable checkout and restored it every time; `git status` is clean.

1. **Strengths**
   - `ownership.go:184-195`: `ApplyManaged` is now "reclaim stages → prepare → commit", and `CheckManaged` uses the same prepare step. There is one inspection routine instead of a second copy (ARCH-DRY).
   - `main.go:567-594`: every pass is checked before any pass writes, and the real apply still re-runs prepare against the actual state. If the check's prediction is ever wrong, the worst outcome is a refusal partway through, never a wrong write.
   - `ownership.go:263-266`: the overlap check (at, above or below an earlier pass's output) replaces the old exact-path loop and covers both directions of BR-2.
   - `startup_test.go:149-176` checks exactly what Done-when asks for: `.gitignore` byte-identical, no inventory, no mount.
   - Items 3 and 4 are small and each has a test (`tools.go:17-27`, `apply.go:126-129`).

2. **Critical findings:** none.

3. **Important findings**
   - `ownership.go:268-273`: the retired-path skip (`within(id.Path, retired)` … `continue`) has no committed test. `TestCompileMovesPathFromDataMountToGeneratedOutput` uses a manifest symlink, which is not a strict output, and the projected inventory alone is enough for that case.
     - **Failure if the skip regresses:** a leaf with data mounted at `construct/generated/example` that drops the mount and gains a generator writing `construct/generated/example/SKILL.md`. Compile then refuses permanently: "publication path … has non-directory parent". This is the BR-1 failure again, for generated outputs.
     - **Fix:** commit a compile-level test for that move. The throwaway one I wrote is about 25 lines and modelled on `TestCompileRefusalLeavesCheckoutUntouched`.

4. **Minor findings**
   - In `prepareManaged`, `retired` is a `map[string]bool` where `false` means "an earlier pass produced this path", and `within` only matches `true` keys. That works, but it's easy to misread; two separate sets would be clearer.
   - ARCH-DRY: `prepareManaged`'s `after` loop restates `commitManaged`'s retirement rule (a matching previous output that is not produced gets removed). If the commit rule changes, the projection can drift from it. A shared helper or a comment cross-reference would keep them tied together.

5. **Test coverage notes**
   - BR-1: reverting the projected inventory and the retired-path skip together turns the transition test red ("belongs to data").
   - BR-2: removing the overlap check turns both compile collision tests red. Removing only the descendant scan turns `TestCheckManagedSeesEarlierPassOutputs` red.
   - The issue Log says "both review findings reproduce as test failures with the projection disabled". That holds for the projection as a whole, but not for each of its parts (see the Important finding).

6. **Architectural notes**
   - **ARCH-DRY:** pass, apart from the Minor note above.
   - **ARCH-PURE:** pass. Prepare and commit go through the injected `weavefs.FS`, as before.
   - **ARCH-PURPOSE:** pass. All four items are delivered.
   - **ARCH-MOCK:** pass. Tests use temp dirs and local git repos, with no network.
   - **ARCH-CONSTRAINTS:** pass. Inspecting each pass twice is acknowledged in the Log for a batch step of tens of outputs.
   - **ARCH-SECURE:** pass. The symlinked-Makefile refusal narrows what gets trusted.
   - **ARCH-ORDER:** pass. Ordering is now explicit (check every pass → apply data → apply artifacts), and the re-prepare inside apply covers changes made between check and apply.
   - **ARCH-FUNERAL:** pass. Nothing durable is new; the generation stage, now created before the check, is still removed by the existing `defer`.
   - A carry-forward outside this diff: moving a path the other way, from an artifact output to a data mount, still fails, because the data pass refuses with "belongs to artifacts" and the artifacts pass never runs. That already happened with the old order. A future issue could give the data pass the same projection, run in reverse.

7. **Plan revision recommendations**
   - Add a "## Revisions" entry (or Log line) saying that the transition test covers the projected inventory, and that a second test pins the retired-path skip for a mount → generated-output move.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Projected inventory plus retired-path skip; reverting both turns TestCompileMovesPathFromDataMountToGeneratedOutput red. The skip alone is not pinned (new finding).
  - id: BR-2
    disposition: addressed
    note: |
      Overlap check at/above/below replaces the exact-path loop; removing it reddens both compile collision tests, removing the descendant scan reddens TestCheckManagedSeesEarlierPassOutputs.
  - id: BR-3
    disposition: addressed
    note: |
      ownership.go CheckManaged doc now says publication stages are reclaimed only by an apply.
  - id: BR-4
    disposition: addressed
    note: |
      Moot: the per-owner loop was replaced by CheckManaged; prepareManaged computes ProducedPathSet once.
findings:
  - id: new
    severity: Important
    family: fix-branch-without-regression-fixture
    title: |
      Retired-path skip in prepareManaged (ownership.go:268-273) is reached by no committed test
    detail: |
      Reverting only the within(id.Path, retired) skip leaves every committed test green, because the BR-1 transition test uses a non-strict manifest symlink that the projected inventory alone handles. A throwaway test (data mount at construct/generated/example, then a generator writing construct/generated/example/SKILL.md) passes at head and fails with the skip disabled ("non-directory parent"), so it is a permanent refusal. Commit that compile-level test.
  - id: new
    severity: Minor
    family: loop-invariant-recompute
    title: |
      Check-time projection restates commitManaged's retirement rule instead of sharing it
    detail: |
      prepareManaged's after loop encodes retire = previous and not produced, the same rule commitManaged applies, so a change to one can drift from the other. Share a helper, or cross-reference the two in a comment. Belongs only loosely to the family (a duplicated derivation, not a recomputation).
```

---

## Re-review — 2026-10-12T19:07:28-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 265 — weave robustness: atomic compile, symlinked Makefile, legacy include, dependency writes |
| repo | ariadne |
| issue file | workshop/issues/000265-weave-robustness.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5655861efb765f10a5e5e9bd514dc3743c23fc56..23728ab3a28a815633f4ed6a27624a8b4f140719 |
| command | sdlc close --issue 265 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T19:07:28-07:00 |
| verdict | SHIP |

## Review

The last commit claims to address BR-5 and BR-6. Let me look at it first, then the full diff.

Now verify BR-5 claim: disable the retired skip and run the new test.

Now revert the skip and confirm the BR-5 test goes red.

```verdict
verdict: SHIP
confidence: high
```

Both open findings are fixed, and I raised nothing new, so this ships. For BR-5, I disabled the retired-path skip in the pinned checkout. The new `TestCompileMovesDataMountToGeneratedParent` then failed with exactly the reported refusal ("has non-directory parent construct/generated/example"), and it passes with the skip restored. For BR-6, a shared `managedPass.retiring()` helper now holds the retirement rule, and both the check projection and the commit use it. All four Done-when items are covered by tests, and the atlas covers the new compile ordering, the symlinked-Makefile refusal and the `$(WF_WORKFLOW)` recognition.

One test, `TestArchiveCompileNeedsNoGitExecutable`, fails in the review sandbox. It fails the same way at the base commit, so this branch didn't cause it. It's an environment problem: the temp directory sits inside the review tree and the sandbox wraps `git`.

1. **Strengths**
   - `ownership.go:156-167`: `retiring()` is now the only definition of "previous and not produced". `prepareManaged` (`ownership.go:226`) and `commitManaged` (`ownership.go:390`) both use it, so the check and the commit can't drift apart.
   - `ownership.go:201-212`: one `prepareManaged` routine does every inspection for both `CheckManaged` and `ApplyManaged`. The checks run before any writes without a second copy of the inspection logic.
   - `main.go:567-582`: every owner's data pass and the leaf's data and artifacts are checked before the first write. Writes then happen in the same order as before, so a refused compile leaves every checkout untouched.
   - `startup/tools.go:17-27`: switching to `Lstat` plus an explicit symlink branch gives a clear, actionable message instead of silently building another repo's tools.
   - The Log (2026-10-12) lists what the data pass changes (records, created paths at/above/below, retired paths) and maps each to a test. That turns the `cross-pass-check-ignores-prior-pass-effects` family into a stated rule, which is also written up in `workshop/lessons.md`.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings:** none filed. One cosmetic point: the atlas paragraph at `atlas/workflow/weave.md:33-41` has one very long line where the text was re-wrapped.

5. **Test coverage notes**
   - Every branch of the projection now has a test that fails if that branch is disabled:
     - records: the BR-1 test
     - a mount at the same path as a generated output: the collision test
     - a mount above a generated output: the BR-2 test
     - a mount below a generated output: `TestCheckManagedSeesEarlierPassOutputs`
     - a retired mount or its parent: the BR-5 test, confirmed red above
   - `go test ./cmd/weave/...`: every package passes except `cmd/weave`, where the only failure is the environment-related `TestArchiveCompileNeedsNoGitExecutable`.
   - Running these tests in the review sandbox needed `GIT_CEILING_DIRECTORIES` unset. Without that, most compile tests fail before reaching the code under test.

6. **Architectural notes**
   - **ARCH-DRY: pass.** The retirement rule and the inspection logic each live in one place.
   - **ARCH-PURE: pass.** The new decision logic sits in `prepareManaged` behind the injected `weavefs.FS`. Helpers like `within` and `overlap` are pure.
   - **ARCH-PURPOSE: pass.** All four problems are delivered, including atomic refusal at compile level, not just inside a single pass.
   - **ARCH-MOCK: pass.** No new external-binary calls. The `make` and `git` call sites are unchanged.
   - **ARCH-CONSTRAINTS: pass.** Each pass is now inspected twice (once to check, once to apply). The Log notes this and argues it's negligible for a batch compile over tens of outputs.
   - **ARCH-SECURE: pass.** It now refuses to follow a symlinked Makefile, which tightens trust in a path from another repo. No secrets are involved.
   - **ARCH-ORDER: pass.** Pass order is explicit: the ordered `after` list in `CheckManaged`. Each later pass's check reads a projection of what the earlier passes leave.
   - **ARCH-FUNERAL: pass.** Nothing new is created that lasts. `managedPass` lives only in memory, and inventory and stage lifetimes are unchanged.
   - For future work: the check and the apply still each re-read the inventory. If more passes join the same root, consider having the apply consume the prepared pass from the check, as long as nothing on disk changed between them.

7. **Plan revision recommendations:** none. The plan and Log match the code.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      TestCompileMovesDataMountToGeneratedParent (cmd/weave/startup_test.go:279) fails with the within(id.Path, retired) skip disabled ("non-directory parent construct/generated/example") and passes at head; verified by revert in the pinned checkout.
  - id: BR-6
    disposition: addressed
    note: |
      managedPass.retiring() (ownership.go:156) is the single retirement rule, used by prepareManaged's projection (ownership.go:226) and commitManaged's retirement loop (ownership.go:390).
```
