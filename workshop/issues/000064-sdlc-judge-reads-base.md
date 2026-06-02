---
id: 000064
status: open
deps: []
github_issue:
created: 2026-06-02
updated: 2026-06-02
estimate_hours:
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

## Plan

- [ ] M1 — locate where the plan-completeness judge loads issue content
      (the verb behind `sdlc judge plan` / the pre-merge hook in
      `sdlc push`); confirm it reads base, not HEAD.
- [ ] M2 — switch it to read HEAD (and follow `issues/ → history/`
      renames in the push range); add the base→HEAD regression test.
- [ ] M3 — verify against a synthetic close-and-archive push; confirm a
      genuinely-incomplete archive still fails.

## Log

### 2026-06-02

Filed from the nous shared-brain project close-down. Workaround used at
the time: `sdlc push --no-judge --yes` after manually verifying HEAD
state was correct (issues done, ticked, logged) via `git show HEAD:`.
The push otherwise had nothing wrong with it — the judge was reading the
wrong side of the merge.
