# Boundary Review — ariadne#338 (whole-issue close)

| field | value |
|-------|-------|
| issue | 338 — main is red: TestPlanningReviewCompetingResults fails on clean main (planning review stale: issue file changed) |
| repo | ariadne |
| issue file | workshop/issues/000338-main-is-red-testplanningreviewcompetingresults-fails-on-clean-main-planning-review-stale-issue-file-changed.md |
| boundary | whole-issue close |
| milestone | — |
| window | cdadf0ec114dcbc77918895b30846f0a1c17d33b..436a1f42e317dc7fd1fe0fb48ea9bf5c4aacec9b |
| command | sdlc close --issue 338 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-12T15:40:07-07:00 |
| verdict | SHIP |

## Review

The planning-review test family is still running in the background. In the meantime I'll check the timeout constant and run just the competing-results test on its own.

The test passed 3 of 3 runs at about 6s each. Now I'm running the whole `TestPlanningReview` family once with timings, to find what used up the 10 minutes earlier.

```verdict
verdict: SHIP
confidence: high
```

This is a small, test-only fix, and the root cause holds up. Since #325, the fake reviewer's wait loop has read its spin bound from `PLANNING_SPINS`. `startPlanningReviewCLI` set that variable, but the competing CLI in `TestPlanningReviewCompetingResults` built its own environment and left it out. I checked what that does: `[ "$n" -lt "$PLANNING_SPINS" ] || exit 1` with the variable unset prints `integer expression expected` and exits 1. So the second reviewer wrote its `ready` file and quit right away. Under `--force` that CLI could then finish first and change the issue file, and the first CLI would fail as stale. The fix moves the environment into one helper, `planningAgentEnv` (`cmd/sdlc/planningreview_test.go:32`), used by both CLIs. It also replaces the remaining fixed 25s/10s/5s deadlines with `testfix.HangCeiling`. The planning-review tests contain no other environment built by hand. `TestPlanningReviewCompetingResults` passed 3 of 3 runs (`-count=3`, about 6s each), and the whole `TestPlanningReview` family passed once (184s, mostly build time). An earlier `-count=5` run of the whole family hit Go's default 10-minute package timeout. The output was cut off before the test list, so I can't name the test that was running. Since a single pass takes about 45s of test time, it looks like load or the default timeout rather than a hang, but I didn't prove that. Nothing blocks SHIP.

1. **Strengths**
   - The fix goes after the cause, not just the one call site. Building the environment in a single helper means a CLI can no longer be launched with only part of it (`planningreview_test.go:32-36`, used at `:90` and `:403`). This is ARCH-DRY applied the right way.
   - The leftover fixed deadlines now follow #325's own `HangCeiling` rule (`:398`, `:413`, `:418`).
   - The cause in the Log is specific and correct: I confirmed the shell behaviour directly. The Log also explains why the failure was intermittent (it only happens when the second CLI wins the race), which matches the plan's "1 in 5" reproduction.
   - It stays inside the issue's scope and does not quarantine the test, as the Spec required.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - If `PLANNING_SPINS` is unset, the fake reviewer still fails silently and only sometimes: it writes `ready` and then exits 1 (`planningreview_test.go:73`). Adding `: "${PLANNING_SPINS:?}"` (and the same for `PLANNING_BARRIER`) before the loop would make a wrong environment fail every time and loudly. That would turn this flaky failure into one that a regression test catches deterministically.
   - The Plan marks "Verify: 10 runs … then make test" as done, but the Log doesn't record the outcome (pass count, `make test` result). Add one line so Done-when has evidence behind it.

5. **Test coverage notes**
   - The fix is test-only, and the bug was a race in the fixture, so no test fails deterministically without it. The `:?` guard above would fix that cheaply.
   - I didn't revert the fix to re-check the 1-in-5 failure: it's probabilistic and a run takes several minutes. The mechanism is confirmed from the shell semantics instead.

6. **Architectural notes**
   - ARCH-DRY: pass. One environment builder now replaces two copies that had drifted apart.
   - ARCH-PURE: pass. The change is only test-fixture wiring.
   - ARCH-PURPOSE: pass. I searched for every place that builds the fake reviewer's environment; `planningAgentEnv` is now the only one.
   - ARCH-MOCK: pass. This is the stateful fake-reviewer seam, unchanged except for correct wiring.
   - ARCH-CONSTRAINTS: pass. The deadlines now come from the shared `HangCeiling` (2 minutes), and the spin bound comes from the same constant.
   - ARCH-SECURE: N/A. The diff handles no secrets and parses no untrusted input; the test writes only under `t.TempDir()`.
   - ARCH-ORDER: pass. The barrier files `ready`/`release` give the test an explicit seam for controlling ordering. Before the fix, the oracle couldn't tell a reviewer that exited early from one that was paused. The minor guard above would close that gap.
   - ARCH-FUNERAL: N/A. The diff creates nothing durable; temporary directories belong to `t.TempDir()`, and the deferred function reaps the CLI.
   - **Docs gate:** N/A. There is no new user-facing or architectural surface, so atlas/README need no changes.

7. **Plan revision recommendations:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: fixture-env-fails-silently
    title: |
      Fake reviewer exits silently when PLANNING_SPINS is unset instead of failing loudly
    detail: |
      planningreview_test.go:73 turns a missing PLANNING_SPINS into an immediate exit 1 after writing ready, which is the intermittent failure this issue chased. Add a guard such as : "${PLANNING_SPINS:?}" (and the same for PLANNING_BARRIER) so a partial env fails every time.
  - id: new
    severity: Minor
    family: done-when-evidence-unrecorded
    title: |
      Log does not record the 10-run and make test results that Done-when requires
    detail: |
      The Plan's Verify item is checked, but the Log has no line with the run count or the make test outcome.
```
