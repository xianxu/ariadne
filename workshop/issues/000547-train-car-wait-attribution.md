---
id: 000547
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '400b2f9dac75b6d44389bf14eb15d0aa1c3a5104' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T04:59:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
---

# Train car active-time counts conductor-owned queue wait and status polls as focused work

## Problem

pair#778 renewed close on 2026-10-21 adopted 4.41h and identified a 213.9-minute/80% dominant segment from20:18 to23:52. That segment includes the retained car waiter PID51323 from20:37 to23:44 (about187minutes) while train261 was owned by pair:2 and waiting/running its combined suite, main baseline and prefix suite. pair:5 only polled/reported retained state during that interval; branch changes resumed after its waiter exited. The timing report excluded66.57minutes across2 earlier own-suite waits, but did not exclude this conductor-owned waiting. Frequent status polling appears as engagement. Do not treat the full213.9m as pure wait: it also contains work before/after the retained waiter.

## Spec

Investigate active-time-v3 engagement/queue attribution for a train car waiting on another workspace conductor. Use proven car membership, retained landing ticket and conductor-owned suite/baseline/prefix wait evidence to distinguish waiting from focused work. Preserve actual code/review attention and fail conservatively for unproved identity/time intervals; no guessed manual replacement hours. Determine whether existing lease/car evidence already provides the necessary binding.

## Done when

- Reproduce the train-car attribution with a controlled timeline containing real issue work, an owned car waiter, another slot conductor and periodic read-only status polls.
- Proven wait-only intervals do not become focused hours merely because of polls, including main-baseline and split-prefix phases; actual work around the wait remains attributable.
- Ordinary solo suite waits, concurrent unrelated work, absent/unbound car identity and interrupted/ended waiters retain explicit conservative behavior.
- Record the correction or limitation against pair#778 evidence; regression proves waiting N versus2N does not grow focused work from polls alone.

## Plan

- [ ] Prove the train-car poll inflation and add exact producer-owned waiter identity/lifetime evidence at existing landing resources.
- [ ] Bind validated car waits into fresh measurement while preserving #548, solo-suite ownership, concurrent work and saved measurement replay.
- [ ] Record the pair#778 correction or explicit historical limitation, update provenance docs, and verify/review the final implementation.

Durable plan: [train car wait attribution](../plans/000547-train-car-wait-attribution-plan.md). Full flow, one close boundary, awaiting TL approval; no implementation started.

## Log

Filed unclaimed by pair:5 while its corrected pair#778 suite/review run. No active-time implementation or recorded actual was hand-edited. Evidence in pair workshop/evidence/000778/train261-compile-failure.* and the issue Log pins train, attempts and terminal state. Renewed close report on source b9ea92045914dc1017e4e59ab473c7c73c015fe8 printed4.41h, dominant213.9m segment, and only66.57m excluded suite queue. Earlier close before landing wait adopted1.11h. Both are tool measurements, not replacement recommendations.


### 2026-10-22 — claim and bounded design investigation

- #507 cleanup completed before this claim: PR376 merged at4660dda0751d1d2dffeee40c68bd2ff4784ea62e, clean resting slot restored and owned temporary build caches removed. Claimed #547 and entered start-plan from main; #548 remains a pending landing dependency, read-only comparison used its accepted branch eb0e3d8a8396adb9cb7af9dba44987daef66ed3d. No stacked development or production code changes.
- #548 establishes immutable tested-head issue membership before existing own-slot queue selection. Its MainBaseline skip and owner-slot guard correctly remain; relaxing them would not prove a retained car's waiting. Current landing cleanup removes ticket/readiness/outcome identity, and conductor logging records terminal car outcomes but not a durable boarding/lifetime proof.
- Bounded pair#778 audit found enqueue20:37:15.152800-07 to refused23:43:56.723901-07 for ticket262/train261:186m41.571101s envelope, not a trustworthy actual Wait-return interval. Combined attempt7b86027ca92a83fe2d21d23d8c90fc4c, baseline fe30c82b2d661f4d810f192bcfe27b2d and prefix b70b199a850719de2da728cf829fd0d6 prove conductor phases. Existing logs omit waiter PID incarnation/slot/head and cleanup removed sidecars. Preserve original adopted hours; design explicitly allows a historical limitation rather than inventing a correction.
- Recommended mechanism: use the existing held ticket as lifetime authority, append bounded start/binding/end facts to the existing retained audit, and pin source-bound car idle in explicit train-aware measurement snapshots. No per-poll history, new phase journal, global slot relaxation, guessed timestamps or automatic actual edits. ARCH-DRY/ORDER/EVOLVE/FUNERAL shaped the plan. Same-source substantive-work ambiguity is presented for TL policy approval, not guessed from text.

## Revisions

### 2026-10-22 — proposed producer-bound attribution contract

The investigation found that existing train261 suite and audit metadata can establish a train envelope and conductor phases but not the complete retained waiter identity/return boundary. Propose exact producer evidence and conservative incomplete-historical handling, with explicit persisted measurement-method compatibility. The original Spec/Done-when remain the contract; this durable plan is awaiting TL approval before code or estimates.
