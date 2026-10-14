# Boundary Review — ariadne#332 (milestone M3)

| field | value |
|-------|-------|
| issue | 332 — sdlc merge integrates a moved main itself, then re-runs the suite before landing |
| repo | ariadne |
| issue file | workshop/issues/000332-sdlc-merge-integrates-a-moved-main-itself-then-re-runs-the-suite-before-landing.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 05de5db107a4db2927f2e6e424e4bedb51d28c03..6868d4d5693919e9bbd323a02d71f3f73b6ce0be |
| command | sdlc milestone-close --issue 332 --milestone M3 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T02:17:27-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

M3 separates parsing from IO, carries failed-run evidence through a typed error, and documents the diagnostic baseline. However, the baseline cannot reliably run the default `make test`: its reentrant lock acquisition loses the holder identity, causing the nested suite to wait on the lock held by its own `sdlc` process. This blocks the promised failure classification.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: lock-reentry-preserves-owner-identity
    title: |
      Diagnostic baseline loses lock ownership and waits on its own process
    detail: |
      cmd/sdlc/suitebaseline.go:27 calls runLocalCheck a second time while the first run's suite lock remains held until sdlc exits. acquireSuiteSlot returns an empty holder for "reentrant" at cmd/sdlc/localcheck_run.go:206, so the baseline subprocess receives no WF_TEST_LOCK_HELD_BY; nested make test then waits on sdlc's occupied slots instead of running. An isolated reproduction with the real helper returned acquired, then reentrant, then timeout without the holder identity; supplying that identity returned reentrant immediately. Preserve the actual lock owner's identity across every repeated run, and cover baseline, integration retries, and multiple named checks with stateful lock tests. ARCH-PURPOSE, ARCH-MOCK, and ARCH-ORDER apply.
```

1. **Strengths**

   - `localCheckFailure` carries the check and completed run without parsing refusal prose (`cmd/sdlc/cigate.go:386`).
   - Failure parsing and classification are pure, directly tested functions (`cmd/sdlc/suitefailures.go`).
   - The baseline skips status and comment publication, and cannot authorize landing (`cmd/sdlc/localcheck_run.go:125`).
   - README and atlas describe the new diagnostic flow in this range.

2. **Critical findings**

   - **Lock ownership is lost on the baseline run** (`cmd/sdlc/suitebaseline.go:27`; underlying handling at `cmd/sdlc/localcheck_run.go:206`). Preserve the actual owner when acquisition is reentrant. Sweep all repeated suite executions in one process. Otherwise, the default baseline can consume its timeout waiting for itself and misleadingly report “main is red (unparsed).”

3. **Important findings:** None additional.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Focused parser, classifier, landing-label, context-guard, local-runner, lock-weight, and recovery-context tests passed in two runs: **5.940s** and **7.441s**.
   - The real lock-helper reproduction used isolated temporary lock files and confirmed the blocker.
   - `integrationFixture` supplies a helper that always prints `skipped` (`cmd/sdlc/landingintegrate_test.go:36`), so the new baseline tests cannot catch lock reentry failures.
   - Pinned-range `git diff --check` passed. The full repository suite was not rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** baseline execution reuses the existing runner.
   - **ARCH-PURE — pass:** parsing and classification require no IO or mocks.
   - **ARCH-PURPOSE — flag:** the default baseline fails to deliver meaningful comparison.
   - **ARCH-MOCK — flag:** the stateless lock helper omits the interaction this change introduces.
   - **ARCH-CONSTRAINTS — flag:** self-wait can consume the baseline’s entire suite timeout.
   - **ARCH-SECURE — pass:** execution uses argv, and diagnostic runs publish no status.
   - **ARCH-ORDER — flag:** repeated acquisitions fail to preserve authoritative lock ownership.
   - **ARCH-FUNERAL — pass:** baseline logs reuse existing retention; worktrees reuse existing cleanup.

7. **Plan revision recommendation**

   Add a `## Revisions` entry specifying that every suite run in one `sdlc` process preserves the actual lock owner through acquisition and subprocess execution. Enumerate baseline, integration retries, and multiple named checks, with regression tests that fail when owner propagation is removed.

---

## Re-review — 2026-10-14T02:35:05-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 332 — sdlc merge integrates a moved main itself, then re-runs the suite before landing |
| repo | ariadne |
| issue file | workshop/issues/000332-sdlc-merge-integrates-a-moved-main-itself-then-re-runs-the-suite-before-landing.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 05de5db107a4db2927f2e6e424e4bedb51d28c03..cc93bb2aeae75a159e754ce3cd42f33bca61fded |
| command | sdlc milestone-close --issue 332 --milestone M3 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T02:35:05-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

M3 delivers the documented failure classification and diagnostic baseline without permitting a failed suite to land. BR-15 is addressed with reachable regression tests that fail when the correction is removed. No new blocking findings emerged from the pinned range.

```findings
dispose:
  - id: BR-15
    disposition: addressed
    note: |
      cmd/sdlc/localcheck_run.go:224-241 preserves lock ownership on reentry. The real-helper baseline and repeated-check tests pass on the pinned head; restoring the empty reentrant holder in a disposable clone makes both fail with nested lock waits (3.220s). Coverage includes the baseline, successive heads, and multiple named checks. The clone was removed.
```

1. **Strengths**

   - Typed failure propagation avoids interpreting refusal text (`cmd/sdlc/cigate.go:385`).
   - The baseline uses the same selected check, suppresses status/comment publication, and preserves the original refusal (`cmd/sdlc/suitebaseline.go:25`).
   - Parser and classifier logic remain pure, with direct tests for package attribution, deduplication, and unparsed failures.
   - README and atlas describe the new diagnostic behavior.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Focused diagnostic, ownership, context, and recovery tests passed in 25.556s. Broader landing/local-check regressions passed in 30.080s. Parser fuzzing passed with 4,026 executions. The BR-15 mutation failed both regression tests as required. Diff checks passed and the checkout remains clean. I did not rerun the full suite.

6. **Architectural notes**

   - **ARCH-DRY — pass:** baseline reuses the existing runner and selected suite.
   - **ARCH-PURE — pass:** parsing/classification are separate from diagnostic IO.
   - **ARCH-PURPOSE — pass:** classification, unknown-result handling, and refusal are delivered.
   - **ARCH-MOCK — pass:** stateful Git/GitHub fixtures and the real isolated lock helper exercise production seams; existing live conformance remains available.
   - **ARCH-CONSTRAINTS — pass:** sequential baseline execution retains existing suite bounds.
   - **ARCH-SECURE — pass:** read failures remain visible; diagnostic execution cannot publish main status.
   - **ARCH-ORDER — pass:** gate rejection precedes diagnostics; diagnostics cannot authorize landing.
   - **ARCH-FUNERAL — pass:** baseline logs reuse capped storage and temporary-worktree cleanup.

7. **Plan revision recommendations:** None. R9–R11 describe the delivered M3 behavior and ownership correction accurately.
