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
