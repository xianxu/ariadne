---
gate: boundary-review
issue: 337
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T15:04:03-07:00"
      dispatched: "2026-10-13T14:58:58-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: Publication trusts stale checkout policy instead of current destination policy
          detail: 'cmd/sdlc/assetpolicy.go:23 reads only the working-tree policy; republish.go and assetbundlegit.go reuse it during publication. A temporary regression tightened main to text-only while leaving the caller allowed, and publication still pushed new asset bytes. ARCH-SECURE / ARCH-PURPOSE: resolve authorization from the pinned destination snapshot on every publication attempt and sweep all admission callers, including first-publication recovery, legacy publication, and abandonment.'
          family: destination-policy-authority
          round: 1
        - id: BR-2
          severity: Critical
          title: Reopening can commit converted asset bytes under their original content name
          detail: 'cmd/sdlc/abandon.go:464 merges and materializes archived assets before line 504 checks only active destinations. A temporary regression using archive-only text eol=crlf and active -text reopened successfully but changed LF evidence to CRLF under its unchanged hash filename. ARCH-PURPOSE: validate all materialization paths before effects and verify restored bytes against the pinned bundle; add a production reopen regression.'
          family: exact-bytes-at-materialization
          round: 1
        - id: BR-3
          severity: Critical
          title: The core-concepts table incorrectly declares archive proof PURE
          detail: 'workshop/plans/000337-issue-assets-plan.md:82 places landing archive plan and proof in the Pure entity table, while cmd/sdlc/landingarchive.go:556 confirmLandingArchive executes Git, loads snapshots, and reads time.Now; its proof tests use temporary repositories. ARCH-PURE: retain the planner as PURE, classify the proof as INTEGRATION, and append a Revisions entry correcting the table.'
          family: core-concept-classification
          round: 1
        - id: BR-4
          severity: Important
          title: Host visibility has no live conformance check
          detail: 'cmd/sdlc/assetvisibility_test.go:11 supplies canned JSON through a replacement gh executable, and assetpolicy_test.go models visibility with a state file; neither verifies the real RepoVisibility invocation against the host. ARCH-MOCK: add an opt-in read-only live conformance check through the production seam and document when it runs.'
          family: external-dependency-conformance
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - "n": 2
      timestamp: "2026-10-13T16:02:44-07:00"
      dispatched: "2026-10-13T15:57:42-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: Pinned destination policy now reaches publication and recovery callers. Existing regressions pass; disabling the destination check makes republish, first-publication, legacy and abandonment cases fail.
          round: 2
        - id: BR-2
          disposition: not-addressed
          note: 'Reopen is corrected, but the materialization-class sweep remains incomplete. cmd/sdlc/issuerestore.go:127 checks out assets, then line 136 commits working-tree paths without verifying pinned bytes. A temporary production regression installed a post-checkout hook; restore returned success and committed "hook-tampered" under the original hash filename. ARCH-PURPOSE / ARCH-SECURE: apply exact-byte verification across every materialization caller before committing.'
          round: 2
        - id: BR-3
          disposition: addressed
          note: workshop/plans/000337-issue-assets-plan.md:82 now classifies only planLandingArchive as PURE; line 95 classifies confirmLandingArchive as INTEGRATION. The Revisions correction matches the planner and Git/clock-backed proof implementations.
          round: 2
        - id: BR-4
          disposition: addressed
          note: TestLiveAssetVisibilityConformance calls the production adapter and fails opted-in prerequisite errors. Atlas documents invocation and cadence. Independently queried visibility was public; the live production test passed against that expectation.
          round: 2
      findings:
        - id: BR-5
          severity: Critical
          title: Host visibility can authorize assets using a different repository host
          detail: 'cmd/sdlc/assetpolicy.go:55 strips github.com before cmd/sdlc/ghclient.go:187 invokes gh with bare owner/repo and inherited GH_HOST. A temporary regression showed PRIVATE visibility from an enterprise host authorizing a public GitHub destination; the installed gh independently confirmed this routing in a loopback probe. This is the 2nd finding in family destination-policy-authority. ARCH-SECURE / ARCH-PURPOSE: enforce the rule that authorization retains the complete destination identity through every seam; preserve the hostname and cover conflicting GH_HOST values across admission callers.'
          family: destination-policy-authority
          round: 2
        - id: BR-6
          severity: Critical
          title: Legacy publication skips missing assets introduced only through the body
          detail: 'cmd/sdlc/transferguard.go:254–259 returns success when asset directories have no branch history. A temporary production regression added a Markdown link to a missing canonical asset without creating an asset directory; guardTransferredDetails accepted it. ARCH-PURPOSE: derive validation requirements from the authoritative body index as well as changed asset paths, and test body-only references through legacy publication.'
          family: body-indexed-publication-validation
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - "n": 3
      timestamp: "2026-10-13T17:02:49-07:00"
      dispatched: "2026-10-13T16:55:46-07:00"
      agent: codex
      dispose:
        - id: BR-2
          disposition: addressed
          note: Source/destination preflight and post-materialization proofs are present. Reopen and restore regressions pass; removing restore's final validation makes both byte-tampering and attribute-tampering cases commit successfully and fail their tests.
          round: 3
        - id: BR-5
          disposition: addressed
          note: assetpolicy.go:57 retains the full GitHub URL and realGH rejects incomplete identities. Conflicting-GH_HOST tests pass across admission paths; restoring hostname stripping and the old permissive seam reproduces unauthorized admission.
          round: 3
        - id: BR-6
          disposition: addressed
          note: Legacy publication now validates changed body indexes without requiring asset-directory history. Restoring the historical directory-history shortcut makes the legacy guard regressions fail. A distinct persisted-transfer recovery gap is raised below.
          round: 3
        - id: BR-1
          disposition: addressed
          note: 'Retained: publication intersects local restrictions with pinned destination policy, with destination-race and recovery regression coverage.'
          round: 3
        - id: BR-3
          disposition: addressed
          note: 'Retained: the Core concepts table identifies planLandingArchive as pure and confirmLandingArchive as integration, matching their implementations.'
          round: 3
        - id: BR-4
          disposition: addressed
          note: 'Retained: TestLiveAssetVisibilityConformance calls the production adapter; atlas/workflow/issue-assets.md documents its invocation and cadence.'
          round: 3
      findings:
        - id: BR-7
          severity: Critical
          title: Prepared transfer recovery can publish missing body-indexed assets
          detail: 'cmd/sdlc/issuemovedetail.go:409 returns success for empty SourceAssets, while internal/tracker/transferop.go:130 publishes resumed candidates without rerunning Prepare''s bundle validation. A temporary regression persisted an older prepared transfer containing a canonical missing-asset link, interrupted before main publication, then invoked runRecoveryReconcile: it returned nil and advanced remote main. This is the 2nd finding in this family. Do not fix only this instance: enforce body-index validation before every fresh or resumed publication, including prepared and uncertain receipts regardless of asset-pin presence, and sweep those states with regression tests. ARCH-PURPOSE / ARCH-SECURE / ARCH-ORDER.'
          family: body-indexed-publication-validation
          round: 3
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - "n": 4
      timestamp: "2026-10-13T18:00:05-07:00"
      dispatched: "2026-10-13T17:56:13-07:00"
      agent: codex
      dispose:
        - id: BR-7
          disposition: addressed
          note: Shared checkedMainPublisher validation covers creation and transfer apply/probe plus confirmed transfer continuation. The production recovery test passed; disabling the guard made prepared, uncertain, landed, landed-uncertain, and confirmed cases fail.
          round: 4
        - id: BR-1
          disposition: addressed
          note: 'Retained: publication admission checks the pinned destination policy on each attempt, intersecting local restrictions.'
          round: 4
        - id: BR-2
          disposition: addressed
          note: 'Retained: materialization callers check attributes and compare actual asset bytes with pinned snapshots before local commits.'
          round: 4
        - id: BR-3
          disposition: not-addressed
          note: 'ARCH-PURE: the proof classification was corrected, but the earlier acceptance of the planner as PURE was mistaken. The plan at line 82 still labels planLandingArchive pure; landingarchive.go:444 calls trackedArchiveBytes, whose line 55 invokes the production ReadCardBlob callback bound at line 728. ReadCardBlob obtains a Git snapshot at internal/tracker/candidates.go:192. Reclassify the planner as INTEGRATION and correct the comment/Revisions claim, or preload card bytes outside it. Retain Critical under the explicit core-concepts contradiction rule; this remains family core-concept-classification.'
          round: 4
        - id: BR-4
          disposition: addressed
          note: 'Retained: the opt-in live visibility test uses the production adapter, and atlas documents its invocation and cadence.'
          round: 4
        - id: BR-5
          disposition: addressed
          note: 'Retained: visibility authorization preserves the explicit destination host through the gh seam.'
          round: 4
        - id: BR-6
          disposition: addressed
          note: 'Retained: publication validation selects affected issue bodies as well as asset paths, covering missing body-indexed members.'
          round: 4
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#337 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T15:04:03-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `destination-policy-authority` Publication trusts stale checkout policy instead of current destination policy
  cmd/sdlc/assetpolicy.go:23 reads only the working-tree policy; republish.go and assetbundlegit.go reuse it during publication. A temporary regression tightened main to text-only while leaving the caller allowed, and publication still pushed new asset bytes. ARCH-SECURE / ARCH-PURPOSE: resolve authorization from the pinned destination snapshot on every publication attempt and sweep all admission callers, including first-publication recovery, legacy publication, and abandonment.
