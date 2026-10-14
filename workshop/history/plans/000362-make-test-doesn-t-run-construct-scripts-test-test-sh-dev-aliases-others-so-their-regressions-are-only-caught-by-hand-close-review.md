# Boundary Review — ariadne#362 (whole-issue close)

| field | value |
|-------|-------|
| issue | 362 — make test doesn't run construct/scripts/test/*.test.sh (dev-aliases, others), so their regressions are only caught by hand |
| repo | ariadne |
| issue file | workshop/issues/000362-make-test-doesn-t-run-construct-scripts-test-test-sh-dev-aliases-others-so-their-regressions-are-only-caught-by-hand.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3e2d787f542e015bde879ebd671607db96ce7fb1..e246d7e5c7b97298bf7de66455e6693b3d9899ee |
| command | sdlc close --issue 362 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T22:26:13-07:00 |
| verdict | SHIP |

## Review

Diff and issue read; now running the new runner and the Go test to confirm they pass and that a broken test turns them red.

Now the deliberate-break proof (in the disposable checkout), plus a check that the Go package is picked up by test-shard.

```verdict
verdict: SHIP
confidence: high
```

This change does what the issue asked for, and I confirmed it by running it. `scripts/test/construct-scripts.test.sh` is now the single runner for all six `construct/scripts/test/*.test.sh`. The CI merge check (`20-ci-workflow-tests.sh`) picks it up through its existing `scripts/test/*.test.sh` glob. `make test` picks it up through the new Go package `scripts/test/scripts_test.go`, which `test-shard.py`'s `go list ./...` includes (I checked: `github.com/xianxu/ariadne/scripts/test` is listed). In this checkout the runner passes 6/6 in about 9s under /bin/bash 3.2.57, and `go test` on the package passes. For the "a deliberately broken test fails `make test`" clause, I added `exit 1` to `vm-log.test.sh`. `go test` then failed with `construct-scripts: 5 passed, 1 failed` and named `vm-log.test.sh`. I then reverted the file. Both modes of the Done-when clause are covered: everything runs when nothing is excluded, and the `EXCLUDED` list is where an exclusion would be named. Nothing blocks the change; the remaining notes are minor.

1. **Strengths**
   - One list for both CI and `make test` (`scripts/test/construct-scripts.test.sh:11`). The Go test calls the same runner instead of keeping its own list, so the two can't drift.
   - It refuses a missing directory instead of passing quietly: an unmatched glob is caught at `construct-scripts.test.sh:17`, and the array expansion is guarded for bash 3.2 with `set -u` (line 19).
   - Each test runs in its own process, failures are named, and the last 20 lines of output are shown (lines 24–28). That makes CI failures easy to act on.
   - The Go entry point uses `testgate.Enter()`, matching the convention from #324.
   - Both atlas pages (`ci-merge-check.md`, `base-layer.md`) now say where these tests run, and what they say matches the code.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `scripts_test.go:22`: `exec.Command` has no context or deadline. A shell test that hangs would stall its shard until go test's 30m timeout. The repo already has `testgate.HangCeiling` for this; `exec.CommandContext` with that ceiling would bound the wait.
   - `construct-scripts.test.sh:6`: the comment says "macOS /bin/bash 3.2 locally". That is only true when /bin/bash comes first on PATH. On a machine with Homebrew bash it would run bash 5. The comment could say "whichever bash is on PATH".
   - The local pre-merge run of `20-ci-workflow-tests` and a `make test` in the same session will each run the construct tests, about 9s twice. That's acceptable; just noting it.

5. **Test coverage:** The proof that a broken test fails the suite was reproduced here, not taken from the Log. The runner has no self-test for the `EXCLUDED` branch. That's acceptable while the list is empty.

6. **Architecture**
   - **ARCH-DRY: pass.** The exclusion list lives in one place, and the Go test reuses the shell runner. Its loop looks like the one in `20-ci-workflow-tests.sh`, but it adds reporting and runs over a different directory. Merging them isn't worth doing yet.
   - **ARCH-PURE: pass.** This is test-harness glue only; there's no business logic in it.
   - **ARCH-PURPOSE: pass.** All six construct tests run in both places, nothing was deferred, and nothing is silently skipped.

