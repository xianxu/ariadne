# Boundary Review — ariadne#348 (whole-issue close)

| field | value |
|-------|-------|
| issue | 348 — issues: a hierarchical area path (couch/broadcast/web) and typed facets (severity), queryable by prefix |
| repo | ariadne |
| issue file | workshop/issues/000348-issues-a-hierarchical-area-path-couch-broadcast-web-and-typed-facets-severity-queryable-by-prefix.md |
| boundary | whole-issue close |
| milestone | — |
| window | f6774623fcf1d4e44a25780fbf5a4e7061241229..83f46fe584de41c2384ac3b4e71a414b11b99bfb |
| command | sdlc close --issue 348 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-13T18:20:00-07:00 |
| verdict | REWORK |

## Review

Now the full diff, in parts.

Reviewing the #348 diff now. Next I'm checking two suspected edge cases with scratch tests: `issue validate` in a repo that has no `main`, and numeric-looking area values.

Two edge cases confirmed so far. Running the new `cmd/sdlc` tests next, then working through the architecture checklist.

```verdict
verdict: REWORK
confidence: high
```

Most of #348 is solid. The model, the pure `repoenum` core, the shape-only card kinds, the setters, the gate wiring, the list filter, help and atlas all match the plan and its R1 revision. The core packages pass their tests. One change breaks an existing command, and it blocks SHIP. `sdlc issue validate` now fails **every** file in a repo that has no `main` ref (no `origin/main` and no local `main`), including files that carry no area or severity at all. It reports "classification: could not read the declaration: no main to read .sdlc/vocabulary.json from".

I reproduced it with the head binary in a `master` repo: exit 1. The base binary passes the same file. The existing `TestRunIssueValidate` also **fails on a clean checkout of the pinned head**, because the checkout has no main ref. The test reads the real repo's git refs through `repoRootOf(cwd)`. So the logged "all shards green" depends on the environment.

## 1. Strengths
- `cmd/sdlc/internal/repoenum/repoenum.go`: a small, pure, strict parser. It rejects unknown fields, trailing data, a wrong version, undeclared keys, bad syntax and duplicates, and every error names the declaration file. The path regex is built from the model's `segment`/`maxDepth`, so Go restates no syntax (ARCH-DRY at the model level).
- `cmd/sdlc/internal/issue/card.go:279`: card parsing checks shape only, never membership. A value later removed from the declaration cannot freeze the tracker. This is the right call and is well documented.
- The R1 sweep of helpers that assumed scalar values was actually done: `sameCardValue` handles sequences, `dropBlockContinuation` drops block-list lines, `IssueRecord.List`/`Text` exist, and migration converts block lists to flow form.
- The fuzz side-quest (`parseCardDocument`: refuse frontmatter containing `\r`/NEL/U+2028/U+2029) fixes a real panic, and the seed is kept as a regression test.
- The declaration is read from main or the gate's base, never from HEAD (`repoEnumsMainRev`, `TestLoadRepoEnums_NeverReadsHead`). The governance model holds.

## 2. Critical findings
- **`cmd/sdlc/issue.go:122` / `:171-178`: `issue validate` fails closed on unclassified files when there is no main ref.**
  - `issueValidateDeclaration` runs on every invocation. `classificationProblems` turns any `declErr` into a problem for every file, even when the file has no `area`/`severity`.
  - Repro: `git init -b master` plus a conforming issue file. Base exits 0 ("conforms"); head exits 1.
  - `TestRunIssueValidate/clean_file_passes` and `/multiple_files_all_conforming_passes` fail in this checkout. They pass only where the developer's real repo happens to have `main`, so the tests are not hermetic (ARCH-MOCK: they reach real repo state).
  - Fix:
    - Read the declaration lazily, only when a file or card actually carries an `area`. Severity needs no declaration.
    - Report `declErr` only for those files.
    - Make the test hermetic: chdir into a temp repo, or inject the declaration reader the way `repoEnumsAtBaseFn` is injected.
    - Add a regression test: a repo without main, an unclassified file → conforms.

