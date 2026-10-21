# Boundary Review — ariadne#509 (whole-issue close)

| field | value |
|-------|-------|
| issue | 509 — Landing scoped local checks |
| repo | ariadne |
| issue file | workshop/issues/000509-landing-scoped-local-checks.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6573f0810474e4de9a82fc0a4b7a4c7670c78996..8ed79fac33b27fd91dda3182514bf06e9fc54d1d |
| command | sdlc close --issue 509 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T01:33:17-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range implements the issue’s landing gate across trains and both single-landing paths. Selection, evidence binding, runner resource isolation, and cleanup retention have meaningful regression coverage. No new blocking findings emerged; all four prior dispositions remain supported.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      landingchecks_git.go:51 rejects tracked blank policy blobs while preserving genuine absence. TestLandingCheckTrackedBlankPolicyRefuses passes at head and fails when that guard is removed.
  - id: BR-2
    disposition: addressed
    note: |
      localcheck_resource_unix.go:70-75 independently confirms completion after escalation. TestRunnerCleanupNilTerminationIsNotCompletion passes at head and fails when the final confirmation guard is removed.
  - id: BR-3
    disposition: addressed
    note: |
      The pinned README diff adds schema, selection, --all exclusion, resource and recovery documentation at README.md:598-622, consistent with localchecks_v2.go, selectLocalChecks and runUnderResource. No wording test is needed for this prose correction.
  - id: BR-4
    disposition: addressed
    note: |
      Immediate cleanup and stale collectors retain artifacts under shared runner ownership. Artifact retention, reconciliation and panic controls pass; disabling runnerArtifactProtected makes TestRunnerCollectorsRespectPrivateAuthority fail on premature scratch deletion.
```

1. **Strengths**

   - One shared gate serves train, tracked single and legacy landing paths, reducing policy divergence.
   - Selection uses original car path unions and immutable base/head caching; counted tests cover repeated reads and growth.
   - Proof requires matching identity and a verified retained-log hash. Saved statuses cannot independently authorize landing.
   - Runner cleanup distinguishes successful signalling from confirmed completion and preserves recovery artifacts when completion remains unknown.
   - README, CLI help and atlas changes document the new configuration and operational surface.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Targeted command tests passed in **88.268s**, including production train/single/legacy execution, selection, codec, resource and cleanup controls.
   - Full `landingline` and `golint` package tests passed.
   - In a disposable pinned-head copy, removing the BR-1, BR-2 and artifact-protection guards caused all three corresponding regression tests to fail. The copy was removed.
   - The reviewed checkout remains unchanged. I did not rerun the repository-wide suite or drive a live VM.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared gate, matcher, runner and artifact-protection predicate.
   - **ARCH-PURE — pass:** codecs, selection, proof matching and train transitions remain deterministic; IO sits in separate seams.
   - **ARCH-PURPOSE — pass:** consumer sweep covers per-car exclusion, final-prefix checks, reintegration and both single-landing paths.
   - **ARCH-MOCK — pass:** stateful GitHub fixtures and controlled runner/process seams exercise production boundaries.
   - **ARCH-CONSTRAINTS — pass:** bounded inputs, command timeout, cleanup budgets and counted caching controls support the declared batch workload.
   - **ARCH-SECURE — pass:** pinned regular policy blobs, strict parsing and verified evidence fail visibly on invalid inputs.
   - **ARCH-ORDER — pass:** explicit checking events/effects enforce suite-before-check ordering and terminal check refusal.
   - **ARCH-FUNERAL — pass:** ordinary retirement is retained; unknown runner outcomes explicitly require reconciliation before artifact removal.
   - **ARCH-EVOLVE — pass:** v1 compatibility remains, while v2 preserves opaque metadata and strictly validates known members.

7. **Plan revision recommendations:** None required.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ca40037b-7136-4f69-814d-86b0365d9ac1","repository":"github.com/xianxu/ariadne","issue_id":"000509","recipe":"milestone-review","round":5,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/ce565ed04f5e1b55","reviewed_head":"8ed79fac33b27fd91dda3182514bf06e9fc54d1d","slot":"ariadne:3","requested_at":"2026-10-21T08:29:10.031011Z"},"phase":"finished","pid":10167,"started_at":"2026-10-21T08:29:10.117315Z","finished_at":"2026-10-21T08:33:16.653258Z","session_id":"01a15d94-1d64-7cd3-9b79-2d8f11423f39","outcome":"returned"}] -->
