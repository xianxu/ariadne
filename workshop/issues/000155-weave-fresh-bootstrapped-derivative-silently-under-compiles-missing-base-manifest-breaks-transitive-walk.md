---
id: 000155
status: open
deps: []
github_issue:
created: 2026-07-01
updated: 2026-07-01
estimate_hours:
---

# weave: fresh-bootstrapped derivative silently under-compiles (missing base.manifest breaks transitive walk)

## Problem

Bootstrapping a **new** derivative the natural way — `mkdir foo && cd foo && weave link <parent> && weave compile` — produces `construct/deps` but **no `construct/base.manifest`**. The layergraph walk (`cmd/weave/internal/walk/walk.go`, `pkg/layergraph`) treats a candidate as a traversable layer **only if it ships `construct/base.manifest`** (`discover_ancestors`' filter). So a manifest-less repo is invisible as a layer, with two silent failure modes:

1. **A manifest-less intermediate breaks the whole chain.** A downstream consumer's `weave compile` reaches the manifest-less parent, doesn't recognize it as a layer, and **never recurses to the foundation** — silently emitting only the appended `EnsureGitignore` action: `weave: applied 1 action(s)`. No error, no warning. No `Makefile`, no composed `AGENTS.md`, no skills.
2. Even a leaf fresh-bootstrap silently omits its own `internal prose AGENTS.local.md` until a base.manifest exists.

**Observed 2026-07-01** building the 3-level chain `kbench → kaggle → metis → ariadne` (nous/kaggle-ml-base-layer). `metis` compiled to 97 actions *for itself* (it's the root; ariadne is its direct substrate), but `kaggle` and `kbench` compiled to **1 action each** — diagnosed only by noticing the missing `Makefile`/`AGENTS.md`, not by any tool output. Root cause: `weave link` establishes the `substrate` edge but not the `base.manifest` that marks a repo as a layer; and the walk **silently skips** a declared substrate that lacks one. Every existing derivative (nous, pair, you-decide, 42shots) has a hand-authored `base.manifest` from its #95 cutover branch, which is why this never surfaced — no one had fresh-bootstrapped a brand-new derivative (let alone a multi-level chain) since weave landed.

## Spec

Fix the footgun. Options (combine 1 + 2 preferred):

1. **Loud failure (minimum).** `weave compile` must error (or at least warn hard) when a declared `substrate <path>` resolves to a directory that lacks `construct/base.manifest` — e.g. `substrate ../metis is not a compilable layer (missing construct/base.manifest)`. A silent 1-action compile that drops the entire base layer should never be possible. Natural home alongside the existing `verify-complete` under-production check.
2. **Seed on establish.** `weave link <path>` (or a dedicated `weave init`/`weave new`) scaffolds a minimal `construct/base.manifest` stub — a header comment + `internal prose AGENTS.local.md`, matching what every #95-cutover repo ships — so a fresh derivative is a valid, traversable layer out of the box.
3. **Consider** whether the walk should traverse *through* a manifest-less intermediate (pass-through layer contributing nothing but still recursing to ancestors). This may be intentionally disallowed because the base.manifest also drives internal-prose selection; if so, (2) is the correct fix and (3) is a no.

## Done when

- A fresh `mkdir foo && weave link ../bar && weave compile` either (a) produces a valid layer because `base.manifest` was seeded, or (b) fails loudly with an actionable message naming the missing file.
- A ≥3-level substrate chain compiles fully (foundation reached) without any hand-authored `base.manifest`.
- Regression test covers the multi-level-chain walk and the manifest-less-substrate error path.

## Plan

- [ ]

## Log

### 2026-07-01

Filed from the kaggle-ml-base-layer bootstrap (metis/kaggle/kbench, in nous). **Workaround already applied** in those three repos: hand-authored a minimal `construct/base.manifest` (`internal prose AGENTS.local.md` + a header noting that shipping the manifest is what marks the repo as a traversable layer) in each; all three then compiled to 97 actions and the transitive chain composed correctly (kbench `AGENTS.md` carries the ariadne Constitution). This ticket is to fix the *tooling* so the next fresh derivative doesn't hit the silent failure.
