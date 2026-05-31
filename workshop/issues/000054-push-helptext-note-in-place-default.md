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

- [ ] Add a short note near the top of `cmd/sdlc/helptext/push.md` (and/or its
  RELATED section) that in-place branch via `sdlc change-code` is the default
  since #51; `sdlc push` is the direct-on-main shortcut.

## Done when

- [ ] `sdlc push --help` (push.md) states that `sdlc change-code` (in-place
  branch) → `sdlc pr` → `sdlc merge` is the default flow since #51, and frames
  `sdlc push` as the direct-on-main shortcut.
- [ ] `go test ./cmd/sdlc/...` stays green, including a new `TestPushEmbedded`
  that pins push.md to reference the default `change-code` flow (the prior
  embed suite only proved push.md *embeds*, not that the note is present).

## Log
