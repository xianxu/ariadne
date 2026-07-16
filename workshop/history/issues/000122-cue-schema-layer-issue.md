---
id: 000122
status: done
deps: []
github_issue:
target: issue-lifecycle
created: 2026-06-24
updated: 2026-06-25
estimate_hours: 10
started: 2026-06-24T13:13:10-07:00
actual_hours: 4.67
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

## Estimate

```estimate
model: estimate-logic-v2
familiarity: 1.0
design-buffer: 0.30
item: typed-data-prototype   design=0.4 impl=0.8
item: atlas-docs             design=0.2 impl=0.4
item: milestone-review       design=0.0 impl=0.6
item: greenfield-go-module   design=0.5 impl=1.2
item: skill-or-dispatcher    design=0.1 impl=0.5
item: smaller-go-module      design=0.2 impl=0.7
item: milestone-review       design=0.0 impl=0.6
item: smaller-go-module      design=0.3 impl=0.8
item: cross-cutting-refactor design=0.3 impl=1.2
item: milestone-review       design=0.0 impl=0.6
total: 10.0
```

Derivation: M1 (vocabulary model + lifecycle target + review) ≈ 2.6h · M2
(`cmd/vocabulary` + weave wiring + `ensure-cue`/build-order + freshness) ≈ 4.0h ·
M3 (`IssueModel` + rewire + conformance + `parked` acceptance) ≈ 3.4h. Σdesign 2.0
×1.30 + Σimpl 7.4 ×1.0 = 10.0.

## Plan

