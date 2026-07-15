---
id: 000099
status: done
deps: []
github_issue:
target: base-layer-mechanics
created: 2026-06-14
updated: 2026-06-16
estimate_hours: 5
actual_hours: 0.56
---

# weave: export/internal visibility mechanism

Implements the visibility axis of [[base-layer-mechanics]] — `𝒜(R) = all-exports(L₀..Lₙ) ⊎ leaf-internals(Lₙ)`. THE core mechanism: without it, a derivative inherits an ancestor's *internal* artifacts (the parley bug) and never composes its own.

## Problem

weave composes `prose` (and will compose `skill`/`settings`) across the layer DAG, but has **no visibility axis**: every `base.manifest` row is implicitly inherited, and `loadLayer` reads each `prose` fragment from the *declaring layer's* directory. So `prose AGENTS.local.md` (declared in ariadne's manifest) always pulls **ariadne's** local into every consumer, and a manifest-less leaf contributes nothing of its own. Empirically (ariadne#95 M5 parley tart pass): parley's composed `AGENTS.md` = ariadne-base + **ariadne's** local, with parley's own local missing — the `@AGENTS.local.md` bug reproduced one level down. `weave golden` missed it because it diffs the composed file against the live *symlink* (an intended divergence) and never inspects composed *content*.

## Spec (fully explicit — "B1")

**Grammar** (`intent.ParseManifest`, backward-compatible): a row becomes `[export|internal] <type> <files…>`; **visibility defaults to `export`** (every existing row unchanged = an export). The only new token is a leading `internal`. `type` is the existing kind (`prose|skill|merge|symlink|seed|scaffold|touch|tool`) — it picks the compose operator; visibility picks the operands.

**`Intent.Visibility`** — add `Export|Internal` to the typed intent; the walk associates visibility with each intent **and each resolved prose fragment**.

**Planner selection** — compiling repo R over `⟨L₀..Lₙ=R⟩` selects `𝒜(R) = { exports of every Lᵢ } ⊎ { internals of Lₙ }`. An ancestor's internals are excluded; the leaf's internals + all exports are included. prose then composes foundation-first with the leaf's internal **last** (per the algebra).

**Fully-explicit, no convention:** each repo declares its own `internal prose AGENTS.local.md` in **its own** `base.manifest`. ariadne's manifest: `prose AGENTS.base.md` → `export prose AGENTS.base.md`; `prose AGENTS.local.md` → `internal prose AGENTS.local.md`. Each derivative (parley, brain, nous, metis…) gets `internal prose AGENTS.local.md` added to its own manifest **on that repo's #95 cutover branch** (a manifest-less leaf gains a minimal `base.manifest`). This issue ships the mechanism + **ariadne's** annotation; the per-derivative rows land in their tart-pass prep.

**golden + verify-complete** become visibility-aware: a derivative's expected composition = ancestors' exports + its own internal.

## Done when

- Compiling any repo yields **ancestors' exports + its own internal, never an ancestor's internal** (multi-layer fixture: foundation `export prose` + foundation `internal prose` + leaf `internal prose` → leaf composes foundation-export + leaf-internal only).
- parley's composed `AGENTS.md` (temp-copy, with `internal prose AGENTS.local.md` declared) contains `# Parley.nvim Local Extensions` and **NOT** `# Ariadne Workshop Extensions`.
- ariadne self-walk unchanged (ariadne is the leaf → gets its own internal local).
- `build`/`test`/`vet`/`gofmt` clean; honors [[base-layer-mechanics]].

## Plan

- [x] grammar + `Intent.Visibility` (parse, default export) — TDD
- [x] walk carries visibility per intent + prose fragment
- [x] planner selection `𝒜(R)`; prose leaf-internal last
- [x] ariadne `base.manifest`: `export prose AGENTS.base.md` + `internal prose AGENTS.local.md`
- [x] golden + verify-complete visibility-aware
- [x] multi-layer prose fixture + parley temp-copy re-verify

## Log


