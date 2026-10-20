---
gate: boundary-review
issue: 466
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a4eb394f-4bbe-4df5-a635-1e1749cdb8ed
            repository: github.com/xianxu/ariadne
            issue_id: "000466"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/7bf7d43ca245c258
            reviewed_head: 3e1c744a6513c1278812620c0ca5def6db84e2e6
            slot: ariadne:2
            requested_at: "2026-10-19T20:49:43.607381Z"
          phase: finished
          pid: 54877
          started_at: "2026-10-19T20:49:43.707778Z"
          finished_at: "2026-10-19T20:54:18.935532Z"
          session_id: f3c633d4-10fb-4739-aa9b-d9edbaa122b3
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 3e1c744a6513c1278812620c0ca5def6db84e2e6
        tracker_head: 53a3c8abef6381fb8694f6f57e34a8e3ae95f798
        branch_point: a0e8766188c855529ac82a38d66dcf30085f47b4
        since: "2026-10-19T10:37:08-07:00"
        until: "2026-10-19T11:10:41-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T10:37:08-07:00"
              end: "2026-10-19T11:10:41-07:00"
              began: "2026-10-19T10:37:08-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 476c4956729113c873a15ba75368d4fb0a0d4ed4cc64f627cae217b4afe79197
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 335426564
              windows:
                - start: "2026-10-19T10:37:08-07:00"
                  end: "2026-10-19T11:10:41-07:00"
      "n": 1
      timestamp: "2026-10-19T13:54:19-07:00"
      dispatched: "2026-10-19T13:49:43-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Unexplained (success guard) and flake list use exact top-level Lookup, so a registered subtest never explains a failure
          detail: 'evidence.go:74 and flake_cmd.go:610 call Registry.Lookup on records'' top-level failures. A probe confirmed that an Unknown attempt failing pkg.TestA stays unexplained once pkg.TestA/sub is registered, so guardSuitePass keeps refusing and its "register then republish" hint fails; flake list shows the case as not registered. Fix: judge each failure by Leaves plus Registry.Cover, using the attempt log''s subtests through the existing flakeDeps.subtests seam and failing closed when the log is unreadable; add tests.'
          family: coverage-consumers-must-derive
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Attempt records' known_flakes may now hold subtest names, but the plan says records are unchanged; README and help omit the older-reader refusal
          detail: teststats_store.go:172 validates known_flakes[].test with ValidTest, so a pre-466 reader refuses such a v2 record, and guardSuitePass refuses on any store warning, which blocks every suite-success post on that store. Correct the plan's Persisted formats section and add this hazard to the README rollout bullet and the register help.
          family: persisted-format-change-declared
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: No test proves concurrent additions survive unregister/narrow or that a repeat writes nothing
          detail: The flakeWorld.editTests fake applies edits with no race, and TestFlakeEditOnTrackerCards only checks that a repeated narrow returns no error. Add a racing-addition case on real tracker cards and assert ErrNoChange (no new commit) on a repeat, as the catalog proof claims.
          family: done-when-evidence-pinned
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: RerunMissing does not also require the top-level test's PASS line, as the plan says
          detail: Exit 0 already implies the parent passed, so this is plan drift only; fix the plan text or pass RerunTests to it too.
          family: plan-code-drift
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: MaxRerun counts names, not top-level tests rerun; PersistentMain matches at top-level granularity
          family: subtest-granularity-labels
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: unregister on a card without flake_tests writes an empty field instead of no-op
          family: idempotent-setter-no-write
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 106b5287-e10e-47ac-9b79-5281985be73a
            repository: github.com/xianxu/ariadne
            issue_id: "000466"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/15ac6d4d01c68eb4
            reviewed_head: e7ed662d6420d15d58b13c317224c8e42404d136
            slot: ariadne:2
            requested_at: "2026-10-19T21:05:08.020127Z"
          phase: finished
          pid: 25814
          started_at: "2026-10-19T21:05:08.200323Z"
          finished_at: "2026-10-19T21:19:56.812052Z"
          session_id: c30f1b78-e776-44ba-90e8-2a82e05ae25b
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: e7ed662d6420d15d58b13c317224c8e42404d136
        tracker_head: b0564b2d5c96b4f5e3bc42f42477d9a8a4828eba
        branch_point: a0e8766188c855529ac82a38d66dcf30085f47b4
        since: "2026-10-19T10:37:08-07:00"
        until: "2026-10-19T14:04:43-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T10:37:08-07:00"
              end: "2026-10-19T14:04:43-07:00"
              began: "2026-10-19T10:37:08-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 671fe3e205e72fc76083a12dbbada436c246de57374c700341007574b42a33da
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 337927910
              windows:
                - start: "2026-10-19T10:37:08-07:00"
                  end: "2026-10-19T14:04:43-07:00"
        queue_waits:
            - attempt_id: fe5fa2d62500c79bcedef05f0b82516e
              slot: ariadne:2
              record_digest: d9143902a070d8f22cd30019237a1b5f75858e7b276da98c3979fcccc93cb169
              start: "2026-10-19T18:29:25.729801Z"
              end: "2026-10-19T18:29:43.012858Z"
              queue_seconds: 17.283057
              raw_seconds: 17.283057
              exact: true
            - attempt_id: bdbbb4f266b2483c65c45a66764397f0
              slot: ariadne:2
              record_digest: ec1833675c452f4d3852a101df2f72d00c43b29fd4b51e1867209b71945f57d4
              start: "2026-10-19T18:50:11.822885Z"
              end: "2026-10-19T18:56:21.506993Z"
              queue_seconds: 369.684108
              raw_seconds: 369.684108
              exact: true
            - attempt_id: c8f19cac55e14f9807fd76dbeb5141f9
              slot: ariadne:2
              record_digest: 2dfbcaa77aa1fca3c97b1e9cf077dcae97c7b727befac03d730f7869aa1cd751
              start: "2026-10-19T19:15:04.604133Z"
              end: "2026-10-19T20:38:19.569376Z"
              queue_seconds: 4994.965243
              raw_seconds: 4994.965243
              exact: true
        protected:
            - issue: "469"
              slot: ariadne:2
              start: "2026-10-19T18:26:47.378Z"
              end: "2026-10-19T18:29:25.729801Z"
            - issue: "469"
              slot: ariadne:2
              start: "2026-10-19T18:29:43.012858Z"
              end: "2026-10-19T11:43:08-07:00"
            - issue: "469"
              slot: ariadne:2
              start: "2026-10-19T11:43:08-07:00"
              end: "2026-10-19T18:44:40.01Z"
        cuts:
            - dispatched: "2026-10-19T13:49:43-07:00"
              finalized: "2026-10-19T13:54:19-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-19T14:19:57-07:00"
      dispatched: "2026-10-19T14:05:08-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: not-addressed
          note: Code fix correct, but reverting withLogSubtests at localcheck_republish.go:165 and flake_cmd.go:597 keeps all targeted cmd/sdlc tests green; add guardSuitePass and runFlakeList regression tests that fail without the wiring.
          round: 2
        - id: BR-2
          disposition: addressed
          note: README rollout bullet, flake help (flake_cmd.go:298-299), atlas sdlc-binary.md Rollout line and plan Revisions now state the older-reader known_flakes refusal.
          round: 2
        - id: BR-3
          disposition: addressed
          note: TestFlakeEditKeepsAConcurrentRegistration on real tracker cards asserts "changed concurrently" on the raced edit, both names after rerun, and an unchanged issue-tracker tip on repeats; passes.
          round: 2
        - id: BR-4
          disposition: not-addressed
          note: Plan still says RerunMissing requires the top-level PASS line too; code checks entries only.
          round: 2
        - id: BR-5
          disposition: not-addressed
          note: Unchanged; MaxRerun still counts known entries (names).
          round: 2
        - id: BR-6
          disposition: not-addressed
          note: Unchanged; editFlakeTests still calls SetCardField(..., "") when the field is absent.
          round: 2
      findings:
        - kind: defect
          id: BR-7
          severity: Minor
          title: guardSuitePass refusal hint says --test <pkg.TestName>, omitting the subtest form
          detail: This is the 2nd finding in family subtest-granularity-labels. The rule is that every user-facing text naming a registerable test uses the grammar PKG.TestName[/sub…]. A sweep of cmd/sdlc for pkg.TestName> finds only localcheck_republish.go:190.
          family: subtest-granularity-labels
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: ccaf6a69-91b3-4f4e-a649-972298649eab
            repository: github.com/xianxu/ariadne
            issue_id: "000466"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/d7abc595aeb1da9b
            reviewed_head: ee4d22050707b73f6c81b9a365082237d3ac687e
            slot: ariadne:2
            requested_at: "2026-10-19T21:23:12.296918Z"
          phase: finished
          pid: 93534
          started_at: "2026-10-19T21:23:12.584286Z"
          finished_at: "2026-10-19T21:25:39.345602Z"
          session_id: 87c09e13-6272-4816-9059-f31006c9a08b
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: ee4d22050707b73f6c81b9a365082237d3ac687e
        tracker_head: 563021372cbac20f5e44e3a593315c40b61b5d52
        branch_point: a0e8766188c855529ac82a38d66dcf30085f47b4
        since: "2026-10-19T10:37:08-07:00"
        until: "2026-10-19T14:22:09-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T10:37:08-07:00"
              end: "2026-10-19T14:22:09-07:00"
              began: "2026-10-19T10:37:08-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: e1abd5cac3a693296987b601712fab5746507ffee9b5c76f01a7b3a932795657
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 340805289
              windows:
                - start: "2026-10-19T10:37:08-07:00"
                  end: "2026-10-19T14:22:09-07:00"
        queue_waits:
            - attempt_id: fe5fa2d62500c79bcedef05f0b82516e
              slot: ariadne:2
              record_digest: d9143902a070d8f22cd30019237a1b5f75858e7b276da98c3979fcccc93cb169
              start: "2026-10-19T18:29:25.729801Z"
              end: "2026-10-19T18:29:43.012858Z"
              queue_seconds: 17.283057
              raw_seconds: 17.283057
              exact: true
            - attempt_id: bdbbb4f266b2483c65c45a66764397f0
              slot: ariadne:2
              record_digest: ec1833675c452f4d3852a101df2f72d00c43b29fd4b51e1867209b71945f57d4
              start: "2026-10-19T18:50:11.822885Z"
              end: "2026-10-19T18:56:21.506993Z"
              queue_seconds: 369.684108
              raw_seconds: 369.684108
              exact: true
            - attempt_id: c8f19cac55e14f9807fd76dbeb5141f9
              slot: ariadne:2
              record_digest: 2dfbcaa77aa1fca3c97b1e9cf077dcae97c7b727befac03d730f7869aa1cd751
              start: "2026-10-19T19:15:04.604133Z"
              end: "2026-10-19T20:38:19.569376Z"
              queue_seconds: 4994.965243
              raw_seconds: 4994.965243
              exact: true
        protected:
            - issue: "469"
              slot: ariadne:2
              start: "2026-10-19T18:26:47.378Z"
              end: "2026-10-19T18:29:25.729801Z"
            - issue: "469"
              slot: ariadne:2
              start: "2026-10-19T18:29:43.012858Z"
              end: "2026-10-19T11:43:08-07:00"
            - issue: "469"
              slot: ariadne:2
              start: "2026-10-19T11:43:08-07:00"
              end: "2026-10-19T18:44:40.01Z"
        cuts:
            - dispatched: "2026-10-19T13:49:43-07:00"
              finalized: "2026-10-19T13:54:19-07:00"
              finalizing: false
            - dispatched: "2026-10-19T14:05:08-07:00"
              finalized: "2026-10-19T14:19:57-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-19T14:25:40-07:00"
      dispatched: "2026-10-19T14:23:12-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Guard and list both use withLogSubtests and flakes.Explained; reverting either wiring site makes TestSuitePassGuardReadsSubtestsFromTheLog / TestFlakeListJudgesSubtestsFromTheLog fail (verified locally).
          round: 3
        - id: BR-4
          disposition: not-addressed
          note: localcheck_flakes.go:104 still passes flakes.Tests(o.Known) only; plan text unchanged.
          round: 3
        - id: BR-5
          disposition: not-addressed
          note: classify.go:166 still compares len(known) covering names against MaxRerun.
          round: 3
        - id: BR-6
          disposition: not-addressed
          note: editFlakeTests sets flake_tests to "" via SetField, which appends the key to a card that lacked it.
          round: 3
        - id: BR-7
          disposition: addressed
          note: localcheck_republish.go:190 now reads PKG.TestName[/sub…]; flake help, register Use and flags agree; README and atlas follow the synopsis with explicit subtest text.
          round: 3
      recipe: milestone-review
      isolation: fenced
      reviewed: ee4d22050707b73f6c81b9a365082237d3ac687e
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: c835c65b-ba31-4a83-83f3-09912ab40028
            repository: github.com/xianxu/ariadne
            issue_id: "000466"
            recipe: fix-review
            round: 4
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/41feaf1b2beb96bc
            reviewed_head: c0bc9b48b215b10f936e147fb142d2421aaeab3b
            slot: ariadne:2
            requested_at: "2026-10-20T00:52:14.681863Z"
          phase: finished
          pid: 81535
          started_at: "2026-10-20T00:52:14.783161Z"
          finished_at: "2026-10-20T00:53:25.25073Z"
          session_id: af0abda6-fe5a-43a4-857d-8dc434d645c2
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: c0bc9b48b215b10f936e147fb142d2421aaeab3b
        tracker_head: c104d8072b0ba1e28b7b13a0d360055e190aff18
        branch_point: dfe6340c899a099185bf11e85f1e77b7940bf2bd
        since: "2026-10-19T10:37:08-07:00"
        until: "2026-10-19T14:25:47-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T10:37:08-07:00"
              end: "2026-10-19T14:25:47-07:00"
              began: "2026-10-19T10:37:08-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6bebdbe5e64554c781c7bcb75aff3ea39642423410279fc4c3c3a830712663f8
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 346077281
              windows:
                - start: "2026-10-19T10:37:08-07:00"
                  end: "2026-10-19T14:25:47-07:00"
        queue_waits:
            - attempt_id: fe5fa2d62500c79bcedef05f0b82516e
              slot: ariadne:2
              record_digest: d9143902a070d8f22cd30019237a1b5f75858e7b276da98c3979fcccc93cb169
              start: "2026-10-19T18:29:25.729801Z"
              end: "2026-10-19T18:29:43.012858Z"
              queue_seconds: 17.283057
              raw_seconds: 17.283057
              exact: true
            - attempt_id: bdbbb4f266b2483c65c45a66764397f0
              slot: ariadne:2
              record_digest: ec1833675c452f4d3852a101df2f72d00c43b29fd4b51e1867209b71945f57d4
              start: "2026-10-19T18:50:11.822885Z"
              end: "2026-10-19T18:56:21.506993Z"
              queue_seconds: 369.684108
              raw_seconds: 369.684108
              exact: true
            - attempt_id: c8f19cac55e14f9807fd76dbeb5141f9
              slot: ariadne:2
              record_digest: 2dfbcaa77aa1fca3c97b1e9cf077dcae97c7b727befac03d730f7869aa1cd751
              start: "2026-10-19T19:15:04.604133Z"
              end: "2026-10-19T20:38:19.569376Z"
              queue_seconds: 4994.965243
              raw_seconds: 4994.965243
              exact: true
        protected:
            - issue: "469"
              slot: ariadne:2
              start: "2026-10-19T18:26:47.378Z"
              end: "2026-10-19T18:29:25.729801Z"
            - issue: "469"
              slot: ariadne:2
              start: "2026-10-19T18:29:43.012858Z"
              end: "2026-10-19T11:43:08-07:00"
            - issue: "469"
              slot: ariadne:2
              start: "2026-10-19T11:43:08-07:00"
              end: "2026-10-19T18:44:40.01Z"
        cuts:
            - dispatched: "2026-10-19T13:49:43-07:00"
              finalized: "2026-10-19T13:54:19-07:00"
              finalizing: false
            - dispatched: "2026-10-19T14:05:08-07:00"
              finalized: "2026-10-19T14:19:57-07:00"
              finalizing: false
            - dispatched: "2026-10-19T14:23:12-07:00"
              finalized: "2026-10-19T14:25:40-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-19T17:53:25-07:00"
      dispatched: "2026-10-19T17:52:14-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-4
          disposition: not-addressed
          note: Interdiff is only the conflict resolution; plan text and RerunMissing call are unchanged.
          round: 4
        - id: BR-5
          disposition: not-addressed
          note: Unchanged in this window; MaxRerun still counts covering names.
          round: 4
        - id: BR-6
          disposition: not-addressed
          note: Unchanged in this window; editFlakeTests still writes an empty flake_tests field.
          round: 4
      findings:
        - kind: defect
          id: BR-8
          severity: Minor
          title: Merged fixed-issue filter (fixedIssueCovers) matches top-level names exactly, not via the shared subtest coverage rule
          detail: 'This is the 2nd finding in family coverage-consumers-must-derive. Rule: every consumer that asks "does a card''s flake_tests cover this failure" must call flakes.Covers/Explained (leaf or ancestor) rather than string equality; sweep all such matchers, of which flake_fixed.go:104 is the one left after the #458/#466 merge. It fails safe (a subtest-registered fixed issue never hides or labels a row as recurred), so this is Minor.'
          family: coverage-consumers-must-derive
          round: 4
      recipe: fix-review
      isolation: fenced
      reviewed: c0bc9b48b215b10f936e147fb142d2421aaeab3b
      blocked: false
