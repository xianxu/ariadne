---
id: 000051
status: done
deps: []
created: 2026-05-30
updated: 2026-06-03
estimate_hours: 3
actual_hours: 4
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
- **In-place merge-back verb.** In-place and worktree branches **share the same publish path** — push the branch, open a PR, merge server-side (`sdlc pr` + `sdlc merge`) — so the PR-only CI merge-check (ariadne #52) actually runs on the publish. They differ *only* in whether a worktree dir exists: the in-place mode skips worktree creation/cleanup and, after the GitHub merge, switches the single checkout back to `main` and `git pull`s the merged result, then deletes the branch + archives completed issues. Detect in-place vs. worktree automatically.
- **Retire direct-on-main.** `sdlc push` (ship-from-main) and `make push` are demoted/removed; the standard close path becomes branch → merge. (Decide: delete, or keep as a guarded fallback that warns "you're on main; start a branch.")
- The merge-to-main itself goes through a **GitHub PR** (server-side merge), so the PR-triggered CI merge-check (ariadne #52) runs on every publish. A **local `git merge` + `git push origin main`** is the *acknowledged escape path* (fast, ungated, used knowingly) — not the default. (Updated 2026-05-31: earlier this said local-merge was acceptable as the default; the PR-only CI decision in you-decide #4 / ariadne #52 overrides that — a local merge is invisible to PR CI, so the standard publish must pass through a PR.)

Branch stays a **transient lane** (deleted post-merge), not a durable tag; durable "published state" markers, if ever wanted, are real `git tag`s.

## Plan

Grounding (from the 2026-05-31 investigation): the in-place primitive already exists
(`branchcreate.go:171 createInPlaceBranch()` — `git checkout -b`, carries the tree
forward); `pr.go` and `claim.go` work for in-place already; `sdlc merge` merges
**server-side** (`merge.go:211 ghClient.PRMerge`), so the PR/CI gate already applies. The
gap is the *local post-merge cleanup* (worktree-shaped today) + the `change-code` default
+ retiring direct-on-main + docs.

**M1 — core logic (sdlc Go) + tests:** ✅ 2026-05-31
- [x] `changecode.go` `resolveBranchingStrategy()` — empty `--worktree` → **in-place** (silent, with an info line). `--worktree=yes` worktree; added `--worktree=ask` to reach the tty prompt / `ASK_BRANCHING_STRATEGY` sentinel (so #39's contract + tested prompt funcs stay reachable, not dead). Verified: empty→mode=no (exit 0, no sentinel); `=yes`→mode=yes; `=ask` piped→sentinel+exit 2.
- [x] `merge.go` — topology split. New `isInPlaceCheckout(gitDir)` (linked worktree iff git-dir under `.git/worktrees/`). In-place: after `PRMerge`, `git switch main` → `pull` → archive in-dir → `git branch -D`. Worktree path unchanged. `findMainWorktree` is now worktree-only. No-PR + in-place aborts cleanly (run `sdlc pr`).
- [x] Tests — `TestIsInPlaceCheckout` (table-driven) added; full suite + vet green.

**M2 — keep `sdlc push` (operator decision):** ✅ confirmed, no change
- [x] `sdlc push` / `make push` stay as-is; `make worktree` correctly keeps `--worktree=yes`. Nothing else hardcodes the default away from in-place.

**M3 — docs:** ✅ 2026-05-31
- [x] `AGENTS.md` §0 branching-decision line — in-place default, worktree opt-in, push still available.
- [x] helptext `change-code.md` + `merge.md` — rewritten for in-place default + dual-topology merge.
- [x] `atlas/workflow/sdlc-binary.md` (verb table) + `issue-lifecycle.md` (flow + branching-decision section).
- [x] `construct/base.manifest` — no make-target shape change, so untouched (correct).

**Verify:**
- [x] `go test ./cmd/sdlc/...` + `go vet` green; `bin/sdlc` rebuilt.
- [x] Worktree path still works via `--worktree=yes` (verified mode=yes).
- [x] **End-to-end live dogfood:** satisfied in production — every branch-based close since (incl. `#54` logged explicitly as a "#51 dogfood", and `#75`/`#69` this cycle) ran `change-code` (in-place) → `pr` → `merge` with the new `switch main`/`pull`/`branch -D` cleanup against real git+gh. The `#69` merge output shows it verbatim ("In-place branch … will merge, then switch this checkout back to main / Switching to main… / Pulling… / Deleting merged branch"). The plumbing is proven, not just the unit-tested decision.
- Downstream inheritance (a derivative's next `make refresh`) trails — a per-derivative propagation event, not an ariadne deliverable; it happens at refresh time and doesn't block closing #51 in ariadne. (Demoted from a checkbox to a note.)

**Revised estimate:** ~6–8h across M1–M3 (was 3h; the investigation showed it's mostly merge.go cleanup-path + docs, low risk since in-place creation is proven).

## Notes / cross-refs

- Builds on **#39** (`sdlc start` split into `claim` + `change-code`; worktree decision deferred) and **#40** (workflow management → sdlc). Related: **#36** (sdlc-push-followups).
- Downstream **you-decide#4** (pre-push/PR publish gate): orthogonal — that's the optional *enforcement* layer; this ticket is the *workflow shape*. The publish gate composes with whatever merge path this lands on.
- Primitive already present: `sdlc change-code --worktree=no` (in-place branch, carries working tree forward) — this ticket promotes it to default and completes its merge-back.

## Log

### 2026-06-03 — retroactive close
- 2026-06-03: closed — in-place branch is the default workflow (changecode.go resolveBranchingStrategy + merge.go topology split + TestIsInPlaceCheckout); proven in production — #54/#75/#69 all ran change-code(in-place)→pr→merge with switch-main/pull/branch-D cleanup against real git+gh. Retroactive close: code shipped 2026-05-31, actual=judgment ~4h (v3 0.56h rejected — stale multi-issue window, #68 limit); --no-judge (months-wide window meaningless); --no-atlas (atlas shipped with the May work).; review verdict: not-run
- Code (M1–M3) shipped 2026-05-31 (`ed4b550 #51 M1-M3: in-place branch becomes
  the default workflow`) and has been the live default ever since; the issue was
  just never formally closed — exactly the close-off hygiene drift `sdlc` exists
  to catch. The remaining "live dogfood" checkbox was satisfied in production many
  times over (see Plan); downstream-refresh demoted to a note (derivative-side).
- **Actual = judgment (~4h).** `sdlc actual --issue 51` measured 0.56h, but over a
  months-spanning window attributed across #49/#51/#52/#53 — the stale, multi-issue
  window v3 can't resolve (the #68 limitation). Rejected the measurement; judgment
  from "M1–M3 done in one 2026-05-31 session, mostly merge.go cleanup + docs."
- **`--no-judge`:** the #69 boundary review would run over `firstSHA(#51)^..HEAD`,
  a months-wide diff of unrelated work — meaningless. The work was reviewed by
  history (shipped + dogfooded a week). Closed direct-on-main (doc-only state flip).
