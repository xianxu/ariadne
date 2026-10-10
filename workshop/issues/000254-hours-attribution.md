---
id: 000254
status: working
deps: []
github_issue:
created: 2026-09-27
updated: 2026-10-09
estimate_hours:
card_mirror: 'fdec70ad27d57893c50dee12bc9a3daf3e831e0a' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-09T21:04:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# Hours attribution: ignore GitHub PR numbers; window starts at the claim

## Problem

`sdlc actual` (which `close` adopts as actual hours) splits active time across
the issues referenced by commits in the issue's window. Found closing
parley.nvim#290 in the #252 A4 canary; both parts pre-date #252.

1. **GitHub PR numbers are counted as issues.** `gitx.DiscoverWindowIssues`
   feeds every commit subject in the window through `issueref.LocalNums`.
   GitHub's merge-commit subject `Merge pull request #204 from xianxu/<branch>`
   names a *PR* number, yet it becomes a peer "#204". In the #290 run it took
   2.7 minutes, and "#203" appeared too. GitHub's squash merges add the same
   kind of suffix (`Title (#204)`). Where PR and issue numbers overlap, real
   hours go to the wrong issue.
2. **The window starts when the issue was filed.** `windowStart` takes the
   earlier of the parent of the first `#N` commit and the claim (`started`).
   #290 was filed at 11:46 (`#290: issue-sync: new issue`) but claimed at
   15:41. The window covered the whole afternoon, including #264's work, and
   the time was split across #203, #204, #264, #281, #289, #290 and #291
   (warnings: "mention fallback without issue commit boundary"). #113 widened
   the start on purpose, to count design attention before the claim, but
   filing is not design. An issue filed days before it is worked absorbs, and
   dilutes, everything in between.

## Spec

1. **PR-aware reference extraction.** One extractor, used by every
   attribution consumer (window peers and the per-interval mention
   attribution), ignores GitHub PR references:
   - `Merge pull request #N from …` (a merge commit's PR number);
   - a trailing ` (#N)` on a squash-merge subject.
   The branch name inside a merge subject (`from xianxu/000264-…`) may still
   attribute to its issue (#264) — that is the landed work. Table-driven
   tests over real subject shapes.
2. **Window start = the claim, when there is one.** For an issue with a
   claim (card `started`, or the legacy working-transition commit), the
   window starts at the claim. Commits made before it that name the issue
   (filing, early design syncs) still count as that issue's activity, but do
   not widen the window over unrelated work. Without a claim, keep today's
   anchor (parent of the first `#N` commit). Record the change against #113's
   intent in the Log (design attention after the claim is still captured).
3. Re-measure one known issue before and after, and note the delta, so
   calibration consumers know.

## Done when

- A `Merge pull request #N` subject's PR number never becomes a peer issue,
  boundary or mention (tests cover merge, a plain `#N`, and a trailing
  ` (#N)` issue ref that stays a ref).
- A claimed issue's hours window starts at its claim; a test pins it with an
  issue filed hours before its claim, with other issues' work in between. (Landed with #270: `TestWindowStart`.)
- `sdlc actual --issue 290` in parley.nvim, run again, shows no #203/#204
  peers and a window starting at 15:41.

## Plan

- [ ] `issueref.Find` masks GitHub's `Merge pull request #N` lead, so every
      consumer drops it: commit boundaries and peers (`LocalNums`), foreign-ref
      warnings, and transcript mentions (`CountLocal`). Table test covers
      merge, a plain `#N`, a trailing ` (#N)` issue ref (kept, see Revisions)
      and a merge subject that also names an issue.
- [x] Claim-anchored window start — landed with #270 (folded there).
- [ ] Re-measure parley.nvim#290 before/after; note the delta in the Log.

## Log

### 2026-09-27

- Filed from #252's A4 canary (parley.nvim#290's close). Pre-existing; not a
  #252 regression.

### 2026-10-09

- Claimed in ariadne:3 after #270 landed (TL dispatch, ariadne-robustness-1).
- Spec 2 (claim-anchored window) shipped in #270: scoped boundaries made it
  necessary there (a 23.26 h first measure). #113's intent holds, since design
  after the claim is in-window.

## Revisions

- 2026-10-09 — Dropped the squash-suffix rule from Spec 1. A trailing
  ` (#N)` is the fleet's own convention for real issue refs (e.g.
  `lessons: … (#179)` in ariadne, `… (#264)` in parley.nvim), and none of
  these repos squash-merge: all 170 GitHub merges in ariadne are merge
  commits. Stripping the suffix would drop real attribution to remove a PR
  number that never occurs. The `Merge pull request #N` lead is unambiguous
  and is still masked. The branch name inside it is not parsed into an issue:
  under #270's scoping a merge commit is never one of the branch's own
  commits, so it would add surface for no measurement. Done-when 1 now pins
  the trailing ` (#N)` as a kept ref instead of a squash case.
