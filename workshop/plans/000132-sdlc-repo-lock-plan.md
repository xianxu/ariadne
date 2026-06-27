# sdlc Repo Transaction Lock Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Serialize local mutating `sdlc` transactions in one checkout with an SDLC-owned lock under `.git`.

**Architecture:** Add one shared transaction-lock integration and wrap mutating Cobra commands at the command boundary. Keep lock-state parsing and stale/holder decisions pure, with filesystem/process details isolated in a thin adapter (`ARCH-PURE`), and drive all covered verbs through one command metadata hook instead of per-command ad hoc lock calls (`ARCH-DRY`). The scope covers every mutating verb named in the issue, with read-only verbs explicitly left lock-free (`ARCH-PURPOSE`). Review-bearing commands (`change-code`, `close`, `milestone-close`, `merge`, `push`) deliberately hold the coarse repo transaction lock across their synchronous judges because those judges can dirty tracked files or precede branch/git mutation; the default wait timeout is sized for that long-held case and status messages tell quick commands that a review/ship transaction is in progress. The lock uses the Git common dir, so linked worktrees of the same repo serialize with each other; that is intentional because they share the issue namespace, object store, and remote refs that the motivating races touched.

**Tech Stack:** Go, Cobra, stdlib filesystem/process APIs, existing `cmd/sdlc` test harnesses.

---

## Core Concepts

### Pure Entities

| Name | Lives in | Status |
|------|----------|--------|
| `LockMetadata` | `cmd/sdlc/internal/repolock/repolock.go` | new |
| `LockObservation` | `cmd/sdlc/internal/repolock/repolock.go` | new |
| `MutatingCommand` | `cmd/sdlc/repolock.go` | new |

- **LockMetadata** - the command, pid, hostname, cwd, argv, and acquisition time written inside `.git/sdlc.lock`.
  - **Relationships:** 1:1 with one held lock directory.
  - **DRY rationale:** All wait/error messages should render the same holder details instead of each verb inventing its own wording.
  - **Future extensions:** Add agent/session IDs without changing command wrapping.

- **LockObservation** - the pure decision result for an existing lock: active, stale-by-missing-process, stale-by-age, or unreadable.
  - **Relationships:** N:1 observations over time for one lock directory.
  - **DRY rationale:** Stale detection and user guidance must be one policy shared by all mutating verbs.
  - **Future extensions:** Platform-specific process checks can widen behind the same decision shape.

- **MutatingCommand** - command annotation that says a Cobra command requires the repo transaction lock.
  - **Relationships:** 1:1 with each mutating leaf command.
  - **DRY rationale:** The root persistent wrapper can enforce the lock from command metadata; command bodies stay focused on workflow logic.
  - **Future extensions:** Add command groups or lock modes if a later verb needs a different resource.
  - **Invariant:** Locking belongs at the Cobra command boundary. Internal calls to `run*` helpers must not acquire a second process lock; `withRepoTransactionLock` is also command-context re-entrant so a future nested Cobra dispatch inheriting the parent `cmd.Context()` is a no-op instead of a self-deadlock. Independent Cobra executions in the same process must use independent contexts and therefore still serialize on the filesystem lock.

### Integration Points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `RepoLock` | `cmd/sdlc/internal/repolock/repolock.go` | new | `.git/sdlc.lock`, process liveness, clock, sleep |
| `withRepoTransactionLock` | `cmd/sdlc/repolock.go` | new | Cobra command execution |

- **RepoLock** - acquires `.git/sdlc.lock` with `mkdir`, writes metadata, waits with timeout/status messages, removes on release, and installs best-effort signal cleanup.
  - **Injected into:** Tests inject root dir, clock, sleeper, PID/hostname/process checks; production uses stdlib adapters.
  - **Future extensions:** More precise stale cleanup or lock breaking can live here without touching command code.
  - **Timeout policy:** Default wait timeout is 30 minutes, not a short Git-index timeout, because `change-code`/`close`/`milestone-close`/`merge`/`push` can hold the transaction while synchronous LLM judges run and while their outputs are committed/amended. Timeout errors must say the lock may be a long-running review/ship transaction and include the recovery path.

- **withRepoTransactionLock** - root-level wrapper that inspects the leaf command metadata and runs the command under `RepoLock` only when needed.
  - **Injected into:** Cobra `RunE` wrapping during root construction.
  - **Future extensions:** Support `--lock-timeout` as a root persistent flag if the default proves too coarse.

## Chunk 1: Lock Helper

### Task 1: Add pure lock metadata and observation tests

**Files:**
- Create: `cmd/sdlc/internal/repolock/repolock.go`
- Create: `cmd/sdlc/internal/repolock/repolock_test.go`

- [ ] **Step 1: Write failing tests for metadata render/parsing and stale decisions**

Add table tests for:
- metadata round-trips through the on-disk file format.
- holder messages include pid and command.
- holder messages identify known long-running review/ship commands as review/ship transactions.
- same-host missing pid classifies as stale.
- different-host missing pid does not auto-stale.
- old lock age classifies as stale with a recovery message.

