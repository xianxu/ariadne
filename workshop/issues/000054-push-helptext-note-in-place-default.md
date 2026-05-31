---
id: 000054
status: working
deps: []
created: 2026-05-31
updated: 2026-05-31
estimate_hours: 0.25
---

# push.md helptext: note in-place branch is the default since #51

## Spec

#51 made the in-place branch the default `change-code` path and added the
dual-topology `sdlc merge`. The `change-code` and `merge` helptext were updated,
but `cmd/sdlc/helptext/push.md` was missed — it still presents direct-on-main
`sdlc push` without noting that it is no longer the default close path. A reader
landing on `sdlc push --help` should learn that the standard flow is now
`sdlc change-code` (in-place branch) → `sdlc pr` → `sdlc merge`, and that `push`
remains available as the direct-on-main shortcut for quick one-liners. This also
serves as the live end-to-end dogfood of the #51 in-place branch→pr→merge flow.

## Plan

- [x] Add a short note near the top of `cmd/sdlc/helptext/push.md` (and/or its
  RELATED section) that in-place branch via `sdlc change-code` is the default
  since #51; `sdlc push` is the direct-on-main shortcut.

## Done when

- [x] `sdlc push --help` (push.md) states that `sdlc change-code` (in-place
  branch) → `sdlc pr` → `sdlc merge` is the default flow since #51, and frames
  `sdlc push` as the direct-on-main shortcut.
- [x] `go test ./cmd/sdlc/...` stays green, including a new `TestPushEmbedded`
  that pins push.md to reference the default `change-code` flow (the prior
  embed suite only proved push.md *embeds*, not that the note is present).

## Log

- 2026-05-31 — Served as the live end-to-end dogfood of the #51 in-place
  branch flow (ariadne #53 Phase B). Ran for real: `claim` → `change-code
  --worktree=no` (in-place branch created, working tree carried forward;
  plan-quality judge passed INFO) → push.md edit + `TestPushEmbedded` guard
  → commit `c011d9d` → `pr` (PR #4) → `merge`.
- **Dogfood finding (the bug this was meant to catch):** the *deployed*
  `sdlc` binary on PATH (`you-decide/bin/sdlc`) was a month stale (May 28,
  pre-#51) and failed the in-place merge with `find main worktree: could
  not find a worktree on branch 'main'`. The current ariadne source/binary
  handles in-place correctly (switch-back-to-main topology). Root cause:
  downstream prebuilt binaries don't auto-rebuild on base-layer tool
  changes. Captured in `atlas/workflow/sdlc-binary.md` (downstream staleness
  gotcha); you-decide binary refreshed via `make sdlc-build`.
