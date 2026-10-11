---
id: 000274
status: working
deps: []
github_issue:
created: 2026-09-29
updated: 2026-10-11
estimate_hours:
card_mirror: 'e3f8fbc5ca6e79ba8690b79bdf81da0ba4dd0706' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-11T09:37:30-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "7e9f66b8", done: "10d26823"}
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

Revised against today's code (2026-10-11): #285 already replaced branch names
with the card's claimant, and #307 made the claimant the home slot. Case 2 of
the Problem (the owner's edits from a branch not named for #N) is therefore
accepted already; what remains is that an **unclaimed** card is judged "not
the owner" and refused, and that the refusal recommends `claim`, which cannot
take an issue another workspace holds.

Design: one policy, read by both writers of published details (ARCH-DRY):
**a foreign claimant blocks; mine or unclaimed may edit**, always subject to
optimistic concurrency (based on main's latest version of the file).

- [ ] `detailsFacts.Owner` → `Foreign` (another workspace holds the card);
      `verdictNotOwner` → `verdictForeign`. Unclaimed + based accepts;
      unclaimed + behind is `verdictBehind` (merge main, resolve, rerun).
      `TestDetailsVerdict` table rewritten.
- [ ] Foreign refusal names the owner and drops the `claim` advice:
      coordinate with the owner (or operator `sdlc reclaim`), or drop this
      edit with `sdlc issue restore`. Restore is only offered for another
      workspace's issue, never for this workspace's own record.
- [ ] `requireOwnedToPublish` accepts an unclaimed card (refuses only a
      foreign claimant); the publish's own main-moved check is the
      concurrency guard.
- [ ] Tests: `TestTransferGuardOverBranchShapes` gains unclaimed rows
      (rewrite after merging main accepts; stale re-add is behind) and an
      assertion that the foreign refusal names the owner and not `claim`;
      the owner's edit from another issue's branch (`000013-…`) accepts;
      a publish test for an unclaimed card succeeds.
- [ ] Docs: helptext `issue.md`, `atlas/workflow/issue-tracker.md`,
      `AGENTS.base.md` "Later edits" line.
- [ ] Close.

## Log

### 2026-09-29

Split out of ariadne#272 (ban stacked development), where both refusals were
observed.
