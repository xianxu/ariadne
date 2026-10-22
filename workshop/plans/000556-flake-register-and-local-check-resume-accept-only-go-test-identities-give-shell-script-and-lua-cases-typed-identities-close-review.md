# Boundary Review — ariadne#556 (whole-issue close)

| field | value |
|-------|-------|
| issue | 556 — flake register and local-check resume accept only Go test identities; give shell-script and Lua cases typed identities |
| repo | ariadne |
| issue file | workshop/issues/000556-flake-register-and-local-check-resume-accept-only-go-test-identities-give-shell-script-and-lua-cases-typed-identities.md |
| boundary | whole-issue close |
| milestone | — |
| window | ffd859eb5d7298b0844bc1c6876edad9b7574fd2..337a9cb6f5eb0cfd07430333b1d961f90f58084e |
| command | sdlc close --issue 556 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T14:49:30-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The typed identity, private Lua evidence channel, process ownership, and saved-proof integration are well covered, and all review tests passed. Shipping is blocked by typed registration accepting targeted-rerun evidence despite the plan’s complete-suite requirement, an inaccurate Core concepts inventory, and missing README coverage.

```findings
findings:
  - id: new
    severity: Critical
    family: typed-confirmation-requires-complete-suite
    title: |
      Typed registration accepts a targeted rerun as its confirming suite pass
    detail: |
      cmd/sdlc/teststats.go:256 marks only affected checks as Subset. flakeEvidence includes successful suite-rerun records through isFlakeEvidence, so SameHeadFailPass accepts a matching v4 targeted child as confirmation for script/Lua registration without a later complete suite pass. This contradicts plan contract item 6. Preserve check scope in the shared evidence model and reject targeted confirmations for typed identities while retaining legacy Go behavior; add script and Lua regressions covering registration and Confirmed. ARCH-PURPOSE.
  - id: new
    severity: Critical
    family: core-concept-traceability
    title: |
      The Core concepts table names missing and incorrectly located entities
    detail: |
      workshop/plans/000556-typed-flake-identities-plan.md:35 lists RerunPlan and RerunMissingJSON as modified in rerun.go, but that file is unchanged and RerunPlan is new in rerun_typed.go:12. Line 37 names typedAttemptCapability, which does not exist in the pinned code. Correct the inventory and verification strategies to identify the delivered helpers and their actual PURE/INTEGRATION classification, recording the correction under Revisions. Critical under the required Core concepts cross-check. ARCH-PURPOSE.
  - id: new
    severity: Important
    family: user-surface-documentation
    title: |
      README update is missing for typed identities and runner adoption
    detail: |
      README.md:850 still documents Go-only registration and reruns, and README.md is unchanged in the review range. Add script/Lua identity syntax, flake run --unit, version-3 flake_runners configuration, and the fleet-upgrade/adoption prerequisite, linking the detailed atlas instructions. Help and atlas updates do not satisfy the separate README gate.
```

1. **Strengths**

   - Shared identity parsing preserves opaque Lua labels and exact non-Go coverage.
   - `flake_run.go` separates child diagnostics from authoritative evidence and projects literal argv.
   - The synchronized wrapper-death test checks inherited process groups and prevents retirement before descendant join.
   - Saved-proof tests cover malformed evidence, contract mismatch, ignored failures, immutable logs, and retention compatibility.

2. **Critical findings**

   The confirmation-scope defect and Core concepts contradictions are detailed above. The confirmation fix needs regressions that fail with the current projection; the inventory correction needs a plan revision.

3. **Important findings**

   Update README for the new user-facing surface and adoption sequence.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed during review:

   - `internal/flakes`, `internal/issue`, and `pkg/golint` tests.
   - Focused root tests matching `Typed|FlakeRun|FlakeUnit|LocalChecksV3`.
   - Native Plenary adapter conformance.
   - Private production CLI typed conformance, including omitted-case rejection and crash recovery.
   - Frozen older-reader refusal and legacy-Go resume conformance.

   The confirmation tests cover an explicitly marked subset, but omit the production projection of `suite-rerun`. No full recorded suite was run during this review. The checkout remained unchanged.

6. **Architectural notes for upcoming work**

   - **ARCH-DRY — pass:** shared identity, coverage, and unit-evidence logic serve multiple consumers.
   - **ARCH-PURE — pass:** deterministic parsing/transitions remain separate from execution and storage.
   - **ARCH-PURPOSE — flag:** complete-suite confirmation is not enforced; the entity inventory also needs correction.
   - **ARCH-MOCK — pass:** private stateful fixtures exercise production boundaries, supplemented by native conformance.
   - **ARCH-CONSTRAINTS — pass:** explicit channel, case, argv, and log bounds have focused tests.
   - **ARCH-SECURE — pass:** strict decoding, contract binding, and diagnostic isolation defend the new boundaries.
   - **ARCH-ORDER — pass:** sticky unusable evidence and controlled process-death tests cover relevant ordering.
   - **ARCH-FUNERAL — pass:** scratch cleanup and existing retention paths cover the new artifacts.
   - **ARCH-EVOLVE — pass:** capability barriers, immutable attempt bytes, and nested declaration preservation are tested.

7. **Plan revision recommendations**

   Append a dated `## Revisions` entry correcting the Core concepts paths, statuses, missing helper name, and associated verification strategies. Retain and enforce the documented complete-suite requirement for typed registration.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"0bf5421d-be3c-4afd-b999-a81535afaa9d","repository":"github.com/xianxu/ariadne","issue_id":"000556","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/04d23021471cfcdd","reviewed_head":"337a9cb6f5eb0cfd07430333b1d961f90f58084e","slot":"ariadne:3","requested_at":"2026-10-22T21:45:33.061585Z"},"phase":"finished","pid":21188,"started_at":"2026-10-22T21:45:33.26658Z","finished_at":"2026-10-22T21:49:26.319171Z","session_id":"01a16593-979b-7120-a1b5-9f2499081837","outcome":"returned"}] -->

