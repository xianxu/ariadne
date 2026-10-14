# Boundary Review — ariadne#332 (whole-issue close)

| field | value |
|-------|-------|
| issue | 332 — sdlc merge integrates a moved main itself, then re-runs the suite before landing |
| repo | ariadne |
| issue file | workshop/issues/000332-sdlc-merge-integrates-a-moved-main-itself-then-re-runs-the-suite-before-landing.md |
| boundary | whole-issue close |
| milestone | — |
| window | 96d587ce291afac3eb6b69ebf117d2d83c090671..f129a2933854d6b90cdf9885eab8c77221f85caf |
| command | sdlc close --issue 332 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T03:00:24-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation satisfies the issue’s Spec and revised Plan. Clean integrations rerun the suite, conflicts preserve the branch, concurrent pushes retain lease protection, and suite failures receive baseline comparison. No new findings. All selected tests passed; the full suite and authenticated live conformance check were not rerun during this review.

1. **Strengths**

   - Integration decisions use typed events in `cmd/sdlc/integrationstate.go:41`, with tests covering retries, recovery, and invalid transitions.
   - `cmd/sdlc/landingintegrate.go:155` checks publish replay before pushing and preserves uncertain push outcomes for reconciliation.
   - CI deadlines cover inflight reads and the real Makefile probe; subprocess tests verify descendant termination.
   - `cmd/sdlc/suitebaseline.go:14` keeps diagnostic baseline results unrecorded and reports unparseable failures explicitly.
   - README and atlas updates explain the suite command, weight, integration flow, and recovery limitations.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Integration, classification, ancestor-evidence, state-transition, and lock-reentry tests passed: **86.880s**.
   - Deadline, cancellation, subprocess cleanup, legacy refusal, enforcement, and clean-checkout tests passed: **19.865s**.
   - Selected shared Git adapter tests passed: **7.861s**.
   - Removing BR-15’s owner-preservation fix in a separate disposable checkout made both regression tests fail with the expected lock self-wait. That checkout was removed.
   - `git diff --check` passed; the reviewed checkout remains unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** suite derivation and changed-path arguments have shared helpers.
   - **ARCH-PURE — pass:** transition decisions and failure classification remain free of IO.
   - **ARCH-PURPOSE — pass:** integration, landing enforcement, conflict recovery, baseline labeling, and bounded retries are delivered.
   - **ARCH-MOCK — pass:** stateful GitHub fakes and isolated bare Git remotes exercise production seams; live status conformance has an explicit invocation and cadence.
   - **ARCH-CONSTRAINTS — pass:** integration attempts, polling, command execution, and cleanup have explicit bounds.
   - **ARCH-SECURE — pass:** resumed integrations require structural evidence; lease pushes protect concurrent work; failed reads refuse admission.
   - **ARCH-ORDER — pass:** typed transitions govern integration progression, including uncertain publication and recovery.
   - **ARCH-FUNERAL — pass:** temporary worktrees and probe files have cleanup paths; logs reuse existing retention.

   The documented server-side race remains: GitHub pins the PR head, but does not atomically pin main.

7. **Plan revision recommendations:** None. Existing revisions account for the delivered names, contextual APIs, transition model, and clean-checkout correction.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      ChangedPathArgs disables rename collapsing; direct rename and ancestor-evidence regressions passed.
  - id: BR-2
    disposition: addressed
    note: |
      Direct status-parser and ancestor-walk tests cover ordering, malformed input, bounds, roots, and code changes; selected tests passed.
  - id: BR-3
    disposition: addressed
    note: |
      Parent-read and ancestor-walk errors propagate; explicit root handling and injected-error regressions passed.
  - id: BR-4
    disposition: addressed
    note: |
      parseMakeTestFacts excludes prerequisite-only Not a target entries; its direct regression passed.
  - id: BR-5
    disposition: addressed
    note: |
      The Makefile probe has a contextual deadline and process-group cleanup; real deadline and descendant regressions passed.
  - id: BR-6
    disposition: addressed
    note: |
      Both the CI gate and local-check command use withDefaultSuiteContext for effective suite derivation.
  - id: BR-7
    disposition: addressed
    note: |
      Plan revisions R2, R5, R6, and R8 document delivered names, signatures, default weight semantics, and transition ownership consistently with source.
  - id: BR-8
    disposition: addressed
    note: |
      effectiveLocalChecks forces a declared suite to remain required; omitted and false required-value gate regressions passed.
  - id: BR-9
    disposition: addressed
    note: |
      README.md documents the recorded close-suite command, default make test, mandatory declaration behavior, weight, and ancestor reuse, matching localchecks.go and localchecks_make.go.
  - id: BR-10
    disposition: addressed
    note: |
      Cleanup tests require descendant exit within one second against a sixty-second natural lifetime; probe and runner cases passed.
  - id: BR-11
    disposition: addressed
    note: |
      The opt-in live conformance test checks independently pinned status history through realGH.CommitStatuses; atlas documents its invocation and cadence.
  - id: BR-12
    disposition: addressed
    note: |
      Remaining-budget contexts cover GitHub reads, Git reads, and the real Makefile probe; late-result and cancellation regressions passed.
  - id: BR-13
    disposition: addressed
    note: |
      Landing and shared Git execution enforce command bounds and descendant cleanup; publish, ownership, and uncertain-push regressions passed.
  - id: BR-14
    disposition: addressed
    note: |
      Production integration progression uses transitionIntegration; state, retry-bound, recovery, and production sequence tests passed.
  - id: BR-15
    disposition: addressed
    note: |
      Baseline and repeated-check regressions passed on head and both failed after removing owner preservation in a disposable checkout.
```
