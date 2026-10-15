---
gate: boundary-review
issue: 350
id_prefix: BR
rounds:
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: b696b7d263d37bc9e4287a0c73ebd13f920afdf5
        tracker_head: 11346d7d8670215cdb8a925e96e50b21ab84730d
        branch_point: 1c39182532fa3cc78ea4935abff4ab1b87a6f01d
        since: "2026-10-13T17:19:46-07:00"
        until: "2026-10-13T18:32:41-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-13T17:19:46-07:00"
              end: "2026-10-13T18:32:41-07:00"
              began: "2026-10-13T17:19:46-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: bd759883-0d13-49e6-8d0c-73ff96809176
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5c34a25ce33e7b3b7cd7e70959429d55d73e25fcd681beea2202618f73af9d4a
              started: "2026-10-14T00:28:59.212Z"
              spawned: false
              bytes: 421205
              windows:
                - start: "2026-10-14T00:28:59.212Z"
                  end: "2026-10-13T18:32:41-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 8782a827ae1f577e22e936106a153674d6e285b9a73da68ab1ac78a37b528986
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 3258388
              windows:
                - start: "2026-10-13T17:19:46-07:00"
                  end: "2026-10-13T18:32:41-07:00"
      "n": 1
      timestamp: "2026-10-14T08:47:13-07:00"
      dispatched: "2026-10-14T08:44:50-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: README still lists only sdlc/datatype/vocabulary/doc-review as ariadne's make tools binaries
          detail: README.md:195 is stale now that the Makefile tools target builds tl-worker and tl-lead. Add both binaries and a one-line pointer to their --help.
          family: readme-tool-surface-sync
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: TestWorkerHelpDoesNotRestateSection15 omits the planned "whole Bash call" refusal; help.md:11 contains it
          detail: Either reword the citation in cmd/tl-worker/help.md or add a plan Revisions entry saying the one-line citation is allowed.
          family: plan-test-contract-drift
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: Live session-sync conformance never exercises fold, so the fake's fold refusals are unchecked against couch
          detail: cmd/tl-lead/conformance_test.go never folds. The argv form matches couch help, but the fold-not-owed refusal the fake models has no live check. Revisit at the M3 close.
          family: conformance-covers-modeled-behavior
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: Plan names ErrDeliveryGap and sdlc_fake_test.go; the code has neither (fake is fake_test.go, no gap error)
          detail: Add Revisions entries so the Core concepts table and the SessionSync description match the code.
          family: plan-core-concepts-drift
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: A fleet-inventory failure aborts the whole trains render, including the independent Trains block
          detail: cmd/tl-lead/trains.go:68 returns fleetErr before rendering the Trains block. It could render the off-train block as ? instead.
          family: degrade-visibly-not-abort
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 392c1079c8f81086a51ede0b695c4cc6e8bebdcc
        tracker_head: 11346d7d8670215cdb8a925e96e50b21ab84730d
        branch_point: 1c39182532fa3cc78ea4935abff4ab1b87a6f01d
        since: "2026-10-13T17:19:46-07:00"
        until: "2026-10-14T08:49:10-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-13T17:19:46-07:00"
              end: "2026-10-14T08:49:10-07:00"
              began: "2026-10-13T17:19:46-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: bd759883-0d13-49e6-8d0c-73ff96809176
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5c34a25ce33e7b3b7cd7e70959429d55d73e25fcd681beea2202618f73af9d4a
              started: "2026-10-14T00:28:59.212Z"
              spawned: false
              bytes: 421205
              windows:
                - start: "2026-10-14T00:28:59.212Z"
                  end: "2026-10-14T08:49:10-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5f209512d67c50e4e8f85ef6895bdc36f617afbb66fb503454f3f37c953c4313
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 3433594
              windows:
                - start: "2026-10-13T17:19:46-07:00"
                  end: "2026-10-14T08:49:10-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.052116666666666665
        cuts:
            - boundary: M1
              dispatched: "2026-10-14T08:44:50-07:00"
              finalized: "2026-10-14T08:47:13-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-14T08:51:31-07:00"
      dispatched: "2026-10-14T08:49:27-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: README.md:195 and atlas base-layer.md:75, setup-and-replication.md:54, sandbox.md:13 all list tl-worker and tl-lead; a grep sweep finds no stale list.
          round: 2
        - id: BR-2
          disposition: addressed
          note: help.md:11 reworded; the test forbids "whole Bash call"; restoring the old help.md fails TestWorkerHelpDoesNotRestateSection15.
          round: 2
        - id: BR-3
          disposition: addressed
          note: conformance_test.go:53-68 folds an ID not owed against real couch; it passed live on this host with TL_LEAD_LIVE=1.
          round: 2
        - id: BR-4
          disposition: addressed
          note: ErrDeliveryGap exists and is mapped in decodeReply; reverting the mapping fails TestDecodeReplyDeliveryGap; a plan Revisions entry fixes the fake file names.
          round: 2
        - id: BR-5
          disposition: addressed
          note: trains.go:63-70 renders Trains and marks Slots unread; reverting fails TestGatherTrainsFleetFailureDegrades.
          round: 2
      findings:
        - kind: defect
          id: BR-6
          severity: Minor
          title: TestSessionSyncConformance doc comment still says the test never folds
          detail: cmd/tl-lead/conformance_test.go:16-17 says "the test never folds", but lines 53-68 now fold a not-owed ID; reword it to say the fold names an ID that isn't owed.
          family: comment-tracks-code
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      reviewed: 392c1079c8f81086a51ede0b695c4cc6e8bebdcc
      blocked: false
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 9cb63797c7e4567798c31622edb61fe68a432110
        tracker_head: 11346d7d8670215cdb8a925e96e50b21ab84730d
        branch_point: 1c39182532fa3cc78ea4935abff4ab1b87a6f01d
        since: "2026-10-13T17:19:46-07:00"
        until: "2026-10-14T08:51:52-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-13T17:19:46-07:00"
              end: "2026-10-14T08:51:52-07:00"
              began: "2026-10-13T17:19:46-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: bd759883-0d13-49e6-8d0c-73ff96809176
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5c34a25ce33e7b3b7cd7e70959429d55d73e25fcd681beea2202618f73af9d4a
              started: "2026-10-14T00:28:59.212Z"
              spawned: false
              bytes: 421205
              windows:
                - start: "2026-10-14T00:28:59.212Z"
                  end: "2026-10-14T08:51:52-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 9b0d8d962954805f1159c7b4cf6633c5e305da24252ea3b93be3d39a5ea87038
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 3490413
              windows:
                - start: "2026-10-13T17:19:46-07:00"
                  end: "2026-10-14T08:51:52-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.052116666666666665
        cuts:
            - boundary: M1
              dispatched: "2026-10-14T08:44:50-07:00"
              finalized: "2026-10-14T08:47:13-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-14T08:49:27-07:00"
              finalized: "2026-10-14T08:51:31-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-14T08:53:42-07:00"
      dispatched: "2026-10-14T08:52:50-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-7
          severity: Minor
          title: tl-skills atlas does not mention the degraded off-train render when the fleet read fails
          detail: 'atlas/workflow/tl-skills.md:33 says a failed car read renders as ?, but not that a failed sdlc fleet inventory now renders "? off-train slots unread: <reason>" (TestGatherTrainsFleetFailureDegrades); add one clause at M3''s atlas pass.'
          family: atlas-tracks-behavior
          round: 3
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      reviewed: 9cb63797c7e4567798c31622edb61fe68a432110
      blocked: false
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 81db31ea07a5b1eda21876c95d20150b5f2b1345
        tracker_head: bf0ee6bf72cf5ebd646dc1115c840d9028ec7807
        branch_point: 1c39182532fa3cc78ea4935abff4ab1b87a6f01d
        since: "2026-10-13T17:19:46-07:00"
        until: "2026-10-14T08:57:16-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-13T17:19:46-07:00"
              end: "2026-10-14T08:57:16-07:00"
              began: "2026-10-13T17:19:46-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: bd759883-0d13-49e6-8d0c-73ff96809176
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5c34a25ce33e7b3b7cd7e70959429d55d73e25fcd681beea2202618f73af9d4a
              started: "2026-10-14T00:28:59.212Z"
              spawned: false
              bytes: 421205
              windows:
                - start: "2026-10-14T00:28:59.212Z"
                  end: "2026-10-14T08:57:16-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 26b66958451af76f2edef3405602e7c39713065af3838fd1d6e7c293b372e415
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 3658872
              windows:
                - start: "2026-10-13T17:19:46-07:00"
                  end: "2026-10-14T08:57:16-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.052116666666666665
        cuts:
            - boundary: M1
              dispatched: "2026-10-14T08:44:50-07:00"
              finalized: "2026-10-14T08:47:13-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-14T08:49:27-07:00"
              finalized: "2026-10-14T08:51:31-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-14T08:52:50-07:00"
              finalized: "2026-10-14T08:53:42-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-14T08:59:14-07:00"
      dispatched: "2026-10-14T08:57:39-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-8
          severity: Important
          title: TestMilestoneTickRegex pins a stale hand copy of the pre-fix tick regex, not markdoc.TickMilestone
          detail: cmd/sdlc/close_test.go:164-172 claims to mirror runClose's tick regex but builds the old pattern without milestoneEmphasis, which diverged from pkg/markdoc/plan.go:67 in this window. Its boundary cases (M40, M4b, M4-extra, [x] unchanged) never reach the production function. Move them into a pkg/markdoc table over TickMilestone, including bold forms, and delete the copy.
          family: tests-exercise-production-code
          round: 4
      boundary: M3
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 0058e96c01c1287501f0129f9735bc5062029e78
        tracker_head: 8fe4bb4ef49ab9e0e465e4e00fc521af3821dd7d
        branch_point: 1c39182532fa3cc78ea4935abff4ab1b87a6f01d
        since: "2026-10-13T17:19:46-07:00"
        until: "2026-10-14T09:01:12-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-13T17:19:46-07:00"
              end: "2026-10-14T09:01:12-07:00"
              began: "2026-10-13T17:19:46-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: bd759883-0d13-49e6-8d0c-73ff96809176
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5c34a25ce33e7b3b7cd7e70959429d55d73e25fcd681beea2202618f73af9d4a
              started: "2026-10-14T00:28:59.212Z"
              spawned: false
              bytes: 421205
              windows:
                - start: "2026-10-14T00:28:59.212Z"
                  end: "2026-10-14T09:01:12-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 80c2a4d5ff16e30e57972af3dce14bca575777e0482a7a30486fd068cb2df22e
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 3748112
              windows:
                - start: "2026-10-13T17:19:46-07:00"
                  end: "2026-10-14T09:01:12-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.052116666666666665
        cuts:
            - boundary: M1
              dispatched: "2026-10-14T08:44:50-07:00"
              finalized: "2026-10-14T08:47:13-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-14T08:49:27-07:00"
              finalized: "2026-10-14T08:51:31-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-14T08:52:50-07:00"
              finalized: "2026-10-14T08:53:42-07:00"
              finalizing: true
            - boundary: M3
              dispatched: "2026-10-14T08:57:39-07:00"
              finalized: "2026-10-14T08:59:14-07:00"
              finalizing: false
      "n": 5
      timestamp: "2026-10-14T09:12:37-07:00"
      dispatched: "2026-10-14T09:01:52-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: addressed
          note: Hand copy deleted from cmd/sdlc/close_test.go; boundary cases now drive markdoc.TickMilestone (plan_test.go:146-168, incl. bold); reverting milestoneEmphasis at plan.go:67 turns both new tests red; sweep finds no other regex copies.
          round: 5
      boundary: M3
      recipe: milestone-review
      isolation: fenced
      reviewed: 0058e96c01c1287501f0129f9735bc5062029e78
      blocked: false
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 60e462468f2815550b4026cda2fe6e2b5c10db3b
        tracker_head: 2f329519753ec1c2fb07d4dcdcb1477435a6ec97
        branch_point: 1c39182532fa3cc78ea4935abff4ab1b87a6f01d
        since: "2026-10-13T17:19:46-07:00"
        until: "2026-10-14T09:53:52-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-13T17:19:46-07:00"
              end: "2026-10-14T09:53:52-07:00"
              began: "2026-10-13T17:19:46-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: bd759883-0d13-49e6-8d0c-73ff96809176
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5c34a25ce33e7b3b7cd7e70959429d55d73e25fcd681beea2202618f73af9d4a
              started: "2026-10-14T00:28:59.212Z"
              spawned: false
              bytes: 421205
              windows:
                - start: "2026-10-14T00:28:59.212Z"
                  end: "2026-10-14T09:53:52-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: e6dfb40c22bd678974a23a7275f9842088b2da40966acb301ebd0ba6a7002310
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 4336683
              windows:
                - start: "2026-10-13T17:19:46-07:00"
                  end: "2026-10-14T09:53:52-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.052116666666666665
        cuts:
            - boundary: M1
              dispatched: "2026-10-14T08:44:50-07:00"
              finalized: "2026-10-14T08:47:13-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-14T08:49:27-07:00"
              finalized: "2026-10-14T08:51:31-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-14T08:52:50-07:00"
              finalized: "2026-10-14T08:53:42-07:00"
              finalizing: true
            - boundary: M3
              dispatched: "2026-10-14T08:57:39-07:00"
              finalized: "2026-10-14T08:59:14-07:00"
              finalizing: false
            - boundary: M3
              dispatched: "2026-10-14T09:01:52-07:00"
              finalized: "2026-10-14T09:12:37-07:00"
              finalizing: true
      "n": 6
      timestamp: "2026-10-14T10:30:47-07:00"
      dispatched: "2026-10-14T10:28:56-07:00"
      agent: claude
      dispose:
        - id: BR-6
          disposition: addressed
          note: cmd/tl-lead/conformance_test.go:12-17 now says the only fold names an ID that isn't owed, matching lines 53-68.
          round: 6
        - id: BR-7
          disposition: withdrawn
          note: Overtaken by the scope revision (cc315116, db844120). The off-train render was removed; no "off-train" text remains in cmd/tl-lead or atlas.
          round: 6
      findings:
        - kind: defect
          id: BR-9
          severity: Critical
          title: Plan Core concepts and Revisions still claim RenderOffTrain and the off-train degraded render; slots and threads entities are missing
          detail: '2nd in family plan-core-concepts-drift. Plan line 28 lists RenderOffTrain in cmd/tl-lead/render.go, and lines 7, 38, 138, 153 and 164 describe the off-train block, TestRenderOffTrain and TestGatherTrainsFleetFailureDegrades. None of these exist after db844120. There are no rows for slotLines/renderSlots/gatherSlots (slots.go) or upsertThread (threads.go), and the scope decision was logged only in the issue. Rule: any commit that adds or removes an entity, or records a scope decision, updates Core concepts and Revisions in the same commit. Sweep every row and revision claim against the tree.'
          family: plan-core-concepts-drift
          round: 6
        - kind: defect
          id: BR-10
          severity: Minor
          title: Comments name a nonexistent carReadConcurrency and claim the slot Phase carries the read error
          detail: '2nd in family comment-tracks-code. trains.go:66 names carReadConcurrency (the constant is readConcurrency). slots.go:18 says Phase is "?" with the read error, but gatherSlots (slots.go:86) discards the error. Rule: a comment naming an identifier or carried data must match the code. Sweep with grep for identifiers that don''t resolve.'
          family: comment-tracks-code
          round: 6
        - kind: defect
          id: BR-11
          severity: Minor
          title: threadBlocks splits on any line starting with "### ", including inside a code fence
          detail: cmd/tl-lead/threads.go:15 isn't fence-aware. A pasted thread body with a fenced line starting "### " would be split into a spurious subsection. pkg/markdoc already has a fence scanner it could reuse (ARCH-DRY).
          family: reuse-fence-aware-scanner
          round: 6
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: f1fde220fd4ea232791be65302e340a1728e4f59
        tracker_head: 2f329519753ec1c2fb07d4dcdcb1477435a6ec97
        branch_point: 1c39182532fa3cc78ea4935abff4ab1b87a6f01d
        since: "2026-10-13T17:19:46-07:00"
        until: "2026-10-14T10:32:55-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-13T17:19:46-07:00"
              end: "2026-10-14T10:32:55-07:00"
              began: "2026-10-13T17:19:46-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: bd759883-0d13-49e6-8d0c-73ff96809176
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5c34a25ce33e7b3b7cd7e70959429d55d73e25fcd681beea2202618f73af9d4a
              started: "2026-10-14T00:28:59.212Z"
              spawned: false
              bytes: 421205
              windows:
                - start: "2026-10-14T00:28:59.212Z"
                  end: "2026-10-14T10:32:55-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 91afa70c23133a5073cf54461bb186c9561c46be7ef2e71431e00e720ed71e77
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 4517963
              windows:
                - start: "2026-10-13T17:19:46-07:00"
                  end: "2026-10-14T10:32:55-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.052116666666666665
        cuts:
            - boundary: M1
              dispatched: "2026-10-14T08:44:50-07:00"
              finalized: "2026-10-14T08:47:13-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-14T08:49:27-07:00"
              finalized: "2026-10-14T08:51:31-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-14T08:52:50-07:00"
              finalized: "2026-10-14T08:53:42-07:00"
              finalizing: true
            - boundary: M3
              dispatched: "2026-10-14T08:57:39-07:00"
              finalized: "2026-10-14T08:59:14-07:00"
              finalizing: false
            - boundary: M3
              dispatched: "2026-10-14T09:01:52-07:00"
              finalized: "2026-10-14T09:12:37-07:00"
              finalizing: true
            - dispatched: "2026-10-14T10:28:56-07:00"
              finalized: "2026-10-14T10:30:47-07:00"
              finalizing: false
      "n": 7
      timestamp: "2026-10-14T10:38:15-07:00"
      dispatched: "2026-10-14T10:33:42-07:00"
      agent: claude
      dispose:
        - id: BR-9
          disposition: addressed
          note: Core concepts lists slotLines/renderSlots/SlotLine, threadBlocks/upsertThread, lineDiff, renderedSections; RenderOffTrain/ThreadEntries marked deleted; Revisions 2026-10-14 supersedes TestRenderOffTrain and TestGatherTrainsFleetFailureDegrades by name; grep of plan identifiers resolves except deleted/superseded ones.
          round: 7
        - id: BR-10
          disposition: addressed
          note: trains.go:66 now names readConcurrency; slots.go:18 says "?" when the observation failed, matching gatherSlots; comment-identifier sweep over tl-lead/tl-worker/pkg found no unresolved names.
          round: 7
        - id: BR-11
          disposition: addressed
          note: threads.go:23 uses markdoc.FenceSpans; neutralizing the guard turns TestUpsertThread's fenced-### row red.
          round: 7
      findings:
        - kind: defect
          id: BR-12
          severity: Minor
          title: Core concepts rows name workerHelp/leadHelp and the note names PlanBar; the code has rootHelp and planBar
          detail: 3rd in family. Plan rows 23-24 cite workerHelp/leadHelp but both embeds are rootHelp (cmd/tl-lead/main.go:19, cmd/tl-worker/main.go:18); entity note line 41 says PlanBar while code and table say planBar. The lessons.md rule (grep each cited identifier before close) covers this; mechanize that grep as a close step rather than patching instances.
          family: plan-core-concepts-drift
          round: 7
      recipe: milestone-review
      isolation: fenced
      reviewed: f1fde220fd4ea232791be65302e340a1728e4f59
      blocked: false
