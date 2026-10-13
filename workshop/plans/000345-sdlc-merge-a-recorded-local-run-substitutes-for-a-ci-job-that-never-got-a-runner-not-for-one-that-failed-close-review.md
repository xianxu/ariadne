# Boundary Review — ariadne#345 (whole-issue close)

| field | value |
|-------|-------|
| issue | 345 — sdlc merge: a recorded local run substitutes for a CI job that never got a runner (not for one that failed) |
| repo | ariadne |
| issue file | workshop/issues/000345-sdlc-merge-a-recorded-local-run-substitutes-for-a-ci-job-that-never-got-a-runner-not-for-one-that-failed.md |
| boundary | whole-issue close |
| milestone | — |
| window | 33fb0feac8abc2274f3a4fe38198614d113170b5..1b1bc2df4abcbc1fbac07bdeca5641839f3b74b8 |
| command | sdlc close --issue 345 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-13T14:14:21-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Core code read; now checking the suite-lock helper's reentrancy semantics and running the tests.

Tests pass; checking README coverage, then writing the verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

This is a solid, well-separated delivery with no blocking problems. I read the full diff, ran `go vet ./cmd/sdlc`, and ran the targeted suites (`TestParseLocalChecks|TestJob|TestCIFakeGH|TestRunCIGate|TestDecideCIGate|TestRunMerge_|TestLanding|TestRunLocalCheck|TestLocalCheck|TestSelectLocal|TestLive`). All passed (`ok … 207s`).

The Done-when pair is covered end-to-end through both merge paths: a job that never got a runner merges after a local run, and a job that ran and failed still blocks. The ARCH-ORDER hazards the plan named all have tests: the head moving mid-run, a pass that was never recorded, and a declaration that exists only on the PR head.

One thing should be fixed before the gate: the new user-typed verb `sdlc local-check run` is not mentioned in README.md. The remaining findings are minor: a weakness in how the gate trusts a recorded pass, two small departures from the plan, and a table typo in the atlas.

1. **Strengths**
   - **Pure decision core.** `decideCIGate(ciInputs)` (`cmd/sdlc/cigate.go:83`) takes the runner check, the declaration and the recorded evidence as plain data. The table tests at `cigate_test.go:89-95` pin every rule in the plan, including "pending CI outranks a needed local run".
   - **Fails closed when GitHub's answer is incomplete.** `jobRun.runnerless` (`localchecks.go:166`) uses pointer and nil-slice fields, so a jobs-API body without `conclusion` or `steps` never counts as runnerless. The parser is strict: it rejects unknown fields, trailing data and duplicate labels.
   - **Re-reads after the inline run.** The `read()` closure in `runCIGate` checks the head before and after, then re-decides after the local run. Because of this, a status that never reached GitHub, or a push during the run, refuses instead of merging. Tests for these: `TestRunCIGateLocalChecks` "head moves during the run" and "a pass that was not recorded refuses".
   - **The test double tracks state.** `ciFakeGH` folds posted statuses into `PRChecks` and returns a 422 for an unknown SHA. That is the interaction a function mock would miss. The real-repo e2e tests drive the actual `runLocalCheck`, including the worktree and the fake lock helper.
   - **Reuses the existing suite lock.** `acquireSuiteSlot` finds `scripts/test-lock.py` the same way `Makefile.workflow` does, instead of re-implementing flock in Go. I checked that `--watch <pid>` makes a second `--all` acquire come back `reentrant` (`scripts/test-lock.py:121`).

2. **Critical findings:** none.

3. **Important findings**
   - **README update appears missing for `sdlc local-check run` and `.sdlc/local-checks.json`.** The README's landing paragraph (`README.md:269-276`) describes `sdlc pr` / `sdlc merge --yes` but not the new verb, or the fact that a repo can declare local substitutes. The atlas and help text cover it in depth; README needs one or two sentences plus a link to `atlas/workflow/ci-merge-check.md#recorded-local-checks-345`.

4. **Minor findings**
   - **The gate trusts any check with the right name (ARCH-SECURE).** In `decideCIGate`, `recorded[c.Name]` (`cigate.go:85-89`) accepts any check named `sdlc-local/<x>` as a recorded pass, including an Actions job or a third-party check run with that name. The local-gates guard on `.github/workflows/*` limits the damage. Still, keying on provenance (`c.Workflow == ""` and no Actions job link) would make "commit status only" hold by construction.
   - **`--rev` check is weaker than planned.** `localcheck_cmd.go` uses `git branch -r --contains sha`, which accepts any commit reachable from any remote branch. The plan said "the tip of the current branch on origin". The effect is harmless (the status just lands on an older commit), but plan and code disagree.
   - **Inline local runs happen before the `ciMissing` check.** In `runCIGate`, the `requiredChecks` / `hasCheck` check (missing `local-gates-guard/verdict`) only runs after the inline local checks. So merge can spend up to `timeout_minutes` running checks and then refuse for a reason it already knew. Moving the required-check presence test before the inline run would fix it.
   - **Atlas table typo.** `atlas/workflow/sdlc-binary.md:42`: the `merge` row now ends with a stray extra empty cell (`… cleanup | |`).
   - **Fake boilerplate (ARCH-DRY).** `e2eGH` and `landingFakeGH` each gained three delegation methods. The plan said the existing fakes would "compile unchanged". `landingFakeGH.JobRun`/`PostStatus`/`PRComment` dereference `g.ci` without a nil guard (only reachable in tests).

