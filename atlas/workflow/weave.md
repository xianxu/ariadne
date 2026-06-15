# weave — the layer-composition compiler (replacing setup.sh)

`cmd/weave` is ariadne's intent compiler: it composes each repo's agentic
context from its layer DAG, replacing the bash `construct/setup.sh` (see
[Setup & Replication](setup-and-replication.md)). Status: **in progress on
branch `000095-weave`** — issue [#95](../../workshop/issues/000095-weave.md),
design [plan](../../workshop/plans/000095-weave-plan.md).

## Shape (ARCH-PURE)
A pure pipeline — `read deps+manifests → Resolve → Plan → []Action → Apply` —
wrapped by a thin injected IO seam (filesystem only; weave does not edit
`go.mod`). Cloning stays in the `bootstrap.sh` shell stub, so weave's IO is
filesystem, not git. Pure entities are unit-tested mock-free.

## Key decisions
- **Layer edges from `construct/deps` only** — resolved repo-root-relative for
  any path (directory-agnostic; `go.mod` is *not* a layer-discovery channel).
- **Hybrid intent vocabulary** — ported file-op verbs (`symlink`/`seed`/
  `scaffold`/`touch`/`merge`) + new semantic `prose` (composes `AGENTS.md`,
  replacing the buggy `@AGENTS.local.md` @-import) and `skill` (served via
  `weave skill`). The `tool` verb was **retired** in #95 M5 — Go-tool ownership
  is location-based (`construct/dev-aliases.sh` + build-in-owner), so weave never
  edits `go.mod`; the substrate edge comes from `weave link` / `construct/deps`.
- **Symlink-only** (vendor mode retired); agent-agnostic floor = a system prompt
  + shell, no `.claude/` assumptions in the core.

## Surface (grows per milestone)
- `cmd/weave/internal/layer` — `Resolve` (foundation-first topo-sort + dedup;
  ports `discover_ancestors`) and `ParseDeps` (`construct/deps` substrate-edge
  parser; ports `lib-deps.sh:deps_substrate_targets`). **[M1]**
- `cmd/weave/internal/{intent,plan,walk,weavefs,golden}` + `main.go` —
  `intent.ParseManifest` (base.manifest → hybrid intents) · `plan.{composeProse,
  Plan,Action,Apply}` (pure lowering + idempotent file-op apply, porting
  `create_symlink`/`create_scaffold` + inline `touch`) · `walk.Walk` (transitive
  `construct/deps` walk; ports `deps_substrate_targets` + the `_seen_or_add` +
  self-reference filters) · `weavefs.FS` (injectable IO seam) · `golden` (pure
  divergence classifier) + the `weave` / `weave --dry-run` / `weave golden` CLI.
  Prose/skill are exempt from the self-reference filter (a repo composes its own
  prose into its `AGENTS.md` — the `@AGENTS.local.md` fix). **[M2]**
- `cmd/weave/internal/skill` + `walk.GatherSkills` + `weave skills`/`weave skill
  <name>` — agent-agnostic skill server: `SkillIndex` (foundation-first,
  namespaced, downstream-overrides), menu compiled into `AGENTS.md` + bodies on
  demand, ports `sync-local-skills.sh` discovery (no `.claude/skills/` reliance).
  Plus `weave link <path>` (records `substrate <path>` verbatim — directory-
  agnostic; the module-include verb of weave's repo-composition dialect). **[M3]**
  (M3 originally also shipped a `tool` lowering — bimodal derivative→`substrate` /
  owner→`go mod edit -tool` via a `weavefs.GoModEditor` exec seam — **retired in
  M5**: ownership is location-based, weave does not edit `go.mod`.)
- `cmd/weave/internal/settingsx` + the `merge` lowering — the `settings`
  backend: pure `Merge`/`SemanticEqual` porting `merge-settings.sh`
  (`$merge_keys` union, `$remove` filter, meta-key strip, local-overrides-base);
  the `MergeSettings` action reads `.claude/settings.ariadne.json` + optional
  `settings.local.json` → `.claude/settings.json`; the golden classifier
  compares **semantically** (not byte-wise). No formal `Backend` interface — the
  `Action` sum type is the seam (YAGNI with a single backend). **[M4]**

Full spec, dep-model rule, and revisions live in the issue + plan above.
