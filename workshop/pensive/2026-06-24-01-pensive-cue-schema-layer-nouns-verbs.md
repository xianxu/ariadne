---
type: pensive
date: 2026-06-24
topic: schema layer — nouns + verbs as a contract-bearing model
mode: ideas
description: What "architecture" means in the agentic age — formalize the system's nouns (data schema) and verbs (lifecycle/state machine) once, as an authoritative source compiled to every consumer (LLM prose, infra shell, app code, tests), so the LLM generates from one source instead of re-deriving and duplicating. CUE as the source language; `issue` as the first guinea pig (ariadne#122).
references: [workshop/issues/000122-cue-schema-layer-issue.md, construct/datatype/type.md, cmd/sdlc/internal/issue/structural.go, workshop/pensive/2026-06-08-01-pensive-shim-state-machines.md, workshop/pensive/2026-06-17-01-pensive-layer-graph-as-platform-primitive.md, atlas/workflow/weave.md]
---

# Pensive: Schema layer — nouns + verbs as a contract-bearing model

Trace the evolution of how we manage the `issue` noun and you can read the whole problem off it. GitHub (external model) → `make issue 42` mirroring it into the repo (Claude invents a sectional model, seeded by the external one) → parley's `<C-y>c` create mode + status-line frontmatter autocompletion (the internal model getting schematized, but split between the Makefile and parley.nvim) → the status enumeration written as prose in `AGENTS.md` with a described lifecycle → `sdlc` taking over the gates, with the issue model embedded *in Go itself*. The endpoint: the `issue` noun is defined implicitly in at least three places — base prose, sdlc's scattered Go string literals, parley's Lua status cycle — and none of them is the source. The state machine is real but smeared across string comparisons (`isTerminalStatus`, `prev != "open"`, `== "working"`); there's no `Issue` struct and no central status enum anywhere. The oldest noun is the most scattered.

This is the thing I actually want to name: **what is "architecture" in the agentic age?** My answer is *limited formalization* — declare the system's **nouns** (data schema) and **verbs** (lifecycle/state machine) once, as an authoritative source, and compile every downstream artifact from it. The payoff is specifically that when an LLM generates code, it generates *from* a single source of truth instead of re-deriving the model and duplicating it. That's the same bet the whole stack already makes — skill-binary, dynamic-skill, weave-merge — extended from *behavior* to *data + lifecycle*.

## Two goals, both pointing at "small and simple"

Formalizing buys two different things, and they need different design altitudes:

1. **A precise human↔LLM interchange vocabulary** (the foremost one). English is imprecise; if we share a small agreed-on formal vocabulary of nouns/states/verbs, then "move #42 to blocked on #51" is unambiguous in a way the English equivalent isn't. This wants *concise and simple* — small enough to hold in my head.
2. **A deterministic shell** — one source compiled downstream so consumers can't drift. This is like a `target` file but harder: a `target` is committed *documentation*; a schema is *compiled from*. This wants *completeness* — every field and constraint the code needs.

The nice result: both goals point *away* from heavyweight DSLs and toward a small, simple formal language — they don't trade off. And the apparent tension (concise vs. complete) dissolves with one move: the interchange vocabulary is a **generated projection** of the full schema — a filtered *view*, not a second artifact. One source, two faces. Tag the public surface with a CUE attribute (`@vocab(public)`); a small projector emits the spoken vocabulary from it. Keep that surface small enough to read — the moment I can't read the model, it stops being *my* design surface and becomes the agent's private encoding.

## The pattern: Design by Contract over a statechart

