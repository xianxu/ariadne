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
| `construct/datatype/<name>.md` | Prototypes OWNED by this layer. Each layer owns its own `construct/datatype/`; the `datatype` binary reads the **DAG-merged union** across the layer graph (every layer's `construct/datatype/` + the leaf's project-local `datatype/`). NOT propagated by symlink (#115 retired the `symlink construct/datatype` manifest row). |
| `<repo>/datatype/<name>.md` | Project-local prototype override (opt-in, not scaffolded). Shadows shared completely. Top-level `datatype/` keeps prototype definitions cleanly separated from instances (which live under `<repo>/data/` or wherever the user puts them). |
| `construct/generated/datatype/SKILL.md` | The dispatcher skill body, lowered to `.claude/skills/xx-datatype/`. **Generated per-repo + GITIGNORED (#115)** — `weave compile` runs the package's `.dynamic-skill` marker (→ the `datatype` binary) to materialize it in THIS repo's tree with the live datatype-noun list in the `description`; never committed. **Edit `cmd/datatype/SKILL.md.tmpl`, not this file** (a hand edit here is overwritten by the next `weave compile`). |
| `construct/local/datatype/.dynamic-skill` | Executable marker (hand-authored, the only TRACKED file in the package dir) that makes the package a dynamic skill — `weave compile` execs it to materialize `construct/generated/datatype/SKILL.md`. The skill entry is discovered from THIS marker even before first compile. See [weave atlas → Dynamic skills](weave.md). |

Lookup precedence: project-local → shared. Migration outcome (#115): nous owns `event`/`travel-plan`/`reference`; ariadne owns the generic remaining set; a repo depending on ariadne directly (pair, 42shots, …) sees only ariadne's set.

## Activation

The dispatcher fires on three triggers, in priority order:

1. **Conversational capture** — "capture this trip", "save these meeting notes", "remember this list", "track this launch". This is the common case. The dispatcher uses three-step judgment (classify turn → discriminate substance from generative → semantic-match against prototype descriptions); no enumerated noun→type table. See `construct/generated/datatype/SKILL.md` §1 (the per-repo materialized body) for the judgment procedure; at apply time the dispatcher enumerates the DAG-merged set with `datatype list` and reads a resolved prototype with `datatype show <name>`.
2. **Slash invocation** — `/xx-datatype <type> [path]`. Bypasses judgment.
3. **Edit-time** — opening a file with `type: <X>` in frontmatter applies `<X>.md`'s authoring instructions to the edit.

## Adding a new type

1. User says they want to start tracking a new kind of artifact.
2. Dispatcher applies `type.md` (the meta-prototype).
3. Per `type.md`'s instructions, that delegates to `superpowers-brainstorming` to design the frontmatter shape, body skeleton, and authoring instructions.
4. Result: a new `<name>.md` written to `construct/datatype/` (shared) or `<repo>/datatype/` (project-local), depending on scope.
5. Available immediately as a type — no skill change, no propagation step beyond the existing construct flow.

## Built-in types

This is the cross-stack catalog. Post-#115 it is **per-layer-owned**: `event`,
`travel-plan`, `reference` are **nous-owned** (`nous/construct/datatype/`, a
personal-assistant concern), the rest are ariadne's generic base set; the
`datatype` binary DAG-merges per repo, so a given repo sees the union over its own
layer graph (ariadne sees the base set without the three nous nouns).

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
| `continuation` | The **connective narrative over a session's durable artifacts** (pensive/issues/targets) — next action, the thread's arc + a model of the user's intention, open questions to resolve on resume, decisions/dead-ends, and lessons — so work resumes later / on another machine / by another person / under another agent. Distilled from the *rendered* session, not the native store. Lives at `workshop/continuation/<timestamp>-slug.md`; the **one type committed+pushed on creation** (disaster-recovery). |

The trio `product` + `roadmap` + `project` together model a small company-or-team's structure:

- **product** = what is being built (durable, static)
- **roadmap** = what we want true by month T (forward-looking, monthly)
- **project** = what we're working on right now to advance it (operator-POV, time-bounded by an MVP)

The pair `pensive` + `prose` together cover thinking-out-loud at two different granularities:

- **pensive** = a *session* — one focused topic per file, write-then-publish, hundreds-to-thousands of words. Filename encodes the topic (`<date>-NN-pensive-<topic>.md`); the artifact stands alone.
- **prose** = a *ledger* — many fragments per file, append-as-you-go, sentence-to-paragraph each. Filename is just `prose.md`; the file is bound to a parent artifact (typically a long-running `product` like a book or blog), and fragments graduate into the parent's drafts.

Heuristic: *session or ledger?* A fragment that grows past ~3 paragraphs and develops a thesis has become a session — graduate it from prose to pensive. A pensive dominated by one-line observations rather than connected argument should have been prose entries.

`continuation` is the human-understanding cousin of a native session **`resume`**: `resume` restores *machine state* (the agent's own transcript + session id, byte-faithful); a continuation restores *human understanding* (next action, the thread's arc + the user's intention, open questions, decisions + dead ends, lessons), distilled from the *rendered* session. You earn its terseness by **flushing** loose understanding into `pensive` first, then narrating over the artifacts rather than restating them. Use a continuation to park or hand off work across time, machines, people, or agent stacks; use `resume` to reattach to a still-live session. It's also the one datatype committed + pushed on creation — a deliberate exception to the dispatcher's "never auto-commit" default, because an unpushed recovery doc is useless.

## Rules of thumb

- **Frontmatter is for queryable, stable, externally-referenced fields.** Free-form prose stays in the body.
- **Directory names carry categorical meaning.** `memory/life/family-travel/` says different things from `memory/work/travel/`. The dispatcher respects existing conventions; it doesn't invent a parallel structure.
- **One prototype per type, one type per file.** Filename and `type:` field must agree.
- **Local override shadows shared completely.** No field-level merging.

## Pointers

- Skill (generated, per-repo + gitignored): `construct/generated/datatype/SKILL.md` — **codegen; do not hand-edit**
- Skill prose source: `cmd/datatype/SKILL.md.tmpl` (the generator's `go:embed` template — **edit here**)
- Generator: `cmd/datatype/` (the `datatype` binary; `weave compile` execs the marker to materialize `SKILL.md` with the live datatype-noun list; also serves `datatype list`/`datatype show <name>` at apply time)
- Dynamic-skill marker (tracked): `construct/local/datatype/.dynamic-skill` (weave execs it at compile)
- Determinism guard: `make weave-drift-check` (asserts the render is byte-deterministic across runs; the #111 committed-file drift guard is retired)
- Mechanism: [weave atlas → Dynamic skills](weave.md) (#111, reshaped by #115)
- Shared libraries: `pkg/layergraph` (the `construct/deps` layer-graph walk, shared by weave + datatype) + `pkg/frontmatter` (flat-YAML `description:` parser)
- Prototypes: per-layer-owned `construct/datatype/`, DAG-merged by the `datatype` binary across the layer graph + the leaf's project-local `datatype/` (no `symlink construct/datatype` manifest row — retired in #115)
- Issue: `workshop/issues/000012-typed-markdown-documents-via-construct.md`
