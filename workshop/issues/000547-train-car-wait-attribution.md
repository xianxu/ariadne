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

Durable plan: [train car wait attribution](../plans/000547-train-car-wait-attribution-plan.md). Full flow, one close boundary, TL approved on 2026-10-22; no implementation started. The approval and current next action are recorded below.

**Current contract:** The visibility revision below supersedes the original subtraction Done-when and checklist. Report retained train waits as counted unverified coverage; implement the revised durable-plan checklist only.

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

### 2026-10-22T05:18-07:00 — TL approval and same-source policy

Authenticated TL receipt 898cfb60-b77a-42a4-81f4-11b13628ed6d approves bounded start/bind/end facts under the held ticket in the existing 30-day audit, a typed source-bound car-idle method, unchanged old snapshots, no per-poll history and no automatic revision of past actuals. Substantive same-source work during a retained wait stays counted and is labelled **unverified coverage**. Only positively proven idle waiting is excluded; command-text heuristics and guessed credit are forbidden. This resolves the policy question in the initial design; ambiguity about substantive work is counted coverage, distinct from malformed or conflicting provenance.

#548 is now done and landed at 80602233df36a670023b60af8dabb0b6f3c2c071, verified by the tracker and Git ancestry in origin/main. The dependency hold is cleared. Next: integrate landed main, establish the precise structural idle-proof seam without command parsing, then run change-code's full-flow plan-quality gate before estimating or implementing. No tests, production changes or own recorded checks have started. If the structural seam requires a new mechanism beyond the approved facts, return that design question to the TL before code.

### 2026-10-22 — integrated dependency and idle-proof boundary

Integrated landed origin/main (including #548) with no conflicts; no production or test changes. Continued in place under authenticated TL receipt abf57c58-6cc9-4eeb-83fa-e18e58381e1b, which explicitly authorizes proceeding without a continuation writer in this setup.

Read-only inspection and independent bounded exploration agree: current evidence has no structural bridge from a ticket's PID incarnation/operation to the harness runtime job or invocation. activetime.Event drops invocation identity; native Claude tool IDs and Codex call IDs pair calls/results but do not identify the ticket process. Codex exec wrappers can contain arbitrary mixed work, so even an outer call ID cannot prove a poll-only invocation. Existing queue tests deliberately subtract same-source stage/Edit activity; reusing that policy would violate the newly approved counted-work requirement.

The simplest proposed reuse (subtract the entire source-bound held-ticket interval) therefore fails the same-source substantive-work control. Proven blocked-process duration is not proven source idle. A structural job/invocation binding and a pure-observation distinction would be an additional harness/tool-runtime seam; do not implement it without TL approval. Unsupported/ambiguous overlap must stay counted as unverified coverage. This leaves the poll-only Done-when unfulfilled until a positive idle-proof seam exists, so it cannot be claimed solved merely by lifecycle facts. Bring this gap to TL before change-code; no command/result-text reconstruction or extra producer history has been added.

### 2026-10-22T05:23-07:00 — Revision of Spec, Done when and Plan: visibility

Authenticated TL receipt bd4da1f3-415c-467a-a065-e0d099edd3d2 explicitly narrows #547: preserve the planned held-ticket start/bind/end facts and report their overlap as **unverified coverage: N min in a retained train wait** beside measured hours, with **no subtraction**. Existing hours, calibration arithmetic and saved method meanings remain unchanged. Do not build the cross-repo runtime-binding prerequisite here; file it separately at low priority. Reason: blocked-command evidence cannot establish source idle, while diagnostic coverage is useful and honest.

Revised Done when: complete retained train-wait coverage is issue/head/owner/source-bound, clipped and unioned; baseline and split-prefix phases remain visible even when later prefixes omit the car. Incomplete facts report a named limitation, not guessed duration or a new calibration refusal. Same-source substantive work and status polls remain counted, verified on Claude/Codex timelines; no claim that longer polls stop growing focused time. N/2N proves descriptive coverage growth and bounded producer/reader cost instead. Optional snapshot facts replay without live log dependence and preserve old snapshots/methods. Record pair#778's historical limitation without changing its adopted actual; file the low-priority future binding issue. The revised tasks in the durable plan supersede the original implementation checklist. Proceed through full-flow change-code before code.

### 2026-10-22 — plan-quality refinement

First plan review requested executable interval semantics (PQ-1), optional diagnostic budgets/fallback isolation (PQ-2), and function-named test strategies (PQ-3). The appended governing plan refinement defines coverage as the whole closed retained invocation with at least one verified car binding, including pre-boarding/released gaps; it is not train-ride or idle time. No release/phase journal is needed. Concrete audit/snapshot caps fall back to incomplete diagnostic coverage while preserving the valid accounting snapshot. The new function strategy table supersedes the old exclusion test inventory. Runtime-binding follow-up filed as pair#800, unclaimed and low severity; no Pair checkout changed.
