# Vocabulary Layer (formal nouns + lifecycles)

`construct/vocabulary/` holds formal **CUE** models of the system's nouns and their
lifecycles — the single authoritative source each consumer *derives* from instead of
re-encoding. Propagates across the layer graph like `datatype` (shared in
`construct/vocabulary/`, repo-local would be `<repo>/vocabulary/`). Motivated by
ariadne#122; the invariant is defended by the `issue-lifecycle` target
(`workshop/targets/issue-lifecycle.md`).

## Current state (#122 M1–M2 — landed)

**The model (M1).**
- `construct/vocabulary/issue.cue` — the `issue` noun. `categories` is the **single
  concrete source** of status membership (open / active / terminal); `#Status`,
  `#Active`, `#Terminal` are *derived* from it via `or()` (so membership is stated
  once and the `#`-defs can't drift). Also: `when` (per-status semantics),
  `lifecycle` (the transition table, with *named* guards whose implementations live
  in sdlc), and `laws` (documented-value + reachable/escapable, enforced by `cue vet`).
- `construct/vocabulary/vet_test.sh` — the M1 gate: the valid model vets, the
  `testdata/issue_invalid.cue` fixture fails, and the **export carries `categories` +
  `lifecycle`** (CUE `#`-definitions don't `cue export`). Test fixtures live under
  `construct/vocabulary/testdata/` so the export doesn't treat them as nouns.

**The compiler + pipeline (M2).**
- `cmd/vocabulary` — `vet` (`cue vet` the merged set), `export --output <dir>` (per-noun
  JSON + a `.source-sha` freshness stamp + a served `SKILL.md`) / `--noun <name>`
  (stdout), `check --output <dir>` (stale-detection vs the merged source). The `cue`
  calls sit behind an injected runner (ARCH-PURE). The DAG-merge is
  `pkg/layergraph.MergeByName` — the shared "merge `*.X` across the layer graph,
  leaf-wins by filename" primitive, also consumed by `cmd/datatype` (ARCH-DRY).
- `construct/local/vocabulary/.dynamic-skill` — weave discovers it by convention and
  execs it (vocabulary binary on PATH) at `weave compile`, materializing the gitignored
  `construct/generated/vocabulary/{issue.json,.source-sha,SKILL.md}`. The `SKILL.md`
  carries the **touch-time breadcrumb**: read `construct/vocabulary/<noun>.cue` before
  editing a lifecycle.
- `Makefile.workflow` — `ensure-cue` (mirrors `ensure-go`; in `bootstrap`),
  `vocabulary-build` (build-in-owner) + the `weave` target puts the vocabulary owner's
  `bin/` on the compile PATH, `issue-json-gen` regenerates the **committed**
  `cmd/sdlc/internal/issue/issue.json` (the M3 `go:embed` input), and `issue-json-check`
  (owner-only, ariadne CI) fails on a stale committed json.

## Planned (#122 M3 — NOT yet wired)

- **M3:** `sdlc` consumes `cmd/sdlc/internal/issue/issue.json` (a pure `IssueModel` +
  category/transition predicates via `go:embed`) in place of the scattered status
  literals (`isTerminalStatus`, the `setstatus`/`state`/`claim` enum/transition checks);
  a conformance test guards the domain. The committed json + `issue-json-check` are in
  place; the Go consumption is M3.

## Relationship to existing entries

- The *operational* status flow (GitHub → local → archive) is
  [Issue Lifecycle](issue-lifecycle.md); **this** entry is the *formal model* those
  statuses will derive from once #122 M3 lands.
- Propagation reuses the layer-graph mechanism — see [weave](weave.md) and the
  datatype DAG-merge in [Data Artifacts](data-artifacts.md).
