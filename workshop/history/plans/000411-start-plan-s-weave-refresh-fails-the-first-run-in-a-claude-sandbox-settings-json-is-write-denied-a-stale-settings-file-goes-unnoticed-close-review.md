# Boundary Review — ariadne#411 (whole-issue close)

| field | value |
|-------|-------|
| issue | 411 — start-plan's weave refresh fails the first run in a Claude sandbox (settings.json is write-denied); a stale settings file goes unnoticed |
| repo | ariadne |
| issue file | workshop/issues/000411-start-plan-s-weave-refresh-fails-the-first-run-in-a-claude-sandbox-settings-json-is-write-denied-a-stale-settings-file-goes-unnoticed.md |
| boundary | whole-issue close |
| milestone | — |
| window | 094cd4294c2d7a0a8102e8b291c638a8eb5d45d6..09e1566a82c57068474e738f704a74c849e71b02 |
| command | sdlc close --issue 411 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T16:21:39-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The normal merge path correctly defers settings publication, preserves ownership evidence, and supports operator recovery. However, removing the settings declaration bypasses deferral and lets gate refresh delete the protected file. The plan also misclassifies filesystem ownership code as PURE. Focused tests passed; native sandbox verification was blocked by this review environment.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: protected-output-exclusion
    title: |
      Gate deferral misses retirement of previously managed settings
    detail: |
      cmd/weave/settings.go:20-39 creates DeferredOutput only for an existing MergeSettings action. After compiling settings and removing its manifest declaration, I reproduced weave settings exiting silently and refresh --substrate --defer-settings deleting .claude/settings.json (output: "retired .claude/settings.json"). Under write denial, this retirement can fail refresh again. ARCH-PURPOSE/ARCH-ORDER: protect the settings destination throughout reconciliation, including retirement and other action types; add a regression covering declaration removal, retained inventory/backups, denied mutations, and eventual operator cleanup.
  - id: new
    severity: Critical
    kind: defect
    family: core-concept-purity-classification
    title: |
      The Core concepts table labels filesystem ownership reconciliation PURE
    detail: |
      workshop/plans/000411-protected-settings-plan.md:16 lists Managed output ownership under Pure entity, but ownership.go:200-208 performs publication reclamation and commits filesystem changes. Its regression at deferred_test.go:78-98 uses temporary files and deferredDenyFS. This contradicts the required Core concepts contract and ARCH-PURE. Move ownership reconciliation to INTEGRATION, identify its injected weavefs seam, and append a Revisions entry; retain only genuinely deterministic entities in the PURE table.
  - id: new
    severity: Important
    kind: unverifiable
    family: native-boundary-conformance
    title: |
      Native protected-settings conformance could not execute at the pinned head
    detail: |
      ARIADNE_NATIVE_SANDBOX=1 go test -trimpath ./cmd/sdlc -run '^TestStartPlanProtectedSettingsNativeSandbox$' -count=1 -v -timeout 8m failed in both subtests because sandbox-exec exited 71 with "sandbox_apply: Operation not permitted", before launching the child. This prevents verification of the real OS-denial invariant (ARCH-MOCK). Supply a successful run on a host permitting sandbox-exec at the reviewed head, including denied-write controls and stale/current notice assertions, or enable that capability in the review environment.
```

1. **Strengths**

   - Deferral happens before managed reconciliation, preserving prior identities and displaced backups for the covered merge path.
   - Ownership tests exercise repeated deferral, edited/missing outputs, mutation denial, recovery, and cleanup.
   - The recovery test runs real `refresh --substrate` on an ahead issue branch and verifies branch/HEAD preservation.
   - README and atlas both document the new command, flag, recovery action, and retained rebuilt-binary rerun.

2. **Critical findings**

   - `cmd/weave/settings.go:20`: protect the destination independently of whether a merge action remains. The reproduced retirement bypass violates the gate’s protected-settings contract.
   - `workshop/plans/000411-protected-settings-plan.md:16`: reclassify ownership reconciliation as INTEGRATION. This is a plan correction; it does not require wording-presence tests.

3. **Important findings**

   - `cmd/sdlc/settings_native_test.go:106`: native conformance remains unverifiable here for the environment failure recorded above.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed:

   - Full `cmd/weave/internal/plan` and `settingsx` tests.
   - Focused settings/deferred-refresh tests in `cmd/weave`.
   - SDLC tests matching `Test(StartPlan|RunInstalledWeave|GateSettings)`.

   Existing tests always supply the deferred action explicitly or retain the merge declaration, so they miss its disappearance. The scratch CLI reproduction confirmed deletion and was removed afterward. I did not rerun the full repository suite.

6. **Architectural notes**

   - **ARCH-DRY — pass:** reuses settings composition, comparison, and managed ownership.
   - **ARCH-PURE — flag:** Core concepts classification contradicts filesystem behavior.
   - **ARCH-PURPOSE — flag:** protected-output exclusion does not cover retirement.
   - **ARCH-MOCK — flag:** injected filesystem coverage exists; native conformance needs host evidence.
   - **ARCH-CONSTRAINTS — pass:** observation introduces no dependency installation, builds, generators, or concurrent work.
   - **ARCH-SECURE — pass:** no sandbox bypass or new credential handling; observation errors remain visible.
   - **ARCH-ORDER — flag:** declaration removal escapes the deferred lifecycle.
   - **ARCH-FUNERAL — pass:** existing operator reconciliation provides retirement and backup cleanup.

7. **Plan revision recommendations**

   Append `## Revisions` entries that:

   - Classify managed ownership as INTEGRATION and name its filesystem seam.
   - Enumerate protected destination transitions, including declaration removal and alternate action types, with regressions proving gates preserve evidence until operator reconciliation.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"120f3049-2005-48f1-b59e-d112ebd3d7a8","repository":"github.com/xianxu/ariadne","issue_id":"000411","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/4a02288cdc53df8b","reviewed_head":"09e1566a82c57068474e738f704a74c849e71b02","slot":"ariadne:5","requested_at":"2026-10-16T23:20:10.563667Z"},"phase":"finished","pid":99638,"started_at":"2026-10-16T23:20:10.665585Z","finished_at":"2026-10-16T23:21:38.118439Z","session_id":"01a14704-0fdf-7c00-97ba-2be1c2a6cb2f","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T16:37:27-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 411 — start-plan's weave refresh fails the first run in a Claude sandbox (settings.json is write-denied); a stale settings file goes unnoticed |
