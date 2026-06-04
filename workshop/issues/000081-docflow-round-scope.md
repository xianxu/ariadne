---
id: 000081
status: working
deps: []
github_issue:
created: 2026-06-04
updated: 2026-06-04
estimate_hours: 0.5
---

# docflow round: stage only in-scope started docs, not git add -u

## Problem

`docflow round` (no explicit files) stages with `git add -u` — *every* tracked
file that's modified, not just the doc(s) under review. In a working tree that
holds unrelated tracked WIP (another post being edited, a config tweak, a peer's
change), a review round sweeps those into the journaled round commit, conflating
unrelated changes with the prose review. Surfaced as dogfooding feedback during
#79's first real use (the blog-post review) — harmless there only because the
post was the sole tracked-modified file. Same hazard family as `lessons.md`
"[[git add -A sweeps unrelated untracked WIP]]" — docflow shipped with a latent
instance of the very rule that lesson states.

## Spec

Make the **started doc set** the explicit source of truth for "in-scope," rather
than inferring it from "whatever's modified." Record each started file at `start`
time in git config — `branch.<rb>.docflowFile` (multi-valued) — mirroring the
existing `docflowBase` key (`ARCH-DRY`: reuse the branch-config pattern already
there). `round`'s default staging (no explicit files passed) stages exactly those
recorded paths via `git add -- <files>`, never `git add -u`. `finish` cleans up
the config keys alongside `docflowBase`. Explicit `round -- <files>` is unchanged
(power-user escape). `status`/`finish` scope stays correct for free: since `round`
now only ever commits started files, the existing `git diff --name-only base..HEAD`
scope naturally contains only in-scope docs.

## Done when

- `start` records each started file in `branch.<rb>.docflowFile` (add-path on the
  existing-review-branch path too).
- `round` with no explicit files stages only the recorded in-scope files; an
  unrelated tracked-modified file is **not** swept into the round commit and is
  left untouched in the working tree.
- `finish` unsets `docflowFile` (no config leak after the branch is deleted).
- e2e gains an assertion: with an unrelated tracked file modified, a `round`
  commits only the in-scope doc and leaves the other file dirty.
- All existing docflow.test.sh assertions still pass.

## Plan

Single-pass refinement of #79 (one review boundary) → plain checkboxes.

- [ ] `cmd_start`: `git config --add branch.<rb>.docflowFile "$f"` per started file
  (both the create-branch and add-to-existing-branch paths).
- [ ] add `inscope_files()` helper: `git config --get-all branch.<cur>.docflowFile`.
- [ ] `cmd_round`: default (no explicit files) → `git add -- $(inscope_files)`
  instead of `git add -u`.
- [ ] `cmd_finish`: `git config --unset-all branch.<cur>.docflowFile` in cleanup.
- [ ] `docflow.test.sh`: assert an unrelated tracked-modified file is left unstaged
  across a round (the regression test for this fix).
- [ ] atlas `atlas/workflow/docflow.md`: note the `docflowFile` scope record.

## Log

### 2026-06-04

Filed from #79 dogfooding (blog-post review). The operator hit the gap conceptually
(round could sweep unrelated WIP in a busy repo) and asked to scope `round` to the
started docs. Fix = record started files in branch config, stage exactly those.
