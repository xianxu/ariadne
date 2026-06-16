# skill-system v2 — unified, visibility-aware skill composition — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make weave's skill composition ONE intent-driven, visibility-aware discovery that feeds both lowerings (claude `.claude/skills` symlinks + codex/agy menu) identically — replacing the three disagreeing mechanisms — then migrate every repo onto per-layer skill dirs + repo-name prefixes.

**Architecture:** Today three code paths discover skills and none consult visibility (issue #104 §A/§B). v2 collapses them: `walk.GatherSkills` becomes the single IO discovery (reads each layer's `skill <dir>` intents, carrying each skill's `Visibility` + `LayerIndex`); a pure `skill.SelectVisible` applies the SAME `intent.Selected` 𝒜(R) filter prose uses; the menu (`skill.Build`) and the claude symlinks (new pure `plan.SkillSymlinks`) both derive from that one selected set. The separate `walk.LowerSkillSymlinks` IO scan is deleted (ARCH-DRY). Then a cross-repo migration drops the whole-dir `construct/{local,adapted}→ariadne` inheritance symlinks, gives each layer its own `config.json` (repo-name prefix default), and moves nous's skills into the convention.

**Tech Stack:** Go (`cmd/weave`), the existing `weavefs.FS` IO seam, `t.TempDir`-rooted OSFS tests (no mocks). Manifest/config are plain text + JSON. The migration re-runs `make weave` per repo (the M5 cutover machinery).

**Milestones (review boundaries):** M1 = the unified discovery + visibility (weave code only, behavior-preserving for today's two dirs); M2 = per-layer prefix + the `internal skill construct/skill` declaration; M3 = the cross-repo migration. M1 is task-detailed below; M2–M3 are task-sketched (detail them at their `sdlc start-plan`/`change-code`, per the #128 plan convention).

---

## Core concepts

The skill subsystem's conceptual model. The pure core is "a set of `Entry`s, selected by visibility, rendered two ways"; the only IO is reading SKILL.md/config.json off disk in one seam.

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `skill.Entry` (+ `Visibility`, `LayerIndex`) | `cmd/weave/internal/skill/skill.go` | modified |
| `skill.SelectVisible` | `cmd/weave/internal/skill/skill.go` | new |
| `plan.SkillSymlinks` | `cmd/weave/internal/plan/skill_symlinks.go` | new |
| `skill.Build` | `cmd/weave/internal/skill/skill.go` | (unchanged — now fed selected entries) |
| `intent.Selected` | `cmd/weave/internal/intent/intent.go` | (unchanged — reused for skills, ARCH-DRY) |

- **`skill.Entry`** — one parsed skill. v2 adds **`Visibility intent.Visibility`** (from the declaring `skill <dir>` row's `export|internal` token) and **`LayerIndex int`** (which layer in the foundation-first walk declared it). Name/Description/BodyPath unchanged.
  - **Relationships:** N:1 with a layer (a layer declares many entries); the `(Visibility, LayerIndex)` pair is exactly the input `intent.Selected` needs.
  - **DRY rationale:** Carrying visibility ON the entry lets the SAME `intent.Selected` that filters prose/file-ops filter skills — no skill-specific visibility logic. The entry is the single skill record both lowerings read; today the menu path and the symlink path build *different* records from *different* scans.
  - **Future extensions:** A `SourceRow` field if per-skill (frontmatter) visibility ever overrides the per-dir convention (issue §C3 / open question).

- **`skill.SelectVisible(entries []Entry, leafIdx int) []Entry`** — the pure 𝒜(R) filter for skills: keep `e` iff `intent.Selected(e.Visibility, e.LayerIndex == leafIdx)`. Foundation-first order preserved.
  - **Relationships:** 1:1 with the gathered entries → selected subset; consumed by both `Build` (menu) and `SkillSymlinks` (claude).
  - **DRY rationale:** This is the skill instantiation of the base-layer-mechanics visibility axis. It REUSES `intent.Selected` (the one selection rule the planner + completeness guard already use) — no parallel skill filter. Eliminates §B1 (skills consulting visibility "in two more places, differently").
  - **Future extensions:** Suppression (issue open question B2) would extend the predicate, not add a second pass.

- **`plan.SkillSymlinks(entries []skill.Entry) []Symlink`** — pure: maps each selected entry to `Symlink{Src: filepath.Dir(e.BodyPath), Dst: ".claude/skills/" + e.Name}`. The claude lowering, derived from the same entries as the menu.
  - **Relationships:** 1:1 entry→Symlink. Lives in `plan` (where `Symlink` lives) to avoid a `skill`→`plan` import cycle (`plan` already imports `skill`).
  - **DRY rationale:** Replaces the WHOLE `walk.LowerSkillSymlinks` IO function, whose separate dir-scan was §A4's "two paths, two operand sets." Now there is one scan (GatherSkills) and two pure renderings.
  - **Future extensions:** A `--target` that wants a third rendering adds a function here, not a fourth discovery.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `walk.GatherSkills` (intent-driven) | `cmd/weave/internal/walk/skills.go` | modified | filesystem (SKILL.md, config.json) |
| `walk.LowerSkillSymlinks` | `cmd/weave/internal/walk/skill_symlinks.go` | deleted | (was the duplicate scan) |
| `walk.skillPrefix` | `cmd/weave/internal/walk/skills.go` | modified | config.json + repo name |

- **`walk.GatherSkills(fs, layers) ([]skill.Entry, error)`** — the SINGLE skill discovery. v2: for each layer, iterate its `skill <dir>` **intents** (not the hardcoded `construct/local`+`construct/adapted`), scan each declared dir for `<name>/SKILL.md`, and emit one `Entry` per skill carrying the discovery-derived name (prefix rule below), the frontmatter description, the BodyPath, the row's `Visibility`, and the layer's index.
  - **Injected into:** `main.buildSkillIndex` (→ `SelectVisible` → `Build`) and the claude lowering (→ `SelectVisible` → `plan.SkillSymlinks`). The pure entities receive its output; it's the only place that touches disk for skills.
  - **Future extensions:** A new provider (a non-`construct/*` skill source) is just another `skill <dir>` row — no code change.

- **`walk.skillPrefix(fs, layerRoot, repoName) string`** — resolves a layer's skill prefix. v2: `config.json localPrefix` if present and non-empty, else **the layer's repo name + "-"** (issue decision 4), still `xx-` for ariadne (its own `config.json` sets it). The prefix applies to a layer's OWN dirs (`construct/local`, `construct/skill`); `construct/adapted` skills stay bare (external names preserved). The bare-vs-prefixed rule keys on the dir: `adapted` → bare, else → prefix.
  - **Injected into:** `GatherSkills` (per layer, per dir).
  - **Future extensions:** repo-name derivation source (dir basename vs go.mod module) is the one knob; default to the layer-root dir basename.

**Test surface.** Every pure entity gets a colocated `_test.go` running against a `t.TempDir`-rooted OSFS (the repo's established pattern — no function mocks). The load-bearing new tests: an `internal skill <dir>` of an ANCESTOR does not reach a consumer but DOES on the ancestor's self-walk (visibility); a `skill <dir>` outside `construct/local|adapted` appears in BOTH the menu and the `.claude/skills` set (the §A2/§A4 unification); a layer with no `config.json` prefixes by repo name (§C2).

---

## Chunk 1: M1 — unified intent-driven discovery + visibility

**Outcome:** One discovery feeds both lowerings; skills honor export/internal + leaf-position. **Behavior-preserving for ariadne today** (its manifest already declares `skill construct/local` + `skill construct/adapted`, both export) — the golden self-walk + the live `.claude/skills` set are unchanged; the NEW capability (intent-driven dirs, visibility, one operand set) is what the tests pin. No repo migration in M1.

### Task 1: `skill.Entry` carries visibility + layer index

**Files:**
- Modify: `cmd/weave/internal/skill/skill.go` (the `Entry` struct)
- Test: `cmd/weave/internal/skill/skill_test.go`

- [ ] **Step 1: Write the failing test** — `skill.Build` still works, and an `Entry` can hold visibility + layer index.

```go
func TestEntryCarriesVisibilityAndLayer(t *testing.T) {
	e := Entry{Name: "xx-fix", Description: "d", BodyPath: "/a/fix/SKILL.md",
		Visibility: intent.Internal, LayerIndex: 2}
	if e.Visibility != intent.Internal || e.LayerIndex != 2 {
		t.Fatalf("entry did not carry visibility/layer: %+v", e)
	}
}
```

- [ ] **Step 2: Run — expect FAIL** (`unknown field Visibility`).
  Run: `go test ./cmd/weave/internal/skill/ -run TestEntryCarriesVisibility -v`

- [ ] **Step 3: Add the fields** (import `intent`):

```go
type Entry struct {
	Name        string
	Description string
	BodyPath    string
	Visibility  intent.Visibility // from the declaring `skill <dir>` row (export default)
	LayerIndex  int               // foundation-first index of the declaring layer
}
```

- [ ] **Step 4: Run — expect PASS.** Run: `go test ./cmd/weave/internal/skill/ -v`
- [ ] **Step 5: Commit.** `git commit -am "#104 M1: skill.Entry carries Visibility + LayerIndex"`

### Task 2: `skill.SelectVisible` — the pure 𝒜(R) skill filter

**Files:**
- Modify: `cmd/weave/internal/skill/skill.go`
- Test: `cmd/weave/internal/skill/skill_test.go`

- [ ] **Step 1: Write the failing test** — an ancestor's internal is dropped; the leaf's internal + everyone's export are kept; order preserved.

```go
func TestSelectVisible(t *testing.T) {
	leaf := 2
	in := []Entry{
		{Name: "base-export", Visibility: intent.Export, LayerIndex: 0},
		{Name: "ancestor-internal", Visibility: intent.Internal, LayerIndex: 0}, // DROP
		{Name: "leaf-internal", Visibility: intent.Internal, LayerIndex: 2},     // KEEP
		{Name: "leaf-export", Visibility: intent.Export, LayerIndex: 2},
	}
	got := names(SelectVisible(in, leaf))
	want := []string{"base-export", "leaf-internal", "leaf-export"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SelectVisible = %v, want %v", got, want)
	}
}
```
(add a `names([]Entry) []string` test helper.)

- [ ] **Step 2: Run — expect FAIL** (`undefined: SelectVisible`).
- [ ] **Step 3: Implement** (reuse `intent.Selected` — ARCH-DRY):

```go
// SelectVisible keeps the entries that participate in 𝒜(R): every layer's
// exports plus the leaf's internals (leafIdx). The single source of the
// visibility rule is intent.Selected — the same predicate prose + file-ops use.
func SelectVisible(entries []Entry, leafIdx int) []Entry {
	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if intent.Selected(e.Visibility, e.LayerIndex == leafIdx) {
			out = append(out, e)
		}
	}
	return out
}
```

- [ ] **Step 4: Run — expect PASS.**
- [ ] **Step 5: Commit.** `git commit -am "#104 M1: skill.SelectVisible reuses intent.Selected for the skill 𝒜(R) filter"`

### Task 3: `GatherSkills` becomes intent-driven + carries visibility/layer

**Files:**
- Modify: `cmd/weave/internal/walk/skills.go` (`GatherSkills`, `scanSkillDir`, `localPrefix`)
- Test: `cmd/weave/internal/walk/skills_test.go`

- [ ] **Step 1: Write the failing test** — a layer that declares `skill <dir>` for a dir OTHER than construct/local|adapted has its skills gathered, and an `internal skill <dir>` row stamps `Internal` on those entries; layer index is set.

```go
func TestGatherSkillsIntentDriven(t *testing.T) {
	// layer with: export skill construct/mine, internal skill construct/priv
	root := t.TempDir()
	writeSkill(t, root, "construct/mine/tool", "tool desc")
	writeSkill(t, root, "construct/priv/secret", "secret desc")
	layers := []layer.Layer{{Path: root, Intents: []intent.Intent{
		{Kind: intent.Skill, Visibility: intent.Export, Source: "construct/mine"},
		{Kind: intent.Skill, Visibility: intent.Internal, Source: "construct/priv"},
	}}}
	got, err := GatherSkills(weavefs.OSFS{}, layers)
	// expect: an Export entry for construct/mine/tool, an Internal entry for construct/priv/secret;
	// both with LayerIndex 0; hardcoded construct/local|adapted NOT auto-scanned.
}
```
(`writeSkill` helper: mkdir `<root>/<dir>/` + a SKILL.md with frontmatter `description:`.)

- [ ] **Step 2: Run — expect FAIL** (GatherSkills ignores intents today).
- [ ] **Step 3: Rewrite `GatherSkills`** to iterate `l.Intents` where `Kind == intent.Skill`, scanning `in.Source`; stamp `Visibility = in.Visibility` and `LayerIndex = i` on each entry; apply the prefix per the dir rule (Task 4 refines the prefix). Delete the hardcoded `localSkillRel`/`adaptedSkillRel` scan loop. `scanSkillDir` gains a `visibility`+`layerIdx` parameter (or the caller stamps them post-scan).

```go
func GatherSkills(fs weavefs.FS, layers []layer.Layer) ([]skill.Entry, error) {
	var entries []skill.Entry
	for i, l := range layers {
		prefix := skillPrefix(fs, l.Path) // Task 4
		for _, in := range l.Intents {
			if in.Kind != intent.Skill {
				continue
			}
			rowPrefix := prefix
			if in.Source == adaptedSkillRel { // external skills keep their bare names
				rowPrefix = ""
			}
			es, err := scanSkillDir(fs, filepath.Join(l.Path, in.Source), rowPrefix)
			if err != nil {
				return nil, err
			}
			for j := range es {
				es[j].Visibility = in.Visibility
				es[j].LayerIndex = i
			}
			entries = append(entries, es...)
		}
	}
	return entries, nil
}
```

- [ ] **Step 4: Run — expect PASS.** Also run `go test ./cmd/weave/internal/walk/ -v` (the existing GatherSkills tests must be updated to declare intents, not rely on hardcoded dirs).
- [ ] **Step 5: Commit.** `git commit -am "#104 M1: GatherSkills is intent-driven; stamps Visibility + LayerIndex"`

### Task 4: per-dir prefix rule (still config.json-driven; repo-name default lands in M2)

**Files:**
- Modify: `cmd/weave/internal/walk/skills.go` (`localPrefix` → `skillPrefix`)
- Test: `cmd/weave/internal/walk/skills_test.go`

- [ ] **Step 1: Write the failing test** — `construct/adapted` skills are bare; a layer's own dir gets the `config.json` prefix (still `xx-` default in M1).
- [ ] **Step 2: Run — expect FAIL.**
- [ ] **Step 3: Rename `localPrefix`→`skillPrefix`** (reads `config.json localPrefix`, default `xx-` for now — the repo-name default is M2 Task). The adapted-bare rule lives in `GatherSkills` (Task 3). Keep `defaultPrefix = "xx-"` for M1 so ariadne is unchanged.
- [ ] **Step 4: Run — expect PASS.**
- [ ] **Step 5: Commit.**

### Task 5: `plan.SkillSymlinks` (pure) + delete `walk.LowerSkillSymlinks`

**Files:**
- Create: `cmd/weave/internal/plan/skill_symlinks.go`
- Create: `cmd/weave/internal/plan/skill_symlinks_test.go`
- Delete: `cmd/weave/internal/walk/skill_symlinks.go` + its test

- [ ] **Step 1: Write the failing test** — entries → `.claude/skills/<name>` Symlinks pointing at `Dir(BodyPath)`.

```go
func TestSkillSymlinks(t *testing.T) {
	got := SkillSymlinks([]skill.Entry{
		{Name: "xx-fix", BodyPath: "/ws/ariadne/construct/local/fix/SKILL.md"},
		{Name: "nous-tools", BodyPath: "/ws/nous/construct/local/tools/SKILL.md"},
	})
	want := []Symlink{
		{Src: "/ws/ariadne/construct/local/fix", Dst: ".claude/skills/xx-fix"},
		{Src: "/ws/nous/construct/local/tools", Dst: ".claude/skills/nous-tools"},
	}
	if !reflect.DeepEqual(got, want) { t.Fatalf("got %v want %v", got, want) }
}
```

- [ ] **Step 2: Run — expect FAIL.**
- [ ] **Step 3: Implement** the pure function; **delete** `walk/skill_symlinks.go` + `walk/skill_symlinks_test.go`.

```go
// SkillSymlinks lowers selected skill entries to the claude .claude/skills/<name>
// links — pure, derived from the SAME entries the menu uses (no second scan).
func SkillSymlinks(entries []skill.Entry) []Symlink {
	out := make([]Symlink, 0, len(entries))
	for _, e := range entries {
		out = append(out, Symlink{Src: filepath.Dir(e.BodyPath), Dst: filepath.Join(".claude", "skills", e.Name)})
	}
	return out
}
```

- [ ] **Step 4: Run — expect FAIL elsewhere** (`main.go` + `golden` still call `walk.LowerSkillSymlinks`). That's the next task.
- [ ] **Step 5: Commit** after Task 6 wires it (or commit the compile-broken state on a WIP branch — prefer wiring first).

### Task 6: rewire `main.planActions` + `buildSkillIndex` to the ONE discovery

**Pin the seam exactly (plan-quality finding #1 — this is the spot the whole issue turns on).** `buildSkillIndex` must return BOTH the index AND the selected `[]skill.Entry` it was built from, so the menu (`skill.Build`) and the claude symlinks (`plan.SkillSymlinks`) read the IDENTICAL slice from ONE `GatherSkills`. If an implementer instead leaves `buildSkillIndex → SkillIndex` and adds a SECOND `GatherSkills`+`SelectVisible` before `SkillSymlinks`, the §A4 duplication silently survives and tests still pass — the issue's core claim is unmet. Do not do that.

**Files:**
- Modify: `cmd/weave/main.go` (`buildSkillIndex`, `planActions`, `resolveSkillIndex`; delete the `skillSymlinks` helper)
- Modify: `cmd/weave/internal/golden/*` if it calls `walk.LowerSkillSymlinks`

- [ ] **Step 1: Change `buildSkillIndex` to return the selected entries too.**

```go
func buildSkillIndex(fs weavefs.FS, layers []layer.Layer) (skill.SkillIndex, []skill.Entry, error) {
	entries, err := walk.GatherSkills(fs, layers)
	if err != nil {
		return skill.SkillIndex{}, nil, err
	}
	selected := skill.SelectVisible(entries, len(layers)-1) // 𝒜(R): leaf = last layer
	return skill.Build(selected), selected, nil
}
```

- [ ] **Step 2: `planActions` consumes ONE gather.** `idx, selected, err := buildSkillIndex(fs, layers)`. Menu from `idx.Menu()` when `target.IncludeSkillMenu()`. For `target.EmitSkillSymlinks()`: `actions = append(actions, asActions(plan.SkillSymlinks(selected))...)` — the SAME `selected`. Delete the `skillSymlinks(fs, layers)` helper and its `walk.LowerSkillSymlinks` call.
- [ ] **Step 3: `resolveSkillIndex`** (serves `weave skills`/`weave skill <name>`) uses the `SkillIndex` from `buildSkillIndex` (discard the entries) — so a *served* skill is exactly a *composed* one (same select, no third path).
- [ ] **Step 4: Run the full suite** `go test ./cmd/weave/...` + `go vet ./cmd/weave/...` + `gofmt -l cmd/weave` — all green. Fix any `golden` caller of the deleted `walk.LowerSkillSymlinks`.
- [ ] **Step 5: Live verify (behavior-preserving):** in ariadne `weave compile --target claude` then `weave golden --target claude` → MATCH unchanged / 0 UNEXPECTED; `weave skills` lists the same set; re-weave diff empty.
- [ ] **Step 6: Commit.** `git commit -am "#104 M1: one discovery → SelectVisible → {Build menu, SkillSymlinks claude}; delete LowerSkillSymlinks"`

### Task 7: M1 milestone close

**Done-when scoping (plan-quality finding #2):** M1 closes the *unification* (one discovery → two lowerings) and the *pure* visibility predicate (`SelectVisible` test). It does NOT close the issue's full Done-when — the multi-layer end-to-end internal-skill assertion ("an ancestor's internal does NOT reach a consumer, the leaf's DOES") is **M2 Task B** (needs a real `internal skill` declared), and the formula-binding golden is **M3**. State this in the M1 close so M1 isn't held to the whole list.

- [ ] Update `atlas/workflow/weave.md` (skill discovery is now intent-driven + visibility-aware — one operand set feeds menu + symlinks).
- [ ] `sdlc milestone-close --issue 104 --milestone M1 --verified 'one intent-driven GatherSkills → SelectVisible → {Build menu, SkillSymlinks claude}; LowerSkillSymlinks deleted; ariadne golden unchanged (behavior-preserving); SelectVisible pure predicate test green'` (dispatches the boundary review; fix Critical/Important).

---

## Chunk 2: M2 — per-layer prefix + the `internal skill construct/skill` declaration (SKETCH)

**Outcome:** A layer's prefix defaults to its repo name; the `construct` skill becomes internal-by-declaration; the three-dir convention is real in ariadne's manifest. Still no derivative migration (those land in M3). Detail at M2's `change-code`.

- **Task A — repo-name prefix default. DONE** (`3655a0e`). `skillPrefix` returns
  `config.json localPrefix` when set, else `<layer-dir-basename> + "-"`. ariadne
  keeps `xx-` via its own `config.json`; behavior-preserving while derivatives'
  `config.json` is still symlinked to ariadne's (M3 un-symlinks it). Tests:
  `RepoNamePrefixWhenNoConfig` (repo-name) + the existing config-override cases;
  ariadne golden MATCH 25/0 unchanged. (Issue decision 4; §C2.)
- **Task B — internal-skill 𝒜(R) end-to-end. DONE** (`3655a0e`). The M1 review's
  deferred finding: `TestBuildSkillIndexExcludesAncestorInternalSkill` — a 2-layer
  fixture (base declares `internal skill construct/skill`) routed through
  `walk → buildSkillIndex` (NOT a direct `SelectVisible` call), proving the
  `len(layers)-1` leaf-index plumbing: an ancestor's internal skill is excluded
  from the consumer but present on the base's own self-walk. (§B1 closed.)
- **Task C — M2 close** (atlas + milestone-close).

> **Moved to M3:** declaring `internal skill construct/skill` in ariadne's *real*
> manifest + the `construct/skill` layout move + the construct-skill name
> (`construct` vs `xx-construct`) + the tracked `.claude/skills/construct` copy —
> these are ariadne-manifest/disk changes that belong with M3's migration (and the
> name is operator-facing, so it gets decided there). M2's `internal skill`
> CAPABILITY is built + proven (Task B); M3 applies it to the real construct skill.

**Open at M2:** the `construct/skill` name (bare `construct` vs `xx-construct`) and whether `.claude/skills/construct` (a tracked real copy per the self-sync rule) stays tracked. Resolve in the M2 plan.

---

## Chunk 3: M3 — the cross-repo migration (SKETCH)

**Outcome:** Every derivative owns its skill dirs + prefix; nous's skills live in `construct/local`; no inheritance-by-symlink. This is a coordinated re-weave like the #95 M5 cutover — same machinery, same safety (ancestors byte-pristine; sandbox-aware on the brains). Detail at M3's `change-code`.

- **Task A — drop the whole-dir inheritance symlinks.** Remove `symlink construct/local` + `symlink construct/adapted` from ariadne's `base.manifest` (so derivatives stop receiving the dir symlinks); the layer-walk reads each ancestor's real dirs directly. Per derivative: `git rm` the now-orphaned `construct/{local,adapted}` symlinks (weave's prune handles them on re-weave). Verify `weave skills` still lists ariadne's skills in every derivative (sourced from the ariadne LAYER, not a local symlink) — and `.claude/skills/xx-*` now point DIRECTLY at `../../../ariadne/construct/local/<name>` (already weave's output, issue D1).
- **Task B — per-layer config.json.** Un-symlink each derivative's `construct/config.json` (currently → ariadne's `xx-`). nous gets its own with no `localPrefix` (→ `nous-` default) or explicit `localPrefix: nous-`. Re-weave: nous's local skills become `nous-*`.
- **Task C — migrate nous `construct/skills` → `construct/local`.** `git mv construct/skills/nous-tools construct/local/tools` + `nous-resolve`→`resolve`; replace the two plain `symlink construct/skills/X .claude/skills/X` manifest rows with one `skill construct/local` (export) intent; re-weave → `.claude/skills/nous-tools` + `nous-resolve` now intent-lowered AND in the menu AND servable via `weave skill nous-tools` (issue §A3/§E1 closed; #102 folded in).
- **Task C2 — the construct skill: internal-by-declaration** (moved from M2). In
  ariadne's `base.manifest` add `internal skill construct/skill`; move the flat
  `construct/skill/SKILL.md` → `construct/skill/construct/SKILL.md` (uniform
  `<name>/SKILL.md` layout); **decide the name with the operator** (bare `construct`
  via a special-case vs `xx-construct` from the prefix rule); reconcile the tracked
  `.claude/skills/construct` real copy (weave now lowers the construct skill, so the
  self-sync copy is redundant). Verify: ariadne self-walk INCLUDES the construct
  skill (leaf-internal); NO derivative gets it (ancestor-internal — replacing the
  old exclude-by-location). The §B1 capability (M2 Task B) applied to the real skill.
- **Task D — re-weave + verify ALL 10 repos** (ariadne, nous, parley, pair, 42shots, xianxu.dev, you-decide, brain, brain-family, brain-private): clean `git status`, ancestors byte-pristine, `weave golden`/`verify-complete` clean, the brains via the sandbox-safe path (out-of-sandbox `make weave` for brain). Mirror the M5 cutover runbook.
- **Task E — M3 close** + retire #101 + #102 (folded in) + update the `skill-system` target (the invariants are now test-bound — the "DESIGN-ONLY" banner on the algebra's skill slice can be lifted).

---

## Open questions (carry into the milestone that hits them)

- **`construct/skill` layout + the construct skill's name** (M2) — flat→`<name>/SKILL.md`; bare `construct` vs prefixed.
- **Suppression** (B2) — a derivative dropping an inherited export. Out of scope for v2 (additive/override-only); note where the predicate would extend.
- **Menu for the claude target** — claude composes a prose-only AGENTS.md (no `## Skills` menu); the menu is exercised only by codex/agy + `weave skills`. v2 makes the menu CORRECT; it doesn't change which target shows it.
