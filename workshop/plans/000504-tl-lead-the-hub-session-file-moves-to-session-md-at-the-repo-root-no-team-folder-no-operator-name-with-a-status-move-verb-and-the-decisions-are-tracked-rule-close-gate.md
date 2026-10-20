---
gate: boundary-review
issue: 504
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 80cb73ad-e614-4abb-9c54-b7c3f4b0dab1
            repository: github.com/xianxu/ariadne
            issue_id: "000504"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/25dc9c80b0c19429
            reviewed_head: ff57bc497da70f69e3e03043ea38168870e41899
            slot: ariadne:4
            requested_at: "2026-10-20T17:47:02.538692Z"
          phase: finished
          pid: 58982
          started_at: "2026-10-20T17:47:02.677981Z"
          finished_at: "2026-10-20T17:49:09.870174Z"
          session_id: 01a15a6c-821e-7eb3-bd5c-d1a71058d553
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: ff57bc497da70f69e3e03043ea38168870e41899
        tracker_head: 0072d9e4991b99e41ff38f88fe2aa6f4112be614
        branch_point: fdfc3ca3c297c1c6c2802e03f3ea54b73f4a2b3e
        since: "2026-10-20T09:56:17-07:00"
        until: "2026-10-20T10:46:40-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T09:56:17-07:00"
              end: "2026-10-20T10:46:40-07:00"
              began: "2026-10-20T09:56:17-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: codex
              id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 64260ed5b95913c2221bd63750587aa4a12b5365282d022348dc9f43deea7c46
              started: "2026-10-20T16:55:54.591Z"
              spawned: false
              bytes: 3892127
              windows:
                - start: "2026-10-20T09:56:17-07:00"
                  end: "2026-10-20T10:46:40-07:00"
            - harness: codex
              id: 01a15a3e-8844-7f72-b197-96fc83f81539
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 99f4fe4c8943f116c817c6f49191601750ba7359a2b97219b98d0b55624c89a9
              started: "2026-10-20T16:56:50.162Z"
              spawned: true
              bytes: 429479
              windows:
                - start: "2026-10-20T16:56:50.162Z"
                  end: "2026-10-20T10:46:40-07:00"
            - harness: codex
              id: 01a15a42-5379-7c20-88c4-31004da81ecc
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 4e4c9bc1741e2528e5051d11802a748f5f372ca5332b3434ded06fc5caf6adf7
              started: "2026-10-20T17:00:59.003Z"
              spawned: false
              bytes: 763820
              windows:
                - start: "2026-10-20T17:00:59.003Z"
                  end: "2026-10-20T10:46:40-07:00"
            - harness: codex
              id: 01a15a45-27e9-7720-875d-4c8a5b2772e1
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 1135a9640ff15cd6fbb4ea5f6dbd7032bd62470fa26f54b5b9887c31c685fc66
              started: "2026-10-20T17:04:04.565Z"
              spawned: false
              bytes: 426336
              windows:
                - start: "2026-10-20T17:04:04.565Z"
                  end: "2026-10-20T10:46:40-07:00"
            - harness: codex
              id: 01a15a46-07a6-72c0-82cb-44bec02002b5
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 5bfb3fe7a915a4d656ff9fc353b586004651c3a5db2eca26b8f02cd1dbe0c13f
              started: "2026-10-20T17:05:01.724Z"
              spawned: false
              bytes: 456616
              windows:
                - start: "2026-10-20T17:05:01.724Z"
                  end: "2026-10-20T10:46:40-07:00"
            - harness: codex
              id: 01a15a46-8683-7420-a06f-df0dff734fa1
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: d639dea7a97c7fda61a1498c3bd55ae064ba78bcc121eedf02ec8c40ab988e6e
              started: "2026-10-20T17:05:34.342Z"
              spawned: false
              bytes: 394281
              windows:
                - start: "2026-10-20T17:05:34.342Z"
                  end: "2026-10-20T10:46:40-07:00"
            - harness: codex
              id: 01a15a47-64cc-7310-b6d4-0d5c94a3f3cc
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 820b4ce19aac4522ef3faa693fcbcb6f4509dc0881c4797ac9437133db6f0cbb
              started: "2026-10-20T17:06:30.831Z"
              spawned: true
              bytes: 611122
              windows:
                - start: "2026-10-20T17:06:30.831Z"
                  end: "2026-10-20T10:46:40-07:00"
            - harness: codex
              id: 01a15a47-8d09-72d0-bc5e-59ffcf634459
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: ef8f65fe4b671657e5c95d5c071a40a3eac4103d889709a3ab12ed99196d3573
              started: "2026-10-20T17:06:41.209Z"
              spawned: true
              bytes: 428852
              windows:
                - start: "2026-10-20T17:06:41.209Z"
                  end: "2026-10-20T10:46:40-07:00"
            - harness: codex
              id: 01a15a69-4c0c-7a52-ab62-17119854390f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: ebd12c3a091cc94658e41e7188525eb145a39b575f44d918aef9073b74d621bc
              started: "2026-10-20T17:43:32.788Z"
              spawned: true
              bytes: 464087
              windows:
                - start: "2026-10-20T17:43:32.788Z"
                  end: "2026-10-20T10:46:40-07:00"
      "n": 1
      timestamp: "2026-10-20T10:49:10-07:00"
      dispatched: "2026-10-20T10:47:02-07:00"
      agent: codex
      reviewer: 'codex (openai; author openai); degraded: same-family by choice'
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Accepted YAML parent values can retain the deleted session path
          detail: 'cmd/tl-lead/status_move.go:70 delegates parent rewriting to sidedoc.RebaseLinks, whose rebaseParent only handles plain values and uncommented double-quoted values. A valid parent: ''../tl-status-op.md'' or parent: ../tl-status-op.md # backlink passes snapshot metadata validation but remains unchanged during relocation; apply then deletes the original session and reports success. ARCH-PURPOSE / ARCH-SECURE: decode and rewrite supported parent scalars structurally, preserving surrounding metadata, or refuse unsupported forms before effects. Add migration regressions covering quotes, tags, and inline comments.'
          family: relocation-preserves-reference-targets
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Existing negated ignore rules defeat migration's local-file guarantee
          detail: 'cmd/tl-lead/status_move.go:102 treats an ignore pattern''s presence anywhere as sufficient. With /session.md followed by !/session.md, it adds no later corrective rule, and runSessionMove copies and switches without checking effective ignored status. The same applies to the other migration-owned patterns. ARCH-PURPOSE: ensure the final rules take precedence and verify effective ignores before publishing copies; add a regression with later negations and assert git check-ignore results.'
          family: migration-enforces-effective-ignore-policy
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#504 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T10:49:10-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `relocation-preserves-reference-targets` (kind: defect) Accepted YAML parent values can retain the deleted session path
  cmd/tl-lead/status_move.go:70 delegates parent rewriting to sidedoc.RebaseLinks, whose rebaseParent only handles plain values and uncommented double-quoted values. A valid parent: '../tl-status-op.md' or parent: ../tl-status-op.md # backlink passes snapshot metadata validation but remains unchanged during relocation; apply then deletes the original session and reports success. ARCH-PURPOSE / ARCH-SECURE: decode and rewrite supported parent scalars structurally, preserving surrounding metadata, or refuse unsupported forms before effects. Add migration regressions covering quotes, tags, and inline comments.
- **BR-2** [Important] `migration-enforces-effective-ignore-policy` (kind: defect) Existing negated ignore rules defeat migration's local-file guarantee
  cmd/tl-lead/status_move.go:102 treats an ignore pattern's presence anywhere as sufficient. With /session.md followed by !/session.md, it adds no later corrective rule, and runSessionMove copies and switches without checking effective ignored status. The same applies to the other migration-owned patterns. ARCH-PURPOSE: ensure the final rules take precedence and verify effective ignores before publishing copies; add a regression with later negations and assert git check-ignore results.

## Open findings

- **BR-1** [Critical] `relocation-preserves-reference-targets` (kind: defect) Accepted YAML parent values can retain the deleted session path
- **BR-2** [Important] `migration-enforces-effective-ignore-policy` (kind: defect) Existing negated ignore rules defeat migration's local-file guarantee
