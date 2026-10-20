# Boundary Review — ariadne#473 (milestone M1)

| field | value |
|-------|-------|
| issue | 473 — test-economy: affected-suite as close evidence; the full suite runs once, in the train or at merge (option B') |
| repo | ariadne |
| issue file | workshop/issues/000473-test-economy-affected-suite-as-close-evidence-the-full-suite-runs-once-in-the-train-or-at-merge-option-b.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 0956e5e1c71bd6734f195929fc151e0c752dbc14..36f16ff1e7a6c5b775a86aaaea601918089d89a9 |
| command | sdlc milestone-close --issue 473 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-20T01:57:31-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing the #473 M1 diff: the planner, the `go list` seam, and the runner are read. Now I'm checking how the runner handles edge cases and running the tests.

I've finished the review. **Verdict: FIX-THEN-SHIP** (medium confidence): there are no Critical findings, but three Important ones should be fixed before M1 closes.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M1 delivers what the plan claims. That covers the strict `affected.Config` parser and the `**` glob matcher, the package graph and planner, and the `go list` seam with a conformance test against a fixture module. It also covers the `affected` name reservation, `readAffectedConfig`, and `runAffectedCheck` with its subset, full and docs-only paths. All of it is wired into `local-check run --name affected`. Every Core-concepts entity scheduled for M1 is at its stated path. `SuiteStandIn`, `isTestRun` and `productionCarGates` are correctly left for M2.

I ran the tests and they pass: `go test ./cmd/sdlc/internal/affected/` passed, `go test ./cmd/sdlc/ -run 'Affected|ParseLocalChecks|LocalCheck'` passed (272 s), and `go vet` was clean.

Three things keep this from SHIP:
- **Root-package bug:** in a module whose root directory is a Go package, the "unclassified → full suite" safety valve never fires.
- **Empty diff recorded as a pass:** an empty change set is recorded as a passing "docs-only" `affected` status without anything running.
- **README:** it doesn't cover the new surface, and the plan never schedules it.

### 1. Strengths
- **The planner is pure, and its tests prove it.** `Plan` takes the docs rule as a parameter (`plan.go:25`), and the caller passes `publishGateHasCodeSurface` (`localcheck_affected.go:37`), so "docs" means the same here as in the CI gate. The table test and a 500-case property test (`plan_test.go:138`) check that the selection is closed under dependents and never too wide. No mocks are needed.
- **The `go list` parser is careful.** It folds test variants into their package through `ForTest` and the `.test` suffix, and drops dependencies outside the module (`golist.go:18-77`). The conformance test runs the real `go list`, and the parser is fuzzed.
- **Failures fail closed.** A graph that won't load (`localcheck_affected.go:34`), a stale `go` root (`plan.go:77`), a bad glob and a strict config refusal all fall back to the full suite or a refusal, never to an empty selection.
- **The env contract is passed as an argument list through `env(1)`, never a shell string** (`localcheck_affected.go:54`). The contract fixture checks both variables end to end through a fake `make` consumer.
- **D1 keeps older sdlc binaries working:** the separate `.sdlc/affected.json` is invisible to them.

### 2. Critical findings
None.

### 3. Important findings
- **I1. A root package claims every path, so the full-suite safety valve is defeated** (`graph.go:30-39` `packageIn`, used at `plan.go:44`; the behaviour is pinned by `plan_test.go:125`, where `b/x.go` maps to the root package).
  - When the module root is a package (`Dir "."`), the walk up from any file ends at the root and claims it.
  - So `scripts/deploy.sh` or `tests/lib/helper.sh` selects the root package and its dependents, usually nothing more, instead of the full suite.
  - That contradicts the Spec ("anything unclassifiable → full suite") and plan rule 4.
  - Fix sketch: only the package's own directory, its `testdata/` subtree, and non-package subdirectories of a non-root package can claim a file. Or exclude `Dir "."` from claiming anything beyond its direct children and `testdata/`.
  - Add a table row: a root package plus `scripts/x.sh` gives Full. (ARCH-PURPOSE, ARCH-SECURE.)
- **I2. An empty diff records a "docs-only" pass without running anything** (`localcheck_affected.go:33,43-46`).
  - When `--rev` is at or behind `origin/main` (for example a freshly pushed branch with no commits yet), merge-base equals the sha and the change set is empty.
  - `publishGateHasCodeSurface([])` is false, so a success status reading "nothing affected (docs-only)" is posted on what may be a code commit.
  - That is a made-up value that M2's stand-in and ancestor walk will read as evidence.
  - Fix sketch: refuse when `paths` is empty ("no change against origin/main; nothing to plan"), or label it distinctly and don't post. Add a test with the sha equal to main. (ARCH-SECURE.)
- **I3. README update appears missing for `sdlc local-check run --name affected` and `.sdlc/affected.json`.**
  - `README.md:557-622` documents `local-check run` and the close suite, but not the new name or config file.
  - The plan's Task 8 lists help text and atlas but not README, so this gap would survive M2 too.
  - Fix sketch: add README to Task 8's file list (a plan revision) or update it now.

### 4. Minor findings
- `--all --name affected` silently drops `--all`, because the `selectLocalChecks` call is skipped (`localcheck_cmd.go:148`). Before this change, combining the two flags was an error.
- The docs-only record saves no attempt, so it has no test-stats row and `republish` can't recover it if the status post fails.
- `ParseConfig` accepts duplicate JSON keys (last one wins), while ARCH-EVOLVE asks for strict handling of duplicate keys.
- Test files behind build tags (for example `//go:build integration`) are excluded from `go list` dependencies, so their imports don't create dependents. The full suite catches this, but it isn't documented.
- The `affected` check carries the suite's `KnownFlakes` policy, but the runner's known-flake step and success guard still key on the name `suite`. This is inert until M2 Task 7; worth stating in the Log so nobody assumes the D4 behaviour is already live.

### 5. Test coverage notes
- What's covered: the planner (table plus property test), the glob matcher (fuzzed against a reference), the parser (fuzzed), the real `go list` conformance test, and all three runner outcomes against a real git repo with a fake GitHub that records statuses.
- Gaps matching I1 and I2: no test of a root package with a non-Go path, and none of an empty change set.
- Done-when fixtures for M1 are present: the leaf-change selection (`plan_test.go:58`), go.mod forcing the suite (`localcheck_affected_test.go`), and the contract fixture.

### 6. Architectural notes
- **ARCH-DRY: pass.** The docs rule is reused via injection, `ChangedPathArgs` and `detachedWorktreeWithCleanup` are reused, and the group-name rule matches the local-check one.
- **ARCH-PURE: pass.** Config, glob, graph and plan are pure; `LoadGoGraph` and `productionAffectedDeps` are the thin IO layer.
- **ARCH-PURPOSE: flag (I1).** The unclassified fallback is the Spec's safety valve, and a root package defeats it.
- **ARCH-MOCK: pass.** The `go` toolchain sits behind a function seam and has a live conformance test. git is used for real in a temporary repo; GitHub is a stateful fake.
- **ARCH-CONSTRAINTS: pass.** It's a batch path; the graph is loaded once per run and the run uses the existing suite slots, weight and timeout.
- **ARCH-SECURE: flag (I2).** An empty diff becomes passing evidence. Otherwise the config is parsed strictly from main and package names reach the command as arguments.
- **ARCH-ORDER: pass.** No state is held between events; each record is a commit status on an exact sha.
- **ARCH-FUNERAL: pass.** The graph worktree is removed after loading, and logs use the existing pruned directory.
- **ARCH-EVOLVE: pass, with the minor duplicate-key note.** A new file with `version: 1` and strict parsing; `local-checks.json` stays unchanged, except that the name `affected` becomes reserved.
- **For M2:**
  - The stand-in must not let the I2-style empty record vouch for a code head.
  - The success guard and flake rerun must key on the check name before `affected` records count for a car.

### 7. Plan revision recommendations
- Task 8: add `README.md` (the `local-check run --name affected` surface and `.sdlc/affected.json`).
- Core concepts, rule 3: say that a root package (or any ancestor directory) claims only its own files and its `testdata/`, so rule 4's full-suite fallback still applies to unrelated paths.
- Task 5: add an "empty change set" outcome (refuse) alongside subset, full, docs-only, no config and graph unavailable.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: unclassified-path-fails-closed
    title: |
      A root Go package claims every changed path, so unclassified files never force the full suite
    detail: |
      packageIn (graph.go:30) walks up to Dir "." and Plan (plan.go:44) marks the path claimed, as pinned by plan_test.go:125 where b/x.go maps to the root package. In a module whose root is a package, scripts/ and tests/lib/ changes select the root package and its dependents instead of Full, against the Spec's unclassified-to-full-suite valve. Fix: limit ancestor claims (own dir plus testdata/, never the root for subdirectories) and add a root-package plus non-Go path table row expecting Full.
  - id: new
    severity: Important
    kind: defect
    family: no-evidence-recorded-as-pass
    title: |
      An empty change set (rev at or behind origin/main) posts a "docs-only" affected success without running anything
    detail: |
      In runAffectedCheck (localcheck_affected.go:33,43), publishGateHasCodeSurface of an empty path list is false, so a branch tip equal to main records sdlc-local/affected success "nothing affected (docs-only)" on a possibly-code commit. That is fabricated evidence the M2 stand-in and ancestor walk will trust. Refuse (or post nothing) when paths is empty, and add a test with sha equal to main.
  - id: new
    severity: Important
    kind: defect
    family: docs-gate-new-surface
    title: |
      README update appears missing for local-check run --name affected and .sdlc/affected.json
    detail: |
      README.md:557-622 documents local-check run and the close suite but not the new name or config file. Plan Task 8 lists help text and atlas only, so README would be missed in M2 as well. Add README.md to Task 8 (plan revision) or update it now.
  - id: new
    severity: Minor
    kind: defect
    family: cli-flag-conflict
    title: |
      --all combined with --name affected silently drops --all
    detail: |
      localcheck_cmd.go:148 skips selectLocalChecks when only affected is named, so the existing all-versus-names conflict error no longer fires.
  - id: new
    severity: Minor
    kind: defect
    family: strict-persisted-format
    title: |
      affected.ParseConfig accepts duplicate JSON keys (last one wins)
    detail: |
      ARCH-EVOLVE asks that duplicate keys stay strict; encoding/json silently takes the last value.
  - id: new
    severity: Minor
    kind: defect
    family: deferred-scope-visibility
    title: |
      The affected check carries KnownFlakes, but the flake step and success guard still key on "suite" until M2
    detail: |
      This is inert in M1 (localcheck_run.go:136,184 and localcheck_republish.go:198). The docs-only record also saves no attempt, so it is not republishable. Note both in the Log so D4 is not assumed live.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"6565c97c-7663-4161-a479-04a5d599c1d3","repository":"github.com/xianxu/ariadne","issue_id":"000473","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/2c46c994bc92bfc1","reviewed_head":"36f16ff1e7a6c5b775a86aaaea601918089d89a9","slot":"ariadne:2","requested_at":"2026-10-20T08:49:20.730538Z"},"phase":"finished","pid":82816,"started_at":"2026-10-20T08:49:20.876995Z","finished_at":"2026-10-20T08:57:31.108352Z","session_id":"6f9f2218-be5e-4d86-9158-14cdee8133d6","outcome":"returned"}] -->
