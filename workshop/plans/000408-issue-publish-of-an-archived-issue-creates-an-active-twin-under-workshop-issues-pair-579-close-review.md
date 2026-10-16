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

---

## Re-review — 2026-10-16T12:25:23-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 408 — issue publish of an archived issue creates an active twin under workshop/issues (pair#579) |
| repo | ariadne |
| issue file | workshop/issues/000408-issue-publish-of-an-archived-issue-creates-an-active-twin-under-workshop-issues-pair-579.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1a8c52756b74160f23240dd123ddc25f5b20ed32..411260a95cd22e012c9e0b74b5e73983b843001b |
| command | sdlc close --issue 408 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T12:25:23-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The archive-location fixes work, and all three prior findings are addressed. One related correctness gap remains: terminal-card admission is checked only at preflight. A card becoming `done` before publication can still receive new active details, including through saved-receipt recovery.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The shared BeforePublish hook now rechecks ID-wide locations. All six archive regressions pass on head and fail when this guard is removed, covering both archive layouts across first publication, moved-main retry and receipt recovery.
  - id: BR-2
    disposition: addressed
    note: |
      issuepublish.go:104 filters local candidates through the canonical filename parser before either local decision. All six discovery regressions pass on head and fail when the filter is removed.
  - id: BR-3
    disposition: addressed
    note: |
      The issue's Done when at lines 41–42 explicitly requires local-only archive refusal, byte preservation, unchanged main and no active details. This matches issuepublish.go:121–122 and TestPublishUnpublishedLocalArchiveRefuses, which passed.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: publication-location-revalidation
    title: |
      Terminal-card admission still bypasses publication-time revalidation
    detail: |
      cmd/sdlc/issuemovedetail.go:439–451 revalidates locations but never calls requireInitialDetails against the current card; that check remains confined to preflight. In a disposable pinned-head reproduction using archiveRaceTransfer, changing the card to done before transfer.main preparation made Drive return success, publish workshop/issues/000008-spin-off.md and remove the local source. Changing the card after saving the candidate also let recovery succeed; subsequent cleanup archived the newly published file instead of refusing. This violates Spec lines 32–33. This is the 2nd finding in family publication-location-revalidation: do not fix only this instance. State and enforce the shared rule that every unconfirmed initial publication requires both an eligible current card and no existing ID-wide details. Sweep first publication, candidate replay and moved-base retry, with terminal-status regressions and confirmed-publication recovery preserved. ARCH-PURPOSE, ARCH-ORDER.
```

1. **Strengths**

   - `publishedDetailPath` provides one resolver shared by republish and transfer admission.
   - Archive tests cover both layouts, mixed active/archive publication, duplicate refusal and source preservation.
   - Removing the BR-1 and BR-2 fixes independently exposed their intended failures across all twelve regression cases.
   - Atlas documents the changed resolution and recovery flow.

2. **Critical findings**

   - **Terminal-card revalidation**, `cmd/sdlc/issuemovedetail.go:439`: read the current card through the existing tracker seam and apply terminal admission before an unconfirmed publication. Preserve the existing reachability-first handling for already-published candidates.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Focused pinned-head tests passed in **117.139s**.
   - Removing the two prior behavioral fixes produced the expected twelve failures.
   - Additional terminal-card probes failed against unchanged production code: publication succeeded despite `status: done`. The first-publication probe confirmed an active main file and removed local source.
   - Disposable checkout removed; reviewed checkout remains clean at the pinned head. The full suite was not independently rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared resolver and canonical filename grammar.
   - **ARCH-PURE — pass:** the declared PURE admission helper and its direct test require no IO; Core concepts entities exist and match their stated changes.
   - **ARCH-PURPOSE — flag:** the no-resurrection contract remains incomplete for terminal-card changes.
   - **ARCH-MOCK — pass for this delta:** existing Git seam and isolated, stateful bare-origin fixtures exercise production operations.
   - **ARCH-CONSTRAINTS — pass:** bounded snapshot facilities are reused; no added concurrency.
   - **ARCH-SECURE — pass:** canonical paths and filename parsing are reused; read errors propagate.
   - **ARCH-ORDER — flag:** preflight card eligibility becomes stale across publication and receipt replay.
   - **ARCH-FUNERAL — pass:** no new durable artifact family.
   
   Atlas is updated. No new command, flag or configuration syntax requires a README addition.

7. **Plan revision recommendations**

   Append a `## Revisions` entry defining initial-publication eligibility as **current nonterminal card plus absent ID-wide details**. Enumerate checks and regression coverage across first publication, moved-base retry and saved-candidate recovery, while retaining confirmed-publication recovery.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b223f882-64e4-4c00-a786-1ed9ea257d6e","repository":"github.com/xianxu/ariadne","issue_id":"000408","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/fc9c54246fc5f9a6","reviewed_head":"411260a95cd22e012c9e0b74b5e73983b843001b","slot":"ariadne:3","requested_at":"2026-10-16T19:20:08.85452Z"},"phase":"finished","pid":57179,"started_at":"2026-10-16T19:20:09.200677Z","finished_at":"2026-10-16T19:25:19.349407Z","session_id":"01a14628-506c-7f80-b623-fe47c2be186c","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T13:02:44-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 408 — issue publish of an archived issue creates an active twin under workshop/issues (pair#579) |
