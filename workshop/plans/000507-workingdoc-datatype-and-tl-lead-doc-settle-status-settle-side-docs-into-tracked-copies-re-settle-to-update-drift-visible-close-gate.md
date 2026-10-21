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

## Open findings

- **BR-1** [Critical] `managed-metadata-roundtrip` Allowed destination paths can produce an unreadable settlement pointer
- **BR-2** [Critical] `first-settlement-destination-ownership` First product settlement can overwrite an unrelated committed destination
- **BR-3** [Important] `settlement-interruption-coverage` Claimed coordinator recovery sequences are missing from the tests
- **BR-4** [Important] `integration-test-dependency-isolation` Ops publication tests require an installed host sdlc binary
