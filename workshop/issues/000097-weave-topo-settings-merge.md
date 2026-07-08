---
id: 000097
status: working
deps: [ariadne#95]
github_issue:
created: 2026-06-14
updated: 2026-07-07
estimate_hours: 3
started: 2026-07-07T22:21:33-07:00
---

# weave: topological multi-layer settings merge

## Problem

weave composes `prose` topologically across the whole layer DAG (a leaf gets
ariadne's `AGENTS.base.md` + every ancestor's `AGENTS.local.md`, foundation-first),
but `settings` does **not**. The `merge` verb lowers each `merge <src> <dst>` row to
an independent `MergeSettings{Source, Target}`, and `settingsx.Merge(base, local)` is
a **two-input** fold: `base` = the row's source (`settings.ariadne.json`), `local` =
the repo's `settings.local.json`. So "higher overrides lower" only ever means
"repo-local overrides ariadne-base" — a **middle** layer cannot contribute settings.

Concretely: brain (ariadne→nous→brain) merges ariadne-base + brain-local; nous is
skipped. The day metis (ML layer) or nous wants its own settings fragment — ML
permissions, layer-specific hooks — the current model can't express it. This is an
inconsistency with `prose`, not a `setup.sh` regression (`merge-settings.sh` was also
two-input), so it's an enhancement deferred out of the #95 cutover.

No consumer exists today: nous's `construct/base.manifest` has 0 `merge` rows and no
`settings.<layer>.json` file exists anywhere. The natural first consumer is metis.

## Spec

Make `settings` compose across the layer stack like `prose`: each layer may declare a
`merge settings.<layer>.json settings.json` row in its own `base.manifest`; weave folds
all such sources for a given target **foundation-first**, then the repo-local on top,
into the final `settings.json` — with the existing per-key semantics preserved at every
step.

**The trap (must be designed for, not discovered):** `settingsx.Merge` strips
`$merge_keys` (and all `$`-meta) from its output. A naive fold
`Merge(Merge(ariadne, nous), brain)` loses `$merge_keys` after the first step, silently
flipping `permissions.allow`/`deny`-style arrays from **union** to **replace** for every
layer past the first. The fold must carry `$merge_keys` (from the foundation) through all
intermediate steps and strip meta only once, at the end.

**Work:**
1. **settingsx** — add a `MergeChain(sources [][]byte) ([]byte, error)` (or refactor
   `Merge` to delegate to it) that folds `deepMerge` across N ordered sources preserving
   meta, applies `$remove` from the topmost (local) layer, and strips meta only at the
   end. `deepMerge`/`stripMeta`/`applyRemovals`/the mergeKeys set already exist — this is
   a rewire. Keep `Merge(base, local)` working (it becomes `MergeChain` of two). It is the
   M4 differentially-verified core, so add a dedicated multi-source differential test.
2. **action shape** — `MergeSettings{Source string}` → `{Sources []string}`, threaded
   through `plan.Plan` (lower), `apply.applyMergeSettings`, and the golden
   `classifyMergeSettings`. Keep the `default` omission-guards (ARCH: the Action fan-out).
3. **plan** — group `merge` rows by `Target` across the walked layers, ordered
   foundation-first (the walk order weave already produces), into one `MergeSettings` with
   the ordered `Sources`.
4. **tests** — a 3-layer fixture (foundation defines `$merge_keys`; a middle layer unions
   into an array + overrides a scalar; a leaf + a local) proving topological override AND
   that `$merge_keys` survives the fold (the union-not-replace assertion is the whole point).

## Done when

- A middle layer's `settings.<layer>.json` fragment composes into a downstream repo's
  `settings.json`, foundation-first, repo-local last.
- `$merge_keys` array-union semantics hold across **every** layer in the fold, not just
  the first (regression test for the union→replace trap).
- `weave golden` / `verify-complete` classify the multi-source merge correctly.
- Existing two-input behavior (ariadne-base + local) is byte-for-byte unchanged.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec             design=0.3 impl=0.1
item: smaller-go-module      design=0.4 impl=0.45
item: cross-cutting-refactor design=0.4 impl=0.75
item: atlas-docs             design=0.1 impl=0.12
item: milestone-review       design=0.0 impl=0.2
design-buffer: 0.15
total: 3.0
```

## Plan

- [x] Write the durable implementation plan at
      `workshop/plans/000097-weave-topo-settings-merge-plan.md`.
- [x] Add the pure `settingsx.MergeChain` fold and keep `Merge(base, local)`
      compatible.
- [x] Change `MergeSettings` to carry ordered sources, group merge rows by target
      in `plan.Plan`, and update apply/prune/dry-run consumers.
- [x] Update golden and verify-complete so they classify and cover every source
      in the chain, not just the target.
- [x] Add an end-to-end 3-layer compile fixture, update atlas, and run the full
      weave/Go verification suite.

## Log

### 2026-07-07
- Claimed the issue and entered planning. Current design keeps merge semantics in
  the pure `settingsx` core, keeps filesystem reads/writes in `plan.Apply`, and
  updates all `MergeSettings` consumers to derive from the ordered source-chain
  action shape (ARCH-PURE, ARCH-DRY, ARCH-PURPOSE).
- Implemented `settingsx.MergeChain`, grouped merge rows into ordered
  `MergeSettings{Sources, Target}` actions, updated apply/golden/completeness
  consumers, and added a 3-layer compile fixture proving middle-layer settings
  composition. Targeted verification passed: `go test
  ./cmd/weave/internal/settingsx -count=1`, `go test ./cmd/weave/internal/plan
  -count=1`, `go test ./cmd/weave/internal/golden -count=1`, `go test
  ./cmd/weave -run TestCompileMergesSettingsAcrossLayerChain -count=1`.
- Full verification passed: `go test ./cmd/weave/internal/settingsx -count=1`;
  `go test ./cmd/weave/internal/plan -count=1`; `go test
  ./cmd/weave/internal/golden -count=1`; `go test ./cmd/weave -count=1`; `go
  test ./...`; `git diff --check`.

### 2026-06-14
- Filed from the ariadne #95 tart pass: operator asked whether settings merged
  topologically (higher layer overrides lower, per-key semantics). Per-key semantics
  ARE implemented + M4-verified; topological multi-layer is NOT (two-input, = setup.sh).
  Complexity assessment: more than a loop because of the `$merge_keys`-stripping fold
  trap + the single→multi Source action change; no consumer today → deferred to a ticket.
