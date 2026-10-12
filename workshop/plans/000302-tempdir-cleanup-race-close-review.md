# Boundary Review — ariadne#302 (whole-issue close)

| field | value |
|-------|-------|
| issue | 302 — Test TempDir cleanup race: directory not empty under sharded make test |
| repo | ariadne |
| issue file | workshop/issues/000302-tempdir-cleanup-race.md |
| boundary | whole-issue close |
| milestone | — |
| window | bf837e55b7cfb4af873529f7a744cc4c8cefabbe..6caa63c639fc0859603049140327749f441ee879 |
| command | sdlc close --issue 302 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T12:40:11-07:00 |
| verdict | SHIP |

## Review

Now checking whether the atlas claim "every context-bound git runner kills the whole process group" holds up.

Running the new tests, then reverting the group kill to check that the regression test fails without it.

Regression test confirmed sensitive (fails with group kill reverted; restored). Now checking the atlas package-list claim.

```verdict
verdict: SHIP
confidence: high
```

This range does what the issue asked. The Log names the writer with a reproduction: a cancelled `git` whose local-transport far end (receive-pack and its hook) keeps writing into the origin. The source fix had already landed in #286 M1. This range adds regression tests that pin it for both context-bound push runners, and closes a measured gap in `QuietBackgroundGit` with a `GIT_TEMPLATE_DIR` config, because git strips `GIT_CONFIG_*` on the receive side. I checked the regression test by reverting the fix. With the `processgroup.Terminate` Cancel removed from `cmd/sdlc/trackerenv.go:114`, `TestTrackerGitCancelEndsTheReceiveSide` fails ("the receive side outlived the cancelled push: pid 15601 still running after 2m0s"). With the line restored, it passes. `TestQuietBackgroundGitReachesTheReceiveSide` passed on this machine without skipping, so its baseline (a push does start receive-side maintenance) held and the assertion is real. Only Minor findings remain.

1. **Strengths**
   - `cmd/sdlc/internal/testfix/blockingorigin.go:22`: the blocking pre-receive hook writes its pid atomically (tmp then `mv`). It gives up on its own after twice `HangCeiling`, so if the bug comes back the test fails instead of hanging, and the hook doesn't run forever.
   - `cmd/sdlc/internal/testfix/realgit_test.go:65`: the test first checks that a plain push does start maintenance, and skips if this git never does. That means it can't pass for the wrong reason.
   - The Log evidence is careful. It notes that Go's RemoveAll reports the deepest directory it failed to remove, which rules out `maintenance.lock` as the observed writer instead of assuming it.
   - Every `QuietBackgroundGit` caller handles the new cleanup return (4 TestMains). There are no stale callers of the old signature.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **The atlas's "every context-bound git runner" claim is too strong** (`atlas/workflow/sdlc-binary.md:1568`). `gitx.TrunkFile.HistoryStream` (`cmd/sdlc/internal/gitx/trunkfile.go:483`) uses `exec.CommandContext` with the default Cancel and a direct `Process.Kill`. It is a read-only `git log` with no far end, so it causes no race, but the sentence is literally false.
     - Family in this window: that one sentence.
     - Elsewhere in the tree (not sdlc): `cmd/weave/internal/acquire/git.go:50` and `cmd/weave/environment.go:20`.
     - Fix: narrow the claim to runners that push or fetch, or that spawn hooks or transports.
   - **The two cancel tests are near-copies (ARCH-DRY).** `cmd/sdlc/internal/gitx/cancel_test.go:14-36` and `cmd/sdlc/trackerenv_cancel_test.go:15-35` differ only in the runner call. Both could become one `testfix` helper, e.g. `AssertCancelEndsReceiveSide(t, push func(ctx, root, origin string) error)`. That would also make the atlas's "a new one must too" rule cheaper to follow.

5. **Test coverage**
   - The gitx test was not reverted here. The Log says it was checked the same way, and its structure is identical to the cmd/sdlc test, which I did revert.
   - The "`make test` ×5" clause relies on the Log: 0/5 cleanup races, with 2 runs hitting an unrelated `TestPlanningReviewCompetingResults` failure that was reported separately. I did not re-run it.
   - The template test checks both states: baseline maintenance starts, and the quieted push starts none.

6. **Architecture**
   - **ARCH-PURE:** pass. All changes are test infrastructure, with no business logic mixed into IO.
   - **ARCH-DRY:** flagged as Minor (the duplicated cancel tests above).
   - **ARCH-PURPOSE:** pass. The writer is identified, and both runners that push are pinned.
   - Note for later: `cmd/weave` tests (e.g. `cmd/weave/internal/refresh/refresh_test.go:52`) push into bare origins without receive-side quieting, and they can't import `cmd/sdlc/internal/testfix`. If a cleanup race shows up there, the template helper would need to move to a shared package.

