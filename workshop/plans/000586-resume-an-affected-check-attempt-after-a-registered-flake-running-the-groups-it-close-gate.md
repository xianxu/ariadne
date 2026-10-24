---
gate: boundary-review
issue: 586
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 4aa63ddb-3c84-4f59-9c99-52a67cdd9908
            repository: github.com/xianxu/ariadne
            issue_id: "000586"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/b516fd6f0ea1f078
            reviewed_head: 7a090b0b89683c76028d1ea9f1f93be7168710ec
            slot: ariadne:2
            requested_at: "2026-10-24T09:33:43.372342Z"
          phase: finished
          pid: 35815
          started_at: "2026-10-24T09:33:43.469422Z"
          finished_at: "2026-10-24T09:40:02.18165Z"
          session_id: 9312c287-ca72-4c8e-bfef-8dd92cfcf629
          outcome: returned
      "n": 1
      timestamp: "2026-10-24T02:40:08-07:00"
      dispatched: "2026-10-24T02:33:43-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: README update appears missing for local-check resume --name affected
          detail: README.md:1016-1018 still documents resume as suite-only (--name suite, "runs only the exact package tests"). The new --name affected value and its rerun-every-selected-group behavior are user-facing CLI surface; help text and atlas were updated but README was not.
          family: readme-surface-sync
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Affected resume production wiring (affectedResumePlan) is untested and diverges from the plan's affectedDeps seam
          detail: resumeSaved takes an opaque resumeSelectionFunc and every test substitutes the whole planner, so the origin/main pin, the missing-affected.json refusal, the GraphPrepare signal scope and the planAffectedSelection call (localcheck_resume_cmd.go:411-433) are never exercised. The plan said tests inject the existing fake affectedDeps (ARCH-MOCK, ARCH-PURE). Add a test that drives the planner through a fake affectedDeps, and record the seam change in the plan's Revisions.
          family: production-seam-untested
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: Resumed affected Scope reflects the replanned selection, not the Go packages the parent ran
          detail: If main's affected.json or the graph moved since the parent ran, the status can name Go packages that never ran in parent or child (D3 accepts this). Document it in the help text, or record that only the groups were rerun.
          family: status-scope-truthfulness
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: No old-reader fence test for a version 5 resume-reconciliation receipt
          detail: TestOldReaderFencesAffectedLinkPerHead covers only the link; a v5 receipt passing decodeUnsupportedProof is assumed, not tested.
          family: old-reader-fence-coverage
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: Plan names tests and a resumeSaved signature that differ from what shipped; plan checkboxes unchecked
          detail: TestAffectedKnownFlakeGroupsFailRefuses, TestResumeAffectedGroupFailureRecordsFailure and TestResumeAffectedFullSelectionRefuses exist under other names. resumeSaved takes a planner closure, not affectedDeps. Add a Revisions entry.
          family: plan-code-drift
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#586 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-24T02:40:08-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `readme-surface-sync` (kind: defect) README update appears missing for local-check resume --name affected
  README.md:1016-1018 still documents resume as suite-only (--name suite, "runs only the exact package tests"). The new --name affected value and its rerun-every-selected-group behavior are user-facing CLI surface; help text and atlas were updated but README was not.
- **BR-2** [Important] `production-seam-untested` (kind: defect) Affected resume production wiring (affectedResumePlan) is untested and diverges from the plan's affectedDeps seam
  resumeSaved takes an opaque resumeSelectionFunc and every test substitutes the whole planner, so the origin/main pin, the missing-affected.json refusal, the GraphPrepare signal scope and the planAffectedSelection call (localcheck_resume_cmd.go:411-433) are never exercised. The plan said tests inject the existing fake affectedDeps (ARCH-MOCK, ARCH-PURE). Add a test that drives the planner through a fake affectedDeps, and record the seam change in the plan's Revisions.
- **BR-3** [Minor] `status-scope-truthfulness` (kind: defect) Resumed affected Scope reflects the replanned selection, not the Go packages the parent ran
  If main's affected.json or the graph moved since the parent ran, the status can name Go packages that never ran in parent or child (D3 accepts this). Document it in the help text, or record that only the groups were rerun.
- **BR-4** [Minor] `old-reader-fence-coverage` (kind: defect) No old-reader fence test for a version 5 resume-reconciliation receipt
  TestOldReaderFencesAffectedLinkPerHead covers only the link; a v5 receipt passing decodeUnsupportedProof is assumed, not tested.
- **BR-5** [Minor] `plan-code-drift` (kind: defect) Plan names tests and a resumeSaved signature that differ from what shipped; plan checkboxes unchecked
  TestAffectedKnownFlakeGroupsFailRefuses, TestResumeAffectedGroupFailureRecordsFailure and TestResumeAffectedFullSelectionRefuses exist under other names. resumeSaved takes a planner closure, not affectedDeps. Add a Revisions entry.

## Open findings

- **BR-1** [Important] `readme-surface-sync` (kind: defect) README update appears missing for local-check resume --name affected
- **BR-2** [Important] `production-seam-untested` (kind: defect) Affected resume production wiring (affectedResumePlan) is untested and diverges from the plan's affectedDeps seam
- **BR-3** [Minor] `status-scope-truthfulness` (kind: defect) Resumed affected Scope reflects the replanned selection, not the Go packages the parent ran
- **BR-4** [Minor] `old-reader-fence-coverage` (kind: defect) No old-reader fence test for a version 5 resume-reconciliation receipt
- **BR-5** [Minor] `plan-code-drift` (kind: defect) Plan names tests and a resumeSaved signature that differ from what shipped; plan checkboxes unchecked
