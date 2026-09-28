---
id: '000061'
status: done
created: 2026-06-01
updated: 2026-06-01
estimate_hours: 1
actual_hours: 0.5
---

# ariadne bootstrap must provision its own go toolchain (sdlc build dependency)

## Problem

`make bootstrap` in a fresh tart VM (no Go) hard-fails: `bootstrap-peers`
recurses `make bootstrap` into the ariadne peer → `tools` → `sdlc-build` →
`/bin/sh: go: command not found` → `Error 127` → the whole cascade aborts. The
only go-check today lives buried in `scripts/sdlc-install.sh` (step 4 of 5) and
just `die`s — too late, after the peer-clone cascade, and it never fixes it.

Root cause: ariadne ships `cmd/sdlc` and compiles it in `tools`, so **go is a
hard build-dependency of the base layer itself** — but bootstrap never
provisions it. Pre-sdlc, ariadne needed only shell + python (effectively always
present), so dependency-provisioning was never wired in; when sdlc (Go) landed
we added the build but not the dep. nous owns its richer toolchain (Homebrew,
GPG, gh, …) separately; ariadne just needs to guarantee go.
