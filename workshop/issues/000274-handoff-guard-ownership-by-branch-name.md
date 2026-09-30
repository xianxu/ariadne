---
id: 000274
status: open
deps: []
github_issue:
created: 2026-09-29
updated: 2026-09-29
estimate_hours:
card_mirror: '5000f7c3cb9ae2dbfe1561b97f046c95cd2fb489' # card fields mirrored from issue-cards; edit via sdlc
---

# Handoff guard judges detail ownership by branch name

## Problem

`guardTransferredDetails` (`cmd/sdlc/transferguard.go`) exempts a landing's
changes to handed-off issue details only when every commit touching them is on
the owner's issue branch (`refs/heads/<NNNNNN-slug>` or its remote copy,
`ownerAuthored`). Ownership is inferred from a branch name, not from who
authored the change, so two legitimate cases are refused as stale-owner
overwrites:

1. **Unclaimed issue edited on main.** After `move-detail`/publish, an
   unclaimed issue has no owner branch. Any edit to its details on main
   (a spec correction, a log line) is refused at `sdlc push` — observed on
   ariadne#272's own follow-ups, and why #273's Spec cannot be corrected
   today. `issue recovery reconcile` reports nothing to reconcile.
2. **Owner's work authored off its branch.** In the parley.nvim#300–#303
   stack (ariadne#272 Log), #300's re-review and merge resolutions were
   committed on #303's branch; `sdlc merge` refused to land #300's completed
   record over main's initial template. Stacking is now banned (#272), but
   any off-branch authorship by the rightful owner hits the same rule.

## Spec

- Decide ownership from evidence about the change, not the branch name: e.g.
  a commit tagged `#N` that changes only #N's details, or the card's claim
  state (an unclaimed issue has no competing owner to protect).
- Keep the protection the guard exists for: a branch that re-adds, deletes or
  rewrites details a *different* live owner has since changed is still refused.
- The refusal's next action must not recommend restoring main's copy when that
  would discard a completed record.

## Done when

- Editing an unclaimed issue's details on main and publishing succeeds.
- The owner's `#N`-tagged edits land from a branch not named for #N.
- Stale-owner overwrites (details changed on main by their claimed owner, then
  rewritten by another branch) are still refused; tests cover all three.

## Plan

## Log

### 2026-09-29

Split out of ariadne#272 (ban stacked development), where both refusals were
observed.
