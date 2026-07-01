---
issue: 000141
title: sdlc merge confirmation should support agent-safe defaults
created: 2026-07-01
---

# Plan — merge agent-safe defaults: fix the tty detection root cause (#141)

## Root cause (not what the issue assumed)

`sdlc merge` **already** fail-fasts before the pre-merge judges when stdin is
non-interactive — the guard landed in `fcd6b1e` (#56 side-quest, 2026-05-31, a
month before this issue). `runMerge` checks `mergeNeedsTTY(f.Yes, f.DryRun,
isTTY(os.Stdin))` at `merge.go:272`, *before* the step-5 judges at `merge.go:312`.
So pair#84's "ran judges then aborted at the prompt" was a stale downstream binary.

BUT the guard is defeated by a bug in the shared `isTTY` helper
(`changecode.go:522`):

```go
return info.Mode()&os.ModeCharDevice != 0   // ← wrong proxy for "is a terminal"
```

`os.ModeCharDevice` is true for **`/dev/null`** (and `/dev/zero`, …) — all
character devices, none a terminal. An agent almost always runs with stdin from
`/dev/null`, so `isTTY` returns **true**, `mergeNeedsTTY` returns false, the
fail-fast is skipped, the judges run, and the final prompt reads EOF → `Ask`
returns "" → treated as "N" → **abort after the judges**. Exactly the #141
symptom. Reproduced: `sdlc merge </dev/null` sails past the TTY check to the push
check; `[ -t 0 ]` correctly says "not a tty".

## Fix — a real `isatty`, stdlib-only (ARCH: root cause, shared helper)

Replace the char-device proxy with a genuine terminal test (the `TIOCGETA`/`TCGETS`
ioctl — the exact technique `golang.org/x/term` uses internally). Keep it
dependency-free: the module is stdlib + cobra only, and `isTTY`'s own comment
records the deliberate "no `golang.org/x/term`" choice. Confirmed available:
`syscall.{SYS_IOCTL,TIOCGETA,Termios}` (darwin), `syscall.{TCGETS,Termios}` (linux).

`isTTY` (in `changecode.go`) keeps its `*os.File` guard, then delegates:

```go
func isTTY(r io.Reader) bool {
    f, ok := r.(*os.File)
    if !ok { return false }
    return isTerminal(f.Fd())
}
```

Platform split (mirrors x/term's layout; one func body, per-OS constant — ARCH-DRY):
- `tty_unix.go`   `//go:build darwin || linux` — `isTerminal(fd)` = `SYS_IOCTL(fd, ioctlReadTermios, &Termios{})` succeeds (errno 0).
- `tty_darwin.go` `//go:build darwin` — `const ioctlReadTermios = syscall.TIOCGETA`.
- `tty_linux.go`  `//go:build linux`  — `const ioctlReadTermios = syscall.TCGETS`.
- `tty_other.go`  `//go:build !darwin && !linux` — `isTerminal` stub → false (safe default: treat as non-tty → force `--yes`).

This is a **shared-helper** fix: `isTTY` also backs `change-code --worktree=ask`
(the `ASK_BRANCHING_STRATEGY` protocol), so a `/dev/null`-stdin agent there now
correctly gets the sentinel instead of a false "interactive" — the same bug, one
fix (ARCH-DRY / ARCH-PURPOSE: serve the real purpose, everywhere it manifests).

Scope note: the issue's "stdin/stdout" wording — prompts write the question to
stderr and read the answer from **stdin**, so stdin is the channel that decides
answerability. Keep the check stdin-only; a piped stdout with a tty stdin is
still interactive. Documented, not an omission.

## Pure/IO split (ARCH-PURE)

- `mergeNeedsTTY(yes, dryRun, stdinIsTTY)` stays the pure decision (already
  unit-tested) — the bug was never here.
- `isTTY`/`isTerminal` is the thin IO seam over the fd. Testable at its false
  cases deterministically (the true case needs a real pty; not asserted).

## Plan

- [ ] Add `tty_unix.go` / `tty_darwin.go` / `tty_linux.go` / `tty_other.go` with
      `isTerminal(fd uintptr) bool` (real ioctl) + per-OS `ioctlReadTermios`.
- [ ] Rewrite `isTTY` to delegate to `isTerminal`; drop the `os.ModeCharDevice`
      proxy + its stale comment.
- [ ] Test `isTTY`: `/dev/null` → **false** (the #141 regression), regular file →
      false, `os.Pipe` reader → false, non-`*os.File` → false (keep existing).
- [ ] Round out `TestMergeNeedsTTY` with the interactive case (tty, no `--yes` →
      proceeds) so all four Done-when combinations are pinned.
- [ ] Update `sdlc merge --help` / the flag text to say agents should pass
      `--yes` (state when the fail-fast fires).
- [ ] Live-verify: `sdlc merge </dev/null` now fail-fasts at the TTY check
      *before* the push/judge steps with the actionable `--yes` message.
- [ ] `GOOS=linux go build ./cmd/sdlc` cross-compiles (build tags correct).

## Done when (maps to issue)

- Non-interactive (`/dev/null` stdin) merge without `--yes` fails fast before
  judges — because `isTTY` now reports false for non-terminals.
- Interactive merge still prompts (real tty → `isTTY` true).
- `--yes` / `--dry-run` unchanged. Tests cover all four.

## ARCH notes

- **Root cause** (AGENTS.md): fix the wrong terminal test, not a `/dev/null`
  special-case band-aid.
- **ARCH-DRY:** one shared `isTTY`/`isTerminal`; one func body + per-OS constant.
- **ARCH-PURPOSE:** the fix lands wherever `isTTY` is consumed (merge *and*
  change-code), which is the actual purpose (agent-safe interactive detection).
- **ARCH-PURE:** `mergeNeedsTTY` decision stays pure; IO isolated in `isTerminal`.
