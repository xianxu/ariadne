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

---

## Re-review — 2026-10-17T20:14:18-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 431 — sdlc local-check: republish a completed attempt whose GitHub status post failed |
| repo | ariadne |
| issue file | workshop/issues/000431-sdlc-local-check-republish-a-completed-attempt-whose-github-status-post-failed.md |
| boundary | whole-issue close |
| milestone | — |
| window | 32fb3f7d056d59bc46d88e2b2d81cb035c4f2d51..906790a75819bfb7cc96924294914a61955e6e2b |
| command | sdlc close --issue 431 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-17T20:14:18-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Targeted tests pass. Now checking the README section on local-check and doing a revert check on the claimed fixes.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The change does what the issue asks. `savedAttemptRecord` refuses an attempt from another repo, commit or check, a main baseline, an incomplete record, or a log whose hash no longer matches. The run path retries the status post and, if it still fails, names the exact republish command. Merge's runner only reuses a saved pass when it verifies and is the newest attempt on that head. The new verb is now tested end to end. All five prior findings are fixed. I re-ran the targeted tests and `go vet`: both clean. I removed the `!newest` guard, and `TestLocalCheckRunnerSavedPassEdges` failed, then passed again with the guard back. One thing should be fixed before shipping: the docs a user reads were not updated. README.md doesn't mention `local-check republish`, and `atlas/workflow/ci-merge-check.md` still says merge always runs a missing check itself.

Done-when: current. The three bullets still cover the core contract: publish without a rerun, only for the attempt's own commit and result; the refusal tests; and the recovery command in the failure message. The later spec additions (merge reusing a saved pass, the newest-attempt rule) are separate Plan items and are tested, so leaving them out of Done-when doesn't hide any deliverable.

1. **Strengths**
   - The pure core is separate from IO: `savedAttemptRecord` (`cmd/sdlc/localcheck_republish.go:25`) takes the log hash and tail as arguments, and its table test has 9 rows and does no IO.
   - The run path and the verb post through one function, `postLocalCheckRecord`, which holds the retry and the PR comment (ARCH-DRY).
   - When merge's republish post fails, it returns `ok=true` with an error, so merge stops with the recovery line instead of starting a 12-minute rerun. `TestLocalCheckRunnerSavedPassEdges` pins this.
   - The layout knowledge now lives next to the store (`suiteAttemptLogFile`, `teststats_store.go:245`), and the branch-to-PR lookup is one shared function (`currentBranchPR`).
   - TestMain shrinks the backoff to 1 ms, so the failing fakes don't sleep.

2. **Critical findings**: none.

3. **Important findings**
   - **The docs a user reads don't cover the new verb.** README.md:471-478 documents `local-check run` and `cancel` but not `republish`. `atlas/workflow/ci-merge-check.md:160` ("When a needed run is missing, merge runs it itself") no longer matches merge's behavior. Fix: add one README sentence for `sdlc local-check republish --name N [--rev SHA]`, and change the ci-merge-check Merge bullet to say merge first posts a saved, verified pass if it is the newest attempt on the exact head.

4. **Minor findings**
   - In `localcheck_republish_test.go:257-271`, the pass still verifies only because `true` and `false` write identical (empty) logs. A failed attempt also overwrites the shared `local-checks/<sha>/<check>.log`, so if the two logs differed, the pass would stop verifying and the test would no longer exercise the `!newest` branch. Write distinct output and assert the guard explicitly.
   - Every attempt on a commit shares one log file, so after a later run on the same head an earlier pass can't be republished. That is safe, but the help text doesn't say so.

5. **Test coverage notes**
   - BR-1: `TestLocalCheckRepublishCmd` drives `runLocalCheckRepublish` through the swapped-in `ghClient` and `detectRepo`. It covers a missing `--name`, a failure posted as a failure with its PR comment, another commit refused, and a bad `--rev`.
   - BR-3: the edges test's first subtest covers merge's republish post failing.
   - BR-5: the second subtest covers the newest-attempt rule; the revert check above confirms it catches a regression.
   - No test runs the verb on a saved pass. That path goes through the same `postLocalCheckRecord` as the failure case, so the risk is low.

