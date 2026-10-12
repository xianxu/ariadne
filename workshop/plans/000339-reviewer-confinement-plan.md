# Reviewer confinement (#339) Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A dispatched reviewer (boundary review, plan-quality, estimate-quality, `sdlc judge`) can't silently write a repository or files outside its own directories, even with no OS write fence; a failure to set up its directories aborts the review; a post-review check fails loudly on an outside write.

**Architecture:** One confinement step owned by `judge.Dispatch`, so every reviewer gets it (ARCH-DRY: the four callers share the one dispatch). Before the first attempt it creates a scratch base (inside the review checkout when there is one, else under the sdlc process's temp dir), verifies it writable, writes a `git` shim, and snapshots what's outside. The reviewer's environment points TMPDIR (and Claude Code's `CLAUDE_CODE_TMPDIR`) into the base, sets `GIT_CEILING_DIRECTORIES` so git can't discover a repository above its root, and puts the shim first on PATH so `git init`/`clone` outside its directories is refused. After the run, the check compares the snapshot; a difference turns the dispatch into an error (the review did not run, so the gate halts).

**Tech Stack:** Go (cmd/sdlc), POSIX sh for the shim.

## Cause (for the Log)

Claude Code's Bash tool sets `TMPDIR` from `CLAUDE_CODE_TMPDIR`, which the reviewer inherited from the worker's own agent session (`/tmp/claude-501`). Inside the nested sandbox, `mktemp -d` failed there; the reviewer's script ran `cd "$(mktemp -d)"`-style setup, and a `cd` with an empty argument lands in `$HOME`. Its next steps (`git init`, writing `scripts/merge-checks.d/50-é.sh`, `git add -A`, `git commit -m a`) then ran in home. Nothing in the dispatch gave the reviewer a temp dir of its own, nothing stopped git from creating a repository in home, and nothing noticed afterward. The prompt and the checkout were sound; the environment was inherited unchecked.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `OutsideSnapshot` | `cmd/sdlc/internal/reviewbox/confine.go` | new |
| `gitShim` | `cmd/sdlc/internal/reviewbox/confine.go` | new |

- **OutsideSnapshot** — what the check compares: for each ancestor directory of the reviewer's roots (and `$HOME`), whether `<dir>/.git` exists; and the set of non-hidden top-level entries in `$HOME`. `Diff(before, after) []string` is pure (unit-tested with literal snapshots). Hidden home entries are excluded: tools touch dotfiles routinely (shell history, caches), while a new non-hidden home entry (`~/scripts`) or a new `.git` above the root is the incident's signature. Taking the snapshot is the IO half (`takeSnapshot`).
- **gitShim** — renders the shim script from (real git path, allowed roots). It parses git's leading options (`-C dir`, `-c k=v`, `--opt[=v]`) to find the effective directory and the subcommand; for `init` and `clone` it refuses (exit 128, a message naming #339) when the effective directory or any absolute-path argument is outside the allowed roots. Every other subcommand execs the real git unchanged. Tested by running the rendered script with sh against a fake "real git" (records its argv).

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `Confinement` | `cmd/sdlc/internal/reviewbox/confine.go` | new | filesystem, env |
| `DispatchOptions.Checkout` | `cmd/sdlc/internal/judge/dispatch.go` | modified | — |
| `runProcess` env seam | `cmd/sdlc/internal/judge/dispatch.go` | modified | `exec.Cmd.Env` |

