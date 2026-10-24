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
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: f7282e4a-166b-4cc9-9f52-a7252a3d3f10
            repository: github.com/xianxu/ariadne
            issue_id: "000586"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/582a3acab5c47cb2
            reviewed_head: fe8d4999b53cb6f415809bd7c8413c962e58cdc7
            slot: ariadne:2
            requested_at: "2026-10-24T09:51:07.632288Z"
          phase: finished
          pid: 28216
          started_at: "2026-10-24T09:51:07.786298Z"
          finished_at: "2026-10-24T09:57:15.206534Z"
          session_id: 2cb3a0e3-3156-49ca-b031-4d77e6d4a7cc
          outcome: returned
      "n": 2
      timestamp: "2026-10-24T02:57:16-07:00"
      dispatched: "2026-10-24T02:51:07-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: README.md:1016-1023 now documents --name affected, replanning against main's affected.json, running every group, and the full/empty refusal toward suite.
          round: 2
        - id: BR-2
          disposition: addressed
          note: TestAffectedResumePlanWiring drives affectedResumePlan through a fake affectedDeps; pinning main instead of origin/main turns it red. Only the GraphPrepare signal branch is unexercised.
          round: 2
        - id: BR-3
          disposition: addressed
          note: helptext/local-check.md now says the scope is the replanned selection and can name packages neither run tested; the plan Revisions records it.
          round: 2
        - id: BR-4
          disposition: addressed
          note: TestOldReaderFencesAffectedReceiptPerHead decodes a v5 receipt through decodeUnsupportedProof as proof fenced to that head.
          round: 2
        - id: BR-5
          disposition: addressed
          note: The plan Revisions entry (2026-10-24) lists the shipped test names and the resumeSaved planner signature; checkboxes are ticked.
          round: 2
      findings:
        - id: BR-6
          severity: Minor
          title: GraphPrepare signal-context block duplicated in affectedResumePlan and runAffectedCheck
          detail: localcheck_resume_cmd.go:572-576 copies localcheck_affected.go:56-60 (ARCH-DRY); fold both into one helper so the signal ownership of graph planning has one source.
          family: shared-helper-duplication
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: fe8d4999b53cb6f415809bd7c8413c962e58cdc7
      blocked: false
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

## Round 2 — 2026-10-24T02:57:16-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — README.md:1016-1023 now documents --name affected, replanning against main's affected.json, running every group, and the full/empty refusal toward suite.
- BR-2 — addressed — TestAffectedResumePlanWiring drives affectedResumePlan through a fake affectedDeps; pinning main instead of origin/main turns it red. Only the GraphPrepare signal branch is unexercised.
- BR-3 — addressed — helptext/local-check.md now says the scope is the replanned selection and can name packages neither run tested; the plan Revisions records it.
- BR-4 — addressed — TestOldReaderFencesAffectedReceiptPerHead decodes a v5 receipt through decodeUnsupportedProof as proof fenced to that head.
- BR-5 — addressed — The plan Revisions entry (2026-10-24) lists the shipped test names and the resumeSaved planner signature; checkboxes are ticked.

### Raised

- **BR-6** [Minor] `shared-helper-duplication` GraphPrepare signal-context block duplicated in affectedResumePlan and runAffectedCheck
  localcheck_resume_cmd.go:572-576 copies localcheck_affected.go:56-60 (ARCH-DRY); fold both into one helper so the signal ownership of graph planning has one source.

## Open findings

- **BR-6** [Minor] `shared-helper-duplication` GraphPrepare signal-context block duplicated in affectedResumePlan and runAffectedCheck
