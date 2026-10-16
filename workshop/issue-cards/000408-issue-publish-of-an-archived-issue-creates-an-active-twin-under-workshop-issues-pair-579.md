---
id: 000408
status: working
created: 2026-10-16
updated: 2026-10-16
estimate_hours: 0.94
github_issue:
started: 2026-10-16T11:18:56-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# issue publish of an archived issue creates an active twin under workshop/issues (pair#579)

## Problem

2026-10-16: pair:3 ran `sdlc issue publish --issue 579` to add a Log entry to pair#579, which is done and archived
(`workshop/history/issues/000579-*.md`). With the default issues dir, publish created a NEW active twin
`workshop/issues/000579-*.md` (47 lines) on pair main (commit 7af5c7b1). Re-running with
`--issues-dir workshop/history/issues` published correctly to the archive. No verb removes the stray active copy.
