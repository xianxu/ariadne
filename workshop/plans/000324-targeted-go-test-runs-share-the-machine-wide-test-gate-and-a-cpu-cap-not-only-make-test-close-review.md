# Boundary Review — ariadne#324 (whole-issue close)

| field | value |
|-------|-------|
| issue | 324 — Targeted go test runs share the machine-wide test gate and a CPU cap, not only make test |
| repo | ariadne |
| issue file | workshop/issues/000324-targeted-go-test-runs-share-the-machine-wide-test-gate-and-a-cpu-cap-not-only-make-test.md |
| boundary | whole-issue close |
| milestone | — |
| window | 19a4cf454e4fb41f738bc8b198076508f005a2df..3a8bb9c28e6932b608696ee7476cdba98ed1cbc3 |
| command | sdlc close --issue 324 |
| reviewer | claude |
| timestamp | 2026-10-10T18:51:22-07:00 |
| verdict | SHIP |

## Review

Checking whether `make test` sets its own `-parallel`, since the gate applies its cap even under make test.

The review is done. Verdict: SHIP. I found no Critical or Important problems, only four Minors.

```verdict
verdict: SHIP
confidence: high
```

I inspected the pinned window with the stat, name-status and full-diff recipes, and read `pkg/testgate` in full. I also checked the issue's Spec, Plan and Log, `scripts/test-lock.py`, `Makefile.workflow` and `scripts/test-shard.py`. The diff delivers the issue's purpose: a direct `go test` of every heavy package goes through one machine-wide gate. That gate lives in the test binary, so it doesn't depend on the slot environment. Each gated binary waits out a full suite holding #319's lock, then takes one of N slots held with flock. Waits are loud and the #319 hatches carry over. The timeout-budget fix that the measurement surfaced is real and tested.

Ran at HEAD:
- `go test ./pkg/testgate/`: ok.
- `go vet ./pkg/testgate/`: clean.
- `go test -run TestGitDrivingPackagesPreferRealGit ./cmd/sdlc/`: ok.

**Strengths**
1. **Gate placement.** Putting the gate in `TestMain` (`pkg/testgate/testgate.go:106`) reaches every harness and shell. The kernel releases the slot when the process exits, so there's no stale-lock recovery to get wrong.
2. **Timeout budget.** Waiting at most a quarter of `-test.timeout`, then shrinking `-test.timeout` by the wait (`testgate.go:118-120`), handles cmd/go's wall-clock kill properly. It is proven end-to-end through a re-exec'd binary in `TestDirectRunShrinksTimeoutByTheWait`.
3. **Suite probe.** The shared, non-blocking probe of `test-suite.lock`, dropped at once (`testgate.go:212`), is the right primitive: targeted runs can't starve the suite's poll for an exclusive lock. A test confirms the suite can still lock exclusively afterwards (`testgate_test.go:149`).
4. **Enforced coverage.** The existing `realgit_guard_test.go` was extended so every git-driving `cmd/sdlc` package must call `testgate.Enter()`. Coverage is enforced, not just listed (ARCH-PURPOSE).
5. **Hermetic tests.** The helper re-exec sets `WF_TEST_GATE_DIR` and empties `WF_TEST_LOCK_FILE`, so the tests never touch the real `~/.local/state` (ARCH-SECURE).

**Critical:** none.

**Important:** none.

**Minor**
1. **The parallel cap applies even when the gate "stands aside".** `capParallel` runs before the bypass and off checks (`testgate.go:108`). So under `make test` (whose packages job passes no `-parallel`) and under `WF_TEST_LOCK=off`, `-test.parallel` is still capped at 2. The atlas and package doc say the gate stands aside. The impact today is negligible: only 2 `t.Parallel()` calls exist across the gated packages, all in cmd/sdlc. Fix: return before `capParallel` on bypass, or document the cap as always on. Add a `parallel=` check to the `WF_TEST_LOCK_HELD_BY` case in `TestDirectRunHonorsCapAndGate`.
2. **`writeRecord` forks git.** It runs `git rev-parse` on every slot acquire (`testgate.go:234`), before `PreferRealGit`, so it goes through the xcrun shim. That is an extra git fork in a feature meant to cut git forks, and an external call outside any seam (ARCH-MOCK). The repo name could come from the working directory, or be dropped.
3. **No guard for weave.** Weave's gated packages are listed by hand, and no test like the cmd/sdlc guard covers `cmd/weave/...`. A new git-driving weave package would go ungated silently. The current weave coverage does match the git-spawning packages (`acquire`, `refresh`, root).
4. **Unrelated churn.** `cmd/sdlc/publishgate_test.go` has gofmt-only whitespace changes.

