---
id: 000196
status: open
deps: []
github_issue:
created: 2026-08-20
updated: 2026-08-20
estimate_hours:
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

## Spec

The exemption must be conditioned on the mechanical question, not the semantic
one: **would `git switch main` tolerate this file?**

Preferred fix — **commit tracker files rather than exempting them.** `sdlc`
already owns tracker state and already auto-commits it elsewhere as
`issue-sync: update issues` (`sdlc issue new`, `sdlc claim`). Doing the same at
the top of merge removes the hazard entirely instead of reasoning around it, and
is consistent with how every other verb treats these files. The semantic claim
behind #82 ("this is not code contention") stays true and becomes the *reason*
committing is safe, rather than the reason for skipping the check.

Alternatives, if that is rejected:

- Narrow the exemption to files that would not be overwritten — i.e. those whose
  content matches `main`'s. Correct but needs a per-file diff against the merge
  target, and gets subtle when the merge itself changes those files.
- Stash the tracker paths around the switch. Works, but leaves a stash to lose.

Either way the **recovery path must be reported**. A merge that succeeds remotely
and fails locally is the worst state this tool can produce, and the operator is
currently told only `git switch main: exit status 1`. It should name what
happened ("the PR is merged; this checkout is stranded") and the exact remaining
steps.

## Done when

- [ ] A dirty tracker file cannot strand a merge — either committed first, or
      genuinely proven safe for the switch.
- [ ] `assessDirty`'s comment and behaviour agree; today the comment names the
      hazard the classification then ignores.
- [ ] A test reproduces the observed sequence: dirty tracker file, merge, switch
      — asserting the merge either refuses up front or completes end to end.
      Never "remote merged, local stranded".
- [ ] If the local half fails after a successful remote merge, the error states
      the state and the recovery commands.
- [ ] `#82 M2`'s intent survives: a dirty issue file still never *blocks* the
      operator; it stops being left dirty across a branch switch.

## Plan

- [ ] Design via `sdlc start-plan` before implementing.

## Log

### 2026-08-20

Third of three filed from a session in the `tools` repo, with `#194` (review
blocks the tree) and `#195` (reviews have no memory).

The bug is narrow and the reasoning behind it was sound in isolation: #82 M2
asked "is this file code contention?" — a question about meaning — where `git
switch` asks "does this file differ?" — a question about bytes. Both guards in
merge route through `Refuse()`, so the semantic answer is the only one either
guard ever consults.
