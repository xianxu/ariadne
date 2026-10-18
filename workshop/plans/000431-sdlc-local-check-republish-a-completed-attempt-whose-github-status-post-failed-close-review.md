# Boundary Review — ariadne#431 (whole-issue close)

| field | value |
|-------|-------|
| issue | 431 — sdlc local-check: republish a completed attempt whose GitHub status post failed |
| repo | ariadne |
| issue file | workshop/issues/000431-sdlc-local-check-republish-a-completed-attempt-whose-github-status-post-failed.md |
| boundary | whole-issue close |
| milestone | — |
| window | 32fb3f7d056d59bc46d88e2b2d81cb035c4f2d51..98bb4a04bde03b3e74e1736ee4b3917101154416 |
| command | sdlc close --issue 431 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-17T19:43:06-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Diff read; now checking test harness seams and running the new tests.

Tests pass. Now reverting the merge-reuse hook to confirm its test goes red.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The change does what the issue asks. A completed attempt is posted only after its repo, commit, check, completeness and log sha256 have all been checked. The status post now goes through one helper with a bounded retry, shared by the run path, the new verb and merge's runner. Merge reuses a saved pass on the exact head, never a saved failure. The tests pass: the targeted local-check set ran with `-count=1` and returned `ok` in 268 s. I also disabled the hook at `cmd/sdlc/cigate.go:400` and `TestLocalCheckRunnerReusesSavedPass` failed with "the check was rerun despite a saved pass", so the merge-reuse test catches a real regression. One thing stops a plain SHIP. Plan item 2 says the `local-check republish` verb has its own fake-GH test, but no test calls `runLocalCheckRepublish`. The tests call the functions it uses (`loadSavedAttempt` and `postLocalCheckRecord`), so the code that resolves `--rev`, finds the store path, looks up the PR and wraps errors is untested. That gap is cheap to close.

**1. Strengths**
- `savedAttemptRecord` (`cmd/sdlc/localcheck_republish.go:39`) is a pure function with a 9-row table test. It refuses an attempt for another repo, commit or check, a main baseline, an incomplete record, or a log whose hash has changed.
- ARCH-DRY: the run path and the verb post through the same function, `postLocalCheckRecord` (`localcheck_republish.go:121`). It holds the retry and the best-effort PR comment, and the old inline post in `localcheck_run.go` is gone. `suiteStoreRoot` and `hostName` also replace path and hostname code that had been written out in several places.
- The recovery message appears only when the attempt was actually saved (the `saved` flag at `localcheck_run.go:188`, used at line 236). A run on the legacy log path, where nothing is saved, does not point users at a republish that would refuse.
- The tests use the real `runLocalCheck` store. Logs are tampered with and deleted on disk (`TestLoadSavedAttemptRefuses`), and `flakyPostGH` sits behind the existing `ghCaller` seam. That is a realistic fake of a stateful service, not a mock that repeats the implementation.
- Merge reuses only a *pass* (`localcheck_republish.go:160`). A saved failure always reruns, as the spec requires, and the second subtest covers this.

**2. Critical findings**
None.

**3. Important findings**
- **The republish verb's own code is untested, despite plan item 2** (`cmd/sdlc/localcheck_cmd.go:183-222`). The step at `localcheck_republish_test.go:100` that claims to recover "as the republish verb runs it" calls the internal functions directly. Nothing exercises `--rev` resolution, the store path derived from `localCheckLogDir`, the PR lookup for a failure's comment, the missing `--name` refusal, or the error wrapping. `runLocalCheckCmd` already has tests in `localcheck_cmd_test.go` that swap `ghClient`. **Fix:** add a test that swaps `ghClient` for `ciFakeGH` and calls `runLocalCheckRepublish`. It should cover: a passing attempt gets posted; a different `--rev` is refused; a failed attempt posts a failure status plus a PR comment.

**4. Minor findings**
- **ARCH-CONSTRAINTS:** the spec says the retry "holds nothing scarce". That is true for the suite slot, but the merge paths (`landing.go:517`, `merge.go:623`) now hold the landing turn through up to about 2.5 minutes of backoff while GitHub is down. Before this change they failed at once. Either correct the claim or use a shorter backoff for merge.
- **Untested path:** when merge's republish post fails (`localcheck_republish.go:165`), merge should stop and print the recovery line instead of rerunning. No test covers that.
- **Choosing an attempt:** both the verb and merge pick "the newest attempt that verifies", not "the newest attempt". If the newest attempt on a commit is a failure whose log has been pruned, merge will reuse an older pass from the same commit. `--attempt <old pass>` can also overwrite a failure that was already posted. It is the same tree, so this only matters for flaky tests, but the help text should say so.
- **ARCH-DRY:** `savedAttemptLog` (`localcheck_republish.go:25`) hard-codes the reverse of the log layout `publishSuiteAttempt` writes (`teststats_store.go:272,287`). A helper next to the store that resolves the path would keep the two in step.
- **ARCH-DRY:** the verb's branch→PR lookup (`localcheck_cmd.go:213-217`) copies the lookup in the run verb's glue.
- **No retry classification:** `postLocalCheckRecord` retries every error, including permanent ones like a 4xx. A 422 waits 2.5 minutes for nothing.
- **Plan naming:** the plan and spec call the function `newestSavedAttempt(storeRoot, repo, sha, check)` and `savedAttemptRecord(..., log []byte, ...)`. The code has `loadSavedAttempt(..., id, host)` and takes `logSHA, tail`. The Log mentions the code names, but the Plan checkbox and Spec still use the old ones.

