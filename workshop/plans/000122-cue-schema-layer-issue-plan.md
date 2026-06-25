# Formal Vocabulary Layer (issue noun) Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make one CUE source the single authoritative model of the `issue` noun + its lifecycle, compiled to JSON that `sdlc` reads, so the scattered status literals collapse to derived reads — proving the formal-vocabulary-layer pattern (and **CUE as the human/LLM design interface**) on the oldest, most-scattered noun.

**Architecture:** The vocabulary source lives in `construct/vocabulary/` and is DAG-merged across the `construct/deps` layer graph exactly like `datatype` (ARCH-DRY — reuse `pkg/layergraph`, not a new propagation mechanism). A `vocabulary` skill-binary validates (`cue vet`) and exports JSON into `construct/generated/vocabulary/` at `weave compile` via a `.dynamic-skill` (ARCH-DRY with the `datatype` `.dynamic-skill`). `sdlc` consumes the **exported JSON** (CUE stays a build-time tool — no `cuelang` Go dependency), unmarshalling it into a pure `IssueModel` whose category/transition predicates replace the scattered literals (ARCH-PURE — pure functions; the only IO seam is the CUE/JSON boundary).

**Tech Stack:** CUE (build-time `cue` CLI v0.16+; export-to-JSON + `go:embed`, no runtime CUE dependency), Go, `pkg/layergraph`, the weave dynamic-skill mechanism, `Makefile.workflow` bootstrap.

**Estimate:** ~10h derived (M1 ≈ 2.6h, M2 ≈ 4.0h, M3 ≈ 3.4h) — see the issue's `## Estimate` block. `estimate_hours: 10`.

**Decided defaults (were open forks):**
- **Location: `construct/vocabulary/`** (in construct ⇒ propagates like `datatype`). Generated outputs in `construct/generated/vocabulary/` (gitignored). Repo-local nouns would live in `<repo>/vocabulary/` (shadows by filename), same as `datatype`'s shared-vs-local axis.
- **`categories` is the concrete single source; the `#`-definitions derive from it via `or()`** (so categories export to JSON and the defs still validate, with no double-statement — fixes the `cue export` gap; see M1).
- **One generated face: `issue.json`.** Justified in a sentence: "Lua/Go can't read CUE." The `.cue` source serves humans and the LLM directly; the `issue-lifecycle` target is the human narrative. **No `.md` render, no `@vocab(public)` projector** — dropped as YAGNI (nothing consumes a projection yet). Public/private stays a *design lens*; the projection is built only when a consumer (e.g. always-loaded routing, parley autocomplete) needs it.
- **The eager breadcrumb is an instruction, not an artifact:** a one-line pointer in the vocabulary skill — "read `construct/vocabulary/issue.cue` before changing the issue lifecycle."
- **`issue.cue` is a tier-1, human-reviewed design interface.** M1 ends with an explicit human design-review checkpoint before any consumer is wired.
- **`parked` is NOT in the base model** — it's the M3 acceptance scenario proving clean evolvability.

**Out of scope (follow-up issues):** parley.nvim consuming the JSON; migrating `datatype` prototypes onto this layer; building the `@vocab` projection.

---

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `issue.cue` model (`categories`, derived `#Status`/`#Active`/`#Terminal`, `#Issue`, `when`, `lifecycle`, `laws`) | `construct/vocabulary/issue.cue` | new |
| `issue-lifecycle` target (the defended invariant — *why*, references the cue) | `workshop/targets/issue-lifecycle.md` | new |
| `IssueModel` (Go) — unmarshalled `{categories, when, lifecycle, laws}` | `cmd/sdlc/internal/issue/model.go` | new |
| category predicates — `IsTerminal`, `IsActive`, `IsOpen`, `AllStatuses` | `cmd/sdlc/internal/issue/model.go` | new |
| transition predicate — `CanTransition(from,to)`, `Guards(from,to)` | `cmd/sdlc/internal/issue/model.go` | new |
| `isTerminalStatus` (sdlc) | `cmd/sdlc/push.go` | deleted |
| `validStatuses` (sdlc) | `cmd/sdlc/setstatus.go` | deleted |