---

# Gate ledger — ariadne#466 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T13:54:19-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `coverage-consumers-must-derive` (kind: defect) Unexplained (success guard) and flake list use exact top-level Lookup, so a registered subtest never explains a failure
  evidence.go:74 and flake_cmd.go:610 call Registry.Lookup on records' top-level failures. A probe confirmed that an Unknown attempt failing pkg.TestA stays unexplained once pkg.TestA/sub is registered, so guardSuitePass keeps refusing and its "register then republish" hint fails; flake list shows the case as not registered. Fix: judge each failure by Leaves plus Registry.Cover, using the attempt log's subtests through the existing flakeDeps.subtests seam and failing closed when the log is unreadable; add tests.
- **BR-2** [Important] `persisted-format-change-declared` (kind: defect) Attempt records' known_flakes may now hold subtest names, but the plan says records are unchanged; README and help omit the older-reader refusal
  teststats_store.go:172 validates known_flakes[].test with ValidTest, so a pre-466 reader refuses such a v2 record, and guardSuitePass refuses on any store warning, which blocks every suite-success post on that store. Correct the plan's Persisted formats section and add this hazard to the README rollout bullet and the register help.
- **BR-3** [Important] `done-when-evidence-pinned` (kind: defect) No test proves concurrent additions survive unregister/narrow or that a repeat writes nothing
  The flakeWorld.editTests fake applies edits with no race, and TestFlakeEditOnTrackerCards only checks that a repeated narrow returns no error. Add a racing-addition case on real tracker cards and assert ErrNoChange (no new commit) on a repeat, as the catalog proof claims.
