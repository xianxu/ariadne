---
type: target
slug: skill-system
status: active
created: 2026-06-15
updated: 2026-06-17
sources:
  - "ariadne#104 — skill-system v2 (the build); gaps surfaced on the #95 nous/brain cutover"
  - "ariadne#111 — dynamic skills (the .dynamic-skill maintenance stage; cmd/datatype)"
  - "base-layer-mechanics — the compose stage this target defers to"
---

# Target: Ariadne skill system — one declaration, every harness, visibility-aware

The ariadne-styled skill system turns a layer's authored skills into
agent-discoverable capability across an arbitrary layer DAG and across every
harness (claude, codex, gemini, …). The durable commitment this target defends: a
skill is **declared once, semantically** — as a skill, with a visibility — and the
system guarantees it is discovered once, composed by the algebra, lowered
identically *in meaning* to every harness, and servable. Behavior must NOT depend
on the declaration MECHANISM or the active TARGET; it depends only on the skill's
nature (its visibility + the layer it lives in).

This is the **subsystem** target. The cross-cutting compose math lives in
[[base-layer-mechanics]] (its `skill` slice); this target OWNS the rest of
the pipeline — declaration, identity, the per-harness lowerings, serving,
inheritance — and *references* the algebra for the compose stage rather than
re-deriving it.

## Why now — and the lesson that justifies a separate target