---

# Gate ledger — ariadne#350 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T08:47:13-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `readme-tool-surface-sync` (kind: defect) README still lists only sdlc/datatype/vocabulary/doc-review as ariadne's make tools binaries
  README.md:195 is stale now that the Makefile tools target builds tl-worker and tl-lead. Add both binaries and a one-line pointer to their --help.
- **BR-2** [Minor] `plan-test-contract-drift` (kind: defect) TestWorkerHelpDoesNotRestateSection15 omits the planned "whole Bash call" refusal; help.md:11 contains it
  Either reword the citation in cmd/tl-worker/help.md or add a plan Revisions entry saying the one-line citation is allowed.
- **BR-3** [Minor] `conformance-covers-modeled-behavior` (kind: defect) Live session-sync conformance never exercises fold, so the fake's fold refusals are unchecked against couch
  cmd/tl-lead/conformance_test.go never folds. The argv form matches couch help, but the fold-not-owed refusal the fake models has no live check. Revisit at the M3 close.
- **BR-4** [Minor] `plan-core-concepts-drift` (kind: defect) Plan names ErrDeliveryGap and sdlc_fake_test.go; the code has neither (fake is fake_test.go, no gap error)
  Add Revisions entries so the Core concepts table and the SessionSync description match the code.
- **BR-5** [Minor] `degrade-visibly-not-abort` (kind: defect) A fleet-inventory failure aborts the whole trains render, including the independent Trains block
  cmd/tl-lead/trains.go:68 returns fleetErr before rendering the Trains block. It could render the off-train block as ? instead.

