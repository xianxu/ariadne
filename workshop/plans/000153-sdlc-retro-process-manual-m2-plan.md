# sdlc process-manual — M2: judge prompts → embedded markdown

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) for the execution approach (superpowers-subagent-driven-development or superpowers-executing-plans). Steps use checkbox (`- [ ]`) syntax.

**Goal:** Move the 8 per-category judge prompts out of `prompts.go`'s `fmt.Sprintf` string literals into embedded, readable `judge/prompts/*.md` files (placeholder substitution), with **byte-identical** output; then relink `process-manual`'s `judgeSources` to the readable `.md`.

**Architecture:** Finish the pattern the package already uses — `architecture.md` and `code-review.md` are `//go:embed`'d markdown loaded by `ArchitectureBlock`/`CodeReviewBody`. Each category's static prose moves to `judge/prompts/<category>.md` with `{{TOKEN}}` placeholders; `BuildPrompt` loads the template and substitutes via one `strings.Replacer` (the `renderLong` pattern). No behavior change — a golden test locks byte-fidelity.

**Tech Stack:** Go, `embed`/`io/fs`, `strings.Replacer`, existing `cmd/sdlc/internal/judge` package.

**Non-negotiable constraint:** These prompts drive the fresh-context reviews and are documented "byte-faithful." The refactor must not change a single byte of rendered output. Golden tests captured from the *current* `BuildPrompt` are the acceptance gate.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `promptTemplate(Category) string` | `cmd/sdlc/internal/judge/prompts.go` | new |
| `promptSubstitutions(Category, PromptInput) *strings.Replacer` | `cmd/sdlc/internal/judge/prompts.go` | new |
| `BuildPrompt` | `cmd/sdlc/internal/judge/prompts.go` | modified |
| `judge/prompts/*.md` (7 templates) | `cmd/sdlc/internal/judge/prompts/` | new |

- **`promptTemplate(c)`** — reads the embedded `prompts/<c>.md` (via a package `//go:embed prompts/*.md`). Pure over the embed FS. Panics on a missing template (build-time bug, like `helptext.MustGet`).
  - **DRY rationale:** one loader for all category templates; mirrors `helptext.Get`.

- **`promptSubstitutions(c, in)`** — builds the single `strings.Replacer` mapping every `{{TOKEN}}` to its value. `{{ARCH_BLOCK}}` resolves to `ArchitectureBlock("at-plan")` for `plan-quality`, `"at-review"` otherwise; `{{REF}}`/`{{PLAN_CONTENT}}` carry the same empty-fallbacks the current code has (`<unknown>`, `(no separate plan file)`). Pure (its inputs — `ArchitectureBlock`, `ContractPreamble`, `CodeReviewBody(in)`, `estimate.CurrentModel()`, `in.*` — are all pure/string).
  - **DRY rationale:** one token table instead of per-category `fmt.Sprintf` arg lists. A `.md` uses only the tokens it needs; unused tokens are simply absent from that file.
  - **Note:** `strings.Replacer` is single-pass (it never re-scans inserted text), so a value that itself contains `{{…}}` (e.g. a diff) is inserted literally — no accidental double-substitution.

- **`BuildPrompt(category, in)`** (modified) — becomes: `if category == Lessons { return "" }; return promptSubstitutions(category, in).Replace(promptTemplate(category))`. The big `switch` of `fmt.Sprintf` literals is deleted; the prose now lives in the `.md` files.

- **`judge/prompts/*.md`** — one per agent-emitting category: `dry.md`, `pure.md`, `plan.md`, `plan-quality.md`, `estimate-quality.md`, `specs.md`, `milestone-review.md`. `lessons` has no prompt (BuildPrompt returns `""`), so **no** `lessons.md`. Each file is the exact current prose with `%s` positions replaced by named `{{TOKEN}}`s.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `//go:embed prompts/*.md` | `cmd/sdlc/internal/judge/prompts.go` | new | embedded FS |
| `judgeSources` | `cmd/sdlc/internal/processmanual/collect.go` | modified | judge prompt links |