- **BR-2** [Critical] `exact-bytes-at-materialization` Reopening can commit converted asset bytes under their original content name
  cmd/sdlc/abandon.go:464 merges and materializes archived assets before line 504 checks only active destinations. A temporary regression using archive-only text eol=crlf and active -text reopened successfully but changed LF evidence to CRLF under its unchanged hash filename. ARCH-PURPOSE: validate all materialization paths before effects and verify restored bytes against the pinned bundle; add a production reopen regression.
- **BR-3** [Critical] `core-concept-classification` The core-concepts table incorrectly declares archive proof PURE
  workshop/plans/000337-issue-assets-plan.md:82 places landing archive plan and proof in the Pure entity table, while cmd/sdlc/landingarchive.go:556 confirmLandingArchive executes Git, loads snapshots, and reads time.Now; its proof tests use temporary repositories. ARCH-PURE: retain the planner as PURE, classify the proof as INTEGRATION, and append a Revisions entry correcting the table.
- **BR-4** [Important] `external-dependency-conformance` Host visibility has no live conformance check
  cmd/sdlc/assetvisibility_test.go:11 supplies canned JSON through a replacement gh executable, and assetpolicy_test.go models visibility with a state file; neither verifies the real RepoVisibility invocation against the host. ARCH-MOCK: add an opt-in read-only live conformance check through the production seam and document when it runs.

