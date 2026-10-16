# Boundary Review — ariadne#408 (whole-issue close)

| field | value |
|-------|-------|
| issue | 408 — issue publish of an archived issue creates an active twin under workshop/issues (pair#579) |
| repo | ariadne |
| issue file | workshop/issues/000408-issue-publish-of-an-archived-issue-creates-an-active-twin-under-workshop-issues-pair-579.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1a8c52756b74160f23240dd123ddc25f5b20ed32..7ca9a41743d8986f4ee5211494eac15355da3737 |
| command | sdlc close --issue 408 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T11:51:25-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The main archive-publishing correction works in the focused tests. However, initial-publication recovery still permits a concurrent archive to gain an active twin. The local scan also misclassifies non-Markdown entries, and acceptance criteria omit the added local-only archive behavior.

```findings
findings:
  - id: new
    severity: Critical
    family: publication-location-revalidation
    title: |
      Initial-publication recovery does not revalidate archive locations
    detail: |
      cmd/sdlc/issuemovedetail.go:134 checks locations only at preflight. If an archive appears afterward, transfer candidate preparation uses DetailBundlePublication (internal/tracker/detailpublish.go:27), which checks only the active destination; newIssueTransferOp's BeforePublish checks assets only. The resulting candidate can publish an active twin. ARCH-PURPOSE: sweep initial candidate preparation, retry preparation, and recovery replay for both legacy flat history and history/issues; republish candidate preparation already performs the ID-wide check. Add regression coverage that introduces an archive after preflight and verifies refusal without an active write.
  - id: new
    severity: Important
    family: detail-filename-grammar
    title: |
      Local detail discovery accepts non-Markdown entries
    detail: |
      cmd/sdlc/issuepublish.go:103 passes LocalPathsByID results directly to classifyFamily, whose default class accepts any matching ID prefix. Thus 000012-note.md.bak beside valid details causes a false duplicate refusal. The same unfiltered collection also drives the local-archive refusal at line 119, where an archived backup alone can block initial publication. These are all instances in this window. Filter through issue.ParseFilename before family classification, matching the published resolver, and test backup and non-Markdown entries.
  - id: new
    severity: Important
    family: acceptance-contract-completeness
    title: |
      Done when omits the local-only archive revision
    detail: |
      The issue's Revisions section at line 104 adds refusal when archived details exist locally but not on main, implemented at issuepublish.go:119 and tested by TestPublishUnpublishedLocalArchiveRefuses. Done when at lines 38–39 covers archive updates, missing done-card details, and twins, but not this distinct state with an open card and one unpublished archive. Append explicit acceptance for refusal, preservation of the local archive, and no active file or main change.
```

**Strengths**

- Reuses canonical directories, filename parsing, family classification, and vocabulary terminal membership.
- Real-Git fixtures verify both archive layouts and atomic mixed active/archive republishing.
- Republish candidate preparation rechecks location before writing, including a concurrent-twin regression.

**Critical findings**

- Initial-publication location revalidation is incomplete; see `publication-location-revalidation` above.

**Important findings**

- Local filename filtering and acceptance-contract completeness need correction as detailed above.

**Minor findings**

- None.

**Test coverage notes**

All six added focused test functions passed: `go test ./cmd/sdlc -run 'TestPublish(ArchivedDetails|TerminalCardWithoutDetailsRefuses|DuplicateLocationsRefuses|RefusesTwinIntroducedBeforeCandidate|UnpublishedLocalArchiveRefuses)$' -count=1` (21.134s). Diff whitespace checks passed. The concurrent test exercises republishing only; initial transfer and filename false positives remain uncovered. I did not independently rerun the reported full suite.

Done-when: stale — the local-only unpublished archive refusal added in Revisions has no corresponding acceptance clause.

**Architectural notes**

- **ARCH-DRY: pass.** Existing shared primitives supply the new resolution behavior.
- **ARCH-PURE: pass.** Terminal classification remains pure; repository reads stay in publication orchestration.
- **ARCH-PURPOSE: flag.** Initial transfer and recovery omit the location safeguard enforced by republishing.

Atlas documents the correction. No new command, flag, or configuration surface requires a README update. There is no Core concepts table to cross-check.

**Plan revision recommendations**

Append `## Revisions` entries covering complete transfer/recovery location revalidation and the missing local-only archive acceptance criteria. Reopen the candidate-recheck Plan item until those paths are covered.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"22ab9433-737e-4cb6-a283-c4155d8d806e","repository":"github.com/xianxu/ariadne","issue_id":"000408","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/abe73d1142956477","reviewed_head":"7ca9a41743d8986f4ee5211494eac15355da3737","slot":"ariadne:3","requested_at":"2026-10-16T18:49:51.637226Z"},"phase":"finished","pid":92349,"started_at":"2026-10-16T18:49:51.865379Z","finished_at":"2026-10-16T18:51:24.477328Z","session_id":"01a1460c-965f-7601-91f4-1c8509cf8602","outcome":"returned"}] -->