## Round 2 — 2026-10-14T08:51:31-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — README.md:195 and atlas base-layer.md:75, setup-and-replication.md:54, sandbox.md:13 all list tl-worker and tl-lead; a grep sweep finds no stale list.
- BR-2 — addressed — help.md:11 reworded; the test forbids "whole Bash call"; restoring the old help.md fails TestWorkerHelpDoesNotRestateSection15.
- BR-3 — addressed — conformance_test.go:53-68 folds an ID not owed against real couch; it passed live on this host with TL_LEAD_LIVE=1.
- BR-4 — addressed — ErrDeliveryGap exists and is mapped in decodeReply; reverting the mapping fails TestDecodeReplyDeliveryGap; a plan Revisions entry fixes the fake file names.
- BR-5 — addressed — trains.go:63-70 renders Trains and marks Slots unread; reverting fails TestGatherTrainsFleetFailureDegrades.

### Raised

- **BR-6** [Minor] `comment-tracks-code` (kind: defect) TestSessionSyncConformance doc comment still says the test never folds
  cmd/tl-lead/conformance_test.go:16-17 says "the test never folds", but lines 53-68 now fold a not-owed ID; reword it to say the fold names an ID that isn't owed.

## Round 3 — 2026-10-14T08:53:42-07:00 (claude) — passed

