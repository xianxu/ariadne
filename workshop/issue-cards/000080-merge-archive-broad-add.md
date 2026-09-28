---
id: '000080'
status: done
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
