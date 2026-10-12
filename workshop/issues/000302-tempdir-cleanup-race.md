---
id: 000302
status: working
deps: []
github_issue:
created: 2026-10-08
updated: 2026-10-12
estimate_hours:
card_mirror: '685fdb50fe8e74a834d7f3d6c7bd9afc72118c30' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T10:29:38-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ] Reproduce under load with a diagnostic cleanup hook; identify the writer.
- [ ] Fix at the source; verify with repeated sharded runs.

## Log

### 2026-10-08

### 2026-10-12
- Context: #286 M2 (`e0c2e844`, 2026-10-08 16:58) already added `testfix.QuietBackgroundGit` (env `GIT_CONFIG_COUNT`: gc.auto=0, maintenance.auto=false) — landed *after* both observations here, speculatively.
- Measured hole in it: `GIT_TRACE=1 git push` to a local bare origin runs `unset GIT_CONFIG_COUNT GIT_CONFIG_PARAMETERS GIT_PREFIX; git-receive-pack …` (git's local_repo_env strip), and receive-pack then starts `git maintenance run --auto --quiet --detach` in origin.git with no suppressor; trace2 shows the detached child exiting *after* `git push` returns. On these tiny repos it runs no task (foreground run: no files changed); its only write is the transient `objects/maintenance.lock`.
- Go's RemoveAll reports the deepest dir whose own rmdir failed — so the observed errors (`…/001/.git`, `…/003/origin.git`) mean a new entry appeared *directly* in `.git`/`origin.git` after listing. A maintenance.lock would report `…/objects`, so detached maintenance doesn't fit the observed path by itself.
- Repro attempts: push→os.RemoveAll loop, 341 iterations at load ~100: 0 failures. `go test -run 'TestIssueNew|TestHandoff|TestTakeover' -count=15` with a diagnostic cleanup hook in newTrackerRepo (runs after TempDir's RemoveAll; dumps leftovers + git procs + lsof): 0 hits.
- Ruled out: repolock (Release is synchronous, no heartbeat); tracker env/TrunkFile git calls already kill the process group on cancel.
