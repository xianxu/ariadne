---
id: 000196
status: codecomplete
deps: []
github_issue:
created: 2026-08-20
updated: 2026-10-12
estimate_hours:
card_mirror: '6064d1dd103dfac6d1a77407cf5a749057d0b6c4' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T07:50:13-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "e515125c", done: "6151fd84"}
actual_hours: 0.92
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

Scope today (2026-10-12): primary and numbered-slot checkouts land through
`runDurableMerge` (`landing.go`), whose `landingClean` already refuses *any*
tracked dirt (tracker files included) before the merge, so the hazard is closed
there. It survives on the legacy `runMerge` path's **in-place** topology
(reachable as a `dependency` checkout). The linked-worktree topology never
switches the issue checkout, so its exemption stays.

- [x] Resolve topology (in-place vs worktree) before step 2, so both dirt guards know it.
- [x] One decision for steps 2 and 9b: `mergeRefusal` = Blocking, plus (in-place
      only) the tracker paths that differ between `main` and `HEAD` or between
      `main` and `origin/main` (git would refuse them on switch or on pull).
      Step 9b fetches `origin/main` first; a diff error refuses (fail safe).
- [x] Remaining tracker dirt still warns, never blocks (#82 M2 intent).
- [x] Post-merge `git switch main` / `git pull` failure: say the PR is merged and
      this checkout is stranded, and give the recovery (fix the named files, re-run
      `sdlc merge --yes`, which resumes via #62 M3).
- [x] Rewrite `assessDirty`/`Refuse` comments to agree with the behaviour.
- [x] Tests: unit for the stranding-path selection; e2e in-place: tracker file
      that differs on main refuses pre-merge (PRMerge never called); identical
      tracker file still lands end to end; switch failure after merge names the
      recovery.

## Log

### 2026-10-12
- 2026-10-12: closed — make test on 505a771a..e1703785: 12 shards, 1092 cmd/sdlc tests + 46 packages, 0 failed. After review round 1 fixes (7c48c569): make test 0 failed except two load-sensitive tests outside merge (TestCLISignalCancelsOwnedReviewer, TestLegacyRepoLockSetStatusMutationWaits; known flake class #325), both pass 3/3 in isolation; merge tests all pass. New e2e: dirty tracker file differing from main refuses pre-merge, PRMerge never called; main moving the file after step 2 refused at 9b after fetch; switch and pull failures after merge name the merged state and the resuming re-run, never a commit. Mutation: disabling the in-place tracker check fails all three new refusal tests.; review verdict: SHIP

Claimed at ariadne:2 (Merge and CI train). The spec's preferred fix, committing
tracker files at the top of merge, no longer fits: merge requires HEAD synced
with origin (step 4) and unchanged since close (#160 publish gate), so a commit
made by merge itself refuses both. Taking the spec's mechanical alternative
instead (exempt only what git would carry across switch and pull).

### 2026-08-20

Third of three filed from a session in the `tools` repo, with `#194` (review
blocks the tree) and `#195` (reviews have no memory).

The bug is narrow and the reasoning behind it was sound in isolation: #82 M2
asked "is this file code contention?" — a question about meaning — where `git
switch` asks "does this file differ?" — a question about bytes. Both guards in
merge route through `Refuse()`, so the semantic answer is the only one either
guard ever consults.

Implemented (`cmd/sdlc/merge.go`): `mergeRefusal` is the single decision for
steps 2 and 9b. `strandingTracker` asks git's question with two path-limited
diffs (`main..HEAD` for the switch, `main..origin/main` for the pull; the
branch's own changes are in the first). The guard skips the diff when there are
no tracker paths, since a bare `git diff --` lists every file. A stale
origin/main at step 2 only weakens the early check; 9b fetches first.
`strandedMessage` covers both post-merge failures; a pull failure in place has
already switched, so it says `git switch <branch>` before the re-run.
Mutation check: disabling the in-place branch of `mergeRefusal` fails all three
new refusal tests. The Done-when bullet "stops being left dirty across a branch
switch" now holds as "is carried across only when git can carry it": a tracker
file that matches main still rides the switch dirty, which is #82 M2's intent.

Close review round 1 (FIX-THEN-SHIP): BR-1 Important, the stranded and 9b
messages advised committing, which the resume path refuses (step 4, #148). Both
now say restore or move aside, and the test pins "do not commit". Minors fixed
in the same round: the tracker warning names the real reason per topology; the
code-only `Refuse()` is removed (mergeRefusal is the one decision); pathspecs
are pinned `:(top,literal)` so a subdirectory run or a glob character can't
slip a stranding file through (pinned in the unit fake). Lesson added.
