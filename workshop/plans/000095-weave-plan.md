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
- **`Manifest`** — parsed `construct/base.manifest` for one layer: `[]Intent` only **(M2)**. Layer *edges do NOT live here* — `base.manifest` has no deps header; edges come from `construct/deps`, parsed by `ParseDeps` (`cmd/weave/internal/layer/deps.go`, shipped M1). `lib-deps.sh` is the grammar source of truth.
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
- **Verify:** golden-diff ledger clean (all divergences explained) **+ an independent completeness check** (enumerate `setup.sh`'s *full* output — run it into a temp copy + diff the whole tree, or enumerate the live repo's full managed set — and assert weave covers every non-deferred path; closes the M2-review I-1 under-production blind spot, since `weave golden` alone only validates paths weave *plans*); **fresh Claude session in brain loads brain's own local prose, not ariadne's** (the bug's real-world proof — manual, can't be unit-tested); `sdlc state` clean across repos.
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

### 2026-06-14 — directory-agnostic substrate paths (testing/rollout enabler)

Confirmed against `lib-deps.sh:deps_substrate_targets`: `construct/deps` substrate edges **already** resolve repo-root-relative for *any* path (relative or absolute, `raw="$repo_root/$target"`) — there is no `../<name>` naming assumption in the resolution. And build-in-owner (`make sdlc-build` → `dev-aliases.sh --list`) finds the owner *among those same `construct/deps`-discovered peers*. So pointing a consumer's `construct/deps` at an arbitrary ariadne checkout drives **both** the layering symlinks and the binary build — no shell repointing needed (this obviates the "repoint the binary chain" step). Deltas:

