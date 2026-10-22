# Boundary Review — ariadne#526 (whole-issue close)

| field | value |
|-------|-------|
| issue | 526 — landing: drop one car from a composed train before merge (no per-car hold today) |
| repo | ariadne |
| issue file | workshop/issues/000526-landing-drop-one-car-from-a-composed-train-before-merge-no-per-car-hold-today.md |
| boundary | whole-issue close |
| milestone | — |
| window | 510aab6ff3a55bfdf5472e45eedf5512f6d04420..8fb25eb263f95e8c8169e4db3bc2560f4be1c1f3 |
| command | sdlc close --issue 526 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T13:56:57-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The normal three-car drop flow, cleanup barrier, ownership checks, and documentation are well covered. One reproduced ordering bug blocks shipping: an accepted drop can lose precedence to a concurrent membership update, preventing survivor re-formation.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: drop-preempts-completion
    title: |
      Accepted drops must fence non-merge completion updates
    detail: |
      cmd/sdlc/landingtrain.go:176 checks pending drops before computing and publishing a completion, but UpdateTrain at cmd/sdlc/internal/landingline/traincontrol_io.go:178 commits membership changes without checking accepted drops under line.lock. A drop accepted between those operations can target a car removed by a bisect completion; the next ApplyTrainDrops then fails with “drop does not pin a unique riding car,” and the surviving conductor exits instead of automatically re-forming. A temporary Go-overlay regression reproduced this failure. Fence completion updates atomically against pending drops and route the pending result through applyTrainDrops using the pre-completion train; cover membership-changing and terminal completions. ARCH-ORDER, ARCH-PURPOSE.
```

1. **Strengths**

   - Drop acceptance and merge admission share `line.lock`; the durable merge-start marker survives reintegration.
   - Cleanup requires positive PR closure and leased-ref deletion evidence, including lost replies and retargeted remotes.
   - Real Git fixtures verify survivor ancestry, dropped-commit exclusion, and both dropped-conductor and dropped-middle cases.
   - README, help, and atlas explain authority, recovery, compatibility, and the merge boundary.

2. **Critical findings**

   - **Accepted drop versus completion race:** [traincontrol_io.go:178](/Volumes/TestScratch/review-checkouts/7df8449c136c8f51/cmd/sdlc/internal/landingline/traincontrol_io.go:178), called from [landingtrain.go:199](/Volumes/TestScratch/review-checkouts/7df8449c136c8f51/cmd/sdlc/landingtrain.go:199). The merge path handles `ErrTrainDropPending`, but ordinary completion updates lack the equivalent fence. Preserve the original train until the locked update confirms that no accepted drop takes precedence.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed:

   - Complete `landingline` package tests.
   - Targeted command drop and authorization tests, including real three-car re-formation.
   - Race-enabled control, transition, and publication tests.
   - Seven Python deadline/process controls.
   - `git diff --check`.

   The temporary overlay probe reproduced: bisect completion computed → drop accepted → membership update committed → drop application fails. Existing tests omit this ordering. Repository files remained unchanged; the full repository suite was not rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** reuses locks, codec, abort effects, and normal admission.
   - **ARCH-PURE — pass:** decisions and transitions remain free of IO; integration seams are separate.
   - **ARCH-PURPOSE — flag:** the reproduced race interrupts promised automatic survivor re-formation.
   - **ARCH-MOCK — pass:** stateful GH doubles and portable real Git fixtures exercise production seams.
   - **ARCH-CONSTRAINTS — pass:** bounded records, scoped polling, and counted batch-work checks.
   - **ARCH-SECURE — pass:** caller/target separation, strict decoding, and destination revalidation.
   - **ARCH-ORDER — flag:** pending-drop precedence is enforced at merge admission but not every completion commit.
   - **ARCH-FUNERAL — pass:** cleanup obligations persist; records and exclusions have bounded replacement/retirement paths.
   - **ARCH-EVOLVE — pass:** strict known fields and nested unknown-field preservation are implemented and tested.

7. **Plan revision recommendations**

   Add a `## Revisions` entry requiring every pre-merge completion commit to serialize against accepted drops, including membership reductions and terminal refusals. Require deterministic tests for both lock orderings and preservation of the pre-completion train when drop wins.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"82b3bd5e-cfe2-4e36-9dcd-4aa6bb778cff","repository":"github.com/xianxu/ariadne","issue_id":"000526","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/7df8449c136c8f51","reviewed_head":"8fb25eb263f95e8c8169e4db3bc2560f4be1c1f3","slot":"ariadne:1","requested_at":"2026-10-21T20:54:36.203878Z"},"phase":"finished","pid":28366,"started_at":"2026-10-21T20:54:36.304738Z","finished_at":"2026-10-21T20:56:57.182586Z","session_id":"01a1603e-952a-7120-aafc-4ee8b5a583f5","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T14:14:16-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 526 — landing: drop one car from a composed train before merge (no per-car hold today) |
