---
gate: boundary-review
issue: 443
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 5d67fbf6-8be0-41bc-a333-1cfe4b5dc604
            repository: github.com/xianxu/ariadne
            issue_id: "000443"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/99d57ab5275b223c
            reviewed_head: 9d55411b82bca7659de5a1cd0d34aa9c911edc32
            slot: ariadne:1
            requested_at: "2026-10-19T00:38:10.879495Z"
          phase: finished
          pid: 62582
          started_at: "2026-10-19T00:38:10.989691Z"
          finished_at: "2026-10-19T00:42:02.8808Z"
          session_id: 01a15198-32f2-7b20-83fa-e328e26688cb
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 9d55411b82bca7659de5a1cd0d34aa9c911edc32
        tracker_head: 202d6ff1fcab1d503009681d88c2f222037e6a0c
        branch_point: f92193cf4afc9d2b302e132ec5a639bbe3e05098
        since: "2026-10-18T10:57:47-07:00"
        until: "2026-10-18T17:37:44-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T10:57:47-07:00"
              end: "2026-10-18T17:37:44-07:00"
              began: "2026-10-18T10:57:47-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 8276510c14bc0ba6fa7e0aa6bdc8578a4038628283b5fe6c6a03bd1c75e84e5e
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 125906497
              windows:
                - start: "2026-10-18T10:57:47-07:00"
                  end: "2026-10-18T17:37:44-07:00"
            - harness: codex
              id: 01a1502d-d103-7d90-a966-36555f34085d
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              digest: 1ac460d124eac981ad596f0d48705823e1c2f7bbba9a8812ee061446b28e6fba
              started: "2026-10-18T18:02:22.443Z"
              spawned: true
              bytes: 610760
              windows:
                - start: "2026-10-18T18:02:22.443Z"
                  end: "2026-10-18T17:37:44-07:00"
        queue_waits:
            - attempt_id: 9e88fbb939bad2636b6afa592c4ff95e
              slot: ariadne:1
              record_digest: 3f7ccdd2b0f140a22b2e4a1ced2a98f8a247a4840ac10627c091117a92d21580
              start: "2026-10-18T18:18:21.221443Z"
              end: "2026-10-18T18:21:08.81741Z"
              queue_seconds: 167.595967
              raw_seconds: 167.595967
              exact: true
            - attempt_id: 698f69f0a003585cc432cdcd1f5eb3b8
              slot: ariadne:1
              record_digest: 709f89a6b2198a258b4d1e2a5ddf0e5677d05acaa6155085826842b2ed6b0122
              start: "2026-10-18T18:23:20.078884Z"
              end: "2026-10-18T18:58:32.709201Z"
              queue_seconds: 2112.630317
              raw_seconds: 2112.630317
              exact: true
            - attempt_id: 67f520b25e748b321532b2f03ec9af4b
              slot: ariadne:1
              record_digest: 72df114728cf3a28018e5e325035a5d290f479c604f30207eb1b57e07b678dec
              start: "2026-10-18T19:31:16.893432Z"
              end: "2026-10-18T20:18:20.731582Z"
              queue_seconds: 2823.83815
              raw_seconds: 2823.83815
              exact: true
            - attempt_id: 8933544146d410dbede8a7922f155adc
              slot: ariadne:1
              record_digest: d54bd24444ac435992d1e6c1cc5ee33dd031586c386a95de57aaab50afae30c9
              start: "2026-10-18T20:36:52.829415Z"
              end: "2026-10-18T20:50:46.737447Z"
              queue_seconds: 833.908032
              raw_seconds: 833.908032
              exact: true
            - attempt_id: 99819fa702f1e021121137f9c9b8f1ed
              slot: ariadne:1
              record_digest: d57fecad313c026440466a9bfd494122100073222a9c97905dc1ac691342f554
              start: "2026-10-18T21:12:48.480858Z"
              end: "2026-10-18T21:27:47.819081Z"
              queue_seconds: 899.338223
              raw_seconds: 899.338223
              exact: true
            - attempt_id: 3590fe8eed4602bfab53764dda88e258
              slot: ariadne:1
              record_digest: 93e77f2027052a9bf6da90fee0a07c334c6958f416065a6d439bf49d21af2bf1
              start: "2026-10-18T21:46:29.803654Z"
              end: "2026-10-18T22:08:43.529371Z"
              queue_seconds: 1333.725717
              raw_seconds: 1333.725717
              exact: true
            - attempt_id: 5e2efbf6565ab3e4e6b060cd8cf2c8e1
              slot: ariadne:1
              record_digest: 46581d8031ab24ce87ac237c42c3edbd47d7331400cfa5cccc3b1c07e8e0e726
              start: "2026-10-18T22:22:15.218323Z"
              end: "2026-10-18T22:40:39.914532Z"
              queue_seconds: 1104.696209
              raw_seconds: 1104.696209
              exact: true
            - attempt_id: a2dfc29275b4d784bac97b2a5fb3595a
              slot: ariadne:1
              record_digest: dfa9e4133c1cd0bf47bdaba7ca3b370f40a25f0166713573e3fc82403e834fa9
              start: "2026-10-18T22:55:49.346762Z"
              end: "2026-10-18T23:31:41.735155Z"
              queue_seconds: 2152.388393
              raw_seconds: 2152.388393
              exact: true
            - attempt_id: 1e8ad5cbea933ba71506bf7fb216970d
              slot: ariadne:1
              record_digest: 2aced6673d27ef0ad7c6d0014bf28c82fe8d269bfe2a4609999786323bdd2045
              start: "2026-10-18T23:46:16.404167Z"
              end: "2026-10-18T23:58:12.906455Z"
              queue_seconds: 716.502288
              raw_seconds: 716.502288
              exact: true
            - attempt_id: acf49aad220c79aeefa60a93cabe2c90
              slot: ariadne:1
              record_digest: 53c802ca0b304e38e4396146da0640c824e172d949988bbd56e9659a79f53d36
              start: "2026-10-19T00:27:40.089756Z"
              end: "2026-10-19T00:27:40.145983Z"
              queue_seconds: 0.056227
              raw_seconds: 0.056227
              exact: true
      "n": 1
      timestamp: "2026-10-18T17:42:03-07:00"
      dispatched: "2026-10-18T17:38:10-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Clear protected intervals from the excluded-brain diagnostic pass
          detail: 'cmd/sdlc/actualmeasurement.go:38 adds ProtectedIntervals to options, but the diagnostic copy at line 98 retains them while clearing SourceSlots. An overlapping unowned brain session therefore aborts a valid candidate with "excluded brain provenance: ... without unique slot ownership". A temporary overlay of TestOwnedActualEngineReplayAndExcludedBrainProvenance reproduced this with five protected minutes. Clear ProtectedIntervals alongside QueueIntervals in the diagnostic population and add a combined regression.'
          family: measurement-population-isolation
          round: 1
        - kind: defect
          id: BR-2
          severity: Critical
          title: Do not protect verification-tail time omitted from adopted actuals
          detail: 'cmd/sdlc/actualexclusion.go:206 converts all replayed segments into protection, although actualmeasurement.go:69 adopts only Build+Converge and excludes Tail. An overlay of TestActualQueueMeasurementReplay demonstrated protection of 00:50–01:00 while the result explicitly reported those ten minutes as uncounted tail. This can silently remove another issue''s valid activity. ARCH-PURPOSE: derive protected intervals from the same counted phases that determine adopted hours; cover wholly-tail runs and runs crossing phase boundaries.'
          family: adopted-credit-authority
          round: 1
        - kind: defect
          id: BR-3
          severity: Critical
          title: Rule out unrelated slots before refusing missing adoption evidence
          detail: 'cmd/sdlc/actualexclusion.go:153 requires a saved measurement window before checking same-slot engagement at line 182. A temporally plausible legacy adoption without ActualInputs blocks candidates even when its engagement belongs exclusively to another slot. An overlay of the parked-claim fixture, using candidate slot ariadne:7 versus adopted slot ariadne:6 and omitting the saved inputs, reproduced "issue #649 adopted actual has no trustworthy saved window". Use trustworthy ownership history to rule out unrelated slots before refusing missing pins; retain refusal when same-slot overlap remains possible. ARCH-PURPOSE.'
          family: overlap-relevance-before-evidence
          round: 1
        - kind: defect
          id: BR-4
          severity: Important
          title: Deliver the planned unrelated-card scaling regression
          detail: 'workshop/plans/000443-sdlc-actual-a-parked-claim-double-counts-time-the-slot-spent-on-other-issues-plan.md:37 claims tests count replays for N and 2N unrelated cards, but no such test exists in the reviewed changes. ARCH-CONSTRAINTS: add representative loader coverage proving unrelated cards do not trigger transcript replay, including other-slot and disjoint-window cases, or revise the unsupported verification claim and supply equivalent measured evidence.'
          family: declared-workload-verification
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: bc02a72d-5886-421d-8761-dd310d146b97
            repository: github.com/xianxu/ariadne
            issue_id: "000443"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/4ea3881561c3edfa
            reviewed_head: aea775197d15df067e707abfd47e0f80308db0a9
            slot: ariadne:1
            requested_at: "2026-10-19T01:26:06.012622Z"
          phase: finished
          pid: 11898
          started_at: "2026-10-19T01:26:06.102826Z"
          finished_at: "2026-10-19T01:32:12.373346Z"
          session_id: 01a151c4-1107-7010-b5ee-450d3e820140
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: aea775197d15df067e707abfd47e0f80308db0a9
        tracker_head: 259fad25d00c8cea8109a1945805e0db5c9cd2e4
        branch_point: f92193cf4afc9d2b302e132ec5a639bbe3e05098
        since: "2026-10-18T10:57:47-07:00"
        until: "2026-10-18T18:07:32-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T10:57:47-07:00"
              end: "2026-10-18T18:07:32-07:00"
              began: "2026-10-18T10:57:47-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 4fd8175b6ffaf61ec7234b5edb163c6ee4b4632bd44b6960c5b1aa2d8e82dcaf
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 127503655
              windows:
                - start: "2026-10-18T10:57:47-07:00"
                  end: "2026-10-18T18:07:32-07:00"
            - harness: codex
              id: 01a1519d-702f-77a0-8d09-6318d7215afb
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              digest: e3e4a431c11d0a31d2481ee6ab312b55394e8f1b56ad992b875f00d7e9479585
              started: "2026-10-19T00:43:54.859Z"
              spawned: true
              bytes: 747735
              windows:
                - start: "2026-10-19T00:43:54.859Z"
                  end: "2026-10-18T18:07:32-07:00"
        queue_waits:
            - attempt_id: 9e88fbb939bad2636b6afa592c4ff95e
              slot: ariadne:1
              record_digest: 3f7ccdd2b0f140a22b2e4a1ced2a98f8a247a4840ac10627c091117a92d21580
              start: "2026-10-18T18:18:21.221443Z"
              end: "2026-10-18T18:21:08.81741Z"
              queue_seconds: 167.595967
              raw_seconds: 167.595967
              exact: true
            - attempt_id: 698f69f0a003585cc432cdcd1f5eb3b8
              slot: ariadne:1
              record_digest: 709f89a6b2198a258b4d1e2a5ddf0e5677d05acaa6155085826842b2ed6b0122
              start: "2026-10-18T18:23:20.078884Z"
              end: "2026-10-18T18:58:32.709201Z"
              queue_seconds: 2112.630317
              raw_seconds: 2112.630317
              exact: true
            - attempt_id: 67f520b25e748b321532b2f03ec9af4b
              slot: ariadne:1
              record_digest: 72df114728cf3a28018e5e325035a5d290f479c604f30207eb1b57e07b678dec
              start: "2026-10-18T19:31:16.893432Z"
              end: "2026-10-18T20:18:20.731582Z"
              queue_seconds: 2823.83815
              raw_seconds: 2823.83815
              exact: true
            - attempt_id: 8933544146d410dbede8a7922f155adc
              slot: ariadne:1
              record_digest: d54bd24444ac435992d1e6c1cc5ee33dd031586c386a95de57aaab50afae30c9
              start: "2026-10-18T20:36:52.829415Z"
              end: "2026-10-18T20:50:46.737447Z"
              queue_seconds: 833.908032
              raw_seconds: 833.908032
              exact: true
            - attempt_id: 99819fa702f1e021121137f9c9b8f1ed
              slot: ariadne:1
              record_digest: d57fecad313c026440466a9bfd494122100073222a9c97905dc1ac691342f554
              start: "2026-10-18T21:12:48.480858Z"
              end: "2026-10-18T21:27:47.819081Z"
              queue_seconds: 899.338223
              raw_seconds: 899.338223
              exact: true
            - attempt_id: 3590fe8eed4602bfab53764dda88e258
              slot: ariadne:1
              record_digest: 93e77f2027052a9bf6da90fee0a07c334c6958f416065a6d439bf49d21af2bf1
              start: "2026-10-18T21:46:29.803654Z"
              end: "2026-10-18T22:08:43.529371Z"
              queue_seconds: 1333.725717
              raw_seconds: 1333.725717
              exact: true
            - attempt_id: 5e2efbf6565ab3e4e6b060cd8cf2c8e1
              slot: ariadne:1
              record_digest: 46581d8031ab24ce87ac237c42c3edbd47d7331400cfa5cccc3b1c07e8e0e726
              start: "2026-10-18T22:22:15.218323Z"
              end: "2026-10-18T22:40:39.914532Z"
              queue_seconds: 1104.696209
              raw_seconds: 1104.696209
              exact: true
            - attempt_id: a2dfc29275b4d784bac97b2a5fb3595a
              slot: ariadne:1
              record_digest: dfa9e4133c1cd0bf47bdaba7ca3b370f40a25f0166713573e3fc82403e834fa9
              start: "2026-10-18T22:55:49.346762Z"
              end: "2026-10-18T23:31:41.735155Z"
              queue_seconds: 2152.388393
              raw_seconds: 2152.388393
              exact: true
            - attempt_id: 1e8ad5cbea933ba71506bf7fb216970d
              slot: ariadne:1
              record_digest: 2aced6673d27ef0ad7c6d0014bf28c82fe8d269bfe2a4609999786323bdd2045
              start: "2026-10-18T23:46:16.404167Z"
              end: "2026-10-18T23:58:12.906455Z"
              queue_seconds: 716.502288
              raw_seconds: 716.502288
              exact: true
            - attempt_id: acf49aad220c79aeefa60a93cabe2c90
              slot: ariadne:1
              record_digest: 53c802ca0b304e38e4396146da0640c824e172d949988bbd56e9659a79f53d36
              start: "2026-10-19T00:27:40.089756Z"
              end: "2026-10-19T00:27:40.145983Z"
              queue_seconds: 0.056227
              raw_seconds: 0.056227
              exact: true
        protected:
            - issue: "437"
              slot: ariadne:1
              start: "2026-10-18T10:57:47-07:00"
              end: "2026-10-18T18:18:21.221443Z"
            - issue: "437"
              slot: ariadne:1
              start: "2026-10-18T18:02:23.518Z"
              end: "2026-10-18T18:04:14.653Z"
            - issue: "437"
              slot: ariadne:1
              start: "2026-10-18T18:21:08.81741Z"
              end: "2026-10-18T18:23:20.078884Z"
            - issue: "437"
              slot: ariadne:1
              start: "2026-10-18T18:58:32.709201Z"
              end: "2026-10-18T12:10:24-07:00"
            - issue: "437"
              slot: ariadne:1
              start: "2026-10-18T12:10:24-07:00"
              end: "2026-10-18T19:31:16.893432Z"
            - issue: "437"
              slot: ariadne:1
              start: "2026-10-18T20:18:20.731582Z"
              end: "2026-10-18T20:36:52.829415Z"
            - issue: "437"
              slot: ariadne:1
              start: "2026-10-18T20:50:46.737447Z"
              end: "2026-10-18T21:00:12.19Z"
            - issue: "444"
              slot: ariadne:1
              start: "2026-10-18T21:37:59.877Z"
              end: "2026-10-18T21:46:29.803654Z"
            - issue: "444"
              slot: ariadne:1
              start: "2026-10-18T22:08:43.529371Z"
              end: "2026-10-18T15:18:56-07:00"
            - issue: "444"
              slot: ariadne:1
              start: "2026-10-18T15:18:56-07:00"
              end: "2026-10-18T22:21:52.006Z"
        cuts:
            - dispatched: "2026-10-18T17:38:10-07:00"
              finalized: "2026-10-18T17:42:03-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-18T18:32:20-07:00"
      dispatched: "2026-10-18T18:26:06-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: The diagnostic pass clears ProtectedIntervals. Removing that line makes TestOwnedActualEngineReplayAndExcludedBrainProvenance fail with the original ownership error.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Protection intersects the same counted phases used for adopted hours. Restoring full-segment protection fails both TestCreditedIntervalsProtectOnlyCountedPhases and TestActualQueueMeasurementReplay, covering wholly-Tail and boundary-crossing runs.
          round: 2
        - id: BR-3
          disposition: addressed
          note: 'Proven other-slot ownership is checked before missing-window refusal. Disabling differentSlot makes the parked-claim regression fail on legacy issue #650 without saved inputs.'
          round: 2
        - id: BR-4
          disposition: addressed
          note: The parked-claim fixture now checks four and eight unrelated adoptions while retaining same-slot protection, plus a disjoint-window case after transcript loss. These pass; disabling unrelated-slot rejection makes the four-card assertion fail.
          round: 2
      findings:
        - kind: defect
          id: BR-5
          severity: Critical
          title: Refuse overlapping adopted actuals whose slot relationship remains unknown
          detail: 'cmd/sdlc/actualexclusion.go:232–243 treats absence of a proven same-slot engagement as proof of irrelevance. Legacy owners may have an empty Workspace: differentSlot correctly declines to prove them unrelated, but the subsequent loop finds no match and silently continues. In a disposable copy of the parked-claim fixture, clearing only issue #649''s claimant workspace produced a measured 0.333333h for #636 with no exclusions, recounting #649''s adopted interval. An explicit refusal regression fails on head and passes with a temporary unknown-slot guard. ARCH-PURPOSE, ARCH-SECURE: this is the 2nd finding in family overlap-relevance-before-evidence. Earlier rounds fixed instances; do NOT fix only this instance. State and enforce the rule that skipping requires affirmative evidence of disjoint time or slot, enumerate unknown/empty/mixed ownership cases, and sweep them with regressions.'
          family: overlap-relevance-before-evidence
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 8966d7ed-836f-417b-ace2-6fac9be315af
            repository: github.com/xianxu/ariadne
            issue_id: "000443"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/16caa765eb6cea21
            reviewed_head: 27b80215792fd4d22d3316830c11f88bbcabf1b5
            slot: ariadne:1
            requested_at: "2026-10-19T02:24:00.304988Z"
          phase: finished
          pid: 67593
          started_at: "2026-10-19T02:24:00.45813Z"
          finished_at: "2026-10-19T02:29:57.291668Z"
          session_id: 01a151f9-14d7-77e2-8c9a-02399a906eea
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 27b80215792fd4d22d3316830c11f88bbcabf1b5
        tracker_head: af2f2946264d2ed9c569075f9372532b20bd8faa
        branch_point: f92193cf4afc9d2b302e132ec5a639bbe3e05098
        since: "2026-10-18T10:57:47-07:00"
        until: "2026-10-18T18:52:08-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T10:57:47-07:00"
              end: "2026-10-18T18:52:08-07:00"
              began: "2026-10-18T10:57:47-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: b159cecc22e3f429e33ded71ea3a7608e41be401c9c11b86d7137addbc5830ca
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 128640569
              windows:
                - start: "2026-10-18T10:57:47-07:00"
                  end: "2026-10-18T18:52:08-07:00"
            - harness: codex
              id: 01a1519d-702f-77a0-8d09-6318d7215afb
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              digest: e3e4a431c11d0a31d2481ee6ab312b55394e8f1b56ad992b875f00d7e9479585
              started: "2026-10-19T00:43:54.859Z"
              spawned: true
              bytes: 747735
              windows:
                - start: "2026-10-19T00:43:54.859Z"
                  end: "2026-10-18T18:52:08-07:00"
        queue_waits:
            - attempt_id: 9e88fbb939bad2636b6afa592c4ff95e
              slot: ariadne:1
              record_digest: 3f7ccdd2b0f140a22b2e4a1ced2a98f8a247a4840ac10627c091117a92d21580
              start: "2026-10-18T18:18:21.221443Z"
              end: "2026-10-18T18:21:08.81741Z"
              queue_seconds: 167.595967
              raw_seconds: 167.595967
              exact: true
            - attempt_id: 698f69f0a003585cc432cdcd1f5eb3b8
              slot: ariadne:1
              record_digest: 709f89a6b2198a258b4d1e2a5ddf0e5677d05acaa6155085826842b2ed6b0122
              start: "2026-10-18T18:23:20.078884Z"
              end: "2026-10-18T18:58:32.709201Z"
              queue_seconds: 2112.630317
              raw_seconds: 2112.630317
              exact: true
            - attempt_id: 67f520b25e748b321532b2f03ec9af4b
              slot: ariadne:1
              record_digest: 72df114728cf3a28018e5e325035a5d290f479c604f30207eb1b57e07b678dec
              start: "2026-10-18T19:31:16.893432Z"
              end: "2026-10-18T20:18:20.731582Z"
              queue_seconds: 2823.83815
              raw_seconds: 2823.83815
              exact: true
            - attempt_id: 8933544146d410dbede8a7922f155adc
              slot: ariadne:1
              record_digest: d54bd24444ac435992d1e6c1cc5ee33dd031586c386a95de57aaab50afae30c9
              start: "2026-10-18T20:36:52.829415Z"
              end: "2026-10-18T20:50:46.737447Z"
              queue_seconds: 833.908032
              raw_seconds: 833.908032
              exact: true
            - attempt_id: 99819fa702f1e021121137f9c9b8f1ed
              slot: ariadne:1
              record_digest: d57fecad313c026440466a9bfd494122100073222a9c97905dc1ac691342f554
              start: "2026-10-18T21:12:48.480858Z"
              end: "2026-10-18T21:27:47.819081Z"
              queue_seconds: 899.338223
              raw_seconds: 899.338223
              exact: true
            - attempt_id: 3590fe8eed4602bfab53764dda88e258
              slot: ariadne:1
              record_digest: 93e77f2027052a9bf6da90fee0a07c334c6958f416065a6d439bf49d21af2bf1
              start: "2026-10-18T21:46:29.803654Z"
              end: "2026-10-18T22:08:43.529371Z"
              queue_seconds: 1333.725717
              raw_seconds: 1333.725717
              exact: true
            - attempt_id: 5e2efbf6565ab3e4e6b060cd8cf2c8e1
              slot: ariadne:1
              record_digest: 46581d8031ab24ce87ac237c42c3edbd47d7331400cfa5cccc3b1c07e8e0e726
              start: "2026-10-18T22:22:15.218323Z"
              end: "2026-10-18T22:40:39.914532Z"
              queue_seconds: 1104.696209
              raw_seconds: 1104.696209
              exact: true
            - attempt_id: a2dfc29275b4d784bac97b2a5fb3595a
              slot: ariadne:1
              record_digest: dfa9e4133c1cd0bf47bdaba7ca3b370f40a25f0166713573e3fc82403e834fa9
              start: "2026-10-18T22:55:49.346762Z"
              end: "2026-10-18T23:31:41.735155Z"
              queue_seconds: 2152.388393
              raw_seconds: 2152.388393
              exact: true
            - attempt_id: 1e8ad5cbea933ba71506bf7fb216970d
              slot: ariadne:1
              record_digest: 2aced6673d27ef0ad7c6d0014bf28c82fe8d269bfe2a4609999786323bdd2045
              start: "2026-10-18T23:46:16.404167Z"
              end: "2026-10-18T23:58:12.906455Z"
              queue_seconds: 716.502288
              raw_seconds: 716.502288
              exact: true
            - attempt_id: acf49aad220c79aeefa60a93cabe2c90
              slot: ariadne:1
              record_digest: 53c802ca0b304e38e4396146da0640c824e172d949988bbd56e9659a79f53d36
              start: "2026-10-19T00:27:40.089756Z"
              end: "2026-10-19T00:27:40.145983Z"
              queue_seconds: 0.056227
              raw_seconds: 0.056227
              exact: true
            - attempt_id: 7057b08b3caa03d551d383ea29b03eec
              slot: ariadne:1
              record_digest: a28b4ba022428f928aae72d6af2c3244e2514442bccae2e22440e3eeb7f1c198
              start: "2026-10-19T01:07:55.710499Z"
              end: "2026-10-19T01:13:44.838921Z"
              queue_seconds: 349.128422
              raw_seconds: 349.128422
              exact: true
        protected:
            - issue: "437"
              slot: ariadne:1
              start: "2026-10-18T10:57:47-07:00"
              end: "2026-10-18T18:18:21.221443Z"
            - issue: "437"
              slot: ariadne:1
              start: "2026-10-18T18:02:23.518Z"
              end: "2026-10-18T18:04:14.653Z"
            - issue: "437"
              slot: ariadne:1
              start: "2026-10-18T18:21:08.81741Z"
              end: "2026-10-18T18:23:20.078884Z"
            - issue: "437"
              slot: ariadne:1
              start: "2026-10-18T18:58:32.709201Z"
              end: "2026-10-18T12:10:24-07:00"
            - issue: "437"
              slot: ariadne:1
              start: "2026-10-18T12:10:24-07:00"
              end: "2026-10-18T19:31:16.893432Z"
            - issue: "437"
              slot: ariadne:1
              start: "2026-10-18T20:18:20.731582Z"
              end: "2026-10-18T20:36:52.829415Z"
            - issue: "437"
              slot: ariadne:1
              start: "2026-10-18T20:50:46.737447Z"
              end: "2026-10-18T21:00:12.19Z"
            - issue: "444"
              slot: ariadne:1
              start: "2026-10-18T21:37:59.877Z"
              end: "2026-10-18T21:46:29.803654Z"
            - issue: "444"
              slot: ariadne:1
              start: "2026-10-18T22:08:43.529371Z"
              end: "2026-10-18T15:18:56-07:00"
            - issue: "444"
              slot: ariadne:1
              start: "2026-10-18T15:18:56-07:00"
              end: "2026-10-18T22:21:52.006Z"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.11616666666666667
        cuts:
            - dispatched: "2026-10-18T17:38:10-07:00"
              finalized: "2026-10-18T17:42:03-07:00"
              finalizing: false
            - dispatched: "2026-10-18T18:26:06-07:00"
              finalized: "2026-10-18T18:32:20-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-18T19:30:02-07:00"
      dispatched: "2026-10-18T19:24:00-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: The excluded-brain pass clears ProtectedIntervals; its provenance regression passes.
          round: 3
        - id: BR-2
          disposition: addressed
          note: creditedIntervals intersects positive allocations with counted Build/Converge phases; Tail and boundary-crossing regressions pass.
          round: 3
        - id: BR-3
          disposition: addressed
          note: Proven different-slot ownership skips missing replay authority, including the no-completion legacy fixture; regressions pass.
          round: 3
        - id: BR-4
          disposition: addressed
          note: The four/eight unrelated-adoption fixture passes without replaying unavailable sources while preserving same-slot protection.
          round: 3
        - id: BR-5
          disposition: not-addressed
          note: 'Empty, absent, and mixed ownership are covered, but actualexclusion.go:154–177 treats malformed nonempty workspace identity as proven different. In a disposable pinned-head checkout, changing only #649''s workspace to ariadne:06 makes #636 measure 0.333333h with no exclusions. A refusal regression fails on head and passes after requiring canonical parsed slot identity. This remains the same overlap-relevance-before-evidence rule failure.'
          round: 3
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#443 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T17:42:03-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `measurement-population-isolation` (kind: defect) Clear protected intervals from the excluded-brain diagnostic pass
  cmd/sdlc/actualmeasurement.go:38 adds ProtectedIntervals to options, but the diagnostic copy at line 98 retains them while clearing SourceSlots. An overlapping unowned brain session therefore aborts a valid candidate with "excluded brain provenance: ... without unique slot ownership". A temporary overlay of TestOwnedActualEngineReplayAndExcludedBrainProvenance reproduced this with five protected minutes. Clear ProtectedIntervals alongside QueueIntervals in the diagnostic population and add a combined regression.
