# Boundary Review — ariadne#337 (whole-issue close)

| field | value |
|-------|-------|
| issue | 337 — issue assets: a per-issue folder for images and other non-text files, with the issue's lifecycle |
| repo | ariadne |
| issue file | workshop/issues/000337-issue-assets-a-per-issue-folder-for-images-and-other-non-text-files-with-the-issue-s-lifecycle.md |
| boundary | whole-issue close |
| milestone | — |
| window | 650e3075df9e7cd5a1780e9e53a2c0949a1f9c6f..c0712e618c6a29c9b1f702e7300cef135042d931 |
| command | sdlc close --issue 337 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-13T15:04:03-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The shared bundle model, attachment recovery, and documentation are substantial and well tested. Two reproduced correctness bugs block shipping: stale local policy permits publication after main becomes text-only, and reopening can change archived asset bytes through Git attributes. The plan also misclassifies the archive proof, and visibility lacks live conformance coverage.

1. **Strengths**

   - `internal/issueassets` centralizes Markdown parsing, content naming, budgets, and bundle deltas.
   - The opaque attachment state and injected filesystem boundary cover ownership-aware rollback and uncertain body-write outcomes.
   - `FilesSelected` filters metadata before reading blobs; its aggregate-volume regression passed.
   - README, atlas, command help, and recovery documentation cover the new surface and explicit migration exclusion.

2. **Critical findings**

   - **Stale publication authorization:** `cmd/sdlc/assetpolicy.go:23` reads working-tree policy. Publication retries therefore ignore a newer destination policy. Resolve admission against the pinned destination snapshot on every attempt, retaining appropriate local restrictions.
   - **Reopen corrupts evidence:** `cmd/sdlc/abandon.go:464` materializes archived files before the destination-only attribute check at line 504. Archive-specific conversion survives the subsequent move into a nonconverting active path. Check every materialization path before effects and verify restored bytes against the pinned bundle.
   - **Core-table contradiction:** `workshop/plans/000337-issue-assets-plan.md:82` classifies archive proof as PURE, but `confirmLandingArchive` invokes Git and the clock. Split the pure planner from the integration proof in the table.

3. **Important findings**

   - **Missing visibility conformance:** `cmd/sdlc/assetvisibility_test.go:11` tests fabricated responses but does not validate the real `gh` invocation or compare the stateful fake with live behavior. Add an opt-in read-only conformance test and document its cadence.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Independently passed:

   - `go test ./cmd/sdlc/internal/issueassets -count=1`
   - Focused `TestFilesSelected` tests.
   - `go test ./cmd/sdlc -run 'Test(Asset|IssueAttach)' -count=1 -timeout 180s`
   - Pinned-range `git diff --check`.

   Temporary test overlays reproduced both behavioral defects:

   - Tightening only main’s policy to `text-only` still published new bytes from a checkout retaining `allowed`.
   - Reopening with archive-only `text eol=crlf` changed `"line one\nline two"` to `"line one\r\nline two"` under the unchanged content-hash filename.

   The overlays were removed; the checkout remains clean. The full suite was not independently rerun. No prior findings required disposition.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared parsing, bundle validation, and attribute helpers.
   - **ARCH-PURE — flag:** archive proof’s declared classification contradicts its implementation.
   - **ARCH-PURPOSE — flag:** privacy enforcement and exact lifecycle preservation remain incomplete.
   - **ARCH-MOCK — flag:** stateful visibility fake exists; live conformance coverage does not.
   - **ARCH-CONSTRAINTS — pass:** explicit byte/count limits, bounded visibility responses, and metadata-first discovery.
   - **ARCH-SECURE — flag:** stale checkout policy becomes authority for destination publication.
   - **ARCH-ORDER — pass:** attachment transitions enforce ownership and reconcile uncertain outcomes.
   - **ARCH-FUNERAL — pass:** temporary cleanup, bundle removal/archive paths, and retained Git-history costs are documented.

