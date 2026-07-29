# Tune the change-code gate — stateful plan review, estimate reorder, churn metrics

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `sdlc change-code`'s plan-quality gate **converge** — remember what it
asked for, block only on undisposed Critical/Important, cost the plan only after the
plan is accepted — and make the gate's own cost **measurable** at close.

**Architecture:** The root cause is that `runPlanQualityJudge` is stateless: it builds a
prompt, prints, and forgets. We give it memory by making the gate's output a **schema'd
handoff** (a fenced ` ```findings ` block modeled in CUE) instead of prose, persisting the
accumulated findings to a durable sidecar with binary-assigned stable IDs, and feeding the
prior rounds back into the next prompt. The *blocking decision* moves out of the LLM
(`judge.Classify` on a verdict token) and into a **pure function over the accumulated
ledger** — which is what structurally stops the "descend to the next-deepest layer every
round" loop the postmortem observed. The persistence + ID + decision layer is a
general-purpose `gatestate` package so #183 (`milestone-close --fixed-to-ship`) consumes
the same notion of gate state rather than inventing a second one (`ARCH-DRY`).

**Tech Stack:** Go 1.26, cobra, `go.yaml.in/yaml/v3` (already a direct dep), CUE via
`cmd/vocabulary` + `make vocab-embed`, embedded markdown prompts/helptext.

**Issue:** ariadne#187. **Milestones:** M1 = A+B+C (the gate fix), M2 = D (cost metrics).

**On this document's own form:** it reproduces a fair amount of final implementation
verbatim. That is this repo's plan-doc house style and it front-loads design where review
is cheapest — but note the tension with the "no procedural restatement of the diff" ask
Task 7 introduces, and that reproduced code goes stale silently if implementation drifts.
The reproduced blocks are **illustrative of shape and contract**, not authoritative: where
they and the committed code disagree, the code wins and this plan gets a `## Revisions`
entry. The parts that ARE authoritative are the entity tables, the test cases, and the
stated invariants.

---

## Deviations from the issue Spec (recorded here, folded into the issue's `## Revisions`)

1. **Sidecar filename is `NNNNNN-slug-plan-gate.md`, not `-plan-review.md`.**
   `construct/vocabulary/verdict.cue` declares `discovery: {home: "workshop/plans", glob:
   "*-review.md"}` — i.e. "a `*-review.md` sidecar **is** a verdict instance." A
   plan-quality gate record carries no boundary verdict, so filing it under that glob would
   hand a future verdict consumer a document it cannot validate. `-plan-gate.md` keeps the
   two artifact families disjoint at zero cost.

2. **Severity vocabulary is `Critical | Important | Minor`, not `…| Info`.** The Spec's
   A3 says "New findings at Info do not cost a round-trip", but `code-review.md` already
   uses `Critical/Important/Minor` for exactly this taxonomy. We single-source the existing
   three names (`ARCH-DRY`); `Minor` **is** the Spec's `Info` bucket. The judge-output
   *verdict token* `INFO` (contract.go) is a different noun and is untouched.

3. **`--force` becomes durable** (Spec D says it "prints to stderr and is not durable").
   A forced change-code appends a round to the sidecar with `forced: true` + the rationale.
   That is what makes D3's accepted-vs-forced count computable at all.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `FindingModel` (`finding.cue`) | `construct/vocabulary/finding.cue` | new |
| `FindingModel` (Go binding) | `pkg/vocab/finding.go` | new |
| `Finding` | `cmd/sdlc/internal/gatestate/ledger.go` | new |
| `Disposition` | `cmd/sdlc/internal/gatestate/ledger.go` | new |
| `Round` | `cmd/sdlc/internal/gatestate/ledger.go` | new |
| `Ledger` | `cmd/sdlc/internal/gatestate/ledger.go` | new |
| `ParseFindingsBlock` | `cmd/sdlc/internal/gatestate/parse.go` | new |
| `AssignIDs` / `Apply` / `OpenFindings` | `cmd/sdlc/internal/gatestate/ledger.go` | new |
| `Decide` | `cmd/sdlc/internal/gatestate/decide.go` | new |
| `Render` / `ParseSidecar` | `cmd/sdlc/internal/gatestate/render.go` | new |
| `RenderPriorFindings` | `cmd/sdlc/internal/gatestate/prompt.go` | new |
| `Bucket` / `ClassifyPath` | `cmd/sdlc/internal/churn/classify.go` | new (M2) |
| `Buckets` / `Report` / `Summarize` | `cmd/sdlc/internal/churn/report.go` | new (M2) |
| `LedgerRow` | `cmd/sdlc/internal/estimate/ledger.go` | modified (M2) |
| `PromptInput` | `cmd/sdlc/internal/judge/prompts.go` | modified |
| `sidecarPathFor` | `cmd/sdlc/reviewsidecar.go` | new |

- **`FindingModel` (CUE + Go binding)** — the single source for finding severities, which
  of them block a gate, and the legal dispositions. Mirrors `verdict.cue`/`verdict.go`
  exactly: `categories` is the one concrete datum, `#Severity`/`#Disposition` are derived
  via `or()`, only `categories`/`when` export to JSON, and `pkg/vocab/finding.json` is a
  committed `go generate` artifact validated by `make vocab-embed`.
  - **Relationships:** 1:N with `Finding` (each finding names one severity). Consumed by
    the plan-quality prompt (renders the emitted set), `ParseFindingsBlock` (validates
    against it), and `Decide` (reads `blocking`).
  - **DRY rationale:** Today the severity names live only in `code-review.md` prose. The
    moment a *binary* branches on severity, prose is the wrong source — this is precisely
    the `agent-binary-handoff-schema` target's named next boundary ("the change-code
    plan/estimate judges").
  - **Future extensions:** a `weight` per severity for a cost metric; a fourth severity.

- **`Finding` / `Disposition` / `Round` / `Ledger`** — the accumulated state of one gate
  across invocations. `Ledger` is the durable noun; `Round` is one invocation's delta
  (dispositions of prior findings + newly raised findings); `Finding` carries a
  binary-assigned stable ID.
  - **Relationships:** `Ledger` 1:N `Round`; `Round` 1:N `Finding` (raised) and 1:N
    `Disposition` (referencing findings from *earlier* rounds). A `Finding`'s ID is unique
    within a `Ledger`.
  - **DRY rationale:** #183 needs the identical shape at the *close* boundary — "a gate
    that remembers what it asked for". Building it once, gate-agnostic (`Ledger.Gate` is a
    string), is the whole reason this is a package and not three functions in
    `changecode.go`.
  - **Future extensions:** `Finding.Files []string` (the files a finding points at) is the
    hook #183's `--fixed-to-ship` hangs on — it can ask "did the files this finding named
    actually change?" without a new persistence format.

- **`ParseFindingsBlock`** — extracts the LAST fenced ` ```findings ` block from agent
  output and unmarshals it with `yaml/v3` into a `RoundReport`, validating every severity
  and disposition against `FindingModel`. Returns `ok=false` on a missing/invalid block —
  a genuine protocol miss, not a value to guess.
  - **DRY rationale:** mirrors `judge.ParseVerdictBlock`; the fence-extraction regex shape
    is the same pattern, one level of structure richer.

- **`Decide`** — the pure gate decision. **This is the load-bearing entity.** Given the
  post-apply `Ledger` and a round cap, returns block/pass + a human reason.
  - **Relationships:** consumes `Ledger` + `FindingModel.Blocking`; consumed by
    `runPlanQualityJudge`'s IO shell.
  - **DRY rationale:** first occurrence of "gate verdict computed from accumulated state
    rather than read off the LLM's token" — the pattern #183 repeats.

- **`Render` / `ParseSidecar`** — the two projections of a `Ledger` and the inverse of the
  machine one. `Render` emits YAML frontmatter (the machine view, the ONLY thing
  `ParseSidecar` reads) followed by generated human prose per round. The prose is
  *derived* from the frontmatter, so there is one source of truth in the file.
  - **Future extensions:** a `--json` dump for the metrics side (M2 reads the ledger, not
    the prose).

- **`ClassifyPath` / `Summarize` (M2)** — bucket a repo-relative path into
  `code-prod | code-test | atlas | workshop` and sum `git --numstat` insertions.
  - **DRY rationale:** the window base is *already* single-sourced in
    `boundaryWindowBase` (milestoneclose.go); churn reuses it rather than re-deriving, so
    the churn number provably covers the same commits as the review and the atlas gate.

- **`sidecarPathFor`** — `sidecarPath`'s suffix-parameterized core, so the boundary-review
  sidecar and the plan-gate sidecar share one stem derivation instead of two.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `planGateStore` | `cmd/sdlc/planreview.go` | new | filesystem (`workshop/plans/`) |
| `runPlanQualityJudge` | `cmd/sdlc/changecode.go` | modified | agent CLI subprocess |
| `churnForWindow` | `cmd/sdlc/churnreport.go` | new (M2) | `git diff`/`git log --numstat` |
| `appendCalibrationRow` | `cmd/sdlc/close.go` | modified (M2) | filesystem (TSV ledger) |

- **`planGateStore`** — read/write the plan-gate sidecar. Two functions
  (`readPlanGateLedger`, `writePlanGateLedger`) over `os.ReadFile` + the existing
  `atomicWriteFile`; all parsing/rendering delegates to `gatestate`.
  - **Injected into:** nothing — it is the shell. `gatestate` never touches the
    filesystem, so every ledger/decision test runs on in-memory strings with no mocks.
  - **Future extensions:** #183 adds `readMilestoneGateLedger` beside it; same package,
    same `gatestate` core.

- **`runPlanQualityJudge`** — already wraps the agent subprocess via `judge.Dispatch`
  (the existing seam; `judge.Run` is the injectable shim tests already fake). It gains:
  read ledger → render prior findings into `PromptInput` → dispatch → parse findings block
  → assign IDs → apply → `Decide` → write ledger. **No new external dependency**, so no
  new stateful fake is required (`ARCH-MOCK` satisfied by the existing `judge.Run` seam,
  which `dispatch_test.go` already fakes).

- **`churnForWindow` (M2)** — the only new external-command surface. It shells `git` via
  the error-returning `gitx.RunGit` seam (NOT `gitx.Capture`, which flattens errors to
  `""` and so could never fire Task 13's promised warning) — so
  tests drive it against a real temp git repo the way `closereview_test.go` already does
  (`ARCH-MOCK`: `git` is exercised for real in a disposable repo, not mocked).

---

## Chunk 1 — Milestone M1: the gate converges (A + B + C)

### Task 1: Model the finding vocabulary in CUE

**Files:**
- Create: `construct/vocabulary/finding.cue`
- Create: `pkg/vocab/finding.go`
- Create: `pkg/vocab/finding_test.go`
- Modify: `pkg/vocab/conformance_test.go`

- [ ] **Step 1: Write `finding.cue`**, mirroring `verdict.cue`'s shape exactly (concrete
  `categories`, derived `#Severity` via `or()`, `when` gloss, closed `#Finding`):

```cue
// finding — the vocabulary of a gate finding: the severities a fresh-context judge may
// emit, which of them BLOCK a gate, and the dispositions a later round may assign to a
// prior finding. Single source of truth for the `finding` noun (ariadne#187). The
// `agent-binary-handoff-schema` target names the change-code plan judge as the next
// boundary to schema; this is that schema.
package finding

import "list"

// ── categories: the single concrete source of severity membership ──
// The two categories PARTITION the severity set (a conformance test pins this).
categories: {
	blocking: ["Critical", "Important"] // undisposed ⇒ the gate refuses
	advisory: ["Minor"]                 // recorded for the close review; never blocks
}

// hardBlocking: the subset that blocks even PAST the round cap. Modeled rather than
// left as a `!= "Critical"` literal in Decide, for the same reason `blocking` is.
hardBlocking: ["Critical"]

#Severity: or(list.Concat([categories.blocking, categories.advisory]))

// ── dispositions: what a LATER round may say about an EARLIER finding ──
// PARTITIONED by the semantics the binary branches on, not a flat list. A flat list plus
// a prose gloss would put the closes-vs-leaves-open decision in a Go switch — the exact
// posture this file exists to reject. Concretely: adding `deferred` to a flat list makes
// ApplyChecked accept it while OpenFindings has no case for it, so the finding stays open
// forever — a silent wedge in the one package whose job is remembering state correctly.
dispositions: {
	closing: ["addressed", "withdrawn"] // the finding is settled; stops blocking
	open:    ["not-addressed"]          // still open; keeps blocking
}

#Disposition: or(list.Concat([dispositions.closing, dispositions.open]))

when: {
	"Critical":  "must fix before the gate is crossed"
	"Important": "fix before the gate if cheap; blocks until disposed"
	"Minor":     "note for the close review; never blocks a gate"
}

whenDisposed: {
	"addressed":     "the plan changed to satisfy this finding"
	"not-addressed": "still open — the judge re-raises it this round"
	"withdrawn":     "the judge retracts it (mistaken, or overtaken by a design change)"
}

// ── discovery: where instances of this noun live ──
discovery: {
	home: "workshop/plans"
	glob: "*-plan-gate.md"
}

// ── #Finding: the structured handoff the judge emits + the binary validates ──
// Closed (fail-closed): a finding is an atomic judgment, not a growing record.
#Finding: {
	id:       string    // "new" for a freshly raised finding; the binary assigns the real ID
	severity: #Severity
	title:    string
	detail?:  string
}

#Dispose: {
	id:          string       // a prior finding's binary-assigned ID
	disposition: #Disposition
	note?:       string
}
```

- [ ] **Step 2: Regenerate + verify the embed**

Run: `make vocab-embed`
Expected: creates `pkg/vocab/finding.json`; the drift assertion passes. (`finding.go`'s
`//go:generate` line must exist first — write Step 3 before running this, or run it twice.)

- [ ] **Step 3: Write `pkg/vocab/finding.go`**, byte-parallel to `verdict.go`:

```go
package vocab

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:generate sh -c "vocabulary export --noun finding > finding.json"

//go:embed finding.json
var findingJSON []byte

// FindingModel is the read-only, parsed `finding` noun: severities by blocking/advisory
// category, the legal dispositions, and per-token semantics. Derived from
// construct/vocabulary/finding.cue at generate time; never hand-edited (ariadne#187).
// The single Go read of the finding vocabulary — the plan-quality prompt's emitted set,
// gatestate's block validation, and gatestate.Decide's blocking test all derive from here.
type FindingModel struct {
	Categories   map[string][]string `json:"categories"`
	HardBlocking []string            `json:"hardBlocking"`
	Dispositions map[string][]string `json:"dispositions"` // "closing" / "open"
	When         map[string]string   `json:"when"`
	WhenDisposed map[string]string   `json:"whenDisposed"`
}

var findingModel = mustLoadFinding()

func mustLoadFinding() *FindingModel {
	var m FindingModel
	if err := json.Unmarshal(findingJSON, &m); err != nil {
		panic(fmt.Sprintf("vocab: corrupt embedded finding.json (run `make vocab-embed`): %v", err))
	}
	return &m
}

// Finding returns the embedded `finding` model.
func Finding() *FindingModel { return findingModel }

// Severities returns every severity, blocking first then advisory — the order the prompt
// renders and the sidecar groups by.
func (m *FindingModel) Severities() []string {
	out := make([]string, 0, len(m.Categories["blocking"])+len(m.Categories["advisory"]))
	out = append(out, m.Categories["blocking"]...)
	out = append(out, m.Categories["advisory"]...)
	return out
}

// IsSeverity reports whether s is a modeled severity (the parser's accepted set).
func (m *FindingModel) IsSeverity(s string) bool { return contains(m.Severities(), s) }

// Blocks reports whether an UNDISPOSED finding at severity s refuses the gate.
func (m *FindingModel) Blocks(s string) bool { return inCat(m.Categories, "blocking", s) }

// BlocksPastCap reports whether s still refuses the gate after the round cap is reached.
func (m *FindingModel) BlocksPastCap(s string) bool { return contains(m.HardBlocking, s) }

// AllDispositions returns every modeled disposition (closing then open).
func (m *FindingModel) AllDispositions() []string {
	return append(append([]string{}, m.Dispositions["closing"]...), m.Dispositions["open"]...)
}

// IsDisposition reports whether d is a modeled disposition.
func (m *FindingModel) IsDisposition(d string) bool { return contains(m.AllDispositions(), d) }

// Closes reports whether disposition d SETTLES a finding (stops it blocking). Derived
// from the model's partition, so a disposition added to finding.cue can never reach
// OpenFindings as an unhandled case.
func (m *FindingModel) Closes(d string) bool { return inCat(m.Dispositions, "closing", d) }
```

