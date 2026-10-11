# Boundary Review — ariadne#303 (whole-issue close)

| field | value |
|-------|-------|
| issue | 303 — Catch the slot and its dependencies up to main before new work |
| repo | ariadne |
| issue file | workshop/issues/000303-catch-up-before-new-work.md |
| boundary | whole-issue close |
| milestone | — |
| window | 15b365d31f1d17f8217476a7baf62fb741f47549..208cd44509e6df3b84ffd75c411de91aed65cead |
| command | sdlc close --issue 303 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-11T09:08:32-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Still waiting on the targeted start-plan/catch-up test run. While it finishes I'm checking `onRest` and how `weave refresh` behaves.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The design delivers the issue's purpose. `start-plan` checks the slot and its dependency repos against origin/main on the resting branch, and hands the actual catch-up to `weave refresh`. It never fast-forwards a dependency itself, which is the right call. Refuse-vs-warn uses the same freshness rule as close's integrate gate, and a successful catch-up stops for a re-run because the running `sdlc` predates it. I ran the targeted tests (`go test ./ -run 'Catch|StartPlan|Tolerat'`) and they pass in 183s. Two things should be fixed before shipping:

- **Partial refresh is treated as plain failure.** `weave refresh` keeps whatever Git updates it completed when it errors, including when every repo moved and only the compile failed. `catchUpBeforeWork` treats any error as "nothing moved". In that case it prints a misleading warning and goes on to plan with the old binary, which breaks the spec's rule that a catch-up always stops for a re-run.
- **Spec drift on unchecked repos.** A repo with no `origin`, or a dependency checkout that is missing, is silently treated as current. The spec says these must warn as unchecked.

**1. Strengths**
- ARCH-DRY: the catch-up goes entirely through `weave refresh`, behind the `weaveRefresh` seam (`cmd/sdlc/catchup.go:24`). The "Stale" decision reuses `freshness.LoadLocation`/`Cutoff`/`Check` and the existing `integrateNow` clock, so `start-plan` cannot refuse anything close would accept.
- The wiring is tightly scoped: it only runs when on the resting branch (`env.onRest()`), and it runs before `preparePlanningBranch` and before the card write (`cmd/sdlc/startplan.go:317-322`). That gives the "nothing branched or started" guarantee, and the tests check it.
- TestMain installs a weave stub that fails loudly (`testmain_test.go:163`), so no test can reach the real installed binary.
- Each fetch has a time limit (`catchUpFetchTimeout`), and a fetch or `rev-list` failure produces a "could not check … NOT verified as current" warning rather than a false pass.
- The tests cover both tolerated cases: code that landed after the cutoff, and docs that landed before it.

**2. Critical findings:** none.

**3. Important findings**
- **Partial refresh progress is collapsed into failure** (`cmd/sdlc/catchup.go:58-88`, ARCH-ORDER).
  - `cmd/weave/internal/refresh/refresh.go:83,228` shows that a failed refresh keeps the updates it finished: "completed Git updates remain", "compile failed; Git updates remain".
  - Example: every repo fast-forwards, then the compile fails. `freshness.Check` now finds HEAD current, so nothing is Stale. Start-plan prints "what these lack landed after … or is docs only" (false) and plans with the pre-refresh `sdlc` on top of a failed compile.
  - Fix sketch: record each behind repo's HEAD before calling weave. After an error, if any HEAD moved, refuse with "refresh partially applied; resolve weave's error, run `weave refresh`, then re-run", and never take the warn path. Add a test whose fake weave does the fast-forward and then returns an error.
- **No-origin and missing checkouts silently count as current** (`cmd/sdlc/catchup.go:96-98`, spec drift).
  - The spec says: "A repo with no `origin`, or a fetch that fails, is reported as unchecked: a warning, never 'up to date'."
  - The code returns `0, ""` whenever `git remote get-url origin` fails for any reason. That includes a missing dependency path, since `substrateChain` also returns paths that don't exist on disk, and a directory that isn't a Git repo.
  - Fix: return a `why` (for example "no origin" or the git error) so the existing warning fires. If silence for repos without an origin is wanted, the spec needs a Revisions entry saying so. Either way, add a test.

