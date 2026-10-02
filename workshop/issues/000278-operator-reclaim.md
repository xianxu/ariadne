---
id: 000278
status: working
deps: [ariadne#277]
github_issue:
created: 2026-10-01
updated: 2026-10-01
estimate_hours:
card_mirror: '6bd7daaa803b59bf95092eced08f9f98e1d40935' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-01T17:25:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# Add operator-directed reclaim for recovery

## Problem

After a crash or deliberate relocation, responsibility can need reassignment. Neither unreachable-machine inference nor the publishing-oriented sdlc move operation expresses an operator-authorized transfer.

## Spec

Project: pair/workshop/projects/cross-slot-work-scheduling.md. Captured for operator review; no implementation is authorized by this issue creation.

Add sdlc reclaim as an operator-directed recovery operation assisted by an LLM. Read and show current and proposed operator/machine/slot/worktree assignment; perform a guarded tracker update against the observed revision. Record the operator’s recovery reason and transfer in Git history. Exact CLI shape is design work, not settled here.

Do not infer abandonment from a timeout, shutdown, parking or unreachable machine. Operators coordinate out of band, then issue reclaim on one machine. Do not overload sdlc move or automatically transfer responsibility as a side effect of publishing movement. Reclaim changes assignment; it must not reset worktrees, replay prompts, kill remote workers or pretend to fence off an old machine. Specify how an identical retry and a lost publication response are reconciled from tracker evidence.

## Done when

- An operator can inspect and reclaim a recorded assignment, with the old and new responsibility and reason discoverable afterward.
- Concurrent/stale reclaim attempts cannot silently overwrite a newer assignment; lost-response recovery observes the authoritative result.
- No automated crash/reachability path invokes reclaim, and publishing-oriented sdlc move remains distinct.
- Tests demonstrate wrong-owner resume refusal after reclaim and preservation of dirty work; help explains out-of-band operator coordination.

## Plan

Implementation plan to be designed after issue claim and start-plan; these are requirements, not an approved implementation plan.

## Log

### 2026-10-01

Captured from the performance → messaging guarantees → SDLC ownership/observability → recovery discussion. No implementation started.

### 2026-10-01 (implementation session)

- Operator authorized the work ("work on #278"). Claimed in ariadne:1 with
  :0's pre-#277 binary, deliberately: a claimant card would start the #277
  flag day before the fleet is rebuilt. This card is therefore unattributed
  and will need `--adopt` after the rollout. Ran start-plan; the branch
  `000278-operator-reclaim` sits at main, which includes #277.
- Tension to resolve in design: the Spec says "do not overload sdlc move",
  but #277 (operator decision) already has move relocate the owner's *own*
  work on the same machine, using positive move evidence. Reclaim is the
  general, operator-directed transfer across workspaces and machines; move's
  relocation stays the narrow same-owner case.
- Operator decisions:
  - **CLI shape:** inspect, then confirm. `sdlc reclaim --issue N` shows the
    current owner, the proposed owner and past reclaims, and prints the
    confirm command. `--expect <card-rev> --reason '...'` performs it as a
    CAS against that revision.
  - **New owner:** only the running workspace.
- side-quest: corrected #277's rollout procedure in the atlas and README.
  Each environment builds `sdlc` from its own ariadne checkout, so the
  procedure is: update ariadne:0, then `weave refresh` in every :1+ slot.
  The earlier `weave compile` / `make weave-all` advice would not have
  advanced any Git revision. The operator is running the refresh.

