---
gate: boundary-review
issue: 567
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: cfb2e440-fe80-4229-9c0c-b0c7d952f81d
            repository: github.com/xianxu/ariadne
            issue_id: "000567"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/b787b09c2e352652
            reviewed_head: 2dddb0b59de75742cb9e288d679285c327184c4f
            slot: ariadne:3
            requested_at: "2026-10-23T12:33:01.754033Z"
          phase: finished
          pid: 69737
          started_at: "2026-10-23T12:33:01.855061Z"
          finished_at: "2026-10-23T12:35:47.327745Z"
          session_id: 01a168c0-18b9-7b83-847e-81de3406c869
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 2dddb0b59de75742cb9e288d679285c327184c4f
        tracker_head: 925a3d69543f3ef60259987e9ac626760c9ce405
        branch_point: 28c40fd5dc3485dce3ec4936485238a56743f4ab
        since: "2026-10-23T05:17:26-07:00"
        until: "2026-10-23T05:32:42-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T05:17:26-07:00"
              end: "2026-10-23T05:32:42-07:00"
              began: "2026-10-23T05:17:26-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 7843c8625a6862dbeb8c038eecc4d0b355af9ea2afd9c667a4a68c6f8b2fa9c2
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 126824902
              windows:
                - start: "2026-10-23T05:17:26-07:00"
                  end: "2026-10-23T05:32:42-07:00"
      "n": 1
      timestamp: "2026-10-23T05:35:48-07:00"
      dispatched: "2026-10-23T05:33:01-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Recognize whitespace-separated ps state format fields
          detail: pkg/golint/ps_liveness.go:56–61 splits format text only on commas and does not trim fields. Native /bin/ps accepted "pid stat=", "pid, stat=", and "pid, state=" during review, but the predicate misses all three, allowing the same snapshot oracle through the new guard (ARCH-PURPOSE). The complete affected family in this window is this shared parser across Command and CommandContext, including separate -o arguments, attached -o formats, and --format= formats containing whitespace separators or comma-adjacent whitespace. Normalize supported field separators and add cases for stat/state through both Check and CheckSources.
          family: literal-command-format-parity
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: adc631ff-17f8-4fb3-9fd4-9b134d918ba4
            repository: github.com/xianxu/ariadne
            issue_id: "000567"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/8d3a008bc9780160
            reviewed_head: 839fc18ae652343c0777a8a515ac943476fc32d3
            slot: ariadne:3
            requested_at: "2026-10-23T12:39:33.831684Z"
          phase: finished
          pid: 86717
          started_at: "2026-10-23T12:39:33.932846Z"
          finished_at: "2026-10-23T12:41:40.981433Z"
          session_id: 01a168c6-1449-74d1-b7da-9b057e7ccabf
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 839fc18ae652343c0777a8a515ac943476fc32d3
        tracker_head: 925a3d69543f3ef60259987e9ac626760c9ce405
        branch_point: 28c40fd5dc3485dce3ec4936485238a56743f4ab
        since: "2026-10-23T05:17:26-07:00"
        until: "2026-10-23T05:38:41-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T05:17:26-07:00"
              end: "2026-10-23T05:38:41-07:00"
              began: "2026-10-23T05:17:26-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 4b10c636cb170b09a353b8a479725490b104a7f260b3cfffec26556a6e57905e
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 127139728
              windows:
                - start: "2026-10-23T05:17:26-07:00"
                  end: "2026-10-23T05:38:41-07:00"
        cuts:
            - dispatched: "2026-10-23T05:33:01-07:00"
              finalized: "2026-10-23T05:35:48-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-23T05:41:42-07:00"
      dispatched: "2026-10-23T05:39:33-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: pkg/golint/ps_liveness.go:65 now normalizes comma and whitespace separators in the shared predicate. TestPSLivenessQueriesFormatSeparators covers both constructors, separate and attached formats, and both Check entry points. All 40 cases pass at head and fail with the previous comma-only parser restored through a temporary Go overlay.
          round: 2
      recipe: small-diff-review
      isolation: fenced
      reviewed: 839fc18ae652343c0777a8a515ac943476fc32d3
      blocked: false
---

# Gate ledger — ariadne#567 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T05:35:48-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `literal-command-format-parity` (kind: defect) Recognize whitespace-separated ps state format fields
  pkg/golint/ps_liveness.go:56–61 splits format text only on commas and does not trim fields. Native /bin/ps accepted "pid stat=", "pid, stat=", and "pid, state=" during review, but the predicate misses all three, allowing the same snapshot oracle through the new guard (ARCH-PURPOSE). The complete affected family in this window is this shared parser across Command and CommandContext, including separate -o arguments, attached -o formats, and --format= formats containing whitespace separators or comma-adjacent whitespace. Normalize supported field separators and add cases for stat/state through both Check and CheckSources.

## Round 2 — 2026-10-23T05:41:42-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — pkg/golint/ps_liveness.go:65 now normalizes comma and whitespace separators in the shared predicate. TestPSLivenessQueriesFormatSeparators covers both constructors, separate and attached formats, and both Check entry points. All 40 cases pass at head and fail with the previous comma-only parser restored through a temporary Go overlay.

## Open findings

(none — every finding has been disposed)
