---
gate: boundary-review
issue: 475
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: e23a1d03-611b-4c3a-8af2-99eb9b7564b9
            repository: github.com/xianxu/ariadne
            issue_id: "000475"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/5b6b7dbec9edb75f
            reviewed_head: b97bc6dd7d96fcd482b2fc57c6231b394cb0bf64
            slot: ariadne:4
            requested_at: "2026-10-20T11:40:01.533863Z"
          phase: finished
          pid: 39821
          started_at: "2026-10-20T11:40:01.627835Z"
          finished_at: "2026-10-20T11:51:26.839134Z"
          session_id: ce54466e-eca4-404e-8a65-78527918ab82
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: d75999f8bd8cab6da9b65500899c89f5364ae986
        tracker_head: d54ab3581632e28192fe40464fdd682d4fadf7e2
        branch_point: 7918d3d63c4c12e9a4537b51680e2143b29fd286
        since: "2026-10-20T03:03:17-07:00"
        until: "2026-10-20T04:38:21-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T03:03:17-07:00"
              end: "2026-10-20T04:38:21-07:00"
              began: "2026-10-20T03:03:17-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 3d08a2cde284a5ec5c5d1007a1ed6ba5714c4ff43cca95d0874bb8ea8ecb69df
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 96520513
              windows:
                - start: "2026-10-20T03:03:17-07:00"
                  end: "2026-10-20T04:38:21-07:00"
            - harness: claude
              id: 3fc56c78-9cad-4e12-a1a7-aed05bd0c63d
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: b676d6d103897ad92941db345e4e23c27533a35891f17b7b537450d5976ee8ad
              started: "2026-10-20T11:21:49.422Z"
              spawned: false
              bytes: 384376
              windows:
                - start: "2026-10-20T11:21:49.422Z"
                  end: "2026-10-20T04:38:21-07:00"
            - harness: claude
              id: 95466ede-cf18-4115-b13f-1faa8f33645d
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 1b281c55e873b03c51fd775f439ef68f3e2a49b0a3f6e6097df1fe069c81db79
              started: "2026-10-20T11:21:30.104Z"
              spawned: false
              bytes: 514596
              windows:
                - start: "2026-10-20T11:21:30.104Z"
                  end: "2026-10-20T04:38:21-07:00"
            - harness: claude
              id: ac351657-0a41-4804-b564-3e710ea2dd55
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 762c45ca71122a0c54e4c27a0ff8e23460c28233ba39f58b9420ecf8d981cda4
              started: "2026-10-20T11:19:02.044Z"
              spawned: false
              bytes: 587879
              windows:
                - start: "2026-10-20T11:19:02.044Z"
                  end: "2026-10-20T04:38:21-07:00"
            - harness: claude
              id: f5886db6-bd63-4d0e-aec8-021fea4f0187
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: ff3b2eef3104ceeab9229868a244b067cce02222561142ea5df5f3dc7ec0d977
              started: "2026-10-20T11:20:37.435Z"
              spawned: false
              bytes: 530611
              windows:
                - start: "2026-10-20T11:20:37.435Z"
                  end: "2026-10-20T04:38:21-07:00"
        queue_waits:
            - attempt_id: 3611f175d5a4d5f65847189d1c3190ac
              slot: ariadne:4
              record_digest: 0bfdc61fc34a7d4333b805a3bf32ca5a4d0d8dabb13eea6559cf546adf6609a2
              start: "2026-10-20T10:14:59.477681Z"
              end: "2026-10-20T10:30:07.515298Z"
              queue_seconds: 908.037617
              raw_seconds: 908.037617
              exact: true
            - attempt_id: b648e3ff6d96d6e5b355164d3dbee74c
              slot: ariadne:4
              record_digest: 11df85f04d90f5060d15b092a898573c63d1a6d63d3891bf06bff55a7853326d
              start: "2026-10-20T11:03:21.374039Z"
              end: "2026-10-20T11:04:29.093112Z"
              queue_seconds: 67.719073
              raw_seconds: 67.719073
              exact: true
        protected:
            - issue: "490"
              slot: ariadne:4
              start: "2026-10-20T10:05:34.25Z"
              end: "2026-10-20T10:11:07.378Z"
      "n": 1
      timestamp: "2026-10-20T04:51:27-07:00"
      dispatched: "2026-10-20T04:40:01-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: README departure rule is stale and omits train_max_wait_minutes
          detail: README.md:521-531 lists only train_interval_minutes/train_max_cars and describes departure as first train / one interval after the last / at once after left-ready. It omits the full-train and max-wait early departures, the new key, and that lone and granted landings no longer reset the clock. helptext/landing.md is equally stale.
          family: user-docs-track-new-surface
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: T5 is ticked, but landing status lacks the promised reasons and golden test
          detail: The plan promised train due (reason) / next train at HH:MM (slot or max-wait), including cars-left-ready, plus a golden test. The code prints a bare train due for slot and left-ready, and an RFC3339 time with no source. The only test change is a 1-ready substring (landingcmd_test.go:196). Add the reasons and a status test, or narrow T5 in Revisions.
          family: plan-claim-matches-delivery
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: Event.Departed has no reader; the slot check uses the stored NextDue, not the plan's lastDeparture
          detail: PQ-2 says lastDeparture reads Departed. The code compares a NextDue fixed at write time, so Departed is written but never read, and a later offset change only applies after the next train.
          family: plan-claim-matches-delivery
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: NextDeparture does not ignore old-format alone records by Result as the Ordering section claims
          detail: train.go:84 filters on non-zero NextDue only, so an alone record written before the upgrade (departed + interval) still delays the next train once.
          family: plan-claim-matches-delivery
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: The max-wait predicate is duplicated in DepartureDue
          detail: train.go:228 and :234 repeat the same three-part condition; one wait-deadline helper would do.
          family: shared-helper-not-repeated-predicate
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: An older sdlc re-vouching carries a stale since through Extra across a head change
          detail: mergeExtra keeps since as an unknown member for an old binary, so after a head change the new binary trusts the stale Since and may depart early on max-wait.
          family: unknown-field-carry-stale-semantics
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: c159b511-dfc9-4d80-91e3-f34150a94d55
            repository: github.com/xianxu/ariadne
            issue_id: "000475"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/30ef4d045dbeb182
            reviewed_head: fcc0be143bb399a3b0176477be2ed52ca16da8e0
            slot: ariadne:4
            requested_at: "2026-10-20T11:55:52.686773Z"
          phase: finished
          pid: 58606
          started_at: "2026-10-20T11:55:52.902227Z"
          finished_at: "2026-10-20T12:08:04.201777Z"
          session_id: dfa6f6eb-bdcf-4fe1-87d5-aa1bbe132208
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: fcc0be143bb399a3b0176477be2ed52ca16da8e0
        tracker_head: d54ab3581632e28192fe40464fdd682d4fadf7e2
        branch_point: 7918d3d63c4c12e9a4537b51680e2143b29fd286
        since: "2026-10-20T03:03:17-07:00"
        until: "2026-10-20T04:55:08-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T03:03:17-07:00"
              end: "2026-10-20T04:55:08-07:00"
              began: "2026-10-20T03:03:17-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 11d1cc2b288eedfc58db0d538b7bf9200b92bd8d2c485cff50fd5feed17411f1
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 96775005
              windows:
                - start: "2026-10-20T03:03:17-07:00"
                  end: "2026-10-20T04:55:08-07:00"
            - harness: claude
              id: 3fc56c78-9cad-4e12-a1a7-aed05bd0c63d
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: b676d6d103897ad92941db345e4e23c27533a35891f17b7b537450d5976ee8ad
              started: "2026-10-20T11:21:49.422Z"
              spawned: false
              bytes: 384376
              windows:
                - start: "2026-10-20T11:21:49.422Z"
                  end: "2026-10-20T04:55:08-07:00"
            - harness: claude
              id: 95466ede-cf18-4115-b13f-1faa8f33645d
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 1b281c55e873b03c51fd775f439ef68f3e2a49b0a3f6e6097df1fe069c81db79
              started: "2026-10-20T11:21:30.104Z"
              spawned: false
              bytes: 514596
              windows:
                - start: "2026-10-20T11:21:30.104Z"
                  end: "2026-10-20T04:55:08-07:00"
            - harness: claude
              id: ac351657-0a41-4804-b564-3e710ea2dd55
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 762c45ca71122a0c54e4c27a0ff8e23460c28233ba39f58b9420ecf8d981cda4
              started: "2026-10-20T11:19:02.044Z"
              spawned: false
              bytes: 587879
              windows:
                - start: "2026-10-20T11:19:02.044Z"
                  end: "2026-10-20T04:55:08-07:00"
            - harness: claude
              id: f5886db6-bd63-4d0e-aec8-021fea4f0187
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: ff3b2eef3104ceeab9229868a244b067cce02222561142ea5df5f3dc7ec0d977
              started: "2026-10-20T11:20:37.435Z"
              spawned: false
              bytes: 530611
              windows:
                - start: "2026-10-20T11:20:37.435Z"
                  end: "2026-10-20T04:55:08-07:00"
        queue_waits:
            - attempt_id: 3611f175d5a4d5f65847189d1c3190ac
              slot: ariadne:4
              record_digest: 0bfdc61fc34a7d4333b805a3bf32ca5a4d0d8dabb13eea6559cf546adf6609a2
              start: "2026-10-20T10:14:59.477681Z"
              end: "2026-10-20T10:30:07.515298Z"
              queue_seconds: 908.037617
              raw_seconds: 908.037617
              exact: true
            - attempt_id: b648e3ff6d96d6e5b355164d3dbee74c
              slot: ariadne:4
              record_digest: 11df85f04d90f5060d15b092a898573c63d1a6d63d3891bf06bff55a7853326d
              start: "2026-10-20T11:03:21.374039Z"
              end: "2026-10-20T11:04:29.093112Z"
              queue_seconds: 67.719073
              raw_seconds: 67.719073
              exact: true
        protected:
            - issue: "490"
              slot: ariadne:4
              start: "2026-10-20T10:05:34.25Z"
              end: "2026-10-20T10:11:07.378Z"
        cuts:
            - boundary: M1
              dispatched: "2026-10-20T04:40:01-07:00"
              finalized: "2026-10-20T04:51:27-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-20T05:08:05-07:00"
      dispatched: "2026-10-20T04:55:52-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: README.md:521-538 and helptext/landing.md now list train_max_wait_minutes, full and max-wait early departures, and that lone/granted landings never push the train back.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Status calls DepartureDue and prints the reason (slot, interval, max wait, full, cars left ready); landingcmd_test.go:852-862 pins "next train at 2026-10-20 10:00 (interval)".
          round: 2
        - id: BR-3
          disposition: addressed
          note: nextDeparture works out the slot from departedOf(e) when it decides (train.go:373); TestNextDepartureReadsTheDeparture covers it.
          round: 2
        - id: BR-4
          disposition: addressed
          note: The alone case is skipped by Result (train.go:370); reverting it makes TestNextDepartureReadsTheDeparture fail (checked in the checkout).
          round: 2
        - id: BR-5
          disposition: addressed
          note: waitDeadline helper (train.go:330) used for both the due check and the reported next time.
          round: 2
        - id: BR-6
          disposition: addressed
          note: SinceHead guard in firstReady (train.go:449); reverting it makes TestStaleSinceFromAnOlderSdlc fail (checked). A sibling case is raised separately.
          round: 2
      findings:
        - kind: defect
          id: BR-7
          severity: Minor
          title: An older sdlc's not-ready then ready re-vouch at the same head still carries a stale since
          detail: 'This is the 2nd finding in this family. The rule: a derived field that an older writer can carry through Extra must be stamped against every condition that resets it, not just one. SinceHead covers a head change but not a not-ready vouch, so mergeExtra keeps since across old-binary not-ready then ready at the same head, and firstReady trusts it (train.go:448-465). Fix: stamp the reset conditions too, for example a since_at equal to the At of the vouch that set Since, read as valid only while every vouch since then was ready. Prevalence: 2 of 2 carried fields in this family (since; head-reset fixed, readiness-reset open). The window needs mixed binaries on one line, so this is Minor.'
          family: unknown-field-carry-stale-semantics
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      reviewed: fcc0be143bb399a3b0176477be2ed52ca16da8e0
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 40efdd64-4cd8-46fb-84f7-9868fcb0db5a
            repository: github.com/xianxu/ariadne
            issue_id: "000475"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/2e9a64e93116eecf
            reviewed_head: d67b8db5ac58de96a3c1cab4ee5cc461d14deee2
            slot: ariadne:4
            requested_at: "2026-10-20T12:24:38.408162Z"
          phase: finished
          pid: 19138
          started_at: "2026-10-20T12:24:38.568946Z"
          finished_at: "2026-10-20T12:27:53.021787Z"
          session_id: 390f745e-756b-46cd-ac8f-b0502e20737e
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: d67b8db5ac58de96a3c1cab4ee5cc461d14deee2
        tracker_head: 2ac63bb6e238df70335b90818a478054f4e7a5b3
        branch_point: 7918d3d63c4c12e9a4537b51680e2143b29fd286
        since: "2026-10-20T03:03:17-07:00"
        until: "2026-10-20T05:23:51-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T03:03:17-07:00"
              end: "2026-10-20T05:23:51-07:00"
              began: "2026-10-20T03:03:17-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 8c0152f07a5118e184695d666e8e33094d8972984eff162f41403c31e5351ea5
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 97266442
              windows:
                - start: "2026-10-20T03:03:17-07:00"
                  end: "2026-10-20T05:23:51-07:00"
            - harness: claude
              id: 3fc56c78-9cad-4e12-a1a7-aed05bd0c63d
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: b676d6d103897ad92941db345e4e23c27533a35891f17b7b537450d5976ee8ad
              started: "2026-10-20T11:21:49.422Z"
              spawned: false
              bytes: 384376
              windows:
                - start: "2026-10-20T11:21:49.422Z"
                  end: "2026-10-20T05:23:51-07:00"
            - harness: claude
              id: 95466ede-cf18-4115-b13f-1faa8f33645d
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 1b281c55e873b03c51fd775f439ef68f3e2a49b0a3f6e6097df1fe069c81db79
              started: "2026-10-20T11:21:30.104Z"
              spawned: false
              bytes: 514596
              windows:
                - start: "2026-10-20T11:21:30.104Z"
                  end: "2026-10-20T05:23:51-07:00"
            - harness: claude
              id: ac351657-0a41-4804-b564-3e710ea2dd55
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 762c45ca71122a0c54e4c27a0ff8e23460c28233ba39f58b9420ecf8d981cda4
              started: "2026-10-20T11:19:02.044Z"
              spawned: false
              bytes: 587879
              windows:
                - start: "2026-10-20T11:19:02.044Z"
                  end: "2026-10-20T05:23:51-07:00"
            - harness: claude
              id: f5886db6-bd63-4d0e-aec8-021fea4f0187
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: ff3b2eef3104ceeab9229868a244b067cce02222561142ea5df5f3dc7ec0d977
              started: "2026-10-20T11:20:37.435Z"
              spawned: false
              bytes: 530611
              windows:
                - start: "2026-10-20T11:20:37.435Z"
                  end: "2026-10-20T05:23:51-07:00"
        queue_waits:
            - attempt_id: 3611f175d5a4d5f65847189d1c3190ac
              slot: ariadne:4
              record_digest: 0bfdc61fc34a7d4333b805a3bf32ca5a4d0d8dabb13eea6559cf546adf6609a2
              start: "2026-10-20T10:14:59.477681Z"
              end: "2026-10-20T10:30:07.515298Z"
              queue_seconds: 908.037617
              raw_seconds: 908.037617
              exact: true
            - attempt_id: b648e3ff6d96d6e5b355164d3dbee74c
              slot: ariadne:4
              record_digest: 11df85f04d90f5060d15b092a898573c63d1a6d63d3891bf06bff55a7853326d
              start: "2026-10-20T11:03:21.374039Z"
              end: "2026-10-20T11:04:29.093112Z"
              queue_seconds: 67.719073
              raw_seconds: 67.719073
              exact: true
        protected:
            - issue: "490"
              slot: ariadne:4
              start: "2026-10-20T10:05:34.25Z"
              end: "2026-10-20T10:11:07.378Z"
        cuts:
            - boundary: M1
              dispatched: "2026-10-20T04:40:01-07:00"
              finalized: "2026-10-20T04:51:27-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-20T04:55:52-07:00"
              finalized: "2026-10-20T05:08:05-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-20T05:27:53-07:00"
      dispatched: "2026-10-20T05:24:38-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-8
          severity: Important
          title: landing status builds its schedule without the machine plan, so it reports a different next departure than admission
          detail: 'landingcmd.go:182 calls readLandingSchedule directly (Offset always nil); admission uses landingScheduleSource, which applies the plan (landingqueue.go:248). Example: offset 15, a full train leaves 08:40; admission''s next departure is the 09:15 slot, status says 09:40. The state table''s plan-missing and plan-unreadable lines are never printed. 2nd finding in this family. Rule: every consumer of a repo''s effective schedule gets it from one constructor that combines config and plan. Sweep every readLandingSchedule caller: landingqueue.go:237 and landingcmd.go:182 move to it; landingplan.go:181 needs config only.'
          family: shared-helper-not-repeated-predicate
          round: 3
        - kind: defect
          id: BR-9
          severity: Important
          title: T9's fixture-fleet golden test is ticked but fleetPlanInputs has no test; T7 and T9 sources and the M2 live dry-run claim also don't match delivery
          detail: 'fleetPlanInputs (landingplan.go:161) has no test: trains-on filter, local-checks weight, the no-sample default, the test-stats name join. It hard-codes execGitRunner{}, so it can''t be faked. Also: T7''s WF_TEST_LOCK_FILE path is untested (tests set it to ""); T9 names sdlc fleet inventory but the code uses fleetRepoPaths; the issue''s M2 checkbox claims a live dry run while T11 is unchecked. 4th finding in this family. Rule: tick a task only when each named artifact (test, data source, live step) exists as named; otherwise revise the plan text in the same commit. Sweep T6-T11 and the issue''s M2 line.'
          family: plan-claim-matches-delivery
          round: 3
        - kind: defect
          id: BR-10
          severity: Important
          title: README update appears missing for sdlc landing plan [--write], the landing-plan.json path and WF_TEST_SUITES
          detail: 'README.md:521-557 describes the departure rule but not the new subcommand or its state file. 2nd finding in this family. Rule: every new subcommand, flag, config key, env var or state-file path in a range gets a README line in the same range. Sweep newLandingPlanCmd''s flags, plus the env inputs of machinePlanPath and machineSuiteSlots.'
          family: user-docs-track-new-surface
          round: 3
        - kind: defect
          id: BR-11
          severity: Important
          title: PackOffsets can place suites over the machine's slot count and landing plan never warns (ARCH-CONSTRAINTS)
          detail: When no offset fits, plan.go:208-216 takes the lowest-peak offset even if the peak is over slots; planFromCosts emits no warning, and the property test skips exactly that case (total > 60). Return each placement's peak, warn when it exceeds slots, and test an infeasible input.
          family: silent-out-of-envelope
          round: 3
        - kind: defect
          id: BR-12
          severity: Minor
          title: Trains-on e2e tests read the real user's landing-plan.json via landingScheduleSource
          detail: Only the two new plan tests set XDG_STATE_HOME and WF_TEST_LOCK_FILE. Set both to a temp dir in the shared landing fixture.
          family: tests-isolate-user-state
          round: 3
        - kind: defect
          id: BR-13
          severity: Minor
          title: strictObject silently accepts duplicate JSON keys in the plan file (last one wins)
          detail: ARCH-EVOLVE asks for duplicate keys to stay strict. Detect them with a token-level decode.
          family: strict-known-fields
          round: 3
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 5a3f636c-4d8b-4c38-8982-9f6fc759b654
            repository: github.com/xianxu/ariadne
            issue_id: "000475"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/e39ea9a134db5f9c
            reviewed_head: 30d34a5a4f7cf329e6bc570f760f6669454efb3b
            slot: ariadne:4
            requested_at: "2026-10-20T12:31:44.411385Z"
          phase: finished
          pid: 9152
          started_at: "2026-10-20T12:31:44.572169Z"
          finished_at: "2026-10-20T13:00:53.351977Z"
          session_id: 6a22593e-ed3c-41ae-bec6-c3e918be71de
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 30d34a5a4f7cf329e6bc570f760f6669454efb3b
        tracker_head: 590710092c264e352c84aa2d4ba25e6c7c9929c1
        branch_point: 7918d3d63c4c12e9a4537b51680e2143b29fd286
        since: "2026-10-20T03:03:17-07:00"
        until: "2026-10-20T05:31:12-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T03:03:17-07:00"
              end: "2026-10-20T05:31:12-07:00"
              began: "2026-10-20T03:03:17-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: ac1674ccac44a0fe9ba5d9fd6331ac1e0b93f999348b528517be12cbee12005e
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 97557917
              windows:
                - start: "2026-10-20T03:03:17-07:00"
                  end: "2026-10-20T05:31:12-07:00"
            - harness: claude
              id: 3fc56c78-9cad-4e12-a1a7-aed05bd0c63d
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: b676d6d103897ad92941db345e4e23c27533a35891f17b7b537450d5976ee8ad
              started: "2026-10-20T11:21:49.422Z"
              spawned: false
              bytes: 384376
              windows:
                - start: "2026-10-20T11:21:49.422Z"
                  end: "2026-10-20T05:31:12-07:00"
            - harness: claude
              id: 95466ede-cf18-4115-b13f-1faa8f33645d
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 1b281c55e873b03c51fd775f439ef68f3e2a49b0a3f6e6097df1fe069c81db79
              started: "2026-10-20T11:21:30.104Z"
              spawned: false
              bytes: 514596
              windows:
                - start: "2026-10-20T11:21:30.104Z"
                  end: "2026-10-20T05:31:12-07:00"
            - harness: claude
              id: ac351657-0a41-4804-b564-3e710ea2dd55
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 762c45ca71122a0c54e4c27a0ff8e23460c28233ba39f58b9420ecf8d981cda4
              started: "2026-10-20T11:19:02.044Z"
              spawned: false
              bytes: 587879
              windows:
                - start: "2026-10-20T11:19:02.044Z"
                  end: "2026-10-20T05:31:12-07:00"
            - harness: claude
              id: f5886db6-bd63-4d0e-aec8-021fea4f0187
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: ff3b2eef3104ceeab9229868a244b067cce02222561142ea5df5f3dc7ec0d977
              started: "2026-10-20T11:20:37.435Z"
              spawned: false
              bytes: 530611
              windows:
                - start: "2026-10-20T11:20:37.435Z"
                  end: "2026-10-20T05:31:12-07:00"
        queue_waits:
            - attempt_id: 3611f175d5a4d5f65847189d1c3190ac
              slot: ariadne:4
              record_digest: 0bfdc61fc34a7d4333b805a3bf32ca5a4d0d8dabb13eea6559cf546adf6609a2
              start: "2026-10-20T10:14:59.477681Z"
              end: "2026-10-20T10:30:07.515298Z"
              queue_seconds: 908.037617
              raw_seconds: 908.037617
              exact: true
            - attempt_id: b648e3ff6d96d6e5b355164d3dbee74c
              slot: ariadne:4
              record_digest: 11df85f04d90f5060d15b092a898573c63d1a6d63d3891bf06bff55a7853326d
              start: "2026-10-20T11:03:21.374039Z"
              end: "2026-10-20T11:04:29.093112Z"
              queue_seconds: 67.719073
              raw_seconds: 67.719073
              exact: true
        protected:
            - issue: "490"
              slot: ariadne:4
              start: "2026-10-20T10:05:34.25Z"
              end: "2026-10-20T10:11:07.378Z"
        cuts:
            - boundary: M1
              dispatched: "2026-10-20T04:40:01-07:00"
              finalized: "2026-10-20T04:51:27-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-20T04:55:52-07:00"
              finalized: "2026-10-20T05:08:05-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-20T05:24:38-07:00"
              finalized: "2026-10-20T05:27:53-07:00"
              finalizing: false
      "n": 4
      timestamp: "2026-10-20T06:01:08-07:00"
      dispatched: "2026-10-20T05:31:44-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-8
          disposition: not-addressed
          note: effectiveSchedule is shared, but status keys the plan on filepath.Base(repo) and --repo defaults to ".", so the key is "."; admission uses Base(WorktreeRoot). The default `sdlc landing status` still reports the interval departure when admission uses the slot. Derive the key inside the constructor from the resolved top level, and test --repo . or a subdirectory.
          round: 4
        - id: BR-9
          disposition: addressed
          note: fleetPlanInputs is injectable and TestFleetPlanInputs covers the filter, weight, no-sample default and name join. TestMachinePlanPath covers WF_TEST_LOCK_FILE and XDG. T9 now names fleetRepoPaths, and the live dry run moved to an unchecked line in the issue.
          round: 4
        - id: BR-10
          disposition: addressed
          note: README.md:540-554 documents sdlc landing plan [--write], the plan path, WF_TEST_LOCK_FILE, WF_TEST_SUITES and the fallback; train_max_wait_minutes is at :523.
          round: 4
        - id: BR-11
          disposition: addressed
          note: PackOffsets returns overloads with peaks; reverting the bestPeak > slots check fails TestPackOffsetsNamesAnOverload. The planFromCosts warning text has no test (new Minor below).
          round: 4
        - id: BR-12
          disposition: addressed
          note: isolateMachinePlan is in landingFixture, which integrationFixture also uses; the remaining unit callers of landingScheduleSource have trains off.
          round: 4
        - id: BR-13
          disposition: addressed
          note: The token-level duplicate check works at every level; reverting it makes TestMachinePlanCodec accept all three new duplicate cases.
          round: 4
      findings:
        - kind: defect
          id: BR-14
          severity: Minor
          title: landing plan's "does not fit" warning has no test; deleting it keeps the suite green
          detail: 'This is the 2nd finding in family silent-out-of-envelope. The rule: every out-of-envelope signal is pinned by a test at the surface the operator reads, not only in the pure helper. PackOffsets'' over map is tested; the planFromCosts loop that turns it into the warning (landingplan.go:53-61) is not. Add an overloaded stubPlanInputs case to TestLandingPlanCommand that checks for "does not fit".'
          family: silent-out-of-envelope
          round: 4
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 8afdfe85-1b12-4840-aed6-ba1fbfab7c97
            repository: github.com/xianxu/ariadne
            issue_id: "000475"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/1dd7f04750a43b8e
            reviewed_head: 61af3a94f9a6f05669bbadd80bd5ab4d9d44aebf
            slot: ariadne:4
            requested_at: "2026-10-20T13:04:56.076592Z"
          phase: finished
          pid: 32275
          started_at: "2026-10-20T13:04:56.244633Z"
          finished_at: "2026-10-20T13:09:40.377767Z"
          session_id: 75f34f1f-15d8-4dc4-9061-ee3bf1c3e240
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 61af3a94f9a6f05669bbadd80bd5ab4d9d44aebf
        tracker_head: 0087c152913b3e3d966d63b3e4dd33187eeaccb8
        branch_point: 7918d3d63c4c12e9a4537b51680e2143b29fd286
        since: "2026-10-20T03:03:17-07:00"
        until: "2026-10-20T06:04:20-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T03:03:17-07:00"
              end: "2026-10-20T06:04:20-07:00"
              began: "2026-10-20T03:03:17-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: da46c75bbf129df2ea49859dc99051cf2da4546478f72831e78088c35f08e606
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 97811596
              windows:
                - start: "2026-10-20T03:03:17-07:00"
                  end: "2026-10-20T06:04:20-07:00"
            - harness: claude
              id: 3fc56c78-9cad-4e12-a1a7-aed05bd0c63d
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: b676d6d103897ad92941db345e4e23c27533a35891f17b7b537450d5976ee8ad
              started: "2026-10-20T11:21:49.422Z"
              spawned: false
              bytes: 384376
              windows:
                - start: "2026-10-20T11:21:49.422Z"
                  end: "2026-10-20T06:04:20-07:00"
            - harness: claude
              id: 95466ede-cf18-4115-b13f-1faa8f33645d
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 1b281c55e873b03c51fd775f439ef68f3e2a49b0a3f6e6097df1fe069c81db79
              started: "2026-10-20T11:21:30.104Z"
              spawned: false
              bytes: 514596
              windows:
                - start: "2026-10-20T11:21:30.104Z"
                  end: "2026-10-20T06:04:20-07:00"
            - harness: claude
              id: ac351657-0a41-4804-b564-3e710ea2dd55
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 762c45ca71122a0c54e4c27a0ff8e23460c28233ba39f58b9420ecf8d981cda4
              started: "2026-10-20T11:19:02.044Z"
              spawned: false
              bytes: 587879
              windows:
                - start: "2026-10-20T11:19:02.044Z"
                  end: "2026-10-20T06:04:20-07:00"
            - harness: claude
              id: f5886db6-bd63-4d0e-aec8-021fea4f0187
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: ff3b2eef3104ceeab9229868a244b067cce02222561142ea5df5f3dc7ec0d977
              started: "2026-10-20T11:20:37.435Z"
              spawned: false
              bytes: 530611
              windows:
                - start: "2026-10-20T11:20:37.435Z"
                  end: "2026-10-20T06:04:20-07:00"
        queue_waits:
            - attempt_id: 3611f175d5a4d5f65847189d1c3190ac
              slot: ariadne:4
              record_digest: 0bfdc61fc34a7d4333b805a3bf32ca5a4d0d8dabb13eea6559cf546adf6609a2
              start: "2026-10-20T10:14:59.477681Z"
              end: "2026-10-20T10:30:07.515298Z"
              queue_seconds: 908.037617
              raw_seconds: 908.037617
              exact: true
            - attempt_id: b648e3ff6d96d6e5b355164d3dbee74c
              slot: ariadne:4
              record_digest: 11df85f04d90f5060d15b092a898573c63d1a6d63d3891bf06bff55a7853326d
              start: "2026-10-20T11:03:21.374039Z"
              end: "2026-10-20T11:04:29.093112Z"
              queue_seconds: 67.719073
              raw_seconds: 67.719073
              exact: true
        protected:
            - issue: "490"
              slot: ariadne:4
              start: "2026-10-20T10:05:34.25Z"
              end: "2026-10-20T10:11:07.378Z"
        cuts:
            - boundary: M1
              dispatched: "2026-10-20T04:40:01-07:00"
              finalized: "2026-10-20T04:51:27-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-20T04:55:52-07:00"
              finalized: "2026-10-20T05:08:05-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-20T05:24:38-07:00"
              finalized: "2026-10-20T05:27:53-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-20T05:31:44-07:00"
              finalized: "2026-10-20T06:01:08-07:00"
              finalizing: false
      "n": 5
      timestamp: "2026-10-20T06:09:41-07:00"
      dispatched: "2026-10-20T06:04:56-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-8
          disposition: addressed
          note: effectiveSchedule is used by landingScheduleSource and runLandingStatus; reverting status to readLandingSchedule fails TestLandingStatusShowsTrainsAndReadiness ("status ignores the machine plan").
          round: 5
        - id: BR-14
          disposition: addressed
          note: TestLandingPlanCommand has an overloaded stubPlanInputs case; deleting the planFromCosts "does not fit" loop turns it red (verified by revert).
          round: 5
      findings:
        - kind: defect
          id: BR-15
          severity: Minor
          title: Plan prose says out-of-range offsets are errors and plan-read failures warn once per TTL; the code reduces offsets modulo the interval and warns once per source
          detail: 'This is the 5th finding in family plan-claim-matches-delivery. Rule: every behavioural sentence in the plan''s Core concepts names a test or code site that enforces it, or the sentence goes in Revisions. Instances: plan.go:38 (MachinePlan.Offset accepts 0..1439 and takes it modulo the interval, never refuses) and landingqueue.go:242 (planWarned is set once for the source''s lifetime, not per TTL). Fix by adding a Revisions entry for both, not by changing the code.'
          family: plan-claim-matches-delivery
          round: 5
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      reviewed: 61af3a94f9a6f05669bbadd80bd5ab4d9d44aebf
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: ac1c618b-8c05-464a-a239-1388169a2a59
            repository: github.com/xianxu/ariadne
            issue_id: "000475"
            recipe: milestone-review
            round: 6
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/f41731c89a9d1958
            reviewed_head: fc32559e438281ea63d14ab4fa4905e0c299e0b7
            slot: ariadne:4
            requested_at: "2026-10-20T13:29:31.273042Z"
          phase: finished
          pid: 48153
          started_at: "2026-10-20T13:29:31.587198Z"
          finished_at: "2026-10-20T13:41:00.853025Z"
          session_id: 37e6231c-1a97-4ea7-9a3c-7967a3f85c3f
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: fc32559e438281ea63d14ab4fa4905e0c299e0b7
        tracker_head: e2b10738a58996c2a8bd0f36651313df5b6986f3
        branch_point: c6fb9859856d942bef3be366946737aa3c47ba64
        since: "2026-10-20T03:03:17-07:00"
        until: "2026-10-20T06:10:06-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T03:03:17-07:00"
              end: "2026-10-20T06:10:06-07:00"
              began: "2026-10-20T03:03:17-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 7a1469d873de507d664114760758f4d07526f2b31d04ba197b28dcda5abebcd5
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 98145956
              windows:
                - start: "2026-10-20T03:03:17-07:00"
                  end: "2026-10-20T06:10:06-07:00"
            - harness: claude
              id: 3fc56c78-9cad-4e12-a1a7-aed05bd0c63d
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: b676d6d103897ad92941db345e4e23c27533a35891f17b7b537450d5976ee8ad
              started: "2026-10-20T11:21:49.422Z"
              spawned: false
              bytes: 384376
              windows:
                - start: "2026-10-20T11:21:49.422Z"
                  end: "2026-10-20T06:10:06-07:00"
            - harness: claude
              id: 95466ede-cf18-4115-b13f-1faa8f33645d
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 1b281c55e873b03c51fd775f439ef68f3e2a49b0a3f6e6097df1fe069c81db79
              started: "2026-10-20T11:21:30.104Z"
              spawned: false
              bytes: 514596
              windows:
                - start: "2026-10-20T11:21:30.104Z"
                  end: "2026-10-20T06:10:06-07:00"
            - harness: claude
              id: ac351657-0a41-4804-b564-3e710ea2dd55
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 762c45ca71122a0c54e4c27a0ff8e23460c28233ba39f58b9420ecf8d981cda4
              started: "2026-10-20T11:19:02.044Z"
              spawned: false
              bytes: 587879
              windows:
                - start: "2026-10-20T11:19:02.044Z"
                  end: "2026-10-20T06:10:06-07:00"
            - harness: claude
              id: f5886db6-bd63-4d0e-aec8-021fea4f0187
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: ff3b2eef3104ceeab9229868a244b067cce02222561142ea5df5f3dc7ec0d977
              started: "2026-10-20T11:20:37.435Z"
              spawned: false
              bytes: 530611
              windows:
                - start: "2026-10-20T11:20:37.435Z"
                  end: "2026-10-20T06:10:06-07:00"
        queue_waits:
            - attempt_id: 3611f175d5a4d5f65847189d1c3190ac
              slot: ariadne:4
              record_digest: 0bfdc61fc34a7d4333b805a3bf32ca5a4d0d8dabb13eea6559cf546adf6609a2
              start: "2026-10-20T10:14:59.477681Z"
              end: "2026-10-20T10:30:07.515298Z"
              queue_seconds: 908.037617
              raw_seconds: 908.037617
              exact: true
            - attempt_id: b648e3ff6d96d6e5b355164d3dbee74c
              slot: ariadne:4
              record_digest: 11df85f04d90f5060d15b092a898573c63d1a6d63d3891bf06bff55a7853326d
              start: "2026-10-20T11:03:21.374039Z"
              end: "2026-10-20T11:04:29.093112Z"
              queue_seconds: 67.719073
              raw_seconds: 67.719073
              exact: true
        protected:
            - issue: "490"
              slot: ariadne:4
              start: "2026-10-20T10:05:34.25Z"
              end: "2026-10-20T10:11:07.378Z"
        cuts:
            - boundary: M1
              dispatched: "2026-10-20T04:40:01-07:00"
              finalized: "2026-10-20T04:51:27-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-20T04:55:52-07:00"
              finalized: "2026-10-20T05:08:05-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-20T05:24:38-07:00"
              finalized: "2026-10-20T05:27:53-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-20T05:31:44-07:00"
              finalized: "2026-10-20T06:01:08-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-20T06:04:56-07:00"
              finalized: "2026-10-20T06:09:41-07:00"
              finalizing: true
      "n": 6
      timestamp: "2026-10-20T06:41:01-07:00"
      dispatched: "2026-10-20T06:29:31-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-7
          disposition: addressed
          note: SinceStamp (train.go:189-194) makes firstReady trust Since only when stamped by the current vouch; reverting the stamp check turns TestStaleSinceFromAnOlderSdlc red (ready again, same head case).
          round: 6
        - id: BR-15
          disposition: addressed
          note: Revisions entry plan.md:248-259 replaces both claims; cited tests exist and assert them (plan_test.go:75 modulo case, :93 offset 1440 refusal; landingplan_test.go:131 warned exactly once).
          round: 6
      recipe: milestone-review
      isolation: fenced
      reviewed: fc32559e438281ea63d14ab4fa4905e0c299e0b7
      blocked: false