| repo | ariadne |
| issue file | workshop/issues/000526-landing-drop-one-car-from-a-composed-train-before-merge-no-per-car-hold-today.md |
| boundary | whole-issue close |
| milestone | — |
| window | 510aab6ff3a55bfdf5472e45eedf5512f6d04420..e48ce11ae824a0ece2a49fa32cf5613040027edf |
| command | sdlc close --issue 526 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T14:14:16-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

BR-1 is addressed: completion updates now atomically defer to accepted drops, and mutation checks confirm the regression tests exercise both the storage fence and conductor routing. The real three-car scenarios and cleanup controls pass. One Important compatibility defect remains: accepting another drop can erase unknown nested fields from the durable cleanup record.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      UpdateTrain now checks pending drops under line.lock, and commitTrainCompletion routes pending results through ApplyTrainDrops using the original train. TestTrainCompletionAndDropOrdering and TestLandingDropPreemptsComputedCompletion pass at head. Removing the storage guard fails all three drop-first completion cases; removing the conductor routing fails bisect, terminal, phase and abort cases.
findings:
  - id: new
    severity: Important
    kind: defect
    family: preserve-unknown-fields-on-rewrite
    title: |
      Drop acceptance erases unknown cleanup publication fields
    detail: |
      cmd/sdlc/internal/landingline/traincontrol.go:190 replaces Cleanup.Publication with Publication, discarding unknown fields stored only in the cleanup snapshot. A temporary overlay regression accepted one drop, persisted future_cleanup_proof under cleanup.publication, then accepted another member's drop; ReadTrainControl returned an empty Extra map. ARCH-EVOLVE requires nested unknown fields to survive every rewrite. Retain the existing pending cleanup snapshot, or preserve its unknown fields when refreshing known values, and add a regression through RequestTrainDrop with metadata unique to cleanup.publication.