## Round 2 — 2026-10-13T16:02:44-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — Pinned destination policy now reaches publication and recovery callers. Existing regressions pass; disabling the destination check makes republish, first-publication, legacy and abandonment cases fail.
- BR-2 — not-addressed — Reopen is corrected, but the materialization-class sweep remains incomplete. cmd/sdlc/issuerestore.go:127 checks out assets, then line 136 commits working-tree paths without verifying pinned bytes. A temporary production regression installed a post-checkout hook; restore returned success and committed "hook-tampered" under the original hash filename. ARCH-PURPOSE / ARCH-SECURE: apply exact-byte verification across every materialization caller before committing.
- BR-3 — addressed — workshop/plans/000337-issue-assets-plan.md:82 now classifies only planLandingArchive as PURE; line 95 classifies confirmLandingArchive as INTEGRATION. The Revisions correction matches the planner and Git/clock-backed proof implementations.
- BR-4 — addressed — TestLiveAssetVisibilityConformance calls the production adapter and fails opted-in prerequisite errors. Atlas documents invocation and cadence. Independently queried visibility was public; the live production test passed against that expectation.

### Raised

- **BR-5** [Critical] `destination-policy-authority` Host visibility can authorize assets using a different repository host
  cmd/sdlc/assetpolicy.go:55 strips github.com before cmd/sdlc/ghclient.go:187 invokes gh with bare owner/repo and inherited GH_HOST. A temporary regression showed PRIVATE visibility from an enterprise host authorizing a public GitHub destination; the installed gh independently confirmed this routing in a loopback probe. This is the 2nd finding in family destination-policy-authority. ARCH-SECURE / ARCH-PURPOSE: enforce the rule that authorization retains the complete destination identity through every seam; preserve the hostname and cover conflicting GH_HOST values across admission callers.