- **BR-4** [Minor] `plan-code-drift` (kind: defect) RerunMissing does not also require the top-level test's PASS line, as the plan says
  Exit 0 already implies the parent passed, so this is plan drift only; fix the plan text or pass RerunTests to it too.
- **BR-5** [Minor] `subtest-granularity-labels` (kind: defect) MaxRerun counts names, not top-level tests rerun; PersistentMain matches at top-level granularity
- **BR-6** [Minor] `idempotent-setter-no-write` (kind: defect) unregister on a card without flake_tests writes an empty field instead of no-op

## Round 2 — 2026-10-19T14:19:57-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — not-addressed — Code fix correct, but reverting withLogSubtests at localcheck_republish.go:165 and flake_cmd.go:597 keeps all targeted cmd/sdlc tests green; add guardSuitePass and runFlakeList regression tests that fail without the wiring.
- BR-2 — addressed — README rollout bullet, flake help (flake_cmd.go:298-299), atlas sdlc-binary.md Rollout line and plan Revisions now state the older-reader known_flakes refusal.
- BR-3 — addressed — TestFlakeEditKeepsAConcurrentRegistration on real tracker cards asserts "changed concurrently" on the raced edit, both names after rerun, and an unchanged issue-tracker tip on repeats; passes.
- BR-4 — not-addressed — Plan still says RerunMissing requires the top-level PASS line too; code checks entries only.
- BR-5 — not-addressed — Unchanged; MaxRerun still counts known entries (names).
- BR-6 — not-addressed — Unchanged; editFlakeTests still calls SetCardField(..., "") when the field is absent.

