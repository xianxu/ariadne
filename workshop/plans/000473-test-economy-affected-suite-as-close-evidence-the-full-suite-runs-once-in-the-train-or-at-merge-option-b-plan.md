---
collaboration-mode: discussion
---

# Affected Suite as Branch Evidence Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution
> approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or
> superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A branch proves itself with the *affected suite* (changed Go packages, their reverse dependents, and the
declared non-Go groups they touch), and the full suite runs once, in the train or at merge.

**Architecture:** A pure planner turns a branch's changed paths, the module's package graph and a per-repo declaration
(`.sdlc/affected.json`) into a selection: nothing, a subset, or full. `sdlc local-check run --name affected` runs the
repo's declared affected command with the selection in its environment and records `sdlc-local/affected`, or, when the
selection is full, runs the `suite` check and records `sdlc-local/suite`. A train car's readiness accepts either record;
`sdlc merge` on a repository without trains keeps requiring `suite` on the integrated head, so it still runs the full
suite there once.

**Tech Stack:** Go (cmd/sdlc), `go list`, git, the existing local-check runner, test-stats store and landing train.

---

## Design status

The TL approved decisions D1–D6 on 2026-10-19 (ops:0). The env contract below is a fixed interface: pair's
counterpart is filed against it once M1 lands.

### What the code does today (measured 10-19, read of `cmd/sdlc`)

- `sdlc close` has **no mechanical suite gate**: `--verified` is free text, and `sdlc-local/suite` appears only as its
  example (`close.go:1898`). The record is enforced by the CI gate (`decideCIGate`, `cigate.go:96`) at `sdlc merge`
  and, read-only, when a train car vouches (`productionCarGates`, `landingcar.go:90`).
- A check runs a fixed argv (`runUnderSlot`, `localcheck_run.go:239`); nothing passes it a package list.
- An ancestor's record counts only while everything since it is docs-only (`localRecordsOnAncestors`,
  `cigate.go:488`, using `publishGateHasCodeSurface`).
