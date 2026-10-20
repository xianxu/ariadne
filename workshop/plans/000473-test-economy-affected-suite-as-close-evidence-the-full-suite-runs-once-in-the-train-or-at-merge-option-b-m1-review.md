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

---

## Re-review — 2026-10-20T02:05:58-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 473 — test-economy: affected-suite as close evidence; the full suite runs once, in the train or at merge (option B') |
| repo | ariadne |
| issue file | workshop/issues/000473-test-economy-affected-suite-as-close-evidence-the-full-suite-runs-once-in-the-train-or-at-merge-option-b.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 0956e5e1c71bd6734f195929fc151e0c752dbc14..387922e63d8c9d31022c1481dc6b0aa83a1fa270 |
| command | sdlc milestone-close --issue 473 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-20T02:05:58-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I've read the fix commit and the current sources. Next I'll check the regression tests and run the affected package tests.

Both regression suites pass. Now I'll revert each fix in this disposable checkout and confirm its test goes red.

Both reverts go red, so BR-1 and BR-2 have real regression tests. Next I'm checking an edge case: a file embedded by one package but living in another package's directory.

I've finished the review. **Verdict: FIX-THEN-SHIP.** Three of the six earlier findings are fully fixed, two have no regression test, and one fix opened a new, narrower gap.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The fix commit `387922e63` deals with all six earlier findings:
- **Fixes I confirmed:**
  - **BR-1:** a root package no longer claims every changed path. A package now owns only files in its own directory, under its `testdata/`, or files it embeds.
  - **BR-2:** a head with no change against main now refuses instead of recording a pass.
  - **BR-3:** the README now covers `--name affected` and `.sdlc/affected.json`.
  - **BR-6:** the Log now records that the D4 flake behaviour isn't live until M2.
- **What I ran:**
  - I reverted the BR-2 guard: `TestAffectedNoChangeRefuses` fails.
  - I put back the old walk-up-to-ancestor ownership: `TestPlan`, `TestPlanRootPackageDoesNotClaimSubdirectories`, `TestPackageOfIsExact` and `TestLoadGoGraphConformance` all fail.
  - With the fixes in place, `go test ./cmd/sdlc/internal/affected/ ./cmd/sdlc/internal/jsonstrict/` passes, and so does `go test ./cmd/sdlc/ -run 'TestAffected|TestReadAffectedConfig|TestLocalCheckCmd'`.
- **What's still open:**
  - **The new gap:** the ownership lookup in `packageIn` stops at the first owner it finds. A file inside one package's directory but embedded by another package selects only the first one. I checked that Go allows this: with the real `go list`, a root package embedding `sub/data.txt` gives `ex.com/m [sub/data.txt]`. That breaks the plan's PQ-1 rule that selectors add up, and under-selects tests. The full suite in the train or at merge still catches it, so it delays a failure rather than landing one.
  - **No regression test:** the `--all` conflict fix (BR-4) and the duplicate-key check (BR-5) both lack a test that fails without them. BR-5's tests cover only the `affected.json` path, which is all the finding asked for. Both are Minor.

### 1. Strengths
- **Package ownership is now exact and checked end to end.** The fixture module embeds `leaf/assets/banner.txt` and also holds a plain `leaf/notes/plain.txt`, and the conformance test runs the real `go list` to confirm that only the embedded file is claimed (`golist_test.go:127-132`).
- **The new embed parsing is defensive.** Absolute and `..` embed paths are dropped, embeds that point at an unknown package are pruned (`golist.go:52-60, 85-89`), and the fuzz test now asserts that no embed escapes the module.
- **The no-change refusal comes before anything is posted.** The test asserts that no status was posted, no consumer ran and no suite ran (`localcheck_affected_test.go:154-169`).
- **`jsonstrict.RejectDuplicateKeys` catches duplicate keys at any depth.** It is tested on nested objects and arrays, and `ParseConfig` calls it before the strict decode.

### 2. Critical findings
None.

