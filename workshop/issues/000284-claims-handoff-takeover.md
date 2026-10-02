---
id: 000284
status: open
deps: [ariadne#283]
github_issue:
created: 2026-10-02
updated: 2026-10-02
estimate_hours:
card_mirror: 'd41c6ff7de14dff8d60d7f8627bb9755bbd53ffd' # card fields mirrored from issue-cards; edit via sdlc
---

# Claims: multi-claim, publish, handoff, takeover

## Problem

Part of project `claimant-ownership` (see its PRD; design rationale in #283's Log). Once #283 makes the claimant the lock, claims must support shaping several issues before implementation and handing started work between slots, machines and operators.

## Spec

- **Atomic multi-claim:** `sdlc claim --issue a,b,c` compare-and-swaps all the cards in one tracker commit. On a race it re-reads, re-checks every card and retries; it is all-or-nothing.
- **Refresh on claim:** claiming refreshes local details from main, so the lock never protects a stale copy.
- **`issue publish --issue a,b,c`:** narrowly republishes owned details edits to main from any branch. `move-detail` stays the first publication. On an issue branch the identical content is committed there too, so the later merge is a no-op on those files; a resting branch fast-forwards.
- **`issue sync` re-scoped:** commits the details of issues you hold a claim on, locally, on any branch (resting branches included). Its `--push` is retired in favour of `issue publish`.
- **Unclaim:** clears the owner; status is unchanged. For an unstarted (`open`) claim, unpublished details edits are published automatically. For started work it is a handoff: refuse on a dirty tree, write a handoff note in `## Log`, push the issue branch, record `{branch, head}` on the card.
- **Takeover claim:** claiming an active, unowned card fetches the branch, checks origin's head equals the recorded head, and checks it out in a clean slot on a resting branch. This replaces `claim --adopt` (pre-#277 cards are active and unowned).
- **`move` vs. `reclaim`:** `move` stays the same-machine local relocation. `reclaim` (with a reason) stays the forced takeover and gets the last pushed state.
- **`sdlc state`:** shows both views, by slot and by operator, with claim age.

## Done when

- Claiming three issues is one tracker commit; a concurrent claim of one of them makes the whole set fail cleanly, with no partial claims.
- A claim, edit, `issue publish`, unclaim cycle on a resting branch lands the details on main, and nothing is left behind.
- Unclaim of a `working` issue pushes the branch and records `{branch, head}`; a claim from another worktree (fixture: a second machine) resumes at that head. A mismatched head refuses with the next action.
- A pre-#277 `codecomplete` card is taken over by a plain claim.
- `sdlc state` shows the slot and operator views with claim age.

## Plan

- [ ]

## Log

### 2026-10-02

From #283's close review: `requireCardOwnership`'s "moved here, finish with `sdlc claim`" hint (`claimant.go:162`) assumes claim finishes relocations, but `claim.go` does so only for `move`'s statuses (active). An open, held card that `sdlc move` relocated (start-plan's card write lost, then move) gets a refusal instead. Decide here whether `move` applies to open claims; then align claim's repair and the hint.

Also from #283: the model's `unclaim` lists every holdable status, `codecomplete` included, following round 2's handoff decision (which supersedes round 1's "not from codecomplete"). Confirm that when building the verb.

