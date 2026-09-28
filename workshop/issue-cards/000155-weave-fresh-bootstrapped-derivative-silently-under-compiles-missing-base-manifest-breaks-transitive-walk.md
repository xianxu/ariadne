---
id: '000155'
status: done
started: 2026-07-06T10:00:21-07:00
created: 2026-07-01
updated: 2026-07-06
estimate_hours: 1.08
actual_hours: 1.34
---

# weave: fresh-bootstrapped derivative silently under-compiles (missing base.manifest breaks transitive walk)

## Problem

Bootstrapping a **new** derivative the natural way — `mkdir foo && cd foo && weave link <parent> && weave compile` — produces `construct/deps` but **no `construct/base.manifest`**. The layergraph walk (`cmd/weave/internal/walk/walk.go`, `pkg/layergraph`) treats a candidate as a traversable layer **only if it ships `construct/base.manifest`** (`discover_ancestors`' filter). So a manifest-less repo is invisible as a layer, with two silent failure modes:

1. **A manifest-less intermediate breaks the whole chain.** A downstream consumer's `weave compile` reaches the manifest-less parent, doesn't recognize it as a layer, and **never recurses to the foundation** — silently emitting only the appended `EnsureGitignore` action: `weave: applied 1 action(s)`. No error, no warning. No `Makefile`, no composed `AGENTS.md`, no skills.
2. Even a leaf fresh-bootstrap silently omits its own `internal prose AGENTS.local.md` until a base.manifest exists.

**Observed 2026-07-01** building the 3-level chain `kbench → kaggle → metis → ariadne` (nous/kaggle-ml-base-layer). `metis` compiled to 97 actions *for itself* (it's the root; ariadne is its direct substrate), but `kaggle` and `kbench` compiled to **1 action each** — diagnosed only by noticing the missing `Makefile`/`AGENTS.md`, not by any tool output. Root cause: `weave link` establishes the `substrate` edge but not the `base.manifest` that marks a repo as a layer; and the walk **silently skips** a declared substrate that lacks one. Every existing derivative (nous, pair, you-decide, 42shots) has a hand-authored `base.manifest` from its #95 cutover branch, which is why this never surfaced — no one had fresh-bootstrapped a brand-new derivative (let alone a multi-level chain) since weave landed.
