---
id: '000078'
status: done
created: 2026-06-03
updated: 2026-06-03
estimate_hours: 1.0
actual_hours: 1.0
---

# sdlc merge clean-tree guard refuses on unrelated untracked files

## Problem

`sdlc merge`'s clean-tree guard (`worktreeDirty`, `merge.go:112-126`) calls
`git status --porcelain` and refuses the merge if the output is **non-empty** —
which includes **untracked** files (`??` lines), not just modified/staged tracked
changes. So unrelated, pre-existing **untracked** WIP in the working tree —
local-only work belonging to a *different* effort — blocks an otherwise-ready
merge. There is no bypass flag (`merge.go:103-106` exposes only `--yes`,
`--no-judge`, `--dry-run`, `--issues-dir`, `--history-dir`).

Hit live shipping #58: two untracked dirs (`.claude/skills/xx-pair-doctor`,
`construct/local/pair-doctor/`) — a separate in-progress skill, nothing to do
with #58 — made `sdlc merge` refuse with "Uncommitted changes found." The
workaround was to `git stash -u` them around the merge and pop after; worse, the
stash itself needed the sandbox disabled because git couldn't unlink under
`.claude/skills` (a sandbox write-deny path). A multi-step, sandbox-fighting
detour for files that have nothing to do with the merge.

**Why untracked-only is materially safer than tracked-dirty** for this guard's
purpose: the guard (and its #62 M1 re-assert at step 9b, `merge.go:295`) exists
to avoid merging server-side and then **stranding** the operator on the
post-merge `git switch main` / `git pull`. Modified **tracked** files genuinely
risk that (switch/pull refuse to clobber local edits). **Untracked** files
carry across a branch switch untouched, *unless* `main` introduces a tracked
file at the exact same path — in which case `git switch`/`pull` refuse on their
own (self-protecting), and at worst this is the pre-existing strand the resume
path (#62 M3, `actionResume`) already recovers idempotently. So untracked files
are a near-zero strand risk, yet treated identically to the real risk.
