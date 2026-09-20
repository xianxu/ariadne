# Setup & Replication

`weave` owns source preparation and layer composition. `construct/deps` declares
the graph; `construct/base.manifest` declares what each layer contributes.
`go.mod` remains a Go dependency file and does not declare substrate edges.
The old setup.sh, vendor mode, and recursive shell bootstrap are retired.

## Adopt or restore a repository

With a compatible weave on PATH, run from the repository adopting the base:

```sh
weave link github.com/xianxu/ariadne
# Or: weave link ../ariadne
weave compile
```

An address link clones a sibling and records its source. A local link records
an available Git origin. Matching existing checkouts are reused without pulling
or resetting, including dirty checkouts. Link seeds a minimal local manifest
when absent. A layer declaration has the form:

```text
substrate ../ariadne https://github.com/xianxu/ariadne.git
data https://github.com/example/content.git data/content
```

The substrate source is optional while its local path exists; restoration of a
missing checkout requires a source. Data rows declare a source and owner-relative
mount. Paths and sources cannot contain whitespace or `#` in this format.

## Preparation and compilation

`weave dependencies` restores transitive sources and runs each layer's root
`Brewfile` foundation-first with `brew bundle install --no-upgrade --file=Brewfile`.
Homebrew is the package manager on both macOS and Linux. This command does not
build tools, mount data, or generate artifacts. Layers without a Brewfile need
no package operation. A dry-run mutates nothing and reports an incomplete preview
with failure when a missing source prevents reading its declarations.

`weave compile` performs these phases in order:

1. Restore sources and provision layer Brewfiles.
2. Run each participating owner's `make tools`, foundation-first.
3. Reconcile data mounts in each declaring owner.
4. Run selected generators with leaf cwd and isolated output directories, then
   publish their validated outputs with the composed artifacts through ownership
   checks.

A failure stops subsequent phases. Ancestor preparation does not recursively
compile ancestor contexts. Owner `tools` targets must build from tracked sources
before generated helpers exist; each owner declares its explicit tools in its
own root Makefile and writes them to its own `bin/`. Ariadne's target builds
`sdlc`, `datatype`, `vocabulary`, and `doc-review`. Distributed weave is built
separately for development or release.

Compilation supplies owner bin directories to child processes and prints them
for humans to add to PATH explicitly. It never edits a shell rc. Compile dry-run
skips builds, package installation, generators, and writes; generator output and
retirement are not previewed.

## Bootstrap and Make

The real, seeded `bootstrap.sh` is a small gateway: change to its own repo root,
reuse a compatible weave on PATH, otherwise install `xianxu/ariadne/weave` with
Homebrew, then exec `weave compile`. When a legacy source binary shadows the
installation, bootstrap invokes the formula's binary directly. Missing Homebrew
produces installation guidance; bootstrap does not install Homebrew itself.
The formula is pending publication in #241. Until then, build `cmd/weave` from
ariadne and put that candidate on PATH.

Each consumer authors its root `Makefile`; ariadne no longer seeds it. An
optional `-include Makefile.workflow` exposes shared workflow commands after
compilation while keeping standalone product targets available. Shared
`make weave` and `make bootstrap` delegate to `weave compile`; the latter uses a
prerequisite so consumer setup extensions can remain additive. Generic startup
has no recursive layer Make builds, shell clone walker, or sdlc-install phase.

The seeded merge-check workflow sets up Homebrew on Linux and compiles before
running the optional executable `scripts/ci-setup.sh` and local merge-check
runner. Ariadne's source job builds and uses its current candidate CLI after
provisioning its root Brewfile; consumer jobs use bootstrap's published formula.

## Ownership and retirement

Composition applies manifest intents such as `symlink`, `seed`, `scaffold`,
`touch`, `merge`, `prose`, and `skill`. Seeds are content-tracking real files for
entrypoints that must survive an isolated checkout, including bootstrap and CI.
Authored roots and owner-local build declarations are not shared seeds.

`construct/generated/weave/ownership.json` records output identities separately
for artifact and data scopes. On a later compile, an output removed from the
plan is retired only while it still matches its recorded identity. Edited and
unrecognized files are preserved; no inventory means no speculative cleanup.
Managed ignore entries cover the generated outputs while preserving authored
ignore content. Data reconciliation in an ancestor retains its artifact scope.

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

## Generated artifacts and local extensions

Binaries build in their source owner's `bin/`; they are not distributed as
manifest-owned Go source. Dynamic skills publish under the leaf's
`construct/generated/` only after staged output passes validation and ownership
checks. An executable `.dynamic-skill` must declare the exact comment
`# weave-output: argv1`, accept an absolute isolated output directory as its first
argument, and write a regular, nonempty `SKILL.md` there. Additional regular
files are collected too. Cwd remains the leaf so generators can read its graph.
Legacy markers lacking this declaration fail before execution; this is a trusted
layer-code contract, not a shell sandbox.

The tracked datatype and vocabulary markers preserve direct invocation with
`"${1:-construct/generated/datatype}"` and
`"${1:-construct/generated/vocabulary}"`. During compile, weave supplies the
staging path. Generation stages use the same PID/host ownership lifecycle as
clone stages: failures do not publish generator output, and retry reclaims dead
stages. Marker authors migrate the script; users keep the same compile command.

The static sdlc skill points to `sdlc --help`, avoiding
a duplicate generated copy of the workflow contract.

Shared text is edited in its source layer; per-repository rules and settings
belong in local fragments. Recompile after changing layer declarations or
manifest inputs. Data mounts consume content without applying that content
repo's manifest; see [data-deps.md](data-deps.md).

## Related

- [weave.md](weave.md) — compiler structure and intent semantics.
- [base-layer.md](base-layer.md) — shared surface, local extensions, and VMs.
- `construct/base.manifest` — ariadne's exported and internal intents.
