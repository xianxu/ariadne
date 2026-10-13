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

## Open findings

- **BR-2** [Critical] `exact-bytes-at-materialization` Reopening can commit converted asset bytes under their original content name
- **BR-5** [Critical] `destination-policy-authority` Host visibility can authorize assets using a different repository host
- **BR-6** [Critical] `body-indexed-publication-validation` Legacy publication skips missing assets introduced only through the body
