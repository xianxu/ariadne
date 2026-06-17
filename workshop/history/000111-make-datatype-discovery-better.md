---
id: 000111
status: done
deps: []
created: 2026-06-16
updated: 2026-06-17
estimate_hours: 7
actual_hours: 0.35
---

# make datatype discovery better

In one instance, when I asked: "ok, make a continuation for pair-pair". I expect agent to read continuation datatype definition and go from there. that didn't happen. agent took some detour to eventually discover continuation is a datatype. 

looking at datatype/SKILL.md, ideally, the supported datatype "noun" should be listed in the description. ideally we would have something like the following:

> description: "Use when the user is requesting an artifact (capture, save, record, create) AND the substance to preserve is conversational context they've already produced. Skip when the user is stating facts, asking questions, or asking the agent to generate substance from scratch. Also trigger when editing markdown with known frontmatter type: {weave datatypes}". 

essentially some templating pattern, i.e. have some deterministic code to parse some directory and insert a more condensed format as prose. 

I guess there are several ways to go about it, if we assume we can rely on "compile time expansion", i.e. `weave compile` will either process some template, as mentioned above, or maybe some of the skills are generated dynamically by some binaries, such that `weave compile` will write out the datatype skill to ./construct/local/datatype/SKILL.md. internally we can do templating itself. 


## Spec

Generalize skill maintenance into **dynamic skills**: a skill package that weave
regenerates at compile time by running an executable marker. `cmd/datatype` is
the first consumer — it generates the datatype `SKILL.md` with a live datatype-
noun list in the description, fixing the truncated `…known frontmatter type:`
that caused the discovery miss in the motivating case.

The two stages are kept distinct (the framing this issue clarified):
**maintenance** (keeping the canonical source skill current — what this issue is
about) vs **lowering** (how weave serves a skill to each agent — already solved,
unchanged here).

### Mechanism — dynamic skills (weave)

- **Marker = an executable `.dynamic-skill`** (tracked) in a skill package. It is
  an ordinary script; weave never parses it (language-neutral). Example
  (`construct/local/datatype/.dynamic-skill`):

  ```sh
  #!/bin/sh
  go run ../../../cmd/datatype --output .
  ```

