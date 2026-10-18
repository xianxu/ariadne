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
---

# sdlc local-check: republish a completed attempt whose GitHub status post failed

## Problem

A completed `sdlc local-check run --name suite` whose GitHub status POST fails (network error to api.github.com after the run) leaves the attempt record and log intact locally, but no `sdlc-local/suite` status on the commit. There is no verb that publishes only that saved result, so the slot must rerun the whole suite (12-15 min, and a machine suite slot) or hand-post the status with `gh api`, which bypasses the gate. Hit twice on 2026-10-17 within minutes: pair:7 (pair#621, attempt a1f1f773, integrated head 055cb82b) and pair:5 (pair#651, head 80063e6d).

## Spec

A publication-only retry: republish the saved result of a completed attempt for its exact commit, verifying that the attempt record, its log hash and its commit all match, and that the attempt finished with exit 0 (or publish the failure as a failure). Refuse when the record is missing, incomplete or for another commit. The run path should also retry the POST a few times before giving up, and its failure message should name the republish command.

### Design (2026-10-17)

**Why the POST fails** (see Log for the measurement): every agent Bash call gets its own proxy credential in
`HTTPS_PROXY` (`user:pass@localhost:port`, distinct per call). A long-lived `sdlc` hands the credential it was born
with to every `gh` it spawns, so once that credential stops working, no in-process retry can succeed; only a *new*
command (a fresh Bash call) gets a working path. So the recovery is a verb run as a new process, and merge reuses a
saved green attempt before it reruns anything. No in-process POST retry loop (it would only delay the failure).

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
- **Run path** — when the POST fails after a completed run, the error names the recovery exactly:
  `run "sdlc local-check republish --name N --rev <sha>" as a new command (this process's network path may be
  stale); the result is saved`. No rerun is needed.
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

- [ ] `savedAttemptRecord` + `newestSavedAttempt` with table tests
- [ ] `local-check republish` verb (fake-GH test: posts the saved outcome, refuses mismatches)
- [ ] Run path: the POST-failure error names the republish command (test with a failing fake POST)
- [ ] Merge: `newLocalCheckRunner` republishes a saved green attempt for the exact head before running (test: no rerun)
- [ ] Help (local-check.md), atlas row

## Log

### 2026-10-17
