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
