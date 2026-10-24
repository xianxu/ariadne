---
id: 000583
status: done
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
card_mirror: '597bf1a5149333e133a18355e6e07d6f63470238' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T20:24:51-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "de8d12eb", done: "9787f76c"}
actual_hours: 0.15
---

# Pre-close and affected checks reuse an existing same-head full-suite pass instead of queueing a new full run

## Problem

## Spec

**`--pre-close` reruns a full suite on a head that already has a passing full-suite record (pair:7, ariadne#578
rollout, 10-23).** On pair#739's unchanged head 00758b61, which had a just-completed recorded full-suite pass (with a
known-flake rerun), `sdlc local-check run --pre-close` printed "affected", then the affected selector fell back to a
full selection (a change to `cmd/couch/main_test.go` selects everything) and queued a **new** full suite on the same
head. A full-suite pass on the exact head already covers every affected selection.

Fix: before running, `--pre-close` (and `--name affected`) accept an existing passing `sdlc-local/suite` record for
the exact pushed head (including a known-flake-qualified pass) as satisfying the check, print that it's reusing it,
and run nothing. A record for another head never counts.

## Done when

- With a passing suite record on the exact head, `--pre-close` and `--name affected` run nothing and report reuse;
  without one, behaviour is unchanged (tests).

## Plan

Reuse, don't add (ARCH-DRY): `republishSavedPass` (#431, `localcheck_republish.go`) already turns a saved, verified
pass of a check on the exact head into its posted record, with the newest-attempt and known-flake revalidation
rules; merge's runner uses it. The CLI never asks it.
- [x] `runLocalCheckCmd`: with `--pre-close`, or `--name affected` (boarding included: a suite pass also boards),
      first ask `republishSavedPass` for `suite` on the exact head; on a reuse, print it and run nothing (no graph,
      no selection). Otherwise behave as today. A plain `--name suite` keeps rerunning on request.
- [x] Tests (CLI fixture from #578's `TestLocalCheckCmdPreClose`): after a recorded suite pass, `--pre-close` and
      `--name affected` run nothing and report reuse (no affected status, no second attempt); a new commit (another
      head) runs as before; a failed saved suite is not reused.
- [x] Scope fold (TL, from pair:1): `cmd/sdlc/helptext/close.md` FIXED-THEN-SHIP prose says `--full-suite` records
      "a full test run" and "the suite still runs"; say the pre-close check (`sdlc local-check run --pre-close`) as
      the flag help does. Pin with a help test.
- [x] local-check help, atlas `ci-merge-check.md` pre-close paragraph: reuse of a same-head suite pass.
- [x] Targeted tests, pre-close check, close, land.

ARCH-SECURE: the saved attempt is local evidence verified by `loadSavedAttemptCurrent` (log hash, host, newest);
a record for another head never matches (keyed by sha). ARCH-ORDER/FUNERAL/EVOLVE: no new state, files or formats.

## Log

### 2026-10-23
- 2026-10-23: closed — targeted: go test ./cmd/sdlc -run TestPreClose|TestLocalCheckCmdPreClose pass; round 1 fixed in-round: TestPreCloseReusesASameHeadSuitePass covers pre-close->affected, pre-close->suite (no trains), --name affected, --boarding reuse with one suite attempt, plain --name suite rerun (2 attempts), another head runs, failed suite not reused; known-flake reuse via republishSavedPass in TestKnownFlakeGate; TestCloseHelpNamesThePreCloseCheck; recorded pre-close check queued on d4a79fa2; review verdict: SHIP

- Implemented in `runLocalCheckCmd` (localcheck_cmd.go): before running, `--pre-close` or an affected-only run asks
  `republishSavedPass` for `suite` on the exact head; a verified pass is posted and nothing runs. Tests:
  `TestPreCloseReusesASameHeadSuitePass` (same head reuses for both flags with one suite attempt total; another head
  runs affected; a failed saved suite is not reused, and #394's guard then refuses the affected pass over it).
- Scope fold: close help FIXED-THEN-SHIP prose names the pre-close check (`TestCloseHelpNamesThePreCloseCheck`);
  swept `--full-suite` docs: atlas `ci-merge-check.md` conflict-resolution line updated, the others were neutral.
- Close review round 1 (FIX-THEN-SHIP), fixed in-round: the Important (reuse branches untested) adds pre-close
  resolving to suite (no trains), plain `--name suite` rerunning (2 attempts), and `--boarding` reuse to
  `TestPreCloseReusesASameHeadSuitePass`; known-flake reuse is `republishSavedPass`'s, covered by `TestKnownFlakeGate`.
  The Minor: `--pre-close` now prints "resolved to X" before the reuse check, and the reuse line names `--pre-close`.