- **judge prompt embed** — `//go:embed prompts/*.md var promptFS embed.FS`, mirroring `architecture.go`/`review.go`.
- **`judgeSources` relink** — `Link` changes from the bare `cmd/sdlc/internal/judge/prompts.go` to the per-category readable file `cmd/sdlc/internal/judge/prompts/<c>.md` (for categories that have one; `lessons` keeps pointing at `prompts.go`, where `LessonsReminder` lives). Body stays the rendered `BuildPrompt` gist/`--full` (the reader sees the *rendered* prompt inline; the link now lands on the readable source, not Go code). This is the payoff: judge prompts join help text/skills as file-backed sources.

**Architecture principle citations:**
- **ARCH-DRY** — extends the existing `//go:embed` prose pattern (`architecture.md`, `code-review.md`) instead of a parallel string-literal mechanism; one `strings.Replacer` token table replaces 7 bespoke `fmt.Sprintf` arg lists.
- **ARCH-PURE** — `promptTemplate`/`promptSubstitutions`/`BuildPrompt` stay pure (embed reads + string work, no IO/clock); golden-tested directly.
- **ARCH-PURPOSE** — the purpose is *legible prompts*: every agent-emitting category gets a real `.md`, and `process-manual` links to it. Not a subset — the shadow-check is that all 7 `.md` exist and `BuildPrompt` renders each byte-identically.

---

## Chunk 1: M2 — extract prompts to markdown

### Task 1: Golden capture of current `BuildPrompt` (the safety net, FIRST)

**Files:**
- Create: `cmd/sdlc/internal/judge/golden_test.go`
- Create (generated): `cmd/sdlc/internal/judge/testdata/golden/<category>.prompt` (7 files)

- [ ] **Step 1: Write the golden test with an `-update` capture mode**

```go
package judge

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var updateGolden = flag.Bool("update-golden", false, "rewrite testdata/golden/*.prompt from current BuildPrompt")

// goldenInput exercises EVERY placeholder (non-empty diff, issue, plan, refs,
// changed issues, and the milestone-review repo-orientation fields).
var goldenInput = PromptInput{
	Diff:          "DIFF-BODY-LINE-1\nDIFF-BODY-LINE-2",
	ChangedIssues: []string{"workshop/issues/000001-a.md", "workshop/issues/000002-b.md"},
	Base:          "BASE_SHA", Head: "HEAD",
	IssueRef:      "pair#31 M2",
	IssueContent:  "ISSUE-CONTENT-BODY",
	PlanContent:   "PLAN-CONTENT-BODY",
	Repo:          "pair", RepoRoot: "/abs/pair",
	IssueFile:     "workshop/issues/000031-x.md",
	Boundary:      "milestone M2 close",
	RepoNote:      "REPO-ORIENTATION-NOTE",
}

func goldenCategories() []Category {
	return append(append([]Category{}, AllCategories()...), EstimateQuality)
}

func TestBuildPrompt_Golden(t *testing.T) {
	for _, c := range goldenCategories() {
		if c == Lessons {
			continue // no prompt (BuildPrompt returns "")
		}
		got := BuildPrompt(c, goldenInput)
		path := filepath.Join("testdata", "golden", string(c)+".prompt")
		if *updateGolden {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("golden missing for %s (run: go test ./cmd/sdlc/internal/judge -update-golden): %v", c, err)
		}
		if got != string(want) {
			t.Errorf("BYTE DRIFT for %s: BuildPrompt output changed vs golden.\n--- got ---\n%s", c, got)
		}
	}
}
```

- [ ] **Step 2: Capture goldens from the CURRENT (pre-refactor) BuildPrompt**

Run: `go test ./cmd/sdlc/internal/judge -run TestBuildPrompt_Golden -update-golden`
Then: `go test ./cmd/sdlc/internal/judge -run TestBuildPrompt_Golden` → PASS (goldens match current code).
Expected: 7 files under `testdata/golden/`. **These are the frozen truth for the whole refactor.**

> ⛔ **CRITICAL — never re-run `-update-golden` after this step.** The entire point of
> M2 is zero-byte drift in the review prompts. From Task 2 on, a golden **failure means
> a `.md` drifted — fix the `.md`, NOT the golden.** Re-capturing (`-update-golden`) to
> "make the test pass" silently bakes the drift into the frozen truth and corrupts the
> fresh-context review prompts with no test left to catch it. If a subagent executes
> this plan, this instruction is load-bearing.