> Note on helpers: `inCat(cats map[string][]string, cat, s string)` already exists at
> `pkg/vocab/lifecycle.go:14` ("shared by every noun model") — use it for both
> map-shaped lookups above rather than adding a parallel helper (`ARCH-DRY`). The flat-slice
> `contains` needed by `IsSeverity`/`IsDisposition`/`BlocksPastCap` lives at
> `pkg/vocab/conformance_test.go:135` and is **test-only**, so `finding.go` cannot call it
> as-is: promote it verbatim to `pkg/vocab/vocab.go` and delete the test-file copy.

- [ ] **Step 4: Write the conformance test** in `pkg/vocab/finding_test.go` — derived from
  the model, not a maintained list (the `TestIssueConformance` posture):

```go
func TestFindingConformance(t *testing.T) {
	m := Finding()
	// Every severity is in exactly one category and carries a non-empty `when`.
	for _, s := range m.Severities() {
		n := 0
		for _, cat := range []string{"blocking", "advisory"} {
			if contains(m.Categories[cat], s) { n++ }
		}
		if n != 1 { t.Errorf("severity %q is in %d categories, want exactly 1", s, n) }
		if m.When[s] == "" { t.Errorf("severity %q has no `when` semantics", s) }
	}
	// Dispositions partition into exactly closing|open, each with semantics; Closes
	// agrees with the partition. Derived from the model — adding a disposition to
	// finding.cue without placing it in a category fails HERE, before it can reach
	// OpenFindings as an unhandled case.
	for _, d := range m.AllDispositions() {
		n := 0
		for _, cat := range []string{"closing", "open"} {
			if contains(m.Dispositions[cat], d) { n++ }
		}
		if n != 1 { t.Errorf("disposition %q is in %d categories, want exactly 1", d, n) }
		if m.WhenDisposed[d] == "" { t.Errorf("disposition %q has no semantics", d) }
		if !m.IsDisposition(d) { t.Errorf("IsDisposition(%q) = false", d) }
		if got := contains(m.Dispositions["closing"], d); m.Closes(d) != got {
			t.Errorf("Closes(%q) = %v, disagrees with the partition", d, m.Closes(d))
		}
	}
	// hardBlocking must be a SUBSET of blocking — a severity that blocks past the cap
	// but not before it would be incoherent.
	for _, s := range m.HardBlocking {
		if !m.Blocks(s) { t.Errorf("hardBlocking %q is not in categories.blocking", s) }
	}
	for _, s := range m.Categories["blocking"] {
		if !m.Blocks(s) { t.Errorf("Blocks(%q) = false for a blocking severity", s) }
	}
	for _, s := range m.Categories["advisory"] {
		if m.Blocks(s) { t.Errorf("Blocks(%q) = true for an advisory severity", s) }
	}
	if m.IsSeverity("Bogus") { t.Error("IsSeverity accepted an unmodeled severity") }
}
```

- [ ] **Step 5: Write the drift guard** — `code-review.md` must not name a severity the
  model doesn't have (this is what keeps the two prompts on one taxonomy). Add to
  `cmd/sdlc/internal/judge/judge_test.go`:

```go
// TestCodeReviewSeveritiesMatchModel pins code-review.md's severity buckets to the
// finding model (#187): the boundary review and the plan gate share ONE taxonomy, so a
// severity renamed in finding.cue can't leave the prose behind.
func TestCodeReviewSeveritiesMatchModel(t *testing.T) {
	body := CodeReviewBody(PromptInput{})
	for _, s := range vocab.Finding().Severities() {
		if !strings.Contains(body, s+" (") {
			t.Errorf("code-review.md does not name severity %q from the finding model", s)
		}
	}
}
```

- [ ] **Step 6: Run**

Run: `make vocab-embed && go test ./pkg/vocab/... ./cmd/sdlc/internal/judge/...`
Expected: PASS. If `TestCodeReviewSeveritiesMatchModel` fails, the model and
`code-review.md` disagree — fix the model, not the test.

- [ ] **Step 7: Commit**

```bash
git add construct/vocabulary/finding.cue pkg/vocab/finding.go pkg/vocab/finding.json \
        pkg/vocab/finding_test.go cmd/sdlc/internal/judge/judge_test.go
git commit -m "#187 M1: model the finding vocabulary in CUE

The change-code plan judge is the boundary the agent-binary-handoff-schema
target names next: deterministic code is about to branch on severity, so the
severity set stops being prose and becomes a schema every consumer derives from.

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>"
```

---

### Task 2: The `gatestate` ledger — pure core

**Files:**
- Create: `cmd/sdlc/internal/gatestate/ledger.go`
- Create: `cmd/sdlc/internal/gatestate/ledger_test.go`

- [ ] **Step 0: Write the shared test helpers FIRST**, in
  `cmd/sdlc/internal/gatestate/helpers_test.go`. Tasks 2, 3 and 5 all use them, so leaving
  them implicit means each task invents its own version. The contract is load-bearing —
  `findings(...)` must assign IDs from the same `PQ-<n>` sequence `AssignIDs` produces, or
  the `dispose("PQ-1", …)` calls in Task 3 silently reference nothing:

```go
// findings("Critical/seam", "Minor/naming") → []Finding with severity/title split on the
// first "/", and NO ids (Apply/AssignIDs assigns them).
func findings(specs ...string) []Finding

// round(n, dispositions, newFindings) → a Round with IDs already stamped as PQ-1, PQ-2, …
// in first-raised order across the whole ledger being built, so a later dispose("PQ-2",…)
// refers to the second finding ever raised. Timestamp/agent are fixed constants.
func round(n int, disp []Disposition, fs []Finding) Round

// dispose("PQ-1","addressed","PQ-2","withdrawn") → []Disposition from id/state pairs.
// Panics on an odd argument count.
func dispose(pairs ...string) []Disposition

// ledgerWith(rounds...) → Ledger{Gate:"plan-quality", IDPrefix:"PQ", Rounds: rounds}.
func ledgerWith(rs ...Round) Ledger

// ids(fs) → []string of finding IDs, for readable failure messages.
func ids(fs []Finding) []string
```

- [ ] **Step 1: Write the failing test** for ID assignment + apply + open-set:

```go
func TestAssignIDsAndApply(t *testing.T) {
	l := Ledger{Gate: "plan-quality", IDPrefix: "PQ"}

	// Round 1: three new findings get sequential stable IDs.
	r1 := AssignIDs(l, RoundReport{New: []Finding{
		{ID: "new", Severity: "Critical", Title: "seam in wrong layer"},
		{ID: "new", Severity: "Important", Title: "absorb layer swallows replies"},
		{ID: "new", Severity: "Minor", Title: "naming"},
	}}, 1, "2026-07-29T10:00:00Z", "claude")
	if got := []string{r1.New[0].ID, r1.New[1].ID, r1.New[2].ID}; !reflect.DeepEqual(got, []string{"PQ-1", "PQ-2", "PQ-3"}) {
		t.Fatalf("IDs = %v, want [PQ-1 PQ-2 PQ-3]", got)
	}
	l = Apply(l, r1)
	if n := len(OpenFindings(l)); n != 3 {
		t.Fatalf("open after round 1 = %d, want 3", n)
	}

	// Round 2: dispose two, raise one new — the new one continues the ID sequence.
	r2 := AssignIDs(l, RoundReport{
		Dispositions: []Disposition{
			{ID: "PQ-1", State: "addressed"},
			{ID: "PQ-2", State: "withdrawn"},
		},
		New: []Finding{{ID: "new", Severity: "Minor", Title: "comment density"}},
	}, 2, "2026-07-29T10:30:00Z", "claude")
	if r2.New[0].ID != "PQ-4" {
		t.Fatalf("new ID = %q, want PQ-4 (IDs never reuse)", r2.New[0].ID)
	}
	l = Apply(l, r2)

	open := OpenFindings(l)
	if len(open) != 2 {
		t.Fatalf("open after round 2 = %v, want PQ-3 + PQ-4", ids(open))
	}
}

// A disposition naming an unknown ID is a protocol error, not a silent no-op.
func TestApplyRejectsUnknownDispositionID(t *testing.T) {
	l := Apply(Ledger{Gate: "plan-quality", IDPrefix: "PQ"},
		Round{N: 1, New: []Finding{{ID: "PQ-1", Severity: "Critical", Title: "x"}}})
	r := Round{N: 2, Dispositions: []Disposition{{ID: "PQ-99", State: "addressed"}}}
	if _, err := ApplyChecked(l, r); err == nil {
		t.Error("disposing an unknown finding ID should error")
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./cmd/sdlc/internal/gatestate/ -run 'TestAssignIDs|TestApplyRejects' -v`
Expected: FAIL — package does not exist.

- [ ] **Step 3: Write `ledger.go`**

```go
// Package gatestate is the durable memory of an SDLC gate: the findings a
// fresh-context judge raised, the stable IDs the binary assigned them, how later
// rounds disposed of them, and the pure decision of whether the gate may be crossed.
//
// It exists because a stateless gate cannot converge (ariadne#187). `sdlc change-code`
// re-dispatched a brand-new plan reviewer on every invocation; with no memory of its own
// prior findings it re-derived an absolute bar each round and surfaced the next-deepest
// layer of a plan that kept improving — five rejections for one 126-line change. A gate
// that remembers says "you addressed my three findings, ship."
//
// Everything here is PURE (ARCH-PURE): no filesystem, no clock, no subprocess. The
// timestamp and agent name are captured at the IO boundary (cmd/sdlc/planreview.go) and
// passed in. That is what lets the whole convergence policy be tested on in-memory
// strings with no mocks.
//
// The package is deliberately gate-agnostic (Ledger.Gate + Ledger.IDPrefix are data, not
// constants) so ariadne#183's milestone-close `--fixed-to-ship` consumes the same notion
// of gate state rather than inventing a second one (ARCH-DRY).
package gatestate

import (
	"fmt"
	"strconv"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// Finding is one judgment the gate raised, with the binary-assigned stable ID that lets a
// LATER round refer to it. Severity is validated against the `finding` model at parse time.
type Finding struct {
	ID       string `yaml:"id"`
	Severity string `yaml:"severity"`
	Title    string `yaml:"title"`
	Detail   string `yaml:"detail,omitempty"`
	Round    int    `yaml:"round"` // the round that first raised it
}

// Disposition is a later round's verdict on an EARLIER finding.
type Disposition struct {
	ID    string `yaml:"id"`
	State string `yaml:"disposition"`
	Note  string `yaml:"note,omitempty"`
	Round int    `yaml:"round,omitempty"` // the round that disposed it
}

// RoundReport is what the judge emitted this round, BEFORE the binary assigns IDs.
type RoundReport struct {
	Dispositions []Disposition `yaml:"dispose,omitempty"`
	New          []Finding     `yaml:"findings,omitempty"`
}

// Round is one gate invocation's durable record: what it disposed, what it raised, and
// whether the operator forced past it.
type Round struct {
	N            int           `yaml:"n"`
	Timestamp    string        `yaml:"timestamp"`
	Agent        string        `yaml:"agent"`
	Dispositions []Disposition `yaml:"dispose,omitempty"`
	New          []Finding     `yaml:"findings,omitempty"`
	Forced        string `yaml:"forced,omitempty"`         // --force rationale, set ONLY when this gate blocked
	Blocked       bool   `yaml:"blocked"`                  // what Decide said, recorded for D3
	ProtocolError string `yaml:"protocol_error,omitempty"` // set when the judge emitted no valid findings block
}

// Ledger is the accumulated state of ONE gate on ONE issue across every invocation.
type Ledger struct {
	Gate     string  `yaml:"gate"`      // e.g. "plan-quality"
	IssueNum int     `yaml:"issue"`
	IDPrefix string  `yaml:"id_prefix"` // e.g. "PQ" — IDs are <prefix>-<n>
	Rounds   []Round `yaml:"rounds"`
	// ContentHash is sha256(issue+plan) as of the last PASSING round — the pass-through
	// key (#187 Task 8 Step 4a). #183's `--fixed-to-ship` is the same mechanism at the
	// close boundary, which is why it lives on the shared Ledger and not in changecode.go.
	ContentHash string `yaml:"content_hash,omitempty"`
}

// ContentHash is the pass-through key: sha256 of the issue + plan text the gate reviewed.
func ContentHash(issueContent, planContent string) string { … }

// PassesUnchanged reports whether the gate may skip dispatch: there is at least one round,
// the most recent one did NOT block, and the content is byte-identical to what it passed.
// Pure — the caller supplies the hash.
func PassesUnchanged(l Ledger, hash string) bool {
	if len(l.Rounds) == 0 || l.ContentHash == "" || hash != l.ContentHash {
		return false
	}
	return !l.Rounds[len(l.Rounds)-1].Blocked
}

// nextIDSeq returns one past the highest ID sequence number ever assigned in l. IDs are
// never reused, even after a finding is withdrawn — a stable ID that changes meaning
// between rounds would let the judge dispose the wrong thing.
func nextIDSeq(l Ledger) int {
	max := 0
	for _, r := range l.Rounds {
		for _, f := range r.New {
			if n, err := strconv.Atoi(trimPrefix(f.ID, l.IDPrefix)); err == nil && n > max {
				max = n
			}
		}
	}
	return max + 1
}

// AssignIDs stamps binary-assigned stable IDs onto a RoundReport's new findings and
// returns the durable Round. The judge emits `id: new`; assigning IDs here (not in the
// prompt) means the judge never has to invent a globally-unique identifier — it only has
// to REFER to the ones we handed it.
func AssignIDs(l Ledger, rr RoundReport, n int, timestamp, agent string) Round {
	seq := nextIDSeq(l)
	out := Round{N: n, Timestamp: timestamp, Agent: agent, Dispositions: rr.Dispositions}
	for _, f := range rr.New {
		f.ID = fmt.Sprintf("%s-%d", l.IDPrefix, seq)
		f.Round = n
		seq++
		out.New = append(out.New, f)
	}
	for i := range out.Dispositions {
		out.Dispositions[i].Round = n
	}
	return out
}

// Apply appends a round to the ledger. Use ApplyChecked when the round came from an agent.
func Apply(l Ledger, r Round) Ledger {
	l.Rounds = append(l.Rounds, r)
	return l
}

// ApplyChecked is Apply plus the protocol validation an agent-sourced round needs: every
// disposition must name a finding raised in an EARLIER round. A judge disposing an ID we
// never issued is a genuine protocol error to surface, not a value to guess
// (agent-binary-handoff-schema target).
func ApplyChecked(l Ledger, r Round) (Ledger, error) {
	known := map[string]bool{}
	for _, prev := range l.Rounds {
		for _, f := range prev.New {
			known[f.ID] = true
		}
	}
	for _, d := range r.Dispositions {
		if !known[d.ID] {
			return l, fmt.Errorf("round %d disposes unknown finding %q", r.N, d.ID)
		}
		if !vocab.Finding().IsDisposition(d.State) {
			return l, fmt.Errorf("round %d: unmodeled disposition %q for %s", r.N, d.State, d.ID)
		}
	}
	return Apply(l, r), nil
}

// OpenFindings returns every finding never disposed `addressed` or `withdrawn`, in
// ID order. A `not-addressed` disposition leaves the finding OPEN — that is the whole
// point: the judge saying "still not addressed" must keep blocking.
func OpenFindings(l Ledger) []Finding {
	// Closed-ness comes from the MODEL's closing/open partition, never a switch on
	// literals: a disposition added to finding.cue must not be able to reach here as an
	// unhandled case that silently leaves a finding open forever.
	m := vocab.Finding()
	closed := map[string]bool{}
	for _, r := range l.Rounds {
		for _, d := range r.Dispositions {
			closed[d.ID] = m.Closes(d.State)
		}
	}
	var out []Finding
	for _, r := range l.Rounds {
		for _, f := range r.New {
			if !closed[f.ID] {
				out = append(out, f)
			}
		}
	}
	return out
}
```