### 3. Important findings
- **New: package ownership stops at the first owner instead of adding up every owner** (`graph.go:31-46`, `graph.go:13`).
  - `packageIn` returns as soon as the own-directory check matches, then the `testdata/` check, and only then looks at embeds.
  - `Graph.Embeds` maps each file to a single package, so if two packages embed the same file, one silently replaces the other.
  - Concrete case: the root package embeds `sub/data.txt` while `sub` is its own package. Changing `sub/data.txt` selects only `ex.com/m/sub`, so the root package's tests are skipped.
  - **This is the 2nd finding in family `unclassified-path-fails-closed`.** The rule that covers both: *a changed path selects the union of everything that claims it (its own-directory package, its `testdata/` owner, every package that embeds it, every group whose paths match); it is Full only when that union is empty, and no lookup stops at the first hit.*
  - Fix sketch:
    - Make `Embeds` a `map[string][]string`, so a file can have several embedding packages.
    - Make `packageIn` return a set of owners.
    - In `Plan`, add every owner to `changed`.
    - Add table rows for an embedded file inside another package's directory, and for one file embedded by two packages.

### 4. Minor findings
- The BR-4 fix (`localcheck_cmd.go:145`) has no regression test. It also repeats the error string from `selectLocalChecks` (`localcheck_cmd.go:98`), which is an ARCH-DRY nit.
- The plan text still describes the old ownership rule, and the plan has no Revisions entry for the change:
  - Core concepts rule 3 says a file "selects its nearest package".
  - In Task 2, the `PackageOf` comment says "deepest dir that is a package", and the `Graph` struct has no `Embeds` field.
  - The atlas already has the new rule.

### 5. Test coverage notes
- **BR-1 and BR-2:** the regression tests are real and fail when the fix is reverted (shown above).
- **BR-5:** covered by `config_test.go:44-45` and `jsonstrict_test.go`.
- **The new gap:** there's no test of an embedded file inside a different package's directory.
- **Build tags:** files excluded by build tags are still not modelled in the graph. The full suite in the train covers this, but it isn't documented.

### 6. Architectural notes
- **ARCH-DRY: pass, with a note.** `internal/jsonstrict` becomes a fourth copy of the duplicate-key scanner, alongside `observe/json.go:159`, `fleet/types.go:727` and `teststats_store.go:313`. Consolidating them is filed as #489. That is a separate change, not the point of this issue, so deferring it is acceptable.
- **ARCH-PURE: pass.** The embed folding is in the pure `ParseGoList`, and ownership is pure on `Graph`.
- **ARCH-PURPOSE: flag** (the new finding). The selectors are supposed to add up (PQ-1), and the ownership lookup still stops at the first match.
- **ARCH-MOCK: pass.** The conformance test runs the real `go list` and now covers the embed fields.
- **ARCH-CONSTRAINTS: pass.** The extra `go list` template fields cost almost nothing.
- **ARCH-SECURE: pass.** Embed paths from `go list` are sanitised, and the empty change set no longer produces fake passing evidence.
- **ARCH-ORDER: pass.** Nothing is held between events.
- **ARCH-FUNERAL: pass.** Nothing new and durable is created.
- **ARCH-EVOLVE: pass.** `affected.json` now rejects duplicate keys, alongside unknown fields.
- **For M2:** the stand-in should still refuse a docs-only `affected` record on a head whose own diff contains code.

