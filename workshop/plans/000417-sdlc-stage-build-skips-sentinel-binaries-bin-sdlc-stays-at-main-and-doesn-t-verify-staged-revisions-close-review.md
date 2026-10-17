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
