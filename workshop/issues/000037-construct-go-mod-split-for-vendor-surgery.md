---
id: 000037
status: working
deps: [000031]
created: 2026-05-26
updated: 2026-05-26
estimate_hours: 5
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

## Spec

### Design — two `go.mod` files per derivative

Each derivative with substrate-Go-dependencies gets a second `go.mod` scoped to *substrate concerns only*. Layout:

```
<derivative>/
  go.mod                    # app deps (charmbracelet, creack/pty, etc.) —
                            # operator-managed, may or may not vendor
  construct/
    go.mod                  # substrate tool deps (require + replace + tool
                            # for each ancestor that ships Go tools)
    vendor/                 # vendored substrate-tool sources only (small)
  bin/sdlc                  # built from construct/vendor/
  bin/<other-tools>         # other ancestor-shipped tools build here too
```

The two `go.mod` files are independent Go modules — Go natively supports this (one module per directory containing a `go.mod` file). The construct/ submodule lives entirely in the substrate-managed space; the operator never edits it manually.

### Why this layout wins

- **Surgical vendoring.** `construct/vendor/` contains only the ancestor-tool closure (~600KB for sdlc-only). The derivative's root `vendor/` is unchanged by substrate refresh.
- **Single substrate path.** Multiple ancestors that ship Go tools (e.g., baby-brain consuming both ariadne's sdlc + nous's cmd/nous) all converge in one `construct/go.mod` — go.mod natively supports multiple require + tool entries.
- **Clear ownership boundary.** Substrate-managed = `construct/go.mod` (operator hands-off). App-managed = root `go.mod` (operator owns).
- **No custom walkers.** Standard `go build`, `go mod tidy`, `go mod vendor` against `construct/go.mod` does everything. No bespoke merge logic across multiple substrate go.mod files.

### Multi-layer composition

For derivatives consuming N upstreams that each ship Go tools, the substrate walks ancestor manifests in topological order. Each `tool <path>` action in an ancestor's `base.manifest` adds the corresponding require + replace + tool to the derivative's `construct/go.mod`. After all walks complete, `go mod vendor` in `construct/` produces one tree containing all ancestor closures.

Example — baby brain consuming both ariadne (sdlc) and nous (cmd/nous):

```
<baby-brain>/construct/go.mod:
  module github.com/<owner>/<baby-brain>-construct
  require github.com/xianxu/ariadne v0.0.0-...
  require github.com/xianxu/nous    v0.0.0-...
  replace github.com/xianxu/ariadne => ../../ariadne
  replace github.com/xianxu/nous    => ../../nous
  tool    github.com/xianxu/ariadne/cmd/sdlc
  tool    github.com/xianxu/nous/cmd/nous
```

### Module path for construct/go.mod

Convention: `github.com/<owner>/<derivative>-construct` (real-shaped path, but local-only — never published since replace directives always point at sibling paths). The `-construct` suffix makes the role explicit.

For ariadne itself: ariadne's root go.mod IS the substrate source; no separate construct/go.mod needed for ariadne.

### make bootstrap

New target in `Makefile.workflow`:

```make
bootstrap:
    @# Post-clone setup. Run once after `git clone <derivative>`.
    @# Does NOT require ../<upstream> sibling — vendored content is
    @# already in the repo. Builds tools from local vendor/.
    @if [ -d construct ] && [ -f construct/go.mod ]; then \
        $(MAKE) sdlc-build; \
    fi
    @# Other ancestor-shipped tools — built same way; targets added
    @# by ancestor base.manifest entries (out of scope for now).
```

Bootstrap is the *no-sibling-needed* verb. `make refresh` remains the *sibling-required* verb (it updates the vendored content from upstream).

### sdlc-build target update

```make
sdlc-build:
    @mkdir -p bin
    @echo "==> building bin/sdlc"
    @cd construct && go build -o ../bin/sdlc github.com/xianxu/ariadne/cmd/sdlc
```

Builds from construct/'s module context — picks up construct/vendor/ for the tool source.

### setup.sh changes

1. **`ensure_go_tool_dependency`** writes to `$TARGET_DIR/construct/go.mod` instead of `$TARGET_DIR/go.mod`. Creates `construct/go.mod` if missing (stub: `module github.com/<derived-from-root-module>-construct`, `go 1.24`). Self-walk case (ARIADNE_DIR == TARGET_DIR) keeps writing to root go.mod since ariadne has no construct/go.mod.

2. **`go mod vendor` post-processing step** runs in `$TARGET_DIR/construct/`, not at root. Same `MODE == vendor && ARIADNE_DIR != TARGET_DIR` gate.

