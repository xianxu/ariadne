---
id: 000543
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '489cd14006bae11b2c12f55d8fd24fea858c2b78' # card fields mirrored from issue-cards; edit via sdlc
---

# new repositories start on the legacy issue path: initial setup should create the issue tracker (non-brain repos)

## Problem

## Spec

The operator asked (via ops-echo:0, 10-21) why echo, created 10-18, weeks after the issue tracker landed (#252),
started as a legacy repository. Confirmed: new-repository setup never creates the tracker. `construct/base.manifest`
scaffolds `workshop/issues` (line 182) but nothing writes `workshop/issue-tracker.json` or the `issue-tracker`
branch, so every new repo starts on the retiring legacy path and only reaches the tracker by a later
`sdlc issue migrate`. Inventory of the workspace on 10-21: ducks (09-27) and ops (10-09) were migrated by hand (ducks
the same day; ops on 10-20); echo (10-18) was migrated tonight after three legacy-path surprises (stray sync commits,
start-plan not cutting a branch); **ops-echo (10-19) and emma-private (10-05) are still legacy**. The brains
(brain, brain-family, brain-private) are legacy by design and must stay so.

Fix: initial setup of a new, non-brain repository creates the tracker (the empty-repo case of `sdlc issue migrate`,
or an equivalent `init` step), so new repos start on the tracker. Brain repos (`.brain/config.md`) and repos that
opt out are left alone. Existing legacy repos get a one-line hint from `sdlc issue new` and `start-plan` pointing at
`sdlc issue migrate`.

## Done when

- Bootstrapping a fresh non-brain repo yields `workshop/issue-tracker.json` and an `issue-tracker` branch on origin, so
  `sdlc issue new` writes no checkout and `start-plan` cuts the branch (test with a bare remote).
- Bootstrapping a brain repo creates no tracker (test).
- In a legacy repo, `issue new` and `start-plan` print the migrate hint (test).

## Plan

- [ ]

## Log

### 2026-10-21
