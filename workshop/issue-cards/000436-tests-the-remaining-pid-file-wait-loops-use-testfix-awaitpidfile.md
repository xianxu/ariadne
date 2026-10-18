---
id: 000436
status: open
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
github_issue:
---

# tests: the remaining pid-file wait loops use testfix.AwaitPIDFile

## Problem

The #434 close review found hand-written pid-file wait loops outside cmd/sdlc that duplicate `testfix.AwaitPIDFile`: `internal/processgroup/process_unix_test.go` (2), `internal/gitx/lifetime_test.go` (2), `internal/gitx/rebasedpatch_context_test.go` (1), `internal/judge/cancellation_unix_test.go` (1). One helper keeps one readiness idiom (ARCH-DRY).
