# weave Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace `construct/setup.sh` with `cmd/weave`, a Go intent-compiler that composes each repo's agentic context from a DAG of layers and serves skills at runtime.

**Architecture:** A **pure compiler core** — parse layer manifests into typed intents, resolve the layer DAG (topo-sort + dedup), compute the cascade, and render an ordered list of filesystem `Action`s — wrapped by a **thin filesystem IO seam** that applies the Actions (ARCH-PURE). Git/cloning stays in the existing `bootstrap.sh` shell seed; weave only ever sees already-present sibling dirs, so its IO seam is the filesystem, not git. A **golden-diff harness** compares weave's computed output against `setup.sh`'s real output on the live repos, with every divergence explained in a ledger.

**Tech Stack:** Go 1.26 (`github.com/xianxu/ariadne`), stdlib + `encoding/json`. JSON-merge semantics ported from `construct/scripts/merge-settings.sh`. New binary `cmd/weave/`; bootstrap stays the thin shell stub.

**Spec:** `workshop/issues/000095-weave.md`. **Principles:** ARCH-PURE, ARCH-DRY (delivered via `sdlc start-plan`).

---

## Core concepts

The compiler is one pipeline: `read deps+manifests → Resolver → []Layer → Planner → []Action → apply`. Everything left of `apply` is pure and table-tested; `apply` and the readers are the thin seam.

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `Intent` | `cmd/weave/internal/intent/intent.go` | new |
| `Manifest` | `cmd/weave/internal/intent/manifest.go` | new |
| `Layer` | `cmd/weave/internal/layer/layer.go` | new |
| `Resolve` (topo-sort + dedup) | `cmd/weave/internal/layer/resolve.go` | new |
| `Action` | `cmd/weave/internal/plan/action.go` | new |
| `Plan` (lower `[]Layer` → `[]Action`) | `cmd/weave/internal/plan/plan.go` | new |
| `composeProse` | `cmd/weave/internal/plan/prose.go` | new |
| `mergeSettings` | `cmd/weave/internal/plan/settings.go` | new |
| `SkillIndex` (menu + body lookup) | `cmd/weave/internal/skill/skill.go` | new |

**Test surface:** each entity gets a colocated `_test.go` running with **no IO mocks** — that's the purity proof. The Planner is tested by asserting the `[]Action` it computes from in-memory `[]Layer`; the golden-diff harness (M2) asserts those Actions render to the same tree `setup.sh` produces.

- **`Intent`** — a typed manifest entry. Kinds are a **hybrid** (verified against the live `base.manifest`, which is ~all generic infrastructure verbs): the **ported file-op verbs** `Symlink | Seed | Scaffold | Touch | Merge | Tool` (the dominant case — most of every manifest), plus the **new semantic intents** `Prose | Skill`. `Merge` *is* the `settings` cascade. `Prose` replaces the `@AGENTS.local.md` @-import; `Skill` unifies the `sync-local-skills.sh` hook + per-layer skill symlinks. So #95 **adds** semantic intents *atop* the ported file-ops — it does not replace file-ops wholesale (the earlier "rises from file-ops to intents" framing over-promised).
  - **Relationships:** N Intents : 1 Manifest : 1 Layer.
  - **DRY rationale (ARCH-DRY):** one lowering switch, one `case` per kind. File-op kinds lower **near-identity** (a `Symlink` intent → a `Symlink` action — ported directly from `walk_manifest`'s `case`); semantic kinds **compose** (all layers' `Prose` → one composed `AGENTS.md`). One source of truth per kind.
  - **Future extensions:** a `Suppress` intent (v2); new kinds slot into the switch.
- **`Manifest`** — parsed `construct/base.manifest` for one layer: `deps []string` (the header, sole layer-edge source — kills the `go.mod`-replace channel) + `[]Intent`.
  - **Relationships:** 1:1 with Layer.
