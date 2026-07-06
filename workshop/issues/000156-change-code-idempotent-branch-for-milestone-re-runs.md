---
id: 000156
status: codecomplete
deps: []
github_issue:
created: 2026-07-01
updated: 2026-07-06
estimate_hours: 0.68
started: 2026-07-06T15:08:58-07:00
actual_hours: 0.59
---

# change-code idempotent branch for milestone re-runs

## Problem

`sdlc change-code --issue N` errors at its branching step when re-run for a
**subsequent milestone** on an issue whose feature branch already exists and is
already checked out:

```
==> branching: in-place (default; --worktree=yes for an isolated worktree)
Error: git checkout -b 000153-sdlc-retro-process-manual: exit status 128
fatal: a branch named '000153-sdlc-retro-process-manual' already exists
```

Multi-milestone issues legitimately re-run `change-code` per milestone to get that
milestone's **plan-quality + estimate-quality gates** on the (re-)designed plan. In
the failing run the gates ran fine (both INFO) — only the final branch-creation step
failed, because it unconditionally does `git checkout -b <branch>` instead of
detecting that the branch already exists / we're already on it.

Encountered on ariadne#153 M2 (re-running `change-code` for the M2 milestone on the
existing `000153-*` branch).

## Spec

`change-code`'s branching (the in-place `git checkout -b` path in
`cmd/sdlc/changecode.go`, and the worktree path in `createWorktreeBranch`) should be
**idempotent**:

- If already on the target branch → info message ("already on branch X"), skip
  creation, proceed (exit 0).
- If the branch exists but isn't checked out → `git checkout <branch>` (switch), don't
  `-b`.
- Only `git checkout -b` when the branch doesn't exist.

The plan-quality + estimate-quality gates should run regardless (they already do).

## Done when

- Re-running `sdlc change-code --issue N` for a 2nd milestone, while already on that
  issue's branch, runs the gates and exits 0 — no "branch already exists" error.