7. **Plan revision recommendations**

   Add `## Revisions` entries that:

   - Define destination-snapshot policy authority and enumerate republish, first publication/recovery, legacy publication, and abandon admission checks.
   - Extend attribute validation to every source and destination materialization before mutation.
   - Classify `planLandingArchive` as PURE and `confirmLandingArchive` as INTEGRATION; reconcile table names and paths with implemented entities.
   - Name the visibility conformance test and execution cadence.

```findings
findings:
  - id: new
    severity: Critical
    family: destination-policy-authority
    title: |
      Publication trusts stale checkout policy instead of current destination policy
    detail: |
      cmd/sdlc/assetpolicy.go:23 reads only the working-tree policy; republish.go and assetbundlegit.go reuse it during publication. A temporary regression tightened main to text-only while leaving the caller allowed, and publication still pushed new asset bytes. ARCH-SECURE / ARCH-PURPOSE: resolve authorization from the pinned destination snapshot on every publication attempt and sweep all admission callers, including first-publication recovery, legacy publication, and abandonment.
  - id: new
    severity: Critical
    family: exact-bytes-at-materialization
    title: |
      Reopening can commit converted asset bytes under their original content name
    detail: |
      cmd/sdlc/abandon.go:464 merges and materializes archived assets before line 504 checks only active destinations. A temporary regression using archive-only text eol=crlf and active -text reopened successfully but changed LF evidence to CRLF under its unchanged hash filename. ARCH-PURPOSE: validate all materialization paths before effects and verify restored bytes against the pinned bundle; add a production reopen regression.
  - id: new
    severity: Critical
    family: core-concept-classification
    title: |
      The core-concepts table incorrectly declares archive proof PURE
    detail: |
      workshop/plans/000337-issue-assets-plan.md:82 places landing archive plan and proof in the Pure entity table, while cmd/sdlc/landingarchive.go:556 confirmLandingArchive executes Git, loads snapshots, and reads time.Now; its proof tests use temporary repositories. ARCH-PURE: retain the planner as PURE, classify the proof as INTEGRATION, and append a Revisions entry correcting the table.
  - id: new
    severity: Important
    family: external-dependency-conformance
    title: |
      Host visibility has no live conformance check
    detail: |
      cmd/sdlc/assetvisibility_test.go:11 supplies canned JSON through a replacement gh executable, and assetpolicy_test.go models visibility with a state file; neither verifies the real RepoVisibility invocation against the host. ARCH-MOCK: add an opt-in read-only live conformance check through the production seam and document when it runs.
```

---

## Re-review — 2026-10-13T16:02:44-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 337 — issue assets: a per-issue folder for images and other non-text files, with the issue's lifecycle |
| repo | ariadne |
| issue file | workshop/issues/000337-issue-assets-a-per-issue-folder-for-images-and-other-non-text-files-with-the-issue-s-lifecycle.md |
| boundary | whole-issue close |
| milestone | — |
| window | f2266a7eea4af6b71969c7b6c02363ed815689c8..64b20a3dc228170515bbddbf087da03fe0e5064a |
| command | sdlc close --issue 337 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-13T16:02:44-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The destination-policy and reopen fixes have meaningful regression coverage, and the documentation gates are satisfied. Three reproducible defects still block shipping: restore can commit corrupted asset bytes, visibility can come from the wrong host, and legacy publication can skip validation of body-only asset references.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Pinned destination policy now reaches publication and recovery callers. Existing regressions pass; disabling the destination check makes republish, first-publication, legacy and abandonment cases fail.
  - id: BR-2
    disposition: not-addressed
    note: |
      Reopen is corrected, but the materialization-class sweep remains incomplete. cmd/sdlc/issuerestore.go:127 checks out assets, then line 136 commits working-tree paths without verifying pinned bytes. A temporary production regression installed a post-checkout hook; restore returned success and committed "hook-tampered" under the original hash filename. ARCH-PURPOSE / ARCH-SECURE: apply exact-byte verification across every materialization caller before committing.
  - id: BR-3
    disposition: addressed
    note: |
      workshop/plans/000337-issue-assets-plan.md:82 now classifies only planLandingArchive as PURE; line 95 classifies confirmLandingArchive as INTEGRATION. The Revisions correction matches the planner and Git/clock-backed proof implementations.
  - id: BR-4
    disposition: addressed
    note: |
      TestLiveAssetVisibilityConformance calls the production adapter and fails opted-in prerequisite errors. Atlas documents invocation and cadence. Independently queried visibility was public; the live production test passed against that expectation.
