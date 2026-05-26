---
id: 000012
status: done
deps: []
created: 2026-04-27
updated: 2026-04-28
---

# typed markdown documents via construct

Add a "typed markdown" capability to ariadne so descendants can declare and use document types (e.g. travel-plan, meeting-notes, postmortem) whose frontmatter shape, body skeleton, and agent authoring instructions are bundled into a single prototype file. New types are pure data — adding one does not require a new skill.

## Done when

- A prototype lives in a well-known location under `construct/` and propagates to descendants via `construct adopt` / `upgrade`.
- A single dispatcher skill (working name: `markdown-types`) handles lookup-by-name and applies the prototype when generating or editing a typed file.
- Descendants can override or extend types by dropping a `.md` file into a project-local directory; no skill changes needed.
- A self-hosting meta-prototype defines the shape of a type prototype, so the system documents itself the same way `SKILL.md` does for skills.
- At least one seed prototype ships and is exercised end-to-end from a descendant repo (candidate target: nous).

## Spec

### Motivation

Markdown frontmatter is already a piece of structured mediation between humans and agents. The next step is to declare *typed documents* whose schema is read by the agent at write/edit time, not merely validated. The schema bundles three concerns:

1. **Frontmatter shape** — fields, types, required vs optional
2. **Body skeleton** — required sections, ordering, leading summary blocks
3. **Authoring instructions** — guidance the agent reads when generating or editing a doc of this type

This generalizes the `SKILL.md` / agentskills.io pattern: a typed document carries its own contract, and the contract doubles as a prompt.

### Why a dispatcher, not one skill per type

Skills are *trigger-based*: the harness picks them up by intent matching. Type prototypes are *looked up by name* ("user said `type: travel-plan`, find the spec"). Different mechanisms. One SKILL.md per type bloats the registry and makes activation noisy. A single dispatcher skill that consults a known directory keeps types as data and the skill list bounded.

### Storage and inheritance

- **Shared types:** under `construct/` in ariadne (exact path TBD during planning — likely `construct/markdown/<type>.md` as a peer to `construct/skill/`). Adopted into descendants via the existing construct flow — no new propagation mechanism.
- **Project-local types:** `<target>/markdown/<type>.md` (path TBD). Lookup is local-first so a descendant can override a shared type without forking.
- **Self-hosting:** a meta-prototype `type.md` defines the shape of a type prototype itself, mirroring how `SKILL.md` describes how to write `SKILL.md`.

### Dispatcher activation contract

Two triggers, combined:

- **Explicit creation:** `/markdown <type>` — user invokes when starting a new typed doc.
- **Implicit editing:** when the agent opens a file with `type: <X>` in frontmatter and `<X>.md` exists in a known location, it applies the prototype's authoring instructions.

Avoid description-based fuzzy intent matching as the primary trigger; reserve it as a fallback.

## Plan

- [x] Author meta-prototype `construct/data/type.md` (self-hosting; delegates new-type design to superpowers-brainstorming).
- [x] Author seed prototypes: `meeting-notes.md`, `travel-plan.md`.
- [x] Migrate xx-memory's 3 taxonomies to prototypes — renamed to `reference.md`, `procedure.md`, `event.md` (avoids `data` colliding with the system name; cleaner read in fuzzy intent matching).
- [x] Author dispatcher skill at `construct/local/data/SKILL.md` (slash: `/xx-data`). Conversation trigger is primary; slash and file-edit are secondary. Universal logic (location discovery, conversation distillation, ask-when-ambiguous) lives here, not per-prototype.
- [x] Update `construct/base.manifest`: `symlink construct/data`. Project-local prototype overrides at `<repo>/data/meta/` are opt-in (not scaffolded). Instances live under `<repo>/data/` or wherever the user chooses (`memory/...` etc.).
- [x] Remove `construct/local/memory/` (superseded).
- [x] Run `construct/scripts/sync-local-skills.sh` and verify symlinks.
- [x] Smoke test — synthesized a meeting-notes instance, verified prototype + dispatcher cohere, cleaned up.
- [x] Author `atlas/data-artifacts.md`, link from `atlas/index.md`.
- [ ] Adopt into nous out-of-band (separate session) for descendant validation.

## Log

### 2026-04-27