6. **Architectural notes**
   - **ARCH-DRY:** pass. The layout is still written in two places in `teststats_store.go` (`publishSuiteAttempt` writes it, `suiteAttemptLogFile` resolves it), but they sit next to each other.
   - **ARCH-PURE:** pass.
   - **ARCH-PURPOSE:** pass on code; the docs gap is the Important finding above.
   - **ARCH-MOCK:** pass. GitHub is reached through the `ghCaller` seam with the stateful `ciFakeGH`, and `flakyPostGH` wraps it.
   - **ARCH-CONSTRAINTS:** pass. The about 2.5-minute backoff, including that it holds the landing turn under merge, is now stated in the spec, help text and code comment.
   - **ARCH-SECURE:** pass. Records go through `decodeSuiteAttempt`, the log hash ties the result to the bytes on disk, and the exact commit and check are enforced even with `--attempt`.
   - **ARCH-ORDER:** pass. Reposting a status overwrites it with the same state, and merge's newest-attempt rule covers a later attempt whose log was pruned.
   - **ARCH-FUNERAL:** pass. Nothing durable is created; it reads the existing store, which is pruned.

7. **Plan revision recommendations**: none. Revisions already records the round-1 changes.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      TestLocalCheckRepublishCmd calls runLocalCheckRepublish with a swapped ghClient/detectRepo: missing --name, failure posts failure plus PR comment, other commit and bad --rev refuse; passes at head.
  - id: BR-2
    disposition: addressed
    note: |
      Spec, help text (local-check.md REPUBLISH) and the localCheckPostBackoff comment now state the wait holds the landing turn under merge; prose-only correction verified against cigate.go newLocalCheckRunner.
  - id: BR-3
    disposition: addressed
    note: |
      TestLocalCheckRunnerSavedPassEdges subtest 1 makes the republish post fail and asserts the recovery line and no rerun.
  - id: BR-4
    disposition: addressed
    note: |
      suiteAttemptLogFile now sits beside publishSuiteAttempt in teststats_store.go; currentBranchPR is shared by run and republish.
  - id: BR-5
    disposition: addressed
    note: |
      Merge requires newest (loadSavedAttempt reports it); help text documents the verb's newest-that-verifies caveat; removing the !newest guard turns TestLocalCheckRunnerSavedPassEdges red.
findings:
  - id: new
    severity: Important
    kind: defect
    family: docs-surface-not-updated
    title: |
      README and atlas ci-merge-check.md not updated for local-check republish and merge's saved-pass reuse
    detail: |
      README.md:471-478 documents local-check run and cancel but not the new republish subcommand; atlas/workflow/ci-merge-check.md:160 still says merge runs a missing check itself, though newLocalCheckRunner now first posts a saved verified newest pass. Add a README sentence for the verb and update the ci-merge-check Merge bullet.
  - id: new
    severity: Minor
    kind: defect
    family: untested-error-path
    title: |
      The newest-attempt edge test passes only because the true and false checks write identical empty logs
    detail: |
      This is the 2nd finding in family untested-error-path. The rule: every guard branch needs a fixture that reaches it on purpose, not by chance. Here a failed attempt overwrites the shared local-checks/<sha>/<check>.log, so the earlier pass verifies only because both logs are empty. Use distinct output and assert the not-newest branch explicitly (localcheck_republish_test.go:257-271).
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e8b024d2-76e2-4d90-9a7c-307013a4e8bb","repository":"github.com/xianxu/ariadne","issue_id":"000431","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/16a783c437ec35a6","reviewed_head":"906790a75819bfb7cc96924294914a61955e6e2b","slot":"ariadne:2","requested_at":"2026-10-18T03:12:28.228565Z"},"phase":"finished","pid":22613,"started_at":"2026-10-18T03:12:28.321564Z","finished_at":"2026-10-18T03:14:17.853227Z","session_id":"98f8c616-8a4a-4839-927e-cccf505ce969","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T20:47:03-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 431 — sdlc local-check: republish a completed attempt whose GitHub status post failed |
| repo | ariadne |
| issue file | workshop/issues/000431-sdlc-local-check-republish-a-completed-attempt-whose-github-status-post-failed.md |
| boundary | whole-issue close |
| milestone | — |
| window | 32fb3f7d056d59bc46d88e2b2d81cb035c4f2d51..5b2d792df73a42e43e6d71d6e1b22121a54e5462 |
| command | sdlc close --issue 431 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-17T20:47:03-07:00 |
| verdict | SHIP |

## Review

Prior-round fixes look in place; now running the new tests, then reverting the newest-attempt guard to confirm its test goes red.

```verdict
verdict: SHIP
confidence: high
```