7. **Plan revisions:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: unbounded-subprocess-wait
    title: |
      TestConstructScripts runs the shell runner without a deadline
    detail: |
      scripts/test/scripts_test.go:22 uses exec.Command with no context, so a hanging construct test stalls the shard until the 30m go test timeout. Use exec.CommandContext with testgate.HangCeiling. This is the only subprocess call in the window.
  - id: new
    severity: Minor
    family: doc-claim-accuracy
    title: |
      Runner header says "macOS /bin/bash 3.2 locally", but it runs whatever bash is on PATH
    detail: |
      construct-scripts.test.sh:6 says the runner uses /bin/bash 3.2 on macOS. That only holds when /bin/bash comes first on PATH; with Homebrew bash first it runs bash 5. Reword to "the bash on PATH". This is the only such claim in the window; the atlas wording is accurate.
```

---

## Re-review — 2026-10-13T22:28:34-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 362 — make test doesn't run construct/scripts/test/*.test.sh (dev-aliases, others), so their regressions are only caught by hand |
| repo | ariadne |
| issue file | workshop/issues/000362-make-test-doesn-t-run-construct-scripts-test-test-sh-dev-aliases-others-so-their-regressions-are-only-caught-by-hand.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4f7bf5e82ca0426450afba28abb288e3053efc18..52fd1fd07d6b9b764a1a9b4af9b36044738034cb |
| command | sdlc close --issue 362 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T22:28:34-07:00 |
| verdict | SHIP |

## Review

Tests pass. I'm checking whether the new `syscall` calls need a unix build tag.

The interdiff is small and passes. I'll check BR-1 by forcing a hang in the disposable checkout: a short deadline plus a test that leaves a background child holding the pipe.

```verdict
verdict: SHIP
confidence: high
```

Both earlier findings are fixed. The new changes add no Critical or Important problems. The only code changes are the deadline and process-group kill in `scripts/test/scripts_test.go` and the reworded header comment in `scripts/test/construct-scripts.test.sh`. The issue file only gains the close entry, its Log note and the status change to codecomplete. Spec, Revisions and Done-when are untouched, so the scope did not change.

I tested BR-1 in the disposable checkout. I added a construct test that starts a background `sleep 600` and then sleeps itself, and cut `runnerDeadline` to 20s. `go test` failed at 20.13s with "construct/scripts/test did not finish within 20s". The background child did not hold the output pipe open. I then reverted both changes and confirmed with `git status` that the checkout is clean. Without the changes, `go vet` passes and `go test -count=1` passes in 8.1s.

1. **Strengths**
   - `scripts_test.go:33-43` doesn't just add a timeout. It runs the runner in its own process group, kills the whole group (`SIGKILL` to `-pid`, ignoring "process already gone"), and sets `WaitDelay` as a backstop. Without that, a background child of a hanging test could keep the pipe open and `CombinedOutput` would block after the deadline. This is the same approach as `cmd/weave/internal/weavefs/runner.go:66-71`.
   - `scripts_test.go:45-47` checks `ctx.Err()` before checking the exit error, so a timeout is reported as a timeout and not as a failing test.
   - The new header wording (`construct-scripts.test.sh:6-8`) is accurate: it runs the first bash on PATH, which on macOS is 3.2 unless a newer bash comes first.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `scripts_test.go:25` adds its own `runnerDeadline = 5 * time.Minute` instead of using `testgate.HangCeiling` (2m). That's what BR-1 suggested, and the ceiling's own doc says it exists "so every module package's tests share one value" (`pkg/testgate/ceiling.go:11`). The runner takes about 13s, so 2m still leaves roughly 9× headroom. **ARCH-DRY:** use `testgate.HangCeiling`, or explain why a whole-suite runner needs a different ceiling.
   - (Already in the tree, not raised as new.) The process-group kill is written out again here, even though `cmd/sdlc/internal/processgroup` and `weavefs/runner.go` already have it. Go's `internal` rule stops this package importing the first one, so there's nothing cheap to fix in this change.

5. **Test coverage:** The deadline is shown to work by the hang I reproduced, not by a test kept in the repo. That's acceptable for test-harness glue; a permanent meta-test would need an injectable runner, which isn't worth it for a Minor.

6. **Architecture** (applied to the new changes only)
   - **ARCH-DRY:** flagged as Minor; the separate deadline constant, as above.
   - **ARCH-PURE:** passes. This is test glue only.
   - **ARCH-PURPOSE:** passes. BR-1 was the only subprocess call in the window and BR-2 the only such claim, so fixing them covers the whole class.

7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      exec.CommandContext with its own process group killed whole, plus WaitDelay. Checked by forcing a hang with a 20s deadline and a test that leaves a background child: it failed at 20.13s with "did not finish within 20s". Reverted afterwards.
  - id: BR-2
    disposition: addressed
    note: |
      construct-scripts.test.sh:6-8 now says "the first bash on PATH (on macOS that is /bin/bash 3.2 unless a newer bash comes first)", which is accurate.
findings:
  - id: new
    severity: Minor
    family: test-ceiling-single-source
    title: |
      TestConstructScripts adds its own 5m runnerDeadline instead of the shared testgate.HangCeiling
    detail: |
      scripts/test/scripts_test.go:25. pkg/testgate/ceiling.go says HangCeiling exists so every package's tests share one value. The runner takes about 13s, so the 2m ceiling still leaves about 9x headroom. ARCH-DRY: use testgate.HangCeiling, or explain why a whole-suite runner needs its own ceiling.
```