3. **Refresh-target sibling-missing error** unchanged — still errors with hint when `make refresh` runs without ariadne available. (`make bootstrap` is the sibling-not-needed alternative.)

### Migration plan per derivative (nous, pair, parley.nvim)

1. Create `<derivative>/construct/go.mod` with module path `github.com/<owner>/<derivative>-construct`, `go 1.24`.
2. Move the ariadne require + replace + tool directives from root `go.mod` to `construct/go.mod`.
3. `go mod tidy` at root → strips ariadne entries.
4. `go mod tidy && go mod vendor` in `construct/` → populates `construct/vendor/` with just sdlc closure.
5. Delete root `vendor/` if it was only there because of substrate (judgment call per-derivative; pair likely keeps its own app vendor anyway).
6. `make sdlc-build` builds `bin/sdlc` from `construct/vendor/` via the new target.
7. Commit.

For brain (private brain, ariadne-styled derivative): same flow.

## Plan

Rough shape — to be detailed when work starts:

- [x] **M1 — Substrate changes (ariadne main).** Done in commit 239b9c2. setup.sh's `ensure_go_tool_dependency` targets `construct/go.mod` (auto-creates stub); `go mod vendor` runs in `construct/`; `sdlc-build` cd's into `construct/` when present; new `make bootstrap` target. Atlas entry updated with the "Go source vendoring — construct/go.mod split" section.
- [x] **M2 — Migrate nous.** Done in nous commit (post-M1 refresh + cleanup). `nous/construct/go.mod` auto-created with module `github.com/xianxu/nous-construct`. Root go.mod stripped of ariadne entries via `go mod edit -dropreplace -droprequire -droptool`. Root vendor/ rebuilt without ariadne (~24MB of nous's own app deps; previously mixed with sdlc closure).
- [x] **M3 — Migrate parley.nvim.** Done in parley.nvim commit. Root go.mod minimal (just `module` + `go` directive — no Go app code). Root vendor/ removed entirely (was substrate-only).
- [x] **M4 — Migrate pair.** Done in pair commit. Root vendor/ rebuilt to ~14MB (was ~15MB; substrate contribution ~944K moved to construct/vendor/). Substrate-only deps fully separated from app deps.
- [x] **M5 — Document operator workflow.** Atlas entry covers `make bootstrap` (no-sibling-needed) vs `make refresh` (sibling-required) distinction. Substrate-side docs sufficient; AGENTS.md doesn't need updating (existing prose stays accurate at its abstraction level).

## Out of scope (followups)

- **Auto-cloning ariadne sibling on `make refresh`.** Operator handles manually; `make refresh` errors with hint if sibling missing.
- **Nous as a re-export host for cmd/nous.** Nous's base.manifest could add `tool cmd/nous` in the future to advertise nous's binary to baby brains. Defer until baby-brain work makes this concrete.
- **Removing `--vendor` / `--symlink` mode dichotomy.** Currently both supported; could simplify to vendor-only since symlink mode is now mostly a development convenience. Defer until operational data shows symlink mode is unused.

## Log

### 2026-05-26 — issue created

Issue extracted from in-session discussion about vendoring scope. The operator noticed pair's substrate refresh inflating its vendor/ from minimal to 15MB+, because `go mod vendor` operates module-level and vendors the union of app + tool dep closures. We considered four alternatives:

1. **Make ariadne public** — clean structurally but conflicts with privacy preference for in-repo strategic-intent prose (workshop/, atlas/, docs/vision pensives).
2. **Split sdlc into its own public repo** — would be cleanest long-term but introduces double-repo structure for code that's intent-tied to ariadne's conventions.
3. **Don't auto-vendor; defer to operator** — solves immediate complaint but breaks fresh-clone-without-sibling experience.
4. **Two go.mod files per derivative** (this issue) — surgical, single substrate path, no public exposure.

The operator chose (4). Pair's vendor inflation was the load-bearing complaint; the rest of the substrate Go-vendoring infrastructure (tool action, ensure_go_tool_dependency, go mod vendor step) stays — just retargeted to construct/go.mod.

`make bootstrap` was raised as the answer to the fresh-clone DX question — confirmed not to need ariadne sibling (vendored content is self-contained); only `make refresh` does.

Multi-layer composition (baby-brain consuming both ariadne + nous) verified to work with single construct/go.mod — Go's native multiple-require/tool support handles the merge without custom walkers.

deps: [000031] — depends on the substrate Go-vendoring infrastructure landed in #31.
