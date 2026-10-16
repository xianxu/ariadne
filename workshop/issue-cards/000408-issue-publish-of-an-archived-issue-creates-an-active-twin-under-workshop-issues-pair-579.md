---
id: 000408
status: open
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
github_issue:
---

# issue publish of an archived issue creates an active twin under workshop/issues (pair#579)

## Problem

2026-10-16: pair:3 ran `sdlc issue publish --issue 579` to add a Log entry to pair#579, which is done and archived
(`workshop/history/issues/000579-*.md`). With the default issues dir, publish created a NEW active twin
`workshop/issues/000579-*.md` (47 lines) on pair main (commit 7af5c7b1). Re-running with
`--issues-dir workshop/history/issues` published correctly to the archive. No verb removes the stray active copy.
