---
id: 000032
status: open
deps: [000031]
created: 2026-05-25
updated: 2026-05-26
estimate_hours:
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

## Spec

### Architectural decision

Adopt Go modules as the cross-layer dependency-management mechanism for all file replication — both text artifacts (skills, scripts, datatypes, configs) and Go code. Each layer is its own Go module. Transitive resolution + `replace` directives subsume today's bespoke per-layer `setup.sh` + vendored ancestor-manifest model.

Today's two scripts (`ariadne/construct/setup.sh` walking one manifest, `nous/nous/setup.sh` walking `ariadne-base.manifest` snapshot + `nous.manifest`) are structurally one algorithm at different depths. Unification: **one canonical `setup.sh` that walks N manifests dynamically**, where N is the dependency-chain depth resolved via `go list -m all`.

### Three operating modes

| Mode | go.mod state | Use case |
|---|---|---|
| **Sibling** | `replace github.com/xianxu/<upstream> => ../<upstream>` | Local cross-repo development; trunk-follow; no auth needed even for private upstream |
| **Vendored-clone** | `replace github.com/xianxu/<upstream> => ./.upstream/<upstream>` (clone in-tree at a tag) | Pinned version inside derivative's tree; one git clone at adopt time; works for private upstream |
| **Public-proxy** | `require github.com/xianxu/<upstream> <version>` (no replace) | Public upstreams only; Go proxy fetches + caches; pinned by tag in `go.sum` |

Sibling + vendored-clone are sufficient for ariadne's current private state. Public-proxy is purely additive — unlocked if ariadne goes public; not required.

### Canonical setup.sh

Source-of-truth at `ariadne/construct/setup.sh`. Vendored down to every layer via `base.manifest` so each consumer has the same script at `<repo>/construct/setup.sh`. Algorithm:

```
# Discover all upstreams in dependency order (ancestors first).
upstreams=$(go list -m -f '{{if not (eq .Path "<self>")}}{{.Dir}}{{end}}' all)
for upstream in $upstreams; do
  process_manifest "$upstream/construct/base.manifest"
done
# Then apply this layer's own additions.
process_manifest "$(dirname "$0")/base.manifest"
```

Depth scales naturally with the dep chain: ariadne walks just its own manifest; nous walks ariadne's then its own; baby brain walks ariadne's, nous's, then its own (if any).

### Path convention

Every layer puts substrate management at `<repo>/construct/` — ariadne, nous, baby brains, all the same. Resolves the `nous/nous/` quirk (the layer-name leaking into its own substrate path).

### Per-binary build opt-out: sentinel files

Replace the hardcoded `case "$$name" in nous) continue ;;` skip in `Makefile.workflow build:` with a sentinel-file convention. The base scanner skips `cmd/<name>/` iff `cmd/<name>/.skip-make-build` exists.

The opt-out is a **correctness** mechanism (not just a workflow nicety): nous distributes its binary signed + notarized. Overwriting `bin/nous` with a freshly-built local copy at the same path invalidates the operator's macOS keychain ACL grants for the signed identity and breaks notification capabilities. The sentinel protects against accidental overwrite from `make build`.

Each opted-out binary documents the rationale in its sentinel file contents (free-form prose). Future binaries with similar distribution semantics (`cmd/charon/`, `cmd/gmail/`, etc.) drop their own sentinel with their own rationale. Base layer no longer knows derivative-specific names.

### Generated artifacts

Artifacts that are *functions of code* — e.g., `construct/local/sdlc/SKILL.md` from `sdlc --index` — are regenerated at the consumer via the binary's install/refresh path, NOT shipped via `base.manifest`. The version of the binary determines the version of every derived artifact automatically.

This eliminates the text-vs-code lockstep concern: pinning ariadne to a specific version via go.mod pins both the binary and all its derived artifacts in one stroke.

### Privacy posture

Sibling + vendored-clone modes don't involve Go's public proxy, so they work fine for private ariadne. Public-proxy mode is the *convenience* mode for public consumers — strictly optional. Open-source derivatives that want to consume only ariadne's text artifacts can continue using `construct/setup.sh` directly with `--upstream ../ariadne` (sibling mode); they get the text fragments but skip the Go bits.

### Open questions (resolved)

- **`go.mod` at ariadne root or `cmd/sdlc/`-only?** Root. `cmd/sdlc/internal/{issue,gitx,judge,project}` rely on Go's internal-import rule (only the enclosing module sees `internal/`); sub-moduling would force the internal packages up to the same level as `main.go`, which is uglier. Ariadne's root `go.mod` doesn't conflict with derivatives because nothing in `base.manifest` symlinks it.
- **`sdlc-bootstrap` upgrade path?** Local-compile via the same `make refresh` flow. `go.mod`'s pinned version determines what gets built; no need for binary distribution channels.

## Plan

### Phase 1 — ariadne side (branch: `branch-32` in `~/workspace/ariadne`)

