# Boundary Review — ariadne#543 (whole-issue close)

| field | value |
|-------|-------|
| issue | 543 — new repositories start on the legacy issue path: initial setup should create the issue tracker (non-brain repos) |
| repo | ariadne |
| issue file | workshop/issues/000543-new-repositories-start-on-the-legacy-issue-path-initial-setup-should-create-the-issue-tracker-non-brain-repos.md |
| boundary | whole-issue close |
| milestone | — |
| window | a4600f5050712a27c7979cfd035cccd8b5a1ceca..cf6b8a95b62cb4547f2ec5a8031ab72187619115 |
| command | sdlc close --issue 543 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T23:21:44-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The initialization workflow, recovery model, exclusions, and documentation are well covered. However, init can incorrectly declare the inventory empty when an issue exists on a remote branch absent from local tracking refs. I reproduced this with the pinned production CLI; it blocks shipping.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: authoritative-inventory-completeness
    title: |
      Initialization treats cached remote branches as the complete publication inventory
    detail: |
      cmd/sdlc/issueinit.go:166-172 authorizes empty initialization using migrationInventory, whose migrationBranchRefs implementation (issuemigrate.go:219-237) enumerates cached refs/remotes entries. The preceding snapshots fetch only main and issue-tracker. In a disposable two-clone reproduction, clone B pushed legacy-work containing workshop/issues/000001-old.md after clone A last fetched; clone A's issue init returned 0 and published an empty tracker and cutover marker. This violates the complete-inventory refusal contract (ARCH-PURPOSE, ARCH-SECURE, ARCH-ORDER). Refresh and pin the publication remote's branch inventory before authorizing initialization, refuse failed observations, and add regressions for remote branches created or advanced since the last fetch.
```

1. **Strengths**

   - Opaque `InitState` and `InitTransition` enforce publication ordering and retain confirmed identities after failure.
   - Deterministic effect barriers exercise retries, lost acknowledgments, cancellation, dirty checkout protection, and competing clones.
   - Initialization reuses migration’s bootstrap CAS, exact-root verification, and cutover publication.
   - README, atlas, command help, and TL setup instructions document the explicit setup step and exclusions.

2. **Critical findings**

   - **Incomplete authoritative inventory:** [issueinit.go:166](/Volumes/TestScratch/review-checkouts/5ad7586956a8926b/cmd/sdlc/issueinit.go:166). The machine-readable finding above contains the reproduction and fix sketch. Correct the shared inventory boundary so both initialization and migration receive fresh remote evidence.

3. **Important findings**

   None additional.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Targeted initialization, eligibility, transition, hint, recovery-contract, and TL procedure tests passed.
   - Existing `TestIssueMigrate*` tests passed.
   - `git diff --check` passed; checkout remained unchanged.
   - The production CLI reproduction exposed missing remote-only branch coverage. Existing branch-inventory coverage creates its issue branch locally.

6. **Architectural notes for upcoming work**

   - **ARCH-DRY — pass:** shares migration publication helpers and hint selection.
   - **ARCH-PURE — pass:** pure entities have direct tests without IO; effects remain injected.
   - **ARCH-PURPOSE — flag:** complete inventory verification is not delivered.
   - **ARCH-MOCK — pass:** portable bare-Git fixtures exercise the production adapter through controllable effect boundaries.
   - **ARCH-CONSTRAINTS — pass:** process-growth tests cover file and branch scaling; no new worker fan-out.
   - **ARCH-SECURE — flag:** cached remote refs are accepted as authoritative absence evidence.
   - **ARCH-ORDER — flag:** incomplete observations authorize publication; invocation sequencing itself is explicit and tested.
   - **ARCH-FUNERAL — pass:** reuses existing artifact lifecycles.
   - **ARCH-EVOLVE — pass:** introduces no persisted-format change.

   Core-concept locations and classifications match the implementation. README and atlas gates are satisfied.

7. **Plan revision recommendations**

   Append a `## Revisions` entry specifying that authoritative inventory enumerates and pins current publication-remote branches, with failed observations refusing initialization. Extend verification to branches created and advanced remotely since the initializer’s last fetch.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"32790977-e9cb-4c29-8c50-5f680f5c59c8","repository":"github.com/xianxu/ariadne","issue_id":"000543","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/5ad7586956a8926b","reviewed_head":"cf6b8a95b62cb4547f2ec5a8031ab72187619115","slot":"ariadne:3","requested_at":"2026-10-22T06:18:30.566774Z"},"phase":"finished","pid":81680,"started_at":"2026-10-22T06:18:30.772652Z","finished_at":"2026-10-22T06:21:43.262979Z","session_id":"01a16242-dbb6-7432-a192-5fdfdbd64e3e","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T23:36:01-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 543 — new repositories start on the legacy issue path: initial setup should create the issue tracker (non-brain repos) |
