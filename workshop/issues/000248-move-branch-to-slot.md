---
id: 000248
status: open
deps: []
github_issue:
created: 2026-09-23
updated: 2026-09-23
estimate_hours:
---

# Agent procedure: move this branch to :N

## Problem

Operators say "move this branch to :0" (or `:N`): take the issue branch
checked out in the current slot and check it out in slot `:N` instead, usually
`:0`, because that is the checkout whose built binary and runtime files the
operator's live sessions run from, so a smoke test needs the branch there.
Git allows a branch in only one worktree at a time, so the source slot must
first go back to its resting branch.

No agent procedure names this. `atlas/workflow/workspace-branching.md` covers
"In :2, create an issue branch from :1", refreshing a resting slot, and
independent work, but not moving an existing branch. On 2026-09-23 (pair#316)
an agent did not recognise the phrase until the operator spelled out the
steps, then improvised them: it guessed checkout paths instead of using
`sdlc workspace :N`, and it switched a `:0` holding untracked operator files
without a stated rule.

## Spec

Add a **"Move this branch to :N"** section to
`atlas/workflow/workspace-branching.md`, alongside the other procedures and
reusing its "Before switching or refreshing" preflight (ARCH-DRY). Like the rest
of that doc, it is an agent procedure over existing Git/SDLC commands, not a new
CLI verb.

1. Resolve source (current) and destination with `sdlc workspace --json` /
   `sdlc workspace :N --json`. Same `repo_identity`; destination on its
   `resting_branch`; source on a non-resting issue branch.
2. Preflight both trees as the doc already requires, except for untracked
   files in the destination. `:0` routinely holds operator scratch files, so
   adopt #246's rule: keep untracked files that don't collide with the branch,
   refuse collisions, never stash/reset/auto-commit. State this exception
   explicitly rather than silently relaxing the empty-status rule.
3. Switch the source to its resting branch, then `git switch <branch>` in the
   destination. Report destination-resting commits that are not on the branch
   (e.g. `:0`'s `main` ahead of origin): they stay on the resting branch and
   are absent from what gets tested.
4. Repo-specific post-move step: a repo whose runtime is a built artifact (pair:
   `make build` in `:0`) declares it in its `AGENTS.local.md`; the procedure
   says "run the destination repo's declared post-move build, if any" and
   verifies the branch head is what was built.
5. Tell the operator that already-running sessions keep the old binary; a fresh
   thread/relaunch picks up the build.
6. Reverse ("move it back" / after merge): destination returns to its resting
   branch with the same preflight; `sdlc merge`'s durable-slot landing (#246)
   owns the post-merge return where it applies.

Discoverability: the base-layer AGENTS.md "Peer/slot" guidance names the
phrase ("move this branch to :N") and links the section, so a fresh session
maps the operator's words to the procedure.

## Done when

- `workspace-branching.md` has the "Move this branch to :N" section covering
  steps 1–6, including the untracked-file rule and the local-only-commits
  report.
- The AGENTS.md base layer points the phrase at it; propagated downstream.
- pair's `AGENTS.local.md` declares its post-move build (`make build` in :0).
- Dry run: a fresh agent session told "move this branch to :0" from a pair slot
  performs the procedure without further explanation.

## Plan

- [ ]

## Log

### 2026-09-23

- Filed from pair#316's session, where the move was done by hand: pair-slot1
  → `main-slot1`; `~/workspace/pair` → `000316-couch-altn-binding-latency`
  with untracked `.nvimlog`, `.qoder/` and workshop drafts preserved and its
  4 local-only `main` commits left behind; `make build`; the fix confirmed in
  `bin/pair` via `go tool nm`.
- Related: #244 (slots v2 workflows), #246 (durable-slot landing: return-to-rest,
  untracked-collision rule to reuse), #247.
