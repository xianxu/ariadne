---
id: 000155
status: done
deps: []
github_issue:
created: 2026-07-01
updated: 2026-07-06
estimate_hours: 1.08
started: 2026-07-06T10:00:21-07:00
actual_hours: 1.34
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

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module    design=0.15 impl=0.3
item: smaller-go-module    design=0.1  impl=0.25
item: milestone-review     design=0.0  impl=0.2
design-buffer: 0.30
total: 1.08
```

## Plan

Design (combine Spec's Option 1 + 2). `ParseDeps` returns ONLY `substrate` rows
(layer edges), so a declared substrate resolving to a **present** dir without
`construct/base.manifest` is unambiguously a broken layer edge — not a legitimate
"non-layer peer". The current walk conflates that with an **absent** peer and
silently drops both; split them.

- [x] **Fix 1 — loud failure in the shared walk** (`pkg/layergraph/walk.go`,
      `discoverEdges`). When a substrate target dir is **present** but ships no
      `construct/base.manifest`, return an actionable error naming the substrate +
      the missing file (helper `pathExists(fs, dep)` via `fs.Stat`). An **absent**
      target keeps the silent present-skip (peer not checked out — legitimate).
      This is the single-source-of-truth fix: all three `layergraph.Walk` consumers
      (weave, datatype, vocabulary) share the footgun (ARCH-DRY). Rewrite the two
      `TestWalkPresentSkipNonLayerDep` tests (pkg/layergraph + cmd/weave/internal/walk)
      that currently PIN the silent skip → assert the error; add an absent-peer
      silent-skip test + a depth-3 chain where a MID layer lacks a manifest (the
      exact kbench→kaggle→metis footgun) → error.
- [x] **Fix 2 — seed on establish** (`cmd/weave/main.go`, `runLink`). After
      recording `substrate <path>`, seed a minimal `construct/base.manifest`
      (header comment + `internal prose AGENTS.local.md`, matching #95-cutover
      repos) in the CURRENT repo when absent, so a fresh derivative is a valid
      traversable layer out of the box. Idempotent: never overwrite an existing
      manifest; announce what it did. With each repo seeding its own manifest at
      link time, a chain bootstrapped foundation-leafward compiles fully without
      hand-authoring — and Fix 1 is the backstop if a substrate still lacks one.

Verified reading: `loadLayer` gracefully skips a declared-but-absent prose
fragment, so the seed manifest alone marks the repo traversable (no need to also
create `AGENTS.local.md`).

## Log

### 2026-07-01

Filed from the kaggle-ml-base-layer bootstrap (metis/kaggle/kbench, in nous). **Workaround already applied** in those three repos: hand-authored a minimal `construct/base.manifest` (`internal prose AGENTS.local.md` + a header noting that shipping the manifest is what marks the repo as a traversable layer) in each; all three then compiled to 97 actions and the transitive chain composed correctly (kbench `AGENTS.md` carries the ariadne Constitution). This ticket is to fix the *tooling* so the next fresh derivative doesn't hit the silent failure.

### 2026-07-06 — implemented (Fix 1 + Fix 2)
- 2026-07-06: closed — Real weave binary end-to-end: (1) present-but-manifest-less substrate → weave compile errors loudly naming the missing base.manifest (was silent 1-action no-op); (2) absent substrate → silent skip preserved; (3) weave link seeds base.manifest, a fresh 3-level chain compiles fully — leaf AGENTS.md composes the foundation constitution through the seeded intermediate + symlinks foundation Makefile; (4) verify-complete over all 10 present siblings = 0 under-produced, none trips new error. Tests: both TestWalkPresentSkipNonLayerDep pins rewritten to assert the error, + absent-skip + kbench-chain-broken + 3 runLink seed tests. Full go test ./... green (25 pkg), build/vet/gofmt clean.; review verdict: FIX-THEN-SHIP

Shipped both preferred options:

- **Fix 1** (`pkg/layergraph/walk.go` `discoverEdges`): a `substrate` target that is
  PRESENT on disk but lacks `construct/base.manifest` now returns a loud, actionable
  error (names the substrate, the declaring `construct/deps`, and the missing file).
  An ABSENT target keeps the silent present-skip. New `pathExists(fs, dep)` helper
  splits the two. One fix, all three `layergraph.Walk` consumers (weave/datatype/
  vocabulary) — ARCH-DRY.
- **Fix 2** (`cmd/weave/main.go` `runLink`): after recording `substrate <path>`,
  seeds a minimal `construct/base.manifest` (header + `internal prose AGENTS.local.md`)
  in the linking repo when absent. Idempotent (never clobbers), runs even when the
  deps row was already present (repairs a pre-#155 repo). Seed content is one-source
  in `seededBaseManifest` (ARCH-DRY — no third hand-copy).

**Verification (real `weave` binary, behavior diff vs main):**

- **Loud failure** — a fresh `consumer` with `substrate ../broken` (present, no
  manifest) → `weave compile` errors: *"substrate …/broken … is present but not a
  compilable layer: missing …/broken/construct/base.manifest — seed its base.manifest
  (`weave link` does this) …"* (exit non-zero). Was a silent `applied 1 action`.
- **Absent-peer skip preserved** — `substrate ../notcheckedout` (absent) → compiles
  clean, no error (partial checkouts don't hard-fail).
- **Seed + full chain** — `mkdir mid && weave link ../base` → *"seeded
  construct/base.manifest (marks mid a traversable layer)"*; `weave link ../mid` in
  leaf seeds leaf. `leaf` compile then reaches the foundation THROUGH the seeded mid:
  composes `AGENTS.md` carrying the exported foundation constitution + symlinks the
  foundation's `Makefile` (not a 1-action no-op). Confirmed the composed
  `leaf/AGENTS.md` contains the foundation's exported prose.
- **No live-repo regression** — `weave verify-complete` over every present sibling
  (ariadne, nous, brain, metis, kaggle, kbench, 42shots, pair, brain-family,
  brain-private) reports 0 under-produced; none trips the new error.
- **Tests**: rewrote both `TestWalkPresentSkipNonLayerDep` pins (pkg/layergraph +
  cmd/weave/internal/walk) → assert the loud error; added `TestWalkAbsentSubstrateSilentlySkipped`,
  `TestWalkChainBrokenByManifestlessIntermediate` (the kbench→kaggle→metis repro),
  and three `runLink` seed tests (seeds-when-absent, never-clobbers, end-to-end
  chain-traversable). Full `go test ./...` green (25 pkg), `go build`/`vet`/`gofmt` clean.

Atlas: `atlas/workflow/weave.md` Key-decisions — added the present-substrate-must-be-a-layer
rule + the `weave link` seed companion.

### 2026-07-06 — boundary review (FIX-THEN-SHIP) applied

Verdict **FIX-THEN-SHIP** (high, no Critical/Important). Applied all three actionable
notes before shipping:

- Stale doc: `runLink`'s "Read-only on everything but the one deps file" was false
  (it now also writes `base.manifest`) — corrected to name both writes; updated
  `buildLink` `Short` + the `buildLink` doc comment to mention the seed.
- Coverage gap: the "repair a pre-#155 repo" path (seed when the deps row already
  exists) was exercised by `TestLinkIdempotent` but unasserted — a regression
  re-adding an early `return nil` before `ensureBaseManifest` would have passed.
  Extended `TestLinkIdempotent` to assert the manifest is seeded in that case.
- The third note (raw substrate path interpolated into seed `#` comments) is
  harmless (comment-stripped by both parsers; a newline-bearing path isn't a real
  CLI arg) — no change, acknowledged.

Re-ran weave + layergraph suites green; build/vet/gofmt clean.
