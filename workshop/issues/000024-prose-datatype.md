---
id: 000024
status: working
deps: []
created: 2026-05-11
updated: 2026-05-11
estimate_hours: 2
---

# prose datatype (per-product scratchpad for half-thoughts before they have a home)

## Done when

- A new `prose` typed-data prototype lives in `construct/datatype/prose.md`, defining frontmatter + body skeleton + authoring instructions.
- The `xx-datatype` skill recognizes `prose` as a valid type, can author a new `prose.md` next to a parent artifact, and can append new dated entries to an existing one.
- A `capture-prose` skill (or extension to existing capture-flow skills) handles trigger phrases — *"capture this prose for X"*, *"note this for X"*, *"save this thought for X"* — including product-name disambiguation when X matches multiple products.
- Atlas entry (under `data-artifacts.md` or similar) documents the prose convention and the prose-vs-pensive distinction.
- A short cross-reference in `construct/datatype/product.md` mentions prose as a recognized sibling artifact (one paragraph in the entity-nested-folder section).
- At least one brain has a working `prose.md` under the new convention (already in flight: `brain/data/life/42shots/book-4/prose.md` — predates this issue, will be conformed to the spec once landed).

## Spec

### Motivation

When a single-person endeavor — a book, a long-running blog, an essay — accumulates over weeks or months, half-thoughts arrive constantly: a sentence the author wants to keep but doesn't yet know where to put, a framing analogy that just clicked, a voice fragment, a comp-title observation. Current datatypes don't fit:

- **`pensive`** captures a *thinking session* (one topic, hundreds-to-thousands of words, sit-down posture). Wrong granularity for "I just thought of a line."
- **`product` body prose** is for durable charter description, not capture of fragments.
- **Ad-hoc files** lose the lifecycle and create proliferation.

Prose is the missing primitive: a *journal* of fragments tied to one parent artifact, append-as-you-go, lifecycle ends when fragments graduate into the parent (a chapter, a post, a spec).

### Distinction from `pensive`

| | Pensive | Prose |
|---|---|---|
| Unit | Session (one topic per file) | Fragment (many topics per file) |
| Shape | Write-then-publish | Append-as-you-go |
| Length | Hundreds-to-thousands of words | Sentence-to-paragraph |
| Filename | `<date>-NN-pensive-<topic>.md` | `prose.md` (sibling to parent) |
| Location | `docs/vision/` (visionary; may be cross-cutting) | Sibling to a specific parent artifact |
| Lifecycle | Stays as durable thinking record | Entries graduate into parent; original stays as history |

Heuristic: *is this a session or a ledger?* If a prose entry grows past ~3 paragraphs and develops a thesis, graduate it to a pensive.

### Frontmatter shape

| Field | Required | Notes |
|---|---|---|
| `type` | yes | `prose` |
| `parent` | yes | Relative path to the parent artifact (typically a `product` file, e.g. `book-4.md` for sibling). Resolves capture-skill's "for X" reference. |
| `created` | yes | ISO date. |
| `updated` | yes | ISO date of the last entry. |

Optional (used rarely):

| Field | Notes |
|---|---|
| `target_repo` | If graduated entries get pushed to a peer repo (e.g. blog product publishes to `../xianxu.dev/`), record the path. Inherited from parent product when unset. |

### Body skeleton

1. `# <parent-slug> prose scratchpad` — title.
2. One paragraph stating what this file is, format conventions, lifecycle. (Boilerplate; can be lifted from book-4 prose.md.)
3. A `---` separator.
4. Reverse-chronological entries, each:
   - `## YYYY-MM-DD HH:MM` header (timestamp at minute precision; multiple entries per day get distinct timestamps).
   - Optional `**topic:**`, `**tag:**`, `**candidate home:**` metadata lines.
   - Free prose body — fragment.
   - Optional `*[framing — agent annotation]*` sub-block if the agent added an interpretive note alongside the user's fragment.

Newest entries on top so the file opens to the most recent thought.

### Default location

Sibling to the parent artifact:

- Parent at `data/life/42shots/book-4/book-4.md` → prose at `data/life/42shots/book-4/prose.md`
- Parent at `data/life/personal/xianxu-dev/xianxu-dev.md` → prose at `data/life/personal/xianxu-dev/prose.md`

For products that currently live at the flat default `data/product/<slug>.md`: the product *graduates to a folder* when it gains a prose sibling. `data/product/<slug>.md` becomes `data/product/<slug>/<slug>.md` + `data/product/<slug>/prose.md`. Update product datatype's §5 to note this.

If a product has multiple distinct prose streams (rare), use `prose-<subtopic>.md`. Don't design this in upfront.

### Lifecycle

1. **Capture.** User says "capture this prose for X" → entry appended to top of `prose.md`.
2. **Mature.** Entry sits in prose.md. May get tagged, annotated, or referenced by later entries.
3. **Graduate.** When the fragment finds a home (chapter draft, blog post, spec), copy it into the parent artifact's draft. Leave the original in prose.md as history — don't delete.
4. **History.** Old entries remain in prose.md as the trace of how thinking arrived where it did.

### `capture-prose` skill behavior

**Trigger phrases:** *capture this prose for X*, *note this for X*, *save this thought for X*, *jot this for X*.

**Resolution flow:**

