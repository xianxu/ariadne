---
gate: boundary-review
issue: 415
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 9f699c5f-e7f1-489f-a49c-126afeef5ca2
            repository: github.com/xianxu/ariadne
            issue_id: "000415"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/6daf07d4d8089694
            reviewed_head: 614b8f11f0eb8bd986d57eac0df1963831c3eaf9
            slot: ariadne:2
            requested_at: "2026-10-17T02:43:21.74706Z"
          phase: finished
          pid: 37113
          started_at: "2026-10-17T02:43:21.883985Z"
          finished_at: "2026-10-17T02:45:26.004376Z"
          session_id: 283abace-6516-4233-8215-cc8a4f1b8d71
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 614b8f11f0eb8bd986d57eac0df1963831c3eaf9
        tracker_head: b6ecd1c8f9948b5b3a320b934335c122a2328e2d
        branch_point: 63b1a010fc175881b8311e8588676a70991a8ad3
        since: "2026-10-16T18:46:26-07:00"
        until: "2026-10-16T18:58:58-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T18:46:26-07:00"
              end: "2026-10-16T18:58:58-07:00"
              began: "2026-10-16T18:46:26-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: cef1fc1627d99cff087e3c83a99d7861f3cafbec55cc07f7565ce86517d131ac
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 18695192
              windows:
                - start: "2026-10-16T18:46:26-07:00"
                  end: "2026-10-16T18:58:58-07:00"
      "n": 1
      timestamp: "2026-10-16T19:45:26-07:00"
      dispatched: "2026-10-16T19:43:21-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: datatype lint is not invoked by any make target, gate, or CI in derived repositories
          detail: Only TestEveryTypeDeclaresACollaborationMode exercises it in ariadne; a derived repo's local prototype without a collaboration-mode row passes unless the operator runs the command by hand.
          family: lint-not-wired-to-gate
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: The collaboration-mode value set is restated in Go, pensive.cue and collaboration-modes.md; the CUE scaffold value is untyped
          detail: Only the Go vocab test rejects a typo in scaffold.collaboration_mode; a shared CUE enum would make the CUE model enforce it too.
          family: enum-single-source
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: collaborationModeDefault accepts the first backticked token anywhere in the notes, not at the start; CRLF prototypes report a misleading error
          detail: The docs say the notes start with the default. A CRLF file fails with "no Frontmatter shape section".
          family: parser-matches-stated-contract
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: TestEveryTypeDeclaresACollaborationMode indexes protos[i] without checking i != -1
          detail: If project ever stops resolving, the test panics instead of reporting a failure.
          family: test-guards-lookup
          round: 1
      recipe: milestone-review
      isolation: fenced
      reviewed: 614b8f11f0eb8bd986d57eac0df1963831c3eaf9
      blocked: false
---

# Gate ledger — ariadne#415 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-16T19:45:26-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `lint-not-wired-to-gate` (kind: defect) datatype lint is not invoked by any make target, gate, or CI in derived repositories
  Only TestEveryTypeDeclaresACollaborationMode exercises it in ariadne; a derived repo's local prototype without a collaboration-mode row passes unless the operator runs the command by hand.
- **BR-2** [Minor] `enum-single-source` (kind: defect) The collaboration-mode value set is restated in Go, pensive.cue and collaboration-modes.md; the CUE scaffold value is untyped
  Only the Go vocab test rejects a typo in scaffold.collaboration_mode; a shared CUE enum would make the CUE model enforce it too.
- **BR-3** [Minor] `parser-matches-stated-contract` (kind: defect) collaborationModeDefault accepts the first backticked token anywhere in the notes, not at the start; CRLF prototypes report a misleading error
  The docs say the notes start with the default. A CRLF file fails with "no Frontmatter shape section".
- **BR-4** [Minor] `test-guards-lookup` (kind: defect) TestEveryTypeDeclaresACollaborationMode indexes protos[i] without checking i != -1
  If project ever stops resolving, the test panics instead of reporting a failure.

## Open findings

- **BR-1** [Minor] `lint-not-wired-to-gate` (kind: defect) datatype lint is not invoked by any make target, gate, or CI in derived repositories
- **BR-2** [Minor] `enum-single-source` (kind: defect) The collaboration-mode value set is restated in Go, pensive.cue and collaboration-modes.md; the CUE scaffold value is untyped
- **BR-3** [Minor] `parser-matches-stated-contract` (kind: defect) collaborationModeDefault accepts the first backticked token anywhere in the notes, not at the start; CRLF prototypes report a misleading error
- **BR-4** [Minor] `test-guards-lookup` (kind: defect) TestEveryTypeDeclaresACollaborationMode indexes protos[i] without checking i != -1
