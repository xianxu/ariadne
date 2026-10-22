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