- **Generate stage** (new — runs **after `walk.Walk`** so the parsed `skill <dir>`
  intents exist to reuse, and **before `GatherSkills`/`planActions`** so the
  regenerated `SKILL.md` is what discovery reads). `walk.Walk` itself reads only
  `construct/deps`/`base.manifest`/`prose` (NOT `SKILL.md`), and the `skill <dir>`
  intents are a product of the walk — so the scan set reuses those already-parsed
  intents (DRY, the skill-system "one discovery"; no second manifest parse, no
  hardcoded dir list). For each package under those dirs containing an executable
  `.dynamic-skill`, weave execs it with **cwd = the package dir**.
  `construct/adapted` (foreign-origin superpowers) is excluded. Read-only paths
  (`--dry-run`, `golden`, `verify-complete`) do **not** run the generate stage
  (they must not mutate the tree); they operate on the committed output, and the
  CI drift guard below catches staleness.
  - **Leaf-layer-only (this is what preserves "ancestors byte-pristine"):**
    generation scans ONLY the **leaf** layer's `skill` intents
    (`layers[len(layers)-1]` — the repo being compiled), never an ancestor's.
    weave iterates ALL resolved layers (ancestors at their real on-disk paths),
    so "no inheritance symlinks post-#104" is necessary but NOT sufficient —
    without leaf-scoping a *derivative's* compile would find
    `ariadne/construct/local/datatype/.dynamic-skill` and exec it with cwd =
    ariadne's dir, mutating an ancestor's tree. Leaf-only scoping is the actual
    guarantee. (A derivative with no dynamic skills of its own generates nothing
    and just symlinks ariadne's committed `SKILL.md`.)
- **Exec seam:** weave gains a narrow, injected exec interface — **separate from
  `weavefs.FS`** (which is filesystem-only by documented stance), e.g.
  `Runner.Run(dir string, argv []string) error` — so the generate stage is
  unit-testable with a fake (no real binary). Production wraps `exec.Command`.
  This is a deliberate, bounded reversal of the filesystem-only stance (#95 M5
  retired the open-ended `go.mod` editor and `cmd/weave` currently has zero
  `os/exec`; this re-adds only "run a package's `.dynamic-skill`"). **Failure
  semantics:** a non-zero exit from `.dynamic-skill` **fails the compile** (loud,
  never a silent skip). `ARCH-PURE`: the stage is a thin exec shell over a pure
  plan ("which intent dirs' packages carry an executable marker").
- **Lower stage:** unchanged — after generation the now-current package dirs
  symlink to each agent exactly as today.
- **Generated outputs are tracked + committed** (NOT gitignored — corrected after
  spec review): the datatype `SKILL.md` is generated by ariadne but consumed by
  *derivatives via symlink* (`brain/.claude/skills/xx-datatype` →
  `ariadne/construct/local/datatype`), and derivatives do not regenerate it — so
  it must physically exist in ariadne's tree (fresh clone, read-only `weave
  skills`/`golden`/`verify-complete`) or the skill silently vanishes from
  discovery. Therefore it is committed codegen: `weave compile` regenerates it
  in place (deterministic/idempotent), and a real change surfaces as a reviewable
  `git diff`. A **drift guard** keeps it honest as a **CI step** (not a golden
  in-memory compare — impossible for an opaque marker whose `--output` weave
  can't redirect): CI runs `weave compile` (tree mutation is fine there) then
  asserts `git diff --exit-code` on the generated skill files, so a stale list
  (datatype added, not regenerated) fails loudly.
  Only `.dynamic-skill` is hand-authored; `SKILL.md` (+ any sub-files) is
  generated-and-committed, source-of-truth being the generator's embedded
  template. `PruneOrphans` never touches it (it GCs only lowered *symlinks*, not
  source-side files).

This is the **skill-binary pattern** (sdlc-style) applied to the *static* surface:
a binary owns the prose; `--help` carries the dynamic-at-runtime instructions; and
`compile`/`--output` generates the *eager* static surface (the description) that
must be present before the agent ever invokes the binary — the one piece that
can't hide behind `--help`. (`ARCH-DRY`: one mechanism — an executable marker —
rather than a weave-internal templating engine or per-skill data manifests.)

### Consumer #1 — `cmd/datatype`

- New Go binary. `//go:embed`s the datatype `SKILL.md` prose as a template
  (`cmd/datatype/SKILL.md.tmpl` — the current ~178-line body, kept readable
  there).
- On `--output <dir>`: enumerates `construct/datatype/*.md` taking each type name
  from the **filename** (without `.md`) — the authoritative convention
  (`SKILL.md` "filename without `.md` is the type name"); the prototype's `type:`
  field is NOT reliable (e.g. `product.md` has `type: type, name: product`).
  Renders the template filling the description's noun list, writes `SKILL.md` into
  `<dir>`. Deterministic: sorted filenames → byte-identical output across runs (so
  the drift guard is meaningful).
- **Names-only** in the description list to start (terse — the description is
  eager-loaded for triggering). The body's existing `awk` enumeration still
  handles full per-type matching at apply-time; revisit a per-noun gloss only if
  names-only proves insufficient.

### Outcome

`construct/local/datatype/` becomes `.dynamic-skill` (hand-authored) + `SKILL.md`
(generated, tracked-and-committed). The description's `…known frontmatter type:`
ends with the live noun list (`continuation, event, meeting-notes, pensive,
procedure, product, project, prose, reference, roadmap, target, travel-plan,
type`), so the agent triggers on "make a continuation for pair-pair" without the
detour that motivated this issue.

### Out of scope

- `construct/adapted/*` dynamic skills (foreign origin — deferred).
- Per-repo *local* datatype lists in the description (it carries the owning
  layer's shared set; a consumer's local datatypes still surface via the body's
  enumeration at apply-time).
- Migrating datatype's runtime helpers into `cmd/datatype` subcommands (the
  fuller skill-binary unification) — this issue is scoped to static-prose
  generation; further consolidation can follow.

## Done when

- A skill package with an executable `.dynamic-skill` is exec'd by `weave compile`
  (generate stage runs after `walk.Walk`, before `GatherSkills`; cwd = package
  dir) — covered by a weave unit test driving the injected `Runner` fake (no real
  binary spawned); a non-zero exit fails the compile (tested).
- The scan set is derived from the **leaf** layer's `skill <dir>` intents (DRY with
  `GatherSkills`; leaf-only so an ancestor's marker is never exec'd — tested),
  with `construct/adapted` excluded; read-only paths
  (`--dry-run`/`golden`/`verify-complete`) do not run it.
- `cmd/datatype --output <dir>` writes a `SKILL.md` whose description noun-list is
  enumerated from `construct/datatype/*.md` (filenames, sorted → byte-identical
  output), prose from the embedded `go:embed` template; unit-tested.
- After `make weave` in ariadne, `construct/local/datatype/SKILL.md` is the
  committed generated file, its description ends with the live datatype nouns, the
  tree is clean (idempotent re-gen), and the skill lowers to the agents unchanged.
