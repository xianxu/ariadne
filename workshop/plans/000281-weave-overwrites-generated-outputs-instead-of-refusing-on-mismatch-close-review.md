# Boundary Review — ariadne#281 (whole-issue close)

| field | value |
|-------|-------|
| issue | 281 — weave overwrites generated outputs instead of refusing on mismatch |
| repo | ariadne |
| issue file | workshop/issues/000281-weave-overwrites-generated-outputs-instead-of-refusing-on-mismatch.md |
| boundary | whole-issue close |
| milestone | — |
| window | d9e3f106e22675735339cd4c9c4b01fcd90fcd70..7464c648210a10002921812721c5aa0dd0ba76fd |
| command | sdlc close --issue 281 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T17:26:30-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Diff read; now running the targeted weave tests to confirm the claimed behavior.

Head-checkout tests fail in my scratch worktree; checking whether that's environmental.

Still environment-related in the scratch worktree; running the new tests in the main checkout instead.

Mutations were caught. Checking one more edge: how the abandoned-output probe handles stat errors compared with `Matches`.

Tests and mutation checks are done. Cleaning up my scratch worktree and temp directories.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The change does what the issue asked. A recorded output that has changed is now moved under `construct/generated/weave/displaced/<path>` with a rename and then rewritten, instead of being refused. A directory at a recorded path, and an unowned destination for a data mount or staged output, still refuse, and the error now says to move it aside. A retired output that has changed is left in place, reported once, and dropped from the inventory. Vendored outputs that git tracks get the "commit it" hint. `atlas/workflow/weave.md` describes the new rule, and "preserving authored replacement" no longer appears in the code.

All plan items are delivered. The plan tests and every `./cmd/weave/...` and `./pkg/weaveownership/...` package passed except the top-level `cmd/weave` package, which fails in the review sandbox's default temp directory. That failure also happens at the base commit: the sandbox's git settings and a block on `git init` outside the checkout break the fixtures. With TMPDIR moved inside the checkout, the targeted `cmd/weave` compile tests all pass, including both new compile tests. I did not re-run the full `cmd/weave` package that way.

I broke the code on purpose twice and the tests caught both times. With the tracked-output recording removed, `TestCompileConvergesRevertedTrackedOutput` failed. With the abandoned reporting removed, the plan package's abandoned test failed.

One gap blocks SHIP: if a compile moves a file to `displaced/` and then fails, the operator is never told the file moved. Fixing it is cheap.

1. **Strengths**
   - Moving the old file with a rename keeps its bytes, mode and link target exactly. The tests check all three: `publication_test.go` checks the mode, `staged_test.go` checks the link target.
   - `ApplyManaged` returns one typed `ManagedReport`, and one `printManagedReport` serves both the data pass and the artifacts pass (`main.go:593`), so the warning logic exists only once.
   - The tracked-output check runs `git ls-files -z` once over the whole index with no pathspec, so there is no globbing and no argument-length limit. If git fails, compile prints a warning and keeps going.
   - The parley case is covered end to end (`main_test.go:255`), including a check that an untracked output does not get the commit hint.
   - `EditHome` is pure and has a test for each family of output.

2. **Critical**
   - None.

3. **Important**
   - **A displacement can go unannounced** (`ownership.go:244-258`, `main.go:520,571`). Files are moved to `displaced/` before `Apply` runs. If `Apply` or a later step fails, `ApplyManaged` returns a report with `Displaced` filled in, but both callers return the error without printing it. On the next compile the destination is simply missing, so it is written fresh with no warning. The operator's edit has disappeared from its path, and no message ever names where it went. That breaks the spec's "never lose bytes silently" rule (ARCH-ORDER: the error path drops the in-flight effect). **Fix:** print the report before returning the error, or add the displaced paths to the error text. Add a test with an `Apply` that fails after displacement.
   - **Data-pass warnings don't say which layer** (`main.go:600`). The data pass runs `ApplyManaged` once per owner layer, but the warning prints `construct/generated/weave/displaced/<path>` relative to that owner without naming it. When the owner is a base layer, the operator will look in the leaf and find nothing. **Fix:** prefix the displaced path with `root` (or print the owner) when it isn't the leaf root.

4. **Minor**
   - In a root that isn't a git repo, every compile that rewrites an output prints "cannot tell whether regenerated outputs are tracked". `Regenerated` includes ordinary content updates, not only displacements.
   - Displaced paths can collide (ARCH-DRY / robustness). Displacing `a` runs `RemoveAll` on `displaced/a`, which also deletes an earlier displaced copy of `a/b`. Displacing `a/b` when `displaced/a` is a file makes `MkdirAll` fail, and compile errors after some files have already been moved.
   - `DisplacedRoot` is defined in `plan`, while the reserved-path constants (`RootRel`) live in `pkg/weaveownership`. It would sit better next to them as `RootRel + "/displaced"` (ARCH-DRY).
   - Retention (ARCH-FUNERAL): a displaced copy of a path weave no longer generates is never removed. The design's bound is every path ever generated, and only the operator can clean up.
   - `printManagedReport` reads `git ls-files` through `CombinedOutput`, so any stderr text from git is mixed into the path list. It's harmless, because those lines only become map keys that never match a path.

