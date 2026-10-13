---
id: 000282
status: working
deps: []
github_issue:
created: 2026-10-01
updated: 2026-10-12
estimate_hours:
card_mirror: '0228d95a8842503ec2a5a0d0d1ed1fe2162960ab' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T18:48:26-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "c5d39ef1", done: "619396ab"}
---

# sdlc move refreshes the destination's details mirror after relocating the owner

## Problem

`sdlc move` (#277) re-stamps the card's `claimant` to the destination slot
after its verified switches. It does not refresh the destination checkout's
details mirror, so the moved issue file still shows the previous owner and
an old `card_mirror` baseline until a later verb (claim, set-status,
change-code, reclaim) refreshes it. Anyone reading the details there sees a
stale owner. That is the confusion #275 set out to remove.

Observed in the #278 smoke test: after `sdlc move :2`, ariadne:2's committed
details still named `ariadne:1`. They changed only when a later
`sdlc reclaim` in :2 refreshed the mirror.

## Spec

After a successful relocation in `relocateAfterMove` (`cmd/sdlc/move.go`),
refresh the destination's details mirror. Use `refreshLocalMirror` against
the destination's tracker environment (`dest`), as every other
post-publication refresh does: a warning on refusal, never a failure. The
move is already complete.

Decide whether the refresh is committed (a narrow commit on the issue branch,
like close's mirror commit #275) or left as a worktree edit (like claim and
the setters). The destination is a working slot on the issue branch, so a
worktree edit is the existing convention. Never refresh on a resting branch.
If relocation does not apply (unattributed or foreign card), leave the
mirror alone.

## Done when

- After `sdlc move :N` relocates the owner, the destination's details show
  the new claimant and current `card_mirror`.
- A refusal (a hand-edited mirrored field) warns without failing the move.
  A non-applicable relocation leaves details untouched.
- A real-git test extends `TestMoveRelocatesItsOwner` to assert the
  destination mirror.

## Plan

Placement rule (B2, ARCH-DRY): every worktree mirror write goes through
`refreshLocalMirror` (claim, reclaim, start-plan, the setters, now move). It
writes only in a checkout on that issue's own branch
(`issue.BranchName(detailPath)`), the branch `start-plan` creates and whose PR
carries the details. Any other checkout (a resting branch, which it already
skipped; another issue's code branch; a detached HEAD) is left alone: its copy
is main's, refreshed through main by `issue publish` and close's archive
commit. Committed writers (close's mirror commit, archive, publish,
move-detail) are already placed on the issue branch or main.

- [x] `refreshLocalMirror` refreshes only on the issue's own branch. Tests:
      a setter on another issue's code branch leaves those details untouched
      (B2 d); a claim from another issue's branch leaves the tree clean, so
      start-plan does not refuse (B2 e); the own-branch refresh still happens.
- [x] `relocateAfterMove` refreshes the destination mirror after a successful
      relocation (or one already done). Extend `TestMoveRelocatesItsOwner` to
      assert the new claimant and current `card_mirror`; a refusal warns.
- [x] B2 c (a `git revert` rolled the mirror back): a reverted mirror is an
      older, untouched projection, so the next refresh on the issue branch
      converges it. Test it, and say "never hand-restore; rerun a verb" in the
      claim help.
- [x] B2 b (set-title vs unpublished details) needs no code: since #308 the
      details are on main from filing, and main's copy refreshes at publish.
      Note it in the Log.

## Log

### 2026-10-12

- Dispatched by the TL (ops:0) with the 10-09 scope event: broadened from
  move's destination mirror to all mirror placement (B2 b–f in
  `workshop/pensive/2026-10-09-01-pensive-sdlc-robustness-evidence.md`).
- One choke point: every worktree mirror write is `refreshLocalMirror`, so the
  placement rule lives there (ARCH-DRY); the committed writers (close's mirror
  commit, archive, publish) were already placed. Rule: refresh only when the
  checkout's branch is `issue.BranchName(detailPath)`.
- B2 b needs no code: since #308 the details are on main from filing, so no
  "unpublished details" exist for set-title to miss; elsewhere the copy is
  main's and converges when the branch lands.
- B2 c: re-running `sdlc claim` (already mine) refreshes, so it is the repair;
  `TestRevertedMirrorConvergesOnTheNextRefresh` pins it. Claim help says so.
- Quota hold on 10-12 at 2cf6f8d3; resumed 10-13, merged origin/main.

### 2026-10-01

- Filed at the operator's request from the #278 smoke test. Kept out of #278,
  which is codecomplete; the gap is #277's.

## Revisions

### 2026-10-12 — scope broadened to all mirror placement

The project's 10-09 scope event folds B2 (b)–(e) into this issue: `set-status`
rewrote another issue's details on a code branch, `claim` dirtied the tree so
`start-plan` refused, a `git revert` rolled back the mirror, and `set-title`
left unpublished details stale. The Plan now fixes placement at the one
worktree writer rather than adding move's refresh alone.
