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

## Design status (this revision is for the TL)

This revision fixes the concepts, the decisions and the milestone shape. The step-level tasks (code, commands) are
written after the TL approves the decisions below, then `sdlc change-code` reviews the whole plan.

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

### Decisions for the TL

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
- **affected.Plan(paths, graph, cfg) Selection**: classifies each changed path. The first matching rule wins:
  1. Docs (the existing `publishGateHasCodeSurface` rule) → ignored.
  2. Matches `full` → Full, with the path as the Reason.
  3. A `.go` file, or any file inside a package directory (testdata, embedded assets) → its nearest package.
  4. Matches a group's `paths` → that group.
  5. Anything else → Full ("unclassified: path").

  It then expands the packages to their dependents and selects the groups whose `go` roots depend on them. It is pure,
  with a table test.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `loadGoGraph` | `cmd/sdlc/internal/affected/golist.go` | new | `go list -test -json ./...` |
| `runAffectedCheck` | `cmd/sdlc/localcheck_affected.go` | new | worktree, git diff, runner |
| `readAffectedConfig` | `cmd/sdlc/localchecks.go` | new | `git show origin/main:.sdlc/affected.json` |
| `productionCarGates` | `cmd/sdlc/landingcar.go` | modified | CI gate (stand-in) |

- **loadGoGraph**: runs `go list` in the head's detached worktree; a parse or exit failure means Full ("package graph
  unavailable"). ARCH-MOCK: `go` is the toolchain the suite itself runs. The seam is a function value
  (`func(dir) (Graph, error)`), and the planner tests use literal Graphs. A conformance test runs the real `go list` on
  a fixture module of three packages (a leaf, a dependent, and a test-only dependent) and checks the graph.
- **runAffectedCheck**: in the existing detached worktree, it diffs `merge-base(origin/main, sha)..sha`
  (`gitx.ChangedPathArgs`), loads the graph and plans. Then:
  - Empty → record a pass that says nothing was affected.
  - Full → run the declared `suite`.
  - Otherwise → run `cfg.command` under `runUnderSlot` with `SDLC_AFFECTED_GO_PKGS` (space-separated import paths)
    and `SDLC_AFFECTED_GROUPS` (comma-separated) in the environment, and record `sdlc-local/affected` with a status
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
