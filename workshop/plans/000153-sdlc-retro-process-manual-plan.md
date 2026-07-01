# sdlc retro — Process Manual (M1) Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `sdlc retro` that deterministically unrolls every always-on injection source (sdlc-injected prompts, help text, skills, lessons, the AGENTS chain, persisted memories) into one human-readable markdown "process manual" with live links back to each source.

**Architecture:** A new `cmd/sdlc/internal/retro` package with a **pure core** (`InjectionSource` record + `renderManual` pure function) and a **thin IO shell** of per-kind *collectors* that each return `[]InjectionSource`. Judge prompts collect purely (`judge.BuildPrompt` is already pure); file-backed sources (helptext embed.FS, skills, lessons, AGENTS chain, memories) are read through injected `fs.FS`/root-path seams so each collector is testable with a fake FS or temp dir. `cmd/sdlc/retro.go` is cobra glue mirroring existing read-only verbs.

**Tech Stack:** Go, cobra, `embed`/`io/fs`, existing `cmd/sdlc/internal/judge` + `cmd/sdlc/helptext` packages.

**Scope:** This plan covers **M1 (the static catalog)** in full. **M2 (dynamic session reconstruction)** is a coarse forward-reference at the end — it consumes M1's catalog as its baseline and is re-planned at its own `start-plan`.

---

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `InjectionSource` | `cmd/sdlc/internal/retro/source.go` | new |
| `renderManual` | `cmd/sdlc/internal/retro/source.go` | new |
| `claudeProjectSlug` | `cmd/sdlc/internal/retro/memory.go` | new |
| `judgeSources` | `cmd/sdlc/internal/retro/collect.go` | new |

- **`InjectionSource`** — one catalogued point where content enters an agent's context.
  - Fields: `Kind` (grouping: `"sdlc-prompt"`, `"help-text"`, `"skill"`, `"lessons"`, `"agents-chain"`, `"memory"`), `Title`, `When` (prose: the trigger that injects it — e.g. "boundary review at `sdlc close`"), `Link` (repo-root-relative path, optional `#anchor`), `Body` (what to show: full rendered text for code-generated prompts that have no readable standalone file; a short excerpt for file-backed sources the human will click through to).
  - **Relationships:** N:1 into `Kind` groups; `renderManual` consumes a `[]InjectionSource`.
  - **DRY rationale:** One record type for every source, so the renderer, tests, and (later, M2) the "did it fire?" diff all speak one vocabulary instead of per-source shapes.
  - **Future extensions (M2):** add `Fired bool` / `Order int` when the same record is matched against a session transcript.

- **`renderManual(sources []InjectionSource, linkPrefix string) string`** — pure: groups sources by `Kind` into stable-ordered sections and emits, per source, a linked heading + `When` + `Body`. `linkPrefix` is prepended to each `Link` so the doc's links resolve from wherever it is written (`""` = paths from repo root for stdout; `"../"` when written under `workshop/`, etc.). No IO, no clock — deterministic. Colocated test asserts on the produced markdown string.
  - **DRY rationale:** Single rendering path; collectors never format markdown themselves (they only produce records).

- **`claudeProjectSlug(absRepoRoot string) string`** — pure: maps an absolute path to the Claude harness project-dir slug by replacing `/` with `-` (e.g. `/Users/x/workspace/ariadne` → `-Users-x-workspace-ariadne`). Pulled out as a pure function so the memory-locator's one non-trivial transform is unit-testable without touching `$HOME`.

- **`judgeSources() []InjectionSource`** — pure (calls only `judge.BuildPrompt(cat, judge.PromptInput{})`, which is a pure function): one record per catalogued judge category, `Body` = the full rendered prompt with empty dynamic sections. Lives with the collectors but is PURE and tested directly with no fake.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `helptext.Names` / `helptext.FS` | `cmd/sdlc/helptext/embed.go` | modified | embedded `*.md` |
| `helptextSources` | `cmd/sdlc/internal/retro/collect.go` | new | `fs.FS` (embed) |
| `skillSources` | `cmd/sdlc/internal/retro/collect.go` | new | `.claude/skills/*` filesystem |
| `fileSources` (lessons + AGENTS chain) | `cmd/sdlc/internal/retro/collect.go` | new | repo files |
| `memorySources` | `cmd/sdlc/internal/retro/memory.go` | new | `~/.claude/projects/<slug>/memory` |
| `Collect` | `cmd/sdlc/internal/retro/collect.go` | new | aggregates all collectors |
| `RetroCmd` | `cmd/sdlc/retro.go` | new | cobra / stdout / `--out` file |

