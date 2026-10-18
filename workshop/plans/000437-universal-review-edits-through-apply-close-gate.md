---
gate: boundary-review
issue: 437
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: d2cda983-e7ec-413e-bb89-5e6b9abf4157
            repository: github.com/xianxu/ariadne
            issue_id: "000437"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/73816bf80a42a6d2
            reviewed_head: 930ab339c87d9cedeae94897f2e226cf3e8a6272
            slot: ariadne:1
            requested_at: "2026-10-18T19:10:24.93917Z"
          phase: finished
          pid: 93984
          started_at: "2026-10-18T19:10:25.141143Z"
          finished_at: "2026-10-18T19:15:40.044587Z"
          session_id: 01a1506c-1efa-7f80-886d-d33831248123
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 930ab339c87d9cedeae94897f2e226cf3e8a6272
        tracker_head: 99c748780f213666ba796b0898ecbd308eb15f6c
        branch_point: 2e783ac22e1a79d3a6cddce34016f5353679415c
        since: "2026-10-18T04:55:12-07:00"
        until: "2026-10-18T12:09:43-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T04:55:12-07:00"
              end: "2026-10-18T12:09:43-07:00"
              began: "2026-10-18T04:55:12-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 014c7df54b1c3ea9abf81e52972e6e42a20ec226e34120f0bee5916699195447
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 115628763
              windows:
                - start: "2026-10-18T04:55:12-07:00"
                  end: "2026-10-18T12:09:43-07:00"
            - harness: codex
              id: 01a14eea-4c39-71d2-b9c8-8735b8b58509
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              digest: f270f489ab6fd0ce4c6f068ee3444f317a7e530f737c7281b8517abdf1f0a650
              started: "2026-10-18T12:09:00.431Z"
              spawned: true
              bytes: 543162
              windows:
                - start: "2026-10-18T12:09:00.431Z"
                  end: "2026-10-18T12:09:43-07:00"
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
                  end: "2026-10-18T12:09:43-07:00"
        queue_waits:
            - attempt_id: 167ef685d14ef984a1ccc9b83a28a5b1
              slot: ariadne:1
              record_digest: 0e39c9d6338f94cf910a2bfc8d333312439b3c7acfaa80702d390a1cee843b1a
              start: "2026-10-18T12:21:07.418501Z"
              end: "2026-10-18T12:28:33.48289Z"
              queue_seconds: 446.064389
              raw_seconds: 446.064389
              exact: true
            - attempt_id: 97e9c79d9d1d102df2d40c32801f3864
              slot: ariadne:1
              record_digest: 302427f7fcd59fc6f235b929456fe28e0d9dee2ebfd0c5ec057a740997e6047c
              start: "2026-10-18T13:00:46.040063Z"
              end: "2026-10-18T13:09:43.849079Z"
              queue_seconds: 537.809016
              raw_seconds: 537.809016
              exact: true
            - attempt_id: 8c5a5e18aa58707db347c709c810f563
              slot: ariadne:1
              record_digest: e3d4f017eb327e3efee5b4d6035c7cfc077e114ad6e05dcb010f951cd3ff0b17
              start: "2026-10-18T13:27:17.095135Z"
              end: "2026-10-18T13:38:49.768584Z"
              queue_seconds: 692.673449
              raw_seconds: 692.673449
              exact: true
            - attempt_id: 90320ab29b17a67bea80a2671c62c12a
              slot: ariadne:1
              record_digest: 2d18cf65a08f926a2e69fb28f4d9985e7505b65b6ed8c309fe046dc6d1471a51
              start: "2026-10-18T13:54:20.761552Z"
              end: "2026-10-18T14:15:05.235935Z"
              queue_seconds: 1244.474383
              raw_seconds: 1244.474383
              exact: true
            - attempt_id: 4a60e90cc408df07a4638b2c53e283ac
              slot: ariadne:1
              record_digest: b4058fa81683863384dde7421a488aa896deb34a4d79ace8aeaa02202c45ca6e
              start: "2026-10-18T14:32:04.166177Z"
              end: "2026-10-18T14:59:49.477614Z"
              queue_seconds: 1665.311437
              raw_seconds: 1665.311437
              exact: true
            - attempt_id: 28f483f4799341fd56dad2c463677506
              slot: ariadne:1
              record_digest: 1f189210f8493e3900895b7da9b87104d5b38c60a7c2af139e525891aeeb29dc
              start: "2026-10-18T15:15:06.25148Z"
              end: "2026-10-18T15:39:30.12085Z"
              queue_seconds: 1463.86937
              raw_seconds: 1463.86937
              exact: true
            - attempt_id: 77d0c88769993b116c9a81f17f6899e5
              slot: ariadne:1
              record_digest: 691c68f7041c15337939e4f7fd16375e059cabd17226d2133de3bd4015692592
              start: "2026-10-18T15:53:13.613306Z"
              end: "2026-10-18T16:18:19.150532Z"
              queue_seconds: 1505.537226
              raw_seconds: 1505.537226
              exact: true
            - attempt_id: 632f45e915f1038d4da0444be6b2d4c6
              slot: ariadne:1
              record_digest: 8a092f4b63f32315462d1158109232274b3edd8788c472a9da1038c1571854a3
              start: "2026-10-18T16:31:45.679488Z"
              end: "2026-10-18T16:38:32.998977Z"
              queue_seconds: 407.319489
              raw_seconds: 407.319489
              exact: true
            - attempt_id: 9a4b90a9e4ed50c8436c7b67033a7564
              slot: ariadne:1
              record_digest: 8798a38cb345ca670dbcb83cab1f00bc65d041520ca0e11680a189d636482f99
              start: "2026-10-18T16:58:08.688864Z"
              end: "2026-10-18T17:06:12.210667Z"
              queue_seconds: 483.521803
              raw_seconds: 483.521803
              exact: true
            - attempt_id: 8f7bced5bdd91f8b15a622646731b7f4
              slot: ariadne:1
              record_digest: 177a6a087cdf56f4f74ecb8da67ad98f43fed20a154c46d7a9f14a8a931ca302
              start: "2026-10-18T17:21:24.866846Z"
              end: "2026-10-18T17:42:50.99172Z"
              queue_seconds: 1286.124874
              raw_seconds: 1286.124874
              exact: true
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
      "n": 1
      timestamp: "2026-10-18T12:15:41-07:00"
      dispatched: "2026-10-18T12:10:24-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: BR-1
          severity: Critical
          title: Untracked edits bypass the document's proposal presentation
          detail: 'construct/local/universal-review/round.md:50-56 now requires clean edits, forbids agent-written proposal markers, and promises proposals under presentation: propose. However, adapters/file-edit.md:124-127 still directs untracked documents straight through harness file tools, with no converter. An unregistered document outside Couch therefore receives clean replacements instead of proposals; round.md:115 also requires round --json output that this path cannot produce. The family includes the submitted untracked document and any untracked parent routed through adapters/file-edit.md:115-118. Define a supported presentation and reporting path for both, and exercise propose on tracked and untracked documents. ARCH-PURPOSE: the document-level contract must reach every supported writer.'
          family: presentation-all-write-paths
          round: 1
        - id: BR-2
          severity: Critical
          title: The proposal fallback lacks a supported path when no ready marker exists
          detail: 'adapters/file-edit.md:95-99 unconditionally routes large or unencodable rewrites to an applied proposal link, but its direct-link exception at lines 62-69 requires an operator''s ready marker. A marker-free developmental pass is explicitly supported by round.md:45 and modes/developmental.md:7-10, yet cannot use that exception. The same gap occurs in fix/SKILL.md:229''s hosted fallback and collaboration-modes.md:135-137''s marker-free chat answer links, routed through the shared link contract at lines 204-215. Provide a supported marker-free structural-link path, preserving link/index consistency, and add conformance cases without ready markers. The new large-proposal fixture covers only the marker-present case. ARCH-PURPOSE: sweep these consumers together.'
          family: structural-links-without-markers
          round: 1
        - id: BR-3
          severity: Critical
          title: A current stored round does not prove the attempted apply landed
          detail: construct/local/fix/SKILL.md:234-238 instructs the agent to commit stored Body whenever Current is true after a lost response. If a previous round is still current and the new apply never reaches Pair, these instructions commit the previous Body while leaving the new requested edits unapplied. Branch/file checks cannot distinguish those operations. The family includes lost-response recovery here and interrupted-commit recovery at lines 258-261, which also needs the successful operation's identity. Correlate the stored round with the attempted apply using a request or round identity and its base/result hashes; test both an old current round with an unreceived request and a newly landed request with a lost response.
          family: apply-recovery-operation-correlation
          round: 1
        - id: BR-4
          severity: Important
          title: README update appears missing for the new review surface
          detail: README.md is unchanged in the pinned range and contains no presentation or review-apply guidance. This window introduces the document presentation key and the pair review apply / pair review round workflow, documented in round.md:50-62 and fix/SKILL.md:211-243. These are all instances of the new user-facing surface requiring README coverage under this boundary's docs gate. Add concise usage and migration guidance, linking to the updated atlas.
          family: user-surface-readme-coverage
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#437 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T12:15:41-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `presentation-all-write-paths` Untracked edits bypass the document's proposal presentation
  construct/local/universal-review/round.md:50-56 now requires clean edits, forbids agent-written proposal markers, and promises proposals under presentation: propose. However, adapters/file-edit.md:124-127 still directs untracked documents straight through harness file tools, with no converter. An unregistered document outside Couch therefore receives clean replacements instead of proposals; round.md:115 also requires round --json output that this path cannot produce. The family includes the submitted untracked document and any untracked parent routed through adapters/file-edit.md:115-118. Define a supported presentation and reporting path for both, and exercise propose on tracked and untracked documents. ARCH-PURPOSE: the document-level contract must reach every supported writer.
