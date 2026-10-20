# Boundary Review — ariadne#486 (whole-issue close)

| field | value |
|-------|-------|
| issue | 486 — Landing train checkout lacks the workflow provider (Makefile.workflow symlink): pair's first train abandoned; instance gate doesn't name the file |
| repo | ariadne |
| issue file | workshop/issues/000486-landing-train-checkout-lacks-the-workflow-provider-makefile-workflow-symlink-pair-s-first-train-abandoned-instance-gate-doesn-t-name-the-file.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1cc7620f848ce10c0834d213257b8ea5c47d214d..4a966a8883623a8eefd5582498139a7dbcaf13b8 |
| command | sdlc close --issue 486 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-20T01:21:16-07:00 |
| verdict | REWORK |

## Review

Reviewing the #486 diff. I've confirmed that pair gitignores `Makefile.workflow`, so the file is missing from the train tree. Now checking whether the fix covers that case.

```verdict
verdict: REWORK
confidence: high
```

The fix doesn't cover pair's actual failure. It only handles a tree that contains a dangling relative `Makefile.workflow` symlink. In pair that symlink is a generated, gitignored file: `.gitignore:78` has `/Makefile.workflow`, `git ls-files` doesn't list it, and the slot has it as an untracked link to `../ariadne/Makefile.workflow`. The TL log in the issue says the same. So `git worktree add` produces a train tree with no `Makefile.workflow` at all. In `workflowFile`, `os.Readlink(p)` then fails with ENOENT and the function returns the original error. I confirmed this with a throwaway probe test in this checkout (since removed): `workflowFile(<tree without the file>, <slot with a working symlink>)` returns `lstat …/Makefile.workflow: no such file or directory`, which is pair's exact error. The new e2e test passes only because its fixture commits the symlink into git (`landingtrain_test.go:820-824`), which no ariadne-layered repo does. The three new tests pass (`go test -run '…' ok 11.8s`). The instance-gate change and the environment-stop pin look right.

**Strengths**
- There is one resolver, `workflowFile` (`helpercapability.go:67`), called through `prepareSuiteHelper` for both local-check and the train, rather than a second copy of the logic (ARCH-DRY pass).
- The instance-gate error flattens the validator's multi-line output into the error's single line (`validategate.go:132-139`). Its test checks this through `firstLine`, which is all the train records.
- `TestLandingTrainStoppedByEnvironmentReleasesItsCars` pins the release, no-hold, closed-PR and serial-landing behaviour, so the Spec's "check what happens today and pin it" clause is met.

**Critical**
- `cmd/sdlc/helpercapability.go:67-78`: the anchor fallback only fires when the tree has a dangling relative link. When the file is absent, it should resolve `filepath.Join(anchor, "Makefile.workflow")` directly, which is how local-check resolves it from the slot. Fix sketch: `if err != nil && anchor != "" { return filepath.EvalSymlinks(filepath.Join(anchor, "Makefile.workflow")) }`. The fixture should leave `Makefile.workflow` untracked: gitignore it on main and create the symlink only in each car checkout. That is pair's real shape (ARCH-PURPOSE).

**Important**
- None beyond the test-fixture gap folded into the Critical finding.

**Minor**
- `landingtrain.go:229-232`: the comment still says the helper resolves "from the train's tree". It now falls back to the anchor checkout.
- Plan claims "It fails without the fix with pair's exact error". Pair's exact error comes from a missing file, not a dangling link, so that claim is wrong.

**Test coverage**
- The e2e needs a variant where the file is untracked or absent. The current fixture tests a repo shape that doesn't exist in the fleet.
- The Done-when "rerun pair's first train live" item is unchecked and deferred until after landing. That is acceptable, but with the current code that rerun would fail again.

**Architecture**
- ARCH-DRY: pass.
- ARCH-PURE: pass. `workflowFile` is a small IO seam and the decision logic is trivial.
- ARCH-PURPOSE: flagged. The fix covers a dangling tracked link, not the absent generated provider that blocks pair.
- ARCH-EVOLVE: pass. No persisted format changes.

**Docs**
- No new user-facing surface, so no atlas or README update is needed.

**Plan revisions**
- Add a `## Revisions` entry: "The train tree lacks Makefile.workflow entirely (it is gitignored and generated in consumers). `workflowFile` falls back to `anchor/Makefile.workflow` when the tree's own copy does not resolve, whether it is missing or dangling. The fixture leaves the link untracked."

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: fixture-diverges-from-real-repo-shape
    title: |
      workflowFile's anchor fallback misses pair's case: Makefile.workflow is gitignored and absent from the train tree
    detail: |
      helpercapability.go:72-76 falls back only through os.Readlink of the tree's own link. In pair (and every consumer that slot-dependencies provisions), Makefile.workflow is untracked (.gitignore:78), so the train worktree has no such file, Readlink gets ENOENT, and the function returns pair's exact error "lstat .../Makefile.workflow: no such file or directory". A probe test confirmed it: workflowFile(treeWithoutFile, slotWithSymlink) returns that error. Fix: when root's copy fails to resolve and anchor != "", return filepath.EvalSymlinks(filepath.Join(anchor, "Makefile.workflow")). Every instance of the family in this window: symlinkedWorkflowTrain (landingtrain_test.go:820-824) commits the symlink to git, so TestLandingTrainResolvesASymlinkedWorkflow and TestLandingTrainStoppedByEnvironmentReleasesItsCars both run on a shape no layered repo has. Make the fixture gitignore Makefile.workflow and create the symlink only in each car checkout, then confirm the e2e test fails without the fix.
  - id: new
    severity: Minor
    kind: defect
    family: stale-comment-after-behavior-change
    title: |
      landingtrain.go:229-232 comment still says the lock helper resolves from the train's tree
    detail: |
      With WorkflowAnchor it now falls back to the checkout the train was built from. Update the comment, and the plan's claim that the test "fails without the fix with pair's exact error".
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"aa0528ba-ebf8-4857-8a39-fc201ee6876e","repository":"github.com/xianxu/ariadne","issue_id":"000486","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/425b922304953e34","reviewed_head":"4a966a8883623a8eefd5582498139a7dbcaf13b8","slot":"ariadne:4","requested_at":"2026-10-20T08:19:39.26669Z"},"phase":"finished","pid":74671,"started_at":"2026-10-20T08:19:39.35695Z","finished_at":"2026-10-20T08:21:15.75727Z","session_id":"432bde0a-908a-4f2c-815a-5b5be639d0a1","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T01:36:54-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 486 — Landing train checkout lacks the workflow provider (Makefile.workflow symlink): pair's first train abandoned; instance gate doesn't name the file |
| repo | ariadne |
| issue file | workshop/issues/000486-landing-train-checkout-lacks-the-workflow-provider-makefile-workflow-symlink-pair-s-first-train-abandoned-instance-gate-doesn-t-name-the-file.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1cc7620f848ce10c0834d213257b8ea5c47d214d..dc5a9f192f259a627a04e183e8197fc405a0edbe |
| command | sdlc close --issue 486 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-20T01:36:54-07:00 |
| verdict | SHIP |