- Known-flake handling (#394/#466), the success guard and the test-stats aggregates all key on check name `suite`.
- `go list -test -f '…Deps…' ./...` over pair (145 packages): 1.5 s. Computing the graph per run is cheap.
- pair's `make test` runs ~37 script/Lua/smoke targets and then `go test ./...` (couchcore alone ~580 s); it has no
  package-list entry point.

### Decisions (approved)

- **D1 — the declaration is a new file, `.sdlc/affected.json`, not new fields in `.sdlc/local-checks.json`.**
  `parseLocalChecks` uses `DisallowUnknownFields`, so a new field there makes every sdlc built before this refuse the
  whole local-check declaration, and with it every merge in that repo, until the fleet rebuilds (ARCH-EVOLVE). A
  separate file is invisible to old binaries: they keep requiring `suite`, which is conservative. The file is read from
  main, like local-checks.json, and sits under `.sdlc/`, so the local-gates guard already protects it. Its absence is
  the per-repo off switch the Spec asks for.
- **D2 — "close accepts affected" means the evidence text, not a new gate.** Close has never checked a suite record;
  this changes `close`'s VERIFIED example and the publish gate's full-suite instruction to name `affected`, and does not
  add a mechanical close gate. If the TL wants close to check a recorded `affected` or `suite` on the head, that is a
  separate gate with its own `--no-` flag. I'd file it rather than fold it in.
- **D3 — only the car readiness relaxes.** A car boards on `affected` *or* `suite`. `sdlc merge` without a train keeps
  requiring `suite` on the integrated head (it runs it once there, as today). The train's prefix suite is unchanged.
  In code: `ciInputs` gains `SuiteStandIn string` (`"affected"` for the car gate, `""` for merge); `decideCIGate`
  treats a recorded stand-in as satisfying `suite`.
- **D4 — `affected` is a test run for flake purposes, and its own row for timing.** It gets the suite's
  `known_flakes` policy, the targeted rerun (`flakes.RerunCommand` is already Go-only `go test -run`), the success
  guard (no `affected` pass over an unexplained `affected` failure on the same head), and counts as flake evidence.
  Duration aggregates stay per check name, so affected runs don't skew the suite's medians.
- **D5 — a full selection runs and records `suite`.** When the planner says full (an infra path changed, or a path
  nothing classifies), `local-check run --name affected` runs the declared `suite` check instead and says so. The
  car then has a `suite` record, which every gate already accepts. There's no `affected` record claiming to be full.
- **D6 — pair's side is a cross-repo item for the TL.** pair needs `.sdlc/affected.json` and a `test-affected` make
  target that reads `SDLC_AFFECTED_GO_PKGS` and `SDLC_AFFECTED_GROUPS`. Per tl-worker I don't file in pair; I'll send
  the proposal with a draft of both. The Done-when's one-week measurement depends on it landing.

### The affected command contract (fixed interface)

The repo's `affected.json` `command` is an argv, run at the root of a detached worktree of the head under the
machine-wide suite slots, like any local check. sdlc runs it as `env SDLC_AFFECTED_GO_PKGS=… SDLC_AFFECTED_GROUPS=…
<command…>`, through argv and never a shell string.

| Variable | Value | Empty when |
|---|---|---|
| `SDLC_AFFECTED_GO_PKGS` | the import paths to test, sorted, space-separated (`go test $(SDLC_AFFECTED_GO_PKGS)`) | only groups were selected |
| `SDLC_AFFECTED_GROUPS` | the selected group names from `affected.json`, sorted, space-separated (`$(filter lua,$(SDLC_AFFECTED_GROUPS))`) | no group was selected |

- The command never runs with both variables empty. sdlc records that case (a docs-only change) itself.
- It never runs for a full selection. sdlc runs the `suite` check instead (D5).
- Exit 0 is a pass. Output should be the suite's own `go test` output, because the log parser attributes failures and
  known-flake reruns from it.
- Group names follow the local-check name rule `^[a-z0-9][a-z0-9-]{0,40}$`, so they need no quoting in make.
- Changing a variable's name or meaning bumps `affected.json`'s `version` and, with it, every consumer (ARCH-EVOLVE).

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `affected.Config` | `cmd/sdlc/internal/affected/config.go` | new |
| `affected.Graph` | `cmd/sdlc/internal/affected/graph.go` | new |
| `affected.Selection` | `cmd/sdlc/internal/affected/plan.go` | new |
| `affected.Plan` | `cmd/sdlc/internal/affected/plan.go` | new |
| `ciInputs.SuiteStandIn` | `cmd/sdlc/cigate.go` | modified |
| `isTestRun` (was `isSuiteRun`) | `cmd/sdlc/teststats.go` | modified |

- **affected.Config**: the parsed `.sdlc/affected.json`. It's strict (unknown fields, a bad version, a bad glob or an
  empty command all refuse), because a misread selection would skip tests. A refusal shows as "affected unavailable;
  run the suite", never as an empty selection (ARCH-SECURE).
  ```json
  {"version": 1,
   "command": ["make", "test-affected"],
   "weight": 1, "timeout_minutes": 30,
   "full":   ["Makefile*", "go.mod", "go.sum", ".sdlc/**", "tests/lib/**"],
   "groups": [
     {"name": "lua",      "paths": ["**/*.lua", "tests/lua/**"]},
     {"name": "pair-cli", "paths": ["tests/pair-*.sh"], "go": ["./cmd/pair"]}
   ]}
  ```
  A group is selected when a changed path matches its `paths`, or when one of its `go` roots depends on an affected Go
  package. That's how a script test that drives a built binary reruns when the binary's code changes.
  - **Relationships:** 1 per repo, read from main (as `readLocalChecks`); N groups.
  - **Future extensions:** more selectors per group (e.g. env-gated live tests) go in the group object.
- **affected.Graph**: in-module packages, each with its directory and its transitive in-module dependencies,
  test-only imports included (`go list -test`). It has one method, `Dependents(set) set`.
  - **DRY rationale:** this is the first package graph in the repo (the exploration found none).
