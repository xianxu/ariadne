---
collaboration-mode: discussion
---

# Merge and stage lock regression plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach. Steps use checkbox syntax.

**Goal:** Pin command-level evidence that stage clear/build proceeds during merge suite waits while actual mutation transactions remain serialized.

**Architecture:** Keep #370's existing manually scoped merge repository transactions. Exercise production buildRoot/NewMergeCmd and NewStageCmd against real disposable Git worktrees, controlled suite processes and the existing stateful GitHub fake. No runtime change is authorized unless these tests fail; an unexpected failure returns to diagnosis and a revised plan.

**Tech Stack:** Go tests, Cobra, existing repository lock, real local Git/bare origin, shell suite/build fixtures, landingFakeGH and private suite admission.

## Incident and decision

The initial thirty-minute suite-lock claim conflated merge process age with lock duration. Recorded #803 suite ended11:03:30; pair:9's wait was11:07:34–11:11:04. It belongs to a short legitimate post-suite transaction. Exact lock transitions and executable launch revision are unavailable. Do not widen the claim to an older binary or redesign stage exclusion on that evidence.

TL approved retaining the current locking and closing as “not reproduced; regressions pinned” if the command-level matrix passes. Existing durable/legacy unlock tests already pass, but assemble mergeFlags and synthetic manual commands; the new test covers real command construction and root wrapping.

## Core concepts

No new pure production entity: existing landing identity, stage facts, stage lease and issue-head revalidation are unchanged. Existing pure transition tests remain their owner; this issue adds integration coverage only.

| Integration | Lives in | Status | Wraps |
|---|---|---|---|
| Command-level merge/stage fixture | cmd/sdlc/merge_stage_lock_test.go | new test helper | buildRoot and existing integrationFixture |
| Repository transaction exclusion | cmd/sdlc/repolock.go | reused unchanged | common-directory transaction lock |
| Suite barriers | cmd/sdlc/merge_stage_lock_test.go | new test fixture | private admission helper and real suite subprocess |
| landingFakeGH | cmd/sdlc/landing_test.go and existing helpers | reused unchanged | stateful PR heads, pushes, exact-head statuses and merge effects |

The command fixture owns its temporary repository, fixed cwd, root commands, controlled subprocesses and cancellation. It reuses existing Git and GH boundaries rather than adding a production lock abstraction. The scope ends only after every worker has joined (ARCH-MOCK/ARCH-FUNERAL).

## Observable state and actions

Actions: clear stage, stage/build, release the fixture barrier, cancel the fixture command, retry after a refusal.

| State | Visible result | Recovery | Possible loss |
|---|---|---|---|
| Merge queued for capacity | Stage clear/build completes without releasing admission | Release admission | None |
| Merge suite running | Stage clear/build completes without ending suite | Release suite | None |
| Merge mutation transaction owns repo lock | Stage waits, target stays unchanged | Release transaction | None |
| Source head changes while suite is held | Merge refuses before GH merge | Retry on explicitly pushed/reviewed head | None; new work retained |
| Command cancelled | Wait stops and fixture children join | Retry if desired | Only disposable test output |

The existing common repository lock remains the single exclusion resource for shared mutations. No target-only bypass, extra persisted fields or lock order is introduced.

## Execution

### 1. Incident fixture and command path

Files: add cmd/sdlc/merge_stage_lock_test.go; reuse landingintegrate_test.go, merge_main_test.go, repolock_test.go and stage tests.

- [x] Build a disposable integrationFixture with fixed cwd and an origin alias to the existing bare remote. Add a minimal committed Makefile build target whose external marker proves the rebuild happened.
- [x] Establish the stage through the actual root command before starting the merge. Start merge --yes --no-judge --no-validate through executeSDLCTestCommandWith, so NewMergeCmd's context/lock wiring and root wrappers are exercised.
- [x] Use explicit barriers for admission and suite execution; isolate helper resources/environment from the outer suite. Set all globals and environment before goroutines start. Every fixture return cancels/releases and joins all workers.

### 2. Admission and execution matrix

- [x] Write table-driven cases for clear and stage --build while merge is separately blocked at admission and during execution. Use fakeSuiteLockHelper only for bounded admission control, retaining the real acquire helper process and suite subprocess.
- [x] Wait for positive barrier evidence, execute the actual stage command, and assert completion before releasing merge. Assert target branch/head, lease state and build marker; source branch/head remains unchanged by stage.
- [x] Run the new test on current main. It should pass without runtime changes. Temporarily mutate only the root test command's lock mode to whole-command automatic wrapping; the same regression must fail because stage cannot complete before release. Restore it and rerun green. Never commit the intentional mutation.

### 3. Safety controls

- [x] Hold a real short repository transaction across a controlled barrier. Start stage and observe the existing lock-wait diagnostic; assert target remains unchanged, then release and require completion.
- [x] Change the source issue head during a held suite. Release the suite and require actual CLI merge to refuse at reacquisition, with no GitHub merge and the new commit preserved. Reuse the existing head-drift fixture pattern.
- [x] Verify cancellation paths join worker/process lifetimes without waiting on a consumed completion channel. Derive test ceilings from t.Deadline where available; timeout output includes goroutines and named barrier.

### 4. Verification and close

- [x] Run focused command-lock tests, existing durable/legacy unlock and head-drift tests, plus stage lease/clear tests. Record exact commands and results.
- [x] Add the incident's process-age-versus-lock-duration lesson without claiming unrecoverable timing evidence. Update atlas only if an existing explanation needs correction; otherwise acknowledge no new architecture at close.
- [ ] Push and run the recorded suite on the exact head. Mandatory sdlc close provides the fresh-context review. Close as “not reproduced; regressions pinned” if all tests pass; ask the hub for a revised design if any real locking defect is exposed.
- [ ] Land via sdlc merge --yes on SHIP and report PR/result. No stage on the live fleet is proposed: this is deterministic real-Git/process conformance; ask the TL to confirm that verification choice.

## Risk and guard table

| Risky function/path | Adversarial ordering | Independent guard |
|---|---|---|
| buildRoot/NewMergeCmd/root wrapper | Hold admission and execution separately, concurrent actual stage | Stage finishes before barrier release; mutation test restores whole-command locking and fails |
| withRequiredRepoTransactionLock/NewStageCmd | Short transaction held before stage | Existing wait diagnostic, unchanged target until release |
| revalidateLanding after suite | Source head changes during held suite | Zero GH merges and changed head retained |
| fixture worker cleanup | Early assertion or cancellation | Worker join before cwd/global teardown and bounded process exit |

## Operating envelope and principles

This is bounded integration-test work, not a new hot path. Four action/phase cases use a constant number of disposable worktrees and subprocesses per case; private admission avoids occupying additional real machine suite capacity. Waits use barriers and test deadlines, not elapsed-time performance assertions. No N-dependent runtime path or new production scan exists.

ARCH-DRY: keep the existing scoped transaction owner. ARCH-PURE: no new business logic is hidden in IO. ARCH-PURPOSE: test actual CLI wiring, not another synthetic lock callback. ARCH-MOCK: reuse stateful landingFakeGH and real local Git. ARCH-CONSTRAINTS: bounded cases/private capacity/deadline-derived joins. ARCH-SECURE: preserve transaction exclusion and no peer mutations. ARCH-ORDER: assert both unlocked waits and locked mutations with forced order. ARCH-FUNERAL: fixture cleanup owns all temporary state/processes. ARCH-EVOLVE: no persisted format changes.
