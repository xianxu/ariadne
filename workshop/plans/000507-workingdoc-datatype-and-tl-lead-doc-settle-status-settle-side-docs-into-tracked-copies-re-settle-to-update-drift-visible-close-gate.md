---
gate: boundary-review
issue: 507
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: ab9c7877-c39d-4050-9048-73e8c522bff3
            repository: github.com/xianxu/ariadne
            issue_id: "000507"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/6c9313a65937d49f
            reviewed_head: f8d5927d1879cdae1f12ae63c0f402a30db3a7ef
            slot: ariadne:4
            requested_at: "2026-10-21T20:59:23.610973Z"
          phase: finished
          pid: 70818
          started_at: "2026-10-21T20:59:23.727838Z"
          finished_at: "2026-10-21T21:02:53.636432Z"
          session_id: 01a16042-f81e-75d2-819e-85b7984b7adc
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: f8d5927d1879cdae1f12ae63c0f402a30db3a7ef
        tracker_head: 13432893d98a4059d567961eaad962202d499dd3
        branch_point: 510aab6ff3a55bfdf5472e45eedf5512f6d04420
        since: "2026-10-21T12:19:30-07:00"
        until: "2026-10-21T13:59:02-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T12:19:30-07:00"
              end: "2026-10-21T13:59:02-07:00"
              began: "2026-10-21T12:19:30-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: codex
              id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 395f021cd7d434c7c1221be02fd22ee38651ac37dc13405cf4a55d10e5001102
              started: "2026-10-20T16:55:54.591Z"
              spawned: false
              bytes: 20265581
              windows:
                - start: "2026-10-21T12:19:30-07:00"
                  end: "2026-10-21T13:59:02-07:00"
            - harness: codex
              id: 01a15fe8-0504-7302-b73c-8c46a42b6d88
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: e5d66c8928eb8b725694a21e03b582e965eba489a09c38951305c9757465d5d8
              started: "2026-10-21T19:20:03.64Z"
              spawned: true
              bytes: 585465
              windows:
                - start: "2026-10-21T19:20:03.64Z"
                  end: "2026-10-21T13:59:02-07:00"
            - harness: codex
              id: 01a15fed-e412-7912-9c62-dafe05658042
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 3d83601a152c04161f3f95899757396577f495302789a64bb82fd130c6f3e3d1
              started: "2026-10-21T19:26:29.036Z"
              spawned: false
              bytes: 620770
              windows:
                - start: "2026-10-21T19:26:29.036Z"
                  end: "2026-10-21T13:59:02-07:00"
            - harness: codex
              id: 01a15ff0-7f50-7f21-ad10-ff631fbe06ef
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: b16a15a4e0b832df1ed87f5c9616f207f4270ddec6cab04ae1c6854427e45517
              started: "2026-10-21T19:29:19.567Z"
              spawned: false
              bytes: 403448
              windows:
                - start: "2026-10-21T19:29:19.567Z"
                  end: "2026-10-21T13:59:02-07:00"
            - harness: codex
              id: 01a15ff2-206e-7f00-bcd0-f8fd39e9fdd5
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 07bad111f94eb1661bfd56e53d592df95a77a87e9955e13ee0ece0567d0dc5d0
              started: "2026-10-21T19:31:06.27Z"
              spawned: false
              bytes: 387170
              windows:
                - start: "2026-10-21T19:31:06.27Z"
                  end: "2026-10-21T13:59:02-07:00"
            - harness: codex
              id: 01a15ff2-8c6c-75f2-954f-6e999ca17b20
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: bee1381e9c0091735897b2c87d5e7b142bfcb1958500b03119e82c1dedd64d7c
              started: "2026-10-21T19:31:33.916Z"
              spawned: false
              bytes: 412032
              windows:
                - start: "2026-10-21T19:31:33.916Z"
                  end: "2026-10-21T13:59:02-07:00"
            - harness: codex
              id: 01a15ff3-4d1e-7b13-8bba-b11e1b274a62
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 8472ff0e10a25643d9b13c3044204e09d762ed72e0d02d92c4018604bfea68eb
              started: "2026-10-21T19:32:23.118Z"
              spawned: true
              bytes: 1836074
              windows:
                - start: "2026-10-21T19:32:23.118Z"
                  end: "2026-10-21T13:59:02-07:00"
            - harness: codex
              id: 01a15ff3-7af8-7342-a64f-b5e7cccd8b22
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: da23c55942c17fa4436ef46e12232b4fb55a8eb36e8f43c14e3f19319e571c88
              started: "2026-10-21T19:32:34.874Z"
              spawned: true
              bytes: 459313
              windows:
                - start: "2026-10-21T19:32:34.874Z"
                  end: "2026-10-21T13:59:02-07:00"
            - harness: codex
              id: 01a15ff6-385e-7cf3-9e8c-12335805d338
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: a90e455795e152ae83c98629823ba05178739baf1e1306b9bf81b05e8d301835
              started: "2026-10-21T19:35:34.456Z"
              spawned: true
              bytes: 937404
              windows:
                - start: "2026-10-21T19:35:34.456Z"
                  end: "2026-10-21T13:59:02-07:00"
        queue_waits:
            - attempt_id: 3bc05ff279d4794c4a8f6bb176e55c71
              slot: ariadne:4
              record_digest: 826068f401579bb474474655ad43cbb732ead13f8f46ba7d288c98f60dbb2c6e
              start: "2026-10-21T20:03:29.668881Z"
              end: "2026-10-21T20:16:17.372475Z"
              queue_seconds: 767.703594
              raw_seconds: 767.703594
              exact: true
      "n": 1
      timestamp: "2026-10-21T14:02:54-07:00"
      dispatched: "2026-10-21T13:59:23-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: BR-1
          severity: Critical
          title: Allowed destination paths can produce an unreadable settlement pointer
          detail: 'cmd/tl-lead/doc_model.go:113 writes paths directly into a plain YAML scalar, but workingdocPath accepts "workshop/workingdoc/a #b.md". YAML treats "#b.md @ COMMIT" as a comment, so settlement can publish, stamp, and report success while subsequent status and settle reject the source. Validate generated metadata with the production parser before effects and refuse unrepresentable paths; add round-trip regressions. ARCH-SECURE, ARCH-EVOLVE.'
          family: managed-metadata-roundtrip
          round: 1
        - id: BR-2
          severity: Critical
          title: First product settlement can overwrite an unrelated committed destination
          detail: cmd/tl-lead/doc_settle.go:349 accepts an occupied local target matching the payload and publishes with no expected commit. If committed main and the branch baseline contain different content, cmd/sdlc/trunkpublish.go:283 permits replacing it through ordinary edit semantics. A locally deleted tracked target also bypasses the collision check. Enforce absent-or-identical remote content for first settlement inside each CAS attempt, preserving ordinary trunk behavior; add product regressions. ARCH-PURPOSE.
          family: first-settlement-destination-ownership
          round: 1
        - id: BR-3
          severity: Important
          title: Claimed coordinator recovery sequences are missing from the tests
          detail: workshop/plans/000507-workingdoc-settle-plan.md:120 marks uncertain Apply and late-lock sequences delivered, but doc_fake_test.go:31 cannot return a lost acknowledgment after applying, and doc_settle_test.go only injects a source edit after publication. Add controlled Apply outcomes and late-lock/retry sequences through settleWorkingdoc, asserting confirmed pointer evidence and no duplicate publication. ARCH-ORDER.
          family: settlement-interruption-coverage
          round: 1
        - id: BR-4
          severity: Important
          title: Ops publication tests require an installed host sdlc binary
          detail: doc_operating_test.go:28 does not provide a controlled sdlc agent-id dependency for doc_operating.go's operatingAgentTrailer. Running PATH=/usr/bin:/bin /opt/homebrew/bin/go test ./cmd/tl-lead -run '^TestOperatingPublishExactPayloadAndRetry$' -count=1 fails with executable "sdlc" not found. Inject or fixture the identity dependency through the production seam so portable tests do not depend on an unrelated installed revision. ARCH-MOCK.
          family: integration-test-dependency-isolation
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: af5eed74-e535-412b-9833-04db67d9f721
            repository: github.com/xianxu/ariadne
            issue_id: "000507"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/5bfd30daaf9753de
            reviewed_head: 925290911ba2a36408fdbf416ee06922ad5eee0a
            slot: ariadne:4
            requested_at: "2026-10-22T01:59:43.319496Z"
          phase: finished
          pid: 16075
          started_at: "2026-10-22T01:59:43.479128Z"
          finished_at: "2026-10-22T02:04:40.239309Z"
          session_id: 01a16155-ee04-7892-ba23-4bea4502c786
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 925290911ba2a36408fdbf416ee06922ad5eee0a
        tracker_head: 283ef68f0f6131b92fa9af0fe14b44fe5d678724
        branch_point: 510aab6ff3a55bfdf5472e45eedf5512f6d04420
        since: "2026-10-21T12:19:30-07:00"
        until: "2026-10-21T18:59:09-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T12:19:30-07:00"
              end: "2026-10-21T18:59:09-07:00"
              began: "2026-10-21T12:19:30-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: codex
              id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: eafcb65e59d2ebbdf8d1292d0dd2a938460cec634843cbc755ffd8a79f677c4f
              started: "2026-10-20T16:55:54.591Z"
              spawned: false
              bytes: 21242726
              windows:
                - start: "2026-10-21T12:19:30-07:00"
                  end: "2026-10-21T18:59:09-07:00"
            - harness: codex
              id: 01a15fe8-0504-7302-b73c-8c46a42b6d88
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: e5d66c8928eb8b725694a21e03b582e965eba489a09c38951305c9757465d5d8
              started: "2026-10-21T19:20:03.64Z"
              spawned: true
              bytes: 585465
              windows:
                - start: "2026-10-21T19:20:03.64Z"
                  end: "2026-10-21T18:59:09-07:00"
            - harness: codex
              id: 01a15fed-e412-7912-9c62-dafe05658042
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 3d83601a152c04161f3f95899757396577f495302789a64bb82fd130c6f3e3d1
              started: "2026-10-21T19:26:29.036Z"
              spawned: false
              bytes: 620770
              windows:
                - start: "2026-10-21T19:26:29.036Z"
                  end: "2026-10-21T18:59:09-07:00"
            - harness: codex
              id: 01a15ff0-7f50-7f21-ad10-ff631fbe06ef
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: b16a15a4e0b832df1ed87f5c9616f207f4270ddec6cab04ae1c6854427e45517
              started: "2026-10-21T19:29:19.567Z"
              spawned: false
              bytes: 403448
              windows:
                - start: "2026-10-21T19:29:19.567Z"
                  end: "2026-10-21T18:59:09-07:00"
            - harness: codex
              id: 01a15ff2-206e-7f00-bcd0-f8fd39e9fdd5
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 07bad111f94eb1661bfd56e53d592df95a77a87e9955e13ee0ece0567d0dc5d0
              started: "2026-10-21T19:31:06.27Z"
              spawned: false
              bytes: 387170
              windows:
                - start: "2026-10-21T19:31:06.27Z"
                  end: "2026-10-21T18:59:09-07:00"
            - harness: codex
              id: 01a15ff2-8c6c-75f2-954f-6e999ca17b20
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: bee1381e9c0091735897b2c87d5e7b142bfcb1958500b03119e82c1dedd64d7c
              started: "2026-10-21T19:31:33.916Z"
              spawned: false
              bytes: 412032
              windows:
                - start: "2026-10-21T19:31:33.916Z"
                  end: "2026-10-21T18:59:09-07:00"
            - harness: codex
              id: 01a15ff3-4d1e-7b13-8bba-b11e1b274a62
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 8472ff0e10a25643d9b13c3044204e09d762ed72e0d02d92c4018604bfea68eb
              started: "2026-10-21T19:32:23.118Z"
              spawned: true
              bytes: 1836074
              windows:
                - start: "2026-10-21T19:32:23.118Z"
                  end: "2026-10-21T18:59:09-07:00"
            - harness: codex
              id: 01a15ff3-7af8-7342-a64f-b5e7cccd8b22
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: da23c55942c17fa4436ef46e12232b4fb55a8eb36e8f43c14e3f19319e571c88
              started: "2026-10-21T19:32:34.874Z"
              spawned: true
              bytes: 459313
              windows:
                - start: "2026-10-21T19:32:34.874Z"
                  end: "2026-10-21T18:59:09-07:00"
            - harness: codex
              id: 01a15ff6-385e-7cf3-9e8c-12335805d338
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: a90e455795e152ae83c98629823ba05178739baf1e1306b9bf81b05e8d301835
              started: "2026-10-21T19:35:34.456Z"
              spawned: true
              bytes: 937404
              windows:
                - start: "2026-10-21T19:35:34.456Z"
                  end: "2026-10-21T18:59:09-07:00"
            - harness: codex
              id: 01a160cd-43fe-7493-beca-70aa3d5c327b
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 8e83ff34310b3c76f8766a6d5e81568e97da24c1f6223db90b213f8fd60efa3f
              started: "2026-10-21T23:30:27.524Z"
              spawned: true
              bytes: 446256
              windows:
                - start: "2026-10-21T23:30:27.524Z"
                  end: "2026-10-21T18:59:09-07:00"
            - harness: codex
              id: 01a160cd-55f6-7062-a9a7-e0ad8a26aab0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 29cd538d8f11948d340cb0c482d789dde29aea4a0c3c0be2beed41baef8e5249
              started: "2026-10-21T23:30:32.135Z"
              spawned: true
              bytes: 482983
              windows:
                - start: "2026-10-21T23:30:32.135Z"
                  end: "2026-10-21T18:59:09-07:00"
            - harness: codex
              id: 01a160cd-6859-7580-b835-ba9b6504cd61
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 5e4a02068103f7d49fc266535e21799b15d804a28cd43e74bd0b2edd52664bcf
              started: "2026-10-21T23:30:36.9Z"
              spawned: true
              bytes: 503005
              windows:
                - start: "2026-10-21T23:30:36.9Z"
                  end: "2026-10-21T18:59:09-07:00"
        queue_waits:
            - attempt_id: 3bc05ff279d4794c4a8f6bb176e55c71
              slot: ariadne:4
              record_digest: 826068f401579bb474474655ad43cbb732ead13f8f46ba7d288c98f60dbb2c6e
              start: "2026-10-21T20:03:29.668881Z"
              end: "2026-10-21T20:16:17.372475Z"
              queue_seconds: 767.703594
              raw_seconds: 767.703594
              exact: true
            - attempt_id: 339841fb96fa15c43e6a69f42180754a
              slot: ariadne:4
              record_digest: 108ed183e363a5edb12ef960541d7c735b9549753fd5b14a2a9be137140b604f
              start: "2026-10-21T23:35:29.719462Z"
              end: "2026-10-22T00:32:25.75598Z"
              queue_seconds: 3416.036518
              raw_seconds: 3416.036518
              exact: true
        cuts:
            - dispatched: "2026-10-21T13:59:23-07:00"
              finalized: "2026-10-21T14:02:54-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-21T19:04:41-07:00"
      dispatched: "2026-10-21T18:59:43-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: doc_model.go:120 reparses generated metadata before effects. Adversarial pointer and coordinator refusal tests pass; removing the roundtrip guard in the disposable checkout makes them fail.
          round: 2
        - id: BR-2
          disposition: addressed
          note: doc_repository.go supplies --expect-absent for first settlement, and trunkpublish.go checks ownership inside every CAS callback. Real-remote ownership tests pass; disabling that predicate reproduces successful replacement of unrelated committed content.
          round: 2
        - id: BR-3
          disposition: addressed
          note: doc_fake_test.go now supports pre-write failure and post-write lost acknowledgment. Coordinator tests exercise late locks, unavailable readback, intervening edits and duplicate-free retries. Disabling uncertain-Apply reconciliation makes the lost-ack regression fail.
          round: 2
        - id: BR-4
          disposition: addressed
          note: doc_operating_test.go:28 supplies a strict fixture sdlc agent-id executable. All operating tests pass with PATH=/usr/bin:/bin; removing the fixture PATH injection reproduces the missing-sdlc failure.
          round: 2
      findings:
        - kind: defect
          id: BR-5
          severity: Critical
          title: Backticks across Markdown blocks hide unresolved review items
          detail: 'cmd/tl-lead/doc_markers.go:171 searches for matching backticks across every non-fenced region, and line 206 marks the intervening blocks literal. Unmatched backticks in separate paragraphs, headings or list items therefore hide an intervening human question. Three coordinator probes confirmed zero actual CodeSpan nodes using the repository''s Goldmark dependency, yet settleWorkingdoc returned success with Published=true and Stamped=true. ARCH-PURPOSE, ARCH-SECURE: restrict literal masking to actual Markdown code contexts and add scanner plus no-effects coordinator regressions across block boundaries, preserving legitimate multiline spans.'
          family: marker-literal-boundaries
          round: 2
        - kind: defect
          id: BR-6
          severity: Important
          title: Settlement transitions observe effects instead of authorizing them
          detail: 'cmd/tl-lead/doc_settle.go:359 publishes before consulting transitionSettlement, and line 395 applies the source stamp before its transition. doc_model.go:153 returns no effects, while failure and unconfirmed events have no production consumers. This contradicts the plan''s line 49 claim that only the transition model grants the next effect. ARCH-ORDER: make the pure model authorize effects and consume their outcomes, route failures through it, and test forbidden effect sequences through the production coordinator.'
          family: state-transition-authority
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 7a6bb200-38d0-4645-b21d-b37d87f6e6dc
            repository: github.com/xianxu/ariadne
            issue_id: "000507"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/32605a8eb9a9ff96
            reviewed_head: 98e25bfe17f3fe372fc59d10c5a40fdeb871a4b4
            slot: ariadne:4
            requested_at: "2026-10-22T04:00:03.646849Z"
          phase: finished
          pid: 47681
          started_at: "2026-10-22T04:00:03.778922Z"
          finished_at: "2026-10-22T04:03:18.813413Z"
          session_id: 01a161c4-1a0d-7001-ba85-c0de1ca5673b
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 98e25bfe17f3fe372fc59d10c5a40fdeb871a4b4
        tracker_head: 50c98718de0ef5d33937d665512dae400391af08
        branch_point: 510aab6ff3a55bfdf5472e45eedf5512f6d04420
        since: "2026-10-21T12:19:30-07:00"
        until: "2026-10-21T20:59:32-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T12:19:30-07:00"
              end: "2026-10-21T20:59:32-07:00"
              began: "2026-10-21T12:19:30-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: codex
              id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 9de53925dd644380686876f3296c201c4a696e7a576bbff9f2f0f9579fb075c5
              started: "2026-10-20T16:55:54.591Z"
              spawned: false
              bytes: 21726742
              windows:
                - start: "2026-10-21T12:19:30-07:00"
                  end: "2026-10-21T20:59:32-07:00"
            - harness: codex
              id: 01a15fe8-0504-7302-b73c-8c46a42b6d88
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: e5d66c8928eb8b725694a21e03b582e965eba489a09c38951305c9757465d5d8
              started: "2026-10-21T19:20:03.64Z"
              spawned: true
              bytes: 585465
              windows:
                - start: "2026-10-21T19:20:03.64Z"
                  end: "2026-10-21T20:59:32-07:00"
            - harness: codex
              id: 01a15fed-e412-7912-9c62-dafe05658042
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 3d83601a152c04161f3f95899757396577f495302789a64bb82fd130c6f3e3d1
              started: "2026-10-21T19:26:29.036Z"
              spawned: false
              bytes: 620770
              windows:
                - start: "2026-10-21T19:26:29.036Z"
                  end: "2026-10-21T20:59:32-07:00"
            - harness: codex
              id: 01a15ff0-7f50-7f21-ad10-ff631fbe06ef
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: b16a15a4e0b832df1ed87f5c9616f207f4270ddec6cab04ae1c6854427e45517
              started: "2026-10-21T19:29:19.567Z"
              spawned: false
              bytes: 403448
              windows:
                - start: "2026-10-21T19:29:19.567Z"
                  end: "2026-10-21T20:59:32-07:00"
            - harness: codex
              id: 01a15ff2-206e-7f00-bcd0-f8fd39e9fdd5
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 07bad111f94eb1661bfd56e53d592df95a77a87e9955e13ee0ece0567d0dc5d0
              started: "2026-10-21T19:31:06.27Z"
              spawned: false
              bytes: 387170
              windows:
                - start: "2026-10-21T19:31:06.27Z"
                  end: "2026-10-21T20:59:32-07:00"
            - harness: codex
              id: 01a15ff2-8c6c-75f2-954f-6e999ca17b20
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: bee1381e9c0091735897b2c87d5e7b142bfcb1958500b03119e82c1dedd64d7c
              started: "2026-10-21T19:31:33.916Z"
              spawned: false
              bytes: 412032
              windows:
                - start: "2026-10-21T19:31:33.916Z"
                  end: "2026-10-21T20:59:32-07:00"
            - harness: codex
              id: 01a15ff3-4d1e-7b13-8bba-b11e1b274a62
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 8472ff0e10a25643d9b13c3044204e09d762ed72e0d02d92c4018604bfea68eb
              started: "2026-10-21T19:32:23.118Z"
              spawned: true
              bytes: 1836074
              windows:
                - start: "2026-10-21T19:32:23.118Z"
                  end: "2026-10-21T20:59:32-07:00"
            - harness: codex
              id: 01a15ff3-7af8-7342-a64f-b5e7cccd8b22
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: da23c55942c17fa4436ef46e12232b4fb55a8eb36e8f43c14e3f19319e571c88
              started: "2026-10-21T19:32:34.874Z"
              spawned: true
              bytes: 459313
              windows:
                - start: "2026-10-21T19:32:34.874Z"
                  end: "2026-10-21T20:59:32-07:00"
            - harness: codex
              id: 01a15ff6-385e-7cf3-9e8c-12335805d338
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: a90e455795e152ae83c98629823ba05178739baf1e1306b9bf81b05e8d301835
              started: "2026-10-21T19:35:34.456Z"
              spawned: true
              bytes: 937404
              windows:
                - start: "2026-10-21T19:35:34.456Z"
                  end: "2026-10-21T20:59:32-07:00"
            - harness: codex
              id: 01a160cd-43fe-7493-beca-70aa3d5c327b
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 9eca989ea9440dcd7886dfe89c4872d332ae721e4654f8f12f9e5eb43fac8b54
              started: "2026-10-21T23:30:27.524Z"
              spawned: true
              bytes: 689403
              windows:
                - start: "2026-10-21T23:30:27.524Z"
                  end: "2026-10-21T20:59:32-07:00"
            - harness: codex
              id: 01a160cd-55f6-7062-a9a7-e0ad8a26aab0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 29cd538d8f11948d340cb0c482d789dde29aea4a0c3c0be2beed41baef8e5249
              started: "2026-10-21T23:30:32.135Z"
              spawned: true
              bytes: 482983
              windows:
                - start: "2026-10-21T23:30:32.135Z"
                  end: "2026-10-21T20:59:32-07:00"
            - harness: codex
              id: 01a160cd-6859-7580-b835-ba9b6504cd61
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 5e4a02068103f7d49fc266535e21799b15d804a28cd43e74bd0b2edd52664bcf
              started: "2026-10-21T23:30:36.9Z"
              spawned: true
              bytes: 503005
              windows:
                - start: "2026-10-21T23:30:36.9Z"
                  end: "2026-10-21T20:59:32-07:00"
        queue_waits:
            - attempt_id: 3bc05ff279d4794c4a8f6bb176e55c71
              slot: ariadne:4
              record_digest: 826068f401579bb474474655ad43cbb732ead13f8f46ba7d288c98f60dbb2c6e
              start: "2026-10-21T20:03:29.668881Z"
              end: "2026-10-21T20:16:17.372475Z"
              queue_seconds: 767.703594
              raw_seconds: 767.703594
              exact: true
            - attempt_id: 339841fb96fa15c43e6a69f42180754a
              slot: ariadne:4
              record_digest: 108ed183e363a5edb12ef960541d7c735b9549753fd5b14a2a9be137140b604f
              start: "2026-10-21T23:35:29.719462Z"
              end: "2026-10-22T00:32:25.75598Z"
              queue_seconds: 3416.036518
              raw_seconds: 3416.036518
              exact: true
            - attempt_id: ed2ea51d86f781c65c1be8eae1c4a383
              slot: ariadne:4
              record_digest: 24822d415bb3ef10c2a42cf0d7d2b4083131ce028a7d9c5365d1e741e48f857a
              start: "2026-10-22T03:03:15.261174Z"
              end: "2026-10-22T03:31:04.998823Z"
              queue_seconds: 1669.737649
              raw_seconds: 1669.737649
              exact: true
        cuts:
            - dispatched: "2026-10-21T13:59:23-07:00"
              finalized: "2026-10-21T14:02:54-07:00"
              finalizing: false
            - dispatched: "2026-10-21T18:59:43-07:00"
              finalized: "2026-10-21T19:04:41-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-21T21:03:19-07:00"
      dispatched: "2026-10-21T21:00:03-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Pointer generation reparses its output and checks exact equality; adversarial path and byte-preservation regressions pass.
          round: 3
        - id: BR-2
          disposition: addressed
          note: First settlement supplies --expect-absent, enforced inside every publication CAS callback; ownership regressions pass.
          round: 3
        - id: BR-3
          disposition: addressed
          note: Production coordinator tests cover lost acknowledgments, pre-write failure, late locks, unavailable readback, source edits and retries; the full package passes.
          round: 3
        - id: BR-4
          disposition: addressed
          note: Operating fixtures provide an isolated sdlc agent-id executable and assert its trailer; operating publication tests pass.
          round: 3
        - id: BR-5
          disposition: addressed
          note: Literal masks derive from actual Goldmark code nodes. Restoring the previous scanner makes all four block-boundary scanner and no-effects coordinator regressions fail.
          round: 3
        - id: BR-6
          disposition: addressed
          note: Publication and stamping require model permissions before execution, and failure/unconfirmed outcomes retire permissions. Removing the production authorization guard makes both forbidden-effect regressions fail.
          round: 3
      findings:
        - kind: defect
          id: BR-7
          severity: Critical
          title: Settlement publishes rewritten code examples
          detail: 'cmd/tl-lead/doc_project.go:28 delegates to sidedoc.RebaseLinks, whose line-based literal detection misses multiline code spans, indented code, and fences nested in lists or blockquotes. Disposable-checkout probes through projectWorkingdoc and the production settleWorkingdoc coordinator reproduced all four cases: literal [example](../README.md) became [example](../../README.md), with Published=true and Stamped=true. This violates the plan''s line 29 preservation contract (ARCH-PURPOSE); the scanner and rebaser also apply inconsistent Markdown literal rules (ARCH-DRY). Establish one shared rule for actual Markdown code contexts and sweep all projection rewrites, with pure preservation tests and coordinator regressions covering these four contexts plus ordinary-link controls.'
          family: markdown-literal-preservation
          round: 3
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: feafd331-8c06-4d6e-b9f4-03f55b893387
            repository: github.com/xianxu/ariadne
            issue_id: "000507"
            recipe: milestone-review
            round: 4
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/3de02616dd1bd58c
            reviewed_head: 0d41e87788e51ba8bd7af58c052bf11042500ebc
            slot: ariadne:4
            requested_at: "2026-10-22T10:00:03.954647Z"
          phase: finished
          pid: 8277
          started_at: "2026-10-22T10:00:04.101846Z"
          finished_at: "2026-10-22T10:02:41.047908Z"
          session_id: 01a1630d-b27b-7182-934f-371541b51ff9
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 0d41e87788e51ba8bd7af58c052bf11042500ebc
        tracker_head: 502cad20bc4d39514fc43f46c4817ce7c09e267e
        branch_point: 510aab6ff3a55bfdf5472e45eedf5512f6d04420
        since: "2026-10-21T12:19:30-07:00"
        until: "2026-10-22T02:59:31-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T12:19:30-07:00"
              end: "2026-10-22T02:59:31-07:00"
              began: "2026-10-21T12:19:30-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: codex
              id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 1a3bcc2af1e42721edcf0416c6bd1caf17eaa1575e1f227f28587637f7811b44
              started: "2026-10-20T16:55:54.591Z"
              spawned: false
              bytes: 22284162
              windows:
                - start: "2026-10-21T12:19:30-07:00"
                  end: "2026-10-22T02:59:31-07:00"
            - harness: codex
              id: 01a15fe8-0504-7302-b73c-8c46a42b6d88
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: e5d66c8928eb8b725694a21e03b582e965eba489a09c38951305c9757465d5d8
              started: "2026-10-21T19:20:03.64Z"
              spawned: true
              bytes: 585465
              windows:
                - start: "2026-10-21T19:20:03.64Z"
                  end: "2026-10-22T02:59:31-07:00"
            - harness: codex
              id: 01a15fed-e412-7912-9c62-dafe05658042
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 3d83601a152c04161f3f95899757396577f495302789a64bb82fd130c6f3e3d1
              started: "2026-10-21T19:26:29.036Z"
              spawned: false
              bytes: 620770
              windows:
                - start: "2026-10-21T19:26:29.036Z"
                  end: "2026-10-22T02:59:31-07:00"
            - harness: codex
              id: 01a15ff0-7f50-7f21-ad10-ff631fbe06ef
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: b16a15a4e0b832df1ed87f5c9616f207f4270ddec6cab04ae1c6854427e45517
              started: "2026-10-21T19:29:19.567Z"
              spawned: false
              bytes: 403448
              windows:
                - start: "2026-10-21T19:29:19.567Z"
                  end: "2026-10-22T02:59:31-07:00"
            - harness: codex
              id: 01a15ff2-206e-7f00-bcd0-f8fd39e9fdd5
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 07bad111f94eb1661bfd56e53d592df95a77a87e9955e13ee0ece0567d0dc5d0
              started: "2026-10-21T19:31:06.27Z"
              spawned: false
              bytes: 387170
              windows:
                - start: "2026-10-21T19:31:06.27Z"
                  end: "2026-10-22T02:59:31-07:00"
            - harness: codex
              id: 01a15ff2-8c6c-75f2-954f-6e999ca17b20
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: bee1381e9c0091735897b2c87d5e7b142bfcb1958500b03119e82c1dedd64d7c
              started: "2026-10-21T19:31:33.916Z"
              spawned: false
              bytes: 412032
              windows:
                - start: "2026-10-21T19:31:33.916Z"
                  end: "2026-10-22T02:59:31-07:00"
            - harness: codex
              id: 01a15ff3-4d1e-7b13-8bba-b11e1b274a62
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 8472ff0e10a25643d9b13c3044204e09d762ed72e0d02d92c4018604bfea68eb
              started: "2026-10-21T19:32:23.118Z"
              spawned: true
              bytes: 1836074
              windows:
                - start: "2026-10-21T19:32:23.118Z"
                  end: "2026-10-22T02:59:31-07:00"
            - harness: codex
              id: 01a15ff3-7af8-7342-a64f-b5e7cccd8b22
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: da23c55942c17fa4436ef46e12232b4fb55a8eb36e8f43c14e3f19319e571c88
              started: "2026-10-21T19:32:34.874Z"
              spawned: true
              bytes: 459313
              windows:
                - start: "2026-10-21T19:32:34.874Z"
                  end: "2026-10-22T02:59:31-07:00"
            - harness: codex
              id: 01a15ff6-385e-7cf3-9e8c-12335805d338
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: a90e455795e152ae83c98629823ba05178739baf1e1306b9bf81b05e8d301835
              started: "2026-10-21T19:35:34.456Z"
              spawned: true
              bytes: 937404
              windows:
                - start: "2026-10-21T19:35:34.456Z"
                  end: "2026-10-22T02:59:31-07:00"
            - harness: codex
              id: 01a160cd-43fe-7493-beca-70aa3d5c327b
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: f84fef466d65859c3fb67464fe72ba5b75c38b2328eba8b72dda2d620e2eb03a
              started: "2026-10-21T23:30:27.524Z"
              spawned: true
              bytes: 861347
              windows:
                - start: "2026-10-21T23:30:27.524Z"
                  end: "2026-10-22T02:59:31-07:00"
            - harness: codex
              id: 01a160cd-55f6-7062-a9a7-e0ad8a26aab0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 29cd538d8f11948d340cb0c482d789dde29aea4a0c3c0be2beed41baef8e5249
              started: "2026-10-21T23:30:32.135Z"
              spawned: true
              bytes: 482983
              windows:
                - start: "2026-10-21T23:30:32.135Z"
                  end: "2026-10-22T02:59:31-07:00"
            - harness: codex
              id: 01a160cd-6859-7580-b835-ba9b6504cd61
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 5e4a02068103f7d49fc266535e21799b15d804a28cd43e74bd0b2edd52664bcf
              started: "2026-10-21T23:30:36.9Z"
              spawned: true
              bytes: 503005
              windows:
                - start: "2026-10-21T23:30:36.9Z"
                  end: "2026-10-22T02:59:31-07:00"
        queue_waits:
            - attempt_id: 3bc05ff279d4794c4a8f6bb176e55c71
              slot: ariadne:4
              record_digest: 826068f401579bb474474655ad43cbb732ead13f8f46ba7d288c98f60dbb2c6e
              start: "2026-10-21T20:03:29.668881Z"
              end: "2026-10-21T20:16:17.372475Z"
              queue_seconds: 767.703594
              raw_seconds: 767.703594
              exact: true
            - attempt_id: 339841fb96fa15c43e6a69f42180754a
              slot: ariadne:4
              record_digest: 108ed183e363a5edb12ef960541d7c735b9549753fd5b14a2a9be137140b604f
              start: "2026-10-21T23:35:29.719462Z"
              end: "2026-10-22T00:32:25.75598Z"
              queue_seconds: 3416.036518
              raw_seconds: 3416.036518
              exact: true
            - attempt_id: ed2ea51d86f781c65c1be8eae1c4a383
              slot: ariadne:4
              record_digest: 24822d415bb3ef10c2a42cf0d7d2b4083131ce028a7d9c5365d1e741e48f857a
              start: "2026-10-22T03:03:15.261174Z"
              end: "2026-10-22T03:31:04.998823Z"
              queue_seconds: 1669.737649
              raw_seconds: 1669.737649
              exact: true
            - attempt_id: 8df37a23bfa69ec7d175c01207476ef2
              slot: ariadne:4
              record_digest: ff717783cd23ceba66466127d84696491c5b7a90d036cc9acb8e268b278dfaaa
              start: "2026-10-22T04:39:43.454895Z"
              end: "2026-10-22T05:16:34.231477Z"
              queue_seconds: 2210.776582
              raw_seconds: 2210.776582
              exact: true
        cuts:
            - dispatched: "2026-10-21T13:59:23-07:00"
              finalized: "2026-10-21T14:02:54-07:00"
              finalizing: false
            - dispatched: "2026-10-21T18:59:43-07:00"
              finalized: "2026-10-21T19:04:41-07:00"
              finalizing: false
            - dispatched: "2026-10-21T21:00:03-07:00"
              finalized: "2026-10-21T21:03:19-07:00"
              finalizing: false
      "n": 4
      timestamp: "2026-10-22T03:02:42-07:00"
      dispatched: "2026-10-22T03:00:03-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Pointer generation reparses the completed document before effects; adversarial-path and byte-preservation regressions pass in doc_model_test.go.
          round: 4
        - id: BR-2
          disposition: addressed
          note: First settlement supplies --expect-absent, enforced inside every publication CAS attempt; ownership and racing-destination regressions pass.
          round: 4
        - id: BR-3
          disposition: addressed
          note: Coordinator tests exercise failed writes, lost acknowledgments, late locks, unavailable readback and newer source work through retry.
          round: 4
        - id: BR-4
          disposition: addressed
          note: Ops publication fixtures provide the agent-id executable dependency; the complete tl-lead package passes.
          round: 4
        - id: BR-5
          disposition: addressed
          note: Marker scanning consumes actual Goldmark code contexts; pure and coordinator tests cover paragraph, heading, list and blockquote boundaries.
          round: 4
        - id: BR-6
          disposition: addressed
          note: Publication and stamping pass runSettlementEffect before execution; transition permissions and forbidden-effect regressions pass.
          round: 4
        - id: BR-7
          disposition: addressed
          note: pkg/sidedoc/literals.go supplies the shared code-context mask to scanning, rebasing and discovery. Pure projection and production-coordinator regressions cover all four reported contexts plus ordinary fence/inline controls. Disabling the rebasing mask in a disposable pinned-head checkout failed all twelve preservation subtests; the checkout was removed afterward.
          round: 4
      recipe: milestone-review
      isolation: fenced
      reviewed: 0d41e87788e51ba8bd7af58c052bf11042500ebc
      blocked: false
