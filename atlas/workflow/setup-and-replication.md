# Setup & Replication

`construct/setup.sh` is the unified base-layer-replication mechanism. One
canonical script lives in ariadne and gets vendored down to every derivative
via `construct/base.manifest`. Same source-of-truth file at every layer.

Design rationale: `workshop/issues/000032`.

## What it does

For a target repo, walks each transitive upstream layer's
`construct/base.manifest` in topological order (ancestors first), applying
the manifest's `symlink` / `copy` / `merge` / `scaffold` / `touch` actions.
Then runs post-processing: creates `Makefile` + `Makefile.local` if absent,
applies `.gitignore` entries, syncs local-skill symlinks, records mode.

The "walk N manifests" generalizes today's depth-specific scripts:
- ariadne at depth 0: no ancestors, self-refresh just runs skill sync.
- nous at depth 1 (post-migration): walks ariadne's manifest, applies into nous.
- baby brain at depth 2: walks ariadne's, then nous's, then its own
  contributions (if any).

## Ancestor discovery — go.mod is the authoritative manifest

**Every ariadne-style derivative declares its upstream(s) in its own `go.mod`
via `replace` directives.** This is the convention regardless of whether the
derivative is itself a Go project — a pure-Lua plugin like parley.nvim still
writes a 3-line `go.mod` purely to declare its substrate ancestor:

```
module github.com/xianxu/parley.nvim

go 1.22

replace github.com/xianxu/ariadne => ../ariadne
```

The `go.mod` here is functioning as the **dependency-management manifest**,
not as a "this is a Go project" declaration. It explicitly records what the
repo consumes and where to find it. Transitive chains (baby brain → nous →
ariadne) just need the immediate `replace`; setup.sh's recursive walker
follows each upstream's own `go.mod` to discover the rest.

### Why go.mod (and not a separate `.upstreams` file)

- **It's already the convention** in any repo that has Go code at all.
  Post-ariadne#31 every ariadne-style repo will gain Go code eventually
  (sdlc binary, project-specific tooling); deferring the convention until
  Go arrives means churn later. Better to write the 3-line go.mod up-front
  and have one consistent dependency mechanism across the whole layer
  chain.
- **Transitive resolution is free.** setup.sh's recursive replace-walk uses
  `go.mod`'s own grammar; no parallel parser to maintain.
- **Versioning evolves naturally.** When ariadne (or any upstream) goes
  public, the same `replace` line becomes `require <module> <version>` for
  pin-mode — no migration of the dependency mechanism itself.
- **Explicit in-tree record.** Anyone reading the derivative can see in
  `go.mod` exactly which upstreams it consumes. The pre-#32 model
  communicated this only by invocation path (`../ariadne/construct/setup.sh`)
  with no on-disk evidence.

### Three discovery sources, in priority order

When setup.sh runs, ancestor candidates are collected from:

1. **Recursive `replace` walk** starting at target's `go.mod`. Each replaced
   path's own `go.mod` is then probed for further replaces. BFS through the
   chain, reversed to topological order (deepest = foundation first).
2. **`go list -m -f '{{.Dir}}' all`** — picks up modules referenced by real
   Go-import code (require lines that survive `go mod tidy`). Added to
   ancestors not already discovered.
3. **Script's own resolved upstream** — last-resort fallback when no `go.mod`
   exists at all. Preserved for first-time bootstrap (running
   `../ariadne/construct/setup.sh` from a brand-new directory) and for
   genuinely-old consumers that haven't yet written `go.mod`. **Not the
   recommended steady state** — derivatives should write `go.mod` after
   first adoption.

Candidates are filtered to dirs that ship `construct/base.manifest` and
deduped. Target's own manifest is walked separately after ancestors.

## Single deployment model — symlink + bootstrap-cascade

Per ariadne#38: substrate is symlink-only. There's no `--vendor` /
`--symlink` mode flag, no `.ariadne-mode` marker, no copy action. All
substrate text symlinks from upstream peers (sibling-checkout); Go tool
sources resolve via `replace => ../<peer>` directives in
`construct/go.mod`.

**Six manifest actions:** `symlink`, `tool`, `scaffold`, `touch`, `merge`,
`seed`. Earlier versions also had `copy` — retired in #38. For per-derivative
divergence (operator wants to customize a substrate file), the pattern is
**per-operator branches in upstream source repos**, not per-derivative
copies in the derivative tree. One branch per operator's preferences,
shared across all that operator's derivatives.

