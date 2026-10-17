# Boundary Review — ariadne#417 (whole-issue close)

| field | value |
|-------|-------|
| issue | 417 — sdlc stage --build skips sentinel binaries (bin/sdlc stays at main) and doesn't verify staged revisions |
| repo | ariadne |
| issue file | workshop/issues/000417-sdlc-stage-build-skips-sentinel-binaries-bin-sdlc-stays-at-main-and-doesn-t-verify-staged-revisions.md |
| boundary | whole-issue close |
| milestone | — |
| window | f5927dfedbfb2e6dbc7d8c91ac56085732181b7b..8ea8f99e16ae3d9f3ae4117f715de9909e1da5fb |
| command | sdlc close --issue 417 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T20:54:17-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The shared build path covers the main staging and clear-retry scenarios, and the stage tests pass. However, verification still accepts some unverifiable binaries, and sentinel declarations can be interpreted as Make options rather than targets. The external-process testing contract and README gate also need completion.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: fail-closed-build-verification
    title: |
      Missing or corrupt build evidence can pass verification
    detail: |
      cmd/sdlc/stagebuild.go:64-75 skips errors classified by notGoBinary, but Go's buildinfo reader also returns "not a Go executable" for corrupt build-info headers. A corrupt non-sentinel Go binary can therefore disappear from verification and stage succeeds. Also, buildRevision defaults missing or invalid vcs.modified values to false, accepting an unproven clean stamp despite the plan requiring vcs.modified == false. Distinguish recognizable Go candidates with unreadable metadata from genuine non-Go files, require an explicit valid clean stamp, and add regressions for both cases.
  - id: new
    severity: Important
    kind: defect
    family: target-declaration-validation
    title: |
      Sentinel declarations accept Make options and assignments
    detail: |
      cmd/sdlc/stagebuild.go:125-129 accepts --version, -n, and NAME=value as single tokens, then runStageMake passes them directly to make. For example, --version exits successfully without running an owner target; existing matching output can let verification pass afterward. Reject option and assignment syntax, define the supported target grammar, and test malformed and duplicate declarations (ARCH-SECURE).
  - id: new
    severity: Important
    kind: defect
    family: external-dependency-test-seam
    title: |
      Owner-target execution lacks an injected stateful fake and conformance check
    detail: |
      cmd/sdlc/stagebuild.go:91-97 invokes real make directly, and stagebuild_test.go exercises real Makefiles while replacing only readBuildInfo. No injectable Make runner, stateful fake, or live conformance check covers the expanded generic-build → owner-build → verification interaction. Add the shared production/test boundary and test execution order, failure propagation, and retries against its persisted output state (ARCH-MOCK).
  - id: new
    severity: Important
    kind: defect
    family: user-facing-docs-gate
    title: |
      README update is missing for changed stage behavior and sentinel syntax
    detail: |
      The range changes stage --build to refuse unverifiable revisions, makes stage --clear rebuild and verify binaries, and introduces # stage-build: <target>. Atlas and help text explain these changes, but README.md is unchanged. Add concise usage and recovery guidance, linking to the updated atlas documentation.
