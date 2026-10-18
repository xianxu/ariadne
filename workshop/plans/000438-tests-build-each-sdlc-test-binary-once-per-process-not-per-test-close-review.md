# Boundary Review — ariadne#438 (whole-issue close)

| field | value |
|-------|-------|
| issue | 438 — tests: build each sdlc test binary once per process, not per test |
| repo | ariadne |
| issue file | workshop/issues/000438-tests-build-each-sdlc-test-binary-once-per-process-not-per-test.md |
| boundary | whole-issue close |
| milestone | — |
| window | 81acc7b58973a0612d17c9f8e4bcdba565410e11..dacdc2a2fcf584094ce86e48cb5ea7a14fe192fa |
| command | sdlc close --issue 438 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-18T03:41:49-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Now run the converted call-site tests.

I've finished checking the diff and running the tests. Here is the review.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The change does what the Spec asks. `testfix.BuiltBinary` caches one build per (directory, package, flags) per test process. A `sync.Once` per entry makes it safe when tests run in parallel, and a failed build is cached too, so every caller gets the same error. All six `go build` sites and both `go run ../vocabulary` calls now use it. All six `//lint:allow test-go-build` suppressions are gone; a grep finds no build or run left in any `_test.go` file outside the lint rule's own fixtures. The lifecycle is covered: one temp directory per process, removed by `TestMain` after `m.Run`, and directories older than 24 h are swept.

I ran the helper's tests under `-race` and they pass. The migrated tests pass, and no `sdlc-testbin-*` directory is left behind afterwards. `go vet` is clean. `TestFleetEndToEnd` fails when TMPDIR is inside the review checkout, because the fleet scan picks up the checkout repo as an ancestor of the temp dirs. It passes with `TMPDIR=/tmp/rv438`, so the failure comes from my environment, not from this diff. One cheap gap remains: the atlas testing section lists every `testfix` helper as a convention, but not the new one.

**1. Strengths**
- `cmd/sdlc/internal/testfix/build.go:47-68`: the map lock is released before `once.Do`. Parallel callers with different keys build at the same time, and callers with the same key wait for one build.
- The failure is cached and shared, so a broken build reports once per caller instead of recompiling for every test. `TestBuiltBinaryFailureIsSharedNotRetried` pins this.
- The ARCH-FUNERAL lifecycle is complete: lazy creation, removal after `m.Run` (`testmain_test.go:176`), and a prefix-scoped 24 h sweep for processes that were killed. `TestSweepBuiltDirsRemovesOnlyStaleOnes` checks that the sweep leaves unrelated directories alone.
- The build command is behind a seam (`buildCommand`), so the cache logic is tested without running a real `go build`.
- The settings_native weave binary still sits alone in its own hashed subdirectory, so putting `filepath.Dir(binary)` first on PATH exposes only `weave`, as before.

**2. Critical:** none.

**3. Important**
- `atlas/workflow/sdlc-binary.md:1849-1895`: the atlas update is missing for `testfix.BuiltBinary` / `RemoveBuiltBinaries`. That section documents each `testfix` convention (`PreferRealGit`, `QuietBackgroundGit`, `Quarantine`, `Watch`, `AwaitExit`, `DeadlineOnCue`), and a new end-to-end test needs to know to call `BuiltBinary` instead of `go build`. Fix: add a short paragraph saying a test that needs a built binary calls `testfix.BuiltBinary(t, dir, pkg, flags...)`, which builds once per process into `sdlc-testbin-*`, and that a package using it calls `RemoveBuiltBinaries()` after `m.Run`.

**4. Minor**
- `pkg/golint/rules.go:30`: the `test-go-build` message says "build it once per package (TestMain)". It could name `testfix.BuiltBinary`, so the rule points at the fix that now exists.
- `settings_native_test.go:91`: the weave build lost its 5-minute `CommandContext` deadline, and `buildCommand` has no timeout. A hung build now runs until the test binary's own `-timeout`.
- The cached build runs under the environment of whichever test calls first (any `t.Setenv` of HOME, GOFLAGS or PATH). Later callers get that binary whatever their own environment. Nothing breaks today, but the doc comment should say so.
- `build_test.go:43`: `BuiltBinary(t, ...)` is called from goroutines, so a `t.Fatal` there would come from a goroutine other than the test's own. It only matters on failure.

