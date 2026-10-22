# Boundary Review — ariadne#544 (whole-issue close)

| field | value |
|-------|-------|
| issue | 544 — issue migrate drops a legacy close's completion record; later issues' actual measurement then refuses (echo#1 → echo#3) |
| repo | ariadne |
| issue file | workshop/issues/000544-issue-migrate-drops-a-legacy-close-s-completion-record-later-issues-actual-measurement-then-refuses-echo-1-echo-3.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2dd8537609f7442797016b44bdf1e957de342ba5..912d87634143d023fab266ded7e212c53082db70 |
| command | sdlc close --issue 544 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T02:09:07-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The migration, conservative exclusion, and combined measurement paths have substantial passing coverage. Two reproduced correctness gaps block shipping: older readers silently discard hybrid measurement evidence, and repair can publish a generation that actual exclusion subsequently rejects.

1. **Strengths**

   - Migration and repair share immutable legacy-close resolution, including review, identity, ancestry, and damaged-evidence checks.
   - Prefix and suffix retain separate source evidence; production tests verify replay after transcript appends.
   - Repair tests exercise stale-card CAS, cancellation, and lost publication acknowledgment.
   - README, command help, and atlas document the new command and measurement behavior.

2. **Critical findings**

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: persisted-measurement-compatibility
    title: |
      Hybrid measurements remain readable as incomplete measurements by older binaries
    detail: |
      cmd/sdlc/internal/gatestate/measurement.go:48 adds legacy_prefix and legacy_cutover while cmd/sdlc/actualinputs.go:233 retains version 1 and the existing engagement method. A temporary regression using the pinned base decoder accepted a valid hybrid snapshot, passed Validate, and silently discarded both fields; the base computation consequently has only suffix evidence to replay, and rewrites lose the prefix permanently. ARCH-EVOLVE requires an explicit method/version boundary that older readers reject because ignoring this evidence changes the total. Add cross-version regression coverage alongside unknown-field preservation checks.
  - id: new
    severity: Critical
    kind: defect
    family: legacy-generation-continuity
    title: |
      Repair uses a weaker history fence than the consumer of its repaired evidence
    detail: |
      cmd/sdlc/issuemigrate_repair.go:217 accepts histories containing removal of a matching completion and does not inspect provenance transitions. A production-fixture probe imported an old card, repaired it, restored its original unbound bytes, and successfully repaired it again; verifyLegacyAdoption then rejected that published card with “legacy adoption tracker history removes completion evidence” at legacycompletion.go:608. Consolidate repair and exclusion onto the same complete, monotone generation validator, including completion/provenance removal and replacement. Add sequence regressions through repair followed by actual exclusion. This violates ARCH-DRY, ARCH-ORDER, and ARCH-PURPOSE.
```

3. **Important findings**

   None beyond the blocking defects above.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Passed the complete `issue`, `gatestate`, `tracker`, and `observe` package tests.
   - Passed the targeted migration/repair/exclusion/prefix matrix: **122.780s**.
   - Passed the existing `TestOwnedActual*` tests: **61.488s**.
   - Temporary overlay regressions reproduced both findings, including successful repair followed by consumer refusal. The overlays were removed; repository implementation files were unchanged.
   - The full repository suite and live Echo conformance were not independently rerun.

6. **Architectural notes**

   | Principle | Result |
   |---|---|
   | ARCH-DRY | **Flag:** competing generation-history validators differ materially. |
   | ARCH-PURE | **Pass:** selection, digest, temporal comparison, and result joining have directly tested pure logic. |
   | ARCH-PURPOSE | **Flag:** repair can report success without restoring usable exclusion evidence. |
   | ARCH-MOCK | **Pass:** private Git fixtures and stateful publication fault injection cover the affected seams. |
   | ARCH-CONSTRAINTS | **Pass:** bounded histories, overflow refusal, aggregate snapshot limits, and scaling controls are present. |
   | ARCH-SECURE | **Pass:** immutable source validation and strict damaged-evidence refusal are well covered. |
   | ARCH-ORDER | **Flag:** repair admits authority-removal sequences rejected by exclusion. |
   | ARCH-FUNERAL | **Pass:** bounded metadata extends existing cards and snapshots; normal repair retries are idempotent. |
   | ARCH-EVOLVE | **Flag:** hybrid evidence changes interpretation without older-reader refusal. |

7. **Plan revision recommendations**

   Add `## Revisions` entries specifying the hybrid compatibility boundary and one shared generation-continuity contract. Correct the Core concepts repair/digest location at plan line 27 to `cmd/sdlc/internal/tracker/legacyrepair.go`, and update revised prefix helper names to match the delivered functions.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"6cc64491-1677-4d7c-9e5b-980b3d51a3fe","repository":"github.com/xianxu/ariadne","issue_id":"000544","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/024361c488a18dd0","reviewed_head":"912d87634143d023fab266ded7e212c53082db70","slot":"ariadne:1","requested_at":"2026-10-22T09:01:05.161845Z"},"phase":"finished","pid":89706,"started_at":"2026-10-22T09:01:05.39856Z","finished_at":"2026-10-22T09:09:07.304627Z","session_id":"01a162d7-b377-7da1-8d72-74abf5a5fd94","outcome":"returned"}] -->

---

## Re-review — 2026-10-22T02:42:57-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 544 — issue migrate drops a legacy close's completion record; later issues' actual measurement then refuses (echo#1 → echo#3) |
| repo | ariadne |
| issue file | workshop/issues/000544-issue-migrate-drops-a-legacy-close-s-completion-record-later-issues-actual-measurement-then-refuses-echo-1-echo-3.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2dd8537609f7442797016b44bdf1e957de342ba5..761da41c42be5520b570b0d18ce5f9d58c7990af |
| command | sdlc close --issue 544 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T02:42:57-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

Both prior defects are fixed, with regression tests that fail when their fixes are removed. Migration, repair, exclusion, and cutover replay checks passed. SHIP remains blocked by a separate persisted-format defect: the new hybrid format accepts unknown fields and silently erases them during rewrites.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Hybrid roots now require version 2 and distinct methods. The frozen decoder matches the pinned base apart from its package/comment; its regression passes and fails when the old discriminators are restored.
  - id: BR-2
    disposition: addressed
    note: |
      Repair and exclusion share legacyGenerationHistory. The restored-authority regression passes; bypassing repair's validator makes all four removal/replacement cases fail.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: persisted-measurement-compatibility
    title: |
      Hybrid snapshot rewrites silently discard unknown root and nested members
    detail: |
      cmd/sdlc/internal/gatestate/measurement.go:47-68 adds the persisted hybrid shape without unknown-member retention. Temporary overlay regressions confirmed that valid snapshots with unknown members at the hybrid root, legacy_prefix, and legacy_prefix.sources[0] all pass ParseSidecar but lose those members through Render; JSON decoding also loses root and prefix members. This reaches production through readGateLedger/writeGateLedger in cmd/sdlc/planreview.go:57-80. ARCH-EVOLVE requires lossless additive evolution. This is the 2nd finding in family persisted-measurement-compatibility. Earlier rounds fixed instances. Do NOT fix only these examples: state and enforce the rule that every accepted unknown member survives every supported decode/rewrite path, enumerate the snapshot's nested records and JSON/YAML paths, and add preservation regressions while retaining strict known-field and duplicate-key validation.
  - id: new
    severity: Minor
    kind: defect
    family: documented-history-bounds
    title: |
      Documented tracker-history limit differs from the implemented limit
    detail: |
      atlas/workflow/issue-tracker-migration.md:62 and workshop/plans/000544-legacy-completion-plan.md:150 describe a 5,000-card-commit limit, but legacyGenerationHistory at cmd/sdlc/legacycompletion.go:556-562 uses legacyCompletionHistoryLimit, which is 10,000. Align the documented operating envelope with the intended enforced limit (ARCH-CONSTRAINTS).
```

1. **Strengths**

   - Repair and exclusion now enforce one monotone generation rule, including removed or replaced authority.
   - The pinned old-decoder fixture establishes meaningful compatibility evidence; mutation testing confirms the regression detects the original defect.
   - Production fixtures cover retained replay, conservative legacy ordering, repair idempotence, and publication failures.
   - README and atlas updates cover the new repair command and cutover measurement flow.

2. **Critical findings**

   - Unknown-field data loss described above. Capture and preserve unknown members throughout the hybrid snapshot tree and its rewrite paths.

3. **Important findings**

   None.

4. **Minor findings**

   - Correct the documented tracker-history limit described above.

5. **Test coverage notes**

   - Passed the gatestate, issue, tracker, and observation packages.
   - Passed two targeted command matrices: 92.335s and 83.453s.
   - Both prior fixes have independently demonstrated failing mutation controls.
   - New preservation probes reproduced loss at all three tested YAML locations, plus JSON root/prefix decoding.
   - Temporary overlays were removed; the checkout remains unchanged. The full repository suite was not rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared proof and generation-history validation.
   - **ARCH-PURE — pass:** selection, transformation, temporal decisions, and result joining remain separate from IO. Core-concept locations match the plan’s revisions.
   - **ARCH-PURPOSE — pass:** import, repair, read-only recovery, and prefix/suffix measurement are delivered.
   - **ARCH-MOCK — pass:** private Git fixtures and stateful publication faults exercise the production seams.
   - **ARCH-CONSTRAINTS — flag, Minor:** documented history bound differs from enforcement.
   - **ARCH-SECURE — pass:** immutable pins, repository identity, malformed evidence, and history failures receive explicit checks.
   - **ARCH-ORDER — pass:** repair/exclusion share the monotone authority fence; mutation controls verify reachability.
   - **ARCH-FUNERAL — pass:** bounded metadata extends existing cards; repair retries create no duplicate commit.
   - **ARCH-EVOLVE — flag, Critical:** old-reader refusal is fixed, but current-reader unknown-field preservation remains absent.

7. **Plan revision recommendations**

   Add `## Revisions` entries that:

   - Enumerate the hybrid snapshot’s JSON/YAML decode and rewrite paths and nested records, with a lossless unknown-member preservation contract and regression coverage.
   - Reconcile the documented tracker-history limit with the enforced value.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d04bacdd-c7d2-452b-8634-57cf67811bb9","repository":"github.com/xianxu/ariadne","issue_id":"000544","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/96a5c3de1f0b85d6","reviewed_head":"761da41c42be5520b570b0d18ce5f9d58c7990af","slot":"ariadne:1","requested_at":"2026-10-22T09:36:51.295773Z"},"phase":"finished","pid":18393,"started_at":"2026-10-22T09:36:51.501296Z","finished_at":"2026-10-22T09:42:56.588208Z","session_id":"01a162f8-7302-7910-8d47-feb6cbb9f08e","outcome":"returned"}] -->

---

## Re-review — 2026-10-22T03:09:31-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 544 — issue migrate drops a legacy close's completion record; later issues' actual measurement then refuses (echo#1 → echo#3) |
| repo | ariadne |
| issue file | workshop/issues/000544-issue-migrate-drops-a-legacy-close-s-completion-record-later-issues-actual-measurement-then-refuses-echo-1-echo-3.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2dd8537609f7442797016b44bdf1e957de342ba5..c90b0bd43d24d8b3e06c02d65c2569d08169eddc |
| command | sdlc close --issue 544 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T03:09:31-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation satisfies the issue’s revised migration, repair, exclusion, and cutover-measurement contracts. Both open findings are addressed with supporting evidence; no new blocking findings surfaced. This review does not establish the separately required final-head full-suite PASS.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      Shared codecs retain unknown members throughout measurement records across JSON/YAML rewrites. The preservation matrix and production ledger rewrite pass; temporarily disabling retention makes all 40 matrix cases and the production rewrite regression fail.
  - id: BR-4
    disposition: addressed
    note: |
      The atlas history-bound passage and plan Revisions now specify 10,000 entries plus an overflow sentinel, matching legacyCompletionHistoryLimit and legacyGenerationHistory.
```

1. **Strengths**

   - Migration, repair, and exclusion share immutable evidence verification and generation-continuity checks.
   - Repair validates every publication base, including retries. Tests exercise restored authority, unrelated-card movement, cancellation, and uncertain publication.
   - Hybrid measurements preserve distinct prefix/suffix evidence and replay after transcript appends.
   - The frozen older decoder exactly matches the pinned base; compatibility tests confirm explicit hybrid refusal.
   - README, help, and atlas document the new repair command and measurement behavior.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed the gatestate, issue, tracker, and observe packages; selected command-level regressions passed in 126.150 seconds. Golint, shelllint, touched-package vet, and range whitespace checks passed. Mutation checks confirmed BR-3’s tests fail without retention. Temporary overlays were removed; the checkout remains clean. Echo conformance and the full recorded suite were not independently rerun here.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared codecs, evidence verifier, generation validator, and source-retention helper.
   - **ARCH-PURE — pass:** selection, repair planning, temporal decisions, codecs, and result joining remain independently testable without IO.
   - **ARCH-PURPOSE — pass:** import, existing-import repair, terminal recovery, and cutover-spanning replay are implemented.
   - **ARCH-MOCK — pass:** existing seams support isolated Git repositories and controlled publication failures.
   - **ARCH-CONSTRAINTS — pass:** history, snapshot, source-count, and nesting bounds have refusal paths and coverage.
   - **ARCH-SECURE — pass:** persisted evidence receives strict validation; missing or contradictory authority refuses visibly.
   - **ARCH-ORDER — pass:** continuity checks and conditional publication protect retries; deterministic schedules exercise relevant races.
   - **ARCH-FUNERAL — pass:** bounded metadata uses existing artifact families; confirmed repair retries add no commits.
   - **ARCH-EVOLVE — pass:** unknown members survive rewrites; incompatible hybrid interpretation has an explicit version/method boundary.

7. **Plan revision recommendations:** None required. Existing Revisions describe the delivered compatibility and publication-fence corrections.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"1f7baefa-4f22-4429-b540-3b5df657a955","repository":"github.com/xianxu/ariadne","issue_id":"000544","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/7fd41feeb807828e","reviewed_head":"c90b0bd43d24d8b3e06c02d65c2569d08169eddc","slot":"ariadne:1","requested_at":"2026-10-22T10:06:05.891901Z"},"phase":"finished","pid":59743,"started_at":"2026-10-22T10:06:05.986698Z","finished_at":"2026-10-22T10:09:30.758234Z","session_id":"01a16313-37a8-7af1-82a5-788fd7064deb","outcome":"returned"}] -->