| repo | ariadne |
| issue file | workshop/issues/000543-new-repositories-start-on-the-legacy-issue-path-initial-setup-should-create-the-issue-tracker-non-brain-repos.md |
| boundary | whole-issue close |
| milestone | — |
| window | a4600f5050712a27c7979cfd035cccd8b5a1ceca..96e8ff1cfaf539a80c5c174ce4ba7bbf83992bae |
| command | sdlc close --issue 543 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T23:36:01-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-1’s cached-branch defect is fixed and backed by meaningful regressions. Initialization, recovery, and documentation are otherwise well covered. One blocking gap remains: archived issues outside main are omitted from the inventory, allowing initialization to publish an empty tracker for a nonempty legacy repository.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Verified the all-heads refresh and pinned issue/anchor reads. The targeted suite passes at head; removing the fetch correction in a disposable copy makes all six remote-create/advance/narrow-fetch cases, failed-observation test, and moved-main test fail.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: authoritative-inventory-completeness
    title: |
      Empty initialization ignores archived issues on branches and in other worktrees
    detail: |
      cmd/sdlc/issuemigrate.go:192-199 scans branch edits and dirty worktrees only under d.Home; archived issues are read only from pinned main at line 185. Two disposable-head regressions demonstrated successful empty-tracker publication with workshop/history/issues/000001-old.md existing only on a remote branch or uncommitted in another worktree. This violates the complete empty-inventory contract consumed by cmd/sdlc/issueinit.go:166-172 (ARCH-PURPOSE, ARCH-SECURE). This is the 2nd finding in family authoritative-inventory-completeness. Earlier rounds fixed instances. Do NOT fix only these instances: state and enforce the inventory rule across active/archive roots and main/local-branch/remote-branch/worktree scopes, refusing unsupported evidence conservatively, and test the enumeration through the shared init/migrate boundary.
```

1. **Strengths**

   - The opaque `InitState` and `InitTransition` keep publication ordering in the pure core.
   - Recovery tests exercise deterministic interruptions before and after every effect, cancellation, dirty checkout preservation, and competing clones.
   - BR-1 now refreshes publication-remote heads independently of configured fetch filters and pins subsequent reads.
   - README, atlas, command help, recovery catalog, and TL setup instructions cover the new command.

2. **Critical findings**

   - The inventory omission above blocks shipping. Centralize the roots and scopes that constitute legacy issue evidence; ensure archive evidence cannot disappear before `EmptyInitEligible` evaluates it.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Targeted tracker, SDLC, recovery, inventory, and TL tests passed.
   - BR-1’s regressions failed when its fetch correction was removed.
   - Both additional archive-inventory regressions failed against pinned head because initialization incorrectly succeeded.
   - Disposable review files were removed; the checkout remains clean at the pinned head. The full recorded suite was not independently verified.

6. **Architectural notes for upcoming work**

   | Marker | Result |
   |---|---|
   | ARCH-DRY | Pass: initialization reuses migration publication and shared hint policy. |
   | ARCH-PURE | Pass: pure eligibility/transitions have tests without IO; effects use an injected boundary. |
   | ARCH-PURPOSE | **Flag:** complete-inventory refusal remains incomplete. |
   | ARCH-MOCK | Pass: portable real-Git fixtures exercise the production effect seam. |
   | ARCH-CONSTRAINTS | Pass: explicit batch path; process-growth tests cover representative file/branch counts. |
   | ARCH-SECURE | **Flag:** omitted archive evidence is treated as absence. |
   | ARCH-ORDER | Pass: explicit transitions stop after uncertainty and retain confirmed identities. |
   | ARCH-FUNERAL | Pass: existing artifact families and lifecycle owners are reused. |
   | ARCH-EVOLVE | Pass: no persisted-format change introduced. |

   Core-concept locations and classifications match the implementation. The README and atlas update gates are satisfied.

7. **Plan revision recommendations**

   Append a `## Revisions` entry defining authoritative inventory as the active/archive-root × repository-scope enumeration, identifying conservative handling of unsupported archive evidence and the regressions that enforce it. Current acceptance claims should remain incomplete until that rule is implemented and verified.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"bdf2e958-9634-4fd9-976d-f534634ba6c2","repository":"github.com/xianxu/ariadne","issue_id":"000543","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/e12132d265e05549","reviewed_head":"96e8ff1cfaf539a80c5c174ce4ba7bbf83992bae","slot":"ariadne:3","requested_at":"2026-10-22T06:33:05.574907Z"},"phase":"finished","pid":31549,"started_at":"2026-10-22T06:33:05.67848Z","finished_at":"2026-10-22T06:36:00.662126Z","session_id":"01a16250-34c1-72d0-aa44-d6d2b0f4d510","outcome":"returned"}] -->

