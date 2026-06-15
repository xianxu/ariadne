---
type: target
slug: weave-composition-algebra
status: active
created: 2026-06-14
updated: 2026-06-14
sources:
  - "ariadne#95 — weave cutover; the prose-composition bug surfaced on the parley tart pass (Phase 2)"
---

# Target: Weave composition algebra — how artifacts merge across a layer DAG

weave compiles a repo's agentic context by **composing artifacts across its layer DAG**. The durable commitment this target defends: that composition is not ad-hoc per artifact type — it is one model, **select-then-fold**, with two axes that must hold for *every* artifact type weave grows.

1. **A visibility axis.** Each artifact a layer declares is either `export` (it flows down to consumers) or `internal` (it stays with the declaring repo). A repo compiles from *every ancestor's exports* plus *its own internal* artifacts — and **never** an ancestor's internal artifacts.
2. **A per-type compose operator.** Each artifact type (`prose`, `skill`, `settings`, file-ops, `tool`) has its own algebra for folding the selected artifacts across the DAG: prose concatenates, skills union, settings deep-merge. The **type** picks the operator; the **visibility** picks the operands.

Getting this math right *is* the point of weave. If composition is wrong, a derivative silently inherits the wrong context — exactly the bug that motivated the rewrite (see *Why now*). This target is the reference the implementation (ariadne#95 and successors) must honor, and the home where each artifact type's formula lands as it crystallizes.

## Why now

`setup.sh` carried prose via a *symlinked* `AGENTS.md` whose `@AGENTS.local.md` import silently resolved to **ariadne's** local file in every derivative — so a downstream repo never loaded its own local prose. weave was meant to fix that. But the first derivative tart pass (parley, ariadne#95 M5) showed weave **reproduced the same bug one level down**: it read each `prose` fragment from the *declaring layer's* directory, so `prose AGENTS.local.md` (declared in ariadne's manifest) always pulled ariadne's local, and a manifest-less leaf contributed nothing of its own. parley's composed `AGENTS.md` came out as ariadne-base + ariadne-local, with parley's own local missing.

The fix is not a prose patch — it is recognizing that **composition needs a precise, type-uniform model**, because the same "who contributes what to whom" question recurs for skills, settings, and every future artifact type. The bug was a symptom of composing without a stated algebra. We are committing to the algebra; `link`/`unlink` (module include/exclude) and the export/internal axis are its surface.

## The model

**Linearize the DAG.** For a target repo R, `Resolve` walks the `substrate` edges (`construct/deps`) into a deduplicated, foundation-first order `⟨L₀, L₁, …, Lₙ⟩` — L₀ the foundation (ariadne), `Lₙ = R` (the leaf being compiled). A diamond collapses to one application per layer (single on-disk copy; no versioning).

**Select by visibility.** Each layer Lᵢ declares a manifest of artifacts, each tagged `export` (default) or `internal`. The multiset that participates in compiling R:

```
𝒜(R) =  ⊎ᵢ { a ∈ Lᵢ : visibility(a) = export }      — every layer's exports
      ⊎    { a ∈ Lₙ : visibility(a) = internal }     — the leaf's internals only
```

The invariant: **an ancestor's `internal` artifacts never reach R; R's own `internal` always do; all ancestors' `export`s do.** (The parley bug was `𝒜(R)` computed wrong — ariadne's internal leaked in, parley's own dropped.)

**Compose by type.** Partition `𝒜(R)` by type; fold each type's artifacts in layer order (foundation-first, the leaf's internal last) with that type's operator `⊕ₜ`.

### prose — clarity: HIGH (this is what stabilized)
Ordered concatenation, foundation-first, the leaf's own local last. Non-commutative, associative, collision-free (concatenate, never overwrite).
```
prose(R) = ⟦export-prose(L₀)⟧ ∥ ⟦export-prose(L₁)⟧ ∥ … ∥ ⟦export-prose(Lₙ)⟧ ∥ ⟦internal-prose(Lₙ)⟧
           ∥ = join with a blank line
```
Floor case (2 layers): `prose(parley) = AGENTS.base.md(ariadne, export) ∥ AGENTS.local.md(parley, internal)`. ariadne's own `AGENTS.local.md` is `internal` to ariadne, so it is never in a derivative.

### skill — clarity: HIGH
Namespaced union; commutative, idempotent, collision-free by prefix (`xx-`/`nous-`/`metis-`).
```
skills(R) = ⋃ᵢ export-skills(Lᵢ)  ∪  internal-skills(Lₙ)        (keyed by namespaced name)
```
The *composition* is target-independent; only the *lowering* differs (claude → `.claude/skills` symlinks; codex/agy → the AGENTS.md menu).

### settings — clarity: MEDIUM (formalized in ariadne#97; the DAG fold is not yet shipped)
Deep-merge with per-key semantics, foundation-first, specific-over-general; `$merge_keys` carried through, meta stripped once at the end.
```
settings(R) = stripMeta( foldl(deepMerge, ⟨export-settings(L₀), …, export-settings(Lₙ), internal-settings(Lₙ)⟩) )
deepMerge(b,l): dicts recurse · $merge_keys arrays union (b, then new l) · $remove filters before union · other arrays: l replaces · scalars: l wins
```
Today only a two-input merge (ariadne-base + repo-local) ships; the DAG fold is ariadne#97.

### file-ops (symlink / seed / scaffold / touch) — clarity: MEDIUM
Provisioning ops keyed by target path, with the self-reference filter (a layer never provisions a file onto its own canonical source). Composition is **conflict-accumulating**, NOT silent last-writer-wins: when two layers provision the same target path from *different* sources, weave does not quietly pick a winner — it **accumulates every such collision across the whole compile and surfaces them as warnings** (an error-monad / `Validation` shape: collect, don't fail-fast, report all). A deterministic winner is still chosen (later layer / specific-over-general) so the compile proceeds, but the conflict is reported, never hidden. Collisions are rare (namespaces `xx-`/`nous-`/`metis-`) — which is exactly why they must be loud when they happen.
```
files(R)[p] = accumulate all Lᵢ provisioning p; if ≥2 with differing source → warning(p, {sources}); resolve to the latest in order and continue
```

### tool — RETIRED (no longer a weave-managed composition type, #95 M5)
Go-tool ownership is resolved by **location**, not by weave: `construct/dev-aliases.sh` scans sibling `cmd/X` dirs, and build-in-owner (#60 M2) builds each tool to `OWNER/bin/`. The substrate edge to the owner comes from `weave link` / `construct/deps` — the same mechanism every other layer dep uses — so the old `tool` row's derived `substrate` was redundant. The owner-side `go mod edit -tool` directive served only goland (which we don't use), so weave **does not edit go.mod at all**.

Consequently weave no longer lowers a `tool` intent: the verb is dropped from the manifest grammar (a stale `tool` row falls through the parser's unknown-verb skip, like the retired `copy`), there is no `ToolDep` action, and the golden harness carries no `tool`/`go.mod` divergence. There is nothing to fold here — `tool` is not a composition operator, it is a non-concern.

## What this is NOT

- **Not a versioning system.** Single on-disk copy per layer (clone-as-peer + symlink); a diamond is one application, not version reconciliation.
- **Not auto-discovery.** Artifacts are *declared* in `base.manifest`. Under the fully-explicit model, even a leaf's own local prose is an explicit `internal prose AGENTS.local.md` row in that repo's manifest — no magic filename inferred from disk.
- **Not commutative for prose.** Order is the DAG linearization; the leaf's own local is always last. Reordering changes meaning.
- **Not suppression.** v1 is additive/override-only; a derivative cannot *remove* an inherited export (deferred — see Open questions).

## Open questions

- **Per-type formulas not yet crystallized.** settings (the DAG fold — ariadne#97) and file-op collision precedence (is last-writer-wins right, or should derivative-over-foundation be explicit?) deserve the rigor prose now has. (`tool` is no longer a composition type — retired in #95 M5; ownership is location-based.)
- **Suppression.** Should a derivative be able to *retract* an inherited export (drop a skill, remove a prose fragment)? Today: no (additive/override-only). If yes, that's a third visibility-adjacent operation.
- **Intermediate-layer prose ordering.** In a 3-layer stack (ariadne→nous→brain), where does nous's *exported* prose sit relative to brain's own? Current model: foundation-first, so nous before brain's internal — but unexercised (nous exports no prose today).
- **Transitive vs direct export.** Is an `export` visible to *all* descendants or only direct consumers? Today: transitive.