- 2026-06-16: closed — Retroactive cleanup close — the export/internal visibility mechanism shipped 2026-06-14 (818daff5, on main): Intent.Visibility grammar (default export), walk carries per-intent visibility, planner 𝒜(R) selection (ancestors export ⊎ leaf internal), ariadne base.manifest export/internal prose rows, visibility-aware golden + verify-complete; all 6 plan items done + TDD-covered. Now load-bearing + exercised end-to-end by #104 (SelectVisible reuses the same intent.Selected; the internal construct skill + xx-construct). actual 0.56h is the commit-anchored measure of the 06-14 work (estimate 5h was high).; review verdict: not-run
### 2026-06-14
- Split from ariadne#95 M5: the prose-composition bug surfaced on the parley tart pass is a symptom of composing without a visibility axis. Operator chose to build the general mechanism now (it is THE core mechanism) rather than a point fix, fully explicit (no AGENTS.local.md convention). Algebra captured in target [[base-layer-mechanics]].

### 2026-06-14 — implemented (TDD, single-pass)
- **Grammar** (`intent`): `ParseManifest` consumes an OPTIONAL leading `export|internal` token (disjoint from the verb set → unambiguous); default `Export` (the zero value, so every pre-#99 row is unchanged). Added `Visibility` (`Export|Internal`) to `Intent`. Examples: `export prose AGENTS.base.md` / `internal prose AGENTS.local.md`.
- **Visibility through walk→plan**: `Layer.ProseFragments` changed `[]string` → `[]ProseFragment{Visibility, Content}`; `walk.loadLayer` tags each fragment with its intent's visibility. The pure planner selects 𝒜(R) — the leaf Lₙ is the last layer (Resolve emits root last); prose composes `[export-prose of all layers, foundation-first] ++ [internal-prose of leaf]`; the SAME export-or-leaf filter (`participates` → `intent.Selected`, one source of truth, ARCH-DRY) applies uniformly to every kind (file-ops are all export today → behavior-preserving). `verify-complete`'s `CheckCompleteness` reuses `intent.Selected` to SKIP an ancestor's internal (excluded, not under-produced); the leaf's own internal is still checked. golden recomputes AGENTS.md via the same visibility-aware `plan.Plan`, so the self-walk MATCH is preserved structurally (no golden code change for content).
- **ariadne `base.manifest`**: `prose AGENTS.base.md` → `export prose AGENTS.base.md`; `prose AGENTS.local.md` → `internal prose AGENTS.local.md` (+ a comment block explaining the axis). ariadne's local now lands ONLY on its own self-walk.
- **Multi-layer fixture (the 𝒜(R) proof)**: `plan.TestPlanProseVisibilitySelection` + the end-to-end `main.TestCompileMultiLayerVisibility` (foundation export+internal, middle export, leaf internal, compiled at the leaf via real Walk→Plan→Apply) assert the leaf composes `FOUNDATION-EXPORT ∥ MIDDLE-EXPORT ∥ LEAF-INTERNAL` and does NOT contain either ancestor's internal. PASS.
- **ariadne self-walk unchanged (proven byte-identical)**: HEAD binary + original manifest vs. this branch's binary + visibility-annotated manifest both compose AGENTS.md = 11143 bytes, md5 `cf25b47ea6b002cbb696b0ea2804025f`, `diff` clean. (The golden "content drift live 10784 vs 11143" is the pre-existing staleness of the committed/live AGENTS.md vs. the current AGENTS.base.md edit — present with the baseline binary too, NOT a #99 regression. The other two UNEXPECTED — merge settings.json, go.mod tool directive — also pre-date #99.)
- **parley temp-copy re-verify (the bug fixed)**: `cp -R parley.nvim → $TMPDIR/parley-vis`; wrote `construct/base.manifest` = `internal prose AGENTS.local.md`, `construct/deps` = `substrate /Users/xianxu/workspace/ariadne`, `rm AGENTS.md`, compiled with the #99 binary. Assertions: composed AGENTS.md **CONTAINS** `# Parley.nvim Local Extensions` (PASS) and does **NOT contain** `# Ariadne Workshop Extensions` (PASS); ariadne's EXPORTED base constitution still flows down (PASS). The #95 bug is fixed.
- **Verify**: `go build ./... && go test ./cmd/weave/... && go vet ./cmd/weave/... && gofmt -l cmd/weave` all clean.
