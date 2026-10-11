# Isolated Review Checkout Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every boundary review (milestone-close / close) runs its reviewer in a fresh, disposable shared clone at the pinned Head. Where the OS allows it, an OS write fence keeps the reviewer out of the source checkout and its Git state. The round records which guarantee it had.

**Architecture:** A new package `cmd/sdlc/internal/reviewbox` owns three things:
- the checkout lifecycle: create, close, and a crash sweep driven by flocks;
- fence probing;
- a pure argv wrapper, `[sandbox-exec -p P | bwrap …] env -C <checkout> <agent> …`.

`judge.DispatchOptions` gains a `Wrap` hook and a `Fenced` flag. Under the fence, `Fenced` turns off each agent's own sandbox, which cannot nest. Isolation is therefore one argv transform, the 61 existing `judge.Run` fakes keep their signature, and `--dry-run` shows the isolation. `dispatchBoundaryReview` creates the checkout, points the manifest's `RepoRoot` at it, and removes it in a defer. It stamps `isolation: fenced|checkout` on the ledger round and the sidecar.

**Tech Stack:** Go, git (`clone --shared`), macOS Seatbelt (`sandbox-exec`), Linux bubblewrap (`bwrap`), `syscall.Flock`.

**Decision record (TL, 2026-10-11, option B):** Seatbelt does not nest. When `sdlc` already runs inside a sandbox (Claude Code's Bash sandbox), the fence probe fails. The review then degrades automatically to checkout-only isolation, with a loud warning and `isolation: checkout` on the round and sidecar. Not nested, the full fence applies.

---

## Reviewer operations → primitive

| Reviewer operation | Needs |
|---|---|
| `git diff <base> <head>` recipes in the manifest, `git log`, blame | history (objects + commits) |
| revert-the-fix / mutation, then run tests | writable working tree at Head, parent commits |
| read issue + plan text | the live files (may be uncommitted at close) |
| none | source refs, config, index, hooks, remote |

A `git archive` snapshot lacks history. A linked worktree shares refs and config with the source. **`git clone --shared --no-checkout` + `remote remove origin` + `checkout --detach <head>`** keeps every object (read through alternates, measured at 70 ms and 26 MB of working files for ariadne) and makes refs, config, index and hooks private. With `origin` removed, a `git push` has no path back to the source.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Fence` | `cmd/sdlc/internal/reviewbox/fence.go` | new |
| `Isolation` (+ `Wrap`, `Level`) | `cmd/sdlc/internal/reviewbox/fence.go` | new |
| `seatbeltProfile` | `cmd/sdlc/internal/reviewbox/fence.go` | new |
| `DispatchOptions.Wrap` / `.Fenced` | `cmd/sdlc/internal/judge/dispatch.go` | modified |
| `BuildArgs` | `cmd/sdlc/internal/judge/dispatch.go` | modified |
| `gatestate.Round.Isolation` | `cmd/sdlc/internal/gatestate/ledger.go` | modified |
| `sidecarMeta.Isolation` | `cmd/sdlc/reviewsidecar.go` | modified |

- **Fence**: `seatbelt | bwrap | none`, the OS write fence that the probe found usable.
- **Isolation**: `{Fence, Dir, Protect []string, Reason string}`.
  - `Wrap(name, args) (name, args)` builds:
    - seatbelt: `sandbox-exec -p <profile> env -C Dir name args…`
    - bwrap: `bwrap --dev-bind / / --ro-bind P P … --chdir Dir name args…`
    - none: `env -C Dir name args…`
  - `Level()` is `fenced` when Fence ≠ none, and `checkout` otherwise.
  - DRY: one argv transform shared by every agent. That makes claude, codex and gemini equivalent by construction.
- **seatbeltProfile**: `(version 1)(allow default)(deny file-write* (subpath "P")…)`, with each path's `"` and `\` escaped. It is a deny list, so agents keep writing their own state and caches.
- **BuildArgs**: with `Fenced`, claude gains `--settings {"sandbox":{"enabled":false}}` and codex gets `-s danger-full-access` in place of `--full-auto`. Gemini starts no sandbox unless given `-s`, so it is unchanged. `Wrap` is applied last, so `FormatCommandLine` (dry-run) prints the isolated command.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `Probe` | `cmd/sdlc/internal/reviewbox/probe.go` | new | `sandbox-exec` / `bwrap` |
| `Create` / `(*Checkout).Close` | `cmd/sdlc/internal/reviewbox/checkout.go` | new | git, filesystem, flock |
| `Sweep` | `cmd/sdlc/internal/reviewbox/checkout.go` | new | filesystem, flock |
| `openReviewIsolation` | `cmd/sdlc/reviewisolation.go` | new | git rev-parse, reviewbox |
| `dispatchBoundaryReview` | `cmd/sdlc/milestoneclose.go` | modified | — |

- **Probe** returns `(Fence, reason)`:
  - darwin: runs `sandbox-exec -p '(version 1)(allow default)' /usr/bin/true`. A failure (nested: `sandbox_apply: Operation not permitted`) gives `none` plus the stderr.
  - linux: runs `bwrap --dev-bind / / /bin/true`.
  - others: `none`.
  - It is a package var (`reviewbox.Probe`), so cmd tests inject `none` or `seatbelt` without exec.
- **Create(root, src, head, overlays)**:
  1. Sweep `root`.
  2. Make `<id>.lock.tmp`, flock it with LOCK_EX, and rename it to `<id>.lock`, so the lock is held before its name is visible.
  3. `git clone --shared --no-checkout src <id>`, then `remote remove origin`, then `checkout --detach head`.
  4. Copy each overlay (a repository-relative path) from `src` to `<id>`.

  It returns `*Checkout{Dir, lock}`. On any failure it removes what it made.
- **Close** runs `RemoveAll(Dir)`, removes the lock file, then unlocks and closes. It is idempotent.
- **Sweep(root)**: for each `*.lock` whose flock is acquirable with LOCK_NB, the owner is dead, so it removes `<id>` and the lock. A directory with no lock file is debris from a sweep that crashed, and is also removed. The kernel releases a flock when its holder dies, so there is no pid-reuse ambiguity. Go opens files O_CLOEXEC, so the reviewer never inherits the lock fd.
- **Root**: `WF_REVIEW_CHECKOUT_ROOT`, else `<os.UserCacheDir>/sdlc/review-checkouts`.
- **openReviewIsolation(stderr, manifest)**:
  1. Resolves `--show-toplevel` and the absolute `--git-common-dir` as `Protect`.
  2. Calls `Create` at `manifest.HeadSHA`, overlaying `manifest.IssueFile` and `manifest.PlanFile` when they are relative and present.
  3. Probes the fence.
  4. With `none`, warns: `write fence unavailable (<reason>); reviewer isolated to a disposable checkout only (isolation: checkout)`.

### Lifecycle (ARCH-FUNERAL / ARCH-STATE)

- **Created by** `dispatchBoundaryReview`, once per round, just before `judge.Dispatch`.
- **Last needed by** the reviewer process. The review text is already in memory (`output`), and the sidecar and ledger never read the checkout.
- **Removed by** a `defer co.Close()` in `dispatchBoundaryReview`. Success, reviewer failure, timeout, cancellation and SIGINT/SIGTERM (`main.go` NotifyContext → ctx) all unwind through it.
- **Crash (SIGKILL of sdlc):** the next review's `Create` sweeps the orphan. The reviewer group of a SIGKILLed sdlc is not reaped by sdlc; it ends with its own turn. That is today's behavior and is noted in the atlas.
- **Bound:** one directory plus one lock file per live review.

States per id: `absent → locked (lock named, held) → checked-out → removed`. A sweep transitions only `locked-but-unheld → removed`.

## Tasks

### Task 1: reviewbox pure argv + profile

**Files:** Create `cmd/sdlc/internal/reviewbox/fence.go`, `fence_test.go`.

- [ ] Failing tests:
  - `Wrap` for each Fence produces the exact argv.
  - The profile escapes `"` in a path.
  - `Level`.
- [ ] Implement `Fence`, `Isolation`, `Wrap`, `Level`, `seatbeltProfile`.
- [ ] `go test ./cmd/sdlc/internal/reviewbox/ -run 'Wrap|Profile|Level'` passes. Commit `#204: reviewbox: pure isolation argv`.

### Task 2: checkout lifecycle + sweep

**Files:** Create `cmd/sdlc/internal/reviewbox/checkout.go`, `checkout_test.go`. Uses a real git repo in `t.TempDir()`.

- [ ] Failing tests:
  - **SourceUntouched:** snapshot the source's tracked file contents, untracked list, `git for-each-ref`, and `.git/config` bytes. In the checkout, edit a tracked file, add an untracked file, commit, `git config x.y z`, `git branch foo`, `git push` (no remote, so it fails). Re-snapshot the source: identical.
  - **PristinePerRound:** round 1 at H1 writes residue and closes. Round 2 at H2 has no residue, HEAD = H2, and a clean `git status`.
  - **CloseRemoves:** the directory and lock are gone, and a second Close is a no-op.
  - **SweepDeadOwner:** a lock plus directory with no holder are removed by `Create`.
  - **SweepLiveOwner:** a held Checkout survives another `Create`'s sweep.
  - **Overlay:** an uncommitted issue edit appears in the checkout.
  - **CreateFailure:** a bad head leaves nothing behind.
- [ ] Implement.
- [ ] Targeted tests pass. Commit.

### Task 3: probe + live fence conformance

**Files:** Create `cmd/sdlc/internal/reviewbox/probe.go`, `probe_test.go`.

- [ ] Live test: probe, and skip when `none` (nested sandbox, or no bwrap). Under the fence, a `sh -c 'echo x > src/f; echo y > checkout/g'` run through `Isolation.Wrap` leaves `src/f` unchanged and writes `checkout/g`. A `git -C src branch evil` also fails.
- [ ] Implement `Probe` (package var).
- [ ] Commit.

### Task 4: judge dispatch hooks

**Files:** Modify `cmd/sdlc/internal/judge/dispatch.go`, `judge_test.go`.

- [ ] Failing tests:
  - Without `Fenced` or `Wrap`, BuildArgs is byte-identical to today (existing tests).
  - For each agent, `Fenced` adds that agent's inner-sandbox-off flag.
  - `Wrap` prefixes the argv for all three agents identically.
  - `FormatCommandLine` shows the wrap.
- [ ] Implement: add `Wrap func(string, []string) (string, []string)` and `Fenced bool`. Apply them in `BuildArgs`, which `dispatchOnce` and `FormatCommandLine` already share.
- [ ] Commit.

### Task 5: boundary wiring + stamp

**Files:**
- Create `cmd/sdlc/reviewisolation.go`.
- Modify `cmd/sdlc/milestoneclose.go`, `cmd/sdlc/boundaryledger.go`, `cmd/sdlc/reviewsidecar.go` and its deferred sidecar write site, and `cmd/sdlc/internal/gatestate/ledger.go`.
- Tests in `cmd/sdlc/reviewisolation_test.go`.

- [ ] Split `boundaryReviewDispatchOptions` into a manifest step and an options step. The options step takes an `*reviewbox.Isolation`: it sets `manifest.RepoRoot = iso.Dir` before rendering, and sets `opts.Wrap` and `opts.Fenced`. Dry-run passes an Isolation with `Dir: "<review-checkout>"` and the probed fence, and creates nothing.
- [ ] `dispatchBoundaryReview` runs `openReviewIsolation`, then `defer Close`. A Create failure maps to `VerdictNotRun` (`review checkout: …`). `rr.Isolation = iso.Level()`.
- [ ] `persistBoundaryRound` stamps `Round.Isolation` on rounds that ran (`omitempty` keeps old ledgers byte-identical). The sidecar renders a `| isolation | fenced/checkout |` row.
- [ ] Failing tests use a fake `judge.Run` plus a fake Probe and real git:
  - The argv has `env -C <dir>`, with `<dir>` under `WF_REVIEW_CHECKOUT_ROOT`.
  - The prompt's `repository root:` is `<dir>`.
  - The fake reviewer writes residue into `<dir>`; afterwards `<dir>` is gone and the source is clean.
  - Cleanup holds on success, on a Run error (reviewer failure), and on ctx cancel (interruption).
  - A Probe returning `none` emits the warning and stamps `checkout`. `seatbelt` stamps `fenced`, and the argv starts with `sandbox-exec`.
  - Two consecutive dispatches use different, fresh directories.
- [ ] Commit.

### Task 6: contract surface

**Files:**
- the atlas boundary-review section (`atlas/workflow/sdlc-binary.md`, or wherever boundary review is mapped);
- `cmd/sdlc/helptext` for close/milestone-close if they describe the reviewer's environment;
- `cmd/sdlc/internal/judge/code-review.md`.

- [ ] Change "revert the fix in a scratch copy" to say that the review runs in a disposable checkout, so mutate and revert there, in the foreground.
- [ ] Atlas: lifecycle, fence, degrade rule, sweep, `WF_REVIEW_CHECKOUT_ROOT`.
- [ ] Update the golden prompt tests. Commit.

### Verification

- [ ] Targeted `go test` per package while iterating; `make test` before close.
- [ ] Manual check: `sdlc close --dry-run` (or `milestone-close --dry-run`) shows the `env -C <review-checkout>` command.
- [ ] The live fence test runs where the probe succeeds (unsandboxed operator or CI). It skips inside an agent sandbox, which the test's skip message says.

## Revisions

### 2026-10-11 — close review round 1 (REWORK)

Reason: the boundary review found 1 Critical, 4 Important and 3 Minor.

Delta:
- **Revert instruction (BR-1):** the in-place revert instruction is rendered
  only for isolated dispatches (`PromptInput.Isolated`). Unisolated prompts
  keep the pre-#204 scratch-copy text byte for byte, and the goldens are
  unchanged.
- **Retries (BR-2):** `DispatchOptions.BeforeRetry` gets `Checkout.Reset`,
  which re-clones in place, so a no-verdict retry starts pristine.
- **Sweep (BR-3):**
  - `Sweep` moved out of `Create` and is best effort: problems are warned
    about, never a refusal.
  - `removeTree` restores write permission.
  - `Close` keeps the lock file when removal fails.
  - An unheld `.lock.tmp` is swept (the Minor).
- **Probe (BR-4):** the probe is self-verifying. It must see a protected write
  refused through the production `Wrap` before reporting a fence. Protect paths
  and the root are canonicalized. The live conformance test still needs an
  unsandboxed run, which this slot cannot do (the classifier refused
  unsandboxed execution).
- **Protect set (BR-5):** it covers every `git worktree list` path. The atlas
  states the remaining HOME widening.
- **Gemini (Minor):** fenced gemini gets `--sandbox=false`. This plan's
  "gemini unchanged" was wrong, because a gemini config can turn its sandbox
  on.
- **Untracked file (Minor):** `TestArchitecture_NarrativeRoutesToArchPrinciples`
  reads the tracked `AGENTS.base.md`. `AGENTS.md` is assembled and gitignored,
  so it is absent in a clean clone.
