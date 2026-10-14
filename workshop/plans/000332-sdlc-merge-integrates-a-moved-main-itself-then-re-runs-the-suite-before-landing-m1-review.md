# Boundary Review — ariadne#332 (milestone M1)

| field | value |
|-------|-------|
| issue | 332 — sdlc merge integrates a moved main itself, then re-runs the suite before landing |
| repo | ariadne |
| issue file | workshop/issues/000332-sdlc-merge-integrates-a-moved-main-itself-then-re-runs-the-suite-before-landing.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 96d587ce291afac3eb6b69ebf117d2d83c090671..4b9a769a82608186cce56d06c54c300e6df06e27 |
| command | sdlc milestone-close --issue 332 --milestone M1 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-13T22:34:25-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking how dependent repos obtain `Makefile.workflow`, since the probe runs in a clean worktree.

Confirmed one rename-detection gap. Now running the new tests.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 delivers what the plan promises for this milestone, and the new tests pass. I ran `go test -run 'TestParseLocalChecks|TestEffectiveLocalChecks|TestParseMakeTestFacts|TestDefaultSuiteFactsComeFromMain|TestProbeMakeTestFactsAriadne|TestAcquireSuiteSlot|TestDecideCIGate|TestRunCIGate|TestLocalCheck|TestLanding|TestRunMerge'` and it came back `ok` in 181s. What M1 delivers:
- The suite is a required, recorded local check, and repos without a declaration get `make test` by default.
- The default's existence and weight come from main, not from the branch (R1 PQ-1).
- A declared weight now reaches `test-lock.py --slots`, which fixes the bug where `make test` held one slot instead of its weight.
- A suite record on an earlier commit still counts when only docs changed since then.

One correctness gap is worth fixing before the gate. The docs-only check runs `git diff --name-only` with rename detection on, so a code file renamed to `.md` shows only its new name. That makes the change look docs-only, and an old suite record still counts even though code was removed. The parser for GitHub's status list also has no test.

1. **Strengths**
   - `cmd/sdlc/localchecks_make.go:547`: the probe runs make with a goal other than `test`, so `Makefile.workflow` never takes the machine lock while parsing. `TestDefaultSuiteFactsComeFromMain` proves this with a marker file, and also checks that no worktree is left behind.
   - `cmd/sdlc/localcheck_run.go:406`: the `detachedWorktree` helper is pulled out of `runLocalCheck` and reused by the probe, following the R1 ARCH-DRY note.
   - `cmd/sdlc/cigate.go:151`: the docs-only rule reuses `publishGateHasCodeSurface`, so this gate and the publish gate draw the docs/code line in the same place. The walk stops at the first commit that differs in code, which errs on the strict side.
   - A failing `sdlc-local/suite` on the head still makes `decideCIGate` refuse, so an older pass on an earlier commit cannot cover a newer failure.
   - `TestRunCIGate_SuiteRecordedBeforeClose` runs the real `runCIGate` against a real git repo and the stateful `ciFakeGH`. It checks both outcomes: no rerun after close's docs-only commits, and a rerun after a code commit.

2. **Critical findings:** none.

3. **Important findings**
   - **The docs-only check misses renames** (`cmd/sdlc/cigate.go:473`). `git diff --name-only a head` has rename detection on by default, so renaming `pkg/f.go` to `pkg/f.md` lists only `pkg/f.md`. I reproduced this in a scratch repo. That counts as docs-only, so the suite record from before the rename still satisfies the gate, though code was deleted. The repo already follows the right rule in `claim.go:673-680` and `issuemigrate.go:257`. **Fix:** use `--no-renames -z` and split on NUL rather than `strings.Fields`, which also fixes paths with spaces or quoting. Add a test where a code file is renamed to a doc and a rerun is required.
   - **`parseCommitStatuses` and `localRecordsOnAncestors` have no direct tests** (`cmd/sdlc/cigate.go:81`, `:133`). The parser assumes GitHub lists statuses newest first and keeps the first entry for each context. If that ordering is wrong, an older `success` could hide a newer `failure` on an earlier commit, and no test would catch it. The walk's limits are also untested:
     - stopping at the merge base;
     - a root commit with no parent;
     - stopping once every name is found;
     - the 30-commit cap.

     **Fix:** add a table test for the parser (two entries for one context, newest first, and an empty context). Add a fixture test for the walk.

