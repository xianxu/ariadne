---
id: 000302
status: open
deps: []
github_issue:
created: 2026-10-08
updated: 2026-10-08
estimate_hours:
card_mirror: 'b181a5617425d8b0ed1e35e7969bdc48fdfb8375' # card fields mirrored from issue-cards; edit via sdlc
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