- [ ] **Step 2b: Cover the empty-`IssueRef` fallback.** `goldenInput` sets a non-empty `IssueRef`, so the `ref == "" → "<unknown>"` branch (used by `plan-quality` + `estimate-quality`) is exercised by no test. Add a tiny unit assertion (not a golden) alongside the golden test:

```go
func TestBuildPrompt_EmptyIssueRefFallback(t *testing.T) {
	in := goldenInput
	in.IssueRef = ""
	for _, c := range []Category{PlanQuality, EstimateQuality} {
		if !strings.Contains(BuildPrompt(c, in), "<unknown>") {
			t.Errorf("%s with empty IssueRef should render <unknown>", c)
		}
	}
}
```

(The `PlanContent == "" → "(no separate plan file)"` fallback is already covered by `TestBuildPrompt_PlanQuality_NoSeparatePlan`.)

- [ ] **Step 3: Commit the safety net**

```bash
git add cmd/sdlc/internal/judge/golden_test.go cmd/sdlc/internal/judge/testdata/golden/
git commit -m "#153 M2: golden-capture current BuildPrompt output (byte-fidelity net)"
```

---

### Task 2: Extract the two simplest prompts (`dry`, `pure`) → `.md`

**Files:**
- Create: `cmd/sdlc/internal/judge/prompts/dry.md`, `cmd/sdlc/internal/judge/prompts/pure.md`
- Modify: `cmd/sdlc/internal/judge/prompts.go` (embed + loader + substitutions + the two cases)

- [ ] **Step 1: Add the embed + loader + substitutions helper.** Create `dry.md`/`pure.md` (Steps 2–3) *before* the first `go build` — `//go:embed prompts/*.md` errors if the glob matches no files.

```go
import "embed"

//go:embed prompts/*.md
var promptFS embed.FS

func promptTemplate(c Category) string {
	b, err := promptFS.ReadFile("prompts/" + string(c) + ".md")
	if err != nil {
		panic("judge: prompt template missing: " + string(c) + ".md")
	}
	return string(b)
}

func promptSubstitutions(c Category, in PromptInput) *strings.Replacer {
	archLens := "at-review"
	if c == PlanQuality {
		archLens = "at-plan"
	}
	return strings.NewReplacer(
		"{{ARCH_BLOCK}}", ArchitectureBlock(archLens),
		"{{CONTRACT}}", ContractPreamble,
		"{{BOUNDARY_CONTRACT}}", BoundaryReviewContract,
		"{{CODE_REVIEW_BODY}}", CodeReviewBody(in),
		"{{DIFF}}", in.Diff,
		"{{CHANGED_ISSUES}}", strings.Join(in.ChangedIssues, "\n"),
		"{{ISSUE_CONTENT}}", in.IssueContent,
		"{{PLAN_CONTENT}}", orDefault(in.PlanContent, "(no separate plan file)"),
		"{{REF}}", orDefault(in.IssueRef, "<unknown>"),
		"{{MODEL}}", estimate.CurrentModel(),
	)
}
```

> **ARCH-DRY:** reuse the existing `orDefault(s, def)` helper (`review.go:50`, already
> used by `CodeReviewBody`) for the empty-`IssueRef`/`PlanContent` fallbacks — don't
> re-inline the `if s == "" {…}` logic.

- [ ] **Step 2: Create `dry.md`** — the exact current prose, `%s` → tokens. The current `dry` template (`prompts.go:137-153`) maps: first `%s` = `{{ARCH_BLOCK}}`, second `%s` = `{{CONTRACT}}`, third `%s` = `{{DIFF}}`:

```
You are a code reviewer checking the diff for ARCH-DRY violations.
The principle is authored once in the registry below (#75):

{{ARCH_BLOCK}}

Apply ARCH-DRY's at-review lens to the diff: duplicated logic, copy-pasted blocks,
near-identical functions that should be one shared helper. Report file:line + the
consolidation. Do NOT modify any files; only report.

{{CONTRACT}}

Tokens for this check:
  CLEAN   = no ARCH-DRY violations.
  FAILURE = duplicated logic that should be consolidated.

Diff:
{{DIFF}}
```

> **Fidelity note for the implementer:** copy the prose verbatim from the current literal (indentation, blank lines, trailing newline). The original `fmt.Sprintf` string ends with a newline after `%s`(Diff); the `.md` file must end with exactly one trailing newline. The golden test is the arbiter — if it fails, diff `got` against the golden to find the drifted byte.

