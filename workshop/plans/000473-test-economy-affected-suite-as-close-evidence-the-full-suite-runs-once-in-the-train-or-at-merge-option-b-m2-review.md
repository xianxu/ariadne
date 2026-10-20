# Boundary Review — ariadne#473 (milestone M2)

| field | value |
|-------|-------|
| issue | 473 — test-economy: affected-suite as close evidence; the full suite runs once, in the train or at merge (option B') |
| repo | ariadne |
| issue file | workshop/issues/000473-test-economy-affected-suite-as-close-evidence-the-full-suite-runs-once-in-the-train-or-at-merge-option-b.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 2f1b32e90074ea0442e41a628a7d8e76eceeb865..d6309691af55cb39765a55623ed01f5ab31212e2 |
| command | sdlc milestone-close --issue 473 --milestone M2 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-20T02:29:43-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking whether `local-check republish --name affected` works for a saved known-flake pass, since the guard's error message tells the operator to run it.

Found one real defect so far (republishing a saved affected known-flake pass can never succeed). Now running the touched packages' tests to confirm the head is green.

Tests pass. Now a quick revert check: removing the `Subset` guard should turn its regression test red.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M2 does what it says on the main points. A train car now boards on `affected` in place of `suite`. Merge without a train still requires `suite`: the stand-in is set only in `carCIGateFacts`, and the train always runs `RunSuite` on its prefix (`landingtrain.go:176,237`), so the D3 claim that "the train runs the one full suite" holds. Affected runs now get the known-flake rerun, the success guard covers both checks, an affected pass can never confirm a fail→pass, and affected runs have their own test-stats row. The targeted tests pass at head (`go test ./cmd/sdlc/ -run 'CIGate|Car|Affected|…'` ok, 248s; `internal/flakes` and `internal/affected` ok). I reverted two of the fixes in the checkout: removing `!b.Subset` turns `TestSubsetPassNeverConfirms` red, and reverting `isFlakeEvidence` to `isSuiteRun` turns `TestTestRunGuardSpansSuiteAndAffected` red. So those tests do catch the bug. One thing blocks SHIP. M2 is the first point where an `affected` record can have the outcome `KnownFlakePassed`. When such a record fails to post, the error tells the operator to run `sdlc local-check republish --name affected`, and that command can never succeed. The plan also hasn't been updated where the code deliberately departed from it.

1. **Strengths**
   - `decideCIGate`'s stand-in (`cigate.go:123-133`) asks for the stand-in rather than `suite` when neither is recorded, so the car's refusal names the run that actually helps (`affected`, which falls back to `suite` itself when it must). `unrecordedRequired` returns both names, so the ancestor walk carries an affected record across docs-only commits. Pinned at `landingcar_affected_test.go:89-103`.
   - `landingcar_affected_test.go` runs the car gate in production form against a real repo and a fake GitHub. It covers both pair#723 shapes from the plan's Revisions, and a merge on the same head that still refuses without `suite`.
   - The decision to make D4 stricter is sound and recorded in the plan. The success guard covering both checks (`localcheck_republish.go:155`) closes the green-wash hole that #394 forbids. `Attempt.Subset` lives in the one shared `Confirmed` rule, so both test-stats and flake registration inherit it (ARCH-DRY).
   - `aggregateSuiteAttempts` becomes a thin wrapper over `aggregateCheckAttempts`, so there's no copied aggregate. The weekly consumers stay on `isSuiteRun`.

2. **Critical findings:** none.

3. **Important findings**
   - **The republish path can't find the `affected` check.** `cmd/sdlc/localcheck_cmd.go:255-258`: when `runLocalCheckRepublish` finds a saved `KnownFlakePassed` or `KnownFlakeAccept` attempt, it calls `declaredCheck(ctx, git, f.Name)`. That only searches `local-checks.json` plus the default suite (`localcheck_cmd.go:290-307`), and `parseLocalChecks` reserves the name `affected` (`localchecks.go:121`), so the lookup always fails with "origin/main declares no local check "affected"".
     - **How it's reached:** this became reachable in M2. `localcheck_run.go:184` now runs `knownFlakeStep` for `affected`. If that pass then fails to post (a GitHub outage, #431), `localcheck_run.go:230` tells the operator to run `republish --name affected`, which refuses. The only way out is a full rerun, which is exactly what the saved-result path exists to avoid.
     - **The rule:** every consumer that turns a check *name* back into a `localCheck` must resolve `affected` from `affected.json` (via `affectedLocalCheck`/suite policy), not from `local-checks.json`. I grepped the name-keyed consumers of `suiteCheckName`, `isSuiteRun` and `declaredCheck`; `declaredCheck` is the only one where M2 makes `affected` reachable.
     - **Fix sketch:** give `declaredCheck` an `affected` branch that reads `readAffectedConfig(origin/main)` plus the declared suite and returns `affectedLocalCheck(cfg, Selection{}, suite)`, or have revalidation take the suite's `knownFlakePolicy`. Add a republish test with a saved affected `KnownFlakePassed` attempt.
     - Architecture marker: ARCH-PURPOSE (shadow-sweep of the consumers that "affected is a test run" implies).
   - **The plan still claims work the code deliberately didn't do.** This is the 2nd finding in family `plan-tracks-code`.
     - D2 and Task 8 still say the publish gate's full-suite instruction (`publishgate.go:302-308`) will name `affected`, and Task 8 lists `helptext/merge.md:140-160`. Neither file changed.
     - The atlas (`ci-merge-check.md`, "Conflict resolution") explains why: a resolution changes main's code, so the publish gate keeps asking for the full suite. That's correct, but the plan still claims the opposite.
     - **The rule:** any departure from a plan's Decisions or Files list ships with a `## Revisions` entry in the same commit. Sweeping D1–D6 and Tasks 6–8 against the diff turns up exactly these two departures.

4. **Minor findings**
   - `landingcar.go:121`: a malformed `affected.json` on main makes every car's CI gate error, which wedges the train. The spec says a refusal should read "affected unavailable; run the suite". Dropping the stand-in (require `suite`) with a warning would be the conservative fallback. This is the 3rd finding in family `unclassified-path-fails-closed`. The rule covering all three: an affected-suite input that can't be trusted degrades to requiring the full suite, never to a hard stop or an empty selection. The impact is small because the local-gates guard protects the file on main.
   - Task 6 Step 1 planned to assert that "the runner is called with suite" at merge. The test instead asserts that the error names `--name suite` with a nil runner (`landingcar_affected_test.go:96`). That's equivalent in practice, but weaker.
   - `TestAffectedStatsAreTheirOwnRow` uses only passing attempts. A failing suite attempt followed by an affected pass on the same SHA would also pin that `ConfirmedFlakeAttempts` for the suite row is unchanged.

5. **Test coverage notes**
   - Stand-in decisions: a pure table in `cigate_test.go`, plus an integration test through `productionCarGates`.
   - Flake behaviour: covered through the real `runLocalCheck` with a script rerun.
   - Gap: `local-check republish --name affected` (the Important finding above).

6. **Architecture notes**
   - **ARCH-DRY:** pass. `isTestCheck` is the one predicate, and the aggregate wrapper is shared.
   - **ARCH-PURE:** pass. The stand-in logic lives in pure `decideCIGate`, and IO stays in `carCIGateFacts`.
   - **ARCH-PURPOSE:** flag (the republish sweep above).
   - **ARCH-MOCK:** pass. Real git plus the existing stateful `ciFakeGH`.
   - **ARCH-CONSTRAINTS:** pass. One extra `git ls-tree`/`show` per vouch.
   - **ARCH-SECURE:** pass. The config is read from main and parsed strictly; see the Minor on how a bad config fails.
   - **ARCH-ORDER:** pass. No new state carried between events; facts are recomputed on every vouch.
   - **ARCH-FUNERAL:** pass. No new durable family; affected attempts go into the existing store under its retention.
   - **ARCH-EVOLVE:** pass. `RepoStats.Check` is additive with `omitempty`, `Attempt.Subset` is derived in memory, and weekly snapshots are unchanged.
   - **Docs gate:** pass. The atlas, `landing.md`, `local-check.md` and `AGENTS.base.md` are updated. M2 adds no new flags, so no README change is needed.

7. **Plan revision recommendations**
   - Add `## Revisions — 2026-10-20 (M2 Task 8)`: the publish gate's `--fixed-then-ship --full-suite` instruction stays on `suite`, because a conflict resolution changes main's code. `merge.md` stays as it is, because merge without a train still requires `suite`. Amend D2 to match.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: name-keyed-check-resolution
    title: |
      local-check republish --name affected always refuses for a saved known-flake pass: declaredCheck cannot resolve the reserved name
    detail: |
      runLocalCheckRepublish (localcheck_cmd.go:255-258) calls declaredCheck for KnownFlakePassed/Accept attempts, but declaredCheck (localcheck_cmd.go:290-307) only searches local-checks.json plus the default suite, where affected is reserved (localchecks.go:121). M2 newly lets affected runs reach KnownFlakePassed (localcheck_run.go:184), and a failed post then tells the operator to run this exact command (localcheck_run.go:230), so the only recovery is a full rerun. Rule (ARCH-PURPOSE sweep): every consumer that resolves a check name back to a localCheck must resolve affected from affected.json with the suite's policy. Fix declaredCheck accordingly and add a republish regression test.
  - id: new
    severity: Important
    kind: defect
    family: plan-tracks-code
    title: |
      Plan D2/Task 8 still claim publishgate.go and merge.md wording changes that the code deliberately did not make
    detail: |
      2nd in family plan-tracks-code. The rule: any departure from a plan's Decisions or Files list ships with a ## Revisions entry in the same commit. The publish gate's --full-suite instruction stays on suite (the atlas explains why: a conflict resolution changes main's code), and merge.md is untouched. D1-D6 and Tasks 6-8 swept against the diff show no other departures. Add the Revisions entry and amend D2.
  - id: new
    severity: Minor
    kind: defect
    family: unclassified-path-fails-closed
    title: |
      A malformed affected.json on main makes every car's CI gate error instead of degrading to requiring suite
    detail: |
      3rd in family. Rule covering all three: an affected-suite input that cannot be trusted degrades to requiring the full suite, never to a hard stop or an empty selection. carCIGateFacts (landingcar.go:121) returns an error, which wedges the train. Dropping the stand-in with a warning matches the spec's "affected unavailable; run the suite".
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"62303091-c0a4-40ac-99d5-ad201c918fd0","repository":"github.com/xianxu/ariadne","issue_id":"000473","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/89369333f40a517f","reviewed_head":"d6309691af55cb39765a55623ed01f5ab31212e2","slot":"ariadne:2","requested_at":"2026-10-20T09:22:21.514451Z"},"phase":"finished","pid":82302,"started_at":"2026-10-20T09:22:21.654275Z","finished_at":"2026-10-20T09:29:43.137115Z","session_id":"a890961a-646b-4097-807b-a6708375d274","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T02:45:04-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 473 — test-economy: affected-suite as close evidence; the full suite runs once, in the train or at merge (option B') |
| repo | ariadne |
| issue file | workshop/issues/000473-test-economy-affected-suite-as-close-evidence-the-full-suite-runs-once-in-the-train-or-at-merge-option-b.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 2f1b32e90074ea0442e41a628a7d8e76eceeb865..22460760f3d5f188c542a29f95358a8ef7899a8a |
| command | sdlc milestone-close --issue 473 --milestone M2 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-20T02:45:04-07:00 |
| verdict | SHIP |

## Review

Checking the three prior findings now: I'm running the tests and reverting each fix to confirm its regression test fails without it.

Found a possible gap in how the train finds the suite. Tracing it before running the tests.

The train turns out to be safe: its CI wait still requires a recorded `suite` on the prefix, so it can't land on `affected` alone. Now running the targeted tests.

That broad run only hit Go's 10-minute test timeout because my regex was too wide; the targeted M2 tests all pass. Next I'll revert the BR-9 and BR-11 fixes to confirm their tests fail without them.

Both fixes are confirmed: each test fails without its fix. Last, I'm checking whether anything reads `test-stats` JSON and would break on a second row per repo.

```verdict
verdict: SHIP
confidence: high
```

M2 does what D3/D4 say and the pinned head closes all three prior findings. A train car can now board on an `affected` record. The stand-in only exists in the car's CI gate; merge still requires `suite`. The train can't get around that either: `waitCI` on the train prefix still asks for a recorded `suite`. The test-run rules also cover `affected` now: it gets the known-flake rerun, the success guard covers both checks, and a subset pass never counts as a confirmed fail→pass. For BR-9 and BR-11 I reverted each fix in the disposable checkout; its regression test failed, and passed again once restored. BR-10 has its `## Revisions` entry. I found one new Minor issue (a zero-run suite row in `test-stats`). Nothing blocks.

**1. Strengths**
- **The stand-in is car-only.** `StandIn` exists only in `carCIGateFacts` (`cmd/sdlc/landingcar.go:117`). `decideCIGate` stays pure (`cmd/sdlc/cigate.go:123-133`). The tests at `cmd/sdlc/cigate_test.go:103-109` check both sides, including "merge: affected does not stand in".
- **One place turns the name `affected` into a check.** That is `affectedCheckDecl` (`cmd/sdlc/localcheck_affected.go:72`). Both the run path and `declaredCheck` use it, so the republish path can no longer drift from the run path (ARCH-DRY).
- **The guard is stricter than the plan.** A failed `suite` blocks an `affected` pass on the same head, and the reverse too (`cmd/sdlc/localcheck_republish.go:155`). This stops a passing subset run from hiding a failed full run; the Revisions entry records the decision.
- **Subset passes don't confirm anything.** `flakes.Attempt.Subset` is set in one place (`cmd/sdlc/teststats.go:239`) and checked inside `Confirmed`, so the flake-registration path and the stats path both get it.
- **The car tests are real end-to-end tests.** `cmd/sdlc/landingcar_affected_test.go` uses a real git repo and the stateful `ciFakeGH` to drive `productionCarGates` and `runCIGate`, including pair#723's docs-only shape.

**2. Critical findings**
None.

**3. Important findings**
None.

**4. Minor findings**
- `aggregateCheckAttempts` groups rows by repo before filtering by check (`cmd/sdlc/teststats.go:128`). A repo whose window has only `affected` attempts therefore gets a `suite` row with `Runs: 0`, which didn't exist before M2. The `test-stats` command only skips empty rows for `affected` (`cmd/sdlc/teststats_cmd.go:154`), not for `suite`.
- The runner-lookup refusal at `cmd/sdlc/cigate.go:648` can't currently be reached. Only cars set `StandIn`, and cars pass `run=nil`. It's a harmless defensive guard with no test.
- `carCIGateFacts` drops the `readAffectedConfig` error without a warning (`cmd/sdlc/landingcar.go:126`). It fails in the safe direction, but the operator only learns the file is broken by running `affected`.

**5. Test coverage notes**
- I ran the ten M2 tests named in the window (`TestDecideCIGateLocalChecks`, the `TestCar*` tests, `TestAffectedKnownFlakeReruns`, `TestTestRunGuardSpansSuiteAndAffected`, `TestSubsetPassNeverConfirms`, `TestAffectedStatsAreTheirOwnRow`, `TestDeclaredCheckResolvesAffected`, `TestCarWithBrokenAffectedConfigNeedsSuite`). All pass.
- `internal/flakes` and `internal/affected` pass.
- My first, broader regex run of `./cmd/sdlc` hit Go's default 10-minute timeout. That was my run being too wide, not an M2 failure, but it means I did not run the whole package in one go.
- Revert checks: with `declaredCheck`'s `affected` branch disabled, `TestDeclaredCheckResolvesAffected` failed. With `carCIGateFacts` changed to return the parse error, `TestCarWithBrokenAffectedConfigNeedsSuite` failed.
- Gap: no test runs `runTestStats` with `affected`-only attempts. That test would catch the zero-run row above.

**6. Architectural notes**
- **ARCH-DRY: pass.** There is one resolver (`affectedCheckDecl`) and one test-run predicate (`isTestCheck`).
- **ARCH-PURE: pass.** `decideCIGate` and `flakes.Confirmed` stay pure. `carCIGateFacts` is a thin IO layer.
- **ARCH-PURPOSE: pass.** Swept every site that turns a name into a check (`selectLocalChecks`, `declaredCheck`, the `cigate.go:648` runner lookup, the train's suite lookup): each handles `affected` correctly or never sees it.
- **ARCH-MOCK: pass.** Tests use the stateful fake GitHub and real git.
- **ARCH-CONSTRAINTS: pass.** Each vouch adds one `ls-tree` plus one `show`.
- **ARCH-SECURE: pass.** An untrusted `affected.json` falls back to requiring `suite`.
- **ARCH-ORDER: pass.** The guard is tested in both orders (suite fails then affected passes, and the reverse).
- **ARCH-FUNERAL: pass.** M2 creates no new kind of durable artifact.
- **ARCH-EVOLVE: pass.** `RepoStats.Check` is `omitempty` and additive. `test-stats --json` can now list two rows per repo; nothing in this repo reads that JSON.
- One thing for later: the train's suite lookup (`cmd/sdlc/landingtrain.go:231-241`) doesn't apply `withDefaultSuiteContext`. In a repo that relies on the implicit Makefile suite, the train would stall at its CI wait rather than land. That's safe, but it now matters more, because M2 relies on the train's prefix suite. This predates M2 (ariadne itself declares `suite` explicitly).

**7. Plan revision recommendations**
None. The 2026-10-20 entries in the plan's `## Revisions` match the code.

```findings
dispose:
  - id: BR-9
    disposition: addressed
    note: |
      declaredCheck resolves affected via affectedCheckDecl (localcheck_cmd.go:301-314); TestDeclaredCheckResolvesAffected goes red with the branch disabled (verified by revert).
  - id: BR-10
    disposition: addressed
    note: |
      Plan Revisions 2026-10-20 (M2 Task 8, BR-9/BR-10) amends D2 and Task 8 Files; publishgate.go and merge.md are unchanged in the diff, matching it.
  - id: BR-11
    disposition: addressed
    note: |
      carCIGateFacts drops the stand-in on readAffectedConfig error (landingcar.go:126); TestCarWithBrokenAffectedConfigNeedsSuite goes red when the error is returned (verified by revert).
findings:
  - id: new
    severity: Minor
    kind: defect
    family: report-row-requires-runs
    title: |
      test-stats emits a zero-run suite row for a repo whose window holds only affected attempts
    detail: |
      aggregateCheckAttempts groups by repo over isFlakeEvidence, which now includes affected, before filtering rows by check (teststats.go:128,157), so the suite aggregate returns Runs 0 for such a repo and runTestStats appends it unconditionally (teststats_cmd.go:150-152). Filter suite rows with Runs > 0 as the affected row is, and add a runTestStats test with affected-only attempts.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"150b7f6e-2da5-482a-bd39-20dfe5171031","repository":"github.com/xianxu/ariadne","issue_id":"000473","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/a5c7ac163e197ce0","reviewed_head":"22460760f3d5f188c542a29f95358a8ef7899a8a","slot":"ariadne:2","requested_at":"2026-10-20T09:32:28.682985Z"},"phase":"finished","pid":28315,"started_at":"2026-10-20T09:32:28.878294Z","finished_at":"2026-10-20T09:45:03.672704Z","session_id":"496da64a-c039-4d5d-b570-4cd323b4fdb3","outcome":"returned"}] -->
