---
id: 000115
status: open
deps: []
github_issue:
created: 2026-06-17
updated: 2026-06-17
estimate_hours:
---

# DAG-merged dynamic skills: per-repo datatype enumeration across the layer graph

## Problem

#111 shipped the datatype dynamic skill **repo-agnostic**: `cmd/datatype`
enumerates ariadne's `construct/datatype/*.md`, the SKILL.md is generated +
committed *once* in the owner, and consumers **symlink ariadne's copy**. But the
effective datatype set is **per-repo** — the union over a repo's layer DAG
(ariadne's shared set + each intermediate layer's + the leaf's local
`datatype/`), local-shadows-shared. So a derivative's eager skill *description*
lists ariadne's nouns, not its own local datatypes — a derivative-defined type
won't trigger `xx-datatype` until the skill body's apply-time enumeration runs.
#111 scoped this out ("per-repo local datatype lists in the description") as a
known limitation; this issue closes it.

It is also the first concrete instance of a broader pattern (see the pensive
referenced below): **DAG-awareness as a capability available to subsystems
*alongside* weave, not a weave monopoly** — the layer graph is a platform
primitive, not weave's private state.

## Spec

Make the datatype dynamic skill **DAG-merged + materialized per-repo**, with weave
staying content-blind. Converged design:

- **Materialize per-repo into the mirror path.** Each repo gets its own
  `<repo>/construct/local/datatype/SKILL.md`, rendered with *that repo's*
  DAG-merged datatype set — **gitignored in every repo, including ariadne's own**
  (ariadne's is just its own render). The only tracked source is the template
  (`cmd/datatype/SKILL.md.tmpl`) + the marker. #111's committed
  `construct/local/datatype/SKILL.md` goes away.
- **Lowering switch (the uniform rule).** `.claude/skills/xx-<name>` →
  `<owner>/construct/local/<name>` for a **static** skill; →
  **`<this-repo>`**`/construct/local/<name>` for a **dynamic** skill. The only
  difference is owner-vs-local target; the materialized dir is shaped exactly like
  a normal skill source, so discovery (GatherSkills) reads the local copy with no
  special case.
- **Generator is DAG-aware at runtime.** Run in repo R, the generator reads R's
  `construct/deps`, merges the `datatype/` dirs across R's layers (union;
  local-shadows-shared by filename), and renders R's SKILL.md. weave supplies
  nothing about datatypes — it just runs the marker. (The merge POLICY lives in
  the generator; it is intentionally simpler than weave's `𝒜(R)`/visibility math —
  a deliberate per-subsystem choice, not a missed unification.)
- **Generator becomes a build-in-owner binary on PATH** (`datatype`), replacing
  #111's owner-relative `go run ../../../cmd/datatype` — the single change that
  makes the marker repo-agnostic. This is NOT #110's anti-pattern: #110 retired
  *runtime substrate-resolution of a script file*; a PATH binary distributed by
  build-in-owner (like `sdlc`/`weave`) is the sanctioned distribution. Reading
  ancestor `datatype/` *dirs* (data) is exactly what weave's walk already does.
- **Single shared DAG-walk (the invariant).** Extract the transitive
  `construct/deps` walk into a **module-level package** (NOT `cmd/weave/internal/*`,
  which `cmd/datatype` can't import) that both weave and the generator import. The
  "single mechanism" rule applies to the *walk* — one source of truth for "what is
  repo R's layer graph" — so subsystems never diverge on the graph even while
  differing on merge policy.
- **weave stays content-blind.** Its only awareness (unchanged from #111) is "is
  this skill dynamic?" (does it carry the marker) — to pick the lowering target
  (local vs owner) and sequence the generate stage. It never learns what
  datatypes are. weave owns WHEN; the generator owns WHAT; weave never owns
  CONTENT.
- **Gitignore/prune gain a new class:** per-repo *materialized* skill dirs under
  `construct/local/` (gitignored everywhere; pruned when the owner drops the
  dynamic skill). Legibility rule: generated-gitignored vs tracked-hand-authored
  is how `git status` keeps materialized-inherited skills distinct from a repo's
  own hand-authored ones.

## Done when

- A derivative that defines a local `datatype/<x>.md` gets `<x>` in its *own*
  `construct/local/datatype/SKILL.md` description (materialized per-repo,
  DAG-merged), while ariadne's render lists only its set — verified across a
  two-repo fixture (owner + derivative-with-a-local-datatype).
- The generator is a PATH binary (build-in-owner), runs from any repo, reads the
  current repo's DAG; the marker is repo-agnostic (no owner-relative path).
- The transitive `construct/deps` walk is one shared module package imported by
  both `weave` and the generator (no duplicate walk).
- Lowering: static → owner symlink, dynamic → local symlink; weave still detects
  only the dynamic flavor. Materialized SKILL.md is gitignored everywhere; prune
  GCs an orphaned materialized dir when the owner drops the skill.
- `go build ./... && go test ./... && go vet`; `make weave` idempotent + clean;
  `make harness-check` green.

## Scope check (gate before building)

Validate the premise first: **do derivatives actually define local datatypes
worth eager-triggering on?** #111 already fixes the motivating case (shared
datatypes like `continuation` are in the description), and the skill body's
`awk` enumeration already surfaces local datatypes at *apply* time. The marginal
benefit here is narrowly "a derivative's *local* datatype in the *eager* trigger
surface." Build only if that's a real, recurring need — otherwise #111's
"shared-in-description + local-via-body-enumeration" is the YAGNI resting point.

## Non-goals / related

- The **generalized "DAG-aware subsystem" pattern** (shared-mechanism /
  per-subsystem-policy / weave-sequences-blind; layer-graph-as-platform-primitive)
  — explored in the pensive
  `workshop/pensive/2026-06-17-01-pensive-layer-graph-as-platform-primitive.md`.
  This issue is its first concrete instance; the pattern stays prose until it
  matures (→ maybe a `target` later).
- AI-cost modeling — out (deferred).
- Related: #111 (the repo-agnostic version this fixes), #110 (the owner-resolution
  distinction), #092 (segment/window concern), #112–#113 (estimate/actual).

## Plan

- [ ]

## Log

### 2026-06-17