## 3. Important findings
- **Some valid declared areas can't be written** (`repoenums.go:660` `flowList`, `scaffold.go:242` `Render`).
  - The segment rule `[a-z0-9][a-z0-9-]*` lets a declaration include `2026`, `1e5`, `true`, `null`.
  - `flowList`/`Render` emit them unquoted, so YAML tags them `!!int`/`!!float`/`!!bool`/`!!null`. `validateRepoEnumNode` requires `!!str` (`area: [1]` is deliberately refused in `cardenum_test.go`).
  - So `set-area 2026` passes membership and is then refused with "not a valid area", and `issue new --area 2026` fails at card split.
  - Meanwhile `FieldList` (gates, legacy details) decodes the same values as strings, so the writers and the gates disagree.
  - Fix: quote any value that doesn't resolve to `!!str` in `flowList`/`Render`, or tighten the CUE segment rule so such values can't be declared. Add a test with a numeric-looking segment.
- **README not updated for new surface:** `issue set-area`, `issue set-severity`, `issue new --area/--severity`, `issue list --area/--severity`, and the per-repo `.sdlc/vocabulary.json` that operators now author. The atlas and help were updated; README.md was not. One sentence plus an atlas link in "Concurrent issue work" would do.

## 4. Minor findings
- `set-area` uses `slices.Compact`, which only removes *adjacent* duplicates. `set-area a b a` passes membership, then fails at `SetCardField` with "listed twice". `issue new --area a --area a` fails only at render/split. `checkClassification` never checks for duplicates; it should, so all writers refuse the same way.
- ARCH-DRY: severity membership is checked in three places (`card.go` enum case, `checkClassification`, `newIssueListFilter`), and area syntax in three. `FieldList`/`FieldString` each unmarshal the whole frontmatter. These could be one helper each.
- `repoenum.ValidValue` compiles the regex on every call (per area, per issue, per gate). Cache it per spec.
- `state.go:88-96`: the new `withClassification` was inserted between `withOwnership`'s doc comment and its function, so `withOwnership` lost its doc.
- `runIssueValidate --all` adds card problems to `nonconforming`, but the error still says "N of M issue file(s)", so N can exceed M.
- In legacy details, `FieldList` on `area: [null]` yields an empty list, which passes the gate silently.

## 5. Test coverage notes
- `repoenum` tests (table, fuzz, property), the card shape tests and `TestIssueListFilter` are good and run without IO.
- No test covers `issue validate` in a repo without main, or a file with no classification when the declaration can't be read. That is exactly the shipped bug.
- `TestRunIssueValidate` silently depends on the real checkout's refs. It should be isolated.
- No test covers a declared value that YAML would not read as a string.

## 6. Architecture
- **ARCH-DRY: flag (Minor).** See the triplicated validation above. The model is the single source, and Go derives from it.
- **ARCH-PURE: pass.** `repoenum` and `issueListFilter` are pure. IO sits in `repoenums.go` and behind `repoEnumsAtBaseFn`.
- **ARCH-PURPOSE: pass.** All consumers derive from the model: setters, gates, list, state, help (`{{SEVERITY_NAMES}}`, `{{REPO_ENUMS_PATH}}`). Per-repo area lists are correctly left to operators.
- **ARCH-MOCK: flag**, as part of the Critical finding. `runIssueValidate` has no seam for the declaration reader, so its test reads the real repo's refs.
- **ARCH-CONSTRAINTS: pass.** The regex recompile is the only (minor) cost.
- **ARCH-SECURE: pass**, with the note above that the writers' YAML encoding and the gates' YAML decoding disagree. The declaration is strictly parsed from main.
- **ARCH-ORDER: pass.** Nothing is held between events. The membership check runs inside the CAS closure. The shrinking-declaration race is documented as intended governance.
- **ARCH-FUNERAL: pass.** Two optional card fields and one operator-owned file; nothing else durable.

## 7. Plan revision recommendations
- Add a Revisions entry: "R2: `issue validate` reads the declaration only for files and cards that carry an `area`; a missing main is reported only for those files (it is not fail-closed for unclassified files). Tests are isolated from the real repo's refs." This covers PQ-4's "else refuse", which was written for writers and then applied to every validated file.
- Note that declared values must survive the card's YAML flow encoding: either the writers quote them, or the segment rule excludes values that YAML reads as numbers, booleans or null.

