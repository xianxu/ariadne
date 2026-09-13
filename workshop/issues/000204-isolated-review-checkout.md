---
id: 000204
status: open
deps: []
github_issue:
created: 2026-08-25
updated: 2026-08-25
estimate_hours:
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

- [ ] Brainstorm the reviewer operations that require Git history versus a Head
      snapshot and select the smallest isolation primitive that supports them.
- [ ] Define checkout ownership, cleanup, and crash-recovery semantics.
- [ ] Implement the shared review-checkout lifecycle around boundary dispatch.
- [ ] Add cross-agent isolation and residue regression tests.
- [ ] Update the boundary-review atlas/help contract.

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
