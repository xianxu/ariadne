# Vocabulary Layer (formal nouns + lifecycles)

`construct/vocabulary/` holds formal **CUE** models of the system's nouns and their
lifecycles — the single authoritative source each consumer *derives* from instead of
re-encoding. Propagates across the layer graph like `datatype` (shared in
`construct/vocabulary/`, repo-local would be `<repo>/vocabulary/`). Motivated by
ariadne#122; the invariant is defended by the `issue-lifecycle` target
(`workshop/targets/issue-lifecycle.md`).

## Current state (#122 M1 — landed)

- `construct/vocabulary/issue.cue` — the `issue` noun. `categories` is the **single
  concrete source** of status membership (open / active / terminal); `#Status`,
  `#Active`, `#Terminal` are *derived* from it via `or()` (so membership is stated
  once and the `#`-defs can't drift). Also: `when` (per-status semantics),
  `lifecycle` (the transition table, with *named* guards whose implementations live
  in sdlc), and `laws` (documented-value + reachable/escapable, enforced by `cue vet`).
- `construct/vocabulary/vet_test.sh` — the M1 gate: the valid model vets, the
  `issue_invalid.cue` fixture fails, and the **export carries `categories` +
  `lifecycle`** — CUE `#`-definitions don't `cue export`, so the gate guards the
  concrete data consumers actually read.

## Planned (#122 M2/M3 — NOT yet wired)

- **M2:** `cmd/vocabulary` (DAG-merge + `cue export`, reusing `pkg/layergraph`) and a
  `.dynamic-skill` materializing `construct/generated/vocabulary/issue.json` at
  `weave compile`; an `ensure-cue` bootstrap recipe in `Makefile.workflow`.
- **M3:** `sdlc` consumes the exported JSON (a pure `IssueModel` + category/transition
  predicates) in place of the scattered status literals; a conformance test guards the
  domain.

## Relationship to existing entries

- The *operational* status flow (GitHub → local → archive) is
  [Issue Lifecycle](issue-lifecycle.md); **this** entry is the *formal model* those
  statuses will derive from once #122 M3 lands.
- Propagation reuses the layer-graph mechanism — see [weave](weave.md) and the
  datatype DAG-merge in [Data Artifacts](data-artifacts.md).
