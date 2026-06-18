---
id: 000115
status: working
deps: []
github_issue:
created: 2026-06-17
updated: 2026-06-18
estimate_hours: 12
started: 2026-06-18T13:22:13-07:00
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

## Estimate

```estimate
model: estimate-logic-v2
familiarity: 1.0
item: smaller-go-module        design=0.4 impl=1.1
item: smaller-go-module        design=0.4 impl=1.3
item: cross-cutting-refactor   design=0.7 impl=2.2
item: cross-repo-refactor-small design=0.4 impl=1.1
item: skill-or-dispatcher      design=0.2 impl=0.5
item: atlas-docs               design=0.1 impl=0.5
item: milestone-review         design=0.0 impl=0.6
item: milestone-review         design=0.0 impl=0.6
item: milestone-review         design=0.0 impl=0.6
item: milestone-review         design=0.0 impl=0.6
design-buffer: 0.30
total: 12.0
```

- `smaller-go-module` ×2 — M1 (extract `pkg/layergraph` shared walk) + M2 (DAG-aware
  `datatype` binary: `mergeTypes` + `list`/`show` + build-in-owner).
- `cross-cutting-refactor` — M3 weave surgery (generate stage, marker-aware discovery,
  lowering switch, gitignore + prune; the interlocking core).
- `cross-repo-refactor-small` — M4 migration (event/travel-plan/reference ariadne→nous,
  retire the whole-dir symlink, re-weave consumers).
- `skill-or-dispatcher` — rewrite the datatype SKILL.md body for binary-based apply-time
  access. `atlas-docs` — reconcile atlas + skill-system + base-layer-mechanics targets.
- `milestone-review` ×4 — M1/M2/M3 milestone-closes + the end-of-issue integration review.

recomputed = 2.2×1.30 + 9.1 = 11.96 ≈ total 12.0 (frontmatter `estimate_hours: 12`).

## Revisions

### 2026-06-18 — scope-check cleared by operator premise + first concrete consumer

The scope-check gate ("do derivatives actually define local datatypes worth
eager-triggering on?") was run against ground truth: **zero** live derivatives
(brain, nous, pair, 42shots, you-decide, xianxu.dev) define any local
`datatype/*.md` — every one consumes ariadne's shared 13 verbatim. The only
non-canonical hits in the whole workspace are two **stale pre-#104 worktrees**
(`nous-14-m1`, ariadne `000031`) carrying inherited copies, not local types.

So the *current* need is empirically nil — but the operator supplied the missing
premise: **repo-specific nouns are coming, deliberately, to keep ariadne from
bloating.** ariadne is the base layer; it is already over-broad — `event`,
`travel-plan`, `reference` are personal-assistant concerns that fit **nous** (the
personal-assistant layer), not a generic base. Decision: **build the mechanism,
then migrate `event` + `travel-plan` + `reference` from ariadne → nous** as the
first concrete consumer. That migration (a) validates the DAG-merge against a real
case, retiring the "speculative generality / single consumer" risk the pensive
flagged, and (b) is a genuine layering cleanup — those nouns drop out of
pair/42shots/you-decide/xianxu.dev (ariadne-direct, no personal-assistant need)
while staying live for nous and brain (brain → nous → ariadne).

**Scope delta:** #115 now also covers the cross-repo prototype migration +
re-weave of every consumer, so the "Done when" gains a migration-correctness check
(post-migration: nous/brain eager-trigger on `event`; pair does not).

## Plan

Durable plan: `workshop/plans/000115-dag-merged-datatype-skills-plan.md` (authored
via `superpowers-writing-plans`). Design fork resolved with operator: **Model A** —
the `datatype` binary is the single DAG-aware access point (apply-time becomes
`datatype list` / `datatype show <name>`; weave never lowers prototypes; the
whole-dir `symlink construct/datatype` row is retired).

- [x] M1 — Extract the transitive `construct/deps` walk into module-level
      `pkg/layergraph` (both weave + datatype import; behavior-preserving; weave
      suite green).
- [x] M2 — `datatype` becomes DAG-aware (`mergeTypes` union local-wins) + gains
      `list`/`show` subcommands + build-in-owner PATH binary.
- [x] M3 — weave generate-stage redesign (all-layers visible set, marker run with
      cwd=R root → leaf-rooted output), lowering switch (dynamic → this-repo),
      materialized SKILL.md gitignored everywhere, prune the orphan materialized
      class, repo-agnostic marker.
- [ ] M4 — Migrate `event`/`travel-plan`/`reference` ariadne→nous, retire the
      whole-dir datatype symlink, SKILL.md body → binary-based apply-time access,
      reconcile atlas + skill-system + base-layer-mechanics, E2E migration proof.

## Log




- 2026-06-18: closed M3 — M3: weave generate-stage surgery. construct/generated/<dir> materialization (gitignored everywhere); marker-aware discovery (entry from tracked marker, BodyPath→leafRoot/construct/generated, fresh-clone safe); DynamicSkills all-layers visible-set (intent.Selected, adapted excl, dedup); generate cwd=leafRoot (ancestor tree never mutated); prune generated class; repo-agnostic marker + datatype-build PATH wiring. VERIFIED: go build/vet/test ./... green; make weave ariadne+nous+pair clean+idempotent, each materializes OWN construct/generated/datatype; C1 GATE — weave skills shows exactly 1 xx-datatype in nous+pair (no <repo>-datatype); committed construct/local/datatype/SKILL.md removed (marker kept); harness-check 6/0.; review verdict: FIX-THEN-SHIP
- 2026-06-18: closed M2 — M2: datatype DAG-aware (mergeTypes union local-wins-by-filename, pure over layergraph.FS; product.md filename-trap + shadow + leaf-local tested) + list/show subcommands + datatype-build PATH binary. go build/vet/test ./... green; datatype list → 13 nouns; show unknown → exit 1; make weave byte-identical construct/local/datatype/SKILL.md (gap-bridge: old marker + deprecated --datatype-dir still work); harness-check 6/0.; review verdict: FIX-THEN-SHIP
- 2026-06-18: closed M1 — M1: pkg/layergraph (FS+ParseDeps+Resolve+Walk) + pkg/frontmatter extracted module-level; go build/vet/test ./... green incl. full weave suite (behavior-preserving regression proof) + new pkgs; make weave idempotent + clean tree; no discoverEdges/substrateTargets/ParseDeps/Resolve/frontmatterDescription survives in cmd/weave (grep empty, ARCH-DRY, one walk). 5 TDD commits 6adb3bf..aca8d81.; review verdict: FIX-THEN-SHIP
### 2026-06-17