- **BR-2** [Critical] `adopted-credit-authority` (kind: defect) Do not protect verification-tail time omitted from adopted actuals
  cmd/sdlc/actualexclusion.go:206 converts all replayed segments into protection, although actualmeasurement.go:69 adopts only Build+Converge and excludes Tail. An overlay of TestActualQueueMeasurementReplay demonstrated protection of 00:50–01:00 while the result explicitly reported those ten minutes as uncounted tail. This can silently remove another issue's valid activity. ARCH-PURPOSE: derive protected intervals from the same counted phases that determine adopted hours; cover wholly-tail runs and runs crossing phase boundaries.
- **BR-3** [Critical] `overlap-relevance-before-evidence` (kind: defect) Rule out unrelated slots before refusing missing adoption evidence
  cmd/sdlc/actualexclusion.go:153 requires a saved measurement window before checking same-slot engagement at line 182. A temporally plausible legacy adoption without ActualInputs blocks candidates even when its engagement belongs exclusively to another slot. An overlay of the parked-claim fixture, using candidate slot ariadne:7 versus adopted slot ariadne:6 and omitting the saved inputs, reproduced "issue #649 adopted actual has no trustworthy saved window". Use trustworthy ownership history to rule out unrelated slots before refusing missing pins; retain refusal when same-slot overlap remains possible. ARCH-PURPOSE.