### Raised

- **BR-7** [Minor] `atlas-tracks-behavior` (kind: defect) tl-skills atlas does not mention the degraded off-train render when the fleet read fails
  atlas/workflow/tl-skills.md:33 says a failed car read renders as ?, but not that a failed sdlc fleet inventory now renders "? off-train slots unread: <reason>" (TestGatherTrainsFleetFailureDegrades); add one clause at M3's atlas pass.

## Round 4 — 2026-10-14T08:59:14-07:00 (claude) — BLOCKED

### Raised

- **BR-8** [Important] `tests-exercise-production-code` (kind: defect) TestMilestoneTickRegex pins a stale hand copy of the pre-fix tick regex, not markdoc.TickMilestone
  cmd/sdlc/close_test.go:164-172 claims to mirror runClose's tick regex but builds the old pattern without milestoneEmphasis, which diverged from pkg/markdoc/plan.go:67 in this window. Its boundary cases (M40, M4b, M4-extra, [x] unchanged) never reach the production function. Move them into a pkg/markdoc table over TickMilestone, including bold forms, and delete the copy.

## Round 5 — 2026-10-14T09:12:37-07:00 (claude) — passed

### Disposed

- BR-8 — addressed — Hand copy deleted from cmd/sdlc/close_test.go; boundary cases now drive markdoc.TickMilestone (plan_test.go:146-168, incl. bold); reverting milestoneEmphasis at plan.go:67 turns both new tests red; sweep finds no other regex copies.

