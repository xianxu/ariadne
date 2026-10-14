# Merge Transaction Lock Scoping Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let other slots use the repository transaction lock while `sdlc merge` waits for CI or a full suite, without allowing a stale head or main to land.

**Architecture:** Change `merge` from Cobra's whole-command lock to the existing manual lock helper. The durable landing flow captures and revalidates Git state around short integration and final publication transactions; review replay, GitHub registration polling, and the CI/suite gate run outside the lock. Existing integration state and Git/GitHub evidence make retries converge after an interrupted phase.

**Tech Stack:** Go, Cobra, Git worktrees, stateful local Git/GitHub test fixtures.

---

**Observed failure:** A live `sdlc merge` held `.git/sdlc.lock` while waiting for and running its first suite. Main advanced, so the same merge started a second suite and continued holding the lock. Independent `stage --clear`, issue filing, claim and branch handoff all waited behind it.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|

No pure entity is added or changed. `integrationState` remains in `cmd/sdlc/integrationstate.go` and already represents the observations and bounded retries this change needs.

It remains the sole owner of integrated head, observed main, retry count and phase. Reobserving main after an unlocked phase uses its existing transition events; this prevents a second recovery state machine (ARCH-DRY, ARCH-PURE).

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| Merge lock scope | `cmd/sdlc/merge.go`, `cmd/sdlc/repolock.go` | modified | Git-common-dir transaction lock |
| Durable landing | `cmd/sdlc/landing.go`, `cmd/sdlc/landingintegrate.go` | modified | Git refs, remote push, GitHub PR, tracker archive |
| Stateful landing fixture | `cmd/sdlc/landingintegrate_test.go` | modified | Local bare Git remote, fake GitHub statuses and PR |

The production command supplies `withRequiredRepoTransactionLock`; direct package tests inject a lock callback. GitHub's fake follows real pushes, so lock tests exercise actual Git state and the same phase transitions as production. No new journal is needed: clean merge commit validation, leased push, expected-head merge and archive provenance are existing retry evidence.

The lock callback has one operation: `func(func() error) error`. Production binds it to the Cobra command, so `withRequiredRepoTransactionLock` still owns cancellation, reentrancy and cleanup. A missing callback is allowed only for existing direct unit calls of `runMerge`; the Cobra command must always bind one. This small seam avoids a new lock implementation (ARCH-DRY). No pure data shape changes are planned.

## Task 1: Prove the lock regression and drift behavior

**Files:** Modify `cmd/sdlc/repolock_test.go`, `cmd/sdlc/landingintegrate_test.go`; inspect `cmd/sdlc/landing_test.go` fixture.

- [ ] Add a command-metadata test asserting `merge` requires a repo lock but is manually scoped, like `close`.
- [ ] Add a real Git fixture test that blocks the declared suite after it starts, attempts to acquire the same Git-common-dir lock with a second command, and proves acquisition completes before unblocking the suite. Capture a bounded timeout so a regression fails rather than hanging.
- [ ] Add a head-drift test that moves the issue ref during an unlocked suite and expects refusal with the new ref intact and no GitHub merge.
- [ ] Extend main-drift coverage: move main during the unlocked first suite, expect another integration and a second suite on the exact integrated head before landing. Assert a second command acquires the repo lock while either suite is blocked. Include movement during replay-to-push registration if the split exposes that gap.
- [ ] Run targeted tests and confirm they fail on the whole-command lock or missing revalidation: `go test ./cmd/sdlc -run 'TestRepoLockCommandMetadata|TestLanding.*Lock|TestLanding.*Drift' -count=1 -timeout 10m`.

Use `landingFixture` with its local bare remote and stateful `landingFakeGH`. The suite command creates a signal file and waits for a release file in `t.TempDir()`, both absolute paths because the suite runs in a detached worktree. Start `runMerge` with a production-style lock callback in a goroutine, wait for the signal, then call `withRequiredRepoTransactionLock` on another marked command with the same Git common dir. A two-second timeout before creating the release file is the regression assertion. Use a second signal/release pair when main moves to show the lock is free for the repeated suite as well. Keep all goroutines bounded and release files in test cleanup.

