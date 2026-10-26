# Boundary Review — ariadne#630 (milestone M1)

| field | value |
|-------|-------|
| issue | 630 — Shard timeouts with no failing test are 'incomplete', not failures: retry once, skip the main-baseline diagnosis |
| repo | ariadne |
| issue file | workshop/issues/000630-shard-timeouts-with-no-failing-test-are-incomplete-not-failures-retry-once-skip.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | e19cdb92b2c58b6d454360925f6feaee47e5c7e6..5f778995696dac5a09a060712b68dea378705e6e |
| command | sdlc milestone-close --issue 630 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T11:33:30-07:00 |
| verdict | REWORK |

## Review

Reading the run path and its consumers; next checking baseline diagnosis, immutability/pruning, and running tests.

Checking whether the republish/saved-attempt path can turn an incomplete attempt back into a posted failure.

Running the targeted Go and Python tests now.

Found a FreeBSD build regression (no `hostload_other.go` fallback). Now checking that the guard test fails when its fix is reverted.

```verdict
verdict: REWORK
confidence: high
```

M1's core works. The shard runner prints an alarm marker, the parser reads it, `classifyIncomplete` is pure and has table and fuzz tests, the attempt is saved at v5 with the load, and the typed error stops the main-baseline diagnosis and the train split. All targeted tests pass: `python3 scripts/test-shard.test.py` (10 OK), the Go tests named in the window, `./internal/flakes`, and a 30 s fuzz run of `FuzzClassifyIncomplete`. I reverted the `isFlakeEvidence` change and both `TestSuiteIncompleteIsNotATestResult` and `TestFlakeListIncompleteRunFailures` went red, so that guard is pinned. One finding blocks: the plan's Core concepts table names an entity (`SuiteLogFacts.Alarmed`) that the code doesn't have, and the review rubric treats any table-versus-code contradiction as Critical. The fix is a one-line Revisions entry. Four Important findings are also cheap to fix:
- The commit breaks the FreeBSD build. The base commit builds on FreeBSD.
- `local-check republish` can still post a `failure` status for an incomplete attempt.
- No test pins the "typed runners stay on today's classification" rule.
- Nothing checks the hand-written Go alarm fixture against the real Go toolchain.

### 1. Strengths
- `cmd/sdlc/suiteincomplete.go:278` `classifyIncomplete` is a pure function of `SuiteLogFacts`. It refuses on parse warnings, typed evidence, any failed process without the alarm, and any unattributed reason the alarm doesn't explain. It fails closed in every direction, and the fuzz test's properties hold.
- `teststats_parse.go:745`: a marker that comes before its process summary becomes a parse warning, which then refuses classification. Old logs without markers can never classify as incomplete (tested).
- The v5 version bump has a concrete reason (an older reader would misread the new outcome value). The Revision correctly drops the separate load record in favour of the additive `load_average` field. `decodeSuiteAttempt` (`teststats_store.go:190-193`) keeps it strict: `load_average` is allowed only on v5, must hold three finite non-negative values, and v5 forbids `rerun_of`, `timed_out` and `exit ≤ 0`. The old-reader rule is exercised with `decodeUnsupportedProofAt(…, 4)`.
- The typed-error routing reuses existing early returns: `diagnoseSuiteFailure` passes any error that isn't a `localCheckFailure` straight through, and the train reports "could not run". A baseline on main that is itself incomplete shows up as "baseline on main could not run", not as main-red (`suitebaseline.go:33`).
- `scripts/test-shard.py:875` keys streams by package, so the marker is found however test2json attributes the panic output. Each alarmed package gets its own marker (tested).