```findings
findings:
  - id: new
    severity: Critical
    family: classification-check-gates-unclassified-input
    title: |
      issue validate fails every file when no main ref exists, even files with no area/severity
    detail: |
      issueValidateDeclaration runs unconditionally (issue.go:122) and classificationProblems turns any declErr into a problem for every file (issue.go:176). Reproduced: in a master-only repo the base binary exits 0 and head exits 1. The existing TestRunIssueValidate fails on a clean checkout of the pinned head (no main ref) because it reads the real repo's refs via repoRootOf(cwd). Fix: read the declaration lazily, only for files or cards that carry an area. Inject the reader and make the test hermetic.
  - id: new
    severity: Important
    family: writer-gate-encoding-divergence
    title: |
      Declarable areas such as 2026, 1e5, true and null cannot be written: flowList/Render emit them unquoted, so they become non-string YAML scalars
    detail: |
      The segment rule allows digit-led and keyword segments. flowList (repoenums.go:660) and Render (scaffold.go:242) emit them unquoted, and validateRepoEnumNode requires !!str, so set-area and issue new refuse a declared value. FieldList (gates) decodes the same value as a string, so the writers and the gates disagree. Fix: quote non-!!str values, or tighten the segment rule in issue.cue. Add a test.
  - id: new
    severity: Important
    family: docs-gate-readme
    title: |
      README update appears missing for issue set-area/set-severity, the --area/--severity flags and .sdlc/vocabulary.json
  - id: new
    severity: Minor
    family: writer-validation-parity
    title: |
      checkClassification does not refuse duplicate areas; slices.Compact only removes adjacent ones, so set-area a b a fails later in SetCardField
  - id: new
    severity: Minor
    family: arch-dry-shared-validator
    title: |
      Severity membership and area syntax are each checked in three places (card.go, checkClassification, newIssueListFilter); consolidate them into one helper
  - id: new
    severity: Minor
    family: doc-comment-placement
    title: |
      state.go: withClassification was inserted under withOwnership's doc comment, so withOwnership lost its doc
  - id: new
    severity: Minor
    family: hot-path-recompile
    title: |
      repoenum.ValidValue recompiles the path regex on every call; cache it per spec
```

---

## Re-review — 2026-10-13T19:38:31-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 348 — issues: a hierarchical area path (couch/broadcast/web) and typed facets (severity), queryable by prefix |
| repo | ariadne |
| issue file | workshop/issues/000348-issues-a-hierarchical-area-path-couch-broadcast-web-and-typed-facets-severity-queryable-by-prefix.md |
| boundary | whole-issue close |
| milestone | — |
| window | f6774623fcf1d4e44a25780fbf5a4e7061241229..d0cfeff3d6927a0ba7229df8969fcc34dcbda22b |
| command | sdlc close --issue 348 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-13T19:38:31-07:00 |
| verdict | SHIP |

## Review

Waiting on the test suite; meanwhile checking the remaining area writers for the BR-2 class.

The main package hit go test's 10m default timeout. Running the #348 tests directly instead.

Pre-existing environment failure (also fails on base). Now reverting the BR-1 and BR-2 fixes to confirm their tests go red.

Both fixes go red when reverted, and the checkout is restored. Next, sweeping for any leftover duplicate validators or unquoted area writers.

All prior fixes check out. Last step: a quick look at the plan's checklist and the atlas diff for any gaps.

```verdict
verdict: SHIP
confidence: high
```

All seven findings from round 1 are fixed, and I found nothing new. I checked each fix against the code and tests at head `d0cfeff3`, not against the commit message. For the two behavior fixes (BR-1 and BR-2), I reverted each fix in this checkout and confirmed its regression tests then fail. Both files are restored and the checkout is clean. This checkout has no `main` ref, which is the exact setup BR-1 described. Here `TestRunIssueValidate` now passes; with BR-1 reverted it fails. The one test failure I saw, `TestPushPublishSequence_CodecompleteFlippedThenArchived`, fails the same way on the base commit `f6774623`, so this change didn't cause it.

**1. Strengths**
- `cmd/sdlc/repoenums.go:579-606`: one memoized reader, `lazyDecl`, now handles the declaration for every caller. Its typed error, `declUnreadable`, means the merge/push gate still refuses to pass when it needs the declaration and can't read it (`validategate.go:109-112`). Issues with no area no longer depend on the declaration at all. This one design fixes BR-1 at every call site: validate, the writers, publish and the gate.
- `cmd/sdlc/internal/issue/classify.go`: there is now one validator per field: `CheckSeverity`, `CheckAreaSyntax`, `RepoEnumValid` and `EnumValid`. The card parser, `checkClassification` and the list filter all use them. The only other places that read the severity list are help text and flag descriptions, which are generated from the same model.
- `FlowList` decides whether to quote a value by asking the YAML decoder itself, so it doesn't keep its own list of special words. Both writers that emit areas (`scaffold.go:242` and `cardsetters.go:175,179`) use it, and `TestFlowList` checks that every rendered list reads back as strings.
- `TestValidateChangedIssuesClassification` now asserts how many times the declaration is read in each case (0 or 1). That pins the "read only when needed" contract, not just whether the gate passes or fails.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings:** none new.