---

# Gate ledger — ariadne#475 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T04:51:27-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `user-docs-track-new-surface` (kind: defect) README departure rule is stale and omits train_max_wait_minutes
  README.md:521-531 lists only train_interval_minutes/train_max_cars and describes departure as first train / one interval after the last / at once after left-ready. It omits the full-train and max-wait early departures, the new key, and that lone and granted landings no longer reset the clock. helptext/landing.md is equally stale.
- **BR-2** [Important] `plan-claim-matches-delivery` (kind: defect) T5 is ticked, but landing status lacks the promised reasons and golden test
  The plan promised train due (reason) / next train at HH:MM (slot or max-wait), including cars-left-ready, plus a golden test. The code prints a bare train due for slot and left-ready, and an RFC3339 time with no source. The only test change is a 1-ready substring (landingcmd_test.go:196). Add the reasons and a status test, or narrow T5 in Revisions.
- **BR-3** [Minor] `plan-claim-matches-delivery` (kind: defect) Event.Departed has no reader; the slot check uses the stored NextDue, not the plan's lastDeparture
  PQ-2 says lastDeparture reads Departed. The code compares a NextDue fixed at write time, so Departed is written but never read, and a later offset change only applies after the next train.
- **BR-4** [Minor] `plan-claim-matches-delivery` (kind: defect) NextDeparture does not ignore old-format alone records by Result as the Ordering section claims
  train.go:84 filters on non-zero NextDue only, so an alone record written before the upgrade (departed + interval) still delays the next train once.
