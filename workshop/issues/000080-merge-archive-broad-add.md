---
id: 000080
status: working
deps: []
github_issue:
created: 2026-06-03
updated: 2026-06-04
estimate_hours: 1.5
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

- [ ]

## Log

### 2026-06-03
Filed from #77's ship. After #78 made the merge guard tolerate untracked files,
the archive step's pre-existing `git add <issuesDir>/ <historyDir>/` (`merge.go:421`)
became reachable and swept an untracked #79 onto main (`791e309`). Root cause:
directory-wide `git add` instead of adding the specific moved paths. Base-layer
change to `cmd/sdlc/merge.go` (`base.manifest`) — weigh downstream impact. Sibling
of [[000078]]; same broad-`git add` hazard recorded in `lessons.md`.