### 2. Critical findings
- **The Core concepts table doesn't match the code (plan-table-drift).** The first Pure-entities row names "`suiteAlarm` facts (`SuiteLogFacts.Alarmed`, `SuiteLogFacts.Processes`)", and Task 2 says "Add … `Alarmed []string` to `SuiteLogFacts`". Neither `SuiteLogFacts.Alarmed` nor `suiteAlarm` exists. The alarm data lives in `SuiteProcess.Alarmed` / `.Running` (`teststats_parse.go:725-730`), and the regex is called `shardAlarm`. No Revision covers this. **Fix:** add a Revisions entry and correct the table row. The issue file's M1 Plan bullet also still says "the v5 attempt plus the separate load record", which the first Revision dropped.

### 3. Important findings
- **FreeBSD build broken (ARCH-CONSTRAINTS).** `GOOS=freebsd go build ./cmd/sdlc` fails with `suiteincomplete.go:128: undefined: hostLoadAverage`. The base commit builds on FreeBSD. The package's own convention is a `_other.go` fallback (`queue_boot_other.go`, `tty_other.go`, `localcheck_resource_other.go`). **Fix:** add `hostload_other.go` with `//go:build !darwin && !linux` that returns an error. `hostLoad` already turns that into "host load unavailable".
- **`local-check republish` posts a failure for an incomplete attempt (ARCH-PURPOSE).** `savedAttemptRecord` (`localcheck_republish.go:28-53`) refuses could-not-run (v3) attempts but not `incomplete-timeout`. `selectSavedAttempt` only skips v4. So `sdlc local-check republish --name suite --rev <sha>` turns a v5 incomplete attempt into a record with `Exit 1`, and `postLocalCheckRecord` posts `failure`. That breaks the contract stated at `localcheck_run.go:452` and in the atlas: "nothing is posted, so no status claims the head failed". The merge path (`republishSavedPass`) is safe because it only posts passes. **Fix:** add a `flakes.Kind(a.Outcome) == flakes.IncompleteTimeout` refusal case beside the could-not-run one, with a test.
- **No test for the typed-runner exclusion.** Revision item 6 says "v5 is written only when `d.flakeContract == ""` … A test pins it." No test sets `flakeContract` together with an alarmed log. Today the rule is guarded twice (`localcheck_run.go:202` and `f.Typed.Seen` in the classifier), and removing either guard would turn nothing red. **Fix:** add a case where a typed suite times out with markers, and assert the outcome is not `incomplete-timeout` and the version is 4.
- **No live check against Go's real alarm output (ARCH-MOCK).** The whole feature depends on two things only fixtures cover: Go's `panic: test timed out after D` / `running tests:` text, and how test2json splits it into events. The fixture at `scripts/test-shard.test.py:936` is written by hand, so a Go toolchain change would silently stop the marker. **Fix:** add one test that runs a real `go test -json -timeout 1s` on a temporary package with a sleeping test, passes it through `report()`, and asserts the marker appears.

### 4. Minor findings
- `incompleteRun.Running` and `suiteIncompleteError.Retried` have no consumers in M1. `Running` is parsed but never reported. They are M2 preparation; either wire them in or note them.
- When a baseline on main is incomplete, the nested message reads "baseline on main could not run: … no main baseline was run; rerun sdlc local-check run … <main sha>", which contradicts itself.
- `decodeSuiteAttempt` accepts typed identities in v5 `failures`/`known_flakes` (the `a.Version < 4` gates), though the plan says v5 "carries exactly v2's fields".
- `hostload_darwin.go:79`: the `len(raw) < 16` check assumes the 24-byte LP64 layout. Checking `len(raw) != 24` would be more exact.
- No `internal/flakes` test pins `IncompleteTimeout.Lands()==false` (Task 3 lists one). It holds trivially through the allow-list.
- Weekly stats and lane eligibility count incomplete attempts as slow or failed suites. That is probably the intended measurement, but it isn't stated.

