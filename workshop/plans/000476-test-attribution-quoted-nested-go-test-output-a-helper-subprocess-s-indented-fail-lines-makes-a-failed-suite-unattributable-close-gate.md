---
gate: boundary-review
issue: 476
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 7cda764e-5187-463e-af6c-58cd83575115
            repository: github.com/xianxu/ariadne
            issue_id: "000476"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/7a8d63d3c22ef16d
            reviewed_head: 74500f04dd8acac5aad8ebaf9cc81267f5a2b2b0
            slot: ariadne:3
            requested_at: "2026-10-20T04:10:59.508699Z"
          phase: finished
          pid: 98824
          started_at: "2026-10-20T04:10:59.683475Z"
          finished_at: "2026-10-20T04:13:00.105894Z"
          session_id: 01a15781-64ae-7422-a86f-670669da8a69
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 74500f04dd8acac5aad8ebaf9cc81267f5a2b2b0
        tracker_head: 95496a0a11396eab41e4d4e1625f370a8cbacce3
        branch_point: 0956e5e1c71bd6734f195929fc151e0c752dbc14
        since: "2026-10-19T20:57:12-07:00"
        until: "2026-10-19T21:10:31-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T20:57:12-07:00"
              end: "2026-10-19T21:10:31-07:00"
              began: "2026-10-19T20:57:12-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: ac507a2fe3c322e2b772ec4382a271bcf131ea79e509a588ba1583c8acf08d23
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 59943433
              windows:
                - start: "2026-10-19T20:57:12-07:00"
                  end: "2026-10-19T21:10:31-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 9.608666666666666
      "n": 1
      timestamp: "2026-10-19T21:13:01-07:00"
      dispatched: "2026-10-19T21:10:59-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Shared-name helper quotations still become real failing subtests
          detail: 'cmd/sdlc/suitefailures.go:62–63 admits every indented FAIL whose name starts with outerTest+"/", regardless of its quotation context. Compiling the pinned scanner with an outer TestOuter, real child TestOuter/real, and a twelve-space quoted "--- FAIL: TestOuter/helper" produces both example.com/outer.TestOuter/real and example.com/outer.TestOuter/helper. The invented leaf can block a registered retry; without a real child, it can also replace the outer failure as the leaf judged by the registry. ARCH-PURPOSE: quoted output must remain inert even when names overlap. The family has one admission predicate, affecting plain output, sharded summary context, CRLF output, and historical-log readers through the shared scanner. Preserve structural indentation/ancestor context rather than using name prefix alone, and add regression cases for overlapping helper names in these modes.'
          family: suite-evidence-provenance
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#476 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T21:13:01-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `suite-evidence-provenance` (kind: defect) Shared-name helper quotations still become real failing subtests
  cmd/sdlc/suitefailures.go:62–63 admits every indented FAIL whose name starts with outerTest+"/", regardless of its quotation context. Compiling the pinned scanner with an outer TestOuter, real child TestOuter/real, and a twelve-space quoted "--- FAIL: TestOuter/helper" produces both example.com/outer.TestOuter/real and example.com/outer.TestOuter/helper. The invented leaf can block a registered retry; without a real child, it can also replace the outer failure as the leaf judged by the registry. ARCH-PURPOSE: quoted output must remain inert even when names overlap. The family has one admission predicate, affecting plain output, sharded summary context, CRLF output, and historical-log readers through the shared scanner. Preserve structural indentation/ancestor context rather than using name prefix alone, and add regression cases for overlapping helper names in these modes.

## Open findings

- **BR-1** [Critical] `suite-evidence-provenance` (kind: defect) Shared-name helper quotations still become real failing subtests
