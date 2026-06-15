---
type: target
slug: skill-system
status: active
created: 2026-06-15
updated: 2026-06-15
sources:
  - "ariadne#104 — skill-system v2 (the build); gaps surfaced on the #95 nous/brain cutover"
  - "base-layer-mechanics — the compose stage this target defers to"
---

# Target: Ariadne skill system — one declaration, every harness, visibility-aware

The ariadne-styled skill system turns a layer's authored skills into
agent-discoverable capability across an arbitrary layer DAG and across every
harness (claude, codex, agy, …). The durable commitment this target defends: a
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
honor (and a test must pin):

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
     the exemplar — today kept out of derivatives by its *location*; v2 makes that
     an *internal declaration*.
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
4. **Lower** — the composed set renders per harness target: claude → a
   `.claude/skills/<name>` symlink into the SOURCE LAYER's skill dir; codex/agy →
   the `## Skills` menu compiled into AGENTS.md. **The set is target-independent;
   only the rendering differs.**
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
- **One discovery** — the menu/serve path and the `.claude/skills` path see the
  IDENTICAL composed set (no `GatherSkills`-vs-`LowerSkillSymlinks` divergence).
- **Visibility honored** — an ancestor's `internal` skill never reaches a
  consumer; the leaf's `internal` always does; all `export` do. (Same selection
  prose uses.)
- **Target-independence of the SET** — a skill behaves the same whether claude or
  codex reads it; only the lowering differs.
- **Every skill servable** — `weave skill <name>` resolves any composed skill.
- **Per-layer identity** — each layer sets its own prefix (its own `config.json`),
  no double-prefix, no mis-prefix from a borrowed config.

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

## Open questions (to settle in ariadne#104's plan)

- **Suppression** — should a derivative be able to RETRACT an inherited export
  (drop a skill)? Today no (additive/override-only), shared with the spine's
  open question.
- **`construct/skill` shape** — today it holds one SKILL.md directly (the
  `construct` skill); align it with `local`/`adapted`'s `<name>/SKILL.md` layout
  so the private dir scans uniformly.
