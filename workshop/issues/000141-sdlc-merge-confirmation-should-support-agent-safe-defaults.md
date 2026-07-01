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

- [ ] `sdlc merge` detects non-interactive stdin/stdout before running
      pre-merge judges.
- [ ] Non-interactive merge without `--yes` fails fast with an actionable error.
- [ ] Interactive merge still prompts before irreversible actions.
- [ ] `sdlc merge --yes` keeps the current scripted flow.
- [ ] Tests cover interactive, non-interactive, `--yes`, and `--dry-run`
      combinations.

## Plan

Design detail + root-cause: `workshop/plans/000141-merge-agent-safe-tty-detection-plan.md`.

**Root cause differs from the issue's framing.** The fail-fast-before-judges
guard *already exists* (`merge.go:272`, from #56/`fcd6b1e`, a month before this
issue). The real bug is the shared `isTTY` helper (`changecode.go:522`): it uses
`os.ModeCharDevice` as a proxy for "is a terminal", but `/dev/null` (an agent's
usual stdin) is a char device, so `isTTY` returns true, the guard is skipped, and
merge aborts *after* the judges. Fix = make `isTTY` a real `isatty` (stdlib
ioctl). This also fixes `change-code --worktree=ask` (same shared helper).

- [ ] Add `tty_{unix,darwin,linux,other}.go`: `isTerminal(fd)` via the
      `TIOCGETA`/`TCGETS` ioctl (stdlib-only, per-OS constant; ARCH-DRY).
- [ ] Rewrite `isTTY` to delegate to `isTerminal`; drop the `ModeCharDevice` proxy.
- [ ] Test `isTTY`: `/dev/null` → false (the #141 regression), regular file /
      pipe / non-`*os.File` → false.
- [ ] Round out `TestMergeNeedsTTY` with the interactive (tty → proceed) case.
- [ ] Update `sdlc merge --help` flag text: agents should pass `--yes`.
- [ ] Live-verify `sdlc merge </dev/null` fail-fasts before judges; cross-compile
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
