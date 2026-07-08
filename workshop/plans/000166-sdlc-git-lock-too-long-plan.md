# sdlc Git Lock Too Long Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Shorten `.git/sdlc.lock` hold time for `sdlc close` / `sdlc milestone-close` by releasing it during long boundary-review dispatch while preserving serialized repo mutation.

**Architecture:** Keep one lock implementation in `cmd/sdlc/repolock.go` (ARCH-DRY). Add a manual-lock command mode for commands whose critical sections are narrower than their full `RunE`, and have close/milestone-close run compute and finalization under explicit lock sections while the external judge runs unlocked. Before finalization, validate that the reviewed HEAD and issue file are unchanged so an unlocked review cannot finalize stale state (ARCH-PURPOSE).

**Tech Stack:** Go, Cobra command annotations, existing `cmd/sdlc/internal/repolock`, existing `judge.Run` seam, hermetic git test repos.

---

## Core Concepts

### Pure Entities

| Name | Lives in | Status |
|------|----------|--------|
| `RepoLockMode` | `cmd/sdlc/repolock.go` | new |
| `CloseReviewSnapshot` | `cmd/sdlc/close.go` | new |

**RepoLockMode** — command annotation value that distinguishes automatic whole-command locking from manual phase locking.

- **Relationships:** 1:1 with a Cobra command that needs repo serialization.
- **DRY rationale:** Reuses the existing command annotation registry instead of creating a separate list of phase-locked commands.
- **Future extensions:** Other long-running mutating commands can opt into manual mode without changing the lock primitive.

**CloseReviewSnapshot** — the reviewed state captured before dispatch and checked before finalization.

- **Relationships:** 1:1 with a boundary review dispatch; owns the reviewed HEAD SHA, original issue text, and original project text when a project edit is prepared.
- **DRY rationale:** Gives both whole-issue close and milestone-close the same stale-review guard.
- **Future extensions:** Can grow to include any additional repo files whose writes are prepared before review dispatch.

### Integration Points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `withRequiredRepoTransactionLock` | `cmd/sdlc/repolock.go` | new | `.git/sdlc.lock` acquisition/release |
| `runCloseWithReviewLocked` | `cmd/sdlc/close.go` | new | close command `RunE` |
| `runMilestoneCloseLocked` | `cmd/sdlc/milestoneclose.go` | new | milestone-close command `RunE` |

**withRequiredRepoTransactionLock** — explicit critical-section helper for manual-lock commands.

- **Injected into:** close/milestone command runners through the existing Cobra command context.
- **Future extensions:** Reusable by any command that needs multiple lock sections in one invocation.

**runCloseWithReviewLocked** — command-level orchestration that computes under lock, dispatches outside lock, then finalizes under lock.

- **Injected into:** `NewCloseCmd().RunE`.
- **Future extensions:** Can be folded back into `runCloseWithReview` if tests no longer need the unlocked convenience runner.

**runMilestoneCloseLocked** — milestone equivalent of `runCloseWithReviewLocked`.

- **Injected into:** `NewMilestoneCloseCmd().RunE`.
- **Future extensions:** Same finalization helper as close once duplication is visible.

## Chunk 1: Manual Lock Mode

### Task 1: Teach the lock wrapper about manual commands

**Files:**
- Modify: `cmd/sdlc/repolock.go`
- Modify: `cmd/sdlc/repolock_test.go`

- [x] **Step 1: Write the failing test**

Add a test proving a manually locked command still reports `commandNeedsRepoLock(cmd) == true` but `wrapRepoLockCommands` does not automatically acquire the lock for its whole `RunE`.

- [x] **Step 2: Run the focused test**

Run: `go test ./cmd/sdlc -run 'TestRepoLockManual|TestWrapRepoLockCommands' -count=1`

Expected: FAIL because all marked commands are currently auto-wrapped.

- [x] **Step 3: Implement manual mode**

Replace the boolean annotation value with lock modes:

- `repoLockAuto` for existing `markMutatingCommand`.
- `repoLockManual` for new `markManualLockCommand`.
- `commandNeedsRepoLock` returns true for either.
- `wrapRepoLockCommands` wraps only auto mode.

Refactor the common acquire/release body so `withRepoTransactionLock` and `withRequiredRepoTransactionLock` share the same implementation.

- [x] **Step 4: Run the focused lock tests**

Run: `go test ./cmd/sdlc -run 'TestRepoLock|TestWrapRepoLockCommands' -count=1`

Expected: PASS.

## Chunk 2: Close Review Unlocking

### Task 2: Make close/milestone-close phase locked

**Files:**
- Modify: `cmd/sdlc/close.go`
- Modify: `cmd/sdlc/milestoneclose.go`
- Modify: `cmd/sdlc/repolock_test.go`