- **Confinement** — `Confine(root string, inside bool, pathPrefix string) (*Confinement, error)`: `root` is the directory the reviewer owns (the checkout, or the operator tree for unisolated judges). With `inside`, the base is `<root>/.sdlc-review` (added to the checkout's `.git/info/exclude`); otherwise `os.MkdirTemp(os.TempDir(), "sdlc-review-")`. `prepare()` (idempotent) makes `base/tmp`, writes and removes a probe file there, and writes `base/bin/git`; any failure returns an error and Confine fails. `Env()` returns `TMPDIR`, `TMP`, `TEMP`, `CLAUDE_CODE_TMPDIR` = `base/tmp`; `GIT_CEILING_DIRECTORIES` = `dirname(root):base`; `PATH` = `base/bin:<pathPrefix>`. `Check()` retakes the snapshot and returns an error listing each difference. `Close()` removes the base (not the checkout).
  - Lifecycle (ARCH-FUNERAL): Dispatch creates it, Dispatch's defer removes it. Inside a checkout it also dies with the checkout (Close / Sweep). Outside, a killed sdlc leaves one small dir in the OS temp dir, which the OS temp reaper owns.
  - Retry: `checkout.Reset` deletes the tree including the base, so Dispatch calls `prepare()` again after `BeforeRetry`.
- **DispatchOptions.Checkout** — the review checkout's directory (boundary review sets it to `iso.Dir`); empty means "unisolated: root is the cwd, base under the temp dir".
- **runProcess env seam** — Dispatch passes the confinement's env to `Run` through the context (`withEnv(ctx, env)`), and `runProcess` overlays it onto `binAugmentedEnv(...)` (PATH from the confinement already includes the owner bin). The `Run` signature and the argv stay unchanged, so the in-process fakes and argv assertions across the suite are untouched.

## Tasks

### Task 1: pure pieces (snapshot diff, shim)
- [ ] Failing tests in `reviewbox/confine_test.go`: `Diff` reports a new ancestor `.git` and a new non-hidden home entry, ignores a new dotfile and unchanged state; the shim refuses `git init` with cwd outside roots, `git -C /outside init`, `git init /outside/x`, `git clone src /outside/x`, and passes `git init` / `clone` inside roots and `git status` anywhere through to the fake real git.
- [ ] Implement `OutsideSnapshot`, `Diff`, `gitShim`; tests green.

### Task 2: Confinement
- [ ] Failing tests: `Confine` inside a checkout creates `.sdlc-review/{tmp,bin}`, excludes it from `git status`, and `Env` carries the six vars; `Confine` with an unwritable base (parent dir mode 0500) errors and leaves nothing; `Check` errors after a test creates `$HOME/scripts` (HOME pointed at a temp dir) or `<ancestor>/.git`; `prepare` after the base was removed recreates it.
- [ ] Implement; tests green.

### Task 3: Dispatch wiring + repro
- [ ] Failing test in `judge` (real subprocess through `runProcess`, `HOME` set to a temp home, a fake `claude` script on PATH): the fake runs the incident's script, `cd "$(mktemp -d)" && git init && mkdir -p scripts && git add -A && git commit -m a`.
  - With the base forced unwritable (temp dir parent chmod 0500 via `TMPDIR` for the unisolated case), Dispatch returns an error before launching, and the fake never ran (a marker file is absent).
  - With mktemp itself made to fail inside the reviewer (the fake sets `TMPDIR` to a nonexistent dir, standing in for the inherited bad `CLAUDE_CODE_TMPDIR`), `cd` lands in the temp home, `git init` is refused by the shim, and if the script writes `~/scripts` anyway the post-check makes Dispatch return an error naming it. No `.git` appears in the temp home.
  - A clean fake run returns its verdict text and leaves no base behind.
- [ ] Wire `Confine` into `Dispatch` (before attempt 1, `prepare` after `BeforeRetry`, `Check` after the last attempt, `Close` deferred); `withEnv` + the `runProcess` overlay; boundary review sets `opts.Checkout = iso.Dir` in `boundaryReviewOptionsFor`.
- [ ] Targeted tests: `go test ./cmd/sdlc/internal/reviewbox/ ./cmd/sdlc/internal/judge/` and `go test ./cmd/sdlc/ -run 'BoundaryReview|PlanningReview|Judge|ChangeCode'`.

### Task 4: docs and close
- [ ] Prompt: no change needed for correctness (the environment now carries the facts); the boundary-tail/code-review text says nothing about temp dirs today.
- [ ] Atlas: the review-isolation section of `atlas/workflow/sdlc-binary.md` (and `helptext/close.md`'s isolation paragraph) gain the confinement (TMPDIR in the base, git ceiling + shim, post-review outside-write check, applies to every judge). Grep "read-only"/"isolation: checkout" for stale claims (lesson from #310).
- [ ] Lesson in `workshop/lessons.md` only if not code-enforced (the code now enforces it; likely none).
- [ ] `make test` before close; main's #338 failure is pre-existing (ariadne:1 fixing).

## Limits (stated, not hidden)

- Without an OS fence, a reviewer that calls `/usr/bin/git` directly, or writes files with plain shell commands, is not prevented; the post-check catches a new `.git` above its root or a new non-hidden entry in `$HOME`, which covers this incident's class, not every possible write. Claude Code's Bash may rebuild PATH from the shell snapshot; if so, the shim is bypassed and the ceiling + post-check remain. Task 3's live check (if a real `claude` can run here) records which.

## Revisions

- 2026-10-12, during implementation:
  - **Cause corrected.** macOS `mktemp` without a template ignores TMPDIR and writes under `_CS_DARWIN_USER_TEMP_DIR`, which the inherited agent sandbox denies. The inherited CLAUDE_CODE_TMPDIR was not the cause. Delta: `bin/` also holds a macOS-only `mktemp` shim, which gives template-less calls a template in TMPDIR.
  - **PATH route.** Claude Code's Bash rebuilds PATH from the shell snapshot, measured live. Delta: the base also holds `env.sh`, named by `CLAUDE_ENV_FILE`, which puts `bin/` first on PATH and sources any inherited env file first. The PATH prepend in the process environment stays in place for codex and gemini.
  - **Signature.** `Confine(root, inside)` dropped `pathPrefix`; `runProcess` prepends `BinDir()` with the existing `binAugmentedEnv`.
  - **Test change.** `TestDispatchParentDeadlineCannotBecomeVerdict` deadline went from 20 ms to 500 ms, because confinement setup comes before the reviewer starts.