### 5. Test coverage notes
- Classifier table, parse, fuzz, `runLocalCheck` recording, the flake view, merge-gate pass-through, the train's could-not-run path, and a later pass posting are all covered. The guard revert was confirmed red.
- Done-when asks for "a real test failure still diagnoses as now". The non-incomplete branches are covered only through the classifier table plus the existing `TestTypedSuiteBaselineUsesSharedFacts`. Task 5's "new case covers a mixed run" is not present end to end: no test runs a mixed alarmed plus non-alarmed log through `runLocalCheck` and `diagnoseSuiteFailure`.

### 6. Architectural notes for upcoming work
- **ARCH-DRY:** pass. The parser was extended rather than duplicated, and the typed error follows the shape of `suiteEnvironmentError`.
- **ARCH-PURE:** pass. The classifier is pure, the load seam is injected, and the platform IO is thin.
- **ARCH-PURPOSE:** flagged. The republish path is a consumer of the new outcome that hasn't been swept. Also, in M1 alone, a real named failure inside an alarmed process is reported as "not a test result" with no retry. M2 must land before this reaches main, or the Done-when ("a real failure still diagnoses") is violated in the meantime.
- **ARCH-MOCK:** flagged (no live check of the alarm format).
- **ARCH-CONSTRAINTS:** flagged for the FreeBSD build. Runtime cost is negligible: one sysctl read per incomplete run.
- **ARCH-SECURE:** pass. The load is validated on decode. In M1 the classification runs on the fresh log; M2 must re-derive the package selection from the retained, digest-bound log, as the plan states.
- **ARCH-ORDER:** pass for M1, which holds no state between events beyond the immutable attempt. M2's single-retry bound (`incompleteRetried`) and the process-death-between-save-and-retry case need sequence tests.
- **ARCH-FUNERAL:** pass. v5 attempts use the existing attempt and failed-log pruning, and the separate record was dropped.
- **ARCH-EVOLVE:** pass with a note. The version bump is justified and tested against an old-reader rule. That test runs the current code with `readVersion=4`, not a real older binary. Separately, an older sdlc sharing the clone will refuse passes on that head for as long as the v5 record survives, which the plan accepts.
- **README:** no new command, flag or config key, so no README change is needed. The atlas was updated.