---

## Re-review — 2026-10-22T00:03:32-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 543 — new repositories start on the legacy issue path: initial setup should create the issue tracker (non-brain repos) |
| repo | ariadne |
| issue file | workshop/issues/000543-new-repositories-start-on-the-legacy-issue-path-initial-setup-should-create-the-issue-tracker-non-brain-repos.md |
| boundary | whole-issue close |
| milestone | — |
| window | a4600f5050712a27c7979cfd035cccd8b5a1ceca..e344c6cec0332b0d8316be96dce73dd04f18280c |
| command | sdlc close --issue 543 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T00:03:32-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The presence-only initializer fixes the original archive cases, and the targeted tests pass. However, its branch inventory still excludes evidence by branch name: archived issues held only on a local `issue-tracker` branch do not prevent empty initialization. BR-2 remains blocking.

```findings
dispose:
  - id: BR-2
    disposition: not-addressed
    note: |
      The original archive regressions pass and fail when the presence guard is removed. However, issueinit_inventory.go:47 inherits migrationBranchRefs's unconditional exclusion of refs/heads/issue-tracker at issu emigrate.go:242 (path spelling: cmd/sdlc/issuemigrate.go). A disposable pinned-head regression placed workshop/history/issues/000001-old.md only on that unchecked-out local branch; issue init succeeded, changed published main, and created the remote tracker. Removing that exclusion made the regression pass. The complete-inventory rule remains unenforced (ARCH-PURPOSE, ARCH-SECURE). This is the third demonstrated instance in family authoritative-inventory-completeness; sweep every branch exclusion and require evidence for exemptions rather than fixing only this branch name.
```

1. **Strengths**

   - The opaque `InitState` and `InitTransition` enforce publication ordering and preserve confirmed identities.
   - Recovery tests inject failures before and after production effects, including lost acknowledgments and competing initialization.
   - The presence matrix covers three roots across branch, detached-worktree, index, ignored, and untracked evidence.
   - README, atlas, command help, and TL setup instructions document the explicit initialization step.

2. **Critical findings**

   - **BR-2 remains open:** [issuemigrate.go:239](/Volumes/TestScratch/review-checkouts/35117b907c7222b3/cmd/sdlc/issuemigrate.go:239) supplies migration-specific exclusions to [initializationEvidence:47](/Volumes/TestScratch/review-checkouts/35117b907c7222b3/cmd/sdlc/issueinit_inventory.go:47). Initialization must enumerate all relevant tips; a reserved name alone cannot prove that a branch contains no legacy issues. Separate complete observation from consumer-specific filtering and test all exemptions across the three roots.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Targeted tracker, SDLC initialization/migration/recovery/hint, and TL tests passed; SDLC tests completed in 61.298s.
   - Removing the presence guard made both selected archive regressions fail.
   - The additional local `issue-tracker` branch regression failed on pinned head and passed after removing its exclusion.
   - `git diff --check` passed. Disposable files were removed; the reviewed checkout remains unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared path predicate and existing publication helpers.
   - **ARCH-PURE — pass:** pure transition policy with an injected effect dispatcher.
   - **ARCH-PURPOSE — flag:** BR-2’s complete-inventory contract remains incomplete.
   - **ARCH-MOCK — pass:** portable real-Git fixtures and controllable effect barriers exercise production seams.
   - **ARCH-CONSTRAINTS — pass:** batch-only execution, bounded snapshot readers, and process-growth coverage.
   - **ARCH-SECURE — flag:** a branch name substitutes for evidence of absence.
   - **ARCH-ORDER — pass:** explicit transitions, uncertainty stops, and retry reconstruction.
   - **ARCH-FUNERAL — pass:** existing Git artifact lifecycles are reused.
   - **ARCH-EVOLVE — pass:** no persisted-format change introduced.