- **BR-4** [Important] `declared-workload-verification` (kind: defect) Deliver the planned unrelated-card scaling regression
  workshop/plans/000443-sdlc-actual-a-parked-claim-double-counts-time-the-slot-spent-on-other-issues-plan.md:37 claims tests count replays for N and 2N unrelated cards, but no such test exists in the reviewed changes. ARCH-CONSTRAINTS: add representative loader coverage proving unrelated cards do not trigger transcript replay, including other-slot and disjoint-window cases, or revise the unsupported verification claim and supply equivalent measured evidence.

## Round 2 — 2026-10-18T18:32:20-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — The diagnostic pass clears ProtectedIntervals. Removing that line makes TestOwnedActualEngineReplayAndExcludedBrainProvenance fail with the original ownership error.
- BR-2 — addressed — Protection intersects the same counted phases used for adopted hours. Restoring full-segment protection fails both TestCreditedIntervalsProtectOnlyCountedPhases and TestActualQueueMeasurementReplay, covering wholly-Tail and boundary-crossing runs.
- BR-3 — addressed — Proven other-slot ownership is checked before missing-window refusal. Disabling differentSlot makes the parked-claim regression fail on legacy issue #650 without saved inputs.
- BR-4 — addressed — The parked-claim fixture now checks four and eight unrelated adoptions while retaining same-slot protection, plus a disjoint-window case after transcript loss. These pass; disabling unrelated-slot rejection makes the four-card assertion fail.

