---
id: '000064'
status: done
created: 2026-06-02
updated: 2026-06-03
estimate_hours: 0.75
actual_hours: 0.75
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
