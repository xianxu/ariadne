---
id: '000057'
status: done
created: 2026-06-01
updated: 2026-06-01
estimate_hours: 4
actual_hours: 2
---

# dev-aliases.sh — expose Go binaries fresh; retire cmd source symlinks

## Problem

Two related pains, surfaced shipping #56:

1. **Dev staleness.** Developing a Go tool (sdlc) meant rebuilding the binary
   after every edit — and *two* binaries (`ariadne/bin/sdlc` + the on-PATH
   `pair/bin/sdlc`). `go run`/`go tool` don't fix it uniformly: in a *derivative*
   the `tool`+`replace` live in `construct/go.mod` (a separate module from the
   app `go.mod`, per #37/#38), so `go tool X` only resolves from `construct/`
   — whose cwd is wrong for the tool to operate on the repo. Only the *owner*
   repo (tool in root go.mod) can `go run`/`go tool` directly.
2. **Code distributed by source-symlink (deprecated).** `nous` ships its Go
   *source* to derivatives via **3** manifest directives — `symlink lib/gmail`,
   `symlink cmd/gmail`, `symlink cmd/oneshot` — producing **9 symlinks** across
   `brain`, `brain-family`, `brain-private` (3 repos × 3). That's code flowing
   through the *substrate* (file-symlink) channel instead of the module channel
   — the same pattern already retired for sdlc (ariadne→pair/brain use
   `tool`+`replace`, no source copy). (`lib/gmail` is the shared library the
   gmail cmd imports; the derivatives only carry it because they used to compile
   gmail *locally* from the `cmd/gmail` symlink — once gmail builds in its owner
   (nous) via the dev-alias, the derivatives need neither the cmd nor the lib
   source.) Note: the *many* other brain* symlinks (Makefiles, `scripts/`,
   `.claude/skills`, `atlas/workflow`, AGENTS.md, …) are doc/config/script
   substrate with no module mechanism — they stay; only the Go source leaves.