- A **drift guard** (CI: `weave compile` then `git diff --exit-code` on generated
  skill files) fails when the committed `SKILL.md` is stale vs regeneration
  (datatype added but not regenerated).
- A fresh `git clone` of ariadne already contains `construct/local/datatype/
  SKILL.md` (it is committed) — `weave skills`/`golden`/`verify-complete` discover
  the datatype skill without first running the generate stage.
- weave atlas + `base-layer-mechanics` + `skill-system` targets reconciled to
  document the exec seam + the dynamic-skill convention.
- `go build ./... && go test ./cmd/weave/... ./cmd/datatype/...` green; `go vet`
  clean.

## Plan

Detailed build sequence: `workshop/plans/000111-dynamic-skills-plan.md`.

- [x] M1 — weave dynamic-skill mechanism: `Runner` exec seam (injected, non-zero
      → compile fails) + the generate stage (runs first in `weave compile`,
      derives its scan set from the `skill <dir>` intents, execs each owned
      package's executable `.dynamic-skill` with cwd = package dir; adapted
      excluded; read-only paths skip it). Unit-tested with a fake Runner.
- [x] M2 — datatype consumer + retire-the-truncation: `cmd/datatype` (`go:embed`
      template + sorted filename enumeration of `construct/datatype/*.md`); the
      datatype `.dynamic-skill`; the committed regenerated `SKILL.md` with the
      live noun list; CI drift guard; atlas + `base-layer-mechanics` +
      `skill-system` reconciled.

## Revisions

### 2026-06-17 — plan-quality findings folded in (3, all blocking, all paper-fixes)
- **Ordering:** `walk.Walk` reads `construct/deps`/`base.manifest`/`prose`, NOT
  `SKILL.md` (that's `GatherSkills`, later). The `skill` intents are a *product*
  of the walk, so "before walk.Walk" + "reuse the walk's intents" contradicted.
  Corrected: generate runs **after `walk.Walk`, before `GatherSkills`**.
- **Leaf-only:** "owned-only by construction (no inheritance symlinks)" was
  necessary-not-sufficient — weave iterates ancestor *layers* at real paths, so a
  derivative compile could exec ariadne's marker. Corrected: scan **only the leaf
  layer's** intents; test asserts an ancestor marker is NOT invoked.
- **Drift guard:** an in-memory/buffer compare in `golden` is impossible for an
  opaque marker (weave can't redirect its `--output`). Corrected to a CI step:
  `weave compile` then `git diff --exit-code` on generated skill files.

## Log



- 2026-06-17: closed — go build/test/vet ./... green; make harness-check 6/0/0; weave compile idempotent (clean tree on recompile). Fix is LIVE: construct/local/datatype/SKILL.md regenerates via the dynamic-skill (cmd/datatype, committed codegen) with the 13 datatype nouns in the description — git diff vs main shows ONLY the description line changed (byte-faithful prose migration, proven by cmd/datatype faithfulness test). CI drift guard (make weave-drift-check) verified fail-on-stale + pass-when-current. M1 (Runner exec seam + leaf-only generate stage) milestone-closed FIX-THEN-SHIP. atlas/weave.md + base-layer-mechanics + skill-system reconciled to the bounded exec seam (the M1-review hard gate). --no-verdict: M2 is the final milestone — this full-issue close auto-dispatches the end-of-issue integration review over the whole M1+M2 diff, which IS M2 boundary review (no separate milestone-close M2). actual 0.35h is the v3 in-window measure and undercounts heavily: design+brainstorm+4 reviews predate the first commit and both M1+M2 were implemented by delegated subagents, so most effort is outside the operator transcript window.; review verdict: FIX-THEN-SHIP
- 2026-06-17: closed M1 — M1 weave dynamic-skill mechanism: go build ./... + go test ./cmd/weave/... + go vet green (independently re-run). Runner exec seam (fake-tested + real /bin/sh exit-code integration test); leaf-only DynamicSkillDirs selection (ancestor marker NOT exec'd, construct/adapted excluded, non-exec marker ignored — all tested); generate stage wired after walk.Walk, before planActions, gated if \!dryRun (golden/verify-complete excluded by construction — verified at main.go:459). No M2 surface touched. --no-atlas: M1 is mechanism-only, no user-facing surface (atlas reconciliation + the datatype consumer land in M2). actual 0.23h is the v3 in-window measure and undercounts — design+plan+3 reviews predate the first M1 commit and M1 was implemented by a delegated subagent, both outside the transcript window.; review verdict: FIX-THEN-SHIP
### 2026-06-16

