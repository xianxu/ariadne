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

---

## Re-review — 2026-09-13T12:45:10-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 225 — Publish a portable root Makefile as a safe seed |
| repo | ariadne |
| issue file | workshop/issues/000225-portable-root-makefile-seed.md |
| boundary | whole-issue close |
| milestone | — |
| window | fa8746eb8127d18e1329cc8465c4504c5adc209c..5cb5857c37c7738addca5e5a38de5068eb231be7 |
| command | sdlc close --issue 225 |
| reviewer | codex |
| timestamp | 2026-09-13T12:45:10-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: medium
```

The implementation matches the intended portable seed design, and both prior documentation defects are corrected in the pinned HEAD. However, neither correction has the regression test required by this review’s explicit disposition rule, so both remain `not-addressed`. No new runtime correctness defect was identified. Integration suites were inspected but not executed because they require filesystem writes unavailable in this review environment.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      The plan now correctly classifies workflow discovery as integration and includes a revision. No regression test pins that classification; reverting the documentation correction leaves the existing runtime tests unchanged. The required failing-without-fix evidence is missing.
  - id: BR-2
    disposition: not-addressed
    note: |
      README now documents seed ownership, Makefile.local, bootstrap, and CI hook ordering and failure behavior. Neither portable test reads README, so removing this documentation leaves those tests unchanged. The required failing-without-fix evidence is missing.
```

1. **Strengths**

   - [apply.go:209](/Users/xianxu/workspace/ariadne/cmd/weave/internal/plan/apply.go:209) reads seed bytes before unlinking and shares a fail-closed destination inspection helper with composed writes.
   - [apply_test.go:578](/Users/xianxu/workspace/ariadne/cmd/weave/internal/plan/apply_test.go:578) covers matching, differing, dangling, and missing-source states, plus operation failures and retry convergence.
   - Bootstrap uses sequential recursive Make phases; the fixture checks ordering under `-j4` and prevents post-weave work after weave failure.
   - CI tests execute the workflow’s actual shell blocks and exercise consumer-runner precedence, fallback, and hook failures.

2. **Critical findings**

   BR-1 remains open solely under the mandated regression-evidence rule. Its original classification contradiction is corrected. Add a durable contract check that rejects the former PURE classification, and demonstrate that reverting the correction fails it.

3. **Important findings**

   BR-2 remains open solely under the same rule. Add a documentation contract check covering the required consumer instructions, with removal of the README addition producing failure. These are dispositions of existing findings, not new family instances.

4. **Minor findings**

   None raised.

5. **Test coverage notes**

   Pinned stat and name-status inspections succeeded; checkout HEAD matches the requested SHA. Shell syntax checks and `make -s help` passed. `git diff --check` reported trailing whitespace at README.md:7. Runtime integration and revert tests were not run in the read-only sandbox; implementor-reported passes were not treated as fresh verification.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared symlink guard and upstream-owned templates.
   - **ARCH-PURE — pass in code:** filesystem behavior stays behind integration seams; BR-1’s remaining gap is regression evidence.
   - **ARCH-PURPOSE — pass:** covers standalone targets, migration, bootstrap, and retained consumer CI setup.
   - **ARCH-MOCK — pass by inspection:** stateful scratch fixtures and real weave conformance are present.
   - **ARCH-CONSTRAINTS — pass:** bounded batch setup within the declared environment.
   - **ARCH-SECURE — pass:** source-first reads and destination-link removal protect ancestor files under the stated single-writer contract.
   - **ARCH-ORDER — pass:** explicit bootstrap phases and materialization retry tests.
   - **ARCH-FUNERAL — pass:** fixed output slots and fixture cleanup introduce no growing artifact family.

7. **Plan revision recommendations**

   Append a `## Revisions` entry identifying the BR-1/BR-2 regression checks and their mutation evidence once implemented. No runtime redesign is needed.

---

## Re-review — 2026-09-13T12:47:19-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 225 — Publish a portable root Makefile as a safe seed |
| repo | ariadne |
| issue file | workshop/issues/000225-portable-root-makefile-seed.md |
| boundary | whole-issue close |
| milestone | — |
| window | fa8746eb8127d18e1329cc8465c4504c5adc209c..41bf6affc137d545c5f5175ad751ac73489b98b6 |
| command | sdlc close --issue 225 |
| reviewer | codex |
| timestamp | 2026-09-13T12:47:19-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: medium
```

The pinned implementation matches the portable Makefile, safe seed migration, and CI setup design. Both prior defects are corrected in the prose, but neither has the regression evidence explicitly required by this review’s disposition contract. They therefore remain open; no new runtime correctness defect was found.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      The plan correctly classifies workflow discovery as INTEGRATION and records the correction. No test pins that correction; plan lines 88–93 explicitly decline that evidence. The requested fails-without-the-fix condition remains unmet.
  - id: BR-2
    disposition: not-addressed
    note: |
      README lines 15–34 document seed ownership, Makefile.local, bootstrap, and hook ordering/failure behavior. No test pins this documentation correction, so the requested fails-without-the-fix condition remains unmet.
```

