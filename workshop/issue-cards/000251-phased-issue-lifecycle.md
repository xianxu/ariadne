---
id: 000251
status: wontfix
created: 2026-09-24
updated: 2026-10-09
estimate_hours:
github_issue:
started: 2026-10-09T18:40:42-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
tracker:
    version: 1
    abandoned: {}
---

# Phased issue lifecycle: spec, plan and code phases land on main

## Problem

Issue files keep getting stranded or diverging, because `sdlc` publishes
*copies* of commits and never brings the copy back into the branch the
original came from. Evidence:

- **Resting branch diverges** (#249): `issue sync` makes local-only commits on
  `main-slotN`; any peer publish then leaves the branch ahead and behind.
- **Published copies strand, and drop content** (#249, re-diagnosed): `change-code`
  publishes one selected commit's diff (`PublishCommit`: `merge-tree` +
  `commit-tree` onto `origin/main`) and leaves the local branch alone. In pair,
  `15f5220f` (1 line) published as `6bc2340e`. The earlier syncs `24199071`
  (+24) and `6aa7d819` (+41/−7) never reached origin, and none of them is an
  ancestor of the #311 close commit. #311's design text was lost from `main` and
  had to be re-added on the feature branch (`dc94d4be`). They weren't duplicates.
- **Wrong-branch commits** (#240, #220): issue updates made while on a feature
  branch are committed there. Seen again on 2026-09-24 in ariadne slot 1:
  `issue new` for #251 reserved on origin correctly, *and* committed a local copy
  on top of #250's closed feature branch. That moved HEAD past #250's reviewed
  head, which would have made #250's `sdlc merge` refuse. It was removed by hand.
  The same slot's `main-slot1` also shows the #249 pattern (`ff972bf`,
  `d2640b3`).

Underneath: the issue system is a design artifact with its own history (specs,
plans, cross-referencing docs), but the lifecycle only models the code half.
`open → working → codecomplete → done` puts everything from claim to close in
one `working` state. Design never lands on `main` in its own right. It reaches
`main` only as a side effect of `change-code`, one copied commit at a time.