> `trimPrefix(id, prefix)` strips `"<prefix>-"`; write it as a 3-line unexported helper.

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./cmd/sdlc/internal/gatestate/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/sdlc/internal/gatestate/
git commit -m "#187 M1: gatestate ledger — stable IDs, dispositions, open-set

The binary assigns IDs so the judge only has to REFER to findings, never invent
globally-unique names for them. IDs never reuse: a stable ID that changed meaning
between rounds would let a later round dispose the wrong finding.

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>"
```

---

### Task 3: `Decide` — the convergence policy

**Files:**
- Create: `cmd/sdlc/internal/gatestate/decide.go`
- Create: `cmd/sdlc/internal/gatestate/decide_test.go`

This is the entity that makes the gate converge. Spell the policy out:

1. Block iff at least one **open** finding has a **blocking** severity (`Critical` or
   `Important`, from the model).
2. New `Minor` findings never block — they land in the ledger for the close review.
3. **Round cap** (A3): once `round > cap` (default 3, `WF_PLAN_ROUND_CAP` overrides),
   only open `Critical` blocks. Important findings are recorded and reported, not blocking.
   Rationale: the observed failure was a reviewer descending Critical → Important → Info
   across rounds; the cap bounds the tail without ever letting a Critical through.

- [ ] **Step 1: Write the failing tests** — these are the Done-when criteria as code:

```go
// The headline regression: round 2 on a plan whose round-1 Critical/Important were
// addressed PASSES, even though the judge raised a new Minor.
func TestDecideConvergesWhenPriorBlockersAddressed(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Critical/seam", "Important/absorb-layer")),
		round(2, dispose("PQ-1", "addressed", "PQ-2", "addressed"), findings("Minor/naming")),
	)
	d := Decide(l, 3)
	if d.Block {
		t.Fatalf("gate blocked after all blockers addressed: %s", d.Reason)
	}
	if d.OpenMinor != 1 {
		t.Errorf("OpenMinor = %d, want 1 (recorded, not blocking)", d.OpenMinor)
	}
}

func TestDecideBlocksOnOpenCritical(t *testing.T) {
	l := ledgerWith(round(1, nil, findings("Critical/seam")))
	if d := Decide(l, 3); !d.Block {
		t.Fatal("an open Critical must block")
	}
}

func TestDecideBlocksOnNotAddressed(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Important/absorb-layer")),
		round(2, dispose("PQ-1", "not-addressed"), nil),
	)
	if d := Decide(l, 3); !d.Block {
		t.Fatal("`not-addressed` must leave the finding blocking")
	}
}

func TestDecideWithdrawnDoesNotBlock(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Critical/mistaken")),
		round(2, dispose("PQ-1", "withdrawn"), nil),
	)
	if d := Decide(l, 3); d.Block {
		t.Fatalf("a withdrawn finding must not block: %s", d.Reason)
	}
}

// Past the cap, Important is recorded but no longer costs a round-trip; Critical still does.
func TestDecideRoundCapDemotesImportantButNotCritical(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Minor/a")), round(2, nil, findings("Minor/b")),
		round(3, nil, findings("Minor/c")), round(4, nil, findings("Important/late")),
	)
	if d := Decide(l, 3); d.Block {
		t.Fatalf("past the cap an open Important must not block: %s", d.Reason)
	}
	l2 := Apply(l, round(5, nil, findings("Critical/real")))
	if d := Decide(l2, 3); !d.Block {
		t.Fatal("an open Critical must block even past the round cap")
	}
}

