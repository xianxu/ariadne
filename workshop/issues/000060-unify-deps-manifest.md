---
id: 000060
status: working
deps: []
github_issue:
created: 2026-06-01
updated: 2026-06-01
estimate_hours: 5
---

# Unify substrate + data dependencies into one manifest; retire construct/go.mod as the peer graph

## Problem

Substrate dependencies (the ariadne-styled ancestor a repo inherits its base
layer + Go tools from) are declared today inside `construct/go.mod`, via a
`require` + `replace => ../../ariadne` + `tool …/cmd/sdlc` stub module
(`<name>-construct`). This was an artifact of an imperfect grasp of the go.mod
system — go.mod is being used as a substrate-peer graph, which is **not what it
is intended for**. Consequences:

- A whole second Go module (`<name>-construct`) exists per derivative that
  builds nothing of the repo's own — it carries only the substrate declaration.
- Non-Go substrate consumers (a markdown-only brain) are forced to carry a fake
  go.mod just to declare lineage.
- The peer graph is split across two carriers: substrate lives in
  `construct/go.mod`, content lives in `construct/data-deps` (a clean, flat,
  language-agnostic, two-column manifest). Two mechanisms for the same
  primitive — "a sibling clone, floating-HEAD, surfaced via symlink."

Meanwhile `construct/go.mod`'s **build** role (compile `cmd/sdlc` through the
`replace`) is already redundant: both the deploy path and the dev path can —
and the dev path (`dev-aliases.sh`) already does — **build the tool from the
owner's checkout** (`cd ../ariadne && go build ./cmd/sdlc`), using ariadne's
own go.mod. See [[000055-sdlc-binary-single-owner-path-freshness]] and
[[000049-data-dependencies]].

## Spec

Collapse substrate + data dependencies into **one language-agnostic manifest**,
and converge tool-building on **build-in-owner** so `construct/go.mod` can be
deleted from derivatives entirely.

### Design decisions (settled with operator)

1. **One manifest subsumes both.** `construct/data-deps` grows (or is replaced
   by `construct/deps`) so each row carries a `kind: substrate | data`
   attribute. Substrate and data are the same primitive (sibling clone,
   floating-HEAD, symlink-mounted); `kind` switches install behavior:
   - `substrate` → clone + symlink base-layer files per `base.manifest` +
     settings merge + **recurse** (`make bootstrap` cascade) + drive the tool
     build. Transitive; needs the existing cycle/depth guards.
   - `data` → clone + symlink at one declared mount path. Flat, no recursion.

2. **Build-in-owner for BOTH audiences** (the lynchpin). Tool source lives only
   in its owner (sdlc → ariadne). Both paths build there, differing only in
   cadence — neither needs a derivative-side go.mod:
   - **Deploy** (consumer: clone `pair`, `./bootstrap.sh`, done): `sdlc-build`
     becomes `cd <owner-sibling> && go build -o <thisrepo>/bin/sdlc ./cmd/sdlc`,
     run **once** at bootstrap into the consumer's bin. Owner sibling is already
     present at that point (bootstrap-peers clones ancestors before `tools`).
     Consumer still gets a stable binary; goes stale until rerun — unchanged
     semantics, just a different go.mod resolves the build.
   - **Dev** (substrate developer): `dev-aliases.sh` unchanged — fresh-per-call,
     builds in owner, PATH function shadows the deployed binary.
   - #37 dependency-isolation (keep cobra etc. out of the consumer's app go.mod)
     becomes **moot**: the tool is never declared as a consumer dep at all, so
     nothing leaks regardless.