Run: `go test ./cmd/sdlc/internal/repolock`

Expected: FAIL because the package does not exist.

- [ ] **Step 2: Implement the pure types and decisions**

Implement:
- `type Metadata struct { PID int; Hostname, CWD, Command string; Args []string; StartedAt time.Time }`
- `func Encode(Metadata) []byte`
- `func Decode([]byte) (Metadata, error)`
- `func HolderLine(Metadata) string`
- `func IsLongRunningCommand(Metadata) bool`
- `func Observe(meta Metadata, now time.Time, host string, processAlive func(int) bool, maxAge time.Duration) Observation`

Keep filesystem calls out of these functions (`ARCH-PURE`).

- [ ] **Step 3: Verify the pure helper**

Run: `go test ./cmd/sdlc/internal/repolock`

Expected: PASS.

### Task 2: Add atomic acquire/wait/release integration

**Files:**
- Modify: `cmd/sdlc/internal/repolock/repolock.go`
- Modify: `cmd/sdlc/internal/repolock/repolock_test.go`

- [ ] **Step 1: Write failing integration tests with temp directories**

Add tests for:
- first acquire creates `.git/sdlc.lock` and writes metadata.
- second acquire waits and emits a holder status line.
- long-held review/ship locks emit wording that tells quick commands a review/ship transaction is in progress.
- release removes the lock directory.
- stale lock returns a clear recovery error and does not silently delete a live-looking lock.

Use a fake sleeper/clock so tests do not wait in real time.

Run: `go test ./cmd/sdlc/internal/repolock`

Expected: FAIL because acquire/release are missing.

- [ ] **Step 2: Implement `Acquire` and `Release`**

Implement `Acquire(ctx, opts)` using `os.Mkdir(lockDir, 0700)` as the atomic operation. On success, write metadata. On `EEXIST`, read metadata, emit `waiting for sdlc repo lock held by ...`, wait with a 30-minute default timeout, and return a recovery-oriented error for stale/unreadable/timeout states. If the holder is a known review/ship command, the status/timeout wording must say a long-running review/ship transaction is in progress. `Release` removes the lock directory best-effort.

- [ ] **Step 3: Verify the integration helper**

Run: `go test ./cmd/sdlc/internal/repolock`

Expected: PASS.

## Chunk 2: Command Wrapping

### Task 3: Mark mutating commands and wrap at the root boundary

**Files:**
- Create: `cmd/sdlc/repolock.go`
- Modify: `cmd/sdlc/main.go`
- Modify: command constructors for mutating leaves:
  - `cmd/sdlc/claim.go`
  - `cmd/sdlc/changecode.go`
  - `cmd/sdlc/close.go`
  - `cmd/sdlc/issue.go`
  - `cmd/sdlc/merge.go`
  - `cmd/sdlc/milestoneclose.go`
  - `cmd/sdlc/pr.go`
  - `cmd/sdlc/push.go`

- [ ] **Step 1: Write failing command metadata tests**

Add or extend tests so these commands are lock-marked:
- `claim`
- `change-code`
- `close`
- `issue new`
- `issue set-status`
- flat hidden `set-status`
- hidden `fetch`
- `merge`
- `milestone-close`
- `pr`
- `push`

Assert these commands are not lock-marked:
- `issue list`
- `issue show`
- `issue validate`
- `state`
- `start-plan`
- `actual`
- `active-time`
- `judge`
- `arch-principles`
- `estimate-source`
- root/help commands

Run: `go test ./cmd/sdlc -run 'RepoLock|Command'`

Expected: FAIL because no metadata exists.

- [ ] **Step 2: Implement command annotations**

Add:
- `func markMutatingCommand(*cobra.Command) *cobra.Command`
- `func commandNeedsRepoLock(*cobra.Command) bool`
- `func withRepoTransactionLock(*cobra.Command) error`

Use Cobra annotations or command context instead of string-matching command names. Mark the leaf constructors directly so the mutation contract lives next to each command (`ARCH-DRY`).

- [ ] **Step 3: Wrap `RunE` at registration time**

In `buildRoot`, wrap each command tree after construction so a mutating leaf acquires the lock before its existing `RunE` runs. Resolve the git common dir through the existing `gitx.Capture("rev-parse", "--git-common-dir")` plumbing and fall back to `.git` only when appropriate. Do not lock read-only commands.

The wrapper must be command-context re-entrant: if the current `cmd.Context()` carries the held-lock marker from a parent command dispatch, nested acquisition returns a no-op release function. A separate Cobra execution in another goroutine must use a fresh context and still serialize on the filesystem lock. Document this in `withRepoTransactionLock` so future maintainers do not push locking into `run*` helpers and self-deadlock.

- [ ] **Step 4: Verify metadata and wrapper behavior**

Run: `go test ./cmd/sdlc -run 'RepoLock|Command'`

Expected: PASS.

- [ ] **Step 5: Verify re-entrant command wrapping**

