# Boundary Review — ariadne#407 (whole-issue close)

| field | value |
|-------|-------|
| issue | 407 — A curated Go lint catalog: golangci-lint + shared ruleguard rules for known mistakes (per-test builds, sleeps, wall-clock asserts, temp dirs, strict wire decode) |
| repo | ariadne |
| issue file | workshop/issues/000407-a-curated-go-lint-catalog-golangci-lint-shared-ruleguard-rules-for-known-mistakes-per-test-builds-sleeps-wall-clock-asserts-temp-dirs-strict-wire-decode.md |
| boundary | whole-issue close |
| milestone | — |
| window | e5f48d8326eab4ce46e7c8fae2d89ae96498f431..21453b203f07329e7153db3fa97a171fde0f6be2 |
| command | sdlc close --issue 407 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-18T00:59:41-07:00 |
| verdict | REWORK |

## Review

Checking the existing code and running the golint tests now.

```verdict
verdict: REWORK
confidence: medium
```

The four-rule catalog is well built and works. `pkg/golint` is stdlib-only and every finding cites its incident. A suppression must give a reason, and a suppression that names an unknown rule or matches nothing is itself a finding. The repo test passes over 1,143 tracked files in about 0.26 s. The judge, gitx, tl-worker and tlrules packages also pass at head. I rated two findings Critical, which makes the verdict REWORK; neither is a runtime bug. First, the plan's Core concepts table and integration claims don't match the code. The review protocol rates any such contradiction Critical, though the fix is a plan revision only. Second, one of the Spec's seed rules (wall-clock duration assertions) was dropped without saying so. It is not among the three approved deviations, and the tree has 14 such assertions. There is also a coverage regression: the gitx deadline test no longer exercises the default `gitCommandTimeout` limit.

**1. Strengths**
- `pkg/golint/check.go:66-99`: suppression checks (reason required, unknown rule, unused) are built into the engine, so stale suppressions can't build up. This covers what nolintlint would have done.
- `pkg/golint/rules_test.go`: each rule has tests that should flag and tests that shouldn't, including the edge cases (`:=` vs `=`, `FromSlash`, `time.Hour`, polling under 100 ms). Each test also checks that the message cites the incident.
- Three tests that raced readiness now expire on cue through `testfix.DeadlineOnCue` + `AwaitPIDFile` (`gitx/lifetime_test.go`, `judge/preparation_extra_test.go:104-138`). This fixes the class, not just the instance.
- The process rule is written once, in `pkg/tlrules/planning.md`, and guarded by a phrase in `rules.tsv`. Both TL helps carry it from that single source (ARCH-DRY).

**2. Critical findings**
- **The plan's Core concepts and integration claims contradict the code.** The plan places `Rule`/`Finding` in `pkg/golint/rule.go` and `suppressions` in `pkg/golint/suppress.go`. Neither file exists; everything is in `check.go`. The plan gives `Match func(*ast.File, *token.FileSet) []token.Pos`, but the code has `func(f *ast.File) []ast.Node`. The plan says the repo test "walks `cmd/` and `pkg/`, skips testdata/ and generated files", wraps the filesystem, and "calls no external binary". In fact `repo_test.go:17` runs `git ls-files` over the whole repo, excludes only `workshop/` and `testdata/`, and doesn't skip generated files. **Fix:** add a `## Revisions` entry that corrects the table, the signature, and the repo test's description, including its `git` dependency. Or skip generated files (`// Code generated … DO NOT EDIT.`) if that was intended.

**3. Important findings**
- **The wall-clock-assertion seed rule was dropped without declaring it (ARCH-PURPOSE).** The Spec's seed list and the issue title include "wall-clock asserts". The plan's three deviations cover the engine, pair, and wire decode, but not this rule. `test-shrunk-deadline` is a different class. The test tree has 14 `if … time.Since(…)` assertions (for example `boundarypush_test.go`, `processgroup/process_unix_test.go`). **Fix:** either add the rule, or declare it as a fourth deviation with a filed follow-up issue and cite it in the plan and atlas.
- **Coverage regression: the default git deadline is no longer tested.** `gitx/lifetime_test.go`'s `default-deadline` case used to run `run(...)` with `gitCommandTimeout=500ms`. It now calls `boundedGitOutput(deadline, …)` with a cued context. No test in gitx now fails if the `context.WithTimeout(ctx, gitCommandTimeout)` in `lifetime.go:25` is removed, or if a nil context loses its limit. **Fix:** add a check that doesn't depend on timing. For example, call `boundedGitCommand(nil, …)` and assert the returned context has a deadline within `gitCommandTimeout` of now.

**4. Minor findings**
- `shrunkDeadlines` flags any plain identifier assigned with `=`, including local variables, while the plan says "package-level". It also misses field and selector targets such as `cfg.Timeout = 10*time.Millisecond`.
- `goBuilds` matches only literal `"go","build"`. A variable `goBin, "build"` slips through.
- `tmpLiterals` matches packages by identifier name (`os`, `filepath`, `ioutil`), so an aliased import evades it. This is acceptable without type-checking but worth documenting.
- A suppression on a line covers findings on both that line and the next. A trailing suppression on statement N therefore also covers statement N+1.
- The Spec's "a rule that never fires in a quarter is reviewed for removal" has no mechanism and is not listed as deferred.
- I couldn't confirm follow-ups #438/#439/#440 exist in this checkout; they aren't on the pinned branch. The citations are consistent in the atlas and the Log.

