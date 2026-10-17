# Boundary Review — ariadne#422 (whole-issue close)

| field | value |
|-------|-------|
| issue | 422 — A downstream repo's committed generated vocabulary goes stale on every ariadne vocabulary change and turns its CI red |
| repo | ariadne |
| issue file | workshop/issues/000422-a-downstream-repo-s-committed-generated-vocabulary-goes-stale-on-every-ariadne-vocabulary-change-and-turns-its-ci-red.md |
| boundary | whole-issue close |
| milestone | — |
| window | 18a4476aa996c36a8df7d647c750e3c130ab2b57..62afe6f15e4672de6f38d1ba55a22c683bdcf24d |
| command | sdlc close --issue 422 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-17T02:33:08-07:00 |
| verdict | SHIP |

## Review

Running the tests and checking whether README documents the manifest verbs.

The new e2e test passes once TMPDIR is outside the confinement; the earlier failure came from the environment (git ceiling dir). Next I'll revert the pair rule to confirm the tests catch its absence.

```verdict
verdict: SHIP
confidence: high
```

The pair rule works and is tested. A `ship` row plus a committed `<path>.weave-sha256` sidecar lets a cold checkout regenerate a shipped output that weave can show it wrote. A hand edit is still refused on a cold checkout and moved aside on a warm one. I checked the regression evidence directly: I replaced `own = own || paired[id.Path]` (`cmd/weave/internal/plan/ownership.go:370`) with a no-op, and the parley row and the pulled-copy row of `TestShippedOutputOwnership` both went red. I then restored the file. `go test` passes for the `plan`, `intent` and `golden` packages. All three cases of `TestCompileRegeneratesShippedOutputCold` pass once TMPDIR is outside the review confinement. Inside it, git cannot run in the temp dir (`GIT_CEILING_DIRECTORIES`), so that failure comes from the environment, not the code. Every checked item in the plan is delivered, and only Minor findings remain.

1. **Strengths**
   - The rule narrows the old one instead of overwriting blindly. `vouched` (`ownership.go:405-436`) Lstats both files and accepts regular files only. A missing, malformed or symlinked file proves nothing, and the result is checked through `safeParents`. Damaged input therefore falls back to today's refusal and never leads to an overwrite (ARCH-SECURE).
   - The pair rule covers both files, which closes plan-quality PQ-1. `paired` holds both the output and its sidecar, so a pulled newer sidecar is not moved aside on a warm checkout. That case is pinned by the "warm, pulled a newer committed pair" row, which asserts `Displaced` is empty.
   - `ShippedPaths` and `ShipStaged` (`shipped.go`) are pure and are tested without touching the filesystem beyond the fixture. All filesystem reads sit in `prepareManaged` behind `weavefs.FS`.
   - The e2e test pins three things: the main.go wiring, today's refusal when no row exists, and the warning for a typo'd row (`startup_test.go:308-367`).
   - The reason for option 2 is recorded: parley ships the file to plugin installs, as found in parley#208. That rules out option 1, and option 3 only reports the breakage without preventing it.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - The matrix doesn't cover a sidecar or output that is a symlink or directory. `vouched` handles it by rejecting non-regular files, but no test enters that branch.
   - The "names no generated output" warning also fires when a `ship` row names a staged *symlink* output, which is misleading. The same goes for an exported ancestor `ship` row reaching a consumer that doesn't generate that path.
   - Plan drift:
     - The Core concepts entry says `ShippedPaths(layers []walk.Layer) []string`, but the code is `ShippedPaths([]layer.Layer) ([]string, error)`.
     - The task checkboxes in the plan file are still `[ ]`, while the issue's checklist is all `[x]`.
   - Done-when says "parley's issue.json is handled by the chosen rule". On the ariadne side this is shown by the live run on parley's real tree. The parley-side declaration and sidecar commit are filed as parley#371 and haven't landed yet, which is fine since it's cross-repo.
   - The e2e test computes sha256 itself instead of sharing the package's `digest`. That's acceptable because the test is in `package main`.

5. **Test coverage notes**
   - `TestShippedOutputOwnership` has 7 rows covering the named cases: parley cold, cold hand edit, no sidecar, malformed sidecar, undeclared, warm pulled pair, and warm hand edit. A real check fails without the fix, as confirmed by the revert above.
   - Without a `ship` row, the e2e test proves the refusal (the parley CI failure) is unchanged.
   - Not covered: first adoption, meaning a cold checkout with a stale output and no sidecar yet. That case is refused. It's correct, but parley#371 should describe the bootstrap: commit the sidecar from a warm apply.