**5. Test coverage notes**
- **BR-1 reverted** (made `checkClassification` read the declaration unconditionally). These went red:
  - `TestRunIssueValidate`
  - `TestIssueValidateWithoutMain`
  - `TestCheckClassification`
  - five subtests of `TestValidateChangedIssuesClassification`
- **BR-2 reverted** (stopped quoting in `FlowList`). These went red:
  - `TestFlowList`
  - `TestIssueSetAreaQuotesAndDedupes`
- **BR-4:** covered by the "twice" case in `TestCheckClassification`, the duplicate-area gate subtest, and the dedupe check in the set-area test.
- **Full suite:** `go test ./cmd/sdlc/...` hits Go's default 10-minute timeout on the `cmd/sdlc` package in this checkout. Every other package passes. The targeted #348 tests pass.

**6. Architecture**
- **ARCH-DRY: pass.** BR-5 is consolidated into `classify.go`, and `lazyDecl` replaces four copies of hand-written lazy loading.
- **ARCH-PURE: pass.** The validators and `FlowList` are pure. Reading the declaration is passed in as a function argument.
- **ARCH-PURPOSE: pass.** The fixes cover the whole class each time: every writer quotes its output, and every checker is lazy.
- **ARCH-MOCK: pass.** Tests swap the declaration reader through a replaceable function variable (`repoEnumsAtBaseFn`) and use real temporary git repos via `testfix`.
- **ARCH-CONSTRAINTS: pass.** The compiled path pattern is now cached (BR-7), and the declaration is read at most once per run.
- **ARCH-SECURE: pass.** If the declaration can't be read or parsed, the gate refuses to pass, and the per-file problem is reported where you can see it.
- **ARCH-ORDER: pass.** The memoized reader only lives for one command or gate run.
- **ARCH-FUNERAL: pass.** The pattern cache is keyed by the model's area specs, which are a fixed set, so it can't grow without limit. Nothing durable is created.

**7. Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      checkClassification reads the declaration via lazyDecl only when there are areas; reverting this turns TestRunIssueValidate, TestIssueValidateWithoutMain, TestCheckClassification and the gate subtests red, and TestRunIssueValidate passes in this no-main checkout.
  - id: BR-2
    disposition: addressed
    note: |
      issue.FlowList (classify.go) quotes values that would not decode as strings, and both area writers (scaffold.go:242, cardsetters.go:175/179) use it; reverting the quoting turns TestFlowList and TestIssueSetAreaQuotesAndDedupes red.
  - id: BR-3
    disposition: addressed
    note: |
      README.md gains a "Classifying issues" section covering set-area/set-severity with --clear, the --area/--severity flags on new and list, .sdlc/vocabulary.json and the gates; this matches cardsetters.go and issue.go.
  - id: BR-4
    disposition: addressed
    note: |
      checkClassification refuses duplicates ("listed twice") and writers apply uniqueAreas; both are covered by TestCheckClassification, the gate's duplicate-area subtest and TestIssueSetAreaQuotesAndDedupes.
  - id: BR-5
    disposition: addressed
    note: |
      card.go, checkClassification and newIssueListFilter all go through issue.EnumValid/RepoEnumValid/CheckSeverity/CheckAreaSyntax; the remaining reads of the severity list are help text generated from the model.
  - id: BR-6
    disposition: addressed
    note: |
      The doc comment for withOwnership is back above its function in state.go.
  - id: BR-7
    disposition: addressed
    note: |
      pathPattern caches the compiled pattern in a sync.Map keyed by the model's area spec, a fixed set.
```

---

## Re-review — 2026-10-13T20:14:20-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 348 — issues: a hierarchical area path (couch/broadcast/web) and typed facets (severity), queryable by prefix |
| repo | ariadne |
| issue file | workshop/issues/000348-issues-a-hierarchical-area-path-couch-broadcast-web-and-typed-facets-severity-queryable-by-prefix.md |
| boundary | whole-issue close |
| milestone | — |
| window | 62da7e26dabe6ffeafb04ed13f59b3c610c59a2d..534f27540c7435c86d09a8ce25d3d8c406fff09f |
| command | sdlc close --issue 348 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-13T20:14:20-07:00 |
| verdict | SHIP |

## Review

Still waiting on the build and test run. The conflict resolution I've read so far looks correct.

Two RepoLock tests are failing. I'm checking whether they also fail on main.

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** All seven earlier findings (BR-1 to BR-7) were already disposed, and this window adds no new defect. The interdiff does three things:

- **Conflict resolutions in `cmd/sdlc/issuepublish.go` and `cmd/sdlc/legacymode.go`.** These are the real content of the window. They keep both sides' intent.
- **One test call-site update.** `cmd/sdlc/repoenums_test.go:169` was changed to match main's new signature.
- **Bookkeeping.** The card moved from working to codecomplete and the 348 plan files moved from `workshop/history/plans/` back to `workshop/plans/`. `## Spec`, `## Revisions` and `## Done when` are unchanged, so scope did not change.

