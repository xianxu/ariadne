---
id: '000062'
status: done
created: 2026-06-02
updated: 2026-06-02
estimate_hours: 3
actual_hours: 0.7
---

# sdlc merge: re-check preconditions before irreversible PR merge + recoverable cleanup

## Problem

`sdlc merge` stranded a real merge (nous PR #1, 2026-06-02): it merged the PR
server-side (irreversible) and then aborted the local cleanup, leaving remote
main merged but the local checkout stuck on the feature branch with a dirty
file. Recovery was fully manual (stash → switch → pull → pop → archive issues →
delete branch).

### Timeline
1. Start-of-flow refusal checks ran, incl. **"working tree clean"** — passed
   (tree was clean at that moment).
2. **Pre-merge judges** ran. The atlas/specs judge is **write-capable** — it
   edits stale docs in place. On the final attempt it modified an atlas file and
   **left it uncommitted while returning a passing INFO verdict.**
3. Judges "passed" → `gh pr merge` (server-side, **irreversible**) → merged.
4. `git switch main` (in-place cleanup) → **refused**: "local changes … would be
   overwritten by checkout" (the judge's uncommitted edit). Merge aborted here.
5. Remote = merged; local = stranded; archive/switch/pull/branch-delete never
   ran.

### Root cause
A **gating step (the judge) mutated the working tree**, and the verb **crossed
its irreversible boundary (`gh pr merge`) without re-asserting the clean-tree
precondition** that the subsequent `git switch` depends on. The invariant was
checked once at the top, violated by sdlc's own judge, never re-checked, and the
irreversible action ran anyway — with no recovery path afterward.

General principle being violated: *order the irreversible action last, and gate
it on a fresh re-check of every precondition it depends on — nothing between the
initial guard and the irreversible step (judges, hooks, linters) may invalidate
a precondition unchecked.*
