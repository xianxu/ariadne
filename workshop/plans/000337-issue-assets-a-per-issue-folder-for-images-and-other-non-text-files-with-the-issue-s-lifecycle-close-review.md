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
