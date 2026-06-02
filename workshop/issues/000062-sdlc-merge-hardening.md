---
id: 000062
status: open
deps: []
github_issue:
created: 2026-06-02
updated: 2026-06-02
estimate_hours: 3
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

## Spec

Four candidate fixes; the recommended set is **#1 + #2 + #3** (see "On #3 vs #4"
for why #4 is largely redundant once #1 lands):

1. **Re-assert clean tree immediately before `gh pr merge`** (after judges run).
   Refuse cleanly if dirty: "pre-merge judges left uncommitted changes in X —
   commit and re-run." Converts the worst failure (remote-merged / local-
   stranded) into a safe **pre-merge refusal**. Highest value, smallest change.
2. **Pre-merge judges must be side-effect-free** (read-only / suggest-only):
   report findings, let the agent apply + commit + re-run. A gate that silently
   writes files and then passes is the core design smell. If auto-fix is wanted,
   sdlc must deterministically stage+commit the judge's edits as part of the
   merge — never leave them loose.
3. **Recoverable post-merge cleanup.** Everything after `gh pr merge` (switch /
   pull / archive / branch-delete) must be idempotent and resumable: a re-run
   (or `sdlc merge --resume`) detects "PR already merged" and finishes the local
   cleanup instead of erroring on "no open PR." Crossing the point-of-no-return
   must be **tool-recoverable, not hand-recoverable**.
4. *(demoted — see below)* Auto-stash around the switch (`stash → switch → pull
   → pop`) so a dirty tree degrades to a warning, not an abort.

### On #3 vs #4 (operator's question)

They cover different scopes, so #4 does **not** subsume #3:

- **#4** only tolerates *one* cause of post-merge failure: a dirty tree at switch
  time. **#3** recovers from *any* interruption after the irreversible merge —
  a transient `git pull` failure, an archive conflict, a Ctrl-C, a detached
  state, etc. #3 is the general "make the irreversible boundary recoverable" net;
  #4 is a single-cause band-aid.
- The real redundancy is **#1 vs #4**, not #3 vs #4. **#1** prevents a dirty tree
  from ever reaching `gh pr merge`; with #1 (and #2, which stops the judge
  dirtying the tree at all) in place, the tree is clean at switch time, so **#4
  becomes unnecessary**.

So: do **#1 + #2** (root-cause: never cross the boundary dirty) and **keep #3**
(the recovery net for everything else). **Drop #4** (or keep only as cheap
defense-in-depth). #3 is exactly what would have turned the manual recovery into
a one-command `sdlc merge --resume`.

## Done when

- `sdlc merge` re-checks the working tree is clean immediately before
  `gh pr merge` and refuses with an actionable message if not.
- Pre-merge judges do not leave the working tree dirty on a passing verdict
  (read-only, or their edits are committed deterministically by the verb).
- An interrupted-after-`gh pr merge` state is recoverable by re-running the verb
  (detects merged PR → completes switch/pull/archive/branch-delete).
- A regression test exercises: judge dirties tree → merge refuses pre-merge
  (not post-merge); and resume-after-merge finishes cleanup.

## Plan

- [ ] M1 — #1: re-assert clean tree before `gh pr merge`; actionable refusal.
- [ ] M2 — #2: make pre-merge judges side-effect-free (or commit their edits).
- [ ] M3 — #3: idempotent/resumable post-merge cleanup (`--resume` / detect
  already-merged PR).

## Log

### 2026-06-02

Filed from the nous PR #1 merge post-mortem. Base-layer fix (`cmd/sdlc` merge
verb). #4 (auto-stash) deliberately demoted: redundant once #1 lands; #3 is the
load-bearing recovery net and is NOT subsumed by #4.