**5. Test coverage**
- The helper's tests cover: 16 parallel callers giving 1 build, separate builds per package and per flag set, the cached failure, removal, and the stale sweep.
- Not covered: two calls whose `dir` values resolve to the same absolute path, such as `"."` and the absolute path. These should share a key, and the code does call `filepath.Abs`.
- The "1 sdlc build + 1 vocabulary build" before/after figure in the Log is plausible from the code, but I did not re-measure it.

**6. Architecture**
- **ARCH-DRY:** pass. Seven hand-rolled build blocks became one helper, and `buildFleetE2EBinary` is now a one-line wrapper.
- **ARCH-PURE:** pass. The IO goes through the injectable `buildCommand`, and the cache logic is tested without exec.
- **ARCH-PURPOSE:** pass. The sweep covers every site plus the `go run` calls, and the lint rule now guards the whole class with no suppressions left.
- **ARCH-MOCK:** pass. `go` is a build tool, not a runtime dependency. The fake builder is enough for the cache logic, and the real call sites run the real build.
- **ARCH-CONSTRAINTS:** pass, with a note. Builds drop to one per package per process; the missing build timeout is listed under Minor.
- **ARCH-SECURE:** pass. The sweep only touches `sdlc-testbin-*` names under the process's own TMPDIR, and no credentials are involved.
- **ARCH-ORDER:** pass. The only state is a per-key once-cell; `RemoveBuiltBinaries` resets it, and the only caller is after `m.Run`.
- **ARCH-FUNERAL:** pass. Creator, last user, remover (`TestMain`), backstop (24 h sweep) and bound (one binary per package per live process) are all named.

**7. Plan revisions:** none needed. The plan matches the code. The issue file has a duplicated `## Spec` heading, a duplicated `### 2026-10-18` heading and an empty `- [ ]` item, which are worth tidying.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: docs-gate-new-surface
    title: |
      atlas update appears missing for testfix.BuiltBinary / RemoveBuiltBinaries
    detail: |
      atlas/workflow/sdlc-binary.md documents each testfix test convention (PreferRealGit, QuietBackgroundGit, Quarantine, Watch, AwaitExit, DeadlineOnCue) but not the new build-once helper. Add a paragraph saying a test needing a built binary calls testfix.BuiltBinary (one build per process into sdlc-testbin-*, 24h stale sweep) and that the package TestMain calls RemoveBuiltBinaries after m.Run.
  - id: new
    severity: Minor
    kind: defect
    family: lint-message-names-fix
    title: |
      test-go-build lint message should name testfix.BuiltBinary as the fix
    detail: |
      pkg/golint/rules.go:30 says build it once per package (TestMain); the helper now exists and is the intended remedy.
  - id: new
    severity: Minor
    kind: defect
    family: subprocess-bounded-deadline
    title: |
      settings_native weave build lost its 5-minute CommandContext deadline
    detail: |
      buildCommand uses exec.Command with no timeout, so a hung go build now runs until the test binary's -timeout instead of failing at 5 minutes.
  - id: new
    severity: Minor
    kind: defect
    family: memoized-env-capture
    title: |
      memoized build runs under the first caller's environment
    detail: |
      Whichever test calls BuiltBinary first decides the HOME, GOFLAGS and PATH the shared build sees. Nothing breaks today, but the BuiltBinary doc comment should state it.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ae8f83c9-b1b9-421a-838b-7a23e79470eb","repository":"github.com/xianxu/ariadne","issue_id":"000438","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/fb6ca8d3b568555d","reviewed_head":"dacdc2a2fcf584094ce86e48cb5ea7a14fe192fa","slot":"ariadne:2","requested_at":"2026-10-18T10:40:03.811796Z"},"phase":"finished","pid":54094,"started_at":"2026-10-18T10:40:03.902522Z","finished_at":"2026-10-18T10:41:49.176327Z","session_id":"72f26c6f-cf0d-43a7-aa33-ae74e09327a7","outcome":"returned"}] -->