**Test coverage**
- Covered: wait then acquire, waiting out a full suite, the holder described in the wait lines, going ahead past the budget, the timeout refusal, the hatch, the cap only when unset, the make-test bypass, the timeout shrink end-to-end, and config parsing.
- Not covered: an invalid `WF_TEST_PARALLEL` value (`flag.Set` fails silently), and the "unidentified holder" fallback for a truncated or hand-edited record. That fallback is simple, and it does fail visibly rather than fabricating a holder.

**Architecture principles**
- **ARCH-DRY: pass.** The env-var names and hatches reuse #319's. The record format is compatible with `test-lock.py`'s.
- **ARCH-PURE: pass.** `enter` takes injected `stderr` and `sleep`. IO is limited to flock and file reads.
- **ARCH-PURPOSE: pass.** Coverage is enforced for cmd/sdlc. Minor 3 covers weave.
- **ARCH-MOCK:** Minor 2 only. Tests use a real flock on temp directories, which is fine for a kernel primitive.
- **ARCH-CONSTRAINTS: pass.** The Log has load and wall-time measurements, gate off and on, with and without full suites. The slot count and the wait budget are bounded.
- **ARCH-SECURE: pass.** Persisted records are parsed and degrade to "an unidentified holder", and the tests are isolated from real state.
- **ARCH-ORDER: pass.** State is kernel flocks that live as long as the process. The wait loop's transitions (acquired, budget-exceeded, timeout, suite running) are explicit and tested. The suite-starts-after-probe race is an accepted, benign interleaving.
- **ARCH-FUNERAL: pass.** There are N single-record files, rewritten on every acquire, bounded by `WF_TEST_GATE_SLOTS`, as the Plan's line says. Lowering the slot count leaves the higher-numbered files behind, which is bounded and harmless.

**Docs:** the atlas section is added and accurate apart from Minor 1. The README has no #319 lock coverage either, so leaving it unchanged is consistent.

**Plan revisions:** none needed. The Plan matches the code; the full-suite probe was added under the Spec's "take the #319 lock" item and is recorded in the Log.

```findings
findings:
  - id: new
    severity: Minor
    family: bypass-path-still-applies-side-effects
    title: |
      -test.parallel cap is applied even when the gate stands aside (make test bypass, WF_TEST_LOCK=off)
    detail: |
      capParallel runs before enter()'s bypass/off checks (testgate.go:108), contradicting the atlas/package-doc "stands aside"; the HELD_BY helper test asserts only held=false. Impact is small today (2 t.Parallel calls in gated packages).
  - id: new
    severity: Minor
    family: external-call-outside-seam
    title: |
      writeRecord forks git rev-parse on every slot acquire, before PreferRealGit
    detail: |
      An extra (xcrun-shimmed) git fork in a gate meant to curb git forks; the repo name could come from the cwd or be dropped (ARCH-MOCK).
  - id: new
    severity: Minor
    family: coverage-enforced-by-guard
    title: |
      weave's gated packages are hand-listed with no source guard like TestGitDrivingPackagesPreferRealGit
    detail: |
      A new git-driving cmd/weave package would go ungated silently; current coverage matches today's git-spawning packages.
  - id: new
    severity: Minor
    family: unrelated-churn
    title: |
      gofmt-only whitespace change in cmd/sdlc/publishgate_test.go rides along in the diff
```

---

## Re-review — 2026-10-10T19:02:42-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 324 — Targeted go test runs share the machine-wide test gate and a CPU cap, not only make test |
| repo | ariadne |
| issue file | workshop/issues/000324-targeted-go-test-runs-share-the-machine-wide-test-gate-and-a-cpu-cap-not-only-make-test.md |
| boundary | whole-issue close |
| milestone | — |
| window | 19a4cf454e4fb41f738bc8b198076508f005a2df..f28f494e3ba2de9bc5f9f378caba894f1cfdbeca |
| command | sdlc close --issue 324 |
| reviewer | claude |
| timestamp | 2026-10-10T19:02:42-07:00 |
| verdict | SHIP |

## Review

