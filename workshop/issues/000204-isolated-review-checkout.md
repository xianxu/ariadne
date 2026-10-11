---
id: 000204
status: codecomplete
deps: []
github_issue:
created: 2026-08-25
updated: 2026-10-11
estimate_hours:
card_mirror: '05b9098d083bc82a688a4f8fb477859ebfa14908' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-11T00:56:11-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.99
---

# Run each boundary reviewer in a disposable isolated checkout

## Problem

Boundary reviewers are instructed to be read-only and leave fixes to the main
agent, but the review protocol also asks them to prove fixes by reverting or
mutating them and running tests. The process currently inherits the real
checkout as its working directory. Claude retains Bash under
`bypassPermissions`; Codex and Gemini do not consume Claude's tool allowlist.
The read-only claim is therefore prose, not an isolation boundary.

An audit of pair#146's ten most recent couch reviews found tool use in all ten
and deliberate mutation or generated artifacts in eight. Some reviewers
voluntarily used scratch worktrees or `git archive` copies, while one Codex
review created and then removed a binary in the real checkout. Safety currently
depends on each reviewer independently inventing the same containment practice.

## Spec

`sdlc` should give every boundary-review invocation a fresh disposable checkout
at the already-resolved immutable Head. The reviewer may read, run tests, and
mutation-test inside that checkout, while the source checkout and its Git state
are outside the reviewer's writable boundary.

- One review round gets one new checkout; filesystem state never carries across
  rounds. Prior review state crosses rounds only through the structured gate
  ledger.
- The checkout is removed after the review's final response and findings are
  durably captured. Interrupted processes have an owned, discoverable cleanup
  path rather than becoming anonymous `/tmp` debris.
- Base, Head, issue/plan paths, and prior findings remain explicit review inputs.
  The reviewer can inspect the pinned change incrementally instead of receiving
  unrelated host/repository context.
- Choose the isolation primitive during design: a linked worktree isolates
  files but shares Git refs/config; a plain archived snapshot removes Git
  access; a disposable shared clone retains history while isolating Git
  metadata. The decision must state which reviewer operations are required.
- Filesystem enforcement is part of the feature. A prompt saying "read-only"
  does not satisfy the source-checkout protection invariant.

## Done when

- A mutation test can edit and test the disposable checkout without changing
  the source checkout's tracked files, untracked files, refs, or config.
- Two consecutive rounds start from their respective pristine pinned Heads;
  residue from round one is absent from round two.
- Normal success, reviewer failure, cancellation, and process interruption have
  tested cleanup/recovery behavior.
- Claude, Codex, and Gemini run with equivalent writable-scope guarantees.

## Plan

Durable plan: `workshop/plans/000204-isolated-review-checkout-plan.md`.

- [x] Brainstorm the reviewer operations that require Git history versus a Head
      snapshot and select the smallest isolation primitive that supports them.
- [x] Define checkout ownership, cleanup, and crash-recovery semantics.
- [x] reviewbox pure argv + Seatbelt profile (plan Task 1).
- [x] Checkout lifecycle + flock sweep (Task 2).
- [x] Fence probe + live conformance (Task 3).
- [x] judge `Wrap`/`Fenced` dispatch hooks (Task 4).
- [x] Boundary wiring + `isolation` stamp on round and sidecar (Task 5).
- [x] Update the boundary-review atlas/help/prompt contract (Task 6).

## Log

### 2026-08-25

- Filed from the pair#146 review audit. Deliberately deferred until #201 makes
  persisted review artifacts semantic rather than raw process transcripts.

### 2026-09-12 — field evidence from parley.nvim#237 (four M1 rounds)

The invariant this issue defends broke twice in one boundary, and the
voluntary containment practice left its own debris:

- **Round 1 destroyed operator work.** The claude reviewer ran
  `git checkout <head> -- .` in the source checkout ("where I meant `git
  diff`", in its own words). That reverted the operator's uncommitted edit to
  an unrelated file and restored a file the operator had deleted. The edit
  survived only in a Time Machine local snapshot, which takes root to mount.
- **Round 3 wrote into the source checkout.** Its sandbox refused `mktemp`, so
  `sed` edits meant for a scratch copy landed in two tracked Lua files; the
  reviewer reversed them by hand.
- **Rounds 3 and 4 contained themselves voluntarily** (`git archive` into
  `/private/tmp/claude-501/parley-scratch-237*`), and left at least five
  orphaned plenary nvims with fixture children there. That is the missing
  "owned, discoverable cleanup path" this Spec asks for.
- Workaround on the parley side meanwhile: snapshot the operator's WIP with
  `git stash create` before each dispatch (parley `workshop/lessons.md`).

### 2026-09-12 — a review round that backgrounded its own work (parley.nvim#237 M2)

The M2 boundary's second round ended without a verdict. The claude reviewer
started its revert checks and the full suite as background tasks, then ended
its turn "waiting for those before writing the verdict". A headless review
cannot be woken by their completion, so the round recorded `unknown` and
carried no findings. When the session ended, its background suite was killed
mid-run, which orphaned a plenary child nvim and a `fake_cliproxy` in the
operator's live worktree. The round before had left two stuck nvims whose
scratch worktrees (`~/.cache/parley-review-wt*`) it had already deleted.

For the Spec: the reviewer must run its checks in the foreground, or the
harness must refuse background tasks in a review invocation. And the review's
process tree must be owned and reaped when it ends, with a verdict or without.

### 2026-10-11 — design start (ariadne:2), one fork open with the TL
- 2026-10-11: closed — Round-1 REWORK findings BR-1..BR-5 + 3 minors fixed with regression tests (TestRevertCheckFollowsIsolation, TestDispatchRunsBeforeRetryBetweenAttempts, TestBoundaryReviewRetryStartsFromAPristineCheckout, TestBoundaryReviewRunsDespiteSweepProblem, TestCloseFailureLeavesLockForSweep, TestResetRestoresPristineCheckout, TestProbeRejectsAFenceThatDoesNotDeny, sibling-worktree assertion in TestBoundaryReviewFencedArgvAndStamp). make test: 12 shards / 1074 cmd/sdlc tests + 22 packages green except processgroup TestCancellationKillsDescendants (sandbox blocks /bin/ps; fails identically on origin/main). One earlier full run failed TestCLISignalCancelsOwnedReviewer once under shard load with no captured detail; it passed 4/4 standalone and in the following full run. Live fence conformance cannot run in this sandboxed slot (Seatbelt cannot nest; unsandboxed execution refused); the probe now proves a denial on every dispatch before stamping fenced.; review verdict: SHIP
- 2026-10-11: flow upgraded quick → full — 753 added lines in code files (limit 100); an earlier round of this close already ran the full review

Measured surface: `judge.runProcess` (`cmd/sdlc/internal/judge/dispatch.go`)
launches every reviewer (plan-quality, `sdlc judge`, boundary) with the
caller's cwd; the boundary prompt and its `git -C <RepoRoot>` recipes name
the source checkout. `processgroup` already reaps the reviewer's group on
normal exit and on cancellation.

Settled direction:

- **Primitive: a disposable `git clone --shared --no-checkout` + detached
  checkout of the pinned Head.** The reviewer needs history (the diff recipes
  run `git diff <base> <head>`, and revert-the-fix checks need the parent
  state), so an archived snapshot is too small. A linked worktree shares refs
  and config, so a `git checkout`/`branch -f`/`config` in it reaches the
  source. A shared clone reads objects through alternates (no copy) and keeps
  refs, config and index private. The clone becomes the reviewer's cwd and the
  manifest's `RepoRoot`. The live issue and plan files are overlaid at the
  same relative paths, so the reviewer reads the same text it reads today.
- **Ownership (ARCH-FUNERAL):** `<UserCacheDir>/sdlc/review-checkouts/<id>/`
  plus a sibling `<id>.lock` held under an exclusive flock for the review's
  lifetime. sdlc creates it per round and removes it in a defer after the
  output is captured (success, reviewer failure, cancellation and SIGINT/TERM
  all unwind through that defer). Each dispatch sweeps first: any `<id>`
  whose lock can be acquired has a dead owner (SIGKILL/crash) and is removed.
  No pid-reuse ambiguity, because the kernel drops the flock with its holder.
  Bound: one directory per live review.
- **Enforcement:** macOS uses `sandbox-exec` with `(allow default)` and
  `(deny file-write* (subpath <source root>) (subpath <git common dir>))`
  around the agent process. That is outside the agent, so claude, codex and
  gemini get the same writable scope. Linux uses `bwrap --ro-bind` of the
  same two paths.

Open fork, asked of the TL: **Seatbelt does not nest.** Inside Claude Code's
Bash sandbox, which is where slot agents run `sdlc close`, `sandbox-exec`
fails with `sandbox_apply: Operation not permitted` (measured). Options:

- (A) Fail closed, with an explicit `WF_REVIEW_ISOLATION=checkout`
  downgrade.
- (B) Degrade automatically to checkout-only isolation, with a loud warning
  and `isolation: checkout` stamped on the round and sidecar.

Recommended (B). Couch delivery to ops:0 failed: the sandbox blocks the
socket, and the classifier refused the unsandboxed retry. The question was
relayed to the operator instead.

### 2026-10-11 — TL decision: option B

The TL (ops:0) chose B: when nested, degrade automatically to clone-only
isolation, with a loud warning and `isolation: checkout` recorded on the review
round; apply the full write fence when not nested. Under the fence, each
agent's own sandbox is turned off (claude `--settings` sandbox disabled, codex
`-s danger-full-access`), because Seatbelt cannot nest inside our fence. One
outer fence is what makes the three agents' writable scope equivalent.

### 2026-10-11 — implementation (ariadne:2)

- `cmd/sdlc/internal/reviewbox`: the pure `Isolation.Wrap` / Seatbelt profile;
  `Create`/`Close`/`Sweep`, a shared clone under a flock; `Probe`. The live
  fence test skips inside an agent sandbox. It measured
  `sandbox_apply: Operation not permitted` here, so the fenced path is proven
  by argv assertions plus that test wherever sdlc runs unsandboxed.
- Isolation is an argv transform (`judge.DispatchOptions.Wrap`), not a new Run
  seam. The 61 `judge.Run` fakes keep their signature, and `--dry-run` prints
  the wrapped command. Under the fence, `Fenced` turns each agent's own sandbox
  off.
- Mutation check: a test that pushes to `HEAD:main` passed even without
  `remote remove origin`, because the source refuses a push to its checked-out
  branch. The test now pushes to a new ref and fails without the removal.
- `atlas/process-manual.md` is generated, but regenerating rewrote about 2,000
  unrelated lines (the checked-in copy is stale against main). Only the
  changed review-prompt passage was edited; a full regen belongs to its own
  change.
- Out of scope, as the Spec names boundary reviews: `sdlc judge
  milestone-review` and plan-quality reviews still run in the caller's
  checkout.
- cmd/sdlc `TestMain` pins `WF_REVIEW_CHECKOUT_ROOT` to a temp dir and the
  probe to checkout-only, so no test depends on the host's sandbox.

### 2026-10-11 — close review round 1: REWORK, all findings fixed

BR-1 through BR-5 and the three Minors are fixed, with regression tests (see
the plan's Revisions). BR-4: the fenced path can't be run live from a sandboxed
slot. The probe now proves the fence on every dispatch, through the production
argv, before claiming `fenced`. `TestProbeRejectsAFenceThatDoesNotDeny` pins
that check, and `TestFenceDeniesSourceWrites` remains for an unsandboxed run.

### 2026-10-11 — post-SHIP: the two advisory Minors fixed

- **Sweep race:** a concurrent sweep could reclaim a just-created
  `.lock.tmp` before Create locked it. The rule is that the sweep acts only on
  a name whose owner is provably dead, so it now skips `.lock.tmp` younger
  than `tmpLockGrace` (1m). The test asserts that a fresh tmp survives a sweep.
- **Test-only helper:** `boundaryReviewDispatchOptions` (no isolation) was
  reached only from tests. It became `plannedBoundaryReviewOptions` (the
  dry-run builder, with the probed fence and a placeholder checkout), so the
  prompt tests check what production sends. The small-diff replay harness
  opens a real checkout.

Landed through `sdlc close --fixed-then-ship`.
