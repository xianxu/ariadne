---
id: 000289
status: working
deps: [ariadne#288]
github_issue:
created: 2026-10-02
updated: 2026-10-02
estimate_hours:
card_mirror: 'e0b0f6d300080f1b3e621777ceec65a1f428bccd' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-02T14:54:16-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
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
- needs-recovery rows give the reasons per checkout (dirty count, active
  operation, detached HEAD); paths are not listed (operator, 2026-10-02).
- Membership comes from weave's declared dependencies, tested against a slot
  with an undeclared sibling directory.
- The command's recovery entry states read-only, with its proof.

## Plan

Durable plan: `workshop/plans/000289-slot-readiness-in-sdlc-fleet-inventory-plan.md`.

- [ ] M1 — per-checkout readiness facts and verdict (shared `gitx` operation detector, dirty paths + operation on facts, pure `JudgeCheckout`)
- [ ] M2 — fleet inventory reports one readiness row per slot (declared membership via `layergraph.ParseRows`, `AssembleSlots`, versioned `slots`, real-git fixtures, recovery entry, help, atlas)

## Log

### 2026-10-02

Filed from pair#384's design review at the operator's direction (pair#384 closed
wontfix; its slot facts belong to sdlc). To be picked up after ariadne#288.
Consumers: pair#363 (bulk resume/reboot), pair#367 (slot view and recovery).

Details were left untracked in pair-slot1/ariadne by the filing session;
published to main with `sdlc issue move-detail` from that checkout at the
operator's direction, then claimed. Design findings: membership is declared in
`construct/deps` (`pkg/layergraph.ParseRows`, transitive); weave records no
clone set, so membership is re-derived; dependency clones rest on `main`;
operation-marker detection exists twice with different lists (to be unified).

## Revisions

- 2026-10-02 — operator: drop path listing from needs-recovery (reasons only).
  Dependency clones join inventory as rows (they are not rows today). Paused
  until the inventory slowdown #288 introduced is fixed in its own issue.