---

## Re-review — 2026-10-22T15:04:39-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 556 — flake register and local-check resume accept only Go test identities; give shell-script and Lua cases typed identities |
| repo | ariadne |
| issue file | workshop/issues/000556-flake-register-and-local-check-resume-accept-only-go-test-identities-give-shell-script-and-lua-cases-typed-identities.md |
| boundary | whole-issue close |
| milestone | — |
| window | ffd859eb5d7298b0844bc1c6876edad9b7574fd2..36357d9e02a152854a5eab4ed647454e565a96c0 |
| command | sdlc close --issue 556 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T15:04:39-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The typed registration, execution, resume, and compatibility paths passed focused and native conformance checks. All three prior findings are addressed. One Important gap remains: weekly snapshots omit typed log-proof hydration and therefore misclassify confirmed script/Lua flakes.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      teststats.go preserves Targeted scope and evidence.go rejects targeted typed confirmation. TestTypedConfirmationRequiresCompleteSuiteScope passes; removing the scope projection through a temporary overlay makes script and Lua registration, Confirmed, and stats assertions fail.
  - id: BR-2
    disposition: addressed
    note: |
      Plan lines 31–58 now identify the delivered symbols, paths, statuses, and PURE/INTEGRATION seams. RerunPlan/RerunMissingProof exist in rerun_typed.go; schema and IO helpers replace typedAttemptCapability. The correction is recorded under Revisions at line 204.
  - id: BR-3
    disposition: addressed
    note: |
      README.md:880 documents typed syntax, complete-suite confirmation, flake run --unit, version-3 flake_runners, v4 attempts, and fleet/adoption prerequisites, linking the atlas contract. These passages agree with the runner, evidence, and declaration implementation.
findings:
  - id: new
    severity: Important
    kind: defect
    family: typed-evidence-consumer-parity
    title: |
      Weekly snapshots never hydrate typed confirmation evidence
    detail: |
      cmd/sdlc/teststats_cmd.go:155 hydrates typed facts for interactive reporting, but weeklystats_inputs.go:224 still passes raw snapshot.Attempts to the weekly aggregator. Because evidence.go:85 requires TypedFacts, valid v4 script/Lua fail→complete-suite-pass histories always lose confirmation there. A temporary production-reader regression reproduced interactive confirmed=1 versus weekly confirmed=0 and failed_tests.kind=unknown for both kinds. ARCH-PURPOSE: finish the plan's reporting-consumer sweep by applying shared digest-verified hydration at the weekly IO boundary and adding persisted-log regressions for both kinds.
```

1. **Strengths**

   - Typed identities preserve opaque Lua labels and enforce exact non-Go coverage.
   - Runner argv comes from verified declarations; Lua proof uses a private channel.
   - Wrapper-death tests exercise inherited process-group ownership and controlled cleanup.
   - Private CLI conformance validates mixed Go/script/Lua registration, resume, omitted-case refusal, and interrupted recovery.

2. **Critical findings**

   None.

3. **Important findings**

   The weekly reporting gap above needs shared hydration and regression coverage before crossing the boundary.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed flakes and golint packages; focused typed, flake, declaration, resume, and weekly tests; real Plenary producer conformance; private production CLI conformance; and frozen older-reader compatibility checks. BR-1’s mutation test failed as required. The new weekly reproduction failed for both kinds. Temporary overlays were removed; tracked files remain unchanged. A full recorded suite was not run during this review.

6. **Architectural notes**

   | Marker | Result |
   |---|---|
   | ARCH-DRY | Pass: shared identity, coverage, and proof helpers. |
   | ARCH-PURE | Pass: parsing and decisions remain separate from IO. |
   | ARCH-PURPOSE | Flag: weekly reporting remains an unwired evidence consumer. |
   | ARCH-MOCK | Pass: private filesystem/provider fixtures and native conformance. |
   | ARCH-CONSTRAINTS | Pass: bounded channels, cases, argv, and serial retries. |
   | ARCH-SECURE | Pass: typed input validation and declaration-owned execution authority. |
   | ARCH-ORDER | Pass: sticky invalid evidence and controlled process-death coverage. |
   | ARCH-FUNERAL | Pass: scratch cleanup and existing retention ownership are reused. |
   | ARCH-EVOLVE | Pass: justified capability barriers, preserved declaration extras, immutable historical records. |

7. **Plan revision recommendations**

   Append a dated `## Revisions` entry recording the reporting-consumer rule: every confirmation-reporting IO reader must supply shared verified typed facts. Add weekly reader wiring and script/Lua persisted-log parity tests to Task 3’s consumer inventory.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"564304dc-0c55-49ca-99e4-d35065113b41","repository":"github.com/xianxu/ariadne","issue_id":"000556","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/abbd556a87f9b7e6","reviewed_head":"36357d9e02a152854a5eab4ed647454e565a96c0","slot":"ariadne:3","requested_at":"2026-10-22T22:00:40.407763Z"},"phase":"finished","pid":3696,"started_at":"2026-10-22T22:00:40.591624Z","finished_at":"2026-10-22T22:04:38.059389Z","session_id":"01a165a1-6f4a-79c3-a011-84892416d47a","outcome":"returned"}] -->
