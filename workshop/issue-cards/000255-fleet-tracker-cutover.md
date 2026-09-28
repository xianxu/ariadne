---
id: 000255
status: working
created: 2026-09-27
updated: 2026-09-27
estimate_hours:
github_issue:
started: 2026-09-27T23:21:50-07:00
---

# Fleet cutover to the issue tracker; delete the legacy writers

## Problem

#252 built the issue tracker (cards on `issue-tracker`, details with
`card_mirror`), the one-time migration (`sdlc issue migrate`), the cutover
guard, and a legacy mode that runs the pre-#252 workflow wherever a repository
has not cut over. It proved them through smoke tests (A1, A2), a legacy-mode
soak and a canary cutover of parley.nvim (A3, A4).

What #252 does not do is move the rest of the fleet. Its "existing issue files
are migrated" criterion moves here. The legacy writers, which the binary keeps
only for repositories not yet cut over, are deleted once every repository has
moved.
