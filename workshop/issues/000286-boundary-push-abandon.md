---
id: 000286
status: working
deps: [ariadne#283]
github_issue:
created: 2026-10-02
updated: 2026-10-08
estimate_hours:
card_mirror: '11706acc2d4a13c09c607b40feedae8dc3bc48ea' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-08T14:22:50-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

- [x] Claim, run start-plan, and design against the Spec and the project PRD (`workshop/projects/claimant-ownership.md`); size the flow at change-code.
- [ ] M1 — Boundary pushes at start-plan, milestone-close, close (and reconcile's close completion); lease push in `pr`; merge deletes the remote branch. Plan Tasks 1–3.
- [ ] M2 — `sdlc abandon`, the set-status redirect, and a reopen that restores the branch and un-archives the details. Plan Tasks 4–7.

Durable plan: `workshop/plans/000286-boundary-push-abandon-plan.md`.

## Log

### 2026-10-08

Claimed in ariadne:2. Survey: `pushIssueBranch` (handoff.go) is the lease push to reuse; nothing pushes at start-plan/close today; milestone-close makes no commit (it pushes HEAD as-is, D3); the durable merge (`gh pr merge` without `--delete-branch`) leaves the remote branch; set-status has no wontfix/punt guard; no reopen un-archives details. Operator decision: reopen of an abandoned issue restores the branch AND un-archives the details (merge main, move history/X back), so it lands cleanly; done-reopen's same gap goes to a follow-up. Boundary push failures warn, never fail the verb (D2); the set-status redirect can't be forced (D8).

### 2026-10-02