- [x] Rewrite `construct/setup.sh`: discover upstreams via `go list -m`, walk N transitive manifests in topological order, drop fixed-depth assumption. Internally idempotent.
- [x] Add `construct/setup.sh` itself to `base.manifest` so it can be vendored down through layers.
- [x] Rewrite `Makefile.workflow build:` — drop hardcoded `case "$$name" in nous)` skip; introduce `cmd/<name>/.skip-make-build` sentinel mechanism.
- [x] Atlas entry: `atlas/workflow/setup-and-replication.md` covering the unified model (3 modes, canonical setup.sh, sentinel convention, generated-artifact policy).
- [x] Self-test: re-run `make refresh` in ariadne against itself; verify depth-1 case still produces correct on-disk state. (Self-refresh exits via the depth-0 early-exit silently — correct.)
- [x] Verify backward-compat: fresh-derivative tempdir bootstrap produces correct symlink layout via the fallback-to-script-upstream path. Nous's current state (go.mod present but no `require ariadne` yet) also hits the fallback as expected — Go discovery returns no ancestors, single-ancestor mode kicks in. Phase 2 unblocks the multi-ancestor walk by adding `require ariadne` to nous's go.mod.

### Phase 2 — nous side (branch: `branch-32` in `~/workspace/nous`)

Commits land in nous; tracked here.

- [x] `nous/go.mod`: add `replace github.com/xianxu/ariadne => ../ariadne` for local dev. (The require line gets stripped by `go mod tidy` without a code import; canonical setup.sh parses replace directives directly as one of three discovery sources.)
- [x] Move nous's substrate from `nous/` to `construct/` — `nous/skills/` → `construct/skills/` via `git mv`; bespoke `nous/setup.sh` deleted; `nous/` directory removed entirely.
- [x] Delete `nous/setup.sh` — vendored canonical script from ariadne replaces it via `base.manifest`.
- [x] Delete `nous/ariadne-base.manifest` + the self-refresh loop — ariadne's location now comes from `go list -m` / replace directive parsing.
- [x] Rename `nous/nous.manifest` → `nous/construct/base.manifest` (with plugin manifests inlined; source paths updated for moved skills).
- [x] Add `cmd/nous/.skip-make-build` with rationale in contents (signed + notarized distribution; overwriting bin/nous invalidates keychain ACL grants). Other cmd/* binaries (brain-sync, charon, gmail, oneshot) didn't need the sentinel — they don't have signing/distribution constraints.
- [ ] Verify baby-brain spawn flow end-to-end: a fresh derivative running `../nous/construct/setup.sh` materializes ariadne's base layer + nous's additions correctly. **Deferred to phase 3 verification** — the setup.sh run is destructive enough (would switch nous's currently-vendored Makefile.workflow to a symlink) that doing it under operator supervision is safer than running it autonomously.

### Phase 3 — documentation + opt-in

- [ ] Update `docs/vision` pointers and atlas to reference the unified model.
- [ ] Future derivatives (parley.nvim, pair, baby brains) migrate at their own pace — phase 1 doesn't force phase 2's path on them.

### Verification milestones

- **After phase 1**: `~/workspace/ariadne` on `branch-32` still self-refreshes cleanly. Existing nous on its main branch can invoke `../ariadne/construct/setup.sh` and get the same on-disk result as before. Phase 1's setup.sh must be backward-compatible with depth-1 invocation.
- **After phase 2**: `~/workspace/nous` on `branch-32` consumes `../ariadne` (also on `branch-32`) via the `replace` directive. `make refresh` in nous walks both layers via `go list`. Test by spawning a throwaway baby brain and verifying the file layout matches today's flow.

### Branch + workflow notes

- Both repos use branch name `branch-32`.
- **No worktree**: peer-repo path assumptions (`../ariadne`, `../nous`) are pervasive enough that working in the canonical sibling locations is simpler than fighting it. Branches live in the main checkouts.
- No push during development; cross-repo verification happens through local sibling resolution.

## Log

### 2026-05-25 — issue created

Spun out from the #31 design conversation. Issue #31 introduces the first base-layer Go binary; we need to decide how that coexists with derivatives that are also Go modules. Currently `Makefile.workflow` has a hardcoded `case "$$name" in nous)` skip that already shows the layering pressure — we should resolve it before adding more derivatives.

### 2026-05-26 — spec refresh after design conversation

Worked through the architecture in conversation. Key decisions:

- **Go modules for cross-layer dep management** (text + code), unifying today's bespoke per-layer setup.sh + manifest-vendoring model.
- **Privacy concern resolved**: sibling + vendored-clone modes work fine for private ariadne (no public proxy involvement). Public-proxy mode is additive.
- **One canonical setup.sh** at `ariadne/construct/setup.sh`, vendored down. Walks N manifests dynamically via `go list -m all` topological resolution. Today's two scripts (ariadne + nous) collapse to one + per-layer base.manifest.
- **Path convention**: every layer puts substrate at `<repo>/construct/`. Resolves the `nous/nous/` namespace leak.
- **Generated artifacts** (SKILL.md from `sdlc --index`) regenerate at consumer via binary's install path — never shipped via base.manifest. Decouples text-vs-code lockstep.
- **Per-binary opt-out via sentinel files** (`cmd/<name>/.skip-make-build`), motivated by macOS keychain/notarization correctness (overwriting a signed binary's path invalidates ACL grants). Base layer no longer knows derivative-specific names.
- **Branch convention**: `branch-32` in both `~/workspace/ariadne` and `~/workspace/nous`. No worktree — peer-repo path assumptions are pervasive enough that direct sibling work is simpler than fighting it.
- **Two-phase plan**: ariadne first (canonical setup.sh + sentinel mechanism), then nous (consume new ariadne, retire bespoke `nous/nous/` tree). Backward compatibility at each phase gate.
