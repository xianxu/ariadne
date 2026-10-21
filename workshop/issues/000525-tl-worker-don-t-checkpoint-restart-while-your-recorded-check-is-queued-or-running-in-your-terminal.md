---
id: 000525
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'beff78c58d3d3854cd5b2701a6ad593d88ba96fb' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T05:46:19-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "8be05a9b", done: "ca43fa8d"}
---

# tl-worker: don't checkpoint/restart while your recorded check is queued or running in your terminal

## Problem

## Spec

A worker that checkpoints and restarts its session (continuation) while a recorded check is queued or running in
its own background terminal loses that check: the terminal dies with the agent, so the lease, in-flight log and queue
place vanish. Observed 10-21 05:15 on pair:9 (pair#752): its retained final-head suite (queue position 6, after
earlier waits of ~40 min) was gone after the continuation restart and had to be requeued from the back.

Add a tl-worker rule (and a one-line hint where the continuation writer is described):
- **Checkpoint before queuing a long check, or after it finishes; never while one you own is queued or running in
  your session's terminal.** If context pressure forces a restart mid-check, say so to the hub first, so it can decide
  (cancel and requeue, or wait).
- After any restart, verify retained checks with `sdlc local-check` (lease/run status) before assuming they survived.

Optional follow-up (separate): run recorded checks detached from the agent's terminal so they survive restarts.

## Done when

- tl-worker's rules state the above; rendered-help phrase guards cover it.

## Plan

- [ ] Add rendered-help phrase guards for safe checkpoint timing, owned queued/running checks, hub-first unavoidable restarts, and post-restart verification; observe failure before changing help.
- [ ] Add the short TESTS rule, a one-line continuation-writer pointer, and README/atlas pointers; verify rendered help and scope guards.

The existing issue Spec settles behavior. The restriction is continuation checkpoint/restart, not ordinary issue/code commits. The hub decides whether to wait or cancel/requeue when context pressure cannot wait; no new check daemon or status verb is added.

Actions: wait for the check; checkpoint before queueing or after completion; tell the hub when restart cannot wait; verify retained check evidence after restart.

| Situation | Visible rule | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Owned recorded check queued/running in the session terminal | No continuation checkpoint/restart | Wait for the check | None |
| Context pressure forces a restart mid-check | Hub hears first and decides wait or cancel/requeue | Tell the hub | Cancellation/restart can lose work and FIFO place |
| Session restarted | Survival is unproven | Verify live run or exact-head completed result | A lost run needs recovery, not assumed success |


## Log

### 2026-10-21

- Claimed and started from fresh main after #511. User dispatch specifies quick-flow help plus phrase guards; issue adds the one-line writer hint and records pair:9/pair#752 losing a queued final suite after continuation restart. #502 is queued after this landing, plan first; not claimed or stacked here.
- ARCH-DRY/PURPOSE: tl-worker remains the policy owner; the continuation prototype points to it rather than duplicating recovery rules. Reuse rules.tsv and existing rendered-help/scope guards. A broader detached-runner mechanism is expressly separate; no runtime code or stored state changes (ARCH-PURE/EVOLVE), no new lifetime or hot-path cost.
- Quick plan posted to ops:0, proceeding unless objected. The continuation file is a prototype being changed as engineering work, not a continuation instance; its writer/approval/no-writer instance workflow is not invoked by this one-line prototype edit.
