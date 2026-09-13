# Boundary Review — ariadne#225 (whole-issue close)

| field | value |
|-------|-------|
| issue | 225 — Publish a portable root Makefile as a safe seed |
| repo | ariadne |
| issue file | workshop/issues/000225-portable-root-makefile-seed.md |
| boundary | whole-issue close |
| milestone | — |
| window | fa8746eb8127d18e1329cc8465c4504c5adc209c..c69fb7285a62a467656955b3478277a0d113877a |
| command | sdlc close --issue 225 |
| reviewer | codex |
| timestamp | 2026-09-13T12:41:58-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: medium
```

The implementation addresses the seed overwrite hazard and adds portable bootstrap/CI paths with meaningful regression fixtures. Two contract gaps remain: the plan classifies filesystem-dependent Make logic as PURE, and README documentation is missing for the new consumer setup surface. No additional runtime correctness defect was identified. Full test execution was unavailable under the read-only filesystem restriction.

1. **Strengths**

   - `applySeed` reads the source before removing destination links; the shared guard also makes `applyWriteFile` fail closed on inspection errors.
   - Seed tests cover matching, differing, and dangling links, ancestor bytes/mode preservation, injected failures, and retry convergence.
   - Bootstrap phases explicitly sequence weave before dependent tooling, including parallel Make invocation.
   - CI fixtures execute actual workflow shell blocks and the real generic runner; both relevant atlas pages were updated.

2. **Critical findings**

   - **ARCH-PURE — incorrect core-concept classification.** [Plan:21](/Users/xianxu/workspace/ariadne/workshop/plans/000225-portable-root-makefile-seed-plan.md:21) labels workflow selection a “pure Make expression.” [Makefile:11](/Users/xianxu/workspace/ariadne/Makefile:11) uses `wildcard` to inspect filesystem state; its regression test creates files and invokes Make. This contradicts the required PURE contract. Classify this entity as INTEGRATION and append a plan revision. No runtime redesign is necessary.

3. **Important findings**

   - **README update missing for consumer setup.** [Workflow:56](/Users/xianxu/workspace/ariadne/.github/workflows/merge-check.yml:56) introduces executable `scripts/ci-setup.sh` as a consumer extension point, alongside changed standalone/bootstrap behavior. `README.md` is unchanged in the pinned range. Add concise usage guidance covering seed ownership, `Makefile.local`, bootstrap, and the executable hook’s ordering and failure behavior.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Required stat, name-status, and patch inspection completed; checkout HEAD matches the pinned head. Shell syntax checks, range whitespace checks, and `make -s help` passed. Git/Make emitted sandbox cache warnings but returned successful results. Go and scratch integration suites were inspected, not executed: they require filesystem writes. No prior boundary findings required disposition.

6. **Architectural notes**

   - **ARCH-DRY: pass** — shared materialization guard; upstream templates remain authoritative.
   - **ARCH-PURE: flag** — classification contradiction described above.
   - **ARCH-PURPOSE: pass** — changes cover migration, missing helpers, repeat weave, and CI ownership.
   - **ARCH-MOCK: pass on inspection** — stateful filesystem faults and portable tool fixtures exercise production seams.
   - **ARCH-CONSTRAINTS: pass** — bounded operator-triggered setup; no new growing workload.
   - **ARCH-SECURE: pass** — destination inspection fails closed; fixtures isolate their working state.
   - **ARCH-ORDER: pass** — explicit bootstrap phases and failure/retry coverage.
   - **ARCH-FUNERAL: pass** — fixed output slots and temporary fixture cleanup.

7. **Plan revision recommendations**

   Append a dated `## Revisions` entry classifying workflow discovery as INTEGRATION because it reads filesystem state, and add README delivery to the documentation task.

```findings
findings:
  - id: new
    severity: Critical
    family: core-concept-classification
    title: |
      Workflow discovery is classified PURE despite filesystem-dependent behavior
    detail: |
      Plan line 21 contradicts Makefile line 11 and its filesystem/subprocess test. ARCH-PURE: classify workflow discovery as INTEGRATION and append a plan revision; no runtime redesign is required.
  - id: new
    severity: Important
    family: user-surface-documentation
    title: |
      README update is missing for portable consumer setup
    detail: |
      The workflow introduces executable scripts/ci-setup.sh at line 56, but README.md is unchanged in the review range. Document seed ownership, Makefile.local, bootstrap, and the hook's execution order and failure behavior.
```
