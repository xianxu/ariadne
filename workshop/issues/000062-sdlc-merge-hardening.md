---
id: 000062
status: done
deps: []
github_issue:
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

- [x] M1 — re-assert clean tree before `gh pr merge` (step 9b) via the extracted
  `worktreeDirty`; actionable refusal instead of merge-then-strand.
- [x] M2 — judges read-only: `Category.AllowedTools()` is `Read,Grep,Glob,Bash`
  for ALL categories (Specs lost `Edit,Write`); Specs prompt rewritten to REPORT
  stale docs for the main agent to fix. Chose read-only over auto-commit
  (operator: fresh/different-model judge has less context than the doer).
- [x] M3 — resumable cleanup: `ghCaller.PRMergedForBranch` + `decideMergeAction`;
  a re-run detects an already-merged PR and finishes switch/pull/archive/delete
  (auto-detect on plain re-run, no flag). #4 (auto-stash) dropped as planned.

## Log


- 2026-06-02: closed — M1+M2+M3 implemented; full sdlc suite green (worktreeDirty, decideMergeAction, all-read-only AllowedTools); fresh-eyes review SHIP — wiring manually verified (PRMergedForBranch matches post --delete-branch; switch-main no-op when on main; idempotent resume). --force: M1-M3 reviewed in one combined pass (not 3 separate verdict commits). e2e harness deferred (fix-forward, operator-approved)
### 2026-06-02

Filed from the nous PR #1 merge post-mortem. Base-layer fix (`cmd/sdlc` merge
verb). #4 (auto-stash) deliberately demoted: redundant once #1 lands; #3 is the
load-bearing recovery net and is NOT subsumed by #4.

Implemented M1+M2+M3 (commit d3ac25f). Fresh-eyes review verdict **SHIP** — no
Critical/Important; reviewer *manually verified* the load-bearing behaviors:
`PRMergedForBranch` (`gh pr list --head --state merged`) still matches after
`--delete-branch` (GitHub retains headRefName); `git switch main` is a no-op
when already on main (resume doesn't die); re-run archive/branch-delete are
idempotent. 3 Minors noted (stale `origin/<branch>` ref can falsely refuse if
the operator `fetch --prune`d before resuming; `.[0]` PR ordering; error-swallow
in PRMergedForBranch) — all narrow/consistent-with-existing-posture, not fixed.

**Test-coverage caveat (fix-forward, per operator):** the Done-when asks for
end-to-end regression tests (dirty→refuse-pre-merge; resume-finishes-cleanup).
Shipped coverage is decision-level unit tests (`worktreeDirty`,
`decideMergeAction`, all-read-only `AllowedTools`) + the reviewer-verified
wiring. A full `runMerge` e2e harness would be the first of its kind (every
`run*` verb uses `os.Exit` via `die`, which is why none are e2e-tested) —
deferred as a separate test-infra investment (make `die` injectable + stub
gitx.run/mergeRunner/ghClient). We'll add it (or fix forward) when a real
workflow surfaces a wiring bug.
