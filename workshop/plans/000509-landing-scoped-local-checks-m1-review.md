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
