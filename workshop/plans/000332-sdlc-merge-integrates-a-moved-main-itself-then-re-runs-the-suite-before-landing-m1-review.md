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