### Raised

- **BR-7** [Minor] `subtest-granularity-labels` (kind: defect) guardSuitePass refusal hint says --test <pkg.TestName>, omitting the subtest form
  This is the 2nd finding in family subtest-granularity-labels. The rule is that every user-facing text naming a registerable test uses the grammar PKG.TestName[/sub…]. A sweep of cmd/sdlc for pkg.TestName> finds only localcheck_republish.go:190.

## Round 3 — 2026-10-19T14:25:40-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Guard and list both use withLogSubtests and flakes.Explained; reverting either wiring site makes TestSuitePassGuardReadsSubtestsFromTheLog / TestFlakeListJudgesSubtestsFromTheLog fail (verified locally).
- BR-4 — not-addressed — localcheck_flakes.go:104 still passes flakes.Tests(o.Known) only; plan text unchanged.
- BR-5 — not-addressed — classify.go:166 still compares len(known) covering names against MaxRerun.
- BR-6 — not-addressed — editFlakeTests sets flake_tests to "" via SetField, which appends the key to a card that lacked it.
- BR-7 — addressed — localcheck_republish.go:190 now reads PKG.TestName[/sub…]; flake help, register Use and flags agree; README and atlas follow the synopsis with explicit subtest text.

## Round 4 — 2026-10-19T17:53:25-07:00 (claude) — passed

