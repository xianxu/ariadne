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

---

## Re-review — 2026-10-13T17:02:48-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 337 — issue assets: a per-issue folder for images and other non-text files, with the issue's lifecycle |
| repo | ariadne |
| issue file | workshop/issues/000337-issue-assets-a-per-issue-folder-for-images-and-other-non-text-files-with-the-issue-s-lifecycle.md |
| boundary | whole-issue close |
| milestone | — |
| window | da2861cd8c8033f21726a6040de1dec6b12279bf..04376ef7082d1c615e3ce23ad4b7f52d3fcfa29f |
| command | sdlc close --issue 337 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-13T17:02:48-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The three open findings have working corrections backed by regressions that fail when the fixes are removed. However, an older prepared transfer can still publish missing body-indexed assets through recovery. This independently reproduced correctness gap blocks shipping. The reviewed checkout is unchanged; the disposable mutation-test copy was removed.

1. **Strengths**

   - Shared Markdown parsing and bundle validation enforce content names, membership and byte budgets.
   - Materialization checks now cover source/destination attributes and actual restored bytes.
   - Visibility queries preserve the destination hostname through the production adapter.
   - README and atlas document the command, policy, lifecycle, limits and recovery.

2. **Critical findings**

   **Prepared transfer recovery bypasses body-index validation** — `cmd/sdlc/issuemovedetail.go:409`, `cmd/sdlc/internal/tracker/transferop.go:130`.

   Recovery can resume an already-prepared `transfer.main` candidate without executing `TransferOp.Prepare`, where bundle validation lives. The publication callback returns immediately when `SourceAssets` is empty.

   A temporary regression constructed a valid persisted transfer using the previous text-only candidate preparation, interrupted before its main push, then invoked production `runRecoveryReconcile`. Its body linked a missing canonical asset. Recovery returned success and advanced remote main.

   **This is the 2nd finding in family `body-indexed-publication-validation`.** Fix the rule across fresh and resumed operations: validate the pinned body and complete candidate bundle before publication, regardless of whether asset pins exist. Include prepared and uncertain receipt states; do not fix only this early return. **ARCH-PURPOSE / ARCH-SECURE / ARCH-ORDER.**

3. **Important findings**

   None additional.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Focused command regressions passed: **26.783s**.
   - Pure asset package passed: **0.087s**.
   - Targeted tracker and Git snapshot tests passed: **10.691s / 5.149s**.
   - Removing the fixes reproduced BR-2, BR-5 and BR-6 failures.
   - New recovery regression failed as expected: `err=<nil> main_changed=true`.
   - Broader asset/lifecycle run timed out after **240.318s**, during temporary-directory cleanup; it is **not a pass**.
   - `git diff --check` passed. Full acceptance and live conformance were not rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared bundle, policy and materialization helpers.
   - **ARCH-PURE — pass:** declared pure entities exist and have IO-free core tests; archive proof is correctly classified as integration.
   - **ARCH-PURPOSE — flag:** recovery remains an unvalidated publication consumer.
   - **ARCH-MOCK — pass:** portable Git fixtures, host doubles and opt-in production-adapter conformance.
   - **ARCH-CONSTRAINTS — pass:** explicit bundle limits, bounded visibility responses and metadata-first discovery.
   - **ARCH-SECURE — flag:** an older persisted candidate inherits validation assumptions from its producer.
   - **ARCH-ORDER — flag:** prepared-candidate recovery skips the new invariant; retain a regression covering that restart sequence.
   - **ARCH-FUNERAL — pass:** asset removal/archive paths, temporary-file recovery and retained Git-history costs are documented.

7. **Plan revision recommendations**

   Add a `## Revisions` entry extending the body-index validation inventory to persisted transfer candidates. Require equivalent validation for fresh preparation, prepared-candidate replay and uncertain-outcome recovery, including receipts created before asset validation existed.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      Source/destination preflight and post-materialization proofs are present. Reopen and restore regressions pass; removing restore's final validation makes both byte-tampering and attribute-tampering cases commit successfully and fail their tests.
  - id: BR-5
    disposition: addressed
    note: |
      assetpolicy.go:57 retains the full GitHub URL and realGH rejects incomplete identities. Conflicting-GH_HOST tests pass across admission paths; restoring hostname stripping and the old permissive seam reproduces unauthorized admission.
  - id: BR-6
    disposition: addressed
    note: |
      Legacy publication now validates changed body indexes without requiring asset-directory history. Restoring the historical directory-history shortcut makes the legacy guard regressions fail. A distinct persisted-transfer recovery gap is raised below.
  - id: BR-1
    disposition: addressed
    note: |
      Retained: publication intersects local restrictions with pinned destination policy, with destination-race and recovery regression coverage.
  - id: BR-3
    disposition: addressed
    note: |
      Retained: the Core concepts table identifies planLandingArchive as pure and confirmLandingArchive as integration, matching their implementations.
  - id: BR-4
    disposition: addressed
    note: |
      Retained: TestLiveAssetVisibilityConformance calls the production adapter; atlas/workflow/issue-assets.md documents its invocation and cadence.
