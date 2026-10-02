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

## Open findings

- **BR-4** [Minor] `silent-error-swallow` rev-parse failure on the tracker ref is dropped, leaving tracker present with an empty ref
- **BR-5** [Minor] `contract-enum-validation` Validate checks landing.outcome against its enum but not assignment.relation or claimant_worktree
- **BR-6** [Minor] `state-semantics-overload` M2 sections are emitted as unknown ("not observed by this build"), overloading the read-failed meaning; remove in M2