- **`helptext.Names() []string`** (and `helptext.FS() fs.FS`) — expose the already-embedded `*.md` set so retro can enumerate help text. Minimal export added to the existing package (the `embed.FS` var is currently unexported).
  - **Injected into:** `helptextSources`, which takes an `fs.FS` so it can be tested against a fake FS holding two throwaway `.md` files.

- **`helptextSources(fsys fs.FS) []InjectionSource`** — one record per embedded help-text file; `Link` → `cmd/sdlc/helptext/<name>.md`, `When` → "printed by `sdlc <verb> --help` and on verb error", `Body` → the file's first paragraph (excerpt). Raw embedded text is shown (placeholders like `{{LIFECYCLE}}` left literal) with a one-line note that the live render substitutes them — avoids coupling M1 to `main.go`'s `renderLong`.

- **`skillSources(skillsDir string) []InjectionSource`** — walk `<skillsDir>/*/SKILL.md`, parse frontmatter `name` + `description`; `When` = the description (the trigger), `Body` = excerpt, `Link` → the resolved `SKILL.md` (symlinks resolve into `construct/…`). Injected with the dir path → tested against a temp dir.

- **`fileSources(root string) []InjectionSource`** — the fixed repo files: `workshop/lessons.md` (Kind `lessons`) and the AGENTS chain `AGENTS.md`, `AGENTS.base.md`, `AGENTS.local.md`, `CLAUDE.md`, `GEMINI.md` (Kind `agents-chain`), each with a `When` noting which agent injects it (CLAUDE.md → Claude at session start; GEMINI.md → Gemini; AGENTS.md → agent-neutral; `.base`/`.local` → merge inputs). Missing files are skipped, not errors.

- **`memorySources(homeDir, absRepoRoot string) []InjectionSource`** — best-effort: compute `homeDir/.claude/projects/<claudeProjectSlug(absRepoRoot)>/memory`, and if it exists emit the `MEMORY.md` index + each memory file; else emit one record noting no persisted memories were found. Agent-specific (Claude) + outside the repo — flagged as such in `When`. This is a **documented blind-spot** per the issue, surfaced rather than silently dropped.

- **`Collect(opts CollectOptions) []InjectionSource`** — the one IO aggregator: concatenates `judgeSources()`, `helptextSources(helptext.FS())`, `skillSources`, `fileSources`, `memorySources`. `CollectOptions{RepoRoot, SkillsDir, HomeDir}` are injected so the aggregator itself is exercised with a temp fixture in one integration test.

