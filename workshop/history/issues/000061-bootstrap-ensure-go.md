---
id: 000061
status: done
deps: []
github_issue:
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

## Spec

A base-layer `ensure-go` target that guarantees the Go toolchain before anything
builds sdlc. Idempotent and non-intrusive:

- `command -v go` present → **no-op** (don't fight gvm/asdf/manual installs).
- absent + `brew` available → `brew install go`.
- absent + no brew → **fail fast** with `https://go.dev/dl/` guidance (the
  sdlc-install message, surfaced *before* the costly peer-clone cascade).

Wiring: `ensure-go` as the **first** prerequisite of `bootstrap`
(`bootstrap: ensure-go bootstrap-peers refresh tools sdlc-install data-deps`) so
go is provisioned before the cascade AND before the recursive ariadne bootstrap's
`tools`; plus a prereq on `sdlc-build` so direct `make sdlc-build`/`make tools`
self-provision too. Both idempotent (the second hit is a no-op).

Scope: macOS-centric (`brew`) — matches dev machines + tart VMs. Linux/openshell
hits the fail-fast-with-instructions path; an apt branch is deferred until needed.

## Done when

- `make ensure-go` is a no-op when go is present; auto-installs via brew when
  absent+brew; fails fast with guidance when absent+no-brew.
- `bootstrap` provisions go before `bootstrap-peers`; `sdlc-build` self-provisions.
- A Go-less tart VM `make bootstrap` no longer dies at sdlc-build (manually
  reproduced before/after, or reasoned + unit-tested the conditional).
- Atlas note (`sdlc-binary.md` build section) records the dependency.

## Plan

- [x] M1 — add `ensure-go` to Makefile.workflow; wire into `bootstrap` (first)
      + `sdlc-build` + `build`; unit-test the conditional (present/absent±brew);
      atlas note. Soften sdlc-install.sh's early go-die to defer to ensure-go.

## Log


- 2026-06-01: closed — ensure-go.test.sh 3/3 (real recipe, stubbed PATH: no-op when present, brew-install when absent+brew, fail-fast otherwise); `make ensure-go`/`build`/`sdlc-build` work; review SOUND, the -j build-gate race fixed. ACTUAL estimated (--force): v3 has 0 events in the #61 window (recent same-session work not yet in transcript)
### 2026-06-01

`ensure-go` added (Makefile.workflow): no-op when go present, `brew install go`
on macOS when absent, fail-fast with go.dev/dl otherwise. Wired as first
`bootstrap` prereq + prereq of BOTH go-build targets (`sdlc-build`, `build`) —
the latter (review finding) closes a `make -j` race where `build`'s go-build
could run before the toolchain install. sdlc-install.sh's early go-die softened.

ensure-go.test.sh runs the real recipe under stubbed PATH across all 3 branches
(3/3); full suite green; real `make ensure-go` no-ops, `make build`/`sdlc-build`
build. **Review SOUND** (1 Important — the -j `build` gate — fixed; rest
verified incl. nous derivative composition + sdlc-install softening safety).