- **BR-5** [Minor] `shared-helper-not-repeated-predicate` (kind: defect) The max-wait predicate is duplicated in DepartureDue
  train.go:228 and :234 repeat the same three-part condition; one wait-deadline helper would do.
- **BR-6** [Minor] `unknown-field-carry-stale-semantics` (kind: defect) An older sdlc re-vouching carries a stale since through Extra across a head change
  mergeExtra keeps since as an unknown member for an old binary, so after a head change the new binary trusts the stale Since and may depart early on max-wait.

## Round 2 — 2026-10-20T05:08:05-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — README.md:521-538 and helptext/landing.md now list train_max_wait_minutes, full and max-wait early departures, and that lone/granted landings never push the train back.
- BR-2 — addressed — Status calls DepartureDue and prints the reason (slot, interval, max wait, full, cars left ready); landingcmd_test.go:852-862 pins "next train at 2026-10-20 10:00 (interval)".
- BR-3 — addressed — nextDeparture works out the slot from departedOf(e) when it decides (train.go:373); TestNextDepartureReadsTheDeparture covers it.
- BR-4 — addressed — The alone case is skipped by Result (train.go:370); reverting it makes TestNextDepartureReadsTheDeparture fail (checked in the checkout).
- BR-5 — addressed — waitDeadline helper (train.go:330) used for both the due check and the reported next time.
- BR-6 — addressed — SinceHead guard in firstReady (train.go:449); reverting it makes TestStaleSinceFromAnOlderSdlc fail (checked). A sibling case is raised separately.

