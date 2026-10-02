---
id: 000289
status: open
deps: [ariadne#288]
github_issue:
created: 2026-10-02
updated: 2026-10-02
estimate_hours:
card_mirror: '1b338f4501b9e8f4743105cea1146b9d94a6ea3f' # card fields mirrored from issue-cards; edit via sdlc
---

# Slot readiness in sdlc fleet inventory

## Problem

After a restart the operator needs to know, per local slot, whether it can take
new work, holds work to resume, or needs recovery. `sdlc fleet inventory
--json` already gathers most inputs for every worktree under the fleet root
(branch, detached, ahead/behind, `dirty_count` including untracked files, and
the issue its branch prefix names with that issue's status), but it reports
worktrees, not slots. It has no `pair:N` address, does not know that a slot's
product checkout and its private substrate checkouts (e.g.
`worktree/pair-slot1/pair` and `worktree/pair-slot1/ariadne`) form one unit, and
gives raw counts rather than a verdict. Couch (pair#363, pair#367) should
consume one per-slot verdict instead of re-deriving git state.

## Spec

Project: pair/workshop/projects/cross-slot-work-scheduling.md. Captured for
operator review; no implementation is authorized by this issue creation.

Extend the read-only fleet observation with slots: group each workspace's
checkouts (the product worktree plus the substrate dependencies weave declares
for it, not every sibling directory), attach the workspace address (`repo:N`,
`:0`) and resting branch, and compute one readiness verdict per slot over all
its checkouts:

- **ready** — every checkout clean (no modified or untracked files), no active
  Git operation, on its resting branch or on a branch already merged to main.
- **holds-work** — clean, but a checkout is on a branch with unlanded commits
  or one whose name points at a non-terminal issue (even with zero commits).
  Claim-awareness (a slot named on a claim while still on its resting branch)
  comes from ariadne#288's observation.
- **needs-recovery** — modified or untracked files, an active merge/rebase, or a
  detached HEAD; name each checkout and the paths behind the verdict.
- **missing** / **unknown** — a declared checkout is absent, or a probe failed.
  Unknown is never ready.

Read-only, versioned JSON, registered in the #280 recovery catalog. The
verdict is an observation: an action that reuses a slot (pair#363's reboot,
then `weave refresh`) must re-check it at action time.

## Done when

- Fleet inventory reports one row per slot with address, member checkouts and
  a verdict; tested with stateful fixtures covering ready, holds-work (unlanded
  commits; issue branch with zero commits), needs-recovery (modified,
  untracked, active rebase, detached) in the product checkout and in a substrate
  checkout, missing checkout, and a failed probe.
- needs-recovery rows list the offending paths per checkout.
- Membership comes from weave's declared dependencies, tested against a slot
  with an undeclared sibling directory.
- The command's recovery entry states read-only, with its proof.

## Plan

Implementation plan to be designed after issue claim and start-plan.

## Log

### 2026-10-02

Filed from pair#384's design review at the operator's direction (pair#384 closed
wontfix; its slot facts belong to sdlc). To be picked up after ariadne#288.
Consumers: pair#363 (bulk resume/reboot), pair#367 (slot view and recovery).
