---
reviewer_runs:
    - version: 1
      intent:
        id: 4e705b71-0f1c-44ee-95b5-2bdd304074b4
        repository: github.com/xianxu/ariadne
        issue_id: "000475"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: claude
        root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        reviewed_head: 7231ba65fa6748dabd24a019ff91de42e2bcb4d9
        slot: ariadne:4
        requested_at: "2026-10-20T11:21:48.340883Z"
      phase: finished
      pid: 95209
      started_at: "2026-10-20T11:21:48.551161Z"
      finished_at: "2026-10-20T11:22:11.893008Z"
      session_id: 3fc56c78-9cad-4e12-a1a7-aed05bd0c63d
      outcome: returned
gate: plan-quality
issue: 475
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 4e08658d-0d83-4a29-8048-4bf05c617224
            repository: github.com/xianxu/ariadne
            issue_id: "000475"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: 16d61f048357e848bc272feae5aae5988ab0a379
            slot: ariadne:4
            requested_at: "2026-10-20T11:19:01.03011Z"
          phase: finished
          pid: 55134
          started_at: "2026-10-20T11:19:01.139084Z"
          finished_at: "2026-10-20T11:19:59.592374Z"
          session_id: ac351657-0a41-4804-b564-3e710ea2dd55
          outcome: returned
      "n": 1
      timestamp: "2026-10-20T04:19:59-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: PQ-1
          severity: Important
          title: Ready waiters do NOT reach trainAdmission; AdmissionRule gets only the ticket's own readiness
          detail: AdmissionRule is func(events, t, r *Readiness, granted) (line.go:76), called per ticket in tryAdmit (line.go:266-279). The full/max-wait reasons need the ready count and oldest vouch, so the landingline AdmissionRule seam must change (pass the ready set or a summary, counting only vouches valid under r.For). Name this seam change in the plan.
          family: unbacked-existing-behavior-claim
          round: 1
        - kind: defect
          id: PQ-2
          severity: Important
          title: 'No persisted last-departure time: Event.At is the train''s completion time, and departure lives only inside NextDue'
          detail: 'recordTrain (landingtrain.go:591-598) records after the train runs; departed survives only as NextDue = departed+interval. The slot check Prev(now) > lastDeparture needs a source: a new additive departed field (ARCH-EVOLVE note plus round-trip test), or a derivation that breaks when the interval changes or cars are left ready, or At, which lets a train running across a boundary swallow a slot. Pick one.'
          family: unbacked-existing-behavior-claim
          round: 1
        - kind: defect
          id: PQ-3
          severity: Important
          title: Early (full/max-wait) departures can land inside another repo's planned window, contradicting the never-overlap Done-when
          detail: With D1's fixed grid, an early departure can start at any minute (pair full at :05 lands in ariadne's :00 window). Decide whether early departures respect the machine plan or accept overlap and rely on the suite lock; state it as a rule or non-goal, and narrow T8 and Done-when to match.
          family: done-when-contradicts-design
          round: 1
        - kind: defect
          id: PQ-4
          severity: Minor
          title: Lone landing with leftReady > 0 loses its immediate-next-train signal once alone records drop NextDue
          family: implicit-behavior-change
          round: 1
        - kind: defect
          id: PQ-5
          severity: Minor
          title: Readiness.At is overwritten on each re-vouch, so max-wait is not measured from the first vouch as D3 states
          family: unbacked-existing-behavior-claim
          round: 1
        - kind: defect
          id: PQ-6
          severity: Minor
          title: T1 and T6 enumerate test cases in prose; compress to one strategy line per risky function
          family: test-prose-enumeration
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: 76222420-3009-41f6-a0c4-18510668a0ad
            repository: github.com/xianxu/ariadne
            issue_id: "000475"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: 6a23e85ed321fa8bf8c38d798b85b4b4ffc1c948
            slot: ariadne:4
            requested_at: "2026-10-20T11:20:36.530935Z"
          phase: finished
          pid: 83324
          started_at: "2026-10-20T11:20:36.625822Z"
          finished_at: "2026-10-20T11:21:03.359526Z"
          session_id: f5886db6-bd63-4d0e-aec8-021fea4f0187
          outcome: returned
      "n": 2
      timestamp: "2026-10-20T04:21:03-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: AdmissionRule gains a Waiters arg computed once per tryAdmit pass under head.lock, counting r.For-valid vouches only.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Additive Event.Departed (omitzero) set by recordTrain; old events fall back to At (conservative); round-trip test planned.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: 'Stated rule: the plan schedules slot departures only, early departures may overlap and are bounded by the suite lock; Done-when and T8 narrowed.'
          round: 2
        - id: PQ-4
          disposition: addressed
          note: An alone record that left cars ready carries NextDue = departed, so it can only bring the next train forward; test planned.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: Additive Readiness.Since, kept across re-vouches, reset when the head changes; old sidecars fall back to At.
          round: 2
        - id: PQ-6
          disposition: addressed
          note: T1 and T6 compressed to one strategy line each (boundary table + property, golden + overlap property).
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: 6bd34644-9349-4297-a5cf-c28ec6a7f2c4
            repository: github.com/xianxu/ariadne
            issue_id: "000475"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: 7231ba65fa6748dabd24a019ff91de42e2bcb4d9
            slot: ariadne:4
            requested_at: "2026-10-20T11:21:29.108578Z"
          phase: finished
          pid: 92283
          started_at: "2026-10-20T11:21:29.220961Z"
          finished_at: "2026-10-20T11:21:46.850874Z"
          session_id: 95466ede-cf18-4115-b13f-1faa8f33645d
          outcome: returned
      "n": 3
      timestamp: "2026-10-20T04:21:47-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Revisions PQ-1 defines the Waiters seam and its computation under head.lock.
          round: 3
        - id: PQ-2
          disposition: addressed
          note: Event.Departed added on train events, with a stated fallback to At.
          round: 3
        - id: PQ-3
          disposition: addressed
          note: The plan schedules slot departures only; Done-when and T8 narrowed.
          round: 3
        - id: PQ-4
          disposition: addressed
          note: An alone record carries NextDue=departed only when it left ready cars.
          round: 3
        - id: PQ-5
          disposition: addressed
          note: Readiness.Since is kept across re-vouches.
          round: 3
        - id: PQ-6
          disposition: addressed
          note: T1 and T6 compressed to one strategy line each.
          round: 3
      blocked: false
