# Boundary Review — ariadne#589 (whole-issue close)

| field | value |
|-------|-------|
| issue | 589 — Recorded suites report leaked SysV shared-memory segments; PostgreSQL test harnesses must not leak them |
| repo | ariadne |
| issue file | workshop/issues/000589-recorded-suites-report-leaked-sysv-shared-memory-segments-postgresql-test.md |
| boundary | whole-issue close |
| milestone | — |
| window | 945a8168578afcbd94ed27965c283d0a6483aeb1..570681b8038d97147d5a9bd3c7529d528c0625f9 |
| command | sdlc close --issue 589 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-24T13:17:39-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation separates observation from classification, preserves warnings in ordinary saved-result replay, and adds useful lifecycle tests. Focused tests pass, but diagnostics can change timeout outcomes, creator attribution does not establish process incarnation, and mutable metadata can make existing segments appear new. The concept table also contradicts the pinned diff.

Done-when: current — warning IDs, an abandoned-segment fixture, harness guidance, and a filed downstream issue still cover the revised advisory scope.

```findings
findings:
  - id: new
    severity: Critical
    family: advisory-outcome-isolation
    title: |
      SysV observation time can turn a completed command into a timeout
    detail: |
      cmd/sdlc/localcheck_run.go:676-687 runs observations under an already-running command timeout, then line 722 checks that timeout after observation completes. A successful command followed by a slow after-snapshot can therefore become TimedOut/LaneOverrun; the before-snapshot also consumes command execution time. ARCH-CONSTRAINTS and ARCH-PURPOSE: isolate diagnostic budgets from execution budgets and capture the execution outcome before post-run diagnostics. Add regressions with delayed before/after observers proving diagnostics cannot change an otherwise successful outcome.
  - id: new
    severity: Critical
    family: creator-incarnation-evidence
    title: |
      A nonempty start time does not establish that the witnessed process created the segment
    detail: |
      cmd/sdlc/internal/sysvshm/report.go:29-33 matches only the numeric creator PID and checks that Witness.Start is nonempty. No code compares that incarnation with evidence from the segment. If a concurrent creator exits and its PID is reused by a run-group process, the unrelated segment receives that process's name and a witnessed-creator label. ARCH-SECURE and ARCH-ORDER: require evidence linking the creator incarnation to the allocation, otherwise retain unknown attribution. Test PID reuse into the owned group; the existing fake lookup returning false does not exercise this case.
  - id: new
    severity: Critical
    family: allocation-identity-stability
    title: |
      Mutable change time is treated as allocation identity
    detail: |
      cmd/sdlc/internal/sysvshm/report.go:11-26 uses the entire Segment, except Attachments, as identity. Changed comes from CTIME/ctime, which is the last shmctl change time rather than an immutable allocation generation. Updating permissions through IPC_SET can therefore make a pre-existing unattached segment appear newly allocated. ARCH-PURPOSE: separate stable identity from mutable metadata, treat ambiguous changes conservatively, and add a regression where one existing segment changes metadata without being replaced.
  - id: new
    severity: Critical
    family: concept-table-code-consistency
    title: |
      The Core concepts table claims a modification absent from the pinned range
    detail: |
      workshop/plans/000589-recorded-suites-report-leaked-sysv-shared-memory-segments-postgresql-test-plan.md:40 marks savedAttemptRecord in cmd/sdlc/localcheck_republish.go as modified, but that file has no base-to-head diff. Replay already carries the supplied tail; the new behavior resides in statusDescription. Per the required Core concepts cross-check, correct the row to existing and append a Revisions entry explaining reuse of the existing validation path.
  - id: new
    severity: Important
    family: diagnostic-consumer-completeness
    title: |
      Resume publication drops warnings produced only by the child run
    detail: |
      cmd/sdlc/localcheck_resume_saved.go:117-138 builds the composite record using only the parent's tail, while cmd/sdlc/localcheck_resume.go:228 discards the child's hashAndTail tail. The newly instrumented resume batch can warn in its child log, yet fresh and saved composite publication omit the promised SysV warning prefix when the parent was clean. ARCH-PURPOSE: derive diagnostic status from both digest-verified logs and test a clean parent followed by a leaking child through publication and republish.
  - id: new
    severity: Important
    family: user-surface-documentation
    title: |
      README update appears missing for the new qualification usage
    detail: |
      atlas/workflow/sdlc-binary.md:3394 introduces TestNativeOwnedSysVQualification and WF_SYSV_QUALIFY=1, but README.md is unchanged in the pinned range. The requested README gate covers new usage steps and configuration keys. Add a concise README pointer explaining host warnings and linking the operator-cleared qualification procedure.
```

1. **Strengths**

   - Strict parsers reject partial tables rather than reporting fabricated empty observations.
   - Observation is explicitly injected into production constructors; unconfigured unit tests avoid live host probing.
   - Ordinary replay retains the digest-verified warning footer, with compatibility coverage for older logs.
   - Atlas and embedded help explain attribution uncertainty, VM exclusions, and PostgreSQL shutdown responsibilities.

2. **Critical findings**

   The four Critical findings above require fixes before admission. The concept-table correction is documentation-only; the three behavior corrections need meaningful regression tests.

3. **Important findings**

   Preserve child-run warnings in resume publication and update README usage documentation.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed on the pinned head:

   `go test ./cmd/sdlc/internal/sysvshm ./cmd/sdlc -run 'SysV|Sysv|TestVerbContextsReachTrackerReads' -count=1 -timeout 3m`

   Existing coverage validates ordinary warning persistence, cancellation, observation failure, admission refusal, and probe counts. It does not cover delayed diagnostics crossing the execution deadline, creator PID reuse into the owned group, mutable CTIME, or child-only resume warnings. Native allocation and private-executable qualification were not independently exercised in this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared classification, rendering, and status extraction.
   - **ARCH-PURE — pass:** deterministic parsing/classification separated from host observation.
   - **ARCH-PURPOSE — flag:** allocation classification and resume publication leave contract gaps.
   - **ARCH-MOCK — flag:** creator lifecycle tests do not reproduce the reuse scenario behind the attribution claim.
   - **ARCH-CONSTRAINTS — flag:** independently bounded diagnostics still affect execution deadlines.
   - **ARCH-SECURE — flag:** numeric PID matching substitutes for creator-incarnation evidence.
   - **ARCH-ORDER — flag:** creator exit/reuse between allocation and observation is not distinguished.
   - **ARCH-FUNERAL — pass:** observations remain scoped values; durable diagnostics reuse existing logs.
   - **ARCH-EVOLVE — pass:** no persisted schema change; older logs remain accepted.

7. **Plan revision recommendations**

   Append `## Revisions` entries documenting execution-budget isolation, conservative allocation/creator identity rules, resume diagnostic propagation, and `savedAttemptRecord` as an existing reused entity. Update completed test claims to name the added regressions.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"79949986-fe8e-4cc8-ac8f-ecea0b0d19a3","repository":"github.com/xianxu/ariadne","issue_id":"000589","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/f334c81f122f4d93","reviewed_head":"570681b8038d97147d5a9bd3c7529d528c0625f9","slot":"ariadne:3","requested_at":"2026-10-24T20:15:09.741732Z"},"phase":"finished","pid":10796,"started_at":"2026-10-24T20:15:09.901024Z","finished_at":"2026-10-24T20:17:38.753123Z","session_id":"01a16f8d-8eaf-7242-bed6-0be76fe8711f","outcome":"returned"}] -->
