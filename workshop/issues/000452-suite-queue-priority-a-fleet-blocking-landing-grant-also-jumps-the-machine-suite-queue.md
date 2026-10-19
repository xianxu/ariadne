---
id: 000452
status: working
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'd51b2d98a33710489a378179089b3e75b9da21cc' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T18:04:52-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# suite queue priority: a fleet-blocking landing grant also jumps the machine suite queue

## Problem

A fleet-blocking landing grant (sdlc landing grant) orders the merge line but not the machine-wide suite queue (WF_TEST_SUITES, two slots). On 2026-10-18 the urgent pair#672 and pair#676 each waited behind ordinary suites for a machine slot even with the grant; the weekly snapshot shows a median queue wait of 11 to 20 minutes and a 90th percentile over 30 minutes.

## Spec

A suite started for an issue holding an active fleet-blocking grant (in that repo's landing line) is admitted to the machine suite queue ahead of every waiting normal suite. It never preempts or kills a running suite: it takes the next free slot. Ordering among several priority suites is FIFO. The wait message names the priority and what it waits behind. Weight (#346) is still respected. The grant's expiry ends the priority.

## Done when

- With two slots busy and normal suites waiting, a granted issue's suite starts on the next free slot before the earlier waiters (regression with the real lock helper).
- Without a grant, ordering is unchanged FIFO.
- An expired grant gets no priority.

## Plan

Durable design: [implementation plan](../plans/000452-suite-queue-priority-a-fleet-blocking-landing-grant-also-jumps-the-machine-suite-queue-plan.md).

- [ ] Pin priority, FIFO, expiry, weight and lifecycle behavior with the real lock helper.
- [ ] Reuse the existing grant authority and reelect weighted admission under the existing mutex.
- [ ] Verify isolated live rollout, recorded suite and close/landing gates.

## Log

### 2026-10-18

### 2026-10-18 — diagnosis and design checkpoint

Claimed before design; start-plan prepared the issue branch from main. Read-only inspection found the machine queue's first waiter becomes a sticky head holding .queue until it acquires capacity; sorting later tickets alone cannot satisfy this issue. The proposed admission keeps the existing slot/ticket lifetime locks and uses the existing allocation mutex for each election and all-or-none scan (ARCH-PURPOSE). The other key constraint is grant authority: the existing Go ValidGrants owns expiry, revocation and authorized-grantor policy. The plan proposes one read-only Go evidence protocol, memory-only cache and a pure Python stable partition, avoiding a second grant store or copied validator (ARCH-DRY, ARCH-PURE).

Simplest alternatives considered: a startup priority boolean/expiry flag goes stale after revocation or late grants, and Python-only grant validation duplicates the current authority. Moving the entire lock helper into Go is a broader rewrite than this issue needs. Full-flow design is saved; no production code, test code or configuration has changed. TL approval is requested for unavailable-provider degradation and the existing-old-helper drain boundary before change-code.

## Revisions

2026-10-18 — TL approved the durable plan. Missing/old/unreadable provider warns and falls back to normal FIFO for all suites. Old sticky waiters drain naturally. Explicit expiry regression proves a granted waiter loses priority on expiry and an older normal waiter proceeds. Approval receipt: 3b55ef84-675d-4f81-834d-80f67f178da1.
