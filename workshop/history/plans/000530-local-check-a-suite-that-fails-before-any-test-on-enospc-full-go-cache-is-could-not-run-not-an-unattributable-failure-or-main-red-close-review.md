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

---

## Re-review — 2026-10-21T16:48:40-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 530 — local-check: a suite that fails before any test on ENOSPC (full Go cache) is could-not-run, not an unattributable failure or main-red |
| repo | ariadne |
| issue file | workshop/issues/000530-local-check-a-suite-that-fails-before-any-test-on-enospc-full-go-cache-is-could-not-run-not-an-unattributable-failure-or-main-red.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4ba9e4210107b9ad50d0e3e731c4c0c93bcf6071..99c89cf3f2181f380fafd754afd28eb65e10674d |
| command | sdlc close --issue 530 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T16:48:40-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned changes satisfy the issue’s startup-classification and audited-recovery contracts. Both prior findings are addressed, with regression evidence for the behavior change. No new blocking findings emerged. This review does not replace the separately required final-head recorded suite.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The before/after diff corrects the Core concepts row to cmd/sdlc/teststats_reclassify.go, matching validateReclassificationProof at line 67; the plan's Revisions entry at line 309 records the correction.
  - id: BR-2
    disposition: addressed
    note: |
      Live errors carry the executed SHA and preserve the parent suite/affected check. Caller-HEAD-advanced regressions pass across direct, affected, rerun, baseline and missing-Go paths; removing revision advice through a temporary compiler overlay makes all six cases fail.
```

1. **Strengths**

   - Strict startup grammar rejects test activity, mixed diagnostics and incomplete output.
   - Audit-first publication preserves original bytes and verifies record, audit and log identity before projecting neutral evidence.
   - Recovery cannot publish success, promote an older saved pass or explain a separate code failure.
   - README, help and atlas updates cover the new command, compatibility and retention.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed focused startup, reclassification, recovery-catalog, scoped-proof and lock-metadata tests; the complete flakes package; and the actual pinned #524 reader compatibility control. The temporary overlay confirmed that BR-2’s regressions fail without revision advice. Repository files remain unchanged; private verification artifacts were removed.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Shared classifier and evidence projection cover the consumers.
   - **ARCH-PURE — pass:** Diagnosis, eligibility and proof validation remain separate from file effects.
   - **ARCH-PURPOSE — pass:** Candidate, affected, rerun, baseline/train and historical-recovery paths are covered.
   - **ARCH-MOCK — pass:** Private repositories, recorded output and the existing stateful GitHub fake exercise production seams.
   - **ARCH-CONSTRAINTS — pass:** Bounded reads and counted scan/proof tests enforce the declared envelope.
   - **ARCH-SECURE — pass:** Ownership, strict decoding and digest checks precede authority changes.
   - **ARCH-ORDER — pass:** Audit-first publication, interrupted retry and lost-response controls preserve failure authority.
   - **ARCH-FUNERAL — pass:** Attempt-linked pruning and orphan-audit cleanup cover durable residue.
   - **ARCH-EVOLVE — pass:** Unknown values survive correction; the historical reader safely scopes its refusal.

7. **Plan revision recommendations:** None. The Core concepts locations and documented implementation match the code; subsequent suite and landing gates remain explicitly pending.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"96761d7e-422a-4005-b19e-97c73abe1dd3","repository":"github.com/xianxu/ariadne","issue_id":"000530","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/89447e6c45aa52ab","reviewed_head":"99c89cf3f2181f380fafd754afd28eb65e10674d","slot":"ariadne:5","requested_at":"2026-10-21T23:46:42.664658Z"},"phase":"finished","pid":19215,"started_at":"2026-10-21T23:46:42.820403Z","finished_at":"2026-10-21T23:48:40.072441Z","session_id":"01a160dc-27a9-7fb1-9e7f-9040574bc097","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T19:28:00-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 530 — local-check: a suite that fails before any test on ENOSPC (full Go cache) is could-not-run, not an unattributable failure or main-red |
| repo | ariadne |
| issue file | workshop/issues/000530-local-check-a-suite-that-fails-before-any-test-on-enospc-full-go-cache-is-could-not-run-not-an-unattributable-failure-or-main-red.md |
| boundary | whole-issue close |
| milestone | — |
| window | f62f6fc7e8f08502b427d4ec4f368de0d5a682cd..375050fa4c18294459edd172359d8d644c268b73 |
| command | sdlc close --issue 530 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T19:28:00-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The interdiff preserves both sides of all three conflict resolutions and enforces the existing genuine-rerun contract across composite reuse and posting. Both prior findings remain addressed. No new defects found; the revision entry records integration fixes without changing scope.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The Core concepts row names cmd/sdlc/teststats_reclassify.go, matching validateReclassificationProof at line 67.
  - id: BR-2
    disposition: addressed
    note: |
      Executed-revision advice remains intact. Current tests pass; removing revision advice through a temporary compiler overlay makes all six direct, rerun, baseline and missing-Go cases fail.
```

1. **Strengths**
   - Command wiring retains reclassify, resume and reconcile (`localcheck_cmd.go:75`).
   - Recovery guards cover composite selection and cached posting (`localcheck_republish.go:108,218`).
   - Publication shares the existing directory lock; pruning test seams share one implementation.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes:** Focused recovery, startup, resume-store, retention and catalog tests passed. Temporary overlays confirmed regressions fail without the composite guards or publication lock. Full suite was not rerun. Repository files remain unchanged.

6. **Architectural notes**
   - **ARCH-DRY — pass:** Shared pruning implementation and existing lock reused.
   - **ARCH-PURE — pass:** File coordination remains in the IO boundary.
   - **ARCH-PURPOSE — pass:** Composite reuse, cached posting and completed-resume reposting enforce genuine recovery.
   - **ARCH-EVOLVE — pass:** No persisted-format change introduced.

   Atlas documents the lock integration. No new user-facing command or flag requires a README update in this window.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"c7fbedd3-9101-44eb-b4b5-69ce580d44b7","repository":"github.com/xianxu/ariadne","issue_id":"000530","recipe":"fix-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/bdd81efd1868cb26","reviewed_head":"375050fa4c18294459edd172359d8d644c268b73","slot":"ariadne:5","requested_at":"2026-10-22T02:25:57.397992Z"},"phase":"finished","pid":71041,"started_at":"2026-10-22T02:25:57.485491Z","finished_at":"2026-10-22T02:27:59.826295Z","session_id":"01a1616d-f1f9-73f1-ab12-8386ce1d5f2a","outcome":"returned"}] -->