## Round 6 — 2026-10-14T10:30:47-07:00 (claude) — BLOCKED

### Disposed

- BR-6 — addressed — cmd/tl-lead/conformance_test.go:12-17 now says the only fold names an ID that isn't owed, matching lines 53-68.
- BR-7 — withdrawn — Overtaken by the scope revision (cc315116, db844120). The off-train render was removed; no "off-train" text remains in cmd/tl-lead or atlas.

### Raised

- **BR-9** [Critical] `plan-core-concepts-drift` (kind: defect) Plan Core concepts and Revisions still claim RenderOffTrain and the off-train degraded render; slots and threads entities are missing
  2nd in family plan-core-concepts-drift. Plan line 28 lists RenderOffTrain in cmd/tl-lead/render.go, and lines 7, 38, 138, 153 and 164 describe the off-train block, TestRenderOffTrain and TestGatherTrainsFleetFailureDegrades. None of these exist after db844120. There are no rows for slotLines/renderSlots/gatherSlots (slots.go) or upsertThread (threads.go), and the scope decision was logged only in the issue. Rule: any commit that adds or removes an entity, or records a scope decision, updates Core concepts and Revisions in the same commit. Sweep every row and revision claim against the tree.
- **BR-10** [Minor] `comment-tracks-code` (kind: defect) Comments name a nonexistent carReadConcurrency and claim the slot Phase carries the read error
  2nd in family comment-tracks-code. trains.go:66 names carReadConcurrency (the constant is readConcurrency). slots.go:18 says Phase is "?" with the read error, but gatherSlots (slots.go:86) discards the error. Rule: a comment naming an identifier or carried data must match the code. Sweep with grep for identifiers that don't resolve.