`seed` (added #42) is the one copy-shaped action — a **write-once** copy of a
real file into the target, mode-preserving, never overwriting. It exists for
the single case symlinks can't serve: a **fresh-clone entrypoint that must run
before any substrate is present**. It is not a `copy` revival — `copy` let
operators diverge substrate (discouraged); `seed` delivers a generic,
not-meant-to-be-edited file once. Sole user today: `bootstrap.sh`.

### Fresh-clone first-run — `./bootstrap.sh`

A bare `git clone` of a derivative has only dangling symlinks (Makefile,
construct/setup.sh, AGENTS.md, … all point at a not-yet-cloned upstream), so
`make` can't read its own Makefile and **no target — including `bootstrap` —
exists**. `bootstrap.sh` (real committed file, `seed`ed from ariadne) breaks the
chicken-and-egg: it reads the real `go.mod`/`construct/go.mod` and clones the
upstream peers as siblings (URL derived from `origin`, same convention as
bootstrap-peers.sh), then `exec make bootstrap` once the symlinks resolve.
Idempotent. See #42.

The clone walk is **transitive** (in-process BFS), not direct-only (#45). A
3-deep chain (`foo → mid → ariadne`) symlinks `foo/Makefile → ../mid/Makefile →
../ariadne/Makefile`, so `make` can't read its Makefile until the *whole* chain
is present — cloning only the direct peer (`mid`) would dangle at `ariadne` and
the `make bootstrap` that would clone `ariadne` could never start. So
bootstrap.sh clones a peer, reads *that peer's* go.mod, and continues until the
tree is complete, then hands off once. (In-process BFS rather than recursing
into each peer's `bootstrap.sh`: the latter's `exec make` tail would orphan the
top repo, and BFS needs only each peer's go.mod, not a committed bootstrap.sh.)
Test hooks `BOOTSTRAP_DRY_RUN` / `BOOTSTRAP_CLONE_ONLY`; hermetic coverage in
`construct/scripts/test/bootstrap-transitive.test.sh`.

### Bootstrap cascade (`make bootstrap`)

```
make bootstrap (in any repo)
  → make ensure-go: guarantee the Go toolchain first (#61) — sdlc is a
    base-layer build dep; no-op if present, brew-installs on macOS, else
    fails fast before the costly peer-clone cascade. See sdlc-binary.md.
  → construct/scripts/bootstrap-peers.sh: parses construct/deps for
    `substrate ../<name>` rows (#60). For each missing peer:
       - Derives clone URL from current repo's origin (substitutes
         this-repo-name → peer-name; operator override via Makefile.local)
       - git clone to ../<name>
       - Recursively `make bootstrap` in the peer
    Carries visited-set + depth-limit (max 5) via env vars.
  → make weave (peers exist now → symlinks materialize)
  → make tools  (builds tools via Go's replace → sibling resolution)
  → derivative's local-env setup hook (Makefile.nous/Makefile.local extends)
```

### `construct/deps` — the substrate-peer graph carrier (#60)

The substrate peer graph lives in **`construct/deps`** — a flat, positional,
grep-parseable manifest (`<kind> <target> [<mount>]`, `kind ∈ {substrate,data}`;
substrate targets are repo-root-relative, e.g. `../ariadne`). It replaced the
old `construct/go.mod`-as-peer-graph hack (which forced a fake `<name>-construct`
Go module onto every derivative, even markdown-only brains). The parse lives in
the shared **`construct/scripts/lib-deps.sh`** (`deps_substrate_targets`,
`deps_data_rows`), sourced by the symlinked walkers; `bootstrap.sh` keeps an
inline `walk_deps` mirror (it runs pre-substrate, can't source the symlink),
locked by the drift test.

**Same peer set, four consumers** — `substrate` rows in `construct/deps` drive
four independent walkers that must agree:

1. **`construct/setup.sh`** (`discover_ancestors`) — substrate symlink resolution.
2. **`construct/scripts/bootstrap-peers.sh`** — the clone cascade above.
3. **`.tart/scripts/tart-list-peers.sh`** → `list-peers.sh` — which repos to
   APFS-clone into the tart VM (`make tart`, #32).
4. **`bootstrap.sh`** — fresh-clone entrypoint (#42); runs *before* any substrate
   exists (the others run after, via the symlinked Makefile), so it keeps an
   **inline copy** of the parser.

Each walker reads `construct/deps` for the substrate ancestor(s) **and** the
repo-root `go.mod` for real Go app-dep siblings (e.g. brain's `replace nous =>
../nous`, which also makes nous an ancestor). It resolves depth-≥2 chains
(brain → nous → ariadne) by walking each node's own `construct/deps`
recursively. The legacy `construct/go.mod` substrate carrier is **no longer
read** (#60 M4 retired the dual-read fallback once every derivative carried
`construct/deps`). Regression-guarded by the test trio, which keeps an explicit
root-`go.mod` app-dep case so that path can't silently regress.

Two walk *modes* over the one grammar (#45): **list-present** (setup.sh, tart)
*skips* absent dirs; **clone-absent** (bootstrap.sh, bootstrap-peers.sh) resolves
syntactically and clones them. bootstrap.sh's inline parser is locked to
`lib-deps.sh` by the drift test in
`construct/scripts/test/bootstrap-transitive.test.sh`.

**Data deps** (content this repo consumes, not substrate it inherits) are
`kind: data` rows in the same `construct/deps`, mounted by `clone-data-deps.sh`
(clone sibling + relative symlink). The legacy two-column `construct/data-deps`
file was retired in #60 M5.

### Weave vs bootstrap

- **`make weave`** — pure substrate-state sync. Peers must exist
  (errors with hint if missing). Runs `construct/setup.sh` to update
  symlinks. NO clone, NO build.
- **`make bootstrap`** — first-time setup OR full state recovery. Cascades:
  clones missing peers, refreshes substrate, builds tools, runs local-env
  setup. Idempotent — re-runnable.

### How divergence works (no `copy` action)

Operators who want to customize a substrate artifact:

1. `cd ../<peer>` (the source repo)
2. `git checkout -b <operator>-customizations`
3. Edit on that branch
4. All derivatives consuming that peer see the changes via their symlinks

To merge upstream improvements: standard `git merge main` in the source
branch periodically. Per-operator branches stay small (one per operator);
all their derivatives share.

## Adopting a fresh derivative

The recommended pattern at every layer — Go-using or not — writes a
`go.mod` upfront so the upstream is explicitly declared:

```bash
cd /path/to/new-derivative
git init

# Minimal go.mod — declares the upstream relationship regardless of
# whether this derivative has its own Go code.
cat > go.mod <<EOF
module github.com/<owner>/<derivative>

go 1.22

replace github.com/xianxu/ariadne => ../ariadne
EOF

../ariadne/construct/setup.sh --yes
```

After this:
- `AGENTS.md`, `Makefile.workflow`, scripts, etc. are linked into the derivative.
- `Makefile`, `Makefile.local`, `AGENTS.local.md` are created if absent.
- `.ariadne-mode` records `symlink` (or `vendor` if `--vendor` was passed).
- `construct/setup.sh` itself is linked, so the derivative can self-refresh.
- The derivative's `go.mod` declares ariadne as its upstream — future
  refreshes use this explicit record instead of falling back to script-
  upstream inference.

### Skipping go.mod for first-time bootstrap

The fallback path (no go.mod → script's resolved upstream) makes it possible
to run setup without writing `go.mod` first. This is fine for the very first
invocation when you don't know the module path yet. But the recommended
workflow is to write `go.mod` either before or immediately after — having
an explicit on-disk record of what your repo consumes is worth the three
lines.

### Subsequent updates

`make weave` re-runs setup.sh against the upstream location Go resolves.
Bumping a pinned version = editing the `require` line. Switching to
trunk-follow on a sibling = changing the `replace` RHS to `../<upstream>`.
All upstream-relationship changes happen in `go.mod`; setup.sh just acts on
what it finds there.

## Per-binary build opt-out

`Makefile.workflow build:` scans `cmd/*/main.go` and builds each. Some binaries
shouldn't be auto-built (e.g., the nous binary is distributed signed +
notarized — overwriting `bin/nous` with an unsigned local build invalidates
macOS keychain ACL grants and notification capabilities).

Opt out per-binary by dropping a sentinel:

    cmd/<name>/.skip-make-build

Contents are free-form prose explaining the rationale (future operators read
it). Each opted-out binary owns its sentinel; the base layer doesn't carry
derivative-specific name lists.

## Generated artifacts (not vendored)

Artifacts that are *functions of code shipping through Go modules* (e.g.,
`bin/sdlc`, built from `cmd/sdlc`) are NOT shipped via `base.manifest`. They
build at the consumer via the install/refresh path (`make sdlc-build`), so the
version of the source determines the version of the artifact automatically — no
text-vs-code lockstep drift possible.

The inverse choice — keeping a derived doc *static* rather than generated —
applies to `construct/local/sdlc/SKILL.md` (the `xx-sdlc` skill): it's a thin
pointer to `sdlc --help` carrying no copy of the contract, so it can't drift
either. (It was previously regenerated from a now-retired `sdlc --index`; the
generator existed only to keep a duplicated copy in sync, which the pointer
makes unnecessary.)

## Who writes the substrate declaration

> **Retired (#95 M5).** The `tool <path>` manifest verb described below was
> `setup.sh`'s mechanism; weave **retired it**. Go-tool ownership is now resolved
> by **location** — `construct/dev-aliases.sh` scans sibling `cmd/X` dirs and
> build-in-owner builds each to `OWNER/bin/`. The substrate edge is declared
> directly via `weave link` / `construct/deps`, not derived from a `tool` row,
> and weave **never edits `go.mod`** (the owner-side `go mod edit -tool` served
> only goland). The section below is historical, describing setup.sh's behavior.

A derivative declared its substrate ancestor via the `tool <path>` manifest
action (ariadne's base.manifest carried `tool cmd/sdlc`). setup.sh's
`ensure_go_tool_dependency` was the writer, split by whether the target IS the
tool's owner:

- **Cross-target** (a derivative): appended `substrate ../ariadne` to
  `construct/deps` (#60). Repo-root-relative, idempotent, language-agnostic — no
  Go needed. The walkers read it; `make sdlc-build` resolves + builds the tool in
  its owner (build-in-owner, #60 M2). `construct/go.mod` is no longer written or
  read (the stubs were deleted from every derivative in #60 M4).
- **Self-walk** (the owner, e.g. ariadne): added a `go mod edit -tool` directive
  to the owner's own root go.mod so `go tool <name>` worked locally. Ariadne had
  no substrate ancestor of its own, so it wrote no `construct/deps` row.

**Multi-layer composition.** If a derivative descends from multiple tool-owning
ancestors, each owner's `tool` action appended its own `substrate` row; rows
dedupe by resolved path. The bootstrap cascade clones every declared peer.

### Historical: `construct/go.mod` — the retired writer target (pre-#60)

Before #60, the cross-target writer stubbed a **second go.mod**
(`<name>-construct`) per derivative carrying `require + replace + tool` for each
ancestor, and `make sdlc-build` built `cd construct && go build` through its
`replace => ../../ariadne`. That used go.mod as a substrate-peer graph — not its
purpose — and forced a fake Go module onto even non-Go (markdown brain)
consumers. #60 retired it: M1 moved the peer graph to `construct/deps`, M2 moved
the build to build-in-owner, M3 flipped this writer, M4 deleted the stub modules
from every derivative + dropped the walkers' dual-read fallback, and M5 retired
the legacy `construct/data-deps` reader. `construct/go.mod` is now unused.

Design rationale: `workshop/issues/000037` (the original split), `000038`
(symlink-only), `000060` (the deps-manifest unification that retired it).

## Data dependencies (content peers, not substrate)

A **data dependency** is content a repo *consumes* from another repo (a sibling
clone surfaced via relative symlink), distinct from the *substrate* peer
mechanism above — it's clone-only, never applies a `base.manifest`, and is
language-agnostic. Full how-to (add / remove / bootstrap) lives in its own doc:
**[data-deps.md](data-deps.md)**.

## Related

- `construct/base.manifest` — what ariadne contributes (action + path
  pairs).
- `construct/setup.sh` — this script.
- `workshop/issues/000032` — design rationale, three operating modes,
  migration plan.
- `atlas/workflow/base-layer.md` — adopting ariadne's base layer (this
  document supersedes parts of it; cross-reference as the system stabilizes).