| repo | ariadne |
| issue file | workshop/issues/000411-start-plan-s-weave-refresh-fails-the-first-run-in-a-claude-sandbox-settings-json-is-write-denied-a-stale-settings-file-goes-unnoticed.md |
| boundary | whole-issue close |
| milestone | — |
| window | 094cd4294c2d7a0a8102e8b291c638a8eb5d45d6..1813e25dbdb94a2e86e7f500c1950e6d89cf559a |
| command | sdlc close --issue 411 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T16:37:27-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The revised implementation satisfies the issue’s operator-publication alternative and resolves both Critical findings. The native conformance dispute is supported by an inspected host log and executable-source parity with the pinned head. One minor atlas description remains outdated.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      settings.go now always defers the destination, including declaration removal and alternate writers. The regression suite passes in a disposable repository; restoring pre-fix settings.go makes TestDeferredSettingsProtectsRemovedDeclarationAndBackup fail in four states. Ownership tests verify denied mutations, retained evidence/backups, and eventual cleanup.
  - id: BR-2
    disposition: addressed
    note: |
      The plan's Integration table now identifies managed ownership and its weavefs IO seam; the close-review Revisions entry explains the correction. ownership.go confirms this classification. Only the DeferredOutput action value remains in the Pure table.
  - id: BR-3
    disposition: withdrawn
    note: |
      Inspected the supplied aa326d59 native log: stale/current/retired passed with denial controls and notice counts 1/0/1. That source head differs from pinned HEAD only in tracker/review records. My rerun reproduced sandbox_apply: Operation not permitted, confirming the nested-fence restriction; the measured host evidence answers it.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: documentation-behavior-parity
    title: |
      Atlas still describes merge-only settings deferral
    detail: |
      atlas/workflow/sdlc-binary.md:1340 says gate refresh replaces only the Claude merge action. settings.go:23-68 now excludes every action producing the destination and always appends DeferredOutput, including after declaration removal. Update the description to match destination-wide protection.
```

1. **Strengths**

   - Protection precedes publication, displacement, and retirement; prior ownership identities and backups survive repeated deferral.
   - Recovery uses real `refresh --substrate` on an issue branch and verifies branch/HEAD preservation.
   - Settings observation failures remain visible without refusing planning; moved-HEAD rerun safeguards remain intact.
   - README and atlas both document the new commands and recovery flow.

2. **Critical findings:** None remaining.

3. **Important findings:** None remaining.

4. **Minor findings:** Atlas wording correction above.

5. **Test coverage notes**

   - Plan package passed.
   - Focused settings/refresh regressions passed in a disposable repository, removed afterward.
   - BR-1 negative control failed without the fix.
   - All selected non-native SDLC tests passed. Native execution was restricted locally; inspected host evidence covers all three native states.
   - Full repository suite was not rerun; recorded failures remain disclosed and passing merge validation remains required.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Reuses composition, identity observation, and ownership reconciliation.
   - **ARCH-PURE — pass:** Corrected integration classification; deterministic settings logic remains separate from IO.
   - **ARCH-PURPOSE — pass:** Covers declaration removal and alternate output actions.
   - **ARCH-MOCK — pass:** Production seams have filesystem doubles, stateful installed-weave fixtures, and native conformance.
   - **ARCH-CONSTRAINTS — pass:** Read-only checks avoid acquisition, builds, and generators.
   - **ARCH-SECURE — pass:** Existing parsing and containment checks remain; no sandbox bypass introduced.
   - **ARCH-ORDER — pass:** Deferral preserves observed ownership evidence until operator reconciliation.
   - **ARCH-FUNERAL — pass:** Deferred backups retain their existing operator-reconciliation cleanup path.

7. **Plan revision recommendations:** None; the existing BR-1/BR-2 revision matches the implementation.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"cc5b93a5-d103-4dfa-a25a-5ff2fe94247b","repository":"github.com/xianxu/ariadne","issue_id":"000411","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/5ccc6c8ac0afcf68","reviewed_head":"1813e25dbdb94a2e86e7f500c1950e6d89cf559a","slot":"ariadne:5","requested_at":"2026-10-16T23:35:51.933003Z"},"phase":"finished","pid":3240,"started_at":"2026-10-16T23:35:52.147618Z","finished_at":"2026-10-16T23:37:26.977527Z","session_id":"01a14712-6da9-7943-aca5-7095c4fd6046","outcome":"returned"}] -->