findings:
  - id: new
    severity: Critical
    family: body-indexed-publication-validation
    title: |
      Prepared transfer recovery can publish missing body-indexed assets
    detail: |
      cmd/sdlc/issuemovedetail.go:409 returns success for empty SourceAssets, while internal/tracker/transferop.go:130 publishes resumed candidates without rerunning Prepare's bundle validation. A temporary regression persisted an older prepared transfer containing a canonical missing-asset link, interrupted before main publication, then invoked runRecoveryReconcile: it returned nil and advanced remote main. This is the 2nd finding in this family. Do not fix only this instance: enforce body-index validation before every fresh or resumed publication, including prepared and uncertain receipts regardless of asset-pin presence, and sweep those states with regression tests. ARCH-PURPOSE / ARCH-SECURE / ARCH-ORDER.
```

---

## Re-review — 2026-10-13T18:00:04-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 337 — issue assets: a per-issue folder for images and other non-text files, with the issue's lifecycle |
| repo | ariadne |
| issue file | workshop/issues/000337-issue-assets-a-per-issue-folder-for-images-and-other-non-text-files-with-the-issue-s-lifecycle.md |
| boundary | whole-issue close |
| milestone | — |
| window | 84d39c1513a4a30f0f61b4d7e5f610ff97b3b9ce..2cd38ae505ffe4413ba9a04225f1e1dfbdd46b6f |
| command | sdlc close --issue 337 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-13T18:00:04-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-7 is addressed: production recovery passes, and disabling the shared validation makes all five recovery-state regressions fail. One classification correction remains: BR-3’s earlier acceptance incorrectly treated the archive planner as pure. The required core-concepts cross-check therefore blocks this boundary. No new executable defect was confirmed.

1. **Strengths**

   - Persisted creation and transfer candidates share body-and-bundle validation before publication or confirmation.
   - Recovery tests cover prepared, uncertain, landed, landed-uncertain, and confirmed transfers, including receipts without asset pins.
   - Attachment uses an opaque transition model, bounded content-addressed bundles, and ownership-aware cleanup.
   - README and atlas document the command, policy, lifecycle, limits, and legacy-receipt recovery restriction.

2. **Critical findings**

   **BR-3 — not-addressed; ARCH-PURE.** [The plan’s PURE table](workshop/plans/000337-issue-assets-plan.md:82) still includes `planLandingArchive`. Its tracked branch calls `trackedArchiveBytes` at `cmd/sdlc/landingarchive.go:444`, which invokes `owned.readCard` at line 55. Production binds that callback to `env.repo.ReadCardBlob` at line 728; `internal/tracker/candidates.go:192` obtains a Git snapshot before reading the blob.

   The earlier recommendation to retain this planner as PURE was mistaken. Classify it as INTEGRATION and correct the corresponding comment and revision statement. Alternatively, preload the required card bytes outside the planner. This is the existing `core-concept-classification` finding, not a new finding. Severity follows the checklist’s explicit rule for table/code contradictions.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Targeted pure-core and tracker tests passed.
   - Production BR-7 recovery test passed: **55.246s**.
   - Attachment, publication, visibility, and archive selection passed: **42.431s**.
   - Disabling the shared guard caused tracker regressions and **all five production recovery states** to fail.
   - Independently reproduced the reported startup failure with `MAKEFLAGS=k`; it passes with `MAKEFLAGS` unset. That package is unchanged in the range.
   - Full suite was not independently rerun. Pinned-range whitespace check passed; temporary overlay removed; checkout clean.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared bundle and replay-validation helpers.
   - **ARCH-PURE — flag:** BR-3 classification contradiction above.
   - **ARCH-PURPOSE — pass:** replay correction covers both persisted main publishers and subsequent transfer completion.
   - **ARCH-MOCK — pass:** portable Git fixtures, stateful visibility fake, and opt-in conformance check.
   - **ARCH-CONSTRAINTS — pass:** explicit byte/count limits and bounded visibility queries.
   - **ARCH-SECURE — pass:** persisted candidates are validated against pinned source bytes.
   - **ARCH-ORDER — pass:** uncertain publication and confirmed-stage recovery receive validation.
   - **ARCH-FUNERAL — pass:** cleanup, archive/removal behavior, and retained Git-history costs are documented.

7. **Plan revision recommendation**

   Add a `## Revisions` entry correcting the earlier claim that `planLandingArchive` is pure, citing its production card-reader dependency. A classification-only repair needs source inspection, not a wording-presence test.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      Shared checkedMainPublisher validation covers creation and transfer apply/probe plus confirmed transfer continuation. The production recovery test passed; disabling the guard made prepared, uncertain, landed, landed-uncertain, and confirmed cases fail.
  - id: BR-1
    disposition: addressed
    note: |
      Retained: publication admission checks the pinned destination policy on each attempt, intersecting local restrictions.
  - id: BR-2
    disposition: addressed
    note: |
      Retained: materialization callers check attributes and compare actual asset bytes with pinned snapshots before local commits.
  - id: BR-3
    disposition: not-addressed
    note: |
      ARCH-PURE: the proof classification was corrected, but the earlier acceptance of the planner as PURE was mistaken. The plan at line 82 still labels planLandingArchive pure; landingarchive.go:444 calls trackedArchiveBytes, whose line 55 invokes the production ReadCardBlob callback bound at line 728. ReadCardBlob obtains a Git snapshot at internal/tracker/candidates.go:192. Reclassify the planner as INTEGRATION and correct the comment/Revisions claim, or preload card bytes outside it. Retain Critical under the explicit core-concepts contradiction rule; this remains family core-concept-classification.
  - id: BR-4
    disposition: addressed
    note: |
      Retained: the opt-in live visibility test uses the production adapter, and atlas documents its invocation and cadence.
  - id: BR-5
    disposition: addressed
    note: |
      Retained: visibility authorization preserves the explicit destination host through the gh seam.
  - id: BR-6
    disposition: addressed
    note: |
      Retained: publication validation selects affected issue bodies as well as asset paths, covering missing body-indexed members.
