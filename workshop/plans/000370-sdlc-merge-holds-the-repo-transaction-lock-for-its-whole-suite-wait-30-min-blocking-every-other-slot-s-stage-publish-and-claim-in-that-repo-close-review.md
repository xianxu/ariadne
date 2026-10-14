# Boundary Review — ariadne#370 (whole-issue close)

| field | value |
|-------|-------|
| issue | 370 — sdlc merge holds the repo transaction lock for its whole suite wait (30+ min), blocking every other slot's stage, publish and claim in that repo |
| repo | ariadne |
| issue file | workshop/issues/000370-sdlc-merge-holds-the-repo-transaction-lock-for-its-whole-suite-wait-30-min-blocking-every-other-slot-s-stage-publish-and-claim-in-that-repo.md |
| boundary | whole-issue close |
| milestone | — |
| window | 675dccf184b69de76bf1f8228c2b818459794a75..7f42a98077dc560b215f99c01c49b22f35bea3fc |
| command | sdlc close --issue 370 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T12:24:05-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The lock scoping works for suite execution and GitHub registration polling, and the targeted tests pass. Shipping remains blocked by a legacy HEAD-pinning race and a missing durable checkpoint required by the plan.

```findings
findings:
  - id: new
    severity: Critical
    family: pinned-evidence-across-lock-gaps
    title: |
      Legacy merge can adopt a newer HEAD after publish validation
    detail: |
      cmd/sdlc/merge.go:582 re-reads origin/<branch> after the unlocked publish gate and confirmation windows, then validates and merges against that newer head at lines 608–622. If another actor commits and pushes during that gap, the new commit can pass CI and reach PRMerge without repeating instance/publish validation; the original branchHead is checked only during cleanup at line 663, after the irreversible merge. ARCH-ORDER/ARCH-SECURE: pin one head before validation, preserve it across every unlocked phase, and refuse drift before effects; add a regression that commits and pushes after publish validation and asserts zero merge calls.
  - id: new
    severity: Important
    family: pre-gate-freshness
    title: |
      Durable merge omits the planned checkpoint after registration polling
    detail: |
      cmd/sdlc/landing.go:444–478 proceeds from integrateLandingMain, including unlocked syncIntegrationPR polling, directly into gates using state.main. The plan explicitly requires a locked exact-head/worktree check and fresh-main observation after polling and before expensive gates (plan lines 69 and 81); main movement during registration therefore starts an already-stale suite and is detected only afterward. ARCH-CONSTRAINTS/ARCH-PURPOSE: add that checkpoint, reintegrate before gates when main moved, and test main advancement during delayed registration.
```

1. **Strengths**

   - Reuses the existing transaction-lock helper and binds it in the production Cobra command.
   - Final durable validation and expected-head merge share one lock scope (`cmd/sdlc/landing.go:487`).
   - Pre-push main drift has an explicit reset/retry transition, with pure transition tests and real Git coverage.

2. **Critical findings**

   The legacy race above violates the promised immutable snapshot. The Core concepts table also names `legacyMergeSnapshot`, which does not exist; the revision records a switch to locals, but those locals currently carry different HEAD values for merge and cleanup. Correct both the implementation and the table.

3. **Important findings**

   Add the post-registration checkpoint above. Its regression should prove that the first suite runs against a head containing main’s registration-time advancement.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed:

   - `go test ./cmd/sdlc -run 'TestRepoLockCommandMetadata|TestIntegrationState|TestLanding|TestRunMerge_Legacy' -count=1 -timeout 5m` — 191 seconds.
   - `git diff --check` for the pinned range.

   Existing tests cover blocked-suite lock acquisition, repeated suites, registration polling, drift during checks, conflicts, leases, and interrupted integrations. They omit the two interleavings identified above. I did not rerun the full suite.

6. **Architectural notes**

   - **ARCH-DRY — pass:** existing lock and Git/GitHub seams are reused.
   - **ARCH-PURE — pass:** transition logic remains pure and directly tested.
   - **ARCH-PURPOSE — flag:** the promised post-registration checkpoint is missing.
   - **ARCH-MOCK — pass:** stateful Git/GitHub fixtures exercise the production boundaries; no new external dependency is introduced.
   - **ARCH-CONSTRAINTS — flag:** registration-time drift can unnecessarily start another lengthy suite.
   - **ARCH-SECURE — flag:** legacy validation evidence can authorize a different commit.
   - **ARCH-ORDER — flag:** legacy HEAD authority changes between phases.
   - **ARCH-FUNERAL — pass:** no new durable artifact family; existing branch/archive cleanup remains.

   Atlas and command help document the changed flow. No new command, flag, or configuration surface requires a README update.

7. **Plan revision recommendations**

   Append a `## Revisions` entry recording the legacy pinning correction and post-registration checkpoint. Update the Core concepts table and snapshot description to match the actual implementation, resolving the absent `legacyMergeSnapshot` declaration.
