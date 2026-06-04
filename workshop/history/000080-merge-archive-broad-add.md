---
id: 000080
status: done
deps: []
github_issue:
created: 2026-06-03
updated: 2026-06-04
estimate_hours: 1.5
actual_hours: 1.5
---

# sdlc merge archive step blanket-adds issues/ + history/, sweeping unrelated untracked files

## Problem

`sdlc merge`'s archive step (step 11) stages the moved issue files with a
**directory-wide** add (`merge.go:421`):

```go
mergeRunner.GitInDir(mainPath, "add", f.IssuesDir+"/", f.HistoryDir+"/")
```

`git add workshop/issues/ workshop/history/` stages *everything* untracked under
those dirs — not just the issue(s) the archive actually moved. So any unrelated
**untracked** issue file sitting in `workshop/issues/` (in-progress local WIP for
a not-yet-claimed issue) gets swept into the "archive completed issues to history"
commit and pushed to main.

Hit live shipping #77: an untracked `000079-doc-review-flow.md` (a separate
in-progress issue, never claimed/pushed) was captured by the archive commit
(`791e309`, 139 lines) and pushed to `origin/main` — committing the operator's
local-only WIP without intent.

This is the **same class** of bug as #78, one layer down: #78 fixed the merge
*guard* to tolerate untracked files; this is the archive *commit* still
capturing them via a broad `git add <dir>/`. The guard now correctly lets the
merge proceed past untracked files — which makes this latent broad-add reachable.

(Note: #78's guard tolerating untracked + this step committing them is a sharp
combination — the merge no longer refuses, then silently commits the untracked
file. Fixing the broad-add closes the loop.)

## Spec

Stage **only the specific paths the archive moved**, not the whole directories.
The archive loop already knows each moved issue's source (`<issuesDir>/<id>-*.md`,
now deleted) and destination (`<historyDir>/<id>-*.md`, now created) — add those
exact paths (e.g. `git add -- <src> <dst>` per moved file, or accumulate the list
and add it once). Never `git add <dir>/`.

Untracked files unrelated to the archive must remain untracked after the merge
(the operator's WIP is theirs to commit deliberately).

## Done when

- `sdlc merge` with an unrelated untracked file in `workshop/issues/` archives the
  done issue(s) WITHOUT committing the untracked file — it stays untracked after.
- The archive commit still contains exactly the moved issue(s) (src deletion +
  history addition), unchanged.
- A regression test in the `merge_e2e_test.go` harness: seed an untracked
  `workshop/issues/000XXX-unrelated.md`, run the full merge, assert the archive
  commit does NOT include it and it remains untracked. `go test ./cmd/sdlc/...`
  green.

## Plan

Single-pass atomic fix. Scope **expanded** per the plan-quality judge (ARCH-DRY):
the identical broad-add lives at three sites — `merge.go:421`, `push.go:175`
(`archiveDoneIssues` commit), `push.go:229` (`recoverInterruptedArchive`). Fixing
merge alone would fork the same archive-and-stage logic and leave the
more-traveled `sdlc push` still sweeping untracked WIP onto main. So converge all
three on the existing `preparedArchiveMove{IssuePath, HistoryPath}` model + one
shared staging helper. No Mx — closes in one `sdlc close`.

- [x] Add a pure `archiveAddArgs(moves []preparedArchiveMove) []string` → the
      precise `["add", "--", <src>, <dst>, …]` arg list (the `--` guards against
      a path being read as a flag). The exactly-moved-paths counterpart to the
      broad `git add <dir>/`. (`push.go`, near `preparedArchiveMove`.)
- [x] Change `archiveDoneIssuesInDir` (merge.go) to return `[]preparedArchiveMove`
      with **`mainPath`-relative** paths (`issuesDir/<base>`, `historyDir/<base>`)
      — `GitInDir(mainPath, "add", …)` needs repo-relative paths. Caller stages
      `mergeRunner.GitInDir(mainPath, archiveAddArgs(moves)...)`, guard on `len`.
- [x] Change `archiveDoneIssues` (push.go) to return `[]preparedArchiveMove`
      (repo-relative). Update both push sites (`archiveDoneIssues` commit +
      `recoverInterruptedArchive`, which already has `moves`) to stage via
      `archiveAddArgs`; fixed the dry-run print to show the precise add.
- [x] Update existing unit-test callers binding the old `int` return:
      `merge_test.go`, `push_test.go` (incl. the recovery test that asserted the
      old broad-add string → now asserts `add -- <src> <dst>`).
- [x] Add a regression test in `merge_e2e_test.go`: seed an unrelated untracked
      `workshop/issues/000888-unrelated.md`, run the full merge, assert the
      archive commit does NOT include it and it stays untracked; assert the
      archive commit still contains exactly the moved issue (src del + history add).
      Verified it FAILS when reverted to the broad add.
- [x] Added a direct table-driven unit test for the pure `archiveAddArgs`
      (judge's non-blocking suggestion) — pins "no directory args" independent of
      either e2e harness.
- [x] `go test ./cmd/sdlc/...` green.

## Log

### 2026-06-03
Filed from #77's ship. After #78 made the merge guard tolerate untracked files,
the archive step's pre-existing `git add <issuesDir>/ <historyDir>/` (`merge.go:421`)
became reachable and swept an untracked #79 onto main (`791e309`). Root cause:
directory-wide `git add` instead of adding the specific moved paths. Base-layer
change to `cmd/sdlc/merge.go` (`base.manifest`) — weigh downstream impact. Sibling
of [[000078]]; same broad-`git add` hazard recorded in `lessons.md`.

### 2026-06-04 (implementation)
- 2026-06-04: closed — go test ./cmd/sdlc/... green (incl. new TestArchiveAddArgs + TestRunMerge_ArchiveDoesNotSweepUntrackedIssue); regression test verified to FAIL when merge reverted to broad add, pass with fix; go vet clean. Converged 3 broad-add sites onto shared archiveAddArgs helper.; review verdict: SHIP
Plan-quality judge (ARCH-DRY) flagged the identical broad-add at three sites, not
one: `merge.go:421` + `push.go:175` (`archiveDoneIssues` commit) + `push.go:229`
(`recoverInterruptedArchive`). Fixing merge alone would fork the archive-stage
logic and leave the more-traveled `sdlc push` still sweeping WIP. Converged all
three on the existing `preparedArchiveMove{IssuePath, HistoryPath}` model + one
pure shared helper `archiveAddArgs(moves) → ["add", "--", <src>, <dst>, …]`. Both
archive scanners (`archiveDoneIssues`, `archiveDoneIssuesInDir`) now return
`[]preparedArchiveMove` instead of an `int`; merge's paths are **`mainPath`-
relative** (most error-prone line — an absolute path would silently miss the
staged move; pinned by a merge_test assertion). Regression verified: reverting
merge to the broad add → e2e fails (sweeps 000888); restored → green. `go test
./cmd/sdlc/...` green. Atlas `sdlc-binary.md` archive-recovery paragraph updated
(was stale: "stages issues/ and history/"). lessons.md Rule 2 already named #80.