5. **Test coverage**
   - The tests check real behavior: actual file contents, modes and link targets after compile, not mocks.
   - Missing: a test where compile fails after a file has been displaced (see Important), and a test of a displacement in the data scope.
   - Pre-existing and outside this change: the `cmd/weave` compile fixtures rely on git finding a repository in a parent directory. They fail under `GIT_CEILING_DIRECTORIES`, and they fail the same way at the base commit.

6. **Architecture**
   - **ARCH-DRY: flagged as Minor** (`DisplacedRoot` placement). Pass otherwise; the report and print helper are shared.
   - **ARCH-PURE: pass.** `EditHome` is pure. The decision logic inside `ApplyManaged` is still mixed with IO, but no worse than before.
   - **ARCH-PURPOSE: pass.** Every case in the spec is delivered.
   - **ARCH-MOCK: pass.** `git ls-files` goes through the existing `setupGitReader` seam, and the tests use a real git repo.
   - **ARCH-CONSTRAINTS: pass.** `ls-files` runs only when some output was rewritten.
   - **ARCH-SECURE: pass.** `displaced/` sits under the reserved, git-ignored `construct/generated/weave/`, and parent paths are checked.
   - **ARCH-ORDER: flagged as Important** (the unannounced displacement on the error path).
   - **ARCH-FUNERAL: Minor.** "Latest copy per path" is enforced, but copies of retired paths are never swept.

7. **Plan revisions**
   - Add a Revisions entry saying that `ApplyManaged`'s report is printed, or attached to the error, when compile fails partway. Today the plan implies warnings are printed only on success.

```findings
findings:
  - id: new
    severity: Important
    family: error-path-drops-effect-report
    title: |
      Displacements done before a failed Apply are never reported; the edit moves to displaced/ silently
    detail: |
      ownership.go renames changed outputs into DisplacedRoot before Apply. On any later error, both compilePrepared callers return without printing the report. The next compile finds the destination absent and writes it with no warning. Print the report, or put the displaced paths in the error, and add a test with an Apply that fails after displacement.
  - id: new
    severity: Important
    family: warning-names-ambiguous-location
    title: |
      Data-pass displacement warning gives an owner-relative path without naming the owner layer
    detail: |
      printManagedReport prints filepath.Join(DisplacedRoot, path) relative to root. The data pass runs per owner layer, so a displaced base-layer mount points the operator at the leaf's displaced/ dir. Prefix the path with root when it is not the leaf.
  - id: new
    severity: Minor
    family: warning-noise
    title: |
      Non-git roots print a "cannot tell whether regenerated outputs are tracked" warning on every compile that rewrites an output
  - id: new
    severity: Minor
    family: displaced-path-collision
    title: |
      Displaced copies of a path and its parent (a vs a/b) collide: RemoveAll deletes the sibling copy, or MkdirAll fails mid-displacement
  - id: new
    severity: Minor
    family: reserved-path-single-source
    title: |
      DisplacedRoot is defined in plan rather than next to RootRel/InventoryPath in pkg/weaveownership
  - id: new
    severity: Minor
    family: artifact-removal-path
    title: |
      Displaced copies of paths weave no longer generates are never swept; only the operator can remove them
```

---

## Re-review — 2026-10-12T17:57:00-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 281 — weave overwrites generated outputs instead of refusing on mismatch |
| repo | ariadne |
| issue file | workshop/issues/000281-weave-overwrites-generated-outputs-instead-of-refusing-on-mismatch.md |
| boundary | whole-issue close |
| milestone | — |
| window | d9e3f106e22675735339cd4c9c4b01fcd90fcd70..05c14052d6ea17d6aa20274086ca67dfc0ffea40 |
| command | sdlc close --issue 281 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T17:57:00-07:00 |
| verdict | SHIP |

## Review

Verifying the round-1 fixes: reading the caller code, then running the tests and reverting each fix to confirm its test fails.

Plan-package and runnable compile tests pass; one git-init test is blocked by the review sandbox. Now reverting each fix to confirm its test goes red.

```verdict
verdict: SHIP
confidence: high
```

All six round-1 findings are fixed, and each fix has a test that fails without it. I reverted each fix in this checkout and ran its test: unwrapping `fail` turns `TestManagedFailureAfterDisplacementNamesCopy` red, a no-op sweep turns `TestManagedSweepsDisplacedCopyWithItsOutput` red, and removing the escaping turns `TestDisplacedPathIsFlat` red. Forcing leaf-relative names, or removing the non-git guard, turns `TestPrintManagedReportNamesOwnerAndSkipsNonRepo` red. The `cmd/weave/internal/plan` and `pkg/weaveownership` suites pass, and so do the runnable compile-level tests (`TestCompileConvergesEditedGeneratorDestination`, `TestCompileRetiresDisconnectedOutputsPreservesAuthoredFiles`, `TestPrintManagedReport…`).

