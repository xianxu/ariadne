---
gate: boundary-review
issue: 388
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 731b8f80-206f-49c0-9f9e-0cc700156ee7
            repository: github.com/xianxu/ariadne
            issue_id: "000388"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/8773d3e5a93da4ab
            reviewed_head: 94c04e25375bc72795d15dff740af616e51ecafd
            slot: ariadne:2
            requested_at: "2026-10-15T18:50:28.889412Z"
          phase: finished
          pid: 82276
          started_at: "2026-10-15T18:50:29.290554Z"
          finished_at: "2026-10-15T18:50:58.035524Z"
          session_id: 2c467df3-27d8-4423-be7b-069b9ca33162
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 94c04e25375bc72795d15dff740af616e51ecafd
        tracker_head: 21ffc918fcf1b9a62d2bb492ea5324405730ae1b
        branch_point: fdd52ed51646063f819155d794b4998c97298b44
        since: "2026-10-15T11:11:59-07:00"
        until: "2026-10-15T11:50:01-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T11:11:59-07:00"
              end: "2026-10-15T11:50:01-07:00"
              began: "2026-10-15T11:11:59-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: fbde86d545b98a4661110b3fa02ec75e7c63b5bd62e9c8636a87691d9ca3ab84
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 8475726
              windows:
                - start: "2026-10-15T11:11:59-07:00"
                  end: "2026-10-15T11:50:01-07:00"
      "n": 1
      timestamp: "2026-10-15T11:51:04-07:00"
      dispatched: "2026-10-15T11:50:28-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: Follow-ups routed to Needs you may strain its "only asks, no status" definition
          detail: help.md:169 routes a later follow-up to Needs you, but help.md:139-141 limits Needs you to asks/decisions/actions with no status. Only instance in the window. Optionally say "a follow-up that needs the operator goes to Needs you; anything else goes into a new bullet's answer or Threads".
          family: section-contract-consistency
          round: 1
      recipe: small-diff-review
      isolation: fenced
      reviewed: 94c04e25375bc72795d15dff740af616e51ecafd
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a2a02621-dcde-4d45-bbda-62493f95a2e3
            repository: github.com/xianxu/ariadne
            issue_id: "000388"
            recipe: fix-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/c9a6ecbdad635c1d
            reviewed_head: 1fada25d45c0c0bfc77d15abb1855a3327ff531c
            slot: ariadne:2
            requested_at: "2026-10-15T18:51:36.792899Z"
          phase: finished
          pid: 4053
          started_at: "2026-10-15T18:51:36.885018Z"
          finished_at: "2026-10-15T18:52:16.508615Z"
          session_id: 2fe6a619-6710-45e1-9c4c-6f4993775d8a
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 1fada25d45c0c0bfc77d15abb1855a3327ff531c
        tracker_head: d79ade94d2c94c2e55d7e009f5fdc709901fa754
        branch_point: fdd52ed51646063f819155d794b4998c97298b44
        since: "2026-10-15T11:11:59-07:00"
        until: "2026-10-15T11:51:21-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T11:11:59-07:00"
              end: "2026-10-15T11:51:21-07:00"
              began: "2026-10-15T11:11:59-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: ae5fd6e0206ee8040a94da0b0943c4f9848f63e4321a6e831b4a2ecba0c1a3da
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 8505434
              windows:
                - start: "2026-10-15T11:11:59-07:00"
                  end: "2026-10-15T11:51:21-07:00"
        cuts:
            - dispatched: "2026-10-15T11:50:28-07:00"
              finalized: "2026-10-15T11:51:04-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-15T11:52:17-07:00"
      dispatched: "2026-10-15T11:51:36-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: help.md:169-171 now routes to Needs you only "if it needs the operator, else into Threads or a new bullet's answer", consistent with help.md:139-141; guards pass.
          round: 2
      findings:
        - kind: defect
          id: BR-2
          severity: Minor
          title: atlas tl-skills.md:34 still routes all later follow-ups to Needs you unconditionally
          detail: '2nd finding in this family: the BR-1 fix updated help.md but the atlas restatement still says "later follow-ups go to Needs you", contradicting the asks-only rule. Rule: any restatement of the routing must keep the "if it needs the operator" condition or not restate it; simplest is for the atlas to keep only "closes its QA file, never into that file" and defer to tl-lead --help. Prevalence: 2 sites restate the routing, 1 still has the old wording.'
          family: section-contract-consistency
          round: 2
      recipe: fix-review
      isolation: fenced
      reviewed: 1fada25d45c0c0bfc77d15abb1855a3327ff531c
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a2e3152b-703f-49d7-8938-b1d746ce8c66
            repository: github.com/xianxu/ariadne
            issue_id: "000388"
            recipe: fix-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/2ca7c60a4ec7eacc
            reviewed_head: 4ccf516b8845760352ecc66b107bc5b384bab86b
            slot: ariadne:2
            requested_at: "2026-10-15T18:52:49.978834Z"
          phase: finished
          pid: 9528
          started_at: "2026-10-15T18:52:50.076366Z"
          finished_at: "2026-10-15T18:53:07.637103Z"
          session_id: d76bc1fb-7d6c-4e24-8c17-8d0eb6025e51
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 4ccf516b8845760352ecc66b107bc5b384bab86b
        tracker_head: e47d3b4d2f8405cbab53ec34c10ac8b5df4fbdbc
        branch_point: ea2cee461c058080a25b03e689e9142c3ea21a63
        since: "2026-10-15T11:11:59-07:00"
        until: "2026-10-15T11:52:34-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T11:11:59-07:00"
              end: "2026-10-15T11:52:34-07:00"
              began: "2026-10-15T11:11:59-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c01b5a49286fb01081a8df0a0f28fd1ebb41d60bf89b9176a5dd87492768ee10
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 8539785
              windows:
                - start: "2026-10-15T11:11:59-07:00"
                  end: "2026-10-15T11:52:34-07:00"
        cuts:
            - dispatched: "2026-10-15T11:50:28-07:00"
              finalized: "2026-10-15T11:51:04-07:00"
              finalizing: true
            - dispatched: "2026-10-15T11:51:36-07:00"
              finalized: "2026-10-15T11:52:17-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-15T11:53:08-07:00"
      dispatched: "2026-10-15T11:52:49-07:00"
      agent: claude
      dispose:
        - id: BR-2
          disposition: addressed
          note: tl-skills.md:34 now says only "closes its QA file, and no follow-up goes into that file" and defers to tl-lead --help; tree sweep finds help.md:168-171 (already conditioned) as the sole remaining routing statement.
          round: 3
      recipe: fix-review
      isolation: fenced
      reviewed: 4ccf516b8845760352ecc66b107bc5b384bab86b
      blocked: false