- **affected.Selection**: `{Full bool; Reason string; GoPkgs []string; Groups []string}`. It is empty for a docs-only
  change, which records a pass that ran nothing and says so.
- **affected.Plan(paths, graph, cfg) Selection**: classifies each changed path:
  1. Docs (the existing `publishGateHasCodeSurface` rule) → ignored.
  2. Matches `full` → Full, with the path as the Reason.
  3. Otherwise the selectors are **additive**: a file inside a package directory (a `.go` file, testdata, an embedded
     asset) selects its nearest package, *and* every group whose `paths` match it is selected too. pair's Lua lives
     inside the embedding package `cmd/internal/runtimebundle`, so a `.lua` change selects that package and the `lua`
     group (PQ-1).
  4. A path that neither a package nor a group claims → Full ("unclassified: path").

  It then expands the packages to their dependents and selects the groups whose `go` roots depend on them. It is pure,
  with a table test.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `loadGoGraph` | `cmd/sdlc/internal/affected/golist.go` | new | `go list -test -json ./...` |
| `runAffectedCheck` | `cmd/sdlc/localcheck_affected.go` | new | worktree, git diff, runner |
| `readAffectedConfig` | `cmd/sdlc/cigate.go` | new | `git show origin/main:.sdlc/affected.json` |
| `productionCarGates` | `cmd/sdlc/landingcar.go` | modified | CI gate (stand-in) |

- **loadGoGraph**: runs `go list` in the head's detached worktree; a parse or exit failure means Full ("package graph
  unavailable"). ARCH-MOCK: `go` is the toolchain the suite itself runs. The seam is a function value
  (`func(dir) (Graph, error)`), and the planner tests use literal Graphs. A conformance test runs the real `go list` on
  a fixture module of three packages (a leaf, a dependent, and a test-only dependent) and checks the graph.
- **runAffectedCheck**: in the existing detached worktree, it diffs `merge-base(origin/main, sha)..sha`
  (`gitx.ChangedPathArgs`), loads the graph and plans. Then:
  - Empty → record a pass that says nothing was affected.
  - Full → run the declared `suite`.
  - Otherwise → run `cfg.command` per the contract above (through the existing `runLocalCheck`, as a `localCheck`
    named `affected` whose argv carries the `env` prefix), and record `sdlc-local/affected` with a status
    description like "12 packages, groups lua".

  `local-check run --name affected` dispatches here; the name is reserved for it in local-checks.json.

### Operating envelope (ARCH-CONSTRAINTS)

This is a batch path, run once per run. `go list` costs 1.5 s on 145 packages (measured), and the planner is
O(paths + packages × deps), well under a second. A wrong selection is the failure that costs: tests skipped on the
branch. The full suite still runs before landing (train or merge), so a mis-selection delays a red; it never lands
one.

### ARCH-ORDER / ARCH-FUNERAL / ARCH-SECURE

- **ARCH-ORDER**: no new state machine. A record is a commit status on an exact sha, the same as `suite`. The
  ordering risk is a stale selection after main moves. The train's full suite covers it, and the record names the
  merge-base it planned against.
- **ARCH-FUNERAL**: the only new durable rows are test-stats attempts with `Check: affected`, in the same store with the
  same retention. Logs go under the existing `local-checks/<sha>/` directory.
- **ARCH-SECURE**: the config comes from main (trusted the way local-checks.json is) and is parsed strictly. Changed
  paths come from git. Package names reach the command through the environment, never through a shell string.

## Milestones

- [ ] M1 — Planner and runner: `internal/affected` (Config, Graph, Plan), `loadGoGraph` with the fixture-module
  conformance test, `runAffectedCheck` with the empty/subset/full paths, `local-check run --name affected`, and the
  Done-when fixtures (a leaf change selects that package and its dependents only; a go.mod change runs and records
  `suite`). ariadne declares no affected.json in this milestone.