Main had committed conflict markers (`<<<<<<< 3d12bf88 … >>>>>>> d0cfeff3`) into base `62da7e26`. Head resolves all of them, and `git grep` finds no markers left outside `workshop/`.

## 1. Strengths
- **The publish resolution keeps both changes** (`cmd/sdlc/issuepublish.go:156-165`):
  - This issue's `checkPublishedClassification(tf, selected, roots[0])` stays in place, after `validateCommitSelection`.
  - Main's change stays too: it builds `changed` and calls `tf.PublishCommitChecked(...)` with `validateAssetPublication`, replacing the bare `tf.PublishCommit`.
  - Ordering is sound: classification is a cheap check before publishing, and the asset check runs inside the checked publish. Nothing from either side is dropped.
- **The import resolution in `legacymode.go:24-25` keeps both `issueassets` (main) and `repoenum` (this issue).** Both are used, and `go build ./...` and `go vet ./cmd/sdlc/` pass.
- **The test fix follows main's API change.** `publishIssueCommit` now takes a `context.Context` first, and `repoenums_test.go` passes `context.Background()`. That keeps the publish-refusal regression test compiling and running.

## 2. Critical findings
None.

## 3. Important findings
None.

## 4. Minor findings
None introduced by this window.

## 5. Test coverage notes
- **Tests for this issue pass at head:** `TestIssuePublish_RefusesUndeclaredArea`, `TestIssueValidateWithoutMain`, `TestCheckClassification`, `TestIssueSetAreaAndSeverity`, `TestIssueSetAreaQuotesAndDedupes`, and the `TestLegacy*` lifecycle and gate tests.
- **Every `cmd/sdlc/internal/*` package passes**, including `repoenum`, `issue`, `tracker` and `gitx`.
- **Three tests in `cmd/sdlc` fail, but the same way on main's parent `3d12bf88`:** `TestPushPublishSequence_CodecompleteFlippedThenArchived`, `TestLegacyRepoLockConcurrentIssueNewSerializesAllocation` and `TestLegacyRepoLockSetStatusMutationWaits`. I ran them in a temporary worktree of main and got identical failures. The error is "issue tracker cutover mismatch … no issue tracker is reachable", which looks like a property of the review checkout's environment (no issue-tracker remote). This window did not cause them.
- **The full `cmd/sdlc` package did not finish within the 600s default `go test` timeout here**, so I couldn't confirm the whole package passes. I covered the window's code with the targeted runs above.

## 6. Architectural notes
- **ARCH-DRY: pass.** The resolution adds no duplicated logic; it splices the two sides together with no copy-paste.
- **ARCH-PURE: pass.** `checkPublishedClassification` still sits in the thin publish shell and reads the declaration lazily through `lazyDecl`. Asset validation stays behind the injected `PublishCommitChecked` callback.
- **ARCH-PURPOSE: pass.** The publish-time membership check, which was the point of the earlier publish-refusal work, survives the merge. It was not dropped in favour of main's version.

## 7. Plan revision recommendations
None.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Already disposed in prior round; interdiff does not regress it (TestIssueValidateWithoutMain passes at head).
  - id: BR-2
    disposition: addressed
    note: |
      Already disposed; TestIssueSetAreaQuotesAndDedupes passes at head.
  - id: BR-3
    disposition: addressed
    note: |
      Already disposed; README untouched by this interdiff.
  - id: BR-4
    disposition: addressed
    note: |
      Already disposed; dedupe test passes at head.
  - id: BR-5
    disposition: addressed
    note: |
      Already disposed; interdiff adds no new validator copies.
  - id: BR-6
    disposition: addressed
    note: |
      Already disposed; state.go untouched by this interdiff.
  - id: BR-7
    disposition: addressed
    note: |
      Already disposed; repoenum untouched and its tests pass.
```
