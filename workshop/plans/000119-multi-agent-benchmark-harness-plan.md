# Multi-Agent Benchmark Harness Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `sdlc bench` — a controlled-experiment harness that runs different coding agents on the same frozen backlog issue in isolated worktrees, then grades each via measured objective signals plus blind head-to-head LLM/operator judging.

**Architecture:** One new package `cmd/sdlc/internal/bench/` with a **pure core** (task/run-record/scorecard/rubric structs + their parse/render, anonymization, judge-output parsing, leaderboard aggregation, prompt building — all unit-tested with no IO mocks) and a **thin IO shell** (a `Store` for the `workshop/benchmarks/` filesystem, a `Worktreer` over git, a `Runner` that dispatches agents via the existing `judge.Dispatch`, a `Measurer` that runs build/test in a worktree). Cobra verbs in `cmd/sdlc/bench.go` wire the shell to the core. On-disk artifacts are markdown with scalar frontmatter + stdlib-`encoding/json` fenced blocks (no new dependency; the repo has only cobra).

**Tech Stack:** Go 1.26, cobra (existing), stdlib `encoding/json`/`os/exec`/`context`. Reuses `cmd/sdlc/internal/judge` (Dispatch shim + AgentCLI), `cmd/sdlc/internal/gitx` (git helpers), `cmd/sdlc/internal/issue` (frontmatter + section extraction), `cmd/sdlc/branchcreate.go` (worktree pattern).

**ARCH alignment.** `ARCH-PURE`: every entity in the *Pure entities* table below is a deterministic function on structs, colocated tests run without mocks; all clock/git/exec/file access is isolated to the *Integration points* table and injected. `ARCH-DRY`: the harness reuses `judge.Dispatch`, `gitx`, `issue.Parse/Compose/SectionBody`, and the `close.go` log-insertion pattern rather than re-implementing dispatch, git, or frontmatter handling — named at each point of use. The `VERDICT` *classifier* is deliberately **not** reused (it is a binary gate; Stage-B needs structured rankings → new parser), and `createWorktreeBranch` is **extended** (it hardcodes `HEAD`; bench needs a base ref).

---

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `Mode` | `cmd/sdlc/internal/bench/types.go` | new |
| `Task` | `cmd/sdlc/internal/bench/task.go` | new |
| `Rubric` / `Dimension` / `ObjectiveCheck` | `cmd/sdlc/internal/bench/rubric.go` | new |
| `RunRecord` / `AgentResult` | `cmd/sdlc/internal/bench/runrecord.go` | new |
| `RawSignals` / `Scorecard` / `Metrics` | `cmd/sdlc/internal/bench/scorecard.go` | new |
| `Submission` / `Scrub` / `BuildSubmissions` | `cmd/sdlc/internal/bench/anonymize.go` | new |
| `JudgeVerdict` / `ParseJudgeOutput` | `cmd/sdlc/internal/bench/judgeparse.go` | new |
| `Leaderboard` / `Aggregate` | `cmd/sdlc/internal/bench/leaderboard.go` | new |
| `SolvePrompt` / `JudgePrompt` builders | `cmd/sdlc/internal/bench/prompts.go` | new |

- **Task** — the immutable frozen experiment: `ID` (slug), `Repo`, `SourceIssue`, `BaseSHA`, `Created`, `Spec` (verbatim issue `## Spec`), `Setup []string`, `Rubric`. `ParseTask(text)→(Task,error)` and `RenderTask(Task)→string` are inverses; the JSON-block round-trips through `encoding/json`.
  - **Relationships:** 1:N with RunRecord (a task is replayed many times). Owns its Rubric.
  - **DRY rationale:** Single source for "what was asked + where it started + how it's graded." Without it, freeze/run/grade each re-derive base SHA and rubric.
  - **Future extensions:** add `Reference` (backfilled real solution) and `Mode` default; the struct widens, parse stays back-compatible (missing field → zero value).
- **Rubric** — `Objective []ObjectiveCheck` (measured) + `Subjective []Dimension` (judged); each carries `Group` (`"quality"`|`"workflow-fit"`) and `Weight`. `DefaultRubric()` returns the standard set.
  - **DRY rationale:** the two score groups (quality / workflow-fit) are defined once; Scorecard scoring and Leaderboard aggregation both read `Group`/`Weight` from here.
- **RunRecord** — one execution: `Task`, `RunID`, `Mode`, `Created`, `JudgeModel`, `Status` (`pending`|`graded`|`reviewed`), `LeaderboardEligible bool`, `Agents []AgentResult`. **AgentResult**: `Agent`, `CLIVersion`, `Branch`, `Commits`, `DiffAdded/Deleted/Files`, `WallClockSec`, `TurnCount`, `Completed`, `TranscriptPath`, `ExitOK`, `Scorecard *Scorecard`, `Verdicts []DimensionVerdict`. `ParseRunRecord`/`RenderRunRecord` round-trip the `## Data` JSON block; `## Summary` tables are rendered one-directionally (never parsed back).
  - **Relationships:** N:1 with Task; N agents per run.
  - **DRY rationale:** the JSON block is the sole source of truth; human tables derive from it — no double-maintained data.
- **Scorecard** — Stage-A output. `ScoreObjective(RawSignals, Rubric)→Scorecard` is pure: maps raw build/test/artifact booleans + metrics into per-check 0..1 scores and weighted `QualityScore`/`WorkflowScore`. **RawSignals** is the measured input (filled by the Measurer IO seam); keeping it a separate struct is what makes scoring unit-testable with zero IO.
  - **DRY rationale:** all "how a raw signal becomes a score" logic lives here, not smeared across the Measurer.
- **Submission / Scrub** — `Scrub(text string, secrets []string)→string` removes identity (agent names, `Co-Authored-By` trailers, `bench/<agent>/…` branch names, self-references) by literal + regex replacement. `BuildSubmissions(results, artifacts, seed)→([]Submission, map[string]string)` assembles anonymized packets and returns the label→agent mapping; label order is a **deterministic seeded shuffle** (seed injected, so pure + reproducible).
  - **DRY rationale:** one scrub routine feeds both the LLM judge and the operator doc.
  - **Future extensions:** `secrets` list widens as new identity tells are found (the anonymization-leak test drives this).
