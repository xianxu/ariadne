# Boundary Review — ariadne#530 (whole-issue close)

| field | value |
|-------|-------|
| issue | 530 — local-check: a suite that fails before any test on ENOSPC (full Go cache) is could-not-run, not an unattributable failure or main-red |
| repo | ariadne |
| issue file | workshop/issues/000530-local-check-a-suite-that-fails-before-any-test-on-enospc-full-go-cache-is-could-not-run-not-an-unattributable-failure-or-main-red.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4ba9e4210107b9ad50d0e3e731c4c0c93bcf6071..b76d0d9fa7959b96ad43a3719dd630513bc3beb1 |
| command | sdlc close --issue 530 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T15:23:28-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The focused behavioral checks passed, including startup classification, audited recovery, genuine-rerun enforcement, and the pinned #524 reader. Two plan-contract gaps remain: the Core concepts table points to the wrong implementation file, and live recovery advice omits the exact failed revision. README and atlas updates are present.

```findings
findings:
  - id: new
    severity: Critical
    family: core-concept-location-consistency
    title: |
      The Core concepts table locates validateReclassificationProof in the wrong file
    detail: |
      workshop/plans/000530-suite-startup-evidence-plan.md:52 lists cmd/sdlc/suitereclassification.go, but the function exists at cmd/sdlc/teststats_reclassify.go:67. This contradicts the required Core concepts cross-check and ARCH-PURPOSE traceability. Correct the table and add a Revisions entry; this prose-only correction needs source inspection, not a wording test.
  - id: new
    severity: Important
    family: exact-subject-recovery-guidance
    title: |
      Live could-not-run advice drops the failed revision
    detail: |
      cmd/sdlc/suiteenvironment.go:45 recommends local-check run --name CHECK without --rev, although the execution contract requires the exact check/head. A failure from run --rev OTHER or a diagnostic baseline therefore recommends a command that defaults to the caller's HEAD. Carry the executed SHA into suiteEnvironmentError and include --rev SHA across direct, affected, rerun and baseline paths; add a regression where the executed SHA differs from HEAD. This is an ARCH-PURPOSE gap.
```

1. **Strengths**

   - Startup recognition rejects test activity, mixed output, incomplete wrappers and truncated lines.
   - Recovery binds original bytes, actor, reason, audit and immutable log before replacing authority.
   - Interruption tests cover publication before and after replacement, including lost acknowledgment and repeat recovery.
   - The actual #524 reader safely refuses related corrected evidence while permitting unrelated proof.

2. **Critical findings**

   - The plan’s entity-location contradiction above is Critical under the explicitly required Core concepts rule. The implementation exists; the correction is documentary.

3. **Important findings**

   - Preserve the executed revision in live recovery guidance. Historical reclassification already prints an exact `--rev`; live errors should provide equivalent advice.

4. **Minor findings**

   - None.

5. **Test coverage notes**

   - Focused startup, recovery, environment, guard and saved-result tests passed: `cmd/sdlc`, 15.346s.
   - Full `internal/flakes` package passed.
   - `scripts/verify-524-reader.sh` passed against pinned historical source. Its disposable checkout was removed.
   - No full repository suite was run during this review. Tracked files remain unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared classifier and evidence projection.
   - **ARCH-PURE — pass:** diagnosis, eligibility and proof validation separate from file effects.
   - **ARCH-PURPOSE — flag:** the two contract gaps above.
   - **ARCH-MOCK — pass:** isolated process/Git fixtures and fake GitHub effects.
   - **ARCH-CONSTRAINTS — pass:** bounded reads and counted N/2N controls.
   - **ARCH-SECURE — pass:** strict decoding, ownership and digest checks.
   - **ARCH-ORDER — pass:** audit-first publication and deterministic interruption coverage.
   - **ARCH-FUNERAL — pass:** attempt-linked pruning and orphan-audit cleanup.
   - **ARCH-EVOLVE — pass:** opaque fields survive correction; historical-reader behavior is exercised.

7. **Plan revision recommendations**

   - Add a `## Revisions` entry correcting `validateReclassificationProof`’s location to `cmd/sdlc/teststats_reclassify.go`.
   - Retain the exact-check/head recovery contract and implement the missing live SHA propagation.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"c297bae0-43c2-4efa-a832-aca5aba269eb","repository":"github.com/xianxu/ariadne","issue_id":"000530","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/1fc084d97a6367d7","reviewed_head":"b76d0d9fa7959b96ad43a3719dd630513bc3beb1","slot":"ariadne:5","requested_at":"2026-10-21T22:21:38.716222Z"},"phase":"finished","pid":74641,"started_at":"2026-10-21T22:21:38.987912Z","finished_at":"2026-10-21T22:23:28.004909Z","session_id":"01a1608e-4778-7b30-a814-a73f0fbf9967","outcome":"returned"}] -->