- A test covers the branch-already-exists / already-checked-out path.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module    design=0.1 impl=0.35
item: milestone-review     design=0.0 impl=0.2
design-buffer: 0.30
total: 0.68
```

## Plan

Design (ARCH-PURE — mirror the existing `estimateRefusal` pure-decision seam so the
branch choice is unit-testable without a git fake). The fns live in
`cmd/sdlc/branchcreate.go` (`createInPlaceBranch:178`, `createWorktreeBranch:144`),
both today unconditional (`checkout -b` / `worktree add -b`). Complete worktree
handling (per operator: a half-idempotent worktree path is just the next bug).

- [x] **Single-source the worktree-porcelain parser (ARCH-DRY, from plan-quality
      review).** The `git worktree list --porcelain` grammar is already walked
      twice — `state.go:145 listWorktrees` and `claim.go:377 findMainWorktree`.
      Do NOT add a third. Extract a pure `parseWorktrees(porcelain string)
      []WorktreeState` from `listWorktrees` (which becomes
      `parseWorktrees(gitx.Capture("worktree","list","--porcelain"))`), and refold
      `findMainWorktree` onto it too (it's the same walk, injected-runner IO stays).
      Then `worktreeForBranch(porcelain, target string) (path string, found bool)`
      is a one-line filter over `parseWorktrees`. Table-test `parseWorktrees` +
      `worktreeForBranch` directly.
- [x] **Pure deciders** (no IO). `decideInPlaceBranch(current, target string, exists
      bool)` → {onTarget | switch | create}. `decideWorktreeBranch(branchExists,
      wtFound bool)` → {reuse | addExisting | addNew}. Table-test both.
- [x] **Thin IO shell + wiring tests.** `createInPlaceBranch` probes current
      (`rev-parse --abbrev-ref HEAD`) + existence (`show-ref --verify --quiet
      refs/heads/<name>`), then: onTarget → info "already on branch X", exit 0;
      switch → `git checkout <name>`; create → `git checkout -b <name>` (today's path).
      `createWorktreeBranch` probes `worktree list --porcelain` + existence, then:
      reuse → rewrite `.goto` to the existing worktree path + info, skip
      (git forbids the same branch in two worktrees, so reuse is the only correct
      move); addExisting → `worktree add <wtPath> <name>` (no `-b`); addNew →
      today's `worktree add -b <name> <wtPath> HEAD`. Drive BOTH shells through the
      `captureRunner` (branchname_test.go:38) and assert the EXACT git command per
      state (the original bug is a shell-wiring bug, not a decider bug): exists →
      `checkout <name>` (no `-b`); onTarget → no `checkout` at all; worktree reuse →
      `.goto` rewrite + no `worktree add`.

## Done when (unchanged — restated for the worktree axis)

- Re-run `change-code --issue N` (in-place) while already on the branch → gates run,
  exit 0, no "branch already exists". Covered.
- Re-run `change-code --issue N --worktree=yes` when the worktree already exists →
  reuse it, exit 0. Covered.
- Branch exists but not checked out (in-place or worktree) → switch/add without `-b`.

## Log

### 2026-07-01

Filed from the ariadne#153 M2 flow (see that issue's Log, "change-code idempotency
bug"). Note: `sdlc issue new` could not auto-sync this to `main` (main not checked out
in a worktree while on the 000153 branch) — this file needs to land on main separately,
NOT bundled into #153's PR.

### 2026-07-06 — implemented (both axes, ARCH-PURE + ARCH-DRY)
- 2026-07-06: closed — Real-git repro (TestCreateInPlaceBranch_RealRepo_IdempotentRerun via hermeticRepo + execGitRunner): re-run while on the branch → idempotent Already-on; switch to main → re-run → Switched. Forcing old unconditional checkout -b reproduces the exact exit-128 "branch already exists". Pure deciders + parseWorktrees + worktreeForBranch table-tested; six captureRunner wiring tests assert the exact git command per state. ARCH-DRY: parseWorktrees single-sourced, findMainWorktree refolded. Full go test ./cmd/sdlc green; build/vet/gofmt clean.; review verdict: SHIP

Both `createInPlaceBranch` + `createWorktreeBranch` (`cmd/sdlc/branchcreate.go`) are
now idempotent via a pure-decision seam (mirrors `estimateRefusal`):

- Pure deciders (no IO): `decideInPlaceBranch(current, target, exists)` →
  onTarget/switch/create; `decideWorktreeBranch(branchExists, wtFound)` →
  reuse/addExisting/addNew. Thin shells probe `rev-parse --abbrev-ref HEAD`
  (`currentBranch`), `show-ref --verify --quiet` (`branchExists`), and
  `worktree list --porcelain` (`worktreeForBranch`), then dispatch.
- **ARCH-DRY (plan-quality review catch):** did NOT add a third porcelain parser.
  Extracted pure `parseWorktrees` from `state.go:listWorktrees` and refolded
  `claim.go:findMainWorktree` onto it (via `worktreeForBranch`) — one grammar,
  three consumers. Removed now-unused `bufio`/`bytes` from claim.go.

**Verification (behavior diff vs main):**

- **Real-git repro** — `TestCreateInPlaceBranch_RealRepo_IdempotentRerun`
  (`hermeticRepo` + real `execGitRunner`): create branch → re-run while ON it →
  idempotent "Already on branch" (exit 0); switch to main → re-run → "Switched to
  existing branch". Drives the real `rev-parse`/`show-ref` output, not a fake.
- **Proved it catches the bug** — forcing the old unconditional `checkout -b` makes
  that test fail with the exact production error: `git checkout -b 000156-real:
  exit status 128 / fatal: a branch named '000156-real' already exists`. Restored.
- Pure/wiring tests: `TestDecideInPlaceBranch`, `TestDecideWorktreeBranch`,
  `TestParseWorktrees`, `TestWorktreeForBranch`, and six `captureRunner` shell tests
  asserting the EXACT git command per state (onTarget → no checkout; exists →
  `checkout <name>` no `-b`; create → `checkout -b`; worktree reuse → `.goto`
  rewrite + no `worktree add`; addExisting → `worktree add <path> <name>` no `-b`;
  addNew → `worktree add -b`).
- Full `go test ./cmd/sdlc/` green; existing claim/findMainWorktree tests unchanged
  (refactor preserved behavior); `go build`/`vet`/`gofmt` clean.

Atlas: `atlas/workflow/sdlc-binary.md` `branchcreate.go` entry — idempotent
branching + the `parseWorktrees` single-source consolidation.

Boundary review: **SHIP** (high, no Critical/Important). Closed the one flagged
coverage gap (added a `bare` case to `TestParseWorktrees`). Two other Minors
acknowledged, not actioned: (a) `createWorktreeBranch`'s `porcelain, _ := r.Git(...)`
swallows the probe error — reviewer confirmed it degrades safely (empty porcelain →
addExisting/addNew, and a genuinely-conflicting `worktree add` still errors from
git); (b) a real-git worktree-*reuse* e2e is future hardening — the worktree axis is
covered via captureRunner synthetic porcelain + the pure deciders, and a real second
worktree risks filesystem escape in the test sandbox.