4. **Minor findings**
   - `cmd/sdlc/cigate.go:143`: when `rev-list` fails, the error is treated as "root commit", which also covers other failures, such as a head object that isn't present locally. That outcome is strict, so nothing is let through, but the comment overstates what the code knows.
   - `cmd/sdlc/localchecks_make.go:606`: `^test:` also matches the `test:` line that `make -p` prints for a file only named as a prerequisite, with no rule (shown under `# Not a target:`). A repo with `all: test` and no test rule would then owe a suite it can't run.
   - `cmd/sdlc/localchecks_make.go:585`: the `make -p` probe has no timeout. A `$(shell …)` in main's Makefile that hangs would stall `sdlc merge`.
   - The Makefile probe followed by `effectiveLocalChecks` is repeated in `cigate.go:177` and `localcheck_cmd.go:349`. Both places that read the declaration already apply the default, but one helper, `readEffectiveLocalChecks(git, rev)`, would keep them from drifting (ARCH-DRY).
   - `helptext/local-check.md:49` says "every landing requires its recorded pass". A declared `suite` with `required: false` is not enforced, because the declaration overrides the default. The help text should say so.
   - The plan says an explicit `weight: 0` is refused, but the code treats 0 as the default of 1, as it does for `timeout_minutes`. The plan should change to match the code.

5. **Test coverage notes**
   - All the M1 tests the plan lists are present (in `suitecheck_test.go`), and they exercise real logic rather than mocks.
   - The live check on ariadne's Makefile (`TestProbeMakeTestFactsAriadne`) runs here and expects has=true, weight=2.
   - Two gaps: the tests above, and a rename case for the docs-only rule.
   - No prior fixes needed checking: this is the first round.

6. **Architectural notes**
   - **ARCH-DRY:** passes. The worktree helper and the docs/code rule are reused. The one repeat (probe plus default) is listed under Minor.
   - **ARCH-PURE:** passes. `effectiveLocalChecks`, `defaultSuiteCheck`, `parseMakeTestFacts` and `decideCIGate` are pure. The IO (`makeTestFacts`, `localRecordsOnAncestors`) sits at the edge, and the probe sits behind `makeTestFactsFn`.
   - **ARCH-PURPOSE:** passes for M1. Both places that read the declaration apply the default, and the scope matches the M1 plan. Integration and failure labels are M2 and M3 by design.
   - **ARCH-MOCK:** passes, with a note. `ciFakeGH.CommitStatuses` reads the same status map that `PostStatus` writes, so it keeps state. Nothing checks the newest-first assumption against the real GitHub API.
   - **ARCH-CONSTRAINTS:** passes, with a note.
     - The commit walk has limits: at most 30 diffs and 30 status calls.
     - The probe checks out a full worktree of main on each `runCIGate` call. That costs seconds, which is acceptable now, but M2's polling (`waitCIGate`) must keep the probe outside the poll loop.
     - The `make -p` probe has no timeout (Minor above).
   - **ARCH-SECURE:** flag, the rename finding. The docs-only claim is weaker than "a doc-only change cannot hide code". The settings that decide whether and how heavily the suite is required come correctly from main.
   - **ARCH-ORDER:** not applicable to M1, because no state carries over between events. The gate reads GitHub again after the inline run. The integration state machine is M2.
   - **ARCH-FUNERAL:** passes. The probe's worktree and temp file are removed on every path, and stray `sdlc-local` worktrees are swept by `sweepStaleLocalWorktrees`. Statuses are capped by GitHub.

7. **Plan revisions to add**
   - The Integration points table names `suiteAncestorRecords`, but the code calls it `localRecordsOnAncestors(ctx, gh, repo, git, head, base, names)`. The table also lists `makeTestFacts(root)`, but the code is `makeTestFacts(git, rev)`, a worktree probe following R1, with the pure parser in `parseMakeTestFacts`. Update both rows.
   - The M1 test list says `weight: 0` is refused, but 0 means the default of 1. Restate it as "a negative weight or one above 8 is refused".
   - Record that the docs-only distance uses `--no-renames`, once that is fixed.