- [ ] M2 — Gates and flakes: `SuiteStandIn` in `decideCIGate`, set by `productionCarGates` only (fixture: a car with
  only `affected` boards, and merge without a train still requires `suite`). Then `isTestRun` across known-flake, the
  success guard and flake evidence; per-check duration rows; close/publish-gate wording (D2); atlas
  (`ci-merge-check.md` "Recorded local checks", `sdlc-binary.md` landing train); helptext `local-check.md`.
- [ ] Rollout (TL): rebuild :0; pair lands its affected.json and `test-affected` (D6); after a week, a test-stats
  snapshot of suite slot-minutes per landed change in pair goes in the Log.

## Chunk 1: M1 — planner and runner

### Task 1: `affected.Config` (strict parse)

**Files:**
- Create: `cmd/sdlc/internal/affected/config.go`, `cmd/sdlc/internal/affected/config_test.go`

- [ ] **Step 1: Write the failing tests.** Strategy: a table in which the Core-concepts example parses and each
  validation rule refuses one minimal violation. `FuzzMatch` compares the `**` matcher against a reference (expanding
  `**` to every segment count up to the path's depth, then `path.Match`), seeded with empty segments, leading, trailing
  and doubled `**`, and `path.Match` error forms (`[`, `\\`).
- [ ] **Step 2: Run it.** `go test ./cmd/sdlc/internal/affected/ -run TestParseConfig`. It fails: the package doesn't
  exist yet.
- [ ] **Step 3: Implement.**
  ```go
  // Package affected plans the affected suite (#473): which Go packages and declared test groups a change can break.
  // It does no IO; loadGoGraph (golist.go) is its one seam.
  package affected

  type Group struct {
      Name  string   `json:"name"`
      Paths []string `json:"paths,omitempty"`
      Go    []string `json:"go,omitempty"` // package patterns relative to the module root, e.g. "./cmd/pair"
  }

  type Config struct {
      Version        int      `json:"version"`
      Command        []string `json:"command"`
      Weight         int      `json:"weight,omitempty"`
      TimeoutMinutes int      `json:"timeout_minutes,omitempty"`
      Full           []string `json:"full,omitempty"`
      Groups         []Group  `json:"groups,omitempty"`
  }

  func ParseConfig(data []byte) (Config, error) // DisallowUnknownFields, single object, the validation above
  ```
  The globs use `Match(pattern, path)` in `glob.go`: `path.Match` per segment, plus `**` for zero or more segments.
- [ ] **Step 4: Run the tests until they pass.**
- [ ] **Step 5: Commit.** `git commit -m '#473 M1: affected.Config and globs' -- cmd/sdlc/internal/affected`

### Task 2: `affected.Graph` and `affected.Plan`

**Files:**
- Create: `cmd/sdlc/internal/affected/graph.go`, `plan.go`, and their `_test.go` files

- [ ] **Step 1: Write the failing tests.** Strategy: a literal Graph (leaf ← mid ← top, a test-only dependent of leaf,
  an unrelated package), with one table row per classification rule. The row the Done-when names is a leaf change
  selecting leaf plus its dependents only. The PQ-1 row is a group-matching file inside a package directory, which
  selects both. Further rows cover a group pulled in by its `go` root, Full beating everything else in one change set,
  and the unclassified fallback. A property test over random graphs and change sets checks that the selection is
  closed under dependents and never contains a package the change can't reach.
- [ ] **Step 2: Run them.** They fail.
- [ ] **Step 3: Implement.**
  ```go
  type Graph struct {
      Module string
      Dir    map[string]string          // import path → module-relative dir ("." for root)
      Deps   map[string]map[string]bool // import path → transitive in-module deps, test imports included
  }
  func (g Graph) PackageOf(file string) (string, bool) // deepest dir that is a package and contains file
  func (g Graph) Dependents(changed map[string]bool) []string // changed ∪ {p : Deps[p] ∩ changed ≠ ∅}, sorted

  type Selection struct {
      Full   bool
      Reason string
      GoPkgs []string
      Groups []string
  }
  func (s Selection) Empty() bool
  // Plan classifies each path by the rules in Core concepts (affected.Plan); package ownership and group
  // matching are additive, never first-match (PQ-1). isDoc is injected: the caller passes
  // !publishGateHasCodeSurface([]string{p}), so docs mean the same here as in the CI gate (ARCH-DRY).
  func Plan(paths []string, g Graph, cfg Config, isDoc func(string) bool) Selection
  ```
  A group's `go` roots resolve through `Dir`: `./top` names the package whose Dir is `top`. A root that names no
  package makes the planner return Full("group cli: ./top is not a package"), because a stale config must not silently
  skip a group.
- [ ] **Step 4: Run the tests until they pass.**
- [ ] **Step 5: Commit.** `#473 M1: affected.Plan over a package graph`

### Task 3: `loadGoGraph` (the `go list` seam) and its conformance test

**Files:**
- Create: `cmd/sdlc/internal/affected/golist.go`, `golist_test.go`, and `testdata/mod/` (go.mod `module example.com/m`;
  `leaf/leaf.go`; `mid/mid.go` importing leaf; `tonly/tonly.go` plus `tonly/tonly_test.go` importing leaf; `other/other.go`)

- [ ] **Step 1: Write the failing test.** `TestLoadGoGraphConformance` runs the real `go list` on `testdata/mod`
  (it is copied to `t.TempDir()` first, because testdata holds a nested go.mod) and asserts that `Deps["example.com/m/mid"]`
  has `leaf`, `Deps[…/tonly]` has `leaf` (through its test), `Deps[…/other]` is empty, and `Dir` maps each package.
  `TestParseGoList` on canned output pins the folding of test variants into their package.
- [ ] **Step 2: Run it.** It fails.
- [ ] **Step 3: Implement.**
  ```go
  // goListFormat prints one line per package (test variants included): import path, dir, ForTest, deps.
  const goListFormat = "{{.ImportPath}}\t{{.Dir}}\t{{.ForTest}}\t{{join .Deps \" \"}}"
  func ParseGoList(module, root, out string) (Graph, error) // pure
  // LoadGoGraph runs `go list -test -f goListFormat ./...` and `go list -m` in dir.
  func LoadGoGraph(ctx context.Context, dir string) (Graph, error)
  ```
  `FuzzParseGoList` feeds malformed lines (missing fields, stray ` [`, empty deps) and checks that the parse errors out
  or yields a graph whose Deps all name known packages. It never panics.
  A package's owner is its `ForTest` value, else its ImportPath with a `.test` suffix trimmed. In a Deps entry, ` [x.test]`
  is stripped, and only paths under the module are kept.
- [ ] **Step 4: Run the tests until they pass.**
- [ ] **Step 5: Commit.** `#473 M1: go list package graph seam`

### Task 4: read `affected.json` from main; reserve the name

**Files:**
- Modify: `cmd/sdlc/cigate.go` (`readAffectedConfig` beside `readLocalChecks`, `:408`), `cmd/sdlc/localchecks.go`
  (constants and the reserved name)
- Test: `cmd/sdlc/localchecks_test.go`

- [ ] **Step 1: Write the failing tests.**
  - `readAffectedConfig(git, rev)` returns `(cfg, true, nil)` when the file exists, `(_, false, nil)` when it is
    absent, and a refusal on a bad file. Use the same fake-git shape as the existing `readLocalChecks` tests.
  - `parseLocalChecks` refuses a check named `affected` ("reserved: declared in .sdlc/affected.json").
- [ ] **Step 2: Run them.** They fail.
- [ ] **Step 3: Implement.** Add `affectedCheckName = "affected"` and `affectedConfigPath = ".sdlc/affected.json"`, and
  extend the reserved-name check at `localchecks.go:110`.
- [ ] **Step 4: Run them until they pass.**
- [ ] **Step 5: Commit.** `#473 M1: read .sdlc/affected.json from main`

### Task 5: `runAffectedCheck` and `local-check run --name affected`

**Files:**
- Create: `cmd/sdlc/localcheck_affected.go`, `cmd/sdlc/localcheck_affected_test.go`
- Modify: `cmd/sdlc/localcheck_cmd.go:116-173` (`runLocalCheckCmd` dispatches the name `affected`)

- [ ] **Step 1: Write the failing tests.** Strategy: the existing local-check run harness in `localcheck_run_test.go`
  (a real git repo, a fake GitHub that records posted statuses, a fake suite helper), with one test per selection
  outcome (subset, full, empty, no config, graph unavailable). The subset test is the contract fixture:
  - **Contract fixture (the fake consumer the TL asked for):** the fixture repo holds the Task 3 module plus a
    `Makefile` whose `test-affected` target runs `printf '%s\n%s\n' "$$SDLC_AFFECTED_GO_PKGS" "$$SDLC_AFFECTED_GROUPS" > "$$OUT"`,
    and an `affected.json` with `command: ["make","test-affected"]` and group `lua`. A branch changes `leaf/leaf.go`
    and `x.lua`. The test asserts the file reads `example.com/m/leaf example.com/m/mid example.com/m/tonly` / `lua`,
    and that the status `sdlc-local/affected` was posted as success with the description `3 packages, groups lua`.
  For the other outcomes, the assertion is which status got posted (`suite`, `affected` with "nothing affected", none
  plus a refusal) and whether a fixture marker shows a command ran.
- [ ] **Step 2: Run them.** They fail.
- [ ] **Step 3: Implement.**
  ```go
  // affectedDeps are runAffectedCheck's seams; production wires git, affected.LoadGoGraph and runLocalCheck.
  type affectedDeps struct {
      changed func(sha string) ([]string, error) // gitx.ChangedPathArgs over merge-base(origin/main, sha)..sha
      graph   func(ctx context.Context, wt string) (affected.Graph, error)
      run     func(ctx context.Context, c localCheck, sha string) (localRunRecord, error) // runLocalCheck
      post    func(ctx context.Context, rec localRunRecord) error                      // postLocalCheckRecord
  }
  func runAffectedCheck(ctx context.Context, d localRunDeps, ad affectedDeps, cfg affected.Config, suite localCheck, sha string) (localRunRecord, error)
  // affectedLocalCheck builds the contract's argv: env SDLC_AFFECTED_GO_PKGS=… SDLC_AFFECTED_GROUPS=… cfg.Command…,
  // with suite's KnownFlakes policy (D4). Pure, with its own test.
  func affectedLocalCheck(cfg affected.Config, sel affected.Selection, suite localCheck) localCheck
  ```
  The graph is loaded from a detached worktree of `sha` (`detachedWorktreeWithCleanup`) because `go list` reads that
  tree's files. The worktree is released before the run, which makes its own.
- [ ] **Step 4: Run them until they pass**, along with `go test ./cmd/sdlc/ -run 'LocalCheck'`.
- [ ] **Step 5: Commit**, then run `sdlc milestone-close --issue 473 --milestone M1`, and tell the TL the contract is
  settled.

## Chunk 2: M2 — gates, flakes, wording

### Task 6: the car boards on `affected` (D3)

**Files:**
- Modify: `cmd/sdlc/cigate.go` (`ciInputs`, `decideCIGate`, `unrecordedRequired`), `cmd/sdlc/cigate_wait.go`
  (`ciGateFacts`), `cmd/sdlc/landingcar.go:90-104`
- Test: `cmd/sdlc/cigate_test.go`, `cmd/sdlc/landingcar_test.go`

- [ ] **Step 1: Write the failing tests.**
  - `decideCIGate` with `StandIn{"suite":"affected"}`: a recorded `affected` satisfies a required `suite`; so does an
    `affected` record on a docs-only ancestor. Without StandIn, the same input still needs `suite`.
  - In `TestCarReadiness`, a car whose head has only `sdlc-local/affected` is ready when main has `affected.json`, and
    not ready when it doesn't.
  - In the merge CI gate tests, a head with only `affected` still runs `suite` at landing (the runner is called with
    `suite`).
- [ ] **Step 2: Run them.** They fail.
- [ ] **Step 3: Implement.** Add `StandIn map[string]string` to `ciInputs` and `ciGateFacts`. In `decideCIGate`, a
  required `c` counts as recorded when `recorded[localStatusContext(c.Name)] || recorded[localStatusContext(StandIn[c.Name])]`.
  `unrecordedRequired` also returns the stand-in's name, so the ancestor walk looks for it. `productionCarGates`
  prepares its facts (`prepareCIGateFactsContext`) and sets `StandIn` when `readAffectedConfig(main)` finds a file.
  Merge's path never sets it.
- [ ] **Step 4: Run them until they pass**, along with `go test ./cmd/sdlc/ -run 'CIGate|Car|Landing'`.
- [ ] **Step 5: Commit.** `#473 M2: a car boards on the affected record`

### Task 7: `affected` is a test run for flakes and stats (D4)

**Files:**
- Modify: `cmd/sdlc/teststats.go:99-107`, `localcheck_run.go:87,136,184`, `localcheck_republish.go:154,198`,
  `localchecks.go:140`; the `test-stats` aggregate (`teststats.go:111+`) and its printer
- Test: `cmd/sdlc/localcheck_flakes_test.go`, `localcheck_republish_test.go`, `teststats_test.go`

- [ ] **Step 1: Write the failing tests.** Strategy: rerun each existing suite-only test of the four predicates
  (known-flake step, success guard, flake evidence, aggregates) with an `affected` attempt. Add one cross-check test:
  a failure under one check never blocks posting the other, and suite aggregates are unchanged by affected rows.
- [ ] **Step 2: Run them.** They fail.
- [ ] **Step 3: Implement.**
  - Add `isTestCheck(name) = name == suiteCheckName || name == affectedCheckName`.
  - `isFlakeEvidence` and the known-flake step (`localcheck_run.go:184`) use it.
  - The success guard compares the attempt's check with the posted one.
  - `isSuiteRun` keeps meaning the full suite, for the weekly stats and the main baseline.
  - `RepoStats` gains `Check string json:"check,omitempty"` (additive; the suite rows leave it empty), with one
    aggregate per (repo, check).
- [ ] **Step 4: Run them until they pass**, along with `go test ./cmd/sdlc/ -run 'Flake|TestStats|Republish|Weekly'`.
- [ ] **Step 5: Commit.** `#473 M2: affected counts as a test run`

### Task 8: wording, help, atlas (D2)

**Files:**
- Modify: `cmd/sdlc/close.go:1905` (the VERIFIED example in `explainVerified`), `cmd/sdlc/publishgate.go:302-308`,
  `cmd/sdlc/helptext/local-check.md`, `cmd/sdlc/helptext/merge.md:140-160`, `atlas/workflow/ci-merge-check.md`
  ("Recorded local checks"), `atlas/workflow/sdlc-binary.md` (landing train paragraph), `AGENTS.base.md:62`
  (base layer: say that close's evidence is the affected suite where a repo declares one)
- Test: the existing helptext/golden tests (`go test ./cmd/sdlc/ -run 'Help|Explain'`)

- [ ] **Step 1:** Update the text so that close's VERIFIED example reads "sdlc local-check run --name affected (or
  suite)", and the help documents the affected.json schema and the env contract.
- [ ] **Step 2:** Run the golden tests and update their expectations.
- [ ] **Step 3:** Commit, run the full suite, then `sdlc close --issue 473`.

## Verification

- Done-when fixtures: Task 2 (leaf plus dependents only), Task 5 (go.mod → suite; the contract fixture), Task 6 (a car
  boards on affected). The close evidence is the recorded `suite` on the final head.
- The one-week pair measurement belongs to the rollout. The TL owns pair's counterpart, and the snapshot (sum of
  `wall_seconds` over pair's suite, affected and suite-rerun attempts ÷ merges to main in the window, compared with the
  week before) goes in this issue's Log.
