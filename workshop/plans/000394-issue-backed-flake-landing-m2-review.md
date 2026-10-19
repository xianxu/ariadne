# Boundary Review — ariadne#394 (milestone M2)

| field | value |
|-------|-------|
| issue | 394 — Issue-backed flake landing and daily triage |
| repo | ariadne |
| issue file | workshop/issues/000394-issue-backed-flake-landing.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 46745426506606d4b4e511f8aa3495830653aacf..a837ec2812c669e71c4e2c66e0e5c53a64a617dc |
| command | sdlc milestone-close --issue 394 --milestone M2 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T17:42:33-07:00 |
| verdict | REWORK |

## Review

Reviewing the M2 diff; now checking the suite log parser for timeout/attribution handling.

Probe confirmed a timed-out suite can land green; now checking test coverage for the baseline label and the declaration validation.

```verdict
verdict: REWORK
confidence: high
```

M2 gets the structure right, but one correctness bug blocks it: a suite that times out can still land green. The decision is a pure `flakes.Classify` called twice (before and after the rerun). The suite and the rerun now share one slot/lease/signal runner (`runUnderSlot`). An interrupted rerun records the failed attempt as unjudged and posts nothing. The `passed()` change also fixed where failed logs are moved. The bug: the runner's own timeout (`rec.TimedOut`) never reaches the classifier. If a known flake failed before the suite was killed, the gate reruns that one test, and the incomplete suite lands with a green "pass · known flake … rerun" status. That breaks the "every test has passed once on this head" argument the D1 decision rests on, and contradicts the new help text, which says a timeout blocks. I reproduced it in the checkout. Three Important gaps remain:
- The registry is not reread at post time, although the plan's ARCH-ORDER note says it is.
- The test-stats readers were not updated for the new rerun attempts.
- README and test coverage are missing for the new knob and labels.

## 1. Strengths
- `cmd/sdlc/localcheck_run.go:643` `runUnderSlot`: the suite and the rerun share one run lease, signal scope, slot and #383 early-release path, with no copy-paste (ARCH-DRY), and `local-check cancel` reaches the rerun.
- `cmd/sdlc/localcheck_flakes.go:43-104`: the IO shell is thin around a pure `Classify` that runs twice. The rerun attempt is saved before its outcome is used, and an unsaved rerun's inflight dir is removed.
- `cmd/sdlc/localcheck_run.go:136`: a log the parser can't read defaults to *unattributable*, so it can never be treated as a flake.
- `cmd/sdlc/localcheck_run.go:609`: the failed log is now moved based on `attempt.Exit`, not `rec.passed()`. That avoids a misplaced log once a known-flake landing counts as passing.
- The v2 store fields are checked against a fixed set (`teststats_store.go:151-163`), so a hand-edited outcome can't get through republish.

## 2. Critical findings
- **A timed-out suite can land green.** `localcheck_flakes.go:45` builds `flakes.Result` only from log facts, and `rec.TimedOut` is never passed in.
  - **Reproduced:** a suite that prints `--- FAIL o/r/pkg TestFlaky`, then sleeps past `timeout_minutes`, with `TestFlaky` registered. Result: `timedout=true passed=true outcome="known-flake-rerun" status=success`, description `"pass · known flake #000448 rerun · timed out in 1s …"`.
  - `unattributedFailures` only catches Go's in-process `panic: test timed out`, not the runner's kill.
  - **Fix sketch:** in `runLocalCheck`, add every failure signal the log doesn't name to `facts.Unattributed` before the classifier runs. Today that means a timeout and a run that did not start. Add a gate test for a timed-out suite whose log names a known flake.

## 3. Important findings
- **The registry is not reread at post time (ARCH-ORDER).** The plan says "the status check rereads the registry at post time, so an issue closed mid-run unregisters its tests".
  - The registry is read once (`localcheck_flakes.go:52`) and reused after a rerun of up to 15 minutes.
  - `republishSavedPass` / `savedAttemptRecord` post a saved known-flake pass with no registry or policy check at all, even days later after the issue closed or the policy became `off`.
  - **Fix sketch:** re-check the outcome against a fresh registry and policy just before posting success (`postLocalCheckRecord`), or change the plan.
