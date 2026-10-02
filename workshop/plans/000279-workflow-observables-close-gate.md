---
gate: boundary-review
issue: 279
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-02T00:11:32-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Critical
          title: 'issue show resolves the observed repository from cwd, not the issues dir: dies outside a repo and observes the wrong repo with --issues-dir'
          detail: 'issueShowRepo (cmd/sdlc/issue.go:626) uses repoRootOf("."). Reproduced: `sdlc issue show 5` in a non-git dir now dies with ". is not inside a repository" (the base version worked via LoadRecords(nil)). Run from ariadne with --issues-dir pointing at a temp dir, it printed the temp issue (open) and then ariadne''s own #5 observation (done, landed). Derive the root from the issues dir unless --repo is given, degrade without a repository, and add regression tests for both.'
          family: repo-root-from-wrong-anchor
          round: 1
        - id: BR-2
          severity: Important
          title: M1 promises a golden fixture for the schema_version 1 JSON; none is in the diff
          detail: The round-trip test compares the document with itself, so a wire-format drift of the v1 contract goes undetected. Add internal/observe/testdata/*.golden.
          family: plan-deliverable-dropped
          round: 1
        - id: BR-3
          severity: Important
          title: TestRunIssueShow_HeadersNotBodies now does a real tracker fetch against the developer's checkout
          detail: With an absolute temp IssuesDir and cwd inside ariadne, the deferred observation runs loadIssueRecordsAt(PreferFresh) on the real repository and updates its remote-tracking ref during go test (ARCH-SECURE). Fixed by the root-anchor fix; check that no other test keeps the cwd fallback.
          family: test-reaches-real-state
          round: 1
        - id: BR-4
          severity: Minor
          title: rev-parse failure on the tracker ref is dropped, leaving tracker present with an empty ref
          family: silent-error-swallow
          round: 1
        - id: BR-5
          severity: Minor
          title: Validate checks landing.outcome against its enum but not assignment.relation or claimant_worktree
          family: contract-enum-validation
          round: 1
        - id: BR-6
          severity: Minor
          title: M2 sections are emitted as unknown ("not observed by this build"), overloading the read-failed meaning; remove in M2
          family: state-semantics-overload
          round: 1
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 2
      timestamp: "2026-10-02T00:17:40-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: issueShowRepo anchors on the abs issues dir or --repo; TestObserveAnchorsOnTheIssuesDir fails under the old cwd anchor (both the non-repo and the wrong-repo halves).
          round: 2
        - id: BR-2
          disposition: addressed
          note: testdata/observation-v1.golden.json, compared byte-for-byte by TestObservationGolden and strictly decoded.
          round: 2
        - id: BR-3
          disposition: addressed
          note: newTestDirs is a non-git temp dir, so root is empty and nothing is fetched; no other test calls runIssueShow with a cwd-relative dir inside the real checkout.
          round: 2
        - id: BR-4
          disposition: not-addressed
          note: TrackerRefErr/RefError is wired up, but no test reaches the rev-parse failure path.
          round: 2
        - id: BR-5
          disposition: not-addressed
          note: 'relation and claimant_worktree are now checked, but there is no rejection test, and Authority on every section, plus the fixed authority per section, is still unvalidated. Rule: every enum-typed contract field validates itself, with one table test covering all of them.'
          round: 2
        - id: BR-6
          disposition: not-addressed
          note: Still present by design until M2, and now also baked into the v1 golden fixture; M2 must update both.
          round: 2
      boundary: M1
      recipe: milestone-review
      blocked: false
    - "n": 3
      timestamp: "2026-10-02T00:37:25-07:00"
      agent: claude
      findings:
        - id: BR-7
          severity: Important
          title: A details parse or read failure at the evidence location yields checkpoints present with plan 0/0 and hides unknown milestones
          detail: 'This is the 2nd finding in this family. In assembleCheckpoints, failures from issue.Parse, flow.FromFrontmatter and PlanItemsBody, and nil Details, are all dropped while the read stays Present. The ticked-milestone set then comes out empty, so a closed milestone with no artifact is omitted instead of reported unknown. collectEvidence also merges a git show failure on the details into "not found". Rule: a read-fed value never falls back to its zero value on failure; it carries its own read quality or degrades the section to unknown with the error. Sweep Parse, FromFrontmatter, PlanItemsBody, the details derr, and workspace.Resolve.'
          family: silent-error-swallow
          round: 3
        - id: BR-8
          severity: Important
          title: Validate leaves review boundary, verdict, and flow kind and provenance unchecked against their sets
          detail: 'This is the 2nd finding in this family; the BR-5 Log claim overstates what was done. boundary is only checked for non-empty, verdict passes sidecar text straight through, and flow is not checked. Rule: every enum-typed contract field is validated against a set derived from its authority, through one table: vocab verdict tokens, flow kinds and provenances, and plan, close or Mx for boundaries. Add a rejection row per field; collector side, a verdict outside the set is unknown with an error.'
          family: contract-enum-validation
          round: 3
        - id: BR-9
          severity: Important
          title: The M1-milestone lifecycle case and the Review-Verdict trailer cross-check were dropped with no Revisions entry; the milestone path is untested
          detail: 'This is the 2nd finding in this family. No real-git test runs milestone-close, so these collector paths have zero fixtures: the m-x review filename parse, the per-milestone FilterBoundary and openScopeFor scoping, and the plan-gate ledger read. The trailer cross-check is not implemented and not withdrawn. Rule: every bullet a boundary claims maps to a named test or a Revisions entry; enumerate the M2 bullets before milestone-close.'
          family: plan-deliverable-dropped
          round: 3
        - id: BR-10
          severity: Important
          title: collectEvidence reads the details from the WF_ISSUES_DIR env default, not the record DetailPath or the given --issues-dir
          detail: 'This is the 2nd finding in this family. collectEvidence never receives issuesDir, so a non-default --issues-dir finds no details at the ref; with the silent-error-swallow finding, the output is present with plan 0/0. Rule (same as the M1 lesson): every path the collector reads derives from the given inputs or the authority''s record, never ambient cwd or env. Apply it to the plans and history roots too.'
          family: repo-root-from-wrong-anchor
          round: 3
        - id: BR-11
          severity: Important
          title: The collector hard-codes sidecar and ledger file names and parses milestones by hand instead of using the existing single sources
          detail: It hard-codes the plan-gate, close-gate and close-review suffixes, and classifies milestones with a HasPrefix "m" test. The single sources are planGateSuffix, boundaryGateSuffix, sidecarPath and sidecarPathFor (reviewsidecar.go), and reviewMilestoneRe and classifyFamily (resolve.go). The prefix test over-matches any stem-m*-review file, and a writer rename would silently empty the reviews list (ARCH-DRY).
          family: artifact-layout-restated
          round: 3
        - id: BR-12
          severity: Minor
          title: workspace.Resolve goes around the counted observeGit seam, so the 20-command bound test undercounts
          family: operating-envelope-unmeasured
          round: 3
        - id: BR-13
          severity: Minor
          title: sidecarRows keeps the last row match, so a verdict row inside a review body can override the metadata table
          family: artifact-layout-restated
          round: 3
        - id: BR-14
          severity: Minor
          title: The --repo test queries from a non-repo temp dir, not from a second tracker repository as the plan states
          family: plan-deliverable-dropped
          round: 3
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 4
      timestamp: "2026-10-02T00:49:06-07:00"
      agent: claude
      dispose:
        - id: BR-7
          disposition: addressed
          note: checkpoints.go:121-145 degrades on DetailsErr, nil Details, Parse and FromFrontmatter failures; collectEvidence keeps the git show error as DetailsErr; Resolve errors are propagated; TestCheckpointsDegradeOnFailedReads has one row per failure.
          round: 4
        - id: BR-8
          disposition: not-addressed
          note: Validate checks verdict and flow (json.go:63-78), but TestValidateRejectsEveryUnknownEnum has no verdict, flow kind or provenance row and only an empty boundary, so removing those checks leaves every test passing; the flow sets are restated from flow.go:169-173 instead of derived.
          round: 4
        - id: BR-9
          disposition: addressed
          note: TestObserveMilestonesThroughTheRealGates covers the m-x filename parse, the per-milestone scoping (M1 0, M2 1) and the plan ledger; the trailer cross-check is withdrawn with reasons in the plan Revisions.
          round: 4
        - id: BR-10
          disposition: addressed
          note: collectEvidence takes the issues dir relative to root; plans and history come from vocab discovery; TestObserveEvidenceFollowsTheGivenIssuesDir.
          round: 4
        - id: BR-11
          disposition: addressed
          note: Names come from sidecarPath, sidecarPathFor, planGateSuffix, boundaryGateSuffix and reviewMilestoneRe, with a round-trip check that a matched file maps back to the same name.
          round: 4
        - id: BR-12
          disposition: not-addressed
          note: Resolve now goes through observeGit, but observe.go:58, :75 and :111 (tracker rev-parse, worktree list, archivedOnMain ls-tree) still call gitx.RunGit directly, so the 20-command bound still undercounts; route every git command one observation runs through the seam.
          round: 4
        - id: BR-13
          disposition: addressed
          note: sidecarRows keeps the first row; the quoted-REWORK case in TestCheckpointsDegradeOnFailedReads.
          round: 4
        - id: BR-14
          disposition: addressed
          note: TestObserveAcrossTrackerRepositories queries repository bravo from inside repository alpha, and alpha without --repo.
          round: 4
      findings:
        - id: BR-15
          severity: Minor
          title: issue.TickedMilestones has no direct unit test for repeated milestone rows or the [.] state
          detail: It is only exercised through the observe tests, with single-row milestones. The rule that a milestone counts as closed only when every row with its tag is ticked is untested.
          family: pure-helper-untested
          round: 4
        - id: BR-16
          severity: Minor
          title: boundaryRE in json.go restates the milestone tag pattern from milestonePlanRE
          detail: 'This is the 3rd finding in family artifact-layout-restated. Rule: a grammar the writer owns (milestone tags, artifact names) is matched by the writer''s exported pattern or helper, never retyped. Export the tag pattern from internal/issue and build boundaryRE from it.'
          family: artifact-layout-restated
          round: 4
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 5
      timestamp: "2026-10-02T00:56:02-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: addressed
          note: json.go validates boundary (grammar from issue.MilestoneTagPattern), verdict (vocab IsEmitted), flow kind/provenance (flow.Valid*); each has a rejection row asserting its own field; collector maps an out-of-set verdict to unknown.
          round: 5
        - id: BR-12
          disposition: addressed
          note: collectHolding passes observeGitReader to workspace.Resolve, so its reads hit the counted observeGit seam.
          round: 5
        - id: BR-15
          disposition: addressed
          note: TestTickedMilestones in internal/issue/plan_test.go covers repeated rows, the in-progress state, lettered and bold tags.
          round: 5
        - id: BR-16
          disposition: addressed
          note: boundaryRE is built from the exported issue.MilestoneTagPattern, which also builds milestonePlanRE.
          round: 5
      findings:
        - id: BR-17
          severity: Minor
          title: Validate does not tie a review's verdict to its read state; a present close review with an empty verdict validates
          detail: 'This is the 3rd finding in family contract-enum-validation. Rule: every conditionally-set contract field states its "set exactly when" invariant in Validate (relation and outcome already do); add verdict set exactly when the read is present and the boundary is not plan, plus a rejection row.'
          family: contract-enum-validation
          round: 5
        - id: BR-18
          severity: Minor
          title: The per-query git bound counts only collector calls; tracker loading (RepositoryForCheckout Resolve, fetch, card reads) bypasses observeGit
          detail: 'This is the 2nd finding in family operating-envelope-unmeasured. Rule: an envelope assertion covers the whole query path or states its scope; either count the tracker layer''s runner or narrow the test comment and commit claim to the collector.'
          family: operating-envelope-unmeasured
          round: 5
        - id: BR-19
          severity: Minor
          title: The milestonePlanRE doc comment in plan.go now attaches to the MilestoneTagPattern const
          family: doc-comment-attachment
          round: 5
      boundary: M2
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#279 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-02T00:11:32-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Critical] `repo-root-from-wrong-anchor` issue show resolves the observed repository from cwd, not the issues dir: dies outside a repo and observes the wrong repo with --issues-dir
  issueShowRepo (cmd/sdlc/issue.go:626) uses repoRootOf("."). Reproduced: `sdlc issue show 5` in a non-git dir now dies with ". is not inside a repository" (the base version worked via LoadRecords(nil)). Run from ariadne with --issues-dir pointing at a temp dir, it printed the temp issue (open) and then ariadne's own #5 observation (done, landed). Derive the root from the issues dir unless --repo is given, degrade without a repository, and add regression tests for both.
- **BR-2** [Important] `plan-deliverable-dropped` M1 promises a golden fixture for the schema_version 1 JSON; none is in the diff
  The round-trip test compares the document with itself, so a wire-format drift of the v1 contract goes undetected. Add internal/observe/testdata/*.golden.
- **BR-3** [Important] `test-reaches-real-state` TestRunIssueShow_HeadersNotBodies now does a real tracker fetch against the developer's checkout
  With an absolute temp IssuesDir and cwd inside ariadne, the deferred observation runs loadIssueRecordsAt(PreferFresh) on the real repository and updates its remote-tracking ref during go test (ARCH-SECURE). Fixed by the root-anchor fix; check that no other test keeps the cwd fallback.
- **BR-4** [Minor] `silent-error-swallow` rev-parse failure on the tracker ref is dropped, leaving tracker present with an empty ref
- **BR-5** [Minor] `contract-enum-validation` Validate checks landing.outcome against its enum but not assignment.relation or claimant_worktree
- **BR-6** [Minor] `state-semantics-overload` M2 sections are emitted as unknown ("not observed by this build"), overloading the read-failed meaning; remove in M2

## Round 2 — 2026-10-02T00:17:40-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — issueShowRepo anchors on the abs issues dir or --repo; TestObserveAnchorsOnTheIssuesDir fails under the old cwd anchor (both the non-repo and the wrong-repo halves).
- BR-2 — addressed — testdata/observation-v1.golden.json, compared byte-for-byte by TestObservationGolden and strictly decoded.
- BR-3 — addressed — newTestDirs is a non-git temp dir, so root is empty and nothing is fetched; no other test calls runIssueShow with a cwd-relative dir inside the real checkout.
- BR-4 — not-addressed — TrackerRefErr/RefError is wired up, but no test reaches the rev-parse failure path.
- BR-5 — not-addressed — relation and claimant_worktree are now checked, but there is no rejection test, and Authority on every section, plus the fixed authority per section, is still unvalidated. Rule: every enum-typed contract field validates itself, with one table test covering all of them.
- BR-6 — not-addressed — Still present by design until M2, and now also baked into the v1 golden fixture; M2 must update both.

## Round 3 — 2026-10-02T00:37:25-07:00 (claude) — BLOCKED

### Raised

- **BR-7** [Important] `silent-error-swallow` A details parse or read failure at the evidence location yields checkpoints present with plan 0/0 and hides unknown milestones
  This is the 2nd finding in this family. In assembleCheckpoints, failures from issue.Parse, flow.FromFrontmatter and PlanItemsBody, and nil Details, are all dropped while the read stays Present. The ticked-milestone set then comes out empty, so a closed milestone with no artifact is omitted instead of reported unknown. collectEvidence also merges a git show failure on the details into "not found". Rule: a read-fed value never falls back to its zero value on failure; it carries its own read quality or degrades the section to unknown with the error. Sweep Parse, FromFrontmatter, PlanItemsBody, the details derr, and workspace.Resolve.
- **BR-8** [Important] `contract-enum-validation` Validate leaves review boundary, verdict, and flow kind and provenance unchecked against their sets
  This is the 2nd finding in this family; the BR-5 Log claim overstates what was done. boundary is only checked for non-empty, verdict passes sidecar text straight through, and flow is not checked. Rule: every enum-typed contract field is validated against a set derived from its authority, through one table: vocab verdict tokens, flow kinds and provenances, and plan, close or Mx for boundaries. Add a rejection row per field; collector side, a verdict outside the set is unknown with an error.
- **BR-9** [Important] `plan-deliverable-dropped` The M1-milestone lifecycle case and the Review-Verdict trailer cross-check were dropped with no Revisions entry; the milestone path is untested
  This is the 2nd finding in this family. No real-git test runs milestone-close, so these collector paths have zero fixtures: the m-x review filename parse, the per-milestone FilterBoundary and openScopeFor scoping, and the plan-gate ledger read. The trailer cross-check is not implemented and not withdrawn. Rule: every bullet a boundary claims maps to a named test or a Revisions entry; enumerate the M2 bullets before milestone-close.
- **BR-10** [Important] `repo-root-from-wrong-anchor` collectEvidence reads the details from the WF_ISSUES_DIR env default, not the record DetailPath or the given --issues-dir
  This is the 2nd finding in this family. collectEvidence never receives issuesDir, so a non-default --issues-dir finds no details at the ref; with the silent-error-swallow finding, the output is present with plan 0/0. Rule (same as the M1 lesson): every path the collector reads derives from the given inputs or the authority's record, never ambient cwd or env. Apply it to the plans and history roots too.
- **BR-11** [Important] `artifact-layout-restated` The collector hard-codes sidecar and ledger file names and parses milestones by hand instead of using the existing single sources
  It hard-codes the plan-gate, close-gate and close-review suffixes, and classifies milestones with a HasPrefix "m" test. The single sources are planGateSuffix, boundaryGateSuffix, sidecarPath and sidecarPathFor (reviewsidecar.go), and reviewMilestoneRe and classifyFamily (resolve.go). The prefix test over-matches any stem-m*-review file, and a writer rename would silently empty the reviews list (ARCH-DRY).
- **BR-12** [Minor] `operating-envelope-unmeasured` workspace.Resolve goes around the counted observeGit seam, so the 20-command bound test undercounts
- **BR-13** [Minor] `artifact-layout-restated` sidecarRows keeps the last row match, so a verdict row inside a review body can override the metadata table
- **BR-14** [Minor] `plan-deliverable-dropped` The --repo test queries from a non-repo temp dir, not from a second tracker repository as the plan states

## Round 4 — 2026-10-02T00:49:06-07:00 (claude) — BLOCKED

### Disposed

- BR-7 — addressed — checkpoints.go:121-145 degrades on DetailsErr, nil Details, Parse and FromFrontmatter failures; collectEvidence keeps the git show error as DetailsErr; Resolve errors are propagated; TestCheckpointsDegradeOnFailedReads has one row per failure.
- BR-8 — not-addressed — Validate checks verdict and flow (json.go:63-78), but TestValidateRejectsEveryUnknownEnum has no verdict, flow kind or provenance row and only an empty boundary, so removing those checks leaves every test passing; the flow sets are restated from flow.go:169-173 instead of derived.
- BR-9 — addressed — TestObserveMilestonesThroughTheRealGates covers the m-x filename parse, the per-milestone scoping (M1 0, M2 1) and the plan ledger; the trailer cross-check is withdrawn with reasons in the plan Revisions.
- BR-10 — addressed — collectEvidence takes the issues dir relative to root; plans and history come from vocab discovery; TestObserveEvidenceFollowsTheGivenIssuesDir.
- BR-11 — addressed — Names come from sidecarPath, sidecarPathFor, planGateSuffix, boundaryGateSuffix and reviewMilestoneRe, with a round-trip check that a matched file maps back to the same name.
- BR-12 — not-addressed — Resolve now goes through observeGit, but observe.go:58, :75 and :111 (tracker rev-parse, worktree list, archivedOnMain ls-tree) still call gitx.RunGit directly, so the 20-command bound still undercounts; route every git command one observation runs through the seam.
- BR-13 — addressed — sidecarRows keeps the first row; the quoted-REWORK case in TestCheckpointsDegradeOnFailedReads.
- BR-14 — addressed — TestObserveAcrossTrackerRepositories queries repository bravo from inside repository alpha, and alpha without --repo.

### Raised

- **BR-15** [Minor] `pure-helper-untested` issue.TickedMilestones has no direct unit test for repeated milestone rows or the [.] state
  It is only exercised through the observe tests, with single-row milestones. The rule that a milestone counts as closed only when every row with its tag is ticked is untested.
- **BR-16** [Minor] `artifact-layout-restated` boundaryRE in json.go restates the milestone tag pattern from milestonePlanRE
  This is the 3rd finding in family artifact-layout-restated. Rule: a grammar the writer owns (milestone tags, artifact names) is matched by the writer's exported pattern or helper, never retyped. Export the tag pattern from internal/issue and build boundaryRE from it.

## Round 5 — 2026-10-02T00:56:02-07:00 (claude) — passed

### Disposed

- BR-8 — addressed — json.go validates boundary (grammar from issue.MilestoneTagPattern), verdict (vocab IsEmitted), flow kind/provenance (flow.Valid*); each has a rejection row asserting its own field; collector maps an out-of-set verdict to unknown.
- BR-12 — addressed — collectHolding passes observeGitReader to workspace.Resolve, so its reads hit the counted observeGit seam.
- BR-15 — addressed — TestTickedMilestones in internal/issue/plan_test.go covers repeated rows, the in-progress state, lettered and bold tags.
- BR-16 — addressed — boundaryRE is built from the exported issue.MilestoneTagPattern, which also builds milestonePlanRE.

### Raised

- **BR-17** [Minor] `contract-enum-validation` Validate does not tie a review's verdict to its read state; a present close review with an empty verdict validates
  This is the 3rd finding in family contract-enum-validation. Rule: every conditionally-set contract field states its "set exactly when" invariant in Validate (relation and outcome already do); add verdict set exactly when the read is present and the boundary is not plan, plus a rejection row.
- **BR-18** [Minor] `operating-envelope-unmeasured` The per-query git bound counts only collector calls; tracker loading (RepositoryForCheckout Resolve, fetch, card reads) bypasses observeGit
  This is the 2nd finding in family operating-envelope-unmeasured. Rule: an envelope assertion covers the whole query path or states its scope; either count the tracker layer's runner or narrow the test comment and commit claim to the collector.
- **BR-19** [Minor] `doc-comment-attachment` The milestonePlanRE doc comment in plan.go now attaches to the MilestoneTagPattern const

## Open findings

- **BR-4** [Minor] `silent-error-swallow` rev-parse failure on the tracker ref is dropped, leaving tracker present with an empty ref
- **BR-5** [Minor] `contract-enum-validation` Validate checks landing.outcome against its enum but not assignment.relation or claimant_worktree
- **BR-6** [Minor] `state-semantics-overload` M2 sections are emitted as unknown ("not observed by this build"), overloading the read-failed meaning; remove in M2
- **BR-17** [Minor] `contract-enum-validation` Validate does not tie a review's verdict to its read state; a present close review with an empty verdict validates
- **BR-18** [Minor] `operating-envelope-unmeasured` The per-query git bound counts only collector calls; tracker loading (RepositoryForCheckout Resolve, fetch, card reads) bypasses observeGit
- **BR-19** [Minor] `doc-comment-attachment` The milestonePlanRE doc comment in plan.go now attaches to the MilestoneTagPattern const
