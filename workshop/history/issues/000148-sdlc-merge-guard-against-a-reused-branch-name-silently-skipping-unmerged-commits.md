---
id: 000148
status: done
deps: []
github_issue:
created: 2026-06-30
updated: 2026-07-05
estimate_hours: 0.48
started: 2026-07-05T20:31:25-07:00
actual_hours: 0.51
---

# sdlc merge: guard against a reused branch name silently skipping unmerged commits

## Problem

When `sdlc merge` finds a **merged** (not open) PR for the head branch, it treats
the work as already shipped and "resumes post-merge cleanup" — switches to main,
pulls, archives, and **deletes the branch** — WITHOUT checking whether the branch
has commits *beyond* what that PR merged. So a reused branch name silently drops
new work: the commits stay on `origin` but never reach main, and the branch is
deleted out from under them, with a green exit code.

Concrete (parley.nvim#116): the M2/M3 work reused the branch name of M1, which had
shipped early via its own merged PR #95 (to unblock #128). `sdlc merge` found #95
(MERGED), resumed cleanup, and switched to main + deleted the branch **without
merging the 16 new M2/M3 commits**. `git rev-list --left-right --count
main...origin/<branch>` showed `0 16` — main never advanced — but nothing warned.
Recovery required re-pushing under a fresh name + `sdlc pr` + `sdlc merge`.

This is a "form gate defends against omission" gap: the merge silently did the
wrong thing instead of refusing with a next-action spec.

## Spec

In the `sdlc merge` post-merge-cleanup path (the branch where an existing PR is
**MERGED** rather than open):

- Before resuming cleanup, compute the unmerged-commit count:
  `git rev-list --count <base>..<head>` (base = the PR's base, e.g. `origin/main`;
  head = the branch tip, e.g. `origin/<branch>`).
- If the count is **0** → the branch is genuinely fully merged → proceed with
  cleanup as today (switch to main, pull, archive, delete branch).
- If the count is **> 0** → **abort** with an actionable error, e.g.:
  `branch '<b>' has N commit(s) not in main despite a merged PR (#<n>) — likely a
  reused branch name. Rename the branch (e.g. <issue>-<short-slug>) and run
  `sdlc pr`, then `sdlc merge`.` Do NOT switch branches, delete, or archive.

Keep this scoped to the merged-PR path; the open-PR and no-PR paths are unchanged.

## Done when

- `sdlc merge` refuses (with the actionable message, non-zero exit, no branch
  deletion) when a merged-PR head branch still has commits not in the base.
- A genuinely fully-merged branch (count 0) still cleans up exactly as before.
- Tests cover both: merged-PR-with-unmerged-commits → abort + tree untouched;
  merged-PR-fully-merged → cleanup proceeds. (The count computation is the pure
  seam to unit-test; the git IO is faked/injected.)
- `sdlc --help` (`cmd/sdlc/helptext/root.md`, PUBLISH) notes the "publish once at
  issue close, not per milestone; don't reuse a branch name with a merged PR"
  guidance (the doc half of this fix).

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module   design=0.15 impl=0.2
item: atlas-docs          design=0.05 impl=0.05
design-buffer: 0.15
total: 0.48
```

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*
Decomposition: (1) `merge.go` — new `actionResumeBlocked`, extend the pure
`decideMergeAction` with an `unmergedCount` param, fetch-base+count guard in the
merged-PR caller path, + tests; (2) `helptext/root.md` PUBLISH note. `impl=` at
v3.1's 40%; familiarity 1.0 (deeply warm — just shipped #145/#146 in this exact
close/merge neighborhood, ran `sdlc merge` several times).

## Plan

Single-pass (no `Mx` — one `sdlc close`). Scoped strictly to `merge.go`'s
**merged-PR** path (open-PR / no-PR paths untouched, per Spec). Design decisions:

- **Extend the pure decision seam (ARCH-PURE/DRY).** Add `actionResumeBlocked` to
  the `mergeAction` enum and give `decideMergeAction(openPR, mergedExists,
  unmergedCount int)` a third arg: `mergedExists && unmergedCount > 0 →
  actionResumeBlocked`; `mergedExists && count 0 → actionResume` (unchanged); open-PR
  / no-PR arms unchanged. The decision stays pure + unit-tested (extends the existing
  `TestDecideMergeAction`); the git IO stays in the caller.
- **Fetch-then-count, fail-safe (correctness subtlety).** At step 10, `origin/main`
  is stale — the merged PR advanced it but the flow doesn't `pull`/`fetch` until
  *after* deciding to merge. So the guard, only when `prNumber == "" && mergedExists`,
  must first `mergeRunner.Git("fetch", "origin", "main")`, then count
  `rev-list --count origin/main..origin/<branch>` (base = origin/main, head =
  `remoteRef`, already resolved at line ~285). Route the count through `mergeRunner`
  (injectable) so it's fakeable. If the fetch or count *errors* (can't verify) →
  **die** (fail-safe: never clean up an unverified branch), don't default to 0.
- **The refusal.** `actionResumeBlocked` → `die` with an actionable message
  (non-zero exit, BEFORE any switch/pull/archive/delete): the branch has N commits
  not in main despite a merged PR — likely a reused name; rename + `sdlc pr` +
  `sdlc merge`; note the commits are safe on `origin/<branch>`.

- [x] Add `actionResumeBlocked` + extend `decideMergeAction(openPR, mergedExists, unmergedCount int)`; update `TestDecideMergeAction` with the count cases (RED→GREEN).
- [x] Extract `countUnmerged(r gitRunner, base, head string) (int, error)` + fake-runner unit test (`"16\n"→16`, `"0"→0`, git-error → propagates, non-numeric → error).
- [x] Wire fetch-base + `countUnmerged(mergeRunner,"origin/main",remoteRef)` into the merged-PR path; fail-safe `die` on error; `actionResumeBlocked` → actionable `die` before any switch/pull/archive/delete.
- [x] `helptext/root.md` PUBLISH block note (renders in `sdlc --help`).
- [x] `go build/vet/test ./...` (25 pkgs green) + **e2e** (both resume branches, real git): cleanup-proceeds when fully merged + refuse-tree-untouched when reused; atlas updated; close.

## Log

### 2026-06-30

Filed from the parley.nvim#116 landing session — see parley `workshop/lessons.md`
(2026-06-30, lesson #4) for the incident. The doc half (the PUBLISH-block wording)
is drafted there too. Root-cause fix lives in `cmd/sdlc/merge.go`'s merged-PR
cleanup branch.

### 2026-07-05
- 2026-07-05: closed — Re-close after FIX-THEN-SHIP fixes. --no-atlas: this delta is only test-hardening (pin the guard fetch in _FinishesCleanup) + one helptext/merge.md REFUSES-IF line — NO new architectural surface; the atlas resume-path description was already updated in the feature commit (05a9e14) earlier in this issue. I1: _FinishesCleanup now pins the fetch (proven — neutering the fetch fails the test). I2: merge.md guard note. Prior review verdict was FIX-THEN-SHIP (recorded "unknown" due to a non-parseable prose verdict from the long-running reviewer). go build/vet/test ./... — 25 pkgs, 0 failures.; review verdict: FIX-THEN-SHIP

Claimed → start-plan → change-code. Both judges **INFO (pass)**. Folded plan-quality
finding #1: extract `countUnmerged` + fake-runner unit test so the count seam is
*automated* coverage (not manual), satisfying Done-when bullet 3. DRY note (finding
#2): `merge.go:291`'s ahead-check also runs `rev-list --count` but via `gitx.Capture`
(not fakeable); it stays as-is (pre-existing) — `countUnmerged` is the fakeable home
for the new guard's count, the divergence is intentional (the guard must be testable).
Estimate findings advisory only (design slightly generous) — no change.

**Implemented + verified.** `go build/vet/test ./...` — 25 pkgs, 0 failures.
Pure decision (`decideMergeAction` +count arg) + count seam (`countUnmerged`, fake
runner) unit-tested. **E2E (real git, `merge_e2e_test.go`):** existing
`…_ResumeMergedPR_FinishesCleanup` fixture was unrealistic (stubbed "merged" but
never merged feature→main); made it genuinely merge (count 0 → cleanup proceeds) —
this also proves finding #3 (the guard's `git fetch origin main` refreshes
`refs/remotes/origin/main`, else it'd count 1 and refuse). Added
`…_UnmergedCommits_Refuses`: feature 1 commit ahead + stubbed-merged → aborts with
the reused-name message, PRMerge not called, still on feature, branch not deleted,
issue not archived (tree untouched). So Done-when bullet 3 ("abort + tree
untouched" / "cleanup proceeds") is AUTOMATED e2e, not manual.

**Boundary review: FIX-THEN-SHIP** (verdict recorded "unknown" — the long-running
reviewer emitted the verdict in prose, not the parseable token; re-close retries).
Both non-blocking Important findings fixed: (I1) `_FinishesCleanup` didn't actually
PIN the guard's fetch — the test's `git push origin main` already refreshed the
local `origin/main` tracking ref, so the fetch was redundant there (my "proves
finding #3" claim was overstated). Fixed by forcing the tracking ref stale
(`update-ref refs/remotes/origin/main <seed>`) after the push — proven: neutering
the guard's fetch now FAILS the test, restoring passes. (I2) `helptext/merge.md`
REFUSES-IF block didn't mention the #148 guard — added it (section-granularity
sweep: I'd updated root.md's PUBLISH block but not the merge subcommand's own help).
