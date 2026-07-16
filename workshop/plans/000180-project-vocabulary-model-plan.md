# Project Vocabulary Model Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Lift the `project` noun from prose-only convention to a formally modeled,
gate-enforced vocabulary — cue model + `pkg/vocab` binding + typed parsing +
conformance gates + a `sdlc project` verb family (lifecycle transitions, derived
kanban board, retro, calibrated close).

**Architecture:** Mirror the `issue` noun end to end (ariadne#122's pattern):
`construct/vocabulary/project.cue` is the single source; `vocabulary export` emits
JSON; `pkg/vocab/project.go` embeds it; every consumer (verbs, gates, helptext)
derives from the model (ARCH-PURPOSE). Pure core / thin IO shell throughout
(ARCH-PURE): parsing, board computation, and guards are pure functions over a
typed `Doc`; git/fs/peer-repo reads stay in thin injected seams.

**Tech Stack:** CUE (build-time only), Go (cobra CLI, embedded JSON), existing
`sdlc` internal packages (`internal/issue` frontmatter helpers, `resolve.go`
peer-repo machinery).

**Issue:** workshop/issues/000180-… — the `## Spec` there is the converged design
input. Milestones M1–M5 below are review boundaries (each ends in
`sdlc milestone-close --issue 180 --milestone Mx`).

---

## Chunk 1: Design overview + core concepts

### Scope boundary vs #171 (read this first)

#180 lands the **model and machinery**; #171 lands the **residency migration and
close-gate lift** and CONSUMES this model. Concretely, in scope here:

- `project.cue` declares the per-repo home `workshop/projects/` — but the repo has
  no project instances yet. The conformance gate (M2) is wired and tested but
  dormant until files exist (dogfood is deferred by operator decision 2026-07-15).
- `sdlc close`'s project gate keeps its brain residency (`--brain-dir`,
  `data/project/*.md` glob in `FindByIssueRef`) — #171 relifts that to cross-peer
  `workshop/projects/` resolution. #180 only retypes the *parsing* under it.
- Brain's 5 legacy files use `status: active` and stay untouched; the
  `active → executing` mapping happens in #171's migration. Nothing in #180 may
  hard-fail on a legacy status (warn only).
- Cross-repo discovery: **the cue declares the per-repo home; resolution owns the
  walk** (settled leaning, matches how `resolveRepoDir` works). No fleet-glob
  encoding in the model.

### Design decisions (with ARCH citations)

1. **Lifecycle funnel** `ideation → defined → committed → executing → done|dropped`
   (+`paused` beside executing), from the Spec. Categories are chosen so Go
   predicates fall out: `forming` (pre-baseline), `committed` (baseline set, not
   broken down), `executing` (live portfolio, incl. paused), `terminal`.
2. **Guards are NAMED in the model, implemented in a Go registry** keyed by name
   (`internal/project/guards.go`). `sdlc project set-status` resolves the matched
   transition's guard list against the registry; an unknown guard name is a
   refusal (model↔code drift is caught at run time, not silently ignored).
   ARCH-PURPOSE: the model is enforced, not documentation.
3. **`deadline:` + `planned_finish:` are the commit-time baseline**, compiled into
   the cue as a conditional requirement (mirrors issue.cue's `actual_hours!`
   compiled guard): any post-commit status except `dropped` requires both.
4. **Kanban split**: baseline stored in `## Breakdown`, progression DERIVED by
   `sdlc project status` (pure `computeBoard` over live cross-repo issue
   frontmatter, ARCH-PURE), re-forecasts appended to `## Log`. No stored board.
5. **`ArchiveSubdirs` widens kind-keyed** (#181 close-review note: scales better
   than a third return): `ArchiveSubdir(root, kind)` with typed kinds; the
   two-return form is deleted and its 9 call sites migrated (ARCH-DRY: one
   derivation point, guard test updated).
6. **Reuse over new code** (ARCH-DRY): frontmatter via `internal/issue`
   `Parse/Compose/GetField/SetField`; ref grammar via the existing `parseRef` in
   `cmd/sdlc/resolve.go` (internal/project returns raw `RefText`; package main
   parses); peer-repo lookup via `resolveRepoDir`; shared lifecycle predicates
   extracted in `pkg/vocab` — the third noun stops the existing `inCategory`
   duplication (vocab.go + verdict.go) and pre-empts triplicating the
   lifecycle predicates.
7. **Which gate owns instance conformance** (Done-when design decision): the
   fail-closed validate gate at `push`/`merge` (same class as issues) — generalize
   `validategate.go` to a noun table. Plus an on-demand `sdlc project validate`.
8. **Phase-A estimation** is a method doc + structured `## Estimate` fields; the
   fog-factor ledger lives beside issue calibration in brain
   (`~/workspace/brain/data/life/42shots/velocity/`) — calibration data is
   explicitly brain-resident (session decision 2026-07-15).
9. **Project close is a dedicated verb** (`sdlc project close`), not a set-status
   edge: it owns retro gate + fog-factor ledger row + archive-to-history + the
   `executing→done` flip. `set-status --to done` refuses and points at it
   (mirrors how issue close/claim own their fixed transitions). **Paused
   projects must resume before closing**: the model deliberately has no
   `paused→done` edge, so `project close` requires `status == "executing"`
   exactly and refuses paused with a "resume first" pointer; `--drop` works
   from both executing and paused (both edges exist, retro-gated).

### Core concepts

#### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `project` noun model (cue) | `construct/vocabulary/project.cue` | new |
| `ProjectModel` | `pkg/vocab/project.go` | new |
| shared lifecycle helpers | `pkg/vocab/lifecycle.go` | new |
| `ArchiveSubdir` (kind-keyed) | `pkg/vocab/vocab.go` | modified |
| `Doc` / `Task` (typed project file) | `cmd/sdlc/internal/project/doc.go` | new |
| YAML-semantic workflow metadata decoder | `cmd/sdlc/internal/project/metadata.go` | new |
| Phase-A tri-state parser | `cmd/sdlc/internal/project/phasea.go` | new |
| `ScaffoldSpec` / `RenderScaffold` | `cmd/sdlc/internal/project/scaffold.go` | new |
| `Summary` renderers | `cmd/sdlc/internal/project/summary.go` | new |
| project slug/path resolver | `cmd/sdlc/internal/project/path.go` | new |
| Tick mutators (typed re-impl) | `cmd/sdlc/internal/project/project.go` | modified |
| Guard registry | `cmd/sdlc/internal/project/guards.go` | new |
| `computeBoard` | `cmd/sdlc/projectstatus.go` | new |
| datatype prose (demoted to cite cue) | `construct/datatype/project.md` | planned M5 |
| prose↔model drift test | `pkg/vocab/prose_drift_test.go` | planned M5 |

- **`project` noun model** — categories/when/lifecycle/laws/discovery/scaffold +
  `#Project` frontmatter shape, mirroring issue.cue.
  - **Relationships:** 1:1 with `ProjectModel` (generate-time export); referenced
    by `vocabulary validate-instance --type project` (auto-registered by
    filename — no Go edit in cmd/vocabulary).
  - **DRY rationale:** replaces the hand-maintained status table in
    `construct/datatype/project.md` (the exact ARCH-PURPOSE gap the issue lift
    closed for issues).
  - **Future extensions:** `product`/`roadmap` nouns follow the same template
    (explicitly out of scope, own tickets).
- **`ProjectModel`** — embedded JSON binding + predicates
  (`IsTerminal/IsExecuting/IsForming`, `CanTransition`, `LegalTransitions`,
  `TransitionFor`, `InitialStatus`, `Sections`, `Discovery`,
  `RenderLifecycleHelp`). Unit tests need no IO (embed is compile-time).
  - **Relationships:** N consumers (verbs, gates, helptext render) → 1 model.
  - **DRY rationale:** third noun forces extraction of shared helpers
    (`lifecycle.go`) that vocab.go/verdict.go currently duplicate.
- **`Doc`/`Task`** — a parsed project file: frontmatter fields (via
  `internal/issue` helpers), typed task rows (`State`, `Title`, `RefText`,
  `LineIdx`), section bodies. Line-preserving render (mutations replace lines,
  never reflow) so untouched bytes are stable.
  - **Relationships:** 1:1 with a project file on disk; consumed by tick
    mutators, guards, `computeBoard`, and the verbs.
  - **DRY rationale:** replaces substring-convention parsing (lessons.md #167 /
    Done-when); one parser feeds close-gate mutation AND the new verbs.
  - **Future extensions:** detail-block typing (the `UpsertDetailBlockFields`
    machinery keeps its current field-level form for now — it is already
    field-typed and battle-tested; retyping it adds churn without a consumer).
- **Guard registry** — `map[string]GuardFunc` over `(Doc, GuardCtx)`; pure
  (evidence + today injected via `GuardCtx`).
- **`computeBoard`** — pure roll-up: task states + per-ref issue meta → done/total,
  remaining Σ estimate hours, days to deadline, blocked list. Issue lookups
  injected as `func(refText string) (issueMeta, error)`.

#### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `vocabulary export/vet` pipeline | `construct/vocabulary/vet_test.sh`, `Makefile.workflow vocab-embed` | modified | cue CLI |
| validate gate (noun table) | `cmd/sdlc/validategate.go` | modified | `vocabulary validate-instance` |
| `sdlc project` verb family | `cmd/sdlc/project.go` (+ `helptext/project.md`) | new | fs + git commit + peer repos |
| issue lookup seam | `cmd/sdlc/projectstatus.go` | new | peer-repo fs reads via `resolveRepoDir` |
| fog-factor ledger append | `cmd/sdlc/projectclose.go` | new | brain fs (sibling repo) |
| Phase-A estimation method (process doc) | `brain/data/life/42shots/velocity/estimate-logic-project-v1.md` | new | human-executed sizing method + calibration ledger |
| close-gate project parsing | `cmd/sdlc/close.go:565-655` (call sites unchanged) | modified | brain project files |

- **validate gate** — generalize `validateChangedIssues` into a per-noun table
  (`{noun, dir, checkSections}`); `--type project` fires on changed
  `workshop/projects/*.md`. Injected seams already exist (`validategate.go:30-35`)
  so hermetic tests need no git/binary.
- **`sdlc project` verbs** — mutating subcommands wrap `markMutatingCommand`
  (mirrors `issue.go:192`) and go through the repo transaction lock like every
  mutating verb.
- **issue lookup seam** — resolves `repo#id` via `parseRef` + `resolveRepoDir`,
  reads issue frontmatter (`status`, `estimate_hours`, `actual_hours`, `deps`).
  Injected into `computeBoard` so board tests use a map-backed fake; a
  hermetic-repo test (existing `hermeticrepo_test.go` pattern) covers the real
  seam with a sibling fixture repo.
- **fog-factor ledger append** — appends one markdown table row; `--brain-dir`
  default `../brain` (same convention as `close.go:140`), `--no-ledger` bypass.

### Verification strategy (threads through every milestone)

- Every pure entity: colocated unit tests, no IO mocks.
- Gates/verbs: hermetic repo tests (existing `hermeticrepo_test.go` pattern).
- `make vocab-embed` green (embed not stale) at every milestone close.
- Live dogfood before M4-close: run the verb family end-to-end against a scratch
  fixture repo with a symlinked cwd (lessons.md 2026-07-15: `$PWD` symlink guard;
  budget a real-fixture pass into every IO-adjacent milestone).
- Full suite `go test ./...` run BARE (never piped — lessons.md 2026-07-15) before
  every commit.

---

## Chunk 2: M1 (model + binding) and M2 (typed parsing + conformance gate)

### M1 — the model and its Go binding

#### Task M1.1: `construct/vocabulary/project.cue`

**Files:**
- Create: `construct/vocabulary/project.cue`
- Create: `construct/vocabulary/testdata/project_invalid.cue`
- Modify: `construct/vocabulary/vet_test.sh`

- [ ] **Step 1: Write the model** — full content:

```cue
// project — the vocabulary of a project: its data shape, its lifecycle funnel,
// and the laws the funnel must satisfy. Single source of truth for the `project`
// noun (ariadne#180). sdlc reads the exported JSON; humans and the LLM read this
// file directly. The prose companion (construct/datatype/project.md) cites this
// file as schema authority — a drift test binds the two.
//
// The organizing insight (#180 Spec): the project lifecycle is the issue
// lifecycle one level up. A project is a structured, TIME-BOUND push for a major
// change, across repos — not merely a container of issues; it carries a deadline
// set at commit.
package project

import "list"

// ── categories: the single concrete source of status membership ──
// forming   = pre-baseline (no deadline yet)
// committed = baseline set (deadline + planned finish), not yet broken down
// executing = broken down, live portfolio (paused keeps its baseline)
// terminal  = closed
categories: {
	forming:   ["ideation", "defined"]
	committed: ["committed"]
	executing: ["executing", "paused"]
	terminal:  ["done", "dropped"]
}

#Forming:   or(categories.forming)
#Terminal:  or(categories.terminal)
#Status:    or(list.Concat([categories.forming, categories.committed, categories.executing, categories.terminal]))

// ── when: one-line semantics per status (the documented-value source) ──
when: {
	ideation:  "idea captured; PRD not yet written (ideation lives in parley, linked via sources)"
	defined:   "PRD exists in the project file; not yet committed to a timeline"
	committed: "baseline set (deadline + planned finish + parallelism intent); not yet broken down"
	executing: "PRD broken down into issues across repos; work in flight"
	paused:    "execution suspended; committed baseline stays intact"
	done:      "done_when met; retro + fog-factor ledger row recorded"
	dropped:   "no longer worth pursuing"
}

// ── discovery: where instances of this noun live, PER REPO. The cue declares
// the per-repo home; cross-repo resolution owns the walk across peers (#171) —
// same division of labor as resolveRepoDir. Repo-relative. ──
discovery: {
	home: "workshop/projects" // repo-relative home for project instances
	glob: "*.md"
	// archive: done/dropped projects move under here (per-kind subdir derived in
	// Go by pkg/vocab.ArchiveSubdir — projects land in <archive>/projects, the
	// #181 layout; operator decision 2026-07-15: archive, don't stay in place).
	archive: "workshop/history"
}

// ── scaffold: the on-disk creation template `sdlc project new` writes. The
// fractal file: sections grow through the gated stages (PRD at define,
// Estimate at commit, Breakdown at breakdown, Log throughout). ──
#ScaffoldSection: {
	name:  string
	seed?: string
}
scaffold: sections: [...#ScaffoldSection] & [
	{name: "PRD"},
	{name: "Estimate"},
	{name: "Breakdown", seed: "- [ ]"},
	{name: "Log"},
]

// ── #Project: the data shape of a project record ──
#Project: {
	type:      "project"
	name:      string // slug; matches filename without .md
	goal:      string // one sentence: why this project exists
	done_when: string // the MVP boundary, falsifiable
	status:    #Status
	// The commit-time baseline (the time-bound attribute distinguishing project
	// from an issue container). Optional pre-commit; compiled-required after.
	// YAML date literals decode as strings (#124 lesson: accept what real
	// frontmatter parses to, don't self-vet only).
	deadline?:       string | null
	planned_finish?: string | null
	operator?:       string | null
	// issue refs ("repo#id"); the MVP commitment. explicitly_out is the
	// load-bearing half of the scoping conversation.
	mvp_scope?:      [...string] | null
	explicitly_out?: [...string] | null
	// compiled guard: every post-commit status except dropped carries the
	// baseline (a dropped project may have died pre-commit).
	if status == "committed" || status == "executing" || status == "paused" || status == "done" {
		deadline!:       string
		planned_finish!: string
	}
	// OPEN (#124 precedent): allow organically-growing frontmatter (created/
	// updated/sources/…) so instance conformance doesn't false-positive on a
	// valid-but-unmodeled field.
	...
}

// ── lifecycle: the transition table (the verbs). Guards are NAMED here; their
// implementations live in sdlc's guard registry (internal/project/guards.go),
// which refuses transitions naming a guard it doesn't implement. ──
#Transition: {
	from:   #Status
	to:     #Status
	event:  string
	guards: [...string]
}

lifecycle: [...#Transition] & [
	// the funnel
	{from: "ideation", to: "defined", event: "define", guards: ["prd-present"]},
	{from: "defined", to: "committed", event: "commit", guards: ["phase-a-estimate", "baseline-set", "reality-check"]},
	{from: "committed", to: "executing", event: "breakdown", guards: ["issues-cover-prd"]},
	// close is a dedicated verb (`sdlc project close`) owning retro + ledger +
	// archive; set-status refuses →done and points at it.
	{from: "executing", to: "done", event: "close", guards: ["retro-recorded", "fog-factor-recorded"]},
	// pause/resume (baseline survives)
	{from: "executing", to: "paused", event: "pause"},
	{from: "paused", to: "executing", event: "resume"},
	// drop at any pre-terminal stage; once executing, a retro is owed
	{from: "ideation", to: "dropped", event: "drop"},
	{from: "defined", to: "dropped", event: "drop"},
	{from: "committed", to: "dropped", event: "drop"},
	{from: "executing", to: "dropped", event: "drop", guards: ["retro-recorded"]},
	{from: "paused", to: "dropped", event: "drop", guards: ["retro-recorded"]},
]

// ── laws: named assertions the graph shape doesn't already guarantee ──
_froms: [for t in lifecycle {t.from}]
_tos: [for t in lifecycle {t.to}]

laws: {
	// every status carries a non-empty `when`
	"documented-value": {
		for s in list.Concat([categories.forming, categories.committed, categories.executing, categories.terminal]) {
			(s): when[s] & !=""
		}
	}
	// every non-initial status is reachable (ideation is the entry point)
	"reachable": {
		for s in list.Concat([["defined"], categories.committed, categories.executing, categories.terminal]) {
			(s): list.Contains(_tos, s) & true
		}
	}
	// every non-terminal status is escapable
	"escapable": {
		for s in list.Concat([categories.forming, categories.committed, categories.executing]) {
			(s): list.Contains(_froms, s) & true
		}
	}
}
```