- Created from chat exploration. Key conclusions captured in Spec: (a) frontmatter-as-mediator generalizes to typed-document templates that prime the agent; (b) prototypes are data, dispatcher is the only skill; (c) ariadne is the right home so descendants inherit via construct, with project-local override for repo-specific types.

### 2026-04-28

- Refined design via chat. Decisions:
  - **Skill name:** `data` (slash: `/xx-data`). Generalizes the existing `xx-memory` skill, which is being deprecated. "memory" overloaded with agent-internal connotations; "data" is the artifact framing.
  - **Prototype path:** `construct/data/<name>.md` (shared, symlinked into descendants); `<repo>/data/<name>.md` for project-local override (opt-in, not scaffolded).
  - **Lookup precedence:** project-local → shared.
  - **Activation:** conversation trigger is primary ("capture this trip", "save this as X"); slash command and file-edit are secondary. Skill description surfaces conversational phrasing for fuzzy intent matching — overruling the original spec's "fallback only" position because it doesn't match how this gets used.
  - **Universal vs per-prototype logic:** location discovery (run `find . -type d`), conversation distillation, and ask-when-ambiguous live in the dispatcher. Per-prototype instructions only declare frontmatter shape, body skeleton, and what to ask for.
  - **Meta-prototype workflow:** creating a new type delegates to `superpowers-brainstorming` to explore the data shape before writing the prototype file.
  - **Implicit edit-time activation via base-layer `AGENTS.md`:** deferred. Conversation trigger covers the main path; revisit if the file-edit case becomes important.
  - **Seed types:** `meeting-notes`, `travel-plan`, plus `data`/`task`/`event` migrated from xx-memory.

- Implemented end-to-end. Files landed:
  - `construct/data/type.md` (meta-prototype, self-hosting)
  - `construct/data/meeting-notes.md`
  - `construct/data/travel-plan.md`
  - `construct/data/reference.md` (was `data.md` — renamed mid-implementation to avoid colliding with the system name "data"; reads better in intent matching: "save this as a reference" vs "save this as data")
  - `construct/data/procedure.md` (was `task.md` — renamed because `task` is overloaded with in-conversation tasks)
  - `construct/data/event.md`
  - `construct/local/data/SKILL.md` (dispatcher)
  - `construct/base.manifest` (added `symlink construct/data`)
  - `atlas/data-artifacts.md` + index link
  - Removed `construct/local/memory/` and the `xx-memory` symlink in `.claude/skills/`.
- All six prototypes structurally conform to the meta-prototype's contract (verified: each has type/name/description frontmatter and the three required body sections).
- One note for descendants on first `setup.sh` run after this commit: the new `construct/data` symlink will be created automatically. No action needed in the descendant beyond re-running setup.sh.
- Nous adoption deferred to a separate session — that's the descendant-side smoke test for end-to-end validation.

- **Iteration after first review.** User caught three issues:
  1. **Project-local prototype path:** `<repo>/data/<name>.md` was wrong — that path is for instances, not metadata. Changed to `<repo>/data/meta/<name>.md` so prototypes (metadata) and instances (data) are namespace-separated under the same `data/` root. Updated everywhere: dispatcher SKILL.md, `construct/data/type.md`, `atlas/data-artifacts.md`. (User had already updated the SKILL.md.)
  2. **Update-existing-instance flow missing.** Common case: "add Florence to our summer trip" implies updating an existing artifact, not creating one. Added `### 6. Update an existing instance from conversational context` to the dispatcher: detect implicit reference, find the file, treat as edit not rewrite, preserve provenance, confirm path once. Also reworded "Never silently overwrite" rule to distinguish unintentional collision (ask) from intentional update (one-line confirm OK).
  3. **Prototype-vs-instance shape ambiguity.** A prototype contains meta-sections (lede, Frontmatter shape, Body skeleton, Authoring instructions) that describe the instance but should NOT be copied into the instance. Made the convention explicit: "the prototype is a specification, not a template." Strengthened both the dispatcher's *Apply the prototype* step and `type.md`'s *Body skeleton* section + Rules. The meta-prototype case (applying `type.md` produces another prototype) is the only self-referential exception, and is called out explicitly.

- Net structural change after iteration: dispatcher gained a Step 6 and a clarified Step 3; meta-prototype gained a critical-note paragraph and a new rule. No new files. No changes to seed prototypes (they didn't violate the spec-vs-template rule, the rule was just implicit).
