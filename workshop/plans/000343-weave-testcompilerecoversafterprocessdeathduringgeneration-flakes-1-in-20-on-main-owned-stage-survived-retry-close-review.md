# Boundary Review — ariadne#343 (whole-issue close)

| field | value |
|-------|-------|
| issue | 343 — weave: TestCompileRecoversAfterProcessDeathDuringGeneration flakes ~1 in 20 on main ('owned stage survived retry') |
| repo | ariadne |
| issue file | workshop/issues/000343-weave-testcompilerecoversafterprocessdeathduringgeneration-flakes-1-in-20-on-main-owned-stage-survived-retry.md |
| boundary | whole-issue close |
| milestone | — |
| window | f2266a7eea4af6b71969c7b6c02363ed815689c8..c2585f059c31350ef226d24e7dc7828e9245e005 |
| command | sdlc close --issue 343 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T15:41:43-07:00 |
| verdict | SHIP |

## Review

The `cmd/weave` package failed. I'm checking which test failed, and whether it also fails at base.

The remaining failures come from the review harness's `git init` guard, not from the code. Next I'm reverting the fix and re-running the tests to check that they catch the bug.

```verdict
verdict: SHIP
confidence: high
```

The fix targets the confirmed root cause. When the owner is dead, `staging.Reclaim` used to make one non-blocking `flock` attempt and give up. It now waits up to a bounded settle for the lease, through a new shared `staging.AwaitExclusive`. The same helper replaces the hand-written retry loop that `finishProducers` used during cancellation (ARCH-DRY). I checked every claim in Done-when myself:

- **Fix reverted, tests fail:** with `Reclaim` changed back to a single `Exclusive`, both new `staging` unit tests fail, and the new end-to-end test fails 3 out of 3 runs.
- **Fix in place, tests pass:** the original flaky test and the new end-to-end test passed 50 runs each with no failures, and the `cmd/weave` lifetime tests (cancellation, parent death) pass.

Some other `cmd/weave` and `internal/acquire` tests fail in this checkout. The cause is the review harness, not the change: its `git init` guard ("outside the reviewer's directories", sdlc #339) and `GIT_CEILING_DIRECTORIES` block the test fixtures. Nothing blocks SHIP. The findings below are minor.

1. **Strengths**
   - `cmd/weave/internal/staging/lease.go:24-42`: one `Settle` constant and one `AwaitExclusive`. `Reclaim` and `finishProducers` now share the bound instead of each keeping its own 1s/10ms loop. `finishProducers` behaves exactly as before: error wrapping is unchanged, and an error other than `ErrInUse` still returns immediately.
   - `cmd/weave/internal/staging/stage.go:88-90`: the change is small and keeps the documented policy that a live orphan keeps its stage. Nothing gets signalled, so there's no new risk from PID reuse.
   - `cmd/weave/internal/staging/settle_test.go` tests both sides of the bound in-process, with no subprocess timing races. One test checks the lease is released within the settle; the other checks a held lease is kept after waiting at least `Settle`. Both fail when the fix is reverted.
   - `cmd/weave/dynamic_test.go:460`: the end-to-end test deliberately widens the race window: it closes stdout/stderr, then sleeps 0.1s while still holding fd 3. That is 10x under the settle, and it reproduces the real mechanism deterministically.
   - The atlas (`atlas/workflow/weave.md:350-354`) is updated, and its claims match the code.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - **ARCH-DRY:** `TestCompileReclaimsStageOfAStillExitingGenerator` (`dynamic_test.go:460-495`) copies about 25 lines of setup from `TestCompileRecoversAfterProcessDeathDuringGeneration` (`dynamic_test.go:405-440`). Both build the fixture, swap the marker, kill the helper child, restore the marker, run, and check that no stage directory survived. The two copies differ only in the marker tail. A shared helper such as `killOwnerMidGeneration(t, tail string) (leaf, marker string, good []byte)` would remove the copy. These two tests are the whole family in this window.
   - **Cost of a live orphan:** each dead-owner stage still held by a live orphan now adds up to 1s to every `Reclaim`, one after another. `Reclaim` is called from `publication.go:34`, `plan/staged.go:40`, `release/main.go:65` and `acquire/stage.go:9`. This is acceptable because the case is rare, but it is worth a line in the atlas or a comment.
   - **Rare test flake:** `deadOwnerStage` relies on the reaped PID of `true` staying unused. If the PID were reused, `TestReclaimPreservesALiveOrphanAfterTheSettle` would fail its elapsed-time assertion. This is very unlikely.

5. **Test coverage notes:** both outcomes the fix can produce are covered at unit level (lease released, so the stage is reclaimed; lease held, so the stage is kept after the bound) and end to end (lingering generator). `Remove` still uses a non-blocking `Exclusive`. That's correct, because its callers run after `finishProducers` has already confirmed the lease is clear.

6. **Architectural notes**
   - **ARCH-DRY:** pass for the production code (the loop is merged into one helper). Flagged only for the copied test setup above.
   - **ARCH-PURE:** pass. The change stays at the IO boundary (flock and clock), and `AwaitExclusive` is a thin wrapper around a retry.
   - **ARCH-PURPOSE:** pass. It fixes the root cause the Spec asked for (wait before reclaiming), not the test's timing. Every non-blocking `Exclusive` that runs while a producer may still be exiting now goes through `AwaitExclusive`: those are `Reclaim` and the post-kill path in `finishProducers`. The first `Exclusive` in `finishProducers` and the one in `Remove` are intentionally probes or post-condition checks.

7. **Plan revision recommendations:** none. The plan still matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: test-setup-duplication
    title: |
      New e2e test copies the killed-owner setup of TestCompileRecoversAfterProcessDeathDuringGeneration
    detail: |
      ARCH-DRY. dynamic_test.go:460-495 repeats about 25 lines from dynamic_test.go:405-440 (fixture, marker swap, helper-child kill, marker restore, run, stage-dir assertion); only the marker tail differs. Extract a helper such as killOwnerMidGeneration(t, tail). These two tests are the only instances in the window.
  - id: new
    severity: Minor
    family: undocumented-latency-bound
    title: |
      Reclaim now spends up to Settle (1s) per dead-owner stage still held by a live orphan
    detail: |
      The stages are waited on one after another inside the Reclaim loop (stage.go:88), and Reclaim is called from publication.go:34, plan/staged.go:40, release/main.go:65 and acquire/stage.go:9. This is acceptable because the case is rare, but worth a sentence in the atlas or code comment.
```