---

# Gate ledger — ariadne#388 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-15T11:51:04-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `section-contract-consistency` (kind: defect) Follow-ups routed to Needs you may strain its "only asks, no status" definition
  help.md:169 routes a later follow-up to Needs you, but help.md:139-141 limits Needs you to asks/decisions/actions with no status. Only instance in the window. Optionally say "a follow-up that needs the operator goes to Needs you; anything else goes into a new bullet's answer or Threads".

## Round 2 — 2026-10-15T11:52:17-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — help.md:169-171 now routes to Needs you only "if it needs the operator, else into Threads or a new bullet's answer", consistent with help.md:139-141; guards pass.

### Raised

- **BR-2** [Minor] `section-contract-consistency` (kind: defect) atlas tl-skills.md:34 still routes all later follow-ups to Needs you unconditionally
  2nd finding in this family: the BR-1 fix updated help.md but the atlas restatement still says "later follow-ups go to Needs you", contradicting the asks-only rule. Rule: any restatement of the routing must keep the "if it needs the operator" condition or not restate it; simplest is for the atlas to keep only "closes its QA file, never into that file" and defer to tl-lead --help. Prevalence: 2 sites restate the routing, 1 still has the old wording.

## Round 3 — 2026-10-15T11:53:08-07:00 (claude) — passed

### Disposed

- BR-2 — addressed — tl-skills.md:34 now says only "closes its QA file, and no follow-up goes into that file" and defers to tl-lead --help; tree sweep finds help.md:168-171 (already conditioned) as the sole remaining routing statement.

## Open findings

(none — every finding has been disposed)
