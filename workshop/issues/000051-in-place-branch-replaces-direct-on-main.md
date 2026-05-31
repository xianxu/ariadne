---
id: 000051
status: open
deps: []
created: 2026-05-30
updated: 2026-05-30
estimate_hours: 3
---

# In-place branch workflow replaces direct-on-main

## Problem

There are three publish modes today, and they overlap awkwardly:

1. **Direct-on-main** — work + commit on `main`, `sdlc push` (`make push`) ships it.
2. **Worktree** — `sdlc change-code --worktree=yes` (separate checkout dir) → `sdlc pr` → `sdlc merge`.
3. **In-place branch** — `sdlc change-code --worktree=no` already creates a branch that carries the working tree forward in the *same* directory. But it's a half-citizen: the **merge-back path is worktree-shaped** (`sdlc merge` / `make merge` locate and clean a worktree — `git worktree list | grep '[main]'`), so there's no clean "merge this in-place branch and switch back to main" verb.

We want to **retire direct-on-main and make in-place branch the default**, with worktree as the opt-in heavyweight (true isolation, parallel work). Rationale from the 2026-05-30 design discussion (you-decide review-gate thread):

- **Commit is a sync/handoff primitive, not a readiness claim.** Committing preserves state and hands off between agents; it doesn't mean "ready." So commits (and branch pushes) should be free, and the readiness boundary is *publishing to main*.
- A **branch is the transient staging lane** that the publish step evaluates — `main` is the published truth. Direct-on-main collapses the lane and the truth into one ref, which is why "commit ≠ ready" had nowhere to live.
- An **in-place branch in the same dir** is ergonomically ~identical to working on main (one `switch -c` up front, one merge at the end — both agent-automatable), so this is a *simplification*, not added ceremony. Worktrees are only worth their setup cost for genuine parallelism.

**Branch protection is explicitly out of scope here.** The operator does not run agents fully autonomously and will notice whether work is on a branch, so the value is *workflow consistency*, not enforcement. (Server-side enforcement — branch protection + a required publish-gate check — is a separate, optional layer tracked in you-decide#4 / its M3.)

## Spec

Make **in-place branch** the standard flow:

- `sdlc change-code` default branching → **in-place** (currently it asks; default should lean no-worktree). Worktree stays available via `--worktree=yes`.
- **In-place merge-back verb.** `sdlc merge` (or a sibling) gains a mode that, when run from a feature branch in the main checkout (not a worktree): merges the branch into `main` in place, switches back to `main`, deletes the merged branch, archives completed issues — *without* any worktree-cleanup machinery. Detect in-place vs. worktree automatically.
- **Retire direct-on-main.** `sdlc push` (ship-from-main) and `make push` are demoted/removed; the standard close path becomes branch → merge. (Decide: delete, or keep as a guarded fallback that warns "you're on main; start a branch.")
- The merge-to-main itself: since branch protection is out of scope, a **local merge + push** is acceptable (no PR required); keep `sdlc pr`/PR-merge as the worktree/collaborative path. Confirm during the audit whether one merge verb can cover both.

Branch stays a **transient lane** (deleted post-merge), not a durable tag; durable "published state" markers, if ever wanted, are real `git tag`s.

## Plan

**Audit (map every direct-on-main / worktree assumption):**
- [ ] `AGENTS.md` — §0 synchronization ("commit, push to origin before commencing"), §2 workflow, the "Close / ship — `sdlc push` (main) or `sdlc pr` → `sdlc merge` (branch)" narrative, `change-code` branching defaults, Task Management. Rewrite so branch-based is the default and direct-on-main is gone.
- [ ] `sdlc` — `push`, `merge`, `pr`, `change-code` (the `--worktree` default + the in-place-vs-worktree detection). Identify what assumes a worktree on merge.
- [ ] `Makefile.workflow` — `push`, `merge`, `worktree`, `pre-merge` targets (base layer; consider downstream impact).
- [ ] `atlas/workflow/*` — the workflow docs that describe push/merge/worktree.
- [ ] `construct/base.manifest` — only if make targets are added/removed.

**Implement:**
- [ ] `sdlc` in-place merge-back (auto-detect; no worktree cleanup; archive done issues; switch to main + delete branch).
- [ ] `sdlc change-code` default → in-place branch.
- [ ] Retire/guard `sdlc push` + `make push`.
- [ ] Update `AGENTS.md`, `Makefile.workflow`, `atlas/workflow/*` to the new default.

**Verify:**
- [ ] End-to-end: `claim` → `change-code` (in-place) → commits → in-place merge → back on main, branch deleted, issue archived — no worktree created at any point.
- [ ] Worktree path still works via `--worktree=yes`.
- [ ] Downstream (you-decide etc.) inherits cleanly via base-layer refresh.

## Notes / cross-refs

- Builds on **#39** (`sdlc start` split into `claim` + `change-code`; worktree decision deferred) and **#40** (workflow management → sdlc). Related: **#36** (sdlc-push-followups).
- Downstream **you-decide#4** (pre-push/PR publish gate): orthogonal — that's the optional *enforcement* layer; this ticket is the *workflow shape*. The publish gate composes with whatever merge path this lands on.
- Primitive already present: `sdlc change-code --worktree=no` (in-place branch, carries working tree forward) — this ticket promotes it to default and completes its merge-back.
