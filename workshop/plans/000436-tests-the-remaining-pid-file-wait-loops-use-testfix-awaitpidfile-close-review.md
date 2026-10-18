# Boundary Review — ariadne#436 (whole-issue close)

| field | value |
|-------|-------|
| issue | 436 — tests: the remaining pid-file wait loops use testfix.AwaitPIDFile |
| repo | ariadne |
| issue file | workshop/issues/000436-tests-the-remaining-pid-file-wait-loops-use-testfix-awaitpidfile.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1f3827acf47152247630497f952516a9c4b24961..00be65f9fe57f03a85b0b3c76df0827877f2bbd9 |
| command | sdlc close --issue 436 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-18T00:09:40-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This change does what the issue asks. Six hand-written pid-file wait loops in four test files under `cmd/sdlc/internal` now call `testfix.AwaitPIDFile(path, testfix.HangCeiling, nil)`. Each call site keeps its own failure handling: the `t.Fatal` message and, where it existed, the cleanup that kills the parent or child process. The diff removes 70 lines and adds 10, and the now-unused `strconv` imports are gone from `lifetime_test.go` and `process_unix_test.go`. I searched `cmd/sdlc/internal` again for `os.ReadFile` + `strconv.Atoi` inside a loop. The only matches left are one-shot reads (`rebasedpatch_context_test.go:45`, `cancellation_unix_test.go:141`, `preparation_extra_test.go:120`). There is also the in-child wait for `exiting-parent` in `cancellation_unix_test.go:42-49`. That one waits for any content rather than a pid, inside the child process, and the Log explains why it was left alone. The `## Done when` clause holds. All tests in `processgroup`, `judge` and `testfix` pass, the three touched `gitx` tests pass, and `go vet` is clean on all three changed packages.

1. **Strengths**
   - Each site keeps its own cleanup on timeout: `waitReviewChild` still sends SIGKILL to `parent` (`cancellation_unix_test.go:112-117`), and `TestForcedTerminationRetiresMemberMissedByFirstSignal` still kills and waits on `cmd`.
   - The helper's `strings.TrimSpace` now applies at every site, so a pid written with a trailing newline no longer causes a silent miss.
   - The scope exclusion for the in-child content wait is justified and written down in the Log. It is a different readiness condition, not a pid wait.
   - Imports were tidied correctly. `strconv` stays in `rebasedpatch_context_test.go` and `cancellation_unix_test.go` because other code there still uses it.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `waitReviewChild` behaves slightly differently now. Before, a pid file with non-numeric content failed the test at once (`t.Fatal(err)`). Now the helper keeps retrying until `HangCeiling` and then reports "did not become ready". This is arguably better, because it tolerates a partial write. Diagnosis is slightly worse for a corrupt file. Acceptable.
   - The issue's `## Plan` ends with an empty `- [ ]` checklist item. Remove it before close.

5. **Test coverage notes:** Only test code changed, so the evidence is that the touched tests still pass. They do: `processgroup` 0.2s, `judge` 25s, `testfix` 1.8s, and the targeted `gitx` tests 7s. `AwaitPIDFile` itself lives in `testfix` and is covered there.

6. **Architecture**
   - **ARCH-DRY: pass.** All six copies are folded into the existing helper, and no new near-duplicates appear in `internal`. The similar loops in `cmd/sdlc/localcheck_*_test.go` are outside the issue's `internal` scope.
   - **ARCH-PURE: pass.** This is test-only glue with no business logic.
   - **ARCH-PURPOSE: pass.** I re-ran the search for the whole family across `cmd/sdlc/internal`, and the one wait left out is a justified exception, not a deferred instance of the same pattern.

7. **Plan revision recommendations:** none. The only plan cleanup needed is removing the empty checklist item noted above.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: tracker-plan-hygiene
    title: |
      Issue Plan ends with an empty "- [ ]" checklist item
    detail: |
      workshop/issues/000436 Plan section has a stray blank unchecked item after the single delivered item; remove it before close so the plan does not look like it has outstanding work.
  - id: new
    severity: Minor
    kind: defect
    family: helper-adoption-semantics-drift
    title: |
      waitReviewChild no longer fails fast on a non-numeric pid file
    detail: |
      Previously a non-numeric pid file in cmd/sdlc/internal/judge/cancellation_unix_test.go caused an immediate t.Fatal(err). AwaitPIDFile retries until HangCeiling and the test then reports "did not become ready". This tolerates partial writes, but a corrupt file is now harder to diagnose. It is the only one of the six converted sites whose behaviour changed this way.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"fde802df-5497-4539-a38d-408f6b5b1b8b","repository":"github.com/xianxu/ariadne","issue_id":"000436","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/8480fcf660e37d9f","reviewed_head":"00be65f9fe57f03a85b0b3c76df0827877f2bbd9","slot":"ariadne:2","requested_at":"2026-10-18T07:08:12.031125Z"},"phase":"finished","pid":53505,"started_at":"2026-10-18T07:08:12.205737Z","finished_at":"2026-10-18T07:09:27.82222Z","session_id":"41e8c072-46c8-4ee8-8b83-cd6282b2110a","outcome":"returned"}] -->