- [ ] **Step 3: Create `pure.md`** — same shape (`prompts.go:157-173`).

- [ ] **Step 4: Rewire the `dry`/`pure` cases** to `return promptSubstitutions(category, in).Replace(promptTemplate(category))` (or fall through to a shared tail — see Task 6).

- [ ] **Step 5: Run the golden test** — `go test ./cmd/sdlc/internal/judge -run TestBuildPrompt_Golden` → PASS (dry + pure byte-identical). Expected FAIL first if a byte drifted; fix the `.md` until green.

- [ ] **Step 6: Commit** — `git commit -m "#153 M2: extract dry/pure prompts to judge/prompts/*.md"`

---

### Task 3: Extract `specs` + `plan` → `.md`

**Files:** Create `prompts/specs.md`, `prompts/plan.md`; Modify `prompts.go`.

- [ ] **Step 1:** `specs.md` from `prompts.go:317-337` (`%s`=`{{CONTRACT}}`, `%s`=`{{DIFF}}`).
- [ ] **Step 2:** `plan.md` from `prompts.go:178-205` (`%s`=`{{CONTRACT}}`, `%s`=`{{CHANGED_ISSUES}}`, `%s`=`{{DIFF}}`). Note: the current code joins `ChangedIssues` with `\n` into `{{CHANGED_ISSUES}}`.
- [ ] **Step 3:** Rewire both cases.
- [ ] **Step 4:** Golden test → PASS.
- [ ] **Step 5:** Commit — `git commit -m "#153 M2: extract specs/plan prompts to markdown"`

---

### Task 4: Extract `plan-quality` + `estimate-quality` → `.md`

**Files:** Create `prompts/plan-quality.md`, `prompts/estimate-quality.md`; Modify `prompts.go`.

- [ ] **Step 1:** `plan-quality.md` from `prompts.go:216-263` (`%s`=`{{REF}}`, then `{{ARCH_BLOCK}}` (at-plan — supplied by `promptSubstitutions`), `{{CONTRACT}}`, `{{ISSUE_CONTENT}}`, `{{PLAN_CONTENT}}`).
- [ ] **Step 2:** `estimate-quality.md` from `prompts.go:271-313` (`%s`=`{{REF}}`, `{{MODEL}}` in the "current model (%s)" line, `{{CONTRACT}}`, `{{ISSUE_CONTENT}}`).
- [ ] **Step 3:** Rewire both cases.
- [ ] **Step 4:** Golden test → PASS (verifies the at-plan-vs-at-review lens routing and the `{{MODEL}}` interpolation).
- [ ] **Step 5:** Commit — `git commit -m "#153 M2: extract plan-quality/estimate-quality prompts to markdown"`

---

### Task 5: Extract `milestone-review` → `.md`

**Files:** Create `prompts/milestone-review.md`; Modify `prompts.go`.

- [ ] **Step 1:** `milestone-review.md` from `prompts.go:352-360`. This one is mostly tokens: `{{CODE_REVIEW_BODY}}`, `{{ARCH_BLOCK}}` (at-review), `{{BOUNDARY_CONTRACT}}`, `Diff:\n{{DIFF}}`. Preserve the exact blank-line spacing between the four blocks.
- [ ] **Step 2:** Rewire the case.
- [ ] **Step 3:** Golden test → PASS.
- [ ] **Step 4:** Commit — `git commit -m "#153 M2: extract milestone-review prompt to markdown"`

---

### Task 6: Collapse `BuildPrompt` to the template path + update package docs

**Files:** Modify `cmd/sdlc/internal/judge/prompts.go`.

- [ ] **Step 1:** Replace the whole category `switch` body with the uniform path:

```go
func BuildPrompt(category Category, in PromptInput) string {
	if category == Lessons {
		return "" // reminder ping, not an agent prompt (LessonsReminder below)
	}
	return promptSubstitutions(category, in).Replace(promptTemplate(category))
}
```

