---
type: target
slug: skill-system
status: active
created: 2026-06-15
updated: 2026-06-15
sources:
  - "ariadne#104 — skill-system v2 (the build); gaps surfaced on the #95 nous/brain cutover"
  - "weave-composition-algebra — the compose stage this target defers to"
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
[[weave-composition-algebra]] (its `skill` slice); this target OWNS the rest of
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
   is claude-only and bypasses the subsystem).
2. **Identify** — a stable **namespaced name** = the declaring layer's prefix +
   base name; this name is the composition key and the harness-facing handle.
   Collision-free by per-layer prefix (`xx-` ariadne, `nous-` nous, `metis-` …).
3. **Compose** — `skills(R) = ⋃ᵢ export-skills(Lᵢ) ∪ internal-skills(Lₙ)`, keyed by
   name, the leaf's internal included, ancestors' internal excluded. (The math is
   [[weave-composition-algebra]]; this target only requires the subsystem *use* it
   — one discovery feeding the composed set.)
4. **Lower** — the composed set renders per harness target: claude → a
   `.claude/skills/<name>` symlink into the SOURCE LAYER's skill dir; codex/agy →
   the `## Skills` menu compiled into AGENTS.md. **The set is target-independent;
   only the rendering differs.**
5. **Serve** — every composed skill is servable via `weave skill <name>` (and
   listed by `weave skills`) regardless of target — the agent-agnostic body access.
6. **Inherit** — a layer owns its OWN skills in a REAL per-layer dir; weave's
   layer-walk aggregates across layers. No inheritance-by-symlink (no derivative
   `construct/local → ../ariadne` that double-scans and blocks per-layer skills).

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
  [[weave-composition-algebra]]; it does not restate the union math.
- **Not target-coupled** — a skill is not "a claude thing"; it is a skill,
  rendered per harness. The declaration never names a target.
- **Not inheritance-by-symlink** — layers don't symlink each other's skill dirs;
  the layer-walk is the aggregation.

## Open questions (to settle in ariadne#104's plan)

- **Visibility granularity** — per-row (`export|internal skill <dir>`, the natural
  grammar fit) vs per-skill (frontmatter)? A dir is the coarse unit; mixed-visibility
  skills in one dir would need either two dirs or a frontmatter override.
- **Identity source** — dir-name (today) vs SKILL.md frontmatter `name:`. Dir-name
  couples layout to identity; frontmatter decouples but adds a parse dependency.
- **Inheritance-symlink removal** — kill `construct/local`/`construct/adapted →
  ariadne` in derivatives in one move, or phase it (it currently masks the
  double-scan via dedup)?
- **Suppression** — should a derivative be able to RETRACT an inherited export
  (drop a skill)? Today no (additive/override-only), shared with the algebra's
  open question.
