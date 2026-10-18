---
gate: boundary-review
issue: 409
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: c4bc4de4-13e2-4ce2-b0f5-a06f4142d8e3
            repository: github.com/xianxu/ariadne
            issue_id: "000409"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/76308b093647e048
            reviewed_head: 7fe6c39b917990a37c04efa94aa6b96e954646c9
            slot: ariadne:2
            requested_at: "2026-10-18T12:32:24.753047Z"
          phase: finished
          pid: 18367
          started_at: "2026-10-18T12:32:24.914333Z"
          finished_at: "2026-10-18T12:35:51.678033Z"
          session_id: 0b2d6a50-30f4-45a0-a498-d2d2c3bafad4
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 7fe6c39b917990a37c04efa94aa6b96e954646c9
        tracker_head: 5578fe23834f9c2160c593a05397eaad33b10d1b
        branch_point: d3ebb55260c6647fbc70b423efed20cd8c734c5b
        since: "2026-10-18T05:16:36-07:00"
        until: "2026-10-18T05:31:57-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T05:16:36-07:00"
              end: "2026-10-18T05:31:57-07:00"
              began: "2026-10-18T05:16:36-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 144a7ee5-cc37-41c5-ac73-b88bbf5223ba
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5f3da10d43be251869bdafd6a98084e32ec627ca1a5a4b930a44fbe4f280c081
              started: "2026-10-18T12:25:28.29Z"
              spawned: false
              bytes: 482080
              windows:
                - start: "2026-10-18T12:25:28.29Z"
                  end: "2026-10-18T05:31:57-07:00"
            - harness: claude
              id: 5ca5160d-491c-4aec-b8f5-f5e40a69efb5
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 42052d978d0f33bca789480a4b74830893f6dd041d8d8d097390922c5c9823d6
              started: "2026-10-18T12:21:14.242Z"
              spawned: false
              bytes: 563079
              windows:
                - start: "2026-10-18T12:21:14.242Z"
                  end: "2026-10-18T05:31:57-07:00"
            - harness: claude
              id: 6be380fe-6c30-4408-b5b9-026237f91e30
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 70f6e935edb05d9fad671da7d73d9b88df89415544fb98457e596558f0fb3332
              started: "2026-10-18T12:23:49.894Z"
              spawned: false
              bytes: 497183
              windows:
                - start: "2026-10-18T12:23:49.894Z"
                  end: "2026-10-18T05:31:57-07:00"
            - harness: claude
              id: 946b723a-6614-4e42-8189-fff378170870
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 18bbe8b8ed1e1d030e367159b09a2e30f9b27a75f8dcc5f5048ff5d6d9bb34f1
              started: "2026-10-18T12:25:42.657Z"
              spawned: false
              bytes: 366511
              windows:
                - start: "2026-10-18T12:25:42.657Z"
                  end: "2026-10-18T05:31:57-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 46da947c66a4904a05bce842137321a76f9d16471d09636a5b6b06196fb57eb8
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 150009716
              windows:
                - start: "2026-10-18T05:16:36-07:00"
                  end: "2026-10-18T05:31:57-07:00"
      "n": 1
      timestamp: "2026-10-18T05:35:52-07:00"
      dispatched: "2026-10-18T05:32:24-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: reviewer_order lacking another family makes Reviewer() silently return the author's own agent
          detail: Mode other with reviewer_order {"bounded":["claude"]} (or an empty list) and a claude main agent makes ReviewPlan return an empty plan, and Reviewer() at agent_resolver.go:85-91 falls through to `return main`, so claude reviews claude with no notice. Validate (agent_resolver.go:120-129) accepts the config. Reject empty, duplicate or single-family orders in Validate, and never return main under other. ARCH-PURPOSE.
          family: no-silent-same-family-review
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Unavailable can never classify a codex or gemini quota or auth failure; the fallback is inert for the default cross-family reviewer
          detail: dispatch.go:629-633 runs ReadStream only for claude, so codex and gemini runs have Stream=false and Unavailable (family.go:131-134) returns false except on a launch error. M1 promised tests for the quota and auth texts each CLI prints; TestUnavailable has only claude stream-json fixtures. Parse codex exec --json (and gemini's output) into Began, or state and test the limit in a plan Revision. ARCH-PURPOSE, ARCH-ORDER.
          family: unavailable-evidence-per-cli
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: close help still says gemini is reviewed by claude; M1 changed it to codex, and the new review.json keys are missing
          detail: cmd/sdlc/helptext/close.md:417 describes the old `other` mapping. The changed judge_test.go:1336 now expects codex. The REVIEWER section omits reviewer_order and reviewer_models, and atlas sdlc-binary.md:1191 still describes other as claude and codex only. A cheap fix, and it should not wait for M2.
          family: docs-track-behavior-change
          round: 1
        - kind: defect
          id: BR-4
          severity: Important
          title: reviewer_models is accepted and validated but has no consumer, and no milestone owns the agentArgs wiring
          detail: reviewconfig.go loads ReviewerModels into Policy.Models and ReviewPlan copies it into Candidate.Model, but agentArgs (dispatch.go:219-258) never passes a model, so a configured model is silently ignored. The plan's Integration section says it is passed as --model/-m, but neither M1 nor M2 lists that work. Add it to M2, or warn while it is unwired.
          family: config-key-without-consumer
          round: 1
        - kind: defect
          id: BR-5
          severity: Important
          title: 'Plan Core concepts table contradicts the code: file paths, FamilyOf and ReviewPlan signatures, missing entities'
          detail: The table puts ReviewPlan in reviewplan.go and Unavailable in apifailure.go (both are in family.go), lists FamilyOf(agent, model) with qoder resolved through its model (the code is FamilyOf(agent) with no qoder handling), lists ReviewPlan(authors, kind, policy) (the code adds main and available), and omits LocalAvailability and AgentRun.Began. Add a Revisions entry.
          family: plan-table-matches-code
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: trimReason in family.go has no caller
          family: dead-code
          round: 1
        - kind: defect
          id: BR-7
          severity: Minor
          title: reviewer_models allows a leading '-', and reviewer_order allows duplicate agents
          detail: Reject a flag-like model value before M2 passes it to the CLI (ARCH-SECURE), and reject an agent listed twice in one order.
          family: parse-config-to-typed-value
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: f2f8d12a-9a73-4926-b20d-9022c59bf6bf
            repository: github.com/xianxu/ariadne
            issue_id: "000409"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/2d32b7370fc25891
            reviewed_head: 948d85dd7dce63dd71cc6514dbf37b8b6f5dcd21
            slot: ariadne:2
            requested_at: "2026-10-18T12:40:21.851335Z"
          phase: finished
          pid: 69034
          started_at: "2026-10-18T12:40:21.997522Z"
          finished_at: "2026-10-18T12:55:36.714862Z"
          session_id: 28e965e1-250f-4b52-8f6c-0784b1160796
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 948d85dd7dce63dd71cc6514dbf37b8b6f5dcd21
        tracker_head: 5578fe23834f9c2160c593a05397eaad33b10d1b
        branch_point: d3ebb55260c6647fbc70b423efed20cd8c734c5b
        since: "2026-10-18T05:16:36-07:00"
        until: "2026-10-18T05:40:03-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T05:16:36-07:00"
              end: "2026-10-18T05:40:03-07:00"
              began: "2026-10-18T05:16:36-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 144a7ee5-cc37-41c5-ac73-b88bbf5223ba
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5f3da10d43be251869bdafd6a98084e32ec627ca1a5a4b930a44fbe4f280c081
              started: "2026-10-18T12:25:28.29Z"
              spawned: false
              bytes: 482080
              windows:
                - start: "2026-10-18T12:25:28.29Z"
                  end: "2026-10-18T05:40:03-07:00"
            - harness: claude
              id: 5ca5160d-491c-4aec-b8f5-f5e40a69efb5
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 42052d978d0f33bca789480a4b74830893f6dd041d8d8d097390922c5c9823d6
              started: "2026-10-18T12:21:14.242Z"
              spawned: false
              bytes: 563079
              windows:
                - start: "2026-10-18T12:21:14.242Z"
                  end: "2026-10-18T05:40:03-07:00"
            - harness: claude
              id: 6be380fe-6c30-4408-b5b9-026237f91e30
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 70f6e935edb05d9fad671da7d73d9b88df89415544fb98457e596558f0fb3332
              started: "2026-10-18T12:23:49.894Z"
              spawned: false
              bytes: 497183
              windows:
                - start: "2026-10-18T12:23:49.894Z"
                  end: "2026-10-18T05:40:03-07:00"
            - harness: claude
              id: 946b723a-6614-4e42-8189-fff378170870
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 18bbe8b8ed1e1d030e367159b09a2e30f9b27a75f8dcc5f5048ff5d6d9bb34f1
              started: "2026-10-18T12:25:42.657Z"
              spawned: false
              bytes: 366511
              windows:
                - start: "2026-10-18T12:25:42.657Z"
                  end: "2026-10-18T05:40:03-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 8058a1b2ab702ea4b6a96fa1fe3037b4eaf10c3a9900648f23705c713e0af0ea
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 152158239
              windows:
                - start: "2026-10-18T05:16:36-07:00"
                  end: "2026-10-18T05:40:03-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-18T05:32:24-07:00"
              finalized: "2026-10-18T05:35:52-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-18T05:55:37-07:00"
      dispatched: "2026-10-18T05:40:21-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'Validate rejects empty, single-family and duplicate orders (agent_resolver.go:124-142), pinned by TestValidateOrdersAndModels; with Validate, Reviewer() cannot reach same-family. The default-order fallback in Reviewer() is unpinned: reverting it leaves the suite green.'
          round: 2
        - id: BR-2
          disposition: addressed
          note: Unavailable now classifies codex (ERROR lines, activity blocks) and gemini (empty stdout, non-zero exit); TestUnavailable has codex quota, bad-model, echoed-prompt and began cases and the gemini IneligibleTierError case; the plan Revision records the per-CLI evidence.
          round: 2
        - id: BR-3
          disposition: addressed
          note: close.md REVIEWER now describes the family-based other mode, reviewer_order and reviewer_models; atlas sdlc-binary.md:1191 no longer says claude and codex only.
          round: 2
        - id: BR-4
          disposition: not-addressed
          note: agentArgs passes the flags (pinned by TestAgentArgsPassTheModel), but deleting Model from milestoneclose.go:874 leaves go test ./cmd/sdlc green, so a configured model can again be silently ignored. Add a boundaryReviewOptionsFor test that asserts opts.Model.
          round: 2
        - id: BR-5
          disposition: addressed
          note: The M1 round-1 Revisions entry corrects the paths (family.go, stream.go, dispatch.go) and the signatures of FamilyOf, ReviewPlan and Unavailable, and adds LocalAvailability and AgentRun.Began; checked against the code.
          round: 2
        - id: BR-6
          disposition: addressed
          note: trimReason is gone; grep finds no definition or caller.
          round: 2
        - id: BR-7
          disposition: addressed
          note: Validate rejects a model starting with '-' and an agent listed twice; both are pinned in TestValidateOrdersAndModels.
          round: 2
      findings:
        - kind: defect
          id: BR-8
          severity: Minor
          title: Close help and atlas describe M2 behaviour (kind order, availability fallback) that the boundary review does not do yet
          detail: 'This is the 2nd finding in docs-track-behavior-change. Rule: help and atlas text must describe what the reviewed head does, and anything that lands later is marked with its milestone. Here close.md:415-421 says a full-flow close tries claude first and that gemini is the fallback, but boundaryReviewOptionsFor resolves through Reviewer(), which always uses KindBounded and ignores availability, so a gemini main agent on a full-flow close gets codex. Mark these as (M2) or land the kind with M2 and sweep both docs together.'
          family: docs-track-behavior-change
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: aa384400-3112-4490-bac5-df118069e10f
            repository: github.com/xianxu/ariadne
            issue_id: "000409"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/98ff97a217b8180e
            reviewed_head: 7753ca91157a2705e04fad8ec4e694835bd36089
            slot: ariadne:2
            requested_at: "2026-10-18T12:57:38.404728Z"
          phase: finished
          pid: 35806
          started_at: "2026-10-18T12:57:38.527913Z"
          finished_at: "2026-10-18T13:00:12.116973Z"
          session_id: c99ef9a6-6af1-47e0-9ed1-c994bf6656aa
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 7753ca91157a2705e04fad8ec4e694835bd36089
        tracker_head: 93fdc011a774cc68b1b0d0721e364a8c638dca8f
        branch_point: d3ebb55260c6647fbc70b423efed20cd8c734c5b
        since: "2026-10-18T05:16:36-07:00"
        until: "2026-10-18T05:57:06-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T05:16:36-07:00"
              end: "2026-10-18T05:57:06-07:00"
              began: "2026-10-18T05:16:36-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 144a7ee5-cc37-41c5-ac73-b88bbf5223ba
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5f3da10d43be251869bdafd6a98084e32ec627ca1a5a4b930a44fbe4f280c081
              started: "2026-10-18T12:25:28.29Z"
              spawned: false
              bytes: 482080
              windows:
                - start: "2026-10-18T12:25:28.29Z"
                  end: "2026-10-18T05:57:06-07:00"
            - harness: claude
              id: 5ca5160d-491c-4aec-b8f5-f5e40a69efb5
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 42052d978d0f33bca789480a4b74830893f6dd041d8d8d097390922c5c9823d6
              started: "2026-10-18T12:21:14.242Z"
              spawned: false
              bytes: 563079
              windows:
                - start: "2026-10-18T12:21:14.242Z"
                  end: "2026-10-18T05:57:06-07:00"
            - harness: claude
              id: 6be380fe-6c30-4408-b5b9-026237f91e30
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 70f6e935edb05d9fad671da7d73d9b88df89415544fb98457e596558f0fb3332
              started: "2026-10-18T12:23:49.894Z"
              spawned: false
              bytes: 497183
              windows:
                - start: "2026-10-18T12:23:49.894Z"
                  end: "2026-10-18T05:57:06-07:00"
            - harness: claude
              id: 946b723a-6614-4e42-8189-fff378170870
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 18bbe8b8ed1e1d030e367159b09a2e30f9b27a75f8dcc5f5048ff5d6d9bb34f1
              started: "2026-10-18T12:25:42.657Z"
              spawned: false
              bytes: 366511
              windows:
                - start: "2026-10-18T12:25:42.657Z"
                  end: "2026-10-18T05:57:06-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 95ce8d58368232d5e9444b72635bedbcfec934bbd530a0f2202fcedc6079dc67
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 154538580
              windows:
                - start: "2026-10-18T05:16:36-07:00"
                  end: "2026-10-18T05:57:06-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-18T05:32:24-07:00"
              finalized: "2026-10-18T05:35:52-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-18T05:40:21-07:00"
              finalized: "2026-10-18T05:55:37-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-18T06:00:31-07:00"
      dispatched: "2026-10-18T05:57:38-07:00"
      agent: claude
      dispose:
        - id: BR-4
          disposition: addressed
          note: DispatchOptions.Model is set from policy.Models at milestoneclose.go:874 and agentArgs emits --model or -m; removing either one makes TestBoundaryReviewPassesTheConfiguredModel or TestAgentArgsPassTheModel fail.
          round: 3
        - id: BR-8
          disposition: addressed
          note: close.md:72-73 and the atlas paragraph now mark the kind order and fallback as M2, and say a boundary review uses the bounded order's first agent, which matches Reviewer() always using KindBounded.
          round: 3
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      reviewed: 7753ca91157a2705e04fad8ec4e694835bd36089
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a459a5a4-a640-42a5-a9f4-5960476d53db
            repository: github.com/xianxu/ariadne
            issue_id: "000409"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/583459365d7fc14b
            reviewed_head: b616aee0de69c94aa54dcbb4832c7e498e6a0ea3
            slot: ariadne:2
            requested_at: "2026-10-18T13:18:55.68487Z"
          phase: finished
          pid: 96620
          started_at: "2026-10-18T13:18:55.856823Z"
          finished_at: "2026-10-18T13:32:28.88951Z"
          session_id: 7b029d83-e311-4e05-a0cf-bf8ad53c264b
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: b616aee0de69c94aa54dcbb4832c7e498e6a0ea3
        tracker_head: 1a31cd21076641eae870767c5dec2d09a37559ca
        branch_point: d3ebb55260c6647fbc70b423efed20cd8c734c5b
        since: "2026-10-18T05:16:36-07:00"
        until: "2026-10-18T06:18:27-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T05:16:36-07:00"
              end: "2026-10-18T06:18:27-07:00"
              began: "2026-10-18T05:16:36-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 144a7ee5-cc37-41c5-ac73-b88bbf5223ba
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5f3da10d43be251869bdafd6a98084e32ec627ca1a5a4b930a44fbe4f280c081
              started: "2026-10-18T12:25:28.29Z"
              spawned: false
              bytes: 482080
              windows:
                - start: "2026-10-18T12:25:28.29Z"
                  end: "2026-10-18T06:18:27-07:00"
            - harness: claude
              id: 5ca5160d-491c-4aec-b8f5-f5e40a69efb5
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 42052d978d0f33bca789480a4b74830893f6dd041d8d8d097390922c5c9823d6
              started: "2026-10-18T12:21:14.242Z"
              spawned: false
              bytes: 563079
              windows:
                - start: "2026-10-18T12:21:14.242Z"
                  end: "2026-10-18T06:18:27-07:00"
            - harness: claude
              id: 6be380fe-6c30-4408-b5b9-026237f91e30
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 70f6e935edb05d9fad671da7d73d9b88df89415544fb98457e596558f0fb3332
              started: "2026-10-18T12:23:49.894Z"
              spawned: false
              bytes: 497183
              windows:
                - start: "2026-10-18T12:23:49.894Z"
                  end: "2026-10-18T06:18:27-07:00"
            - harness: claude
              id: 946b723a-6614-4e42-8189-fff378170870
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 18bbe8b8ed1e1d030e367159b09a2e30f9b27a75f8dcc5f5048ff5d6d9bb34f1
              started: "2026-10-18T12:25:42.657Z"
              spawned: false
              bytes: 366511
              windows:
                - start: "2026-10-18T12:25:42.657Z"
                  end: "2026-10-18T06:18:27-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 543e1a378de2981427f9d1421877c5c98dfb80357b617e5c54e8b3455be02d00
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 163389692
              windows:
                - start: "2026-10-18T05:16:36-07:00"
                  end: "2026-10-18T06:18:27-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-18T05:32:24-07:00"
              finalized: "2026-10-18T05:35:52-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-18T05:40:21-07:00"
              finalized: "2026-10-18T05:55:37-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-18T05:57:38-07:00"
              finalized: "2026-10-18T06:00:31-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-18T06:32:31-07:00"
      dispatched: "2026-10-18T06:18:55-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-9
          severity: Critical
          title: A reviewer whose first attempt ran but whose no-verdict retry is unavailable hands the review to another family
          detail: 'dispatchAttempts (internal/judge/dispatch.go:386-395) returns attempt 2''s UnavailableError unchanged (directly, or via retryFailure). DispatchChoice (reviewchoice.go:98-104) then treats it as a skip, discards attempt 1''s output and runs the next candidate. Reproduced with a probe test: codex call 1 gives a no-verdict review with exit 0, call 2 gives a 429 ERROR line with a non-zero exit; result called=[codex codex gemini], used=gemini, err=nil. This violates the plan''s state-table row "started ... no switch to another family" (ARCH-ORDER). Fix: on attempt 2, demote UnavailableError to a plain did-not-run error, and add this two-call fake as a regression test.'
          family: switch-only-before-review-began
          round: 4
        - kind: defect
          id: BR-10
          severity: Critical
          title: The plan's integration table and recording section describe entities and ledger fields the code does not have
          detail: 'This is the 2nd finding in family plan-table-matches-code. Rule: at every boundary close, each Core concepts row and each named persisted field is checked by grep against the head, and any mismatch goes into a Revisions entry in the same commit. Instances now: a dispatchReview loop in dispatch.go (actually judge.DispatchChoice and ResolveReview in internal/judge/reviewchoice.go, plus resolveReviewChoice in cmd/sdlc/reviewchoice.go); ledger fields family/model/author_families/degraded (actually one Round.Reviewer string).'
          family: plan-table-matches-code
          round: 4
        - kind: defect
          id: BR-11
          severity: Important
          title: 'An empty ReviewChoice becomes agent "" and so claude: a dry run shows a same-family command the real run would refuse'
          detail: 'This is the 2nd finding in family no-silent-same-family-review. Rule: a ReviewChoice with no candidates must never become a dispatchable zero value; First() returns (Candidate, bool), and every caller (boundaryReviewOptionsFor at milestoneclose.go:897, the changecode.go plan/estimate sites, judge.go) reports the NoReviewerError refusal instead. Today agentArgs maps "" to claude (dispatch.go:224), so close --dry-run prints a claude argv for a claude-authored window.'
          family: no-silent-same-family-review
          round: 4
        - kind: defect
          id: BR-12
          severity: Important
          title: No test pins that plan-quality, estimate-quality or sdlc judge choose their reviewer through the policy
          detail: Only the boundary review has a policy-driven test (reviewchoice_test.go, reviewconfig_test.go). If changecode.go:726/978 or judge.go:685 went back to ResolveAgentCLI, every test would still pass. Add one test per site under boundary_reviewer=other with a claude session, asserting codex is dispatched.
          family: each-wired-site-has-a-test
          round: 4
        - kind: defect
          id: BR-13
          severity: Important
          title: The atlas still says every judge dispatch resolves through ResolveAgentCLI, and that the ledger's agent field records the reviewer
          detail: 'This is the 3rd finding in family docs-track-behavior-change. Rule: a change that retires or replaces a symbol greps atlas/ and helptext/ for that symbol in the same commit (ideally a test that every backticked judge.* identifier in atlas exists and has a non-test caller). Instances: atlas/workflow/sdlc-binary.md:1010-1012 and 1195-1197. Measured prevalence this round: 2 stale passages.'
          family: docs-track-behavior-change
          round: 4
        - kind: defect
          id: BR-14
          severity: Minor
          title: Fix reviews are always bounded, but the plan says they inherit their close's kind
          detail: boundaryChangeKind (cmd/sdlc/reviewchoice.go:97) maps FixReview to KindBounded even for a full-flow close. Either inherit the kind or record the deviation in Revisions.
          family: plan-decision-matches-code
          round: 4
        - kind: defect
          id: BR-15
          severity: Minor
          title: ResolveAgentCLI, ResolveReviewerCLI, TallyReviewers and ParseReviewerRecord have no production caller
          detail: 'This is the 2nd finding in family dead-code. Rule: a migration that moves every caller off an exported function deletes it, or names its remaining consumer, in the same change. Tally helpers that only tests use should either back a command or be documented as test-only.'
          family: dead-code
          round: 4
        - kind: defect
          id: BR-16
          severity: Minor
          title: A launch-failure reason truncated to 100 characters puts local PATH text into the trailer and ledger and drops the cause
          detail: 'dispatch.go:242-245 builds the reason as "could not launch: dispatch X (owner bin ... PATH=...)", and oneLine cuts it to 100 characters. Put the underlying runErr first.'
          family: parse-config-to-typed-value
          round: 4
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: f70a3154-0e64-444f-a601-ff0a54cea014
            repository: github.com/xianxu/ariadne
            issue_id: "000409"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/7001651fbb040aee
            reviewed_head: fdd8bffbce40de361f3b14cb5491ce644efca54f
            slot: ariadne:2
            requested_at: "2026-10-18T13:39:11.416407Z"
          phase: finished
          pid: 14858
          started_at: "2026-10-18T13:39:11.615862Z"
          finished_at: "2026-10-18T14:02:14.836676Z"
          session_id: 292b62a9-9001-4c51-b895-95fd8f20e2b3
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: fdd8bffbce40de361f3b14cb5491ce644efca54f
        tracker_head: 1a31cd21076641eae870767c5dec2d09a37559ca
        branch_point: d3ebb55260c6647fbc70b423efed20cd8c734c5b
        since: "2026-10-18T05:16:36-07:00"
        until: "2026-10-18T06:38:38-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T05:16:36-07:00"
              end: "2026-10-18T06:38:38-07:00"
              began: "2026-10-18T05:16:36-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 144a7ee5-cc37-41c5-ac73-b88bbf5223ba
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5f3da10d43be251869bdafd6a98084e32ec627ca1a5a4b930a44fbe4f280c081
              started: "2026-10-18T12:25:28.29Z"
              spawned: false
              bytes: 482080
              windows:
                - start: "2026-10-18T12:25:28.29Z"
                  end: "2026-10-18T06:38:38-07:00"
            - harness: claude
              id: 5ca5160d-491c-4aec-b8f5-f5e40a69efb5
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 42052d978d0f33bca789480a4b74830893f6dd041d8d8d097390922c5c9823d6
              started: "2026-10-18T12:21:14.242Z"
              spawned: false
              bytes: 563079
              windows:
                - start: "2026-10-18T12:21:14.242Z"
                  end: "2026-10-18T06:38:38-07:00"
            - harness: claude
              id: 6be380fe-6c30-4408-b5b9-026237f91e30
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 70f6e935edb05d9fad671da7d73d9b88df89415544fb98457e596558f0fb3332
              started: "2026-10-18T12:23:49.894Z"
              spawned: false
              bytes: 497183
              windows:
                - start: "2026-10-18T12:23:49.894Z"
                  end: "2026-10-18T06:38:38-07:00"
            - harness: claude
              id: 946b723a-6614-4e42-8189-fff378170870
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 18bbe8b8ed1e1d030e367159b09a2e30f9b27a75f8dcc5f5048ff5d6d9bb34f1
              started: "2026-10-18T12:25:42.657Z"
              spawned: false
              bytes: 366511
              windows:
                - start: "2026-10-18T12:25:42.657Z"
                  end: "2026-10-18T06:38:38-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 23ac47880d79a7416b069767232f8cd29cd6334d3f92f320e794592ad623b119
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 167772871
              windows:
                - start: "2026-10-18T05:16:36-07:00"
                  end: "2026-10-18T06:38:38-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-18T05:32:24-07:00"
              finalized: "2026-10-18T05:35:52-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-18T05:40:21-07:00"
              finalized: "2026-10-18T05:55:37-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-18T05:57:38-07:00"
              finalized: "2026-10-18T06:00:31-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-18T06:18:55-07:00"
              finalized: "2026-10-18T06:32:31-07:00"
              finalizing: false
      "n": 5
      timestamp: "2026-10-18T07:02:18-07:00"
      dispatched: "2026-10-18T06:39:11-07:00"
      agent: claude
      dispose:
        - id: BR-9
          disposition: addressed
          note: dispatch.go:388-394 turns an unavailable attempt-2 failure into ErrReviewDidNotRun. With the fix reverted, TestUnavailableRetryNeverHandsOver fails with called [codex codex gemini].
          round: 5
        - id: BR-10
          disposition: addressed
          note: The 2026-10-18 Revisions as-built table supersedes the old tables. Every row checked at the head (family.go, reviewchoice.go, stream.go Began, ledger.go Round.Reviewer, milestoneclose.go reviewTrailers).
          round: 5
        - id: BR-11
          disposition: addressed
          note: First() returns (Candidate, bool) and all four production callers refuse with Refusal(). TestResolveReview pins the empty-choice result.
          round: 5
        - id: BR-12
          disposition: addressed
          note: The three tests in reviewsites_test.go fail when planningReviewChoice and runJudge resolve with an empty policy (checked by revert).
          round: 5
        - id: BR-13
          disposition: addressed
          note: The atlas at sdlc-binary.md:1012 and 1195 now names ResolveReview and the ledger's reviewer field. A grep of atlas/ and helptext/ finds no live mention of a retired symbol.
          round: 5
        - id: BR-14
          disposition: addressed
          note: The deviation (fix reviews are always bounded) is recorded in Revisions with its reason.
          round: 5
        - id: BR-15
          disposition: addressed
          note: ResolveAgentCLI and ResolveReviewerCLI are deleted. The parse and tally helpers moved into a test file and are documented as test-only.
          round: 5
        - id: BR-16
          disposition: not-addressed
          note: 'The code fix is right (Reason holds runErr, Detail holds PATH), but no test pins it: no _test.go mentions could not launch. Minor, non-blocking.'
          round: 5
      findings:
        - kind: defect
          id: BR-17
          severity: Minor
          title: reviewchoice_test.go in internal/judge is not gofmt-clean
          detail: gofmt -l lists it. There is no blank line between TestReviewerRecordRoundTrip and the ReviewerEntry comment, and a trailing blank line at end of file. Run gofmt -w.
          family: gofmt-clean-diff
          round: 5
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      reviewed: fdd8bffbce40de361f3b14cb5491ce644efca54f
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: cc4effe3-dcb7-4d0b-bec4-cca56ce00f0a
            repository: github.com/xianxu/ariadne
            issue_id: "000409"
            recipe: milestone-review
            round: 6
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/c05b825593fbc47e
            reviewed_head: e5dca284acc7c1d10d3b354130fac751304ceca2
            slot: ariadne:2
            requested_at: "2026-10-18T14:34:49.347125Z"
          phase: finished
          pid: 97088
          started_at: "2026-10-18T14:34:49.435409Z"
          finished_at: "2026-10-18T14:51:44.314908Z"
          session_id: 48f48ad3-5673-4dd3-aab1-722acf4a3624
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: e5dca284acc7c1d10d3b354130fac751304ceca2
        tracker_head: 1a31cd21076641eae870767c5dec2d09a37559ca
        branch_point: d3ebb55260c6647fbc70b423efed20cd8c734c5b
        since: "2026-10-18T05:16:36-07:00"
        until: "2026-10-18T07:03:22-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T05:16:36-07:00"
              end: "2026-10-18T07:03:22-07:00"
              began: "2026-10-18T05:16:36-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 144a7ee5-cc37-41c5-ac73-b88bbf5223ba
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5f3da10d43be251869bdafd6a98084e32ec627ca1a5a4b930a44fbe4f280c081
              started: "2026-10-18T12:25:28.29Z"
              spawned: false
              bytes: 482080
              windows:
                - start: "2026-10-18T12:25:28.29Z"
                  end: "2026-10-18T07:03:22-07:00"
            - harness: claude
              id: 5ca5160d-491c-4aec-b8f5-f5e40a69efb5
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 42052d978d0f33bca789480a4b74830893f6dd041d8d8d097390922c5c9823d6
              started: "2026-10-18T12:21:14.242Z"
              spawned: false
              bytes: 563079
              windows:
                - start: "2026-10-18T12:21:14.242Z"
                  end: "2026-10-18T07:03:22-07:00"
            - harness: claude
              id: 6be380fe-6c30-4408-b5b9-026237f91e30
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 70f6e935edb05d9fad671da7d73d9b88df89415544fb98457e596558f0fb3332
              started: "2026-10-18T12:23:49.894Z"
              spawned: false
              bytes: 497183
              windows:
                - start: "2026-10-18T12:23:49.894Z"
                  end: "2026-10-18T07:03:22-07:00"
            - harness: claude
              id: 946b723a-6614-4e42-8189-fff378170870
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 18bbe8b8ed1e1d030e367159b09a2e30f9b27a75f8dcc5f5048ff5d6d9bb34f1
              started: "2026-10-18T12:25:42.657Z"
              spawned: false
              bytes: 366511
              windows:
                - start: "2026-10-18T12:25:42.657Z"
                  end: "2026-10-18T07:03:22-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 781d274e610a2b5bf56ed5a3f832125a7126f412512648422173b0564e45fcf7
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 169885287
              windows:
                - start: "2026-10-18T05:16:36-07:00"
                  end: "2026-10-18T07:03:22-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-18T05:32:24-07:00"
              finalized: "2026-10-18T05:35:52-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-18T05:40:21-07:00"
              finalized: "2026-10-18T05:55:37-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-18T05:57:38-07:00"
              finalized: "2026-10-18T06:00:31-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-18T06:18:55-07:00"
              finalized: "2026-10-18T06:32:31-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-18T06:39:11-07:00"
              finalized: "2026-10-18T07:02:18-07:00"
              finalizing: true
      "n": 6
      timestamp: "2026-10-18T07:51:45-07:00"
      dispatched: "2026-10-18T07:34:49-07:00"
      agent: claude
      dispose:
        - id: BR-16
          disposition: not-addressed
          note: Code is correct at dispatch.go:667 (Reason holds runErr, Detail holds PATH) but still no test asserts that Reason excludes PATH and leads with the cause.
          round: 6
        - id: BR-17
          disposition: addressed
          note: gofmt -l cmd/ at e5dca284 no longer lists internal/judge/reviewchoice_test.go.
          round: 6
      findings:
        - id: BR-18
          severity: Important
          title: 'reviewAuthorFamilies, the Spec''s author-family input from Agent: trailers, has no test'
          detail: 'cmd/sdlc/reviewchoice.go:23-41 turns the window''s first-parent Agent: trailers into families, but every site test runs a trailer-less window and only the pure ReviewPlan is tested with hand-made author sets. A regression in agentid parsing or the cut at the slash would silently send codex-authored work to codex. Add a temp-repo table test (codex trailer with a model suffix, merge commit ignored, unknown agent), or extract a pure log-text to families function and test that.'
          family: io-input-path-tested
          round: 6
        - id: BR-19
          severity: Minor
          title: A failed git log or a missing origin/main silently drops the trailer authors from the exclusion
          detail: 'This is the 3rd finding in family no-silent-same-family-review. reviewchoice.go:29-31 returns only the session family when git log fails (confirmed with a bad base, which gave anthropic only), and branchWindowBase returns empty without origin/main, so a codex-authored change can be reviewed by codex and the trailer records author anthropic. Rule for the class: any input the cross-family exclusion depends on (policy, order, authors) that cannot be read must refuse or be recorded as degraded in ReviewerRecord, never silently narrow the excluded families.'
          family: no-silent-same-family-review
          round: 6
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#409 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T05:35:52-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `no-silent-same-family-review` (kind: defect) reviewer_order lacking another family makes Reviewer() silently return the author's own agent
  Mode other with reviewer_order {"bounded":["claude"]} (or an empty list) and a claude main agent makes ReviewPlan return an empty plan, and Reviewer() at agent_resolver.go:85-91 falls through to `return main`, so claude reviews claude with no notice. Validate (agent_resolver.go:120-129) accepts the config. Reject empty, duplicate or single-family orders in Validate, and never return main under other. ARCH-PURPOSE.
- **BR-2** [Important] `unavailable-evidence-per-cli` (kind: defect) Unavailable can never classify a codex or gemini quota or auth failure; the fallback is inert for the default cross-family reviewer
  dispatch.go:629-633 runs ReadStream only for claude, so codex and gemini runs have Stream=false and Unavailable (family.go:131-134) returns false except on a launch error. M1 promised tests for the quota and auth texts each CLI prints; TestUnavailable has only claude stream-json fixtures. Parse codex exec --json (and gemini's output) into Began, or state and test the limit in a plan Revision. ARCH-PURPOSE, ARCH-ORDER.
- **BR-3** [Important] `docs-track-behavior-change` (kind: defect) close help still says gemini is reviewed by claude; M1 changed it to codex, and the new review.json keys are missing
  cmd/sdlc/helptext/close.md:417 describes the old `other` mapping. The changed judge_test.go:1336 now expects codex. The REVIEWER section omits reviewer_order and reviewer_models, and atlas sdlc-binary.md:1191 still describes other as claude and codex only. A cheap fix, and it should not wait for M2.
- **BR-4** [Important] `config-key-without-consumer` (kind: defect) reviewer_models is accepted and validated but has no consumer, and no milestone owns the agentArgs wiring
  reviewconfig.go loads ReviewerModels into Policy.Models and ReviewPlan copies it into Candidate.Model, but agentArgs (dispatch.go:219-258) never passes a model, so a configured model is silently ignored. The plan's Integration section says it is passed as --model/-m, but neither M1 nor M2 lists that work. Add it to M2, or warn while it is unwired.
- **BR-5** [Important] `plan-table-matches-code` (kind: defect) Plan Core concepts table contradicts the code: file paths, FamilyOf and ReviewPlan signatures, missing entities
  The table puts ReviewPlan in reviewplan.go and Unavailable in apifailure.go (both are in family.go), lists FamilyOf(agent, model) with qoder resolved through its model (the code is FamilyOf(agent) with no qoder handling), lists ReviewPlan(authors, kind, policy) (the code adds main and available), and omits LocalAvailability and AgentRun.Began. Add a Revisions entry.
- **BR-6** [Minor] `dead-code` (kind: defect) trimReason in family.go has no caller
- **BR-7** [Minor] `parse-config-to-typed-value` (kind: defect) reviewer_models allows a leading '-', and reviewer_order allows duplicate agents
  Reject a flag-like model value before M2 passes it to the CLI (ARCH-SECURE), and reject an agent listed twice in one order.

## Round 2 — 2026-10-18T05:55:37-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Validate rejects empty, single-family and duplicate orders (agent_resolver.go:124-142), pinned by TestValidateOrdersAndModels; with Validate, Reviewer() cannot reach same-family. The default-order fallback in Reviewer() is unpinned: reverting it leaves the suite green.
- BR-2 — addressed — Unavailable now classifies codex (ERROR lines, activity blocks) and gemini (empty stdout, non-zero exit); TestUnavailable has codex quota, bad-model, echoed-prompt and began cases and the gemini IneligibleTierError case; the plan Revision records the per-CLI evidence.
- BR-3 — addressed — close.md REVIEWER now describes the family-based other mode, reviewer_order and reviewer_models; atlas sdlc-binary.md:1191 no longer says claude and codex only.
- BR-4 — not-addressed — agentArgs passes the flags (pinned by TestAgentArgsPassTheModel), but deleting Model from milestoneclose.go:874 leaves go test ./cmd/sdlc green, so a configured model can again be silently ignored. Add a boundaryReviewOptionsFor test that asserts opts.Model.
- BR-5 — addressed — The M1 round-1 Revisions entry corrects the paths (family.go, stream.go, dispatch.go) and the signatures of FamilyOf, ReviewPlan and Unavailable, and adds LocalAvailability and AgentRun.Began; checked against the code.
- BR-6 — addressed — trimReason is gone; grep finds no definition or caller.
- BR-7 — addressed — Validate rejects a model starting with '-' and an agent listed twice; both are pinned in TestValidateOrdersAndModels.

### Raised

- **BR-8** [Minor] `docs-track-behavior-change` (kind: defect) Close help and atlas describe M2 behaviour (kind order, availability fallback) that the boundary review does not do yet
  This is the 2nd finding in docs-track-behavior-change. Rule: help and atlas text must describe what the reviewed head does, and anything that lands later is marked with its milestone. Here close.md:415-421 says a full-flow close tries claude first and that gemini is the fallback, but boundaryReviewOptionsFor resolves through Reviewer(), which always uses KindBounded and ignores availability, so a gemini main agent on a full-flow close gets codex. Mark these as (M2) or land the kind with M2 and sweep both docs together.

## Round 3 — 2026-10-18T06:00:31-07:00 (claude) — passed

### Disposed

- BR-4 — addressed — DispatchOptions.Model is set from policy.Models at milestoneclose.go:874 and agentArgs emits --model or -m; removing either one makes TestBoundaryReviewPassesTheConfiguredModel or TestAgentArgsPassTheModel fail.
- BR-8 — addressed — close.md:72-73 and the atlas paragraph now mark the kind order and fallback as M2, and say a boundary review uses the bounded order's first agent, which matches Reviewer() always using KindBounded.

## Round 4 — 2026-10-18T06:32:31-07:00 (claude) — BLOCKED

### Raised

- **BR-9** [Critical] `switch-only-before-review-began` (kind: defect) A reviewer whose first attempt ran but whose no-verdict retry is unavailable hands the review to another family
  dispatchAttempts (internal/judge/dispatch.go:386-395) returns attempt 2's UnavailableError unchanged (directly, or via retryFailure). DispatchChoice (reviewchoice.go:98-104) then treats it as a skip, discards attempt 1's output and runs the next candidate. Reproduced with a probe test: codex call 1 gives a no-verdict review with exit 0, call 2 gives a 429 ERROR line with a non-zero exit; result called=[codex codex gemini], used=gemini, err=nil. This violates the plan's state-table row "started ... no switch to another family" (ARCH-ORDER). Fix: on attempt 2, demote UnavailableError to a plain did-not-run error, and add this two-call fake as a regression test.
- **BR-10** [Critical] `plan-table-matches-code` (kind: defect) The plan's integration table and recording section describe entities and ledger fields the code does not have
  This is the 2nd finding in family plan-table-matches-code. Rule: at every boundary close, each Core concepts row and each named persisted field is checked by grep against the head, and any mismatch goes into a Revisions entry in the same commit. Instances now: a dispatchReview loop in dispatch.go (actually judge.DispatchChoice and ResolveReview in internal/judge/reviewchoice.go, plus resolveReviewChoice in cmd/sdlc/reviewchoice.go); ledger fields family/model/author_families/degraded (actually one Round.Reviewer string).
- **BR-11** [Important] `no-silent-same-family-review` (kind: defect) An empty ReviewChoice becomes agent "" and so claude: a dry run shows a same-family command the real run would refuse
  This is the 2nd finding in family no-silent-same-family-review. Rule: a ReviewChoice with no candidates must never become a dispatchable zero value; First() returns (Candidate, bool), and every caller (boundaryReviewOptionsFor at milestoneclose.go:897, the changecode.go plan/estimate sites, judge.go) reports the NoReviewerError refusal instead. Today agentArgs maps "" to claude (dispatch.go:224), so close --dry-run prints a claude argv for a claude-authored window.
- **BR-12** [Important] `each-wired-site-has-a-test` (kind: defect) No test pins that plan-quality, estimate-quality or sdlc judge choose their reviewer through the policy
  Only the boundary review has a policy-driven test (reviewchoice_test.go, reviewconfig_test.go). If changecode.go:726/978 or judge.go:685 went back to ResolveAgentCLI, every test would still pass. Add one test per site under boundary_reviewer=other with a claude session, asserting codex is dispatched.
- **BR-13** [Important] `docs-track-behavior-change` (kind: defect) The atlas still says every judge dispatch resolves through ResolveAgentCLI, and that the ledger's agent field records the reviewer
  This is the 3rd finding in family docs-track-behavior-change. Rule: a change that retires or replaces a symbol greps atlas/ and helptext/ for that symbol in the same commit (ideally a test that every backticked judge.* identifier in atlas exists and has a non-test caller). Instances: atlas/workflow/sdlc-binary.md:1010-1012 and 1195-1197. Measured prevalence this round: 2 stale passages.
- **BR-14** [Minor] `plan-decision-matches-code` (kind: defect) Fix reviews are always bounded, but the plan says they inherit their close's kind
  boundaryChangeKind (cmd/sdlc/reviewchoice.go:97) maps FixReview to KindBounded even for a full-flow close. Either inherit the kind or record the deviation in Revisions.
- **BR-15** [Minor] `dead-code` (kind: defect) ResolveAgentCLI, ResolveReviewerCLI, TallyReviewers and ParseReviewerRecord have no production caller
  This is the 2nd finding in family dead-code. Rule: a migration that moves every caller off an exported function deletes it, or names its remaining consumer, in the same change. Tally helpers that only tests use should either back a command or be documented as test-only.
- **BR-16** [Minor] `parse-config-to-typed-value` (kind: defect) A launch-failure reason truncated to 100 characters puts local PATH text into the trailer and ledger and drops the cause
  dispatch.go:242-245 builds the reason as "could not launch: dispatch X (owner bin ... PATH=...)", and oneLine cuts it to 100 characters. Put the underlying runErr first.

## Round 5 — 2026-10-18T07:02:18-07:00 (claude) — passed

### Disposed

- BR-9 — addressed — dispatch.go:388-394 turns an unavailable attempt-2 failure into ErrReviewDidNotRun. With the fix reverted, TestUnavailableRetryNeverHandsOver fails with called [codex codex gemini].
- BR-10 — addressed — The 2026-10-18 Revisions as-built table supersedes the old tables. Every row checked at the head (family.go, reviewchoice.go, stream.go Began, ledger.go Round.Reviewer, milestoneclose.go reviewTrailers).
- BR-11 — addressed — First() returns (Candidate, bool) and all four production callers refuse with Refusal(). TestResolveReview pins the empty-choice result.
- BR-12 — addressed — The three tests in reviewsites_test.go fail when planningReviewChoice and runJudge resolve with an empty policy (checked by revert).
- BR-13 — addressed — The atlas at sdlc-binary.md:1012 and 1195 now names ResolveReview and the ledger's reviewer field. A grep of atlas/ and helptext/ finds no live mention of a retired symbol.
- BR-14 — addressed — The deviation (fix reviews are always bounded) is recorded in Revisions with its reason.
- BR-15 — addressed — ResolveAgentCLI and ResolveReviewerCLI are deleted. The parse and tally helpers moved into a test file and are documented as test-only.
- BR-16 — not-addressed — The code fix is right (Reason holds runErr, Detail holds PATH), but no test pins it: no _test.go mentions could not launch. Minor, non-blocking.

### Raised

- **BR-17** [Minor] `gofmt-clean-diff` (kind: defect) reviewchoice_test.go in internal/judge is not gofmt-clean
  gofmt -l lists it. There is no blank line between TestReviewerRecordRoundTrip and the ReviewerEntry comment, and a trailing blank line at end of file. Run gofmt -w.

## Round 6 — 2026-10-18T07:51:45-07:00 (claude) — BLOCKED

### Disposed

- BR-16 — not-addressed — Code is correct at dispatch.go:667 (Reason holds runErr, Detail holds PATH) but still no test asserts that Reason excludes PATH and leads with the cause.
- BR-17 — addressed — gofmt -l cmd/ at e5dca284 no longer lists internal/judge/reviewchoice_test.go.

### Raised

- **BR-18** [Important] `io-input-path-tested` reviewAuthorFamilies, the Spec's author-family input from Agent: trailers, has no test
  cmd/sdlc/reviewchoice.go:23-41 turns the window's first-parent Agent: trailers into families, but every site test runs a trailer-less window and only the pure ReviewPlan is tested with hand-made author sets. A regression in agentid parsing or the cut at the slash would silently send codex-authored work to codex. Add a temp-repo table test (codex trailer with a model suffix, merge commit ignored, unknown agent), or extract a pure log-text to families function and test that.
- **BR-19** [Minor] `no-silent-same-family-review` A failed git log or a missing origin/main silently drops the trailer authors from the exclusion
  This is the 3rd finding in family no-silent-same-family-review. reviewchoice.go:29-31 returns only the session family when git log fails (confirmed with a bad base, which gave anthropic only), and branchWindowBase returns empty without origin/main, so a codex-authored change can be reviewed by codex and the trailer records author anthropic. Rule for the class: any input the cross-family exclusion depends on (policy, order, authors) that cannot be read must refuse or be recorded as degraded in ReviewerRecord, never silently narrow the excluded families.

## Open findings

- **BR-16** [Minor] `parse-config-to-typed-value` (kind: defect) A launch-failure reason truncated to 100 characters puts local PATH text into the trailer and ledger and drops the cause
- **BR-18** [Important] `io-input-path-tested` reviewAuthorFamilies, the Spec's author-family input from Agent: trailers, has no test
- **BR-19** [Minor] `no-silent-same-family-review` A failed git log or a missing origin/main silently drops the trailer authors from the exclusion
