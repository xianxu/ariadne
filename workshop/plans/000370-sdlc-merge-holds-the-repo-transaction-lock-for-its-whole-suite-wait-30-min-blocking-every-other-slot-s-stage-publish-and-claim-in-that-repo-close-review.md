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

---

## Re-review — 2026-10-14T12:45:30-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 370 — sdlc merge holds the repo transaction lock for its whole suite wait (30+ min), blocking every other slot's stage, publish and claim in that repo |
| repo | ariadne |
| issue file | workshop/issues/000370-sdlc-merge-holds-the-repo-transaction-lock-for-its-whole-suite-wait-30-min-blocking-every-other-slot-s-stage-publish-and-claim-in-that-repo.md |
| boundary | whole-issue close |
| milestone | — |
| window | 675dccf184b69de76bf1f8228c2b818459794a75..fb2c0fa0093df3b50bfa0a6631040315c0e91d6c |
| command | sdlc close --issue 370 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T12:45:30-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

Both prior findings are addressed, with independently confirmed red-before/green-after regressions. Lock scoping and freshness checks pass the targeted tests, but the new legacy checkpoint breaks interrupted-merge recovery after remote branch deletion. The new concurrency tests also need cleanup that waits for their workers.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      TestRunMerge_LegacyRefusesHeadMovedAfterPublishGate passes on pinned HEAD. Removing the correction in a disposable checkout produces one PRMerge call before refusal; pinned HEAD produces zero.
  - id: BR-2
    disposition: addressed
    note: |
      TestLandingMainMovesDuringRegistrationBeforeFirstSuite passes on pinned HEAD. Removing the landing.go:450 checkpoint produces two suites, including a stale first suite; pinned HEAD runs one suite containing the advanced main.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: recovery-preconditions-respect-completed-effects
    title: |
      Legacy recovery requires a remote branch that successful merging deletes
    detail: |
      cmd/sdlc/merge.go:418–421 unconditionally calls revalidateLegacyMerge with requireRemote=true before discovering an already-merged PR; its fetch at lines 775–776 fails when the remote feature branch was deleted by production PRMerge's --delete-branch (ghclient.go:199). Extending the existing merged-PR recovery fixture with server-side branch deletion reproduces “fatal: couldn't find remote ref feature”; making this early checkpoint local-only allows cleanup to finish. ARCH-ORDER/ARCH-MOCK: distinguish pre-merge remote-head requirements from confirmed-landing recovery, retain strict checks before a new merge, and add deletion-aware recovery coverage.
  - id: new
    severity: Important
    kind: defect
    family: concurrent-fixtures-own-worker-lifetime
    title: |
      Blocked-suite tests release signals without joining merge workers on failure
    detail: |
      cmd/sdlc/landingintegrate_test.go:137,153 and cmd/sdlc/merge_main_test.go:20,32 start merge goroutines, but cleanup only writes a release file. Fatal startup, lock-acquisition, or completion failures bypass the finished-channel receive, allowing fixture cleanup to restore cwd and package-global adapters while runMerge is still executing Git effects. ARCH-ORDER/ARCH-SECURE: give both workers cancellation contexts and cleanup that releases, cancels, and joins them before restoring fixture state; consolidate this lifecycle helper.
```

1. **Strengths**

   - Production merge uses the existing manual transaction-lock helper.
   - Real Git fixtures verify lock availability during suites and registration polling.
   - Durable main drift uses the pure transition model, with bounded retry and exact-head checks.
   - Atlas and CLI help describe the changed lock behavior.

2. **Critical findings**

   - `merge.go:418`: recovery must tolerate deletion caused by a completed merge. Select the landing state before imposing remote-branch requirements; test interrupted cleanup after server-side deletion.

3. **Important findings**

   - `landingintegrate_test.go:153`, `merge_main_test.go:32`: join workers on every exit path before restoring cwd, adapters, or temporary storage.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Broad landing, legacy merge, transition, and lock-metadata tests passed: **220 seconds**.
   - Focused lock/drift/recovery tests passed: **50.6 seconds**.
   - Both prior regressions failed with their corrections removed.
   - The additional remote-deletion recovery test failed on pinned HEAD and passed with the early remote requirement disabled.
   - `git diff --check` passed. Disposable checkout removed; reviewed files unchanged. Full suite was not rerun.

6. **Architectural notes**

   - **ARCH-DRY — Pass:** production reuses the lock helper; consolidate test-worker cleanup when fixing it.
   - **ARCH-PURE — Pass:** integration transitions remain deterministic and tested without IO.
   - **ARCH-PURPOSE — Flag:** interrupted recovery regresses despite being part of the retained contract.
   - **ARCH-MOCK — Flag:** legacy recovery fixtures omit production remote-branch deletion.
   - **ARCH-CONSTRAINTS — Pass:** expensive waits run unlocked; integration retries remain bounded.
   - **ARCH-SECURE — Flag:** unjoined fixture workers can execute after their environment is restored.
   - **ARCH-ORDER — Flag:** recovery preconditions ignore completed deletion effects; test-worker lifetimes are not bounded by fixture lifetime.
   - **ARCH-FUNERAL — Pass:** no new durable artifact family; existing cleanup remains the intended removal mechanism.

   Core-concept entities exist, and the pure/integration separation matches the implementation. Atlas coverage is present; no new command, flag, or configuration syntax requires a README addition.

7. **Plan revision recommendations**

   Add `## Revisions` entries specifying:

   - Remote-head requirements apply before a new landing; confirmed-landing recovery must tolerate branch deletion and still finish cleanup.
   - Concurrent fixtures must cancel and join workers before restoring shared state, including failure paths.
