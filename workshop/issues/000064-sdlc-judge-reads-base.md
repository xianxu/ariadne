---
id: 000064
status: working
deps: []
github_issue:
created: 2026-06-02
updated: 2026-06-03
estimate_hours: 0.75
---

# sdlc push plan-completeness judge reads pre-merge base, not HEAD — blocks close-and-archive pushes

## Problem

`sdlc push` runs a pre-merge "Check issue plan completeness" judge that
evaluates the issue files as they exist on `origin/main` (the pre-merge
base), **not** as they exist at the branch HEAD being pushed.

This makes a push whose *purpose* is to close-and-archive issues
structurally unsatisfiable: at the base, the very issues you are closing
are still `status: working` with unticked Plan boxes (that's the state
your commits fix). The judge reports them as incomplete, fails, and
aborts — even though HEAD has them `status: done`, fully ticked, with
`## Log` close entries.

Observed 2026-06-02 closing the nous `shared-brain` project (nous
#5/#16/#20/#30/#31/#32):

- HEAD: every issue `status: done`, Plan boxes ticked/`[-]`-deferred,
  `## Log` close entries present (verified with `git show HEAD:<path>`).
- origin/main: same issues `status: working`, boxes `[ ]`, empty Logs.
- Judge output described the **origin/main** state verbatim — it kept
  flagging "#26/#27 Log empty" and "#30/#32 working" *after* those were
  fixed and committed at HEAD. Re-running with new HEAD commits changed
  nothing, confirming it never reads HEAD.

Net effect: the only way to land a legitimate close-and-archive push is
`sdlc push --no-judge`, which disables *all* pre-merge judges (specs,
lessons, plan) — throwing out the judges that would still add value.

## Spec

The plan-completeness judge should evaluate the **post-merge result**
(i.e. the pushing branch HEAD, or the merge of HEAD into main), not the
current `origin/main`. Concretely:

- For the issue files touched in the push range, read their **HEAD**
  content when judging Plan/status/Log completeness.
- A judge that wants to catch "you're about to archive an incomplete
  issue" should still read HEAD — the archive + status flip live in the
  same push, so HEAD is the authoritative "what will main look like
  after this lands" view.

Likely the judge resolves paths against `workshop/issues/` on the base
ref, or shells `git show origin/main:<path>` / reads the working-tree
base, instead of the branch tip. Find where it loads issue content and
point it at HEAD (or the computed merge result).

Secondary consideration: issues that move `issues/ → history/` in the
same push. The judge must follow the rename (read the HEAD path in
`history/`), not the base path in `issues/`, or it'll see the
pre-archive version.

## Done when

- A close-and-archive push (issues flipped to `done` + moved to
  `history/` + Logs added, all at HEAD) passes the plan-completeness
  judge without `--no-judge`.
- The judge still correctly fails when HEAD genuinely archives an issue
  with unticked Plan items and no terminal status / Log rationale
  (i.e. it didn't just get disabled — it reads the right side now).
- A regression test pins: base=working/unticked, HEAD=done/ticked →
  judge PASS.

## Root cause (located 2026-06-03)

Not "reads base vs HEAD" exactly — the **diff filter doesn't follow the
`issues/ → history/` archive rename**. `collectDiff` (cmd/sdlc/judge.go) builds
the Plan judge's inputs with a path filter of **`issuesDir/*.md` only**
(lines 193, 210). When the close+archive is *already committed* in the judged
range, `git diff … -- issues/*.md` shows the issue as **deleted** (moved to
history/); the agent is handed an `issues/` path that no longer holds the
content and reads the stale base version → false "incomplete".

Why it's been dormant: the default `sdlc merge` runs the Plan judge at **step 5,
before archiving at step 11** — so closed issues are still `done`-in-`issues/`
at judge time and pass. The bug bites only when the archive move is pre-committed
in the range (the nous `shared-brain` push it was filed from).

## Plan

- [ ] Fix `collectDiff` (Plan case): include `historyDir/*.md` in the path
  filter, and use `--diff-filter=d` on the name-only pass so the changed-files
  list excludes the **deleted** `issues/` side and keeps the **HEAD-existing**
  path (`history/NNN.md` for an archived issue, `issues/NNN.md` for an in-place
  close). The agent then reads the live done-content, not a stale deleted path.
  Add a temp-repo regression test pinning: base=working-in-issues/,
  HEAD=done+archived-to-history/ → `changedIssues` is the `history/` path (not
  the deleted `issues/` path) and the diff carries the done content.

## Log

### 2026-06-02

Filed from the nous shared-brain project close-down. Workaround used at
the time: `sdlc push --no-judge --yes` after manually verifying HEAD
state was correct (issues done, ticked, logged) via `git show HEAD:`.
The push otherwise had nothing wrong with it — the judge was reading the
wrong side of the merge.