```

1. **Strengths**

   - Stage and clear reuse `stageBuild`, keeping their revision checks aligned (`cmd/sdlc/stage.go:163`).
   - Sentinel outputs match by package path, supporting differently named binaries (`cmd/sdlc/stagebuild.go:49`, `:78`).
   - Clear failure reports that the checkout has already returned to rest and supplies a retry command; the retry test exercises that path (`cmd/sdlc/stage_test.go:192`).
   - Atlas and embedded help accurately describe the new refusal and rebuild behavior.

2. **Critical findings**

   - `cmd/sdlc/stagebuild.go:64`: fail closed for recognizable Go binaries with corrupt metadata.
   - `cmd/sdlc/stagebuild.go:75`: require explicit `vcs.modified=false`; absence or malformed values must remain unknown rather than become clean.

3. **Important findings**

   - `cmd/sdlc/stagebuild.go:125`: validate Make target syntax before execution.
   - `cmd/sdlc/stagebuild.go:91`: introduce the external-process seam and stateful fake required by ARCH-MOCK.
   - `README.md`: document the changed commands and sentinel directive.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Independently passed: `go test ./cmd/sdlc -run '^TestStage' -count=1` (48.328s).
   - Pinned-range `git diff --check` passed.
   - Missing cases include actual corrupt build-info classification, missing/invalid modified stamps, and malformed/duplicate target declarations. Successful fixtures currently omit `vcs.modified`, masking that contract gap.
   - The full suite and live slot check were not independently rerun. There are no prior findings to dispose of.

6. **Architectural notes**

   - **ARCH-DRY — pass:** build ownership stays in Make targets; stage and clear share verification.
   - **ARCH-PURE — pass:** revision extraction is pure; process and filesystem operations remain in the build orchestration boundary.
   - **ARCH-PURPOSE — flag:** corrupt or incomplete evidence can evade the promised verification invariant.
   - **ARCH-MOCK — flag:** expanded Make interactions lack the required fake and conformance coverage.
   - **ARCH-CONSTRAINTS — pass:** target enumeration is finite and execution is sequential.
   - **ARCH-SECURE — flag:** malformed metadata becomes trusted clean state, and marker tokens can become Make options.
   - **ARCH-ORDER — pass:** the tested failed-clear path preserves the restored checkout and retries the remaining rebuild.
   - **ARCH-FUNERAL — pass:** no new durable artifact family; existing build outputs are superseded in place.

7. **Plan revision recommendations**

   Add `## Revisions` entries specifying explicit clean-stamp validation, corrupt-candidate handling, the accepted Make-target grammar, and the injected Make seam/fake/conformance strategy. Add PURE/INTEGRATION classifications to the concepts table. Keep the live check pending until its documented slot grant and execution.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"871fe3d9-74ce-4300-80dc-9877c0106cba","repository":"github.com/xianxu/ariadne","issue_id":"000417","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/aa0ceefdb84dd3c7","reviewed_head":"8ea8f99e16ae3d9f3ae4117f715de9909e1da5fb","slot":"ariadne:1","requested_at":"2026-10-17T03:52:36.775556Z"},"phase":"finished","pid":68765,"started_at":"2026-10-17T03:52:37.023177Z","finished_at":"2026-10-17T03:54:17.00681Z","session_id":"01a147fd-7de4-7b51-8b0e-4aba4e831333","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T21:01:01-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 417 — sdlc stage --build skips sentinel binaries (bin/sdlc stays at main) and doesn't verify staged revisions |
| repo | ariadne |
| issue file | workshop/issues/000417-sdlc-stage-build-skips-sentinel-binaries-bin-sdlc-stays-at-main-and-doesn-t-verify-staged-revisions.md |
| boundary | whole-issue close |
| milestone | — |
| window | f5927dfedbfb2e6dbc7d8c91ac56085732181b7b..d46d47bfcf07ad9faffc142091bf1e01582b37b7 |
| command | sdlc close --issue 417 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T21:01:01-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The targeted tests pass, and three prior findings are addressed. BR-1 remains open: corrupt XCOFF Go binaries still bypass verification. The plan also labels a filesystem-reading parser PURE, contradicting the required concepts contract.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      Clean-stamp validation is fixed, and removing either new guard makes its regression fail. However, cmd/sdlc/stagebuild.go:114-116 omits XCOFF and Plan 9 formats recognized by debug/buildinfo. Independently cross-built an AIX/ppc64 Go executable, corrupted its build-info magic, and exercised production verification through a temporary test overlay: readBuildInfo returned "not a Go executable", nativeExecutable returned false, and verification succeeded. Enumerate every supported executable format and its variants, fail closed across that enumeration, and add regressions (ARCH-PURPOSE, ARCH-SECURE).
  - id: BR-2
    disposition: addressed
    note: |
      stagebuild.go:154 and :162 reject options and assignments through an explicit target grammar. The malformed and duplicate declaration tests pass; removing grammar validation makes the option, dry-run and assignment regressions fail.
  - id: BR-3
    disposition: addressed
    note: |
      Production and tests share stageBuildWithMake. The fake persists stale output through owner failure and verifies retry order; the real-Make fixture exercises owner execution and renamed outputs. Removing the owner-target call makes the failure/retry regression fail.
  - id: BR-4
    disposition: addressed
    note: |
      README.md:17-24 now documents sentinel targets, verification refusal, clear rebuilding, and retry recovery, linking to the atlas. These statements match stage.go:135-176 and stagebuild.go:30-90.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: pure-core-io-separation
    title: |
      Sentinel parser contradicts the plan's PURE classification
    detail: |
      The plan's concepts table at workshop/plans/000417-sdlc-stage-build-skips-sentinel-binaries-bin-sdlc-stays-at-main-and-doesn-t-verify-staged-revisions-plan.md:13 labels the sentinel parser PURE, but sentinelBuildTarget at cmd/sdlc/stagebuild.go:142-145 reads the filesystem, and its tests create temporary files. Extract parsing into a deterministic text-to-target function with direct tests and retain file reading in the INTEGRATION shell, or classify the existing entity INTEGRATION. Add a ## Revisions entry correcting the contract (ARCH-PURE).