---

# Gate ledger — ariadne#507 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-21T14:02:54-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `managed-metadata-roundtrip` Allowed destination paths can produce an unreadable settlement pointer
  cmd/tl-lead/doc_model.go:113 writes paths directly into a plain YAML scalar, but workingdocPath accepts "workshop/workingdoc/a #b.md". YAML treats "#b.md @ COMMIT" as a comment, so settlement can publish, stamp, and report success while subsequent status and settle reject the source. Validate generated metadata with the production parser before effects and refuse unrepresentable paths; add round-trip regressions. ARCH-SECURE, ARCH-EVOLVE.
- **BR-2** [Critical] `first-settlement-destination-ownership` First product settlement can overwrite an unrelated committed destination
  cmd/tl-lead/doc_settle.go:349 accepts an occupied local target matching the payload and publishes with no expected commit. If committed main and the branch baseline contain different content, cmd/sdlc/trunkpublish.go:283 permits replacing it through ordinary edit semantics. A locally deleted tracked target also bypasses the collision check. Enforce absent-or-identical remote content for first settlement inside each CAS attempt, preserving ordinary trunk behavior; add product regressions. ARCH-PURPOSE.
- **BR-3** [Important] `settlement-interruption-coverage` Claimed coordinator recovery sequences are missing from the tests
  workshop/plans/000507-workingdoc-settle-plan.md:120 marks uncertain Apply and late-lock sequences delivered, but doc_fake_test.go:31 cannot return a lost acknowledgment after applying, and doc_settle_test.go only injects a source edit after publication. Add controlled Apply outcomes and late-lock/retry sequences through settleWorkingdoc, asserting confirmed pointer evidence and no duplicate publication. ARCH-ORDER.
