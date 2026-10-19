---
id: 000472
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
---

# merge: issue-only main commits shouldn't cost a ready merge an integration attempt or a CI wait

## Problem

2026-10-19, pair#705 (pair:7): its integrated suite passed and CI passed, then two **issue-only** commits reached main (a
`sdlc issue new` publish and an `abandon` archive, touching only `workshop/`). Each forced `sdlc merge` to integrate main
again and wait for fresh CI, using up attempt 3 of 3. On a busy repo, issue filing alone can starve a ready merge, and the
hub had to ask every slot to stop publishing.
