---
gate: boundary-review
issue: 483
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 1077074f-152d-4a25-afec-931671823f12
            repository: github.com/xianxu/ariadne
            issue_id: "000483"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/e829a2102610eb6e
            reviewed_head: 3c683e9cd174f608896220678ea12328eedd20f2
            slot: ariadne:5
            requested_at: "2026-10-20T13:07:23.179361Z"
          phase: finished
          pid: 66692
          started_at: "2026-10-20T13:07:23.386153Z"
          finished_at: "2026-10-20T13:19:10.676064Z"
          session_id: 3184c72e-9656-425e-9fb0-05b6fe39d7cd
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 3c683e9cd174f608896220678ea12328eedd20f2
        tracker_head: ddbea4968217ddbcecfbb708995e0aa5c429665f
        branch_point: 48c1d7a3c8f766bcd3e2662a6354b6d1cfb287ad
        since: "2026-10-20T04:06:52-07:00"
        until: "2026-10-20T05:29:41-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T04:06:52-07:00"
              end: "2026-10-20T05:29:41-07:00"
              began: "2026-10-20T04:06:52-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 2b1c35f1-9bee-4064-83d0-2a67ebaaf586
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 68a8492a7fa027299110470c99de17c5d452b4a8cffd028d447630d83a64815d
              started: "2026-10-19T16:59:53.709Z"
              spawned: false
              bytes: 9017478
              windows:
                - start: "2026-10-20T04:06:52-07:00"
                  end: "2026-10-20T05:29:41-07:00"
        queue_waits:
            - attempt_id: 3dade12eeb6446cd74f67ff743a9ff3b
              slot: ariadne:5
              record_digest: d8865c4c6156d2b0805bc3acd8213411983b8a70243a6a5d59cbaaacbdd74381
              start: "2026-10-20T11:46:31.794188Z"
              end: "2026-10-20T12:19:28.778072Z"
              queue_seconds: 1976.983884
              raw_seconds: 1976.983884
              exact: true
      "n": 1
      timestamp: "2026-10-20T06:19:11-07:00"
      dispatched: "2026-10-20T06:07:23-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: issueBranchLanded reads only the local canonical ref; an absent local branch reads as landed even when the remote branch holds unlanded commits
          detail: 'landedissue.go:61-66 uses for-each-ref refs/heads/<name> and returns landed when empty. In a slot without the local branch, a remote canonical branch pushed past the merged PR head (mid-issue squash landing, then continued work) is ignored: start-plan refuses instead of restoring it, and close admits a prefixed follow-up branch. Fix by using issueTip (local, else remote) and add a test where the local branch is deleted and the remote is ahead of HeadOID.'
          family: landed-evidence-ignores-remote-branch
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: Plan step names landedIssue(mergedPR, localTip, onMain); code delivers landedPR(prs) plus the issueBranchLanded wrapper
          detail: Add a Revisions entry so the plan matches the delivered entities and the LandingPRs/HeadOID evidence source.
          family: plan-entity-drift
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: closeBranchRecoveryNote claims already landed on any tip mismatch, relying on requireCloseIssueBranch having admitted the branch first
          detail: trackerbranch.go:75; document the precondition or gate the message on issueBranchLanded.
          family: implicit-precondition-coupling
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 392fba76-64cd-4feb-9092-f723d031abf3
            repository: github.com/xianxu/ariadne
            issue_id: "000483"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/b501ea97e89cb8ef
            reviewed_head: 69d2987ceff329b44ff50f7c8478a823c80698e2
            slot: ariadne:5
            requested_at: "2026-10-20T13:48:13.779615Z"
          phase: finished
          pid: 87293
          started_at: "2026-10-20T13:48:13.867911Z"
          finished_at: "2026-10-20T14:00:39.239899Z"
          session_id: 3692e931-8925-4998-a303-40a49d74bbd2
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 69d2987ceff329b44ff50f7c8478a823c80698e2
        tracker_head: b87aeaf90da3d46653d08b4a79e00208927424b0
        branch_point: 48c1d7a3c8f766bcd3e2662a6354b6d1cfb287ad
        since: "2026-10-20T04:06:52-07:00"
        until: "2026-10-20T06:20:54-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T04:06:52-07:00"
              end: "2026-10-20T06:20:54-07:00"
              began: "2026-10-20T04:06:52-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 2b1c35f1-9bee-4064-83d0-2a67ebaaf586
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 29635e2bdabe448df01a769fd21cb4a59c882c68925516a66b438cfc5d2cb7c3
              started: "2026-10-19T16:59:53.709Z"
              spawned: false
              bytes: 9127231
              windows:
                - start: "2026-10-20T04:06:52-07:00"
                  end: "2026-10-20T06:20:54-07:00"
        queue_waits:
            - attempt_id: 3dade12eeb6446cd74f67ff743a9ff3b
              slot: ariadne:5
              record_digest: d8865c4c6156d2b0805bc3acd8213411983b8a70243a6a5d59cbaaacbdd74381
              start: "2026-10-20T11:46:31.794188Z"
              end: "2026-10-20T12:19:28.778072Z"
              queue_seconds: 1976.983884
              raw_seconds: 1976.983884
              exact: true
            - attempt_id: 8a3847f6060794d92dba362ef9a00f99
              slot: ariadne:5
              record_digest: c32a68b6228907396153b4c6d1e1060f11d0e7bf15320e509dd5f170e1e055fc
              start: "2026-10-20T12:29:48.500907Z"
              end: "2026-10-20T12:57:18.13363Z"
              queue_seconds: 1649.632723
              raw_seconds: 1649.632723
              exact: true
        cuts:
            - dispatched: "2026-10-20T06:07:23-07:00"
              finalized: "2026-10-20T06:19:11-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-20T07:00:40-07:00"
      dispatched: "2026-10-20T06:48:13-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: landedissue.go:58 now uses issueTip (local else remote); TestRemoteCanonicalAheadOfMergedHeadIsNotLanded goes red when the local-only for-each-ref read is restored (verified by revert in this checkout).
          round: 2
        - id: BR-2
          disposition: addressed
          note: Issue Revisions 2026-10-20 names landedPRs(prs) plus issueBranchLanded and LandingPRs/HeadOID as the evidence source, matching landedissue.go:17-79.
          round: 2
        - id: BR-3
          disposition: addressed
          note: trackerbranch.go:70-72 doc comment states the precondition (describes only branches requireCloseIssueBranch admitted); requireCloseIssueBranch admits a tip mismatch only after issueBranchLanded.
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: 69d2987ceff329b44ff50f7c8478a823c80698e2
      blocked: false
