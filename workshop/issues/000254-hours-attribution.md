---
id: 000254
status: codecomplete
deps: []
github_issue:
created: 2026-09-27
updated: 2026-10-09
estimate_hours:
card_mirror: '76d140068196f012fcb58bda4cceccbef38f9902' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-09T21:04:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "1eb9b6cf", done: "88861740"}
actual_hours: 0.13
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

- [x] `issueref.Find` masks GitHub's `Merge pull request #N` lead, so every
      consumer drops it: commit boundaries and peers (`LocalNums`), foreign-ref
      warnings, and transcript mentions (`CountLocal`). Table test covers
      merge, a plain `#N`, a trailing ` (#N)` issue ref (kept, see Revisions)
      and a merge subject that also names an issue.
- [x] Claim-anchored window start — landed with #270 (folded there).
- [x] Re-measure parley.nvim#290 before/after; note the delta in the Log.

## Log

### 2026-09-27

- Filed from #252's A4 canary (parley.nvim#290's close). Pre-existing; not a
  #252 regression.

### 2026-10-09
- 2026-10-09: closed — make test green except processgroup TestCancellationKillsDescendants (sandbox blocks /bin/ps; passes unsandboxed). TestFindMasksMergePullRequestNumbers: Merge pull request #N never a peer/boundary/mention, plain #N and trailing (#N) kept; mask removal fails 3 cases. Claim-anchored window landed with #270 (TestWindowStart). parley.nvim#290 rerun: window 2026-09-27T15:41 -> HEAD, peers #290 only (pre-#270: #203 #204 #205 #264 #281 #289 #290 #291); ariadne#287 drops PR #166. Squash-suffix rule dropped per Revisions.; review verdict: SHIP

- Claimed in ariadne:3 after #270 landed (TL dispatch, ariadne-robustness-1).
- Spec 2 (claim-anchored window) shipped in #270: scoped boundaries made it
  necessary there (a 23.26 h first measure). #113's intent holds, since design
  after the claim is in-window.
- `issueref.Find` masks `Merge pull request #N`. `TestFindMasksMergePullRequestNumbers`
  covers boundaries/peers (`LocalNums`) and mentions (`CountLocal`); removing
  the mask fails 3 cases.
- The `actual` window label now names the resolved start (`<ISO> → HEAD`),
  not the first `#N` commit, which since the claim anchor is often the
  filing and lies outside the window.
- Re-measure. The baseline binaries were built from pre-#270 main (77e55b81^1)
  and from #270 (dfee8323), because the `sdlc` on PATH rebuilds from this
  slot's working tree. #270's Log calls its pair#247 baseline the "installed
  binary", but it was the unscoped author-date engine; its #341 conclusion
  holds.

  | Issue | pre-#270 | #270 | #254 |
  |---|---|---|---|
  | parley.nvim#290 hours | 0.60 h | 0.60 h | 0.60 h |
  | parley.nvim#290 peers | #203 #204 #205 #264 #281 #289 #290 #291 | #205 #290 | #290 only |
  | ariadne#287 hours | 0.26 h | 0.26 h | 0.26 h |
  | ariadne#287 peers | 37 issues | #166 #287 #300 #304 #306 | #166 (PR merge) dropped |

  - parley.nvim#290's window now opens at 15:41, its claim.
  - Hours are unchanged because these runs read the :0 transcript dirs, and
    the work happened in slots. The peer sets are the measurable delta here.
    pair#247 (in #270) carries the hours delta.

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
