---
id: 000431
status: working
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'bc88af914b9dcee535f1551b0293953083ab0676' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T18:23:15-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# sdlc local-check: republish a completed attempt whose GitHub status post failed

## Problem

A completed `sdlc local-check run --name suite` whose GitHub status POST fails (network error to api.github.com after the run) leaves the attempt record and log intact locally, but no `sdlc-local/suite` status on the commit. There is no verb that publishes only that saved result, so the slot must rerun the whole suite (12-15 min, and a machine suite slot) or hand-post the status with `gh api`, which bypasses the gate. Hit twice on 2026-10-17 within minutes: pair:7 (pair#621, attempt a1f1f773, integrated head 055cb82b) and pair:5 (pair#651, head 80063e6d).

## Spec

A publication-only retry: republish the saved result of a completed attempt for its exact commit, verifying that the attempt record, its log hash and its commit all match, and that the attempt finished with exit 0 (or publish the failure as a failure). Refuse when the record is missing, incomplete or for another commit. The run path should also retry the POST a few times before giving up, and its failure message should name the republish command.

### Design (2026-10-17)

**Why the POST fails** (measurements in Log): not a stale proxy credential. `sdlc` runs outside the agent sandbox
(the whole-call exemption), with no proxy in its environment, so its `gh` and `git` connect to GitHub directly; the
failures are direct connects timing out (`Failed to connect to github.com port 443 after 75008 ms`, `error
connecting to api.github.com`). A fresh `sdlc pr` failed the same way, so age of the process is not the cause; a
sandboxed background call kept a working path through the proxy for 16 minutes. The direct path has outages lasting
a minute or more, and every long-running command is likely to straddle one at its end. So the remedies are time and
a saved result: a bounded retry spanning a couple of minutes, then a recovery verb that posts the saved result
later, and merge reusing a saved green attempt before it reruns anything.

- **Pure core** — `savedAttemptRecord(a SuiteAttempt, log []byte, repo, sha, check) (localRunRecord, error)` in a new
  `localcheck_republish.go`: refuses unless the attempt is for this repo, exact commit and check, is not a main
  baseline, is complete (finished, wall seconds present — `decodeSuiteAttempt` already rejects malformed records),
  and the log's sha256 equals the recorded `log_sha256`. It returns the record the run would have posted
  (exit/timeout, duration from `wall_seconds`, hash, tail); host is this machine's, as the store is local. Table
  tests: match publishes; tampered log, other commit, other check, other repo, baseline, missing log refuse.
- **Lookup** — `newestSavedAttempt(storeRoot, repo, sha, check)` via `readSuiteAttempts` (the store's one reader,
  ARCH-DRY), newest `finished_at` wins; `--attempt ID` names one exactly.
- **Verb** — `sdlc local-check republish --name N [--rev HEAD] [--attempt ID]`: posts the saved outcome of that
  commit's attempt as `sdlc-local/N` (a failure posts as a failure, plus its PR comment), never runs anything. No
  attempt, or one that fails verification, refuses with what it found.
- **Run path** — the status POST retries with backoff (15 s, 45 s, 90 s: about 2.5 minutes, after the suite slot is
  already released; under merge it holds the landing turn, which still beats a rerun). When it still fails, the error names the recovery exactly:
  `the result is saved; publish it with "sdlc local-check republish --name N --rev <sha>" (no rerun needed)`.
- **Merge** — `newLocalCheckRunner` (the one seam both merge paths use for a missing run) first looks for a saved
  *passing*, verified attempt for the exact head and check and republishes it; only without one does it run. A
  saved failure is never reused (merge reruns, as today). If the republish POST fails, merge stops with the same
  recovery line instead of rerunning a 12-minute suite.
- ARCH-SECURE: the attempt JSON and log come from the local store, an untrusted-disk boundary
  (`decodeSuiteAttempt`); the log hash binds the published result to the bytes on disk, and the exact-commit check
  stops a record being posted on another head. ARCH-FUNERAL: creates nothing durable (it reads the existing store,
  pruned at 30 days). ARCH-ORDER: holds no state between events beyond the store; a republish of an already-posted
  status is an idempotent overwrite with the same state and description. ARCH-MOCK: GitHub via the existing
  `ghCaller` fake (`ciFakeGH`).

## Done when

- A completed attempt whose POST failed can be published without rerunning, and only for its own commit and result.
- Tests cover: matching record publishes; tampered log hash, other commit, incomplete or missing record refuse.
- The POST failure message names the recovery command.

## Plan

- [x] `savedAttemptRecord` + `newestSavedAttempt` with table tests
- [x] `local-check republish` verb (fake-GH test: posts the saved outcome, refuses mismatches)
- [x] Run path: bounded POST retry; the final error names the republish command (fake POST failing n times)
- [x] Merge: `newLocalCheckRunner` republishes a saved green attempt for the exact head before running (test: no rerun)
- [x] Help (local-check.md), atlas row

## Revisions

- 2026-10-17 (close review round 1): merge reuses a saved pass only when it is the *newest* attempt of that check on
  that head (`loadSavedAttempt` reports it), so a later attempt whose log was pruned blocks reuse; the verb keeps
  "newest that verifies" and documents it. The backoff note now says that under merge it holds the landing turn.
  The log-path resolver moved beside the store (`suiteAttemptLogFile` in teststats_store.go) and the branch-PR lookup
  is shared (`currentBranchPR`).

## Log

### 2026-10-17
- 2026-10-17: closed — sdlc-local/suite green on 5b2d792d (exit 0, 8m47s). TestLocalCheckRepublishCmd: the verb from the checkout posts a saved failure as a failure with its PR comment and refuses missing --name, another commit and a bad --rev. TestRunLocalCheckPostRetryAndRecovery: 2 failed posts ridden out; a lasting outage leaves the result saved and names sdlc local-check republish. TestLoadSavedAttemptRefuses, TestSavedAttemptRecord (9 rows), TestLocalCheckRunnerReusesSavedPass, TestLocalCheckRunnerSavedPassEdges (newest-attempt guard reached deliberately; failed republish stops with the recovery line). README, atlas ci-merge-check and help swept. Root cause measured: direct-path outages, not a stale proxy credential. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
- 2026-10-17: flow upgraded quick → full — 289 added lines in code files (limit 100); an earlier round of this close already ran the full review

### 2026-10-17 (ariadne:2)

- Hub hypothesis (stale per-call proxy credential) tested. Each sandboxed Bash call does get a distinct credential
  in `HTTPS_PROXY` (fingerprints differ between two calls). But a background sandboxed call probing
  `gh api rate_limit` at t+0, 31, 60, 121, 240, 381, 540, 721, 960 s succeeded at every point except one
  `Bad Gateway` at 240 s, so the credential is neither revoked when the call returns nor after ~16 minutes.
- An unsandboxed call has no proxy variables at all (`env | grep -ci proxy` → 0) and reaches api.github.com
  directly (5/5 in 0.08–0.18 s). The failing `sdlc` errors are direct-connect timeouts, and the fresh `sdlc pr` that
  opened PR #290 failed with one, so long-lived processes are not the cause; direct-path outages are.
- Implemented: `localcheck_republish.go` (`savedAttemptRecord`, `loadSavedAttempt`, `postLocalCheckRecord` with the
  retry, `republishSavedPass`), the verb in `localcheck_cmd.go`, `newLocalCheckRunner` reuse. `runLocalCheck`'s post
  now goes through `postLocalCheckRecord` (ARCH-DRY with the verb). TestMain shrinks the backoff to 1 ms so failing
  fakes don't sleep. Tests: `TestSavedAttemptRecord` (9 rows), `TestRunLocalCheckPostRetryAndRecovery`,
  `TestLoadSavedAttemptRefuses` (other commit/check/id/repo, tampered and missing log),
  `TestLocalCheckRunnerReusesSavedPass` (pass reused without rerun; failure rerun).
- A broad targeted run hit `TestLocalCheckSignalCancelsQueuedHelper/terminated/cleanup` ("cleanup never reached its
  blocked Git seam") under load; 3/3 alone. Second sighting (first during #424), so filed as flake #434, unclaimed.
- 30-minute direct probe (unsandboxed curl to api.github.com/rate_limit every 15 s, 18:40–19:10 PT): 119/119 returned
  200, max 0.15 s. Outages are intermittent episodes (this slot's were ~16:20–16:50 PT), duration unmeasured; the hub
  declined a standing probe since #431 makes an outage cheap whatever its length.
- Close review round 1 (FIX-THEN-SHIP): BR-1 fixed with `TestLocalCheckRepublishCmd` (the verb from the checkout:
  missing --name, a saved failure posts as a failure with a PR comment, another commit and a bad --rev refuse).
  Minors fixed in the round: the backoff claim, `TestLocalCheckRunnerSavedPassEdges` (a failed republish stops with
  the recovery line and never reruns; a pass with a later unverifiable attempt is rerun), the layout/PR-lookup
  duplication, and the newest-attempt rule (Revisions).
- Close review round 2 (FIX-THEN-SHIP; BR-1..5 disposed addressed): BR-6 swept every doc stating that merge runs a
  missing check itself (README, atlas ci-merge-check, local-check help HOW MERGE USES IT) to say it first posts a
  saved pass; README also names `local-check republish`. Minor (untested-error-path, 2nd in family): the
  newest-attempt test now reaches the guard deliberately (distinct pass/fail output, the pass's bytes restored, the
  failure's own log pruned) and asserts the guard's message.