```

1. **Strengths**

   - Stage and clear share one build-and-verification path.
   - Package-path matching supports sentinel outputs whose names differ from their source directories.
   - Clear’s regression verifies partial progress: the checkout returns to rest, rebuilding fails, and retry repairs the binaries.
   - README, atlas, and command help describe the changed workflow.

2. **Critical findings**

   - **BR-1 remains open**, `cmd/sdlc/stagebuild.go:114`: complete executable-format recognition rather than handling only ELF, PE, and selected Mach-O headers. The independently reproduced XCOFF case passes incorrectly.
   - **PURE contract mismatch**, `cmd/sdlc/stagebuild.go:142` and plan line 13: separate parsing from file reading or correct the classification.

3. **Important findings**

   None newly raised.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Passed: `go test ./cmd/sdlc -run '^(TestStage|TestSentinel)' -count=1` — 24.732s.
   - Passed: pinned-range `git diff --check`.
   - Temporary overlays confirmed failures when the native-file guard, clean-stamp requirement, target validation, or owner-target call was removed.
   - The additional corrupt-XCOFF regression failed against the pinned implementation, confirming BR-1 remains reachable.
   - Full-suite and live-slot checks were not independently rerun. Temporary review artifacts were removed; the checkout remained unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared verification and existing owner targets.
   - **ARCH-PURE — flag:** parser classification contradicts filesystem IO.
   - **ARCH-PURPOSE — flag:** verification still misses members of the corrupt-executable class.
   - **ARCH-MOCK — pass:** shared Make seam, persisted fake output, failure/retry coverage, real-tool fixture.
   - **ARCH-CONSTRAINTS — pass:** finite target enumeration and sequential execution.
   - **ARCH-SECURE — flag:** omitted executable formats turn unreadable provenance into successful verification.
   - **ARCH-ORDER — pass:** clear preserves completed checkout restoration and retries rebuilding.
   - **ARCH-FUNERAL — pass:** build outputs supersede existing files; no new production artifact family.

7. **Plan revision recommendations**

   Append `## Revisions` entries defining the complete executable-format enumeration and correcting the sentinel parser’s PURE/INTEGRATION boundary. Keep the documented live check pending until it is performed.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"1b59809e-d3e2-4e7f-8c01-97b9561e5bf3","repository":"github.com/xianxu/ariadne","issue_id":"000417","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/53d1649bb259f34b","reviewed_head":"d46d47bfcf07ad9faffc142091bf1e01582b37b7","slot":"ariadne:1","requested_at":"2026-10-17T03:58:47.184689Z"},"phase":"finished","pid":53705,"started_at":"2026-10-17T03:58:47.407823Z","finished_at":"2026-10-17T04:01:00.463466Z","session_id":"01a14803-2473-7312-aef6-9e5783335943","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T21:05:23-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 417 — sdlc stage --build skips sentinel binaries (bin/sdlc stays at main) and doesn't verify staged revisions |
