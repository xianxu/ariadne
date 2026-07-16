# Derive `sdlc issue new` Template from the issue.cue Model — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `issue.Render` (the `sdlc issue new` on-disk template) derive its section list, section order, seed placeholders, and initial `status:` from `construct/vocabulary/issue.cue` instead of hardcoded Go literals — so a change in the cue model propagates to created issues with no Go edit.

**Architecture:** Add a concrete `scaffold.sections` block to `issue.cue` (CUE `#`-definitions don't export; only concrete data reaches the generated JSON — matches the existing `discovery:` pattern). `pkg/vocab` grows two pure read-only accessors — `Sections()` (the ordered section list) and `InitialStatus()` (`categories.open[0]`, no new data needed). `issue.Render` reads the embedded model via `vocab.Issue()` and loops over `Sections()` to emit the body, keeping the two dynamic behaviors (`--from-github` body → `## Problem`; dated `### <today>` → `## Log`) as name-keyed Go logic. Regenerate the embedded `pkg/vocab/issue.json` (the go:embed the code reads) and the `construct/generated/vocabulary/issue.json` consumer face.

**Tech Stack:** Go 1.26, CUE (build-time export via the `vocabulary` binary — no cuelang Go runtime dep), `//go:embed`.

**Architecture principles applied (`sdlc arch-principles`):**
- **ARCH-DRY** — this *is* the issue: the section list + `status: open` are duplicated as Go literals with no tie to the cue model. Fix single-sources them in `issue.cue`; `Render` reuses the existing `pkg/vocab` binding (no second cue encoding, no new dep).
- **ARCH-PURE** — `Sections()`, `InitialStatus()`, and `Render` are pure functions over the once-embedded model (build-time `//go:embed` seam). All tests run without mocks.
- **ARCH-PURPOSE** — shadow-sweep of every consumer that could restate the model is in the "Scope boundary" section below: `Render` is the one true consumer of the *creation template* and it is made to derive; `structural.go` and `helptext/issue.md` are documented as **not** consumers of this model (semantic-validation behavior and a human-facing superset doc, respectively) rather than silently skipped.

---

## Scope boundary (ARCH-PURPOSE shadow-sweep)

The purpose is: *the creation template derives from the cue model*. Enumerated consumers that reference issue sections, and the disposition of each:

| Site | What it does | Relationship to `scaffold.sections` |
|------|--------------|-------------------|
| `cmd/sdlc/internal/issue/scaffold.go:Render` | writes the new-issue skeleton | **DERIVES (this plan).** The one true consumer of the creation template — loops the model. |
| `cmd/sdlc/internal/issue/structural.go` | gates Spec (≥50 words) / Plan (≥1 checklist item) / Done-when (bullet-or-`related:`) | **SUBSET — drift-tested, not derived.** It is per-section *semantic validation behavior* over a subset (Problem/Log aren't gated); encoding word-counts/regex/fallback into cue would overmodel (Simplicity-First). But the sections it gates **must be a subset of the model** — a gate that validates a section the template never writes is a bug. Enforced by `TestGatedSectionsSubsetOfModel` (Task 5). |
| `cmd/sdlc/helptext/issue.md` | human `sdlc issue --help` "Body sections" doc | **SUPERSET — drift-tested, not derived.** It documents a superset (lists `## Estimate` + `## Side quests`, which `Render` never emits) as human reference for the file *format*; deriving it would *drop* those. But it **must document every modeled section** — a human doc that silently omits a creation section is drift. Enforced by `TestIssueHelpDocumentsEveryScaffoldSection` (Task 6). |

**The invariant chain (per your direction — enforce the containment, don't just document it):**

```
structural.go gated  ⊆  scaffold.sections (issue.cue)  ⊆  helptext/issue.md documented
     (subset)                 THE MODEL                        (superset)
```

The cue model sits in the middle and pins both neighbours by a *test*, not by prose: the gate can't validate a section the model doesn't create, and the human doc can't omit a section the model does create. Neither is made to *derive* (that would overmodel the gate and truncate the doc) — instead each is held to the model from its own side. If a future edit renames or adds a section in `issue.cue`, exactly one of the two tests fails and names the file to fix.

Consumer of issue **creation** outside this repo: `parley.nvim#116` delegates creation to `sdlc issue new` (confirmed done), so it derives transitively the moment `Render` does. Issue **status/location** it already sources from the emitted `issue.json`.

**Known name-coupling (documented, tested):** `Render` special-cases the `Problem` and `Log` sections by name (GH-body injection; dated subheading). Renaming either in cue needs a matching Go edit. Mitigated by `TestScaffold_SpecialSectionsPresent` (Task 5) which trips if the model no longer carries those names, and by a comment on both sides.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `scaffold.sections` (CUE concrete data) | `construct/vocabulary/issue.cue` | new |
| `Section` / `Scaffold` (Go types) | `pkg/vocab/vocab.go` | new |
| `IssueModel.Sections()` | `pkg/vocab/vocab.go` | new |
| `IssueModel.InitialStatus()` | `pkg/vocab/vocab.go` | new |
| `issue.Render` | `cmd/sdlc/internal/issue/scaffold.go` | modified |
| `gatedSections` (+ section-name consts) | `cmd/sdlc/internal/issue/structural.go` | new |
| `TestGatedSectionsSubsetOfModel` | `cmd/sdlc/internal/issue/structural_drift_test.go` | new |
| `TestIssueHelpDocumentsEveryScaffoldSection` | `cmd/sdlc/helptext/issue_sections_test.go` | new |

- **`scaffold.sections`** — an ordered list of `{name, seed?}` records: the body sections `sdlc issue new` writes, in order, each with an optional literal seed placeholder written under its heading (absent = bare heading). Concrete data so it exports to JSON.
  - **Relationships:** 1:1 with the `issue` noun; ordered (list order = output order). Sits beside `discovery:` as a second concrete "how instances are shaped/located" block.
  - **DRY rationale:** eliminates the hardcoded `## Problem/## Spec/## Done when/## Plan/## Log` list in `scaffold.go:Render` — the section vocabulary now lives once, in cue.
  - **Future extensions:** a `desc`/`optional` field per section if the human help doc is ever made to derive too (explicitly *not* now — see Scope boundary); more seed sections without a Go edit.

- **`Section` / `Scaffold`** — the Go decode targets for the `scaffold:` block. `Section{Name, Seed string}`.
  - **Relationships:** `IssueModel` holds one `Scaffold` (`Scaf` field, `json:"scaffold"`), mirroring the existing `Disc Discovery` field naming so the `Sections()` accessor carries the read name.

- **`IssueModel.Sections()`** — returns the ordered `[]Section`. Pure.

- **`IssueModel.InitialStatus()`** — returns `categories.open[0]` (the created-not-started status). Pure. No new cue data — the status is *already* modeled as the sole member of the `open` category, so this removes the `status: open` literal by derivation, not duplication.

- **`issue.Render`** — unchanged signature (`Render(ScaffoldSpec) string`) and unchanged byte output. Now reads `vocab.Issue()`: emits `status:` from `InitialStatus()` and the body by looping `Sections()`. Frontmatter field rendering (YAML nuances, optional `deps`/`target`/`github_issue`) stays as-is — it is rendering *logic*, not a restatement of the model.

- **`gatedSections`** — a package-level slice (built from new `secSpec`/`secPlan`/`secDoneWhen` constants) naming the sections `CheckSectionsPresence` validates. Introduced so the gated names are single-sourced *within* `structural.go` (ARCH-DRY) and become real data a test can read — the drift test asserts `gatedSections ⊆ model` rather than re-listing the names itself. The gate's *logic* (word counts, regex, fallback) is untouched; only the section-heading string literals become constants.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `//go:embed issue.json` | `pkg/vocab/vocab.go` | existing (regenerated) | build-time CUE export |

- **`//go:embed issue.json`** — the committed cue-export snapshot the Go binary reads (so a standalone `go build` needs no `cue`). Regenerated by `go run ./cmd/vocabulary export --noun issue > pkg/vocab/issue.json`. This is the seam that carries the new `scaffold` block into the runtime.
  - **Injected into:** every `vocab.Issue()` caller, including `Render`. Pure logic stays testable because the model is plain decoded JSON.

---

## Task 1: Model `scaffold.sections` in issue.cue

**Files:**
- Modify: `construct/vocabulary/issue.cue` (add after the `discovery:` block, ~line 53)

- [ ] **Step 1: Add the `scaffold` block + its section definition**

```cue
// ── scaffold: the on-disk creation template `sdlc issue new` writes (#145).
// Concrete data (CUE #-definitions don't export) so sdlc's issue.Render derives
// the section list + order + seed placeholders from HERE instead of a hardcoded
// Go template — add/rename/reorder a section and `sdlc issue new` follows with no
// Go edit. The initial `status:` is categories.open[0] (not restated). `seed` is
// the literal placeholder written under a blank section's heading (absent = bare
// heading). Two sections carry creation BEHAVIOUR that stays in Go, keyed by name
// — `Problem` receives a --from-github body, `Log` seeds a dated `### <today>`
// subheading; renaming either needs a matching Go edit (a test pins the names). ──
#ScaffoldSection: {
	name:  string
	seed?: string
}
scaffold: sections: [...#ScaffoldSection] & [
	{name: "Problem"},
	{name: "Spec"},
	{name: "Done when", seed: "-"},
	{name: "Plan", seed: "- [ ]"},
	{name: "Log"},
]
```

- [ ] **Step 2: Vet the model**

Run: `go run ./cmd/vocabulary vet` (or `cue vet construct/vocabulary/issue.cue`)
Expected: clean exit (0). The `[...#ScaffoldSection]` type unified with the concrete list follows the existing `lifecycle: [...#Transition] & [...]` pattern.

- [ ] **Step 3: Confirm the block exports**

Run: `go run ./cmd/vocabulary export --noun issue | grep -A14 '"scaffold"'`
Expected: a `"scaffold": { "sections": [ {"name":"Problem"}, ..., {"name":"Done when","seed":"-"}, {"name":"Plan","seed":"- [ ]"}, {"name":"Log"} ] }` block (unset `seed?` omitted).

- [ ] **Step 4: Commit**

```bash
git add construct/vocabulary/issue.cue
git commit -m "#145: model scaffold.sections in issue.cue"
```

---

## Task 2: Add `Section`/`Scaffold` types + `Sections()`/`InitialStatus()` to pkg/vocab (TDD)

**Files:**
- Modify: `pkg/vocab/vocab.go`
- Test: `pkg/vocab/vocab_test.go`

- [ ] **Step 1: Write failing tests**

Add to `pkg/vocab/vocab_test.go`:

```go
func TestInitialStatus(t *testing.T) {
	if got := vocab.Issue().InitialStatus(); got != "open" {
		t.Errorf("InitialStatus() = %q, want open (categories.open[0])", got)
	}
}

func TestSections(t *testing.T) {
	secs := vocab.Issue().Sections()
	// Order + seeds must match the cue model exactly (creation template shape).
	want := []struct {
		name, seed string
	}{
		{"Problem", ""},
		{"Spec", ""},
		{"Done when", "-"},
		{"Plan", "- [ ]"},
		{"Log", ""},
	}
	if len(secs) != len(want) {
		t.Fatalf("Sections() len = %d, want %d: %+v", len(secs), len(want), secs)
	}
	for i, w := range want {
		if secs[i].Name != w.name || secs[i].Seed != w.seed {
			t.Errorf("Sections()[%d] = {%q,%q}, want {%q,%q}", i, secs[i].Name, secs[i].Seed, w.name, w.seed)
		}
	}
}
```

(Match the package clause + import style already in `vocab_test.go` — if it's `package vocab` internal, drop the `vocab.` qualifier and call `Issue()` directly. Check the file header before writing.)

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./pkg/vocab/ -run 'TestInitialStatus|TestSections' -v`
Expected: FAIL — `Sections`/`InitialStatus` undefined (won't compile). This confirms the tests bind to new API.

- [ ] **Step 3: Add the types + accessors**

In `pkg/vocab/vocab.go`, add the types near `Discovery`:

```go
// Section is one body section of the new-issue creation template: a heading
// name and an optional literal seed line written beneath it in a blank issue.
// Ordered (list order = created-file order). Source: issue.cue `scaffold.sections`.
type Section struct {
	Name string `json:"name"`
	Seed string `json:"seed"`
}

// Scaffold is the parsed `scaffold:` block — the on-disk creation template shape
// that sdlc's `issue new` renders from (#145), instead of a hardcoded Go template.
type Scaffold struct {
	Sections []Section `json:"sections"`
}
```

Add the field to `IssueModel` (beside `Disc`):

```go
	// Scaf holds the scaffold: block; unexported-name-clash-avoiding (Scaf, not
	// Scaffold) so the Sections() accessor can carry the read name — mirrors Disc.
	Scaf Scaffold `json:"scaffold"`
```

Add the accessors near `Discovery()`:

```go
// Sections returns the ordered creation-template body sections, so the issue
// scaffolder derives the section list from the model instead of hardcoding it
// (#145).
func (m *IssueModel) Sections() []Section { return m.Scaf.Sections }

// InitialStatus returns the status a newly-created issue carries — the sole
// member of the `open` category — so the scaffolder's `status:` line derives
// from the model, not a Go literal (#145). Falls back to "open" only if a
// corrupt model defines no open status (mustLoadIssue already panics on corrupt
// JSON, so this is a belt-and-suspenders guard, not a real path).
func (m *IssueModel) InitialStatus() string {
	open := m.Categories["open"]
	if len(open) == 0 {
		return "open"
	}
	return open[0]
}
```

- [ ] **Step 4: Regenerate the embedded JSON so the new block is present at runtime**

Run: `go run ./cmd/vocabulary export --noun issue > pkg/vocab/issue.json`
Expected: `pkg/vocab/issue.json` now contains the `"scaffold"` block. (`git diff pkg/vocab/issue.json` shows only the added block.)

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./pkg/vocab/ -run 'TestInitialStatus|TestSections' -v`
Expected: PASS.

- [ ] **Step 6: Full vocab package test (no regressions)**

Run: `go test ./pkg/vocab/...`
Expected: PASS (conformance_test + vocab_test).

- [ ] **Step 7: Commit**

```bash
git add pkg/vocab/vocab.go pkg/vocab/vocab_test.go pkg/vocab/issue.json
git commit -m "#145: pkg/vocab — Sections() + InitialStatus() from the model"
```

---

## Task 3: Make `issue.Render` derive from the model (TDD — byte-stability first)

**Files:**
- Modify: `cmd/sdlc/internal/issue/scaffold.go`
- Test: `cmd/sdlc/internal/issue/scaffold_test.go`

- [ ] **Step 1: Add a byte-stability golden test BEFORE changing Render**

This locks the exact current output so the refactor is provably behavior-preserving. Add to `scaffold_test.go`:

```go
// TestRender_ByteStable pins the exact blank-issue output so deriving the
// template from the cue model (#145) provably preserves bytes.
func TestRender_ByteStable(t *testing.T) {
	got := Render(ScaffoldSpec{ID: "000057", Title: "Some new thing", Today: "2026-05-31"})
	want := `---
id: 000057
status: open
deps: []
github_issue:
created: 2026-05-31
updated: 2026-05-31
estimate_hours:
---

# Some new thing

## Problem

## Spec

## Done when

-

## Plan

- [ ]

## Log

### 2026-05-31
`
	if got != want {
		t.Errorf("Render byte-drift:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

// TestRender_ByteStable_FromGitHub locks the --from-github path too, so the
// Problem-body special-case is provably byte-preserved by the refactor.
func TestRender_ByteStable_FromGitHub(t *testing.T) {
	got := Render(ScaffoldSpec{
		ID: "000057", Title: "Imported", Today: "2026-05-31",
		GithubIssue: "42", ProblemBody: "The GH issue body.\n\nMore detail.",
	})
	want := `---
id: 000057
status: open
deps: []
github_issue: 42
created: 2026-05-31
updated: 2026-05-31
estimate_hours:
---

# Imported

## Problem

The GH issue body.

More detail.

## Spec

## Done when

-

## Plan

- [ ]

## Log

### 2026-05-31
`
	if got != want {
		t.Errorf("Render byte-drift (from-github):\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}
```

- [ ] **Step 2: Run it against the CURRENT (unmodified) Render to confirm the golden is exact**

Run: `go test ./cmd/sdlc/internal/issue/ -run TestRender_ByteStable -v`
Expected: PASS *before* any Render change. (If it fails, fix the golden string to match reality — the point is a faithful snapshot.)

- [ ] **Step 3: Add a model-driven invariant test (proves derivation, not coincidence)**

```go
// TestRender_DrivenByModel asserts Render emits exactly the model's sections in
// the model's order — so a cue section add/rename/reorder propagates (#145).
func TestRender_DrivenByModel(t *testing.T) {
	out := Render(ScaffoldSpec{ID: "000057", Title: "x", Today: "2026-05-31"})
	last := -1
	for _, sec := range vocab.Issue().Sections() {
		idx := strings.Index(out, "## "+sec.Name+"\n")
		if idx < 0 {
			t.Errorf("Render output missing section %q", sec.Name)
			continue
		}
		if idx < last {
			t.Errorf("section %q out of model order", sec.Name)
		}
		last = idx
	}
}
```

(Add `"github.com/xianxu/ariadne/pkg/vocab"` to the test's imports.)

- [ ] **Step 4: Rewrite Render to derive status + body from the model**

Add the import `"github.com/xianxu/ariadne/pkg/vocab"` to `scaffold.go`. Replace the hardcoded `status: open` line and the hardcoded section block:

```go
func Render(s ScaffoldSpec) string {
	m := vocab.Issue()
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\n", s.ID)
	fmt.Fprintf(&b, "status: %s\n", m.InitialStatus())
	fmt.Fprintf(&b, "deps: [%s]\n", strings.Join(s.Deps, ", "))
	if s.GithubIssue != "" {
		fmt.Fprintf(&b, "github_issue: %s\n", s.GithubIssue)
	} else {
		b.WriteString("github_issue:\n") // no trailing space when empty
	}
	if s.Target != "" {
		fmt.Fprintf(&b, "target: %s\n", s.Target)
	}
	fmt.Fprintf(&b, "created: %s\n", s.Today)
	fmt.Fprintf(&b, "updated: %s\n", s.Today)
	b.WriteString("estimate_hours:\n")
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "# %s\n\n", s.Title)

	// Body sections come from the cue model (#145): their names, order, and static
	// seed placeholders live in issue.cue `scaffold.sections`. Two sections carry
	// dynamic creation content that stays here, keyed by name — keep these in sync
	// with the modeled names (TestScaffold_SpecialSectionsPresent pins them).
	sections := m.Sections()
	for i, sec := range sections {
		fmt.Fprintf(&b, "## %s\n\n", sec.Name)
		content := sec.Seed
		switch sec.Name {
		case "Problem":
			content = strings.TrimSpace(s.ProblemBody) // --from-github body, else blank
		case "Log":
			content = fmt.Sprintf("### %s", s.Today) // dated session subheading
		}
		if content == "" {
			continue
		}
		b.WriteString(content)
		if i < len(sections)-1 {
			b.WriteString("\n\n")
		} else {
			b.WriteString("\n") // last section closes the file with a single newline
		}
	}
	return b.String()
}
```

Update the `Render` doc comment: note it derives section list + initial status from `vocab.Issue()` (the cue model), and drop the stale "single source of truth for the on-disk template" claim (the cue model is now the source; `Render` is the renderer).

- [ ] **Step 5: Run the full scaffold test suite**

Run: `go test ./cmd/sdlc/internal/issue/ -v`
Expected: PASS — `TestRender_ByteStable` (unchanged bytes), `TestRender_DrivenByModel`, `TestRender_Blank`, `TestRender_FromGitHub`, `TestRender_TargetAndDeps` all green.

- [ ] **Step 6: Commit**

```bash
git add cmd/sdlc/internal/issue/scaffold.go cmd/sdlc/internal/issue/scaffold_test.go
git commit -m "#145: issue.Render derives sections + status from the cue model"
```

---

## Task 4: Pin the name-coupling with a guard test

**Files:**
- Test: `cmd/sdlc/internal/issue/scaffold_test.go`

- [ ] **Step 1: Add the guard test**

```go
// TestScaffold_SpecialSectionsPresent guards the name-coupling: Render injects
// the --from-github body into "Problem" and the dated heading into "Log" by name.
// If the cue model renames either, this trips before a silent creation regression.
func TestScaffold_SpecialSectionsPresent(t *testing.T) {
	have := map[string]bool{}
	for _, sec := range vocab.Issue().Sections() {
		have[sec.Name] = true
	}
	for _, name := range []string{"Problem", "Log"} {
		if !have[name] {
			t.Errorf("scaffold model missing %q — Render special-cases it by name; update Render's switch if renamed", name)
		}
	}
}
```

- [ ] **Step 2: Run it**

Run: `go test ./cmd/sdlc/internal/issue/ -run TestScaffold_SpecialSectionsPresent -v`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add cmd/sdlc/internal/issue/scaffold_test.go
git commit -m "#145: guard the Problem/Log name-coupling in the scaffolder"
```

---

## Task 5: Enforce `structural.go` gated sections ⊆ model (subset invariant)

**Files:**
- Modify: `cmd/sdlc/internal/issue/structural.go` (single-source the gated names)
- Test: `cmd/sdlc/internal/issue/structural_drift_test.go` (new)

- [ ] **Step 1: Single-source the gated section names in structural.go**

Add near the top of `structural.go` (after the imports / before `CheckStructural`):

```go
// Section names the presence gates validate. Constants so the names are
// single-sourced within this file (ARCH-DRY) and pinned to the cue model by
// TestGatedSectionsSubsetOfModel. The gates' logic (word counts, checklist/bullet
// regex, related: fallback) is bespoke per section and intentionally NOT modeled
// in cue — only the section identity is.
const (
	secSpec     = "Spec"
	secDoneWhen = "Done when"
	secPlan     = "Plan"
)

// gatedSections is the set of sections CheckSectionsPresence enforces.
// INVARIANT (TestGatedSectionsSubsetOfModel): a subset of issue.cue
// scaffold.sections — a gate must not require a section the creation template
// never writes. (Note: checkPlan encodes "Plan" in PlanSectionRE, so a rename
// there needs a matching regex edit — the test fires to remind you.)
var gatedSections = []string{secSpec, secPlan, secDoneWhen}
```

Swap the section-heading string literals in the lookups to the constants (leave the human-readable failure `Message` prose and the `Name` tokens like `"spec-present"` UNCHANGED — those are pinned by `structural_test.go`):
- `checkSpecPresent`: `SectionBody(body, secSpec)`
- `checkSpecWordCount`: `SectionBody(body, secSpec)`
- `checkDoneWhen`: `SectionBody(body, secDoneWhen)`
- `checkPlan`: leave as-is (uses `PlanSectionRE`); `secPlan` is referenced by `gatedSections`.

- [ ] **Step 2: Write the drift test**

Create `cmd/sdlc/internal/issue/structural_drift_test.go`:

```go
package issue

import (
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// TestGatedSectionsSubsetOfModel enforces the invariant chain's left edge:
// every section the structural gate validates must exist in the cue creation
// template (gatedSections ⊆ scaffold.sections). A gate that requires a section
// `sdlc issue new` never writes would reject every fresh issue.
func TestGatedSectionsSubsetOfModel(t *testing.T) {
	model := map[string]bool{}
	var names []string
	for _, s := range vocab.Issue().Sections() {
		model[s.Name] = true
		names = append(names, s.Name)
	}
	for _, g := range gatedSections {
		if !model[g] {
			t.Errorf("structural gate targets %q, absent from issue.cue scaffold.sections %v — "+
				"reconcile structural.go (and PlanSectionRE if it's Plan) or issue.cue", g, names)
		}
	}
}
```

- [ ] **Step 3: Run it (and the existing structural tests, to prove the const swap didn't break pins)**

Run: `go test ./cmd/sdlc/internal/issue/ -run 'Structural|GatedSections|Section' -v`
Expected: PASS — `TestGatedSectionsSubsetOfModel` green; all pre-existing structural tests still green (the `Name`/`Message` pins are untouched).

- [ ] **Step 4: Commit**

```bash
git add cmd/sdlc/internal/issue/structural.go cmd/sdlc/internal/issue/structural_drift_test.go
git commit -m "#145: enforce structural gated sections ⊆ cue model"
```

---

## Task 6: Enforce `helptext/issue.md` documents ⊇ model (superset invariant)

**Files:**
- Test: `cmd/sdlc/helptext/issue_sections_test.go` (new)

The help doc stays hand-written (it documents a superset for humans). This test guards the containment so it can't silently *omit* a modeled creation section.

- [ ] **Step 1: Write the drift test**

Create `cmd/sdlc/helptext/issue_sections_test.go`:

```go
package helptext

import (
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// docSectionRE pulls "Name" out of the help doc's "Body sections" list lines,
// which look like `    ## Done when    acceptance criteria ...` — a `##` heading,
// the section name, then a 2+-space gutter before the description. The name may
// contain single spaces ("Done when", "Side quests"); the lazy group stops at the
// first 2+-space run. Line-anchored (?m), so the mid-line `## Plan` inside a
// description is never captured.
var docSectionRE = regexp.MustCompile(`(?m)^ *## (.+?)  +\S`)

// TestIssueHelpDocumentsEveryScaffoldSection enforces the invariant chain's right
// edge: `sdlc issue --help` documents a SUPERSET of the creation template, so
// every modeled scaffold.sections entry must appear in the help doc (the doc may
// list more — e.g. Estimate, Side quests). Catches "added a section to issue.cue
// but forgot to document it for humans."
func TestIssueHelpDocumentsEveryScaffoldSection(t *testing.T) {
	documented := map[string]bool{}
	for _, m := range docSectionRE.FindAllStringSubmatch(MustGet("issue"), -1) {
		documented[strings.TrimSpace(m[1])] = true
	}
	for _, s := range vocab.Issue().Sections() {
		if !documented[s.Name] {
			t.Errorf("issue.md `sdlc issue --help` omits section %q modeled in issue.cue "+
				"scaffold.sections; document it in the Body-sections list (documented=%v)", s.Name, documented)
		}
	}
}
```

- [ ] **Step 2: Run it — first confirm the regex extracts the right set**

Run: `go test ./cmd/sdlc/helptext/ -run TestIssueHelpDocumentsEveryScaffoldSection -v`
Expected: PASS. (If it fails claiming a section is missing, the regex mis-parsed — inspect by temporarily `t.Logf`-ing `documented`; it should be `{Problem, Spec, Done when, Estimate, Plan, Log, Side quests}`.)

- [ ] **Step 3: Commit**

```bash
git add cmd/sdlc/helptext/issue_sections_test.go
git commit -m "#145: enforce issue --help documents ⊇ cue model sections"
```

---

## Task 7: Regenerate the `construct/generated/vocabulary` consumer face

**Files:**
- Modify: `construct/generated/vocabulary/issue.json`, `construct/generated/vocabulary/.source-sha`

The embedded `pkg/vocab/issue.json` (what the code reads) is regenerated in Task 2. This file is the *separate* JSON face for non-Go consumers (e.g. parley Lua). It is stale on **three** axes independent of this change (you can't partially regen a generated face, so all land in one diff): (a) `active` lacks `codecomplete`; (b) `discovery` lacks the `plans`/`archive` fields (#144); (c) `lifecycle` lacks all the `codecomplete` edges (#160). Regenerating picks up `scaffold` **plus** all three latent drifts — name them all in the commit/Log so the close-boundary reviewer isn't surprised by the diff size.

- [ ] **Step 1: Regenerate via the materializer the drift check uses**

Run: `go run ./cmd/vocabulary export --output construct/generated/vocabulary`
Expected: `construct/generated/vocabulary/issue.json` gains the `scaffold` block *and* the `codecomplete` status + `codecomplete` lifecycle edges + `discovery.plans`/`discovery.archive`; `.source-sha` updates to the current `issue.cue` hash. Inspect `git diff --stat construct/generated/vocabulary/` — expect `issue.json` + `.source-sha` changed. If `SKILL.md` also churns, review whether that's intended pre-existing drift; if unrelated, note it in the Log rather than bundling silently.

- [ ] **Step 2: Sanity-check the two JSON faces now agree on scaffold**

Run: `diff <(go run ./cmd/vocabulary export --noun issue) pkg/vocab/issue.json && echo MATCH`
Expected: `MATCH` (embedded snapshot equals a fresh export). And `grep -q scaffold construct/generated/vocabulary/issue.json && echo OK`.

- [ ] **Step 3: Commit**

```bash
git add construct/generated/vocabulary/issue.json construct/generated/vocabulary/.source-sha
git commit -m "#145: regenerate vocabulary JSON face (scaffold + codecomplete status/edges + discovery.plans/archive drift)"
```

---

## Task 8: Full build + suite + manual behavior check

- [ ] **Step 1: Build sdlc and the whole module**

Run: `go build ./... && go vet ./...`
Expected: clean.

- [ ] **Step 2: Full test suite**

Run: `go test ./...`
Expected: PASS. Pay attention to `cmd/sdlc/...` (issue_test.go, validategate_test.go) — any test asserting created-issue shape must still pass byte-for-byte.

- [ ] **Step 3: Manual end-to-end — real `sdlc issue new` output (drive the flow, not just tests)**

Run: `go run ./cmd/sdlc issue new --dry-run "Derive check probe"` (title is a **positional** arg — `issue.go` `Use: "new <title>"`, not a `--title` flag; `--dry-run` does exist).
Expected: printed body has `status: open`, the five sections in order, `- [ ]` under Plan, `-` under Done when, and a dated `### <today>` under Log — identical to pre-change output. This exercises `Render` through the actual command path, satisfying the verification-before-completion bar.

- [ ] **Step 4: Prove propagation (the issue's core Done-when) without shipping a model change**

Temporarily add a throwaway `{name: "Risks"}` to `scaffold.sections` in `issue.cue`, run `go run ./cmd/vocabulary export --noun issue > pkg/vocab/issue.json`, then `go run ./cmd/sdlc issue new --dry-run probe` and confirm `## Risks` appears with **no Go edit**. Then `git checkout construct/vocabulary/issue.cue pkg/vocab/issue.json` to revert. Record the result in the issue `## Log` (this is the acceptance evidence; don't commit the throwaway section).

- [ ] **Step 5: Update the atlas**

Modify `atlas/workflow/vocabulary.md` (and/or `atlas/workflow/issue-lifecycle.md`): note that the issue **creation template** (`scaffold.sections`) is now modeled in `issue.cue` and consumed by `sdlc issue new` via `pkg/vocab` — so the section list is single-sourced with the lifecycle. Keep `atlas/index.md` linkage intact.

```bash
git add atlas/
git commit -m "#145: atlas — creation template now derives from the issue model"
```

---

## Close

Single-pass atomic change — **no `Mx` milestone tags** (one review boundary). Close in one shot:

- [ ] `sdlc close --issue 145 --verified '<byte-stable golden + model-driven propagation e2e; go test ./... green; sdlc issue new --dry-run identical output>'` (let it compute `--actual`).
- [ ] The mandatory fresh-context boundary review is auto-dispatched by `sdlc close`; fix any Critical/Important before crossing; log the `Review-Verdict:` outcome in `## Log`.

---

## Revisions

### 2026-07-05 — Task 7: the generated face is gitignored (no commit)

**Reason:** Implementing Task 7 revealed `/construct/generated/` is a `.gitignore`'d
build artifact (`.gitignore:30`), not a committed consumer face — `git ls-files
construct/generated/` is empty.

**Delta:** Task 7 no longer commits `construct/generated/vocabulary/issue.json` /
`.source-sha` (there is nothing to track). It reduces to: regenerate the **local**
artifact for hygiene via `go run ./cmd/vocabulary export --output construct/generated/vocabulary`
(picks up `scaffold` + the latent `codecomplete`/`discovery.plans`/`archive` drift),
but produce **no commit**. Downstream repos materialize this face themselves via
`make weave`. The single tracked deliverable of the exported model is the embedded
`pkg/vocab/issue.json` (Task 2). Consequently the plan-quality judge's finding #2
(a large generated diff surprising the close reviewer) cannot occur — the face is
never in the diff.