- **Test-stats readers were not updated for rerun attempts (ARCH-PURPOSE).**
  - Rerun attempts are stored as ordinary `suite` attempts. `teststats.go:118-131` counts them in `Runs` and `FlakeDenominator`, adds a few-second wall time to the suite wall distribution, and adds their package units to slow-unit times. Nothing reads `RerunOf` except validation.
  - Every reader of the store (summary, snapshot, flake list) needs a decision on rerun attempts: leave them out of suite wall and the denominator, but keep them as same-head pass evidence.
- **The README is missing the gate behaviour and the `known_flakes` knob.** This is the 2nd finding in family `docs-gate-readme`.
  - `helptext/local-check.md` and `helptext/merge.md` gained the gate behaviour and the knob, but `README.md:486-505` was not touched. The README still describes the registry only, not that a failed suite can now land, and not the knob.
  - **Rule:** any range that changes a helptext's user-facing behaviour or config key updates the matching README section (and the atlas `ci-merge-check.md` recorded-local-checks section) in the same commit. Check every helptext diff in the range against the README.
- **Missing tests for new branches.**
  - Neither new `parseLocalChecks` refusal is tested (an invalid `known_flakes` value; `known_flakes` on a check other than the suite).
  - There is no test for the persistent-main-failure label in `diagnoseSuiteFailure`.
  - There is no gate test for a timed-out suite (the Critical above).

## 4. Minor findings
- The rerun timeout is a fixed 15 minutes (`localcheck_flakes.go:22`); the plan says it runs within the check's `timeout_minutes` (ARCH-CONSTRAINTS).
- A rerun that times out or fails to build is labelled `known-flake-failed-twice`. It does block, but the reason is wrong.
- The rerun passes on exit 0 alone. `ok pkg [no tests to run]` would count, and nothing confirms the named tests actually ran.
- The `suitebaseline.go:52` persistent-main-failure label repeats the `pre-existing:` list beside it (ARCH-DRY).
- The test seams are package-level variables (`knownFlakeRerunArgv`, `knownFlakeRerunTimeout`) rather than `localRunDeps` fields.

## 5. Test coverage notes
`TestKnownFlakeGate` drives the real `runLocalCheck` with `ciFakeGH`. It covers rerun→landed, failed twice, unknown, accept, unreadable registry, unattributable, interrupted, and the baseline skip. The targeted tests pass (`go test -run 'TestKnownFlakeGate|TestRunLocalCheck|…'` ok). Not covered:
- a timed-out suite;
- a known-flake landing through merge's `runCIGate` itself;
- the declaration refusals;
- the baseline label;
- whether republish selects the rerun attempt.

## 6. Architectural notes (ARCH-*)
| Marker | Result | Note |
|---|---|---|
| ARCH-DRY | pass | `runUnderSlot` extraction is good; the duplicated label is Minor |
| ARCH-PURE | pass | `Classify` is pure, `knownFlakeStep` is the shell |
| ARCH-PURPOSE | **flag** | the timeout hole undercuts the D1 argument; readers not updated for rerun attempts |
| ARCH-MOCK | pass | `go test` scripted at the argv seam; registry injected; GitHub via `ciFakeGH` |
| ARCH-CONSTRAINTS | minor flag | rerun timeout is not bounded by `timeout_minutes`; `MaxRerun` is enforced |
| ARCH-SECURE | pass | validated names, argv only, store outcomes come from a fixed set |
| ARCH-ORDER | **flag** | no registry reread at post time; republish trusts a stale registry |
| ARCH-FUNERAL | pass | rerun attempts ride the 30-day pruning; an unsaved rerun's inflight dir is removed |

M3's success-post guard should share the same post-time registry and policy check, so republish and the run's own post decide the same way.

