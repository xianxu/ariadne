---
gate: boundary-review
issue: 589
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 79949986-fe8e-4cc8-ac8f-ecea0b0d19a3
            repository: github.com/xianxu/ariadne
            issue_id: "000589"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/f334c81f122f4d93
            reviewed_head: 570681b8038d97147d5a9bd3c7529d528c0625f9
            slot: ariadne:3
            requested_at: "2026-10-24T20:15:09.741732Z"
          phase: finished
          pid: 10796
          started_at: "2026-10-24T20:15:09.901024Z"
          finished_at: "2026-10-24T20:17:38.753123Z"
          session_id: 01a16f8d-8eaf-7242-bed6-0be76fe8711f
          outcome: returned
      "n": 1
      timestamp: "2026-10-24T13:17:40-07:00"
      dispatched: "2026-10-24T13:15:09-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: BR-1
          severity: Critical
          title: SysV observation time can turn a completed command into a timeout
          detail: 'cmd/sdlc/localcheck_run.go:676-687 runs observations under an already-running command timeout, then line 722 checks that timeout after observation completes. A successful command followed by a slow after-snapshot can therefore become TimedOut/LaneOverrun; the before-snapshot also consumes command execution time. ARCH-CONSTRAINTS and ARCH-PURPOSE: isolate diagnostic budgets from execution budgets and capture the execution outcome before post-run diagnostics. Add regressions with delayed before/after observers proving diagnostics cannot change an otherwise successful outcome.'
          family: advisory-outcome-isolation
          round: 1
        - id: BR-2
          severity: Critical
          title: A nonempty start time does not establish that the witnessed process created the segment
          detail: 'cmd/sdlc/internal/sysvshm/report.go:29-33 matches only the numeric creator PID and checks that Witness.Start is nonempty. No code compares that incarnation with evidence from the segment. If a concurrent creator exits and its PID is reused by a run-group process, the unrelated segment receives that process''s name and a witnessed-creator label. ARCH-SECURE and ARCH-ORDER: require evidence linking the creator incarnation to the allocation, otherwise retain unknown attribution. Test PID reuse into the owned group; the existing fake lookup returning false does not exercise this case.'
          family: creator-incarnation-evidence
          round: 1
        - id: BR-3
          severity: Critical
          title: Mutable change time is treated as allocation identity
          detail: 'cmd/sdlc/internal/sysvshm/report.go:11-26 uses the entire Segment, except Attachments, as identity. Changed comes from CTIME/ctime, which is the last shmctl change time rather than an immutable allocation generation. Updating permissions through IPC_SET can therefore make a pre-existing unattached segment appear newly allocated. ARCH-PURPOSE: separate stable identity from mutable metadata, treat ambiguous changes conservatively, and add a regression where one existing segment changes metadata without being replaced.'
          family: allocation-identity-stability
          round: 1
        - id: BR-4
          severity: Critical
          title: The Core concepts table claims a modification absent from the pinned range
          detail: workshop/plans/000589-recorded-suites-report-leaked-sysv-shared-memory-segments-postgresql-test-plan.md:40 marks savedAttemptRecord in cmd/sdlc/localcheck_republish.go as modified, but that file has no base-to-head diff. Replay already carries the supplied tail; the new behavior resides in statusDescription. Per the required Core concepts cross-check, correct the row to existing and append a Revisions entry explaining reuse of the existing validation path.
          family: concept-table-code-consistency
          round: 1
        - id: BR-5
          severity: Important
          title: Resume publication drops warnings produced only by the child run
          detail: 'cmd/sdlc/localcheck_resume_saved.go:117-138 builds the composite record using only the parent''s tail, while cmd/sdlc/localcheck_resume.go:228 discards the child''s hashAndTail tail. The newly instrumented resume batch can warn in its child log, yet fresh and saved composite publication omit the promised SysV warning prefix when the parent was clean. ARCH-PURPOSE: derive diagnostic status from both digest-verified logs and test a clean parent followed by a leaking child through publication and republish.'
          family: diagnostic-consumer-completeness
          round: 1
        - id: BR-6
          severity: Important
          title: README update appears missing for the new qualification usage
          detail: atlas/workflow/sdlc-binary.md:3394 introduces TestNativeOwnedSysVQualification and WF_SYSV_QUALIFY=1, but README.md is unchanged in the pinned range. The requested README gate covers new usage steps and configuration keys. Add a concise README pointer explaining host warnings and linking the operator-cleared qualification procedure.
          family: user-surface-documentation
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#589 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-24T13:17:40-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `advisory-outcome-isolation` SysV observation time can turn a completed command into a timeout
  cmd/sdlc/localcheck_run.go:676-687 runs observations under an already-running command timeout, then line 722 checks that timeout after observation completes. A successful command followed by a slow after-snapshot can therefore become TimedOut/LaneOverrun; the before-snapshot also consumes command execution time. ARCH-CONSTRAINTS and ARCH-PURPOSE: isolate diagnostic budgets from execution budgets and capture the execution outcome before post-run diagnostics. Add regressions with delayed before/after observers proving diagnostics cannot change an otherwise successful outcome.
