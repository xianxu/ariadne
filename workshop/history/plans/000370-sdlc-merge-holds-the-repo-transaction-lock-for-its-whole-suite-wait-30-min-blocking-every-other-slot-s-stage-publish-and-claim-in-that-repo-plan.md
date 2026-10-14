# Merge Transaction Lock Scoping Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let other slots use the repository transaction lock while `sdlc merge` waits for CI or a full suite, without allowing a stale head or main to land.

**Architecture:** Change `merge` from Cobra's whole-command lock to the existing manual lock helper. Durable and legacy landing each capture pinned Git/PR evidence under short lock scopes, release for their CI/local-check work, and revalidate before publication. Durable integration additionally releases for replay review and GitHub registration polling; its pure integration state owns pre-push main drift and bounded retry.

**Tech Stack:** Go, Cobra, Git worktrees, stateful local Git/GitHub test fixtures.

---

**Observed failure:** A live `sdlc merge` held `.git/sdlc.lock` while waiting for and running its first suite. Main advanced, so the same merge started a second suite and continued holding the lock. Independent `stage --clear`, issue filing, claim and branch handoff all waited behind it.

**Operating envelope (ARCH-CONSTRAINTS):** A repo transaction may span Git/network mutations measured in seconds, while a suite queue/run can take 30+ minutes and can repeat. No suite or CI poll may hold the repo lock. A cancellation releases an acquired lock; a lock reacquisition waits under the command context and refuses before any effect if it fails. No new concurrency or artifact family is introduced.

**Non-goals:** This issue does not change machine-wide suite scheduling, cross-machine serialization, the existing legacy merge refusal when main moved, or GitHub's merge policy. A new journal would duplicate the already durable Git/PR/archive evidence (ARCH-DRY, ARCH-FUNERAL).

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `integrationState` | `cmd/sdlc/integrationstate.go` | modified |

`integrationState` remains the sole owner of integrated head, observed main, retry count and phase. Add a main-drift event valid in `integrationPush` that requests reset of the validated but unpublished integration commit, then a reset-completed event that restores `oldHead`, advances the attempt count, and returns to `integrationGates`. The next main observation either starts a fresh integration or hits the existing limit (ARCH-ORDER). It cannot create a second unpublished merge atop the first because `integrationResume` demands the remote head as first parent (`landingintegrate.go:58-60`).

The legacy path pins its branch and HEAD in local variables before validation, then carries those exact values through the unlocked check. Existing Git and PR adapters revalidate them before publication. A changed head, branch or worktree refuses, and a changed main follows legacy's existing manual-integration refusal. `decideMergeAction` remains the pure selector; no parallel legacy state machine is added.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| Merge lock scope | `cmd/sdlc/merge.go`, `cmd/sdlc/repolock.go` | modified | Git-common-dir transaction lock |
| Durable landing | `cmd/sdlc/landing.go`, `cmd/sdlc/landingintegrate.go` | modified | Git refs, remote push, GitHub PR, tracker archive |
| Legacy landing | `cmd/sdlc/merge.go` | modified | Git refs, CI/local check, PR merge, cleanup |
| Stateful landing fixture | `cmd/sdlc/landingintegrate_test.go` | modified | Local bare Git remote, fake GitHub statuses and PR |

The production command supplies `withRequiredRepoTransactionLock`; direct package tests inject a lock callback. GitHub's fake follows real pushes, so lock tests exercise actual Git state and the same phase transitions as production. No new journal is needed: clean merge commit validation, leased push, expected-head merge and archive provenance are existing retry evidence.

The lock callback has one operation: `func(func() error) error`. Production binds it to the Cobra command, so `withRequiredRepoTransactionLock` still owns cancellation, reentrancy and cleanup. A missing callback is allowed only for existing direct unit calls of `runMerge`; the Cobra command must always bind one. This small seam avoids a new lock implementation (ARCH-DRY). No new persistent state or credential is read; remote Git/PR observations remain untrusted until exact OID/identity checks at each side-effect boundary (ARCH-SECURE).

## Task 1: Write red tests for the lock and transition contracts