### 7. Plan revision recommendations
- Add a Revisions entry: "The alarm facts are per process (`SuiteLogFacts.Processes[].Alarmed/.Running`), with no `SuiteLogFacts.Alarmed`. The regex is `shardAlarm`." Update the Core concepts row and Task 2 to match.
- Mark the "incomplete evidence record" Integration-points row as removed (the first Revision already explains why), and fix the issue's M1 Plan bullet ("plus the separate load record").
- Note that `savedAttemptRecord` refuses `incomplete-timeout`, as one of the consumers the incomplete outcome sweep covers.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: plan-table-drift
    title: |
      Core concepts row names SuiteLogFacts.Alarmed and suiteAlarm, which do not exist
    detail: |
      Plan Pure-entities row 1 and Task 2 claim SuiteLogFacts.Alarmed []string; code puts alarms on SuiteProcess.Alarmed/Running (teststats_parse.go:725-730) and the regex is shardAlarm. No Revisions entry covers it; the issue M1 Plan bullet also still cites the dropped separate load record. Fix with a Revisions entry plus table correction.
  - id: new
    severity: Important
    kind: defect
    family: platform-fallback-missing
    title: |
      FreeBSD (non-darwin/linux) build broken: hostLoadAverage undefined
    detail: |
      GOOS=freebsd go build ./cmd/sdlc fails at suiteincomplete.go:128; base e19cdb92 builds. Add hostload_other.go (//go:build !darwin && !linux) returning an error, per the queue_boot_other.go / tty_other.go convention.
  - id: new
    severity: Important
    kind: defect
    family: new-outcome-consumer-sweep
    title: |
      local-check republish posts a failure status for an incomplete-timeout attempt
    detail: |
      savedAttemptRecord (localcheck_republish.go:28) refuses could-not-run but not incomplete-timeout, and selectSavedAttempt skips only v4, so republish turns the v5 attempt into Exit 1 and posts failure, contradicting the no-status contract at localcheck_run.go:452 and in the atlas. Add a refusal case and a test.
  - id: new
    severity: Important
    kind: defect
    family: plan-claimed-test-missing
    title: |
      No test pins that a typed (flakeContract) suite timeout keeps today's classification
    detail: |
      Plan Revision item 6 says a test pins v5 being written only when d.flakeContract is empty; none exists. The guards at localcheck_run.go:202 and Typed.Seen in classifyIncomplete overlap, so removing either one turns nothing red.
  - id: new
    severity: Important
    kind: defect
    family: external-format-conformance
    title: |
      Go test-alarm panic format and test2json split are modeled only by hand-written fixtures
    detail: |
      scripts/test-shard.test.py hand-writes the panic and running-tests output. Add one test that runs a real go test -json -timeout 1s through report() and asserts the marker, so a toolchain change cannot silently disable classification (ARCH-MOCK live check).
  - id: new
    severity: Minor
    kind: defect
    family: unconsumed-field
    title: |
      incompleteRun.Running and suiteIncompleteError.Retried have no consumers in M1
  - id: new
    severity: Minor
    kind: defect
    family: error-message-context
    title: |
      Incomplete baseline-on-main message says no main baseline was run inside a baseline error
  - id: new
    severity: Minor
    kind: defect
    family: version-envelope-strictness
    title: |
      v5 decode admits typed identities in failures/known_flakes though v5 is meant to carry only v2 fields
  - id: new
    severity: Minor
    kind: defect
    family: plan-claimed-test-missing
    title: |
      Mixed alarmed plus real-failure run is not exercised end to end through diagnoseSuiteFailure
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a0ed87d2-bfb6-43f7-9860-f52905eb6788","repository":"github.com/xianxu/ariadne","issue_id":"000630","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/daa73266b22f5b40","reviewed_head":"5f778995696dac5a09a060712b68dea378705e6e","slot":"ariadne:5","requested_at":"2026-10-25T18:27:52.085911Z"},"phase":"finished","pid":21691,"started_at":"2026-10-25T18:27:52.214531Z","finished_at":"2026-10-25T18:33:29.065523Z","session_id":"2840fc26-bb8d-4233-9906-4e77eb43e61c","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T11:47:05-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 630 — Shard timeouts with no failing test are 'incomplete', not failures: retry once, skip the main-baseline diagnosis |
| repo | ariadne |
| issue file | workshop/issues/000630-shard-timeouts-with-no-failing-test-are-incomplete-not-failures-retry-once-skip.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | e19cdb92b2c58b6d454360925f6feaee47e5c7e6..d81ab02411dfbd8fe96433e634345f6752b4c5b4 |
| command | sdlc milestone-close --issue 630 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T11:47:05-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Verifying the prior findings' fixes now: the build and tests pass. Next, I'll revert each fix and confirm its test goes red.

Checking which callers consume the new incomplete error, especially the train path when main's baseline is incomplete.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

All the Critical and Important findings from earlier rounds (BR-2 through BR-6) are fixed. I checked each fix by reverting it in this checkout and running its test; each test failed without the fix. The two exceptions are BR-3, which I confirmed with a FreeBSD build, and BR-6, whose fix is a new live test that I ran. The FreeBSD build now succeeds. Republish refuses an incomplete-timeout attempt. A typed-contract run never reaches the incomplete classifier. A real `go test -json -timeout 1s` and a real `test2json` shard both produce the marker. The mixed alarm-plus-real-failure run is still diagnosed against main. Nothing blocking is left. Three cheap test gaps remain, all Minor:
- The new "running" part of the report (BR-7) has no test.
- The new message for an incomplete baseline on main (BR-8) has no test.
- The Revisions entry says a classifier test covers the `f.Typed.Seen` guard. That fixture is rejected earlier, by the Unattributed default, so it never reaches the guard. Removing the guard keeps the tests green. BR-1 belongs to M2 and stays open for the issue close.

**Checks I ran (all at the pinned head; the checkout was restored afterwards):**
- `GOOS=freebsd go build ./cmd/sdlc` passes.
- The targeted `go test` run passes.
- `python3 -I scripts/test-shard.test.py` ran 11 tests and all passed; the live alarm test ran rather than being skipped.
- Revert checks:
  - Removing the BR-4 case turned `TestSavedAttemptRecord` red.
  - Removing the BR-5 contract guard turned `TestIncompleteEligible` red.
  - Removing the BR-9 clause turned `TestRunLocalCheckRecordsIncompleteTimeout` red.
  - Removing `f.Typed.Seen` left everything green.

1. **Strengths**
   - `incompleteEligible` (`cmd/sdlc/suiteincomplete.go:34`) moves the eligibility guard out of `runLocalCheck` into a pure function with a test for each guard. That fixes BR-5 properly.
   - `test_real_go_alarm_through_both_runner_paths` (`scripts/test-shard.test.py:80`) is a real live check against the Go toolchain. It covers both `go test -json` and a compiled binary run through `test2json`, and `shard_report_test.go` runs it as part of the Go suite (ARCH-MOCK).
   - The v5 envelope check (`teststats_store.go:190-193`) is strict and each clause has a test: the outcome must match the version, the check must be a test check with exit > 0, `load_average` is allowed only on v5, and v5 rejects known flakes and typed identities.
   - The outcome-consumer sweep in the Revisions entry lists the consumers and how each one is handled. That is the class-level answer that ARCH-PURPOSE asks for.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `suiteincomplete.go:45`: no test reaches `f.Typed.Seen`. The "a typed runner's evidence" fixture (`suiteincomplete_test.go:46`) uses `{"sdlc_unit":1}`, which parses as a malformed unit. Its Unattributed reason is rejected at the `default` branch (line 67) before the guard matters. Fix: use a well-formed typed unit event with no failures, so only `Typed.Seen` can reject it.
   - BR-7 and BR-8: the new report strings have no tests (details in the dispositions).

5. **Test coverage notes**
   - The `republishSavedPass` assertion added to `TestSuiteIncompleteIsNotATestResult` stays green without the BR-4 fix, probably because saved-attempt selection never picks a failed attempt. It is harmless, but `TestSavedAttemptRecord` is the test that actually covers BR-4.
   - The live Go test needs `go` on `PATH` inside `go test`, and skips cleanly without it.

6. **Architecture notes**

   | Principle | Result | Notes |
   |---|---|---|
   | ARCH-DRY | pass | `goIdentities` repeats the check in `classifyIncomplete`'s loop (lines 76-78). Reusing it there is optional. |
   | ARCH-PURE | pass | The classifier, eligibility check and parser are pure; host load is injected through `d.loadAverage`. |
   | ARCH-PURPOSE | pass for M1 | The retry, which is the issue's main purpose, is M2 by plan. BR-1 tracks its cost. |
   | ARCH-MOCK | pass | The live conformance test was added. |
   | ARCH-CONSTRAINTS | pass | The load is recorded, and no new unbounded work was added. |
   | ARCH-SECURE | pass | v5 records are parsed strictly at decode, and `load_average` is validated as finite and non-negative. |
   | ARCH-ORDER | pass for M1 | M1 has no retry state machine yet; the resume branch is M2. |
   | ARCH-FUNERAL | pass | `load_average` is stored on the attempt, which existing pruning removes. The separate record was dropped. |
   | ARCH-EVOLVE | pass | The v5 bump is justified: an older reader must not treat an incomplete run as a pass or a failure. `decodeUnsupportedProofAt` checks the v4-reader rule. |

   For M2: `suiteIncompleteError` has no `errors.As` consumer yet. The train and the merge gate treat it as "could not run" only because it is not a `localCheckFailure`. M2's retry should match on the type explicitly.

7. **Plan revision recommendations**
   - Correct the 2026-10-25 Revisions entry: "plus the classifier's typed-facts case" is not yet a test that covers `Typed.Seen`. Either fix the fixture or reword the entry.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      The retry is M2 scope and is not built in M1. It stays open for the issue close (cost equivalence or operator confirmation of the Done-when).
  - id: BR-2
    disposition: addressed
    note: |
      The Core concepts tables now name SuiteProcess/shardAlarm, incompleteEligible, goIdentities, validLoadAverage, LoadAverage and hostload_other. A 2026-10-25 Revisions entry was added, and the issue M1 bullet now says "v5 attempt with an additive load_average".
  - id: BR-3
    disposition: addressed
    note: |
      hostload_other.go (//go:build !darwin && !linux) was added, and GOOS=freebsd go build ./cmd/sdlc passes at head.
  - id: BR-4
    disposition: addressed
    note: |
      savedAttemptRecord refuses IncompleteTimeout (localcheck_republish.go:32). Removing that case turns TestSavedAttemptRecord red.
  - id: BR-5
    disposition: addressed
    note: |
      TestIncompleteEligible pins the contract guard; removing contract=="" turns it red. The separate Typed.Seen gap is raised as a new Minor.
  - id: BR-6
    disposition: addressed
    note: |
      test_real_go_alarm_through_both_runner_paths runs real go test -json and test2json through report() and asserts the marker. It ran and passed (not skipped).
  - id: BR-7
    disposition: not-addressed
    note: |
      Retried was removed and Running now reaches String() (suiteincomplete.go:92-96), but no test asserts "(running TestSlow)". Add it to the want list at suiteincomplete_test.go:82.
  - id: BR-8
    disposition: not-addressed
    note: |
      The Baseline branch message (suiteincomplete.go:139-141) has no test. Add a diagnoseSuiteFailure case whose main baseline hits the alarm and assert "main's baseline is unknown".
  - id: BR-9
    disposition: addressed
    note: |
      The v5 decode refuses known_flakes and typed identities (teststats_store.go:193). Removing the clause turns TestRunLocalCheckRecordsIncompleteTimeout red.
  - id: BR-10
    disposition: addressed
    note: |
      TestSuiteMixedAlarmAndRealFailureDiagnoses runs the mixed log through runLocalCheck and diagnoseSuiteFailure, and asserts a suite comparison. It passes.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: plan-claimed-test-missing
    title: |
      The classifier's typed-facts case does not pin the f.Typed.Seen guard that the Revisions entry says it pins
    detail: |
      This is the 3rd finding in family plan-claimed-test-missing; earlier rounds fixed instances. Rule: every guard that the plan or a Revisions entry says is pinned needs a test that goes red when that guard alone is removed, with a fixture that reaches the guard and is not rejected by an earlier branch. Instance: the {"sdlc_unit":1} fixture at suiteincomplete_test.go:46 is a malformed unit; its Unattributed reason is rejected at the default branch (suiteincomplete.go:67), and removing f.Typed.Seen at line 45 keeps TestClassifyIncomplete green. Apply the rule to every "pinned" claim in the Revisions entry.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"98c6b230-8e07-4991-b38d-637621735443","repository":"github.com/xianxu/ariadne","issue_id":"000630","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/b4ebf47a64684a7c","reviewed_head":"d81ab02411dfbd8fe96433e634345f6752b4c5b4","slot":"ariadne:5","requested_at":"2026-10-25T18:43:15.358744Z"},"phase":"finished","pid":66690,"started_at":"2026-10-25T18:43:15.45779Z","finished_at":"2026-10-25T18:47:03.653358Z","session_id":"e55cbba0-e701-41b4-8e69-1526cac127ad","outcome":"returned"}] -->