The vocabulary I want to fix:
- **noun** — the entity/data shape (`#Issue`).
- **state** + **transition** — the lifecycle ("verbs"). A transition is an edge `from → to` on an `event`.
- **guard** — a *named* precondition on a transition. The name lives in the model; the implementation may be compiled (a constraint over the noun's own fields) or hand-written code (effectful — `active-time > 0`, `atlas-updated`). Either way it has a name.
- **law** — a *named, universally-quantified* assertion ("for all issues in state X, P holds"). Borrowed from FP laws: inspectable, referenceable, must-hold.
- **case** — a *named example* that witnesses a law or exercises a path.

So a **model = noun + lifecycle + laws**, single-sourced and compiled. The honest lineage is Design by Contract laid over a statechart — both halves legible and battle-tested, no UML/MDE baggage.

The data/code boundary falls out of the structure itself: constraints over a noun's own fields *compile* from the schema (declare and decide collapse — a formal schema genuinely binds them); the residual is effectful predicates over external state, which stay as code. But I don't care about a closed formal system — be practical. The effectful guards still get a **name** in the model and a **case** that tests them; whether a guard is compiled or hand-written becomes an implementation detail *below the name*. The name is the contract; the impl is free. We don't need to push schema expressiveness up (constraint-solver class) yet.

## Concrete shape

**CUE** as the source language — formal, intuitive to read, not XML. Author + validate in CUE; distribute as JSON (the lingua franca everything reads); render markdown for the prose face. The transition table is modeled *in CUE as data*, not a separate markdown table — the table I read is rendered from it. The pipeline rides weave's existing compile-time merge:

```
cue vet schema/issue.cue                       # fail the build on an inconsistent model
cue export schema/issue.cue --out json > generated/issue.json   # Go (go:embed) + nvim Lua (vim.json.decode)
issue-render schema/issue.cue > generated/datatype/issue.md     # the prose face, via .dynamic-skill
```

sdlc reads the exported JSON in place of `isTerminalStatus` and the scattered literals (it's already regex/read-time, so embedding JSON is a *small* change, not a migration). The propagation problem is already solved — weave's DAG-merge, same machinery datatype rides; not a new symlink scheme.

**Who consumes what — and this is where my "make the semantics travel" instinct flipped.** The generated artifacts exist for consumers that *can't read the source*: JSON for the deterministic code (Go `go:embed`, parley Lua `vim.json.decode`), rendered markdown for *humans* skimming docs. The LLM is *not* one of them — CUE is legible, so the agent reads `schema/issue.cue` directly when it touches the domain. That's stronger: the LLM re-reads live every session, so reading source means it's never stale (vs. a generated prose face that's one more thing to regenerate), and at the moment of editing the lifecycle it gets the *full* model — categories, transitions, laws, rationale — not a flattened projection. Keep only a minimal eager breadcrumb (the value names + categories in the always-loaded surface, plus a skill instruction "read the source before touching the issue lifecycle"). Eager = route in the vocabulary; lazy = read the source to mutate it. Two load-bearing conditions: the skill must *point* the agent at the source at the touch-moment (don't hope it opens the file), and the source must stay small enough to read — the same "hold it in your head" constraint.

**Freshness is its own fail-closed gate, one level up from conformance.** Conformance tests catch "a consumer doesn't handle the schema's full domain"; they do *not* catch "the materialized artifact is no longer the current schema." So: regenerate the cheap export on every build (freshness by construction — can't go stale locally), and for the heavy full weave + cross-repo, stamp the artifact with a hash of the *merged* source (weave's DAG-merge — cross-repo is where staleness bites) and a `weave check` that recomputes and compares, run in CI. Source→artifact freshness, then artifact→consumer conformance: the same closed-loop discipline chained.

`issue` is the guinea pig precisely because it's the oldest and most scattered noun — prove the seam there before generalizing.

## Evolving it — change propagates, it isn't tracked

The real test isn't centralizing the model, it's *changing* it cleanly — say adding a `parked` status (in-progress, waiting on an external thing, distinct from `blocked`-on-another-issue). I don't want a maintained "where-used" lineage registry: that's a second source of truth that drifts, the exact failure I'm killing. I want **closed-loop** — consumers fail to compile/boot until they handle the new value, so the toolchain finds the sites for me.

Three layers, in order of leverage:

1. **Categories over values, and define the enum *as* the union of its categories.** `#Status: "open" | #Active | #Terminal`, with `#Active: "working" | "blocked" | "parked"`. Now you *cannot* add `parked` without categorizing it — "every value is categorized" is true by construction, not a checked law. Consumers branch on *categories* (`∈ #Active`, `∈ #Terminal`), not raw values, so adding `parked` is a one-line category edit that propagates for free: `isTerminalStatus`, the active-set filters, `list` — all correct with zero code change. Most "necessary places" evaporate.

2. **Conformance tests are the registry — but a *check*, not a *claim*.** Fair challenge to myself: a set of conformance tests *is* the consumer registry, same information as a maintained list. The difference that matters is form — a maintained registry *asserts* "parley consumes #Status" and can quietly go false; a conformance test *exercises* parley against the schema and goes red when reality diverges. Executable + co-located + grounded in behavior. And the *complete* "who uses this" answer is the **import/dependency graph** — derived by grep/static analysis, never maintained, finds even untested consumers. All three beat a hand-written list because they're grounded in the code, not in a claim about it.

3. **Completeness laws gate `cue vet`.** No orphan states (every state ≥1 inbound + outbound transition catches "added `parked` but no way to reach/leave it"); every value carries a non-empty `when` (forces the *semantic* decision). That last one is the honest residual: meaning can't be compiled — no mechanism decides *for* me what `parked` means. But the law makes its *omission* fail closed. The system makes gaps loud; it doesn't fabricate semantics.

**Cross-language lineage: coarse across the boundary, fine within it.** A single build system over Go + Lua + prose + CUE is too heavy. I don't need fine-grained-across-languages, though: within a language the native toolchain gives fine-grained for free; across the boundary one coarse fact — "repo X consumes schema `issue`" — suffices, because X's own conformance test re-fines locally. Repo-level granularity, not subsystem (Go-vs-Lua is recovered inside X by its toolchain; recording it at the boundary would re-import the cost I just rejected). The coarse edge rides the layer graph (`construct/deps`) — another DAG-aware projection, not a new builder. The LLM consumer doesn't belong in this graph at all: it re-reads live, so it has a different freshness model (never cached, never stale) than the code consumers that cache and need the stamp-check.

## Typing markdown — the general frame

Step back and what #122 is really doing is *typing markdown files*. A typed markdown file has two surfaces: its **frontmatter** (a record of fields) and its **sectional structure** (required/optional/ordered `## ` sections). #122 typed the first for `issue`; the second is the same kind of thing — more fields in the same record. Seen this way, an ariadne **`datatype` = an agent skill (prose) + a schema** — the skill-binary pattern (prose + deterministic code) applied to *artifacts* instead of *behavior*. The prose carries the unconstrainable guidance (when to use it, what to ask, keep the voice); the schema carries the checkable shape.

The validation is mechanical: **artifact → list of typed instances → `cue vet` each.** A type is `(locator, schema)`: the *locator* extracts a normalized instance from a region (frontmatter via YAML, sections via `## ` split, a fenced block, `- [ ]` lines), the *schema* constrains it. The under-appreciated half is the **extractor** — `cue vet` is the easy part; "where do instances live in the artifact" is the new, per-type component. One file is a *container* of instances (1 frontmatter + N sections + the estimate block + M checklist items), each vetted against its type. It's schema-on-read for human/LLM-legible documents — the markdown cousin of JSON-Schema/Schematron, novel because the substrate is prose and the authors are LLMs.

**Where the line falls — and I had it wrong at first.** The schema is for **well-formedness only**: the `age ∈ [0,100]` / `status ∈ enum` / required-field / present-section class — crisp, bounded, objective. It is *not* for **semantic quality proxies**. `Spec ≥ 50 words` looks structural but is a heuristic stand-in for "is this substantive" — a soul-check in disguise, a pre-AI-era hack to approximate judgment in deterministic code. Those stay out of the schema; in the AI age that judgment is the LLM's. So sdlc's existing `CheckStructural` word-count gates are NOT schema material — don't migrate them in; arguably retire them to LLM judgment. **The schema defends the skeleton; the LLM owns the soul — and we shouldn't *attempt* to capture the soul in deterministic code at all.**

**Fixing is the LLM's job too — there is no deterministic write-back.** The validator emits diagnostics; the LLM (or the human in parley) reads them and edits the markdown, then re-validates. Same agentic loop. The lossy-extraction worry dissolves: we never write *back through* the extractor — the deterministic part is only the *check* (the signal), the LLM is the *actuator* (the edit). The one requirement is that diagnostics be clear enough to act on.

So **instance-conformance** is the third closed-loop (alongside artifact-freshness and consumer-conformance), and it's what makes a type actually *defend* its artifacts rather than just guide tooling. One extract→vet binary becomes the deterministic shell at every edit surface — parley on save, the agent after an Edit, the pre-merge gate, CI. This also updates my earlier "datatype migration is low-yield": even prose datatypes get real value from *well-formedness* validation, because their instances are LLM-generated and *will* drift.

## Open questions

- **Resolved (this thread + later rounds):** location = `construct/vocabulary/` (propagates like datatype via the layer graph); dropped the `.md` face + `@vocab` projector — one generated face (`issue.json`) for code, the LLM reads the source; schema scope = **well-formedness only** (semantic/quality gates stay out → LLM); read-time over codegen for sdlc; categories-as-source with `or()`-derived defs; conformance-tests + import-graph over a maintained registry.
- **The extractor-as-declared-locator** is the genuinely open piece — the part that doesn't reduce to "just CUE." Generic for frontmatter/sections; bespoke for fenced blocks and `- [ ]` lines. Is there a way to *declare* a locator per type, or do bespoke ones stay imperative glue?
- Do the existing `CheckStructural` word-count gates get **retired to LLM judgment**, kept as cheap imperative pre-filters, or deleted? (They're soul-checks in deterministic code — the AI age makes them questionable.)
- The **instance-conformance validator** (extract → `cue vet`, fired at edit/pre-merge across *all* typed markdown) is a follow-up beyond #122 — motivated by the hand-edited-bad-status scenario.
- **How to organize builds across languages** stays open (coarse-on-the-layer-graph is the starting position).
- The semantic residual: is a required `when` field + a completeness law enough to keep meaning from rotting, or does it need periodic agent review?

- **Resolved this round:** read-time over codegen for sdlc (embed exported JSON); LLM reads source, not a generated artifact; categories-as-union; conformance-tests + import-graph over a maintained registry; coarse/fine cross-language graph at repo granularity.
- Folder naming and placement: `schema/` for the source, `vocabulary/` for the eager human/LLM breadcrumb? How exactly do they materialize/travel via weave — mirror datatype's `construct/generated/`?
- **How to organize builds across languages stays genuinely open.** Coarse-on-the-layer-graph is the starting position, but the broader question — one orchestrator vs. per-tool builds wired loosely — is undetermined, and ariadne's build-from-scratch lean doesn't resolve it yet.
- The semantic residual: `when`/rationale lives in source for the agent to read on touch — is a required `when` field + completeness law enough to keep meaning from rotting, or does it need periodic agent review?
- Does migrating the `datatype` prototypes to this ever pay off? Probably not soon — most are prose with little lifecycle or constraint, so their "model" is mostly the *sectional structure* of a markdown artifact, not a noun+statechart. datatype is a *consumer* of the schema layer, not the substrate.
- How much guard logic to compile vs. leave as named code — start minimal; don't reach for constraint-solver expressiveness prematurely.
