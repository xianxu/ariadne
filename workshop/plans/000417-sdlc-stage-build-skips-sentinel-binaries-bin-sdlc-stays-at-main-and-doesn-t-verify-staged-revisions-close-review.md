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