## Review

Running the #486 tests. One unrelated train test failed on a host repo-lock wait, so next I'm rerunning only the targeted tests.

Now reverting the anchor fallback to confirm the e2e test goes red.

Both pieces of the fix go red when reverted. Next I'm checking other callers of `prepareSuiteHelper` and rerunning the one unrelated failure on its own.

```verdict
verdict: SHIP
confidence: high
```

Both prior findings are fixed, and I checked the fix by running the tests, not by reading the commit message. `workflowFile` (`cmd/sdlc/helpercapability.go:69-84`) now falls back to the anchor checkout's own `Makefile.workflow`. That covers pair's real setup, where the file is gitignored and missing from the train's tree. The fixture now has the same setup: main gitignores the file, and each car checkout holds an untracked link to its sibling substrate. I reverted the fix two ways in the disposable checkout. Each time `TestLandingTrainResolvesASymlinkedWorkflow` failed with pair's exact error: "no Makefile.workflow to find the suite lock from (lstat …/sdlc-train-…/Makefile.workflow: no such file or directory)". The two reverts were the anchor fallback lines alone, and the `c.local.WorkflowAnchor = c.t.Root` line alone. `TestWorkflowFileResolvesThroughTheAnchor` also failed when the fallback was removed. With the fix in place, all four #486 tests pass and `go vet` is clean. Nothing blocks SHIP. The live rerun of pair's train is still open in the plan; it is a step for after landing, not something this code range has to deliver.

1. **Strengths**
   - There is one resolver, `workflowFile`, used through `prepareSuiteHelper`. Both local-check (empty anchor) and the train (`landingtrain.go:238`) go through it, so the logic lives in one place (ARCH-DRY: pass).
   - `workflowFile` is a small file-resolution function. Its unit test covers the untracked case, the tracked-link case, the no-anchor case and the checkout itself, using only temp directories, with no mocks.
   - The instance-gate error is now on one line and names each file and rule (`validategate.go:132-139`). That matches the train recording only `firstLine(err)`, and `TestLandingInstanceGate_ErrorNamesEachFileOnOneLine` uses a multi-line validator message to check it.
   - A test now pins what happens when the environment stops a train: the other car is released (not held), it lands by the serial path, and the train PR is closed.

2. **Critical:** none.
3. **Important:** none.
4. **Minor:** none new.
5. **Test coverage**
   - Both modes in the Done-when are covered. A train that lands with a reachable substrate is tested, and so is one stopped by its environment with no substrate.
   - `TestLandingTrainLandsEveryCarOnOneSuite/null` failed once in a wider run while the host's sdlc repo lock (pid 5654) was held. Run alone it passed in 31s. This diff doesn't touch that test or its setup, so I put the failure down to the host environment.
6. **Architecture**
   - ARCH-DRY: pass. There is one resolver for both callers.
   - ARCH-PURE: pass. The file-system work is confined to a small leaf function.
   - ARCH-PURPOSE: pass. Every Done-when item except the live rerun is delivered, and that rerun is correctly marked as a step after landing.
   - ARCH-EVOLVE: pass. No stored format changes.
7. **Plan revisions:** none. The plan's claim that the test "fails without the fix with pair's exact error" now holds; I confirmed it by reverting the fix.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      helpercapability.go:80 now falls back to the anchor's own Makefile.workflow. The fixture gitignores it and links it untracked in each car. With the fallback or the WorkflowAnchor line reverted, TestLandingTrainResolvesASymlinkedWorkflow fails with pair's exact lstat error, and with the fallback reverted the unit test fails too.
  - id: BR-2
    disposition: addressed
    note: |
      landingtrain.go:229-232 now describes the fallback to the checkout the train was built from. The plan's fails-without-fix claim is now true; confirmed by reverting.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"f22bf1ca-71f6-4fb9-a088-720d69b78afe","repository":"github.com/xianxu/ariadne","issue_id":"000486","recipe":"small-diff-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/e1624071ff38efaf","reviewed_head":"dc5a9f192f259a627a04e183e8197fc405a0edbe","slot":"ariadne:4","requested_at":"2026-10-20T08:29:52.507787Z"},"phase":"finished","pid":85159,"started_at":"2026-10-20T08:29:52.743624Z","finished_at":"2026-10-20T08:36:54.14004Z","session_id":"146fb686-5cc2-4493-9948-20463cf91b58","outcome":"returned"}] -->