- **JudgeVerdict / ParseJudgeOutput** — the Stage-B judge emits a fenced ```json block `{"dimensions":[{"key","ranking":[labels…],"rationale","confidence"}]}`; `ParseJudgeOutput(output)→(JudgeVerdict,error)` extracts and unmarshals it, tolerant of preamble. This is the **net-new structured parser** that replaces the binary `VERDICT` classifier for bench.
- **Leaderboard / Aggregate** — `Aggregate([]RunRecord)→Leaderboard` (pure): per-`(agent,version)` objective-metric distributions + per-dimension head-to-head win-rates; `RenderLeaderboard(Leaderboard)→string`.
- **Prompt builders** — `SolvePrompt(Task)→string` (the constant worker prompt) and `JudgePrompt(Task, []Submission)→string` (the blind head-to-head prompt + the JSON output contract). Pure string builders, snapshot-tested.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `Store` | `cmd/sdlc/internal/bench/store.go` | new | filesystem (`workshop/benchmarks/`) |
| gitx additions (`WorktreeAddBase`, `DiffStat`, `ShowAtSHA`, `CommitsOnRange`) | `cmd/sdlc/internal/gitx/bench.go` | new | git |
| `Worktreer` | `cmd/sdlc/internal/bench/worktree.go` | new | git worktree |
| `VersionProbe` | `cmd/sdlc/internal/bench/version.go` | new | `exec` (`claude --version` …) |
| `Runner` | `cmd/sdlc/internal/bench/runner.go` | new | `judge.Dispatch` subprocess |
| `Measurer` | `cmd/sdlc/internal/bench/measure.go` | new | build/test/git in a worktree |
| `Responder` (interface) + `AutonomousResponder` | `cmd/sdlc/internal/bench/responder.go` | new | (seam; live/interactive later) |
| bench cobra verbs | `cmd/sdlc/bench.go` | new | CLI |
| helptext | `cmd/sdlc/helptext/bench*.md` | new | embed |

- **Store** — reads/writes Task, RunRecord, Leaderboard files; allocates `RunID` by scanning `runs/<task>-*.md`. Injected into every verb. Pure path/format logic (filenames, next-id) lives in small pure helpers it calls; the struct only does `os.ReadFile`/`os.WriteFile`/`os.ReadDir`.
- **gitx additions** — follow the existing `var run`/`RunGit`/`Capture` pattern so tests fake `run`. `WorktreeAddBase(name, path, base)` runs `git worktree add -b name path <base>` (the **base-ref** the existing `createWorktreeBranch` lacks). `DiffStat(base, head)→(added,deleted,files,err)`. `CommitsOnRange(base, head)→[]LogEntry`. `ShowAtSHA(sha, path)→(string,err)`.
- **Worktreer** — `Add(task, agent, runid)→(path,branch,err)` and `Remove(path)`; thin wrapper over the gitx worktree helper + the branchcreate path convention (`../worktree/<repo>/bench/<task>/<agent>/<runid>`). Injected into Runner.
- **VersionProbe** — `Version(agent)→string` shells `claude --version`/`codex --version`/`gemini --version`; `var probeRun` seam for tests. Captured into AgentResult.
- **Runner** — `RunAgent(ctx, task, agent, wt)→AgentResult`: builds `SolvePrompt`, dispatches via `judge.Dispatch` with a **write allowlist** (`"Read,Edit,Write,Grep,Glob,Bash"`) and a **timeout `ctx`** (the budget — the first Dispatch caller to use both), captures wall-clock + exit + transcript path. Pure assembly (prompt, AgentResult construction) is separated from the Dispatch call.
- **Measurer** — `Measure(task, wt, result)→RawSignals`: runs `task.Setup`/build/test in the worktree, inspects git log (commit conventions, `Review-Verdict:` trailers) and file presence for artifacts. Returns RawSignals; **all scoring is deferred to the pure `ScoreObjective`.**
- **Responder** — `interface { Answer(question string) (string, error) }`; `AutonomousResponder` returns an error/"proceed" (autonomous never blocks). The seam exists so interactive/live slot in later without touching the Runner signature.

---

## Chunk 1: M1 — Data model, datatype, layout, `freeze`

**Milestone M1 review boundary** — closes with `sdlc milestone-close --issue 119 --milestone M1`.

### Task 1.1: Package skeleton + `Mode` + `Task` struct

**Files:**
- Create: `cmd/sdlc/internal/bench/types.go`
- Create: `cmd/sdlc/internal/bench/task.go`
- Test: `cmd/sdlc/internal/bench/task_test.go`

- [ ] **Step 1: Write the failing test** (`task_test.go`)

```go
package bench

import "testing"

func TestTaskRoundTrip(t *testing.T) {
	in := Task{
		ID: "119-demo", Repo: "ariadne", SourceIssue: "119",
		BaseSHA: "abc123", Created: "2026-06-19",
		// Spec deliberately embeds a ```json fence — the config extraction must
		// NOT pick this up (regression guard for the first-block bug).
		Spec:  "Solve the thing.\n\n```json\n{\"example\": true}\n```\n\nDetails here.",
		Setup: []string{"go build ./..."},
		Rubric: DefaultRubric(),
	}
	got, err := ParseTask(RenderTask(in))
	if err != nil {
		t.Fatalf("ParseTask: %v", err)
	}
	if len(got.Setup) != 1 || got.Setup[0] != "go build ./..." {
		t.Fatalf("config leaked from spec's json fence: setup=%v", got.Setup)
	}
	if got.ID != in.ID || got.BaseSHA != in.BaseSHA || got.Spec != in.Spec {
		t.Errorf("scalar/spec mismatch:\n got %+v\nwant %+v", got, in)
	}
	if len(got.Setup) != 1 || got.Setup[0] != "go build ./..." {
		t.Errorf("setup mismatch: %v", got.Setup)
	}
	if len(got.Rubric.Subjective) != len(in.Rubric.Subjective) {
		t.Errorf("rubric not round-tripped")
	}
}
```

- [ ] **Step 2: Run, verify it fails** — `go test ./cmd/sdlc/internal/bench/ -run TestTaskRoundTrip` → FAIL (undefined: Task).

- [ ] **Step 3: Implement** `types.go`:

```go
package bench

type Mode string

const (
	ModeAutonomous  Mode = "autonomous"
	ModeInteractive Mode = "interactive"
	ModeLive        Mode = "live"
)
```

and `task.go`:

```go
package bench

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

type Task struct {
	ID          string
	Repo        string
	SourceIssue string
	BaseSHA     string
	Created     string
	Spec        string
	Setup       []string
	Rubric      Rubric
}

// taskConfig is the JSON payload embedded in the ## Config section.
type taskConfig struct {
	Setup  []string `json:"setup"`
	Rubric Rubric   `json:"rubric"`
}

func RenderTask(t Task) string {
	fm := strings.Join([]string{
		"type: benchmark-task",
		"id: " + t.ID,
		"repo: " + t.Repo,
		"source_issue: " + t.SourceIssue,
		"base_sha: " + t.BaseSHA,
		"created: " + t.Created,
	}, "\n")
	cfg, _ := json.MarshalIndent(taskConfig{Setup: t.Setup, Rubric: t.Rubric}, "", "  ")
	body := fmt.Sprintf("# Benchmark task: %s\n\n## Spec\n\n%s\n\n## Config\n\n```json\n%s\n```\n",
		t.ID, strings.TrimRight(t.Spec, "\n"), string(cfg))
	return issue.Compose(fm, body)
}

