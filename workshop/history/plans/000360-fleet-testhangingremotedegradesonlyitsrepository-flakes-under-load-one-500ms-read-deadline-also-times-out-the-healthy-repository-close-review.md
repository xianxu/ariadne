# Boundary Review — ariadne#360 (whole-issue close)

| field | value |
|-------|-------|
| issue | 360 — fleet: TestHangingRemoteDegradesOnlyItsRepository flakes under load; one 500ms read deadline also times out the healthy repository |
| repo | ariadne |
| issue file | workshop/issues/000360-fleet-testhangingremotedegradesonlyitsrepository-flakes-under-load-one-500ms-read-deadline-also-times-out-the-healthy-repository.md |
| boundary | whole-issue close |
| milestone | — |
| window | 26df5619c1100153b4818c0f2bfae624beedee48..f2013a0178cba078cdbc4f03ea1e534e0ada742a |
| command | sdlc close --issue 360 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T21:43:22-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I'm checking the fleet test and atlas docs, then running the tests with the fix reverted to confirm they fail.

Now reverting the fix pieces to confirm the tests go red.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The fix is correct and does what the issue set out to do. Before, `recordsReadDeadline` was a whole-read context that also covered local git work, and that local work is what timed out the healthy repository. Now it is a remote-only budget (`TrunkFile.BoundRemote` / `runRemote`). That budget is shared across all of one trunk's remote commands, and each repository gets its own. All five remote-contacting git invocations in the tree go through `runRemote`, which matches the plan's count. I checked the regression evidence by reverting each part of the fix. With fleet's `BoundRemote` call removed, `TestHangingRemoteDegradesOnlyItsRepository` fails ("inventory took 1m0.76s with a 500ms read deadline"). With the budget in `runRemote` disabled, `TestRemoteBudgetIsSharedAcrossRemoteCommands` fails and `TestRemoteBudgetBoundsAHangingRemote` hangs until the test timeout. With the fix in place, the tests pass 3/3. Two cheap gaps remain before shipping. The atlas still describes the old whole-read deadline. And the fleet test still gives the healthy repository's real remote calls only 500ms, although the Spec explicitly asked for "real headroom (a few seconds)".

**1. Strengths**
- Root cause found correctly: the deadline was attached to the context that the `tracker.Repository`/`gitx.TrunkFile` carries, so it governed local calls too. The fix moves the bound to the one seam that actually means "the remote doesn't answer" (`trunkfile.go:150-171`).
- The budget is fixed at the first remote command and shared across later ones, so a hanging remote can't multiply it. `TestRemoteBudgetIsSharedAcrossRemoteCommands` checks this in a way that doesn't depend on timing.
- `runRemote` derives its context from `operationContext()`, so caller cancellation still applies. The budget can only tighten the deadline, never extend it.
- The change in behaviour (a hanging remote is now `stale`, not `unknown`) is deliberate. The plan records it, and both the test and `helptext/fleet-inventory.md` were updated. The `unknown`-if-never-fetched path exists at `fleet/issues.go:250`.
- `TestRemoteBudgetDoesNotCountLocalWork` also proves the old shape fails on the same read, so the test shows the contrast rather than only passing.

**2. Critical findings**
None.

**3. Important findings**
- **The atlas wasn't updated for the new deadline meaning** (`atlas/workflow/sdlc-binary.md:365`). It still says "each under a 15s deadline". It should say the 15s bounds only remote contact (ls-remote, fetch), that local git work is not counted (#360), and that a hanging remote now reads as stale when the repository was fetched before. In the window, the other docs restating this are the helptext (updated) and the `recordsReadDeadline` doc comment (updated). The atlas is the only stale copy.
- **The fleet test still races a tight 500ms deadline** (`fleet/recordsdeadline_test.go:108`). The healthy repository's remote calls are real: a `git ls-remote`, and possibly a `fetch`, against a local bare repository, each spawning git plus upload-pack. Under heavy load those can still exceed 500ms; then the healthy repository reads stale instead of present and the test flakes. It is the same failure mode as the original issue, just with fewer process spawns. The Spec asked for "a few seconds" of headroom, and the Log records the same tight-margin problem in the author's own gitx test. 0/100 failures under load is good evidence but not a margin. Fix: `shortDeadline(t, 2*time.Second)`. The hang script sleeps 30s, so the `> 5s` elapsed bound still separates bounded from unbounded. Other tight timing in the window: `TestRemoteBudgetBoundsAHangingRemote` uses 200ms, but there the remote is fake and the hang is meant to happen, so it can't race.

