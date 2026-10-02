---
id: 000282
status: open
deps: []
github_issue:
created: 2026-10-01
updated: 2026-10-01
estimate_hours:
card_mirror: '2f8a75b8b8d947d0e77db4076c986aff183f7848' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ] Refresh the destination mirror in `relocateAfterMove` after a
      successful `moveRelocation`; extend `TestMoveRelocatesItsOwner`.

## Log

### 2026-10-01

- Filed at the operator's request from the #278 smoke test. Kept out of #278,
  which is codecomplete; the gap is #277's.