### 7. Plan revision recommendations
- Add a "## Revisions" entry for M1 BR-1: a package owns its own directory, its `testdata/`, and the files it embeds (from `go list`'s embed fields); a plain subdirectory file is unclassified. Also add `Embeds` to `Graph` in Task 2 and fix the `PackageOf` comment.
- Core concepts rule 3: say the owners add up, as the union of every claimant, including every package that embeds the file.
- Task 5: add the "empty change set refuses" outcome.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      packageIn is now exact (own dir, testdata/, embeds); TestPlanRootPackageDoesNotClaimSubdirectories and TestPackageOfIsExact go red when the old ancestor walk is restored.
  - id: BR-2
    disposition: addressed
    note: |
      runAffectedCheck refuses on empty paths before any post; TestAffectedNoChangeRefuses goes red with the guard disabled.
  - id: BR-3
    disposition: addressed
    note: |
      README.md:569-582 documents --name affected, .sdlc/affected.json, both env vars and all four outcomes; it links the atlas anchor, which exists (ci-merge-check.md:307).
  - id: BR-4
    disposition: not-addressed
    note: |
      localcheck_cmd.go:145 restores the refusal, but no test runs --all with --name affected, so nothing fails without the fix.
  - id: BR-5
    disposition: addressed
    note: |
      ParseConfig calls jsonstrict.RejectDuplicateKeys; config_test.go covers top-level and nested group duplicates.
  - id: BR-6
    disposition: addressed
    note: |
      The issue Log's M1 review round 1 entry states that D4 is inert until M2 and that a docs-only record cannot be republished.
findings:
  - id: new
    severity: Important
    kind: defect
    family: unclassified-path-fails-closed
    title: |
      packageIn stops at the first owner, so a file embedded by one package but inside another package's dir skips the embedder
    detail: |
      graph.go:31-46 returns on the own-dir or testdata match before looking at Embeds, and Embeds is a single-valued map (graph.go:13). Verified with go list: a root package embedding sub/data.txt, where sub is its own package, is legal, yet Plan selects only sub. Rule for the family: a path selects the union of every claimant (own-dir package, testdata owner, every embedder, every matching group), and is Full only when that union is empty. Fix: make Embeds map[string][]string, have packageIn return a set, and add table rows for both shapes.
  - id: new
    severity: Minor
    kind: defect
    family: plan-tracks-code
    title: |
      The plan still describes nearest-ancestor package ownership; no Revisions entry for the exact-ownership change
    detail: |
      Core concepts rule 3 says "selects its nearest package", Task 2 says PackageOf is the "deepest dir that is a package", and the Graph struct has no Embeds field. The code and atlas now use exact ownership (own dir, testdata/, embeds).
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"cf024732-aa00-4e83-ad77-7c77f75d9d74","repository":"github.com/xianxu/ariadne","issue_id":"000473","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/d546c3d9773d17bf","reviewed_head":"387922e63d8c9d31022c1481dc6b0aa83a1fa270","slot":"ariadne:2","requested_at":"2026-10-20T09:03:23.984159Z"},"phase":"finished","pid":58208,"started_at":"2026-10-20T09:03:24.133797Z","finished_at":"2026-10-20T09:05:58.238916Z","session_id":"819a2b26-5aab-4d8f-a094-0d8e187c33ff","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T02:14:01-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 473 — test-economy: affected-suite as close evidence; the full suite runs once, in the train or at merge (option B') |
| repo | ariadne |
| issue file | workshop/issues/000473-test-economy-affected-suite-as-close-evidence-the-full-suite-runs-once-in-the-train-or-at-merge-option-b.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 0956e5e1c71bd6734f195929fc151e0c752dbc14..f4a4a688e1499c1f34e910719f33929d962b619b |
| command | sdlc milestone-close --issue 473 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-20T02:14:01-07:00 |
| verdict | SHIP |

## Review

Focus on the open findings first: BR-4, BR-7, BR-8.

```verdict
verdict: SHIP
confidence: high
```

All three open findings are fixed, and I checked each one by breaking the fix and watching its test fail. BR-7 now returns every owner of a file instead of stopping at the first. BR-4 makes `--all` combined with `--name affected` refuse with an error. BR-8 adds a dated Revisions entry to the plan. Both code fixes have tests that go red when the fix is reverted. I ran the affected-package, jsonstrict and local-check tests at the pinned head and they passed. I found nothing new that would block the boundary.

1. **Strengths**
   - `cmd/sdlc/internal/affected/graph.go:34-50`: `owners` builds the union of three sources: the file's own directory, any `testdata/` ancestor, and every package that embeds it. That is the class-level rule BR-7 asked for, not a fix for the one case it named. A path is unclassified, and forces the full suite, only when nothing claims it.
   - `cmd/sdlc/internal/affected/golist.go:51-58,82-90`: `Embeds` lists every embedding package. It drops paths that escape the module and drops packages the listing doesn't know, then sorts the result so the output is stable. The fuzz test holds that invariant.
   - The BR-7 shape is checked against real `go list`: `testdata/mod/m.go` has the root package embed `other/data.txt` (checked in `golist_test.go:136`). It is also covered in the pure tables (`plan_test.go:80`, `TestOwnersAreExactAndAUnion`).
   - `cmd/sdlc/localcheck_cmd.go:98-104`: `splitAffectedName` is a small pure helper with its own unit test, so the `--all` conflict is decided in one place.

