---
id: 000491
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
severity: low
collaboration-mode: collaboration
card_mirror: '8e61bb5c11a3d6fafb290741eba9372571238354' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T06:03:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "087c96bb", done: "f328b890"}
---

# Suite queue position becomes fictitious on allocation contention

## Problem

Observed while #481's recorded suite waited under real allocation contention: reports alternate between the actual position 3 (ahead contains the train and another waiter) and position 1 with the same next-in-line holder. No suite admission occurred. The new nonblocking secondary-lock callback returns its default empty ahead list when allocation cannot be inspected; acquire interprets len(ahead)+1 as an observed position. This produces false position updates and log churn.

## Spec

When allocation/global admission contention prevents observing FIFO order, report position as unknown or retain explicitly stale last-observed position. Never derive position 1 from a default empty ahead list. Keep admission, priority, weight and FIFO unchanged; do not make secondary locks blocking.

## Done when

- A real held-allocation fixture makes the admission attempt return promptly and never reports an observed position 1 from missing order evidence.
- Once allocation is available, the same ticket reports its actual position and retains FIFO admission.
- Stable contention does not generate alternating fake position updates.

## Plan

- [x] Return an explicit unobserved order from paused, global-gate contention, allocation contention and changed-ticket retries; keep the five-result admission seam and all resource decisions unchanged.
- [x] Retain only the latest observed diagnostic text/key in acquire. An unobserved retry uses explicitly last-observed text; with no prior observation it says position unknown. Observation freshness alone does not trigger another log line; periodic and pause-change reporting still apply.
- [x] RED/GREEN real-flock producer controls, then whole-caller diagnostic checks at 8/16 alternating known/unobserved attempts; retain FIFO, priority and cancellation controls.

Actions: wait, cancel.

| Internal evidence | Visible report | Recovery | Possible loss |
|---|---|---|---|
| Order observed | Observed position and classification | Wait | None |
| Order unavailable, prior observation | Explicitly last-observed position/classification when reporting | Wait | Fresh position temporarily unavailable |
| Order unavailable, no prior observation | Position unknown | Wait | Fresh position temporarily unavailable |
| Pause hold | Fresh pause reason plus available/last-observed order | Wait for resume, or cancel | None |

ARCH-ORDER: an empty observed predecessor list differs from no observation. ARCH-DRY: keep the existing scheduler, locks and reporting interval; no alternate queue or durable state. A sixth validity flag duplicates what a nullable order can express; blocking the allocation lock again would reintroduce #481's gate pin. The one cached report dies with acquire. For W wait attempts the added work/storage is O(W)/O(1), with no new provider/process/queue scans. Counted 8/16 caller tests require one admission call per attempt and stable position logging under unchanged observations.

## Log

Filed from #481 final-suite observation, after its close review. Source: scripts/test-lock.py Tickets.admission default result and acquire position rendering. The #481 suite keeps its original waiter; this is diagnostic reporting, not an admission failure.

### 2026-10-20 — claim and quick plan

Claimed after #481 landed and the hub rebuilt central sdlc. Read the dispatch's canonical submitted receipt 6bed2d3c-b536-40ea-89fb-cc6d9ca45052. Source confirms the same empty-list default for every unobserved gated attempt; len(ahead)+1 fabricates first place and changes the log key. Preserve uncertainty with a nullable order and cache only prior observed diagnostics; no admission state/lock change. The real private helper tests exercise the live scheduler seam; no production pause or staging is needed.

### 2026-10-20 — implementation and targeted verification

The real held-allocation, pre-paused, pause-after-preview and changed-ticket fixtures first failed because unobserved order was returned as []. The producer now returns None only when no order was validated; elect and every resource decision are unchanged. acquire caches only observed diagnostic text/key, labels stale reuse, and excludes observation freshness from the change key. Python private-helper tests pass (18 tests); the real Go-supplied provider and pause integration controls pass (34.856s). Counted 8/16 caller tests observe one admission call per attempt and one stable report; unavailable order never probes the legacy head token. Periodic stale reports and actual position changes remain visible.

This is semantic observation propagation across producer/renderer, not a mechanical Go API class: real-lock and whole-caller behavioral controls enforce it instead of a lexical golint rule. No production queue pause or :0 staging was used. The actual global gate still blocks as designed; unobserved gated retries include a pause arriving after preview, not a new nonblocking global lock.

The broader real-lock shell controls pass in full, including exact observed first/second/third-place strings, weighted admission, stopped/dead waiters and legacy queue heads. An explicit mutation mapping None back to [] failed all five caller assertions (8/16 log churn, spurious head-token reads and false position 1); restoring the fix passed. git diff --check passes.

## Revisions

### 2026-10-20 — separate completed implementation from delivery gates

The final Plan checkbox originally ended "Recorded suite, small-diff close review and normal central landing." Those remain mandatory delivery gates, not completed implementation: the recorded suite on bcd6510d is queued, and the close review may overlap it under tl-worker's explicit rule. A recorded pass on the final code precedes landing. Removed only that suffix from the checkbox so its checked state describes the completed regression controls accurately; no design or acceptance change.