**4. Minor findings**
- ARCH-DRY: `runRemote` (`trunkfile.go:163-170`) copies `run`'s body: the context check, then `runGitInContext`, then the "context error wins" override. A shared `runIn(ctx, env, args...)` that both call would keep one source of truth for the error-precedence rule.
- `remoteDeadline` is set once and never resets unless `BoundRemote` is called again. A `TrunkFile` reused for a second operation would start with its budget already spent. This is safe today because fleet builds a fresh `Repository` per load, but the doc comment should say "per trunk lifetime" explicitly.
- With the budget removed, `TestRemoteBudgetBoundsAHangingRemote` fails as a go-test-timeout panic, because `hangRemote` waits `time.Hour`, rather than as a clean assertion. A delay of about 10s would make the failure clean and quick.
- `TestRemoteBudgetDoesNotCountLocalWork` still passes with the budget disabled, which is expected. Its job is to prove that local work isn't counted, not that the budget exists; the other two tests cover the budget.

**5. Test coverage notes**
- Done-when, "still fails if the stuck repository's deadline handling is removed": confirmed by reverting it (red, 62.7s).
- Done-when, "passes 50x under concurrent load": I didn't re-run this. The Log records 0/100 on the branch against 26/100 on main under load. The 500ms finding above is the remaining risk for that clause.
- Of the five remote commands, the tests exercise only ls-remote (`RemoteExists`) and fetch. The two push paths under the budget have no direct test, which is acceptable because fleet never pushes.

**6. Architectural notes**
- ARCH-DRY: flagged as Minor (`run`/`runRemote` duplication). The `remoteCommand` test helper restates which commands are remote, but it is test-only.
- ARCH-PURE: pass. The budget logic is a thin IO-seam wrapper, and the tests inject latency through the existing `runGitInContext` seam.
- ARCH-PURPOSE: pass. The fix serves the real purpose ("bound a remote that does not answer") instead of just loosening the test deadline. All five remote call sites are covered, and the Spec's per-repository question is answered: each repository gets its own budget.
- Pre-existing and out of scope: `gitDirReader` (`tracker/records.go:297-300`) and `RepositoryFor`'s `RunGit` calls ignore `ctx`. Since the inventory now has no overall deadline, local work is bounded only by the caller's context.

**7. Plan revision recommendations**
None needed. The plan matches the code, including the "five remote-contacting commands" count. If the fleet test deadline is raised, add a Revisions line noting the headroom value.

```findings
findings:
  - id: new
    severity: Important
    family: docs-restate-changed-contract
    title: |
      atlas/workflow/sdlc-binary.md:365 still describes recordsReadDeadline as a whole-read 15s deadline
    detail: |
      Since this change the 15s bounds only remote contact (ls-remote/fetch/push via runRemote); local git work is not counted, and a hanging remote now degrades to stale (unknown only if never fetched). Helptext and the issues.go doc comment were updated; the atlas paragraph was not. It is the only stale restatement in the tree.
  - id: new
    severity: Important
    family: timing-test-margin
    title: |
      Fleet hanging-remote test still gives the healthy repository's real remote calls only 500ms
    detail: |
      recordsdeadline_test.go:108 keeps shortDeadline(500ms). The healthy repository's ls-remote (and possible fetch) against a local bare repo spawns git plus upload-pack, and under load that can exceed 500ms, turning it stale and flaking the test: the same failure mode, reduced. The Spec asked for a few seconds of headroom. Raise it to about 2s; the 30s hang script and the 5s elapsed bound still separate bounded from unbounded. The other timing tests in the window (gitx remotebudget_test.go) already use wide margins or fake remotes.
  - id: new
    severity: Minor
    family: shared-helper-not-extracted
    title: |
      runRemote duplicates run's context-check, call, and context-error-precedence body (ARCH-DRY)
    detail: |
      trunkfile.go:163-170 against trunkfile.go:209-221. Extract runIn(ctx, env, args...) and have run and runRemote both call it.
  - id: new
    severity: Minor
    family: test-failure-mode
    title: |
      TestRemoteBudgetBoundsAHangingRemote fails via test-timeout panic when the budget is removed
    detail: |
      hangRemote(t, time.Hour) means a regression shows up as a go test timeout rather than a clean assertion. A delay of about 10s would make the failure crisp.
  - id: new
    severity: Minor
    family: docs-restate-changed-contract
    title: |
      BoundRemote doc does not say the budget is consumed for the trunk's lifetime
    detail: |
      remoteDeadline is fixed at the first remote command and never reset, so a reused TrunkFile starts its next operation with the budget spent. Safe for fleet, which builds a fresh Repository per load; worth one clause in the trunkfile.go:144 comment.
```

