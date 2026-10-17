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
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 0a7260da-9e79-46e1-95cb-1991e40df217
            repository: github.com/xianxu/ariadne
            issue_id: "000415"
            recipe: fix-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/d3075f6a596001ae
            reviewed_head: 21ba3b6c13b4123a639a58666228c4764436ee22
            slot: ariadne:2
            requested_at: "2026-10-17T02:46:43.158057Z"
          phase: finished
          pid: 25086
          started_at: "2026-10-17T02:46:43.273768Z"
          finished_at: "2026-10-17T02:47:43.271045Z"
          session_id: b61a53e8-fe0f-45f7-b7df-29d05fedb13c
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 21ba3b6c13b4123a639a58666228c4764436ee22
        tracker_head: 68c38c877e3096c702db121c5abb87214fc40e4a
        branch_point: 63b1a010fc175881b8311e8588676a70991a8ad3
        since: "2026-10-16T18:46:26-07:00"
        until: "2026-10-16T19:46:16-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T18:46:26-07:00"
              end: "2026-10-16T19:46:16-07:00"
              began: "2026-10-16T18:46:26-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: ba6e248f3f997cb8ea4f82674690bea8b2b191333cbb9f0b3f52448f87f3aedf
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 18761838
              windows:
                - start: "2026-10-16T18:46:26-07:00"
                  end: "2026-10-16T19:46:16-07:00"
        cuts:
            - dispatched: "2026-10-16T19:43:21-07:00"
              finalized: "2026-10-16T19:45:26-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-16T19:47:53-07:00"
      dispatched: "2026-10-16T19:46:43-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: renderToOutput (run by every repo's weave compile via construct/local/datatype/.dynamic-skill) now calls lintTypes and warns on stderr (main.go:91-96).
          round: 2
        - id: BR-2
          disposition: addressed
          note: TestCollaborationModesAgreeAcrossSources binds pensive.cue's enum and collaboration-modes.md's error text to vocab.CollaborationModes. The scaffold value still has no CUE type; TestScaffoldsDeclareACollaborationMode covers that in Go, which is acceptable at Minor.
          round: 2
        - id: BR-3
          disposition: addressed
          note: CRLF is normalized and the notes must start with the default (CutPrefix on the trimmed cell). Reverting either fix makes TestCollaborationModeDefault fail (CRLF case, and the not-at-the-start case).
          round: 2
        - id: BR-4
          disposition: addressed
          note: collabmode_test.go:71-73 calls t.Fatal when the project lookup returns i < 0.
          round: 2
      findings:
        - kind: defect
          id: BR-5
          severity: Minor
          title: The new compile-time warning loop in renderToOutput has no test
          detail: lintTypes is tested, but no test checks that the render step prints a warning, so the BR-1 wiring could be dropped without any test failing. Acceptable as thin IO glue; a small test of renderToOutput against a fixture tree with a prototype lacking a collaboration-mode row would pin it.
          family: lint-not-wired-to-gate
          round: 2
      recipe: fix-review
      isolation: fenced
      reviewed: 21ba3b6c13b4123a639a58666228c4764436ee22
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: f94fe7c2-3bb0-4aee-b7bf-aa25a0e5c363
            repository: github.com/xianxu/ariadne
            issue_id: "000415"
            recipe: fix-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/c38d9cbf89291776
            reviewed_head: c812b6e5e3756aa302676ecfb1f576a791ac0287
            slot: ariadne:2
            requested_at: "2026-10-17T02:49:05.417384Z"
          phase: finished
          pid: 48653
          started_at: "2026-10-17T02:49:05.594496Z"
          finished_at: "2026-10-17T02:49:46.988917Z"
          session_id: c393b803-80f3-44ab-8211-d37f6f0a3a67
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: c812b6e5e3756aa302676ecfb1f576a791ac0287
        tracker_head: 31666ad49dedf9c42685e429ac123807edea0a76
        branch_point: 63b1a010fc175881b8311e8588676a70991a8ad3
        since: "2026-10-16T18:46:26-07:00"
        until: "2026-10-16T19:48:32-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T18:46:26-07:00"
              end: "2026-10-16T19:48:32-07:00"
              began: "2026-10-16T18:46:26-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 96b304dc7d3fbf5e8d1948937ee13412541276a00707ad93500012fb2d6f024d
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 18817079
              windows:
                - start: "2026-10-16T18:46:26-07:00"
                  end: "2026-10-16T19:48:32-07:00"
        cuts:
            - dispatched: "2026-10-16T19:43:21-07:00"
              finalized: "2026-10-16T19:45:26-07:00"
              finalizing: true
            - dispatched: "2026-10-16T19:46:43-07:00"
              finalized: "2026-10-16T19:47:53-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-16T19:49:54-07:00"
      dispatched: "2026-10-16T19:49:05-07:00"
      agent: claude
      dispose:
        - id: BR-5
          disposition: addressed
          note: TestRenderWarnsOnATypeWithoutADefault (collabmode_test.go:100-124) runs renderToOutput, now taking an io.Writer, on a fixture tree; making the warning print a no-op turns it red ("warnings \"\", want one for bare only").
          round: 3
      recipe: fix-review
      isolation: fenced
      reviewed: c812b6e5e3756aa302676ecfb1f576a791ac0287
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

## Round 2 — 2026-10-16T19:47:53-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — renderToOutput (run by every repo's weave compile via construct/local/datatype/.dynamic-skill) now calls lintTypes and warns on stderr (main.go:91-96).
- BR-2 — addressed — TestCollaborationModesAgreeAcrossSources binds pensive.cue's enum and collaboration-modes.md's error text to vocab.CollaborationModes. The scaffold value still has no CUE type; TestScaffoldsDeclareACollaborationMode covers that in Go, which is acceptable at Minor.
- BR-3 — addressed — CRLF is normalized and the notes must start with the default (CutPrefix on the trimmed cell). Reverting either fix makes TestCollaborationModeDefault fail (CRLF case, and the not-at-the-start case).
- BR-4 — addressed — collabmode_test.go:71-73 calls t.Fatal when the project lookup returns i < 0.

### Raised

- **BR-5** [Minor] `lint-not-wired-to-gate` (kind: defect) The new compile-time warning loop in renderToOutput has no test
  lintTypes is tested, but no test checks that the render step prints a warning, so the BR-1 wiring could be dropped without any test failing. Acceptable as thin IO glue; a small test of renderToOutput against a fixture tree with a prototype lacking a collaboration-mode row would pin it.

## Round 3 — 2026-10-16T19:49:54-07:00 (claude) — passed

### Disposed

- BR-5 — addressed — TestRenderWarnsOnATypeWithoutADefault (collabmode_test.go:100-124) runs renderToOutput, now taking an io.Writer, on a fixture tree; making the warning print a no-op turns it red ("warnings \"\", want one for bare only").

## Open findings

(none — every finding has been disposed)