- [x] Design the durable plan via `superpowers-writing-plans` → `workshop/plans/000122-cue-schema-layer-issue-plan.md` (fresh-eyes reviewed; revised)
- [x] M1 — CUE model single-sourced via `categories` (`or()`-derived `#`-defs) + lifecycle + laws (documented-value, reachable/escapable); `cue vet` + export-shape gate; `issue-lifecycle` target; **human design-interface review** before consumers are wired
- [x] M2 — `cmd/vocabulary` (reuse `pkg/layergraph`) vet/export; `.dynamic-skill` weave wiring → `construct/generated/vocabulary/issue.json`; `ensure-cue` bootstrap + honest build-order; freshness stamp + `weave check`; touch-time skill instruction; atlas
- [x] M3 — Go binding in a shared importable `pkg/vocab` (embed once, consumers `import` it — supersedes M2's per-consumer copy + `issue-json-*` targets); rewire sdlc consumers to `vocab.Issue()` categories (carve out legit literals, honest grep); conformance check; `parked` acceptance; dedup AGENTS.md prose; close
- [x] M4 — widen the lifecycle to the legal set (+6 edges) + gate `set-status` on `CanTransition` with `--force` (operator chose (b) at the M3 boundary); meets the deferred "model-forbidden transition rejected" Done-when

## Log



- 2026-06-25: closed — All 4 milestones SHIP (M1 design-interface sign-off; M2/M3/M4 fresh-context substitute reviews — clean bills, recorded in Log). cue vet gate green; sdlc reads the issue model via pkg/vocab (scattered status literals gone, only annotated value carve-outs); set-status enforces the lifecycle graph with a --force escape; parked-status acceptance passed with zero Go edits; go test ./cmd/sdlc/... ./cmd/vocabulary/... ./pkg/vocab/... green.; review verdict: FIX-THEN-SHIP
- 2026-06-25: closed M4 — M4: set-status gates on the lifecycle graph (illegal transition refused w/ legal-targets msg + --force escape, dogfooded); model widened +6 legitimate edges; cue vet + full suite green; guard tests repointed off now-illegal edges + illegal/force tests added. Actual 0.73 = M4 increment (cumulative 4.55 − M1-3 3.82); review verdict: not-run
### 2026-06-24
- 2026-06-24: closed M3 — M3: pkg/vocab Go binding (embed once, consumers import); sdlc rewired to vocab.Issue() — isTerminalStatus+validStatuses deleted, honest grep shows only annotated #122 carve-outs; conformance test fail-closed; parked acceptance proved category propagation with zero production Go edits; cue vet + full suite green. (actual=M3 increment; M1 1.86+M2 1.07+M3 0.89=3.82 cumulative)
- 2026-06-24: closed M2 — M2: cmd/vocabulary vet/export/check (cue behind injected runner) + shared layergraph.MergeByName; .dynamic-skill → construct/generated/vocabulary + served SKILL.md breadcrumb; ensure-cue + vocabulary-build + weave PATH + committed issue.json + issue-json-check; go test ./cmd/vocabulary/... ./cmd/datatype/... ./pkg/layergraph/... ./cmd/sdlc/... green. (actual 1.07h = 2.93 cumulative − 1.86 M1; overlapping window); review verdict: not-run
- 2026-06-24: closed M1 — M1: cue vet gate green (valid vets, issue_invalid fails, export carries categories+lifecycle); design interface (issue.cue + issue-lifecycle target) reviewed + signed off by operator; no-code milestone (CUE data + target); review verdict: not-run
- Design chat (typing-markdown frame; captured in the pensive): refined the schema
  boundary — `issue.cue` stays **well-formedness only** (enum, required fields, types;
  the `age 0-100` class). The existing `CheckStructural` word-count/heuristic gates are
  *semantic* (soul-checks in disguise) → do NOT migrate them into the schema. **Scopes
  M3:** rewire only the enum/category/transition literals to the model; leave (or
  separately LLM-relocate) the semantic structural gates. Write-back stays an LLM job
  (validator emits diagnostics → LLM edits → re-vet), not a deterministic engine.
- Follow-up beyond #122 (motivated by the hand-edited-bad-status scenario): a general
  **instance-conformance validator** — `artifact → (locator) → typed instances →
  cue vet` against the type — wired at edit/pre-merge surfaces across *all* typed
  markdown. datatype = prose skill + schema; the *extractor/locator* is the new piece
  (cue vet is the easy half). To be filed as its own issue.

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
- Durable plan authored (`workshop/plans/000122-…-plan.md`), estimate set to 12h.
  Fresh-eyes plan review caught a showstopper: CUE `#`-definitions don't `cue export`,
  so the category membership the consumer needs wasn't in the JSON. Fixed *better* than
  the suggested concrete-list-plus-law by making `categories` the concrete single source
  and deriving the `#`-defs via `or()` — DRY by construction, no drift. Review also
  found ≥7 unlisted status-literal sites (close.go/state.go drift-switch/setstatus stamp
  trigger/push GitHub branch) → M3 now enumerates every site + carves out legitimate
  state-writes with an honest grep; and an overstated `go:generate` freshness claim →
  fixed via a make chain (`ensure-cue → schema-install → schema-gen → sdlc-build`) with a
  committed `issue.json`. Added the `ensure-cue` bootstrap, `issue-lifecycle` target, and
  an explicit **M1 human design-interface review** (CUE as the human/LLM design surface).
- M2 fresh-eyes review (subagent — the milestone-close judge CLI needs sandbox-blocked
  network, so closed `--no-judge`; this is the substitute review per §3): **no Critical**;
  the `datatype`→`layergraph.MergeByName` refactor verified behavior-preserving (shadow
  ordering tested). Fixed in `c66d4c3`: Important — `osCue.Export` used `.Output()` and
  dropped CUE's stderr (export failures undiagnosable) → route stderr to a buffer; Nice —
  error on the ambiguous `--noun`+`--output` combo, `.gitignore` the stray `/vocabulary`
  + `/datatype` root binaries. Accepted as-is: `readSources` bypassing the FS seam.
  **Effective verdict: SHIP** — M2 sealed. (Remaining integration: wire `make
  issue-json-check` into ariadne CI — one line.)
- M3 design refinement (per-language binding — from design chat; captured in the plan's
  `## Revisions`): placement is per consumer-*language*, not per instance — the language's
  module system distributes one canonical copy. M3 puts the **Go binding** in a shared
  importable `pkg/vocab` (embed once; sdlc + any future Go consumer `import` it), declared
  by a co-located `//go:generate` and regenerated by one generic `go generate` /
  `make vocab-embed`. This **supersedes** M2's per-entity `issue-json-gen`/`issue-json-check`
  + the `cmd/sdlc/internal/issue/issue.json` copy — so the "wire issue-json-check into CI"
  item above is replaced by a generic go-generate diff check. General per-language bindings
  config `(language, dir, form)` deferred to the 2nd language (Lua/parley). Binding stays
  out of `issue.cue` (no protobuf `go_package` wart).
- M3 fresh-eyes review (subagent; close ran `--no-judge`, sandbox): **clean bill — no
  Critical/Important**. Mutation-verified the conformance test has teeth (drop a
  transition / blank a `when` → it fails); carve-out judgment correct at all 11 sites;
  transition-gating deferral honestly documented (vocab.go, atlas, Log). Fixed 2 stale
  comments naming the deleted `issue-json-gen` (Makefile.workflow:829,
  cmd/vocabulary/export.go:35). **M3 verdict: SHIP — M3 sealed.** Remaining before the
  whole-issue close: operator ratifies the transition-gating deferral (amend the
  "model-forbidden transition rejected" Done-when + file a follow-up), then `sdlc close
  --issue 122`.
- 2026-06-25: operator chose **(b) enforce** (not defer). M4 widens the lifecycle (+6
  legitimate edges — `open→wontfix/punt`, `punt→working`, `wontfix→working`,
  `blocked→wontfix/punt`; edge set approved, iterate later) and gates `set-status` on
  `CanTransition` with a `--force` escape (`claim`/`close` stay ungated). This **meets**
  the previously-deferred "model-forbidden transition rejected" Done-when. Forked the M4
  implementation.
- M4 fresh-eyes review (subagent; close `--no-judge`, sandbox): **clean bill — no
  Critical/Important, SHIP**. Verified the gate covers every mutation entry path
  (`applyStatus`'s two callers — `set-status` gated, `claim`'s is the legal open→working;
  `close`/`milestone-close` write `done` directly by design; `state.go` read-only), the
  gate has teeth (mutation-checked: removing Guard 0 reddens the illegal-transition
  tests), the 3 test repoints are legitimate (none a missing edge), laws hold fail-closed,
  Done-when genuinely met. 2 non-blocking Nice-to-haves (no action). **M4 verdict: SHIP.**
  All four milestones SHIP (substitute fresh-context reviews, recorded here per gate);
  every Done-when criterion met.
- 2026-06-25: **#122 closed** — `sdlc close --issue 122 --actual 4.67` (est 10h, ratio
  2.1×, logged to the calibration ledger). The whole-issue boundary review *did* run (the
  judge CLI was reachable this time) → **FIX-THEN-SHIP**, no Critical: fixed the one
  Important (`set-status` help still said "all other transitions allowed" — now documents
  the M4 lifecycle gate) + the Minor (atlas heading M1–M3 → M1–M4). **#122 DONE** — all 4
  milestones SHIP, every Done-when met.
- 2026-06-25 **scope reconciliation (operator correction):** #122 delivered the *source*
  (`issue.cue`) + the *sdlc-Go consumer* (`pkg/vocab`) + lifecycle *enforcement*. But the
  Done-when said "compiled to consumers" / "Go/**Lua** propagate" — and the **Lua consumer
  (parley) and the operator-prose consumer (sdlc help-text) were NOT wired**; I'd folded
  them into silent "follow-up," leaving `issue.cue` as just-documentation those surfaces
  don't derive from. So #122 proves the pattern + the primary consumer + enforcement;
  *full multi-consumer derivation* completes when **parley#135** (Lua) + **ariadne#125**
  (help-text-from-vocab) land. Lesson recorded in `workshop/lessons.md` (don't let
  "follow-up" offload an issue's purpose; shadow-sweep at close).