### Raised

- **BR-7** [Minor] `unknown-field-carry-stale-semantics` (kind: defect) An older sdlc's not-ready then ready re-vouch at the same head still carries a stale since
  This is the 2nd finding in this family. The rule: a derived field that an older writer can carry through Extra must be stamped against every condition that resets it, not just one. SinceHead covers a head change but not a not-ready vouch, so mergeExtra keeps since across old-binary not-ready then ready at the same head, and firstReady trusts it (train.go:448-465). Fix: stamp the reset conditions too, for example a since_at equal to the At of the vouch that set Since, read as valid only while every vouch since then was ready. Prevalence: 2 of 2 carried fields in this family (since; head-reset fixed, readiness-reset open). The window needs mixed binaries on one line, so this is Minor.

## Round 3 — 2026-10-20T05:27:53-07:00 (claude) — BLOCKED

### Raised

- **BR-8** [Important] `shared-helper-not-repeated-predicate` (kind: defect) landing status builds its schedule without the machine plan, so it reports a different next departure than admission
  landingcmd.go:182 calls readLandingSchedule directly (Offset always nil); admission uses landingScheduleSource, which applies the plan (landingqueue.go:248). Example: offset 15, a full train leaves 08:40; admission's next departure is the 09:15 slot, status says 09:40. The state table's plan-missing and plan-unreadable lines are never printed. 2nd finding in this family. Rule: every consumer of a repo's effective schedule gets it from one constructor that combines config and plan. Sweep every readLandingSchedule caller: landingqueue.go:237 and landingcmd.go:182 move to it; landingplan.go:181 needs config only.