7. **Plan revisions:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: doc-claim-overreach
    title: |
      atlas says every context-bound git runner kills its process group; gitx HistoryStream does not
    detail: |
      atlas/workflow/sdlc-binary.md:1568 claims every context-bound git runner uses processgroup.Terminate, but cmd/sdlc/internal/gitx/trunkfile.go:483 (HistoryStream, read-only git log) uses CommandContext's default Cancel plus a direct Process.Kill. It is harmless since git log has no far end. Narrow the sentence to runners that push or fetch (or spawn hooks or transports). Only instance in the window; weave's acquire/git.go:50 and environment.go:20 are outside the sdlc doc's scope.
  - id: new
    severity: Minor
    family: duplicated-test-scaffold
    title: |
      ARCH-DRY: gitx and cmd/sdlc cancel-ends-receive-side tests are near-identical
    detail: |
      cmd/sdlc/internal/gitx/cancel_test.go:14-36 and cmd/sdlc/trackerenv_cancel_test.go:15-35 differ only in the runner call. Extract a testfix helper (e.g. AssertCancelEndsReceiveSide(t, push func(ctx, root, origin string) error)) so each new context-bound runner adds one line, as the atlas rule asks.
```

---

## Re-review — 2026-10-12T12:49:15-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 302 — Test TempDir cleanup race: directory not empty under sharded make test |
| repo | ariadne |
| issue file | workshop/issues/000302-tempdir-cleanup-race.md |
| boundary | whole-issue close |
| milestone | — |
| window | bf837e55b7cfb4af873529f7a744cc4c8cefabbe..47a4e4636218dcd3b3dba7a160a0773a03ea48f5 |
| command | sdlc close --issue 302 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T12:49:15-07:00 |
| verdict | SHIP |

## Review

Checking the remaining packages for pushes, then running the new tests.

```verdict
verdict: SHIP
confidence: high
```

This change ships. Issue #302 asked for the late writer to be found and fixed where it starts, and the diff does that. The writer is receive-pack, the far end of a local push, which kept writing after a cancelled git returned. The process-group kill from #286 M1 is the actual fix, and this change adds the regression tests that pin it. It also closes a measured gap: git strips the `GIT_CONFIG_*` environment from the receive side of a local push, so the change also sets quiet config through `GIT_TEMPLATE_DIR`. Both prior findings are fixed and I checked them against the code. I ran the three new tests at the pinned head and all pass: `TestQuietBackgroundGitReachesTheReceiveSide`, `TestGitCancelEndsTheReceiveSide` and `TestTrackerGitCancelEndsTheReceiveSide`. The quiet-git test ran its assertion; it did not take its skip branch. So this git does start receive-side maintenance at baseline, and the test is not vacuous. Nothing blocks.

1. **Strengths**
   - `testfix/blockingorigin.go:24-52`: the pre-receive hook records its pid, then blocks for a fixed time. It gives up after twice `HangCeiling`, which is longer than `AwaitExit` waits. A regression therefore fails the test rather than passing it, and the hook can't hang forever.
   - `testfix/realgit_test.go:48-73`: the test first proves the problem exists on this git (a plain push does start `maintenance run --auto`), and skips otherwise. Only then does it assert the fix, so the test can't pass vacuously.
   - `realgit.go:53-60`: the doc comment explains why two mechanisms are needed: git strips the environment config on a local transport, but config stored in the repository survives.
   - Every package whose tests push into a bare origin now calls `QuietBackgroundGit` in `TestMain` and runs its cleanup after `m.Run`. The one other push I found, in reviewbox `checkout_test.go:98`, is expected to fail, so it needs no hook.

2. **Critical:** none.

3. **Important:** none.

4. **Minor:** none new. One note for the record, not a finding: if a test binary dies before its `cleanupQuietGit()` runs, the template directory is left in `$TMPDIR`. It is a few bytes and the OS temp sweep removes it.

5. **Test coverage**
   - The two cancel tests now go through one shared helper, `AssertCancelEndsReceiveSide`, and each caller is a single closure.
   - I did not revert the group kill myself to re-run the sensitivity check. Each run waited 150 s for a slot on the shared test gate, and this round's fixes were a test refactor and a doc edit, not a behaviour change. The issue Log records the check: with the group kill removed, the test fails with "pid … still running after 2m0s".

6. **Architecture**
   - **ARCH-DRY: pass.** The duplicate test scaffold from BR-2 is now one helper.
   - **ARCH-PURE: pass.** The change is test infrastructure only.
   - **ARCH-PURPOSE: pass.** The writer is identified with a reproduction, and the fix is at the source. The atlas rule says any new runner of this kind adds one test.
   - **ARCH-MOCK: pass.** The tests use real git against local bare origins, which is how this repo already works.
   - **ARCH-CONSTRAINTS: pass.** Every wait is bounded by `HangCeiling`.
   - **ARCH-SECURE: pass.** The template directory is created with `MkdirTemp`, and no credentials are involved.
   - **ARCH-ORDER: pass.** The pid handshake is explicit, the hook blocks until cancel, and the cancel ordering is reproducible.
   - **ARCH-FUNERAL: pass.** The template directory is removed after `m.Run`, and the hook stops itself on a timer.

7. **Plan revisions:** none needed.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      atlas/workflow/sdlc-binary.md:1568-1572 now scopes the rule to runners that push, fetch, or spawn hooks or transports, and exempts read-only git log. Checked: both such runners use processgroup.Terminate (trackerenv.go:114, trunkfile.go:87); trunkfile.go:483 is the exempt git log.
  - id: BR-2
    disposition: addressed
    note: |
      Extracted testfix.AssertCancelEndsReceiveSide (blockingorigin.go:56). gitx/cancel_test.go and trackerenv_cancel_test.go are now single closures, and both pass at head.
```