## Task 2: Scope the command lock and integration transactions

**Files:** Modify `cmd/sdlc/merge.go`, `cmd/sdlc/landing.go`, `cmd/sdlc/landingintegrate.go`; optionally `cmd/sdlc/integrationstate.go` only if a new drift event is required.

- [ ] Mark the Cobra merge command `markManualLockCommand`. Supply a callback using `withRequiredRepoTransactionLock(cmd, fn)` to the merge flow. Keep legacy merge under one lock, since it has no CI/suite wait, and preserve direct test injection.
- [ ] Lock every `fetchMain` that writes the shared remote-tracking ref, including the recovery selector and each integration loop observation. During durable integration, lock the local merge or verified resume mutation, including conflict abort; release for publish replay. Reacquire, check workspace identity, exact branch/ref HEAD, clean checkout, remote lease base and freshly fetched main before a leased push. A moved head refuses. If main moved, use the existing integration transition and loop so the next gate snapshot is current; never run the suite on an integration that omits fresh main.
- [ ] Release before `syncIntegrationPR`'s bounded GitHub registration polling. Recheck exact head and fresh main before the expensive instance/publish/CI/suite gates and after their completion.
- [ ] Reacquire for the final gate check, expected-head GitHub merge and remote archive/cleanup. If main moved, release and reintegrate; if head, checkout or target changed, refuse before merging. Keep post-merge archive retryable under the lock.
- [ ] Run `gofmt -w` on edited Go files, then targeted landing/repolock tests. Check each lock acquisition releases on success, error and cancellation.

The production binding is equivalent to:

```go
f.Lock = func(run func() error) error {
    return withRequiredRepoTransactionLock(cmd, run)
}
```

The durable loop's checkpoint order is: lock and observe exact head + fresh main; unlock; if needed, lock and create/recover the integration commit; unlock and replay the reviewed change; lock, revalidate head/main and leased push; unlock and poll GitHub registration; lock and capture a fresh main/head gate snapshot; unlock and run gates plus CI/suite; lock and revalidate exact head/main again. Main drift takes the existing bounded integration loop and discards the stale suite authority; head drift refuses before a push or server merge. No code path may call `waitCIGate` or `syncIntegrationPR` from inside the callback.

On a crash at any seam, rerun `sdlc merge --branch NAME --yes`: the existing `integrationResume` proves an unpublished merge commit; a lease proves or refuses the push; PR head registration reconciles a completed push; a MERGED PR and archive provenance reconcile landing and cleanup. An uncertain network result leaves the branch intact and surfaces this retry instruction.

## Task 3: Prove interrupted-phase recovery and update the contract

**Files:** Modify `cmd/sdlc/landingintegrate_test.go`, `cmd/sdlc/helptext/root.md`, `cmd/sdlc/helptext/merge.md`, `atlas/workflow/sdlc-binary.md`; update the issue Log.

- [ ] Test retry after a local integration commit exists but its push did not complete; verify exact clean two-parent merge and replay before leased push. Test retry after push succeeds but GitHub still reports the old head, then registration catches up. Assert no second unreviewed commit is published.
- [ ] Test lock release while GitHub registration polling is delayed, and retain existing tests for conflict abort, lease rejection, main movement limit and archive retry.
- [ ] Document the short locked phases and unlocked waits in `sdlc --help`, `sdlc merge --help` and the atlas map. State that retry revalidates head and main; no lock spans a suite queue or run.
- [ ] Run targeted Go tests, `make -k tools`, and the recorded full suite using `sdlc local-check run --name suite`. Inspect real `git diff --check`, exact HEAD via `git rev-parse HEAD`, and the local-check result before closing.

## Review and delivery

- [ ] Record verification and review outcome in the issue Log; close through `sdlc close --issue N --verified '...'` with the mandatory fresh-context review, resolving findings without bypasses.
- [ ] Publish through `sdlc pr`, verify CI and recorded suite, merge with `sdlc merge --yes`, then report the landed PR and real merged hash to the hub.
