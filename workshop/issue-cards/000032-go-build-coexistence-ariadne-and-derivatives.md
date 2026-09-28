---
id: '000032'
status: done
created: 2026-05-25
updated: 2026-05-26
actual_hours: 6.9
---

# Go build coexistence: ariadne and derivative repos

## Problem

Issue #31 introduces `cmd/sdlc/` and makes ariadne itself a Go module (it will need its own `go.mod`). Derivative repos (../nous and future ariadne-styled repos) are *also* Go modules with their own `cmd/*/main.go` trees. The shared `Makefile.workflow` is symlinked from ariadne into every derivative (`symlink Makefile.workflow` in `construct/base.manifest`), so both repos run the same `build:` target. That target currently lives at `Makefile.workflow` and does:

```make
build:
    @if [ -f go.mod ]; then \
        found=0; \
        for d in cmd/*/; do \
            name=$$(basename "$$d"); \
            case "$$name" in nous) continue ;; esac; \
            ...
```

Two problems with this shape once ariadne also has `cmd/`:

1. **Downstream-name leak.** The hardcoded `case "$$name" in nous) continue ;;` is a derivative-specific exception sitting in the *base layer*. As more derivatives appear (e.g., charon, gmail), the list either grows here or each repo grows a different workaround. The base layer should not know any derivative's binary names.
2. **Ambiguous ownership of `cmd/`.** Ariadne will own `cmd/sdlc/`. Nous owns `cmd/nous/`, `cmd/charon/`, `cmd/gmail/`, etc. The shared scanner runs in both repos and picks up whatever is in the local `cmd/`. That's fine as long as the *generic scan* is what each repo wants — but the current nous skip says it isn't generic; nous deliberately runs `make nous-build` (its own target in `Makefile.nous`) for its primary binary and uses `make build` only for *other* utilities.

This leaves us with several entangled questions:

- Should ariadne's `cmd/sdlc/` be picked up by every derivative's `make build`? Cross-vendoring binaries from the base layer was never the design intent — `setup.sh` only vendors text-shaped artifacts (skills, datatypes, scripts, configs). Binaries shouldn't be vendored by the same path.
- If ariadne adds `go.mod` at its root, does that conflict with a derivative's `go.mod` when the derivative symlinks ariadne paths into its tree? (Today: nothing in `construct/base.manifest` symlinks `go.mod` or `cmd/`, so this is mostly fine — but the `Makefile.workflow` scanner walks the *target's* `cmd/`, not ariadne's, so the generic loop will work for any repo that drops `cmd/X/main.go` in.)
- Where does `make sdlc-build` live? In ariadne's own `Makefile.local` (so it never leaks downstream)? Or in `Makefile.workflow` (so the base layer ships the convention and downstreams just don't have `cmd/sdlc/`)? Issue #31's spec puts it inline alongside `nous-build` — but that conflates "binaries owned by the base layer" with "binaries owned by the repo."
