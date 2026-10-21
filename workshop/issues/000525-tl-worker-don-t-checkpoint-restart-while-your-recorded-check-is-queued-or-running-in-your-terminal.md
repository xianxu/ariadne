---
id: 000525
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '76208d62e21bb639b9df0d08d749cc5910e4ba8f' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-21
