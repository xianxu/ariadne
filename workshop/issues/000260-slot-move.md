---
id: 000260
status: working
deps: []
github_issue:
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: '9dac2893c0e13151acbb5dd2f221c58452e198e1' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-09-28T11:22:06-07:00
---

# sdlc move: move the current branch to another slot

## Problem

Moving the current issue branch to another slot is a common operator request,
usually to `:0` for testing, since `:0`'s runtime and shell point at that
checkout. #248 wrote this down as a manual procedure
([Move this branch to :N](../../atlas/workflow/workspace-branching.md#move-this-branch-to-n)),
which an agent follows step by step. The steps are deterministic, so they
belong in the binary: the agent currently re-derives the preflight, the
switch order and the checks after the switch on every move.

## Spec

`sdlc move [:N]` moves the current slot's branch into slot `:N`. `sdlc move`
with no argument means `sdlc move :0`. Any slot can be the destination.

Steps, run from the source slot:

1. Resolve both slots with `sdlc workspace --json` and `sdlc workspace :N --json`
   (their `worktree_root`, not a path built from the slot number). Refuse if
   `:N`'s worktree is missing, both resolve to the same worktree, or the
   `repo_identity` differs.
2. Check `:N`: on its resting branch, with no tracked or staged changes, dirty
   submodules or Git operation in progress. Untracked files may stay (for
   example operator scratch files in `:0`), as in #248: list them with
   `git ls-files --others --exclude-standard -z`, refuse any that collide with a
   path `A` checks out (including file/directory ancestor collisions), and keep
   the rest untouched.
3. Check the source slot: clean, with no uncommitted or untracked changes.
   Record its current branch `A`. Refuse if `A` is the source's resting branch.
4. Switch the source slot to its resting branch. This must come first because
   Git allows a branch to be checked out in only one worktree.
5. Switch `:N` to branch `A`.

Carry over from the #248 procedure:
- Compare `:N`'s resting branch against `A` and against its configured
  upstream. Report resting commits that the move will leave out of `A`, and
  refuse a nonempty `A..rest` list unless the operator explicitly accepts it.
- Re-resolve both slots just before switching, and refuse if anything changed
  since the preflight.
- Use `switch --no-overwrite-ignore` with `submodule.recurse=false`.
- If step 5 fails, leave `A` intact on its branch ref, report it, and name the
  retry. Never stash, reset, delete or push.
- After the move, verify: `:N` is on `A` at the recorded HEAD, and the source
  is on its unchanged resting HEAD.
- Run the destination repository's post-move build from its `AGENTS.local.md`,
  if it declares one. Say that already-running sessions keep the old binary.
- Support `--dry-run`.

Moving back is the same command run from `:N` with a slot target, e.g.
`sdlc move :1`.

## Done when

- `sdlc move` and `sdlc move :N` perform steps 1–5 with the carried-over
  checks. Real-Git fixture: two linked worktrees; after the move the branch is
  in the destination and the source is on its resting branch.
- Each refusal has a test that also shows nothing changed: missing `:N`,
  `:N` with tracked changes or not resting, an untracked file in `:N` that
  collides with `A`, source dirty or untracked, source on its resting branch,
  `A..rest` nonempty without acceptance.
- Non-colliding untracked files in `:N` survive the move unchanged (test).
- A failed second switch leaves the branch ref and both resting refs intact
  (test).
- The atlas procedure points to `sdlc move` and keeps only what a person still
  has to decide. The README and `sdlc --help` list the verb.

## Plan

- [ ]

## Log

### 2026-09-28

Filed from ariadne slot 1 at the operator's request. The five steps are the operator's; the extra checks come from the #248 procedure.

## Revisions

- 2026-09-28: the operator relaxed step 2 from "`:N` clean" to #248's rule:
  untracked files in `:N` may stay unless they collide with `A`. The open
  question is resolved and removed; Done when gained the collision and
  survival tests.
