---
id: 000196
status: open
created: 2026-08-20
updated: 2026-08-20
estimate_hours:
github_issue:
---

# merge: dirty tracker files are exempted from the block but still strand the post-merge switch

## Problem

`sdlc merge` classifies a dirty working tree into three buckets
(`assessDirty`, `cmd/sdlc/merge.go:131-155`): `Blocking` (tracked code),
`Untracked`, and `Tracker` (workshop issue/history markdown). Only `Blocking`
refuses (`Refuse()`, `merge.go:147`).

The exemption for `Tracker` (#82 M2) is justified semantically — tracker state is
append-only and syncs to main out-of-band, so a dirty issue file "is never code
contention". But the hazard it is exempting itself from is **mechanical**, and
the same doc comment states it two lines earlier:

> *"Only tracked CODE modifications block a merge: a dirty tracked file makes the
> post-merge `git switch main` / `git pull` refuse, stranding the server-side
> merge."* — `merge.go:151-153`

`git switch` does not know a file is "tracker state". It refuses to check out
whenever a locally-modified tracked file differs between the branches, whatever
the file means to us. So the exemption's premise contradicts the hazard its own
comment names.

### Observed (tools repo, 2026-08-20)

`sdlc claim --issue 2` was run while still on branch `000001-define`; it flipped
the status but its sync failed, leaving `workshop/issues/000002-repl.md`
modified. Then `sdlc merge --yes`:

```
[!] 1 tracker file(s) dirty — not blocking the merge (tracker state syncs to main out-of-band, #82):
M workshop/issues/000002-repl.md
[ok] No uncommitted tracked changes
==> Merging PR #1 (000001-define) into main via GitHub...
==> Switching to main...
Error: git switch main: exit status 1
error: Your local changes to the following files would be overwritten by checkout:
	workshop/issues/000002-repl.md
```

The PR **was merged on GitHub**. The local checkout was stranded on the feature
branch, the `codecomplete → done` flip never ran, and the issue was not archived.
Recovery took committing the file by hand, `git switch`, `git pull`, and then
`sdlc push` to finish the publish flip — none of which the tool suggested.

### Why the #62 re-check does not catch it

`worktreeDirty` is deliberately re-asserted immediately before the irreversible
`gh pr merge` (`merge.go:118-122`), precisely so a hook that dirties the tree
"converts that into a clean pre-merge refusal". But the re-assert asks
`Refuse()`, which exempts tracker files by construction — so this class passes
both guards. The invariant #62 established is real; this is a hole beneath it.