- **Add a `weave depend-on <path>` verb** that records `substrate <path>` **verbatim** (replaces the `tool` action's hardcoded `substrate ../ariadne`), so a test setup captures the real path it was given.
- **M2's IO-walk must port `deps_substrate_targets` faithfully** — repo-root-relative resolution, absolute-path support, present-skip (an absent peer still resolves syntactically; the caller decides skip-vs-clone). The M1 pure parser already returns relpaths verbatim; the *walk* is the part that must match.
- **Auto-clone of an *absent* substrate stays in the shell bootstrap** — it needs a git-URL convention `construct/deps` doesn't carry (only `data` rows do); irrelevant when peers are placed manually (the test schemes).
- **Adopt the side-by-side migration as the integration test** (added to Done-when): clone a real derivative (e.g. parley) + the ariadne worktree anywhere, `weave depend-on <ariadne-path>`, migrate it off `setup.sh`, both on branches — production repos untouched. Stronger than golden-diff alone.

Test/rollout progression (validated as enabled by the above): (1) self-walk in the ariadne worktree, golden-diff vs setup.sh + a live session; (2) spawn a derivative via `weave depend-on`; (3) the side-by-side migration above.

### 2026-06-14 — M1 close: reconcile plan to delivered code (review FIX-THEN-SHIP)

M1 milestone-review verdict **FIX-THEN-SHIP** — code correct/pure/tested, no Critical/Important *code* issues; the fixes are plan/doc reconciliations:

- **P-1 (ordering fidelity).** `Resolve` is DFS post-order — a *valid* foundation-first topo order with dedup — and does **NOT** bit-reproduce `discover_ancestors`' BFS-then-reverse, which mis-orders a foundation that's also a direct dep. Worked example: post-migration brain (`substrate ../ariadne` + `../nous`, `nous→ariadne`) — setup.sh applies **nous before ariadne** (quirk); weave applies **ariadne first** (correct). **Pre-registered M5 golden-diff ledger entry:** brain layer-application order differs — *expected, weave-is-correct*, not a regression. `resolve.go` doc softened; ARCH-DRY anchor is "ports the *intent* (foundation-first + dedup)," not bit-ordering.
- **P-2 (Core-concepts correction).** Layer edges come from `construct/deps` (parsed by `ParseDeps` in `cmd/weave/internal/layer/deps.go`, shipped M1) — **not** a `base.manifest` "header" (it has none; `lib-deps.sh` is the source of truth). `intent/manifest.go` is **M2-only** and parses base.manifest *intents*. M1 "Files" shipped as `layer/{resolve,deps}{,_test}.go` (`layer.go` deferred to M2).
- **M2 carry-forward (judge's architectural notes).** The deps *walk* (IO seam) must reproduce `deps_substrate_targets`' repo-root-relative + absolute-path + present-skip resolution **and** the two `_seen_or_add` filters (base.manifest-existence, target-self-exclusion) that pure `Resolve` rightly omits. `Resolve` emits `root` **last (self-included)**; the Planner must account for root-is-last-and-self.

### 2026-06-14 — M2 close: golden-harness reconciliation + carry-forwards (review FIX-THEN-SHIP)

M2 boundary-review verdict **FIX-THEN-SHIP** (no Critical). Code fixes landed (`6a70aac`: hermetic transitive+diamond walk tests [I-2]; gofmt; `ParseManifest` comment). Plan reconciliations:

- **P-1 (golden harness — design changed + a blind spot).** M2 shipped the golden-diff NOT as the planned `golden_test.go` running both `setup.sh` and `weave` into temp copies and diffing trees, but as a **pure `internal/golden` classifier + a `weave golden` subcommand** that classifies *weave's planned actions* against the **live** repos' current state (live = setup.sh's output). Trade-off: lighter, no hermetic `setup.sh` run — **BUT it cannot detect *under-production*** (a path setup.sh produces that weave never plans is invisible: no action → not classified → not flagged). So the M2 ledger proves weave-produces-*correctly* (verified clean on ariadne-self + metis), not weave-produces-*completely*. **→ M5 MUST add an independent completeness check** (below) before "golden-diff ledger clean" means what the Spec wants. Also: the 6 nous/brain UNEXPECTED are stale-derivative drift the harness surfaces correctly but can't auto-distinguish from a weave bug.
- **P-2 (scope).** `scaffold` + `touch` were forward-pulled from M3 to M2 (needed for the M2 golden-diff to be meaningful against the ~all-file-op live manifests). `create_scaffold`'s `.gitkeep` parity is deferred (`applyMkdir` TODO; currently invisible to the harness, which observes only the dir target).
- **M3/M4/M5 carry-forwards (review §6).** (a) `Action` now fans out across 5 switches (Plan/Apply/Gather/classifyAction/formatActions) — a new kind touches all 5; keep the `default` omission-guards. (b) **M4:** wire the `golden` classifier case in the SAME milestone `Merge`/`Tool` lowering lands, or M4's golden run spuriously goes UNEXPECTED. (c) **M5:** close the P-1 completeness blind spot; the brain `substrate ../nous` migration must make the nous-edge divergence *close* (weave starts planning nous's layer for brain), not get re-ledgered as accepted.

### 2026-06-14 — M3 close: go.mod-parse consolidation + skill-golden deferral (review FIX-THEN-SHIP)

M3 boundary-review verdict **FIX-THEN-SHIP** (no Critical; the same-milestone classifier wiring was praised as nailing the M2 carry-forward). Fixes landed (`b4843e1`): consolidated the duplicated `moduleLine`/`gomodHasTool`/`goDirective` go.mod parsing into a new **pure `cmd/weave/internal/gomodx`** package (ARCH-DRY — go.mod reasoning was spread across plan/golden/weavefs; now one home, no import cycle) + refreshed stale `action.go` docs. Plan notes:

- **Skill-menu golden coverage deferred to M5.** M3's task line said "extend golden-diff to skills/scaffold/tool"; `tool` is delivered + classified and `scaffold` was forward-pulled to M2, but the **skill menu is intentionally NOT golden-gated** (`weave golden` passes a nil menu) because weave's skill mechanism (menu in `AGENTS.md` + `weave skill`) diverges from setup.sh's `.claude/skills/` symlinks. Skill-serving parity is an M5 concern.
- **M5 carry-forwards (review §6):** (a) the owner-vs-derivative `tool` classification hinges on byte-identical absolute paths between the resolved layer and the canonicalized root — **M5 must assert path canonicalization stays consistent across ariadne/nous/brain/metis** (else an owner is misclassified as a derivative); (b) the skill-menu golden gate (above).

### 2026-06-14 — M4 close: settingsx package + no formal Backend interface (review FIX-THEN-SHIP)

M4 boundary-review verdict **FIX-THEN-SHIP** (no Critical; the judge differentially ran live `merge-settings.sh` against the Go tests — byte-identical on the `$remove`-before-union case). Fixes are plan-sync (the code is correct):

- **(a) `mergeSettings` lives at `cmd/weave/internal/settingsx/settingsx.go`, NOT `plan/settings.go`** (Core-concepts table). Required: a leaf package below both `plan` and `golden` (which already imports `plan`) to avoid an import cycle. The table row is **superseded**.
- **(b) The `Backend` interface (`Lower(intent) []Action`) was NOT built** (Spec "backend-interface seam"; the M4 step; the seam done-when). With a single non-floor backend, the `Action` sum type + type-switch dispatch **is** the seam — YAGNI; a one-implementor interface is premature abstraction (the M2 review already endorsed the `Action` fan-out). The "Define the Backend interface" step + the seam done-when are **superseded**.
- **M5 note (review §6): `settings.json` byte-churn.** weave's output is *semantically* equal to `merge-settings.sh`'s but **byte-different** (Go `json.MarshalIndent` HTML-escapes `<>&`; python escapes non-ASCII). The M5 cutover's `git diff` of `settings.json` will show cosmetic churn that is a semantic no-op — the M5 ledger must say so, not read it as a regression.
- **Minor carried:** `gather_test.go`'s "regardless of order" comment overclaims — `observeMerge` relies on `base.manifest` ordering (`symlink` before `merge`); soften or make order-independent (low-value).

### 2026-06-14 — M5 test runbook (approved)

Cutover tested fully on `#95` branches in tart VMs, **isolated from production `main`**; coordinated pause-the-world merge only after all pass. Order: ariadne → parley.nvim → pair → nous → brain (root → leaf → leaf → layer-stack → encrypted-chain). metis excluded (re-init later); ariadne `#31` + `nous-14` worktrees = the final step.

- **Phase 0 [done]:** all descendants parked clean-on-`main`; ariadne on `000095-weave`.
- **Phase 1 [implement; review-before-commit]:** (1a) ariadne self-cutover migration; (1b) completeness-check harness — independently enumerate `setup.sh`'s *full* output vs weave (closes the M2-review under-production blind spot, since `weave golden` alone only validates paths weave *plans*).
- **Phase 2 [per-repo tart, manual sign-off]:** each repo on its `#95` cutover branch → ariadne@#95 → `make tart` → in the VM verify: (1) weave compiles clean + **idempotent**; (2) golden-diff = only *expected* divergences, **zero UNEXPECTED**; (3) **completeness**; (4) **★ a fresh session loads the repo's OWN `AGENTS.local.md` prose, not ariadne's**; (5) skills menu + serve; (6) `settings.json` semantic-correct (byte-churn ok); (7) `sdlc`/`make`/`tart` toolchain intact.
- **Phase 3 [pause-the-world merge, after all pass]:** merge `#95` branches base-first; re-weave all descendants on `main`; post-merge smoke.
- **Phase 4 [cleanup]:** re-weave `#31` + `nous-14`; re-init metis.

gcrypt is orthogonal (weave works on brain's *decrypted* working tree). `settings.json` is semantic-equal but byte-different (cosmetic cutover churn, not a regression).

### 2026-06-14 — M5 Phase 1 done: explicit `weave compile --target` + backend-targeting (Approach 1)

Phase 1 (ariadne self-cutover + completeness harness) is complete on `000095-weave`; two design decisions evolved the M5 surface:

- **Skill lowering is now per-backend-target, not union.** M3/M4 emitted one artifact set carrying BOTH skill faces (the `.claude/skills` symlinks AND the AGENTS.md `## Skills` menu). Operator call (Approach 1): since weave runs *before* an agent (`weave compile --target <t>; <agent>`), it compiles **one target per invocation** — `claude` → `.claude/skills` symlinks + **prose-only** AGENTS.md (no menu); `codex`/`agy` → AGENTS.md **with** the menu + **no** `.claude/skills`. Menu and symlinks are mutually exclusive by construction. Implemented as `plan.Target` (string enum + `EmitSkillSymlinks()`/`IncludeSkillMenu()` predicates); `plan.Plan` stays target-agnostic (composes whatever menu it's handed — nil for claude). **Rejected Approach 2** (one set + a "Claude ignore / codex read `AGENTS.skills.md`" conditional) — it routes via soft prose-obedience, against the deterministic-shell principle. Consequence: artifacts are compiled-per-target (regenerated per launch); switching agents re-runs weave; concurrent same-dir agents conflict (use worktrees) — acceptable, since concurrent same-dir agents already conflict on the *work*.
- **`weave` → `weave compile`.** The root command no longer mutates (bare `weave` prints help — closes the accidental-apply footgun that clobbered `AGENTS.md` earlier). `compile` carries `--dry-run` + `--target` (default `claude`); `golden`/`verify-complete` inherit `--target` (ARCH-DRY — same `planActions`); `skills`/`skill` stay target-agnostic.
- **seed implemented; P-1 completeness blind spot closed.** The deferred `seed` verb now lowers (`plan.Seed`, ports `create_seed`: create-if-missing / refresh-on-drift / no-op / non-fatal-missing-source) — the deferred ledger is now **empty** (every setup.sh verb lowers + classifies). `weave verify-complete` (the P-1 completeness check) reports **zero under-production** on ariadne self-walk + nous derivative, for both targets (a skill intent is covered by *either* backend). seed self-filters on the ariadne self-walk (source==target); it manifests on derivatives (nous: both seeds MATCH).
- **Invocation:** `make refresh` → `weave compile --target claude` (the apply). Retired the `sync-local-skills.sh` SessionStart hook, **not** re-added — `make refresh`/`bootstrap` are the render triggers; the auto-refresh-hook question is deferred to before the merge.

Commits on `000095-weave`: `ac7dc2e` (run-wiring + skill backends), `63449bf` (config migration), `b0bb835` (test retirement), `e5c5f19` (seed lowering), `d1b30c1` (completeness harness), `5530a0b` (`weave compile --target`). M5 milestone-close happens after the coordinated merge (Phase 3).
