# weave — the layer-composition compiler (replacing setup.sh)

`cmd/weave` is ariadne's intent compiler: it composes each repo's agentic
context from its layer DAG, replacing the bash `construct/setup.sh` (see
[Setup & Replication](setup-and-replication.md)). Status: **in progress on
branch `000095-weave`** — issue [#95](../../workshop/issues/000095-weave.md),
design [plan](../../workshop/plans/000095-weave-plan.md).

## Shape (ARCH-PURE)
A pure pipeline — `read deps+manifests → Resolve → Plan → []Action → Apply` —
wrapped by a thin injected IO seam (filesystem; one `go mod edit` exec). Cloning
stays in the `bootstrap.sh` shell stub, so weave's IO is filesystem, not git.
Pure entities are unit-tested mock-free.

## Key decisions
- **Layer edges from `construct/deps` only** — resolved repo-root-relative for
  any path (directory-agnostic; `go.mod` is *not* a layer-discovery channel).
- **Hybrid intent vocabulary** — ported file-op verbs (`symlink`/`seed`/
  `scaffold`/`touch`/`merge`/`tool`) + new semantic `prose` (composes `AGENTS.md`,
  replacing the buggy `@AGENTS.local.md` @-import) and `skill` (served via
  `weave skill`).
- **Symlink-only** (vendor mode retired); agent-agnostic floor = a system prompt
  + shell, no `.claude/` assumptions in the core.

## Surface (grows per milestone)
- `cmd/weave/internal/layer` — `Resolve` (foundation-first topo-sort + dedup;
  ports `discover_ancestors`) and `ParseDeps` (`construct/deps` substrate-edge
  parser; ports `lib-deps.sh:deps_substrate_targets`). **[M1]**

Full spec, dep-model rule, and revisions live in the issue + plan above.