- **BR-9** [Important] `plan-claim-matches-delivery` (kind: defect) T9's fixture-fleet golden test is ticked but fleetPlanInputs has no test; T7 and T9 sources and the M2 live dry-run claim also don't match delivery
  fleetPlanInputs (landingplan.go:161) has no test: trains-on filter, local-checks weight, the no-sample default, the test-stats name join. It hard-codes execGitRunner{}, so it can't be faked. Also: T7's WF_TEST_LOCK_FILE path is untested (tests set it to ""); T9 names sdlc fleet inventory but the code uses fleetRepoPaths; the issue's M2 checkbox claims a live dry run while T11 is unchecked. 4th finding in this family. Rule: tick a task only when each named artifact (test, data source, live step) exists as named; otherwise revise the plan text in the same commit. Sweep T6-T11 and the issue's M2 line.
- **BR-10** [Important] `user-docs-track-new-surface` (kind: defect) README update appears missing for sdlc landing plan [--write], the landing-plan.json path and WF_TEST_SUITES
  README.md:521-557 describes the departure rule but not the new subcommand or its state file. 2nd finding in this family. Rule: every new subcommand, flag, config key, env var or state-file path in a range gets a README line in the same range. Sweep newLandingPlanCmd's flags, plus the env inputs of machinePlanPath and machineSuiteSlots.