Tests pass; verifying the BR-3 guard's coverage, then the verdict.

```verdict
verdict: SHIP
confidence: high
```

Round 2 confirms all four round-1 Minors are resolved. Three were fixed, and each fix has a test that would fail without it. BR-4 is withdrawn because the rationale given for keeping it is sound. `pkg/testgate` is a clean, small gate built on flock. The make-test bypass and `WF_TEST_LOCK=off` now skip it completely, with no parallel cap. Slot records no longer fork git, and cmd/weave has a source guard for its git-driving packages. The targeted tests pass at HEAD: `pkg/testgate`, the cmd/weave guard, and the cmd/sdlc guard plus `TestHasConflictMarkers`. One new Minor came up: both source guards can be satisfied by their own string literal.

1. **Strengths**
   - **BR-1 fix is checked against Go's own default.** `pkg/testgate/testgate_test.go:59-64` sets `GOMAXPROCS=7` and expects `parallel=7`. Before the fix the cap would have printed `parallel=2`, so this test fails without the fix. It covers both the silent bypass and the loud off-switch.
   - **BR-2 fix:** `writeRecord` (`testgate.go:243`) no longer forks a process. `Repo` is now `omitempty`, so suite records written by `scripts/test-lock.py` are still described correctly.
   - **BR-3 fix:** the guard pattern finds exactly the 5 weave test files that spawn git. Those files sit in 3 directories, and every one of them has a `TestMain` that calls `testgate.Enter()`.
   - The budget for waiting and shrinking `-test.timeout` is tested, and the Log grounds it in a real SIGQUIT kill seen during measurement.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **The source guards match their own literal.** `cmd/weave/testgate_guard_test.go` contains the string `"testgate.Enter()"`. `cmd/sdlc/realgit_guard_test.go` contains both `"testgate.Enter()"` and `"testfix.PreferRealGit()"`. Each guard reads every `_test.go` file in its directory, including itself, so the package root always counts as compliant. Deleting `cmd/weave/testmain_test.go` or `cmd/sdlc/testmain_test.go` would leave the guard green.

5. **Test coverage notes:** the helper process re-runs the test binary for real, so it exercises `TestMain`, flag handling and the slot lock end to end. That is a strong test. The guards are the weak spot, as described in the Minor above.

6. **Architecture**
   - **Passes:**
     - ARCH-DRY: the weave guard mirrors the sdlc guard. That is acceptable for two small, separate command trees.
     - ARCH-PURE: the `enter` logic takes an injected sleep function and output writer, so it is testable.
     - ARCH-MOCK: the BR-2 git call is gone.
     - ARCH-CONSTRAINTS: the slot count, parallel cap and wait budget are all stated and measured.
     - ARCH-SECURE: a malformed record degrades to "an unidentified holder".
     - ARCH-ORDER: the flock is released by the kernel, the suite probe never keeps the shared lock, and the wait loop is bounded by `timeout/4`.
     - ARCH-FUNERAL: there are N slot files, each rewritten in place and never appended.
   - **Flag:** the self-matching guard (the Minor above) is a weakness in how coverage is enforced.

7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      testgate.go:104 gates capParallel on !bypass && !off; testgate_test.go:59-64 asserts parallel=7 (GOMAXPROCS) for both hatches — red without the fix.
  - id: BR-2
    disposition: addressed
    note: |
      writeRecord no longer execs git; os/exec import removed; holdSlot fixture and describe() expectation updated.
  - id: BR-3
    disposition: addressed
    note: |
      cmd/weave/testgate_guard_test.go walks weave test files; the 3 git-spawning dirs all call testgate.Enter(); guard passes.
  - id: BR-4
    disposition: withdrawn
    note: |
      The change only gofmt-aligns a table this implementor added in #320; reverting would leave main misformatted. Acceptable housekeeping.
findings:
  - id: new
    severity: Minor
    family: source-guard-matches-own-literal
    title: |
      Source guards are satisfied by their own string literal, so the package-root TestMain is unguarded
    detail: |
      cmd/weave/testgate_guard_test.go and cmd/sdlc/realgit_guard_test.go both contain "testgate.Enter()" (and sdlc "testfix.PreferRealGit()") and are read as part of their own directory, so deleting the root testmain_test.go stays green. Rule: a source guard must skip its own file (or match only a call statement in a TestMain), not a bare substring.
```
