---
id: 000286
status: open
deps: [ariadne#283]
github_issue:
created: 2026-10-02
updated: 2026-10-02
estimate_hours:
card_mirror: 'e9fa3d5ea24c81759cf1fdb8f43f909b557ca579' # card fields mirrored from issue-cards; edit via sdlc
---

# Boundary pushes and sdlc abandon

## Problem

Part of project `claimant-ownership`. Started work is only recoverable after a handoff or a forced `reclaim` if it is on origin. Ending an issue as `wontfix`/`punt` today flips only the card and leaves the local branch, the remote branch and the owner behind.

## Spec

- **Boundary pushes:** push the issue branch at every sdlc boundary (`start-plan`, `milestone-close`, `close`, `unclaim`). The owner is the only writer (claim lock + #272 no-stacking), so rebasing then `push --force-with-lease` is safe. `issue sync` stays local.
- **Cleanup:** `merge` deletes the remote issue branch.
- **`sdlc abandon --issue N --as wontfix|punt --reason …`:**
  - Requires that you own the issue (or claim it first); refuses on a dirty tree.
  - Writes the reason to `## Log` and commits it.
  - Pushes the branch tip to `refs/ariadne/abandoned/NNNNNN`, records that ref on the card, and deletes the issue branch locally and on origin.
  - Narrowly publishes the final details to main and archives them; the card goes terminal, and the owner stays as attribution.
  - The work is kept for both `wontfix` and `punt`. Reopening restores the branch from the archive ref.
- `set-status` to `wontfix`/`punt` from an active status redirects to `abandon`.

## Done when

- After each boundary verb, origin's issue branch equals local HEAD; a rebase followed by the next boundary force-pushes with lease.
- `merge` leaves no remote issue branch.
- `abandon` leaves no issue branch locally or on origin, an archive ref holding the tip, a terminal card, and the details archived on main. Reopening a punted issue restores the branch at that tip.

## Plan

- [ ] Claim, run start-plan, and design against the Spec and the project PRD (`workshop/projects/claimant-ownership.md`); size the flow at change-code.

## Log

### 2026-10-02