func ParseTask(text string) (Task, error) {
	fm, body, err := issue.Parse(text)
	if err != nil {
		return Task{}, err
	}
	get := func(k string) string { v, _ := issue.GetField(fm, k); return v }
	t := Task{
		ID: get("id"), Repo: get("repo"), SourceIssue: get("source_issue"),
		BaseSHA: get("base_sha"), Created: get("created"),
	}
	if spec, ok := issue.SectionBody(body, "Spec"); ok {
		t.Spec = strings.TrimSpace(spec)
	}
	// Scope extraction to the ## Config section — a verbatim spec may itself
	// contain a ```json fence, and extractJSONBlock returns the FIRST one.
	cfgBody, ok := issue.SectionBody(body, "Config")
	if !ok {
		return Task{}, fmt.Errorf("benchmark-task %s: missing ## Config section", t.ID)
	}
	raw, ok := extractJSONBlock(cfgBody)
	if !ok {
		return Task{}, fmt.Errorf("benchmark-task %s: missing ## Config json block", t.ID)
	}
	var cfg taskConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return Task{}, fmt.Errorf("benchmark-task %s: bad config json: %w", t.ID, err)
	}
	t.Setup, t.Rubric = cfg.Setup, cfg.Rubric
	return t, nil
}
```

- [ ] **Step 4: Add the pure `extractJSONBlock` helper** in `task.go` (reused by run-record parsing too — DRY):

```go
// extractJSONBlock returns the content of the first ```json fenced block in body.
func extractJSONBlock(body string) (string, bool) {
	const open = "```json"
	i := strings.Index(body, open)
	if i < 0 {
		return "", false
	}
	rest := body[i+len(open):]
	j := strings.Index(rest, "```")
	if j < 0 {
		return "", false
	}
	return strings.TrimSpace(rest[:j]), true
}
```

- [ ] **Step 5: Run tests** — `go test ./cmd/sdlc/internal/bench/ -run TestTaskRoundTrip` → PASS.

- [ ] **Step 6: Commit** — `git commit -m "#119 M1: bench Task struct + round-trip parse/render"`

### Task 1.2: `Rubric` + `DefaultRubric`

**Files:**
- Create: `cmd/sdlc/internal/bench/rubric.go`
- Test: `cmd/sdlc/internal/bench/rubric_test.go`

- [ ] **Step 1: Failing test** — assert `DefaultRubric()` has both groups represented and weights normalize per group:

```go
func TestDefaultRubricGroups(t *testing.T) {
	r := DefaultRubric()
	groups := map[string]bool{}
	for _, d := range r.Subjective {
		groups[d.Group] = true
	}
	for _, c := range r.Objective {
		groups[c.Group] = true
	}
	if !groups["quality"] || !groups["workflow-fit"] {
		t.Fatalf("both groups must be present, got %v", groups)
	}
}
```

- [ ] **Step 2: Run, verify fail.**

- [ ] **Step 3: Implement** `rubric.go`:

```go
package bench

type Dimension struct {
	Key    string  `json:"key"`
	Group  string  `json:"group"`  // "quality" | "workflow-fit"
	Weight float64 `json:"weight"`
	Prompt string  `json:"prompt"` // what the judge evaluates
}

type ObjectiveCheck struct {
	Key    string  `json:"key"`
	Group  string  `json:"group"`
	Weight float64 `json:"weight"`
}

type Rubric struct {
	Objective  []ObjectiveCheck `json:"objective"`
	Subjective []Dimension      `json:"subjective"`
}

func DefaultRubric() Rubric {
	return Rubric{
		Objective: []ObjectiveCheck{
			{"build", "quality", 1},
			{"existing-tests", "quality", 1},
			{"new-tests", "quality", 1},
			{"completed", "quality", 1},
			{"artifact-log", "workflow-fit", 1},
			{"artifact-plan-ticked", "workflow-fit", 1},
			{"artifact-atlas", "workflow-fit", 1},
			{"gates-run", "workflow-fit", 1},
		},
		Subjective: []Dimension{
			{"elegance", "quality", 1, "Which solution is more elegant — DRY, pure core, root-cause not patch?"},
			{"design-reasoning", "quality", 1, "Which reasons better about design/UI subtleties (read its spec/plan/diff)?"},
			{"doc-quality", "quality", 1, "Which produced clearer, better-judged docs/spec/plan?"},
			{"gate-judgment", "workflow-fit", 1, "Which made better decisions at the SDLC gates?"},
		},
	}
}
```

- [ ] **Step 4: Run → PASS. Step 5: Commit** — `git commit -m "#119 M1: bench Rubric + DefaultRubric (quality + workflow-fit groups)"`

### Task 1.3: `Store` (task read/write) + `benchmark-task` datatype doc + layout

**Files:**
- Create: `cmd/sdlc/internal/bench/store.go`
- Test: `cmd/sdlc/internal/bench/store_test.go`
- Create: `construct/datatype/benchmark-task.md`
- Create: `workshop/benchmarks/README.md`

- [ ] **Step 1: Failing test** — write a task via Store to a temp dir, read it back:

```go
func TestStoreTaskRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	task := Task{ID: "119-demo", Repo: "ariadne", BaseSHA: "abc", Created: "2026-06-19",
		Spec: "do it", Setup: []string{"go build ./..."}, Rubric: DefaultRubric()}
	if err := s.WriteTask(task); err != nil {
		t.Fatal(err)
	}
	got, err := s.ReadTask("119-demo")
	if err != nil {
		t.Fatal(err)
	}
	if got.BaseSHA != "abc" {
		t.Errorf("base_sha = %q", got.BaseSHA)
	}
}
```

- [ ] **Step 2: Run, verify fail.**

- [ ] **Step 3: Implement** `store.go` (thin IO; format logic delegates to RenderTask/ParseTask):

```go
package bench

import (
	"os"
	"path/filepath"
)

type Store struct{ root string } // e.g. "workshop/benchmarks"

func NewStore(root string) *Store { return &Store{root: root} }

func (s *Store) tasksDir() string { return filepath.Join(s.root, "tasks") }
func (s *Store) runsDir() string  { return filepath.Join(s.root, "runs") }

func (s *Store) WriteTask(t Task) error {
	if err := os.MkdirAll(s.tasksDir(), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.tasksDir(), t.ID+".md"), []byte(RenderTask(t)), 0o644)
}

func (s *Store) ReadTask(id string) (Task, error) {
	b, err := os.ReadFile(filepath.Join(s.tasksDir(), id+".md"))
	if err != nil {
		return Task{}, err
	}
	return ParseTask(string(b))
}
```

- [ ] **Step 4: Run → PASS.**

- [ ] **Step 5: Write `construct/datatype/benchmark-task.md`** — match the `construct/datatype/` convention (frontmatter `type: type`, `name`, `description`; body documents the on-disk shape). Content:

```markdown
---
type: type
name: benchmark-task
description: An immutable, replayable benchmark task frozen from a live issue — spec snapshot + base SHA + grading rubric — that different coding agents are run against.
---

# Datatype: benchmark-task