| repo | ariadne |
| issue file | workshop/issues/000417-sdlc-stage-build-skips-sentinel-binaries-bin-sdlc-stays-at-main-and-doesn-t-verify-staged-revisions.md |
| boundary | whole-issue close |
| milestone | — |
| window | f5927dfedbfb2e6dbc7d8c91ac56085732181b7b..a7f66f308e6a2e5348d6d55929844f39fe935043 |
| command | sdlc close --issue 417 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T21:05:23-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range satisfies the stage/clear build-verification contract. Both open findings are addressed, including demonstrated regression failures when the behavioral fixes are removed. No new blocking findings emerged. The documented full-suite rerun and granted live-slot check remain separate verification steps before landing.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      stagebuild.go:76 rejects unreadable native executables, and :85 requires an explicit vcs.modified=false. Targeted tests pass. In a disposable pinned-head tree, restoring the previous format guard made all four new cross-format cases fail; removing the explicit-clean requirement made all three missing/invalid-stamp cases fail. The disposable tree was removed.
  - id: BR-5
    disposition: addressed
    note: |
      The pinned plan changes the sentinel reader/parser classification from PURE to INTEGRATION at line 13 and records the correction under Revisions at line 55. This accurately describes sentinelBuildTarget's filesystem read at stagebuild.go:156; no executable behavior changed.
  - id: BR-2
    disposition: addressed
    note: |
      The target grammar at stagebuild.go:175 rejects options and assignments; the existing regression cases passed.
  - id: BR-3
    disposition: addressed
    note: |
      Production and tests share stageBuildWithMake. The stateful owner-failure/retry fixture and real-Make sequencing fixtures passed.
  - id: BR-4
    disposition: addressed
    note: |
      README.md:17-24 documents owner targets, verification refusals, clear rebuilding, and retry. These statements match the implementation; atlas/workflow/workspace-branching.md:100-109 describes the same flow.
```

1. **Strengths**

   - Stage and clear reuse one build-and-verify path, avoiding divergent revision checks (`cmd/sdlc/stage.go:179`).
   - Sentinel outputs match build-info package paths, supporting renamed binaries (`cmd/sdlc/stagebuild.go:59,88-92`).
   - Failed clear rebuilding leaves the restored checkout retryable; the integration test verifies recovery from the resting branch (`cmd/sdlc/stage_test.go:192`).

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Passed `go test ./cmd/sdlc -run 'Test(Stage|Sentinel)' -count=1`.
   - Passed pinned-range `git diff --check`.
   - Mutation checks confirmed that cross-format rejection and explicit clean-stamp regressions fail without their fixes.
   - Full suite and live-slot checks were not rerun during this review. The plan explicitly leaves those verification steps pending.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared verifier and existing owner build recipes.
   - **ARCH-PURE — pass:** stamp extraction remains deterministic; filesystem-reading entities are accurately classified INTEGRATION.
   - **ARCH-PURPOSE — pass:** sentinel builds, renamed outputs, revision refusal, and clear recovery are delivered.
   - **ARCH-MOCK — pass:** injected Make boundary, stateful retry fake, and real-Make fixtures exercise the production sequence.
   - **ARCH-CONSTRAINTS — pass:** sequential execution over the finite command set; no added concurrency or fan-out.
   - **ARCH-SECURE — pass:** declarations use restricted argv values; invalid stamps and recognized corrupt containers refuse.
   - **ARCH-ORDER — pass:** switch/build ordering preserves actual checkout progress, and already-resting retries reconcile failed rebuilding.
   - **ARCH-FUNERAL — pass:** no new durable artifact family; existing binary outputs and stage-record lifecycle remain in use.

7. **Plan revision recommendations:** None. The existing revisions accurately record the corrections and deferred live check.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"cc596950-4998-4c2e-a750-66520c7caf77","repository":"github.com/xianxu/ariadne","issue_id":"000417","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/a45d4bf521a8f985","reviewed_head":"a7f66f308e6a2e5348d6d55929844f39fe935043","slot":"ariadne:1","requested_at":"2026-10-17T04:03:29.865435Z"},"phase":"finished","pid":36910,"started_at":"2026-10-17T04:03:29.966868Z","finished_at":"2026-10-17T04:05:22.660275Z","session_id":"01a14807-7312-7ab2-8500-9c63e74c981e","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T22:36:46-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 417 — sdlc stage --build skips sentinel binaries (bin/sdlc stays at main) and doesn't verify staged revisions |
| repo | ariadne |
| issue file | workshop/issues/000417-sdlc-stage-build-skips-sentinel-binaries-bin-sdlc-stays-at-main-and-doesn-t-verify-staged-revisions.md |
| boundary | whole-issue close |
| milestone | — |
| window | f5927dfedbfb2e6dbc7d8c91ac56085732181b7b..448595c97a127ed6d1bcd5d6e9b6906ea3018f55 |
| command | sdlc close --issue 417 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T22:36:46-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range implements the issue’s stage and clear contracts, including recovery when resting main predates the sentinel directive. Targeted tests passed, and removing the fixes in disposable copies made their regression tests fail. No new findings block shipping.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Verification requires an explicit clean stamp and rejects unreadable native executables. Removing these guards made the clean-stamp, ELF, XCOFF, Plan 9 and Mach-O regressions fail.
  - id: BR-2
    disposition: addressed
    note: |
      sentinelBuildTarget restricts declarations to target names. Removing grammar validation made the option, dry-run and assignment regression cases fail.
  - id: BR-3
    disposition: addressed
    note: |
      Production and tests share stageBuildWithMake. Stateful owner-failure/retry and real-Make fixtures passed; skipping owner execution made the retry regression fail.
  - id: BR-4
    disposition: addressed
    note: |
      README.md:17-25 documents sentinel targets, verification refusal, clear rebuilding and retry. These statements match stage.go:130-182 and stagebuild.go:32-105; atlas documents the same flow.
  - id: BR-5
    disposition: addressed
    note: |
      The plan's Concepts table classifies the sentinel reader/parser as INTEGRATION and records the correction under Revisions. This matches sentinelBuildTarget's filesystem read at stagebuild.go:172; the correction is prose-only.
```