- [x] **Step 1: Write the failing command metadata test**

Update `TestRepoLockCommandMetadata` to assert `close` and `milestone-close` still need the repo lock but are manual-lock commands.

- [x] **Step 2: Write the failing behavioral test**

In `cmd/sdlc/closereview_test.go`, add a test with:

- a hermetic `closeRepo`;
- a `judge.Run` stub that blocks on a channel after starting;
- a lock-acquire stub that records acquire/release events;
- execution through the Cobra command path, not direct `runCloseWithReview`.

The assertion: while the judge stub is blocked, the lock has been released; after the judge returns `VERDICT: SHIP`, finalization reacquires and releases the lock.

- [x] **Step 3: Run the failing tests**

Run: `go test ./cmd/sdlc -run 'TestRepoLockCommandMetadata|TestCloseWithReviewReleasesLockDuringBoundaryReview' -count=1`

Expected: FAIL because close/milestone-close are still whole-command wrapped.

- [x] **Step 4: Implement phase locking**

Change `NewCloseCmd` and `NewMilestoneCloseCmd` to use `markManualLockCommand`.

Add:

- `runCloseWithReviewLocked(cmd, stdout, stderr, f)`.
- `runMilestoneCloseLocked(cmd, stdout, stderr, f)`.

Keep existing direct helpers for unit tests, but make the command path use the locked variants.

- [x] **Step 5: Run focused close/repolock tests**

Run: `go test ./cmd/sdlc -run 'TestRepoLock|TestRunCloseWithReview|TestRunMilestoneClose|TestCloseWithReviewReleasesLockDuringBoundaryReview' -count=1`

Expected: PASS.

## Chunk 3: Stale Review Guard

### Task 3: Refuse to finalize if state changed during unlocked review

**Files:**
- Modify: `cmd/sdlc/close.go`
- Modify: `cmd/sdlc/close_finalize_test.go`

- [x] **Step 1: Write failing tests for stale finalization**

Add tests that mutate HEAD or the issue file while the judge stub is blocked. After the stub returns a finalizing verdict, the command must return an error and must not write `status: codecomplete`, close log lines, or milestone ticks.

- [x] **Step 2: Run stale-guard tests**

Run: `go test ./cmd/sdlc -run 'TestCloseWithReview.*ChangedDuringBoundaryReview|TestMilestoneClose.*ChangedDuringBoundaryReview' -count=1`

Expected: FAIL because finalization currently trusts the pre-review `closeResult`.

- [x] **Step 3: Implement the guard**

Capture the reviewed HEAD SHA and original issue text before dispatch. In the finalization lock section, re-read HEAD and the issue file. If either differs, emit the review trailer for traceability, warn that the close was not finalized, and return an error instructing the operator to rerun the close.

- [x] **Step 4: Run focused tests**

Run: `go test ./cmd/sdlc -run 'TestRunCloseWithReview|TestRunMilestoneClose|TestCloseWithReview.*ChangedDuringBoundaryReview|TestMilestoneClose.*ChangedDuringBoundaryReview' -count=1`

Expected: PASS.

## Chunk 4: Docs and Verification

### Task 4: Update docs and run verification

**Files:**
- Modify: `cmd/sdlc/helptext/root.md` or lock-related help text if the command contract mentions whole-command lock behavior.
- Modify: `workshop/issues/000166-sdlc-git-lock-too-long.md`
- Modify: `workshop/plans/000166-sdlc-git-lock-too-long-plan.md`

- [x] **Step 1: Update docs**

Adjust lock prose from “close/milestone-close hold the lock during long review transactions” to “close/milestone-close release the lock during external review dispatch and reacquire before finalization.”

- [x] **Step 2: Run targeted tests**

Run: `go test ./cmd/sdlc -count=1`

Expected: PASS.

- [x] **Step 3: Run repository verification**

Run: `go test ./...`

Expected: PASS.

- [x] **Step 4: Format and diff-check**

Run: `gofmt -w cmd/sdlc/repolock.go cmd/sdlc/repolock_test.go cmd/sdlc/close.go cmd/sdlc/closereview_test.go cmd/sdlc/close_finalize_test.go cmd/sdlc/milestoneclose.go`

Run: `git diff --check`

Expected: no output.

## Revisions

### 2026-07-07 — close-review REWORK

- Reason: close boundary review found stale validation covered HEAD/issue state but not precomputed project-file edits; it also found the `RepoLockMode` entity was represented as untyped string constants.
- Delta: `CloseReviewSnapshot` now validates original project-file text whenever close prepares a project edit, and `RepoLockMode` is implemented as a typed internal mode in `repolock.go`.
