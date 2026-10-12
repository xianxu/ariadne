---
id: 000302
status: codecomplete
deps: []
github_issue:
created: 2026-10-08
updated: 2026-10-12
estimate_hours:
card_mirror: '7553a22436e35e3ccb1e797addf1ae972348cea1' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T10:29:38-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 1.54
---

# Test TempDir cleanup race: directory not empty under sharded make test

## Problem

`make test` (sharded) intermittently fails a test at cleanup, not in its body:

    testing.go:1617: TempDir RemoveAll cleanup: unlinkat /tmp/.../001/.git: directory not empty

Seen twice during #284 (2026-10-07/08):
- `TestIssueNewOnFeatureBranchCommitsOnlyItsDetails` (on `.../001/.git`; this test predates #284);
- `TestHandoffRecognisesItsNoteCommit` (on `.../003/origin.git`).

Each passes repeatedly when run alone (8× and 4×). Something is still writing inside a temp repository when the test's TempDir cleanup removes it.

## Spec

Find the writer before fixing anything. #284 deliberately made no speculative fix. Candidates to confirm or rule out:
- a detached `git gc --auto` / `git maintenance run --auto` (unlikely: the temp repos are far below auto-gc thresholds);
- a git process from a push's receive side (`receive.autogc`);
- a child process a test or verb starts and doesn't reap before returning (e.g. a killed `git log` stream, a built-binary race helper).

Approach: run the affected tests under `-count=N` in parallel shards with a cleanup hook that, on failure, lists the processes holding files under the temp dir (`lsof +D`) and the leftover file names. Fix the cause where it lives; if it is git's background maintenance, disable it for the test process in one place (TestMain), with the evidence.

## Done when

- The writer is identified, with evidence recorded in the Log.
- The fix removes it at the source, and a repeated sharded run (e.g. `make test` ×5) shows no cleanup race.

## Plan

- [x] Reproduce under load with a diagnostic cleanup hook; identify the writer.
- [x] Regression test: a cancelled push through `trackerEnv.gitRaw` leaves no receive-side process (hook pid dies; `testfix.AwaitExit`).
- [x] QuietBackgroundGit: template config so receive-side maintenance is off too; test that a local push spawns no `maintenance run`.
- [x] Verify with repeated sharded runs (make test, one at a time).

## Log

### 2026-10-08

### 2026-10-12
- 2026-10-12: closed — Close-review minors fixed: cancel assertion extracted to testfix.AssertCancelEndsReceiveSide (both runner tests pass); atlas claim narrowed to runners that push, fetch or spawn hooks. Prior evidence stands: writer repro (receive-pack writes ~1.5s after a direct-child kill, nothing after a group kill); the regression test fails without the group kill; make test 6 runs, directory not empty 0/6.; review verdict: SHIP
- 2026-10-12: flow upgraded quick → full — 107 added lines in code files (limit 100)
- 2026-10-12: closed — Writer identified with a repro: SIGKILL to only the direct git push leaves receive-pack writing refs into origin ~1.5s later, while a group kill leaves nothing; #286 M1 group kill is the source fix. New regression tests fail without the group kill (pid still running after 2m0s) and pass with it. Receive-side maintenance hole closed via GIT_TEMPLATE_DIR; the test asserts a baseline push spawns maintenance and a quieted one does not. make test after the fix: 6 sharded runs, directory not empty 0/6. The non-race failures were TestPlanningReviewCompetingResults (2x, main is red, #338) and cmd/weave TestCompileCancellationStopsLateGeneratorChild (1x on the merged tree; passes 5/5 alone; the branch has no cmd/weave diff).; review verdict: SHIP
- Context: #286 M2 (`e0c2e844`, 2026-10-08 16:58) already added `testfix.QuietBackgroundGit` (env `GIT_CONFIG_COUNT`: gc.auto=0, maintenance.auto=false) — landed *after* both observations here, speculatively.
- Measured hole in it: `GIT_TRACE=1 git push` to a local bare origin runs `unset GIT_CONFIG_COUNT GIT_CONFIG_PARAMETERS GIT_PREFIX; git-receive-pack …` (git's local_repo_env strip), and receive-pack then starts `git maintenance run --auto --quiet --detach` in origin.git with no suppressor; trace2 shows the detached child exiting *after* `git push` returns. On these tiny repos it runs no task (foreground run: no files changed); its only write is the transient `objects/maintenance.lock`.
- Go's RemoveAll reports the deepest dir whose own rmdir failed — so the observed errors (`…/001/.git`, `…/003/origin.git`) mean a new entry appeared *directly* in `.git`/`origin.git` after listing. A maintenance.lock would report `…/objects`, so detached maintenance doesn't fit the observed path by itself.
- Repro attempts: push→os.RemoveAll loop, 341 iterations at load ~100: 0 failures. `go test -run 'TestIssueNew|TestHandoff|TestTakeover' -count=15` with a diagnostic cleanup hook in newTrackerRepo (runs after TempDir's RemoveAll; dumps leftovers + git procs + lsof): 0 hits.
- Ruled out: repolock (Release is synchronous, no heartbeat); tracker env/TrunkFile git calls already kill the process group on cancel.
- **Writer identified: the other end of a local git transport, outliving a cancelled `git`.** Repro ($TMPDIR, alarm-capped): origin with a 2 s pre-receive hook; `git push` started, SIGKILL to the *direct* `git` only (what `exec.CommandContext`'s default Cancel does): `refs/heads/main` appears in origin.git ~1.5 s *after* the push died — receive-pack carried on and wrote. Had TempDir's RemoveAll already taken `refs/`, git recreates it (safe_create_leading_directories) ⇒ `unlinkat …/origin.git: directory not empty`, the observed shape. Same run with SIGKILL to the push's process group: no late write, origin untouched.
- Before #286 M1 (`f21fbf4a`, 2026-10-08 16:21 — after both sightings) `trackerEnv.gitRaw` used CommandContext's default Cancel (direct child only); M1 moved it to `processgroup.Terminate` (gitx's trunk writes already were). Under sharded load, read/push deadlines fire, so a tracker test's cancelled push left receive-pack writing into the TempDir. That fix is the source fix; nothing pinned it to this mechanism, so #302 adds the regression test. The `.git` (checkout) sighting is the same class on the fetching side; not reproduced individually.
- Residual hole in #286 M2, measured above: env config is stripped on local transport, so receive-side detached maintenance still runs. Its only write on these repos is nested (`objects/maintenance.lock`), but QuietBackgroundGit claims to cover pushes. Fix: also set `GIT_TEMPLATE_DIR` to a template whose `config` turns off gc.auto / maintenance.auto / receive.autogc — repo-local config survives the strip (verified: 0 maintenance spawns on push vs 1).
- Post-#286 sightings: none found (ops project log, issues, lessons).
- Fix (16ce7f3a): `testfix.BlockingOrigin` (pre-receive hook records pid, blocks) + `TestTrackerGitCancelEndsTheReceiveSide` (cmd/sdlc, trackerEnv.gitRaw) and `TestGitCancelEndsTheReceiveSide` (gitx, executeGitInputContext). Sensitivity: with gitx's group kill removed the test fails — `the receive side outlived the cancelled push: pid … still running after 2m0s`; restored, passes. `QuietBackgroundGit` now also sets GIT_TEMPLATE_DIR (config: gc.auto=0, maintenance.auto=false, receive.autogc=false; empty hooks/ and info/ since tests write there) and returns a cleanup; installed in cmd/sdlc, gitx, tracker, fleet TestMains. `TestQuietBackgroundGitReachesTheReceiveSide` asserts a baseline push spawns receive-side maintenance and a quieted one doesn't.
- Full `make test` #1 (diag hook in, pre-fix code): no cleanup race in any shard; every shard tripped the hermeticity guard because I committed in this worktree mid-run (HEAD moved + an untracked diag file) — self-inflicted, not a test failure.
- Verification, `make test` ×5 sequential after the fix (16ce7f3a): `directory not empty` 0/5. Runs 3–5 fully green (wall 360–393 s). Runs 1–2 each failed only `TestPlanningReviewCompetingResults` ("planning review stale: … changed/appeared"; the competing change-code held the repo lock when the first review came back). It is unrelated to this change: it passes 4/4 alone on this branch, and the fix only touches git gc/maintenance config and test helpers. A load-sensitive review-ordering race, reported to the hub as its own item.
- Atlas: sdlc-binary.md "Running the test suite" gains the QuietBackgroundGit / process-group paragraph.