**4. Minor findings**
- The probe runs `fetch origin main` and relies on the configured refspec to update `refs/remotes/origin/main`. weave uses the explicit `+refs/heads/main:refs/remotes/origin/main`, and the probe should match it so a narrow or odd refspec can't leave the ref stale.
- The re-run hint only reconstructs `--issue N`; any other flags the operator passed are dropped.
- The spec scopes this to `:1+` slots, but `onRest()` also fires in the primary checkout, where weave refresh includes shared peers. That is probably fine, but neither the spec nor the help text says so.
- ARCH-MOCK: the weave seam is faked per test, but nothing checks that the real `weave refresh` is invoked correctly (argv and working directory). Low risk because both are in this repo.

**5. Test coverage notes**
- Covered: catching up both repos and stopping for a re-run; the re-run being idempotent; refusing on a Stale checkout while surfacing weave's output; warning when the lag is tolerated.
- Not covered: refresh partially applied, then error (Important #1); no-origin, missing dependency, and fetch-failure warnings (Important #2); weave missing from PATH.
- The Done-when fixture shows the wiring and ordering, but the fast-forward itself is done by the fake. That is acceptable because weave owns the transaction.

**6. Architectural notes**
- ARCH-DRY pass. ARCH-PURE pass: thin IO glue, and the logic delegates to the pure `freshness`. ARCH-PURPOSE pass. ARCH-MOCK pass with the minor conformance gap. ARCH-CONSTRAINTS pass: one bounded, sequential fetch per repo.
- ARCH-SECURE pass: argv arrays, and paths come from the existing bounded deps reader.
- ARCH-ORDER flag: see Important #1. The "refresh error" outcome stands for three different states (nothing moved, some repos moved, all moved but the compile failed), and the code only models the first.
- ARCH-FUNERAL pass: only remote-tracking refs are written.

**7. Plan revision recommendations**
- If silent treatment of repos without an origin is intended, add a "## Revisions" entry changing the spec's no-origin rule. Otherwise fix the code.
- Add a "## Revisions" note: "a refresh that partially applies (Git updates remain) is treated as a catch-up that requires a re-run, never as a tolerated lag."

```findings
findings:
  - id: new
    severity: Important
    family: uncertain-outcome-collapsed
    title: |
      A failed weave refresh that already advanced checkouts is treated as "nothing moved" and may plan with the stale binary
    detail: |
      weave refresh keeps completed Git updates on error, including when every repo moved and only the compile failed (refresh.go:83,228). catchUpBeforeWork then runs freshness on the now-current HEADs, finds nothing Stale, prints a false "landed after cutoff or docs only" warning, and plans with the pre-refresh sdlc. Fix: snapshot HEADs before weave; if any moved, refuse and ask for a re-run. Add a fake-weave test that fast-forwards and then errors.
  - id: new
    severity: Important
    family: unchecked-reported-as-current
    title: |
      A repo with no origin, or a missing/non-git dependency path, is silently treated as current, contrary to the spec
    detail: |
      catchup.go:96-98 returns (0, "") on any `remote get-url origin` failure. The spec requires a no-origin repo to be reported as unchecked (a warning). substrateChain also returns dependency paths that don't exist, which pass silently. Return a why so cwarn fires (or revise the spec), and add a test.
  - id: new
    severity: Minor
    family: refspec-consistency
    title: |
      The probe fetch relies on the configured refspec rather than weave's explicit +refs/heads/main:refs/remotes/origin/main
  - id: new
    severity: Minor
    family: test-coverage-gap
    title: |
      No tests for fetch-failure warnings, weave missing from PATH, or the real weaveRefresh argv/dir
```

---

## Re-review — 2026-10-11T09:28:40-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 303 — Catch the slot and its dependencies up to main before new work |
| repo | ariadne |
| issue file | workshop/issues/000303-catch-up-before-new-work.md |
| boundary | whole-issue close |
| milestone | — |
| window | 15b365d31f1d17f8217476a7baf62fb741f47549..34545b375f2727108d7b89aea95ba1c8c158fe1d |
| command | sdlc close --issue 303 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-11T09:28:40-07:00 |
| verdict | SHIP |

## Review

Now revert the fixes to confirm the regressions go red.

Both prior-round fixes go red when I revert them, so the regression tests are real. I'm waiting for the full test suite to finish.

```verdict
verdict: SHIP
confidence: high
```

All four findings from round 1 are fixed, and the two behavior fixes have regression tests that I checked by reverting each fix.

- **Partial refresh (BR-1):** `catchUpBeforeWork` now records each checkout's HEAD before running `weave refresh`. If any HEAD moved, it stops and asks for a re-run, whether or not refresh failed. It no longer reads a failed refresh as "nothing moved". With that guard disabled, `TestStartPlanStopsWhenAFailedRefreshAdvancedACheckout` fails.
- **Unchecked checkouts (BR-2):** a repo with no origin, a dependency that is declared but not on disk, and a failed fetch now each print a "NOT verified as current" warning. With the fix reverted, `TestCatchUpWarnsUncheckedCheckouts` fails on the missing warnings.

The catch-up, start-plan and seam tests pass at the pinned head. The full `cmd/sdlc` test suite hit the 10-minute background limit and was stopped, so I have no full-suite result. The checkout was restored to `34545b37` and is clean. All that remains are two Minor comment and help-text gaps.

**1. Strengths**
- `catchup.go:73-89` works out what refresh did from the checkouts themselves, not from weave's exit status. It covers a refresh that worked, failed part-way, or was cancelled. The rule is also recorded in `workshop/lessons.md`.
- `startplan.go:433-482` adds `substrateWalk` while `substrateChain` keeps its old behavior for existing callers, so the dependency walk is written once (ARCH-DRY).
- `runInstalledWeave` is now a named function behind the `weaveRefresh` seam. A test runs it with a fake weave on PATH, checks the arguments and working directory, and checks the error when weave is missing.
- There is a new safeguard for the case where refresh reports success but a checkout is still behind. It refuses instead of planning on stale code.

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- `catchup.go:116-118`: the `behindOrigin` comment still says "a checkout with no origin has nothing to trail", but the code now reports "no origin remote". The file header's "or nothing moves" also suggests refresh rolls back completely, which the BR-1 fix shows it does not. Both comments should be updated.
- `helptext/start-plan.md` and the atlas paragraph don't mention two behaviors: the "could not check … NOT verified as current" warning for missing, no-origin or unreachable checkouts, and the stop when a refresh fails after it has already advanced a checkout.

**5. Test coverage**
- With each fix reverted, the BR-1 and BR-2 tests fail. With the fixes in place they pass.
- Fetch failure, a missing weave, and the arguments and directory weave runs with are now covered (BR-4).
- The tests run against real git repos and a fake weave behind the production seam.

**6. Architecture**

| Marker | Result | Note |
|---|---|---|
| ARCH-DRY | pass | |
| ARCH-PURE | pass | The decision logic is a small shell around the git and weave seams; `freshness` stays pure. |
| ARCH-PURPOSE | pass | Every checkout on the dependency path is probed or warned about. |
| ARCH-MOCK | pass | Weave is faked through a seam that TestMain also stubs. Git runs for real on temp repos. |
| ARCH-CONSTRAINTS | pass | Each fetch is capped at one minute and they run one after another, which is fine for a short dependency chain. |
| ARCH-SECURE | pass | Paths from `construct/deps` are resolved and only passed as git `-C` arguments; no secrets are involved. |
| ARCH-ORDER | pass | An uncertain refresh outcome is no longer treated as success or failure; the code reads the observed HEADs. |
| ARCH-FUNERAL | pass | Nothing durable is created beyond remote-tracking refs, which git manages. |

**7. Plan revisions:** none needed. The Log already records the round-1 changes.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      HEADs snapshotted pre-weave; any move stops for re-run (catchup.go:73-89). Reverting the guard turns TestStartPlanStopsWhenAFailedRefreshAdvancedACheckout red.
  - id: BR-2
    disposition: addressed
    note: |
      no-origin returns "no origin remote"; substrateWalk surfaces absent deps as unchecked. Reverting both turns TestCatchUpWarnsUncheckedCheckouts red.
  - id: BR-3
    disposition: addressed
    note: |
      fetch now uses explicit +refs/heads/main:refs/remotes/origin/main (catchup.go:128).
  - id: BR-4
    disposition: addressed
    note: |
      TestCatchUpWarnsUncheckedCheckouts covers fetch failure; TestRunInstalledWeaveRunsRefreshInRoot covers argv/dir and weave missing from PATH.
findings:
  - id: new
    severity: Minor
    family: stale-doc-comment
    title: |
      behindOrigin and file-header comments contradict the round-1 behavior
    detail: |
      catchup.go:116-118 still says a no-origin checkout has nothing to trail (it now warns as unchecked), and the header's "or nothing moves" implies full rollback, which BR-1 showed is false after a compile failure.
  - id: new
    severity: Minor
    family: docs-surface-gap
    title: |
      start-plan help and atlas omit the unchecked-checkout warning and the stop after a partial refresh
```

---

## Re-review — 2026-10-11T09:33:51-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 303 — Catch the slot and its dependencies up to main before new work |
| repo | ariadne |
| issue file | workshop/issues/000303-catch-up-before-new-work.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1829f2fa25299d3f803b0ad6c63bb9faf9abb1e1..1ec8628e5997a7db47ca4040c903a7ca3dff6423 |
| command | sdlc close --issue 303 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-11T09:33:51-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

Both open findings are fixed, and this fix window adds no Critical or Important problems.

- **Comments (BR-5):** the `behindOrigin` comment now says a checkout with no origin is reported unchecked, never current. The file header now says a failed `weave refresh` keeps the updates it already completed. Both match the code at `cmd/sdlc/catchup.go:76-88` and `:123-125`.
- **Help text and atlas (BR-6):** both now describe the stop after a partial refresh. They also list the three kinds of checkout the probe cannot read, each matched to its code path:
  - absent from disk: `substrateWalk`, line 51
  - no origin: line 124
  - failed fetch: line 129

The issue file changes only the status, the card mirror hash, the flow, `actual_hours` and Log entries. `## Spec`, `## Revisions` and `## Done when` are untouched, so the scope did not change. There were no conflict resolutions to judge. The changes are comments, help text and atlas text only, so no new regression test was needed. I did not re-run the test suite: only comments and prose changed, and `gofmt -l` is clean.

1. **Strengths**
   - The new comments describe what the code actually does now. For example, the header's "keeps the updates it completed if a later step fails" is the reason the code reads HEADs before and after the refresh (`catchup.go:73-80`).
   - The atlas paragraph (`atlas/workflow/sdlc-binary.md:1107-1110`) names the code that does each thing: HEAD comparison, `substrateWalk`, and `freshness.Check`.
   - The help text's three unchecked cases match the code paths exactly.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - The new prose is not wrapped. `catchup.go:5` and one added line in `cmd/sdlc/helptext/start-plan.md` (CATCH-UP section) run to about 130–150 characters, while the lines around them wrap at about 80. That is purely cosmetic.

5. **Test coverage:** no runtime behaviour changed. The existing tests from round 2 (`TestStartPlanStopsWhenAFailedRefreshAdvancedACheckout` and `TestCatchUpWarnsUncheckedCheckouts`) still cover the behaviour these docs describe.

6. **Architecture** (checked against this fix window only)
   - ARCH-DRY: pass. Nothing is duplicated.
   - ARCH-PURE: pass. No code changed.
   - ARCH-PURPOSE: pass. Both places BR-6 named (help text and atlas) were updated, and so were both comments BR-5 named. The `stale-doc-comment` and `docs-surface-gap` families each still have only their original finding: the fixes covered every instance, and I found nothing left to repeat.

7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      catchup.go:4-5 header now says refresh keeps completed updates on a later failure; catchup.go:118-120 says no-origin is reported unchecked, matching line 124.
  - id: BR-6
    disposition: addressed
    note: |
      helptext/start-plan.md CATCH-UP section and atlas/workflow/sdlc-binary.md:1107-1110 now describe the stop after a partial refresh and the absent/no-origin/failed-fetch unchecked warnings, matching catchup.go:51,124,129 and 76-88.
findings:
  - id: new
    severity: Minor
    family: prose-line-wrap
    title: |
      New comment and help-text lines are not wrapped to the surrounding width
    detail: |
      catchup.go:5 and one added line in helptext/start-plan.md (CATCH-UP section) run to about 130-150 characters, while neighbouring lines wrap at about 80. Cosmetic only.
```