findings:
  - id: new
    severity: Critical
    family: destination-policy-authority
    title: |
      Host visibility can authorize assets using a different repository host
    detail: |
      cmd/sdlc/assetpolicy.go:55 strips github.com before cmd/sdlc/ghclient.go:187 invokes gh with bare owner/repo and inherited GH_HOST. A temporary regression showed PRIVATE visibility from an enterprise host authorizing a public GitHub destination; the installed gh independently confirmed this routing in a loopback probe. This is the 2nd finding in family destination-policy-authority. ARCH-SECURE / ARCH-PURPOSE: enforce the rule that authorization retains the complete destination identity through every seam; preserve the hostname and cover conflicting GH_HOST values across admission callers.
  - id: new
    severity: Critical
    family: body-indexed-publication-validation
    title: |
      Legacy publication skips missing assets introduced only through the body
    detail: |
      cmd/sdlc/transferguard.go:254–259 returns success when asset directories have no branch history. A temporary production regression added a Markdown link to a missing canonical asset without creating an asset directory; guardTransferredDetails accepted it. ARCH-PURPOSE: derive validation requirements from the authoritative body index as well as changed asset paths, and test body-only references through legacy publication.
```

1. **Strengths**

   - Destination-policy checks use pinned publication bases, including retries and recovery.
   - The shared bundle model validates canonical names, exact bytes, membership and budgets.
   - Attachment transitions are encapsulated in the pure core; README and atlas explain the new surface, recovery and history costs.

2. **Critical findings**

   - **BR-2 remains open:** verify restored bytes before `issue restore` commits them.
   - **Wrong-host authorization:** preserve the fully qualified destination through the visibility seam.
   - **Legacy validation bypass:** inspect body references even when no asset directory has existed.

3. **Important findings:** None additional.

4. **Minor findings:** None.

5. **Test coverage notes**

   Focused command regressions passed (145.426s), selected internal-package tests passed, and live visibility conformance passed. Disabling the prior fixes made their regression tests fail. Three additional temporary regressions reproduced the remaining defects. All disposable checkouts were removed; the reviewed checkout remains unchanged. I did not rerun the full suite.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared bundle and attribute helpers.
   - **ARCH-PURE — pass:** corrected classification and IO-free core.
   - **ARCH-PURPOSE — flag:** restore and legacy callers leave lifecycle contracts incomplete.
   - **ARCH-MOCK — pass:** stateful visibility fake plus executable live conformance.
   - **ARCH-CONSTRAINTS — pass:** explicit byte/count limits and bounded visibility queries.
   - **ARCH-SECURE — flag:** incomplete host identity and unchecked materialized bytes.
   - **ARCH-ORDER — pass:** encapsulated attachment transitions and uncertain-outcome reconciliation.
   - **ARCH-FUNERAL — pass:** asset removal/archive behavior, temporary-file recovery and retained-history costs are documented.

7. **Plan revision recommendations**

   Append `## Revisions` entries establishing complete destination identity as an authorization invariant, enumerating every materialization caller requiring exact-byte verification, and requiring body-driven validation on legacy publication paths. Add the three demonstrated regressions before repeating this boundary review.
