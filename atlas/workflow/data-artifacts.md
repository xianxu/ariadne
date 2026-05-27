# Data Artifacts

A typed-document system for capturing conversational substance into structured markdown files. Implemented by the `xx-datatype` skill plus a directory of pluggable prototypes. (Skill is named `datatype` to keep the meta-concept distinctive; the artifacts it produces are *data* and live under `data/`.)

## Mental model

- **Prototype** — a markdown file declaring the frontmatter shape, body skeleton, and authoring instructions for one *kind* of artifact. Lives in `construct/datatype/`.
- **Instance** — a markdown file shaped by some prototype. Lives wherever the user wants — typically under `memory/` in a directory whose name carries categorical meaning.
- **Dispatcher** — the `xx-datatype` skill. Owns lookup, conversation distillation, location discovery, and prototype application. Same code path for every type.
- **Meta-prototype** — `construct/datatype/type.md`. Self-hosting: applying it produces a new prototype.

New types are pure data. Adding one means writing a new `<name>.md`, not modifying any skill.

## Directory layout

| Path | Role |
|---|---|
| `construct/datatype/<name>.md` | Shared prototypes, propagate to descendants via `base.manifest`. |
| `<repo>/datatype/<name>.md` | Project-local prototype override (opt-in, not scaffolded). Shadows shared completely. Top-level `datatype/` keeps prototype definitions cleanly separated from instances (which live under `<repo>/data/` or wherever the user puts them). |
| `construct/local/datatype/SKILL.md` | The dispatcher skill, symlinked to `.claude/skills/xx-datatype/`. |

Lookup precedence: project-local → shared.

## Activation

The dispatcher fires on three triggers, in priority order:

1. **Conversational capture** — "capture this trip", "save these meeting notes", "remember this list", "track this launch". This is the common case. The dispatcher uses three-step judgment (classify turn → discriminate substance from generative → semantic-match against prototype descriptions); no enumerated noun→type table. See `construct/local/datatype/SKILL.md` §1 for the judgment procedure.
2. **Slash invocation** — `/xx-datatype <type> [path]`. Bypasses judgment.
3. **Edit-time** — opening a file with `type: <X>` in frontmatter applies `<X>.md`'s authoring instructions to the edit.

## Adding a new type

1. User says they want to start tracking a new kind of artifact.
2. Dispatcher applies `type.md` (the meta-prototype).
3. Per `type.md`'s instructions, that delegates to `superpowers-brainstorming` to design the frontmatter shape, body skeleton, and authoring instructions.
4. Result: a new `<name>.md` written to `construct/datatype/` (shared) or `<repo>/datatype/` (project-local), depending on scope.
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
| `product` | Durable charter of a thing being built — vision + components + current state. Spans 0..N peer repos. |
| `roadmap` | Month-level forward-looking plan for one product — capacity, scope decisions, target state per component. Lives at `data/roadmap/YYYYMM/<product>.md`. |
| `project` | Execution container — focused push toward a defined MVP, cutting across issues and possibly products. Operator-POV. One operator per project. Closes the velocity calibration loop. |
| `prose` | Per-parent ledger of pre-manuscript fragments — sentences and half-thoughts captured before they have a home in the parent's drafts. Sibling to a `product` (or other long-running parent). Reverse-chrono, append-only, voice-preserving. Distinct from `pensive` (session vs ledger). |

The trio `product` + `roadmap` + `project` together model a small company-or-team's structure:

- **product** = what is being built (durable, static)
- **roadmap** = what we want true by month T (forward-looking, monthly)
- **project** = what we're working on right now to advance it (operator-POV, time-bounded by an MVP)

The pair `pensive` + `prose` together cover thinking-out-loud at two different granularities:

- **pensive** = a *session* — one focused topic per file, write-then-publish, hundreds-to-thousands of words. Filename encodes the topic (`<date>-NN-pensive-<topic>.md`); the artifact stands alone.
- **prose** = a *ledger* — many fragments per file, append-as-you-go, sentence-to-paragraph each. Filename is just `prose.md`; the file is bound to a parent artifact (typically a long-running `product` like a book or blog), and fragments graduate into the parent's drafts.

Heuristic: *session or ledger?* A fragment that grows past ~3 paragraphs and develops a thesis has become a session — graduate it from prose to pensive. A pensive dominated by one-line observations rather than connected argument should have been prose entries.

## Rules of thumb

- **Frontmatter is for queryable, stable, externally-referenced fields.** Free-form prose stays in the body.
- **Directory names carry categorical meaning.** `memory/life/family-travel/` says different things from `memory/work/travel/`. The dispatcher respects existing conventions; it doesn't invent a parallel structure.
- **One prototype per type, one type per file.** Filename and `type:` field must agree.
- **Local override shadows shared completely.** No field-level merging.

## Pointers

- Skill: `construct/local/datatype/SKILL.md`
- Prototypes: `construct/datatype/`
- Manifest entry: `symlink construct/datatype` in `construct/base.manifest`
- Issue: `workshop/issues/000012-typed-markdown-documents-via-construct.md`