## 7. Plan revision recommendations
- `## Revisions`: say where the post-time registry reread lives (M2 post or M3 guard), or drop the ARCH-ORDER claim.
- `## Revisions`: the rerun timeout is a fixed 15 minutes, not `timeout_minutes`, or change the code to match.
- `## Revisions`: rerun attempts are a separate kind of attempt that the stats readers must treat differently (summary, snapshot, flake list).

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: failure-signal-dropped-before-classify
    title: |
      A suite killed by timeout_minutes lands green when its partial log names only a known flake
    detail: |
      knownFlakeStep builds flakes.Result from log facts only; rec.TimedOut never reaches Classify, and unattributedFailures only matches go's in-process timeout panic. Reproduced in the checkout: a suite that prints a FAIL for a registered test and then sleeps past its timeout gives timedout=true, outcome known-flake-rerun, status success. That contradicts the help text and the D1 union argument. Rule: any failure signal the log does not name (a timeout, a run that did not start, an unparseable log) goes into Unattributed before Classify. Add a gate test for the timed-out case.
  - id: new
    severity: Important
    kind: defect
    family: effect-revalidates-fresh-state
    title: |
      The known-flake pass is posted without rereading the registry or policy, unlike the plan's ARCH-ORDER claim
    detail: |
      The registry is read once at localcheck_flakes.go:52 and reused after a rerun of up to 15 minutes. republishSavedPass and savedAttemptRecord post a saved known-flake-rerun or known-flake-accepted outcome with no registry or policy check, so a closed issue or a policy now set to off still turns green. Recheck the outcome against a fresh registry and policy just before posting success, or revise the plan.
  - id: new
    severity: Important
    kind: defect
    family: new-record-kind-unswept-readers
    title: |
      Rerun attempts are counted as full suite runs by test-stats readers
    detail: |
      Rerun attempts are saved with check suite and only rerun_of to mark them. teststats.go:118-131 counts them in Runs and FlakeDenominator, adds their wall time to the suite wall distribution and their units to slow-unit times, and nothing reads RerunOf apart from validation. List every reader of the store (summary, snapshot, flake list) and decide how each treats rerun attempts.
  - id: new
    severity: Important
    kind: defect
    family: docs-gate-readme
    title: |
      README does not cover known-flake landing or the known_flakes key added to the local-check and merge help
    detail: |
      This is the 2nd finding in family docs-gate-readme. Rule: any range that changes a helptext's user-facing behaviour or config key updates the matching README section, and the atlas ci-merge-check recorded-local-checks section, in the same commit. Here README.md lines 486-505 still describe the registry only.
  - id: new
    severity: Important
    kind: defect
    family: claimed-behavior-without-test
    title: |
      The new declaration refusals and the persistent-main-failure label have no tests
    detail: |
      There is no test for an invalid known_flakes value, for known_flakes on a check other than the suite in parseLocalChecks, or for the diagnoseSuiteFailure PersistentMain label in suitebaseline.go.
  - id: new
    severity: Minor
    kind: defect
    family: declared-envelope-not-enforced
    title: |
      The rerun timeout is a fixed 15 minutes, not bounded by the check's timeout_minutes as the plan states
    detail: |
      knownFlakeRerunTimeout in localcheck_flakes.go is a constant; a rerun that times out is also labelled known-flake-failed-twice.
  - id: new
    severity: Minor
    kind: defect
    family: rerun-evidence-unverified
    title: |
      A rerun passes on exit 0 alone, so a run with no tests to run would count as landing evidence
    detail: |
      Nothing confirms that the named tests appear as passed in the rerun log.
  - id: new
    severity: Minor
    kind: defect
    family: duplicated-io-glue
    title: |
      The persistent-main-failure label repeats the pre-existing list it sits beside
    detail: |
      This is the 2nd finding in family duplicated-io-glue. Here it is one shared-failure list printed twice, in suitebaseline.go around line 52. Compute the list once and label it once.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"9513e4ab-6515-4927-931e-466daee25d25","repository":"github.com/xianxu/ariadne","issue_id":"000394","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/2a8edeac52931d01","reviewed_head":"a837ec2812c669e71c4e2c66e0e5c53a64a617dc","slot":"ariadne:2","requested_at":"2026-10-19T00:38:24.489004Z"},"phase":"finished","pid":64171,"started_at":"2026-10-19T00:38:24.604582Z","finished_at":"2026-10-19T00:42:32.941472Z","session_id":"2edc7a8a-36c0-4e66-bd0f-ae35e1cbce20","outcome":"returned"}] -->