---

# Gate ledger — ariadne#483 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T06:19:11-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `landed-evidence-ignores-remote-branch` (kind: defect) issueBranchLanded reads only the local canonical ref; an absent local branch reads as landed even when the remote branch holds unlanded commits
  landedissue.go:61-66 uses for-each-ref refs/heads/<name> and returns landed when empty. In a slot without the local branch, a remote canonical branch pushed past the merged PR head (mid-issue squash landing, then continued work) is ignored: start-plan refuses instead of restoring it, and close admits a prefixed follow-up branch. Fix by using issueTip (local, else remote) and add a test where the local branch is deleted and the remote is ahead of HeadOID.
- **BR-2** [Minor] `plan-entity-drift` (kind: defect) Plan step names landedIssue(mergedPR, localTip, onMain); code delivers landedPR(prs) plus the issueBranchLanded wrapper
  Add a Revisions entry so the plan matches the delivered entities and the LandingPRs/HeadOID evidence source.
- **BR-3** [Minor] `implicit-precondition-coupling` (kind: defect) closeBranchRecoveryNote claims already landed on any tip mismatch, relying on requireCloseIssueBranch having admitted the branch first
  trackerbranch.go:75; document the precondition or gate the message on issueBranchLanded.

## Round 2 — 2026-10-20T07:00:40-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — landedissue.go:58 now uses issueTip (local else remote); TestRemoteCanonicalAheadOfMergedHeadIsNotLanded goes red when the local-only for-each-ref read is restored (verified by revert in this checkout).
- BR-2 — addressed — Issue Revisions 2026-10-20 names landedPRs(prs) plus issueBranchLanded and LandingPRs/HeadOID as the evidence source, matching landedissue.go:17-79.
- BR-3 — addressed — trackerbranch.go:70-72 doc comment states the precondition (describes only branches requireCloseIssueBranch admitted); requireCloseIssueBranch admits a tip mismatch only after issueBranchLanded.

## Open findings

(none — every finding has been disposed)
