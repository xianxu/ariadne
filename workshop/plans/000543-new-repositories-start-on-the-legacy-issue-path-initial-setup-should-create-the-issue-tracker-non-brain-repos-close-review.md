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