- **BR-6** [Critical] `body-indexed-publication-validation` Legacy publication skips missing assets introduced only through the body
  cmd/sdlc/transferguard.go:254–259 returns success when asset directories have no branch history. A temporary production regression added a Markdown link to a missing canonical asset without creating an asset directory; guardTransferredDetails accepted it. ARCH-PURPOSE: derive validation requirements from the authoritative body index as well as changed asset paths, and test body-only references through legacy publication.

## Round 3 — 2026-10-13T17:02:49-07:00 (codex) — BLOCKED

### Disposed

- BR-2 — addressed — Source/destination preflight and post-materialization proofs are present. Reopen and restore regressions pass; removing restore's final validation makes both byte-tampering and attribute-tampering cases commit successfully and fail their tests.
- BR-5 — addressed — assetpolicy.go:57 retains the full GitHub URL and realGH rejects incomplete identities. Conflicting-GH_HOST tests pass across admission paths; restoring hostname stripping and the old permissive seam reproduces unauthorized admission.
- BR-6 — addressed — Legacy publication now validates changed body indexes without requiring asset-directory history. Restoring the historical directory-history shortcut makes the legacy guard regressions fail. A distinct persisted-transfer recovery gap is raised below.
- BR-1 — addressed — Retained: publication intersects local restrictions with pinned destination policy, with destination-race and recovery regression coverage.
- BR-3 — addressed — Retained: the Core concepts table identifies planLandingArchive as pure and confirmLandingArchive as integration, matching their implementations.
- BR-4 — addressed — Retained: TestLiveAssetVisibilityConformance calls the production adapter; atlas/workflow/issue-assets.md documents its invocation and cadence.

### Raised

- **BR-7** [Critical] `body-indexed-publication-validation` Prepared transfer recovery can publish missing body-indexed assets
  cmd/sdlc/issuemovedetail.go:409 returns success for empty SourceAssets, while internal/tracker/transferop.go:130 publishes resumed candidates without rerunning Prepare's bundle validation. A temporary regression persisted an older prepared transfer containing a canonical missing-asset link, interrupted before main publication, then invoked runRecoveryReconcile: it returned nil and advanced remote main. This is the 2nd finding in this family. Do not fix only this instance: enforce body-index validation before every fresh or resumed publication, including prepared and uncertain receipts regardless of asset-pin presence, and sweep those states with regression tests. ARCH-PURPOSE / ARCH-SECURE / ARCH-ORDER.

## Round 4 — 2026-10-13T18:00:05-07:00 (codex) — BLOCKED

### Disposed

- BR-7 — addressed — Shared checkedMainPublisher validation covers creation and transfer apply/probe plus confirmed transfer continuation. The production recovery test passed; disabling the guard made prepared, uncertain, landed, landed-uncertain, and confirmed cases fail.
- BR-1 — addressed — Retained: publication admission checks the pinned destination policy on each attempt, intersecting local restrictions.
- BR-2 — addressed — Retained: materialization callers check attributes and compare actual asset bytes with pinned snapshots before local commits.
- BR-3 — not-addressed — ARCH-PURE: the proof classification was corrected, but the earlier acceptance of the planner as PURE was mistaken. The plan at line 82 still labels planLandingArchive pure; landingarchive.go:444 calls trackedArchiveBytes, whose line 55 invokes the production ReadCardBlob callback bound at line 728. ReadCardBlob obtains a Git snapshot at internal/tracker/candidates.go:192. Reclassify the planner as INTEGRATION and correct the comment/Revisions claim, or preload card bytes outside it. Retain Critical under the explicit core-concepts contradiction rule; this remains family core-concept-classification.
- BR-4 — addressed — Retained: the opt-in live visibility test uses the production adapter, and atlas documents its invocation and cadence.
- BR-5 — addressed — Retained: visibility authorization preserves the explicit destination host through the gh seam.
- BR-6 — addressed — Retained: publication validation selects affected issue bodies as well as asset paths, covering missing body-indexed members.

## Open findings

- **BR-3** [Critical] `core-concept-classification` The core-concepts table incorrectly declares archive proof PURE