- **`Layer`** — a resolved layer: `Path`, `Name`, `Manifest`. Resolved to one on-disk sibling dir (no versioning).
- **`Resolve`** — pure topo-sort + dedup over the `Path→deps` edge set, foundation-first; a diamond collapses to one application per layer. Behavioral spec = `setup.sh:discover_ancestors` (ARCH-DRY: port its ordering + `_seen_or_add`, don't reinvent).
  - **Future extensions:** cycle detection error (parity with the bash's circular guard).
- **`Action`** — one pending filesystem op: `Symlink{src,dst} | WriteFile{path,content} | Mkdir{path} | GoModEdit{...}`. The pure/IO boundary: the Planner *computes* Actions; the seam *executes* them. Golden tests diff computed Actions (and rendered file content) without touching real repos.
- **`Plan`** — pure: `[]Layer → []Action`. Walks layers foundation-first, lowers each Intent to Action(s), applies the cascade. Behavioral spec = `setup.sh:walk_manifest`'s dispatch + the self-reference filter.
  - **DRY rationale:** the single lowering switch; backends (M4) plug in here behind an interface, not by forking the switch.
- **`composeProse`** — pure: concatenate prose fragments foundation-first into the `AGENTS.md` body + the compiled skill menu. Fixes the `@AGENTS.local.md` bug structurally (no `@`-import; the repo's fragment is concatenated in directly).
  - **Manifest grammar:** each layer declares `prose <relpath>` (e.g. `prose AGENTS.local.md`); weave concatenates all layers' fragments foundation-first into the repo's real `AGENTS.md`. No `prose` verb exists today (prose flows via the `@`-import inside the *symlinked* `AGENTS.md`) — this is new syntax M1's `Manifest` parser targets. Consequence: weave's manifest replaces `symlink AGENTS.md` with `prose …`, so `AGENTS.md` flips from a symlink (setup.sh) to a composed real file — an expected, explained golden-diff divergence (it *is* the bug fix), distinct from the ~40 generic symlinks that must match byte-for-byte.
- **`mergeSettings`** — pure: deep-merge dicts, `$merge_keys` union, `$remove` filter, strip meta keys. Behavioral spec = `merge-settings.sh` (ARCH-DRY: port semantics verbatim; golden-test against its output).
- **`SkillIndex`** — pure: aggregate skills across layers → `(name, description)` menu + name→body-path lookup. Backs `weave skills` / `weave skill <name>`; no `.claude/skills/` assumption.

### Integration points (the thin seam)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `FS` | `cmd/weave/internal/weavefs/fs.go` | new | filesystem (read/write/symlink/mkdir/readlink) |
| `Apply` | `cmd/weave/internal/plan/apply.go` | new | `FS` |
| `GoMod` | `cmd/weave/internal/weavefs/gomod.go` | new | `exec` `go mod edit` |
| `weave` CLI | `cmd/weave/main.go` | new | `os.Args`, stdout |

- **`FS`** — interface for reads (deps/manifests/existing tree) + mutations (symlink/write/mkdir). **Injected into** `Apply` and the manifest readers, so the pure core is tested against an in-memory/`t.TempDir()` FS.
- **`Apply`** — executes `[]Action` idempotently (symlink replaces an existing symlink; write-on-drift for seeds). The *only* mutating code; kept dumb. Behavioral spec = `setup.sh:create_symlink/create_seed/create_scaffold`.
- **`GoMod`** — the one `exec` dependency, for the `tool` intent (`go mod edit` require/replace/tool). Behavioral spec = `ensure_go_tool_dependency`. Injected so the Planner stays pure.
- **`weave` CLI** — subcommands: `weave` (compile, default), `weave skills` (print menu), `weave skill <name>` (print body), `--dry-run` (print the `[]Action`, no apply — the golden-diff hook).

**Out of scope for weave (stays shell):** peer cloning (`bootstrap.sh`/`bootstrap-peers.sh`) and data-dep cloning (`clone-data-deps.sh`) — weave assumes siblings present. Vendor mode is **removed** entirely (symlink-only).

---

## Milestones

> M1–M2 carry full bite-sized TDD steps below. **M3–M5 are task-level**; their bite-sized steps get authored at each milestone-start (per AGENTS.md "map, don't over-specify" — later design depends on M1–M2 reality). Each `Mx` is its own `sdlc milestone-close` review boundary.

### M1 — Dependency model: `construct/deps` as sole layer-edge source

**Goal:** Establish + test the resolver and the dep-model cutover *before* any lowering, so the rest builds on a correct graph.

**Files:** Create `cmd/weave/internal/layer/{layer.go,resolve.go,resolve_test.go}`, `cmd/weave/internal/intent/{manifest.go,manifest_test.go}` (deps-header parse only). Reference: `construct/setup.sh:185` (`discover_ancestors`).

- [ ] **Step 1 — failing test: topo-sort foundation-first.** `resolve_test.go`: given edges `brain→{ariadne,nous}, nous→ariadne`, assert order `[ariadne, nous, brain]`.
- [ ] **Step 2 — run, expect FAIL** (`go test ./cmd/weave/internal/layer/` → undefined `Resolve`).
- [ ] **Step 3 — implement `Resolve`** (Kahn/DFS topo + `seen` dedup, pure over an edge map).
- [ ] **Step 4 — run, expect PASS.**
- [ ] **Step 5 — failing test: diamond.** `D→{B,C}, B→A, C→A`: assert A appears once, before B/C, before D.
- [ ] **Step 6 — run (PASS if dedup correct; else fix), then commit** `#95 M1: layer resolver (topo+dedup)`.
- [ ] **Step 7 — failing test: cycle → error** (parity with bash circular guard). Implement, pass, commit.
- [ ] **Step 8 — failing test: deps parsed from `construct/deps` only** (a `Manifest.deps` populated from the deps file; `go.mod` ignored). Implement parser, pass.
- [ ] **Step 9 — dep-model rule (verified against live source).** weave discovers layer edges *only* from `construct/deps`; it never infers them from `go.mod` **or `go list -m all`**. Per-repo `go.mod` `replace`s are judged by nature: one the Go build needs (imports upstream *packages* / `go tool <upstream>`) **stays**; a discovery-only one is **migrated** (not merely deleted). **Verified:** ariadne/nous have no `replace`; brain has `replace nous => ../nous` with no `.go`/imports/`tool` → **vestigial for the Go build, but LOAD-BEARING for layer discovery**: brain's `construct/deps` names only `substrate ../ariadne`, so that `replace` is the *sole current source* of the nous→brain edge (`setup.sh:232-238`). Removing it is therefore a **required, edge-preserving migration** — add `substrate ../nous` to brain's `construct/deps` *first*, then drop the `replace`; brain's golden-diff only reaches clean once the substrate row lands (M5 must assert the nous-edge divergence **closes**, not re-ledger it as accepted). **Source 2 (`go list -m all`):** `discover_ancestors` also walks `go list -m all` for code-imported deps; weave does **not** consult it — verified no current repo gains a layer edge there (module-cache dirs lack `construct/base.manifest`), stated so the channel isn't silently dropped.
- [ ] **Step 10 — `sdlc milestone-close --issue 95 --milestone M1`.**

**Done when:** resolver passes topo/diamond/cycle tests; deps come only from `construct/deps`; ARCH-DRY cited (ported `discover_ancestors`).

### M2 — Core: intent model + Planner + `prose`→AGENTS.md + golden-diff harness

**Goal:** The pure pipeline end-to-end for the `prose` intent, plus the harness that proves parity.

**Files:** `cmd/weave/internal/intent/intent.go(+test)`, `cmd/weave/internal/plan/{action.go,plan.go,prose.go,*_test.go}`, `cmd/weave/internal/weavefs/fs.go`, `cmd/weave/internal/plan/apply.go(+test)`, `cmd/weave/main.go`, `cmd/weave/golden_test.go`.

- [ ] **Step 1 — failing test: parse a `prose` intent** from a manifest block.
- [ ] Steps 2–4 — implement `Intent` + manifest parse; pass.
- [ ] **Step 5 — failing test: `composeProse` concatenates fragments foundation-first** (base + layer + local → one body); assert the local fragment is present (the `@AGENTS.local.md` bug, fixed structurally).
- [ ] Steps 6–8 — implement `composeProse`; pass; commit.
- [ ] **Step 9 — failing test: `Plan` lowers prose Intents → a `WriteFile{AGENTS.md, body}` Action** over in-memory `[]Layer`. Implement; pass.
- [ ] **Step 10 — failing test: `Apply` writes/symlinks against `t.TempDir()`** (FS seam; symlink replaces an existing symlink). Implement; pass; commit.
- [ ] **Step 11 — wire `cmd/weave/main.go`** (`weave` compiles cwd repo; `--dry-run` prints `[]Action`).
- [ ] **Step 12 — golden-diff harness (`golden_test.go`):** run `setup.sh` and `weave` into two temp copies of each live repo (ariadne/nous/brain/metis); diff the resulting trees; assert divergences ∈ an allow-listed ledger (e.g., `AGENTS.md` content now includes local prose; nous edge). ARCH-PURE: the harness diffs *output*, not internals.
- [ ] **Step 13 — write the divergence ledger** into `## Log` (each diff = intended-fix or regression).
- [ ] **Step 14 — `sdlc milestone-close --issue 95 --milestone M2`.**

**Done when:** `weave` produces AGENTS.md (with correct per-repo prose) for all four repos; golden-diff passes with an explained ledger; pure-core tests run mock-free.

### M3 — Skill server + `scaffold` + `tool` intents

**Goal:** Agent-agnostic skills (no `.claude/skills/`) + remaining structural intents.

**Tasks (detail at milestone-start):**
- `SkillIndex` (pure): aggregate `construct/skills/` across layers → menu + body lookup; namespaced (collision-free).
- `weave skills` / `weave skill <name>` subcommands; compile the menu into AGENTS.md (always-on discovery), bodies served on demand.
- `scaffold` intent → `Mkdir` Actions (port `create_scaffold`).
- `tool` intent → `GoMod` integration (port `ensure_go_tool_dependency`); the one `exec` seam.
- Extend golden-diff to skills/scaffold/tool outputs.
- `sdlc milestone-close --milestone M3`.

**Done when:** `weave skills`/`skill` serve from `construct/skills/` with no `.claude` dependency; scaffold + tool reach golden parity.

### M4 — `settings` backend + backend-interface seam

**Goal:** The one non-floor backend, behind the seam that later backends will reuse.

**Tasks (detail at milestone-start):**
- `mergeSettings` (pure): port `merge-settings.sh` semantics exactly (deep-merge, `$merge_keys` union, `$remove`, strip meta) + table tests vs its output.
- Define the `Backend` interface (`Lower(intent) []Action`); the floor (prose/skill) + a `claude` backend (settings.json) as its first two implementors. `.claude/rules` + native-skill backends explicitly deferred.
- `settings` intent → `claude` backend `WriteFile{settings.json}`; no-op + warn on backends lacking the surface (thin portable core = env only).
- Golden-diff settings.json.
- `sdlc milestone-close --milestone M4`.

**Done when:** settings.json matches `merge-settings.sh` output (golden); the backend seam has ≥1 real client; ARCH-DRY cited (ported merge semantics).

### M5 — Cutover

**Goal:** Make weave the entrypoint; retire setup.sh + vendor; verify in anger.

**Tasks (detail at milestone-start):**
- `bootstrap.sh` builds + invokes `weave` (replace the `setup.sh` handoff); `make bootstrap`/`refresh` → `weave`.
- Apply the brain `substrate ../nous` migration (M1 step 9); add explicit `substrate` lines wherever `go.mod`-replace was doing layer discovery.
- Delete `construct/setup.sh`, `merge-settings.sh`, `sync-local-skills.sh` (logic now in weave); retire vendor mode + the stale `.ariadne-mode` markers.
- **Verify:** golden-diff ledger clean (all divergences explained); **fresh Claude session in brain loads brain's own local prose, not ariadne's** (the bug's real-world proof — manual, can't be unit-tested); `sdlc state` clean across repos.
- `weave` re-run on all live repos (ariadne/nous/brain/metis) idempotent.
- `sdlc milestone-close --milestone M5`, then `sdlc close --issue 95`.

**Done when:** setup.sh gone; weave drives all repos; fresh-session prose check passes; vendor removed.

---

## Notes
- **Worktree:** implementation runs under `sdlc change-code --worktree=yes` (isolates the in-progress compiler from the live repos the golden-diff reads).
- **Moving base:** ariadne #52 is in-flight; re-check `merge-settings.sh`/`setup.sh` aren't changed under us before M2/M4 golden snapshots.
- **Open design details** (resolve in-milestone): skill-menu placement (lean: descriptions in AGENTS.md, bodies via `weave skill`); extent of the portable settings core (lean: env only).

---

## Revisions

### 2026-06-14 — plan-quality judge (sdlc change-code), VERDICT: FAILURE → addressed

The judge blocked with three findings; all verified against live source, all incorporated:

- **(A) Intent vocabulary was incomplete.** The real `base.manifest` is ~all generic `symlink/seed/scaffold/touch/merge/tool`; `prose`/`skill` aren't manifest verbs today. **Delta:** the `Intent` kind is now a **hybrid** — ported file-op verbs + new `Prose`/`Skill` semantic intents (see Core concepts). The "rises from file-ops to intents" framing is corrected to "adds semantic intents atop ported file-ops."
- **(B) `go.mod` dep-model premise was imprecise + carried a stated regression risk.** Verified: ariadne/nous have no `replace`; brain's `replace nous` is **vestigial** (no `.go`, no imports, no `tool`). **Delta:** M1 step 9 now states the precise rule (edges from `construct/deps` only; build-necessary `replace`s stay, discovery-only ones go) and records that no build-regression surface exists today.
- **(C) `prose` grammar undefined.** **Delta:** `composeProse` now defines `prose <relpath>`, foundation-first composition, and the `symlink AGENTS.md` → composed-real-file flip as an expected golden-diff divergence.

Non-blocking: judge flagged `estimate_hours: 28` as tight (clean-sheet compiler + skill server + settings port + golden harness + 4-repo cutover); left as-is — `actual` is measured at close, not re-typed.