1. **Strengths**

   - Stage and clear share one build-and-verify implementation, keeping revision checks consistent (`cmd/sdlc/stagebuild.go:26`).
   - Sentinel outputs are matched by Go package path, supporting differently named binaries while detecting missing outputs (`stagebuild.go:95-99`).
   - Clear preserves the restored checkout after build failure and supports repair through an already-resting retry (`cmd/sdlc/stage.go:130-166`).
   - The legacy fallback is conditional: an already-current generic output avoids unnecessary owner-target execution (`stagebuild.go:64`).

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Passed `go test ./cmd/sdlc -run '^(TestStage|TestSentinel|TestRebuildRestingStage)' -count=1` in 37.889 seconds.
   - Passed pinned-range `git diff --check`.
   - Mutation checks confirmed regressions detect removal of clean-stamp enforcement, corrupt-container rejection, target validation, owner execution and the legacy fallback.
   - Both disposable copies were removed. The full suite and live slot staging were not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared verifier; build recipes remain owned by Make.
   - **ARCH-PURE — pass:** stamp extraction remains pure; filesystem parsing is correctly classified as integration.
   - **ARCH-PURPOSE — pass:** staged builds, restored builds, missing sentinel outputs and legacy-main recovery are covered.
   - **ARCH-MOCK — pass:** shared Make seam has stateful failure/retry coverage and real-Make fixtures.
   - **ARCH-CONSTRAINTS — pass:** owner targets execute sequentially over the finite sentinel set.
   - **ARCH-SECURE — pass:** declarations are validated; incomplete provenance refuses visibly.
   - **ARCH-ORDER — pass:** failed post-switch builds preserve actual checkout state and permit retry.
   - **ARCH-FUNERAL — pass:** runtime changes introduce no durable artifact family.

7. **Plan revision recommendations:** None; the recorded revisions match the implementation.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ad42ca81-b758-4cfc-84e1-796e7cc4561a","repository":"github.com/xianxu/ariadne","issue_id":"000417","recipe":"milestone-review","round":4,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/e40fcdbb105325c6","reviewed_head":"448595c97a127ed6d1bcd5d6e9b6906ea3018f55","slot":"ariadne:1","requested_at":"2026-10-17T05:34:55.16199Z"},"phase":"finished","pid":42411,"started_at":"2026-10-17T05:34:55.259661Z","finished_at":"2026-10-17T05:36:38.315907Z","session_id":"01a1485b-2625-7201-a178-da38d531d10a","outcome":"returned"}] -->
