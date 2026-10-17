---
gate: boundary-review
issue: 412
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 18bcd2c0-3703-4b84-8353-59b03ee47297
            repository: github.com/xianxu/ariadne
            issue_id: "000412"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/a271c0b6d57e2f38
            reviewed_head: 031ea5c641fd81964c0e613178f1bd93dced516c
            slot: ariadne:2
            requested_at: "2026-10-16T23:55:50.847012Z"
          phase: finished
          pid: 21500
          started_at: "2026-10-16T23:55:51.782743Z"
          finished_at: "2026-10-16T23:57:40.343763Z"
          session_id: ca937b3c-788e-48c2-a03c-5828a970022d
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 031ea5c641fd81964c0e613178f1bd93dced516c
        tracker_head: 27a568ffc9e65d519a006f41cc07fbe5af3f0e35
        branch_point: 35ba24e6cad39344c347af981de7ac8fcc1d1c31
        since: "2026-10-16T15:42:55-07:00"
        until: "2026-10-16T16:54:25-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T15:42:55-07:00"
              end: "2026-10-16T16:54:25-07:00"
              began: "2026-10-16T15:42:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 6681a7b6-a1e9-4bf6-8743-64ae382a6945
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: a01d2d2e30cdee9e2ab0f79bd4c475b0767f8fb71bde51c321f0d54701105ed7
              started: "2026-10-16T22:49:11.289Z"
              spawned: false
              bytes: 542081
              windows:
                - start: "2026-10-16T22:49:11.289Z"
                  end: "2026-10-16T16:54:25-07:00"
            - harness: claude
              id: a3c26c68-b277-41fb-8571-456fa2883740
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5320c188df9f1d072a2c1e3788ff32b5cea4f732e8e180b077700b714d95ba8f
              started: "2026-10-16T22:55:51.076Z"
              spawned: false
              bytes: 509821
              windows:
                - start: "2026-10-16T22:55:51.076Z"
                  end: "2026-10-16T16:54:25-07:00"
            - harness: claude
              id: c16f9f99-f7dc-493c-80e6-9950883fe1cf
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 56c17381f2e60fc56f2ae5acb83b94366afd3a2dea497b7c768fec2238d6894b
              started: "2026-10-16T22:53:46.176Z"
              spawned: false
              bytes: 499558
              windows:
                - start: "2026-10-16T22:53:46.176Z"
                  end: "2026-10-16T16:54:25-07:00"
            - harness: claude
              id: e1a45820-de66-478e-bdce-95f833a64410
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 84b46a5b1ffbb97690e35548a8500ac62e056483b444935a24569d2a8335bc77
              started: "2026-10-16T22:56:14.41Z"
              spawned: false
              bytes: 394689
              windows:
                - start: "2026-10-16T22:56:14.41Z"
                  end: "2026-10-16T16:54:25-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 98fd170565074499c786ba50f996f5a1538335550a74a2171021994b22a3f7a2
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 17062094
              windows:
                - start: "2026-10-16T15:42:55-07:00"
                  end: "2026-10-16T16:54:25-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 6.5026
      "n": 1
      timestamp: "2026-10-16T16:57:45-07:00"
      dispatched: "2026-10-16T16:55:50-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Side-doc reuse by (parent, marker) alone can link a new ask to an unrelated, already-linked discussion
          detail: "collaboration-modes.md §4 step 1 reuses any side doc whose parent and marker match, even one already linked from another section, so a second identical marker (e.g. \U0001F916[why?] elsewhere) silently gets the old discussion; same-round identical markers are unspecified. Restrict reuse to unlinked docs (and same section for discussion) and add a case or check."
          family: reuse-identity-too-coarse
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: tl-lead's instructionBullets defines an item differently from the chat mode tl-lead --help now cites as the rule
          detail: cmd/tl-lead/instructions.go only starts items at unindented `N. ` lines, treats `→ done:` as answered, and absorbs `-`/`*` lines into the previous bullet, while collaboration-modes.md §3 counts -, * and N. items at the shallowest indent; a `- foo` bullet is invisible to list and gets the previous bullet's link. Align the parser (with a table test) or narrow the help.
          family: cited-rule-restated-divergently
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: Done-when 2 (session file runs on collaboration-mode) depends on the TL's ops write; record it as owed in the landing report
          family: done-when-deferred
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: Examples checker does not assert the (folded <date>) link mark for fold cases (rows 7 and 10)
          family: checker-coverage-gap
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#412 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-16T16:57:45-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `reuse-identity-too-coarse` (kind: defect) Side-doc reuse by (parent, marker) alone can link a new ask to an unrelated, already-linked discussion
  collaboration-modes.md §4 step 1 reuses any side doc whose parent and marker match, even one already linked from another section, so a second identical marker (e.g. 🤖[why?] elsewhere) silently gets the old discussion; same-round identical markers are unspecified. Restrict reuse to unlinked docs (and same section for discussion) and add a case or check.
- **BR-2** [Important] `cited-rule-restated-divergently` (kind: defect) tl-lead's instructionBullets defines an item differently from the chat mode tl-lead --help now cites as the rule
  cmd/tl-lead/instructions.go only starts items at unindented `N. ` lines, treats `→ done:` as answered, and absorbs `-`/`*` lines into the previous bullet, while collaboration-modes.md §3 counts -, * and N. items at the shallowest indent; a `- foo` bullet is invisible to list and gets the previous bullet's link. Align the parser (with a table test) or narrow the help.
- **BR-3** [Minor] `done-when-deferred` (kind: defect) Done-when 2 (session file runs on collaboration-mode) depends on the TL's ops write; record it as owed in the landing report
- **BR-4** [Minor] `checker-coverage-gap` (kind: defect) Examples checker does not assert the (folded <date>) link mark for fold cases (rows 7 and 10)

## Open findings

- **BR-1** [Important] `reuse-identity-too-coarse` (kind: defect) Side-doc reuse by (parent, marker) alone can link a new ask to an unrelated, already-linked discussion
- **BR-2** [Important] `cited-rule-restated-divergently` (kind: defect) tl-lead's instructionBullets defines an item differently from the chat mode tl-lead --help now cites as the rule
- **BR-3** [Minor] `done-when-deferred` (kind: defect) Done-when 2 (session file runs on collaboration-mode) depends on the TL's ops write; record it as owed in the landing report
- **BR-4** [Minor] `checker-coverage-gap` (kind: defect) Examples checker does not assert the (folded <date>) link mark for fold cases (rows 7 and 10)