1. **Strengths**
   - `applySeed` reads the source before unlinking the destination; the shared guard rejects inspection/removal errors before writing through a link.
   - Seed tests cover matching, differing, dangling, and missing-source states, checking ancestor bytes and permissions.
   - Bootstrap uses sequential recursive Make phases; its fixture checks parallel execution ordering and stops after weave failure.
   - CI tests execute the actual workflow shell blocks with the real check runner, covering hook failure and runner precedence.

2. **Critical findings**
   - **BR-1 remains open under the supplied evidence rule.** The classification itself is corrected at [plan line 21](/Users/xianxu/workspace/ariadne/workshop/plans/000225-portable-root-makefile-seed-plan.md:21). This is the existing finding, not another classification defect.

3. **Important findings**
   - **BR-2 remains open under the supplied evidence rule.** The requested public documentation exists at [README line 15](/Users/xianxu/workspace/ariadne/README.md:15). This is the existing finding, not another documentation omission.

4. **Minor findings**
   - None.

5. **Test coverage notes**
   - Required stat, name-status, and targeted patch inspections succeeded; checkout HEAD matches the pinned head.
   - Bash syntax validation and range whitespace checks passed.
   - Behavioral suites and mutation checks were not executed: they require temporary filesystem writes prohibited by this session. Implementor test results were not treated as independently verified.
   - Existing tests exercise runtime behavior; they cannot detect reverting either prose correction.

6. **Architectural notes**
   - **ARCH-DRY — pass:** shared destination-link guard and reusable helper resolution.
   - **ARCH-PURE — pass structurally:** filesystem operations remain behind the FS seam; discovery is now classified as integration. BR-1 evidence remains open.
   - **ARCH-PURPOSE — pass structurally:** covers migration, missing helpers, standalone targets, and preserved consumer CI setup.
   - **ARCH-MOCK — pass by inspection:** stateful scratch fixtures and production entry points; real-weave conformance is included.
   - **ARCH-CONSTRAINTS — pass:** bounded synchronous setup fits the declared batch workload.
   - **ARCH-SECURE — pass:** source-first handling protects old link targets; destination inspection fails closed.
   - **ARCH-ORDER — pass:** explicit bootstrap phases and retry coverage.
   - **ARCH-FUNERAL — pass:** fixed output slots; fixture cleanup is defined.

7. **Plan revision recommendations**
   - Reconcile [the documentation-only verification revision](/Users/xianxu/workspace/ariadne/workshop/plans/000225-portable-root-makefile-seed-plan.md:88) with this gate’s explicit regression-evidence requirement. The prose repairs are present; the remaining blocker is the conflicting evidence contract.

---

## Re-review — 2026-09-13T12:52:18-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 225 — Publish a portable root Makefile as a safe seed |
| repo | ariadne |
| issue file | workshop/issues/000225-portable-root-makefile-seed.md |
| boundary | whole-issue close |
| milestone | — |
| window | fa8746eb8127d18e1329cc8465c4504c5adc209c..c3c46d39ecf09514ebab6f8da8dd3a3306016453 |
| command | sdlc close --issue 225 |
| reviewer | codex |
| timestamp | 2026-09-13T12:52:18-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: medium
```

The pinned implementation matches the issue’s portability and safe migration contracts. Both prior findings are addressed; no new blocking defects found. Confidence is limited by the read-only environment: runtime suites and mutation checks were inspected, not rerun.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The plan's Core concepts table now classifies workflow discovery as integration and records the correction under Revisions. This matches Makefile:11 filesystem wildcard discovery and the scratch-filesystem Make tests.
  - id: BR-2
    disposition: addressed
    note: |
      README.md:15-34 documents seed ownership, Makefile.local, bootstrap, and hook ordering/skipping/failure behavior. These match construct/base.manifest, Makefile:11-13, Makefile.workflow's bootstrap phases, and .github/workflows/merge-check.yml:28-77.
```

1. **Strengths**

   - `applySeed` reads source bytes before unlinking; shared destination inspection fails closed before writing or changing permissions.
   - Filesystem tests cover matching, differing, dangling links, ancestor preservation, injected failures, and retry convergence.
   - Bootstrap explicitly sequences weave before dependent tooling, including parallel Make execution.
   - CI tests execute actual workflow shell blocks against the real generic runner; README and atlas explain the new surface.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Required stat, name-status, and targeted pinned diffs inspected successfully. `bash -n` and `git diff --check` passed. Runtime tests require temporary writes unavailable here. The prompt-policy regression reaches `BuildPrompt`; its required clauses are absent from the base implementation, and the golden captures the complete changed policy.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared symlink guard and helper resolution.
   - **ARCH-PURE — pass:** filesystem behavior remains behind integration seams.
   - **ARCH-PURPOSE — pass:** addresses migration, standalone targets, first bootstrap, and durable CI customization.
   - **ARCH-MOCK — pass:** portable filesystem backing, stateful fixture tools, and real-runner/weave conformance coverage.
   - **ARCH-CONSTRAINTS — pass:** bounded setup work; no new unbounded workload.
   - **ARCH-SECURE — pass:** ancestor preservation and inspection failures handled; repository hooks retain repository-code authority.
   - **ARCH-ORDER — pass:** explicit synchronous phases and failure propagation.
   - **ARCH-FUNERAL — pass:** fixed output slots and automatic fixture cleanup; no new accumulating artifacts.

7. **Plan revision recommendations:** None.
