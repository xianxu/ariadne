# Boundary Review — ariadne#624 (whole-issue close)

| field | value |
|-------|-------|
| issue | 624 — fleet inventory --json reports merge intents (producer for pair#848) |
| repo | ariadne |
| issue file | workshop/issues/000624-fleet-inventory-json-reports-merge-intents-producer-for-pair-848.md |
| boundary | whole-issue close |
| milestone | — |
| window | fa73ecc8751c6ab4b65e9dd2ca9f0cdd37a038cd..0ccb7f5266b4a25788dbe4a4f477c36e6ebfadd1 |
| command | sdlc close --issue 624 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-25T04:16:04-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The shared classifier, conservative partial-read handling, and private executable tests provide a solid foundation. Two adapter defects remain: an unusable checkout can prevent observation of an accessible shared store, and cancellation does not reach Git subprocesses. The internal fleet test suite also fails, and README coverage is missing.

Done-when: current — it still covers the delivered wire contract, lifecycle tests, capability, and output bound; terminal removal is clarified by the revisions.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: shared-store-observation-authority
    title: |
      An unusable member checkout can hide an accessible shared intent store
    detail: |
      cmd/sdlc/fleet_mergeintents.go:54 selects the lexicographically first TreePath as the sole observation directory. Collection retains existing registered worktree directories even when Git facts fail, so a directory whose .git metadata was removed can be selected ahead of healthy members; landingLineDir then fails and every row becomes unreadable despite the known common-directory store remaining accessible. Observe through RepoIdentity, or verify and select a usable checkout without rescanning the store. Add a regression with an existing prunable directory lacking .git, ordered before a healthy checkout containing an intent.
  - id: new
    severity: Critical
    kind: defect
    family: cancellation-reaches-external-effects
    title: |
      Caller cancellation cannot interrupt merge-intent Git probes
    detail: |
      cmd/sdlc/fleet_mergeintents.go:35 passes execGitRunner{} into the shared observer. Its Git methods use exec.Command without context, and the observer probes every intent before checking ctx again. A stalled Git process therefore prevents cancellation or the declared deadline from completing. ARCH-CONSTRAINTS and ARCH-ORDER: thread a context-aware runner through the shared observation seam, stop further probes on cancellation, and test cancellation while a probe is blocked.
  - id: new
    severity: Important
    kind: defect
    family: decoder-contract-test-consistency
    title: |
      The additive decoder change leaves the existing package suite failing
    detail: |
      cmd/sdlc/internal/fleet/types.go:525 now accepts unknown inventory fields, but json_test.go:410 calls assertJSONRoundTrip, whose lines 426–428 require their rejection. Running go test ./cmd/sdlc/internal/fleet ./cmd/sdlc/internal/landingline -count=1 failed both FuzzJSONContractAlgebra seeds. Update the oracle to distinguish additive inventory/row decoding from intentionally strict types, retaining duplicate-key and known-type rejection checks, then rerun the package suite.
  - id: new
    severity: Important
    kind: defect
    family: user-surface-documentation
    title: |
      README update appears missing for the inventory JSON contract
    detail: |
      README.md:1121 documents fleet inventory --json but is unchanged in this range. The command now exposes merge-intent capability/read-quality fields and refuses output above 1 MiB, which affects consumers and usage. Extend the Fleet queries passage with the capability-versus-empty distinction, read-quality behavior, output bound, and a link to the detailed contract. Atlas and command help already document these changes.
```

1. **Strengths**

   - Inventory, state, and landing status share one report type and classifier.
   - Branch-probe failures retain earlier observations and suppress executable guidance.
   - Private executable conformance covers real locks, dependencies, pause notes, corruption, terminal removal, and unchanged store bytes.
   - JSON encoding checks the exact byte bound before publishing stdout.

2. **Critical findings**

   The first two findings above require implementation fixes and regression tests.

3. **Important findings**

   Repair the existing test oracle and add the README contract summary.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Passed: `go test ./cmd/sdlc -run '^(TestFleet|TestMergeIntent)' -count=1` — 31.516s.
   - Passed: internal landingline package — 2.753s.
   - Failed: internal fleet package — both existing JSON algebra fuzz seeds.
   - `git diff --check` passed; the checkout remains unchanged.
   - Current cancellation tests cover cancellation around an injected observer, not interruption of production Git execution.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared report, classifier, pause-note formatting, and diagnostic ordering.
   - **ARCH-PURE — pass:** placement and validation remain deterministic; IO is outside placement.
   - **ARCH-PURPOSE — flag:** shared-store availability depends on an arbitrary member checkout.
   - **ARCH-MOCK — pass:** stateful observation fake and private real-store/executable conformance exercise the shared seam.
   - **ARCH-CONSTRAINTS — flag:** output and counted growth are bounded, but cancellation does not reach Git.
   - **ARCH-SECURE — pass:** validated branches construct commands; corrupt roots remain diagnostics.
   - **ARCH-ORDER — flag:** subprocess lifetime is not bounded by caller cancellation.
   - **ARCH-FUNERAL — pass:** no new durable runtime artifacts; fixtures use temporary storage.
   - **ARCH-EVOLVE — pass:** intent storage remains preserving v1; inventory additions remain schema 1. The test oracle needs alignment.

   Core-concept entities exist at their stated paths, and the modified entities appear in the pinned diff.

7. **Plan revision recommendations**

   Append a `## Revisions` entry specifying that shared-store observation must remain available when a member checkout is unusable, with the corresponding fixture. Keep the existing cancellation promise and add verification that cancellation interrupts an active production probe.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"72c646a8-e951-43a2-8a4d-6593cbaea176","repository":"github.com/xianxu/ariadne","issue_id":"000624","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/3b3840f017369926","reviewed_head":"0ccb7f5266b4a25788dbe4a4f477c36e6ebfadd1","slot":"ariadne:3","requested_at":"2026-10-25T11:13:32.932778Z"},"phase":"finished","pid":48750,"started_at":"2026-10-25T11:13:33.036311Z","finished_at":"2026-10-25T11:16:02.094119Z","session_id":"01a172c4-0cb8-7b12-a4b5-675c232e6466","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T04:28:25-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 624 — fleet inventory --json reports merge intents (producer for pair#848) |
| repo | ariadne |
| issue file | workshop/issues/000624-fleet-inventory-json-reports-merge-intents-producer-for-pair-848.md |
| boundary | whole-issue close |
| milestone | — |
| window | fa73ecc8751c6ab4b65e9dd2ca9f0cdd37a038cd..7f2957f419edd82be881e442d434ebfbc54da943 |
| command | sdlc close --issue 624 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-25T04:28:25-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The four prior findings are addressed, with negative controls confirming that the behavioral regressions fail without their fixes. Shared-store placement, read-quality reporting, compatibility, output bounds, and documentation are well covered. One additional cancellation defect blocks shipping: interrupting the actual CLI leaves its isolated Git probe running.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Collection observes through RepoIdentity. TestFleetMergeIntentsUsesSharedStoreDespiteBrokenMember passes at head and fails with the previous adapter substituted through a temporary build overlay.
  - id: BR-2
    disposition: addressed
    note: |
      Injected context cancellation now interrupts active Git probes and prevents later probes. TestFleetMergeIntentsCancelsActiveGitProbe passes at head and fails with the previous adapter; the separate CLI signal-entry gap is reported below.
  - id: BR-3
    disposition: addressed
    note: |
      The algebra oracle now distinguishes additive Inventory/TreeRow decoding from strict types and retains duplicate rejection. Both full packages pass; substituting the previous oracle reproduces both failing fuzz seeds.
  - id: BR-4
    disposition: addressed
    note: |
      README.md:1138–1143 now documents capability versus complete empty, incomplete read quality, the 1 MiB refusal, and the detailed contract. These statements match placement, wire validation, and writeFleetInventoryJSON.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: cancellation-reaches-external-effects
    title: |
      CLI interruption leaves the new isolated Git probe running
    detail: |
      cmd/sdlc/fleet_mergeintents.go:35 uses boundedLandingGit, which puts Git in its own process group, but cmd/sdlc/main.go:107–125 excludes fleet inventory from signal-owned cancellation. In a private fixture using the pinned-head executable, I blocked the branch probe, sent SIGTERM to inventory, and observed CLI exit -15 while the Git child remained alive. The existing regression cancels an injected context and cannot detect this production entry-path failure. ARCH-CONSTRAINTS and ARCH-ORDER: this is the 2nd finding in family cancellation-reaches-external-effects. Do not fix only this instance: establish and enumerate the rule that every affected CLI entry point using isolated subprocesses must propagate supported interruption into its context and terminate/reap those subprocesses before exiting. Include fleet inventory and add native SIGINT/SIGTERM regressions proving child cleanup and absence of later probes.
```

Done-when: current — its state/lifecycle, quality, shared-store, legacy, capability, and bounded-output criteria still cover the revised producer contract.

1. **Strengths**

   - One shared report/classifier serves inventory and existing status surfaces; the alias avoids parallel command construction.
   - Pure placement uses exact common-directory/root association, rejects conflicts, and preserves valid observations alongside diagnostics.
   - Private executable conformance exercises real locks, dependency stores, pause notes, terminal removal, and unchanged store/queue contents.
   - README and atlas updates explain the new consumer contract.

2. **Critical findings**

   - **CLI signal cancellation:** see the finding above, anchored at `cmd/sdlc/fleet_mergeintents.go:35` and `cmd/sdlc/main.go:107`. Connect native interruption to subprocess cleanup and prove it through the executable.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed:

   - Full `internal/fleet` and `internal/landingline` package tests.
   - Focused merge-intent command tests, including executable conformance.
   - Race-enabled `TestFleet`, `TestReportMergeIntent`, and `TestMergeIntent` tests.
   - Pinned-range `git diff --check`.

   Temporary build overlays reproduced BR-1, BR-2, and both BR-3 seed failures without changing the checkout. Native CLI signal cleanup remains uncovered by committed tests and failed my private executable check. All review-created processes and temporary fixtures were cleaned up. A full-module suite was not run.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared classification, command construction, and diagnostic sorting.
   - **ARCH-PURE — pass:** placement and report classification remain deterministic; observation uses an injected boundary. Core-concept locations match the implementation.
   - **ARCH-PURPOSE — pass:** linked stores, dependency clones, ambiguous roots, and incomplete evidence are represented.
   - **ARCH-MOCK — pass:** stateful observation doubles share the production seam; private real-store/executable conformance supplements them.
   - **ARCH-CONSTRAINTS — flag:** native interruption bypasses cancellation despite the declared deadline/cancellation envelope.
   - **ARCH-SECURE — pass:** validated records supply commands; corrupt inputs produce inspection evidence.
   - **ARCH-ORDER — flag:** isolated subprocess lifetime can exceed the CLI’s lifetime.
   - **ARCH-FUNERAL — pass for durable artifacts:** no new runtime file family; existing intent cleanup remains authoritative. Process lifetime is flagged above.
   - **ARCH-EVOLVE — pass:** persisted v1 intents retain preserving decoding; inventory additions are observation-only and keep duplicate/authority validation.

7. **Plan revision recommendations**

   Append a `## Revisions` entry defining supported CLI interruption and subprocess ownership, enumerating affected entry points, and requiring native signal cleanup tests. Amend the claim that the command returns with no background work once that invariant is enforced.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"3ff7f6d2-86f0-4ca4-9ceb-bd85ec543882","repository":"github.com/xianxu/ariadne","issue_id":"000624","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/e8e84a00460838eb","reviewed_head":"7f2957f419edd82be881e442d434ebfbc54da943","slot":"ariadne:3","requested_at":"2026-10-25T11:24:27.007581Z"},"phase":"finished","pid":16612,"started_at":"2026-10-25T11:24:27.109345Z","finished_at":"2026-10-25T11:28:24.208577Z","session_id":"01a172ce-079d-7470-ad92-41819f4aa098","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T04:43:12-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 624 — fleet inventory --json reports merge intents (producer for pair#848) |
| repo | ariadne |
| issue file | workshop/issues/000624-fleet-inventory-json-reports-merge-intents-producer-for-pair-848.md |
| boundary | whole-issue close |
| milestone | — |
| window | fa73ecc8751c6ab4b65e9dd2ca9f0cdd37a038cd..5724aaba9338a40537db4841a9e11e762fc73bd8 |
| command | sdlc close --issue 624 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-25T04:43:12-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation delivers the revised producer contract, and BR-5 is addressed with meaningful production-path regression coverage. Shared-store grouping, exact-root placement, explicit uncertainty, additive capability reporting and bounded JSON publication match the issue and plan. No new blocking findings.

Done-when: current — its state, quality, lifecycle, compatibility, bounded-output and capability requirements still cover the revised implementation, including terminal removal as complete empty evidence.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Common-directory observation avoids reliance on member usability; TestFleetMergeIntentsUsesSharedStoreDespiteBrokenMember passed.
  - id: BR-2
    disposition: addressed
    note: |
      The production observer uses context-bound Git; TestFleetMergeIntentsCancelsActiveGitProbe passed and verifies no subsequent branch probe.
  - id: BR-3
    disposition: addressed
    note: |
      The algebra oracle distinguishes additive inventory/row decoding from strict policy decoding while retaining duplicate-key rejection; the complete internal/fleet suite passed.
  - id: BR-4
    disposition: addressed
    note: |
      README.md now documents capability, complete-empty versus unavailable evidence and the output limit; these agree with fleet_mergeintents.go and the expanded atlas contract.
  - id: BR-5
    disposition: addressed
    note: |
      main.go gives fleet inventory signal-owned cancellation, and fleet.go bounds discovery and collection Git alongside intent probes. All six native phase/signal cases passed. Removing both changes in a disposable pinned-head copy made all six fail with surviving Git and pipe-holding children; the copy was removed.
```

1. **Strengths**

   - [fleet_mergeintents.go](/Volumes/TestScratch/review-checkouts/c48da523eebf83f1/cmd/sdlc/fleet_mergeintents.go:52) groups reads by physical common directory and keeps placement separate from IO.
   - [mergeintent.go](/Volumes/TestScratch/review-checkouts/c48da523eebf83f1/cmd/sdlc/mergeintent.go:140) preserves earlier observations and represents failed branch probes as uncertainty without executable guidance.
   - [fleet_mergeintents_process_test.go](/Volumes/TestScratch/review-checkouts/c48da523eebf83f1/cmd/sdlc/fleet_mergeintents_process_test.go:21) exercises the public executable with positive readiness, both supported signals, all three Git phases and pipe-holding descendants.
   - README, atlas and help explain the capability, identity, quality, terminal-removal and publication contracts consistently.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed:

   - Complete `internal/fleet` suite: 32.348s.
   - Complete `internal/landingline` suite: 13.108s.
   - Command tests matching `TestFleetMergeIntents` and `TestMergeIntent`: 48.372s.
   - Pinned-range `git diff --check`.

   The fix-removed native regression failed all six cases as expected. Stateful lifecycle, private executable/store conformance, legacy fixtures, malformed wire data, physical deduplication and growth controls were inspected. The full module suite was not rerun during this review.

6. **Architectural notes for upcoming work**

   - **ARCH-DRY — pass:** report fields, classification and re-entry construction share one owner; diagnostic ordering reuses a shared helper.
   - **ARCH-PURE — pass:** placement and validation remain deterministic; store, Git and queue access stay in the integration shell. Core-concept entities exist at their documented locations.
   - **ARCH-PURPOSE — pass:** linked worktrees and dependency stores participate; failed discovery and unassignable records remain visible.
   - **ARCH-MOCK — pass:** stateful observer tests use the production injection seam, supplemented by real Git, lock and executable conformance.
   - **ARCH-CONSTRAINTS — pass:** counted growth controls exercise shared-store scaling; output is checked against 1 MiB before publication; Git execution is bounded.
   - **ARCH-SECURE — pass:** persisted records pass through the existing strict codec; invalid roots and failed probes produce inspection evidence.
   - **ARCH-ORDER — pass:** native cancellation reaches active subprocess groups, joins their execution and prevents later probes. Inventory retains no state between calls.
   - **ARCH-FUNERAL — pass:** no new durable runtime artifacts; existing intent removal and retention remain authoritative, and fixtures have scoped cleanup.
   - **ARCH-EVOLVE — pass:** v1 storage remains preserving and unchanged; inventory additions retain schema 1 and distinguish unsupported producers. Inventory decoding is not a persisted rewrite authority.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"0049583e-d984-453b-9543-80d911791e10","repository":"github.com/xianxu/ariadne","issue_id":"000624","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/c48da523eebf83f1","reviewed_head":"5724aaba9338a40537db4841a9e11e762fc73bd8","slot":"ariadne:3","requested_at":"2026-10-25T11:40:31.93046Z"},"phase":"finished","pid":59590,"started_at":"2026-10-25T11:40:32.091217Z","finished_at":"2026-10-25T11:43:10.088915Z","session_id":"01a172dc-c1b0-7593-8a0e-ff2e60b1a081","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T05:42:23-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 624 — fleet inventory --json reports merge intents (producer for pair#848) |
| repo | ariadne |
| issue file | workshop/issues/000624-fleet-inventory-json-reports-merge-intents-producer-for-pair-848.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4775d79248fa98a6733e01660e65a16eafbadd45..9229a4df3ed0f94b78810614b5b37cd6858dbc51 |
| command | sdlc close --issue 624 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-25T05:42:23-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The interdiff correctly removes the context-free reader wrapper and threads existing caller contexts through fleet, state, and landing status. The revision records this correction without expanding scope. Focused tests pass, and no new defects were found.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Shared-store authority remains intact; TestFleetMergeIntentsUsesSharedStoreDespiteBrokenMember passed.
  - id: BR-2
    disposition: addressed
    note: |
      Fleet retains its bounded context-aware runner; the active-probe cancellation regression passed.
  - id: BR-3
    disposition: addressed
    note: |
      The complete internal/fleet suite passed, including the decoder algebra controls.
  - id: BR-4
    disposition: addressed
    note: |
      README.md:1138 documents capability, complete-empty evidence, incomplete quality and the output bound, consistent with the implementation.
  - id: BR-5
    disposition: addressed
    note: |
      Native SIGINT/SIGTERM coverage across all three Git phases passed; the interdiff preserves signal ownership and bounded runners.
```

1. **Strengths**

   - `mergeintent.go:132–136` consolidates observation behind context-taking entry points.
   - `landingcmd.go:75`, `landingcmd.go:83`, and `state.go:264` pass their existing contexts explicitly.
   - The existing context guard detects the removed wrapper: running it against disposable pinned-base sources failed at `context.Background()`.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Focused context, status, fleet merge-intent and merge-intent tests passed: **32.047s**, including native-signal and private executable conformance.
   - Complete `internal/fleet` suite passed: **11.745s**.
   - Pinned-range `git diff --check` passed.
   - The negative-control source copy was removed; the checkout remains unchanged.
   - The full module suite was not rerun during this narrow review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** removes the redundant compatibility wrapper.
   - **ARCH-PURE — pass:** context threading remains in the integration shell; pure placement is unchanged.
   - **ARCH-PURPOSE — pass:** all shared-reader callers now supply their owned contexts.
   - **ARCH-EVOLVE — pass:** no persisted formats or rewrite paths change.

   This window introduces no new user-facing or architectural surface requiring README or atlas updates.

7. **Plan revision recommendations:** None. The entry at `workshop/plans/000624-fleet-merge-intent-inventory-plan.md:173` accurately records the correction.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"2cc7972d-7d6d-45e6-94ce-60498f83ac50","repository":"github.com/xianxu/ariadne","issue_id":"000624","recipe":"fix-review","round":4,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/07e72e00dde83d4c","reviewed_head":"9229a4df3ed0f94b78810614b5b37cd6858dbc51","slot":"ariadne:3","requested_at":"2026-10-25T12:40:10.6412Z"},"phase":"finished","pid":24296,"started_at":"2026-10-25T12:40:10.747812Z","finished_at":"2026-10-25T12:42:19.180117Z","session_id":"01a17313-5c34-7831-9b27-d03cc1ff4cfe","outcome":"returned"}] -->