```findings
findings:
  - id: new
    severity: Important
    family: docs-only-classification-must-see-every-changed-path
    title: |
      The docs-only check uses git diff --name-only with renames on, so code renamed to .md looks docs-only
    detail: |
      cigate.go:473 lists only the new name of a renamed file (reproduced: f.go to f.md lists just f.md), so an old suite record still counts after code is removed. Use --no-renames -z, as claim.go:673 and issuemigrate.go:257 do, and add a rename test.
  - id: new
    severity: Important
    family: new-parser-needs-direct-tests
    title: |
      parseCommitStatuses and localRecordsOnAncestors have no direct tests
    detail: |
      The newest-first dedupe in parseCommitStatuses is untested, and so are the walk's limits: merge base, root commit, stopping once every name is found, and the 30-commit cap.
  - id: new
    severity: Minor
    family: error-swallowed-as-benign-case
    title: |
      localRecordsOnAncestors treats any rev-list error as "root commit"
  - id: new
    severity: Minor
    family: make-database-parse-precision
    title: |
      ^test: also matches a prerequisite-only "Not a target" entry in make -p output
  - id: new
    severity: Minor
    family: unbounded-subprocess
    title: |
      The make -p probe has no timeout; a hanging $(shell) in main's Makefile stalls merge
  - id: new
    severity: Minor
    family: shared-helper-for-repeated-derivation
    title: |
      Makefile probe plus effectiveLocalChecks is repeated in cigate.go and localcheck_cmd.go
  - id: new
    severity: Minor
    family: plan-table-matches-code
    title: |
      Plan names suiteAncestorRecords and makeTestFacts(root); code has localRecordsOnAncestors and makeTestFacts(git, rev); weight 0 means default, not refused
```

---

## Re-review — 2026-10-13T23:02:30-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 332 — sdlc merge integrates a moved main itself, then re-runs the suite before landing |
| repo | ariadne |
| issue file | workshop/issues/000332-sdlc-merge-integrates-a-moved-main-itself-then-re-runs-the-suite-before-landing.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 96d587ce291afac3eb6b69ebf117d2d83c090671..5a4e3cdf09b31f8591625586130295767131e7c9 |
| command | sdlc milestone-close --issue 332 --milestone M1 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-13T23:02:30-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

M1’s default-suite derivation, weighted execution, and rename handling are substantially implemented and tested. One enforcement hole blocks shipping: a declared `suite` can remain optional, allowing landing without its recorded pass. README coverage is missing, and two prior Minor findings remain unresolved.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      ChangedPathArgs supplies --no-renames -z to the identified classifiers. Removing --no-renames made both TestLocalRecordsOnAncestors and TestDiffNamesListsBothSidesOfARename fail.
  - id: BR-2
    disposition: addressed
    note: |
      TestParseCommitStatuses directly tests newest-first deduplication and malformed input. TestLocalRecordsOnAncestors exercises merge-base exclusion, early termination, the commit cap, root commits, and code changes; these tests pass.
  - id: BR-3
    disposition: not-addressed
    note: |
      cigate.go:481 now treats every rev-parse failure as a root commit, moving the uncertainty rather than resolving it. No regression test injects either rev-parse or rev-list failure. Positively establish parentlessness and propagate operational errors.
  - id: BR-4
    disposition: addressed
    note: |
      parseMakeTestFacts excludes prerequisite-only entries. TestParseMakeTestFactsSkipsNotATarget passes and fails when that exclusion is removed.
  - id: BR-5
    disposition: not-addressed
    note: |
      localchecks_make.go:79-84 adds CommandContext but no bounded pipe wait or descendant cleanup. A temporary reproduction with a 50 ms deadline and $(shell sleep 1) returned after 1.04 seconds: Output waits for the descendant holding stderr open. No committed timeout regression exists. Bound subprocess lifetime and pipe draining, and test both.
  - id: BR-6
    disposition: addressed
    note: |
      withDefaultSuite is now the shared derivation called by runCIGate and runLocalCheckCmd. Both production paths have passing integration coverage; this consolidation preserves behavior.
  - id: BR-7
    disposition: addressed
    note: |
      Plan R2's “Names as delivered” passage at line 234 explicitly corrects both function names/signatures and zero-weight semantics. These match cigate.go:472, localchecks_make.go:51, and localchecks.go:115.