```

**1. Strengths**

- Completion decisions commit before in-memory replacement or effects; deferred abort uses the same fence ([landingtrain_completion.go](/Volumes/TestScratch/review-checkouts/361c208e96fa5db1/cmd/sdlc/landingtrain_completion.go:8)).
- Cleanup requires positive PR closure and leased-ref absence; tests cover lost replies, stale observations, and retargeted remotes.
- Real three-runner tests verify survivor reachability, dropped-commit exclusion, and obsolete-train cleanup.
- README, help, and atlas document authority, merge-start refusal, and recovery.

**2. Critical findings**

None. BR-1 is addressed.

**3. Important findings**

The nested-field loss above is reproducible. Existing codec tests exercise decode/encode round trips but miss the request-driven rewrite that loses metadata.

**4. Minor findings**

None.

**5. Test coverage notes**

Passed:

- Full `landingline` package tests.
- `TestLanding(Drop|Abort)` command tests, including real train recomposition.
- Targeted race tests for concurrent requests and completion ordering.
- Seven Python deadline/process-ownership tests.
- Diff whitespace checks.

Both BR-1 mutations failed their expected regressions. The new preservation regression fails on pinned head. Temporary overlays were removed; the checkout remains clean. I did not independently rerun the repository-wide suite.

**6. Architectural notes**

- **ARCH-DRY — pass:** existing locks, codecs, abort transitions, and admission paths are reused.
- **ARCH-PURE — pass:** listed pure entities remain deterministic and have direct tests; integration entities use portable fixtures and injected boundaries.
- **ARCH-PURPOSE — pass:** both conductor-drop and middle-car-drop scenarios land the survivors while excluding the dropped head.
- **ARCH-MOCK — pass:** stateful GH doubles and real Git/native-process fixtures exercise the production seams under the approved conformance approach.
- **ARCH-CONSTRAINTS — pass:** bounded record, scoped watcher, short transactions, and counted-work guard are present.
- **ARCH-SECURE — pass:** strict decoding, pinned identities, separate caller authorization, and destination revalidation are implemented.
- **ARCH-ORDER — pass:** completion/drop ordering is enforced under the lock; cancellation joins obsolete work before cleanup.
- **ARCH-FUNERAL — pass:** one bounded control record, ticket-based exclusion retirement, and confirmed external cleanup define artifact lifetimes.
- **ARCH-EVOLVE — flag:** the cleanup-snapshot replacement violates nested unknown-field preservation.

**7. Plan revision recommendations**

Append a `## Revisions` entry naming the cleanup-snapshot preservation correction and its request-driven regression. The Core concepts table matches the implementation; README and atlas updates are present.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"68bcc8aa-4a0c-42c4-9036-5e99a3a5ff23","repository":"github.com/xianxu/ariadne","issue_id":"000526","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/361c208e96fa5db1","reviewed_head":"e48ce11ae824a0ece2a49fa32cf5613040027edf","slot":"ariadne:1","requested_at":"2026-10-21T21:11:25.066043Z"},"phase":"finished","pid":24809,"started_at":"2026-10-21T21:11:25.217993Z","finished_at":"2026-10-21T21:14:15.266406Z","session_id":"01a1604d-faa9-7013-b17f-bf0e4c2cf76a","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T14:21:08-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 526 — landing: drop one car from a composed train before merge (no per-car hold today) |
| repo | ariadne |
| issue file | workshop/issues/000526-landing-drop-one-car-from-a-composed-train-before-merge-no-per-car-hold-today.md |
| boundary | whole-issue close |
| milestone | — |
| window | 510aab6ff3a55bfdf5472e45eedf5512f6d04420..2eb193c29df4550c63266ec065030a071a234d4b |
| command | sdlc close --issue 526 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T14:21:08-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range implements the documented drop protocol, survivor re-formation, and merge-start refusal. BR-2 is addressed with regression evidence that fails without the fix. BR-1 remains addressed. No new Critical, Important, or Minor findings.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      traincontrol.go:193 preserves cleanup.publication-only unknown fields. TestTrainDropRequestsPreserveCleanupPublicationUnknownFields passes through RequestTrainDrop, ApplyTrainDrops, and FinishTrainControl; removing the preservation assignment in a temporary Go overlay makes it fail with lost metadata.
  - id: BR-1
    disposition: addressed
    note: |
      Completion and abort routing retain the authoritative riders until the locked commit. Ordering and shell outcome tests pass; removing UpdateTrain's pending-drop guard in a temporary overlay makes all three drop-first completion regressions fail.
