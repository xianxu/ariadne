---
id: 000015
status: blocked
deps: [000012, 000016]
created: 2026-04-29
updated: 2026-05-03
references: [/Users/xianxu/workspace/brain/memory/life/42shots/ideas/2026-04-28-02-pensive-ariadne-arc.md]
---

# product and roadmap data types

Add two typed-document prototypes to the data system: `product` (the durable charter for a thing being built, spanning 1..N peer repos of the brain) and `roadmap` (a structured monthly snapshot tracking progress against a product's components). Co-design the prototypes and the workflows around them inside the prototype files themselves — the prototype is the spec.

"Product" is the umbrella term, deliberately preferred over "project" because "project" carries too many overloaded meanings (engineering effort, IDE workspace, the ariadne `workshop/` notion, etc.). Externally-sold products, internal efforts, and infra all fit the same charter shape under this name.

## Done when

- `construct/datatype/product.md` exists, conforming to the meta-prototype's contract.
- `construct/datatype/roadmap.md` exists, conforming to the meta-prototype's contract.
- Both prototypes ship with `Search recipes` and authoring instructions sufficient for a fresh agent to author good instances unaided.
- Ariadne itself is described retrospectively using these prototypes — `data/product/ariadne.md` (or equivalent path, decided during planning) plus a first `data/roadmap/202604/ariadne.md`. This is the dogfood test.
- Charon is described as the second test, authored from outside ariadne.
- `atlas/data-artifacts.md` updated to include the two new types.

## Spec

### Motivation

The Ariadne-arc pensive (referenced) argues the substrate Ariadne sells is a typed-data system. To extend that system to model a company, we need types for the things a company organizes around. Product and roadmap are the first two — they capture the durable shape of work and its temporal status. Persona, business, and others are deferred (persona until there's a second operator; business until there's a real company to model).

### Product type — design intent

Product is a *charter*: vision, what it is in one paragraph, the durable shape of what's being built. The shape is expressed as **components** — sections in the body that decompose recursively. Time-invariant content lives here; status snapshots live in roadmap.

Key design points to nail down during brainstorming (delegated per the meta-prototype's authoring instructions):

- **Frontmatter** likely carries: `type: product`, `name`, `repos: [path...]` (1..N peer repos of the brain), `status` (active | paused | sunset), creation/update dates, and lineage fields. Owner is deferred until persona exists. No `kind` or `audience` discriminator for now — everything is a product (internal efforts and infra are products with internal customers); add a discriminator later if a real query demands it.
- **Components as sections with slug-like IDs.** Component headings use `## substrate-skill-management` (slug form), not `## Substrate (skill management)`. This is so `rg` across roadmaps stays cheap. Renaming a component is a deliberate act with a cross-repo `rg` sweep.
- **Recursive decomposition.** Components nest into subcomponents via subsections. No fixed depth.
- **What goes in body vs frontmatter.** Vision, the one-paragraph definition, and the component tree are body content. Repo paths and status are frontmatter (queryable).

### Roadmap type — design intent

Roadmap is the temporal snapshot — *where we are against a product's components at month T*. Roadmaps live in monthly snapshot directories (`data/roadmap/YYYYMM/<product>.md`) so multiple months can be authored in parallel and edited as plans shift; git history is the change log.

Key design points to nail down:

- **Frontmatter** likely carries: `type: roadmap`, `product: <name>` (links to the product artifact), `month: YYYYMM`, `status` per component or aggregate, lineage fields.
- **Body** organized around the components from the product file (referenced by slug ID). Each component section says: where we are, what changed since last month, what's targeted next, blockers.
- **Cadence is monthly; horizon is free.** A 202604 roadmap can talk about a Q3-2026 target — directory cadence is about how often we snapshot, not how far we look.
- **Linkage to product.** A roadmap entry referring to a component the product file no longer contains is a signal — either the product drifted from reality, or the roadmap is stale. Worth making the inconsistency easy to detect (`rg` against component IDs).

### Workflow processes baked into the prototypes

Per the meta-prototype philosophy, workflow lives in the prototype's *Authoring instructions* section. Concretely, both prototypes should encode:

- When to create a new instance vs. update an existing one (creating a new monthly roadmap each month; editing a product file in place when components shift).
- How the dispatcher should ask for missing information (e.g., for roadmap: "which product? which month?" — defaults to current month if unstated).
- How the product and roadmap reference each other (slug IDs, product name field).
- The cross-repo `rg` discipline on component renames.

### Lineage

The Ariadne-arc pensive proposes universal lineage frontmatter — `sources:`, `derived_sha:`, `derived_by:` — as a binding rule in `AGENTS.md` for AI-written typed artifacts. That rule is broader than this issue. **Out of scope here: amending AGENTS.md.** In scope: ensuring both new prototypes declare the lineage fields in their frontmatter shape, so when the AGENTS.md rule lands, the prototypes are already conformant. File a separate issue for the AGENTS.md change if it doesn't already exist.

### Why now, why these two first

These are the smallest typed pair that can test "model a company as data." Persona requires a second operator before it pays off. Business requires a real company to model. Product + roadmap can be exercised on Ariadne (retrospective dogfood) and Charon (forward-looking) immediately, with one operator.

## Plan

- [ ] **Brainstorm `product` prototype** via `superpowers-brainstorming`. Settle frontmatter fields (especially `repos`, `status`), the component-section convention (slug IDs, depth rules), and the authoring-instructions content.
- [ ] **Brainstorm `roadmap` prototype** via `superpowers-brainstorming`. Settle the snapshot-directory path (`data/roadmap/YYYYMM/<product>.md` vs alternatives), how component IDs link back to the product file, and what each component section in a roadmap should contain.
- [ ] Decide where instances live by default — `data/product/<name>.md` and `data/roadmap/YYYYMM/<name>.md` is the working assumption.
- [ ] Write `construct/datatype/product.md`. Self-contained per meta-prototype rules.
- [ ] Write `construct/datatype/roadmap.md`. Self-contained.
- [ ] Both prototypes include `Search recipes` for the `rg` queries each type's downstream tooling will need (find all active products, find all components of product X, find all roadmaps for product X across months, find roadmaps targeting a given quarter).
- [ ] Run `construct/scripts/sync-local-skills.sh` if needed, and verify the dispatcher (`xx-datatype`) can fuzzy-match conversational triggers like "let's roadmap Charon for May" or "set up a product file for Ariadne."
- [ ] **Dogfood test 1 — Ariadne, retrospectively.** Author `data/product/ariadne.md` describing what Ariadne is, its durable shape, and its components (Construct, base layer, sandbox, data system, ...). Then author `data/roadmap/202604/ariadne.md` (this month's snapshot) with status against each component.
- [ ] **Dogfood test 2 — Charon, forward-looking.** From outside ariadne, in Charon's repo or wherever Charon's brain lives, author a Charon product file and an initial roadmap. This validates the prototypes work in a non-self-referential setting.
- [ ] Update `atlas/data-artifacts.md` to list `product` and `roadmap` with one-line descriptions and pointers.
- [ ] If dogfood reveals schema drift, iterate on the prototypes before declaring done. Capture iterations in `## Log`.

## Log

### 2026-04-29

- Issue created from the Ariadne-arc pensive (`brain/memory/life/42shots/ideas/2026-04-28-02-pensive-ariadne-arc.md`). Pensive synthesizes the typed-data-system framing for modeling a company; this issue extracts the immediate next step (product + roadmap types) and defers persona/business.
- Renamed the umbrella type from `project` → `product`. "Project" carries too many overloaded meanings (engineering effort, IDE workspace, ariadne's `workshop/` notion). "Product" is cleaner; internal efforts and infra are conceptualized as products with internal customers.
- Open design questions logged in Spec; brainstorming step in Plan will close them before prototype files are written.

- **Product brainstorming converged.** Decisions:
  - **Body skeleton:** `# title`, lede line, `## vision`, `## components` container with flat `### <slug>` items. Each component: one-line purpose, `**State:**` line (enum + free text), free prose. No `####` nesting; sub-features live in prose.
  - **Component shape:** lightly structured (option B). `rg -A2 "^### "` reads at a glance.
  - **Cross-reference convention:** single-backtick `` `product:slug` `` for cross-product, `` `slug` `` for same-product (when context is unambiguous), `` `product` `` for product itself.
  - **Frontmatter:** `type`, `name`, `repos`, `created`, `updated`, optional `sources`. No `status` field (YAGNI — git history and roadmap recency carry the signal). No `derived_sha` or `derived_by` (model nondeterminism + cross-repo dependencies make rigorous reproducibility tracking too brittle to be worth the cost).
  - **`repos` format:** repo names only, `0..N` (empty list valid). Resolved as `<workspace-root>/<name>`.
  - **Vision authoring rule:** never fabricated by the dispatcher; placeholder + flag-to-human if unstated.
  - **State enum:** `idea | planning | in-progress | shipped | paused | dropped`, followed by `— <free text>`. Lives in product (not roadmap), since "where we are now" is part of the living charter. Updated as work progresses; git history is the trajectory.
  - **Pensive framing correction:** the "time-invariant in product, time-variant in roadmap" framing is wrong as stated. Better: product = durable shape + current state; roadmap = where we want to be at month T.
  - **Default instance path:** `data/product/<slug>.md`.

- **Roadmap brainstorming converged.** Decisions:
  - **Roadmap is a forward-looking *plan*, not a snapshot.** Targets what should be true at end of month T. Not a changelog (git diff between roadmaps is the changelog). Not a current-state report (that's in product).
  - **Per-product, period.** One roadmap = one (product, month) pair. Proto-company view is the aggregate of `data/roadmap/YYYYMM/*.md`. Cross-product dependencies via `` `other-product:slug` `` references in component prose. No cross-product roadmap datatype yet (deferred — add `month-plan` datatype later if needed).
  - **Body skeleton:** `# product — month` title, `**Target:**` lede, `## plan`, `## components`, `## postmortem`. Postmortem starts empty and is filled after the month closes.
  - **`## plan` shape:** `**Capacity:**` (free-form, dev-weeks), `**In scope:**` priority-ordered list, `**Out of scope:**` priority-ordered list (cut between the two = capacity boundary), `**Reasoning:**` paragraph.
  - **`## components` shape:** only components being touched this month appear (sparse, not full snapshot). Each has `**Target state:**` and `**Effort:**` (free-form), plus prose. Component slugs MUST exist in the corresponding product file.
  - **Multi-month gaps allowed.** If 202605 then 202607 (no 202606), the 202607 roadmap covers the 2-month horizon; capacity statement says so explicitly.
  - **Frontmatter:** `type`, `product`, `month`, optional `target_event`, `created`, `updated`, optional `sources`.
  - **Default path:** `data/roadmap/<YYYYMM>/<product>.md`.

- **Files landed:**
  - `construct/datatype/product.md` (revised to add `**State:**` element)
  - `construct/datatype/roadmap.md`
  - Both conform to meta-prototype contract.

- **Reframe (later same day):** the original "product is the umbrella covering products + projects + infra" decision was wrong. Sharper framing:
  - `product` = *durable thing being built* (what this issue defined). Static. Stays as-is.
  - `project` = *execution container* — what we've decided to do for a purpose, with an MVP scope. Operator-POV, time-bounded, cuts across products/repos. Distinct datatype, tracked in follow-up issue 000016.
  - `roadmap` = month-level aggregate of projects + KTLO (KTLO modeled as an issue-priority flag, not a separate datatype).
  - Updated `construct/datatype/product.md` lede paragraph to drop the "umbrella" claim and point at project + roadmap as siblings.
  - Roadmap prototype currently references product components directly. May need rework once project lands (roadmap → projects → product components vs current roadmap → components). Deferred to `000016`'s dogfood phase to discover.

### 2026-05-03

- **Blocked on #16.** Dogfood is happening via `brain/data/project/charon-release-push.md` (the project datatype), not via product/roadmap. Roadmap likely needs to reference projects rather than product components directly — that rework should be informed by #16's velocity-calibration evidence rather than pre-committed. Resume #15 once #16 closes and we have a real project + month of data to author a roadmap against.