// An empty ledger (no round dispatched yet) must not block. Decide is normally consulted
// only after a round is applied, but a defensive read of a fresh ledger must be safe.
func TestDecideEmptyLedgerPasses(t *testing.T) {
	if d := Decide(Ledger{Gate: "plan-quality", IDPrefix: "PQ"}, 3); d.Block {
		t.Fatal("an empty ledger must not block")
	}
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./cmd/sdlc/internal/gatestate/ -run TestDecide -v`
Expected: FAIL — `Decide` undefined.

- [ ] **Step 3: Write `decide.go`**

```go
package gatestate

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// DefaultRoundCap is the round after which only Critical findings block (ariadne#187 A3).
// Three rounds is where the pair#127 postmortem showed the reviewer had stopped finding
// substance and started descending severity levels on a plan that kept improving.
const DefaultRoundCap = 3

// Decision is the gate's answer: block or not, plus the reason to print and the counts
// the close-time metrics (D3) read back.
type Decision struct {
	Block        bool
	Reason       string
	OpenBlocking []Finding // open Critical/Important (post-cap: the ones that still block)
	OpenMinor    int
	Rounds       int
	CapReached   bool
}

// Decide is the gate's pure decision over the accumulated ledger. `roundCap` is spelled
// out rather than `cap` so it doesn't shadow the builtin.
//
// Block iff some finding is still OPEN at a blocking severity. This is the mechanic that
// makes the gate converge: a judge that raises a fresh Minor every round can no longer
// cost a round-trip, and a judge that has disposed its own prior findings sees the gate
// open. Compare the pre-#187 behavior, where the decision was read off the LLM's verdict
// token and every fresh reviewer re-derived an absolute bar.
//
// Past `cap` rounds, only Critical blocks. Important findings raised that late are
// recorded in the ledger and reported to the operator — they reach the close review, which
// is the boundary that catches what a plan review structurally cannot.
func Decide(l Ledger, roundCap int) Decision {
	if roundCap <= 0 {
		roundCap = DefaultRoundCap
	}
	m := vocab.Finding()
	d := Decision{Rounds: len(l.Rounds), CapReached: len(l.Rounds) > roundCap}

	var demoted []Finding
	for _, f := range OpenFindings(l) {
		switch {
		case !m.Blocks(f.Severity):
			d.OpenMinor++
		case d.CapReached && !m.BlocksPastCap(f.Severity):
			demoted = append(demoted, f)
		default:
			d.OpenBlocking = append(d.OpenBlocking, f)
		}
	}

	if len(d.OpenBlocking) == 0 {
		d.Reason = fmt.Sprintf("no open blocking findings after %d round(s)", d.Rounds)
		if len(demoted) > 0 {
			d.Reason += fmt.Sprintf("; %d Important finding(s) recorded but not blocking (round cap %d reached)", len(demoted), roundCap)
		}
		if d.OpenMinor > 0 {
			d.Reason += fmt.Sprintf("; %d Minor finding(s) recorded for the close review", d.OpenMinor)
		}
		return d
	}

	d.Block = true
	var b strings.Builder
	fmt.Fprintf(&b, "%d open blocking finding(s):", len(d.OpenBlocking))
	for _, f := range d.OpenBlocking {
		fmt.Fprintf(&b, "\n  [%s] %s — %s", f.ID, f.Severity, f.Title)
	}
	d.Reason = b.String()
	return d
}
```

- [ ] **Step 4: Run to verify they pass**

Run: `go test ./cmd/sdlc/internal/gatestate/ -v`
Expected: PASS — all six `TestDecide*` cases.

- [ ] **Step 5: Commit**

```bash
git add cmd/sdlc/internal/gatestate/decide.go cmd/sdlc/internal/gatestate/decide_test.go
git commit -m "#187 M1: Decide — the gate decision moves out of the LLM

Blocking is now a pure function of the accumulated ledger, not a token read off
the judge's output. That is what structurally stops the observed round-trip loop:
a fresh Minor can no longer cost a round, and disposed blockers open the gate.

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>"
```

---

### Task 4: Parse the schema'd handoff

**Files:**
- Create: `cmd/sdlc/internal/gatestate/parse.go`
- Create: `cmd/sdlc/internal/gatestate/parse_test.go`

- [ ] **Step 1: Write the failing tests**

```go
const goodBlock = "prose the judge wrote first\n\n" +
	"```findings\n" +
	"dispose:\n" +
	"  - id: PQ-1\n    disposition: addressed\n    note: seam moved to the filter\n" +
	"findings:\n" +
	"  - id: new\n    severity: Critical\n    title: absorb layer swallows solicited replies\n" +
	"    detail: capability negotiation breaks silently\n" +
	"```\n"

func TestParseFindingsBlock(t *testing.T) {
	rr, ok := ParseFindingsBlock(goodBlock)
	if !ok { t.Fatal("valid block should parse") }
	if len(rr.Dispositions) != 1 || rr.Dispositions[0].ID != "PQ-1" ||
		rr.Dispositions[0].State != "addressed" {
		t.Errorf("dispositions = %+v", rr.Dispositions)
	}
	if len(rr.New) != 1 || rr.New[0].Severity != "Critical" || rr.New[0].ID != "new" {
		t.Errorf("findings = %+v", rr.New)
	}
}

// LAST block wins — a judge that shows an example block before its real one must not
// hand us the example (the ParseVerdictBlock precedent).
func TestParseFindingsBlockLastWins(t *testing.T) {
	in := "```findings\nfindings:\n  - id: new\n    severity: Minor\n    title: example\n```\n" + goodBlock
	rr, ok := ParseFindingsBlock(in)
	if !ok || len(rr.New) != 1 || rr.New[0].Severity != "Critical" {
		t.Errorf("last block should win, got %+v", rr.New)
	}
}

// Fail-closed: an unmodeled severity is a protocol error, not a value to guess.
func TestParseFindingsBlockRejectsUnmodeledSeverity(t *testing.T) {
	in := "```findings\nfindings:\n  - id: new\n    severity: Catastrophic\n    title: x\n```\n"
	if _, ok := ParseFindingsBlock(in); ok {
		t.Error("an unmodeled severity must not parse")
	}
}

func TestParseFindingsBlockRejectsUnmodeledDisposition(t *testing.T) {
	in := "```findings\ndispose:\n  - id: PQ-1\n    disposition: maybe\n```\n"
	if _, ok := ParseFindingsBlock(in); ok {
		t.Error("an unmodeled disposition must not parse")
	}
}

// A finding with no title is unusable in the sidecar and in the next round's prompt.
func TestParseFindingsBlockRejectsTitlelessFinding(t *testing.T) {
	in := "```findings\nfindings:\n  - id: new\n    severity: Minor\n```\n"
	if _, ok := ParseFindingsBlock(in); ok {
		t.Error("a titleless finding must not parse")
	}
}

func TestParseFindingsBlockAbsent(t *testing.T) {
	if _, ok := ParseFindingsBlock("VERDICT: CLEAN\n\nlooks good"); ok {
		t.Error("no block ⇒ ok=false (caller falls back + warns)")
	}
}

// An EMPTY block is a valid, meaningful statement: "no findings this round".
func TestParseFindingsBlockEmptyIsValid(t *testing.T) {
	rr, ok := ParseFindingsBlock("```findings\nfindings: []\n```\n")
	if !ok { t.Fatal("an explicitly empty block must parse") }
	if len(rr.New) != 0 || len(rr.Dispositions) != 0 { t.Errorf("got %+v", rr) }
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./cmd/sdlc/internal/gatestate/ -run TestParseFindings -v`
Expected: FAIL — `ParseFindingsBlock` undefined.

- [ ] **Step 3: Write `parse.go`**

```go
package gatestate

import (
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// findingsBlockRE extracts the body of a fenced ```findings … ``` block — the structured
// handoff the plan judge emits. Same shape as judge.verdictBlockRE, one level of structure
// richer. (?s) so `.` spans newlines.
var findingsBlockRE = regexp.MustCompile("(?s)```+[ \t]*findings[ \t]*\r?\n(.*?)\r?\n```+")

// ParseFindingsBlock extracts the LAST fenced ```findings block and validates every
// severity and disposition against the `finding` model. This is the AUTHORITATIVE
// structured handoff (the agent-binary-handoff-schema target): it does NOT parse prose, so
// a missing or model-invalid block is a genuine protocol miss (ok=false), never a
// heuristic read of the judge's paragraphs. Pure.
func ParseFindingsBlock(output string) (RoundReport, bool) {
	ms := findingsBlockRE.FindAllStringSubmatch(output, -1)
	if len(ms) == 0 {
		return RoundReport{}, false
	}
	var rr RoundReport
	if err := yaml.Unmarshal([]byte(ms[len(ms)-1][1]), &rr); err != nil {
		return RoundReport{}, false
	}
	m := vocab.Finding()
	for _, f := range rr.New {
		if !m.IsSeverity(f.Severity) || strings.TrimSpace(f.Title) == "" {
			return RoundReport{}, false
		}
	}
	for _, d := range rr.Dispositions {
		if !m.IsDisposition(d.State) || strings.TrimSpace(d.ID) == "" {
			return RoundReport{}, false
		}
	}
	return rr, true
}
```

- [ ] **Step 4: Run to verify they pass**

Run: `go test ./cmd/sdlc/internal/gatestate/ -v`
Expected: PASS.

- [ ] **Step 5: Fuzz it — the mechanical guard, not more enumerated cases.**

This is the plan's own C1 ask applied to the plan's own riskiest function.
`ParseFindingsBlock` is a parser over **unbounded LLM output**: `title` and `detail` are
free-form agent text. The seven cases above are all syntactically well-formed — which is
verbatim the pathology #187's Problem section indicts (*"30 hand-written cases all fed
syntactically valid sequences, and the close review found a panic on malformed input"*),
and C1's own headline example ("byte scanner over arbitrary device output → fuzz it,
seeded with malformed forms") describes exactly this function. Shipping enumerated-only
coverage here, in the issue that introduces the ask, would be indefensible.

**There is no `func Fuzz` anywhere in this repo today** — this is the first, so it also
becomes the reference instance of the guard the new prompt will demand of every future plan.

```go
// FuzzParseFindingsBlock: the parser must never panic, and must never return ok=true with
// a finding whose severity is unmodeled or whose title is blank — the two invariants every
// downstream consumer (AssignIDs, Decide, Render) assumes without re-checking.
func FuzzParseFindingsBlock(f *testing.F) {
	f.Add(goodBlock)
	f.Add("```findings\nfindings: []\n```\n")
	// Seeds targeting the specific structural hazards, not random noise:
	f.Add("```findings\nfindings:\n  - id: new\n    severity: Minor\n    title: |\n      block scalar\n```\n")
	f.Add("```findings\nfindings:\n  - id: new\n    severity: Minor\n    title: x\n    detail: \"has\\n---\\nfence\"\n```\n")
	f.Add("```findings\nfindings:\n  - id: new\n    severity: Minor\n    title: x\n    detail: \"```findings nested\"\n```\n")
	f.Add("```findings\ndispose:\n  - id: \"\"\n    disposition: addressed\n```\n")
	f.Fuzz(func(t *testing.T, s string) {
		rr, ok := ParseFindingsBlock(s)   // must not panic
		if !ok { return }
		m := vocab.Finding()
		for _, fd := range rr.New {
			if !m.IsSeverity(fd.Severity) || strings.TrimSpace(fd.Title) == "" {
				t.Fatalf("ok=true with an invalid finding: %+v", fd)
			}
		}
		for _, d := range rr.Dispositions {
			if !m.IsDisposition(d.State) || strings.TrimSpace(d.ID) == "" {
				t.Fatalf("ok=true with an invalid disposition: %+v", d)
			}
		}
	})
}
```

Run: `go test ./cmd/sdlc/internal/gatestate -run Fuzz -fuzz FuzzParseFindingsBlock -fuzztime 60s`
Expected: no failures. Commit any `testdata/fuzz/` corpus entries the run produces.

- [ ] **Step 6: Commit**

```bash
git add cmd/sdlc/internal/gatestate/parse.go cmd/sdlc/internal/gatestate/parse_test.go \
        cmd/sdlc/internal/gatestate/testdata/
git commit -m "#187 M1: parse the findings handoff, validated against the model

Fail-closed per the agent-binary-handoff-schema target: an unmodeled severity or
disposition is a protocol error the caller surfaces, not a near-miss we guess at.

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>"
```

---

### Task 5: Render + re-read the sidecar, and render prior findings into the prompt

**Files:**
- Create: `cmd/sdlc/internal/gatestate/render.go`
- Create: `cmd/sdlc/internal/gatestate/render_test.go`
- Create: `cmd/sdlc/internal/gatestate/prompt.go`
- Create: `cmd/sdlc/internal/gatestate/prompt_test.go`

The sidecar is **YAML frontmatter (machine) + generated prose (human)**. `ParseSidecar`
reads *only* the frontmatter; the prose is derived from the same `Ledger`, so the document
has one source of truth.

- [ ] **Step 1: Write the failing round-trip test**

```go
func TestRenderParseRoundTrip(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Critical/seam in wrong layer", "Minor/naming")),
		round(2, dispose("PQ-1", "addressed"), findings("Important/lock contract unstated")),
	)
	l.Gate, l.IssueNum, l.IDPrefix = "plan-quality", 187, "PQ"

	got, err := ParseSidecar(Render(l, "ariadne"))
	if err != nil { t.Fatalf("ParseSidecar: %v", err) }
	if !reflect.DeepEqual(got, l) {
		t.Errorf("round-trip lost data:\n got %+v\nwant %+v", got, l)
	}
}

// The human prose must actually carry the findings — a reader opening the sidecar should
// not have to read YAML.
func TestRenderProseCarriesFindings(t *testing.T) {
	out := Render(ledgerWith(round(1, nil, findings("Critical/seam in wrong layer"))), "ariadne")
	for _, want := range []string{"## Round 1", "PQ-1", "Critical", "seam in wrong layer"} {
		if !strings.Contains(out, want) { t.Errorf("rendered sidecar missing %q", want) }
	}
}

func TestParseSidecarRejectsMissingFrontmatter(t *testing.T) {
	if _, err := ParseSidecar("# Plan gate\n\nno frontmatter here\n"); err == nil {
		t.Error("a sidecar without frontmatter must error, not yield an empty ledger")
	}
}
```

- [ ] **Step 2: Write the failing prior-findings prompt test**

```go
// The prompt block must (a) list every OPEN finding with its ID and severity so the judge
// can dispose them, and (b) list disposed IDs so the judge does not re-raise them — the
// A2/A3 contract.
func TestRenderPriorFindings(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Critical/seam", "Important/absorb", "Minor/naming")),
		round(2, dispose("PQ-1", "addressed"), nil),
	)
	out := RenderPriorFindings(l)
	for _, want := range []string{"PQ-2", "Important", "absorb", "PQ-3", "PQ-1", "addressed"} {
		if !strings.Contains(out, want) { t.Errorf("prior-findings block missing %q", want) }
	}
	if !strings.Contains(out, "round 2") && !strings.Contains(out, "2 prior round") {
		t.Error("prior-findings block must state how many rounds ran")
	}
}

// Round 1 has no prior state — the block must say so explicitly rather than render empty,
// so the judge knows it is the FIRST reviewer, not one whose history was dropped.
func TestRenderPriorFindingsEmpty(t *testing.T) {
	out := RenderPriorFindings(Ledger{Gate: "plan-quality", IDPrefix: "PQ"})
	if !strings.Contains(strings.ToLower(out), "first") {
		t.Errorf("empty ledger should announce a first round, got %q", out)
	}
}
```

- [ ] **Step 3: Run to verify they fail**

Run: `go test ./cmd/sdlc/internal/gatestate/ -run 'TestRender|TestParseSidecar' -v`
Expected: FAIL — undefined.

- [ ] **Step 4: Implement `render.go` + `prompt.go`**

`Render(l Ledger, repo string) string`:
- `---\n` + `yaml.Marshal(l)` + `---\n\n`
- `# Plan gate — <repo>#<issue> (<gate>)\n\n`
- per round: `## Round <n> — <timestamp> (<agent>)`, a `blocked: yes|no` line, the
  `forced:` rationale when present, a `### Disposed` list (`- PQ-1 — addressed — <note>`)
  and a `### Raised` list (`- **PQ-4** [Critical] <title>` + indented detail).
- a trailing `## Open findings` section rendered from `OpenFindings(l)`.

`ParseSidecar(text string) (Ledger, error)`:
- use **`frontmatter.Split(content) (fm, body string, err error)`** (`pkg/frontmatter/frontmatter.go:26`)
  to separate the two halves — do **not** hand-roll a `---` splitter (`ARCH-DRY`) — then
  `yaml.Unmarshal([]byte(fm), &l)`. Propagate `Split`'s error; a sidecar with no
  frontmatter is a corrupt sidecar, not an empty ledger.

`RenderPriorFindings(l Ledger) string`:
- empty ledger → `"This is the FIRST plan-quality round for this issue; there are no prior findings."`
- otherwise: a header naming the round count, an `OPEN FINDINGS — you must dispose EVERY
  one of these before raising anything new` list (ID, severity, title, detail), and an
  `ALREADY DISPOSED — do NOT re-raise these, at any severity` list (ID, title, state).

- [ ] **Step 5: Fuzz the round-trip.** `Render` writes agent-authored `title`/`detail`
  into a `---`-fenced YAML frontmatter that `ParseSidecar` reads back via
  `frontmatter.Split`. A `detail` containing a line that is exactly `---`, or a ``` fence,
  is the structural hazard — and the round-trip is the property, so this is a property
  test, not a case list:

```go
// FuzzRenderParseRoundTrip: for ANY title/detail an agent might write, a rendered sidecar
// must parse back to the identical Ledger. The hazard is a detail containing `\n---\n`
// (which would terminate the frontmatter early) or a code fence.
func FuzzRenderParseRoundTrip(f *testing.F) {
	f.Add("seam in wrong layer", "moves the filter boundary")
	f.Add("x", "has\n---\na fence line")
	f.Add("x", "has ```findings inside")
	f.Add("", "")
	f.Fuzz(func(t *testing.T, title, detail string) {
		l := Ledger{Gate: "plan-quality", IssueNum: 187, IDPrefix: "PQ",
			Rounds: []Round{{N: 1, Timestamp: "2026-07-29T10:00:00Z", Agent: "claude",
				New: []Finding{{ID: "PQ-1", Severity: "Critical", Title: title, Detail: detail, Round: 1}}}}}
		got, err := ParseSidecar(Render(l, "ariadne"))
		if err != nil { t.Fatalf("round-trip failed for title=%q detail=%q: %v", title, detail, err) }
		if !reflect.DeepEqual(got, l) { t.Fatalf("round-trip lost data:\n got %+v\nwant %+v", got, l) }
	})
}
```

> If this fails on the `---` seed — and it plausibly will, since `frontmatter.Split` scans
> for a closing fence — the fix belongs in `Render`, not the test: `yaml.Marshal` must
> quote/indent such values so no emitted line is a bare `---`. Verify what `yaml/v3`
> actually emits before assuming it handles this.

- [ ] **Step 6: Run to verify everything passes**

Run: `go test ./cmd/sdlc/internal/gatestate/ -v` then
`go test ./cmd/sdlc/internal/gatestate -run Fuzz -fuzz FuzzRenderParseRoundTrip -fuzztime 60s`
Expected: PASS, no fuzz failures.

- [ ] **Step 7: Commit**

```bash
git add cmd/sdlc/internal/gatestate/render.go cmd/sdlc/internal/gatestate/render_test.go \
        cmd/sdlc/internal/gatestate/prompt.go cmd/sdlc/internal/gatestate/prompt_test.go \
        cmd/sdlc/internal/gatestate/testdata/
git commit -m "#187 M1: sidecar render/parse + the prior-findings prompt block

Frontmatter is the machine view, prose the human view, both projected from one
Ledger — so the document cannot disagree with itself.

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>"
```

---

### Task 6: The IO shell — persist the plan-gate ledger

**Files:**
- Create: `cmd/sdlc/planreview.go`
- Create: `cmd/sdlc/planreview_test.go`
- Modify: `cmd/sdlc/reviewsidecar.go` (extract `sidecarPathFor`)

- [ ] **Step 1: Write the failing test**

```go
// The plan-gate sidecar lands beside the boundary-review sidecars, under a DISTINCT
// suffix so a verdict consumer globbing *-review.md never sees a plan gate.
func TestPlanGateSidecarPath(t *testing.T) {
	got := planGatePath("workshop/plans", "000187-tune-change-code-gate.md")
	want := "workshop/plans/000187-tune-change-code-gate-plan-gate.md"
	if got != want { t.Errorf("planGatePath = %q, want %q", got, want) }
	if strings.HasSuffix(got, "-review.md") {
		t.Error("plan-gate sidecar must not match verdict.cue's *-review.md discovery glob")
	}
}

// A missing sidecar is the normal round-1 state: an empty ledger, not an error.
func TestReadPlanGateLedgerAbsent(t *testing.T) {
	l, err := readPlanGateLedger(t.TempDir(), "000187-x.md", 187)
	if err != nil { t.Fatalf("absent sidecar should not error: %v", err) }
	if len(l.Rounds) != 0 || l.Gate != "plan-quality" || l.IDPrefix != "PQ" {
		t.Errorf("empty ledger = %+v", l)
	}
}

func TestWriteThenReadPlanGateLedger(t *testing.T) {
	dir := t.TempDir()
	l, _ := readPlanGateLedger(dir, "000187-x.md", 187)
	l = Apply(l, gatestate.AssignIDs(l, gatestate.RoundReport{
		New: []gatestate.Finding{{ID: "new", Severity: "Critical", Title: "seam"}},
	}, 1, "2026-07-29T10:00:00Z", "claude"))
	if err := writePlanGateLedger(dir, "000187-x.md", l, "ariadne"); err != nil {
		t.Fatalf("write: %v", err)
	}
	back, err := readPlanGateLedger(dir, "000187-x.md", 187)
	if err != nil { t.Fatalf("read: %v", err) }
	if len(back.Rounds) != 1 || back.Rounds[0].New[0].ID != "PQ-1" {
		t.Errorf("round-tripped ledger = %+v", back)
	}
}

// A corrupt sidecar must NOT be silently reset to empty — that would erase the memory the
// whole feature exists to keep, and silently re-open every disposed finding.
func TestReadPlanGateLedgerCorrupt(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "000187-x-plan-gate.md"), []byte("---\n:::not yaml:::\n---\n"), 0o644)
	if _, err := readPlanGateLedger(dir, "000187-x.md", 187); err == nil {
		t.Error("a corrupt sidecar must error, not silently start from an empty ledger")
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./cmd/sdlc/ -run 'TestPlanGate|TestReadPlanGate|TestWriteThenRead' -v`
Expected: FAIL — undefined.

- [ ] **Step 3: Refactor `sidecarPath` in `reviewsidecar.go`**

```go
// sidecarPathFor derives a sidecar path from an issue filename stem plus a suffix.
// The one stem derivation shared by the boundary-review sidecar (#136) and the plan-gate
// sidecar (#187), so both track the issue's slug from a single source (ARCH-DRY).
func sidecarPathFor(plansDir, issueFileName, suffix string) string {
	stem := strings.TrimSuffix(filepath.Base(issueFileName), ".md")
	return filepath.Join(plansDir, stem+"-"+suffix+".md")
}

// sidecarPath derives the boundary-review sidecar path: `NNNNNN-slug-close-review.md`
// for a whole-issue close, `NNNNNN-slug-m<x>-review.md` for milestone `Mx` (lowercased).
func sidecarPath(plansDir, issueFileName, milestone string) string {
	suffix := "close-review"
	if milestone != "" {
		suffix = strings.ToLower(milestone) + "-review"
	}
	return sidecarPathFor(plansDir, issueFileName, suffix)
}
```

- [ ] **Step 4: Write `planreview.go`** — thin IO over `gatestate`:

```go
// planreview.go — the IO shell for change-code's plan-gate ledger (#187).
//
// The gate's memory lives in `workshop/plans/NNNNNN-slug-plan-gate.md`. This file is the
// ONLY place that touches the filesystem or the clock for that ledger; all parsing,
// rendering, ID assignment and gate decisions are pure in internal/gatestate (ARCH-PURE).
//
// Deliberately NOT named `-plan-review.md`: construct/vocabulary/verdict.cue declares
// `discovery.glob: "*-review.md"` — that glob means "this document carries a boundary
// verdict". A plan gate carries findings, not a verdict, so it stays out of that family.
package main

const planGateSuffix = "plan-gate"

func planGatePath(plansDir, issueFileName string) string {
	return sidecarPathFor(plansDir, issueFileName, planGateSuffix)
}

// readPlanGateLedger loads the ledger, or returns a fresh empty one when the sidecar does
// not exist yet (the normal round-1 state). A sidecar that EXISTS but does not parse is an
// error, never an empty ledger: silently resetting would erase every disposition and
// re-open findings the operator already addressed.
func readPlanGateLedger(plansDir, issueFileName string, issueNum int) (gatestate.Ledger, error) { … }

// writePlanGateLedger renders and atomically writes the ledger (reusing atomicWriteFile).
func writePlanGateLedger(plansDir, issueFileName string, l gatestate.Ledger, repo string) error { … }
```

- [ ] **Step 5: Pin the archiving coverage.** `archivePlanArtifacts` (`push.go:281`) globs
  `filepath.Join(plansFull, id+"-*")` (`push.go:286`), so `-plan-gate.md` is already
  caught. Note it is called from **`sdlc push` (`push.go:610`) and `sdlc merge`
  (`merge.go:652`) — not from `close`**; archiving happens at publish, not at close.
  `archiveartifacts_test.go` fixtures only `-plan.md` and `-close-review.md`, leaving the
  coverage incidental: add `000143-x-plan-gate.md` to the fixture set in
  **`TestArchivePlanArtifacts`** (`archiveartifacts_test.go:29`, fixtures at ~:33) and to
  its moved-files assertion (~:44), so a future narrowing of the glob fails loudly.

- [ ] **Step 6: Run**

Run: `go test ./cmd/sdlc/ -run 'PlanGate|Sidecar|Archive' -v`
Expected: PASS (including the pre-existing `sidecarPath` tests, unchanged).

- [ ] **Step 7: Commit**

```bash
git add cmd/sdlc/planreview.go cmd/sdlc/planreview_test.go cmd/sdlc/reviewsidecar.go
git commit -m "#187 M1: persist the plan-gate ledger under workshop/plans/

A corrupt sidecar errors rather than resetting to empty — silently forgetting is
the exact failure the issue is about, and a reset would re-open disposed findings.

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>"
```

---

### Task 7: Rewrite the plan-quality prompt (A2 + C1 + C2)

**Files:**
- Modify: `cmd/sdlc/internal/judge/prompts/plan-quality.md`
- Modify: `cmd/sdlc/internal/judge/prompts.go` (`PromptInput.PriorFindings`, `{{PRIOR_FINDINGS}}`, `{{FINDINGS_BLOCK}}`)
- Modify: `cmd/sdlc/internal/judge/golden_test.go` + `testdata/`

- [ ] **Step 1: Add the tokens to `PromptInput` + `promptSubstitutions`**

```go
	PriorFindings string // rendered prior-round findings (plan-quality, #187)
```
and in `promptSubstitutions`:
```go
		"{{PRIOR_FINDINGS}}", orDefault(in.PriorFindings, "(no prior rounds)"),
		"{{FINDINGS_BLOCK}}", vocab.Finding().RenderBlockInstruction(),
```

- [ ] **Step 2: Add `RenderBlockInstruction` as a METHOD on `FindingModel`**, in
  `pkg/vocab/finding.go` — not in `judge/contract.go`. The precedent it mirrors,
  `VerdictModel.RenderBlockInstruction` (`pkg/vocab/verdict.go:67`), is a method on the
  model *consumed from* `judge/review.go:43`. Putting the finding version in the judge
  package would split "a vocabulary model renders its own handoff instruction" across two
  packages on its second instance — the one point where the pattern is still cheap to keep
  single (`ARCH-DRY`). This also means the prompt layer holds no severity names at all.

```go
// RenderBlockInstruction renders the structured findings-handoff instruction — the fenced
// ```findings block template + the per-severity and per-disposition gloss — entirely from
// the model, so the prompt's accepted set never drifts from finding.cue (#187, the
// agent-binary-handoff-schema target). Mirrors VerdictModel.RenderBlockInstruction.
func (m *FindingModel) RenderBlockInstruction() string { … }
```

Then in `judge/prompts.go`'s `promptSubstitutions`, wire it directly:
`"{{FINDINGS_BLOCK}}", vocab.Finding().RenderBlockInstruction(),`.

- [ ] **Step 3: Write the guard test** in `judge_test.go` — the prompt must render the
  model, not a hardcoded list:

```go
func TestPlanQualityPromptRendersFindingModel(t *testing.T) {
	p := BuildPrompt(PlanQuality, PromptInput{})
	for _, s := range vocab.Finding().Severities() {
		if !strings.Contains(p, s) { t.Errorf("plan-quality prompt omits severity %q", s) }
	}
	// AllDispositions(), not .Dispositions — the latter is map[string][]string since the
	// closing/open partition landed, so ranging it yields []string and won't compile.
	for _, d := range vocab.Finding().AllDispositions() {
		if !strings.Contains(p, d) { t.Errorf("plan-quality prompt omits disposition %q", d) }
	}
	if !strings.Contains(p, "```findings") { t.Error("prompt must show the findings block") }
}

// The prior-findings block must actually reach the prompt (A2's whole mechanism).
func TestPlanQualityPromptCarriesPriorFindings(t *testing.T) {
	p := BuildPrompt(PlanQuality, PromptInput{PriorFindings: "SENTINEL-PRIOR-BLOCK"})
	if !strings.Contains(p, "SENTINEL-PRIOR-BLOCK") {
		t.Error("PriorFindings did not reach the rendered plan-quality prompt")
	}
}
```

- [ ] **Step 4: Rewrite `plan-quality.md`.** Keep the framing, `{{ARCH_BLOCK}}`, and the
  issue/plan payload. Change three things:

  **(a) Prior-round disposal comes FIRST** — insert before the failure-mode list:

  ```markdown
  ## Prior rounds — dispose of these BEFORE raising anything new

  {{PRIOR_FINDINGS}}

  This gate has memory. Your FIRST obligation is to state, for every OPEN finding above,
  whether the current plan `addressed` it, left it `not-addressed`, or whether you
  `withdrawn` it as mistaken or overtaken by a design change. Only then may you raise
  something new.

  Do NOT re-raise a finding listed as already disposed — not at the same severity, and
  not at a lower one. If a disposed finding is genuinely still wrong, dispose it
  `not-addressed` by its ID instead of raising it again as new.

  A plan that has addressed every prior Critical and Important finding is DONE. Say so and
  raise nothing further. Perfect is not the bar; executable is.
  ```

  **(b) Replace the "Missing test surface" failure mode (C1).** Delete the bullet that
  reads *"Missing test surface — the Plan changes code but doesn't say what behavior the
  tests will pin"* and add a dedicated section:

  ```markdown
  ## What the plan must say about tests — and what it must NOT

  REQUIRE: the **functions** that will be unit-tested, by name, plus **one line of
  strategy per risky function** — the adversarial input class and the mechanical guard.
  Example of the whole obligation, done right:

      byte scanner over arbitrary device output → fuzz it, seeded with malformed forms

  REJECT (raise a finding telling the plan to compress) any of:
    - an enumerated list of test cases in prose. Every case will be rewritten as code
      within the hour; the prose is a lossy pre-image of an executable artifact, and
      enumeration systematically misses the malformed-input class that enumerated cases
      are by construction blind to.
    - a line-numbered inventory of call sites.
    - a procedural restatement of the diff the implementer is about to write.

  These three cost real authoring time, are stale on arrival, and buy nothing the code
  will not state better. One strategy line per risky function is worth fifteen bullets.
  ```

  **(c) Keep and sharpen what paid off (C2)** — add to the failure-mode list:

  ```markdown
    - Unstated hard-to-reverse decisions — the plan changes a seam, a layer boundary, an
      ownership relation, or a lock/transaction contract without saying so explicitly.
      These are the findings worth a round-trip; almost nothing else is.
    - Unbacked claims about EXISTING behavior — the plan asserts what current code does
      without a `file:line`. Verify each such claim against the code and flag the ones
      that are wrong. (Factual errors by the implementing agent about existing code are
      historically this gate's highest-yield catch.)
    - No stated non-goals — the plan never says what it is deliberately NOT building, and
      why.
  ```

  **(d) Replace the tri-state token section** with the findings block, keeping
  `{{CONTRACT}}`'s `VERDICT:` line as the documented fallback:

  ```markdown
  {{FINDINGS_BLOCK}}

  {{CONTRACT}}

  Tokens for this check (advisory — the BLOCKING decision is computed by the binary from
  the findings block above, not from this token):
    CLEAN   = plan is concrete, testable, scoped, architecturally sound; safe to start.
    INFO    = plan is workable; only Minor findings.
    FAILURE = at least one open Critical or Important finding.
  ```

- [ ] **Step 5: Regenerate the golden fixtures**

Run: `go test ./cmd/sdlc/internal/judge -run TestBuildPrompt_Golden -update-golden`
then `go test ./cmd/sdlc/internal/judge/... -v`
Expected: PASS.

> ⚠️ `golden_test.go:37` carries a ⛔ standing rule: **never** re-run `-update-golden` to
> "fix" a failure — the goldens exist to catch unintended prompt drift. Re-capturing is
> legitimate *here* and only here, because rewriting `plan-quality.md` is the deliberate
> point of this task. Re-capture **once**, then read the resulting `testdata/golden/*.prompt`
> diff line by line and confirm every changed line is one you intended. Any other prompt's
> golden changing is a bug in this task, not a fixture to bless.

- [ ] **Step 6: Commit**

```bash
git add cmd/sdlc/internal/judge/
git commit -m "#187 M1: plan-quality prompt — dispose-first, strategy-not-enumeration

Prior findings lead the prompt so the judge converges instead of re-deriving an
absolute bar. The test ask changes from enumerating cases (a lossy pre-image of
code, which on pair#127 missed the malformed-input panic outright) to naming the
functions plus one adversarial-strategy line each.

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>"
```

---

### Task 8: Wire change-code — stateful judge + gate reorder (A1/A2/A3 + B1)

**Files:**
- Modify: `cmd/sdlc/changecode.go`
- Modify: `cmd/sdlc/changecode_test.go`

- [ ] **Step 1: Write the failing reorder test.** `runChangeCode`'s gate order is currently
  implicit in a 100-line function; extract the order into a pure, testable list first:

```go
// TestGateOrderPlanBeforeEstimate pins B1: the estimate is a FUNCTION of the plan, so it
// is never demanded before plan-quality has accepted the plan. Costing an unapproved plan
// is waste by construction — pair#127 re-derived its estimate five times, four of them
// forced by plan changes.
func TestGateOrderPlanBeforeEstimate(t *testing.T) {
	order := changeCodeGateOrder()
	idx := func(name string) int {
		for i, g := range order { if g == name { return i } }
		t.Fatalf("gate %q missing from the order", name); return -1
	}
	if idx("structural") > idx("plan-quality") {
		t.Error("structural must run before plan-quality (it is free)")
	}
	for _, est := range []string{"estimate", "estimate-recon", "estimate-quality"} {
		if idx(est) < idx("plan-quality") {
			t.Errorf("%s must run AFTER plan-quality (#187 B1)", est)
		}
	}
}
```

- [ ] **Step 2: Write the failing stateful-judge tests.** Drive `runPlanQualityJudge`
  through the existing fake-judge seam — `stubJudge(t, output) (*int, *string)`
  (`closereview_test.go:48`) installs a fake agent and hands back a call counter plus the
  captured prompt, which is exactly what the "prior findings reached round 2" assertion
  needs. Extend it to return a *sequence* of outputs (round 1, round 2) rather than
  writing a second stub:

```go
// Round 1 raises a Critical; the gate refuses and the sidecar records the finding.
func TestPlanQualityRound1BlocksAndPersists(t *testing.T) { … }

// Round 2 disposes it `addressed` and raises only a Minor: the gate PASSES.
// This is the issue's headline Done-when criterion.
func TestPlanQualityRound2ConvergesAfterAddressed(t *testing.T) {
	// fake judge output round 1: Critical "seam in wrong layer"
	// fake judge output round 2: dispose PQ-1 addressed + new Minor
	// assert: round 1 err != nil, round 2 err == nil
	// assert: sidecar has 2 rounds, OpenFindings == [the Minor]
}

// Prior findings actually reach round 2's prompt.
func TestPlanQualityFeedsPriorFindingsIntoPrompt(t *testing.T) {
	// capture the prompt the fake receives on round 2; assert it contains "PQ-1"
}

// No findings block ⇒ warn loudly and fall back to judge.Classify (transitional posture).
func TestPlanQualityFallsBackWithoutFindingsBlock(t *testing.T) {
	// fake emits "VERDICT: FAILURE" and no block
	// assert: err != nil AND stderr mentions the missing block
}

// --force records a durable forced round (D3's accepted-vs-forced signal).
func TestPlanQualityForceRecordsDurableRound(t *testing.T) {
	// assert the sidecar's last round has Forced == "<rationale>"
}
```

- [ ] **Step 3: Run to verify they fail**

Run: `go test ./cmd/sdlc/ -run 'TestGateOrder|TestPlanQuality' -v`
Expected: FAIL.

- [ ] **Step 4: Implement.** In `runChangeCode`, move blocks 3b/3c to *after* the
  plan-quality judge, and split the judge block:

```go
	// 4. Plan-quality judge (fresh-context LLM, now stateful — #187 A1/A2/A3).
	if !f.NoJudge {
		if err := runPlanQualityJudge(…); err != nil { … }
	}

	// 5. Estimate gates (#113/#117) — RELOCATED here, downstream of plan acceptance
	//    (#187 B1). The estimate is a function of the plan: demanding one before the
	//    plan has been looked at once forces a recomputation per plan revision, which is
	//    waste by construction. Nothing above this line mentions the estimate.
	//
	//    B1 is only a net win BECAUSE of the pass-through short-circuit above (step 4a):
	//    without it, every estimate-gate failure would cost a fresh 3-minute plan-quality
	//    dispatch on the retry, where today it costs milliseconds.
	if fail := estimateRefusal(…); fail != nil { … }
	if fail := estimateReconRefusal(…); fail != nil { … }
	if !f.NoJudge {
		if err := runEstimateQualityJudge(…); err != nil { … }
	}
```

**Make the declared order the ACTUAL order — do not restate it (`ARCH-DRY`).** A
hand-written `[]string{...}` plus a test that reads it guards a *restatement*: move the
estimate block back above plan-quality in `runChangeCode` and the test still passes, so
the sole guard for B1 — half this issue — cannot fail on the regression it exists to catch.
That is the same drift mode Task 9 diagnoses for prose. Make `runChangeCode` **iterate**
the declaration, so the list *is* the order:

```go
// gate is one named step in change-code's sequence. Declaring the gates as data and
// RUNNING that data is what makes changeCodeGateOrder() a real guard: reordering the
// literal reorders execution, so TestGateOrderPlanBeforeEstimate fails on the regression
// it exists to catch rather than passing on a restatement (ARCH-DRY).
type gate struct {
	name string
	run  func() error // nil error = passed; the closure owns its own --no-<gate> skip
}

// changeCodeGates is the ordered sequence. #187 B1: plan-quality precedes every estimate
// gate, because the estimate is a function of the plan.
func changeCodeGates(ctx *changeCodeCtx) []gate {
	return []gate{
		{"structural", ctx.structural},
		{"plan-quality", ctx.planQuality},
		{"estimate", ctx.estimate},
		{"estimate-recon", ctx.estimateRecon},
		{"estimate-quality", ctx.estimateQuality},
	}
}

// changeCodeGateOrder returns just the names, for the ordering test.
func changeCodeGateOrder() []string { … }
```

and `runChangeCode` becomes a loop over `changeCodeGates(ctx)` that applies the shared
`--force` handling once, instead of five hand-sequenced blocks each repeating it.

> This also collapses the five near-identical `if f.Force == "" { … exitWithCode(1) } …
> cwarn("… bypassed (--force: %s)")` blocks (`changecode.go:124-136`, `:144-152`,
> `:159-167`, `:172-177`, `:178-183`) into one — the duplication an `ARCH-DRY` reviewer
> would flag in the diff anyway.

Rewrite `runPlanQualityJudge` to the stateful shape:

**Step 4a — the pass-through short-circuit. B1 does not work without this.**

Moving the estimate gates *below* plan-quality inverts a cost that used to be free. Today
an estimate failure exits at `changecode.go:143`/`:158` in **milliseconds**, before the
judge. After B1 the sequence becomes:

```
invocation N:   plan-quality dispatch (~3 min) → passes → estimate gate FAILS
                (no `## Estimate` block yet — which is exactly what B2's new prose tells
                 the agent to expect: derive the estimate only once the plan clears)
invocation N+1: plan-quality dispatch AGAIN (~3 min) → passes → estimate gates run
```

So B2's own instruction guarantees a wasted 3-minute re-dispatch on **every issue**, and
the one genuine estimation error pair#127 hit would now cost a full judge round instead of
milliseconds. Shipping B1 without a short-circuit would make the gate *more* expensive —
the opposite of this issue's purpose.

The fix is the mechanism #187's `## Log` already names: **gate state keyed on a content
hash** — "#183 proposes gate-owned state keyed on content hashes … worth designing the two
together so there is one notion of gate state" (`ARCH-DRY`). One field on `Ledger`:

```go
	ContentHash string `yaml:"content_hash,omitempty"` // sha256 of issue+plan at the last PASSING round
```

When the last round passed **and** `sha256(issueContent + planContent)` equals
`ledger.ContentHash`, the plan has not changed since it was accepted: **skip
`judge.Dispatch` entirely**, print `plan-quality: unchanged since round N — passing
through (cached)`, persist **no** round, and fall straight to the estimate gates. So the
estimate-failure retry costs milliseconds again, and the operator can iterate the
`## Estimate` block freely without re-paying for plan review.

Two consequences this also fixes, both otherwise self-inflicted:
- `Decide`'s `CapReached: len(l.Rounds) > roundCap` counts **invocations**, not substantive
  rounds. Estimate-driven re-runs and `protocol_error` rounds would burn the default cap of
  3, silently demoting a genuinely new Important raised at round 4. Cached invocations
  persisting no round keeps the cap counting review rounds.
- M2's `gate_rounds` — the number introduced to answer "which gates earn their cost" —
  would otherwise count exactly the noise this reorder creates, reporting the tuning as
  more expensive than the status quo for reasons unrelated to review quality.

Add the tests:

```go
// The short-circuit is what makes B1 a net win: an unchanged plan must not re-dispatch.
func TestPlanQualityPassThroughOnUnchangedContent(t *testing.T) {
	// round 1 passes → second call with identical issue+plan content
	// assert: the fake judge was invoked ONCE, err == nil, ledger still has 1 round
}

// An EDITED plan must re-dispatch — the short-circuit must not cache away a real review.
func TestPlanQualityRedispatchesWhenContentChanges(t *testing.T) {
	// round 1 passes → second call with one character changed in the plan
	// assert: the fake judge was invoked TWICE, ledger has 2 rounds
}

// A short-circuit after a BLOCKING round would let a refused plan through unchanged.
func TestPlanQualityNoPassThroughAfterBlockingRound(t *testing.T) {
	// round 1 blocks → second call, content unchanged
	// assert: dispatched again (never cached), still blocks
}
```

**Step 4b — the main path:**

```go
func runPlanQualityJudge(stdout, stderr io.Writer, f *changeCodeFlags, name, issuePath, issueContent, planContent string) error {
	ledger, err := readPlanGateLedger(f.PlansDir, filepath.Base(issuePath), f.Issue)
	if err != nil {
		// A corrupt ledger must halt, not silently forget (see planreview.go).
		return fmt.Errorf("plan-gate ledger: %v", err)
	}
	// Pass-through: unchanged content after a passing round ⇒ no dispatch, no round.
	if hash := gatestate.ContentHash(issueContent, planContent); gatestate.PassesUnchanged(ledger, hash) {
		cok(stderr, fmt.Sprintf("plan-quality: unchanged since round %d — passing through (cached)", len(ledger.Rounds)))
		return nil
	}
	prompt := judge.BuildPrompt(judge.PlanQuality, judge.PromptInput{
		IssueRef: issueRef, IssueContent: issueContent, PlanContent: planContent,
		PriorFindings: gatestate.RenderPriorFindings(ledger),
	})
	// … dispatch as today …

	n := len(ledger.Rounds) + 1
	rr, ok := gatestate.ParseFindingsBlock(output)
	if !ok {
		// Transitional fallback (agent-binary-handoff-schema: the schema'd path is
		// authoritative; a prose fallback may exist transitionally). Warn loudly — this
		// round contributes no FINDINGS, so the gate cannot converge on it.
		//
		// It DOES still contribute a round. Persisting a findings-less round is
		// load-bearing, not bookkeeping: if a protocol miss returned early, `len(Rounds)`
		// would stay 0 forever for a judge whose CLI never emits the fence, so
		// RenderPriorFindings would announce "this is the FIRST round" on invocation six,
		// Decide's round cap could never fire to bound the loop it exists to bound, and
		// M2's `gate_rounds` would report 0 for precisely the most expensive sessions —
		// inverting the signal the D-series exists to produce.
		cwarn(stderr, "plan-quality: no valid ```findings block — falling back to the verdict token; this round carries NO findings, so the gate cannot converge on it")
		blocked := classifyFallback(stderr, output) != nil
		persistRound(stderr, f, issuePath, ledger, gatestate.Round{
			N: n, Timestamp: nowRFC3339(), Agent: agent,
			ProtocolError: "no valid findings block", Blocked: blocked,
			Forced: forcedRationale(f.Force, blocked),
		})
		if blocked {
			return fmt.Errorf("plan-quality failure")
		}
		return nil
	}

	round := gatestate.AssignIDs(ledger, rr, n, nowRFC3339(), agent)
	ledger, aerr := gatestate.ApplyChecked(ledger, round)
	if aerr != nil {
		// Same reasoning: a protocol error is a round that happened and cost latency.
		cwarn(stderr, "plan-quality: "+aerr.Error())
		persistRound(stderr, f, issuePath, ledger, gatestate.Round{
			N: n, Timestamp: nowRFC3339(), Agent: agent,
			ProtocolError: aerr.Error(), Blocked: true, Forced: forcedRationale(f.Force, true),
		})
		return fmt.Errorf("plan-quality protocol error: %v", aerr)
	}
	d := gatestate.Decide(ledger, roundCapFromEnv())
	ledger.Rounds[len(ledger.Rounds)-1].Blocked = d.Block
	// Stamp the content hash only on a PASSING round — that is what the pass-through
	// checks. Stamping on a blocking round would cache away a refusal.
	if !d.Block {
		ledger.ContentHash = gatestate.ContentHash(issueContent, planContent)
	}
	// Record --force ONLY when this gate actually blocked. --force is a GLOBAL bypass
	// (changecode.go:124/144/159 all consult it), so stamping it unconditionally would
	// mark a plan-gate round "forced" when the operator forced past a STRUCTURAL failure,
	// or even when Decide passed cleanly — over-reporting overrides in the one number
	// whose entire purpose is to answer "which gates earn their cost" trustworthily.
	ledger.Rounds[len(ledger.Rounds)-1].Forced = forcedRationale(f.Force, d.Block)
	if werr := writePlanGateLedger(f.PlansDir, filepath.Base(issuePath), ledger, repoIdentity()); werr != nil {
		cwarn(stderr, fmt.Sprintf("plan-gate ledger not persisted: %v", werr))
	}
	if d.Block {
		cwarn(stderr, "plan-quality: "+d.Reason)
		cwarn(stderr, "address the findings above and re-run; the gate remembers what you fixed")
		return fmt.Errorf("plan-quality failure")
	}
	cok(stderr, "plan-quality: "+d.Reason)
	return nil
}

// forcedRationale returns the --force rationale only when this gate actually refused;
// "" otherwise. See the comment at its call site for why the distinction matters.
func forcedRationale(force string, blocked bool) string {
	if blocked {
		return force
	}
	return ""
}

// classifyFallback is the pre-#187 tri-state read of the judge's output — the
// judge.Classify switch EXTRACTED verbatim from runPlanQualityJudge (changecode.go:380-392)
// so the schema'd path and the transitional prose path don't each carry a copy (ARCH-DRY).
// Returns nil on Clean/Info, an error on Failure.
func classifyFallback(stderr io.Writer, output string) error { … }

// persistRound appends one round and writes the ledger, warning (never failing) on a
// write error. The single persistence call site for all three exit paths.
func persistRound(stderr io.Writer, f *changeCodeFlags, issuePath string, l gatestate.Ledger, r gatestate.Round) { … }

// roundCapFromEnv reads WF_PLAN_ROUND_CAP, defaulting to gatestate.DefaultRoundCap.
func roundCapFromEnv() int { … }
```

> `runChangeCode` must now pass `issuePath` into `runPlanQualityJudge` (it already has it
> from `resolveChangeCodeName`).

- [ ] **Step 5: Run**

Run: `go test ./cmd/sdlc/... -v`
Expected: PASS.

- [ ] **Step 6: Prompt-shape smoke test** (cheap, no agent):

```bash
go build -o /tmp/sdlc ./cmd/sdlc
/tmp/sdlc change-code --issue 187 --dry-run
```
Confirm the printed prompt contains the ` ```findings ` block and a prior-findings section.

> `--dry-run` returns at `changecode.go:357-367` **before** `judge.Dispatch` — no agent
> runs, no ledger is written. This step proves the prompt renders, nothing more. The
> real-agent, two-round convergence proof is Task 14, which is built for it.

- [ ] **Step 7: Commit**

```bash
git add cmd/sdlc/changecode.go cmd/sdlc/changecode_test.go
git commit -m "#187 M1: change-code — stateful plan gate, estimate after plan acceptance

The gate reads its own prior findings, blocks only on undisposed Critical/Important,
and records every round (including forced ones) durably. The estimate gates move
below plan-quality: costing a plan nobody has accepted yet is waste by construction.

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>"
```

---

### Task 9: Documentation — say the same thing in all five places (B2)

**Files:**
- Modify: `cmd/sdlc/helptext/change-code.md`
- Modify: `cmd/sdlc/helptext/start-plan.md`
- Modify: `cmd/sdlc/helptext/estimate.md`
- Modify: `cmd/sdlc/startplan.go` (the estimate nudge)
- Modify: `cmd/sdlc/changecode.go:147` — the **sixth** surface, and the one an agent
  actually reads at the moment the gate fires. It currently prints "…(set it at
  start-plan)…"; after B1 that message fires *after* plan-quality, making its own advice
  wrong exactly when it is displayed.
- Modify: `AGENTS.base.md:24` (§2 "Claim early" bullet)
- Modify: `cmd/sdlc/helptext/issue.md:35` — "estimate_hours set at start-plan; required by
  change-code — not at claim (#113)"
- Modify: `cmd/sdlc/helptext/set-status.md:15` — "Claim/start work early; estimate at start-plan."
- Modify: `atlas/workflow/issue-lifecycle.md:35` + `:93` — the "Plan" lifecycle step and the
  frontmatter template comment. **`construct/base.manifest:214` is `symlink atlas/workflow`,
  so these propagate to every downstream repo** — weigh accordingly.
- Modify: `atlas/workflow/sdlc-binary.md:35` — the gate-order table row still reads
  `structural + estimate (#113) + estimate-reconciliation + estimate-quality (#117) +
  plan-quality`, verbatim the order B1 inverts.
- Modify: `cmd/sdlc/startplan_test.go:116` — `TestEstimateNudge` pins the nudge text Step 3
  rewrites; it will fail until updated.
- Modify: `atlas/` — a new `atlas/workflow/gate-state.md`
- Modify: `atlas/index.md` (link the new file)

> **How this list was built (do not hand-extend it):** the round-1 revision's literal sweep
> found only the two strings it already knew about. A **semantic** sweep —
> `rg -i 'estimate.{0,80}start-plan|start-plan.{0,80}estimate' --glob '!workshop/**'` —
> surfaces all of the above. `cmd/sdlc/helptext/estimate.md` and
> `helptext/estimate-source.md` appear in that sweep but carry **no timing claim** (grammar
> and a pointer respectively); audit and leave them. Re-run the semantic sweep before
> declaring Step 4 done — the surface set is derived, not enumerated.

- [ ] **Step 1: Write the consistency guard test** first — the five surfaces must agree, and
  a prose contract needs a test that pins the *semantic* claim, not a token
  (`workshop/lessons.md`, #167):

```go
// TestEstimateTimingConsistency pins B2: every surface that tells an agent WHEN to derive
// the estimate must say "after the plan clears plan-quality", not "at start-plan". A
// prose policy with five consumers drifts unless a test reads all five (#167 lesson).
func TestEstimateTimingConsistency(t *testing.T) {
	surfaces := map[string]string{
		"helptext/change-code.md": helptext.MustGet("change-code"),
		"helptext/start-plan.md":  helptext.MustGet("start-plan"),
		"helptext/estimate.md":    helptext.MustGet("estimate"),
		"AGENTS.base.md":          mustReadRepoFile(t, "AGENTS.base.md"),
		// The two Go sources that PRINT the advice. A test reading only the prose files
		// would miss changecode.go:147 — the very line an agent sees when the gate fires.
		"changecode.go": mustReadRepoFile(t, "cmd/sdlc/changecode.go"),
		"startplan.go":  mustReadRepoFile(t, "cmd/sdlc/startplan.go"),
	}
	// The guard IS the semantic sweep — not a list of literals, which is how the round-1
	// version passed clean while five live surfaces still told the old story. Any
	// co-occurrence of "estimate" and "start-plan" within 80 chars must be on the
	// allowlist of surfaces verified to carry no timing claim.
	allowed := map[string]bool{
		"cmd/sdlc/helptext/estimate.md":        true, // block grammar only
		"cmd/sdlc/helptext/estimate-source.md": true, // a pointer, no timing claim
	}
	re := regexp.MustCompile(`(?i)estimate.{0,80}start-plan|start-plan.{0,80}estimate`)
	for _, path := range repoFilesExcluding(t, "workshop/") {
		if allowed[path] { continue }
		for i, line := range strings.Split(mustReadRepoFile(t, path), "\n") {
			if re.MatchString(line) && !strings.Contains(line, "after the plan clears plan-quality") {
				t.Errorf("%s:%d still ties the estimate to start-plan (#187 B2): %s", path, i+1, line)
			}
		}
	}
	// The positive claim, scoped to the section that owns it.
	for _, s := range []string{"helptext/change-code.md", "AGENTS.base.md"} {
		if !strings.Contains(surfaces[s], "after the plan clears plan-quality") {
			t.Errorf("%s must state the new timing explicitly", s)
		}
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./cmd/sdlc/ -run TestEstimateTimingConsistency -v`
Expected: FAIL on all four surfaces.

- [ ] **Step 3: Edit the four prose surfaces + `startplan.go`.**

`AGENTS.base.md` §2, final sentence of the "Claim early" bullet — replace:
> The estimate is set later, at `start-plan` (required by `change-code`).

with:
> The estimate is derived later still — **after** the plan clears `change-code`'s
> plan-quality gate (#187), when the design is settled and the scope is actually knowable.
> `change-code` runs plan-quality first and only then asks for `estimate_hours` + the
> `## Estimate` block, so an unapproved plan is never costed.

`helptext/change-code.md` — renumber the gate list to the new order (structural →
plan-quality → estimate trio → branching), and add to the plan-quality entry:

> The gate is **stateful** (#187): its findings persist to
> `workshop/plans/NNNNNN-slug-plan-gate.md` with stable IDs, and each re-run must
> dispose of every prior finding before raising new ones. Only *undisposed* Critical or
> Important findings block; new Minor findings are recorded for the close review and cost
> no round-trip. Past `WF_PLAN_ROUND_CAP` rounds (default 3) only Critical blocks.

`helptext/start-plan.md` OUTPUT section — replace "this is where you set the estimate" with
a pointer to the new timing.

`startplan.go`'s estimate nudge — same change (it currently prints "Set `estimate_hours:`
… before `sdlc change-code`"). New text: derive it *when change-code asks*, i.e. after
plan-quality accepts.

`helptext/estimate.md` — audit for the same claim; update the timing sentence.

`cmd/sdlc/changecode.go:147` — replace `(set it at start-plan)` with
`(derive it now — the plan has cleared plan-quality)`.

- [ ] **Step 4: Run**

Run: `go test ./cmd/sdlc/...`
Expected: PASS.

Then re-run the **semantic** sweep that produced the surface list, to confirm nothing was
missed (a literal sweep for the two known strings cannot find what it doesn't already know):

Run: `rg -n -i 'estimate.{0,80}start-plan|start-plan.{0,80}estimate' --glob '!workshop/**'`
Expected: only `helptext/estimate.md` and `helptext/estimate-source.md` (the two audited
no-timing-claim surfaces), plus lines that now read "after the plan clears plan-quality".

> The `workshop/**` exclusion is required, not incidental: this plan quotes the target
> strings and the issue quotes the `AGENTS.base.md` sentence in its Spec. A sweep that
> includes `workshop/` can never return clean, so it would be unfalsifiable.

- [ ] **Step 5: Wire the demotion's promised consumer — `code-review.md` must actually
  read the plan-gate sidecar (`ARCH-PURPOSE`).**

  The whole safety argument for not blocking on Minor and post-cap Important findings is
  that *"they land in the sidecar for the close review to pick up"* (Spec A3) — a claim
  this plan repeats in `decide.go`'s doc comment, Task 3's policy note, and Task 9's
  helptext. **That consumer does not exist today.**
  `boundaryReviewDispatchOptions` (`milestoneclose.go:582`; its `judge.PromptInput` literal
  at `:596-600`) builds `PromptInput{Diff, Base, Head, IssueRef, Repo, RepoRoot, IssueFile,
  Boundary, RepoNote}` — no plan-gate
  content — and `code-review.md` contains no mention of a plan gate (verified: `grep -i
  'plan-gate\|plan gate\|plan-quality'` returns nothing). Shipping the demotion without
  this is the exact ARCH-PURPOSE shape the registry warns about: the cheap half (stop
  blocking) ships, and the half that makes it *safe* stays documentation that doesn't
  derive.

  The boundary reviewer already has `Read` (`AllowedTools()` returns `"Read,Grep,Glob,Bash"`),
  so it needs a pointer, not new plumbing. Add to `code-review.md`, in the review-checklist
  section:

  ```markdown
  Plan-gate carry-forward
    - Read `workshop/plans/<issue-stem>-plan-gate.md` if it exists. It holds the findings
      the pre-implementation plan gate raised but did NOT block on — Minor findings, and
      Important ones demoted past the round cap. They were deferred to THIS boundary by
      design. For each still-open finding, confirm the code either addresses it or that it
      no longer applies; a still-valid deferred finding is a finding here.
  ```

  And a guard test in `judge_test.go`, so the pointer cannot silently disappear:

```go
// TestCodeReviewCarriesPlanGateForward pins #187 A3's safety argument: the plan gate stops
// blocking on Minor/demoted-Important findings ONLY because the close review picks them
// up. If this pointer is ever dropped, that trade becomes a silent loss.
func TestCodeReviewCarriesPlanGateForward(t *testing.T) {
	body := CodeReviewBody(PromptInput{})
	if !strings.Contains(body, "plan-gate.md") {
		t.Error("code-review.md must point the boundary reviewer at the plan-gate sidecar")
	}
}
```

- [ ] **Step 6: Update the atlas** (do not defer to an end-of-project sweep, AGENTS.md §8).
  Add a `atlas/workflow/gate-state.md` describing: what a gate ledger is, where it lives,
  the finding vocabulary, the convergence policy, the plan-gate → close-review
  carry-forward from Step 5, and that #183 is the second intended consumer. Link it from
  `atlas/index.md`, and update the change-code row in the existing workflow map with the
  new gate order.

- [ ] **Step 7: Commit**

```bash
git add cmd/sdlc/helptext/ cmd/sdlc/startplan.go cmd/sdlc/changecode.go AGENTS.base.md \
        cmd/sdlc/internal/judge/code-review.md cmd/sdlc/internal/judge/judge_test.go \
        atlas/ cmd/sdlc/*_test.go
git commit -m "#187 M1: one story about estimate timing across all five surfaces

A prose policy with five consumers drifts unless a test reads all five — the #167
lesson, applied. The guard pins the semantic claim, not a token.

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>"
```

---

### Task 10: Close M1

- [ ] **Step 1:** `go test ./... && go vet ./...` — all green.
- [ ] **Step 2:** Tick M1's Plan rows in `workshop/issues/000187-tune-change-code-gate.md`
      and write the `## Log` entry (round counts observed, ARCH markers cited).
- [ ] **Step 3:** `sdlc milestone-close --issue 187 --milestone M1` — the binary
      auto-dispatches the one mandatory fresh-eyes review (do **not** separately run
      `superpowers-requesting-code-review`). Fix Critical/Important before crossing.

---

## Chunk 2 — Milestone M2: the gate's cost becomes measurable (D)

### Task 11: `churn` — pure bucketing

**Files:**
- Create: `cmd/sdlc/internal/churn/classify.go`
- Create: `cmd/sdlc/internal/churn/classify_test.go`
- Create: `cmd/sdlc/internal/churn/report.go`
- Create: `cmd/sdlc/internal/churn/report_test.go`

- [x] **Step 1: Write the failing classification test.** Pin the rules *and* the
  judgment calls, so a later reader sees the reasoning.

  **The rule, in order:** `atlas/` → atlas; `workshop/` → workshop; a test file
  (`*_test.go`, a `testdata/` segment) → code-test; **everything else → code-prod**, which
  is the stated DEFAULT rather than a fallthrough nobody chose. Embedded markdown counts as
  production here because it ships inside the binary via `//go:embed`; so do build and
  config files, which are versioned, reviewed, and break the build when wrong.

```go
func TestClassifyPath(t *testing.T) {
	cases := map[string]Bucket{
		"cmd/sdlc/changecode.go":               CodeProd,
		"cmd/sdlc/changecode_test.go":          CodeTest,
		"cmd/sdlc/internal/judge/testdata/x.md": CodeTest,
		"pkg/vocab/finding.go":                 CodeProd,
		"atlas/index.md":                       Atlas,
		"atlas/workflow/gate-state.md":         Atlas,
		"workshop/issues/000187-x.md":          Workshop,
		"workshop/plans/000187-x-plan.md":      Workshop,
		// Embedded prompt/helptext markdown is PRODUCTION here — it ships inside the
		// binary via //go:embed and is exactly the surface #187 changes. Counting it as
		// prose would understate the code this repo actually writes.
		"cmd/sdlc/internal/judge/prompts/plan-quality.md": CodeProd,
		"cmd/sdlc/helptext/change-code.md":                CodeProd,
		"construct/vocabulary/finding.cue":                CodeProd,
		"AGENTS.base.md":                                  CodeProd,
		// The DEFAULT bucket, named explicitly rather than left to whichever switch arm
		// happens to be last. Build/config/meta files are production artifacts of the
		// repo: they are versioned, reviewed, and break the build when wrong. Routing
		// them to code-prod is a decision, and a lockfile-sized diff landing there must
		// be a visible choice rather than an accident.
		"go.mod":                    CodeProd,
		"go.sum":                    CodeProd,
		"Makefile.workflow":         CodeProd,
		".github/workflows/ci.yml":  CodeProd,
		"construct/base.manifest":   CodeProd,
		"docs/vision/roadmap.md":    CodeProd,
	}
	for path, want := range cases {
		if got := ClassifyPath(path); got != want {
			t.Errorf("ClassifyPath(%q) = %v, want %v", path, got, want)
		}
	}
}
```

- [x] **Step 2: Write the failing report test**

```go
func TestSummarize(t *testing.T) {
	final := []FileStat{
		{Path: "cmd/sdlc/changecode.go", Insertions: 100},
		{Path: "cmd/sdlc/changecode_test.go", Insertions: 200},
		{Path: "atlas/index.md", Insertions: 5},
		{Path: "workshop/issues/000187-x.md", Insertions: 60},
	}
	// The window's commits inserted 1095 lines total to land those 365.
	r := Summarize(final, 1095)
	if r.Final.CodeProd != 100 || r.Final.CodeTest != 200 || r.Final.Atlas != 5 || r.Final.Workshop != 60 {
		t.Errorf("buckets = %+v", r.Final)
	}
	if r.FinalTotal != 365 { t.Errorf("FinalTotal = %d, want 365", r.FinalTotal) }
	if math.Abs(r.Rework-3.0) > 0.01 { t.Errorf("Rework = %.2f, want 3.00", r.Rework) }
}

// Rework is undefined, not +Inf, when a window lands no insertions (a pure-deletion or
// empty window) — a NaN/Inf in the TSV would poison every downstream reader.
func TestSummarizeZeroFinal(t *testing.T) {
	if r := Summarize(nil, 40); r.Rework != 0 {
		t.Errorf("Rework = %v, want 0 for an empty final diff", r.Rework)
	}
}

// Binary files show "-" in numstat; they must not abort the sum.
func TestSummarizeSkipsBinary(t *testing.T) { … }
```

- [x] **Step 3: Run to verify they fail; Step 4: implement; Step 5: run to verify PASS.**

Run: `go test ./cmd/sdlc/internal/churn/ -v`

- [x] **Step 6: Commit**

```bash
git add cmd/sdlc/internal/churn/
git commit -m "#187 M2: churn — four-bucket classification + rework ratio

Rework is sum-of-commit-insertions over final-insertions, not the final diff:
final-diff alone scored pair#127 as merely process-heavy and missed the actual
waste, which was rewriting one file five times.

No comment-vs-non-comment split (deliberately deferred, #187 D1): this house
style is comment-dense by design, so that ratio is descriptive at best and a
Goodhart target at worst.

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>"
```

---

### Task 12: The git seam — churn over the close window

**Files:**
- Create: `cmd/sdlc/churnreport.go`
- Create: `cmd/sdlc/churnreport_test.go`

- [x] **Step 1: Write the failing test** against a real temp git repo (the pattern
  `closereview_test.go` already establishes — `git` is exercised for real in a disposable
  repo rather than mocked, per `ARCH-MOCK`):

```go
// Three commits that rewrite the same file land 30 final lines from 90 inserted —
// rework 3.0. This is the pair#127 shape the metric exists to catch.
func TestChurnForWindowCountsRework(t *testing.T) {
	// closeRepo(t, issueNum) (closereview_test.go:24) builds + chdir's into a temp git
	// repo with an issue file already committed; it is the established helper for
	// git-touching tests in this package. Reuse it and add a small commitFile helper
	// beside it rather than standing up a second repo fixture (ARCH-DRY).
	closeRepo(t, 187)
	base := commitFile(t, "cmd/x.go", strings.Repeat("a\n", 30), "#187: v1")
	commitFile(t, "cmd/x.go", strings.Repeat("b\n", 30), "#187: v2")
	commitFile(t, "cmd/x.go", strings.Repeat("c\n", 30), "#187: v3")

	r, err := churnForWindow(base)
	if err != nil { t.Fatal(err) }
	if r.Final.CodeProd != 30 { t.Errorf("CodeProd = %d, want 30", r.Final.CodeProd) }
	if math.Abs(r.Rework-3.0) > 0.05 { t.Errorf("Rework = %.2f, want ~3.0", r.Rework) }
}

// An unresolvable base (a docs-only window with no #N commit — resolveReviewWindow
// returns "") must degrade to a zero report, never break the close.
func TestChurnForWindowEmptyBase(t *testing.T) {
	r, err := churnForWindow("")
	if err != nil { t.Fatalf("empty base must not error: %v", err) }
	if r.FinalTotal != 0 { t.Errorf("want a zero report, got %+v", r) }
}
```

- [x] **Step 2–4: run-fail, implement, run-pass.** `churnForWindow(baseLong string)`:
  - `""` base → zero `churn.Report`, nil error.
  - final: `gitx.RunGit("diff", "--numstat", base+"..HEAD")` → `[]churn.FileStat`.
  - commit total: `gitx.RunGit("log", "--numstat", "--format=", base+"..HEAD")` summed.

  **Use `RunGit`, not `Capture`.** `gitx.Capture` returns `""` on ANY error
  (`internal/gitx/window.go:50-56`) and its own doc says it is "not suitable for queries
  where you must distinguish 'ran but empty' from 'errored'" (`:47-49`). Task 13's contract
  is "a git failure warns and leaves the values at zero" — with `Capture` that warning can
  never fire, because a bad base SHA prints `churn: prod 0 / test 0 / …` identically to a
  genuinely empty window, in the one number introduced to answer "which gates earn their
  cost". `gitx.RunGit` (`window.go:38-40`) is the exported error-returning variant and
  exists for exactly this (`ARCH-DRY` — reuse the seam that already distinguishes them,
  don't degrade silently through the flattening one). `churnForWindow` returns
  `(churn.Report, error)`; the CALLER warns and zeroes.
  - `churn.Summarize(final, commitTotal)`.
  - Reuse `boundaryWindowBase` for the base — **do not** re-derive the window
    (`ARCH-DRY`; it keeps churn provably covering the same commits as the review and the
    atlas gate).

- [x] **Step 5: Commit.**

---

### Task 13: Emit the metrics at close (D1/D2/D3)

**Files:**
- Modify: `cmd/sdlc/internal/estimate/ledger.go` (10 appended columns)
- Modify: `cmd/sdlc/internal/estimate/ledger_test.go`
- Modify: `cmd/sdlc/close.go` (`appendCalibrationRow`, the info line)
- Modify: `cmd/sdlc/close_ledger_test.go`

- [x] **Step 1: Write the failing ledger tests.** Columns are **appended**, never
  reordered — `ParseRows` indexes positionally (`c[0]`…`c[9]`) and
  `projectthroughput_test.go` seeds fixture rows, so an insertion would silently
  mis-read history:

```go
// Appending columns must not break the reader for PRE-EXISTING 10-column rows.
func TestParseRowsAcceptsLegacyTenColumnRows(t *testing.T) {
	legacy := Header() + "\nariadne#1\t2\t1\t1\t3\t0.67\tm\t-\tyes\t2026-01-01\n"
	if rows := ParseRows(legacy); len(rows) != 1 || rows[0].Issue != "ariadne#1" {
		t.Errorf("legacy row lost: %+v", rows)
	}
}

func TestFormatRowCarriesChurnColumns(t *testing.T) {
	r := LedgerRow{Issue: "ariadne#187", Estimate: 4, Actual: 5,
		ChurnProd: 554, ChurnTest: 300, ChurnAtlas: 20, ChurnWorkshop: 778,
		Rework: 2.4, GateRounds: 6, GateForced: 1,
		GateAddressed: 2, GateWithdrawn: 1, GateOpen: 3}
	cols := strings.Split(FormatRow(r), "\t")
	if len(cols) != len(strings.Split(Header(), "\t")) {
		t.Fatalf("row has %d columns, header has %d", len(cols), len(strings.Split(Header(), "\t")))
	}
	if cols[10] != "554" { t.Errorf("churn_prod at col 10 = %q", cols[10]) }
	if cols[16] != "1" { t.Errorf("gate_forced at col 16 = %q", cols[16]) }
	if cols[17] != "2" { t.Errorf("gate_addressed at col 17 = %q", cols[17]) }
	if cols[19] != "3" { t.Errorf("gate_open at col 19 = %q", cols[19]) }
}
```

- [x] **Step 2: Run-fail; Step 3: implement.** Append to `LedgerRow` and to
  `ledgerHeader`, in this order:
  `churn_prod  churn_test  churn_atlas  churn_workshop  rework  gate_rounds  gate_forced
  gate_addressed  gate_withdrawn  gate_open` — **ten** appended columns (see the
  disambiguation note in Step 4 for why the last three exist). `ledgerHeader`
  (`internal/estimate/ledger.go:33`) has 10 columns today (indices 0–9), so the new ones
  occupy **10–19** and the row total is **20**: churn_prod=10, churn_test=11,
  churn_atlas=12, churn_workshop=13, rework=14, gate_rounds=15, gate_forced=16,
  gate_addressed=17, gate_withdrawn=18, gate_open=19. Extend `ParseRows` to read them
  **only when present** (`len(c) >= 20` — `>= 19` would admit a 19-column row and then
  panic reading `c[19]`), so legacy 10-column rows keep parsing.

- [x] **Step 4: Wire `close.go` — TWO consumers with DIFFERENT contracts.** Do not fold
  them together; that is the trap here.

  **(a) The operator-facing line is unconditional.** Emit it in `applyClose`, beside the
  `"done — review with `+"`git diff`"+`"` line at `close.go:766` — **not** inside
  `appendCalibrationRow`. That function is reached only when `shouldLogCalibration(f)`
  passes (`close.go:763`: `f.Milestone == "" && f.Actual != "" && !IsActualNotApplicable`)
  and then returns early at `close.go:808` ("no brain dir resolved") and `:816` ("no ledger
  dir"). Putting the print there would mean **no churn output** on a milestone close, under
  `--no-actual`, or in any downstream repo without a sibling `brain/` — against a Done-when
  that says plainly "`sdlc close` prints the four-bucket churn split, rework ratio,
  round-trip count, and finding disposition." The line is diagnostic and inherits none of
  those three conditions.

```go
	// #187 D1–D3: the cost report. Unconditional — unlike the ledger row below, which is
	// gated by #117's calibration-integrity rule (milestone closes carry a partial actual
	// and must not pollute the ledger), this is diagnostic output the operator always gets.
	ch, rounds, forced := closeMetrics(stderr, f, res)
	cok(stderr, fmt.Sprintf("churn: prod %d / test %d / atlas %d / workshop %d (final %d, rework %.1f×)",
		ch.Final.CodeProd, ch.Final.CodeTest, ch.Final.Atlas, ch.Final.Workshop, ch.FinalTotal, ch.Rework))
	cok(stderr, fmt.Sprintf("plan gate: %d round(s), %d forced", rounds, forced))
```

  **(b) The ledger columns stay inside `appendCalibrationRow`**, correctly gated — pass the
  same values in rather than recomputing (`ARCH-DRY`).

  **Disambiguation — "disposition" names two different things here.** Spec D3 glosses it
  as *accepted-vs-forced*; this plan's `Disposition` type means *addressed |
  not-addressed | withdrawn*. Done-when says close prints "finding disposition", which
  reads as the latter — and the latter is the number that actually answers *"did the
  gate's findings get acted on, or worked around?"* Both are free from the ledger, so
  **emit both** and end the ambiguity rather than picking:

```go
	cok(stderr, fmt.Sprintf("plan gate: %d round(s), %d forced; findings %d addressed / %d withdrawn / %d still open",
		rounds, forced, addressed, withdrawn, stillOpen))
```

  Ledger columns become TEN, not seven — append
  `gate_addressed  gate_withdrawn  gate_open` after `gate_forced` — **ten** appended
  columns in total (Step 3 enumerates them), so the appended block is
  `cols[10]`…`cols[19]` (20 columns total) and the `ParseRows` presence check reads
  `len(c) >= 20`. Counts come from a pure
  `gatestate.DispositionCounts(l Ledger) (addressed, withdrawn, open int)` — unit-tested
  beside `OpenFindings`, which already computes the closed-set it needs.

  **Wiring gap to close first:** `closeMetrics` needs an issue path and a plans dir, and
  **neither is currently reachable** from `appendCalibrationRow`'s signature
  (`stderr, f, fm, body, repoName, issueStr, today`). `closeFlags` (`close.go:52-77`) has
  `BrainDir` and `IssuesDir` but **no `PlansDir`**. Sources to use:
  - issue path → `res.issuePath`, already on `closeResult` (`close.go:329`);
  - plans dir → add a `PlansDir` field to `closeFlags` with
    `envOr("WF_PLANS_DIR", "workshop/plans")`, matching how `close.go:959`/`:992` already
    resolve it inline for the boundary review — and switch those two inline calls to the
    new field so there is one source (`ARCH-DRY`).

  **Degrade, never break** — the `appendCalibrationRow` precedent (a missing ledger must
  never break `sdlc close`) applies to churn and gate state too: a git failure or an absent
  plan-gate sidecar warns and leaves the values at zero. `closeMetrics` returns zeroes and
  warns; it never returns an error.

- [x] **Step 5: Write the failing close-side tests** — the unconditional contract is the
  point, so test the three cases that would silently lose it:

```go
func TestCloseWithoutPlanGateSidecarStillCloses(t *testing.T) { … }

// The churn line must print on a MILESTONE close, under --no-actual, and with no brain
// dir — the three conditions that gate the ledger row but must not gate the report.
func TestChurnLinePrintsWhenLedgerRowIsSkipped(t *testing.T) {
	for _, tc := range []struct{ name string; f closeFlags }{
		{"milestone", closeFlags{Milestone: "M1", Actual: "1.0"}},
		{"no-actual", closeFlags{NoActual: true}},
		{"no-brain",  closeFlags{Actual: "1.0", BrainDir: ""}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// … applyClose into a buffer …
			if !strings.Contains(stderr.String(), "churn: prod ") {
				t.Error("churn line must print even when the calibration row is skipped")
			}
		})
	}
}
```

- [x] **Step 6: Run**

Run: `go test ./cmd/sdlc/... ./cmd/sdlc/internal/... -v`
Expected: PASS, including `projectthroughput_test.go`'s ledger fixture.

- [x] **Step 7: Update the atlas** — the calibration-ledger schema row gains the seven
  columns; `atlas/workflow/gate-state.md` gains the "what close reports" section.

- [x] **Step 8: Commit**

```bash
git add cmd/sdlc/internal/estimate/ cmd/sdlc/close.go cmd/sdlc/*_test.go atlas/
git commit -m "#187 M2: close reports churn, rework, and gate round-trips

Columns are APPENDED, never reordered — ParseRows indexes positionally and
existing ledgers must keep reading. Every new metric degrades to zero with a
warning rather than breaking a close.

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>"
```

---

### Task 14: The regression test for "did we weaken review?"

This is the Done-when criterion that protects the gate's *value*. It cannot be a unit test
— it needs a real judge — and it must run through **`sdlc change-code`**, not `sdlc judge`.

> **Why not `sdlc judge plan-quality`:** that verb cannot do either half of this job.
> `cmd/sdlc/judge.go:109-113` builds `judge.PromptInput` with `Diff`, `ChangedIssues`,
> `Base`, `Head` only — `IssueContent` and `PlanContent` are **never populated for any
> category**, so `{{ISSUE_CONTENT}}` renders empty and `{{PLAN_CONTENT}}` renders
> `(no separate plan file)`. The judge would review pair#127's plan without seeing it.
> And `sdlc judge` holds no ledger (Task 8 wires the stateful path in `changecode.go`
> only), so "rounds to acceptance" is uncountable through it. **File that as its own
> issue** (`sdlc issue new`: "`sdlc judge plan-quality` renders an empty issue — no
> `--issue` path populates IssueContent/PlanContent") — it is a real defect in a verb this
> repo's helptext advertises, but it is not #187's purpose and must not be smuggled in.

**Files:**
- Create: `cmd/sdlc/planreplay_test.go` (the harness; build-tagged manual)
- Create: `workshop/plans/000187-replay-pair127.md` (the evidence record)

- [x] **Step 1: Recover pair#127's round-1 state from the ISSUE file — there is no plan
      doc.** `pair/workshop/history/` holds exactly two #127 artifacts:
      `issues/000127-term-pane-stream-corruption.md` and
      `plans/000127-term-pane-stream-corruption-close-review.md`. The close-review sidecar
      archived fine, so the absent plan doc is real rather than an archiving miss — #127's
      Plan lived **in the issue file** (`## Plan` at `:184`, with the prose test-case
      bullets #187's Problem section cites at `:346`).

      So: `git log -p --follow workshop/issues/000127-*.md` in the pair repo; take the
      revision as it stood at the FIRST `change-code` invocation, before any gate feedback
      landed. Seed `writeIssueFile` from that and **drop `writePlanFile`** — the replay's
      round 1 sees `PlanContent` empty, which is exactly what the original round 1 saw.
      State that in the evidence record so it reads as a controlled condition, not a gap.

      #187's Problem section names what round 1 must find: the filter seam relocation and
      the "defense in depth" absorb layer that would have swallowed solicited terminal
      replies.

- [x] **Step 2: Build the replay harness.** A scratch repo driven through the real verb —
      the same shape as the existing git-touching tests, so the ledger, the round counter
      and the convergence policy are all exercised for real:

```go
//go:build manual

// TestReplayPair127 is the regression test for "did the tuning weaken review?" (#187
// Done-when). Manual-tagged: it spends real agent latency, so CI does not run it.
//   go test ./cmd/sdlc -tags manual -run TestReplayPair127 -v -timeout 30m
//
// It drives runPlanQualityJudge DIRECTLY, not runChangeCode. runChangeCode iterates
// changeCodeGates and calls exitWithCode(1) on any gate failure (changecode.go:130-139 →
// term.go:56-59 → os.Exit), so it never RETURNS a gate failure — a `err := runChangeCode(…)`
// loop would os.Exit(1) the test process on round 1, which is precisely the round whose
// blocking is the point. (expectDie doesn't help: it swaps the `die` var, which this path
// bypasses.) runPlanQualityJudge owns the entire ledger path — read → dispatch → parse →
// assign ids → Decide → write — and returns an error, so rounds stay countable in-process.
func TestReplayPair127(t *testing.T) {
	dir := t.TempDir()
	f := &changeCodeFlags{Issue: 900, PlansDir: dir}
	issue := mustReadTestdata(t, "pair127-round1-issue.md")

	// Round 1 against the recovered plan. Expect a refusal.
	err := runPlanQualityJudge(os.Stdout, os.Stderr, f, "000900-replay", "000900-replay.md", issue, "")
	t.Logf("round 1: err=%v", err)

	l, lerr := readPlanGateLedger(dir, "000900-replay.md", 900)
	if lerr != nil { t.Fatalf("ledger: %v", lerr) }
	for _, fd := range gatestate.OpenFindings(l) {
		t.Logf("round 1 finding [%s] %s: %s", fd.Severity, fd.ID, fd.Title)
	}
	// Step 3 authors the between-rounds edit BY HAND from these findings, then re-runs
	// this test with REPLAY_PLAN pointing at the edited plan. The harness deliberately
	// does NOT auto-edit: faking the author's response would prove nothing about
	// convergence.
}
```

> **Isolation note.** Because this drives the plan gate directly, the estimate gates are
> not in the loop at all — no `--no-estimate` needed. Do NOT "seed a reconciling
> `## Estimate` block" instead: `runEstimateQualityJudge` skips silently only when the
> block is ABSENT (`changecode.go:589-593`), so seeding one would dispatch a second real
> agent per round whose failure would end the run for a reason unrelated to the plan gate.

> The loop deliberately does **not** simulate a plan author. Rounds 2+ require a real
> plan edit responding to round 1's findings; that is the session's work, and faking it
> would prove nothing about convergence. The harness's job is to make each round
> reproducible and the ledger inspectable.

- [x] **Step 3: Run the rounds for real.** Round 1 against the recovered plan; then edit
      the scratch plan to address only what round 1 raised; then round 2. Read
      `workshop/plans/000900-*-plan-gate.md` between rounds to confirm the dispositions
      landed.

      **Write that between-rounds edit in the strategy-line form** — name the functions
      under test plus one adversarial-strategy line each, rather than re-enumerating cases.
      This costs nothing extra and turns the edit into C1's **positive control**: Done-when
      has two halves — *"a plan naming test functions + strategy **passes** the gate"* and
      *"a plan enumerating 15 prose test cases **draws a finding**"* — and a prompt rewrite
      that made the judge reject *both* shapes would otherwise ship green.

- [x] **Step 4: Record the evidence** in `000187-replay-pair127.md`. Four questions —
      the fourth is the ONLY verification C1 gets anywhere in this plan:
      - rounds to acceptance (baseline: 6 invocations / 5 rejections);
      - **load-bearing check 1** — did the seam relocation still surface? Quote verbatim.
      - **load-bearing check 2** — did the absorb-layer removal still surface? Quote verbatim.
      - **load-bearing check 3 (C1, negative)** — pair#127's plan is the control for
        *"a plan enumerating 15 prose test cases draws a finding telling it to compress"*:
        the issue cites its ~15 prose test bullets at
        `000127-term-pane-stream-corruption.md:346`. Did the judge raise a finding against
        that enumeration?
      - **load-bearing check 4 (C1, positive)** — did the Step-3 strategy-line rewrite
        **pass** without drawing a test-surface finding? Checks 3 and 4 together are the
        only place C1's *semantics* are exercised anywhere in this plan (Task 7's guards
        test plumbing: that the model renders and that `PriorFindings` arrives). Check 3
        alone would be satisfied by a prompt that rejects every shape of test description.
      - If any of the three checks fails, the tuning **weakened review** (checks 1–2) or
        **didn't land** (check 3): revisit Task 7's prompt and re-run before closing. This
        is a gate on the issue, not a report.

- [x] **Step 5:** Summarize the outcome in the issue's `## Log` and commit.

---

### Task 15: Close the issue

- [x] **Step 1:** `go test ./... && go vet ./...`
- [x] **Step 2:** Tick every Plan row; write the final `## Log` entry.
- [x] **Step 3:** `sdlc actual --issue 187` to preview the measured hours (never
      hand-type them).
- [ ] **Step 4:** `sdlc close --issue 187 --verified '<evidence>'` — omit `--actual` so
      close measures and adopts it. The binary auto-dispatches the mandatory close review.
- [ ] **Step 5:** Add a `workshop/lessons.md` entry **only if** the close review surfaces a
      mistake not already prevented by code or tooling.

---

## Risks and open questions

1. **A judge that ignores the dispose-first instruction.** Mitigation: `ApplyChecked`
   rejects dispositions of unknown IDs, and `Decide` blocks only on *open* findings — a
   judge that silently drops a prior finding leaves it open and blocking, which fails
   *safe* (it costs a round-trip, it never lets a Critical through). The failure mode is
   the status quo, not something worse.

2. **Findings-block adoption across agent CLIs.** `codex`/`gemini` may comply less
   reliably than `claude`. The transitional prose fallback (Task 8 Step 4) keeps
   change-code working; the round is still persisted with `protocol_error` set, so the
   cap and `gate_rounds` stay correct even for a CLI that never emits the fence.
   **Fail-closed trigger, stated in advance so it is not relitigated later:** once M2's
   ledger holds ≥10 closes, if `protocol_error` rounds are **<5%** of all plan-gate rounds,
   drop the prose fallback and make a missing block a hard refusal (the
   agent-binary-handoff-schema target's end state). If it is **>20%**, the block grammar is
   too demanding for the fleet — simplify the schema rather than keeping a permanent
   fallback. Between 5% and 20%, keep the fallback and re-measure at 20 closes.

3. **#162 (`milestone-close` derives the review window from a wrong base) is open** and
   M2's churn reads the same `boundaryWindowBase`. That is deliberate: sharing the bug
   means #162's fix corrects churn for free. Do **not** work around it here.

4. ~~**Plan-gate sidecars may escape artifact archiving.**~~ **Resolved before planning
   closed:** archiving globs `filepath.Join(plansFull, id+"-*")` (`cmd/sdlc/push.go:286`),
   which already matches `000187-…-plan-gate.md`. No change needed — but
   `archiveartifacts_test.go` currently fixtures only `-plan.md` and `-close-review.md`,
   so **add a `-plan-gate.md` case to it in Task 6** to pin the coverage rather than
   leaving it incidental.

5. **Live conformance for the ` ```findings ` fence (`ARCH-MOCK`).** M1 ships a hard
   dependency on three agent CLIs emitting the fence, and Task 8 Step 6 proves only that
   the prompt renders (`--dry-run` returns before `judge.Dispatch`). There is no live
   conformance precedent to reuse — #147's ` ```verdict ` fence shipped without one, which
   is why `ParseVerdict` still carries three prose fallback layers (`classify.go:200-230`).
   **Task 14's real-agent replay IS the de-facto conformance check for `claude`**, and that
   makes it load-bearing for two reasons, not one. `codex`/`gemini` conformance is covered
   only by the fallback until M2's `protocol_error` rate answers it from production data
   (Risk 2). Accepted knowingly rather than building a third-party-CLI harness here.

6. **The findings block adds a second machine-read fence to one agent response** (the
   plan judge emits `VERDICT:` *and* ` ```findings `). That is a transitional redundancy,
   not the end state: once adoption is confirmed, the token drops and `Decide` is the only
   reader. Recorded here so a future reader does not mistake it for a design intent.