```

1. **Strengths**

   - Drop acceptance and merge admission share the durable lock boundary in `cmd/sdlc/internal/landingline/traincontrol_io.go`.
   - Real three-car tests verify both dropped-conductor and dropped-middle-car cases: survivors land together, the dropped commit stays absent, and obsolete PR/ref cleanup completes.
   - Cleanup requires positive external evidence and retains uncertainty after lost replies or missing observations (`cmd/sdlc/landingtrain_drop.go`).
   - README, help, and atlas cover the command, authority, irreversible merge boundary, and recovery.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Independently passed:

   - Full `landingline` package, including counted-work guard.
   - Selected command/integration tests covering drop, authorization, completion, abort, cleanup, retry, and real train re-formation.
   - Focused `landingline` race checks.
   - Seven Python deadline/process-ownership controls.
   - `go vet` for both affected Go packages and range whitespace checks.

   Both prior fixes have mutation evidence. The repository-wide recorded suite was not rerun during this review. Temporary overlays were removed; the checkout remains clean.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Reuses existing transitions, locks, codec, outcomes, and normal formation.
   - **ARCH-PURE — pass:** Decisions and transitions remain free of IO; direct pure tests exist.
   - **ARCH-PURPOSE — pass:** Tests establish actual survivor landing and exclusion of the dropped commit.
   - **ARCH-MOCK — pass:** Production seams support stateful GH doubles and portable real-Git fixtures.
   - **ARCH-CONSTRAINTS — pass:** Bounded record, shared polling, short transactions, and counted linear-work checks match the envelope.
   - **ARCH-SECURE — pass:** Strict decoding, pinned identities, caller authorization, and destination revalidation protect the relevant boundaries.
   - **ARCH-ORDER — pass:** Locked completion/merge fences and controlled ordering tests enforce drop precedence.
   - **ARCH-FUNERAL — pass:** Control retention and replacement follow ticket retirement and confirmed cleanup; watcher lifetime is joined.
   - **ARCH-EVOLVE — pass:** Same-generation rewrites preserve nested unknown fields; BR-2’s regression exercises the corrected path.

   Core-concept table entries match their locations and classifications.

7. **Plan revision recommendations:** None. Existing revisions describe the implemented corrections.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"1457ca26-38d1-4d0a-ae81-766a81d34d8e","repository":"github.com/xianxu/ariadne","issue_id":"000526","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/73a7754a859655e6","reviewed_head":"2eb193c29df4550c63266ec065030a071a234d4b","slot":"ariadne:1","requested_at":"2026-10-21T21:18:52.460537Z"},"phase":"finished","pid":33747,"started_at":"2026-10-21T21:18:52.5517Z","finished_at":"2026-10-21T21:21:01.929383Z","session_id":"01a16054-cd8f-7aa3-84cd-286b03b5e797","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T16:34:15-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 526 — landing: drop one car from a composed train before merge (no per-car hold today) |
| repo | ariadne |
| issue file | workshop/issues/000526-landing-drop-one-car-from-a-composed-train-before-merge-no-per-car-hold-today.md |
| boundary | whole-issue close |
| milestone | — |
| window | f9c0dac6f52b6c7736bcf686777ea01d88c83d22..a887eb11140108b188dd37dd2abaebdb8144a18a |
| command | sdlc close --issue 526 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T16:34:15-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The interdiff fixes compatibility with main’s attribution phase while retaining the approved drop-precedence contract. The atlas conflict resolution preserves both sides’ additions. No scope expansion or new blocking defect found.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Both completion/drop lock orders remain covered at SplitAssessed and attribution entry; targeted conductor tests verify dropped and surviving outcome identities.
  - id: BR-2
    disposition: addressed
    note: |
      The request-driven nested-metadata preservation regression remains intact and passes through acceptance, ApplyTrainDrops and FinishTrainControl.
```

1. **Strengths**
   - `traincontrol.go:98` accepts the existing attribution phase without weakening other validation.
   - `traincontrol_completion_test.go:20` retains the membership-reduction oracle and separately tests attribution entry.
   - `atlas/workflow/sdlc-binary.md:530` preserves attribution, landing-check and drop documentation.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage:** Full landingline package passed; targeted conductor tests passed; focused race checks passed. Removing the phase fix with a temporary overlay caused three regression cases to fail. Repository remained unchanged. The recorded full suite was not rerun during this review.