- **BR-4** [Important] `integration-test-dependency-isolation` Ops publication tests require an installed host sdlc binary
  doc_operating_test.go:28 does not provide a controlled sdlc agent-id dependency for doc_operating.go's operatingAgentTrailer. Running PATH=/usr/bin:/bin /opt/homebrew/bin/go test ./cmd/tl-lead -run '^TestOperatingPublishExactPayloadAndRetry$' -count=1 fails with executable "sdlc" not found. Inject or fixture the identity dependency through the production seam so portable tests do not depend on an unrelated installed revision. ARCH-MOCK.

## Round 2 — 2026-10-21T19:04:41-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — doc_model.go:120 reparses generated metadata before effects. Adversarial pointer and coordinator refusal tests pass; removing the roundtrip guard in the disposable checkout makes them fail.
- BR-2 — addressed — doc_repository.go supplies --expect-absent for first settlement, and trunkpublish.go checks ownership inside every CAS callback. Real-remote ownership tests pass; disabling that predicate reproduces successful replacement of unrelated committed content.
- BR-3 — addressed — doc_fake_test.go now supports pre-write failure and post-write lost acknowledgment. Coordinator tests exercise late locks, unavailable readback, intervening edits and duplicate-free retries. Disabling uncertain-Apply reconciliation makes the lost-ack regression fail.
- BR-4 — addressed — doc_operating_test.go:28 supplies a strict fixture sdlc agent-id executable. All operating tests pass with PATH=/usr/bin:/bin; removing the fixture PATH injection reproduces the missing-sdlc failure.

