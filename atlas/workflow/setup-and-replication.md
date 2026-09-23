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

## Numbered workspace environments

A numbered checkout is nested, for example
`/workspace/worktree/pair-slot1/pair`. Weave verifies its registered host Git
worktree and treats `/workspace/worktree/pair-slot1` as the private source
environment. Every substrate edge, including transitive edges, must name a
direct sibling ordinary checkout. Escapes, symlink aliases, deeper layouts and
linked dependency checkouts are rejected. This rule also applies when invoked
from a dependency or its Git-verified feature worktree.

For a missing source, `construct/deps` must record a remote URL. Weave creates an
ordinary clone at branch main and verifies `origin/main` before publication,
regardless of the remote default branch. Local/file sources are unsupported
inside numbered environments. Existing ordinary checkouts are reused without
fetching, pulling, resetting or switching branches, including dirty state and
unpublished commits. Recorded origins must match. A source-less existing clone
works, but needs a recorded URL before it can be restored. Explicit removal and
recreation uses current remote main, not the prior selected SHA.

To choose another revision, use normal Git in the private dependency, then
recompile the consuming checkout. Publish dependency changes using that clone's
normal issue/review/PR flow, prerequisites first. There is no recursive merge,
shared dependency shelf, lockfile, or local-primary source inference. Dependency
feature worktrees retain SDLC context but may not satisfy literal relative
manifest paths; compose from the environment's dependency primary in that case.

Couch calls `weave compile` from the nested main checkout with default targets.
Exit 0 means preparation and composition completed. Progress is human-readable;
failures exit 1 with diagnostics. Couch records readiness only on success and
does not compile/fetch on ready resume. After failure, display the diagnostic
and explicitly retry the same command after its cause is resolved. Missing
source/main needs metadata/remote correction; contention needs the active setup
to finish. No diagnostic-string parser or automatic destructive recovery is
required. `weave dependencies` prepares sources/packages without composing;
`--dry-run` creates no lock or files and fails for an incomplete graph.

One stable `.weave-setup.lock` per environment serializes setup across host and
dependency callers. Live clone/build/generator descendants inherit the OS-held
lease, so parent death alone cannot admit another writer. Retry reuses complete
clones and existing owned-stage recovery; it never deletes local dependency work.
The lock file persists until explicit environment removal. Different environments
can prepare independently. Trusted producers must preserve inherited descriptors
and remain in their assigned process group.

Source isolation does not isolate Homebrew packages or data mounts. Owner-local
tools stay in each clone's `bin`; normal setup never changes installed tool
selection or shell configuration. Explicit installation remains an operator action.

## Bootstrap and Make

The real, seeded `bootstrap.sh` is a small gateway: change to its own repo root,
reuse a compatible weave on PATH, otherwise install `xianxu/ariadne/weave` with
Homebrew, then exec `weave compile`. When a legacy source binary shadows the
installation, bootstrap invokes the formula's binary directly. Missing Homebrew
produces installation guidance; bootstrap does not install Homebrew itself.
The formula is pending publication in #241. Until then, build `cmd/weave` from
ariadne and put that candidate on PATH.

Each consumer owns its root `Makefile`. If absent, compile seeds the generic
`construct/Makefile.seed` once. If a regular Makefile already exists, compile
prepends `-include Makefile.workflow` and preserves the rest of the file; a
symlink or other non-regular path is preserved and compile prints the same
adoption instruction. Shared `make weave` and `make bootstrap` delegate to
`weave compile`; the latter uses a prerequisite so consumer setup extensions
can remain additive. Generic startup has no recursive layer Make builds, shell
clone walker, or sdlc-install phase.

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

## Owner tool selection

Each layer's authored root `Makefile` explicitly lists the binaries its `tools`
target builds. Startup no longer scans `cmd/*` or requires `.skip-make-build`
sentinels. Layers without tools need no target. Keep signing, service restarts,
and other product development actions in separate explicit targets; bootstrap
prepares dependencies and ordinary local builds. The separate generic
`make build` command retains its `cmd/*` scanner and `.skip-make-build` opt-out;
it is not part of startup.

## Generated artifacts and local extensions

Binaries build in their source owner's `bin/`; they are not distributed as
manifest-owned Go source. Dynamic skills publish under the leaf's
`construct/generated/` only after staged output passes validation and ownership
checks. An executable `.dynamic-skill` must declare the exact comment
`# weave-output: argv1`, accept an absolute isolated output directory as its first
argument, and write a regular, nonempty `SKILL.md` there. Additional regular
files are collected too. Cwd remains the leaf so generators can read its graph.
Legacy markers lacking this declaration fail before execution; this is a trusted
layer-code contract, not a shell sandbox. Markers and children must remain in
the assigned process group, preserve inherited descriptors, and not daemonize.

The tracked datatype and vocabulary markers preserve direct invocation with
`"${1:-construct/generated/datatype}"` and
`"${1:-construct/generated/vocabulary}"`. During compile, weave supplies the
staging path. Generation and clone stages share an ownership lifecycle: cleanup
must retain metadata until all producers stop. Parent death alone does not permit
reclamation; cancellation and retry must account for surviving descendants.
An inherited stage lease protects live writers, and cancellation terminates
the assigned process group. Marker authors migrate the script; users keep the same compile command.

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

## Consumer cutover

#239 prepares and tests startup and migration tooling. #241 owns the actual
release, tap publication, and consumer edits after the implementation merges.
For each consumer, keep these inputs committed before testing a fresh clone:

- A source-bearing `construct/deps` row for every missing peer that bootstrap
  must restore. `weave link` records the checkout origin when available.
- A layer-owned root `Brewfile`, containing its external development prerequisites.
  Use Homebrew's platform conditionals for platform-specific packages.
- An authored root `Makefile` with optional generated workflow includes. Add
  `tools` only for binaries owned by this layer; keep service/signing targets
  explicit and independent.
- The current bootstrap launcher, manifest, local sources, and compatible dynamic
  markers. Markers accept the supplied staging output directory; see
  [the generator contract](weave.md#dynamic-skill-output-contract).

After the gateway is published, test `./bootstrap.sh` from a fresh clone and
repeat it. Add the reported owner `bin/` directories to the developer's PATH.
Use `sdlc propagate-base` only with clean consumer working trees; its
[ownership-aware index migration](base-layer.md#pushing-updates-to-all-consumers)
preserves intentionally tracked ignored files. Do not blanket-untrack ignored
paths. Product checks, signing, and service startup remain separate.

Disposable pilots during #239 established the following boundaries:

| Consumer snapshot | Migration inputs exercised | Evidence and remaining limitation |
|---|---|---|
| parley.nvim `c8dcd56a` | Authored root with optional local workflow include; root Brewfile for Neovim/Python; no owner tools | Linux compile and bootstrap repeat passed. Product sources and intentionally tracked vocabulary `issue.json` remained unchanged. Full product lint was not claimed. |
| nous `2e7756d5` | Replace inherited root Makefile link with authored root and `tools: nous-build`; keep `nous-bootstrap`/`nous-dev` explicit; guard macOS packages in Brewfile | macOS ordinary owner build and safe help passed without service/signing actions. Full bootstrap remains unverified: Linux code requires launchd and the existing Mutagen tap formula targets amd64 and needs Homebrew trust handling. Resolve those consumer choices in #241 rather than porting nous in #239. |

Actual peer repositories were not modified for these pilots.
