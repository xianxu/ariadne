---
gate: boundary-review
issue: 460
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: ba87cce4-a5a9-43eb-b772-f48b40b305f4
            repository: github.com/xianxu/ariadne
            issue_id: "000460"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/dcf792c35ad57902
            reviewed_head: 13be7fd15e87c19edb264fd7d9384b38e2b3e30e
            slot: ariadne:1
            requested_at: "2026-10-19T10:28:51.707302Z"
          phase: finished
          pid: 92026
          started_at: "2026-10-19T10:28:51.925481Z"
          finished_at: "2026-10-19T10:35:01.633718Z"
          session_id: 01a153b4-fc42-7f61-a1eb-c4566455f7ba
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 13be7fd15e87c19edb264fd7d9384b38e2b3e30e
        tracker_head: ec71ed794e86f9299192aff52dd5d59f35c73c85
        branch_point: 2abdfb582380a38e3cf4add6130128aa0d5e67aa
        since: "2026-10-19T00:51:23-07:00"
        until: "2026-10-19T02:43:53-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T00:51:23-07:00"
              end: "2026-10-19T02:43:53-07:00"
              began: "2026-10-19T00:51:23-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 2faa155ff32494254aa11a32755c88c4763833956123ec950d50cfe3d018b916
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 137108094
              windows:
                - start: "2026-10-19T00:51:23-07:00"
                  end: "2026-10-19T02:43:53-07:00"
            - harness: codex
              id: 01a1532b-6586-7fa1-8acb-3efd179b40c5
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: bee81db81aa397b38a11dcbd9168944e4e75a31a3d19ddb3ba96c4faa82af295
              started: "2026-10-19T07:58:36.782Z"
              spawned: false
              bytes: 645988
              windows:
                - start: "2026-10-19T07:58:36.782Z"
                  end: "2026-10-19T02:43:53-07:00"
            - harness: codex
              id: 01a1532e-045a-7482-b76e-ae908931f88d
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: d8bbfbdc745d733f6369334794a1d4384e0e6d5d43b90bab08ff1311669f736a
              started: "2026-10-19T08:01:27.772Z"
              spawned: false
              bytes: 452865
              windows:
                - start: "2026-10-19T08:01:27.772Z"
                  end: "2026-10-19T02:43:53-07:00"
            - harness: codex
              id: 01a1532f-46e0-7510-8cc6-a1f8ae0b00ca
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 5b54466b6be0590d5796a17508e1fd7b0ddbdcdf22cbae0e27d1d78fec89484a
              started: "2026-10-19T08:02:50.411Z"
              spawned: false
              bytes: 380642
              windows:
                - start: "2026-10-19T08:02:50.411Z"
                  end: "2026-10-19T02:43:53-07:00"
            - harness: codex
              id: 01a15331-cc8d-70f1-8ac0-7b2aa1094989
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 1686f24838d0c406633ffc50f25f06995159152715b3acc63a9cb250a08aa6ee
              started: "2026-10-19T08:05:35.816Z"
              spawned: false
              bytes: 425372
              windows:
                - start: "2026-10-19T08:05:35.816Z"
                  end: "2026-10-19T02:43:53-07:00"
            - harness: codex
              id: 01a15332-3230-7401-85f1-610944224225
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 352ad35f74299fb660cabd060217f00568e88a067c7b81ee8c84ea2fef5e47f5
              started: "2026-10-19T08:06:01.544Z"
              spawned: false
              bytes: 380437
              windows:
                - start: "2026-10-19T08:06:01.544Z"
                  end: "2026-10-19T02:43:53-07:00"
      "n": 1
      timestamp: "2026-10-19T03:35:19-07:00"
      dispatched: "2026-10-19T03:28:51-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Tracker validation accepts non-string completion tokens
          detail: cmd/sdlc/internal/issue/card.go:95 delegates nested validation to cardEnvelope, but handoff.go:97 uses yaml.Node.Decode, which coerces numeric and boolean scalars into strings. A temporary regression replacing completion.token with 123 or true showed ParseCard accepts both, violating the Spec's strict known-field types. Validate known YAML node kinds and tags before struct decoding across all nested tracker records, and add permanent rejection tests.
          family: strict-known-field-types
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Duplicate-key validation rejects valid unknown JSON numbers
          detail: cmd/sdlc/teststats_store.go:314 constructs a decoder without UseNumber, so validateJSONValue converts unknown numeric tokens to float64. Adding "future":{"number":1e400} to an otherwise valid attempt makes decodeSuiteAttempt reject it with "cannot unmarshal number 1e400 into Go value of type float64". UseNumber preserves opaque numeric tokens during duplicate checking; add root and nested regressions while retaining strict known-number validation.
          family: opaque-unknown-field-values
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#460 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T03:35:19-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `strict-known-field-types` (kind: defect) Tracker validation accepts non-string completion tokens
  cmd/sdlc/internal/issue/card.go:95 delegates nested validation to cardEnvelope, but handoff.go:97 uses yaml.Node.Decode, which coerces numeric and boolean scalars into strings. A temporary regression replacing completion.token with 123 or true showed ParseCard accepts both, violating the Spec's strict known-field types. Validate known YAML node kinds and tags before struct decoding across all nested tracker records, and add permanent rejection tests.
- **BR-2** [Important] `opaque-unknown-field-values` (kind: defect) Duplicate-key validation rejects valid unknown JSON numbers
  cmd/sdlc/teststats_store.go:314 constructs a decoder without UseNumber, so validateJSONValue converts unknown numeric tokens to float64. Adding "future":{"number":1e400} to an otherwise valid attempt makes decodeSuiteAttempt reject it with "cannot unmarshal number 1e400 into Go value of type float64". UseNumber preserves opaque numeric tokens during duplicate checking; add root and nested regressions while retaining strict known-number validation.

## Open findings

- **BR-1** [Critical] `strict-known-field-types` (kind: defect) Tracker validation accepts non-string completion tokens
- **BR-2** [Important] `opaque-unknown-field-values` (kind: defect) Duplicate-key validation rejects valid unknown JSON numbers
