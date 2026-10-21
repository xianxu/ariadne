---
gate: boundary-review
issue: 517
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: f388efa4-d86e-4ae9-b9cd-311ee9d95164
            repository: github.com/xianxu/ariadne
            issue_id: "000517"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/2cc6a0776fea615d
            reviewed_head: d219882fb8a83a1c5195c0550fb070f3a3407cca
            slot: ariadne:5
            requested_at: "2026-10-21T06:39:44.76938Z"
          phase: finished
          pid: 80871
          started_at: "2026-10-21T06:39:44.864893Z"
          finished_at: "2026-10-21T06:40:47.084969Z"
          session_id: 01a15d2f-efde-7781-8d0a-67ad6b7ed2c0
          outcome: exit-error
        - version: 1
          intent:
            disposable_root: true
            id: e00ec68a-079a-4c8d-9db7-23c38cb7dfcd
            repository: github.com/xianxu/ariadne
            issue_id: "000517"
            recipe: small-diff-review
            round: 1
            attempt: 2
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/2cc6a0776fea615d
            reviewed_head: d219882fb8a83a1c5195c0550fb070f3a3407cca
            slot: ariadne:5
            requested_at: "2026-10-21T06:40:48.886659Z"
          phase: finished
          pid: 92335
          started_at: "2026-10-21T06:40:49.045085Z"
          finished_at: "2026-10-21T06:45:50.605354Z"
          session_id: 01a15d30-eaef-7a73-a0a2-b096cbe46bb6
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: d219882fb8a83a1c5195c0550fb070f3a3407cca
        tracker_head: 9dfb97fa01910c816fc449265800ab3e7f4abf91
        branch_point: 993b4be4acf30b88cf3215fb5270dbfc776d4333
        since: "2026-10-20T23:25:39-07:00"
        until: "2026-10-20T23:39:29-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T23:25:39-07:00"
              end: "2026-10-20T23:39:29-07:00"
              began: "2026-10-20T23:25:39-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: abd882ed2c21a776704a03db1c197314f98d774dd50b3ebb23c3c2dbf909f665
              started: "2026-10-20T15:58:37.432Z"
              spawned: false
              bytes: 13311211
              windows:
                - start: "2026-10-20T23:25:39-07:00"
                  end: "2026-10-20T23:39:29-07:00"
      "n": 1
      timestamp: "2026-10-20T23:45:51-07:00"
      dispatched: "2026-10-20T23:39:44-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Preparation survives SIGTERM because its context does not own CLI signals
          detail: 'cmd/sdlc/localcheck_graphprepare.go:23-33 creates a separate process group but receives the background CLI context: main.go:83-98 excludes local-check run from signal ownership, and runUnderSlot installs signal handling only later. A reproduction using the pinned-head binary observed CLI exit -15 after SIGTERM while the preparation process remained alive; neither its timeout nor deferred checkout cleanup survives the CLI''s death. The complete family in this window is the new preparation seam under SIGINT and SIGTERM; the context-cancellation fixture at localcheck_graphprepare_process_test.go:28 bypasses both. Own signals for this resource lifetime, ensure cancellation cannot fall through to a suite using a live parent context, and add real CLI signal tests asserting descendant shutdown, cleanup, and no passing publication. This leaves ARCH-PURPOSE''s promised cancellation behavior incomplete.'
          family: owned-subprocess-cancellation
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Done when omits the revised immutable-main declaration requirement
          detail: workshop/issues/000517-affected-graph-preparation.md:32-37 remains unchanged although Revisions at lines 88-92 adds immutable main declaration authority and localcheck_cmd.go:142-174 implements it. The complete missing acceptance family here is declaration provenance under candidate graph_prepare edits and origin/main advancing after its commit is pinned; both are exercised by TestAffectedGraphPreparationFromPinnedMain but neither is required by Done when. Add an explicit acceptance clause and a timestamped Revisions entry connecting it to the existing fixture.
          family: acceptance-contract-traceability
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#517 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T23:45:51-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `owned-subprocess-cancellation` (kind: defect) Preparation survives SIGTERM because its context does not own CLI signals
  cmd/sdlc/localcheck_graphprepare.go:23-33 creates a separate process group but receives the background CLI context: main.go:83-98 excludes local-check run from signal ownership, and runUnderSlot installs signal handling only later. A reproduction using the pinned-head binary observed CLI exit -15 after SIGTERM while the preparation process remained alive; neither its timeout nor deferred checkout cleanup survives the CLI's death. The complete family in this window is the new preparation seam under SIGINT and SIGTERM; the context-cancellation fixture at localcheck_graphprepare_process_test.go:28 bypasses both. Own signals for this resource lifetime, ensure cancellation cannot fall through to a suite using a live parent context, and add real CLI signal tests asserting descendant shutdown, cleanup, and no passing publication. This leaves ARCH-PURPOSE's promised cancellation behavior incomplete.
- **BR-2** [Important] `acceptance-contract-traceability` (kind: defect) Done when omits the revised immutable-main declaration requirement
  workshop/issues/000517-affected-graph-preparation.md:32-37 remains unchanged although Revisions at lines 88-92 adds immutable main declaration authority and localcheck_cmd.go:142-174 implements it. The complete missing acceptance family here is declaration provenance under candidate graph_prepare edits and origin/main advancing after its commit is pinned; both are exercised by TestAffectedGraphPreparationFromPinnedMain but neither is required by Done when. Add an explicit acceptance clause and a timestamped Revisions entry connecting it to the existing fixture.

## Open findings

- **BR-1** [Critical] `owned-subprocess-cancellation` (kind: defect) Preparation survives SIGTERM because its context does not own CLI signals
- **BR-2** [Important] `acceptance-contract-traceability` (kind: defect) Done when omits the revised immutable-main declaration requirement