findings:
  - id: new
    severity: Critical
    family: mandatory-checks-cannot-be-disabled-by-overrides
    title: |
      A declared suite can bypass the mandatory landing check
    detail: |
      localchecks_make.go:38 returns a declared suite unchanged; localchecks.go:143 likewise preserves Required=false. A main-side declaration containing {"name":"suite","command":["true"]} therefore makes runCIGate accept green CI without any suite record. A temporary integration regression reproduced this acceptance. ARCH-PURPOSE: preserve command/weight overrides while requiring the suite, or reject optional suite declarations; cover omitted and explicit-false required fields.
  - id: new
    severity: Important
    family: readme-tracks-user-facing-contracts
    title: |
      README update appears missing for the default suite and weight setting
    detail: |
      README.md:318-324 still describes only explicitly declared required checks. This range adds automatic make-test enforcement, the close-suite invocation, ancestor-record reuse, and the weight configuration key without updating README. Update its local-check guidance alongside the existing atlas and embedded-help changes.
```

1. **Strengths**

   - Suite facts come from main’s detached worktree; the fixture verifies that deleting the branch’s test target cannot change those facts.
   - `ChangedPathArgs` consolidates rename-safe classification across the identified consumers.
   - Regression tests demonstrate that docs-only close commits preserve suite evidence while subsequent code changes require another run.
   - The weight test verifies the actual `--slots 2` argument reaching the lock helper.

2. **Critical findings**

   The declared-suite bypass above violates M1’s explicit “required on every landing” contract. Fix both declaration paths through the shared derivation and add a gate-level regression.

3. **Important findings**

   README lacks the new usage/configuration contract. Atlas and embedded help were updated appropriately.

4. **Minor findings**

   - **BR-3:** failed parent probes still become benign absence; error-path regression coverage is missing.
   - **BR-5:** the timeout does not bound descendant-held pipes. Add bounded cancellation and cleanup coverage.

5. **Test coverage notes**

   Targeted tests passed in `cmd/sdlc`, `internal/gitx`, and `internal/freshness`. A subsequent focused run of the M1 tests also passed. BR-1 and BR-4 mutation checks produced the expected failures. Temporary reproductions confirmed the optional-suite bypass and ineffective subprocess deadline.

   All temporary changes were restored; the checkout is clean. The full repository suite was not run.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared suite derivation, worktree helper, and changed-path arguments reduce duplication.
   - **ARCH-PURE — pass:** declaration derivation and gate decisions remain pure; IO resides in integration helpers.
   - **ARCH-PURPOSE — flag:** optional declared suites defeat mandatory enforcement.
   - **ARCH-MOCK — pass:** GitHub uses the existing stateful seam; repository fixtures exercise git interactions and make has live conformance coverage.
   - **ARCH-CONSTRAINTS — flag:** BR-5’s claimed timeout is not an effective wall-time bound.
   - **ARCH-SECURE — flag:** BR-3 still converts a failed external probe into asserted absence.
   - **ARCH-ORDER — flag:** BR-5 leaves descendant lifetime outside effective cancellation control. The larger integration state machine belongs to M2.
   - **ARCH-FUNERAL — pass:** worktrees and probe files have cleanup paths; suite logs retain the existing bounded lifecycle.

   M1’s core-concept entries match the implementation after R2’s corrections. M2/M3 entries are future work, not missing M1 deliverables.

7. **Plan revision recommendations**

   Append an R3 entry specifying that suite overrides cannot disable mandatory enforcement, defining positive root-commit evidence and error propagation, and correcting the probe’s timeout contract to include descendant cleanup and bounded pipe draining. Add the README update to M1’s documentation checklist.
