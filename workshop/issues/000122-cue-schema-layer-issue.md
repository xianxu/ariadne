---
id: 000122
status: working
deps: []
github_issue:
created: 2026-06-24
updated: 2026-06-24
estimate_hours:
started: 2026-06-24T13:13:10-07:00
---

# Formal schema layer: nouns + lifecycle as a contract-bearing model, compiled to consumers

## Problem

The system's **nouns** (data schema) and **verbs** (lifecycle / state machine) are
defined implicitly and duplicated across repos and languages. The `issue` noun is
the worst case and the oldest: its model lives as prose in the base `AGENTS.md`
status enumeration, as *scattered string literals* in sdlc Go (`isTerminalStatus`
hardcodes `done|wontfix|punt`; `claim.go` compares `prev != "open"`; `startplan.go`
compares `== "working"` — there is **no `Issue` struct and no central status enum**),
and as a status cycle in parley.nvim's Lua. The state machine is real but smeared
across those comparisons; nothing is the source, so drift is structural.

More broadly: "architecture in the agentic age" wants the nouns + verbs formalized
**once**, as an authoritative source compiled to every consumer (LLM prose, the
deterministic shell, application code, tests), so the LLM generates *from* one source
of truth instead of re-deriving the model and duplicating it.

Design captured in
`workshop/pensive/2026-06-24-01-pensive-cue-schema-layer-nouns-verbs.md`.

## Spec

Stand up a **formal schema layer**: a single authoritative source for a noun's
data shape + lifecycle + laws, compiled to its consumers. Prove it on `issue`.

- **Source language: CUE.** Formal, intuitive, not XML, constraint-capable. Author +
  `cue vet` (build fails on an inconsistent model). The model is a **contract-bearing
  statechart** (Design by Contract over a statechart):
  - **noun** — `#Issue` field schema + `#Status` enum.
  - **lifecycle** — the transition table modeled *in CUE as data* (`from → to` on an
    `event`, with named `guards`). Not a separate markdown table.
  - **guards** — *named* preconditions. Constraint-over-own-fields guards compile from
    the schema; effectful guards (`active-time > 0`, `atlas-updated`) stay as code but
    are still **named** in the model and covered by a **case**.
  - **laws** — named universally-quantified assertions the graph shape doesn't already
    guarantee; **cases** are named examples that witness them.
- **Distribution: author/validate in CUE → export JSON (the lingua franca) → render
  markdown for the prose face.** Rides weave's existing compile-time DAG-merge; no new
  propagation mechanism.
  - `cue export → generated/issue.json` consumed by Go (`go:embed`) and (follow-up)
    parley Lua (`vim.json.decode`).
  - a renderer (`.dynamic-skill`, skill-binary pattern) emits the datatype prose face
    + lifecycle table *from* the model.
- **Two faces, one source.** The human↔LLM **interchange vocabulary** is a *generated
  projection* — fields/transitions tagged `@vocab(public)`, filtered out by a small
  projector into a `vocabulary/` surface. The full schema lives in `schema/`. Keep the
  public surface small enough to hold in one's head.
- **Rewire sdlc to read the model.** Replace `isTerminalStatus` + the scattered
  `"open"`/`"working"` literals with reads of the generated JSON. sdlc is already
  regex/read-time, so this is a small change, not a migration — lean **read-time
  (embed exported JSON)** over struct codegen.
- **Change propagation (the real test — evolvability).** Adding a value (e.g. a
  `parked` status) must be clean and *closed-loop*, not tracked via a maintained
  registry. (a) Define `#Status` as the **union of its categories**
  (`#Status: "open" | #Active | #Terminal`) so a value can't be added without being
  categorized; consumers branch on **categories** (`∈ #Active` / `∈ #Terminal`), never
  raw values, so most changes propagate for free. (b) Each consumer ships a **generated
  conformance test** in its own framework that fails when it doesn't cover the schema's
  full domain — a fail-closed registry (a *check*, not a maintained *claim*); the
  complete lineage is the import graph, derived not maintained. (c) **Completeness
  laws** gate `cue vet`: no orphan states, every value has a non-empty `when`.