Add a focused test that simulates a mutating command invoking another marked command through Cobra with the inherited parent `cmd.Context()` and asserts the second acquisition is a no-op rather than a timeout. Also assert existing helper-level re-entry paths (`fetch` calling `runIssueNew`, `milestone-close` calling `runClose`) do not acquire twice.

Run: `go test ./cmd/sdlc -run 'RepoLock|Reentrant|Fetch|Milestone'`

Expected: PASS.

### Task 4: Add concurrency coverage for issue allocation and status mutation

**Files:**
- Modify: `cmd/sdlc/issue_test.go`
- Modify: `cmd/sdlc/setstatus_test.go` or add `cmd/sdlc/repolock_integration_test.go`

- [ ] **Step 1: Write a failing same-checkout concurrency test**

Create a temp git repo with a local bare origin and run two `sdlc issue new` command executions concurrently through the Cobra command path with independent command contexts, not by calling `runIssueNew` directly. Assert:
- both commands complete without `.git/index.lock` failures.
- allocated issue IDs are distinct.
- stderr contains at least one lock wait/status line.

Run this in one checkout. Add a separate, smaller test for two linked worktrees only if the common-dir resolution path is not otherwise pinned: both worktrees should resolve the same `.git/sdlc.lock` under the Git common dir and therefore serialize.

Run: `go test ./cmd/sdlc -run TestRepoLockConcurrentIssueNew -count=1`

Expected: FAIL before wrapper integration.

- [ ] **Step 2: Add one mutation-path serialization test**

Use `issue set-status` or `claim` with an injected slow lock holder so the second command waits. Assert the final issue file is valid and the wait message identifies the holder pid/command.

Run: `go test ./cmd/sdlc -run TestRepoLock -count=1`

Expected: FAIL before wrapper integration.

- [ ] **Step 3: Make the tests pass without broad sleeps**

Use channels/fakes where possible. Keep any real subprocess test small and skip only when temp git setup fails before entering the repo.

Run: `go test ./cmd/sdlc -run 'RepoLock|IssueNew|SetStatus' -count=1`

Expected: PASS.

## Chunk 3: Docs and Full Verification

### Task 5: Document lock behavior and recovery

**Files:**
- Modify: `cmd/sdlc/helptext/root.md`
- Modify: relevant command help if needed, likely `cmd/sdlc/helptext/issue.md` or `cmd/sdlc/helptext/claim.md`
- Modify: `atlas/workflow/sdlc-binary.md` or another existing atlas workflow doc if the command transaction model is described there.

- [ ] **Step 1: Add concise user-facing docs**

Document:
- mutating commands take a local repo transaction lock in `.git/sdlc.lock`.
- wait messages show pid/command when metadata is available.
- change-code/close/milestone-close/merge/push may hold the lock for a long-running review/ship transaction; quick commands should wait or retry instead of removing the lock.
- linked worktrees for the same repo share the lock through `git rev-parse --git-common-dir`.
- stale/timeout errors tell the operator how to inspect/remove the lock.
- remote push/ref races remain separate and still need retry guidance.

For Done-when #6, add one focused non-regression assertion around existing push/merge retry guidance if an existing unit test can cover it without a heavy e2e; otherwise log why the existing push/merge tests already exercise that path.

- [ ] **Step 2: Add helptext drift tests if existing tests do not cover the new text**

Run: `go test ./cmd/sdlc -run Helptext`

Expected: PASS.

### Task 6: Final verification

**Files:**
- Modify: `workshop/issues/000132-sdlc-repo-lock.md`

- [ ] **Step 1: Tick completed issue plan rows and add a log entry**

Record the commands run and the lock design notes, citing `ARCH-DRY`, `ARCH-PURE`, and `ARCH-PURPOSE`.

- [ ] **Step 2: Run focused and full tests**

Run:
- `go test ./cmd/sdlc/internal/repolock`
- `go test ./cmd/sdlc -run 'RepoLock|IssueNew|SetStatus|Helptext' -count=1`
- `go test ./cmd/sdlc ./cmd/sdlc/internal/... ./pkg/...`

Expected: PASS.

- [ ] **Step 3: Run one live smoke check**

Run two local `sdlc issue new --dry-run` or temp-repo command invocations concurrently if the test path does not already execute the real binary path. Confirm the lock wait/status line appears and no Git lock error appears.

Expected: serialized execution with holder metadata.

## Revisions

### 2026-06-27 — Boundary review REWORK recovery fixes

Reason: the boundary review found that `die()` calls `os.Exit`, which skips the
wrapper's deferred release, and that waiters could observe the lock directory
before `meta.json` was written.

Delta: add an active-lock cleanup registry drained by `die()` before `os.Exit`;
make release idempotent in the wrapper; add real concurrent `Acquire` coverage;
auto-reclaim confirmed-dead same-host holders; treat missing/partial metadata as
holder initialization during a short grace window instead of an immediate
unreadable-lock failure. This revises the original "never auto-remove" stance:
cross-host and age uncertainty still fail with recovery guidance, but a dead
same-host pid is safe to reclaim.