### Raised

- **BR-5** [Critical] `marker-literal-boundaries` (kind: defect) Backticks across Markdown blocks hide unresolved review items
  cmd/tl-lead/doc_markers.go:171 searches for matching backticks across every non-fenced region, and line 206 marks the intervening blocks literal. Unmatched backticks in separate paragraphs, headings or list items therefore hide an intervening human question. Three coordinator probes confirmed zero actual CodeSpan nodes using the repository's Goldmark dependency, yet settleWorkingdoc returned success with Published=true and Stamped=true. ARCH-PURPOSE, ARCH-SECURE: restrict literal masking to actual Markdown code contexts and add scanner plus no-effects coordinator regressions across block boundaries, preserving legitimate multiline spans.
- **BR-6** [Important] `state-transition-authority` (kind: defect) Settlement transitions observe effects instead of authorizing them
  cmd/tl-lead/doc_settle.go:359 publishes before consulting transitionSettlement, and line 395 applies the source stamp before its transition. doc_model.go:153 returns no effects, while failure and unconfirmed events have no production consumers. This contradicts the plan's line 49 claim that only the transition model grants the next effect. ARCH-ORDER: make the pure model authorize effects and consume their outcomes, route failures through it, and test forbidden effect sequences through the production coordinator.

## Round 3 — 2026-10-21T21:03:19-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — Pointer generation reparses its output and checks exact equality; adversarial path and byte-preservation regressions pass.
- BR-2 — addressed — First settlement supplies --expect-absent, enforced inside every publication CAS callback; ownership regressions pass.
- BR-3 — addressed — Production coordinator tests cover lost acknowledgments, pre-write failure, late locks, unavailable readback, source edits and retries; the full package passes.
- BR-4 — addressed — Operating fixtures provide an isolated sdlc agent-id executable and assert its trailer; operating publication tests pass.
- BR-5 — addressed — Literal masks derive from actual Goldmark code nodes. Restoring the previous scanner makes all four block-boundary scanner and no-effects coordinator regressions fail.
- BR-6 — addressed — Publication and stamping require model permissions before execution, and failure/unconfirmed outcomes retire permissions. Removing the production authorization guard makes both forbidden-effect regressions fail.