- **BR-11** [Important] `silent-out-of-envelope` (kind: defect) PackOffsets can place suites over the machine's slot count and landing plan never warns (ARCH-CONSTRAINTS)
  When no offset fits, plan.go:208-216 takes the lowest-peak offset even if the peak is over slots; planFromCosts emits no warning, and the property test skips exactly that case (total > 60). Return each placement's peak, warn when it exceeds slots, and test an infeasible input.
- **BR-12** [Minor] `tests-isolate-user-state` (kind: defect) Trains-on e2e tests read the real user's landing-plan.json via landingScheduleSource
  Only the two new plan tests set XDG_STATE_HOME and WF_TEST_LOCK_FILE. Set both to a temp dir in the shared landing fixture.
- **BR-13** [Minor] `strict-known-fields` (kind: defect) strictObject silently accepts duplicate JSON keys in the plan file (last one wins)
  ARCH-EVOLVE asks for duplicate keys to stay strict. Detect them with a token-level decode.

## Round 4 — 2026-10-20T06:01:08-07:00 (claude) — BLOCKED

### Disposed

- BR-8 — not-addressed — effectiveSchedule is shared, but status keys the plan on filepath.Base(repo) and --repo defaults to ".", so the key is "."; admission uses Base(WorktreeRoot). The default `sdlc landing status` still reports the interval departure when admission uses the slot. Derive the key inside the constructor from the resolved top level, and test --repo . or a subdirectory.
- BR-9 — addressed — fleetPlanInputs is injectable and TestFleetPlanInputs covers the filter, weight, no-sample default and name join. TestMachinePlanPath covers WF_TEST_LOCK_FILE and XDG. T9 now names fleetRepoPaths, and the live dry run moved to an unchecked line in the issue.
- BR-10 — addressed — README.md:540-554 documents sdlc landing plan [--write], the plan path, WF_TEST_LOCK_FILE, WF_TEST_SUITES and the fallback; train_max_wait_minutes is at :523.
- BR-11 — addressed — PackOffsets returns overloads with peaks; reverting the bestPeak > slots check fails TestPackOffsetsNamesAnOverload. The planFromCosts warning text has no test (new Minor below).
- BR-12 — addressed — isolateMachinePlan is in landingFixture, which integrationFixture also uses; the remaining unit callers of landingScheduleSource have trains off.
- BR-13 — addressed — The token-level duplicate check works at every level; reverting it makes TestMachinePlanCodec accept all three new duplicate cases.

