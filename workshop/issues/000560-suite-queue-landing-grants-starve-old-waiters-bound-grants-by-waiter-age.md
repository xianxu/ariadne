---
id: 000560
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours: 1.61
collaboration-mode: collaboration
card_mirror: '25d4bc803513a7926e12299e7a4121a1d6b2d8cf' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T15:10:08-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# Suite queue: landing grants starve old waiters; bound grants by waiter age

## Problem

## Spec

Landing grants (`sdlc landing grant`, fleet-blocking priority) jump the machine-wide suite queue with no bound, so a
run of grants plus ariadne's weight-2 suite can starve small suites indefinitely. Evidence (ops-echo:0, 10-22): the two
oldest tickets (88, parley.nvim-slot3 #409, queued 11:58; 90, echo-slot1 echo#4, queued 12:01) waited about 2.5 h while
newer granted tickets went first (pair#773 twice, ariadne#552, ariadne#555), each granted for a good local reason.

Fix: **aging bounds grants.** A granted ticket jumps only waiters younger than an age limit (e.g. 60 minutes); any
waiter older than that is admitted first, in FIFO order, before further granted tickets. `sdlc queue status` shows
each waiter's age and when aging overrode a grant. The grant still beats everything younger, so it keeps its purpose.

## Done when

- With a ticket older than the age limit waiting, a newly granted ticket is admitted after it, not before (test).
- Granted tickets still jump younger waiters (test).
- Queue status names the aging override.

## Plan

- [x] Approve the durable plan: [bounded grant aging](../plans/000560-suite-queue-landing-grants-starve-old-waiters-bound-grants-by-waiter-age-plan.md).
- [ ] Implement and verify pure protected-prefix ordering and immutable ticket age.
- [ ] Add shared read-only queue inspection/status with actual admission and compatibility coverage.
- [ ] Complete conformance, recorded suite, review and landing.

## Log

### 2026-10-22

### 2026-10-22 — claimed, design checkpoint

#555's retained merge runner completed successfully before this item began; workspace returned to main-slot1. Claimed #560 and ran start-plan, now on its independent issue branch from main. Read current suite helper, grant provider, queue status and private helper/provider tests. Current effective_order is unconditional grant-first stable partition, which explains repeated-grant starvation; existing started is naive wall time from before identity lookup and is unsuitable for elapsed queue age. Keep fairness in Python admission, leaving Go grant validity and repository landing GrantHold unchanged (ARCH-DRY). A read-only subagent audited the same seams and recommended a protected prefix to avoid malformed-age inversions. No code changed. TL approved the policy/60-minute threshold; complete plan is now ready for approval, including shared status projection and bounded private conformance.

### 2026-10-22 — full plan approved; rollout on next helper start

TL approved the complete plan (delivery b7226f49-4802-4e84-a68e-dd94c93e60c5): one named setting/projection; text ages, limit and override; frozen JSON unchanged; additive monotonic member with unknown ages protected; no new durable history, hot reload or forced cutover. Running helpers pick up the new behavior at their next process start; the operator's planned full shutdown tonight restarts them. This issue does not initiate that shutdown. Enter full-flow change-code before implementation because the scheduler, bounded observer and status integration exceed the quick-flow code shell.

## Revisions

### 2026-10-22 — approved aging policy and60-minute threshold

TL approved immutable monotonic enqueue age under the allocation lock, FIFO prefix protected through the last at-bound/older ticket, and current grant ordering only in the younger suffix. Renewal never resets age; missing or invalid age protects its prefix. The threshold is60minutes, selected because the machine routinely runs an hour deep;30minutes would suppress nearly all useful priority. Preserve running holders and weighted all-or-none admission. The durable plan proposes one named Python policy setting shared by admission/inspection, with queue text status rendered from that observer; frozen --json remains unchanged. Full-plan approval precedes code.

### 2026-10-22 — plan-quality refinement

First change-code review accepted the design but blocked on PQ-1 (function-level test contract). Consolidated repeated case inventories into one strategy/independent guard per named risky function. Corrected PQ-2: existing holder reads are bounded, existing ticket/slot writers are not; this issue bounds inspection without expanding writer scope. No production code changed; rerun plan-quality.

## Estimate

Produced via brain/data/life/42shots/velocity/estimate-logic-v3.1.md against baseline-v3.1.md. Method A only. Calibration is flagged stale by estimate-source; values are provisional. Existing flock, provider and process-group seams cover this familiar stack; no new library or external API is needed. Map the bounded Python policy extension to smaller-go-module (the method has no separate Python primitive), the inspection protocol to greenfield-go-module, and CLI wiring to another smaller-go-module. Use upper implementation ranges for concurrency/process coverage, scaled once by0.40. Detailed approved design receives the0.2 design discount; familiarity1.0 and15% design buffer.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.1 impl=0.08
item: smaller-go-module design=0.06 impl=0.2
item: greenfield-go-module design=0.2 impl=0.32
item: smaller-go-module design=0.06 impl=0.2
item: atlas-docs design=0.04 impl=0.08
item: milestone-review design=0 impl=0.2
design-buffer: 0.15
total: 1.61
```

Design0.46×1.15=0.529h plus implementation1.08h gives1.609h, rounded1.61h.