---

## Re-review — 2026-10-13T21:47:53-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 360 — fleet: TestHangingRemoteDegradesOnlyItsRepository flakes under load; one 500ms read deadline also times out the healthy repository |
| repo | ariadne |
| issue file | workshop/issues/000360-fleet-testhangingremotedegradesonlyitsrepository-flakes-under-load-one-500ms-read-deadline-also-times-out-the-healthy-repository.md |
| boundary | whole-issue close |
| milestone | — |
| window | 26df5619c1100153b4818c0f2bfae624beedee48..45da7ab97efc46037c6ca297e935ecb111f27f63 |
| command | sdlc close --issue 360 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T21:47:53-07:00 |
| verdict | SHIP |

## Review

I'm running the tests, then reverting the fix to confirm the fleet test goes red without it.

```verdict
verdict: SHIP
confidence: high
```

All five prior findings are fixed, and I found nothing new. The fix goes after the root cause rather than just widening the test's budget: `recordsReadDeadline` used to cover the whole repository read, and now it is a remote budget (`TrunkFile.BoundRemote`). That budget covers only the five commands that contact the remote (one fetch, two ls-remote, two push), all sent through `runRemote`, and they share it from the first one. I checked by grep that no remote command in `gitx` still bypasses `runRemote`. Both `## Done when` clauses have evidence:
- **Fails without the fix:** I removed the fleet `BoundRemote` call and the fleet test went red (the inventory took 60.7s against the 10s bound). I also turned `runRemote` into plain `run`, and both the hang test and the shared-budget test went red.
- **Passes 50x under load:** I did not re-run this. The issue Log records the measurement (on origin/main, under full-suite load 30–90, the test failed 26 times in its count; this branch passed).

With the fix in place, the gitx budget tests pass, and the fleet test passed 3 of 3 runs.

1. **Strengths**
   - `trunkfile.go:154-167`: the budget's end is fixed at the first remote command and shared after that, so a hanging remote can't multiply the budget by the number of calls. `TestRemoteBudgetIsSharedAcrossRemoteCommands` pins this, and its pass or fail doesn't depend on load.
   - `remotebudget_test.go:36`: the test checks both the new shape (slow local work still succeeds) and the old one (a whole-operation deadline fails the same read). Its fixture also refuses to pass if the local work didn't actually exceed the budget.
   - The doc changes (atlas, helptext, `issues.go`) now state the "stale (unknown if never fetched)" consequence, which matches the changed test assertion at `recordsdeadline_test.go:133`.
   - `runIn` now holds the rule that the context's error wins over git's own error, so `run` and `runRemote` no longer each carry a copy.

2. **Critical:** none.
3. **Important:** none.
4. **Minor:** none new. One design trade-off, not a finding: fleet no longer has a time limit on a repository's local git work. This is deliberate, it is documented in the Plan, and the caller's context still applies.
5. **Test coverage:** the gitx tests inject delay through `runGitInContext`, and the fleet test uses a real hanging ssh transport. The fleet margins are now wide: a 2s budget against a 30s hang, with a 10s limit on elapsed time.
6. **Architectural notes:** for a long-lived `TrunkFile` serving several reads, the budget is spent over the trunk's lifetime. The `BoundRemote` doc now says this; such callers must call it again for a fresh budget.
7. **Plan revisions:** none. The plan matches the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      atlas/workflow/sdlc-binary.md:365-371 now states the remote-only shared 15s budget, local work not counted, stale/unknown degradation.
  - id: BR-2
    disposition: addressed
    note: |
      recordsdeadline_test.go:112 uses shortDeadline(2s); elapsed bound 10s vs 30s hang; reverting BoundRemote turns it red (60.7s).
  - id: BR-3
    disposition: addressed
    note: |
      trunkfile.go:214-223 runIn extracted; run and runRemote both delegate to it.
  - id: BR-4
    disposition: addressed
    note: |
      remotebudget_test.go:87 hangRemote(t, 10s); with runRemote reverted the test fails cleanly at line 97 in 10.2s, no timeout panic.
  - id: BR-5
    disposition: addressed
    note: |
      trunkfile.go:148-150 documents the lifetime budget and calling BoundRemote again to reset; reset implemented by zeroing remoteDeadline.
```