### Raised

- **BR-14** [Minor] `silent-out-of-envelope` (kind: defect) landing plan's "does not fit" warning has no test; deleting it keeps the suite green
  This is the 2nd finding in family silent-out-of-envelope. The rule: every out-of-envelope signal is pinned by a test at the surface the operator reads, not only in the pure helper. PackOffsets' over map is tested; the planFromCosts loop that turns it into the warning (landingplan.go:53-61) is not. Add an overloaded stubPlanInputs case to TestLandingPlanCommand that checks for "does not fit".

## Round 5 — 2026-10-20T06:09:41-07:00 (claude) — passed

### Disposed

- BR-8 — addressed — effectiveSchedule is used by landingScheduleSource and runLandingStatus; reverting status to readLandingSchedule fails TestLandingStatusShowsTrainsAndReadiness ("status ignores the machine plan").
- BR-14 — addressed — TestLandingPlanCommand has an overloaded stubPlanInputs case; deleting the planFromCosts "does not fit" loop turns it red (verified by revert).

### Raised

- **BR-15** [Minor] `plan-claim-matches-delivery` (kind: defect) Plan prose says out-of-range offsets are errors and plan-read failures warn once per TTL; the code reduces offsets modulo the interval and warns once per source
  This is the 5th finding in family plan-claim-matches-delivery. Rule: every behavioural sentence in the plan's Core concepts names a test or code site that enforces it, or the sentence goes in Revisions. Instances: plan.go:38 (MachinePlan.Offset accepts 0..1439 and takes it modulo the interval, never refuses) and landingqueue.go:242 (planWarned is set once for the source's lifetime, not per TTL). Fix by adding a Revisions entry for both, not by changing the code.

## Round 6 — 2026-10-20T06:41:01-07:00 (claude) — passed

### Disposed

- BR-7 — addressed — SinceStamp (train.go:189-194) makes firstReady trust Since only when stamped by the current vouch; reverting the stamp check turns TestStaleSinceFromAnOlderSdlc red (ready again, same head case).
- BR-15 — addressed — Revisions entry plan.md:248-259 replaces both claims; cited tests exist and assert them (plan_test.go:75 modulo case, :93 offset 1440 refusal; landingplan_test.go:131 warned exactly once).

## Open findings

(none — every finding has been disposed)