| repo | ariadne |
| issue file | workshop/issues/000408-issue-publish-of-an-archived-issue-creates-an-active-twin-under-workshop-issues-pair-579.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1a8c52756b74160f23240dd123ddc25f5b20ed32..9da64867cec00ab9d09b1d3391e4988262134d86 |
| command | sdlc close --issue 408 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T13:02:44-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation satisfies the documented archive-publication contract. BR-4 is addressed through shared publication-time admission, with regression evidence covering all terminal statuses across first publication, retry, and receipt recovery. No blocking findings remain.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Archive revalidation remains in the shared BeforePublish path; all six archive-layout and publication-phase regression cases passed.
  - id: BR-2
    disposition: addressed
    note: |
      localDetailPath filters through the canonical filename parser before classification; all six backup/non-Markdown regression cases passed.
  - id: BR-3
    disposition: addressed
    note: |
      The pinned issue diff adds the local-only archive acceptance clause, matching TestPublishUnpublishedLocalArchiveRefuses and entry-point parity coverage.
  - id: BR-4
    disposition: addressed
    note: |
      checkInitialPublication reads the current card and calls requireInitialDetails before unconfirmed publication. Removing only that call in a disposable pinned-head copy made all nine terminal-status/phase cases fail with publication accepted. The unchanged head passes, including confirmed-publication recovery.
```

1. **Strengths**

   - One admission authority composes current-card eligibility, location checks, and existing asset safeguards (`cmd/sdlc/issuemovedetail.go:404`).
   - Both publication flows resolve details by ID across active and archive locations (`cmd/sdlc/issuepublish.go:182`).
   - Controlled transfer interruptions test retries and durable recovery through production code (`cmd/sdlc/issuepublish_archive_test.go:209`).
   - Confirmed-publication recovery remains reachable after a terminal transition without republishing (`cmd/sdlc/issuepublish_archive_test.go:256`).

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Two focused runs passed: archive publication, twins, terminal admission, filename filtering, local archives, retry, and recovery (23.518s and 51.036s). The BR-4 mutation failed all nine intended cases; its disposable checkout was removed. The reviewed checkout remains clean. The full suite was not independently rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared location resolver and consolidated admission.
   - **ARCH-PURE — pass:** terminal eligibility is directly tested without IO; integration remains in the boundary layer.
   - **ARCH-PURPOSE — pass:** both entry points and first/retry/replay paths consume the shared rule.
   - **ARCH-MOCK — pass:** existing Git seam, stateful fake, and real bare-Git conformance fixtures are retained.
   - **ARCH-CONSTRAINTS — pass:** existing directory inventory and bounded Git reads; no added concurrency.
   - **ARCH-SECURE — pass:** canonical filename filtering, repository containment, and visible read failures remain enforced.
   - **ARCH-ORDER — pass:** existing transition model remains authoritative; reachability precedes admission, preserving confirmed recovery.
   - **ARCH-FUNERAL — pass:** no new durable artifact family or cleanup obligation.

   Core-concept locations and classifications match the code. Atlas documents the changed flow. No new command, flag, or configuration syntax requires a README addition.

7. **Plan revision recommendations:** None. Remaining close/landing checkboxes appropriately await this boundary.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"21597bc7-f346-4c53-a1c3-cb8a6a7339db","repository":"github.com/xianxu/ariadne","issue_id":"000408","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/e080bcf673647fc0","reviewed_head":"9da64867cec00ab9d09b1d3391e4988262134d86","slot":"ariadne:3","requested_at":"2026-10-16T20:00:14.480152Z"},"phase":"finished","pid":69659,"started_at":"2026-10-16T20:00:14.636648Z","finished_at":"2026-10-16T20:02:44.484459Z","session_id":"01a1464d-046a-7882-98b1-1940f9bc2dfd","outcome":"returned"}] -->