3. **Format must stay inline-shell-parseable — no full YAML parser.**
   `bootstrap.sh` parses the graph *before* substrate symlinks exist
   (chicken-and-egg — it can't source a symlinked parser, and has no `yq`). The
   constraint is the *parser*, not the surface syntax: a **flat YAML subset**
   (one field per line, no nesting/anchors/multiline) is fine — it reads as YAML
   to the eye yet parses with grep/awk, same class as today's `data-deps`
   columns. So either a columnar `kind url mount` line or flat `key: value`
   rows works; what's ruled out is anything needing a real YAML engine. **Open
   schema question for planning:** exact layout (columnar vs flat key/value), and
   whether to extend `data-deps` in place or introduce `construct/deps` that
   subsumes + deprecates it.

4. **Pinning is OUT of scope.** Default stays floating-HEAD (loose-submodule
   velocity; minor markdown drift is harmless; co-development of ariadne+nous
   would make pins churn). Revisit when a concrete need appears — and even then,
   pins stay local + advisory, conflicts **warned, never transitively
   reconciled** (we are not rebuilding npm).

### Migration surface (all base-layer; propagates to ~13 siblings)

`construct/go.mod`'s graph is read/written by several walkers that **must agree**
(per `atlas/workflow/setup-and-replication.md`); all migrate together:

- `construct/setup.sh` — the **writer** (stubs `construct/go.mod`, declares
  require/replace/tool). Migrate last so existing repos keep working mid-flight.
- `construct/scripts/bootstrap-peers.sh` — greps `replace => ../<name>`.
- `construct/scripts/list-peers.sh` — reads substrate replaces.
- `bootstrap.sh` — **inline-copied** parser (the constraint forcing flat text).
- `Makefile.workflow` — `sdlc-build` (→ build-in-owner) + `sdlc-install`.
- Fixture tests: `bootstrap-transitive.test.sh`, `discover-ancestors.test.sh`,
  `sandbox-sync-set.test.sh`.
- Delete `construct/go.mod` from all derivatives once readers are migrated.

## Done when

- A single language-agnostic manifest declares both substrate and data deps via
  a `kind` attribute; `construct/go.mod` is gone from all derivatives.
- `make bootstrap` on a fresh consumer clone (no dev-aliases sourced) still
  yields a working `bin/sdlc`, built from the owner checkout.
- `dev-aliases.sh` dev flow unchanged (fresh-per-call shadows deploy binary).
- All four+ peer-graph walkers read the new manifest and agree on the peer set;
  existing transitive/cycle/depth tests pass against the new format.
- A non-Go substrate repo (e.g. a markdown brain) can declare ariadne lineage
  with **no go.mod**.
- Atlas updated: `setup-and-replication.md`, `base-layer.md`, `sdlc-binary.md`.

## Plan

Plan settled (plan mode, 2026-06-01): positional-column `construct/deps`,
foundation-first. Full plan: `~/.claude/plans/curried-launching-bubble.md`.
Scope of THIS execution = M1 + M2 (foundation); M3–M5 deferred to a separate
propagation-gated rollout. Estimate (5h) covers the foundation only.

- [ ] M1 — `construct/deps` positional format + shared `lib-deps.sh` parser +
      dual-read in all 5 walkers (additive; no-op until data exists). Tests:
      construct/deps cases added to the 3 fixtures + drift test, go.mod cases kept.
- [ ] M2 — build-in-owner `sdlc-build` (resolve owner via `dev-aliases.sh
      --list`); add `dev-aliases.sh` to base.manifest. Orthogonal to M1.

Deferred (separate sessions, propagation-gated): **M3** writer flips to
`construct/deps` + folds data-deps; **M4** delete `construct/go.mod` ×13 + drop
dual-read fallback; **M5** retire legacy `data-deps`. See plan file for gates.

## Log

### 2026-06-01

Issue opened from a design conversation. Investigation findings:

- **Only ariadne-owned tool built through any `construct/go.mod` is `sdlc`** —
  scanned all 13 ariadne-styled siblings; every `tool` directive is
  `…/cmd/sdlc`, ariadne's `cmd/` has only sdlc. No derivative builds nous or any
  other ancestor tool. Build role is reducible to "compile sdlc."
- **`construct/go.mod` consumers** (besides bootstrap-peers.sh): `setup.sh`
  (writer + ancestor walker), `list-peers.sh` (reader), `Makefile.workflow`
  (`sdlc-build`), and `bootstrap.sh`'s inline parser; plus 3 fixture tests.
- Atlas confirms (`sdlc-binary.md:132`, `setup-and-replication.md:145,276`) the
  four-walker-agreement requirement and the #37 dep-isolation rationale.
- Operator confirmed: deploy = build-in-owner; peer graph belongs in any file;
  go.mod was incidental ("grasp of go.mod not good"); skip pin for now.