2. **Critical:** none.

3. **Important:** none.

4. **Minor:** The plan body still says "nearest package" (line 127) and describes `PackageOf` (line 247). The Revisions entry at line 427 explicitly supersedes both, which is what the plan convention asks for, so this is not a finding.

5. **Test coverage notes**
   - **BR-7:** I changed `owners` back to returning the first match. Three tests then failed: `TestLoadGoGraphConformance`, `TestPlan/...BR-7` and `TestOwnersAreExactAndAUnion`.
   - **BR-4:** I disabled the `all && len(names) > 0` guard and `TestSplitAffectedName` failed.
   - Both files were restored afterwards and the checkout is clean.
   - **Full runs:** `go test ./internal/affected/ ./internal/jsonstrict/` and `go test -run 'Affected|LocalCheck|SplitAffected|ParseLocalChecks' ./cmd/sdlc` both pass.

6. **Architecture**
   - **ARCH-DRY: pass.** Ownership has a single implementation (`owners`), shared by `Owners` and `Plan`.
   - **ARCH-PURE: pass.** `Plan`, `Owners` and `ParseGoList` are pure and tested without IO. `LoadGoGraph` and `productionAffectedDeps` are the thin IO layer.
   - **ARCH-PURPOSE: pass.** BR-7 was fixed at the level of the rule (a union of all claimants), not just the reported case.
   - **ARCH-MOCK: pass.** The `affectedDeps` seams are injected, and a live conformance test runs real `go list` against `testdata/mod`.
   - **ARCH-CONSTRAINTS: pass.** The only extra cost is one `go list` per run. Any failure falls back to the full suite.
   - **ARCH-SECURE: pass.** `go list` output is parsed strictly and fails closed. The command runs as an argument list through `env`, never through a shell.
   - **ARCH-ORDER: pass.** The command keeps no state between events. Each run plans once and executes once.
   - **ARCH-FUNERAL: pass.** The temporary worktree used to load the graph is removed by a deferred cleanup. No new durable artifacts are created.
   - **ARCH-EVOLVE: pass.** `.sdlc/affected.json` is a new, strict, versioned file in its own path. Older sdlc builds never read it, and the name `affected` is reserved in `local-checks.json`.

7. **Plan revisions:** none needed. The 2026-10-20 entry already records exact, union-based ownership.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      splitAffectedName refuses --all with any --name (localcheck_cmd.go:98-104); disabling the guard makes TestSplitAffectedName fail.
  - id: BR-7
    disposition: addressed
    note: |
      Embeds is map[string][]string and owners returns the union of own-dir, testdata ancestors and every embedder; reverting to first-match fails TestLoadGoGraphConformance (live go list), TestPlan BR-7 row and TestOwnersAreExactAndAUnion.
  - id: BR-8
    disposition: addressed
    note: |
      Plan Revisions entry dated 2026-10-20 (plan.md:427-436) explicitly replaces "nearest package" (rule 3) and Task 2 PackageOf with Graph.Owners union semantics, matching graph.go:24-50.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"48a54735-a380-40bb-b9d2-57d92641e825","repository":"github.com/xianxu/ariadne","issue_id":"000473","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/acd521bd8eac615f","reviewed_head":"f4a4a688e1499c1f34e910719f33929d962b619b","slot":"ariadne:2","requested_at":"2026-10-20T09:08:18.759877Z"},"phase":"finished","pid":68795,"started_at":"2026-10-20T09:08:18.845901Z","finished_at":"2026-10-20T09:13:59.937446Z","session_id":"cdbb0124-abc4-46fd-b39f-cf762c1bbdf6","outcome":"returned"}] -->
