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

## Open findings

- **BR-4** [Minor] `silent-error-swallow` rev-parse failure on the tracker ref is dropped, leaving tracker present with an empty ref
- **BR-5** [Minor] `contract-enum-validation` Validate checks landing.outcome against its enum but not assignment.relation or claimant_worktree
- **BR-6** [Minor] `state-semantics-overload` M2 sections are emitted as unknown ("not observed by this build"), overloading the read-failed meaning; remove in M2
- **BR-7** [Important] `silent-error-swallow` A details parse or read failure at the evidence location yields checkpoints present with plan 0/0 and hides unknown milestones
- **BR-8** [Important] `contract-enum-validation` Validate leaves review boundary, verdict, and flow kind and provenance unchecked against their sets
- **BR-9** [Important] `plan-deliverable-dropped` The M1-milestone lifecycle case and the Review-Verdict trailer cross-check were dropped with no Revisions entry; the milestone path is untested
- **BR-10** [Important] `repo-root-from-wrong-anchor` collectEvidence reads the details from the WF_ISSUES_DIR env default, not the record DetailPath or the given --issues-dir
- **BR-11** [Important] `artifact-layout-restated` The collector hard-codes sidecar and ledger file names and parses milestones by hand instead of using the existing single sources
- **BR-12** [Minor] `operating-envelope-unmeasured` workspace.Resolve goes around the counted observeGit seam, so the 20-command bound test undercounts
- **BR-13** [Minor] `artifact-layout-restated` sidecarRows keeps the last row match, so a verdict row inside a review body can override the metadata table
- **BR-14** [Minor] `plan-deliverable-dropped` The --repo test queries from a non-repo temp dir, not from a second tracker repository as the plan states