A **benchmark-task** is the immutable definition of one controlled experiment for
the `sdlc bench` harness (#119). It is frozen from a live backlog issue at a
point in time and never changes thereafter, so any agent can be replayed against
the exact same starting conditions even after `main` has advanced.

Lives in: `workshop/benchmarks/tasks/<id>.md` (repo-local; the harness is base-layer).

## Frontmatter

- `type: benchmark-task`
- `id` — slug, usually `<issue>-<short-slug>`; matches the filename
- `repo` — repo the task is frozen from
- `source_issue` — the live issue number it was frozen from
- `base_sha` — the immutable commit agents branch from (the reproducibility anchor)
- `created` — ISO date

## Body

- `## Spec` — the issue's `## Spec`, copied **verbatim**; the constant prompt context.
- `## Config` — a fenced ` ```json ` block: `{ "setup": [...], "rubric": {...} }`.
  - `setup` — shell commands to bring a fresh worktree to green before the agent runs.
  - `rubric` — `objective` checks (measured) + `subjective` dimensions (judged), each
    tagged `group` (`quality` | `workflow-fit`) and `weight`. See `DefaultRubric()`.

## Invariants

- Immutable after `sdlc bench freeze`. To change grading, freeze a new task.
- `base_sha` must remain reachable; the harness branches from it and never merges.
```

- [ ] **Step 6: Write `workshop/benchmarks/README.md`** documenting the layout (`tasks/`, `runs/`, `leaderboard.md`) and pointing at `sdlc bench --help` and the datatype doc.

- [ ] **Step 7: Commit** — `git commit -m "#119 M1: bench Store + benchmark-task datatype + workshop/benchmarks layout"`

### Task 1.4: `freeze` command

**Files:**
- Create: `cmd/sdlc/bench.go`
- Create: `cmd/sdlc/helptext/bench.md`, `cmd/sdlc/helptext/bench-freeze.md`
- Modify: `cmd/sdlc/main.go` (register `NewBenchCmd`)
- Test: `cmd/sdlc/bench_freeze_test.go`

- [ ] **Step 1: Failing test** — drive `runBenchFreeze` against a fixture issue + a fake "HEAD sha" injection, assert a task file is written with the issue's Spec and the pinned base sha. Use a `var headSHA = func() (string,error){...}` seam in bench.go so the test injects `"deadbeef"`.

```go
func TestRunBenchFreeze(t *testing.T) {
	root := t.TempDir()
	// seed a fake issue file
	issuesDir := filepath.Join(root, "workshop", "issues")
	os.MkdirAll(issuesDir, 0o755)
	os.WriteFile(filepath.Join(issuesDir, "000119-demo.md"),
		[]byte("---\nid: 000119\nstatus: working\n---\n\n# Demo\n\n## Spec\n\nDo the thing.\n"), 0o644)
	orig := headSHA
	headSHA = func() (string, error) { return "deadbeef", nil }
	defer func() { headSHA = orig }()

	var out, errBuf bytes.Buffer
	err := runBenchFreeze(&out, &errBuf, benchFreezeFlags{Issue: 119, Repo: "ariadne", Root: root})
	if err != nil {
		t.Fatal(err)
	}
	got, err := NewStore(filepath.Join(root, "workshop", "benchmarks")).ReadTask("119-demo")
	if err != nil {
		t.Fatal(err)
	}
	if got.BaseSHA != "deadbeef" || !strings.Contains(got.Spec, "Do the thing") {
		t.Errorf("freeze wrong: %+v", got)
	}
}
```

- [ ] **Step 2: Run, verify fail.**

- [ ] **Step 3: Implement** `bench.go` — parent cmd + `freeze` child + `runBenchFreeze`. Use `gitx.Capture("rev-parse","HEAD")` behind the `headSHA` seam; locate the issue file by globbing `workshop/issues/*<issue>*`; derive the slug from the issue filename; extract `## Spec` via `issue.SectionBody`.

```go
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

var headSHA = func() (string, error) {
	s := gitx.Capture("rev-parse", "HEAD")
	if s == "" {
		return "", fmt.Errorf("could not resolve HEAD")
	}
	return s, nil
}

func NewBenchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "bench", Short: "Benchmark coding agents on the same frozen task",
		Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(c *cobra.Command, _ []string) error { return c.Help() },
	}
	cmd.AddCommand(newBenchFreezeCmd())
	return cmd
}

type benchFreezeFlags struct {
	Issue int
	Repo  string
	Root  string // repo root override (tests); "" → cwd
}

func newBenchFreezeCmd() *cobra.Command {
	var f benchFreezeFlags
	cmd := &cobra.Command{
		Use: "freeze", Short: "Snapshot a live issue into an immutable benchmark task",
		Args: cobra.NoArgs, SilenceErrors: true,
		RunE: func(c *cobra.Command, _ []string) error {
			return runBenchFreeze(c.OutOrStdout(), c.ErrOrStderr(), f)
		},
	}
	cmd.Flags().IntVar(&f.Issue, "issue", 0, "live issue to freeze (required)")
	cmd.Flags().StringVar(&f.Repo, "repo", "ariadne", "repo name recorded on the task")
	return cmd
}

func runBenchFreeze(stdout, stderr io.Writer, f benchFreezeFlags) error {
	if f.Issue == 0 {
		return fmt.Errorf("--issue is required")
	}
	root := f.Root
	if root == "" {
		root = "."
	}
	path, slug, err := findIssueFile(root, f.Issue)
	if err != nil {
		return err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, body, err := issue.Parse(string(b))
	if err != nil {
		return err
	}
	spec, ok := issue.SectionBody(body, "Spec")
	if !ok {
		return fmt.Errorf("issue %d has no ## Spec to freeze", f.Issue)
	}
	sha, err := headSHA()
	if err != nil {
		return err
	}
	task := bench.Task{
		ID: fmt.Sprintf("%d-%s", f.Issue, slug), Repo: f.Repo,
		SourceIssue: fmt.Sprintf("%d", f.Issue), BaseSHA: sha,
		Created: today(), Spec: strings.TrimSpace(spec),
		Setup: []string{"go build ./..."}, Rubric: bench.DefaultRubric(),
	}
	s := bench.NewStore(filepath.Join(root, "workshop", "benchmarks"))
	if err := s.WriteTask(task); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "froze issue %d → task %s (base %s)\n", f.Issue, task.ID, sha[:min(8, len(sha))])
	return nil
}
```

(`findIssueFile`, `today`, `min` are small helpers — `findIssueFile` globs `workshop/issues/*` and parses the 6-digit prefix + derives the slug after `NNNNNN-`; `today` reuses whatever the repo already uses for ISO date, else `time.Now().Format("2006-01-02")` in this IO function.)

- [ ] **Step 4: Register in `main.go`** — add `add(NewBenchCmd(), "bench", "Benchmark coding agents on the same frozen task")` in `buildRoot`, and create `helptext/bench.md` + `helptext/bench-freeze.md`.

- [ ] **Step 5: Run** — `go test ./cmd/sdlc/... -run Freeze` → PASS; `go build ./...`.

- [ ] **Step 6: Manual smoke** — `sdlc bench freeze --issue 119` writes `workshop/benchmarks/tasks/119-multi-agent-benchmark-harness.md`. Inspect it.

- [ ] **Step 7: Commit** — `git commit -m "#119 M1: sdlc bench freeze — snapshot a live issue into an immutable task"`

### Task 1.5: M1 milestone close

- [ ] Update `## Log` (#119) with M1 discoveries; tick the M1 box in `## Plan`.
- [ ] `sdlc milestone-close --issue 119 --milestone M1` → fix any Critical/Important from the boundary-review judge; record the `Review-Verdict:`.

---

## Chunk 2: M2 — Runner (autonomous), responder seam, no-merge isolation

**Milestone M2 review boundary** — closes with `sdlc milestone-close --issue 119 --milestone M2`.

### Task 2.1: gitx worktree-with-base + range helpers

**Files:**
- Create: `cmd/sdlc/internal/gitx/bench.go`
- Test: `cmd/sdlc/internal/gitx/bench_test.go`

- [ ] **Step 1: Failing test** — fake the package `run` shim, assert `WorktreeAddBase` issues `git worktree add -b <branch> <path> <base>` (the base ref, NOT `HEAD`):

```go
func TestWorktreeAddBaseUsesBaseRef(t *testing.T) {
	orig := run
	defer func() { run = orig }()
	var gotArgs []string
	run = func(name string, args ...string) ([]byte, error) { gotArgs = args; return nil, nil }
	if err := WorktreeAddBase("b1", "/tmp/wt", "abc123"); err != nil {
		t.Fatal(err)
	}
	want := []string{"worktree", "add", "-b", "b1", "/tmp/wt", "abc123"}
	if !slices.Equal(gotArgs, want) {
		t.Errorf("args = %v, want %v", gotArgs, want)
	}
}
```

- [ ] **Step 2: Run, verify fail.**
- [ ] **Step 3: Implement** `gitx/bench.go` — `WorktreeAddBase`, `WorktreeRemove`, `DiffStat(base,head)`, `CommitsOnRange(base,head)`, `ShowAtSHA(sha,path)`, all via `run`/`RunGit`. `DiffStat` parses `git diff --numstat base head`.
- [ ] **Step 4: Run → PASS. Step 5: Commit** — `git commit -m "#119 M2: gitx worktree-with-base + diffstat/range/show helpers"`

### Task 2.2: `Responder` seam + `RunRecord`/`AgentResult` + render/parse

**Files:**
- Create: `cmd/sdlc/internal/bench/responder.go`
- Create: `cmd/sdlc/internal/bench/runrecord.go`
- Test: `cmd/sdlc/internal/bench/runrecord_test.go`

- [ ] **Step 1: Failing test** — RunRecord round-trips through the `## Data` JSON block (including a nested `AgentResult` with a Scorecard).
- [ ] **Step 2: Run, verify fail.**
- [ ] **Step 3: Implement** `responder.go`:

```go
package bench

type Responder interface {
	Answer(question string) (string, error)
}

// AutonomousResponder never blocks — autonomous runs proceed without human input.
type AutonomousResponder struct{}

func (AutonomousResponder) Answer(string) (string, error) {
	return "", errProceedAutonomously
}

var errProceedAutonomously = fmt.Errorf("autonomous run: no responder available")
```

and `runrecord.go` (structs + `RenderRunRecord`/`ParseRunRecord`; JSON block is source of truth, `## Summary` tables rendered one-directionally via a pure `renderSummary([]AgentResult) string`). **`ParseRunRecord` must scope its block extraction to the `## Data` section** (`issue.SectionBody(body, "Data")` → `extractJSONBlock`), the same first-block guard as `ParseTask` — so a transcript excerpt or future section carrying a ```json fence can never be mistaken for the data block.

- [ ] **Step 4: Run → PASS. Step 5: Commit** — `git commit -m "#119 M2: RunRecord/AgentResult round-trip + Responder seam"`

### Task 2.3: `SolvePrompt` builder

**Files:**
- Create: `cmd/sdlc/internal/bench/prompts.go`
- Test: `cmd/sdlc/internal/bench/prompts_test.go`

- [ ] **Step 1: Failing test** — `SolvePrompt(task)` contains the issue number, the constant no-merge instruction, and the spec; snapshot-assert key substrings.
- [ ] **Step 2–4: Implement + pass.** `SolvePrompt` is pure:

```go
func SolvePrompt(t Task) string {
	return fmt.Sprintf(`Solve issue %s in this repository, following the repo's own conventions (read AGENTS.md and any skills as you would normally).

Commit your work as you go. Do NOT merge to main and do NOT open a pull request — the branch you are on is the deliverable; it will be graded as-is.

The task:

%s`, t.SourceIssue, t.Spec)
}
```

- [ ] **Step 5: Commit** — `git commit -m "#119 M2: SolvePrompt builder (constant worker prompt)"`

### Task 2.4: `VersionProbe` + `Worktreer` + `Runner` (IO seams)

**Files:**
- Create: `cmd/sdlc/internal/bench/version.go`, `worktree.go`, `runner.go`
- Test: `cmd/sdlc/internal/bench/runner_test.go`

- [ ] **Step 1: Failing test** — inject a fake `judge.Run` (the dispatch shim's seam) + a fake Worktreer; assert `Runner.RunAgent` returns an `AgentResult` with the captured CLI version, branch, non-zero wall-clock, and `ExitOK` true. Reuse `judge.Run` swap exactly as `judge_test.go` does.

```go
func TestRunnerRunAgentAutonomous(t *testing.T) {
	origRun := judge.Run
	defer func() { judge.Run = origRun }()
	judge.Run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return []byte("done"), nil
	}
	r := &Runner{
		Worktree: fakeWorktreer{path: "/tmp/wt", branch: "bench/119-demo/claude/1"},
		Version:  func(a string) string { return a + "-1.2.3" },
		Now:      func() float64 { return 0 }, // injected clock seam
	}
	task := Task{ID: "119-demo", SourceIssue: "119", Spec: "do it"}
	res, err := r.RunAgent(context.Background(), task, "claude", ModeAutonomous)
	if err != nil {
		t.Fatal(err)
	}
	if res.CLIVersion != "claude-1.2.3" || res.Branch == "" || !res.ExitOK {
		t.Errorf("bad result: %+v", res)
	}
}
```

- [ ] **Step 2: Run, verify fail.**
- [ ] **Step 3: Implement** the three seams. `Runner.RunAgent`:
  1. `wt, branch := r.Worktree.Add(task, agent, runid)`
  2. `prompt := SolvePrompt(task)`
  3. `ctx, cancel := context.WithTimeout(ctx, budget)` (budget from a flag, default e.g. 30m)
  4. `out, err := judge.Dispatch(ctx, judge.DispatchOptions{Agent: judge.AgentCLI(agent), Prompt: prompt, AllowedTools: "Read,Edit,Write,Grep,Glob,Bash", IsSandbox: true, Stdout: …, Stderr: …})` — **note** the write allowlist + timeout ctx (first such caller).
  5. capture wall-clock via injected `Now`, version via `r.Version`, transcript to a file under the run dir, set `ExitOK` from err.
  - **Worktreer.Add** runs in `agent`'s own worktree under `../worktree/<repo>/bench/<task>/<agent>/<runid>` via `gitx.WorktreeAddBase(branch, path, task.BaseSHA)` — **branches from `BaseSHA`, never `HEAD`** (the immutability guarantee).
- [ ] **Step 4: Run → PASS. Step 5: Commit** — `git commit -m "#119 M2: Runner (autonomous dispatch, write-allowlist + timeout) + Worktreer + VersionProbe"`

### Task 2.5: `run` command + no-merge isolation test

**Files:**
- Modify: `cmd/sdlc/bench.go` (add `run` child)
- Create: `cmd/sdlc/helptext/bench-run.md`
- Test: `cmd/sdlc/bench_run_test.go`

- [ ] **Step 1: Failing integration test (process-level, real git)** — the **no-merge / immutability** acceptance check:
  1. `git init` a temp repo, commit a file, capture `base = HEAD`.
  2. Freeze a task at `base`. Advance `main` with another commit.
  3. Run with a **stub agent** (`judge.Run` faked to make one commit in the worktree).
  4. Assert: the agent's branch's merge-base with main is `base` (it branched from the frozen SHA, not advanced main); `main` is unchanged (no merge happened); the worktree branch still exists; **and the branch has no upstream/push target** (`git rev-parse --abbrev-ref bench/...@{upstream}` fails) — proving the spec's "structurally can't push," not merely "didn't merge."

- [ ] **Step 2: Run, verify fail.**
- [ ] **Step 3: Implement** `run` child + `runBenchRun`: load task, parse `--agents`, for each agent `Runner.RunAgent`, assemble a `RunRecord{Status: "pending"}`, `Store.WriteRun`. Default `--parallel` runs agents concurrently (each in its own worktree — no shared state); `--mode` defaults to `autonomous` (interactive/live return a "not yet wired" error citing the seam).
- [ ] **Step 4: Run → PASS. Step 5: Manual smoke** with a stub. **Step 6: Commit** — `git commit -m "#119 M2: sdlc bench run + no-merge/base-immutability isolation test"`

### Task 2.6: M2 milestone close

- [ ] Update `## Log`; tick M2. `sdlc milestone-close --issue 119 --milestone M2`; address findings; record verdict.

---

## Chunk 3: M3 — Grader Stage A (objective scorecard) + metrics

**Milestone M3 review boundary** — closes with `sdlc milestone-close --issue 119 --milestone M3`.

### Task 3.1: `Scorecard` + `ScoreObjective` (pure)

**Files:**
- Create: `cmd/sdlc/internal/bench/scorecard.go`
- Test: `cmd/sdlc/internal/bench/scorecard_test.go`

- [ ] **Step 1: Failing test** — `ScoreObjective` maps `RawSignals` → per-check 0..1 + weighted group scores, reading weights/groups from the `Rubric`:

```go
func TestScoreObjective(t *testing.T) {
	sig := RawSignals{
		BuildOK: true, ExistingTestsOK: true, NewTestsAdded: true, NewTestsOK: false,
		Completed: true,
		ArtifactsPresent: map[string]bool{"log": true, "plan-ticked": false, "atlas": true},
		GatesRun:         map[string]bool{"any": true},
		Commits: 3, DiffAdded: 120, DiffDeleted: 10, DiffFiles: 4,
	}
	sc := ScoreObjective(sig, DefaultRubric())
	if sc.Quality["build"] != 1 || sc.Quality["new-tests"] != 0 {
		t.Errorf("per-check scores wrong: %+v", sc.Quality)
	}
	if sc.QualityScore <= 0 || sc.QualityScore > 1 {
		t.Errorf("QualityScore out of range: %v", sc.QualityScore)
	}
	if sc.Metrics.Commits != 3 {
		t.Errorf("metrics not carried")
	}
}
```

- [ ] **Step 2: Run, verify fail.**
- [ ] **Step 3: Implement** `scorecard.go`:

```go
package bench

type Metrics struct {
	Commits      int     `json:"commits"`
	DiffAdded    int     `json:"diff_added"`
	DiffDeleted  int     `json:"diff_deleted"`
	DiffFiles    int     `json:"diff_files"`
	WallClockSec float64 `json:"wall_clock_sec"`
	TurnCount    int     `json:"turn_count"`
	Completed    bool    `json:"completed"`
}

type RawSignals struct {
	BuildOK          bool
	ExistingTestsOK  bool
	NewTestsAdded    bool
	NewTestsOK       bool
	Completed        bool
	ArtifactsPresent map[string]bool // "log","plan-ticked","atlas"
	GatesRun         map[string]bool // detected gate names; non-empty ⇒ ran something
	Commits          int
	DiffAdded        int
	DiffDeleted      int
	DiffFiles        int
	WallClockSec     float64
	TurnCount        int
}

type Scorecard struct {
	Quality       map[string]float64 `json:"quality"`
	Workflow      map[string]float64 `json:"workflow"`
	QualityScore  float64            `json:"quality_score"`
	WorkflowScore float64            `json:"workflow_score"`
	Metrics       Metrics            `json:"metrics"`
}

func b2f(b bool) float64 { if b { return 1 }; return 0 }

// ScoreObjective is PURE: RawSignals + Rubric → Scorecard. No IO, no clock.
func ScoreObjective(s RawSignals, r Rubric) Scorecard {
	checks := map[string]float64{
		"build":                b2f(s.BuildOK),
		"existing-tests":       b2f(s.ExistingTestsOK),
		"new-tests":            b2f(s.NewTestsAdded && s.NewTestsOK),
		"completed":            b2f(s.Completed),
		"artifact-log":         b2f(s.ArtifactsPresent["log"]),
		"artifact-plan-ticked": b2f(s.ArtifactsPresent["plan-ticked"]),
		"artifact-atlas":       b2f(s.ArtifactsPresent["atlas"]),
		"gates-run":            b2f(len(s.GatesRun) > 0),
	}
	sc := Scorecard{Quality: map[string]float64{}, Workflow: map[string]float64{}}
	var qW, wW float64
	for _, c := range r.Objective {
		v := checks[c.Key]
		switch c.Group {
		case "quality":
			sc.Quality[c.Key] = v
			sc.QualityScore += v * c.Weight
			qW += c.Weight
		case "workflow-fit":
			sc.Workflow[c.Key] = v
			sc.WorkflowScore += v * c.Weight
			wW += c.Weight
		}
	}
	if qW > 0 { sc.QualityScore /= qW }
	if wW > 0 { sc.WorkflowScore /= wW }
	sc.Metrics = Metrics{
		Commits: s.Commits, DiffAdded: s.DiffAdded, DiffDeleted: s.DiffDeleted,
		DiffFiles: s.DiffFiles, WallClockSec: s.WallClockSec, TurnCount: s.TurnCount,
		Completed: s.Completed,
	}
	return sc
}
```

- [ ] **Step 4: Run → PASS. Step 5: Commit** — `git commit -m "#119 M3: ScoreObjective — pure RawSignals→Scorecard mapping"`

### Task 3.2: artifact + gate detection (pure)

**Files:**
- Create: `cmd/sdlc/internal/bench/detect.go`
- Test: `cmd/sdlc/internal/bench/detect_test.go`

- [ ] **Step 1: Failing tests** — pure functions over already-fetched git/file data:
  - `DetectArtifacts(changedFiles []string, issueBodyAfter string)→map[string]bool` — `log` (a new `### <date>`/`- ` line in the issue's `## Log`), `plan-ticked` (`issue.CountPlanItems` shows ≥1 newly ticked), `atlas` (any `atlas/` path in `changedFiles`).
  - `DetectGates(subjects []string, trailers []string)→map[string]bool` — scans commit subjects for `#<issue>` conventions and `Review-Verdict:` trailers.
- [ ] **Step 2–4: Implement + pass.** Keep these pure (inputs are plain slices/strings); the Measurer fetches the raw git/file data and calls them.
- [ ] **Step 5: Commit** — `git commit -m "#119 M3: pure artifact + SDLC-gate detection from changed files/commits"`

### Task 3.3: `Measurer` (IO seam) + `grade` Stage A wiring

**Files:**
- Create: `cmd/sdlc/internal/bench/measure.go`
- Modify: `cmd/sdlc/bench.go` (add `grade` child — Stage A only for now)
- Create: `cmd/sdlc/helptext/bench-grade.md`
- Test: `cmd/sdlc/internal/bench/measure_test.go`

- [ ] **Step 1: Failing test** — with a fake command runner + a synthetic worktree dir, `Measure` returns a populated `RawSignals` (build/test exit codes mapped, `DiffStat` parsed, `DetectArtifacts`/`DetectGates` invoked). Fake the exec seam (`var measureRun`).
- [ ] **Step 2: Run, verify fail.**
- [ ] **Step 3: Implement** `Measure(task, wtPath, result)→RawSignals`:
  - run `task.Setup` then `go build ./...` and `go test ./...` in `wtPath` (via `measureRun` with `cmd.Dir=wtPath`), map exit codes → `BuildOK`/`ExistingTestsOK`; a second `go test` over only the agent's added test files → `NewTestsAdded/OK` (detect new `_test.go` in the diff name list).
  - `gitx.DiffStat(task.BaseSHA, branchHead)` → diff metrics; `gitx.CommitsOnRange` → subjects/trailers → `DetectGates`; `gitx.DiffNames` → `DetectArtifacts`.
  - returns `RawSignals` (no scoring here — `ScoreObjective` does that, keeping the seam thin per ARCH-PURE).
- [ ] **Step 4: Wire `grade` Stage A** — load RunRecord, for each agent `Measure` → `ScoreObjective` → set `AgentResult.Scorecard`; set `Status="graded"`; `Store.WriteRun`. (Stage B added in M4.)
- [ ] **Step 5: Run → PASS; build. Step 6: Commit** — `git commit -m "#119 M3: Measurer + sdlc bench grade (Stage A objective scorecard)"`

### Task 3.4: M3 milestone close

- [ ] Update `## Log`; tick M3. `sdlc milestone-close --issue 119 --milestone M3`; address findings; record verdict.

---

## Chunk 4: M4 — Grader Stage B (anonymizer + LLM judge + operator review) + `review`

**Milestone M4 review boundary** — closes with `sdlc milestone-close --issue 119 --milestone M4`.

### Task 4.1: `Scrub` + `BuildSubmissions` (pure) + the anonymization-leak gate

**Files:**
- Create: `cmd/sdlc/internal/bench/anonymize.go`
- Test: `cmd/sdlc/internal/bench/anonymize_test.go`

- [ ] **Step 1: Failing tests:**
  - `TestScrubRemovesIdentity` — `Scrub` strips agent names (case-insensitive `claude`/`codex`/`gemini`), `Co-Authored-By:` trailer lines, `bench/<agent>/…` branch strings, and a provided secrets list; assert none survive.
  - `TestBuildSubmissionsDeterministicShuffle` — same `seed` ⇒ same label→agent mapping; different seed ⇒ (usually) different; every agent gets exactly one label.
- [ ] **Step 2: Run, verify fail.**
- [ ] **Step 3: Implement** `anonymize.go`. Define the struct explicitly:

```go
type Submission struct {
	Label             string            // "A","B","C" (anonymized)
	Diff              string            // scrubbed unified diff
	Artifacts         map[string]string // scrubbed path→content (spec/plan/log/atlas)
	TranscriptSummary string            // scrubbed, summarized supplement
}
```

`Scrub(text string, secrets []string) string` = ordered literal + regex replacements (compile regexes once at package scope, DRY). It strips: agent names (`(?i)claude|codex|gemini`), `Co-Authored-By:` trailer lines, `bench/<agent>/…` branch strings, **and the caller's `secrets` list — which MUST include the contestant agents' identifying commit-message tells, not just trailers** (the diff carries commit subjects; style leaks there). `BuildSubmissions(results []AgentResult, artifacts map[string]map[string]string, seed int64)` builds the `[]Submission` with a deterministic Fisher–Yates over a seeded `math/rand.New(rand.NewSource(seed))` (seed injected by the caller from the RunID → pure + reproducible). Returns `([]Submission, map[string]string)` (label→agent).
- [ ] **Step 4: Run → PASS.**
- [ ] **Step 5: Add the leak-gate test** `TestAnonymizationLeakGate` (the acceptance gate from the spec) — **skipped by default** (`testing.Short()`), opt-in via an env flag because it dispatches a real judge. It builds two real-ish submissions (each with realistic, agent-styled commit subjects in the diff — so the test exercises the *style* leak, not just the `Co-Authored-By` trailer), scrubs them, and across **K=20 trials** asks the pinned judge "which submission was written by Claude?". **Assertion: the judge names the correct agent in ≤ 60% of trials** (12/20) — a falsifiable threshold, not "above chance" by vibe. If it exceeds 60%, the scrub is leaking and `secrets` needs widening (the gate drives that loop). Document it as the gate; in CI it runs as a manual/nightly check, not on every `go test`.
- [ ] **Step 6: Commit** — `git commit -m "#119 M4: Scrub + BuildSubmissions (seeded blind packets) + anonymization-leak gate"`

### Task 4.2: `JudgePrompt` + `ParseJudgeOutput` (pure)

**Files:**
- Modify: `cmd/sdlc/internal/bench/prompts.go` (add `JudgePrompt`)
- Create: `cmd/sdlc/internal/bench/judgeparse.go`
- Test: `cmd/sdlc/internal/bench/judgeparse_test.go`

- [ ] **Step 1: Failing tests** — `JudgePrompt(task, subs)` lists each Submission's label + content and specifies the **exact JSON output contract**; `ParseJudgeOutput` extracts the fenced ```json block and unmarshals into `JudgeVerdict` (tolerant of preamble/markdown), erroring cleanly on a missing block.
- [ ] **Step 2–4: Implement + pass.**

```go
type DimensionVerdict struct {
	Key        string   `json:"key"`
	Ranking    []string `json:"ranking"`    // labels best→worst, e.g. ["B","A"]
	Rationale  string   `json:"rationale"`
	Confidence string   `json:"confidence"` // high|medium|low
}
type JudgeVerdict struct {
	Dimensions []DimensionVerdict `json:"dimensions"`
}

// ParseJudgeOutput reuses extractJSONBlock (DRY) and json.Unmarshal.
func ParseJudgeOutput(output string) (JudgeVerdict, error) { /* extract + unmarshal */ }
```

The prompt embeds the output contract (mirrors the judge package's "first machine-read block" discipline, but a JSON block — **not** the binary `VERDICT` token, which can't carry rankings).

- [ ] **Step 5: Commit** — `git commit -m "#119 M4: JudgePrompt + ParseJudgeOutput (net-new structured Stage-B contract)"`

### Task 4.3: Stage B wiring in `grade` + operator review doc render (pure)

**Files:**
- Modify: `cmd/sdlc/bench.go` (`grade` Stage B)
- Create: `cmd/sdlc/internal/bench/review.go` (pure `RenderReviewDoc` + `ParseOperatorRankings` + `DeAnonymize`)
- Test: `cmd/sdlc/internal/bench/review_test.go`

- [ ] **Step 1: Failing tests** — `RenderReviewDoc(task, subs, dims)` produces a side-by-side markdown with the LLM verdict **withheld**, `🤖`-marker slots per dimension for operator rankings; `ParseOperatorRankings(text)` reads the operator's filled-in rankings back; `DeAnonymize(verdict, mapping)` rewrites labels→agent names.
- [ ] **Step 2–4: Implement + pass.** `grade` Stage B (IO): `BuildSubmissions` (seed from RunID) → `JudgePrompt` → `judge.Dispatch` with the **pinned judge model** (a `--judge-agent` flag, default `claude`, read-only `AllowedTools`, NOT a contestant) → `ParseJudgeOutput` → store anonymized verdict on the RunRecord + write the operator review doc via `Store`.
- [ ] **Step 5: Commit** — `git commit -m "#119 M4: grade Stage B (pinned blind judge) + operator review doc"`

### Task 4.4: `review` command (de-anonymize + merge verdicts)

**Files:**
- Modify: `cmd/sdlc/bench.go` (`review` child)
- Create: `cmd/sdlc/helptext/bench-review.md`
- Test: `cmd/sdlc/bench_review_test.go`

- [ ] **Step 1: Failing test** — given a RunRecord with an anonymized LLM verdict + a filled operator review doc, `runBenchReview` reveals the LLM verdict, `ParseOperatorRankings` + `DeAnonymize` both, folds them into the RunRecord (`Status="reviewed"`), and persists.
- [ ] **Step 2–4: Implement + pass.**
- [ ] **Step 5: Commit** — `git commit -m "#119 M4: sdlc bench review — de-anonymize + merge LLM/operator verdicts"`

### Task 4.5: M4 milestone close

- [ ] Update `## Log`; tick M4. `sdlc milestone-close --issue 119 --milestone M4`; address findings; record verdict.

---

## Chunk 5: M5 — `report`/`leaderboard`, end-to-end demo, atlas

**Milestone M5 review boundary** — closes with `sdlc milestone-close --issue 119 --milestone M5` (then `sdlc close`).

### Task 5.1: `Leaderboard` + `Aggregate` + `RenderLeaderboard` (pure)

**Files:**
- Create: `cmd/sdlc/internal/bench/leaderboard.go`
- Test: `cmd/sdlc/internal/bench/leaderboard_test.go`

- [ ] **Step 1: Failing test** — `Aggregate([]RunRecord)` produces per-`(agent,version)` objective-score means + per-dimension head-to-head win-rates (from the de-anonymized `Verdicts`); only `LeaderboardEligible && Status=="reviewed"` records count (live-mode excluded). `RenderLeaderboard` snapshot.
- [ ] **Step 2–4: Implement + pass.** Pure: input is `[]RunRecord`, output is `Leaderboard` struct + markdown. Win-rate = for each dimension, count pairwise (agent ahead of agent) across all records.
- [ ] **Step 5: Commit** — `git commit -m "#119 M5: Leaderboard Aggregate + render (eligible reviewed runs only)"`

### Task 5.2: `report` + `leaderboard` commands

**Files:**
- Modify: `cmd/sdlc/bench.go` (`report`, `leaderboard` children)
- Create: `cmd/sdlc/helptext/bench-report.md`, `bench-leaderboard.md`
- Test: `cmd/sdlc/bench_report_test.go`

- [ ] **Step 1: Failing test** — `report --run <id>` renders a single run's side-by-side (de-anonymized) summary; `leaderboard` reads all `runs/*.md`, `Aggregate`, writes `workshop/benchmarks/leaderboard.md`.
- [ ] **Step 2–4: Implement + pass. Step 5: Commit** — `git commit -m "#119 M5: sdlc bench report + leaderboard"`

### Task 5.3: End-to-end claude-vs-codex demo (manual verification)

- [ ] **Real run** (documented in `## Log`): `sdlc bench freeze --issue <small-real-issue>`; `sdlc bench run --task <id> --agents claude,codex --mode autonomous`; `sdlc bench grade --run <id>`; fill the operator review doc; `sdlc bench review --run <id>`; `sdlc bench leaderboard`.
- [ ] **Verify:** both worktrees branched from `base_sha` and never merged (`git log main` unchanged); the run record carries both scorecards + both verdicts; the leaderboard shows a `(claude,version)` vs `(codex,version)` row. Capture the artifacts paths in `## Log` as the close evidence.
- [ ] **Pick the demo issue deliberately small** (completes inside the autonomous budget) — this is the mitigation for the headless-completion risk. Note the chosen issue + observed wall-clock in `## Log`.

### Task 5.4: Atlas + docs

**Files:**
- Create: `atlas/workflow/bench.md` (the harness's surface/flow/terminology)
- Modify: `atlas/index.md` (link the new file)

- [ ] Map the `freeze→run→grade→review→leaderboard` flow, the pure-core/IO-shell split, the `workshop/benchmarks/` layout, and the deferred seams (interactive/live, reference backfill, nested-dispatch decision). Map, don't over-specify.
- [ ] **Commit** — `git commit -m "#119 M5: atlas — bench harness surface + flow"`

### Task 5.5: Issue close

- [ ] Update `## Log` with the end-to-end evidence + the nested-dispatch resolution (whether sub-dispatched SDLC reviews were disabled/pinned during the demo run). Tick M5.
- [ ] `sdlc milestone-close --issue 119 --milestone M5`; address findings.
- [ ] `sdlc actual --issue 119` then `sdlc close --issue 119 --verified '<end-to-end evidence + paths>'` (measured actual, atlas updated).

---

## Notes for the executor

- **Reuse, don't rebuild** (`ARCH-DRY`): `judge.Dispatch`/`judge.Run` (dispatch + test seam), `judge.AgentCLI`, `gitx.run`/`RunGit`/`Capture`, `issue.Parse`/`Compose`/`GetField`/`SetField`/`SectionBody`/`CountPlanItems`, the `close.go` log-insertion pattern. `extractJSONBlock` is the one shared block-parser across Task + RunRecord + judge output.
- **Purity boundary** (`ARCH-PURE`): nothing in `task.go`/`rubric.go`/`runrecord.go`/`scorecard.go`/`detect.go`/`anonymize.go`/`judgeparse.go`/`leaderboard.go`/`prompts.go`/`review.go` may call `os`/`exec`/`time`/`git`. All clock/IO is injected (`Now`, `headSHA`, `measureRun`, the Store, the Runner). If a "pure" test needs a mock, the logic is in the wrong layer — move the IO out.
- **First-of-its-kind Dispatch usage:** autonomous `RunAgent` is the first `judge.Dispatch` caller to pass a write `AllowedTools` AND a real `context.WithTimeout`. Keep the budget a flag (`--budget`, default 30m).
- **Deferred (do NOT build now):** interactive/live responder impls, reference backfill, the nested-dispatch resolution (decide + document at the M5 demo). The seams exist; wiring is out of scope for #119 day-one.
- **Token-cost metric is deliberately out for day-one** (the spec lists it; `Metrics` carries wall-clock/turns/diff/commits instead). Cross-agent token accounting isn't symmetric across CLIs — consistent with the spec's "effort = wall-clock/turns" note. Log this as a *decision* in `## Log` at M3, not an omission; `Metrics` can widen later.
