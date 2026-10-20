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

## Open findings

- **BR-1** [Important] `landed-evidence-ignores-remote-branch` (kind: defect) issueBranchLanded reads only the local canonical ref; an absent local branch reads as landed even when the remote branch holds unlanded commits
- **BR-2** [Minor] `plan-entity-drift` (kind: defect) Plan step names landedIssue(mergedPR, localTip, onMain); code delivers landedPR(prs) plus the issueBranchLanded wrapper
- **BR-3** [Minor] `implicit-precondition-coupling` (kind: defect) closeBranchRecoveryNote claims already landed on any tip mismatch, relying on requireCloseIssueBranch having admitted the branch first
