---
id: 000109
status: working
deps: []
github_issue:
target: base-layer-mechanics
created: 2026-06-16
updated: 2026-06-16
estimate_hours: 1
---

# propagate-base — skip dependents with a dirty working tree (clean-tree precheck)

## Problem

`sdlc propagate-base`'s `commitConsumption` stages the consumption with `git add
-A`, which assumes ANY dirty state in a dependent is the re-weave's own output. It
isn't: pre-existing uncommitted/untracked work in a dependent (e.g. **a concurrent
agent session mid-edit in a sibling repo**) is indistinguishable, so it gets
swept into a mislabeled `<ref>: consume base-layer change` commit.

Hit live on the #107 atlas-prose propagation (2026-06-16): a concurrent Claude
session was editing `parley.nvim`'s `workshop/plans/000128-*-plan.md`. The re-weave
itself produced no tracked diff (the woven constitution is gitignored everywhere),
yet `git add -A` committed that session's in-flight plan work under the consumption
message. Operator caught it; it was undone with `git reset --mixed HEAD~1` and the
other session re-committed its work properly — but it *raced*, and only resolved
cleanly by luck. With agents working peer repos in parallel, this is a live hazard.

This is the deferred-follow-on hardening flagged in #106's Revisions
("branch-first per dependent … the manual loop lacked"); the clean-tree precheck
is the minimal, root-cause slice of it.

## Spec

Before re-weaving each dependent, check its working tree is CLEAN. If it has
pre-existing uncommitted/untracked work, **SKIP** that dependent (don't `make
weave`, don't commit) and report it in the status table as
`SKIPPED: dirty working tree`. A skipped dependent is left untouched and STALE
(it didn't get the new base), so the run exits NON-ZERO with a distinct message
("N skipped — commit/stash + re-run"), separate from a genuine FAILED.

Why clean-BEFORE is the right gate: the woven output is gitignored, so a
previously-propagated clean dependent reads as not-dirty (`git status --porcelain`
excludes ignored paths). If the tree is clean before re-weave, any post-re-weave
porcelain delta IS the re-weave's own output → safe to `git add -A` + commit. If
it's dirty before, skipping never touches the unrelated in-flight work. This
upgrades `commitConsumption`'s `git add -A` from "assume" to a caller-guaranteed
precondition.

## Done when

- propagate-base re-weaves+commits only dependents that were clean before the
  re-weave; dirty dependents are SKIPPED (untouched) and reported, never `git add
  -A`'d.
- the run exits non-zero when ≥1 dependent was skipped (incomplete propagation is
  not silently "success"), with a message distinct from FAILED.
- a previously-propagated CLEAN dependent (whose only on-disk delta is gitignored
  woven output) is NOT falsely skipped.
- unit test covers: clean → proceed; gitignored woven output → still clean;
  untracked WIP / modified-tracked → dirty.

## Plan

- [ ] Add `workingTreeDirty(repoRoot) (bool, error)` (`git status --porcelain`
      non-empty ⇒ dirty; gitignored output excluded by default). Call it FIRST in
      the per-dependent loop: dirty ⇒ `SKIPPED: dirty working tree …`, skip make
      weave/verify/commit. Track a `skipped` counter; exit non-zero on skipped
      (distinct msg) after the FAILED check. Tighten `commitConsumption`'s doc to
      state the clean-before precondition (so `git add -A` only stages re-weave
      output). Update package doc + the cmd Long help. Unit-test `workingTreeDirty`
      against a temp git repo (clean / gitignored-output-clean / untracked-WIP /
      modified-tracked) reusing the package `git(t,…)` helper. Single-pass atomic
      (one review boundary).

## Log

### 2026-06-16
- Filed from the live #107-propagation incident (operator-flagged): propagate-base
  swept a concurrent session's uncommitted work in parley.nvim via `git add -A`.
  Root cause = no clean-tree precheck. The #106 MVP Revisions had deferred this.