The review sandbox stopped two kinds of test from running. Tests that call `git init` are refused outside the reviewer's directories, and tests whose temp directories sit inside the checkout fail because `GIT_CEILING_DIRECTORIES` blocks git there. So `TestCompileConvergesRevertedTrackedOutput` and the full `make test` were not re-run here. The issue log records a green full run. I found nothing new that blocks.

Done-when: current. The bullets still describe what ships: convergence with a warning, a kept displaced copy, exit 0, retirement only of matching files, unowned-destination refusal with a next action, the commit hint, and the atlas. The Revisions entry changes how copies are laid out and how long they live, and adds the copy paths to errors. Those changes refine the "keeps a displaced copy" bullet and have their own tests, so no Done-when bullet is wrong.

### 1. Strengths
- **One error wrapper covers every failure after a displacement.** The `fail` closure in `cmd/weave/internal/plan/ownership.go:250-259` wraps every later error with the copy paths, inside the function itself. No caller can drop them on its error path. The rule is also written down in `workshop/lessons.md`.
- **Flat, collision-free copy names.** `weaveownership.DisplacedPath` flattens each output path with `url.PathEscape`. That escapes both `/` and `%`, so each output maps to exactly one entry and `a` and `a/b` can no longer nest or collide. It sits beside `RootRel`/`InventoryPath` as the single source for the location (`pkg/weaveownership/ownership.go:47-53`).
- **Copies have a defined end.** `sweepDisplaced` keys its keep-set on `kept`, which includes the other scope's identities. A data pass therefore never sweeps the artifacts pass's copies. Leftover entries from the earlier mirror-tree layout (never released) are removed automatically.
- **Data-pass messages name the owner.** `printManagedReport` gives absolute paths only when the owner is not the leaf (`cmd/weave/main.go:597-602`).

### 2. Critical findings
None.

### 3. Important findings
None.

### 4. Minor findings
- The non-git check in `cmd/weave/main.go:617` matches git's English text ("not a git repository"). Git translates that message, so under a non-English locale the "cannot tell…" warning comes back. Using exit code 128 together with `rev-parse --is-inside-work-tree`, or setting `LC_ALL=C` on the call, would remove the locale dependence. This is recorded under BR-3, not raised as a new finding.

### 5. Test coverage notes
- Every round-1 fix has a regression test that fails without the fix (the mutation checks above).
- `TestManagedDataScopeDisplacesChangedMount` closes the data-scope displacement gap.
- `TestCompileConvergesRevertedTrackedOutput` (the parley case, with the commit hint) could not run in this sandbox. The issue log records it passing.

### 6. Architectural notes
- **ARCH-DRY: pass.** The `plan.DisplacedPath` wrapper delegates to `weaveownership`, the same pattern as `digest` and `readInventory`.
- **ARCH-PURE: pass.** `EditHome` and `DisplacedPath` are pure. Displacement and the sweep stay in the existing IO function, behind the `weavefs.FS` interface.
- **ARCH-PURPOSE: pass.** The round-1 fixes cover each class, not just the reported instance: every error return after a displacement is wrapped, every path in the data-pass report is named, and every copy is swept.
- **ARCH-MOCK: pass.** The tests drive `managedFailFS` over the real filesystem. The git call goes through the existing `setupGitReader` interface.
- **ARCH-CONSTRAINTS: pass.** The sweep does one `ReadDir` per managed apply, and the number of copies is bounded by the live outputs.
- **ARCH-SECURE: pass, with the locale caveat above.** Escaping makes copy names safe from path traversal (`..` stays inside, `/` is escaped).
- **ARCH-ORDER: pass, with one note.** Error paths now carry the effect. If the process is killed between the rename and `Apply`, there is no error to carry it: the next compile writes the output silently. The copy does survive, because the `prepared` inventory keeps the identity live. If that window ever matters, report unreported copies from the state on disk.
- **ARCH-FUNERAL: pass.** A copy is created when an output is displaced, replaced by the next displacement of the same path, and swept when its output leaves the inventory.

### 7. Plan revision recommendations
None. The Revisions entry matches the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      ApplyManaged's fail() wraps every post-displacement error with the copy paths; TestManagedFailureAfterDisplacementNamesCopy goes red when fail() is unwrapped (verified).
  - id: BR-2
    disposition: addressed
    note: |
      printManagedReport takes leaf and owner and names ancestor-owner paths absolutely; TestPrintManagedReportNamesOwnerAndSkipsNonRepo goes red with leaf-relative names (verified).
  - id: BR-3
    disposition: addressed
    note: |
      Silent outside a repository (mutation-verified). The guard matches git's English stderr, so a non-English locale still warns; consider exit code or LC_ALL=C.
  - id: BR-4
    disposition: addressed
    note: |
      Flat url.PathEscape names under DisplacedRel are injective; TestDisplacedPathIsFlat goes red without the escaping (verified).
  - id: BR-5
    disposition: addressed
    note: |
      DisplacedRel and DisplacedPath now live in pkg/weaveownership beside RootRel and InventoryPath; plan only delegates.
  - id: BR-6
    disposition: addressed
    note: |
      sweepDisplaced removes copies whose outputs left the inventory; TestManagedSweepsDisplacedCopyWithItsOutput goes red when the sweep is a no-op (verified).
```