5. **Test coverage notes**
   - **Covered:** the parser, the runnerless classifier and job-link parsing, the gate decision across all nine planned cases, the gate shell across ten subtests, and the run shell (success, failure with comment, lock timeout, missing `Makefile.workflow`, timeout, pruning). E2E covers both merge paths, plus a local failure after a runnerless job.
   - **Live checks:** the live jobs-API conformance test exists (`TestLiveJobRunConformance`, gated on `SDLC_LIVE_GH=1`). I did not run it live. The issue log says it passed against pair jobs 115490631095 and 115497743342.
   - **Dropped PostStatus self-test:** the plan's live `PostStatus` self-test was dropped, and I checked the reason holds. `local-gates-guard/verdict` is a real commit status posted via `statuses/{sha}` (`.github/workflows/local-gates-guard.yml:79`), and `hasCheck` already relies on `gh pr checks` listing it by context.
   - **Gap:** no test has an Actions job named `sdlc-local/x` standing in as a pass (see the provenance finding above).

6. **Architectural notes**
   - **ARCH-DRY — pass.** One context builder (`localStatusContext`), and `abbrevSHA` / `stubGH` are reused. The fake boilerplate is minor.
   - **ARCH-PURE — pass.** The declaration, classifier, decision and record rendering are pure and tested without IO. The run shell is injected into `runCIGate` as `localCheckRunner`.
   - **ARCH-PURPOSE — pass.** The pair declaration and its docs (Spec item 4, part of Done-when) were moved by the hub's decision to pair#534. That is a separate repo issue, not a deferred piece of this repo's purpose.
   - **ARCH-MOCK — pass.** `ciFakeGH` tracks state behind the `ghCaller` seam, and there is a live conformance test for the jobs shape.
   - **ARCH-CONSTRAINTS — pass.** The timeout is enforced through `runCtx`, the machine-wide lock is shared, and GitHub calls are bounded by the declared labels.
   - **ARCH-SECURE — pass, with the Minor provenance note.** The declaration is read from the base tree only and parsed strictly; commands run as argv with no shell; comment bodies go over stdin.
   - **ARCH-ORDER — pass.** The re-read after the run and the head pin are tested. A process killed mid-run posts no status.
   - **ARCH-FUNERAL — pass.** Logs keep the newest 20 per-commit directories, stale temp worktrees are swept after 24h, and the `defer` removes the worktree.

7. **Plan revision recommendations**
   - Add a "## Revisions" entry: Task 9's live `PostStatus` self-test was dropped as redundant with `local-gates-guard/verdict` (it is in the issue log, but the plan still asks for it).
   - Add a "## Revisions" entry: `--rev` is accepted when any remote branch contains it, not only when it is the tip of the current branch on origin. Either tighten the code or record the looser rule.

```findings
findings:
  - id: new
    severity: Important
    family: docs-gate-readme-surface
    title: |
      README update appears missing for the new `sdlc local-check run` verb and .sdlc/local-checks.json
    detail: |
      README.md:269-276 covers sdlc pr / sdlc merge landing but says nothing about the new user-typed verb or that a repo can declare local substitutes; atlas and help are updated. Add a sentence plus a link to atlas/workflow/ci-merge-check.md.
  - id: new
    severity: Minor
    family: evidence-provenance-by-name
    title: |
      decideCIGate accepts any check named sdlc-local/<x> as a recorded pass, not only a commit status
    detail: |
      cigate.go:85-89 keys recorded[] on c.Name, so an Actions job or third-party check run with that name satisfies a substitute or required check; the local-gates guard on .github/workflows limits this. Require Workflow=="" and a non-Actions link (ARCH-SECURE).
  - id: new
    severity: Minor
    family: plan-code-drift
    title: |
      local-check --rev accepts any commit contained in any remote branch, not the tip of the current branch on origin as planned
  - id: new
    severity: Minor
    family: gate-check-ordering
    title: |
      runCIGate runs inline local checks before the cheap required-check presence test (ciMissing)
    detail: |
      A missing local-gates-guard/verdict refuses only after up to timeout_minutes of local runs; check required presence before running inline.
  - id: new
    severity: Minor
    family: doc-table-format
    title: |
      atlas/workflow/sdlc-binary.md merge row gained a stray trailing empty cell
  - id: new
    severity: Minor
    family: test-fake-duplication
    title: |
      e2eGH and landingFakeGH each add delegation boilerplate for the new seam methods; landingFakeGH derefs a nil ci
```