- [ ] **Step 2: Write the invalid-model fixture** `testdata/project_invalid.cue` —
  it must be SELF-CONTAINED (its own package; a standalone file referencing
  `#Transition`/`#Status` from project.cue would fail vet with "reference not
  found" — a vacuous pass proving nothing; `testdata/issue_invalid.cue` is
  likewise self-contained — own package, minimal inline `when`+law construction,
  though its trip mechanism differs). Copy `categories`, the `#Status`/`#Transition`
  definitions, and the `lifecycle` list from project.cue verbatim, then corrupt
  ONE edge so the failure is genuinely the enum conflict:

```cue
// project_invalid: a broken model — one lifecycle edge targets a status outside
// every category. vet_test.sh asserts `cue vet` REJECTS this file; if it ever
// passes, the model's own constraints have stopped biting.
package projectinvalid

import "list"

categories: { /* verbatim copy from project.cue */ }
#Status: or(list.Concat([categories.forming, categories.committed, categories.executing, categories.terminal]))
#Transition: { from: #Status, to: #Status, event: string, guards: [...string] }

lifecycle: [...#Transition] & [
	{from: "ideation", to: "shipped", event: "define", guards: []}, // "shipped" ∉ #Status → vet must fail HERE
]
```

  Verify the failure message names the `to:` conflict (run
  `cue vet testdata/project_invalid.cue` by hand once) — not a missing
  reference.

- [ ] **Step 3: Extend `vet_test.sh`** — append a project block mirroring the
  issue block (`vet_test.sh:9-18`): `cue vet project.cue` passes;
  `cue vet testdata/project_invalid.cue` fails; export carries `"categories"`,
  `"lifecycle"`, and `"discovery"`.

- [ ] **Step 4: Run** `sh construct/vocabulary/vet_test.sh` — expect `ok`.

- [ ] **Step 5: Commit** — `#180 M1: project.cue — the project noun modeled (funnel, baseline guard, discovery, scaffold, laws)`

#### Task M1.2: shared lifecycle helpers in `pkg/vocab` (DRY refactor, behavior-preserving)

**Files:**
- Create: `pkg/vocab/lifecycle.go`, `pkg/vocab/lifecycle_test.go`
- Modify: `pkg/vocab/vocab.go`, `pkg/vocab/verdict.go`

- [ ] **Step 1: Extract** unexported helpers (a third noun is the trigger —
  vocab.go and verdict.go already duplicate `inCategory`):

```go
// pkg/vocab/lifecycle.go
package vocab

// inCat reports whether s is a member of cats[cat]. Shared by every noun model.
func inCat(cats map[string][]string, cat, s string) bool

// canTransition reports whether l declares a from→to edge.
func canTransition(l []Transition, from, to string) bool

// legalTransitions returns from's legal targets in lifecycle order, de-duplicated.
func legalTransitions(l []Transition, from string) []string

// allStatuses concatenates cats in the given category order.
func allStatuses(cats map[string][]string, order []string) []string

// renderLifecycleHelp renders STATUSES + LEGAL TRANSITIONS for any noun
// (statuses in category order, when-gloss, legal edges). Pure.
func renderLifecycleHelp(statuses []string, when map[string]string, l []Transition) string
```

- [ ] **Step 2: Delegate** — rewrite `IssueModel.inCategory/CanTransition/
  LegalTransitions/AllStatuses/RenderLifecycleHelp` and `VerdictModel.inCategory`
  as thin calls into the helpers (issue's category order literal
  `{"open","active","terminal"}` moves to a package const). No exported-surface
  change.

- [ ] **Step 3: Run** `go test ./pkg/vocab/` — the existing pins
  (`vocab_test.go`, `conformance_test.go`) must pass unchanged. Expected: PASS.

- [ ] **Step 4: Commit** — `#180 M1: pkg/vocab — extract shared lifecycle helpers (third noun incoming)`

#### Task M1.3: `ProjectModel` binding + embed

**Files:**
- Create: `pkg/vocab/project.go`, `pkg/vocab/project.json` (generated)
- Test: extend `pkg/vocab/vocab_test.go`, `pkg/vocab/conformance_test.go`

- [ ] **Step 1: Write failing conformance test** — `TestProjectConformance` in
  `conformance_test.go`, mirroring `TestIssueConformance` (:9-41) exactly but
  over `Project()`: every status in exactly one category with non-empty `when`;
  every transition references known statuses and agrees with `CanTransition`.
  Plus pins in `vocab_test.go`:

```go
func TestProjectDiscovery(t *testing.T) {
	d := Project().Discovery()
	if d.Home != "workshop/projects" || d.Glob != "*.md" || d.Archive != "workshop/history" {
		t.Fatalf("discovery = %+v", d)
	}
}
func TestProjectInitialStatus(t *testing.T) {
	if got := Project().InitialStatus(); got != "ideation" {
		t.Fatalf("initial = %q", got)
	}
}
func TestProjectSetStatusRefusesDone(t *testing.T) { // TransitionFor surface
	tr := Project().TransitionFor("executing", "done")
	if tr == nil || tr.Event != "close" {
		t.Fatalf("executing→done = %+v", tr)
	}
}
```

- [ ] **Step 2: Run** `go test ./pkg/vocab/ -run 'Project'` — expect FAIL
  (undefined `Project`).

- [ ] **Step 3: Implement `pkg/vocab/project.go`** — mirror vocab.go's shape:

```go
//go:generate sh -c "vocabulary export --noun project > project.json"

//go:embed project.json
var projectJSON []byte

// ProjectModel is the read-only, parsed `project` noun (ariadne#180): the
// lifecycle funnel, per-status semantics, discovery, and creation scaffold.
// Derived from construct/vocabulary/project.cue at generate time; never
// hand-edited. Category order: forming → committed → executing → terminal.
type ProjectModel struct {
	Categories map[string][]string `json:"categories"`
	When       map[string]string   `json:"when"`
	Disc       Discovery           `json:"discovery"` // Plans stays empty: projects have no plan sidecars
	Lifecycle  []Transition        `json:"lifecycle"`
	Scaf       Scaffold            `json:"scaffold"`
}

func Project() *ProjectModel
func (m *ProjectModel) Discovery() Discovery
func (m *ProjectModel) Sections() []Section
func (m *ProjectModel) InitialStatus() string // categories.forming[0]
func (m *ProjectModel) IsTerminal(s string) bool
func (m *ProjectModel) IsExecuting(s string) bool // executing category (incl. paused)
func (m *ProjectModel) IsForming(s string) bool
func (m *ProjectModel) AllStatuses() []string
func (m *ProjectModel) CanTransition(from, to string) bool
func (m *ProjectModel) LegalTransitions(from string) []string
// TransitionFor returns the declared from→to edge (nil if none) — the guard
// runner's lookup surface.
func (m *ProjectModel) TransitionFor(from, to string) *Transition
func (m *ProjectModel) RenderLifecycleHelp() string
func (m *ProjectModel) StatusNames(sep string) string
```

  All predicates delegate to the M1.2 helpers. Also add `TransitionFor` to
  `IssueModel`? No — YAGNI; no issue consumer needs it.

- [ ] **Step 4: Generate the embed** — `make vocabulary-build` then
  `make vocab-embed` (runs `go generate ./pkg/vocab/...` + stale-diff check).
  Expected: `pkg/vocab/project.json` created, diff-clean afterward.

- [ ] **Step 5: Run** `go test ./pkg/vocab/` — expect PASS.

- [ ] **Step 6: Commit** — `#180 M1: pkg/vocab.Project() — embedded binding for the project noun`

#### Task M1.4: kind-keyed `ArchiveSubdir`

**Files:**
- Modify: `pkg/vocab/vocab.go:96-108`, and non-test call sites:
  `cmd/sdlc/resolve.go:232`, `cmd/sdlc/merge.go:622-623`, `cmd/sdlc/state.go:299`,
  `cmd/sdlc/push.go:294-295,490,596`, `cmd/sdlc/internal/issue/scaffold.go:32`
- Modify test call sites (they break the compile when the two-return form goes):
  `pkg/vocab/vocab_test.go:144-159` (`TestArchiveSubdirs`, the layout pin —
  rewrite over the kind-keyed form), `cmd/sdlc/resolve_test.go:408`