- **BR-2** [Critical] `structural-links-without-markers` The proposal fallback lacks a supported path when no ready marker exists
  adapters/file-edit.md:95-99 unconditionally routes large or unencodable rewrites to an applied proposal link, but its direct-link exception at lines 62-69 requires an operator's ready marker. A marker-free developmental pass is explicitly supported by round.md:45 and modes/developmental.md:7-10, yet cannot use that exception. The same gap occurs in fix/SKILL.md:229's hosted fallback and collaboration-modes.md:135-137's marker-free chat answer links, routed through the shared link contract at lines 204-215. Provide a supported marker-free structural-link path, preserving link/index consistency, and add conformance cases without ready markers. The new large-proposal fixture covers only the marker-present case. ARCH-PURPOSE: sweep these consumers together.
- **BR-3** [Critical] `apply-recovery-operation-correlation` A current stored round does not prove the attempted apply landed
  construct/local/fix/SKILL.md:234-238 instructs the agent to commit stored Body whenever Current is true after a lost response. If a previous round is still current and the new apply never reaches Pair, these instructions commit the previous Body while leaving the new requested edits unapplied. Branch/file checks cannot distinguish those operations. The family includes lost-response recovery here and interrupted-commit recovery at lines 258-261, which also needs the successful operation's identity. Correlate the stored round with the attempted apply using a request or round identity and its base/result hashes; test both an old current round with an unreceived request and a newly landed request with a lost response.
- **BR-4** [Important] `user-surface-readme-coverage` README update appears missing for the new review surface
  README.md is unchanged in the pinned range and contains no presentation or review-apply guidance. This window introduces the document presentation key and the pair review apply / pair review round workflow, documented in round.md:50-62 and fix/SKILL.md:211-243. These are all instances of the new user-facing surface requiring README coverage under this boundary's docs gate. Add concise usage and migration guidance, linking to the updated atlas.

## Open findings

- **BR-1** [Critical] `presentation-all-write-paths` Untracked edits bypass the document's proposal presentation
- **BR-2** [Critical] `structural-links-without-markers` The proposal fallback lacks a supported path when no ready marker exists
- **BR-3** [Critical] `apply-recovery-operation-correlation` A current stored round does not prove the attempted apply landed
- **BR-4** [Important] `user-surface-readme-coverage` README update appears missing for the new review surface
