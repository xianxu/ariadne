---
id: 000254
status: open
deps: []
github_issue:
created: 2026-09-27
updated: 2026-09-27
estimate_hours:
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

- A merge or squash subject's PR number never becomes a peer issue (tests
  cover merge, squash, and a plain `#N` reference).
- A claimed issue's hours window starts at its claim; a test pins it with an
  issue filed hours before its claim, with other issues' work in between.
- `sdlc actual --issue 290` in parley.nvim, run again, shows no #203/#204
  peers and a window starting at 15:41.

## Plan

- [ ] PR-aware extractor in `issueref`, used by `DiscoverWindowIssues` and the activetime mention attribution
- [ ] Claim-anchored window start in `windowStart`/`resolveWindowStart`
- [ ] Before/after measurement noted in the Log

## Log

### 2026-09-27

- Filed from #252's A4 canary (parley.nvim#290's close). Pre-existing; not a
  #252 regression.
