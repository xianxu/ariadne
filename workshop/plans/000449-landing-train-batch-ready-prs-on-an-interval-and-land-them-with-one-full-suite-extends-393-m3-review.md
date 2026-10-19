# Boundary Review — ariadne#449 (milestone M3)

| field | value |
|-------|-------|
| issue | 449 — landing train: batch ready PRs on an interval and land them with one full suite (extends #393) |
| repo | ariadne |
| issue file | workshop/issues/000449-landing-train-batch-ready-prs-on-an-interval-and-land-them-with-one-full-suite-extends-393.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 9e6849a9525828f036ceb5e909e885fe04de1527..0db9e31fed4559d1859f1640624bbd5118e4c632 |
| command | sdlc milestone-close --issue 449 --milestone M3 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T07:22:34-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reading the plan and issue log for M3, then checking README and the new code paths.

Inspecting the suite/baseline and closePR paths, then running the tests.

Running just the M3-added tests, verbose, with a bounded timeout.

All pass. Now reverting the new grant check to see whether its test fails without it.

Grant fix is pinned (test goes red without it; file restored). Checking the conductor's terminal handling for the new `merge-unconfirmed` phase.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M3 does what it set out to do in the code. A split holds the culprit, lands the clean prefix and sends the remaining cars off at once. A conflicting car is held, red CI on the train PR refuses the whole train, and a granted merge lands alone. A known flake whose rerun passes counts as a pass, `sdlc landing status` reports the last trains, and the new `merge-unconfirmed` terminal phase stops an abort during the merge from claiming the train was abandoned. All 15 targeted tests pass on the pinned head (`go test ./cmd/sdlc/ -run 'TestLandingTrain|TestTrainSuiteOutcome|TestRecentTrainsReport|TestTrainCodeTakesItsGitHubSeam|TestLandingStatus'`, 139s), and so does `./cmd/sdlc/internal/landingline`. I removed the new grant check at `landingtrain.go:665-669` and `TestLandingTrainGrantDepartsAloneWhenNoTrainIsDue` failed ("train PRs [landing-train/3]; the grant did not land alone"); I then restored the file. One fix stands before SHIP: the M3 docs task names README, but README's train paragraph was not updated for any of M3's user-visible behavior.

**1. Strengths**
- `deleteTrainRef` (`landingtrain.go:555`) puts every train-branch delete in one helper: it refuses a non-`landing-train/` name or an invalid tip, then deletes under the repo lock, leased on the tip. It also fixes the earlier unleased delete in `closePR`. This settles a whole class, not one call site.
- `PhaseUnconfirmed` (`trainstate.go:109`, 178-193) adds an explicit state for the uncertain merge outcome instead of folding it into abandoned. The PR is left for the next conductor's reclaim and the cars are released, so each one reads its own PR. Pinned by `TestTrainAbortWhileMergingIsUnconfirmed`.
- The end-to-end split test (`landingtrain_test.go:646`) uses real git and a real failing suite. It checks the culprit is held, the prefix is merged, the car behind the culprit stays open and is admitted next, `Splits==2` (within ⌈log₂3⌉), and that `NextDue` is immediate.
- `recentTrains` (`landingcmd.go:230`) is a pure function, tested directly and with its limit checked.
- Two contract tests now enforce rules that drifted before: one keeps the plan's lists naming every train event and effect, and one keeps train code off the global `ghClient`.

**2. Critical findings**
None.

**3. Important findings**
- **README not updated for M3 (`README.md:491-501`, family `docs-gate-readme`, 2nd finding).** The plan's M3 docs task lists README. The paragraph still ends at "shows the setting, the next departure and each waiting merge's readiness". It says nothing about split/bisect and the held culprit, conflict holds, refusal when main is red or train CI is red, a local-gates-label PR landing alone, a car finishing as a resume, or the last-trains report in `sdlc landing status`. All of these are in `helptext/landing.md` and the atlas. **Rule, not instance:** every user-visible train behavior listed in helptext's `LANDING TRAINS` section must also appear in README's train paragraph. Sweep the whole list in one pass.

**4. Minor findings**
- **Boarding ignores grants (ARCH-PURPOSE).** `Board` only treats `Readiness.Alone` (the local-gates label) as a reason to ride alone. A grant added after the head is admitted but before boarding lets a granted waiter board a train, where another car's red CI can refuse it. The plan said `Alone` covers "when a grant applies".
- **The grant lookup is written out at three sites (ARCH-DRY, family `duplicated-rule`, 4th finding).** `GrantEntries` + `ValidGrants(...)[issue]` appears at `landingqueue.go:50`, `landingqueue.go:178` and `landingtrain.go:665`. Rule: "is issue N granted now" should be one `Line` method that all three call.
- **The seam test checks a fixed file list.** `TestTrainCodeTakesItsGitHubSeam` names four files, so a new `landing*.go` file escapes the check. Glob the landing files instead.
- **`mainIsRed` has no direct test.** `TestTrainSuiteOutcome` pre-sets `c.baseline`, so the baseline comparison inside `mainIsRed` is never run. Its `classifyFailures` path and the "unparsed red main" path are untested.
- **One help-text line is over-long.** `helptext/landing.md` RETENTION: the line ending "…which grants are" runs past the wrap width.

**5. Test coverage notes**
The e2e tests cover split, conflict, red CI and grant. The flake-is-a-pass and main-red mappings are covered only at the unit level. There is no e2e test of an abort during the merge followed by released cars resuming. That is acceptable because each part is pinned on its own: the pure transition, and BR-16's "a car with no word" from M2.

**6. Architectural notes**
- **ARCH-DRY:** flag, the grant lookup (Minor above). The delete helper is a good consolidation.
- **ARCH-PURE:** pass. `recentTrains` is pure; the grant check is thin glue.
- **ARCH-PURPOSE:** flag, README (Important) and boarding grants (Minor). Live conformance moving to the pilot follows the hub's Q7 answer, and the pilot plan is recorded in the issue.
- **ARCH-MOCK:** pass. The GitHub fake gains train CI checks and accepts a status on any commit it holds, which the bisect needs. Conformance is the pilot's evidence, per Q7.
- **ARCH-CONSTRAINTS:** pass. The split test confirms extra suites stay within ⌈log₂ n⌉.
- **ARCH-SECURE:** pass. `deleteTrainRef` validates the branch prefix and the tip OID. An unreadable grant log gives no priority, the same as at admission.
- **ARCH-ORDER:** pass. The uncertain merge outcome is now an explicit terminal state, not collapsed into abandoned.
- **ARCH-FUNERAL:** pass. A train branch closed without a PR is now deleted, and `outcomes/` is swept along with `ready/`.
- **ARCH-EVOLVE:** pass. `merge-unconfirmed` is a new value of the existing `result` field, which is additive.

**7. Plan revision recommendations**
None needed for code drift: the "M2 as built" revision matches `trainstate.go`, and a test now enforces it. Optionally, add an "M3 as built" note that `Readiness.Alone` comes only from the local-gates label and grants are handled at admission and conduct time, or fix Board to match the original sketch.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: docs-gate-readme
    title: |
      README train paragraph not updated for M3 surface (splits, holds, refusals, alone, per-train report)
    detail: |
      This is the 2nd finding in family docs-gate-readme. The plan's M3 docs task names README, but README.md:491-501 still ends at the readiness line and omits the bisect split with the held culprit, conflict holds, main-red and train-CI-red refusal, local-gates-label lands alone, resume on a merged PR with no word, and the last-trains report in sdlc landing status. All of these are in helptext/landing.md and atlas. Rule: every user-visible train behavior in helptext LANDING TRAINS must also appear in README's train paragraph; sweep the whole list.
  - id: new
    severity: Minor
    kind: defect
    family: duplicated-rule
    title: |
      The "is issue granted now" lookup is written out at three sites
    detail: |
      This is the 4th finding in family duplicated-rule. GrantEntries plus ValidGrants(...)[issue] appears at landingqueue.go:50, landingqueue.go:178 and landingtrain.go:665. Rule: grant status is one Line method that every caller uses, so an unreadable log means the same thing everywhere.
  - id: new
    severity: Minor
    kind: defect
    family: grant-bypass-boarding
    title: |
      Board ignores grants, so a waiter granted after admission can board a train
    detail: |
      Readiness.Alone is set only from the local-gates label (landingcar.go:73). A grant added between the head's admission and boarding lets the granted car ride a train that another car's red CI can refuse, contrary to "lands alone, outside the train".
  - id: new
    severity: Minor
    kind: defect
    family: contract-test-hand-list
    title: |
      TestTrainCodeTakesItsGitHubSeam checks a hand-maintained list of four files
    detail: |
      A new landing*.go file escapes the seam check. Glob the landing files, excluding landing.go, the wiring site.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"3a4528e5-92fc-4783-922e-e83d4b9c8f04","repository":"github.com/xianxu/ariadne","issue_id":"000449","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/f7d3e9b01f1175d5","reviewed_head":"0db9e31fed4559d1859f1640624bbd5118e4c632","slot":"ariadne:2","requested_at":"2026-10-19T14:06:38.81747Z"},"phase":"finished","pid":95470,"started_at":"2026-10-19T14:06:39.089237Z","finished_at":"2026-10-19T14:22:34.130251Z","session_id":"6a98b83d-0ceb-4c8e-9d3c-9af442be4c13","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T08:22:43-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 449 — landing train: batch ready PRs on an interval and land them with one full suite (extends #393) |