```

---

## Re-review — 2026-10-13T18:13:45-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 337 — issue assets: a per-issue folder for images and other non-text files, with the issue's lifecycle |
| repo | ariadne |
| issue file | workshop/issues/000337-issue-assets-a-per-issue-folder-for-images-and-other-non-text-files-with-the-issue-s-lifecycle.md |
| boundary | whole-issue close |
| milestone | — |
| window | 84d39c1513a4a30f0f61b4d7e5f610ff97b3b9ce..2bd62b3fa0a643f65e0d2123a838aaf83f1367bf |
| command | sdlc close --issue 337 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-13T18:13:45-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: medium
```

BR-3 is addressed: the plan now correctly classifies both archive planner and proof as INTEGRATION, matching their production dependencies. Inspection of the pinned range found no new blocking defects or documentation gaps. Test confidence is qualified by the CLI batch timeout described below. The checkout remains unchanged.

1. **Strengths**
   - Shared bundle validation enforces body-indexed membership, content hashes and size limits.
   - Publication validates destination policy and persisted candidates; materialization checks compare actual bytes against pinned snapshots.
   - Attachment recovery uses explicit transitions and distinguishes newly installed bytes from pre-existing evidence.
   - README, atlas, help and the recovery catalog document the new surface.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**
   - Passed: full `internal/issueassets`, `internal/tracker` and `internal/gitx` packages.
   - The CLI asset/lifecycle batch timed out after 480 seconds during reopen fixture setup, with no preceding assertion failures reported.
   - The interrupted completed-issue reopen test and remaining stage/clear test passed separately in 111.331 seconds.
   - `git diff --check` passed.
   - Full-module tests and live visibility conformance were not rerun. BR-3 changes only classification and comments; source tracing supplies the appropriate evidence.

6. **Architectural notes**
   - **ARCH-DRY — pass:** shared bundle, publication and materialization helpers.
   - **ARCH-PURE — pass:** pure asset core remains IO-free; archive classifications now reflect reachable IO.
   - **ARCH-PURPOSE — pass:** reviewed consumers cover the approved in-repository lifecycle and legacy recovery.
   - **ARCH-MOCK — pass:** stateful visibility fake, portable Git fixtures and live-conformance entry point.
   - **ARCH-CONSTRAINTS — pass:** bounded files, bundles, queries and metadata-selected reads.
   - **ARCH-SECURE — pass:** strict parsing, explicit host identity, destination-policy checks and byte validation.
   - **ARCH-ORDER — pass:** explicit attachment transitions and persisted-candidate validation preserve uncertainty and partial progress.
   - **ARCH-FUNERAL — pass:** cleanup, retained-file recovery, archive/removal behavior and Git-history costs are documented.

7. **Plan revision recommendations:** None. The correction at `workshop/plans/000337-issue-assets-plan.md:388` explicitly supersedes the earlier pure-planner claim.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      Plan lines 94–95 classify planner and proof as INTEGRATION; line 388 supersedes the earlier claim. landingarchive.go:399/445 reaches trackedArchiveBytes:55, whose readCard callback is bound to ReadCardBlob at line 729; internal/tracker/candidates.go:192 obtains a Git snapshot. The pinned correction changes classification/comments only.
  - id: BR-1
    disposition: addressed
    note: |
      Retained: publication intersects local restrictions with pinned destination policy and rechecks on retries; destination-policy regressions remain present.
  - id: BR-2
    disposition: addressed
    note: |
      Retained: materialization paths check conversion attributes and verify pinned bytes before commits; isolated completed-issue reopen and stage/clear regressions passed.
  - id: BR-4
    disposition: addressed
    note: |
      Retained: the opt-in live visibility test exercises the production adapter, with invocation and change-triggered cadence documented in atlas.
  - id: BR-5
    disposition: addressed
    note: |
      Retained: authorization preserves the explicit destination host through the visibility seam; conflicting-host regression coverage remains present.
  - id: BR-6
    disposition: addressed
    note: |
      Retained: publication selects affected issue bodies as well as asset paths; missing body-indexed member regressions remain present.
  - id: BR-7
    disposition: addressed
    note: |
      Retained: shared candidate validation covers creation and transfer apply/probe plus confirmed transfer continuation; the full tracker test package passed.
```
