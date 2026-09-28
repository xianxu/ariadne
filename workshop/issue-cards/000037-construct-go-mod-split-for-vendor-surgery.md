---
id: '000037'
status: done
created: 2026-05-26
updated: 2026-05-26
estimate_hours: 5
actual_hours: 2.5
---

# construct/go.mod split — surgical vendoring for substrate-only Go tool deps

## Problem

The current substrate Go-vendoring (ariadne #31 follow-on work) declares ariadne's `cmd/sdlc` as a tool dependency in each derivative's **root** `go.mod`. `go mod vendor` then populates the derivative's **root** `vendor/` with the *closure of both* (a) the derivative's app deps and (b) the sdlc tool's deps.

For derivatives that already have Go app code (like `pair` with charmbracelet + creack/pty + golang.org/x/* dependencies), this means substrate refresh produces a ~15MB / 800+ file `vendor/` directory. Only ~600KB of that is sdlc-related; the bulk is the derivative's own app closure that the substrate didn't actually need to vendor.

Symptoms:
- `vendor/` git diffs balloon on every substrate refresh
- Operators are confused about why substrate refresh added `vendor/github.com/charmbracelet/*` (entries unrelated to sdlc)
- "vendor mode" feels indistinguishable from "always vendor everything"

The root cause is structural: `go mod vendor` operates at module level, vendoring the entire closure of whatever's in `go.mod`. There's no partial-vendor flag. The granularity of control is *what's declared in the module*.
