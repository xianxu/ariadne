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