| repo | ariadne |
| issue file | workshop/issues/000449-landing-train-batch-ready-prs-on-an-interval-and-land-them-with-one-full-suite-extends-393.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 9e6849a9525828f036ceb5e909e885fe04de1527..b1ab3a4cec0a2314e38c8a344fee2a9956cee0c8 |
| command | sdlc milestone-close --issue 449 --milestone M3 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T08:22:43-07:00 |
| verdict | SHIP |

## Review

Prior-round fixes all check out on inspection. Next I'm running the landing test suites and reverting the BR-28 fix to confirm its test fails without it.

The changed landing tests all pass (222s), and the BR-28 regression test fails when the fix is reverted. Last step: checking the Board loop ordering.

```verdict
verdict: SHIP
confidence: high
```

All four findings from the previous round are fixed, and I found nothing new. The fix commit `b1ab3a4ce` is small and targeted. The README now lists every train behavior that the landing help describes. Grant status is now answered by one `Line.GrantsNow` method, and every call site that only needs current grants uses it. `Board` now skips a granted waiter, and a pure unit test pins that. The GitHub-seam contract test now globs the landing files instead of checking a hand-written list. I reverted the BR-28 change locally and its new `TestBoard` subtest failed (`cars [1 2 3]`), then restored it; the checkout is clean afterwards. `landingline` passes, and all tests in `landingtrain_test.go`, `landingcar_test.go` and `landingcmd_test.go` pass (222s).

