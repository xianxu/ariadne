# DAG-Merged Datatype Skills Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the datatype dynamic skill DAG-merged and materialized per-repo — each repo's `xx-datatype` skill (eager description *and* apply-time prototype access) reflects the union of datatype prototypes across its own layer DAG (local/leaf shadows shared by filename) — then migrate `event`/`travel-plan`/`reference` from ariadne down to nous as the first real consumer.

**Architecture:** Extract the transitive `construct/deps` walk into a module-level `pkg/layergraph` package both `weave` and a new `datatype` PATH binary import (one graph, no divergence — the pensive's *single mechanism* invariant). The `datatype` binary owns the merge **policy** (union, local-wins-by-filename) and is the single DAG-aware access point: it renders the per-repo `SKILL.md` (eager) and answers `datatype list` / `datatype show <name>` (apply-time). weave stays **content-blind** — it sequences the generate stage (WHEN), detects dynamic-ness by **marker presence**, and points the lowering target at this-repo's materialized copy; it never learns what a datatype is. The whole-dir `symlink construct/datatype` propagation is retired — prototypes are read across the DAG by the binary, never copied or symlinked into consumers.

**Materialization sink (the load-bearing decision, post plan-review C1):** the per-repo materialized `SKILL.md` is written to a **new `construct/generated/<dir>/` tree** — gitignored everywhere and **never scanned as a skill dir** — *not* into `construct/local/<dir>/` (which a consumer's own `skill construct/local` manifest row re-scans, producing a duplicate `<repo>-datatype` skill). Discovery becomes **marker-aware**: a `.dynamic-skill` marker in a scanned skill dir declares a dynamic skill whose body is `<compiling-repo>/construct/generated/<dir>/SKILL.md`. This is uniform for owner and consumer, and means the skill is discovered from the *tracked marker* (never vanishes in a fresh clone — the #111 reason for committing the file dissolves).

**Tech Stack:** Go (module `github.com/xianxu/ariadne`), `weavefs.FS`/`weavefs.Runner` injected seams, `go:embed` template, Makefile build-in-owner distribution (mirrors `weave`/`sdlc`), `construct/deps` layer graph.

---

## Core concepts

The work splits along the pensive's three-way invariant (`workshop/pensive/2026-06-17-01-pensive-layer-graph-as-platform-primitive.md`): **mechanism shared** (the DAG walk — one library), **policy per-subsystem** (datatype's merge lives in the binary), **execution sequenced-but-blind** (weave runs the marker, ignorant of content).

**Naming discipline (plan-review I2):** two distinct names must never be conflated. The skill **Name** is the prefixed `xx-datatype` (used for the lowered symlink *Dst* `.claude/skills/xx-datatype`). The skill **dir name** is the bare `datatype` (used for the materialization *output path* `construct/generated/datatype` and the lowering *Src*). Throughout, `<dir>` = the bare package dir name; `<Name>` = the prefixed skill name.

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `Walk` (ordered layer roots) | `pkg/layergraph/walk.go` | new |
| `ParseDeps` | `pkg/layergraph/deps.go` | new (moved) |
| `Resolve` (DFS post-order; takes root + deps map) | `pkg/layergraph/resolve.go` | new (moved) |
| `frontmatter.Description` + `unquote` | `pkg/frontmatter/frontmatter.go` | new (moved) |
| `frontmatterDescription` (weave) | `cmd/weave/internal/walk/skills.go` | deleted (delegates to `pkg/frontmatter`) |
| `findRepoRoot` (walk up to nearest `construct/`) | `cmd/datatype/main.go` | new |
| `TypeProto` | `cmd/datatype/merge.go` | new |
| `mergeTypes` (DAG union, local-wins) | `cmd/datatype/merge.go` | new |
| `typeNames` | `cmd/datatype/datatype.go` | deleted (subsumed by `mergeTypes`) |
| `renderSkill` | `cmd/datatype/datatype.go` | modified |
| `formatList` (apply-time listing) | `cmd/datatype/list.go` | new |
| `DynamicSkill{Name,Dir,MarkerPath,OutputRel}` | `cmd/weave/internal/walk/dynamic.go` | new |
| `DynamicSkillDirs` (leaf-only) | `cmd/weave/internal/walk/dynamic.go` | deleted |
| `skill.Entry.Dynamic` flag | `cmd/weave/internal/skill/skill.go` | modified |
| `scanSkillDir` (marker-aware) | `cmd/weave/internal/walk/skills.go` | modified |
| `SkillSymlinks` (UNCHANGED — switch realized in `BodyPath`) | `cmd/weave/internal/plan/skill_symlinks.go` | unchanged |
| `walk.GeneratedRel` + `dynamicMarker` (single-source) | `cmd/weave/internal/walk/dynamic.go` | new |
| `generatedGitignore` (`/construct/generated/`) | `cmd/weave/internal/plan/gitignore.go` | modified |
| `shouldPruneGenerated` | `cmd/weave/internal/plan/prune.go` | new |

- **`Walk` (pkg/layergraph)** — given a repo root + an injected `FS`, returns the transitive `construct/deps` layer roots **foundation-first, leaf last** (absolute, canonicalized). *Single source of truth* for "what is repo R's layer graph"; weave's internal `walk.Walk` and the `datatype` binary both call it so they never diverge on topology (ARCH-DRY; the pensive's load-bearing invariant).
  - **Relationships:** 1:1 with a repo root → ordered N-list of ancestor roots. weave's `layer.Layer` loading (manifests, prose) and datatype's per-layer `construct/datatype/` reads layer *on top*; the topology is shared.
  - **DRY rationale:** Eliminates the second graph walk #115 would otherwise add. Today the walk lives in `cmd/weave/internal/{walk,layer}` (un-importable by `cmd/datatype`); this is the extraction that lets a *second* DAG-aware subsystem exist without copy-pasting `discoverEdges`/`resolveOrder`/`ParseDeps`.
  - **Future extensions:** A third DAG-aware subsystem (per-repo config, generated indexes — the pensive's open question) imports the same `Walk`. If it earns a `target`, "the layer walk is one shared library" is the invariant to pin.

- **`TypeProto` / `mergeTypes`** — `TypeProto{Name, Description, BodyPath}` is one resolved datatype prototype; `mergeTypes` takes the ordered layer roots + the leaf's project-local `datatype/` dir and returns the **DAG-merged set**: union over `{each layer's construct/datatype/*.md} ∪ {leaf's datatype/*.md}`, keyed by filename, **local/leaf shadows shared** (a downstream layer's `event.md` wins over an ancestor's). Reads ancestor **source** prototype dirs directly (NOT ancestors' materialized copies), so a compile is independent of whether ancestors were compiled first (plan-review I3 — re-weave order is immaterial to correctness). Pure over a list of `(dir, []filename)` — the only IO (ReadDir per dir) sits at the edge.
  - **Relationships:** N TypeProtos per repo; ownership is the *nearest* (most-leafward) layer defining the filename.
  - **DRY rationale:** The merge **policy** lives here, once. Feeds *both* the eager render (`renderSkill`) and apply-time (`list`/`show`) — so eager and apply-time can never disagree about a repo's type set (the bug whole-dir-symlink masked).
  - **Naming trap to test (plan-review N10):** the type name is the **filename** minus `.md`, NOT the `type:` frontmatter — `product.md` carries `type: type, name: product`. The merge fixture MUST include `product.md` so this invariant is exercised, not just asserted.

- **`DynamicSkill` (weave)** — replaces leaf-only `DynamicSkillDirs`. For repo R, scans **all resolved layers** for skill packages carrying an executable `.dynamic-skill`, cross-referenced against R's *visible* set (`skill.SelectVisible`), returning `{Name (prefixed), Dir (bare), MarkerPath (owner's marker), OutputRel ("construct/generated/<Dir>")}`. The byte-pristine guarantee moves from *leaf-only selection* to *leaf-rooted output*: weave runs the (possibly ancestor-owned) marker with **cwd = R's root** and a **repo-relative `--output`**, so materialization lands in R's tree, never an ancestor's.
  - **DRY rationale:** Reuses the same `skill <dir>` intents `GatherSkills` consumes; one discovery, not a second hardcoded scan.

- **`scanSkillDir` marker-awareness + `SkillSymlinks` lowering switch** — `scanSkillDir` emits an `Entry` for a package dir when it has a `SKILL.md` (static, today's path) **OR** an executable `.dynamic-skill` marker (dynamic, new): a marker-bearing dir yields `Entry{Name, Dynamic:true, BodyPath:<compiling-root>/construct/generated/<Dir>/SKILL.md}` and its description is read from that per-repo body. `skill.Entry` gains `Dynamic bool`. `SkillSymlinks` keeps `Src = filepath.Dir(BodyPath)` for static skills (owner copy) but for dynamic skills sets `Src = <root>/construct/generated/<Dir>` (this-repo's materialized copy). One fork point; every harness face inherits it.
  - **DRY rationale:** Marker presence is the single signal driving BOTH the generate stage and discovery — weave never special-cases "datatype."
  - **Fresh-clone behavior (plan-review I5):** because the entry is emitted from the *tracked marker*, the datatype skill is discovered even in an un-compiled clone (read-only `weave skills`/`golden`/`verify-complete`); only its description body is absent until the first `weave compile`. This is strictly better than #111 (where SKILL.md-keyed discovery would have made the gitignored skill vanish). Documented caveat, not a blocker.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `layergraph.FS` | `pkg/layergraph/fs.go` | new | `os` (ReadFile/ReadDir/Stat) |
| `weavefs.FS` adapter | `cmd/weave/internal/walk/walk.go` | modified | satisfies `layergraph.FS` |
| `weavefs.Runner` (cwd=R root) | `cmd/weave/main.go` | modified | `exec.Command` |
| `datatype` PATH binary build | `Makefile.workflow`, `construct/dev-aliases.sh` | new | build-in-owner distribution |

- **`layergraph.FS`** — minimal filesystem seam the shared walk needs (`ReadFile`, `ReadDir`, `Stat`). Defined in the new package so `cmd/datatype` (which can't import `cmd/weave/internal/weavefs`) has a real-`os` impl, and weave's existing `weavefs.FS` satisfies it structurally (adapter, no behavior change).
  - **Injected into:** `layergraph.Walk` and `mergeTypes`' edge reads. Keeps both pure-testable with an in-memory fake.

- **`weavefs.Runner` run with cwd = R root** — the existing #111 exec seam, unchanged in shape, but the generate stage now invokes the marker with `cwd = root` (the repo being compiled) instead of the package dir, so the binary's DAG walk resolves R's `construct/deps` and the repo-relative `--output` lands in R's tree. Fake-tested (assert cwd == R root, argv == marker).

- **`datatype` PATH binary** — built **in the owner** (`$owner/bin/datatype`) and invoked by name, exactly like `weave`/`sdlc` (Makefile `weave-build` pattern + a `datatype` row in `construct/dev-aliases.sh`). The single change making the marker repo-agnostic. **Not #110's anti-pattern** — a build-in-owner PATH binary is the sanctioned distribution; reading ancestor `construct/datatype/` *data* dirs is exactly what weave's walk already does.

---

## Milestones (review boundaries)

Four boundaries, each its own `sdlc milestone-close`. M1 is a behavior-preserving refactor; M2 makes the binary DAG-aware + on PATH; M3 is the weave surgery (the hardest); M4 migrates + reconciles docs and proves the end state.

---

## Chunk 1: M1 — Extract the shared DAG-walk library

**Outcome:** A module-level `pkg/layergraph` exposes the transitive `construct/deps` walk; `cmd/weave`'s internal `walk.Walk` delegates topology to it; **all existing weave tests stay green** (behavior-preserving). No datatype change yet.

### Task 1.1: `pkg/layergraph` — FS seam + `ParseDeps` (moved)

**Files:** Create `pkg/layergraph/fs.go` (`FS`: ReadFile/ReadDir/Stat), `pkg/layergraph/deps.go` (port of `cmd/weave/internal/layer/deps.go:ParseDeps`), `pkg/layergraph/deps_test.go`.

- [ ] **Step 1:** Copy `ParseDeps` + its tests (`layer/deps_test.go`) verbatim, adjusting package name (pure string→[]string).
- [ ] **Step 2:** `go test ./pkg/layergraph/...` → PASS.
- [ ] **Step 3:** Commit — `#115 M1: pkg/layergraph — ParseDeps + FS seam (moved)`.

### Task 1.2: `resolveOrder` (DFS post-order) moved

**Files:** Create `pkg/layergraph/resolve.go` (port of `layer/resolve.go:Resolve`, signature `Resolve(root string, deps map[string][]string) ([]string, error)`), `resolve_test.go`.

- [ ] **Step 1:** Port the diamond-dedup + cycle-detection tests (foundation-first, leaf last, cycle→error).
- [ ] **Step 2:** Run → fails (absent).
- [ ] **Step 3:** Port `Resolve` verbatim.
- [ ] **Step 4:** `go test ./pkg/layergraph/...` → PASS.
- [ ] **Step 5:** Commit — `#115 M1: pkg/layergraph — resolveOrder (moved)`.

### Task 1.3: `Walk(fs, root) ([]string, error)` — ordered absolute layer roots

**Files:** Create `pkg/layergraph/walk.go` (port `discoverEdges`+`substrateTargets` from `walk/walk.go`, returning ordered roots), `walk_test.go` (in-memory fake FS; ariadne←nous←brain + a diamond).

- [ ] **Step 1:** Failing test: fake `construct/deps` graph; assert foundation-first, leaf last, deduped, `base.manifest`-gating preserved.
- [ ] **Step 2:** Run → fails.
- [ ] **Step 3:** Port to operate over `layergraph.FS`, return `[]string` roots (rich per-layer load stays in weave).
- [ ] **Step 4:** `go test ./pkg/layergraph/...` → PASS.
- [ ] **Step 5:** Commit — `#115 M1: pkg/layergraph — Walk (ordered layer roots)`.

### Task 1.4: weave delegates topology to `pkg/layergraph` (behavior-preserving)

**Files:** Modify `cmd/weave/internal/walk/walk.go` (`Walk` calls `layergraph.Walk` for roots, then `loadLayer` per root; add adapter so `weavefs.FS` satisfies `layergraph.FS`). Delete the now-duplicated `discoverEdges`/`substrateTargets`/`Resolve`/`ParseDeps` bodies in `cmd/weave/internal/{walk,layer}`.

- [ ] **Step 1:** Wire `walk.Walk` → `layergraph.Walk` for topology; keep `loadLayer` for rich fields.
- [ ] **Step 2:** Delete the duplicated walk/resolve/deps code (one source of truth — ARCH-DRY; this is the de-dup the extraction exists for).
- [ ] **Step 3:** `go test ./cmd/weave/...` → PASS (the regression gate; no behavior change).
- [ ] **Step 4:** `go build ./... && go vet ./...` → clean.
- [ ] **Step 5:** `make weave` in ariadne, `git status` → idempotent, clean tree.
- [ ] **Step 6:** Commit — `#115 M1: weave delegates DAG topology to pkg/layergraph`.

### Task 1.5: Extract the frontmatter-description parser into `pkg/frontmatter` (plan-quality Finding 1)

**Why:** `mergeTypes`/`list` (M2) need to read each prototype's `description:`, but `frontmatterDescription`+`unquote` live in `cmd/weave/internal/walk/skills.go` — unreachable from `cmd/datatype` by the same Go-internal rule that justifies `pkg/layergraph`. Lift it once; both weave and datatype call it (ARCH-DRY; one source of truth for "parse a flat-YAML `description:`").

**Files:** Create `pkg/frontmatter/frontmatter.go` (`Description(content string) string` + `unquote`), `frontmatter_test.go` (port weave's existing cases). Modify `cmd/weave/internal/walk/skills.go` — delete the local `frontmatterDescription`/`unquote`, call `frontmatter.Description`.

- [ ] **Step 1:** Move `frontmatterDescription`→`Description` + `unquote` + their tests into `pkg/frontmatter` (pure string→string).
- [ ] **Step 2:** `go test ./pkg/frontmatter/...` → PASS.
- [ ] **Step 3:** weave's `skills.go` calls `frontmatter.Description`; delete the local copies.
- [ ] **Step 4:** `go test ./cmd/weave/...` → PASS (unchanged behavior).
- [ ] **Step 5:** Commit — `#115 M1: pkg/frontmatter — shared flat-YAML description parser`.

- [ ] **M1 — `sdlc milestone-close`** (window = branch point → HEAD). Review checks: extraction truly behavior-preserving, no second walk/parser survives (ARCH-DRY). Fix Critical/Important, log the `Review-Verdict:`.

---

## Chunk 2: M2 — `datatype` becomes a DAG-aware PATH binary

**Outcome:** `cmd/datatype` merges prototypes across the DAG (using `pkg/layergraph`), gains `list` / `show <name>`, and is built + distributed as a PATH binary. Still driven by the *old* marker form until M3 — so M2 is independently testable via direct invocation.

### Task 2.1: `TypeProto` + `mergeTypes` (the merge policy, pure)

**Files:** Create `cmd/datatype/merge.go`, `merge_test.go`.

- [ ] **Step 1:** Failing test — fake layers: ariadne `construct/datatype/{continuation,event,product}.md`, nous `construct/datatype/{event}.md` (shadows), nous leaf `datatype/{trip}.md`. Assert merged = `{continuation(ariadne), event(nous — leaf wins), product(ariadne, name from FILENAME not type:), trip(nous-local)}`, sorted by name, winning `BodyPath`+description each.
- [ ] **Step 2:** Run → fails.
- [ ] **Step 3:** Implement `mergeTypes` — iterate roots foundation-first so leafward writes overwrite the filename key; overlay leaf's `datatype/`; sort by name; filename-without-`.md` naming + `pkg/frontmatter.Description` for the description (the shared parser from M1.5).
- [ ] **Step 4:** `go test ./cmd/datatype/...` → PASS.
- [ ] **Step 5:** Commit — `#115 M2: datatype mergeTypes (DAG union, local-wins-by-filename)`.

### Task 2.2: Rewire `--output` to the merged set; delete `typeNames`

**Files:** Modify `cmd/datatype/main.go` (`--output`: `findRepoRoot(cwd)` → `layergraph.Walk` → `mergeTypes` → `renderSkill` → write; repo-root semantics replace `--datatype-dir`), `datatype.go` (delete `typeNames`), `datatype_test.go` (single-layer case reproduces #111's exact 13-noun byte output — the faithfulness guard).

**Root anchoring (plan-quality Finding 2):** `findRepoRoot` walks up from cwd to the nearest dir containing `construct/` (else `.git`), so apply-time `datatype list`/`show` work when the agent's cwd is a subdirectory — not just at the repo root (where the marker happens to run it). Pure + unit-testable over a fake FS.

- [ ] **Step 1:** Failing tests — `findRepoRoot` from a nested subdir resolves the repo root; the merged render (single-layer) == #111's byte-identical output.
- [ ] **Step 2:** Implement `findRepoRoot` + the merged `--output` path.
- [ ] **Step 3:** `go test ./cmd/datatype/...` → PASS incl. byte-faithfulness.
- [ ] **Step 4:** Commit — `#115 M2: datatype --output renders the DAG-merged noun list`.

### Task 2.3: `datatype list` + `datatype show <name>` (apply-time access)

**Files:** Modify `cmd/datatype/main.go` (subcommand dispatch); create `cmd/datatype/list.go` (`formatList([]TypeProto) string`, pure), `list_test.go`.

- [ ] **Step 1:** Failing tests — `formatList` output shape (name + description per line — the matching surface); `show` resolves the leaf-winning body; `show <unknown>` lists available names + non-zero exit.
- [ ] **Step 2:** Run → fails.
- [ ] **Step 3:** Implement `list`/`show` as thin shells over `mergeTypes` (formatting pure, tested without IO).
- [ ] **Step 4:** `go test ./cmd/datatype/...` → PASS.
- [ ] **Step 5:** Manual smoke — `cd ariadne && go run ./cmd/datatype list` → 13 nouns; `… show continuation` → body.
- [ ] **Step 6:** Commit — `#115 M2: datatype list + show (DAG-resolved apply-time access)`.

### Task 2.4: Build-in-owner distribution (`datatype` on PATH)

**Files:** Modify `Makefile.workflow` (add `datatype-build` mirroring `weave-build` → `$owner/bin/datatype`; ensure the setup/weave path builds it so the marker can call `datatype` by name), `construct/dev-aliases.sh` (add `datatype` ownership row, owner = ariadne).

- [ ] **Step 1:** Add `datatype-build` (location-based owner via `dev-aliases.sh --list`).
- [ ] **Step 2:** Add the alias row; confirm `dev-aliases.sh --list` shows `datatype → <owner>`.
- [ ] **Step 3:** `make datatype-build` → `$owner/bin/datatype` exists; `datatype list` works from any repo on PATH.
- [ ] **Step 4:** `make harness-check` → green.
- [ ] **Step 5:** Commit — `#115 M2: datatype build-in-owner PATH binary`.

- [ ] **M2 — `sdlc milestone-close`.** Review checks: merge policy correct + pure; subcommands DAG-resolve; binary distributed like weave/sdlc; #111 byte-faithfulness preserved for the single-layer case; `product.md` filename-trap exercised.

---

## Chunk 3: M3 — weave generate-stage redesign, marker-aware discovery, lowering switch, gitignore + prune

**Outcome:** weave materializes each repo's own `construct/generated/<dir>/SKILL.md` (gitignored everywhere) by running the (possibly ancestor-owned) marker with cwd = R's root; discovers the dynamic skill from its **marker** and lowers `xx-datatype` to the repo's `construct/generated` copy; GCs orphaned generated dirs. The marker is repo-agnostic. #111's committed ariadne `construct/local/datatype/SKILL.md` is removed (the body now lives, gitignored, under `construct/generated/`).

### Task 3.1: Repo-agnostic marker → `construct/generated/<dir>`, cwd = repo root

**Files:** Modify `construct/local/datatype/.dynamic-skill` — from `go run ../../../cmd/datatype --output . --datatype-dir ../../../construct/datatype` to repo-root-relative: `datatype --output construct/generated/datatype` (PATH binary; cwd = repo root; binary walks `./construct/deps`).

- [ ] **Step 1:** Rewrite the marker + header comment (contract: cwd = repo root, output repo-relative → materialization lands in the *compiling* repo, never an ancestor).
- [ ] **Step 2:** Manual — `cd ariadne && sh construct/local/datatype/.dynamic-skill` writes `ariadne/construct/generated/datatype/SKILL.md` (13-noun render).
- [ ] **Step 3:** Commit — `#115 M3: repo-agnostic .dynamic-skill marker (cwd=repo-root → construct/generated)`.

### Task 3.2: `DynamicSkills` across all layers (visible set), leaf-rooted output

**Files:** Modify `cmd/weave/internal/walk/dynamic.go` — delete leaf-only `DynamicSkillDirs`; add `DynamicSkills(fs, layers, selected []skill.Entry) []DynamicSkill` (`{Name, Dir, MarkerPath, OutputRel:"construct/generated/<Dir>"}`) for each **visible** skill whose source package carries an executable marker. Modify `dynamic_test.go` — an **ancestor**-owned marker IS now selected (the #111 leaf-only test inverts); adapted excluded; non-exec marker ignored; non-visible skill excluded; `OutputRel` uses the bare `Dir` (plan-review I2).

- [ ] **Step 1:** Failing tests (ancestor marker selected; `OutputRel == construct/generated/datatype`; adapted excluded).
- [ ] **Step 2:** Run → fails.
- [ ] **Step 3:** Implement `DynamicSkills` (cross-ref markers against `skill.SelectVisible`).
- [ ] **Step 4:** `go test ./cmd/weave/internal/walk/...` → PASS.
- [ ] **Step 5:** Commit — `#115 M3: DynamicSkills (all-layers, visible-set, construct/generated output)`.

### Task 3.3: Generate stage runs markers with cwd = R root

**Files:** Modify `cmd/weave/main.go:generateDynamicSkills` — for each `DynamicSkill`: `mkdir -p <root>/<OutputRel>` then `runner.Run(root, []string{"sh", <MarkerPath>})` (cwd = `root`). Update the doc comment (byte-pristine now via leaf-rooted output). Modify the generate-stage test — fake Runner asserts cwd == root and the **ancestor** marker is invoked when compiling a derivative; non-zero exit still aborts.

- [ ] **Step 1:** Failing test — compile a derivative fixture (marker owned by ancestor); assert Runner called with cwd = derivative root, output dir created under the derivative, ancestor tree untouched.
- [ ] **Step 2:** Run → fails.
- [ ] **Step 3:** Implement cwd=root exec + mkdir.
- [ ] **Step 4:** `go test ./cmd/weave/...` → PASS.
- [ ] **Step 5:** Commit — `#115 M3: generate stage materializes per-repo (cwd=R root)`.

### Task 3.4: Marker-aware discovery + lowering switch (dynamic → this-repo's `construct/generated`)

**Files:** Modify `cmd/weave/internal/skill/skill.go` (`Entry` gains `Dynamic bool`), `cmd/weave/internal/walk/skills.go` (`scanSkillDir`: a dir with an executable `.dynamic-skill` marker emits `Entry{Dynamic:true, BodyPath:<root>/construct/generated/<Dir>/SKILL.md}` — read description from there; a `SKILL.md`-only dir stays static as today; thread the compiling `root` in), `cmd/weave/internal/plan/skill_symlinks.go` (`SkillSymlinks`: for `Dynamic` entries `Src = <root>/construct/generated/<Dir>`, else owner). Update `skills_test.go`, `skill_symlinks_test.go`.

- [ ] **Step 1:** Failing tests — (a) a marker-bearing dir with NO sibling SKILL.md still yields exactly ONE entry, `Dynamic`, body resolved under `<root>/construct/generated`; (b) **no duplicate `<repo>-datatype`** entry is produced when a consumer's `skill construct/local` row is scanned (the plan-review C1 regression — assert the lowered set has exactly one datatype skill named `xx-datatype`); (c) dynamic lowers to `<root>/construct/generated/datatype`, static lowers to owner.
- [ ] **Step 2:** Run → fails.
- [ ] **Step 3:** Implement marker-awareness + `Dynamic` flag + the switch (thread `root`).
- [ ] **Step 4:** `go test ./cmd/weave/...` → PASS.
- [ ] **Step 5:** Commit — `#115 M3: marker-aware discovery + lowering switch (dynamic → construct/generated)`.

### Task 3.5: Gitignore `construct/generated/` everywhere; drop #111's committed SKILL.md + drift guard

**Files:** Modify `cmd/weave/internal/plan/gitignore.go` (add `/construct/generated/` to the generated-runtime entries), `gitignore_test.go`. `git rm` ariadne's tracked `construct/local/datatype/SKILL.md` (body now under gitignored `construct/generated/`; keep `.dynamic-skill` tracked). Modify `Makefile.workflow` — **remove** `weave-drift-check`'s `git diff --exit-code` form (plan-review I4: a gitignored, regenerated-every-compile output cannot go stale, so the #111 guard's job evaporates). Replace its *value* with what still bites: (a) the renderer's byte-stable unit test (already in M2) and (b) a generate-idempotency assertion (`make weave` twice → identical `construct/generated/datatype/SKILL.md` bytes).

- [ ] **Step 1:** Add the gitignore entry (assert via `gitignore_test.go`).
- [ ] **Step 2:** `git rm --cached construct/local/datatype/SKILL.md`; `make weave`; confirm `git status` clean (body ignored under `construct/generated/`).
- [ ] **Step 3:** Replace `weave-drift-check` with the idempotency assertion (run marker twice, `diff` the two byte outputs).
- [ ] **Step 4:** `go test ./cmd/weave/...` → PASS.
- [ ] **Step 5:** Commit — `#115 M3: gitignore construct/generated; retire committed SKILL.md + stale drift guard`.

### Task 3.6: Prune the generated class (orphan GC when owner drops the marker)

**Files:** Modify `cmd/weave/internal/plan/prune.go` (add `shouldPruneGenerated`: a `construct/generated/<Dir>/` dir for a `<Dir>` no longer in this run's produced dynamic set is removed; never touch a non-generated path), `prune_test.go`.

- [ ] **Step 1:** Failing tests — orphaned `construct/generated/<gone>` removed; an in-use generated dir preserved; nothing outside `construct/generated/` touched.
- [ ] **Step 2:** Run → fails.
- [ ] **Step 3:** Implement `shouldPruneGenerated` (pure decision; IO at edge).
- [ ] **Step 4:** `go test ./cmd/weave/...` → PASS.
- [ ] **Step 5:** `make weave` idempotent + clean in ariadne, nous, pair (a derivative compile materializes its own copy + lowers to it; **assert no `<repo>-datatype` duplicate skill** via `weave skills`).
- [ ] **Step 6:** Add a read-only-path test — `weave verify-complete` stays green for a derivative post-compile (plan-review I6).
- [ ] **Step 7:** Commit — `#115 M3: prune orphaned construct/generated dirs + verify-complete coverage`.

- [ ] **M3 — `sdlc milestone-close`.** Review checks: weave content-blind (marker presence + output-path convention only); byte-pristine via leaf-rooted output (ancestor tree never mutated by a derivative compile — assert directly); **no duplicate skill** (C1 closed); gitignore/prune class correct; drift-guard change is honest (I4).

---

## Chunk 4: M4 — Migration to nous + doc reconciliation + end-to-end proof

**Outcome:** `event`/`travel-plan`/`reference` live in nous; the whole-dir `symlink construct/datatype` is retired and its dangling consumer symlinks removed; the skill body uses the binary for apply-time access; atlas + targets reconciled; an E2E proves nous/brain trigger on `event` while pair does not, with no duplicate skills.

### Task 4.1: Move the three prototypes ariadne → nous; retire the whole-dir symlink

**Files:** Cross-repo move (ariadne `git rm` + nous add): `construct/datatype/{event,travel-plan,reference}.md` → `nous/construct/datatype/`. Modify `ariadne/construct/base.manifest` (delete the `symlink construct/datatype` row + comment block — prototypes are no longer a weave-lowered artifact). Manually remove the dangling `nous|brain|pair/construct/datatype` whole-dir symlinks (plan-review N9: `construct/` is **not** a managed skill location, so `PruneOrphans` will NOT auto-remove them — they must be `rm`'d explicitly).

- [ ] **Step 1:** Move the three files (ariadne loses them; nous gains a real `construct/datatype/`).
- [ ] **Step 2:** Delete the manifest row; `rm nous/construct/datatype brain/construct/datatype pair/construct/datatype` (the symlinks).
- [ ] **Step 3:** Re-weave **foundation-first** (ariadne → nous → brain; order is for tidiness only — `mergeTypes` reads source prototype dirs, not ancestors' materialized copies, so correctness is order-independent, plan-review I3). Confirm no consumer carries a stale prototype copy or dangling symlink.
- [ ] **Step 4:** Commit (two commits, cross-repo) — `#115 M4: migrate event/travel-plan/reference ariadne→nous; retire whole-dir datatype symlink`.

### Task 4.2: SKILL.md template body → binary-based apply-time access

**Files:** Modify `cmd/datatype/SKILL.md.tmpl` — rewrite **§Type lookup** (~49–60), the enumeration in **§When to use Step 3**, and **§Adding a new type**: replace raw `ls construct/datatype/` + `<repo>/datatype/` + `awk` with `datatype list` (enumerate) and `datatype show <name>` (read a body); replace "Prototypes live in two places" with "the `datatype` binary resolves across the layer DAG, local/leaf shadows shared." Keep authoring locations: shared-downward → `<layer>/construct/datatype/<name>.md`; leaf-only override → `<repo>/datatype/<name>.md`.

- [ ] **Step 1:** Rewrite the affected sections (binary-based; DAG-correct).
- [ ] **Step 2:** `make weave`; confirm each repo's materialized `construct/generated/datatype/SKILL.md` carries the new body + correct per-repo noun list.
- [ ] **Step 3:** Commit — `#115 M4: SKILL.md body uses datatype list/show (DAG-correct apply-time)`.

### Task 4.3: Reconcile atlas + targets (+ fix stale Makefile comment N8)

**Files:** Modify `atlas/workflow/weave.md` (dynamic-skill section: `construct/generated` materialization gitignored everywhere, marker-aware discovery, cwd=R-root marker exec, lowering switch, prune class; note `pkg/layergraph` as the shared walk; fix the stale `weave compile --target claude` comment, plan-review N8 — the recipe runs the bare Union). Modify `workshop/targets/skill-system.md` (committed-codegen → gitignored `construct/generated` per-repo materialization; marker-aware discovery; lowering switch; prune/gitignore class). Modify `workshop/targets/base-layer-mechanics.md` (datatype prototypes: whole-dir-symlink → DAG-walked-by-binary, no longer a file-op artifact; record `pkg/layergraph` as the shared topology library — the pensive's *single mechanism* realized cross-tool, first second-consumer of the walk). Confirm `atlas/index.md` links any new file.

- [ ] **Step 1:** Reconcile all docs to the new truth (atlas = current state only; chase stale refs to ground — brain `atlas-current-state-only` lesson).
- [ ] **Step 2:** Commit — `#115 M4: reconcile atlas + skill-system + base-layer-mechanics`.

### Task 4.4: End-to-end migration-correctness proof

- [ ] **Step 1:** `make weave` in ariadne, nous, pair, brain.
- [ ] **Step 2: Eager** — `nous/construct/generated/datatype/SKILL.md` description includes `event, reference, travel-plan` + ariadne's 10; `brain/...` includes them (brain→nous→ariadne); `pair/...` and `ariadne/...` do **not** (only 10).
- [ ] **Step 3: Apply-time** — `cd nous && datatype list` → 13 (incl. the 3); `cd pair && datatype list` → 10; `datatype show event` works in nous, errors (with available names) in pair.
- [ ] **Step 4: No duplicate (C1 gate)** — `weave skills` in nous/brain/pair lists exactly ONE datatype skill named `xx-datatype`; no `nous-datatype`/`brain-datatype`/`pair-datatype`.
- [ ] **Step 5: Hygiene** — `git status` clean in every repo post-weave (generated tree gitignored; no stray prototype copies / dangling symlinks); `make harness-check` green; `go build ./... && go test ./... && go vet ./...` green.
- [ ] **Step 6:** Commit — `#115 M4: E2E migration-correctness proof`.

- [ ] **M4 / issue close — `sdlc close --issue 115`** (auto-dispatches the end-of-issue integration review over the whole M1–M4 diff). Provide `--verified` (the E2E evidence) + measured `--actual`. Atlas reconciled in 4.3 so the atlas gate passes.

---

## Decided design points (resolved during planning + review)

- **D1 — Materialization sink = `construct/generated/<dir>/` (not `construct/local/`).** Closes plan-review C1: `construct/local/` is re-scanned by a consumer's own `skill construct/local` row → a duplicate `<repo>-datatype`. `construct/generated/` is never skill-scanned; discovery is marker-driven. Uniform owner/consumer.
- **D2 — Drift guard retired, not replaced 1:1 (plan-review I4).** A gitignored, every-compile-regenerated output cannot go stale; the residual value (deterministic render + idempotency) is covered by the M2 byte-stable unit test + the M3.5 idempotency assertion. Don't ship a tautological `diff` that always passes.
- **D3 — Fresh-clone read-only paths (plan-review I5).** Marker-driven discovery means the skill is found from the tracked marker even pre-compile; only the description body is absent until first compile. Documented caveat; CI compiles before `golden`/`verify-complete`.

## Revisions

### 2026-06-18 — folded in fresh-eyes plan review (Critical C1 + I2–I6, N7–N10)
- **C1 (Critical):** materialization moved `construct/local/<dir>` → **`construct/generated/<dir>`** + **marker-aware discovery**, to avoid the duplicate-skill collision with a consumer's own `skill construct/local` scan. Reshaped M3.4/3.5/3.6 and the Architecture section. (This refinement also resolves I5 — discovery now keys on the tracked marker, so the skill never vanishes in a fresh clone.)
- **I2:** explicit `<dir>` (bare, for output path/Src) vs `<Name>` (prefixed, for symlink Dst) discipline.
- **I3:** noted `mergeTypes` reads ancestor *source* prototype dirs (not materialized copies) → re-weave order immaterial to correctness.
- **I4:** drift guard retired (D2), not replaced with a tautological diff.
- **I6:** added `verify-complete`-green test for a derivative (M3.6 step 6).
- **N8/N9/N10:** stale `--target claude` Makefile comment fix (M4.3); explicit `rm` of dangling consumer `construct/datatype` symlinks (M4.1); `product.md` filename-trap in the merge fixture (M2.1).

### 2026-06-18 — folded in change-code plan-quality + estimate-quality judges (both INFO)
- **Finding 1 (ARCH-DRY):** the `frontmatterDescription`+`unquote` parser has the *same* import-boundary problem as the walk (lives in `cmd/weave/internal/walk/skills.go`, unreachable from `cmd/datatype`). Added **Task 1.5** — extract to a shared `pkg/frontmatter`; weave delegates; M2's merge/list use it. One parser, not two.
- **Finding 2 (robustness):** `datatype list`/`show` may run from a subdir → added `findRepoRoot` (walk up to nearest `construct/`) so apply-time access anchors the repo root, not the agent's cwd (M2.2).
- Estimate-quality (INFO): design hours slightly high vs the pre-resolved plan, offset by M4 being light; total 12.0 stands.

### 2026-06-18 — M1 boundary review folded in (FIX-THEN-SHIP, no Critical)
- Entity-table name corrected `resolveOrder` → **`Resolve`** (matches the shipped symbol; the table's greppability is the point).
- Important (atlas): added a 2-line `atlas/index.md` pointer for `pkg/layergraph` + `pkg/frontmatter` now (full docs still batched to M4 per Task 4.3); M1 closed `--no-atlas` consciously (relocation-only).
- Minors: `pkg/layergraph/walk_test.go` now uses production `OSFS{}` (deleted the duplicate `testFS`, gives OSFS coverage); stale `discoverEdges' BFS` comment in weave's `walk_test.go` reworded.
- **M3 consideration (review §6):** `cmd/weave/walk.go` derives `canonRoot = order[len-1]`, an invariant coupling to `layergraph.Resolve`'s "root emitted last" post-order. When M3 touches this seam, weigh having `layergraph.Walk` return the canonical root explicitly (`(roots, canonRoot, err)`) instead of re-deriving it positionally, so a future emission-order change can't silently mis-target weave's self-reference filter.

### 2026-06-18 — M3 boundary review folded in (FIX-THEN-SHIP, no Critical)
- **Core-concepts table corrected (review #4):** `SkillSymlinks` is **unchanged**, not "modified." The lowering switch is realized in `scanSkillDir`'s `BodyPath` — a dynamic entry's `BodyPath = <root>/construct/generated/<dir>/SKILL.md`, so `SkillSymlinks`' existing `Src = filepath.Dir(BodyPath)` lowers to the per-repo materialized copy with **no branch** (more DRY than the plan's "SkillSymlinks branch" sketch; the reviewer called it "genuinely more DRY than the plan envisioned").
- **ARCH-DRY consolidated AT the M3 boundary (reviews #1/#2 — not deferred):** single-sourced the `construct/generated` convention via `walk.GeneratedRel` + `GeneratedSkillDir`/`GeneratedSkillBody`, used by the write-path (`dynamic.go`), read-path (`skills.go`), prune (`prune.go`), and gitignore (`gitignore.go`) — they can no longer diverge. Extracted the dynamic-marker predicate into `walk.dynamicMarker`, shared by `DynamicSkills` + `scanSkillDir`.
- **Faithfulness test hardened (review #3):** `TestRenderSkill_FaithfulToTemplate` now asserts the placeholder is present in the template, absent in the render, and the noun list appears — closing the vacuous-pass hole (a removed placeholder would otherwise ship a noun-less skill green).
- **Minor:** `runner.go` doc comment corrected (cwd = repo root, not package dir). **Deferred to M4 cleanup:** `DynamicSkills` is computed twice per compile (generate + prune) — compute once; remove the now-dead `--datatype-dir` flag; dry-run doesn't preview the generated-class prune (comment overstates).

### 2026-06-18 — issue-close integration review folded in (FIX-THEN-SHIP, no Critical)
- **I1 (Important) — peer-symlink cleanup DONE:** M4.1 Step 2's removal of the retired `construct/datatype` whole-dir symlink had been missed in 5 peers (`pair`/`brain`/`42shots`/`you-decide`/`xianxu.dev`), contradicting the "no dangling symlinks" close evidence (`nous` was correctly a real dir). All 5 `git rm`'d + committed; `datatype list` re-verified (ariadne 10, nous 13, brain 13, pair 10); zero `construct/datatype` symlinks remain.
- **M3-deferred cleanups LANDED (no longer deferred):** `walk.DynamicSkills` now computed ONCE in `run()` and reused by the generate stage + the generated-class prune (`generateDynamicSkills` takes the precomputed set); the dead `--datatype-dir` flag removed from `cmd/datatype`; the dry-run prune-preview comment corrected to note `PruneGenerated` isn't previewed (harmless — its targets are gitignored).
- **Not actioned (review §6 watch-item, not a finding):** `skill.Entry.Dynamic` is currently write-only (the BodyPath switch does the lowering); left as a documented future-cleanup candidate, not removed, to avoid test churn at the close boundary.
