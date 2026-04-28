# Data Artifacts

A typed-document system for capturing conversational substance into structured markdown files. Implemented by the `xx-data` skill plus a directory of pluggable prototypes.

## Mental model

- **Prototype** — a markdown file declaring the frontmatter shape, body skeleton, and authoring instructions for one *kind* of artifact. Lives in `construct/data/`.
- **Instance** — a markdown file shaped by some prototype. Lives wherever the user wants — typically under `memory/` in a directory whose name carries categorical meaning.
- **Dispatcher** — the `xx-data` skill. Owns lookup, conversation distillation, location discovery, and prototype application. Same code path for every type.
- **Meta-prototype** — `construct/data/type.md`. Self-hosting: applying it produces a new prototype.

New types are pure data. Adding one means writing a new `<name>.md`, not modifying any skill.

## Directory layout

| Path | Role |
|---|---|
| `construct/data/<name>.md` | Shared prototypes, propagate to descendants via `base.manifest`. |
| `<repo>/data/meta/<name>.md` | Project-local prototype override (opt-in, not scaffolded). Shadows shared completely. The `meta/` segment keeps prototypes namespaced separately from instances, since instances also live under `<repo>/data/` (or wherever the user puts them). |
| `construct/local/data/SKILL.md` | The dispatcher skill, symlinked to `.claude/skills/xx-data/`. |

Lookup precedence: project-local → shared.

## Activation

The dispatcher fires on three triggers, in priority order:

1. **Conversational capture** — "capture this trip", "save these meeting notes", "remember this list", "track this launch". This is the common case.
2. **Slash invocation** — `/xx-data <type> [path]`.
3. **Edit-time** — opening a file with `type: <X>` in frontmatter applies `<X>.md`'s authoring instructions to the edit.

## Adding a new type

1. User says they want to start tracking a new kind of artifact.
2. Dispatcher applies `type.md` (the meta-prototype).
3. Per `type.md`'s instructions, that delegates to `superpowers-brainstorming` to design the frontmatter shape, body skeleton, and authoring instructions.
4. Result: a new `<name>.md` written to `construct/data/` (shared) or `<repo>/data/meta/` (project-local), depending on scope.
5. Available immediately as a type — no skill change, no propagation step beyond the existing construct flow.

## Built-in types

| Type | Use for |
|---|---|
| `type` | Meta-prototype. Apply to add new types. |
| `meeting-notes` | Distilled record of a sync — attendees, decisions, action items. |
| `travel-plan` | One trip — destination, dates, itinerary, bookings, status. |
| `reference` | Evergreen, mostly-static info — lists, vendors, contacts. |
| `procedure` | Steps to follow for a repeatable or in-flight task. |
| `event` | Time-bound plan with a deadline — launch, conference, prep effort. |
| `pensive` | Timestamped train of thought, insight, brainstorm. Captures a moment of thinking-out-loud in the user's voice. |

## Rules of thumb

- **Frontmatter is for queryable, stable, externally-referenced fields.** Free-form prose stays in the body.
- **Directory names carry categorical meaning.** `memory/life/family-travel/` says different things from `memory/work/travel/`. The dispatcher respects existing conventions; it doesn't invent a parallel structure.
- **One prototype per type, one type per file.** Filename and `type:` field must agree.
- **Local override shadows shared completely.** No field-level merging.

## Pointers

- Skill: `construct/local/data/SKILL.md`
- Prototypes: `construct/data/`
- Manifest entry: `symlink construct/data` in `construct/base.manifest`
- Issue: `workshop/issues/000012-typed-markdown-documents-via-construct.md`