---

## Re-review — 2026-10-13T22:29:49-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 362 — make test doesn't run construct/scripts/test/*.test.sh (dev-aliases, others), so their regressions are only caught by hand |
| repo | ariadne |
| issue file | workshop/issues/000362-make-test-doesn-t-run-construct-scripts-test-test-sh-dev-aliases-others-so-their-regressions-are-only-caught-by-hand.md |
| boundary | whole-issue close |
| milestone | — |
| window | 14b19ee280dee33f6e3ec4bef1770bc2443c68f0..ae4b17867fa160731f6472b2fda860574431bea3 |
| command | sdlc close --issue 362 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T22:29:49-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

BR-3, the only open finding, is fixed, and this window doesn't introduce any new problems. It changes one line of code: `scripts/test/scripts_test.go:26` now sets `runnerDeadline = testgate.HangCeiling` instead of its own `5 * time.Minute`. The comment above it now says the deadline is the shared ceiling, and `time` is still imported because `WaitDelay` uses it. The other changes are generated files (the gate ledger and the close-review record) and the issue-file entries that come with a close: the card mirror hash, `actual_hours`, and one Log line. `## Spec`, `## Revisions` and `## Done when` are unchanged, so scope did not change. There were no conflict resolutions to judge. In the review checkout, `go vet ./scripts/test/` passes and `go test -count=1 ./scripts/test/` passes in 7.5s, well under the 2-minute ceiling.

1. **Strengths**
   - `scripts/test/scripts_test.go:26` now takes its deadline from `testgate.HangCeiling`, the shared value the rest of the tree already uses (for example `cmd/weave/internal/staging/stage_test.go:39` and `cmd/weave/internal/weavefs/runner_owned_test.go:45`). That makes it one value for every package, which is what `pkg/testgate/ceiling.go:11` intends.
   - The fix only changes the constant. The process-group kill, `WaitDelay` and the check of `ctx.Err()` before the exit error were verified in round 2, and they are unchanged.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - The issue's newest Log line, added by this window's close, describes the fix as "a 5-minute deadline". The constant is now 2 minutes. The line records the earlier close, so I'm not raising it as a finding.

5. **Test coverage:** The deadline mechanism was shown to work in round 2 by forcing a hang. This window only changes the constant, so a new regression test isn't needed. The runner finishes in about 7–13s, well inside 2 minutes.

6. **Architecture** (this window only)
   - **ARCH-DRY:** pass. The separate constant is gone; the test uses the shared ceiling.
   - **ARCH-PURE:** pass. This is test setup code only.
   - **ARCH-PURPOSE:** pass. This test was the one place in the window that set its own ceiling, and it now uses the shared one.

7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      scripts/test/scripts_test.go:26 now sets runnerDeadline = testgate.HangCeiling (2m), replacing the local 5m constant; go vet and go test -count=1 ./scripts/test pass (7.5s).
```