The two open findings from round 2 are both fixed, and I checked each against the code rather than the commit messages. For BR-6, all three docs now say merge first posts a saved pass before rerunning a missing check: README.md:472-481 (which also names `local-check republish`), atlas/workflow/ci-merge-check.md:160-162, and the help text's "HOW MERGE USES IT" section. For BR-7, the newest-attempt edge test now reaches its guard on purpose. The pass and the failure write different output, the pass's log bytes are put back, and the failure's own log is deleted. The test then checks the guard's stderr message. I reverted the guard (`|| !newest` in `republishSavedPass`) in the disposable checkout and `TestLocalCheckRunnerSavedPassEdges` failed with "an older pass stood in although a later attempt exists". I then restored the file. `go vet` is clean. The six new tests pass, and the wider `LocalCheck|CIGate|TestStats|SuiteAttempt|Help` subset passes too (283 s). Only one Minor finding remains, so it does not block.

Done-when: current — the three bullets (publish without a rerun only on the attempt's own commit and with its own result; tests for match, tampered log, other commit and incomplete or missing record; the failure message naming the recovery command) all match what this diff builds. The merge-reuse rule from Revisions only adds to them, has its own tests, and no bullet contradicts it.

1. **Strengths**
   - `savedAttemptRecord` (localcheck_republish.go:243) is a pure function that refuses on repo, commit, check, baseline, incomplete record or log hash. It is tested with a 9-row table and no IO.
   - `postLocalCheckRecord` is now the only place that posts a status. The run path and the verb share it, so the retry and the PR comment can't drift apart (ARCH-DRY).
   - `suiteAttemptLogFile` sits next to `publishSuiteAttempt` in teststats_store.go:777, so the layout and its reverse live in one file.
   - Merge is stricter than the verb: it needs the saved attempt to be the newest one, and it never reuses a saved failure. When the republish post fails, merge stops with the recovery line instead of quietly rerunning. The tests check that the marker file is absent.
   - The fake GitHub stays stateful: `flakyPostGH` wraps `ciFakeGH` and tracks call counts, and the production seam is used throughout (ARCH-MOCK).

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - localcheck_republish.go:277: `loadSavedAttempt` throws away the warnings from `readSuiteAttempts`. That reader stops at the record limit or byte budget and reads in directory order, so "newest" can be computed from a partial set. This is the 2nd finding in family `attempt-selection-semantics`. The rule: if the store read is incomplete, the newest attempt is unknown, not absent. The fix is to return `newest=false` whenever there are warnings, so merge reruns. The risk is low: the limit is 10000 records and the store is pruned at 30 days.

5. **Test coverage notes:** every guard in `republishSavedPass` has a fixture that reaches it on purpose: a refused attempt, a saved failure, a pass that isn't the newest, and a failed post. The verb is tested end to end from the checkout. The retry is tested by failing the first N posts. `TestMain` shrinks the backoff to 1 ms, so these tests don't sleep.

6. **Architectural notes**
   - ARCH-DRY: pass. The PR lookup, host name, store root and post path are each shared.
   - ARCH-PURE: pass. Record validation is pure, and the lookup and post are a thin IO layer.
   - ARCH-PURPOSE: pass. The verb, the run-path retry, merge reuse and the docs are all delivered.
   - ARCH-MOCK: pass. Tests use the `ciFakeGH` seam; there are no direct `gh` calls.
   - ARCH-CONSTRAINTS: pass. The roughly 2.5-minute backoff, and the fact that it holds the landing turn under merge, are written down in the help text and the code comment.
   - ARCH-SECURE: pass. Store records go through `decodeSuiteAttempt`, and the log hash ties the result to the bytes on disk. The truncated-read case is the Minor finding above.
   - ARCH-ORDER: pass. A republish is an idempotent overwrite. A run still in progress overwrites the shared log, which makes the hash check fail, so merge reruns under the lock instead of reusing.
   - ARCH-FUNERAL: pass. Nothing new is created on disk; this only reads the existing store, which is pruned at 30 days.

7. **Plan revision recommendations:** none. The plan matches the code.

```findings
dispose:
  - id: BR-6
    disposition: addressed
    note: |
      README.md:472-481 names the republish verb and merge's saved-pass-first rule; atlas ci-merge-check.md:160-162 and help HOW MERGE USES IT updated likewise.
  - id: BR-7
    disposition: addressed
    note: |
      Test uses distinct pass/fail output, restores pass bytes, prunes the failure log, asserts the guard message; reverting the newest guard turns it red (localcheck_republish_test.go:293).
findings:
  - id: new
    severity: Minor
    kind: defect
    family: attempt-selection-semantics
    title: |
      loadSavedAttempt computes "newest" over a possibly truncated store read and discards the reader's warnings
    detail: |
      2nd in family attempt-selection-semantics. Rule: when the store read is incomplete, the newest attempt is unknown, not absent. readSuiteAttempts stops at testStatsRecordLimit or the byte budget, reading in directory order, and returns warnings that loadSavedAttempt (localcheck_republish.go:277) drops, so merge could treat an older pass as newest. Fix: report newest=false when warnings are present.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ef69055e-db06-484e-991c-2042eb84bfb8","repository":"github.com/xianxu/ariadne","issue_id":"000431","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/a6d15915ed9f7f19","reviewed_head":"5b2d792df73a42e43e6d71d6e1b22121a54e5462","slot":"ariadne:2","requested_at":"2026-10-18T03:40:34.054194Z"},"phase":"finished","pid":60751,"started_at":"2026-10-18T03:40:34.21089Z","finished_at":"2026-10-18T03:47:01.468032Z","session_id":"fc157129-5341-4ccf-8cca-1c2dfbf0c6c8","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T20:50:04-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 431 — sdlc local-check: republish a completed attempt whose GitHub status post failed |
| repo | ariadne |
| issue file | workshop/issues/000431-sdlc-local-check-republish-a-completed-attempt-whose-github-status-post-failed.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0eb63a95c6a205f2e287929bdee7ed905cfcef76..391c17e163125b715265c3952ba0c1e1c14722c4 |
| command | sdlc close --issue 431 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-17T20:50:04-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

BR-8 is fixed, and I checked that against the code and a test run, not the commit message. `loadSavedAttempt` now keeps the warnings from `readSuiteAttempts`. When that list is not empty, it reports `newest=false` (`cmd/sdlc/localcheck_republish.go:62-65,97`). That covers every way the store read can come up short: the record limit, either byte-budget stop, a file that fails to open or read, and a record that won't decode. So the fix covers the whole family rule ("an incomplete read means the newest attempt is unknown, not absent"), not just the record-limit case the finding named. The new subtest writes a corrupt attempt record. It then checks two things: `loadSavedAttempt` returns a pass that still verifies but has `newest=false`, and merge's runner really reruns the check, because the marker file appears. I reverted the fix in the disposable checkout (keeping it compilable). The subtest then failed with `newest=true <nil>` at `localcheck_republish_test.go:310`. I restored the file afterwards. The test passes at head and `go vet` is clean. Outside the BR-8 fix, the window only adds generated gate and review records for round 3. It does not change the Spec, Revisions or Done-when, and it has no conflict resolutions.

1. **Strengths**
   - The fix uses the warnings `readSuiteAttempts` already returns. There is no second check for "was the read complete?" (ARCH-DRY).
   - The incomplete case fails safe: merge reruns the check rather than letting a saved pass stand in. The `republish` verb (`localcheck_cmd.go:204`) ignores `newest`, so it is unaffected, which is correct.
   - The test checks both the pure result and the end-to-end merge behaviour, and it goes red without the fix.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - Not raised as a finding: one corrupt record anywhere in the store stops merge from reusing a saved pass for every check. That lasts until the store's 30-day pruning removes the record. It is the conservative choice and only costs a rerun. A one-line remark in the help text would make it easier to diagnose.

5. **Test coverage notes:** the new subtest reaches the "store read incomplete" branch through the invalid-record path, on purpose. The limit and byte-budget paths add to the same warnings list, so the same guard covers them.

6. **Architectural notes**
   - **ARCH-DRY:** pass. It reuses the reader's existing warnings.
   - **ARCH-PURE:** pass. The change is one boolean inside the existing load function, with no new IO.
   - **ARCH-PURPOSE:** pass. The fix covers the family rule for every warning source, not only the site the finding named.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-8
    disposition: addressed
    note: |
      localcheck_republish.go:65,97 sets newest=false whenever readSuiteAttempts warns (any limit, budget, open, read or decode warning); the new subtest at localcheck_republish_test.go:299 goes red with newest=true when the guard is reverted.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"15c82fa1-3078-4147-9a01-6b42811c30a6","repository":"github.com/xianxu/ariadne","issue_id":"000431","recipe":"fix-review","round":4,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/2a187b74d4b551c4","reviewed_head":"391c17e163125b715265c3952ba0c1e1c14722c4","slot":"ariadne:2","requested_at":"2026-10-18T03:49:06.310078Z"},"phase":"finished","pid":81027,"started_at":"2026-10-18T03:49:06.504335Z","finished_at":"2026-10-18T03:50:03.437823Z","session_id":"815fe8de-8afc-4d34-8ec8-5507dbda03c0","outcome":"returned"}] -->
