---
id: 000489
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'b2a9f5bc499538435f1895b9fa01097245a5d587' # card fields mirrored from issue-cards; edit via sdlc
---

# jsonstrict: route the three older duplicate-JSON-key scanners through internal/jsonstrict (ARCH-DRY, from #473 M1 review)

## Problem

## Spec

Four copies of a "reject duplicate JSON object keys" scanner exist:
- `cmd/sdlc/internal/observe/json.go` `rejectDuplicateKeys`
- `cmd/sdlc/internal/fleet/types.go` `rejectDuplicateJSONKeys`/`scanJSONNoDuplicateKeys`
- `cmd/sdlc/teststats_store.go` `validateJSONNoDuplicates`
- the canonical `cmd/sdlc/internal/jsonstrict.RejectDuplicateKeys` (from #473 M1)

Move the three older sites onto `jsonstrict`, keeping each caller's error wrapping (ARCH-DRY).

## Done when

- `grep` finds one duplicate-key scanner in `cmd/sdlc`.
- Each former caller's existing duplicate-key test still passes.

## Plan

- [ ]

## Log

### 2026-10-20