7. **Plan revision recommendations**

   Append a `## Revisions` entry specifying that initialization observes every local/publication branch tip before applying any evidence-based exemption. Extend the root × scope tests to cover every excluded ref category.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"90652019-eff1-43e7-bb5a-d2f07cdf733d","repository":"github.com/xianxu/ariadne","issue_id":"000543","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/35117b907c7222b3","reviewed_head":"e344c6cec0332b0d8316be96dce73dd04f18280c","slot":"ariadne:3","requested_at":"2026-10-22T06:59:55.296901Z"},"phase":"finished","pid":71430,"started_at":"2026-10-22T06:59:55.393406Z","finished_at":"2026-10-22T07:03:31.601395Z","session_id":"01a16268-c486-7801-a520-277fda24906c","outcome":"returned"}] -->

---

## Re-review — 2026-10-22T00:36:54-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 543 — new repositories start on the legacy issue path: initial setup should create the issue tracker (non-brain repos) |
| repo | ariadne |
| issue file | workshop/issues/000543-new-repositories-start-on-the-legacy-issue-path-initial-setup-should-create-the-issue-tracker-non-brain-repos.md |
| boundary | whole-issue close |
| milestone | — |
| window | a4600f5050712a27c7979cfd035cccd8b5a1ceca..8b7cbe40b2997f5142d19d5d5f3560ad62d43ddb |
| command | sdlc close --issue 543 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T00:36:54-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation satisfies the revised initialization contract. BR-2 is addressed across the documented roots and scopes, including previously excluded branch categories. Regression tests fail when the relevant corrections are removed. No new blocking findings emerged; the recorded pushed-head suite and landing remain separate, unchecked acceptance steps.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      issueinit_inventory.go enforces presence across all three roots, complete publication branch tips, live worktree HEADs, indexes and filesystem entries. Head tests pass; disposable mutations restoring filtered branch observation or active-only scanning fail the reserved-branch and both archive regressions. The approved plan revision explicitly separates init presence from migration importing.
  - id: BR-1
    disposition: addressed
    note: |
      publicationBranchTips refreshes the complete publication heads mapping and pins names/OIDs. Current-remote, failed-refresh, moved-main, pinned-consumer and fetch-scope regressions pass.
```

1. **Strengths**

   - One presence predicate serves snapshot, index and filesystem scans in `cmd/sdlc/issueinit_inventory.go:39`.
   - The opaque initialization model owns phase transitions; failures preserve confirmed identities and stop subsequent effects.
   - Recovery tests exercise real bare-Git publication with deterministic interruption and competing-clone schedules.
   - README, atlas, command help and TL creation guidance document the explicit final setup step and recovery behavior.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed the targeted tracker/init/hint/recovery tests, migration-inventory tests, existing issue-migration tests, TL tests and recovery-package tests. `git diff --check` passed.

   In a disposable pinned-head checkout, restoring filtered observation failed all four local reserved-branch cases. Restricting inventory to the active root failed remote-only archive and untracked-other-worktree archive cases. The disposable checkout was removed; the reviewed checkout remains clean.

6. **Architectural notes**

   - **ARCH-DRY — Pass:** shared publication owner and common presence predicate.
   - **ARCH-PURE — Pass:** deterministic eligibility/transitions separated from Git and filesystem effects.
   - **ARCH-PURPOSE — Pass:** fresh setup reaches tracker-backed issue creation and planning; the inventory rule covers the documented enumeration.
   - **ARCH-MOCK — Pass:** shared effect seam, portable real-Git backends and controlled failure schedules.
   - **ARCH-CONSTRAINTS — Pass:** explicit batch operation; process-growth tests cover unrelated files and distinct branch tips.
   - **ARCH-SECURE — Pass:** uncertain observations refuse; literal paths and pinned identities constrain reads.
   - **ARCH-ORDER — Pass:** production dispatch follows the opaque transition model; interruption and retry tests preserve partial publication.
   - **ARCH-FUNERAL — Pass:** existing artifact lifecycles are reused; no new durable phase files.
   - **ARCH-EVOLVE — Pass:** existing tracker and marker formats remain unchanged.

7. **Plan revision recommendations:** None. The latest revisions accurately describe the presence-only gate and verified recovery exceptions.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"8f7a1f8b-e180-4300-88fb-eeace446502f","repository":"github.com/xianxu/ariadne","issue_id":"000543","recipe":"milestone-review","round":4,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/21a7c551b65eed37","reviewed_head":"8b7cbe40b2997f5142d19d5d5f3560ad62d43ddb","slot":"ariadne:3","requested_at":"2026-10-22T07:32:23.4977Z"},"phase":"finished","pid":64896,"started_at":"2026-10-22T07:32:23.638511Z","finished_at":"2026-10-22T07:36:53.864723Z","session_id":"01a16286-7f62-73a3-87c9-4ebe56b13e8e","outcome":"returned"}] -->
