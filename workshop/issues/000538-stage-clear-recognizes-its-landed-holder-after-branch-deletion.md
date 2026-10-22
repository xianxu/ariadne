---
id: 000538
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '53855556f8adaa61cb630abf58fb2bf1a604e475' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T02:51:52-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# stage --clear recognizes its landed holder after branch deletion

## Problem

## Spec

After pair#752 landed, landing deleted its issue branch and returned pair:9 to its resting branch. Its subsequent plain `sdlc stage --clear` refused the stage lease because the holder identity is currently derived from the caller branch; the holder had to use `--clear --abandoned 752` even though landing is the normal completion, not abandonment. Reported by ops:0 while dispatching ariadne#537.

Allow the original holder slot to clear its landed stage after normal branch deletion using confirmed durable landing evidence. Preserve refusal for other slots/issues, unlanded or unknown outcomes, replaced leases and explicit abandoned-holder acknowledgement. Do not treat a missing branch alone as landing proof.

## Done when

- A regression stages from a worker issue branch, completes a durable landing and branch cleanup, then clears from that original slot on its rest without --abandoned.
- Another slot, unfinished/unknown holder and replaced lease still refuse without clearing or rebuilding.

## Plan

- [ ] Add a proof-backed original-slot/resting-caller route to clear a landed issue's retained stage, sharing the existing main pin and archive authority.
- [ ] Replay the branch-deleted landing regression plus independent identity/generation/lease/cancellation refusals; preserve ordinary and explicit abandoned clears.
- [ ] Document, verify, close and land. Detailed design: [implementation plan](../plans/000538-stage-clear-recognizes-its-landed-holder-after-branch-deletion-plan.md).

## Log

2026-10-21: Split from #537 per hub permission if not small. Fresh-main restoration is independent of the holder-authentication policy. Existing checkClearLease only sees branch-derived holder strings; safe normal-completion recovery needs durable landed evidence and original-slot binding, rather than a branch-absence or slot-only bypass.


### 2026-10-22

- Claimed and started from fresh main after #536 landed. Hub highlighted four upcoming operator stages and requested claim/plan, not live :0 mutation.
- Exploration confirms runUnstage currently derives holder only from the caller branch. The target already has a lifetime stage lock and records canonical original Slot, Branch and staged Commit, so no new lease format is needed. Reuse #537's main pin and #512's durable archive proof (ARCH-DRY).
- Proof trap found before code: landingArchiveComplete treats an empty issue selection as success. Recovery must explicitly select the held issue's current done completion/generation before using archive confirmation. Missing/reopened/replaced bindings cannot become issue-free success. Durable plan prepared for hub review; no implementation or tests changed.

- 2026-10-22 03:00 PDT — Hub approved the committed full-flow plan and isolated native production-path proof (receipt 6ad82555-6756-417d-9954-ffba37f12d76). Confirmed original canonical resting slot, current done binding/merged PR/staged commit and explicit nonempty archive selection under existing stage lock/main pin, with final caller/lease/generation re-observation. Do not touch the live pair:0 stages. The existing explicit abandoned acknowledgment remains the interim operator path.

## Revisions

- 2026-10-22 02:59 PDT — Proposed full-flow design narrows the new recovery route to the original canonical slot on its own rest with current tracked done completion, matching landed PR and confirmed archive. Existing same-holder/legacy and exact --abandoned paths remain; new recovery grants no staging or move authority. Propose isolated native proof to preserve the four live operator stages; awaiting hub review.
