---
id: 000278
status: open
deps: [ariadne#277]
github_issue:
created: 2026-10-01
updated: 2026-10-01
estimate_hours:
card_mirror: '9da00517ec8098e778d6cbfb4eecfa6d5bad9375' # card fields mirrored from issue-cards; edit via sdlc
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