**Files:** Modify `cmd/sdlc/repolock_test.go`, `cmd/sdlc/integrationstate_test.go`, `cmd/sdlc/landingintegrate_test.go`, `cmd/sdlc/merge_test.go`.

| Function under test | Adversarial input/event class and mechanical guard |
|---------------------|----------------------------------------------------|
| `transitionIntegration` | Main moves after replay acceptance: pure event-sequence test requires reset/retry within the attempt limit and rejects an out-of-phase event. |
| `integrateLandingMain` | Main or head moves while replay is unlocked: real Git fixture verifies exact-ref refusal or reset before any leased push. |
| `runDurableMerge` | A local check blocks on an absolute temp-file signal and main moves between runs: a second command obtains the real Git-common-dir lock during each queue/run; exact-head status and merge assertions guard stale landing. |
| `runMerge` (legacy path) | A required local check blocks: a second command obtains the same lock before the check finishes; changing pinned main/head causes refusal before `PRMerge`. |
| `revalidateLanding` and legacy snapshot validation | Another actor moves a ref or dirties the checkout: equality/cleanliness checks refuse before side effects. |

- [ ] Add the failing manual-lock metadata assertion in `TestRepoLockCommandMetadata`.
- [ ] Write red pure transition tests and real-lock Git/GitHub fixture tests using the existing stateful fakes. Bound waits with context/timeouts and always release blocked subprocesses in test cleanup.
- [ ] Run targeted tests before implementation: `go test ./cmd/sdlc -run 'TestRepoLockCommandMetadata|TestIntegrationState|TestLanding.*Lock|TestLegacy.*Lock' -count=1 -timeout 10m`. Confirm they fail on the missing behavior, not fixture setup.

## Task 2: Scope the command lock and integration transactions

**Files:** Modify `cmd/sdlc/merge.go`, `cmd/sdlc/landing.go`, `cmd/sdlc/landingintegrate.go`, `cmd/sdlc/integrationstate.go`.

- [ ] Mark the Cobra merge command `markManualLockCommand`. Supply a callback using `withRequiredRepoTransactionLock(cmd, fn)` to both durable and legacy paths; preserve direct test injection.
- [ ] Split legacy merge into locked preparation, unlocked `runCIGate`, and locked finalization using pinned local branch and HEAD values. Revalidate immediately after the publish gate and again before PR merge; finalization checks branch/ref/worktree, remote/PR head and fresh main. Legacy's `requireLegacyMainIntegrated` refuses a moved main for manual integration. Keep the existing post-merge cleanup under the final lock. No local check runner is called inside a lock callback.
- [ ] Lock every durable `fetchMain` that writes the shared remote-tracking ref, including the recovery selector and each integration loop observation. During durable integration, lock the local merge or verified resume mutation, including conflict abort; release for publish replay. Reacquire, check workspace identity, exact branch/ref HEAD, clean checkout, remote lease base and freshly fetched main before a leased push. A moved head refuses.
- [ ] For pre-push main drift, transition `integrationPush → integrationReset`, reset the validated unpublished commit to `oldHead` under the lock, then transition `integrationReset → integrationGates` with one attempt consumed. Re-observe main and re-integrate; the limit still bounds repeated drift. A crash before reset is reconciled by `integrationResume`; after reset, local/remote/PR heads match and a rerun starts normally.
- [ ] Release before `syncIntegrationPR`'s bounded GitHub registration polling. Recheck exact head and fresh main before the expensive instance/publish/CI/suite gates and after their completion.
- [ ] Reacquire for the final gate check, expected-head GitHub merge and remote archive/cleanup. If main moved, release and reintegrate; if head, checkout or target changed, refuse before merging. Keep post-merge archive retryable under the lock.
- [ ] Run `gofmt -w` on edited Go files, then targeted landing/repolock tests. Check each lock acquisition releases on success, error and cancellation.

The production binding is equivalent to:

```go
f.Lock = func(run func() error) error {
    return withRequiredRepoTransactionLock(cmd, run)
}
```