- **Consumer model.** Generated artifacts are for consumers that can't read CUE
  (Go/Lua → JSON) or aren't the LLM (humans → markdown). The **LLM reads
  `schema/issue.cue` directly** — legible, full model, never stale because it re-reads
  live; only a minimal eager breadcrumb (value names + categories + a "read the source
  before touching the lifecycle" instruction) is generated for routing.
- **Staleness gate.** Conformance catches "consumer doesn't handle the domain"; a
  separate gate catches "materialized artifact ≠ current schema": regenerate the cheap
  export on build (freshness by construction); for full weave + cross-repo, stamp the
  artifact with a hash of the **merged** source and a `weave check` (recompute +
  compare), run in CI.
- **Cross-language lineage — coarse across the boundary, fine within it.** The native
  toolchain gives fine-grained per language; across the boundary one coarse repo→schema
  edge (repo-level, not subsystem) on the layer graph suffices, because each repo's
  conformance test re-fines locally. No universal build system.

Out of scope (follow-ups): parley.nvim consuming the JSON; migrating `datatype`
prototypes to this layer (likely low-yield — most are prose with little lifecycle, so
their "model" is just the sectional markdown structure; `datatype` is a *consumer* of
the schema layer, not the substrate).

Detailed design (folder placement, CUE shape, weave wiring) to be authored at
`start-plan` via `superpowers-writing-plans` into `workshop/plans/`.

## Done when

- `schema/issue.cue` declares the `#Issue` noun, `#Status` **defined as the union of
  its categories** (`#Active` / `#Terminal` / `open`), the lifecycle transition table,
  and ≥2 laws; `cue vet` passes on a good model and *fails* on a deliberately broken one
  (orphan state, undocumented value).
- The compile pipeline runs under weave: `cue export` → `generated/issue.json` (code
  consumers) and a renderer → human markdown; both regenerate on compile.
- sdlc and every code consumer branch on **categories read from the model**, not raw
  status literals — `isTerminalStatus` and the scattered `"open"`/`"working"` checks are
  gone; sdlc tests pass and a model-forbidden transition is rejected.
- Each consumer ships a **generated conformance test** that goes red if it doesn't cover
  the schema's full domain.
- A **freshness gate** exists: the cheap export regenerates on build; a `weave check`
  stamp (hash over the *merged* source) flags a stale materialized artifact
  (CI-enforceable).
- The **LLM reads the source directly** — no LLM-facing generated artifact beyond a
  minimal eager breadcrumb (names + categories + the touch-time read instruction).
- Duplicated definitions are **deleted, not paralleled** (dependency inverted: consumers
  derive from the schema).
- **Acceptance scenario:** adding a `parked` status touches only schema (its category
  membership, its `working↔parked` transitions, its `when`) plus genuine per-value
  decisions; categories propagate to Go/Lua with no hand-edits; any consumer that
  branched on a raw value is caught by a failing conformance test.

## Plan

- [ ] Design via `superpowers-writing-plans` at `start-plan` (folder convention `schema/` + `vocabulary/`, CUE model shape, weave wiring, build organization across languages)
- [ ] `schema/issue.cue` — `#Issue` noun + `#Status` as union of categories + lifecycle transition table + laws; `cue vet` gate (orphan-state, documented-value)
- [ ] Compile pipeline: `cue export → generated/issue.json`; renderer for the human markdown face + the eager vocabulary breadcrumb; wire into weave
- [ ] Rewire sdlc + code consumers to branch on categories read from the model; delete `isTerminalStatus` + scattered literals; tests
- [ ] Generated conformance test per consumer; freshness gate (build-time regen + `weave check` stamp over merged source)
- [ ] Verify the seam with the `parked` acceptance scenario end-to-end; record the verdict for "worth generalizing"

## Log

### 2026-06-24

- Created from a brainstorm on formalizing nouns + verbs as a single compiled source
  (`ARCH-DRY`: one definition, many consumers; `ARCH-PURE`: schema is pure data,
  guards are the IO/logic). Design captured in the companion pensive.
- Confirmed the duplication empirically: no `Issue` struct / no central status enum in
  sdlc — only `cmd/sdlc/internal/issue` regex accessors + scattered literals
  (`isTerminalStatus`, `claim.go`, `startplan.go`) + base-prose enum + parley Lua.
- Auto-sync push to remote failed (sandbox network block); issue file is on disk.
- Design extended over further rounds (captured in the pensive): **categories-as-union**
  so values can't be added uncategorized; **conformance tests as a fail-closed registry**
  + import graph as derived lineage; **completeness laws** (orphan-state, documented-value)
  gating `cue vet`; **freshness gate** (build-time regen + `weave check` stamp over the
  *merged* source, CI-enforced); **cross-language lineage** coarse (repo→schema on the
  layer graph, repo-level granularity) + fine within each language; the **LLM reads the
  CUE source directly** rather than a generated artifact (re-reads live → never stale),
  with only a minimal eager breadcrumb generated. The `parked`-status add is the
  acceptance scenario.
