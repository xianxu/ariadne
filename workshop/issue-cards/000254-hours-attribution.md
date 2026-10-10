---
id: 000254
status: codecomplete
created: 2026-09-27
updated: 2026-10-09
estimate_hours:
github_issue:
started: 2026-10-09T21:04:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.13
tracker:
    version: 1
    completion:
        token: close-b2fbed26645a
        repository: github.com/xianxu/ariadne
        reviewed_head: ed83e9f8f32f53c89a49e6793b73e1f1f48add0a
        evidence_commit: 1334cfb297c9247d54e281e8ada85c90aada0bb6
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