- [ ] **Step 1b: Fix the import block.** Collapsing the switch deletes the last `fmt.Sprintf`, so `"fmt"` is now unused (the package won't compile). Net import delta for `prompts.go`: **add** `"embed"`, **drop** `"fmt"`, **keep** `"strings"` + `estimate`. (`go build` at Step 3 catches it, but do it here.)
- [ ] **Step 2:** Update the `BuildPrompt` + package doc comments: prose now lives in `prompts/*.md`; "byte-faithful from the shell heredocs" → "byte-faithful, now sourced from embedded `prompts/*.md` (golden-tested)".
- [ ] **Step 3:** Run the FULL judge suite — `go test ./cmd/sdlc/internal/judge` (golden + all existing `TestBuildPrompt_*`, `TestArchitectureRegistry_EmbeddedInPrompts`, `TestAgentPromptsEmbedContract` must still pass — they assert content that the `.md` now provides).
- [ ] **Step 4:** Commit — `git commit -m "#153 M2: collapse BuildPrompt to embedded-template path"`

---

### Task 7: Relink `process-manual` `judgeSources` to the `.md` files

**Files:** Modify `cmd/sdlc/internal/processmanual/collect.go`, `collect_test.go`.

- [ ] **Step 1: Write the failing test** — `judgeSources(false)` for a category with a template (e.g. `dry`) has `Link == "cmd/sdlc/internal/judge/prompts/dry.md"`; `lessons` still links to `prompts.go` (no template).

```go
func TestJudgeSources_LinkToMarkdown(t *testing.T) {
	byTitle := map[string]InjectionSource{}
	for _, s := range judgeSources(false) {
		byTitle[s.Title] = s
	}
	if l := byTitle["dry"].Link; l != "cmd/sdlc/internal/judge/prompts/dry.md" {
		t.Errorf("dry should link to its .md, got %q", l)
	}
	if l := byTitle["lessons"].Link; !strings.HasSuffix(l, "prompts.go") {
		t.Errorf("lessons has no template → link prompts.go, got %q", l)
	}
}
```

- [ ] **Step 2:** Run → FAIL (still links prompts.go).
- [ ] **Step 3: Implement** — in `judgeSources`, compute `Link` per category: `lessons` → `cmd/sdlc/internal/judge/prompts.go`; every other category → `cmd/sdlc/internal/judge/prompts/<c>.md`. Body stays `categoryBody(c, full)` (rendered prompt gist/full). **The breaking assertion is `TestJudgeSources_CoversEveryCategoryIncludingEstimate` (`collect_test.go:35`)** — it currently asserts `strings.Contains(s.Link, "prompts.go")` for all 8 categories; make it category-aware (lessons → `prompts.go`; others → `prompts/<c>.md`). (Note: `source_test.go`'s synthetic `Link: "prompts.go"` fixtures test `renderManual`, not `judgeSources` — leave them; they're unaffected.)
- [ ] **Step 4:** Run → PASS; full `go test ./cmd/sdlc/...`.
- [ ] **Step 5:** Commit — `git commit -m "#153 M2: process-manual links judge prompts to their .md files"`

---

### Task 8: Regenerate the tracked manual + atlas

**Files:** `atlas/process-manual.md` (regen), `atlas/workflow/process-manual.md` (note the `.md` source).

- [ ] **Step 1:** `go run ./cmd/sdlc process-manual --out atlas/process-manual.md`; confirm judge links now point at `../cmd/sdlc/internal/judge/prompts/<c>.md` and the memory section is still redacted (0 leak scan).
- [ ] **Step 2:** Update `atlas/workflow/process-manual.md`: judge prompts are now embedded `.md` (like help text/skills), linked directly; note the golden-test fidelity guard.
- [ ] **Step 3:** Commit — `git commit -m "#153 M2: regenerate manual + atlas note (judge prompts as .md)"`

---

## Estimate

To be set in the issue `## Estimate` block before `sdlc change-code` (recalibrated to
v3.1). Rough shape: one `cross-cutting-refactor`-flavored change confined to one package
(7 prose files moved + a loader + golden harness) + a small `process-manual` relink +
atlas regen + the M2 milestone-review. No new external surface, deterministic, guarded
by goldens → low risk, ~2–3h.

## Location (decided)

Prompt files live at **`cmd/sdlc/internal/judge/prompts/*.md`** (co-located with the
code that renders them) — confirmed by the operator (issue `## Log`, 2026-07-01 scope
entry). Not `cmd/sdlc/helptext/`: its `//go:embed *.md` would pull them into the manual's
Help-text section and mis-categorize them.