- **`issue.cue` model** — the noun's data shape + lifecycle + laws, in CUE, **legible-first** (intent comments per state/transition; `when` semantics inline). Single-source mechanics (resolves the export gap + DRY tension):
  ```cue
  categories: { open: ["open"], active: ["working","blocked"], terminal: ["done","wontfix","punt"] }
  #Active:   or(categories.active)
  #Terminal: or(categories.terminal)
  #Status:   or(categories.open + categories.active + categories.terminal)
  ```
  `categories` is the only place membership is stated (exports to JSON); `#Status` validates issue files. No consistency law needed.
  - **DRY rationale (ARCH-DRY):** collapses the issue model that today lives in ≥3 places (base prose, sdlc Go literals, parley Lua) into one source.

- **`issue-lifecycle` target** — the *global why / invariant* layer: why each state and transition exists, what we rejected, what must not drift. **References `issue.cue` for the *what*; never restates the transition table** (a target that lists `open → working` has drifted back into being the redundant `.md`). The cue holds the *what* (machine truth); inline comments hold the *local why*; the target holds the *global why/invariant*. Three layers, no overlap.

- **`IssueModel` + predicates (Go)** — pure interpretation of the exported JSON (`categories`, `lifecycle`, `when`, `laws`). Unmarshals the concrete fields, never the `#`-definitions (they aren't in the JSON). ARCH-PURE: pure functions, unit-tested with no mocks; the embedded-JSON read is the thin IO seam.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `cue` runner (`vet`/`export`) | `cmd/vocabulary/cue.go` | new | the `cue` CLI |
| vocabulary DAG-merge | `cmd/vocabulary/merge.go` (prefer shared helper — Task 2.1) | new | `pkg/layergraph` |
| `.dynamic-skill` weave hook | `construct/local/vocabulary/.dynamic-skill` | new | `weave compile` |
| committed embed JSON | `cmd/sdlc/internal/issue/issue.json` (`//go:embed`) | new | the generated artifact |
| `ensure-cue` + build-order wiring | `Makefile.workflow` | new | the toolchain bootstrap |
| freshness stamp + `weave check` | `cmd/vocabulary/stamp.go` + weave check surface | new | merged-source hash |

- **`cmd/vocabulary` binary** — `vocabulary vet` (wraps `cue vet`), `vocabulary export [--output dir]` (DAG-merge then `cue export --out json` per noun), `vocabulary stamp`/`check` (merged-source hash). Pure merge/hash logic unit-tested; `cue`-CLI calls go through a small injected runner (ARCH-PURE), mirroring `cmd/datatype` + `cmd/sdlc/internal/gitx.GitRunner`.
- **embed JSON** — `issue.json` is **committed** (so standalone `go build ./cmd/sdlc` works, per the `helptext` embed precedent) and regenerated by the make chain (Task 2.4); CI asserts it isn't stale. The `construct/generated/` artifacts stay gitignored.

---

## Chunk 1: M1 — the model, the vet gate, and the design-interface review

> Milestone **M1** review boundary. Pure data + the human design review; no Go. Closes with `sdlc milestone-close --issue 122 --milestone M1`.

### Task 1.1: The CUE model (legible-first)

**Files:** Create `construct/vocabulary/issue.cue`, `construct/vocabulary/issue_invalid.cue`; Test `construct/vocabulary/vet_test.sh`.

- [ ] **Step 1 — failing gate.** `vet_test.sh` asserts: (a) `cue vet construct/vocabulary/issue.cue` exits 0; (b) vet on the invalid fixture exits non-zero; (c) `cue export … --out json` output contains a top-level `categories` object (with `active`/`terminal`/`open` arrays) **and** a `lifecycle` array — front-loads the "definitions don't export" risk into M1.
- [ ] **Step 2 — run, expect FAIL** (no model yet).
- [ ] **Step 3 — write `construct/vocabulary/issue.cue`:** the `categories`-as-source + `or()`-derived definitions; `#Issue` with the compiled guard `if status == "done" { actual_hours!: number & >0 }`; the `when` map; the `lifecycle` transition list with **intent comments per transition** and named `guards`. This file is the human design interface, not just a build input.
- [ ] **Step 4 — the laws (finalize syntax against `cue` v0.16):** `documented-value` (every status has a non-empty `when`); `reachable`/`escapable` (the orphan-state gate the Done-when requires) via `import "list"` over the concrete lifecycle:
  ```cue
  import "list"
  _froms: [for t in lifecycle {t.from}]
  _tos:   [for t in lifecycle {t.to}]
  laws: {
  	"documented-value": {for s in (categories.open+categories.active+categories.terminal) {(s): when[s] & !=""}}
  	"reachable": {for s in (categories.active+categories.terminal) {(s): list.Contains(_tos, s) & true}}
  	"escapable": {for s in (categories.open+categories.active)     {(s): list.Contains(_froms, s) & true}}
  }
  ```
- [ ] **Step 5 — `issue_invalid.cue`:** violates a law (empty `when`, or an unreachable added state) to prove the gate bites.
- [ ] **Step 6 — run, expect PASS.**
- [ ] **Step 7 — commit.** `#122 M1: issue noun + lifecycle + laws as a CUE vocabulary, single-sourced via categories (ARCH-DRY)`

### Task 1.2: The `issue-lifecycle` target (the *why*, not a copy)

**Files:** Create `workshop/targets/issue-lifecycle.md` (use the `target` datatype); modify issue `000122` frontmatter (`target: issue-lifecycle`).

- [ ] Author the target as the **global why / invariant** only: why each state and transition exists, alternatives rejected, what must not drift. **Point at `construct/vocabulary/issue.cue` as the source of the *what*; do NOT restate the transition table** (that would recreate the redundancy the `.md` was dropped for). Commit.

### Task 1.3: Human design-interface review checkpoint (the experiment)

- [ ] Present `construct/vocabulary/issue.cue` + `workshop/targets/issue-lifecycle.md` to the operator for sign-off **before** M2 wires any consumer — the highest-leverage review, and *the* test of CUE-as-design-interface. Distinct from the milestone-close judge (which reviews code). Record sign-off in `## Log`.
- [ ] `sdlc milestone-close --issue 122 --milestone M1`.

---

## Chunk 2: M2 — bootstrap, compile pipeline, freshness

> Milestone **M2** review boundary. Closes with `sdlc milestone-close --issue 122 --milestone M2`. Tasks specified to function level; bite-sized TDD steps finalized at M2 start against M1's exact model.

### Task 2.1: `cmd/vocabulary` binary — `vet` + `export` (reuse `pkg/layergraph`)

**Files:** Create `cmd/vocabulary/{main.go,cue.go,merge.go}`; Test `cmd/vocabulary/{merge_test.go,cue_test.go}`.

- [ ] **DRY decision (ARCH-DRY):** `cmd/datatype/merge.go`'s `overlayDir`/`mergeTypes` is a generic "merge `*.X` across the layer graph keyed by filename" routine. **Prefer extracting it into `pkg/layergraph` (or a shared helper)** and have both `datatype` and `vocabulary` call it. If that balloons scope, copy with a `TODO(#NNN)` + follow-up issue — do not silently duplicate.
- [ ] TDD: `vocabulary export` DAG-merges `construct/vocabulary/*.cue` (local shadows shared), then shells `cue export --out json`. Pure merge logic tested with a fake layergraph; `cue` behind the injected runner. `vocabulary vet` wraps `cue vet`.

### Task 2.2: weave wiring + the touch-time breadcrumb instruction

**Files:** Create `construct/local/vocabulary/.dynamic-skill` (mirror `construct/local/datatype/.dynamic-skill`); verify `construct/generated/vocabulary/` is gitignored.

- [ ] `.dynamic-skill` runs `vocabulary export --output construct/generated/vocabulary`. Run `make weave`; confirm `construct/generated/vocabulary/issue.json` materializes in the compiling repo's tree.
- [ ] **One-line touch-time instruction** in the vocabulary skill (SKILL.md / `.dynamic-skill` output): "before editing the issue lifecycle, read `construct/vocabulary/issue.cue`." This is the breadcrumb — an instruction, not a generated doc.

### Task 2.3: `ensure-cue` bootstrap + honest build-order

**Files:** Modify `Makefile.workflow`.

- [ ] Add `.PHONY: ensure-cue` mirroring `ensure-go` (lines 237–249): `command -v cue` → no-op; else `brew install cue` on macOS; else fail-fast pointing at cuelang.org. Add `ensure-cue` to the `bootstrap` prereq list (line 258).
- [ ] **Honest freshness/build-order:** `go build` does NOT run `go generate`, and `vocabulary` must exist first. Add a make chain `ensure-cue → vocabulary-install (build cmd/vocabulary) → issue-json-gen (vocabulary export > cmd/sdlc/internal/issue/issue.json) → sdlc-build`, with `sdlc-build`/`build` depending on `issue-json-gen`. Keep a `//go:generate` directive as documentation, but the *guarantee* is the make dependency; `issue.json` is **committed** so standalone `go build` works.

### Task 2.4: freshness stamp + `weave check`

- [ ] **First determine the `weave check` surface:** existing subcommand, or new `cmd/weave` surface? TDD `vocabulary stamp` (hash of merged source into the artifact) + `vocabulary check` (recompute/compare, non-zero on mismatch); CI asserts `git diff --exit-code` on the committed `issue.json` after regen. Pure hash logic unit-tested.

### Task 2.5: M2 milestone close

- [ ] Atlas: add an `atlas/` page for the vocabulary layer (new surface + generated-faces flow); link from `atlas/index.md`. `sdlc milestone-close --issue 122 --milestone M2`.

---

## Chunk 3: M3 — rewire sdlc, conformance, delete duplicates

> Milestone **M3** review boundary. Closes with `sdlc close --issue 122 --milestone M3 --verified '<evidence>'`. The `parked` scenario is the verification.

**Binding model (design refinement — see ## Revisions).** Placement of a generated artifact is a property of the consumer's *language*, not the consumer *instance* — and the language's own module system distributes one canonical copy to every consumer of that language. So M3 does NOT copy `issue.json` per consumer package (the per-entity `issue-json-gen` smell). The **Go binding** is one shared importable package `pkg/vocab` that `go:embed`s the noun JSONs and exports accessors; every Go consumer (sdlc now, anything later) `import`s it — the import graph *is* the distribution. The co-located `//go:generate` directive is the Go binding declaration; one generic `go generate ./...` (`make vocab-embed`) regenerates it (no per-noun/per-consumer Make target). A general repo-local *bindings config* `(language, dir, form)` is deferred until a second language needs placement (the Lua/parley follow-up); for Go, `go:generate` is the native declaration. Keep the binding OUT of the noun — `issue.cue` stays pure (avoid protobuf's `option go_package` wart).

### Task 3.1: `pkg/vocab` — the Go binding (shared, importable; ARCH-PURE/DRY)

**Files:** Create `pkg/vocab/{vocab.go,vocab_test.go,issue.json}` + the `//go:generate` directive. **Delete** M2's per-consumer copy `cmd/sdlc/internal/issue/issue.json` and the per-entity `issue-json-gen`/`issue-json-check` Make targets (superseded by the binding).

- [ ] **Failing test:** in `pkg/vocab`, `Issue().IsTerminal("done")==true`, `IsTerminal("working")==false`, `IsActive("blocked")==true`, `CanTransition("open","working")==true`, `CanTransition("open","done")==false`.
- [ ] Implement: `pkg/vocab/issue.json` `//go:embed`'d once; `//go:generate vocabulary export --noun issue -o issue.json` co-located; unmarshal `{categories, when, lifecycle, laws}` into an `IssueModel`; pure predicates over `categories`/`lifecycle` (no `#`-def parsing). Run → PASS. Commit.
- [ ] Makefile: replace `issue-json-gen`/`issue-json-check` with one generic `vocab-embed` (= `go generate ./...`) + a generic `git diff --exit-code` over the generated files. No per-noun/per-consumer target.

### Task 3.2: Rewire the consumers — read from `pkg/vocab`; complete site list + honest grep

sdlc consumers `import .../pkg/vocab` and branch on `vocab.Issue()` predicates (NOT a sdlc-internal copy).

**Rewire (category/transition branching → model):**
- [ ] `cmd/sdlc/push.go` — delete `isTerminalStatus` (494–495); `vocab.Issue().IsTerminal` at 370, 441, 467.
- [ ] `cmd/sdlc/merge.go:550` — `isTerminalStatus` → `vocab.Issue().IsTerminal`.
- [ ] `cmd/sdlc/setstatus.go` — `validStatuses` (40) → `vocab.Issue().AllStatuses()`; transition guards at 195 (`prev=="open"&&status=="working"`), 222 (`next=="done"`), 240 → `CanTransition` + named events.
- [ ] `cmd/sdlc/state.go:300–319` (detectDrift `switch i.Status`) **and** `:341` → category predicates.
- [ ] `cmd/sdlc/claim.go:125` (`prev!="open"`) → model.
- [ ] `cmd/sdlc/startplan.go:308` (`=="working"`) → `IsActive`/explicit.

**Legitimate literals — carve out with rationale (NOT rewired), excluded from the grep:**
- [ ] `cmd/sdlc/close.go:459` `SetField(...,"done")` — writes a target state; source from the transition `to` or annotate `// literal: terminal write, not a category branch`.
- [ ] `cmd/sdlc/close.go:376`, `cmd/sdlc/push.go:471` (`=="done"`) — assess: rewire to `IsTerminal` if a category test, annotate if genuinely value-specific.

- [ ] **Honest acceptance:** `grep -n '"open"\|"working"\|"blocked"\|"done"\|"wontfix"\|"punt"' cmd/sdlc/*.go | grep -v _test` returns **only** the annotated carve-outs (the parse plumbing now lives in `pkg/vocab`, outside `cmd/sdlc`); the milestone log records the carve-out list. `go test ./cmd/sdlc/... ./pkg/vocab/...` after each file; commit per file.

### Task 3.3: Conformance test — a *check*, not a maintained list

- [ ] `pkg/vocab/conformance_test.go` reads the embedded model at runtime and asserts: every declared transition is accepted by `CanTransition`, every forbidden one rejected, every status has a category and a `when`. Cases derived from the model (no hand-maintained list) — adding a value it can't place fails the test.

### Task 3.4: Acceptance scenario — `parked` proves clean evolvability

- [ ] Add `parked`: append to `categories.active`, add `working↔parked` transitions, add its `when` (≈4 lines). Regenerate (`make vocab-embed` / `go generate ./...`); rebuild.
- [ ] **Verify with no Go edits:** `IsActive("parked")==true`, `IsTerminal("parked")==false`, `setstatus working→parked` accepted; `state`/`list` treat it active. Any raw-value branch that slipped through is caught by Task 3.3.
- [ ] Revert `parked` (or keep); record outcome in `## Log` as verification evidence.

### Task 3.5: Close

- [ ] Atlas: update the vocabulary-layer page — the Go binding is the shared imported `pkg/vocab`, the per-language binding model, and the `go generate` regen (supersedes the M2 `cmd/sdlc/internal/issue` embed + `issue-json-*` targets). Reconcile the `issue-lifecycle` target; dedup the AGENTS.md status-enumeration prose to point at `construct/vocabulary/issue.cue` (ARCH-DRY).
- [ ] `sdlc actual --issue 122` (measured), then `sdlc close --issue 122 --milestone M3 --verified '<evidence>'`.

---

## Chunk 4: M4 — enforce the lifecycle graph in set-status (decision (b))

> Milestone **M4** review boundary. Closes with `sdlc milestone-close --issue 122 --milestone M4`. At the M3 boundary the operator chose **(b)** — enforce now, not defer. Enforcing requires the model to first permit all legitimate flows (else triage/resume break), so M4 widens the lifecycle, then gates the arbitrary-flip surface.

### Task 4.1: Widen the lifecycle to the legal set

**Files:** `construct/vocabulary/issue.cue`; regenerate `pkg/vocab/issue.json`.

- [ ] Add 6 transitions (events reuse the vocabulary, more `from` states): `open→wontfix` (abandon), `open→punt` (defer), `punt→working` (reopen), `wontfix→working` (reopen), `blocked→wontfix` (abandon), `blocked→punt` (defer). Left illegal (→ `--force`): `open→done` (a done needs actuals), `working→open`, `open→blocked`.
- [ ] `cue vet` green — laws still hold (terminals gaining a `reopen` outbound is fine; `escapable` only requires non-terminals to exit). `make vocab-embed` regenerates the embedded JSON; the freshness/conformance checks pass.

### Task 4.2: Gate set-status on CanTransition + --force

**Files:** `cmd/sdlc/setstatus.go` (+ help); tests.

- [ ] In the `set-status` command path ONLY (NOT `claim`/`close`, which perform fixed legal transitions): if `!vocab.Issue().CanTransition(prev, next)` and not `--force` → refuse with a clear error naming the legal targets from `prev` + "pass `--force` to override (logged)". Add `--force`; log the override (per sdlc's per-gate convention).
- [ ] TDD: a legal flip passes; `open→done` and `done→wontfix` are rejected; `--force` overrides; the 6 new edges pass as regressions.

### Task 4.3: Close

- [ ] Atlas + `CanTransition` doc + `issue-lifecycle` target: the graph is now **enforced** at `set-status` (with `--force` escape) — supersede M3's "exposed but not enforced" note; the target's "transitions are gated, not free" is now literally true at the flip surface.
- [ ] `sdlc milestone-close --issue 122 --milestone M4 --no-judge --actual <measured> --verified '<evidence>'`.

---

## Verification (end-to-end)

- `sh construct/vocabulary/vet_test.sh` — valid passes, broken fails, **export contains categories + lifecycle**.
- `make bootstrap` provisions `cue`; `make weave` materializes `construct/generated/vocabulary/issue.json`; `make vocab-embed` (`go generate ./...`) regenerates `pkg/vocab` and sdlc imports it (no per-consumer copy).
- `go test ./cmd/sdlc/... ./cmd/vocabulary/... ./pkg/vocab/...` — green, incl. the conformance test.
- `grep` shows no un-annotated status literals in non-test sdlc code.
- `parked` propagates to every category-based consumer with zero Go edits.

## Revisions

### 2026-06-24 — Per-language binding for the Go consumer (M3)

**Reason:** M2 shipped a committed `cmd/sdlc/internal/issue/issue.json` regenerated by a per-entity Make target (`issue-json-gen`/`issue-json-check`). Operator flagged the smell: a target per (noun × consumer) doesn't scale and violates single-mechanism (`feedback_minimum_mechanism`). The design chat resolved that **placement is per consumer-*language*, not per instance** — the language's own module system distributes one canonical copy to all consumers of that language. (The IDL multi-backend model — protobuf/Thrift — with protobuf's `option go_package` as the wart to avoid.)

**Delta:** M3 puts the **Go binding** in a shared importable `pkg/vocab` (embeds the noun JSONs once; sdlc and any future Go consumer `import` it — the import graph distributes). The co-located `//go:generate` is the Go binding declaration; one generic `go generate ./...` (`make vocab-embed`) + a generic git-diff replaces the per-entity targets, which are deleted along with `cmd/sdlc/internal/issue/issue.json` (superseded). A general per-language bindings config `(language, dir, form)` is deferred to the second language (Lua/parley follow-up). The binding stays repo-local and out of `issue.cue` (the noun stays pure). Affects Tasks 3.1–3.5 above.

### 2026-06-25 — Enforce the lifecycle graph (decision (b), M4)

**Reason:** At the M3 boundary the operator chose to **enforce** the transition graph now (option b) rather than defer it to a follow-up. Enforcing requires the model to first permit every *legitimate* flow — otherwise triage (`open → wontfix/punt`) and resume (`punt → working`) get wrongly rejected — so M4 widens the lifecycle, then gates.

**Delta:** +6 edges (`open→wontfix`, `open→punt`, `punt→working`, `wontfix→working`, `blocked→wontfix`, `blocked→punt`; operator-approved, "iterate later"). `set-status` gates on `CanTransition` with a `--force` escape; `claim`/`close` (fixed legal transitions) stay ungated. M3's "exposed but not enforced" note is superseded, and the Done-when *"a model-forbidden transition is rejected by sdlc"* (deferred at M3) is now **met**. Adds Chunk 4.