- **BR-2** [Critical] `creator-incarnation-evidence` A nonempty start time does not establish that the witnessed process created the segment
  cmd/sdlc/internal/sysvshm/report.go:29-33 matches only the numeric creator PID and checks that Witness.Start is nonempty. No code compares that incarnation with evidence from the segment. If a concurrent creator exits and its PID is reused by a run-group process, the unrelated segment receives that process's name and a witnessed-creator label. ARCH-SECURE and ARCH-ORDER: require evidence linking the creator incarnation to the allocation, otherwise retain unknown attribution. Test PID reuse into the owned group; the existing fake lookup returning false does not exercise this case.
- **BR-3** [Critical] `allocation-identity-stability` Mutable change time is treated as allocation identity
  cmd/sdlc/internal/sysvshm/report.go:11-26 uses the entire Segment, except Attachments, as identity. Changed comes from CTIME/ctime, which is the last shmctl change time rather than an immutable allocation generation. Updating permissions through IPC_SET can therefore make a pre-existing unattached segment appear newly allocated. ARCH-PURPOSE: separate stable identity from mutable metadata, treat ambiguous changes conservatively, and add a regression where one existing segment changes metadata without being replaced.
- **BR-4** [Critical] `concept-table-code-consistency` The Core concepts table claims a modification absent from the pinned range
  workshop/plans/000589-recorded-suites-report-leaked-sysv-shared-memory-segments-postgresql-test-plan.md:40 marks savedAttemptRecord in cmd/sdlc/localcheck_republish.go as modified, but that file has no base-to-head diff. Replay already carries the supplied tail; the new behavior resides in statusDescription. Per the required Core concepts cross-check, correct the row to existing and append a Revisions entry explaining reuse of the existing validation path.
- **BR-5** [Important] `diagnostic-consumer-completeness` Resume publication drops warnings produced only by the child run
  cmd/sdlc/localcheck_resume_saved.go:117-138 builds the composite record using only the parent's tail, while cmd/sdlc/localcheck_resume.go:228 discards the child's hashAndTail tail. The newly instrumented resume batch can warn in its child log, yet fresh and saved composite publication omit the promised SysV warning prefix when the parent was clean. ARCH-PURPOSE: derive diagnostic status from both digest-verified logs and test a clean parent followed by a leaking child through publication and republish.
- **BR-6** [Important] `user-surface-documentation` README update appears missing for the new qualification usage
  atlas/workflow/sdlc-binary.md:3394 introduces TestNativeOwnedSysVQualification and WF_SYSV_QUALIFY=1, but README.md is unchanged in the pinned range. The requested README gate covers new usage steps and configuration keys. Add a concise README pointer explaining host warnings and linking the operator-cleared qualification procedure.

## Open findings

- **BR-1** [Critical] `advisory-outcome-isolation` SysV observation time can turn a completed command into a timeout
- **BR-2** [Critical] `creator-incarnation-evidence` A nonempty start time does not establish that the witnessed process created the segment
- **BR-3** [Critical] `allocation-identity-stability` Mutable change time is treated as allocation identity
- **BR-4** [Critical] `concept-table-code-consistency` The Core concepts table claims a modification absent from the pinned range
- **BR-5** [Important] `diagnostic-consumer-completeness` Resume publication drops warnings produced only by the child run
- **BR-6** [Important] `user-surface-documentation` README update appears missing for the new qualification usage