- **`RetroCmd` / `runRetro(stdout, stderr io.Writer, opts) error`** — cobra glue mirroring `NewEstimateSourceCmd`. Resolves repo root (git top-level) + `$HOME`, calls `Collect`, then `renderManual`, and writes to stdout or `--out <path>` (computing `linkPrefix` from the out-file's dir → repo root). Help text is `helptext/retro.md`, wired by `add(NewRetroCmd(), "retro", …)` in `main.go` (which sets Long via `renderLong` internally).

**Architecture principle citations:**
- **ARCH-PURE** — `InjectionSource`/`renderManual`/`claudeProjectSlug`/`judgeSources` are pure and tested without mocks; every file-touching collector takes an injected `fs.FS`/dir/home seam, so no collector needs a function-call mock to test.
- **ARCH-DRY** — reuse `judge.BuildPrompt` + the existing category constants (never re-transcribe prompt prose); reuse the existing `helptext` `embed.FS` (add a tiny accessor, don't re-list files); one `renderManual` path, collectors never format markdown.
- **ARCH-PURPOSE** — the manual must include **every** injection source, not the easy subset: all **8** judge categories *including `EstimateQuality`* (a real change-code-time injection that `AllCategories()` deliberately omits for standalone-judge validity — so the catalog defines its own complete list), every help-text file, every skill, lessons, the full AGENTS chain, and memories (agent-specific, included-with-caveat). The M1 acceptance runs a **shadow-sweep** confirming each catalogued source appears.

---

## Chunk 1: M1 — static injection catalog

### Task 1: `InjectionSource` type + pure `renderManual`

**Files:**
- Create: `cmd/sdlc/internal/retro/source.go`
- Test: `cmd/sdlc/internal/retro/source_test.go`

- [ ] **Step 1: Write the failing test**

```go
package retro

import (
	"strings"
	"testing"
)

func TestRenderManual_GroupsByKindWithLinks(t *testing.T) {
	sources := []InjectionSource{
		{Kind: KindHelpText, Title: "close", When: "printed by `sdlc close --help`", Link: "cmd/sdlc/helptext/close.md", Body: "Close gate…"},
		{Kind: KindSDLCPrompt, Title: "milestone-review", When: "boundary review at `sdlc close`", Link: "cmd/sdlc/internal/judge/prompts.go", Body: "You are conducting a fresh-context review…"},
	}
	out := renderManual(sources, "")

	// Grouped section headers appear, sdlc-prompt group before help-text (stable order).
	if i, j := strings.Index(out, "## sdlc-injected prompts"), strings.Index(out, "## Help text"); i < 0 || j < 0 || i > j {
		t.Fatalf("sections missing or misordered:\n%s", out)
	}
	// Each source renders a linked title, its When, and its Body.
	for _, want := range []string{
		"[milestone-review](cmd/sdlc/internal/judge/prompts.go)",
		"boundary review at `sdlc close`",
		"You are conducting a fresh-context review",
		"[close](cmd/sdlc/helptext/close.md)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("manual missing %q:\n%s", want, out)
		}
	}
}

func TestRenderManual_LinkPrefixApplied(t *testing.T) {
	out := renderManual([]InjectionSource{{Kind: KindLessons, Title: "lessons.md", Link: "workshop/lessons.md"}}, "../")
	if !strings.Contains(out, "(../workshop/lessons.md)") {
		t.Errorf("linkPrefix not applied:\n%s", out)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd/sdlc/internal/retro/ -run TestRenderManual -v`
Expected: FAIL (package/type/func not defined).

- [ ] **Step 3: Write minimal implementation**

```go
// Package retro assembles the "process manual" — every always-on injection
// source unrolled into one linked markdown document (issue #153).
package retro

import (
	"fmt"
	"sort"
	"strings"
)

type Kind string

const (
	KindSDLCPrompt  Kind = "sdlc-prompt"
	KindHelpText    Kind = "help-text"
	KindSkill       Kind = "skill"
	KindLessons     Kind = "lessons"
	KindAgentsChain Kind = "agents-chain"
	KindMemory      Kind = "memory"
)

// InjectionSource is one catalogued point where content enters an agent's context.
type InjectionSource struct {
	Kind  Kind
	Title string
	When  string // prose trigger: what injects this content
	Link  string // repo-root-relative path (+ optional #anchor)
	Body  string // full text for code-generated prompts; excerpt for file-backed sources
}

// section groups Kinds into stable display order with human headings.
var sections = []struct {
	kind    Kind
	heading string
}{
	{KindSDLCPrompt, "## sdlc-injected prompts"},
	{KindHelpText, "## Help text"},
	{KindSkill, "## Skills"},
	{KindLessons, "## Lessons"},
	{KindAgentsChain, "## AGENTS chain"},
	{KindMemory, "## Persisted memories"},
}

// renderManual is pure: sources → linked markdown, grouped by Kind in `sections`
// order. linkPrefix is prepended to every Link so the doc resolves from wherever
// it is written ("" = from repo root).
func renderManual(sources []InjectionSource, linkPrefix string) string {
	byKind := map[Kind][]InjectionSource{}
	for _, s := range sources {
		byKind[s.Kind] = append(byKind[s.Kind], s)
	}
	var b strings.Builder
	b.WriteString("# Process manual\n\n")
	b.WriteString("_Generated by `sdlc retro`. Every always-on injection source, linked to its origin. Do not hand-edit — regenerate._\n\n")
	for _, sec := range sections {
		items := byKind[sec.kind]
		if len(items) == 0 {
			continue
		}
		sort.SliceStable(items, func(i, j int) bool { return items[i].Title < items[j].Title })
		b.WriteString(sec.heading + "\n\n")
		for _, s := range items {
			fmt.Fprintf(&b, "### [%s](%s%s)\n\n", s.Title, linkPrefix, s.Link)
			if s.When != "" {
				fmt.Fprintf(&b, "**When:** %s\n\n", s.When)
			}
			if s.Body != "" {
				b.WriteString(s.Body)
				if !strings.HasSuffix(s.Body, "\n") {
					b.WriteString("\n")
				}
				b.WriteString("\n")
			}
		}
	}
	return b.String()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd/sdlc/internal/retro/ -run TestRenderManual -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/sdlc/internal/retro/source.go cmd/sdlc/internal/retro/source_test.go
git commit -m "#153 M1: retro InjectionSource + pure renderManual"
```

---

### Task 2: `judgeSources` collector (pure)

**Files:**
- Create: `cmd/sdlc/internal/retro/collect.go`
- Test: `cmd/sdlc/internal/retro/collect_test.go`

- [ ] **Step 1: Write the failing test** — asserts every catalogued category is present, *including* `EstimateQuality` (the ARCH-PURPOSE completeness case), and that bodies are the real rendered prompts.

```go
package retro

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

func TestJudgeSources_CoversEveryCategoryIncludingEstimate(t *testing.T) {
	got := judgeSources()
	titles := map[string]InjectionSource{}
	for _, s := range got {
		if s.Kind != KindSDLCPrompt {
			t.Errorf("judgeSources produced non-prompt kind %q", s.Kind)
		}
		titles[s.Title] = s
	}
	// All 8 categories — AllCategories() omits estimate-quality, the catalog must not.
	want := []judge.Category{
		judge.DRY, judge.PURE, judge.Plan, judge.PlanQuality,
		judge.EstimateQuality, judge.Specs, judge.Lessons, judge.MilestoneReview,
	}
	for _, c := range want {
		s, ok := titles[string(c)]
		if !ok {
			t.Fatalf("judgeSources missing category %q", c)
		}
		if strings.TrimSpace(s.Body) == "" {
			t.Errorf("category %q has empty rendered body", c)
		}
		if !strings.Contains(s.Link, "prompts.go") {
			t.Errorf("category %q link should point at the builder, got %q", c, s.Link)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails** — `go test ./cmd/sdlc/internal/retro/ -run TestJudgeSources -v` → FAIL (undefined).

- [ ] **Step 3: Write minimal implementation**

```go
import (
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// catalogCategories is the COMPLETE injected set — judge.AllCategories() omits
// EstimateQuality (scoped to standalone `sdlc judge` validity), but that prompt
// IS injected at change-code, so the manual must list it (ARCH-PURPOSE).
func catalogCategories() []judge.Category {
	return append(append([]judge.Category{}, judge.AllCategories()...), judge.EstimateQuality)
}

// categoryBody is what the manual shows for a judge category. BuildPrompt is a
// pure function, so an empty PromptInput yields the deterministic prompt
// skeleton — EXCEPT `lessons`, which BuildPrompt renders as "" because it is a
// reminder ping, not a full agent prompt. For lessons the injected content IS
// judge.LessonsReminder. (This judge `lessons` category is a DIFFERENT injection
// point from the workshop/lessons.md file, which fileSources emits under Kind
// `lessons`.)
func categoryBody(c judge.Category) string {
	if body := judge.BuildPrompt(c, judge.PromptInput{}); strings.TrimSpace(body) != "" {
		return body
	}
	if c == judge.Lessons {
		return judge.LessonsReminder
	}
	return ""
}

// judgeSources is pure (judge.BuildPrompt + a constant do no IO).
func judgeSources() []InjectionSource {
	var out []InjectionSource
	for _, c := range catalogCategories() {
		out = append(out, InjectionSource{
			Kind:  KindSDLCPrompt,
			Title: string(c),
			When:  whenForCategory(c),
			Link:  "cmd/sdlc/internal/judge/prompts.go",
			Body:  categoryBody(c),
		})
	}
	return out
}

// whenForCategory maps each category to its injection-trigger prose (pure).
func whenForCategory(c judge.Category) string {
	switch c {
	case judge.PlanQuality, judge.EstimateQuality:
		return "plan-quality gate at `sdlc change-code`"
	case judge.MilestoneReview:
		return "boundary review at `sdlc close` / `sdlc milestone-close`"
	case judge.Lessons:
		return "lessons reminder emitted at the review boundary"
	default:
		return "`sdlc judge " + string(c) + "`"
	}
}
```

- [ ] **Step 4: Run test to verify it passes** — expected PASS.

- [ ] **Step 5: Commit** — `git commit -m "#153 M1: judgeSources — all 8 categories incl estimate-quality (ARCH-PURPOSE)"`

---

### Task 3: `helptext.FS()` accessor + `helptextSources`

**Files:**
- Modify: `cmd/sdlc/helptext/embed.go` (add `FS()` only — one accessor, no dead `Names()`)
- Modify: `cmd/sdlc/internal/retro/collect.go`
- Test: `cmd/sdlc/helptext/embed_test.go` (accessor), `cmd/sdlc/internal/retro/collect_test.go` (collector against a fake FS)

**ARCH-DRY note (from plan review):** keep a single enumeration path. `helptextSources(fsys fs.FS)` owns the `.md` walk; `helptext` exposes only `FS()` (the injectable seam), no separate `Names()` — a public accessor no production caller uses is dead code.

- [ ] **Step 1: Write failing tests** — (a) `helptext.FS()` opens `root.md` (a known embedded file); (b) `helptextSources(fstest.MapFS{...})` on a fake FS with two `.md` files yields two `KindHelpText` records whose `Link` is `cmd/sdlc/helptext/<name>.md`, and the `root` entry's `When` names bare `sdlc --help` (not `sdlc root --help`).

```go
// collect_test.go
func TestHelptextSources_FromFakeFS(t *testing.T) {
	fsys := fstest.MapFS{
		"close.md": {Data: []byte("Close gate.\n\nSecond para.")},
		"root.md":  {Data: []byte("The workflow contract.")},
	}
	got := helptextSources(fsys)
	if len(got) != 2 {
		t.Fatalf("want 2 help-text sources, got %d", len(got))
	}
	byTitle := map[string]InjectionSource{}
	for _, s := range got {
		if s.Kind != KindHelpText || !strings.HasPrefix(s.Link, "cmd/sdlc/helptext/") {
			t.Errorf("bad help-text source: %+v", s)
		}
		byTitle[s.Title] = s
	}
	if w := byTitle["root"].When; !strings.Contains(w, "sdlc --help") || strings.Contains(w, "sdlc root") {
		t.Errorf("root When should name bare `sdlc --help`, got %q", w)
	}
}
```

- [ ] **Step 2: Run to verify fail.**
- [ ] **Step 3: Implement** — add ONE accessor to `helptext/embed.go`. The package-level embed var is `var fs embed.FS`; do **not** `import "io/fs"` here (it collides with that name). Return the concrete `embed.FS` (it satisfies `io/fs.FS` at the call site); no new imports needed:

```go
// FS exposes the embedded help-text filesystem for enumeration (issue #153).
// Returns the concrete embed.FS, which satisfies io/fs.FS where callers need it.
func FS() embed.FS { return fs }
```

and `helptextSources(fsys fs.FS)` in `collect.go` (the retro package **does** import `io/fs`, no collision there): `fs.ReadDir(fsys, ".")`, keep `.md` files, `fs.ReadFile(fsys, name)` for the first-paragraph excerpt (`Body`), `Link` → `cmd/sdlc/helptext/<name>.md`. `When` via a small pure helper `helptextWhen(name)`: `root` → "printed by bare `sdlc --help` (the workflow contract)"; everything else → "embedded help; printed by the matching `sdlc … --help` / on verb error" (kept generic-but-truthful, since some names like `set-status`/`fetch` are sub-verbs, not top-level `sdlc <name>`). Called in production as `helptextSources(helptext.FS())`.

- [ ] **Step 4: Run to verify pass.**
- [ ] **Step 5: Commit** — `git commit -m "#153 M1: expose helptext.FS + helptextSources (truthful When)"`

---

### Task 4: `skillSources` collector

**Files:** Modify `cmd/sdlc/internal/retro/collect.go`; Test `collect_test.go`.

- [ ] **Step 1: Write failing test** — build a temp dir `skills/xx-demo/SKILL.md` with frontmatter (`name: xx-demo`, `description: Use when demoing.`); assert `skillSources(tmp)` returns one `KindSkill` record with `When` containing "Use when demoing" and `Link` ending `SKILL.md`.
- [ ] **Step 2: Run to verify fail.**
- [ ] **Step 3: Implement** — glob `<skillsDir>/*/SKILL.md`, parse frontmatter with the existing `pkg/frontmatter` package (ARCH-DRY): `frontmatter.Split(content) (fm, body, err)` splits the `---` block from the body, and `frontmatter.Description(content)` returns the `description:` value. `When` = `frontmatter.Description(content)` (the trigger); `Title` = the skill dir name (or the `name:` field); `Body` = first non-frontmatter paragraph of `body`. Note skill entries under `.claude/skills/` are symlinks into `construct/…`; resolve the link for `Link` so it points at the real `SKILL.md`.
- [ ] **Step 4: Run to verify pass.**
- [ ] **Step 5: Commit** — `git commit -m "#153 M1: skillSources from .claude/skills/*/SKILL.md"`

---

### Task 5: `fileSources` (lessons + AGENTS chain) + `memorySources`

**Files:** Modify `collect.go`; Create `cmd/sdlc/internal/retro/memory.go` + `memory_test.go`; Test `collect_test.go`.

- [ ] **Step 1: Write failing tests**
  - `claudeProjectSlug("/Users/x/workspace/ariadne")` == `"-Users-x-workspace-ariadne"` (pure).
  - `fileSources(tmpRoot)` where `tmpRoot/workshop/lessons.md` and `tmpRoot/AGENTS.md` exist → a `KindLessons` record + a `KindAgentsChain` record; a missing `GEMINI.md` produces no error and no record.
  - `memorySources(tmpHome, "/w/ariadne")` with `tmpHome/.claude/projects/-w-ariadne/memory/MEMORY.md` present → a `KindMemory` record for the index; with the dir absent → exactly one `KindMemory` record whose `Body` notes none were found.
- [ ] **Step 2: Run to verify fail.**
- [ ] **Step 3: Implement** — `claudeProjectSlug` = `strings.ReplaceAll(abs, "/", "-")`; `fileSources` iterates a fixed list `{path, Kind, When}` skipping absent files; `memorySources` uses `claudeProjectSlug` + `homeDir` to locate the dir (best-effort, agent-specific caveat in `When`).
- [ ] **Step 4: Run to verify pass.**
- [ ] **Step 5: Commit** — `git commit -m "#153 M1: fileSources + memorySources (best-effort, agent-specific)"`

---

### Task 6: `Collect` aggregator + `runRetro` + `NewRetroCmd` + wiring

**Files:**
- Modify: `cmd/sdlc/internal/retro/collect.go` (add `Collect`)
- Create: `cmd/sdlc/retro.go`, `cmd/sdlc/helptext/retro.md`
- Modify: `cmd/sdlc/main.go` (register), `cmd/sdlc/retro_test.go`
- Test: `collect_test.go` (aggregator over a fixture), `retro_test.go` (cobra tree)

- [ ] **Step 1: Write failing tests**
  - `Collect(CollectOptions{RepoRoot, SkillsDir, HomeDir})` over a temp fixture returns records spanning **every** `Kind` (the shadow-sweep in a test: judge, help-text, skill, lessons, agents-chain, memory all present).
  - Cobra: `buildRoot()` with args `["retro"]` (run in the real repo working dir) executes without error and stdout contains `"# Process manual"`, `"milestone-review"`, and `"## Skills"`.

```go
// retro_test.go
func TestRetroCmd_WiredInRoot(t *testing.T) {
	root := buildRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"retro"})
	if err := root.Execute(); err != nil {
		t.Fatalf("`sdlc retro` failed: %v", err)
	}
	for _, want := range []string{"# Process manual", "milestone-review", "## Skills"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("`sdlc retro` output missing %q", want)
		}
	}
}
```

- [ ] **Step 2: Run to verify fail.**
- [ ] **Step 3: Implement**
  - `Collect` concatenates the collectors (judge = pure; the rest from `opts`).
  - `retro.go`: `NewRetroCmd()` mirrors `NewEstimateSourceCmd` — `--out <path>` (default stdout). `runRetro(stdout, stderr, opts)` resolves repo root via `gitx.RepoTopLevel()` (ARCH-DRY — the helper `branchcreate.go`/`state.go` already use), resolves `$HOME` via `os.UserHomeDir()`, `Collect`, `renderManual` (linkPrefix `""` for stdout; for `--out` compute `filepath.Rel(outDir, repoRoot)+"/"` — **assumes `--out` is within the repo tree**, adequate for M1), write. On `--out`, also print the written path to stderr.
  - `helptext/retro.md`: verb contract (what it unrolls, that it's a regeneration not hand-edited, the memory/agent caveat). No `{{…}}` placeholders needed.
  - `main.go`: `add(NewRetroCmd(), "retro", "Unroll every injection source into a linked process manual")`. (`add(...)` already calls `renderLong("retro")` internally to set the Long help from `helptext/retro.md` — no separate call.)
- [ ] **Step 4: Run to verify pass** — `go test ./cmd/sdlc/...` all green; then a real smoke run: `go run ./cmd/sdlc retro | head -40`.
- [ ] **Step 5: Commit** — `git commit -m "#153 M1: Collect aggregator + sdlc retro verb wired in root"`

---

### Task 7: Atlas + acceptance shadow-sweep

**Files:** Modify `atlas/workflow/sdlc-binary.md` (or add `atlas/workflow/retro.md` + link in `atlas/workflow/index.md` and `atlas/index.md`).

- [ ] **Step 1** — Document `sdlc retro` in the atlas: what it is (the process-manual regenerator), the six source kinds, and the stated blind spots (memories are Claude-specific + outside the repo; Task-tool subagent forks are a v2 gap).
- [ ] **Step 2 — Shadow-sweep (ARCH-PURPOSE acceptance):** run `go run ./cmd/sdlc retro` and confirm the output contains, at minimum: all 8 judge categories, ≥1 entry per help-text file (`sdlc retro | grep -c '## Help text'` and spot-check `close`), every `.claude/skills/*` skill, `lessons.md`, all present AGENTS-chain files, and the memory section. Record the sweep result in the issue `## Log`.
- [ ] **Step 3: Commit** — `git commit -m "#153 M1: atlas entry for sdlc retro + shadow-sweep evidence"`

---

## Estimate

The canonical, reconciled derivation is the `## Estimate` block in the issue
(`workshop/issues/000153-sdlc-retro-process-manual.md`) — the estimate-quality gate
reads that one, so this plan does not carry a second (divergent) derivation. In
short: M1 = one greenfield-go-module (`internal/retro`: pure core + 6 collectors) +
a smaller-go-module (`retro.go` glue + `helptext.FS` accessor) + atlas-docs + the
M1 milestone-review cost, recalibrated to the v3.1 corpus ≈ **3.07h**. M2 is
estimated separately at its own `start-plan`.

---

## Milestone 2 (coarse — re-planned at its own `start-plan`)

Not detailed here (would be speculative before the catalog exists). Shape only:

- **Per-agent JSONL parser seam** (Claude `~/.claude/projects/<slug>/*.jsonl` first) that emits an ordered list of injection *events* (Bash `sdlc …` calls, Skill loads, file reads), reusing `introspect`'s normalize/segment plumbing.
- **Sidecar join** — for `sdlc close`/`milestone-close` review forks, read the correlated `workshop/history/…-review.md` sidecar rather than the orphan reviewer JSONL.
- **Match events against the M1 catalog** (the baseline) → an **anomaly-first** report: injected-but-never-referenced instructions, deviations from the modal sequence, with `🤖[]` annotation slots.
- Documented v2 blind spots carried forward: AGENTS.md/memory "step-0" by convention; generic Task-tool subagent forks need cwd+timestamp correlation.

---

## Revisions

- **2026-07-01 — verb + package renamed `retro` → `process-manual`.** User asked to
  make the verb explicit. All `retro` names in this plan (`cmd/sdlc/internal/retro/*`,
  `cmd/sdlc/retro.go`, `RetroCmd`/`runRetro`, `package retro`, `sdlc retro`) map to the
  new names: package `cmd/sdlc/internal/processmanual`, command `cmd/sdlc/processmanual.go`,
  `NewProcessManualCmd`/`runProcessManual`, `package processmanual`, `sdlc process-manual`.
  Also folded in during M1 (delta from the tasks above): dropped the dead
  `helptext.Names()` (kept only `FS()`); `categoryBody` shows a first-paragraph **gist**
  of each judge prompt (not the full body — full inline ran ~1100 lines with the ARCH
  registry ×4); `renderManual` fences heading-bearing bodies and leaves absolute /
  empty links unprefixed. Behaviour is otherwise as planned.
