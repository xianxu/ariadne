# weave — the layer-composition compiler (replaced setup.sh)

`cmd/weave` is ariadne's intent compiler: it composes each repo's agentic
context from its layer DAG, replacing the bash `construct/setup.sh` (see
[Setup & Replication](setup-and-replication.md)). Status: **cutover complete (M5)**
— all 10 ariadne-styled repos compile via weave (`make weave`); `setup.sh` +
the `merge-settings.sh`/`sync-local-skills.sh` hooks retired. Issue
[#95](../../workshop/issues/000095-weave.md), design
[plan](../../workshop/plans/000095-weave-plan.md). The composition invariant lives
in the [base-layer-mechanics](../../workshop/targets/base-layer-mechanics.md) target.

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
- **Cutover surface** — `weave compile --target <claude|codex|agy>` (bare `weave`
  is help-only, mutates nothing) + `weave verify-complete` (completeness companion
  to `golden` — asserts the plan covers every managed path) · the `.claude/skills/<name>`
  symlink lowering (each pointing at the source layer's skill dir — absorbed the
  retired `sync-local-skills.sh` SessionStart hook; **unified into the pure
  `plan.SkillSymlinks` in #104 M1**, see below) ·
  `plan.PruneOrphans` (#96 — GCs orphaned lowered symlinks + the dead
  `setup.sh`/`merge-settings.sh`/`sync-local-skills.sh` cutover links; four
  conjunctive KEEP-unless safety criteria) · `plan.EnsureGitignore` (weave owns
  ignoring its generated-runtime set: `/AGENTS.md`, `/.claude/skills/`,
  `/.claude/settings.json`, `/.colima/`, `/construct/scripts/vm-log.sh`) · the
  **export/internal visibility axis** (#99, `intent.Selected` — `𝒜(R)` = ancestors'
  exports ⊎ leaf's internals) · the `applyWriteFile` clobber-guard (removes a
  symlink at the slot before writing, so a derivative's pre-cutover
  `AGENTS.md`→ancestor symlink is never written through). The `tool` intent + the
  `GoMod`/`GoModEdit` exec seam were **retired** (location-based Go-tool ownership;
  weave's IO is filesystem-only). **[M5 — cutover complete]**

- **Skill discovery unified (intent-driven + visibility-aware)** — the three
  disagreeing skill paths collapse to ONE: `walk.GatherSkills` reads each layer's
  `skill <dir>` INTENTS (not hardcoded `construct/local`+`adapted`) and stamps each
  entry's `Visibility` + `LayerIndex`; `skill.SelectVisible` applies the SAME
  `intent.Selected` 𝒜(R) filter prose uses (an ancestor's `internal` skill never
  reaches a consumer); the menu (`skill.Build`) and the claude links (the pure
  `plan.SkillSymlinks`) lower from the IDENTICAL selected set. The duplicate IO
  `walk.LowerSkillSymlinks` scan is deleted. Each layer prefixes its OWN skills via
  `skillPrefix` (`construct/config.json` `localPrefix`, else the layer's repo-name
  basename — ariadne pins `xx-`; `construct/adapted` stays bare). The subsystem
  invariant lives in the [skill-system](../../workshop/targets/skill-system.md)
  target. **[#104 M1+M2]**

Full spec, dep-model rule, and revisions live in the issue + plan above.
