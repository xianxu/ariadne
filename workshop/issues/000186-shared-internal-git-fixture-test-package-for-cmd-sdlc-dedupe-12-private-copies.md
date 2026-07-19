---
id: 000186
status: working
deps: []
github_issue:
created: 2026-07-17
updated: 2026-07-19
estimate_hours:
started: 2026-07-19T11:17:40-07:00
---

# shared internal git-fixture test package for cmd/sdlc (dedupe ~12 private copies)

## Problem

cmd/sdlc tests carry ~12 private copies of the same git-fixture idiom
(init temp repo on main, config user, initial commit) — closeRepo,
hermeticRepo, initFleetRepo/gitIn (peerwrite_apply_test.go), migrate/resolve
fixtures, plus near-identical writeProject helpers in discover_test.go vs
projectfind_test.go. Flagged by the #171 M3 boundary review and again at the
#171 issue-close review.

## Spec

## Done when

- One internal test-fixture package (e.g. cmd/sdlc/internal/testfix) owns the
  git-repo/fleet fixture idiom; the private copies delegate to it.
- No behavior change; suites stay green.

## Plan

- [ ] inventory the private fixture copies (closeRepo, hermeticRepo, initFleetRepo/gitIn, migrate/resolve fixtures, writeProject duplicates)
- [ ] extract cmd/sdlc/internal/testfix with the shared git-repo/fleet idiom
- [ ] delegate the copies; suites stay green (behavior-identical)

## Log

### 2026-07-17
