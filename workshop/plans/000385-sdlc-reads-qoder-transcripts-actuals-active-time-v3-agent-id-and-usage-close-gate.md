---
gate: boundary-review
issue: 385
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: f19d47bd-d5af-41fd-8399-853500dce3c4
            repository: github.com/xianxu/ariadne
            issue_id: "000385"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/1f99f510016bcc98
            reviewed_head: d4c9d29701d73e39a2718e22398b484d4c694460
            slot: ariadne:3
            requested_at: "2026-10-15T18:42:18.468641Z"
          phase: finished
          pid: 38476
          started_at: "2026-10-15T18:42:18.70588Z"
          finished_at: "2026-10-15T18:45:11.374353Z"
          session_id: 01a140df-4f76-7342-8669-a30a2b5886ed
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: d4c9d29701d73e39a2718e22398b484d4c694460
        tracker_head: 21ffc918fcf1b9a62d2bb492ea5324405730ae1b
        branch_point: 9d03c5735e15a06f70e552dfcbcdd21f4d52e343
        since: "2026-10-15T09:33:04-07:00"
        until: "2026-10-15T11:41:48-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T09:33:04-07:00"
              end: "2026-10-15T11:41:48-07:00"
              began: "2026-10-15T09:33:04-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a138fd-1ac0-7a83-826e-60623c226386
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: c5684de9a8f25398e32a10b0274bfab9f3f1eac2099a79b2cd1351bb3a7f1f36
              started: "2026-10-14T05:58:30.184Z"
              spawned: false
              bytes: 41949541
              windows:
                - start: "2026-10-15T09:33:04-07:00"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a14074-981e-74f2-af8c-a6ddf477fc6d
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: a9886dc5ea07325421724dbd964a263d52c38a01666ce05f66b4eb483be6e957
              started: "2026-10-15T16:45:47.228Z"
              spawned: false
              bytes: 758658
              windows:
                - start: "2026-10-15T16:45:47.228Z"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a14079-9246-7ea3-91fb-c457e3886da9
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 1f1689dceec0617335228ae3ffa7158942cc8f392d39eed6bd1e8e0323cd5c72
              started: "2026-10-15T16:51:14.597Z"
              spawned: false
              bytes: 453960
              windows:
                - start: "2026-10-15T16:51:14.597Z"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a1407a-af6f-7152-833b-a43cafac9231
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 467ba6fd9787954503097fe5d6f76e5f3eacf111156a527937edb4a48121c222
              started: "2026-10-15T16:52:26.427Z"
              spawned: false
              bytes: 462233
              windows:
                - start: "2026-10-15T16:52:26.427Z"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a1407b-1424-7c31-964f-c6839b6c5951
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 9bfc53e280ec295615d2f254082bc30c1f99379efc2f80040203fcef658d93a0
              started: "2026-10-15T16:52:54.44Z"
              spawned: false
              bytes: 420615
              windows:
                - start: "2026-10-15T16:52:54.44Z"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a1407b-ee59-71d1-98d5-f8a27ea6615a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: 9efef5aa5f9e50eef2d508357bf71aee275445db8b2a49fb239a8fb0e8da70a8
              started: "2026-10-15T16:53:46.292Z"
              spawned: true
              bytes: 611271
              windows:
                - start: "2026-10-15T16:53:46.292Z"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a1407d-b788-7921-97ed-2aa273e67570
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: 23ed2409e504187554c6d6aa4eba3868f59c74027970bb50e97cb1bd40c15e76
              started: "2026-10-15T16:55:43.336Z"
              spawned: true
              bytes: 806710
              windows:
                - start: "2026-10-15T16:55:43.336Z"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a1407e-0e37-7f82-9b51-1e2425d58944
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: 55afe6e1788546348bdf2b561bae29401bfb9ab0da7a8b1288c2e96b41876b19
              started: "2026-10-15T16:56:05.521Z"
              spawned: true
              bytes: 1268169
              windows:
                - start: "2026-10-15T16:56:05.521Z"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a14080-72b2-7e43-bc09-63ffe1e09102
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: e1a877e295cf76c410c5c74c7a817e582583faf1e507b98fe977b821fe7d9247
              started: "2026-10-15T16:58:42.319Z"
              spawned: true
              bytes: 1553084
              windows:
                - start: "2026-10-15T16:58:42.319Z"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a140bb-446b-78a2-bad0-8a87ab45e0d6
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: bea2c41d92607fceb2ea2d50a0cf2458e402a022f020adbe379546da89cb68ea
              started: "2026-10-15T18:02:57.072Z"
              spawned: true
              bytes: 811720
              windows:
                - start: "2026-10-15T18:02:57.072Z"
                  end: "2026-10-15T11:41:48-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.03448333333333333
      "n": 1
      timestamp: "2026-10-15T11:45:12-07:00"
      dispatched: "2026-10-15T11:42:18-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Important
          title: Qoder nonzero token counters lack the promised format diagnostic
          detail: cmd/sdlc/internal/usage/qoder.go:117 ignores the parsed token fields. A temporary overlay test with input_tokens=123 and output_tokens=456 returned null tokens but only sources/records coverage and spend_complete=true. The plan at lines 137–140 explicitly requires a visible diagnostic for unrecognized counter semantics. Validate all four placeholder fields, retain null tokens, and add regression coverage for nonzero and invalid values (ARCH-SECURE, ARCH-PURPOSE).
          family: unsupported-format-visibility
          round: 1
        - id: BR-2
          severity: Important
          title: README omits Qoder usage and links to a removed atlas anchor
          detail: README.md:130–145 still describes Claude/Codex tokens only, omitting Qoder credit reporting, calibration support, and schema-version-2 nullable measurements. Its local-token-usage-365 link no longer matches the renamed atlas heading. Update this passage and link in the same boundary, as required by the README docs gate (ARCH-PURPOSE).
          family: user-surface-documentation
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#385 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-15T11:45:12-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `unsupported-format-visibility` Qoder nonzero token counters lack the promised format diagnostic
  cmd/sdlc/internal/usage/qoder.go:117 ignores the parsed token fields. A temporary overlay test with input_tokens=123 and output_tokens=456 returned null tokens but only sources/records coverage and spend_complete=true. The plan at lines 137–140 explicitly requires a visible diagnostic for unrecognized counter semantics. Validate all four placeholder fields, retain null tokens, and add regression coverage for nonzero and invalid values (ARCH-SECURE, ARCH-PURPOSE).
- **BR-2** [Important] `user-surface-documentation` README omits Qoder usage and links to a removed atlas anchor
  README.md:130–145 still describes Claude/Codex tokens only, omitting Qoder credit reporting, calibration support, and schema-version-2 nullable measurements. Its local-token-usage-365 link no longer matches the renamed atlas heading. Update this passage and link in the same boundary, as required by the README docs gate (ARCH-PURPOSE).

## Open findings

- **BR-1** [Important] `unsupported-format-visibility` Qoder nonzero token counters lack the promised format diagnostic
- **BR-2** [Important] `user-surface-documentation` README omits Qoder usage and links to a removed atlas anchor