### Disposed

- BR-4 — not-addressed — Interdiff is only the conflict resolution; plan text and RerunMissing call are unchanged.
- BR-5 — not-addressed — Unchanged in this window; MaxRerun still counts covering names.
- BR-6 — not-addressed — Unchanged in this window; editFlakeTests still writes an empty flake_tests field.

### Raised

- **BR-8** [Minor] `coverage-consumers-must-derive` (kind: defect) Merged fixed-issue filter (fixedIssueCovers) matches top-level names exactly, not via the shared subtest coverage rule
  This is the 2nd finding in family coverage-consumers-must-derive. Rule: every consumer that asks "does a card's flake_tests cover this failure" must call flakes.Covers/Explained (leaf or ancestor) rather than string equality; sweep all such matchers, of which flake_fixed.go:104 is the one left after the #458/#466 merge. It fails safe (a subtest-registered fixed issue never hides or labels a row as recurred), so this is Minor.

## Open findings

- **BR-4** [Minor] `plan-code-drift` (kind: defect) RerunMissing does not also require the top-level test's PASS line, as the plan says
- **BR-5** [Minor] `subtest-granularity-labels` (kind: defect) MaxRerun counts names, not top-level tests rerun; PersistentMain matches at top-level granularity
- **BR-6** [Minor] `idempotent-setter-no-write` (kind: defect) unregister on a card without flake_tests writes an empty field instead of no-op
- **BR-8** [Minor] `coverage-consumers-must-derive` (kind: defect) Merged fixed-issue filter (fixedIssueCovers) matches top-level names exactly, not via the shared subtest coverage rule