The durable loop's checkpoint order is: lock and observe exact head + fresh main; unlock; if needed, lock and create/recover the integration commit; unlock and replay the reviewed change; lock, revalidate head/main and either reset a stale unpublished integration or leased-push it; unlock and poll GitHub registration; lock and capture a fresh main/head gate snapshot; unlock and run gates plus CI/suite; lock and revalidate exact head/main again. Main drift takes the bounded integration loop and discards stale suite authority; head drift refuses before a push or server merge. No code path may call `waitCIGate`, `runCIGate` or `syncIntegrationPR` from inside the callback.

On a crash at any seam, rerun `sdlc merge --branch NAME --yes`: the existing `integrationResume` proves an unpublished merge commit; a lease proves or refuses the push; PR head registration reconciles a completed push; a MERGED PR and archive provenance reconcile landing and cleanup. An uncertain network result leaves the branch intact and surfaces this retry instruction.

## Task 3: Verify phase recovery and update the contract

**Files:** Modify `cmd/sdlc/landingintegrate_test.go`, `cmd/sdlc/merge_test.go`, `cmd/sdlc/helptext/root.md`, `cmd/sdlc/helptext/merge.md`, `atlas/workflow/sdlc-binary.md`; update the issue Log.

- [ ] Extend `integrationResume` tests with interrupted pre-push observations: the adversarial input is an unpushed or partly pushed merge commit; its exact two-parent tree/trailer, remote-first-parent and lease checks must distinguish safe retry from new work.
- [ ] Extend `selectIntegrationPR` and `syncIntegrationPR` tests with delayed PR registration: stateful GH fake remains behind a real pushed ref; registration polling stays unlocked and a retry resolves to the same PR/head.
- [ ] Run the existing conflict, lease, movement-limit and archive recovery tests against the new lock boundary; their invariant is that an uncertain or changed external effect preserves the issue branch.
- [ ] Document the short locked phases and unlocked waits in `sdlc --help`, `sdlc merge --help` and the atlas map. State that retry revalidates head and main; no lock spans a suite queue or run.
- [ ] Run targeted Go tests, `make -k tools`, and the recorded full suite using `sdlc local-check run --name suite`. Inspect real `git diff --check`, exact HEAD via `git rev-parse HEAD`, and the local-check result before closing.

## Review and delivery

- [ ] Record verification and review outcome in the issue Log; close through `sdlc close --issue N --verified '...'` with the mandatory fresh-context review, resolving findings without bypasses.
- [ ] Publish through `sdlc pr`, verify CI and recorded suite, merge with `sdlc merge --yes`, then report the landed PR and real merged hash to the hub.

## Revisions

- 2026-10-14 11:42 PDT — Plan-quality round 1 found the legacy local-check path also executes a suite (`merge.go:570`, `cigate.go:602-605`), the existing pure transition rejects main observation in `integrationPush` (`integrationstate.go:46-48`), and test intent needed named-function strategies. The plan now covers both merge paths, defines reset/retry after pre-push drift, and replaces case inventory with function-level adversarial proofs. It also states non-goals and the 30+ minute operating envelope.
- 2026-10-14 12:12 PDT — During implementation, kept the legacy snapshot in its existing pinned branch/head/main/PR locals instead of introducing a named `legacyMergeSnapshot` type; this preserves the same reacquisition checks with less state duplication. A post-merge check no longer requires the remote feature ref, because `gh pr merge --delete-branch` may remove it before archive. The final GitHub observation moved outside the lock after the expected-head merge request.
- 2026-10-14 12:25 PDT — Close review round 1 found that the legacy path re-read a new remote HEAD after the publish gate and could merge it before the cleanup-only original-head check (BR-1), and that durable PR registration polling lacked the planned fresh-main checkpoint before the suite (BR-2). The implementation now pins and checks one legacy HEAD before effects and re-observes main after registration; both new regression tests failed on the reviewed code and passed after correction. The Core concepts table now reflects the local-variable implementation.
- 2026-10-14 12:48 PDT — Close review round 2 disposed BR-1/BR-2 and found recovery and fixture-lifetime gaps (BR-3/BR-4). A confirmed merged PR can delete its remote branch before local cleanup, so recovery now selects that state before enforcing pushed-head checks and proves integration from the surviving local head and fresh main. New landing still requires the exact remote head. The concurrent blocked-suite fixtures now release, cancel and join merge workers before restoring shared cwd/adapters; a cleanup-order test guards that lifetime.
