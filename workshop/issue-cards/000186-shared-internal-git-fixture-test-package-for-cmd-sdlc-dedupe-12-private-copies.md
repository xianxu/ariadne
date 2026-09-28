---
id: '000186'
status: done
started: 2026-07-19T11:17:40-07:00
created: 2026-07-17
updated: 2026-07-19
estimate_hours: 2.07
actual_hours: 3.67
---

# shared internal git-fixture test package for cmd/sdlc (dedupe ~12 private copies)

## Problem

cmd/sdlc tests carry ~12 private copies of the same git-fixture idiom
(init temp repo on main, config user, initial commit) — closeRepo,
hermeticRepo, initFleetRepo/gitIn (peerwrite_apply_test.go), migrate/resolve
fixtures, plus near-identical writeProject helpers in discover_test.go vs
projectfind_test.go. Flagged by the #171 M3 boundary review and again at the
#171 issue-close review.
