---
id: 000220
status: open
deps: []
github_issue:
created: 2026-09-11
updated: 2026-09-11
estimate_hours:
---

# Issue publishes from a feature branch: false success, empty subjects, add/add at merge

## Problem

Field evidence from parley.nvim#227 (branch `000227-…`, PR #172, 2026-09-10).
Every defect below is in the **trunk path that ariadne#207 adds and has not
merged**. parley ran it because the `sdlc` shell function rebuilds from
ariadne's *working tree* on every call, and that tree is on
`000207-sync-without-worktree`. Line numbers are against that branch at
`671f0f6`.

1. **Empty commit subjects.** `sdlc issue new` from a feature branch published
   blank templates to parley `main` as `1ead2e2`, `4905785` and `c955972`, all
   with empty subjects. The trunk path passes `msg` raw
   (`claim.go:158` → `synctrunk.go:107`), while the on-main path defaults it
   through `syncMessage` (`claim.go:313`).
2. **A publish that never happened, reported as a publish.**
   `sdlc claim --issue 234 --no-start` on the branch printed "Issue changes
   published to the trunk." and `synced`, but nothing reached `main`: its
   first-parent line has no commit between `4905785` (18:15:53) and `c955972`
   (18:32:36). There are two causes.
   - `changedIssueFiles` diffs the working tree against **HEAD**
     (`claim.go:336-341`). The filled #234 had already been committed on the
     branch, so it looked unchanged — but on the trunk path the question is
     what differs from the *trunk*.
   - `syncViaTrunkWithRealloc` returns `(nil, nil)` after "No issue changes to
     sync.", and `syncViaTrunk` then prints the success line and `synced`
     unconditionally (`synctrunk.go:59-62`).
3. **add/add at merge.** GitHub refused PR #172 as "not mergeable" on
   `000233`, `000234` and `000235`. The trunk held the blank stubs from (1); the
   branch held the filled files, committed independently. With no common
   ancestor for those paths, git cannot merge them. It was resolved by hand
   (`63390fe`).
4. **`sdlc pr` is not idempotent.** On a branch that already has an open PR it
   pushes, then `gh pr create` fails with "already exists", and it exits nonzero
   (`pr.go:107-115` never checks for an existing PR).

Two further observations from the same session:

- The `sdlc` function builds from ariadne's working tree, so an in-flight
  ariadne branch silently changes `sdlc` for every repo on the machine. That is
  how (1)–(3) reached parley before #207 merged.
- ariadne's local `main` is 8 commits ahead of `origin/main`: #207's
  `change-code` spec/plan syncs `8970d88`…`ff1b52b` and side-quest `59debf4`,
  all from 2026-09-09. Any on-main `sdlc` publish would push them wholesale —
  the exact hazard `PublishExisting`'s comment warns about. This issue was filed
  by hand from a detached `origin/main` worktree to avoid both that and (1).

### Root cause: two writers, unrelated histories

A trunk publish is a commit built on `origin/main`'s tip, and the publishing
branch never contains it. The branch keeps committing the same paths, because
the files are in its working tree. Git merges cleanly only where the two sides
share an ancestor for a path, or happen to agree byte-for-byte.

(3) is the add/add form of this. The branch's *own* issue file is the
modify/modify form. It merged cleanly on #227 only because the trunk stopped
writing it after the branch point (`555b81a`); any later trunk publish of that
file would have conflicted the same way. (2) hides the divergence instead of
closing it.

## Spec

**Proposed — awaiting the operator's decision.** The operator's constraints
(2026-09-11):

- **(a)** an issue is just a markdown file in the repo;
- **(b)** `origin/main` always shows at least the rough lifecycle — created,
  working — which is the lock.

These constraints do not conflict. What conflicts is letting two histories
write the same file. The proposal: **the trunk holds the lock, the branch holds
the story, and the two always share ancestry.**

1. **One writer per issue file at a time.** The trunk owns an issue file from
   `issue new` until the branch point, and again after merge. The branch named
   `NNNNNN-*` owns issue NNNNNN in between. Every other issue file stays
   trunk-owned, even when it is edited from that branch's checkout.
2. **A publish from a branch is folded back into it.**
   - Build the publish commit P on `merge-base(HEAD, origin/main)`, not on
     `origin/main`'s tip.
   - Land it on `main` as `merge(origin/main, P)` under the existing CAS, then
     merge P into the branch.
   - P touches only issue files, and its parent is already an ancestor of the
     branch. So the branch gains exactly that change — a doc-only delta the
     publish gate already accepts — and none of `main`'s other code enters the
     review window.
   - From then on trunk and branch share P, so neither add/add nor silent
     divergence is possible.
   - A conflict while landing P on `main` means someone else wrote that issue
     since the merge-base. Surface it at publish time as a lock violation, not
     at PR time.
3. **Decide what to publish against the trunk**: working tree vs
   `origin/main`'s blob, never `HEAD`. A publish that writes nothing reports a
   no-op, keeping the machine-readable marker's contract.
4. **`sdlc merge` checks before GitHub does.**
   - A foreign issue file that differs from the trunk: fold it back, or refuse
     and name it.
   - The branch's own issue changed on the trunk since the branch point: refuse
     as a lock violation.
   - GitHub's opaque "not mergeable" becomes a named local refusal with the fix.
5. **Fix regardless of the design:** trunk commits always carry a subject that
   names the issue ids and the verb (reuse `syncMessage`), and `sdlc pr` exits 0
   with the existing PR's URL when one is open.

**Rejected alternatives:**

- **Lock in a separate trunk-only file or git ref.** Status would live outside
  the markdown — two sources of truth, which violates (a).
- **Issue files trunk-only, never committed on branches.** The branch
  checkout's working tree stays dirty forever, and the issue's Plan/Log drop out
  of the reviewed PR.
- **A custom git merge driver for issue files.** GitHub's server-side merge,
  which `sdlc merge` relies on, ignores it.

**Open question for the operator:** should the `sdlc` function build from a
pinned ref (`origin/main`, or a clean dedicated worktree) rather than whatever
branch ariadne's checkout is on?

## Done when

- The #227 sequence is a test against a real bare origin: `issue new` on a
  feature branch → fill → commit/sync → PR merge. After every publish,
  `git merge-tree --write-tree origin/main HEAD` is conflict-free under
  `workshop/issues/`.
- A publish that writes nothing reports a no-op, and no trunk commit has an
  empty subject.
- `sdlc merge` refuses before calling GitHub when an issue file would conflict,
  naming the file and the fix.
- `sdlc pr` with an open PR exits 0 and prints that PR.
- Whichever design the operator picks is recorded in the Spec, with the other
  options' rejection reasons kept.

## Plan

- [ ] Operator decides the Spec: the fold-back design, or an alternative.
- [ ] Small fixes, independent of the design: trunk-path message fallback, a truthful no-op, `pr` idempotence. Candidates to fold into #207 before it merges.
- [ ] Decide what to publish against the trunk, not `HEAD`.
- [ ] Merge-base-parented publish commit, landed as a CAS merge on main and folded back into the branch.
- [ ] Pre-merge issue-file divergence gate in `sdlc merge`.
- [ ] Regression test that replays the #227 sequence against a bare origin.

## Log

### 2026-09-11

Filed from parley.nvim#227's wrap-up. Evidence was confirmed against parley's
`origin/main` first-parent history and the #207 branch source. Filed by hand
(`sdlc issue new --dry-run` for the id, then a fast-forward push from a detached
`origin/main` worktree): every `sdlc` publish route available here was either
defect (1) or a wholesale push of local `main`'s 8 unpushed commits.