6. **Architecture**
   - **ARCH-DRY — pass:** Reuses existing validation and completion paths.
   - **ARCH-PURE — pass:** Phase validation remains pure; storage tests exercise the integration boundary.
   - **ARCH-PURPOSE — pass:** Preserves drop precedence at the actual membership-changing event.
   - **ARCH-EVOLVE — pass:** No format/version change or weakened strictness; nested unknown-field preservation tests pass.

7. **Plan revisions:** None required. The added revision records integration compatibility within approved scope.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"3229c300-5da7-4724-9e20-d605409890d8","repository":"github.com/xianxu/ariadne","issue_id":"000526","recipe":"fix-review","round":4,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/052b471e7ebeda4f","reviewed_head":"a887eb11140108b188dd37dd2abaebdb8144a18a","slot":"ariadne:1","requested_at":"2026-10-21T23:33:10.948882Z"},"phase":"finished","pid":61994,"started_at":"2026-10-21T23:33:11.118153Z","finished_at":"2026-10-21T23:34:15.14084Z","session_id":"01a160cf-c4ef-7322-9dd8-02013617efb4","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T18:18:24-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 526 — landing: drop one car from a composed train before merge (no per-car hold today) |
| repo | ariadne |
| issue file | workshop/issues/000526-landing-drop-one-car-from-a-composed-train-before-merge-no-per-car-hold-today.md |
| boundary | whole-issue close |
| milestone | — |
| window | 640577b38c53a9193453874fca19b2149b78916e..45e47313662ffd2415d6ad58ffce1ca8a6fc1fe1 |
| command | sdlc close --issue 526 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T18:18:24-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The interdiff completes BR-2’s unknown-field preservation fix: cleanup retains its own metadata on key collisions while copying publication-only keys. The regression fails with the previous implementation and passes at the pinned head. The revision records this correction within approved scope; no new blocking defects found.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Completion/drop ordering tests and targeted conductor completion/abort tests pass; this interdiff leaves their fencing behavior unchanged.
  - id: BR-2
    disposition: addressed
    note: |
      The request/apply/finish regression verifies conflicting-key preservation in both objects; restoring the previous merge order through a temporary overlay makes it fail at the second acceptance.
```

1. **Strengths**
   - `traincontrol.go:193` reuses `mergeExtra` with the correct collision precedence.
   - `traincontrol_test.go:535` exercises real persistence through acceptance, apply and finish, checking both metadata scopes.
   - The added revision describes the existing preservation contract without expanding scope.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage:** Full landingline package and targeted conductor completion/abort tests passed. Reverting the fix through a temporary overlay caused the regression to fail as expected. `git diff --check` passed; checkout remained unchanged. Full repository suite was not rerun.

6. **Architectural notes**
   - **ARCH-DRY — pass:** Reuses the existing metadata merge helper.
   - **ARCH-PURE — pass:** The changed decision remains deterministic and contains no IO.
   - **ARCH-PURPOSE — pass:** Corrects collision handling for the shared acceptance path.
   - **ARCH-EVOLVE — pass:** Preserves each object’s unknown values without changing known fields or version semantics.

   No new architectural or user-facing surface requires atlas or README changes in this window.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"8229e720-b20f-4240-bf95-c80b9fce38cb","repository":"github.com/xianxu/ariadne","issue_id":"000526","recipe":"fix-review","round":5,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/7c16e2e08f1dfdf2","reviewed_head":"45e47313662ffd2415d6ad58ffce1ca8a6fc1fe1","slot":"ariadne:1","requested_at":"2026-10-22T01:17:14.379389Z"},"phase":"finished","pid":4685,"started_at":"2026-10-22T01:17:14.518065Z","finished_at":"2026-10-22T01:18:23.64197Z","session_id":"01a1612f-096d-7291-9005-5d17bf5d8897","outcome":"returned"}] -->
