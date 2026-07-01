---
id: 000141
status: working
deps: []
github_issue:
created: 2026-06-29
updated: 2026-07-01
estimate_hours: 0.42
started: 2026-07-01T09:49:03-07:00
---

# sdlc merge confirmation should support agent-safe defaults

## Problem

`sdlc merge` asks for final confirmation before irreversible actions:

- server-side GitHub PR merge;
- remote branch deletion;
- switching/pulling the local checkout;
- archiving completed issues to history;
- deleting the local feature branch or worktree.

That confirmation is sensible for humans in an interactive terminal. In an
agent/non-interactive run, however, the prompt defaults to "no". In pair#84,
`sdlc merge` ran the expensive pre-merge judges successfully and then aborted at
the final prompt because no interactive answer could be supplied. The operator
had to rerun with `--yes`.

The problem is not the confirmation itself. The problem is that non-interactive
contexts discover the need for `--yes` only after spending time on slow judges.

## Spec

Make `sdlc merge` agent-safe without weakening the irreversible-action guard for
humans.

Desired behavior:

- In interactive terminals, keep the final confirmation prompt by default.
- In non-interactive contexts, fail fast before pre-merge judges with a clear
  message: rerun `sdlc merge --yes` after confirming irreversible actions.
- `--yes` remains the explicit opt-in for scripted/agent flows.
- Dry runs remain non-mutating and should not require final confirmation.

The confirmation exists to protect irreversible merge/cleanup actions, not to
guard the read-only pre-merge judges. Therefore the prompt/precondition should
be checked early enough that an agent does not wait through all judges only to
abort at the end.

## Done when

- [x] `sdlc merge` detects non-interactive stdin/stdout before running
      pre-merge judges. (Fail-fast at `merge.go:272` was already there; the fix
      is `isTTY` now correctly reporting false for a non-terminal stdin.)
- [x] Non-interactive merge without `--yes` fails fast with an actionable error.
      (Live-verified: `sdlc merge </dev/null` → "needs interactive confirmation
      … Re-run with --yes", before the push/judge steps.)
- [x] Interactive merge still prompts before irreversible actions. (Real tty →
      `isTerminal` true → `mergeNeedsTTY` false → prompt; unit-pinned.)
- [x] `sdlc merge --yes` keeps the current scripted flow. (Live-verified `--yes`
      + `--dry-run` bypass the gate.)
- [x] Tests cover interactive, non-interactive, `--yes`, and `--dry-run`
      combinations. (`TestMergeNeedsTTY` 4 cases + `TestIsTTY_*` incl. /dev/null.)

## Plan

Design detail + root-cause: `workshop/plans/000141-merge-agent-safe-tty-detection-plan.md`.

**Root cause differs from the issue's framing.** The fail-fast-before-judges
guard *already exists* (`merge.go:272`, from #56/`fcd6b1e`, a month before this
issue). The real bug is the shared `isTTY` helper (`changecode.go:522`): it uses
`os.ModeCharDevice` as a proxy for "is a terminal", but `/dev/null` (an agent's
usual stdin) is a char device, so `isTTY` returns true, the guard is skipped, and
merge aborts *after* the judges. Fix = make `isTTY` a real `isatty` (stdlib
ioctl). This also fixes `change-code --worktree=ask` (same shared helper).

- [x] Add `tty_{unix,darwin,linux,other}.go`: `isTerminal(fd)` via the
      `TIOCGETA`/`TCGETS` ioctl (stdlib-only, per-OS constant; ARCH-DRY).
- [x] Rewrite `isTTY` to delegate to `isTerminal`; drop the `ModeCharDevice` proxy.
- [x] Test `isTTY`: `/dev/null` → false (the #141 regression), regular file /
      pipe / non-`*os.File` → false.
- [x] Round out `TestMergeNeedsTTY` with the interactive (tty → proceed) case.
      (Already present — `merge_test.go:271` — so no change needed.)
- [x] Update `sdlc merge --help` flag text: agents should pass `--yes`.
      (Flag text + a NON-INTERACTIVE section in `helptext/merge.md`.)
- [x] Live-verify `sdlc merge </dev/null` fail-fasts before judges; cross-compile
      `GOOS=linux`.

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.* Small, well-scoped bug fix in a shared helper
+ cross-platform build files + tests; design pre-resolved by the plan doc (+15%
buffer), impl at v3.1's 40%-of-v2 unit.

- smaller-go-module: fix `isTTY` → real `isatty`, add the four `tty_*.go`
  platform files, `isTTY`/`mergeNeedsTTY` tests — design 0.1 + impl 0.15.
- milestone-review: the one boundary review auto-dispatched at `sdlc close` —
  impl 0.15.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
design-buffer: 0.15
item: smaller-go-module   design=0.1 impl=0.15
item: milestone-review    design=0.0 impl=0.15
total: 0.42
```

## Log

### 2026-06-29

- Created from pair#84 dogfooding: the merge judges passed, then `sdlc merge`
  aborted at the final confirmation prompt in a non-interactive agent run. A
  rerun with `--yes` succeeded.

### 2026-07-01

- **Root cause re-diagnosed.** The fail-fast-before-judges guard the issue asked
  for *already existed* (`merge.go:272`, from #56/`fcd6b1e`, 2026-05-31 — a month
  before this issue). pair#84 was a stale downstream `sdlc`. The live bug in
  ariadne was the shared `isTTY` helper: it used `os.ModeCharDevice` as a proxy
  for "is a terminal", but `/dev/null` (an agent's usual stdin) is a char device
  → `isTTY` returned true → the guard was skipped → judges ran → the EOF prompt
  aborted. Reproduced: pre-fix `sdlc merge </dev/null` sailed past the TTY gate to
  the push check.
- **Fix (root cause, shared helper).** Replaced the char-device proxy with a real
  `isatty` — the `TIOCGETA` (darwin) / `TCGETS` (linux) ioctl, the technique
  `golang.org/x/term` uses internally — kept **stdlib-only** (the module is
  stdlib+cobra; `isTTY`'s own comment recorded the deliberate no-`x/term` choice).
  `tty_unix.go` (one `isTerminal` body) + `tty_darwin.go`/`tty_linux.go` (per-OS
  constant) + `tty_other.go` (non-unix stub → false, forces `--yes`). Because
  `isTTY` is shared, this also fixes `change-code --worktree=ask` for
  `/dev/null`-stdin agents (ARCH-DRY / ARCH-PURPOSE — one fix, both callers).
- **Scope note:** the issue's "stdin/stdout" — prompts read **stdin**, so stdin
  is the channel that decides answerability; kept stdin-only (documented).
- **Tests:** `TestIsTTY_RealNonTerminalFilesAreNotTTY` (/dev/null → false [the
  regression], regular file, `os.Pipe` → false); `TestMergeNeedsTTY` already had
  all four combos. gofmt/vet clean; `GOOS=linux` cross-compiles (build tags OK).
- **Live E2E** (post-fix, `bin/sdlc` rebuilt from this branch):
  - `sdlc merge </dev/null` → dies at the TTY gate: *"needs interactive
    confirmation, but stdin is not a terminal. Re-run with --yes…"* — before the
    push/judge steps.
  - `sdlc merge --yes </dev/null` and `--dry-run </dev/null` → bypass the gate
    (reach the expected push check).