### Raised

- **BR-7** [Critical] `markdown-literal-preservation` (kind: defect) Settlement publishes rewritten code examples
  cmd/tl-lead/doc_project.go:28 delegates to sidedoc.RebaseLinks, whose line-based literal detection misses multiline code spans, indented code, and fences nested in lists or blockquotes. Disposable-checkout probes through projectWorkingdoc and the production settleWorkingdoc coordinator reproduced all four cases: literal [example](../README.md) became [example](../../README.md), with Published=true and Stamped=true. This violates the plan's line 29 preservation contract (ARCH-PURPOSE); the scanner and rebaser also apply inconsistent Markdown literal rules (ARCH-DRY). Establish one shared rule for actual Markdown code contexts and sweep all projection rewrites, with pure preservation tests and coordinator regressions covering these four contexts plus ordinary-link controls.

## Round 4 — 2026-10-22T03:02:42-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — Pointer generation reparses the completed document before effects; adversarial-path and byte-preservation regressions pass in doc_model_test.go.
- BR-2 — addressed — First settlement supplies --expect-absent, enforced inside every publication CAS attempt; ownership and racing-destination regressions pass.
- BR-3 — addressed — Coordinator tests exercise failed writes, lost acknowledgments, late locks, unavailable readback and newer source work through retry.
- BR-4 — addressed — Ops publication fixtures provide the agent-id executable dependency; the complete tl-lead package passes.
- BR-5 — addressed — Marker scanning consumes actual Goldmark code contexts; pure and coordinator tests cover paragraph, heading, list and blockquote boundaries.
- BR-6 — addressed — Publication and stamping pass runSettlementEffect before execution; transition permissions and forbidden-effect regressions pass.
- BR-7 — addressed — pkg/sidedoc/literals.go supplies the shared code-context mask to scanning, rebasing and discovery. Pure projection and production-coordinator regressions cover all four reported contexts plus ordinary fence/inline controls. Disabling the rebasing mask in a disposable pinned-head checkout failed all twelve preservation subtests; the checkout was removed afterward.

## Open findings

(none — every finding has been disposed)