1. **Strengths**
   - `landingline/log.go:86-95`: `GrantsNow` uses the line's own clock (`l.o.Now()`), so admission, leave and conduct all judge grants at the same time source. The two remaining direct `ValidGrants` callers are deliberate. `landingcmd.go:167` needs the skipped-entry count for its status output. `suitepriority.go` is #452 code that is outside this window.
   - `landingline/train.go:227-231`: a granted waiter counts toward `LeftReady`, so the next train leaves at once. Grant ordering admits that waiter next, and `maybeConductTrain` sends it serially, so it lands alone. It also does not stop boarding for the cars behind it, which is correct because it never takes a train slot.
   - `landingcar_test.go:425-445`: the glob test has a `checked < 5` guard, so a pattern that silently stops matching fails instead of passing with nothing checked.
   - The README train section (`README.md:491-520`) now matches helptext LANDING TRAINS item for item: vouching, departure, conducting, holds and bisect, flake / main-red / CI-red refusals, riding alone, resume, and the last-trains status.

2. **Critical:** none.
3. **Important:** none.
4. **Minor:** none new. `go test ./cmd/sdlc/ -run 'Landing|Train|Car|Grant'` hit Go's default 600s test timeout because the pattern matches far more than the landing tests, and it panicked at that limit. The same tests pass when selected by name, so this is not a defect.
5. **Test coverage:** the BR-28 behavior is pinned at the pure level (`TestBoard` "a granted waiter never boards"), and I confirmed that test fails without the fix. The grants that `Board` reads at conduct time come from the same `GrantsNow` result as the head's own serial check, so the path is reachable. There is no integration test where a granted waiter sits behind a conducting head. That is acceptable because the logic lives in pure `Board`.
6. **Architecture:**
   - **ARCH-DRY:** pass. The grant lookup is consolidated in `GrantsNow`.
   - **ARCH-PURE:** pass. `Board` stays pure, and grants are passed in as data.
   - **ARCH-PURPOSE:** pass. The README sweep covers the full helptext list, not just the items BR-26 named.
   - **ARCH-MOCK:** pass. The seam contract is enforced by the glob test.
   - **ARCH-CONSTRAINTS:** pass. The grant log is read once per conduct, and it was already read on this path.
   - **ARCH-SECURE:** pass. An unreadable grant log means no priority, at every site.
   - **ARCH-ORDER:** pass. Grant handling is explicit in the `Board` transition, and it fails closed (no priority).
   - **ARCH-FUNERAL:** pass. The fix creates no new artifacts.
   - **ARCH-EVOLVE:** pass. No persisted format changed.
7. **Plan revisions:** none needed.

```findings
dispose:
  - id: BR-26
    disposition: addressed
    note: |
      README.md:491-520 now lists vouching, departure, conducting, holds/bisect, flake/main-red/CI-red refusals, alone (grant and local-gates label), resume, and the last-trains status, matching helptext/landing.md LANDING TRAINS one-for-one.
  - id: BR-27
    disposition: addressed
    note: |
      Line.GrantsNow (log.go:89) is used at landingqueue.go:50, landingqueue.go:178 and landingtrain.go:665; the remaining direct ValidGrants use at landingcmd.go:171 needs the skipped count for status.
  - id: BR-28
    disposition: addressed
    note: |
      Board takes grants and passes over a granted waiter (train.go:227); TestBoard granted-waiter subtest fails when the check is disabled (verified by revert) and passes with it.
  - id: BR-29
    disposition: addressed
    note: |
      TestTrainCodeTakesItsGitHubSeam now globs landing*.go, excluding landing.go and tests, with a checked<5 guard against an outdated glob; passes.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"bacf2150-dcf8-4516-94c0-3df12edaa04b","repository":"github.com/xianxu/ariadne","issue_id":"000449","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/acacff334121b161","reviewed_head":"b1ab3a4cec0a2314e38c8a344fee2a9956cee0c8","slot":"ariadne:2","requested_at":"2026-10-19T15:07:20.518007Z"},"phase":"finished","pid":55761,"started_at":"2026-10-19T15:07:20.793915Z","finished_at":"2026-10-19T15:22:30.264475Z","session_id":"11e611e1-1520-4aaa-9ee9-0af9e6ef36f2","outcome":"returned"}] -->
