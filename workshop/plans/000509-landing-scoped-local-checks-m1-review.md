# Boundary Review — ariadne#509 (milestone M1)

| field | value |
|-------|-------|
| issue | 509 — Landing scoped local checks |
| repo | ariadne |
| issue file | workshop/issues/000509-landing-scoped-local-checks.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 6573f0810474e4de9a82fc0a4b7a4c7670c78996..456fa34aec88b993a6dfbc2636a12884883929b7 |
| command | sdlc milestone-close --issue 509 --milestone M1 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T23:09:30-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

M1 provides a sound separation between declaration policy, pure selection, Git inputs, and resource execution. Targeted tests pass. Two contract gaps block admission: empty tracked configuration silently disables selection, and forced cleanup can report completion without confirming descendant termination. The new user-facing configuration also needs README coverage at this boundary.

**1. Strengths**

- Selection unions per-car paths and preserves ordered exclusions, preventing a later revert from hiding an earlier trigger.
- V2 preserves opaque nested metadata while rejecting duplicate keys, invalid known fields, and incompatible resource settings.
- Runner execution bypasses suite-helper discovery/allocation and removes inherited suite authority. Its test also verifies retained failure output and status.
- Counted tests verify cached Git reads and linear matching work; the six-second cleanup fixture exercises the former five-second limit.

**2–4. Findings**

```findings
findings:
  - id: new
    severity: Critical
    family: policy-input-fails-closed
    title: |
      A tracked empty declaration becomes a successful empty selection
    detail: |
      cmd/sdlc/landingchecks_git.go:64-68 passes both an absent declaration and an existing empty/whitespace blob to parseLocalChecks, whose empty-input branch returns an empty declaration. readLandingCheckInputs then returns successfully at lines 102-103 without source-car evidence. This contradicts the plan's rejection of truncated policy input and fabricates an empty selection from invalid configuration (ARCH-SECURE). Distinguish absence from a present blank blob in the new reader, preserve legacy parser behavior, and add regression cases proving blank tracked declarations refuse while genuine absence remains compatible.
  - id: new
    severity: Critical
    family: cleanup-requires-confirmed-completion
    title: |
      Forced runner cleanup accepts an unconfirmed group termination
    detail: |
      cmd/sdlc/localcheck_resource_unix.go:56-59 returns success after processgroup.Terminate(cmd, true) returns nil, without rechecking group termination. That helper also returns nil when its 100ms repeated-kill budget expires (cmd/sdlc/internal/processgroup/process_unix.go:46-59), so nil does not establish that descendants stopped. The caller then retires the run lease and checkout despite the plan's force-and-join guarantee (ARCH-ORDER, ARCH-FUNERAL). Confirm termination after escalation within a bounded join; if it remains unconfirmed, return a visible cleanup error and preserve the necessary recovery ownership. Add a controllable regression where termination returns nil while the group remains live.
  - id: new
    severity: Important
    family: user-surface-documentation
    title: |
      README coverage is missing for the M1 configuration and selection surface
    detail: |
      This range introduces version 2 with scope, resource, and paths_file, and changes local-check run --all to exclude landing-scoped checks (cmd/sdlc/localcheck_cmd.go:111-118), but contains no README.md update. README.md:591-593 still describes named checks as running under the machine-wide suite lock, which is false for resource:runner. Update README now with the delivered M1 behavior and explicitly identify automatic enforcement as pending M2; the docs gate applies at each milestone.
```

No Minor findings.

**5. Test coverage notes**

Passed:

- Targeted parser, selection, runner, signal, and escalation tests: **31.428s**.
- Narrow runner/proof race tests: **8.164s**.
- `internal/affected` tests and `git diff --check`.

The escalation test covers the direct process; it does not exercise an unconfirmed surviving descendant after the terminator’s budget expires. Git-reader tests lack the tracked-empty-declaration refusal case. No full suite was run. The checkout remains clean.

**6. Architectural notes for upcoming work**

- **ARCH-DRY — pass:** existing matcher, changed-path argv, execution shell, and evidence storage are reused.
- **ARCH-PURE — pass:** codecs, selection, and proof matching have direct tests without IO.
- **ARCH-PURPOSE — pass for M1:** union selection and resource separation are delivered. Train/car/single-landing consumer enforcement remains explicitly M2.
- **ARCH-MOCK — pass for M1:** injected Git, disposable repositories, fake GitHub, and private subprocess fixtures exercise the production boundaries.
- **ARCH-CONSTRAINTS — pass:** counted growth controls, input limits, and bounded runtime/grace are present.
- **ARCH-SECURE — flag:** present-but-empty policy is collapsed into absence.
- **ARCH-ORDER — flag:** a bounded termination attempt is collapsed into confirmed cleanup.
- **ARCH-FUNERAL — flag:** that uncertainty permits premature lease/worktree retirement.
- **ARCH-EVOLVE — pass:** the unsafe semantic change has a version boundary; unknown metadata survives round-trip.

**7. Plan revision recommendations**

Append `## Revisions` entries to:

- Move documentation of the already-delivered config keys and `--all` behavior into M1.
- Explicitly require tracked-blank policy refusal coverage and a deterministic unconfirmed-cleanup regression, alongside the existing process fixtures.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"9d6b9c64-bfb1-40b1-b19d-03a24a67f835","repository":"github.com/xianxu/ariadne","issue_id":"000509","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/6262fcc90694515f","reviewed_head":"456fa34aec88b993a6dfbc2636a12884883929b7","slot":"ariadne:3","requested_at":"2026-10-21T06:05:58.902945Z"},"phase":"finished","pid":87758,"started_at":"2026-10-21T06:05:59.056938Z","finished_at":"2026-10-21T06:09:29.618874Z","session_id":"01a15d11-0711-7573-a4a8-e5a9cfc573ab","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T23:20:17-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 509 — Landing scoped local checks |
| repo | ariadne |
| issue file | workshop/issues/000509-landing-scoped-local-checks.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 6573f0810474e4de9a82fc0a4b7a4c7670c78996..bdff0b680d0eb09be2ef59f6d3731d7b2869743a |
| command | sdlc milestone-close --issue 509 --milestone M1 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T23:20:17-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The three prior findings are addressed, with mutation evidence for both behavioral fixes. M1’s declaration, selection, and resource separation are sound. One related cleanup defect still blocks admission: an unconfirmed runner’s temporary directory is deleted even though descendants may still need it.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      readLandingBlob rejects tracked blank blobs while preserving genuine absence. TestLandingCheckTrackedBlankPolicyRefuses passed and failed when the blank guard was removed in a disposable checkout.
  - id: BR-2
    disposition: addressed
    note: |
      finishRunnerGroup separately confirms completion after escalation. The nil-termination/live-group regression passed and failed without that confirmation; recovery tests verify retained worktree, log, and unknown queue activity.
  - id: BR-3
    disposition: addressed
    note: |
      README.md:598-610 documents v2 fields, --all exclusion, runner resource behavior, and pending M2 enforcement. These descriptions match localchecks_v2.go, selectLocalChecks, and runUnderResource.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: cleanup-requires-confirmed-completion
    title: |
      Unconfirmed runner cleanup still deletes its temporary inputs
    detail: |
      cmd/sdlc/localcheck_run.go:446-450 returns on unconfirmed group cleanup, but the unconditional deferred scratch.Cleanup at line 429 still removes the directory supplied as TMPDIR/GOTMPDIR (testscratch.go:49). Surviving descendants therefore lose inputs despite retained checkout ownership. A disposable regression using the existing cleanup seam reproduced the missing scratch file. This is the 2nd finding in family cleanup-requires-confirmed-completion. Do NOT fix only this instance: state and enforce the rule that every artifact potentially used by an unconfirmed descendant remains owned until completion or explicit reconciliation. Sweep checkout, scratch, logs, and leases, including stale collectors; testscratch.go:115-120 currently removes scratch based on owner PID/age rather than descendant completion. Add deterministic recovery coverage for immediate return and later sweeping (ARCH-PURPOSE, ARCH-ORDER, ARCH-FUNERAL).
```

**1. Strengths**

- Selection unions per-car paths and applies ordered exclusions, preserving eligibility across later reverts.
- V2 preserves opaque metadata while strictly validating known fields and resource constraints.
- Runner execution bypasses suite allocation and removes inherited suite authority; tests verify failure status and retained output.
- README and atlas changes describe the delivered capability and identify enforcement as M2 work.

**2. Critical findings**

- The scratch-lifetime defect above requires a shared retention rule covering all runner-owned artifacts, plus regression coverage.

**3. Important findings**

None.

**4. Minor findings**

None.

**5. Test coverage notes**

Passed:

- Targeted parser, selection, runner, signal, and cleanup tests: **29.325s**.
- Narrow cleanup race controls: **2.190s**.
- `internal/affected` tests and `git diff --check`.

Both prior behavioral regressions failed when their fixes were removed. The additional scratch-retention regression failed on the pinned implementation. All mutation work occurred in a disposable checkout, subsequently removed. The reviewed checkout remains clean. No full suite was run.

**6. Architectural notes**

- **ARCH-DRY — pass:** existing matcher, Git arguments, execution shell, and evidence store are reused.
- **ARCH-PURE — pass:** codecs, selection, and proof matching have direct tests without IO.
- **ARCH-PURPOSE — flag:** cleanup retention fixes covered named artifacts but missed scratch; the artifact class needs enumeration.
- **ARCH-MOCK — pass for M1:** injected boundaries, disposable repositories, fake GitHub, and controlled subprocess fixtures exercise production paths.
- **ARCH-CONSTRAINTS — pass:** input/runtime bounds and counted growth controls are present.
- **ARCH-SECURE — pass:** tracked blank policy now refuses; known-field validation remains strict.
- **ARCH-ORDER — flag:** unconfirmed completion still triggers destructive scratch cleanup.
- **ARCH-FUNERAL — flag:** scratch retirement is disconnected from descendant lifetime and recovery ownership.
- **ARCH-EVOLVE — pass:** v2 provides the unsafe-to-ignore version boundary and preserves unknown metadata.

The M1 pure entities exist and are tested directly. Train transitions and shared landing enforcement remain explicitly assigned to M2.

**7. Plan revision recommendations**

Append a `## Revisions` entry enumerating checkout, scratch, logs, and leases as runner lifetime artifacts. Define their retention and reconciliation behavior after unconfirmed cleanup or owner death, and require stale collectors to honor that ownership.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"608c4232-78b8-46a7-ac8b-f5a5f0fb2c0b","repository":"github.com/xianxu/ariadne","issue_id":"000509","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/ae882f2aa0643871","reviewed_head":"bdff0b680d0eb09be2ef59f6d3731d7b2869743a","slot":"ariadne:3","requested_at":"2026-10-21T06:16:51.743369Z"},"phase":"finished","pid":4228,"started_at":"2026-10-21T06:16:51.873689Z","finished_at":"2026-10-21T06:20:16.823649Z","session_id":"01a15d1a-fd32-7270-bafc-17a629b20b3a","outcome":"returned"}] -->