6. **Architectural notes**
   - **ARCH-DRY: pass.** `stagedOutput`, `digest`, `safeParents` and `participates` are reused, and the sidecar name is defined once in `SidecarPath`.
   - **ARCH-PURE: pass.** The new logic is pure, and IO is limited to `vouched` behind the FS seam.
   - **ARCH-PURPOSE: pass.** The class is fixed: any declared shipped output now survives an upstream change. The parley-side wiring is a separable cross-repo step.
   - **ARCH-MOCK: pass.** No new external dependency is added, and the tests use `weavefs.OSFS` over a temp dir.
   - **ARCH-CONSTRAINTS: pass.** The cost is one extra read and hash per declared output during compile.
   - **ARCH-SECURE: pass.** The sidecar is treated as untrusted (regular files only, trimmed exact comparison), and anything it can't vouch for degrades to refusal.
   - **ARCH-ORDER: pass.** No state is held between events. The read-then-write gap between `vouched` and the write is within one compile, the same as the existing `observe`.
   - **ARCH-FUNERAL: pass.** There is one sidecar per declared output. On a warm checkout weave's retirement removes it. On a cold one it is removed by the consumer's `git rm`, which the atlas entry documents.
   - **Docs gate: pass.** The atlas entry is in `atlas/workflow/weave.md`, and the verb is listed in the `construct/base.manifest` comment. README doesn't list manifest verbs, so it needs no update.

7. **Plan revision recommendations**
   - Add a `## Revisions` entry noting the `ShippedPaths` signature: it takes `[]layer.Layer` and returns `([]string, error)`, refusing unsafe paths.
   - Tick the plan file's task checkboxes to match what was delivered.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: untrusted-input-branch-untested
    title: |
      The non-regular sidecar/output branch of vouched (symlink or directory) has no test
    detail: |
      vouched returns false for a non-regular Lstat result, but no TestShippedOutputOwnership row enters that branch. Add a row where the sidecar is a symlink to a matching digest, expecting a cold refusal.
  - id: new
    severity: Minor
    kind: defect
    family: plan-signature-drift
    title: |
      Plan Core concepts gives ShippedPaths as walk.Layer returning []string; the code takes layer.Layer and returns ([]string, error)
    detail: |
      Add a Revisions entry, and tick the plan file's task checkboxes (still unchecked although the issue checklist is complete).
  - id: new
    severity: Minor
    kind: defect
    family: misleading-diagnostic
    title: |
      The ship-row warning says "names no generated output" for a row naming a staged symlink output or an exported ancestor row
    detail: |
      ShipStaged only matches staged WriteFile outputs, so a symlink output, or an export row reaching a consumer that doesn't generate it, gets a misleading warning. Consider naming the reason.
```

The full `go test ./cmd/weave/...` run has finished, and the SHIP verdict still holds.

- **Passing:** every subpackage under `cmd/weave/internal/`, including `plan`, `intent` and `golden`, where the new pair rule is tested.
- **Failing:** two tests in the top-level `cmd/weave` package. Both fail the same way, with `git rev-parse --show-toplevel` → `not a git repository`. The review checkout keeps its temp directory under `.sdlc-review/`, which is listed in `GIT_CEILING_DIRECTORIES`, so git can't find a repository there.
  - `TestUniversalReviewSkillDiscoveredInDerivedRepo` existed before this change, so the failure isn't caused by this work.
  - `TestCompileRegeneratesShippedOutputCold/typo` is the new end-to-end test. All three of its cases pass when TMPDIR is outside the review checkout.

The rest of `cmd/weave` passed. I didn't run `make test`.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ca8c0394-73c6-4091-8443-b054b3fd9eb8","repository":"github.com/xianxu/ariadne","issue_id":"000422","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/b900e5bf7e217d3a","reviewed_head":"62afe6f15e4672de6f38d1ba55a22c683bdcf24d","slot":"ariadne:4","requested_at":"2026-10-17T09:27:34.140536Z"},"phase":"finished","pid":61921,"started_at":"2026-10-17T09:27:34.234664Z","finished_at":"2026-10-17T09:33:05.626077Z","session_id":"3b982b31-71b9-4769-ac3e-de6df77ca595","outcome":"returned"}] -->