**5. Test coverage notes**
- Covered: the pure verification table; the retry recovering after a short outage (3 posts, then success); a long outage naming the recovery command with "no rerun needed"; refusal for another commit, check, attempt id or repo, a tampered log and a missing log; merge reusing a pass and rerunning a failure. I confirmed the merge-reuse test fails when the hook is disabled.
- `TestMain` shrinks the backoff to 1 ms for every test, so no test sleeps through the real backoff. Good.
- Gaps: no test calls the verb itself, merge's post-failure path is untested, and no test checks that a republished failure posts its PR comment.

**6. Architectural notes for upcoming work**
- ARCH-DRY: flagged as Minor (log-layout inverse and PR lookup); the post path itself is correctly shared.
- ARCH-PURE: pass. The verification is pure and the IO (hashing, reading the store, posting) sits in thin callers.
- ARCH-PURPOSE: pass. All three remedies are delivered: the verb, the run-path retry with a named recovery, and merge reuse.
- ARCH-MOCK: pass. GitHub is reached only through `ghCaller`, with stateful fakes.
- ARCH-CONSTRAINTS: flagged as Minor (the landing turn is held during backoff).
- ARCH-SECURE: pass. The store is untrusted disk: `decodeSuiteAttempt` checks `LogPath` against the two allowed forms, and the hash ties the posted result to the log bytes. Each refusal is reported (the merge path warns on stderr) rather than silently substituted. `--rev` reaches git as an argv argument, not through a shell.
- ARCH-ORDER: pass. Reposting the same status is an idempotent overwrite, so retrying a POST that timed out but actually landed is harmless. The verb holds no state between events beyond the store.
- ARCH-FUNERAL: pass. Nothing new and durable is created; the existing 30-day store prune and 20-commit log prune apply.

**7. Plan revision recommendations**
- Add a "## Revisions" entry saying the lookup became `loadSavedAttempt(storeRoot, repo, sha, check, id, host)` and that `savedAttemptRecord` takes `(logSHA, tail)`, not the log bytes.
- Either add the verb-level test or reword plan item 2 to match what is actually tested.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: plan-claimed-test-missing
    title: |
      The republish verb's own code (runLocalCheckRepublish) has no test, despite plan item 2 claiming a fake-GH test of the verb
    detail: |
      The tests call loadSavedAttempt and postLocalCheckRecord directly, so --rev resolution, the store path from localCheckLogDir, the PR lookup for the failure comment, the missing --name refusal and the error wrapping in localcheck_cmd.go:183-222 are never run. Add a test that swaps ghClient for ciFakeGH and calls runLocalCheckRepublish: a pass gets posted, another --rev is refused, a failure posts a failure status plus a PR comment.
  - id: new
    severity: Minor
    kind: defect
    family: operating-envelope-claim
    title: |
      The status-post backoff (about 2.5 min) holds the landing turn on the merge paths, contradicting the spec's claim that it holds nothing scarce
    detail: |
      newLocalCheckRunner, called from landing.go:517 and merge.go:623, now waits up to about 2.5 minutes inside waitCIGate while GitHub is unreachable; before this change it failed at once. Correct the claim or use a shorter backoff for merge.
  - id: new
    severity: Minor
    kind: defect
    family: untested-error-path
    title: |
      No test covers merge's republish post failing (stop with the recovery line instead of rerunning)
    detail: |
      localcheck_republish.go:165 returns ok=true with an error so the gate stops; no fixture makes that post fail.
  - id: new
    severity: Minor
    kind: defect
    family: duplicated-layout-knowledge
    title: |
      savedAttemptLog hard-codes the reverse of publishSuiteAttempt's log layout; the PR lookup copies the run verb's
    detail: |
      Put a log-path resolver next to teststats_store.go so the layout has one source, and share the branch-to-PR lookup.
  - id: new
    severity: Minor
    kind: defect
    family: attempt-selection-semantics
    title: |
      "Newest that verifies" can choose an older pass over a newer failure whose log was pruned, and --attempt can overwrite a posted failure
    detail: |
      It is the same tree, so this only matters for flaky tests; say so in the help text, or have merge refuse to reuse when a newer attempt exists but does not verify.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"c73562b8-3aa6-4689-b9c1-a2b6d004e780","repository":"github.com/xianxu/ariadne","issue_id":"000431","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/514707126d234259","reviewed_head":"98bb4a04bde03b3e74e1736ee4b3917101154416","slot":"ariadne:2","requested_at":"2026-10-18T02:35:17.936734Z"},"phase":"finished","pid":41616,"started_at":"2026-10-18T02:35:18.060769Z","finished_at":"2026-10-18T02:43:03.853961Z","session_id":"ecc61eb5-39f4-4793-b767-4d02c929af1c","outcome":"returned"}] -->
