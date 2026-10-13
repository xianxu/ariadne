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

---

## Re-review — 2026-10-13T15:44:07-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 343 — weave: TestCompileRecoversAfterProcessDeathDuringGeneration flakes ~1 in 20 on main ('owned stage survived retry') |
| repo | ariadne |
| issue file | workshop/issues/000343-weave-testcompilerecoversafterprocessdeathduringgeneration-flakes-1-in-20-on-main-owned-stage-survived-retry.md |
| boundary | whole-issue close |
| milestone | — |
| window | 7926a8ae93b9eb36d4d3c138e1adda34bc95c504..06c9afd08bbc9ca7b05e54601cb285e558076a13 |
| command | sdlc close --issue 343 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T15:44:07-07:00 |
| verdict | SHIP |

## Review

The failures come from the harness environment: the git ceiling and temp-dir settings. I'm retrying with a temp dir outside the checkout.

Both tests pass 5 out of 5 runs. Next I'm checking that the refactored end-to-end test still fails when the Reclaim fix is reverted.

```verdict
verdict: SHIP
confidence: high
```

Both earlier Minor findings are fixed, and the fixes add no new problems. For BR-1, the duplicated killed-owner setup is now two shared helpers in `cmd/weave/dynamic_test.go`. `killOwnerMidGeneration(t, tail)` at line 407 builds the fixture, swaps the marker, kills the child, and restores the marker. `compileLeavesNoOwnedStage(t, leaf)` at line 436 compiles and checks that no stage directory survived. The two tests now differ only in the marker tail, which is what the finding asked for. For BR-2, the latency cost is now written down in two places. One is the code comment at `cmd/weave/internal/staging/stage.go:88-91`. The other is the atlas at `atlas/workflow/weave.md:354-355`. Both match the code: `Settle = time.Second` at `lease.go:28`, and `Reclaim` waits on stages one after another. The window does not touch Spec, Revisions or Done-when, and it has no conflict resolutions to judge.

1. **Strengths**
   - **Refactor checked by running it.** I ran both refactored tests 5 times with a temp dir outside the checkout and the git ceiling variable unset; they passed every time. Then I temporarily put back the old single `Exclusive(stage)` call in `Reclaim`. `TestCompileReclaimsStageOfAStillExitingGenerator` failed 3 out of 3 runs ("owned stage survived retry"). So the regression test still catches the bug after the refactor. I restored the file afterwards and the tree is clean.
   - **Behaviour unchanged.** The refactor keeps the original test's behaviour. The order changed slightly: the "dead generator published" check now runs after the marker is restored. That is harmless, because the marker lives in `base` and the checked body lives in `leaf`.
   - **Docs match the code.** The cost note in the comment and the atlas agree with `AwaitExclusive` and the `Reclaim` loop.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings:** none new. The "Rare test flake" note from the earlier close review was never a recorded finding, and this window does not touch it.

5. **Test coverage notes**
   - Both refactored tests pass.
   - The end-to-end test fails when the fix is reverted.
   - In this checkout, the tests only fail under the review harness's `GIT_CEILING_DIRECTORIES` and temp-dir setting. They pass with those removed, so the failures come from the environment, not the code.

6. **Architectural notes**
   - **ARCH-DRY:** pass. The copied test setup is now one shared helper.
   - **ARCH-PURE:** pass. Only test helpers and comments changed.
   - **ARCH-PURPOSE:** pass. BR-1's finding said the two tests were the only copies in the window, and the helper now covers both. BR-2's latency bound is documented both at the code site and in the atlas.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Shared killOwnerMidGeneration(t, tail) + compileLeavesNoOwnedStage(t, leaf) at dynamic_test.go:407/436 now cover both tests; the reverted fix still turns the e2e test red (3/3).
  - id: BR-2
    disposition: addressed
    note: |
      Cost written down in the stage.go:88-91 comment and atlas/workflow/weave.md:354-355; matches Settle=1s (lease.go:28) and the serial Reclaim loop.
```