The algebra target already had a `skill` slice marked "clarity: HIGH" with the
formula `skills(R) = ⋃ export-skills(Lᵢ) ∪ internal-skills(Lₙ)`. But that slice
was **generalized from the prose fix (#99) by analogy and never built or tested**:
no skill code consults visibility, there are *three* disagreeing discovery
mechanisms, and only the claude target is ever exercised (and claude routes around
every gap). The target asserted settled math over an unbuilt subsystem — and so
**actively hid the gap**: "skill: HIGH" told us it was solved, so we did not look.

Two structural consequences this target encodes:
1. **A target must separate design-clarity from implementation-status.** A formula
   understood is not a system built. Every "clarity HIGH" claim binds to a test or
   a verified pass, or it is a wish with a confidence label.
2. **The subsystem is more than its compose stage.** Declaration grammar, identity/
   prefix, per-harness lowering, serving, and inheritance are NOT in the algebra —
   they need their own defended invariant. Hence this target.

## The pipeline — the invariant, stage by stage

A skill flows through six stages; each has an invariant the implementation must
honor (and a test must pin). A skill's `SKILL.md` source may optionally be
**maintained** (regenerated) at compile time — stage 0 below — upstream of
Declare:

0. **Maintain (optional) — dynamic skills (#111, reshaped by #115).** A skill package
   may own an executable, tracked **`.dynamic-skill`** script that `weave compile`
   execs to **materialize its `SKILL.md` per-repo at `construct/generated/<dir>/SKILL.md`,
   GITIGNORED in every repo (ariadne included), regenerated on every compile** — never
   committed (#115 retired the old committed `construct/local/datatype/SKILL.md`). This
   keeps the SKILL body current (e.g. `cmd/datatype` injects the live datatype-noun
   list into the description) — it is **maintenance**, kept distinct from
   **composition** (the union) and **lowering** (the per-harness symlinks). The skill
   ENTRY is emitted from the TRACKED `.dynamic-skill` marker, so the skill is
   **discovered even in a fresh, never-compiled clone** (only the description body is
   absent until first compile — the #111 "skill vanishes in a fresh clone" fix). The
   generate stage runs after `walk.Walk`, before `GatherSkills`, over the
   **all-layers visible-set** (NOT leaf-only): each marker — even an ancestor-owned
   one — is exec'd with **cwd = the COMPILING repo's root** + repo-relative
   `--output construct/generated/<dir>`, so materialization lands in THAT repo's tree.
   The byte-pristine guarantee now rests on **leaf-rooted OUTPUT** (an ancestor's tree
   is never mutated by a derivative's compile), not leaf-only selection.
   `construct/adapted` is excluded, the exec runs through an injected `weavefs.Runner`
   (a non-zero exit fails the compile), and read-only paths skip the stage. Lowering
   points the dynamic skill's symlink at THIS repo's `construct/generated/<dir>` (via
   the entry's BodyPath); `PruneOrphans` GCs an orphaned `construct/generated/<dir>`
   when the owner drops the marker. The #111 committed-file drift guard is **retired**
   (a gitignored, every-compile-regenerated output can't go stale); `make
   weave-drift-check` now asserts the render is **byte-deterministic across runs**.
   Only `.dynamic-skill` is hand-authored + tracked; `SKILL.md` is generated +
   gitignored. Most skills are static (no marker) and skip this stage entirely. The
   mechanism lives in the [weave atlas](../../atlas/workflow/weave.md); weave's IO is
   therefore filesystem + this one bounded exec seam, not filesystem-only.

1. **Declare** — ONE mechanism: a `[export|internal] skill <dir>` row in a layer's
   `base.manifest`. No ad-hoc `symlink … .claude/skills/X` skills (that mechanism
   is claude-only and bypasses the subsystem). A layer holds skills in up to three
   conventional dirs, and **the dir encodes the visibility**:
   - **`construct/adapted`** — skills with an EXTERNAL source (e.g. superpowers),
     imported and conformed to ariadne's conventions via the `construct` skill →
     **export**.
   - **`construct/local`** — skills constructed locally in the layer and meant to
     be shared → **export**.
   - **`construct/skill`** — skills PRIVATE to the layer (used while developing it,
     never exported) → **internal**. The `construct` skill itself lives here and is
     the exemplar — now an *internal declaration* (`internal skill construct/skill`),
     lowered as `xx-construct` on ariadne's OWN self-walk and never leaked into a
     derivative (ancestor-internal). v2 M3 replaced the old keep-out-by-location.
2. **Identify** — a stable **namespaced name** = the declaring layer's prefix +
   base name; this name is the composition key and the harness-facing handle.
   Collision-free by per-layer prefix. **The prefix defaults to the layer's repo
   name** (`nous-`, `brain-`, `metis-`), set in the layer's OWN `config.json`
   (no longer a symlink to ariadne's); ariadne keeps its explicit `xx-` override.
   So `nous/construct/local/tools` + `nous-` → `nous-tools` (the move from
   `construct/skills/nous-tools` lands the same name, uniformly).
3. **Compose** — `skills(R) = ⋃ᵢ export-skills(Lᵢ) ∪ internal-skills(Lₙ)`, keyed by
   name, the leaf's internal included, ancestors' internal excluded. (The math is
   [[base-layer-mechanics]]; this target only requires the subsystem *use* it
   — one discovery feeding the composed set.)
4. **Lower** — the composed set renders to each harness's own skill DIR (Option B,
   #107): claude → `.claude/skills/<name>` symlinks; codex + gemini → `.agents/skills/<name>`
   symlinks (the Agent Skills neutral path). Each link points into the SOURCE LAYER's
   skill dir. There is NO `## Skills` menu — every harness discovers its dir natively
   (Codex/Gemini auto-compose their own). The per-harness ENTRY FILES (CLAUDE.md /
   AGENTS.md / GEMINI.md) are pure prose. **The set is target-independent; only the
   destination dir differs.**
5. **Serve** — every composed skill is servable via `weave skill <name>` (and
   listed by `weave skills`) regardless of target — the agent-agnostic body access.
6. **Inherit** — a layer owns its OWN skills in a REAL per-layer dir; weave's
   layer-walk aggregates across layers. Two *different* symlinks must not be
   confused:
   - **KEEP** the per-skill lowering link — `.claude/skills/<name>` → the SOURCE
     LAYER's skill dir (`brain/.claude/skills/xx-fix → ../../../ariadne/construct/
     local/fix`). This IS the claude lowering; it points at whichever layer owns
     the skill.
   - **REMOVE** the whole-dir inheritance link — `<derivative>/construct/local →
     ../ariadne/construct/local`. The layer-walk reads each ancestor's dirs
     directly, so this is redundant; it causes the double-scan and physically
     blocks a layer from owning its own `construct/local`.

## Invariants (each must be test-bound)

- **One declaration mechanism** — no plain-symlink skills; a skill is a `skill`
  intent.
- **One discovery** — every per-harness skill dir (`.claude/skills` + `.agents/skills`)
  and the `weave skill` serve path see the IDENTICAL composed set (one
  `walk.GatherSkills`, rendered into each dir by `plan.SkillSymlinks`).
- **Visibility honored** — an ancestor's `internal` skill never reaches a
  consumer; the leaf's `internal` always does; all `export` do. (Same selection
  prose uses.)
- **Target-independence of the SET** — a skill behaves the same whether claude or
  codex reads it; only the lowering differs.
- **Every skill servable** — `weave skill <name>` resolves any composed skill.
- **Per-layer identity** — each layer sets its own prefix (its own `config.json`),
  no double-prefix, no mis-prefix from a borrowed config.
- **Dynamic-skill maintenance is per-repo materialized + bounded (#111, #115)** —
  the `.dynamic-skill` generate stage execs the all-layers visible-set markers (never
  `construct/adapted`) through an injected `Runner` whose non-zero exit fails the
  compile, each with cwd = the COMPILING repo's root + repo-relative
  `--output construct/generated/<dir>`, so output is leaf-rooted (an ancestor's tree
  is never mutated). The body is materialized at `construct/generated/<dir>/SKILL.md`,
  **gitignored + regenerated every compile**; the skill entry is emitted from the
  tracked marker (discovered in a fresh clone); read-only paths skip the stage;
  `PruneOrphans` GCs an orphaned `construct/generated/<dir>`. The committed-file drift
  guard is retired — `make weave-drift-check` now asserts byte-determinism across
  runs. (The generate-stage fake-Runner + dry-run-skip tests, the marker-aware
  discovery test, and `cmd/datatype`'s faithfulness + determinism tests.)

## What this is NOT

- **Not a second composition algebra** — it defers the compose stage to
  [[base-layer-mechanics]]; it does not restate the union math.
- **Not target-coupled** — a skill is not "a claude thing"; it is a skill,
  rendered per harness. The declaration never names a target.
- **Not inheritance-by-symlink** — layers don't symlink each other's skill dirs;
  the layer-walk is the aggregation.

## Decided (operator, 2026-06-15)

- **Visibility granularity = per-dir convention.** The three dirs encode it
  (`adapted`/`local` → export, `skill` → internal). Per-skill frontmatter is the
  fallback only if mixed visibility *within one dir* ever becomes common.
- **Identity = dir-name** (keep today's; frontmatter `name:` deferred).
- **Inheritance-symlink removal = yes** — drop the whole-dir `construct/{local,
  adapted} → ariadne` links; keep the per-skill `.claude/skills/*` links.
- **Prefix = repo-name default**, set per-layer in the layer's own `config.json`;
  ariadne overrides to `xx-`.

## Status

**Built — ariadne#104 M1–M3 (M3 closed 2026-06-16).** Every invariant above is
implemented and test-bound (`skill.SelectVisible`, `walk.GatherSkills` intent-
driven + repo-name prefix, `plan.SkillSymlinks`, `weave skill`/`skills`), and the
cross-repo migration put all 10 ariadne-styled repos onto per-layer real skill
dirs + repo-name prefixes with the whole-dir inheritance symlinks dropped. The
`construct` skill is the internal exemplar (`xx-construct`, ariadne-only).

**Dynamic-skill maintenance (stage 0) — ariadne#111 (M1 mechanism + M2 datatype
consumer), reshaped by ariadne#115.** The `.dynamic-skill` generate stage + the
injected `weavefs.Runner` exec seam ship test-bound (adapted-excluded, dry-run-skip,
non-zero-fails-compile); `cmd/datatype` is the first consumer.

**#115 (DAG-merged dynamic skills, per-repo datatype enumeration across the layer
graph).** The committed codegen at `construct/local/datatype/SKILL.md` is RETIRED;
the body is now materialized **per-repo at `construct/generated/<dir>/SKILL.md`,
gitignored everywhere + regenerated every compile**. Discovery is marker-aware (the
skill entry comes from the tracked `.dynamic-skill`, so it survives a fresh clone —
#111's "skill vanishes" fix). The generate stage runs the **all-layers visible-set**
(not leaf-only) with cwd = the compiling repo's root + repo-relative `--output`, so
output is leaf-rooted and ancestors stay byte-pristine; lowering points the symlink
at this repo's `construct/generated/<dir>` via BodyPath; prune GCs an orphaned dir.
The committed-file drift guard is retired — `make weave-drift-check` now asserts
byte-determinism. Datatype prototypes are now **per-layer-owned** (the
`symlink construct/datatype` manifest row is retired) and read as a **DAG-merged
union** by the `datatype` binary (every layer's `construct/datatype/` + the leaf's
project-local `datatype/`, local/leaf shadowing shared by filename); apply-time
access is `datatype list` + `datatype show <name>`. Two shared module-level libraries
back this: `pkg/layergraph` (the transitive `construct/deps` walk — the single source
of "repo R's layer graph," imported by BOTH weave and the `datatype` binary) +
`pkg/frontmatter` (the flat-YAML `description:` parser shared by weave + datatype).
Migration outcome: nous owns `event`/`travel-plan`/`reference`; ariadne owns the
generic remaining set; a repo depending on ariadne directly (pair, 42shots, …) sees
only ariadne's set.

## Open questions

- **Suppression** — should a derivative be able to RETRACT an inherited export
  (drop a skill)? Today no (additive/override-only), shared with the spine's
  open question. The `SelectVisible` predicate is where it would extend.
- **Resolved (v2 M3): `construct/skill` shape** — now aligned with `local`/
  `adapted`'s `<name>/SKILL.md` layout (`construct/skill/construct/SKILL.md`), so
  the private dir scans uniformly and weave names it from the dir.
