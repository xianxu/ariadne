---
id: 000016
status: working
deps: [000012, 000015]
created: 2026-04-29
updated: 2026-04-29
references: [/Users/xianxu/workspace/brain/memory/life/42shots/ideas/2026-04-28-02-pensive-ariadne-arc.md, workshop/issues/000015-product-and-roadmap-data-types.md]
---

# project datatype

Add a `project` typed-document prototype to ariadne's datatype system. A project is the *execution container* — "what we've decided to do for a purpose, with an MVP scope, sequenced and tracked day-to-day." Distinct from `product` (durable charter) and `roadmap` (month-level aggregate). Operator-POV, time-bounded, cuts across multiple products and repos.

This is the missing piece between issues (units of work) and roadmaps (month-level targets) — and it's where the velocity calibration loop closes (each completed entry records `actual_hours`).

## Done when

- `construct/datatype/project.md` exists, conforming to the meta-prototype contract.
- Prototype ships with `Search recipes` and authoring instructions sufficient for a fresh agent to author good instances unaided.
- Dogfooded by authoring `data/project/charon-release-push.md` in brain — the immediate active project covering charon issues #13, #14, #15 (with #16 explicitly out of scope).
- `atlas/data-artifacts.md` updated with the project type.
- Decision recorded on whether/how the existing `roadmap` prototype needs adjustment to reference projects rather than product components directly. (Decision can be: "no change needed, defer to v2 of roadmap" — that's a valid outcome.)

## Spec

### Why this is needed

The product/roadmap pair from #000015 covers two of the three concerns: *what is built* (product) and *what we want true by month T* (roadmap). It misses *what we're working on right now and in what order* — the day-to-day execution view.

This isn't an issue tracker. Issues live in product repos and describe units of work that exist regardless of when they're worked on. A project is the operator's view: which units are *currently in flight*, *in what sequence*, with *what blockers*, *toward what defined MVP*.

The Facebook precedent: tasks (issues) shipped day 1; project management came years later. The two are genuinely separate concerns. Project management is downstream of issues and operator-centric; brain (the operator's substrate) is the right home.

### Project datatype — design intent

A project is a *flexible container of execution toward external value*, with a defined MVP. It declares:

- **What it's accomplishing** — a `goal` and `done_when` criterion (the MVP boundary).
- **Which issues are in scope** — referenced by `[<repo>#<id>]` syntax.
- **Which are explicitly out** — the conversation about what *isn't* included is more useful than a vague in-scope list.
- **Sequence + state** — a single ordered task list, top-down execution, with optional per-task details for state and notes.

### Frontmatter

```yaml
---
type: project
name: <slug>                  # matches filename
goal: "..."                   # one-line, why this project exists
done_when: "..."              # the MVP boundary as a falsifiable criterion
status: active | paused | done | dropped
operator: <persona-name>      # exactly one operator per project — see discipline note below
mvp_scope: [<repo>#<id>, ...] # in-MVP issue references
explicitly_out: [<repo>#<id>, ...]
created: YYYY-MM-DD
updated: YYYY-MM-DD
sources: [...]                # lineage
---
```

### Body shape — single ordered task list

Body skeleton:

1. `# <name>` — title.
2. **Lede paragraph** — one short paragraph; explicitly call out the headline omission (what's NOT in MVP) since that's the discipline.
3. `## tasks` — a single ordered list, top-down execution.
4. `## details` — optional. Per-task structured info + prose, with one block per task that has state worth recording.
5. *(Reference definitions at end of file — see "Jump-link convention" below.)*

#### Task line format

Keep one line per task short. **Title + ref only.** No inline est, no inline status, no inline blocking reason — those live in `## details`.

```markdown
- [ ] provider interface skeleton [charon#13 M1]
- [.] OpenAI provider impl [charon#13 M2]
- [ ] Anthropic mirror [charon#13 M3]
- [x] initial provider design sketch [charon#13 sketch]
- [ ] write release notes
```

#### Checkbox conventions

- `[ ]` — open / not started
- `[x]` — done
- `[.]` — blocked (reason in detail block)
- `[-]` — cancelled / removed from scope mid-project

#### Reference syntax

- `[<repo>#<id>]` — issue in any repo with an issue tracker. Product repos (`charon`, `ariadne`, etc.) and shared brain repos (`brain-team`, `brain-family`, etc.) work uniformly.
- `[<repo>#<id> M<N>]` — milestone-level granularity within an issue.
- Plain text — for items that don't fit any issue tracker (e.g., `write release notes`).

### Per-task details (`## details`, optional)

A task earns a detail block when it has state worth recording: estimate, started/closed dates, actual hours, blocking reason, prose notes. Open-not-started tasks with no notes need no detail block.

Detail block format:

```markdown
<a id="charon-13-m2"></a>
### charon#13 M2 — OpenAI provider impl

**est:** 10–16h
**status:** blocked — need OpenAI Admin API access verified before mint testing
**started:** 2026-04-30

(free prose — design notes, partial progress, decisions made during work)
```

When closed:

```markdown
<a id="charon-13-sketch"></a>
### charon#13 sketch — initial provider design

**est:** ~2h
**actual:** 1h
**closed:** 2026-04-29

Reused keychain ACL pattern from M4. Anthropic mirror should be straightforward as a result.
```

Convention:
- **Heading** = `<ref> — <title>`. Repeating the ref makes `rg <ref>` find both the task line and the detail.
- **Bold-labeled fields** for the structured bits: `**est:**`, `**actual:**`, `**status:**`, `**started:**`, `**closed:**`. Free prose follows.
- **Explicit `<a id>` anchor** above the heading — see jump-link convention.

### Jump-link convention (in-file navigation)

Goal: clicking a `[<ref>]` in the task list jumps to the corresponding detail block. Solution: markdown shortcut-reference links + explicit `<a id>` anchors.

**Slug rule (deterministic, renderer-independent):** lowercase the ref, replace each `#` and whitespace character with `-`. Examples:
- `charon#13 M2` → `charon-13-m2`
- `charon#13 sketch` → `charon-13-sketch`
- `brain-team#40 doc-cleanup` → `brain-team-40-doc-cleanup`

**Three pieces, in order in the file:**

1. Task list line (unchanged literal): `- [ ] OpenAI provider impl [charon#13 M2]`
2. Detail block:
   ```
   <a id="charon-13-m2"></a>
   ### charon#13 M2 — OpenAI provider impl
   ...
   ```
3. Reference definition at end of file:
   ```
   [charon#13 M2]: #charon-13-m2
   ```

**Behavior at render time:**
- Task with a matching reference definition → `[charon#13 M2]` becomes a clickable link, jumps to the `<a id>` anchor.
- Task without a reference definition → `[charon#13 M2]` renders as plain bracketed text. Nothing breaks.

**Why this combo:**
- Task line never changes syntax — same `[<ref>]` literal as everywhere else in the system.
- Slug derives from the ref alone, not the heading text — renaming the title doesn't break the link.
- `<a id>` is universal markdown — works in GitHub, parley.nvim, Obsidian, plain CommonMark, regardless of slugify quirks.
- `rg <ref>` still works as a plain-text fallback.

**Maintenance discipline (for the dispatcher):** when adding a detail block, also add the reference definition at the bottom; when removing a detail block, remove the matching reference definition. Authoring instructions should make this explicit.

### Single-operator discipline

Schema carries `operator: <persona-name>` in frontmatter. **Exactly one operator per project.** This is the discipline that makes AI-centric flow viable — no synchronization needed, no blocked-on-each-other patterns, throughput is preserved.

Multiple operators on one project = a smell. The dispatcher's authoring instructions should flag if the user describes "we'll split this between A and B." Suggest splitting into two projects with a clear interface.

Today (solo founder): `operator` is optional, defaults to self. Becomes load-bearing when 2+ people share a brain.

### Default instance path

`data/project/<slug>.md`.

### KTLO handling

KTLO (Keep The Lights On — bug fixes, oncall response, ad-hoc small work) is modeled as an **issue-priority flag**, not a separate datatype. Issues carrying `priority: ktlo` (or similar) are picked off when the active project has slack. This avoids spawning a perpetual "ktlo" pseudo-project that never closes — and keeps the project datatype clean ("a project always has a definable MVP and an end").

The exact frontmatter spelling for the KTLO flag is out-of-scope for this issue (lives in the issue conventions, not the datatype system). Just record the decision so future authoring uses it.

### Velocity calibration loop

A project is where the velocity skill's calibration discipline manifests:

- A detail block tracks `**est:** ~Nh` (mirroring the issue's `estimate_hours` frontmatter) when an estimate exists.
- On task completion, the checkbox flips to `[x]`, and the detail block gains `**actual:** Mh` and `**closed:** YYYY-MM-DD`.
- The `actual_hours` value is also written back to the corresponding issue file's frontmatter (e.g., `charon/workshop/issues/000013-...md`) and appended to `brain/data/life/42shots/velocity/estimate-logic-v1.md`'s validation table.
- Closes the loop: estimate → execute → record actuals → recalibrate.

The project datatype's authoring instructions should explicitly call out this discipline so it's not left to memory.

### Authoring flow

When the dispatcher applies the project prototype:

1. Distill from conversation: name, goal, candidate issues to include.
2. Force the MVP conversation upfront: "what's NOT in this project? what's the smallest version that delivers external value?"
3. Resolve `mvp_scope` and `explicitly_out` as lists of `[<repo>#<id>]` refs — confirm each issue exists.
4. Confirm `operator` (default = self for solo founder).
5. Build the initial `## tasks` list, top-down by execution order. Detail blocks for tasks already in flight; bare task lines for the rest.
6. Add reference definitions at the file bottom for any task with a detail block.
7. Confirm + write.

### Roadmap implications

The current `roadmap` prototype (from #000015) references product components directly. New framing suggests roadmap should aggregate *projects + KTLO* per month, not components.

**Decision: defer the roadmap rework to discovery during dogfood.** Land project first; author a real roadmap that references projects; if it doesn't fit, restructure roadmap as a v2. Don't pre-commit to a roadmap rework based on theory.

### Why now

The immediate trigger: tomorrow's work session needs a project to organize the charon-release-push (issues #13/#14/#15 in charon, with #16 explicitly out for scope discipline). Without a project datatype, that work either uses ad-hoc markdown or fragments across files. With one, the velocity calibration loop closes immediately on the first issue completed.

## Plan

- [x] Design pre-resolved in the Spec above; brainstorming step folded into the conversation that produced this issue. Remaining gaps surfaced during writing, not before.
- [x] Decided default instance path: `data/project/<slug>.md`.
- [x] Wrote `construct/datatype/project.md`.
- [x] Dogfooded by authoring `brain/data/project/charon-release-push.md`. Real project covering charon #13 + #14 + #15 (#16 explicitly out). Ready for tomorrow's work session.
- [x] Updated `atlas/data-artifacts.md` to include `product`, `roadmap`, and `project` (the latter two were missing from the table previously).
- [x] Decided on roadmap rework: **defer.** See Log entry 2026-04-29 below for rationale.
- [ ] **Open** — verify the velocity calibration loop in practice: as charon-release-push tasks complete, confirm `actual_hours` flows to the corresponding issue's frontmatter and to `estimate-logic-v1.md`'s validation table. Requires actual execution; can't verify in this session.

## Log

### 2026-04-29

- Issue created from a design conversation reframing product vs project. Original issue 000015 collapsed product+project+infra into a single "product" datatype with no discriminator; this reframe restores `project` as a distinct execution-container datatype while keeping `product` as the durable charter.
- KTLO modeled as an issue-priority flag rather than a separate long-running pseudo-project. Cleaner.
- Trigger: needing to organize the charon-release-push (issues #13/#14/#15) for tomorrow's work session, with explicit MVP scope (charon #16 deferred).

- **Body shape iterated** during the same conversation:
  - First sketch had kanban columns (`## doing` / `## next` / `## blocked` / `## done`) with rich inline metadata per task.
  - User pushed back: simpler is better. Single ordered task list, top-down execution, checkbox states for status (`[ ]` / `[x]` / `[.]` blocked / `[-]` cancelled).
  - Task line = title + ref only. One line, scan-friendly.
  - Per-task details (est, status, blocking reason, prose) live in optional `## details` blocks.
  - Refs cover product repos and shared brain repos uniformly: `[<repo>#<id>]`, `[<repo>#<id> M<N>]`, or plain text for non-tracked items.
- **Single-operator field added** (`operator: <persona>` in frontmatter). Discipline: exactly one operator per project, makes AI-centric flow viable, multiple-operator-per-project is a smell to flag.
- **Jump-link convention added.** Markdown shortcut-reference links + explicit `<a id>` anchors. Task line `[<ref>]` becomes a clickable jump to the corresponding detail block when a reference definition exists at file bottom. Slug rule: lowercase ref, replace `#` and whitespace with `-` (e.g. `charon#13 M2` → `charon-13-m2`). Robust across renderers; falls back to `rg <ref>` if the editor doesn't render anchors.

- **Files landed:**
  - `construct/datatype/project.md` — prototype, conforming to meta-prototype contract.
  - `atlas/data-artifacts.md` — added rows for `product`, `roadmap`, `project`. Plus a one-paragraph note framing the trio as a small-team-or-company structural model.
  - `brain/data/project/charon-release-push.md` — first dogfood instance. Covers charon #13 + #14 + #15 (#16 explicitly out). 27 tasks ordered top-down; 11 detail blocks pre-populated with estimates from the velocity skill's earlier estimates on those issues.

- **Roadmap rework decision: defer.** Rationale:
  - The current `roadmap` prototype (per-product, per-month, references components) hasn't been authored as a real instance yet. Restructuring without dogfood feels premature.
  - Roadmap and project operate at different levels: roadmap says "by end of month T, component X should be at state Y"; project says "this push completes when criterion Z is met." These compose without conflict — a project's tasks contribute to a roadmap-month's component targets.
  - Instead of pre-rewriting roadmap, author one for the upcoming month against `charon-release-push`, see whether referencing the project (rather than components directly) feels right, and decide then.
  - If a rework lands later, file as a separate issue with the dogfood evidence.

- **Status:** issue stays `working` (not `done`) until the velocity calibration loop is verified end-to-end via at least one closed task in `charon-release-push`. Schema and dogfood are in place; the *behavioral* test happens during execution.