1. **Find candidate products.** `rg -l "^type: product" data/ | xargs rg -l "^name: <X>$"` and fuzzy-match on `name:` field.
2. **Unique match** → append entry to that product's `prose.md` (creating the file if absent, with parent frontmatter pointing back to the product).
3. **Multiple matches** (e.g., user says "blog" and both `xianxu-dev` and `42shots-blog` exist) → ask user: present 2-3 options as a single-select question. Don't proceed without disambiguation.
4. **No match** → ask if user wants to create a new product first, or capture as a freestanding pensive instead.
5. **Implicit "this" / "current"** → use the product whose file is currently open / most recently edited in the session; otherwise ask.

**Entry construction:**

- Timestamp: current local time, minute precision.
- Topic: extract or infer one short noun phrase from the fragment.
- Tags: suggest 2-4 tags from content keywords; prefer parent product's existing tag vocabulary (scan prior entries in same prose.md); user can correct.
- Candidate home: if the agent sees a clear chapter/component fit in the parent product, suggest one; otherwise leave blank.
- Agent annotation: if the agent has analytical context to add (e.g., why this fragment is load-bearing for a particular chapter), append a `*[framing — agent annotation]*` sub-block. Optional.

**Don't ceremoniously confirm.** Like pensive, prose is low-friction capture. Confirm only the destination path once if ambiguous, then write.

### Cross-repo product model — context

This datatype assumes the brain's **Pattern B** layout for products: each product's canonical `product.md` and `prose.md` both live in `brain/data/`, even when the product publishes to a peer repo. The peer repo holds *execution and publication only* (built site, theme, deployed posts). Pattern B keeps half-thoughts in private encrypted brain rather than leaking into public peer repos.

When the product's `target_repo:` is set, the graduate-to-draft step (step 3 in Lifecycle) is a move to that peer repo. The `capture-prose` skill itself doesn't need to know about target_repo — that's the graduate-skill's concern (out of scope for this issue).

## Plan

### M1 — datatype spec lands in ariadne

- [x] Write `construct/datatype/prose.md` matching this issue's spec.
- [x] ~~Add prose to `xx-datatype` skill's recognized-types list.~~ **Moot — dispatcher is data-driven.** Per `construct/datatype/type.md` and the `xx-datatype` skill, new types are pure data: the dispatcher discovers prototypes by scanning `construct/datatype/`. Adding the prototype IS the recognition. No skill change needed.
- [x] Add one-paragraph cross-reference in `construct/datatype/product.md` §5 noting prose as a recognized sibling artifact + the file-vs-folder graduation rule for default-location products.
- [x] Update atlas (`data-artifacts.md`) with prose row in built-in types table.

### M2 — capture-prose skill

- [ ] Decide: new skill `construct/skill/capture-prose/SKILL.md`, or extend existing capture flow in `xx-datatype`? Lean toward new skill — trigger phrases are distinct and the resolution flow is non-trivial.
- [ ] Implement resolution flow (product lookup, disambiguation, append, file creation if absent).
- [ ] Implement entry construction (timestamp, topic inference, tag suggestion).
- [ ] Test against brain's book-4/prose.md.

### M3 — propagate to brain

- [ ] Vendor latest ariadne to brain.
- [ ] Conform existing `brain/data/life/42shots/book-4/prose.md` to the new spec (likely no changes — drafted to anticipate the spec).
- [ ] Add `parent:` frontmatter field to existing prose.md.
- [ ] Verify capture-prose flow works against book-4.

### M4 — atlas + lessons

- [ ] Atlas entry documenting prose convention and prose-vs-pensive distinction.
- [ ] Brief lesson entry: "when in doubt between pensive and prose, ask 'session or ledger?'" (if it earned a place in lessons.md).

## Log

**2026-05-11 — M1 implemented.**

- Wrote `construct/datatype/prose.md` (the prototype). Follows the meta-prototype shape in `type.md`: frontmatter, body skeleton (6 sections — title, intro, format note, lifecycle note, separator, reverse-chrono entries), authoring instructions (parent resolution + disambiguation, sibling location with file→folder graduation rule, entry composition, append-at-top, frontmatter `updated:` bump, don't-commit), search recipes, and rules. Concrete example points readers to `brain/data/life/42shots/book-4/prose.md` which predates this prototype and was the design vehicle.
- **Discovered the skill-change bullet was moot** — `xx-datatype` is data-driven; it auto-discovers prototypes in `construct/datatype/`. Marked the bullet struck-through rather than removed so the reasoning survives in the issue's commit log. The `capture-prose` skill (M2) is a separate concern — that one *will* be a new skill in `construct/local/`.
- Cross-reference paragraph added to `construct/datatype/product.md` §5. Calls out prose as a recognized sibling AND specifies the **file→folder graduation rule** for products at the default flat location (`data/product/<slug>.md` becomes `data/product/<slug>/<slug>.md` plus `data/product/<slug>/prose.md` when prose first appears). Graduation is one `git mv` + `rg` sweep — explicit, not silent.
- Atlas row added to `atlas/data-artifacts.md` in the built-in types table, with the prose-vs-pensive distinction in the cell ("session vs ledger").
- Files touched, not yet committed:
  - `construct/datatype/prose.md` (new)
  - `construct/datatype/product.md` (cross-ref paragraph)
  - `atlas/data-artifacts.md` (row)
  - `workshop/issues/000024-prose-datatype.md` (this log + M1 checkboxes)
- Verification on M1: the dispatcher's data-driven behavior means the prototype is "live" as soon as it lands. To exercise: have `xx-datatype` recognize a "capture this prose for X" trigger. That exercise is M2's job (the `capture-prose` skill formalizes the trigger phrasing and the parent-disambiguation flow); for now the prototype can be applied via `/xx-datatype prose <path>` directly.
