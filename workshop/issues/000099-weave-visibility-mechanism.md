---
id: 000099
status: working
deps: []
github_issue:
target: weave-composition-algebra
created: 2026-06-14
updated: 2026-06-14
estimate_hours: 5
---

# weave: export/internal visibility mechanism

Implements the visibility axis of [[weave-composition-algebra]] — `𝒜(R) = all-exports(L₀..Lₙ) ⊎ leaf-internals(Lₙ)`. THE core mechanism: without it, a derivative inherits an ancestor's *internal* artifacts (the parley bug) and never composes its own.

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
- `build`/`test`/`vet`/`gofmt` clean; honors [[weave-composition-algebra]].

## Plan

- [ ] grammar + `Intent.Visibility` (parse, default export) — TDD
- [ ] walk carries visibility per intent + prose fragment
- [ ] planner selection `𝒜(R)`; prose leaf-internal last
- [ ] ariadne `base.manifest`: `export prose AGENTS.base.md` + `internal prose AGENTS.local.md`
- [ ] golden + verify-complete visibility-aware
- [ ] multi-layer prose fixture + parley temp-copy re-verify

## Log

### 2026-06-14
- Split from ariadne#95 M5: the prose-composition bug surfaced on the parley tart pass is a symptom of composing without a visibility axis. Operator chose to build the general mechanism now (it is THE core mechanism) rather than a point fix, fully explicit (no AGENTS.local.md convention). Algebra captured in target [[weave-composition-algebra]].