### Raised

- **BR-5** [Critical] `overlap-relevance-before-evidence` (kind: defect) Refuse overlapping adopted actuals whose slot relationship remains unknown
  cmd/sdlc/actualexclusion.go:232–243 treats absence of a proven same-slot engagement as proof of irrelevance. Legacy owners may have an empty Workspace: differentSlot correctly declines to prove them unrelated, but the subsequent loop finds no match and silently continues. In a disposable copy of the parked-claim fixture, clearing only issue #649's claimant workspace produced a measured 0.333333h for #636 with no exclusions, recounting #649's adopted interval. An explicit refusal regression fails on head and passes with a temporary unknown-slot guard. ARCH-PURPOSE, ARCH-SECURE: this is the 2nd finding in family overlap-relevance-before-evidence. Earlier rounds fixed instances; do NOT fix only this instance. State and enforce the rule that skipping requires affirmative evidence of disjoint time or slot, enumerate unknown/empty/mixed ownership cases, and sweep them with regressions.

## Round 3 — 2026-10-18T19:30:02-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — The excluded-brain pass clears ProtectedIntervals; its provenance regression passes.
- BR-2 — addressed — creditedIntervals intersects positive allocations with counted Build/Converge phases; Tail and boundary-crossing regressions pass.
- BR-3 — addressed — Proven different-slot ownership skips missing replay authority, including the no-completion legacy fixture; regressions pass.
- BR-4 — addressed — The four/eight unrelated-adoption fixture passes without replaying unavailable sources while preserving same-slot protection.
- BR-5 — not-addressed — Empty, absent, and mixed ownership are covered, but actualexclusion.go:154–177 treats malformed nonempty workspace identity as proven different. In a disposable pinned-head checkout, changing only #649's workspace to ariadne:06 makes #636 measure 0.333333h with no exclusions. A refusal regression fails on head and passes after requiring canonical parsed slot identity. This remains the same overlap-relevance-before-evidence rule failure.

## Open findings

- **BR-5** [Critical] `overlap-relevance-before-evidence` (kind: defect) Refuse overlapping adopted actuals whose slot relationship remains unknown
