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