- Modify: `pkg/vocab/vocab_test.go:166-207` (`TestArchiveSubdirs_SingleDerivationPoint`)
- Modify: `construct/vocabulary/issue.cue:48-52` comment (`ArchiveSubdirs` →
  `ArchiveSubdir` name only; the "will widen" sentence lives in
  `vocab.go:104-105` and dies with Step 3's replacement comment)

- [ ] **Step 1: Write failing test** — in `vocab_test.go`:

```go
func TestArchiveSubdirKinds(t *testing.T) {
	for kind, want := range map[ArchiveKind]string{
		ArchiveIssues: "h/issues", ArchivePlans: "h/plans", ArchiveProjects: "h/projects",
	} {
		if got := ArchiveSubdir("h", kind); got != filepath.FromSlash(want) {
			t.Fatalf("%s = %q, want %q", kind, got, want)
		}
	}
}
```

- [ ] **Step 2: Run** `go test ./pkg/vocab/ -run ArchiveSubdirKinds` — FAIL
  (undefined).

- [ ] **Step 3: Implement** in vocab.go, replacing `ArchiveSubdirs`:

```go
// ArchiveKind names one per-kind archive subdir under the archive root (#181
// layout, widened kind-keyed for #180). THE single derivation point — consumers
// route through ArchiveSubdir; nothing concatenates these literals elsewhere
// (the source-scan guard test enforces this).
type ArchiveKind string

const (
	ArchiveIssues   ArchiveKind = "issues"
	ArchivePlans    ArchiveKind = "plans"
	ArchiveProjects ArchiveKind = "projects"
)

// ArchiveSubdir derives one kind's archive dir from an archive ROOT. Go-owned
// rather than cue-encoded for the same reasons as before (#181): writers derive
// from --history-dir overrides; widening the cue's `archive` string would break
// downstream JSON consumers. Reads stay tolerant of the pre-#181 flat layout.
func ArchiveSubdir(root string, kind ArchiveKind) string {
	return filepath.Join(root, string(kind))
}
```

- [ ] **Step 4: Migrate the 9 call sites** — mechanical, e.g.
  `issuesSub, plansSub := vocab.ArchiveSubdirs(root)` →
  `issuesSub := vocab.ArchiveSubdir(root, vocab.ArchiveIssues)` (+ plans line
  where the second value was used; several sites discard one value — those
  become a single call). Delete `ArchiveSubdirs`.

- [ ] **Step 5: Update the guard test** — extend
  `TestArchiveSubdirs_SingleDerivationPoint`'s source scan to also forbid
  hand-concatenated `"projects"` archive paths and to accept the new call form.

- [ ] **Step 6: Run** `go test ./pkg/vocab/ ./cmd/...` — expect PASS.

- [ ] **Step 7: Commit** — `#180 M1: ArchiveSubdir goes kind-keyed (+projects) — 9 call sites migrated`

#### Task M1.5: close M1

- [ ] Run full suite bare: `go test ./...` and `sh construct/vocabulary/vet_test.sh` — PASS.
- [ ] Tick M1 rows in the issue `## Plan`, log the boundary in `## Log`.
- [ ] `sdlc milestone-close --issue 180 --milestone M1` (binary dispatches the
  mandatory fresh-eyes review; fix Critical/Important before crossing).

### M2 — typed project parsing + the conformance gate

#### Task M2.1: typed `Doc`/`Task` parser

**Files:**
- Create: `cmd/sdlc/internal/project/doc.go`, `doc_test.go`

- [x] **Step 1: Write failing tests** — parse a fixture with frontmatter + all
  four sections + task rows in every checkbox state
  (`[ ]`,`[x]`,`[.]`,`[-]`,`[~]`), refs with and without milestones, and a
  plain-text task (no ref):

```go
func TestParseDoc(t *testing.T) {
	d, err := project.ParseDoc(fixture)
	// d.FM("status") == "executing"; len(d.Tasks) == 5
	// d.Tasks[0] == Task{LineIdx: …, State: ' ', Title: "provider interface skeleton", RefText: "charon#13 M1"}
	// d.SectionBody("PRD") non-empty; d.SectionBody("Estimate") == "" for absent section
}
func TestDocRenderRoundTrip(t *testing.T) {
	// ParseDoc(x).Render() == x byte-for-byte when nothing was mutated
}
func TestDocSetTaskState(t *testing.T) {
	// SetTaskState(i, 'x') rewrites ONLY that line; Render diff is 1 line
}
```

- [x] **Step 2: Run** — FAIL (undefined).

- [x] **Step 3: Implement**:

```go
// Doc is a parsed project file: raw lines (render source of truth), frontmatter
// (via internal/issue's Parse/GetField — same delimiter grammar), typed task
// rows, and section spans. Mutations edit lines in place; Render never reflows.
type Doc struct {
	lines []string
	fm    string // raw frontmatter block
	Tasks []Task
	// sections: name → [start,end) line span of the body under "## <name>"
}

type Task struct {
	LineIdx int
	State   byte   // ' ', 'x', '.', '-', '~'
	Title   string
	RefText string // raw "[repo#id Mx]" innards, "" for plain-text tasks; the
	               // ref GRAMMAR stays owned by cmd/sdlc's parseRef (ARCH-DRY)
}

func ParseDoc(text string) (*Doc, error)
func (d *Doc) FM(field string) string          // via issue.GetField
func (d *Doc) SetFM(field, value string)       // via issue.SetField
func (d *Doc) SectionBody(name string) string
func (d *Doc) AppendToSection(name, block string) error // for Log appends
func (d *Doc) SetTaskState(i int, state byte)
func (d *Doc) Render() string
```

  Task-row grammar (one regex, compiled once):
  `^- \[([ x.\-~])\] (.*)$` for the row, then the ref = the LAST
  `\[([^\]]+)\]` group within the remainder (regexp `FindAllStringSubmatchIndex`,
  take the final match); no bracketed group → plain-text task. Trailing text
  after the ref is legal and preserved (matches the current mutators'
  unanchored behavior, modulo the one accepted narrowing pinned in M2.2
  Step 1). The class includes `~` so the milestone tick can flip it (the
  issue-close tick filters it out by State, per the pinned asymmetry).

- [x] **Step 4: Run** `go test ./cmd/sdlc/internal/project/` — PASS.

- [x] **Step 5: Commit** — `#180 M2: internal/project.Doc — typed project-file records`

#### Task M2.2: retype the tick mutators over `Doc`

**Files:**
- Modify: `cmd/sdlc/internal/project/project.go:62-95` (`TickMilestoneTaskRow`,
  `TickAllTaskRowsForIssue`)

- [x] **Step 1: Note the pin** — the existing `project_test.go` cases define the
  behavior contract, including the intentional asymmetry: the milestone tick
  flips states `[ .\-~]`, the issue-close tick flips only `[ .]` (so
  cancelled/blocked rows aren't silently completed at issue close). Do NOT
  change signatures — `(text string, …) (string, int)` in, out — so `close.go`
  call sites stay untouched this milestone. One behavior the pins DON'T cover:
  today's regexes have no end-of-line anchor, so a row with trailing text after
  the ref (`- [ ] thing [charon#13 M1] (note)`) ticks. That behavior is KEPT
  deliberately — the Doc grammar below treats the LAST bracketed group on the
  line as the ref (trailing text allowed and preserved) — and Step 1 adds a
  pinned test for exactly this case before the reimplementation. One accepted
  narrowing: a row whose post-ref trailing text itself contains a bracketed
  group (`… [ariadne#31 M1] (see [notes](url))`) ticks under today's unanchored
  regexes but is skipped under the last-bracketed-group grammar. The convention
  puts the ref last (`FindTaskTitle` already assumes it), so this is accepted
  deliberately — pin the skip behavior with its own test case so the narrowing
  is visible, not incidental.

- [x] **Step 2: Reimplement bodies** — parse via `ParseDoc`, select tasks by
  matching `RefText` against `repo#id [milestone]` (exact-match on the parsed
  ref text: `RefText == repoName+"#"+issueID` or
  `strings.HasPrefix(RefText, repoName+"#"+issueID+" ")` for the all-rows form;
  `RefText == repoName+"#"+issueID+" "+milestone` for the milestone form),
  filter by current `State` per the pinned character classes, `SetTaskState`,
  `Render`. Legacy `~` state: M2.1's grammar class already accepts it — the
  milestone tick flips it, the issue-close tick doesn't (State filter).

- [x] **Step 3: Run** `go test ./cmd/sdlc/internal/project/` — the pre-existing
  tests must PASS unchanged (that's the point: typed parsing, same semantics).

- [x] **Step 4: Run** `go test ./cmd/sdlc/` (close-gate tests exercise the call
  sites) — PASS.

- [x] **Step 5: Commit** — `#180 M2: tick mutators re-implemented over typed Doc (same contract, no substring convention)`

#### Task M2.3: generalize the validate gate to a noun table

**Files:**
- Modify: `cmd/sdlc/validategate.go` (+ `validategate_test.go`)
- Modify: the `push.go`/`merge.go` invocation sites of `validateChangedIssues`
  (merge.go:326, push.go:126) and the `validateChangedIssuesFn` stub in
  `cmd/sdlc/merge_e2e_test.go:131,140`

- [x] **Step 1: Write failing test** — hermetic (seams injected per
  `validategate.go:30-35`): a diff window touching
  `workshop/projects/demo.md` must invoke the validator with
  `--type project`; an issue file still gets `--type issue` + section presence
  on added files; a project file gets NO issue-section check.

- [x] **Step 2: Implement** — introduce the table and generalize:

```go
// nounGate binds one vocabulary noun to the repo dir whose changed instances
// the fail-closed gate validates at push/merge.
type nounGate struct {
	noun          string
	dir           string // repo-relative home (from the noun's Discovery)
	checkSections bool   // issue-only: section presence on added files
}

// issuesDir = the caller-resolved dir (f.IssuesDir flag / WF_ISSUES_DIR env,
// falling back to the model home) — the override chain must survive.
func nounGates(issuesDir string) []nounGate {
	return []nounGate{
		{noun: "issue", dir: issuesDir, checkSections: true},
		{noun: "project", dir: vocab.Project().Discovery().Home},
	}
}
```

  `validateChangedIssues` becomes `validateChangedInstances(base, head, gates
  []nounGate, …)` iterating the table; `shellValidateFrontmatter` takes the noun
  (`--type <noun>`). Keep the exported/internal naming conventions of the file.
  **Preserve the dir-override path**: today both invocation sites pass
  `f.IssuesDir` (merge.go:326, push.go:126; env fallback `WF_ISSUES_DIR`) — the
  issue row's `dir` must stay the caller-resolved issuesDir, NOT hardcode
  `Discovery().Home` (the model home is the project row's source and the
  issuesDir default, not a replacement for the flag/env override). Also update
  the `validateChangedIssuesFn` stub signature in `merge_e2e_test.go:131,140`
  (add to Files list).

- [x] **Step 3: Run** `go test ./cmd/sdlc/ -run Validate` — PASS.

- [x] **Step 4: Commit** — `#180 M2: validate gate generalized to a noun table — project instances conform at push/merge`

#### Task M2.4: close M2

- [x] Full bare suite + `sh construct/vocabulary/vet_test.sh` — PASS.
- [x] **Live check (IO-adjacent milestone):** in a scratch clone, create
  `workshop/projects/bad.md` with `status: shipped`, run `sdlc push` — expect
  the gate to refuse naming the file and the enum. Delete the scratch.
- [x] Tick M2 plan rows; log; `sdlc milestone-close --issue 180 --milestone M2`.

---

## Chunk 3: M3 (verbs), M4 (board/retro/close + Phase-A), M5 (docs + drift)

### M3 — the `sdlc project` verb family

#### Task M3.1: command skeleton + helptext

**Files:**
- Create: `cmd/sdlc/project.go` (verb family), `cmd/sdlc/helptext/project.md`
- Modify: `cmd/sdlc/main.go` (register after `issue`), `cmd/sdlc/main.go:37-44`
  (`renderLong`: add `{{PROJECT_LIFECYCLE}}`, `{{PROJECT_STATUS_NAMES}}`
  placeholders sourced from `vocab.Project()`)
- Test: `cmd/sdlc/project_cmd_test.go`; helptext anchors mirror
  `cmd/sdlc/helptext/embed_test.go` conventions

- [x] **Step 1: Write failing tests** — cobra tree has `project` with
  subcommands `new,list,show,set-status,validate,status,retro,close`
  (walk the tree like `helptext_render_test.go:71-82` does); rendered Long
  contains the model-derived lifecycle (assert a `when` gloss line surfaces);
  `TestNoCommandLongHasSurvivingPlaceholder` still passes (it auto-covers the
  new placeholders).

- [x] **Step 2: Implement** `NewProjectCmd()` mirroring `issue.go:27-52`
  (parent `RunE → cmd.Help()`; each subcommand a `newProject<X>Cmd()` builder
  with its own flags struct; mutating ones wrap `markMutatingCommand`).
  `helptext/project.md` documents: the funnel (via placeholder), the gated
  sections (PRD/Estimate/Breakdown/Log), residency + archive-on-done, the
  retro convention (`### <ISO date> — retro` headings in `## Log`), and that
  set-status refuses →done in favor of `sdlc project close`. Subcommands
  `status`/`retro`/`close` register in M3 but `RunE` returns "lands in M4"
  errors only if M4 is deferred — otherwise leave them out of M3's commit and
  add in M4 (**do the latter**: register only
  `new,list,show,set-status,validate` now; the tree test grows in M4).

- [x] **Step 3: Run** `go test ./cmd/sdlc/ -run 'Project|Placeholder|Helptext'` — PASS.

- [x] **Step 4: Commit** — `#180 M3: sdlc project — verb family skeleton + model-derived helptext`

#### Task M3.2: `project new` / `list` / `show` / `validate`

**Files:**
- Extend: `cmd/sdlc/project.go`; scaffold render in
  `cmd/sdlc/internal/project/scaffold.go` (+ test)

- [x] **Step 1: Write failing tests** —
  `new --slug demo --goal "…" --done-when "…"` writes
  `workshop/projects/demo.md` with `type: project`, `name:`, `goal:`,
  `done_when:` (ALL four are `#Project`-required — the scaffold's own file must
  pass `sdlc project validate`, and the test asserts it does), `status: ideation`
  (`vocab.Project().InitialStatus()` — no literal), `created:`/`updated:` from
  the runner clock, and the model's four sections with seeds (derive from
  `Sections()`, mirroring `internal/issue/scaffold.go`'s model-derived render);
  refuses an existing slug; `--goal`/`--done-when` are required flags (empty
  values would scaffold a file that fails its own gate). `list` renders name/status/deadline rows from
  `workshop/projects/*.md` (model glob). `show --slug demo` prints path +
  frontmatter + task summary. `validate` shells
  `vocabulary validate-instance --type project <file>` (seam-injected like
  `issue.go:99-124`'s `runIssueValidate` — mirror it).

- [x] **Step 2: Implement.** Pure render/summarize helpers live in
  `internal/project`; the command layer only does fs + output.

- [x] **Step 3: Run** `go test ./cmd/sdlc/ ./cmd/sdlc/internal/project/` — PASS.

- [x] **Step 4: Commit** — `#180 M3: project new/list/show/validate — scaffold + listing derive from the model`

#### Task M3.3: guard registry + `project set-status`

**Files:**
- Create: `cmd/sdlc/internal/project/guards.go`, `guards_test.go`
- Create: `cmd/sdlc/projectsetstatus.go` (set-status subcommand + guard-runner
  wiring — the densest subcommand gets its own file so `project.go` holds only
  the parent + the thin new/list/show/validate builders, mirroring how
  status/close get `projectstatus.go`/`projectclose.go`)

- [x] **Step 1: Write failing guard tests** (pure, table-driven):

```go
// GuardCtx carries the injected world: evidence strings from flags, today's date.
type GuardCtx struct {
	Evidence map[string]string // keyed by guard name
	Today    string            // ISO date
}
type GuardFunc func(d *Doc, ctx GuardCtx) error
func Guards() map[string]GuardFunc
```

  Cases: `prd-present` fails on empty/seed-only `## PRD`, passes on prose;
  `phase-a-estimate` requires `**phase-a:** <N>h` in `## Estimate`;
  `baseline-set` requires non-empty `deadline` + `planned_finish` frontmatter;
  `reality-check` / `issues-cover-prd` require non-empty evidence in ctx
  (supplied via `--reality` / `--coverage` flags, appended to `## Log` by the
  verb on success); `retro-recorded` requires a `(?m)^### \d{4}-\d{2}-\d{2} — retro\b`
  heading in `## Log`; `fog-factor-recorded` is registered but only satisfiable
  by `sdlc project close` (returns an error directing there — set-status can
  never legally need it since →done is refused anyway; registering it keeps the
  unknown-guard check honest).

- [x] **Step 2: Implement guards** — pure over `Doc`.

- [x] **Step 3: Write failing set-status tests** — legality from the model
  (`CanTransition`; refusal message renders `LegalTransitions` like
  `setstatus.go` does for issues — read it first and mirror the UX);
  `--to done` refuses with "use `sdlc project close`"; guards of the matched
  `TransitionFor` edge run in order; a guard name missing from `Guards()`
  refuses loudly (model↔code drift); `--force` waives guards but not the
  legality of writing a status outside the enum; success rewrites `status:` +
  `updated:` via `Doc` and commits per the same convention `sdlc issue
  set-status` uses (verify at implementation and mirror).

- [x] **Step 4: Implement; run** `go test ./cmd/sdlc/... -run 'Guard|ProjectSetStatus'` — PASS.

- [x] **Step 5: Commit** — `#180 M3: project set-status — model-legality + named-guard registry (unknown guard = refusal)`

#### Task M3.4: close M3

- [x] Full bare suite — PASS. Tick plan rows; log.
- [x] `sdlc milestone-close --issue 180 --milestone M3`.

### M4 — derived board, retro, calibrated close, Phase-A method

#### Task M4.1: `computeBoard` + `sdlc project status`

**Files:**
- Create: `cmd/sdlc/projectstatus.go`, `projectstatus_test.go`
- Extend: `cmd/sdlc/project.go` (register `status` subcommand)

- [x] **Step 1: Write failing pure tests** — map-backed lookup fake:

```go
// issueMeta is the cross-repo issue surface the board reads.
type issueMeta struct {
	Status        string
	EstimateHours float64 // 0 when unset
	ActualHours   float64
	Deps          []string // refs; terminal-dep check gates the frontier
}

type boardRow struct {
	RefText, Title, IssueStatus string
	Ticked                      bool
	RemainingHours              float64
}

type board struct {
	Rows                  []boardRow
	Done, Total           int
	RemainingHours        float64
	Deadline, PlannedFinish string
	Frontier              []string // unblocked, not-started refs (deps all terminal)
	Blocked               []string
	// Threads: the parallel threads as independent dep-subgraphs (Spec's third
	// derived computation): connected components over the project's unfinished
	// refs, edges = deps restricted to refs in this project. The component
	// count is what the operator holds against the ~2-concurrent-sessions
	// attention ceiling (#117).
	Threads   [][]string
	LastRetro string // ISO date of newest retro heading, "" if none
}

func computeBoard(d *project.Doc, lookup func(refText string) (issueMeta, error)) (board, error)
```

  Cases: ticked rows counted done regardless of issue status (with a mismatch
  warning row when the issue isn't terminal); plain-text tasks (no ref) roll up
  by checkbox only; unresolvable refs degrade to a warning entry, never an
  error (a peer repo may be absent on this machine); remaining = Σ
  `EstimateHours` of unticked rows with non-terminal issues; frontier excludes
  rows whose deps contain a non-terminal ref; threads = connected components
  over unfinished refs (two rows sharing a dep edge land in one thread; three
  rows with no cross-deps = three threads — test both shapes); `LastRetro`
  parsed from `## Log` via `project.LatestRetroDate` (a pure helper in
  `internal/project`, shared with M4.2's nudge so the close gate never grows a
  lookup dependency).

- [x] **Step 2: Implement pure core; run** — PASS.

- [x] **Step 3: Wire the seam** — real lookup: `parseRef(refText)` →
  `resolveRepoDir` (current repo when `Repo` empty) → read
  `<repoDir>/<vocab.Issue().Discovery().Home>/NNNNNN-*.md` frontmatter via
  `internal/issue` helpers; tolerate archived issues by also checking
  `ArchiveSubdir(historyRoot, ArchiveIssues)`. Render: header (name, status,
  deadline, planned finish, days left computed from injected today), progress
  `done/total`, `Σ remaining ≈ Nh`, frontier, blocked,
  `threads: N — [refs] / [refs]` (the component count is the line the operator
  reads against the ~2-session attention ceiling — computing it without
  rendering it would be dead derived state; assert it in the render test),
  last-retro age. Add a
  hermetic test with a sibling fixture repo (mirror `hermeticrepo_test.go`
  scaffolding) covering one cross-repo ref.

- [x] **Step 4: Run** `go test ./cmd/sdlc/ -run 'Board|ProjectStatus'` — PASS.

- [x] **Step 5: Commit** — `#180 M4: sdlc project status — the derived kanban (baseline stored, progression computed, nothing hand-maintained)`

#### Task M4.2: `sdlc project retro`

**Files:**
- Extend: `cmd/sdlc/project.go`; pure bits in `internal/project`

- [x] **Step 1: Write failing tests** — `retro --slug demo` appends to `## Log`:

```markdown
### 2026-07-20 — retro

**board:** 3/7 done · Σ remaining ≈ 22h · deadline 2026-09-01 (43 days) · frontier: ariadne#182, metis#9

<where we are + what changed + new forecast — replace this line>
```

  (board line rendered from `computeBoard`; the prose placeholder is the
  agent/operator's to fill — mechanism, not mandate). Appending twice on one
  date is allowed (distinct entries). `--dry-run` prints without writing.

- [x] **Step 2: Implement** (reuse `Doc.AppendToSection("Log", …)`); commit
  convention as other mutating verbs. Run tests — PASS.

- [x] **Step 3: Staleness nudge at issue close** — in `close.go`'s project gate
  (after the tick edit is prepared, ~`close.go:584-655`): when the matched
  project's `status` is in the executing category and its newest retro heading
  is older than 7 days (or absent), print a `[!]` info line nudging
  `sdlc project retro` — never a refusal, and skipped under `--no-project`.
  Uses `ParseDoc` + `project.LatestRetroDate` (the pure helper from M4.1) — no
  new parsing and no issue-lookup dependency. Legacy files without the new
  sections must pass through silently (SectionBody("Log") tolerant of absence);
  `status: active` (legacy) is NOT in the model's executing category → no nudge,
  which is correct pre-#171. Hermetic test each branch.

- [x] **Step 4: Run** `go test ./cmd/sdlc/ -run 'Retro|Close'` — PASS.

- [x] **Step 5: Commit** — `#180 M4: project retro verb + stale-retro nudge in the issue-close project gate`

#### Task M4.3: `sdlc project close` — retro gate, fog factor, archive

**Files:**
- Create: `cmd/sdlc/projectclose.go`, `projectclose_test.go`
- Modify: `cmd/sdlc/internal/processmanual/gatesig.go` (gate catalog) +
  `cmd/sdlc/gates_test.go` (`TestGateCatalogMatchesRegisteredFlags` uses a
  fixed spine map — extend it to include `project close`; add `retro`/`ledger`
  rows to `GateCatalog`, from which `GateFlagsFor` derives). The catalog rows
  carry `Grammar`/`AckPat`/`RefusalPat` metadata for the friction classifier —
  give the new rows the same ack grammar the neighboring close-gate rows use
  (read the `verified`/`atlas` rows first) and update the file-header
  gate/row-count comments. Verify the transcript scanner's verb maps
  (processmanual's codex.go / friction.go use single-word verbs) can express
  the two-word `project close`; if not, extend the map key format rather than
  registering a row no invocation ever matches.

- [x] **Step 1: Write failing tests** (hermetic; brain dir = temp fixture):
  1. requires `status == "executing"` EXACTLY (design decision 9): a forming/
     committed status refuses pointing at the funnel; `paused` refuses with a
     "resume first" pointer (`sdlc project set-status --to executing`) — the
     model deliberately has no `paused→done` edge, and the verb must never
     bypass `TransitionFor`;
  2. refuses without a retro entry (`--no-retro` bypasses, recorded in output);
  3. fog factor: reads `**phase-a:** 40h` from `## Estimate`, sums
     `actual_hours` across `mvp_scope` refs (issue frontmatter via the M4.1
     lookup seam; `N/A`/missing → warn + skip that ref), appends one row to the
     brain ledger table AND a `### <date> — close` Log entry with
     `fog = Σactuals / phase-a` to 2 decimals; `--no-ledger` skips the brain
     write (e.g. no sibling brain); missing `**phase-a:**` → warn + `fog: n/a`
     row in the Log entry only (a pre-model project may lack Phase-A);
  4. flips `status: done` + `updated:` via the model transition, then MOVES the
     file to `ArchiveSubdir(<repo>/workshop/history, ArchiveProjects)`
     (creating the dir), all in one commit;
  5. `drop` mode: `--drop` takes the executing→dropped OR paused→dropped edge
     (whichever `TransitionFor` matches the current status; retro still gated),
     archives the same way, writes no ledger row.
  Note on the actuals source: the Spec says "roll up from the calibration
  ledger"; the plan reads issue frontmatter `actual_hours` instead — the same
  measured value (frontmatter is what `sdlc close` adopts into the ledger), and
  frontmatter is resolvable per-ref cross-repo where the ledger is keyed
  differently. Deliberate choice, not an omission.

- [x] **Step 2: Implement.** Ledger row format (the table it appends to is
  created by M4.4's doc):

```markdown
| <name> | <phase-a>h | <Σ actuals>h | <fog> | <ISO date> |
```

  Brain path: `<brain-dir>/data/life/42shots/velocity/estimate-logic-project-v1.md`,
  `--brain-dir` default `../brain` (same as `close.go:140`); if the file or its
  `## Fog ledger` table is missing → refuse with the exact heading to add
  (unless `--no-ledger`).

- [x] **Step 3: Run** `go test ./cmd/sdlc/ -run ProjectClose` and the gate
  catalog test — PASS.

- [x] **Step 4: Commit** — `#180 M4: sdlc project close — retro-gated, fog-factor ledger row, archives to history/projects`

#### Task M4.4: Phase-A estimation method doc (design deliverable)

**Files:**
- Create: `~/workspace/brain/data/life/42shots/velocity/estimate-logic-project-v1.md`
  (brain is a capture repo — write + let its auto-save flow own the commit;
  do NOT run sdlc there)

- [x] **Step 1: Write the v1 method** — contents (complete, this IS the design):
  - **Scope:** PRD-stage estimation (Phase A), before issues exist. Phase B (at
    breakdown) is per-issue estimate-logic-v3.1, unchanged. Same calibration
    process as v3.1, different primitives + an explicit uncertainty multiplier
    (operator refinement, #180 Spec).
  - **Primitives:** decompose the PRD into *workstreams* (a coherent chunk one
    issue-or-few will own). Classify each S / M / L / XL with hour midpoints
    S=3h, M=8h, L=20h, XL=40h (seeded from issue-ledger percentiles; recalibrate
    once fog rows exist). Base = Σ midpoints.
  - **Fog factor:** `phase-a = base × fog`. Default fog **1.5** until the ledger
    has ≥3 rows; then use the ledger's median observed fog. Record both.
  - **Recording:** in the project file's `## Estimate`:

    ```markdown
    **phase-a:** 36h
    **fog:** 1.5
    **basis:** 3 workstreams — model+binding (M/8h), verbs (L/20h), docs (S/3h); base 31h... (agents idles differ)
    ```

  - **Calibration bridge:** at project close, `sdlc project close` appends
    `| project | phase-a | Σ issue actuals | fog | closed |` to the
    `## Fog ledger` table below. Over projects this calibrates the PRD-stage
    multiplier exactly the way issue closes calibrated v3.1.
  - Include the empty `## Fog ledger` table header row.

- [x] **Step 2: Cross-link** — `sdlc estimate-source` surfaces the issue-level
  doc; add a one-line pointer to the project doc from `helptext/project.md`
  ("Phase-A method: …path…"). (Extending `estimatesource.go` to know both docs
  is a natural follow-up, NOT the purpose — note in Log if skipped.)

- [x] **Step 3: Commit** (ariadne side: the helptext pointer edit rides the next
  ariadne commit) — `#180 M4: helptext — Phase-A method pointer`.

#### Task M4.5: live dogfood + close M4

- [x] **Live fixture pass** (lessons.md: IO needs a live run; use a symlinked
  cwd to exercise the `$PWD` branch): in a scratch repo (`$TMPDIR` via a
  symlink), run the full arc: `project new` → hand-write a PRD →
  `set-status --to defined` → fill `## Estimate` per M4.4 →
  `set-status --to committed --reality "fits July"` (deadline+planned_finish
  set first; verify the guard refuses before, passes after) → seed two task
  rows referencing fixture issues → `set-status --to executing --coverage "2
  issues cover PRD"` → `project status` (board renders) → `project retro` →
  tick issues → `project close --no-ledger` (or with a fixture brain) →
  file lands in `workshop/history/projects/`. Record the transcript in `## Log`.
- [x] Full bare suite — PASS. Tick plan rows; log.
- [x] `sdlc milestone-close --issue 180 --milestone M4`.

### M5 — prose demotion, drift binding, atlas

#### Task M5.1: demote `construct/datatype/project.md` to cite the cue

**Files:**
- Modify: `construct/datatype/project.md`

- [ ] **Step 1: Edit** — (procedure refers, registry defines):
  - Frontmatter-shape table: `status` row now reads "see
    `construct/vocabulary/project.cue` (the schema authority) — funnel:
    ideation → defined → committed → executing → done | dropped, + paused";
    add `deadline` / `planned_finish` rows (set at commit).
  - Add a short **Lifecycle** section: the funnel one-liner per status, each
    token backtick-quoted (`` `ideation` — <gloss> `` — the drift test asserts
    the backticked byte form, so write it that way here rather than converging
    via a test-fail loop; copy the `when` glosses), transitions owned by
    `sdlc project set-status` / `close`.
  - Default location: `workshop/projects/<slug>.md` (was `data/project/` — note
    brain-era files migrate under #171); archive-on-done to
    `workshop/history/projects/`.
  - **Rewrite the Body-skeleton section to the four gated sections** (the prose
    currently teaches `## tasks` + `## details`, which the scaffold no longer
    produces): `## PRD` (grown at define), `## Estimate` (Phase-A fields, at
    commit), `## Breakdown` (the committed baseline: deadline rationale, thread
    assignments, sequencing — AND the task list; task-line grammar, checkbox
    states, detail blocks, and jump-link conventions all persist here
    unchanged, since the close gate still consumes them), `## Log` (retro +
    re-forecast + scope-event entries; document the `### <ISO date> — retro`
    heading convention).
  - Search recipes: update paths to `workshop/projects/`, and the
    `rg -l "^status: active"` recipe (project.md:214) becomes
    `rg -l "^status: executing"` (the model's live-portfolio status).
  - Keep everything else (single-operator discipline, scope events, MVP
    conversation) — this is a demotion of the *schema* claims only.

- [ ] **Step 2: Commit** — `#180 M5: datatype prose demoted — project.cue is schema authority`

#### Task M5.2: drift test binding prose ↔ model

**Files:**
- Create: `pkg/vocab/prose_drift_test.go`

- [ ] **Step 1: Write the test** (mirrors the invariant-chain pattern —
  `helptext/issue_sections_test.go` for the ⊆-direction):

```go
// TestProjectProseCitesModel binds construct/datatype/project.md to the model:
// every status token and every scaffold section must appear in the prose, the
// prose must cite the cue as schema authority, and the retired hand-maintained
// enum must not survive anywhere in the file.
func TestProjectProseCitesModel(t *testing.T) {
	prose, err := os.ReadFile("../../construct/datatype/project.md") // go test runs in pkg dir
	// 1. for each Project().AllStatuses(): strings.Contains(prose, "`"+s+"`")
	//    (statuses appear backtick-quoted in the prose gloss)
	// 2. for each Project().Sections(): strings.Contains(prose, "## "+name)
	// 3. positive citation: strings.Contains(prose, "construct/vocabulary/project.cue")
	// 4. the retired enum token, TWO exact byte forms, both must be FALSE:
	//    strings.Contains(prose, "`active`")     — the status-row remnant
	//      (pre-demotion project.md:28 wraps tokens in backticks with escaped
	//      pipes; a loose `status.*active\s*|` regex never matches those bytes
	//      and would guard nothing)
	//    strings.Contains(prose, "status: active") — the search-recipe remnant
	//      (project.md:214 sits inside a sh fence, NOT backticked, so the first
	//      assertion never sees it; nothing post-demotion legitimately
	//      contains this string)
}
```

- [ ] **Step 2: Sanity-check the binding both ways** — `git stash` the M5.1
  edit and run the test: it must FAIL against the pre-demotion prose (the
  backticked `active` + missing citation both trip). Unstash; it must PASS.
  A drift test that passes on the un-demoted prose is vacuous — prove it bites
  before committing it.

- [ ] **Step 3: Commit** — `#180 M5: drift test — prose table bound to project.cue`

#### Task M5.3: atlas + skill claim + issue close

- [ ] **Atlas:** add `atlas/` coverage for the project noun (new file or the
  vocabulary page, matching how the issue noun is mapped — check
  `atlas/index.md` and link from it): model location, funnel, verb family,
  gates, fog-factor loop, #171 handoff note.
- [ ] **xx-vocabulary skill:** verify `.claude/skills/xx-vocabulary/SKILL.md` —
  its claim is now true for project; update any noun enumeration it carries.
- [ ] **lessons.md:** only if something recurred that no code now enforces
  (memory: feedback_lessons_only_if_not_code_enforced).
- [ ] Full bare suite + vet_test.sh — PASS.
- [ ] Tick remaining plan rows. `sdlc close --issue 180 --verified '<evidence:
  test suite + live-fixture transcript + dormant-gate demo>'` (actuals measured,
  not typed; atlas gate satisfied by M5.3). Then `sdlc pr` → `sdlc merge` per
  the publish convention (single publish at issue close).

### Execution notes

- Follow the #174 FIX-THEN-SHIP protocol at every boundary: fix review findings
  pre-commit, bundle into the close commit, anchor == HEAD.
- Estimate (set in issue frontmatter before `sdlc change-code`): see issue
  `estimate_hours:` — derived per estimate-logic-v3.1 after plan approval.
- Milestone → plan-row mapping lands in the issue `## Plan` as five `Mx` rows
  (each a genuine review boundary).

## Revisions

### 2026-07-16 — dogfood reversal (operator, at plan approval)

**Reason:** the operator wants the project-management lift itself as the
guinea pig from day one: "use the creation of project management in ariadne
as a project to guinea pig the project management improvement itself."
Supersedes the 2026-07-15 dogfood-deferral decision recorded in Chunk 1's
scope boundary.

**Delta:** `workshop/projects/project-management-primitive.md` now exists
(hand-authored at ideation to the emerging model shape; mvp_scope
[ariadne#180, ariadne#171], #182 explicitly out). Consequences for tasks:

- Chunk 1 scope boundary: "the repo has no project instances yet … dormant
  until files exist" no longer holds — the conformance gate goes LIVE against
  this instance the moment M2.3 lands. M1.1's `#Project` must accept this
  file as-written (it is the first conformance fixture; if it fails vet, fix
  whichever side is wrong and log the call).
- M2.4 live-check: run the bad-status check against a scratch COPY as
  planned, but ALSO run `vocabulary validate-instance --type project` against
  the real instance — expected PASS.
- M3/M4 dogfood: prefer the real instance over scratch fixtures wherever a
  verb test wants a live file (set-status define→…, status board, retro);
  the M4.5 fixture arc still runs in a scratch repo for the destructive
  close/archive step (the real instance closes only when the project is
  actually done).
- #180's issue scope is unchanged — single multi-boundary issue; the project
  file tracks the wider lift, it does not restructure the issue.

### 2026-07-16 — M1 boundary-review reconciliation

**Reason:** the M1 `FIX-THEN-SHIP` review found one stale served consumer and
three plan statements that no longer matched the implementation or the
operator's same-day scope decision. The generated consumer is refreshed via
`make weave`; this revision records the durable plan deltas without rewriting
the approved plan in place.

**Delta:**

- M1.2's Files list drops the planned `pkg/vocab/lifecycle_test.go`. The shared
  helpers are exercised through the issue, verdict, and project model tests;
  the behavior-preserving extraction kept the existing pins green, so a new
  helper-only test file would duplicate those contracts.
- M1.4 migrated **11 call sites total: 9 non-test + 2 test**. The task's
  implementation text and proposed commit subject saying "9 call sites" refer
  only to the non-test sites; the implementation commit and issue estimate use
  the total and are authoritative.
- The prior dogfood revision's `mvp_scope [ariadne#180, ariadne#171], #182
  explicitly out` was superseded later on 2026-07-16. The live project now has
  `mvp_scope: [ariadne#180, ariadne#171, ariadne#182]` and
  `explicitly_out: [ariadne#15]`: the operator established that computed
  effort-to-calendar feasibility is the defining timeline capability that
  distinguishes a project from an issue, so #182 is in the MVP while remaining
  a separately implemented issue.

### 2026-07-16 — M2 boundary-review reconciliation

**Reason:** the M2 `FIX-THEN-SHIP` review found three implementation ambiguities,
one ownership contradiction, and durable verification evidence that had not
been reflected in the detailed plan rows.

**Delta:**

- Typed `Doc.Tasks` is scoped to the real `## Breakdown`; the legacy close
  mutators retain their historical whole-document scan through an explicitly
  named compatibility seam. Markdown fences are structure-aware so example H2s
  and checkboxes cannot masquerade as project structure.
- Close-time schema validation cannot safely hard-fail until #171 migrates the
  five legacy `status: active` brain projects. #180 supplies typed mutation and
  push/merge conformance; #171 owns activating close-time validation afterward.
- The scratch bad-status push refusal, live-instance conformance, full suite,
  and vocabulary harness were run before review and are now recorded in the
  checked M2 rows and issue Log.
- The milestone-actual deviation correction is an enabling side quest, not an
  expansion of M2's project-model scope: cumulative claim-to-HEAD measurements
  cannot be compared to per-milestone increments.

### 2026-07-16 — M3 REWORK boundary reconciliation

**Reason:** the first M3 boundary review found that free-text scaffold values
were not YAML-safe and that three slug consumers could escape the modeled live
project directory. It also found the new user-facing verb family absent from
README.md.

**Delta:**

- `RenderScaffold` encodes every user-provided frontmatter string as a YAML-safe
  double-quoted scalar, and `Doc.FM` decodes that representation for typed
  consumers. Regression coverage includes colons, comments, quotes, newlines,
  booleans, and date-like text plus a real vocabulary-validator subprocess.
- The PURE `project.ResolvePath` is now the single slug-to-path boundary for
  new/show/validate/set-status. Canonical lowercase kebab slugs are required;
  traversal and nested paths refuse before any read, validator call, or write.
- The Core concepts table now names scaffold, summary, and locator entities;
  README.md carries the runnable M3 project workflow.

### 2026-07-16 — M4 live-dogfood frontier correction

**Reason:** the required symlinked process-level run showed an unblocked
`working` issue missing from the derived frontier. The approved Spec defines
frontier as “what's unblocked”; `IssueModel.IsOpen` is narrower and means only
“not yet started.”

**Delta:** frontier membership is any unfinished, dependency-unblocked,
non-terminal issue (with explicit `blocked` still excluded). A focused
regression covers the active `working` case before the one-predicate fix.

### 2026-07-16 — M4 full-suite integration correction

**Reason:** the first full bare suite found that the new nested lifecycle verb
was present in the shared workflow catalog but had not joined the brain/non-SDLC
repo guard, and two fixed census tests still pinned the pre-M4 gate/lock sets.

**Delta:** `project close` runs `guardSpineRepo` before manual `--slug`
validation (so repository identity remains guard-first), the lifecycle test
executes multiword catalog keys as separate argv tokens, and the repo-lock plus
14-gate census pins include the new boundary.

### 2026-07-16 — M4 REWORK boundary reconciliation

**Reason:** the M4 fresh-context review found that close recognized selected
guard names instead of consuming the transition's complete ordered guard list,
could calibrate from a partial actual sum, and changed the live project before
the sibling ledger write. It also found the Phase-A document misclassified as a
PURE entity, missing README examples, and ambiguity about real-instance M4
dogfood.

**Delta:**

- Close iterates every modeled guard through a close-owned handler registry and
  refuses unknown names. The retro and fog handlers remain distinct because fog
  prepares a cross-repo transaction rather than calling the generic document
  guard.
- Fog calibration requires a positive measured actual for every MVP ref.
  Lookup errors, unset values, explicit `N/A`, and non-positive values make the
  set incomplete: ledger-backed close refuses; acknowledged `--no-ledger`
  archives with `actuals: incomplete` / `fog: n/a` instead of a false row.
- Project and ledger contents are staged before either durable path changes;
  ordered renames have reverse-order compensation. A forced ledger-stage
  failure regression proves both original records remain unchanged.
- The Phase-A method moves from PURE entities to Integration points as a
  human-executed process/document surface. Validation is structural content
  review, helptext linkage, live command use, and close-ledger tests—not a
  fictional colocated pure unit test.
- README now shows status, retro dry-run/write, close/drop, and explicit bypass
  semantics. The real ideation project is exercised with read-only `status` and
  `retro --dry-run`; lifecycle mutation remains correctly deferred until its
  PRD and baseline are genuinely ready.

### 2026-07-16 — M4 YAML-semantics and Phase-A reconciliation

**Reason:** the third fresh review found that raw line extraction did not honor
all YAML representations accepted by `#Project`/`#Issue`, and that the close
path collapsed absent, malformed, and non-positive Phase-A estimates into one
automatic legacy bypass. It also found the compensation test stopped before
ledger replacement and the checked board contract's last-retro age was absent.

**Delta:**

- One typed YAML metadata decoder now supplies quoted scalar, flow-list,
  block-list, and numeric semantics to both project and issue consumers. Status,
  `mvp_scope`, deps, estimates, and actuals are no longer re-parsed as raw lines;
  pure plus real-lookup regressions cover each representation.
- One tri-state Phase-A parser feeds both the commitment guard and close.
  Malformed/non-positive values always refuse. An absent value is the only
  legacy case and requires explicit `--no-ledger`/`--force`, matching README and
  the model's named fog guard.
- Transaction rename is injected in tests: a forced archive rename after the
  ledger has been replaced proves reverse compensation restores both original
  records. Board rendering now includes last-retro age from injected `today`.

### 2026-07-16 — M4 finite-number and residual-consumer reconciliation

**Reason:** the fourth fresh review found that YAML/legacy non-finite numbers
could survive semantic decoding and enter the fog ledger, two M4 consumers had
not yet switched to decoded metadata, and the Core concepts table described M5
entities as already delivered.

**Delta:**

- Numeric metadata is valid only when finite and positive; quoted/unquoted
  NaN/Inf and zero refuse before board aggregation or calibration. Close also
  defends its injected issue-meta seam against non-finite test/future callers.
- Board name and stale-retro status use typed metadata. A close transaction
  captures `today` once so Log, ledger, and frontmatter cannot cross midnight.
- Close/drop endpoints derive from vocabulary event transitions; explicit
  blocked detection derives from the issue model's `block` event target. Event
  names remain deliberate verb-policy identifiers; enum values are not copied.
- Datatype prose and the prose-drift test are labeled `planned M5` in Core
  concepts until that milestone lands.

### 2026-07-16 — M4 canonical-scope calibration reconciliation

**Reason:** a later M4 fresh review found that `mvp_scope` could repeat one
logical issue through identical or current-repository alias references. Close
would sum every list entry and persist a precise-looking but inflated fog
factor.

**Delta:** before reading any actual or mutating either durable record, project
close parses each MVP ref to a canonical resolved-repository-plus-issue-ID
identity and refuses duplicates. Regressions cover both exact duplicates and
`ariadne#1` / `#1` alias equivalence (ARCH-PURPOSE).

### 2026-07-16 — M4 unavailable-peer compatibility reconciliation

**Reason:** the next M4 review found that filesystem-required canonicalization
made an absent peer an unconditional error, overriding the established contract
that lookup failures are incomplete actuals and `--no-ledger` may explicitly
close without calibration.

**Delta:** duplicate identity is now best-effort: resolvable refs use the
canonical repository path, while an absent peer uses a normalized repository
token so repeated spellings are still caught. Unparseable refs continue to the
lookup seam and become unavailable inputs. A missing-peer regression proves a
ledger-backed close refuses as incomplete and `--no-ledger` archives with
`actuals: incomplete` / `fog: n/a` (ARCH-PURPOSE).

### 2026-07-16 — M4 logical-identity board reconciliation

**Reason:** the next M4 review found that dependency components used raw ref
spelling, so current-repository and peer-prefix aliases could split one logical
dependency graph into false parallel threads. Exact duplicate task refs could
also double-count remaining effort.

**Delta:** issue lookup now carries the shared canonical repository-plus-ID
identity into the pure board core. Board maps and components key on that
identity while retaining the first authored ref for display; duplicate task
refs contribute one effort value and one component. Regressions cover
`ariadne#3` / `#3`, peer-prefix / full-peer identity, and exact duplicates. The
same resolver helper now serves close and status (ARCH-DRY, ARCH-PURPOSE).

### 2026-07-16 — M4 structural fog-ledger reconciliation

**Reason:** the next M4 review found that fog-ledger insertion assumed every
split line owned a newline byte (panicking at a valid EOF table) and located the
section with an unanchored substring that could select prose or fenced examples.

**Delta:** ledger transformation is now a pure helper over text. It uses the
project package's shared fence-aware Markdown scanner to locate the real
level-two section, restricts insertion to its first contiguous table, and
reconstructs lines without invented byte offsets. Regressions cover EOF after
the divider, EOF after an existing row, inline/fenced fake headings, and the
real later section (ARCH-DRY, ARCH-PURE, ARCH-PURPOSE).