content_hash: 7c4b45da8a29137d1b4464fc55872bf5e6370b8825cfc0ed284166235dc139c2
---

# Gate ledger — ariadne#475 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T04:19:59-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `unbacked-existing-behavior-claim` (kind: defect) Ready waiters do NOT reach trainAdmission; AdmissionRule gets only the ticket's own readiness
  AdmissionRule is func(events, t, r *Readiness, granted) (line.go:76), called per ticket in tryAdmit (line.go:266-279). The full/max-wait reasons need the ready count and oldest vouch, so the landingline AdmissionRule seam must change (pass the ready set or a summary, counting only vouches valid under r.For). Name this seam change in the plan.
- **PQ-2** [Important] `unbacked-existing-behavior-claim` (kind: defect) No persisted last-departure time: Event.At is the train's completion time, and departure lives only inside NextDue
  recordTrain (landingtrain.go:591-598) records after the train runs; departed survives only as NextDue = departed+interval. The slot check Prev(now) > lastDeparture needs a source: a new additive departed field (ARCH-EVOLVE note plus round-trip test), or a derivation that breaks when the interval changes or cars are left ready, or At, which lets a train running across a boundary swallow a slot. Pick one.
- **PQ-3** [Important] `done-when-contradicts-design` (kind: defect) Early (full/max-wait) departures can land inside another repo's planned window, contradicting the never-overlap Done-when
  With D1's fixed grid, an early departure can start at any minute (pair full at :05 lands in ariadne's :00 window). Decide whether early departures respect the machine plan or accept overlap and rely on the suite lock; state it as a rule or non-goal, and narrow T8 and Done-when to match.
- **PQ-4** [Minor] `implicit-behavior-change` (kind: defect) Lone landing with leftReady > 0 loses its immediate-next-train signal once alone records drop NextDue
- **PQ-5** [Minor] `unbacked-existing-behavior-claim` (kind: defect) Readiness.At is overwritten on each re-vouch, so max-wait is not measured from the first vouch as D3 states
- **PQ-6** [Minor] `test-prose-enumeration` (kind: defect) T1 and T6 enumerate test cases in prose; compress to one strategy line per risky function

## Round 2 — 2026-10-20T04:21:03-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — AdmissionRule gains a Waiters arg computed once per tryAdmit pass under head.lock, counting r.For-valid vouches only.
- PQ-2 — addressed — Additive Event.Departed (omitzero) set by recordTrain; old events fall back to At (conservative); round-trip test planned.
- PQ-3 — addressed — Stated rule: the plan schedules slot departures only, early departures may overlap and are bounded by the suite lock; Done-when and T8 narrowed.
- PQ-4 — addressed — An alone record that left cars ready carries NextDue = departed, so it can only bring the next train forward; test planned.
- PQ-5 — addressed — Additive Readiness.Since, kept across re-vouches, reset when the head changes; old sidecars fall back to At.
- PQ-6 — addressed — T1 and T6 compressed to one strategy line each (boundary table + property, golden + overlap property).

## Round 3 — 2026-10-20T04:21:47-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Revisions PQ-1 defines the Waiters seam and its computation under head.lock.
- PQ-2 — addressed — Event.Departed added on train events, with a stated fallback to At.
- PQ-3 — addressed — The plan schedules slot departures only; Done-when and T8 narrowed.
- PQ-4 — addressed — An alone record carries NextDue=departed only when it left ready cars.
- PQ-5 — addressed — Readiness.Since is kept across re-vouches.
- PQ-6 — addressed — T1 and T6 compressed to one strategy line each.

## Open findings

(none — every finding has been disposed)