**5. Test coverage notes**
- The engine and rules are tested as pure functions on in-memory source. The repo test is the only part with IO.
- The suppression table covers same line, line above, two lines above, wrong rule, no reason, dash only, unknown rule and unused. It has no case for a suppression of a test-only rule in a non-test file; that would be reported as unused, which is the right behaviour, but it isn't pinned.
- The converted deadline tests pass at head (gitx 46.9 s, judge 8.6 s).

**6. Architecture**
- **ARCH-DRY: pass.** The package follows `pkg/shelllint`'s shape, the `calls` helper is shared across rules, and the process line has one source.
- **ARCH-PURE: pass.** `Check` and the rules are pure functions over the syntax tree, and the repo test is a thin IO shell.
- **ARCH-PURPOSE: flag.** The wall-clock rule was dropped without declaring it (Important above). The golangci-lint and pair deviations were approved by the hub.
- **ARCH-MOCK: pass, with a doc correction.** The test calls `git ls-files`, which is read-only and test-only. The plan's claim that no external binary is called is wrong (folded into the Critical finding).
- **ARCH-CONSTRAINTS: pass.** The budget is under 3 s; measured 0.26–0.7 s and logged.
- **ARCH-SECURE: pass.** The package reads only the repo's own sources and handles parse failures by reporting them.
- **ARCH-ORDER: pass.** No state is kept between files; each suppression's `used` flag lives only for one `Check` call.
- **ARCH-FUNERAL: pass.** Nothing durable is created. Unused suppressions are reported, which gives them an end.

**7. Plan revision recommendations**
- Revisions: the Core concepts table should say that `Rule`, `Finding`, `suppression` and `Check` all live in `pkg/golint/check.go`, and that `Rule.Match` is `func(*ast.File) []ast.Node`.
- Revisions: the repo test lists tracked files with `git ls-files` across the module, excluding `workshop/` and `testdata/`. It depends on `git`, and generated files are not skipped.
- Revisions or Decisions: add the wall-clock-assertion seed rule as a declared deviation with its follow-up issue, or deliver it.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: core-concepts-table-drift
    title: |
      Plan Core concepts and integration claims contradict the code (rule.go/suppress.go absent, Match signature, git-based walk)
    detail: |
      Plan places Rule/Finding in pkg/golint/rule.go and suppressions in suppress.go; both live in check.go. Plan Match is func(*ast.File,*token.FileSet) []token.Pos; code is func(*ast.File) []ast.Node. Plan says repo test walks cmd/ and pkg/, skips generated files, and calls no external binary; repo_test.go:17 shells out to git ls-files over the whole repo and skips no generated files. Fix with a Revisions entry (or skip generated files).
  - id: new
    severity: Important
    kind: defect
    family: undeclared-scope-reduction
    title: |
      Spec seed rule for wall-clock duration assertions dropped without a declared deviation or follow-up
    detail: |
      The Spec seed list and issue title include wall-clock asserts; the plan's three deviations (engine, pair, wire decode) do not mention it, and 14 if-time.Since assertions exist in cmd test files. Add the rule or declare it as a fourth deviation with a filed follow-up (ARCH-PURPOSE).
  - id: new
    severity: Important
    kind: defect
    family: test-conversion-drops-coverage
    title: |
      gitx default-deadline test no longer exercises the gitCommandTimeout ceiling of the default runner
    detail: |
      The default-deadline case now calls boundedGitOutput with a cued context instead of run() with a shrunk gitCommandTimeout, so removing context.WithTimeout(ctx, gitCommandTimeout) in lifetime.go:25 fails no test. Add a race-free check, e.g. that boundedGitCommand(nil, ...) returns a context with a deadline within gitCommandTimeout.
  - id: new
    severity: Minor
    kind: defect
    family: lint-rule-matcher-precision
    title: |
      test-shrunk-deadline matches locals (plan says package-level) and misses selector/field targets
    detail: |
      shrunkDeadlines accepts any *ast.Ident LHS with =, and skips cfg.Timeout = 10*time.Millisecond style field overrides.
  - id: new
    severity: Minor
    kind: defect
    family: lint-rule-matcher-precision
    title: |
      test-go-build and tmp-literal match by literal and identifier name only (variable go binary or aliased import evades)
  - id: new
    severity: Minor
    kind: defect
    family: undeclared-scope-reduction
    title: |
      Spec rule-hygiene clause (review a rule that never fires in a quarter) has no mechanism and is not listed as deferred
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"19c0be95-9d9e-4532-9946-2c2512e4431e","repository":"github.com/xianxu/ariadne","issue_id":"000407","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/44cf1ce3178f19f8","reviewed_head":"21453b203f07329e7153db3fa97a171fde0f6be2","slot":"ariadne:2","requested_at":"2026-10-18T07:57:06.375852Z"},"phase":"finished","pid":93630,"started_at":"2026-10-18T07:57:06.475688Z","finished_at":"2026-10-18T07:59:40.540887Z","session_id":"dcb316aa-c569-4dab-9ae1-28e6b6a381e7","outcome":"returned"}] -->
