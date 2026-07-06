# cmd/sdlc Test Hermeticity: Isolated Repo Lock + a Real-Repo Mutation Guard — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `cmd/sdlc` tests hermetic w.r.t. the *real* ariadne repo — (a) command-tree tests that acquire the repo transaction lock must lock in a per-test temp repo, not the developer's `.git/sdlc.lock` (#149); and (b) a package-level `TestMain` backstop must fail loudly if any test mutates the real repo's HEAD/branch/tree or grabs the real lock (#165, folded in — it would have caught the incident that corrupted `main` this session).

**Architecture:** Both bugs share one root cause — a `cmd/sdlc` test invokes sdlc code that resolves "the repo" from the process cwd (`git rev-parse --git-common-dir`), and cwd is `cmd/sdlc/`, *inside* the real repo. Fix = the offending tests run inside an isolated temp git repo (so resolution finds the temp `.git`), and a `TestMain` guard catches any test that still leaks. The guard's decision is a **pure** snapshot-diff; the git IO is the thin shell.

**Tech Stack:** Go 1.26, `testing.TestMain`, git.

**Architecture principles (`sdlc arch-principles`):**
- **ARCH-PURE** — the guard's core is a pure `repoSnapshot` + `snapshotDiff(before, after)` (unit-tested, no IO); `TestMain` is the thin IO shell (git calls + `m.Run()` + `os.Exit`).
- **ARCH-DRY** — reuse the existing chdir-into-temp-repo pattern (collectdiff/closereview/close tests already do it) for the lock-offender fix rather than inventing a parallel one; the single guard covers **both** the #149 (real-lock) and #165 (real-tree) symptom classes instead of two mechanisms.
- **ARCH-PURPOSE** — deliver both halves: the lock isolation (#149 Done-when) AND the mutation backstop (#165). The guard is proven by a deliberately-polluting test, not merely asserted.

---

## Root cause (verified)

`repoLockGitCommonDir()` (`cmd/sdlc/repolock.go:115`) runs `git rev-parse --git-common-dir` from cwd; `repolock.Acquire` (`internal/repolock/repolock.go:143`) locks at `<GitCommonDir>/sdlc.lock`. A mutating command (`markMutatingCommand`) driven via `buildRoot().Execute()` in a test that has **not** chdir'd into a temp git repo resolves the **real** repo's `.git` → grabs the real `.git/sdlc.lock`. Confirmed: `TestSetStatusAlias_BothPathsMutate` (`issue_test.go`) sets a temp `--issues-dir` but never chdir's, so `set-status` (mutating) hits the real lock (#149's hang). The same non-isolation is what let a stray git sequence mutate the real tree this session (#165).

There is **no real `TestMain`** in the package yet (a test merely *named* `TestMainHas…` exists) — clean slate for the guard.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `repoSnapshot` (head, branch, porcelain set, lockPresent) | `cmd/sdlc/testmain_test.go` | new |
| `snapshotDiff(before, after repoSnapshot) []string` | `cmd/sdlc/testmain_test.go` | new |

- **`repoSnapshot`** — a value capturing the real repo's `HEAD` sha, current branch, the set of `git status --porcelain` lines, and whether `.git/sdlc.lock` exists. Plain data.
- **`snapshotDiff`** — pure: returns human-readable descriptions of what a test run CHANGED — HEAD moved, branch switched, NEW porcelain entries (entries in `after` not in `before` — so **pre-existing** untracked files like the session's parley/pensive notes do NOT trip it), or the real lock appearing. Empty slice = hermetic. Unit-tested directly.
  - **DRY rationale:** one diff covers both symptom classes (tree change *and* real-lock appearance).
  - **Future extension:** could grow to snapshot other side-effect surfaces (config, refs) if a new leak class appears.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `TestMain(m *testing.M)` | `cmd/sdlc/testmain_test.go` | new | git IO + test runner |
| hermetic temp-repo setup for lock-offender tests | `cmd/sdlc/issue_test.go` (+ any peer) | modified | git init + chdir |

- **`TestMain`** — resolves the real repo root from the **initial** cwd (before any test chdir) via `git -C <cwd> rev-parse --show-toplevel`; takes `before := snapshot(root)`, runs `code := m.Run()`, takes `after := snapshot(root)`; if `code == 0` but `snapshotDiff(before, after)` is non-empty, prints a loud failure naming each change and exits non-zero. Always queries with `-C <root>` so a test's stray chdir can't fool it. If root can't be resolved (not in a git repo — e.g. CI tarball), skip the guard gracefully (don't fail the suite for the wrong reason).
  - **Injected into:** nothing — it's the package entry point. Its *logic* is `snapshotDiff` (pure, injected data).
- **hermetic temp-repo setup** — the lock-offender tests chdir into an inited temp git repo (mirroring `collectdiff_test`/`closereview_test`), so `repoLockGitCommonDir()` resolves the temp `.git`. Prefer a tiny shared helper if >1 test needs it.

---

## Task 1: Pure guard core — `repoSnapshot` + `snapshotDiff` (TDD)

**Files:** Create `cmd/sdlc/testmain_test.go`.

- [ ] **Step 1: failing unit test** for `snapshotDiff`:
  - identical snapshots → empty.
  - HEAD changed → one entry mentioning "HEAD".
  - branch changed → entry mentioning the branch names.
  - a NEW porcelain line (in after, not before) → entry with the path; a line present in BOTH (pre-existing untracked) → NOT reported.
  - `lockPresent` false→true → entry mentioning `.git/sdlc.lock`.
- [ ] **Step 2:** run → FAIL (undefined).
- [ ] **Step 3:** implement `repoSnapshot` (struct) + pure `snapshotDiff`. Represent porcelain as a `map[string]bool` set so "new entries only" is a set difference.
- [ ] **Step 4:** run → PASS.
- [ ] **Step 5:** commit — `#149: pure repoSnapshot + snapshotDiff (guard core)`.

## Task 2: `TestMain` backstop (the IO shell)

**Files:** `cmd/sdlc/testmain_test.go`.

- [ ] **Step 1:** add `func TestMain(m *testing.M)`: resolve real root from initial cwd; `before := readSnapshot(root)`; `code := m.Run()`; `after := readSnapshot(root)`; if `code == 0 && len(snapshotDiff(before, after)) > 0` → `fmt.Fprintln(os.Stderr, loud message with each change)` and `code = 1`; `os.Exit(code)`. Graceful skip if root unresolved.
  - `readSnapshot(root)` = the thin git IO: `git -C root rev-parse HEAD`, `git -C root branch --show-current`, `git -C root status --porcelain`, `os.Stat(root/.git/sdlc.lock)`.
- [ ] **Step 2:** run `go test ./cmd/sdlc/` — expect it to now **FAIL**, surfacing the real offenders (at minimum `TestSetStatusAlias_BothPathsMutate` if it dirties/locks; and any test that mutates the real tree). Record which tests the guard flags — that list drives Task 3.
- [ ] **Step 3:** commit — `#149: TestMain guard — fail if a test mutates the real repo or grabs the real lock`.

## Task 3: Isolate the lock/tree offenders the guard surfaced

**Files:** `cmd/sdlc/issue_test.go` (`TestSetStatusAlias_BothPathsMutate`) + any peer the guard flags.

- [ ] **Step 1:** for each flagged test, chdir into an inited temp git repo before `buildRoot().Execute()` (mirror `closereview_test.go:26-27`: `t.TempDir()` → `git init -b main` + identity + gpgsign off → `os.Chdir` with `t.Cleanup` restore). Point `--issues-dir` at a dir inside that temp repo. Now the mutating command's lock resolves to the temp `.git`, and any writes land in the temp tree.
  - If ≥2 tests need it, extract a `hermeticRepo(t) (dir string)` helper (ARCH-DRY) rather than copy-paste.
- [ ] **Step 2:** re-run `go test ./cmd/sdlc/` — the guard now passes (no real-repo mutation, no real lock). Green.
- [ ] **Step 3:** commit — `#149: isolate command-tree lock offenders into a temp repo`.

## Task 4: Prove the guard actually fires

**Files:** `cmd/sdlc/testmain_test.go`.

- [ ] **Step 1:** add `TestGuard_FiresOnRealRepoMutation` gated behind an env flag (e.g. `if os.Getenv("SDLC_TEST_POLLUTE") == ""` → `t.Skip`). When enabled it deliberately touches the real repo (a harmless `git status`-visible change in a temp file at root, immediately reverted, OR a HEAD-move it undoes) — enough that `snapshotDiff` would report it.
  - Simpler + safe: unit-test the guard's decision path directly — construct before/after snapshots representing a mutation and assert `TestMain`'s decision (via a small extracted `guardVerdict(before, after, code) int`) returns non-zero. This proves the FIRE without risking the real repo. **Prefer this** (pure, zero real-repo risk); the env-gated live version is optional.
- [ ] **Step 2:** run it (`SDLC_TEST_POLLUTE=1 go test -run TestGuard_Fires` for the live variant, or plain for the pure variant) → confirms the guard reports the mutation.
- [ ] **Step 3:** commit — `#149: prove the mutation guard fires`.

## Task 5: Verify #149 lock isolation + full suite + close

- [ ] **Step 1:** `go build ./... && go vet ./...` clean.
- [ ] **Step 2:** full suite `go test ./...` green — and the `cmd/sdlc` guard passes (no real-repo mutation).
- [ ] **Step 3:** #149 acceptance — assert the offender test's lock is temp-rooted: a focused test that, with cwd = an inited temp repo, `repoLockGitCommonDir()` returns a path under the temp dir (not the real checkout). This is the hermetic proof for "go test runs green while a live sdlc holds the real lock" without spawning a concurrent holder (the tests simply never touch the real lock).
- [ ] **Step 4:** atlas — note in `atlas/workflow/sdlc-binary.md` (or the testing/vocab atlas) that `cmd/sdlc`'s `TestMain` guards real-repo hermeticity and command-tree tests lock in a temp repo. (If no natural home, `--no-atlas` with rationale: test-infra, no production surface.)
- [ ] **Step 5:** `sdlc close --issue 149 --verified '<go test ./... green + guard passes; TestSetStatusAlias (+peers) lock in a temp repo; snapshotDiff unit-tested incl. pre-existing-untracked-not-tripped; guard proven to fire; folds #165>'` (let it compute `--actual`). Single-pass — no `Mx`. The boundary review auto-dispatches; log the `Review-Verdict:`.
- [ ] **Step 6:** in the close/Log, note #165 is delivered here (its Done-when — a TestMain that fails on real-repo mutation — is met).

---

## Revisions

### 2026-07-05 — pure-entities table completeness

The Core-concepts pure-entities table lists `repoSnapshot` + `snapshotDiff`; the
implementation also ships a third pure entity, `guardVerdict(before, after, code)
(int, []string)` (extracted per Task 4 — the exit-code decision that fires on a
passing+mutated run and preserves an already-failing run's code). It lives in
`cmd/sdlc/testmain_test.go` and is unit-tested by `TestGuardVerdict`. Recorded here
per the close-review's completeness note (not drift — every listed row matched code).