- **BR-11** [Minor] `reuse-fence-aware-scanner` (kind: defect) threadBlocks splits on any line starting with "### ", including inside a code fence
  cmd/tl-lead/threads.go:15 isn't fence-aware. A pasted thread body with a fenced line starting "### " would be split into a spurious subsection. pkg/markdoc already has a fence scanner it could reuse (ARCH-DRY).

## Round 7 — 2026-10-14T10:38:15-07:00 (claude) — passed

### Disposed

- BR-9 — addressed — Core concepts lists slotLines/renderSlots/SlotLine, threadBlocks/upsertThread, lineDiff, renderedSections; RenderOffTrain/ThreadEntries marked deleted; Revisions 2026-10-14 supersedes TestRenderOffTrain and TestGatherTrainsFleetFailureDegrades by name; grep of plan identifiers resolves except deleted/superseded ones.
- BR-10 — addressed — trains.go:66 now names readConcurrency; slots.go:18 says "?" when the observation failed, matching gatherSlots; comment-identifier sweep over tl-lead/tl-worker/pkg found no unresolved names.
- BR-11 — addressed — threads.go:23 uses markdoc.FenceSpans; neutralizing the guard turns TestUpsertThread's fenced-### row red.

### Raised

- **BR-12** [Minor] `plan-core-concepts-drift` (kind: defect) Core concepts rows name workerHelp/leadHelp and the note names PlanBar; the code has rootHelp and planBar
  3rd in family. Plan rows 23-24 cite workerHelp/leadHelp but both embeds are rootHelp (cmd/tl-lead/main.go:19, cmd/tl-worker/main.go:18); entity note line 41 says PlanBar while code and table say planBar. The lessons.md rule (grep each cited identifier before close) covers this; mechanize that grep as a close step rather than patching instances.

## Open findings

- **BR-12** [Minor] `plan-core-concepts-drift` (kind: defect) Core concepts rows name workerHelp/leadHelp and the note names PlanBar; the code has rootHelp and planBar
