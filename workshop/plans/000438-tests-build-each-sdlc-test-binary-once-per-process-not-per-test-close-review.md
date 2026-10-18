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

---

## Re-review — 2026-10-18T04:26:17-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 438 — tests: build each sdlc test binary once per process, not per test |
| repo | ariadne |
| issue file | workshop/issues/000438-tests-build-each-sdlc-test-binary-once-per-process-not-per-test.md |
| boundary | whole-issue close |
| milestone | — |
| window | 81acc7b58973a0612d17c9f8e4bcdba565410e11..81c832d5e0008578d7a03783a0d4055cc6358000 |
| command | sdlc close --issue 438 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-18T04:26:17-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Tests pass with TMPDIR outside the checkout. The fleet failure came from the review sandbox's TMPDIR sitting inside the repo, not from this diff. Next I'll check the golint run and the deadline path.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The diff does what #438 set out to do. `testfix.BuiltBinary` builds each (directory, package, flags) binary at most once per test process. All six `go build` sites and both `go run ../vocabulary` calls now use it, all six `//lint:allow test-go-build` suppressions are gone, and the `test-go-build` rule now guards the whole class. The helper tests pass under `-race`, and so do the migrated end-to-end tests when run with TMPDIR outside the checkout. Inside the checkout, `TestFleetEndToEnd` fails because the review sandbox puts TMPDIR inside the repository: the inventory step picks up the repo as an extra row. That step doesn't involve the build, so this diff isn't the cause. No `sdlc-testbin-*` directory is left after the process exits. Prior findings BR-1, BR-2 and BR-4 are fixed. BR-3 is only partly fixed: the 5-minute bound exists, but no test exercises it, and it may not hold against a hung compiler subprocess. It is Minor and doesn't block.

Done-when: current — it requires each binary built at most once per process, the six sites moved to the helper, and their suppressions removed; the diff delivers all three, and the Spec changes since change-code (the timeout and the environment note) are refinements inside that scope.

1. **Strengths**
   - `cmd/sdlc/internal/testfix/build.go:64-86`: one `sync.Once` per key behind a mutex-guarded map. A failed build is remembered too, so every caller sees the same error and nothing retries. Tests cover 16 parallel callers and the shared failure.
   - ARCH-FUNERAL is handled properly: each process gets its own `sdlc-testbin-*` directory, `RemoveBuiltBinaries` runs after `m.Run` (`testmain_test.go:176`), and later processes sweep directories older than 24 h. A test checks that the sweep spares fresh directories and unrelated ones.
   - The key uses `filepath.Abs`, so two spellings of the same directory share one build, and a test pins that.
   - `buildCommand` is a seam the tests replace, so the memoization tests never run the real `go`.
   - In `settings_native_test.go`, prepending `filepath.Dir(binary)` to PATH is safe: each key gets its own subdirectory, so only `weave` lands on PATH.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - BR-3 is still open (see dispositions). `buildCommand` kills only the `go` process. Its output goes to a `bytes.Buffer`, so `Wait` also waits for the output-copy goroutines, and a hung `compile`/`link` child keeps that pipe open. Setting `cmd.WaitDelay` closes that gap. There is also a small race at `build.go:51`: if `Wait` returns right as the timer fires, a successful build is reported as "exceeded".
   - `sweepBuiltDirs` judges age by the directory's mtime. A live process running longer than 24 h could have its directory swept by another process. In practice this doesn't matter.

5. **Test coverage notes:** the build-once path, the shared failure, the stale sweep and cleanup are all tested. Nothing tests the timeout path: no fake build hangs, and `buildTimeout` can't be shortened in a test.

6. **Architecture**
   - ARCH-DRY pass: one helper replaced six copies.
   - ARCH-PURE pass: memoization and sweep logic are separate from the IO behind the `buildCommand` seam.
   - ARCH-PURPOSE pass: every site moved, and the lint rule now enforces the class.
   - ARCH-MOCK pass: the `go` toolchain sits behind a seam with a counting fake.
   - ARCH-CONSTRAINTS: mostly pass. Builds are bounded at one per package per process, with before/after measurements. The bound has the gap noted under BR-3.
   - ARCH-SECURE pass: the sweep only touches its own prefix under TMPDIR, and no credentials are involved.
   - ARCH-ORDER pass: state is a per-key `Once`, and `RemoveBuiltBinaries` runs only after `m.Run`.
   - ARCH-FUNERAL pass.
   - Follow-up: sharing one binary across shards remains out of scope, as the Spec says.

7. **Plan revisions:** none needed.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      atlas/workflow/sdlc-binary.md:1896 adds a paragraph next to the other testfix conventions covering BuiltBinary, the sdlc-testbin-* directory, RemoveBuiltBinaries after m.Run, and the 24 h sweep.
  - id: BR-2
    disposition: addressed
    note: |
      pkg/golint/rules.go:30 now names testfix.BuiltBinary as the remedy.
  - id: BR-3
    disposition: not-addressed
    note: |
      build.go:38-54 adds a 5-minute AfterFunc kill, but no test exercises it, and killing only the go process may not bound a hung compile/link child holding the output pipe (Wait waits for the copy goroutines). Rule for the subprocess-bounded-deadline family: every bounded subprocess sets cmd.WaitDelay along with its kill, and a fake hanging build shows the bound holds.
  - id: BR-4
    disposition: addressed
    note: |
      The BuiltBinary doc comment (build.go:61-62) now says the build runs once under the first caller's environment (HOME, GOFLAGS, PATH).
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"4d1bd3ab-ecb3-4b38-956c-aef080165a15","repository":"github.com/xianxu/ariadne","issue_id":"000438","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/40c1b49688537acd","reviewed_head":"81c832d5e0008578d7a03783a0d4055cc6358000","slot":"ariadne:2","requested_at":"2026-10-18T11:24:20.666817Z"},"phase":"finished","pid":14438,"started_at":"2026-10-18T11:24:20.833827Z","finished_at":"2026-10-18T11:26:16.871267Z","session_id":"a16aa917-701b-4f6b-8336-06bec7db4acd","outcome":"returned"}] -->
