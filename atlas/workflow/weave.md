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
wrapped by a thin injected IO seam: filesystem (`weavefs.FS`) **plus a narrow,
injected `.dynamic-skill` exec seam** (`weavefs.Runner`, #111 — see *Dynamic
skills* below). weave does not edit `go.mod` (the #95 M5 `go.mod` editor was
retired) and does not clone (that stays in the `bootstrap.sh` shell stub), so
weave's IO is filesystem + the one bounded exec; it is NOT git. Pure entities are
unit-tested mock-free; the exec seam is fake-tested (no real binary spawned).

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
- **A present substrate must be a compilable layer (#155)** — `construct/deps`
  `substrate` rows are layer edges (ParseDeps yields only those), so `layergraph.Walk`
  ERRORS (loud, actionable, naming the missing file) when a substrate target is
  present on disk but ships no `construct/base.manifest`. The pre-#155 walk silently
  skipped it, dropping the whole transitive chain below — a fresh-bootstrapped
  derivative under-compiled to a 1-action no-op with no signal. An **absent**
  substrate (peer not checked out) keeps the silent present-skip. The error is the
  single-source backstop for all three `Walk` consumers (weave, datatype,
  vocabulary). Companion: **`weave link` seeds** a minimal `construct/base.manifest`
  (header + `internal prose AGENTS.local.md`, one-source `seededBaseManifest`) in the
  linking repo when absent (idempotent, never clobbers), so a chain bootstrapped
  foundation-leafward — each repo seeding its own manifest at link time — compiles
  fully without any hand-authored manifest.

## Surface (grows per milestone)
- `pkg/layergraph` (module-level — imported by BOTH weave AND the `datatype`
  binary, #115) — the SINGLE source of "what is repo R's layer graph": `Walk`
  (transitive `construct/deps` topology → foundation-first ordered layer roots;
  ports `deps_substrate_targets` + the `_seen_or_add` filters), `Resolve`
  (foundation-first topo-sort + dedup; ports `discover_ancestors`), `ParseDeps`
  (`construct/deps` substrate-edge parser; ports `lib-deps.sh:deps_substrate_targets`),
  `FS` (the walk's IO seam). `cmd/weave/internal/layer` now carries only the
  resolved-layer value types (`Layer`, `ProseFragment`). `pkg/frontmatter` —
  flat-YAML `description:` parser, shared with weave's skill discovery. **[M1]**
- `cmd/weave/internal/{intent,plan,walk,weavefs,golden}` + `main.go` —
  `intent.ParseManifest` (base.manifest → hybrid intents) · `plan.{composeProse,
  Plan,Action,Apply}` (pure lowering + idempotent file-op apply, porting
  `create_symlink`/`create_scaffold` + inline `touch`) · `walk.Walk` (the per-layer
  LOADER on top of `pkg/layergraph.Walk`'s topology — loads each resolved root's
  manifest + prose fragments; the self-reference filter) · `weavefs.FS` (injectable IO seam) · `golden` (pure
  divergence classifier) + the `weave` / `weave --dry-run` / `weave golden` CLI.
  Prose/skill are exempt from the self-reference filter (a repo composes its own
  prose into its `AGENTS.md` — the `@AGENTS.local.md` fix). **[M2]**
- `cmd/weave/internal/skill` + `walk.GatherSkills` + `weave skills`/`weave skill
  <name>` — agent-agnostic skill server: `SkillIndex` (foundation-first,
  namespaced, downstream-overrides). Skills lower as per-harness skill-dir
  symlinks (`.claude/skills` for claude, `.agents/skills` for codex/gemini — #107),
  each harness discovering its own dir natively (NO `## Skills` menu); `weave skills`
  is a diagnostic listing, `weave skill <name>` serves a body on demand. Ports
  `sync-local-skills.sh` discovery (no `.claude/skills/` reliance).
  Plus `weave link <path>` (records `substrate <path>` verbatim — directory-
  agnostic; the module-include verb of weave's repo-composition dialect). **[M3]**
  (M3 originally also shipped a `tool` lowering — bimodal derivative→`substrate` /
  owner→`go mod edit -tool` via a `weavefs.GoModEditor` exec seam — **retired in
  M5**: ownership is location-based, weave does not edit `go.mod`.)
- `cmd/weave/internal/settingsx` + the `merge` lowering — the `settings`
  backend: pure `MergeChain`/`Merge`/`SemanticEqual` porting and extending
  `merge-settings.sh` (`$merge_keys` union, final-source `$remove` filter,
  meta-key strip). `plan.Plan` groups selected `merge` rows by target into one
  `MergeSettings{Sources, Target}`; `Apply` folds ordered layer sources
  foundation-first, then optional sibling `settings.local.json` last, into the
  generated target. The golden classifier recomputes the same chain and compares
  **semantically** (not byte-wise); `verify-complete` checks every manifest merge
  source is represented in the planned chain. No formal `Backend` interface —
  the `Action` sum type is the seam (YAGNI with a single backend). **[M4, #97]**
- **Cutover surface** — `weave compile` (the **Union** over every harness face by
  default; `--target {claude|codex|gemini}` for a lean subset; bare `weave` is
  help-only, mutates nothing) + `weave verify-complete` (completeness companion
  to `golden` — asserts the plan covers every managed path) · the `.claude/skills/<name>`
  symlink lowering (each pointing at the source layer's skill dir — absorbed the
  retired `sync-local-skills.sh` SessionStart hook; **unified into the pure
  `plan.SkillSymlinks` in #104 M1**, see below) ·
  `plan.PruneOrphans` (#96 — GCs orphaned lowered symlinks + the dead
  `setup.sh`/`merge-settings.sh`/`sync-local-skills.sh` cutover links; four
  conjunctive KEEP-unless safety criteria) · `plan.EnsureGitignore` (weave owns
  ignoring its generated-runtime set: `/CLAUDE.md`, `/AGENTS.md`, `/GEMINI.md`,
  `/.claude/skills/`, `/.agents/skills/`, `/.claude/settings.json`, `/.colima/`,
  `/construct/scripts/vm-log.sh`, `/construct/generated/` (#115 per-repo
  dynamic-skill materialization)) · the
  **export/internal visibility axis** (#99, `intent.Selected` — `𝒜(R)` = ancestors'
  exports ⊎ leaf's internals) · the `applyWriteFile` clobber-guard (removes a
  symlink at the slot before writing, so a derivative's pre-cutover
  `AGENTS.md`→ancestor symlink is never written through). The `tool` intent + the
  `GoMod`/`GoModEdit` exec seam were **retired** (location-based Go-tool ownership;
  weave does not edit `go.mod`). After M5 weave's IO was filesystem-only; #111
  re-adds exactly ONE bounded exec — the `.dynamic-skill` generate stage (below) —
  so weave's IO is now filesystem + a narrow exec seam, NOT the open-ended exec the
  retired `go.mod` editor was. **[M5 — cutover complete]**

- **Skill discovery unified (intent-driven + visibility-aware)** — the three
  disagreeing skill paths collapse to ONE: `walk.GatherSkills` reads each layer's
  `skill <dir>` INTENTS (not hardcoded `construct/local`+`adapted`) and stamps each
  entry's `Visibility` + `LayerIndex`; `skill.SelectVisible` applies the SAME
  `intent.Selected` 𝒜(R) filter prose uses (an ancestor's `internal` skill never
  reaches a consumer); the per-harness skill dirs (the pure `plan.SkillSymlinks`,
  ONE renderer per dir) lower from the IDENTICAL selected set — no `## Skills` menu
  (#107). The duplicate IO
  `walk.LowerSkillSymlinks` scan is deleted. Each layer prefixes its OWN skills via
  `skillPrefix` (`construct/config.json` `localPrefix`, else the layer's repo-name
  basename — ariadne pins `xx-`; `construct/adapted` stays bare). The subsystem
  invariant lives in the [skill-system](../../workshop/targets/skill-system.md)
  target. **[#104 M1+M2]**

- **Cross-repo skill migration** — the whole-dir `construct/{local,adapted}` +
  `construct/config.json` inheritance symlinks are GONE from every derivative
  (removed from ariadne's `base.manifest`): weave reads each ancestor's REAL skill
  dirs through the layer walk, so a derivative's `.claude/skills/xx-*` point
  straight at the owning layer (and weave's prune GCs the orphaned symlinks on
  re-weave). Each layer now resolves its OWN prefix (repo-name default; ariadne's
  real `config.json` pins `xx-`). nous owns its skills at `construct/local/{tools,
  resolve}` via one `skill construct/local` export intent — `nous-tools`/
  `nous-resolve` are now menu-listed + servable (`weave skill`), and inherited by
  its dependent brains through the layer. The `construct` skill is declared
  `internal skill construct/skill` (at `construct/skill/construct/SKILL.md`) →
  lowered as `xx-construct` on ariadne's self-walk only. `active-time-v3.py` (an
  ariadne tool that rode the dropped `construct/local` symlink) was owner-resolved
  by `sdlc actual` via `substrateChain` — until #110 ported it into the binary
  (`cmd/sdlc/internal/activetime`), retiring both the script and its
  owner-resolution. All 10 repos re-wove + verified (ancestors byte-pristine).
  **[#104 M3; superseded by #110]**

- **Per-harness skill-dir lowering (Option B)** — the skill backend stops being
  "`.claude/skills` symlinks XOR an AGENTS.md `## Skills` menu". Each harness gets a
  FACE = a pure-prose ENTRY FILE + a skill DIR (`plan.Target.Faces`): claude →
  `CLAUDE.md` + `.claude/skills`; codex → `AGENTS.md` + `.agents/skills`; gemini →
  `GEMINI.md` + `.agents/skills` (codex+gemini share the Agent Skills neutral
  `.agents/skills`). `weave compile` = the **Union** (every face, the default);
  `--target T` = the lean subset. The `## Skills` MENU is RETIRED — Codex/Gemini
  auto-compose their own from `.agents/skills`; `plan.SkillSymlinks(entries, dir)`
  is ONE renderer lowering the same selected set into each dir (ARCH-DRY), and
  `plan.Plan(layers, entryFiles)` fans the ONE composed prose to each entry file.
  Verified against the live CLIs by `scripts/harness-assumptions.test.sh`
  (`make harness-check`). The integration model + per-harness assumption ledger
  live in [harness-integration.md](harness-integration.md). A lean `--target X`
  compile PRUNES every OTHER face's stale artifacts (the original #107 bug): the
  prune scans `ManagedLocations(union-actions)` while the produced-set stays the
  lean compile's, so a codex compile GCs `.claude/skills` and a claude compile GCs
  `.agents/skills` — bidirectional, NO per-target registry (reuses the Union
  primitive + the existing `shouldPrune` safety criteria, ARCH-DRY); the Union
  prunes neither. **Cutover + propagation (M4):** ariadne dropped the `symlink
  CLAUDE.md` bridge + flipped the `Makefile.workflow` weave target to the Union
  default; then `sdlc propagate-base` (#106) re-wove all 10 recursive dependents
  foundation-first — each now carries `CLAUDE.md`/`AGENTS.md`/`GEMINI.md` (prose) +
  `.claude/skills` + `.agents/skills`, with the tracked `CLAUDE.md` bridge untracked
  (it's generated now). All 11 repos clean, ancestors byte-pristine,
  `make harness-check` green. **[#107 M2 produce + M3 prune + M4 propagate; tool #106]**

- **Dynamic skills — the `.dynamic-skill` exec seam (#111, reshaped by #115)** — a
  skill package may regenerate its own `SKILL.md` at compile time. The convention: a
  tracked, **executable `.dynamic-skill`** script in the package dir (language-neutral
  — weave never parses it). A **generate stage** runs in `weave compile`
  **after `walk.Walk`** (so the parsed `skill <dir>` intents exist to reuse, DRY)
  and **before `GatherSkills`/`planActions`** (so discovery reads the regenerated
  body). The output is **materialized per-repo at `construct/generated/<dir>/SKILL.md`,
  GITIGNORED in every repo (ariadne included)** — regenerated on every compile, never
  committed (#115 retired the old `construct/local/datatype/SKILL.md` committed
  codegen). Only the tracked `.dynamic-skill` marker stays in the package dir;
  `cmd/datatype/SKILL.md.tmpl` is the authored prose source.
  - **Marker-aware discovery.** The skill ENTRY is emitted from the TRACKED marker,
    not the generated body — so a dynamic skill is discovered even in a fresh,
    never-compiled clone (only the `description:` body is absent until first compile).
    This fixes #111's "skill vanishes in a fresh clone" failure mode.
  - **All-layers visible-set exec, leaf-rooted output.** The stage runs the
    visible-set markers across ALL layers (not leaf-only). For each marker — even one
    owned by an ANCESTOR — weave execs it with **cwd = the COMPILING repo's root** and
    a repo-relative `--output construct/generated/<dir>`, so materialization always
    lands in THE COMPILING repo's tree. The byte-pristine guarantee now rests on
    **leaf-rooted OUTPUT** (an ancestor's tree is never mutated by a derivative's
    compile), not on leaf-only SELECTION. `construct/adapted` is excluded
    (foreign-origin). The exec goes through the injected `weavefs.Runner` (production
    `ExecRunner` wraps `os/exec`, non-zero exit FAILS the compile loudly) —
    deliberately SEPARATE from `weavefs.FS`. The **read-only paths (`--dry-run`,
    `golden`, `verify-complete`) skip the stage** (they must not mutate).
  - **Lowering via BodyPath.** A dynamic skill's lowered `.claude/skills/xx-<name>`
    symlink points at **THIS repo's** `construct/generated/<dir>` (the skill entry's
    `BodyPath`); a static skill's link points at the owner layer's dir. So a
    derivative serves the dynamic body it materialized in its own tree.
  - **Prune class.** When an owner drops the `.dynamic-skill` marker, `PruneOrphans`
    GCs the now-orphaned `construct/generated/<dir>` (alongside the orphaned lowered
    symlinks).
  - **Drift guard retired; determinism guard in its place.** The #111 committed-file
    drift guard is GONE — a gitignored, regenerated-every-compile output can't go
    stale (git can't even see it). `make weave-drift-check` now asserts the render is
    **byte-deterministic across runs** (two compiles into temp dirs, diff the bytes),
    not that a committed file matches.
  - **Shared module-level libraries (#115 M1).** Two `pkg/` libraries underpin this:
    `pkg/layergraph` — the transitive `construct/deps` walk, the SINGLE source of
    "repo R's layer graph," imported by weave, the `datatype` binary, and (via `MergeByName`, #122) `cmd/vocabulary`
    so the DAG-aware tools never diverge on topology; and `pkg/frontmatter` — a flat-YAML
    `description:` parser shared by weave + datatype.
  - First consumer: `cmd/datatype` (`go:embed` prose template + the DAG-merged union
    of every layer's `construct/datatype/*.md` + the leaf's project-local
    `datatype/*.md`, local/leaf shadowing shared by filename) injects the live
    datatype-noun list into the generated `SKILL.md` description. Its apply-time
    prototype access is `datatype list` (enumerate the merged set) + `datatype show
    <name>` (read a resolved prototype body).
  **[#111 M1 mechanism + M2 datatype consumer; #115 per-repo gitignored materialization
  + DAG-merged datatype + marker-aware discovery]**

Full spec, dep-model rule, and revisions live in the issue + plan above.
