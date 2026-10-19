---
gate: boundary-review
issue: 456
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: e73ca3aa-5f97-4d6a-b600-15c2586ad542
            repository: github.com/xianxu/ariadne
            issue_id: "000456"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/d4423222b0f94ad0
            reviewed_head: 1e324f8a89c20d0da7237b96906a37b05b81af46
            slot: ariadne:4
            requested_at: "2026-10-19T21:02:09.491137Z"
          phase: finished
          pid: 83725
          started_at: "2026-10-19T21:02:09.792951Z"
          finished_at: "2026-10-19T21:05:30.831598Z"
          session_id: a75c1f88-ea07-4b8b-8e66-08a412a1f6c3
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 1e324f8a89c20d0da7237b96906a37b05b81af46
        tracker_head: b0564b2d5c96b4f5e3bc42f42477d9a8a4828eba
        branch_point: 5058591f86fce002653fdca0bf6e7f1b5c4460e6
        since: "2026-10-19T01:26:56-07:00"
        until: "2026-10-19T13:58:18-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T01:26:56-07:00"
              end: "2026-10-19T13:58:18-07:00"
              began: "2026-10-19T01:26:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 494a042fe8736b8a931a0dab6a6081f2883a17c821cfa93d47c68407cc4e8b11
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 83276551
              windows:
                - start: "2026-10-19T01:26:56-07:00"
                  end: "2026-10-19T13:58:18-07:00"
        queue_waits:
            - attempt_id: bab05729e5aa66707d18abc487e82c82
              slot: ariadne:4
              record_digest: 4269c85bb4cee1e5d6d512fb1ca487093fc523f8f4e0c964b593b99172217c84
              start: "2026-10-19T16:56:57.010726Z"
              end: "2026-10-19T17:00:27.968039Z"
              queue_seconds: 210.957313
              raw_seconds: 210.957313
              exact: true
            - attempt_id: f40b02c7b758c556fd24519f024975c3
              slot: ariadne:4
              record_digest: 2e5a51b6f6979f0e1b47fdafd876e549427c0fdad41895c8eadcc5909daf5884
              start: "2026-10-19T17:33:42.737584Z"
              end: "2026-10-19T17:44:43.407011Z"
              queue_seconds: 660.669427
              raw_seconds: 660.669427
              exact: true
            - attempt_id: b00d103272bccd5c7313f6eb9494dd2c
              slot: ariadne:4
              record_digest: 1f5e84b09ee13d47fb9b0bc885d70445c895f5c8c2cfdd897a6ac9e26db82257
              start: "2026-10-19T18:17:04.975887Z"
              end: "2026-10-19T18:17:31.242964Z"
              queue_seconds: 26.267077
              raw_seconds: 26.267077
              exact: true
            - attempt_id: 7d26798c7990c53f884cc3d352da7920
              slot: ariadne:4
              record_digest: c1eb38ed45e3a2afad21d66517efedd0af2886dd6be61c7b21aea05cc65fea6c
              start: "2026-10-19T20:19:38.867364Z"
              end: "2026-10-19T20:27:27.018395Z"
              queue_seconds: 468.151031
              raw_seconds: 468.151031
              exact: true
        protected:
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T16:32:27.029Z"
              end: "2026-10-19T16:56:57.010726Z"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T17:00:27.968039Z"
              end: "2026-10-19T17:12:03.292Z"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T17:12:06.457Z"
              end: "2026-10-19T17:33:42.737584Z"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T17:44:43.407011Z"
              end: "2026-10-19T17:48:45.763Z"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T17:55:31.654Z"
              end: "2026-10-19T10:57:22-07:00"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T10:57:22-07:00"
              end: "2026-10-19T11:11:34-07:00"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T11:15:17-07:00"
              end: "2026-10-19T18:16:45.598Z"
      "n": 1
      timestamp: "2026-10-19T14:05:31-07:00"
      dispatched: "2026-10-19T14:02:09-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: fleet resume's inFlight treats an owner_error or worktree-only owner as no work, classing the slot idle
          detail: cmd/tl-lead/fleet.go:116-129 matches only is.Owner == addr. A non-terminal issue with OwnerErr, or an owner whose workspace is empty (decodeRepoIssues falls back to the worktree path), never matches, so the slot is silently classed idle. Plan-quality round 1 says an unmatched owner is listed as unknown, never matched, and the Done-when says a failed read is unknown. Class such slots unknown and add a fixture with an OwnerErr row.
          family: failed-read-is-never-absence
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: tl-lead's .hub grammar copy omits pair's UTF-8 check and its 0xff test vector
          detail: pair couchcore.ParseHubClaims refuses non-UTF-8 input, and TestHubClaimsParse includes string([]byte{0xff}). cmd/tl-lead/team.go parseHub has no utf8.Valid check, and team_test.go dropped that vector while claiming the vectors are pair's own. Add the check and restore the vector.
          family: mirrored-contract-drift
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: fleet resume matches issue owners to slots by workspace address only, ignoring the owner's machine
          detail: sdlc state owner rows carry machine_name, but inFlight compares owner.workspace alone (fleet.go:125). A claim on another machine's pair:3 makes this machine's down pair:3 count as work in flight, and --apply resumes it when a local conversation is proven. Filter to this machine, or document and pin that sdlc reports only local owners.
          family: observation-bound-to-exact-entity
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: execResume.Candidates has no timeout, so a hung couch hangs fleet resume
          detail: fleet.go:42 uses exec.Command without a context, unlike Resume (3 min) and the SDLC seam (60 s).
          family: bounded-external-calls
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: 'execResume.Resume reports an exec failure that is not an exit error as "refused: (none)"'
          detail: fleet.go:75-77 drops err when stderr is empty (for example, couch not on PATH). Include err in the message.
          family: error-cause-preserved
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: team list prints nothing when run from a checkout whose parent folder is not the workspace
          detail: From a worktree slot, readHubs scans the slot folder, finds no .hub, and the command prints nothing. It should say no teams were found in that folder.
          family: empty-result-says-why
          round: 1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 056232a3-4732-4b9e-adad-58ce9774bffd
            repository: github.com/xianxu/ariadne
            issue_id: "000456"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/c7d8873c8e12dc96
            reviewed_head: d7ec954cfbeb9bf0de22969bc899078c50dfff5b
            slot: ariadne:4
            requested_at: "2026-10-19T21:08:55.851753Z"
          phase: finished
          pid: 60507
          started_at: "2026-10-19T21:08:56.011022Z"
          finished_at: "2026-10-19T21:10:46.126563Z"
          session_id: 943c5591-e926-424c-ab3e-ab9cc466a73e
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: d7ec954cfbeb9bf0de22969bc899078c50dfff5b
        tracker_head: b0564b2d5c96b4f5e3bc42f42477d9a8a4828eba
        branch_point: 5058591f86fce002653fdca0bf6e7f1b5c4460e6
        since: "2026-10-19T01:26:56-07:00"
        until: "2026-10-19T14:08:14-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T01:26:56-07:00"
              end: "2026-10-19T14:08:14-07:00"
              began: "2026-10-19T01:26:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 365a6d97d7cb47d4d0197c39ba7a3011c3c3fba11d513a208ba619b75c5b18d7
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 86652933
              windows:
                - start: "2026-10-19T01:26:56-07:00"
                  end: "2026-10-19T14:08:14-07:00"
        queue_waits:
            - attempt_id: bab05729e5aa66707d18abc487e82c82
              slot: ariadne:4
              record_digest: 4269c85bb4cee1e5d6d512fb1ca487093fc523f8f4e0c964b593b99172217c84
              start: "2026-10-19T16:56:57.010726Z"
              end: "2026-10-19T17:00:27.968039Z"
              queue_seconds: 210.957313
              raw_seconds: 210.957313
              exact: true
            - attempt_id: f40b02c7b758c556fd24519f024975c3
              slot: ariadne:4
              record_digest: 2e5a51b6f6979f0e1b47fdafd876e549427c0fdad41895c8eadcc5909daf5884
              start: "2026-10-19T17:33:42.737584Z"
              end: "2026-10-19T17:44:43.407011Z"
              queue_seconds: 660.669427
              raw_seconds: 660.669427
              exact: true
            - attempt_id: b00d103272bccd5c7313f6eb9494dd2c
              slot: ariadne:4
              record_digest: 1f5e84b09ee13d47fb9b0bc885d70445c895f5c8c2cfdd897a6ac9e26db82257
              start: "2026-10-19T18:17:04.975887Z"
              end: "2026-10-19T18:17:31.242964Z"
              queue_seconds: 26.267077
              raw_seconds: 26.267077
              exact: true
            - attempt_id: 7d26798c7990c53f884cc3d352da7920
              slot: ariadne:4
              record_digest: c1eb38ed45e3a2afad21d66517efedd0af2886dd6be61c7b21aea05cc65fea6c
              start: "2026-10-19T20:19:38.867364Z"
              end: "2026-10-19T20:27:27.018395Z"
              queue_seconds: 468.151031
              raw_seconds: 468.151031
              exact: true
        protected:
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T16:32:27.029Z"
              end: "2026-10-19T16:56:57.010726Z"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T17:00:27.968039Z"
              end: "2026-10-19T17:12:03.292Z"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T17:12:06.457Z"
              end: "2026-10-19T17:33:42.737584Z"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T17:44:43.407011Z"
              end: "2026-10-19T17:48:45.763Z"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T17:55:31.654Z"
              end: "2026-10-19T10:57:22-07:00"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T10:57:22-07:00"
              end: "2026-10-19T11:11:34-07:00"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T11:15:17-07:00"
              end: "2026-10-19T18:16:45.598Z"
        cuts:
            - dispatched: "2026-10-19T14:02:09-07:00"
              finalized: "2026-10-19T14:05:31-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-19T14:10:47-07:00"
      dispatched: "2026-10-19T14:08:55-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Claims now come from inventory rows; a claims_state other than present or absent, or a missing row, is unknown. TestDecodeFleetClaims and TestFleetResumeDryRun go red when the fix is reverted.
          round: 2
        - id: BR-2
          disposition: addressed
          note: utf8.Valid added in team.go parseHub; the 0xff vector is restored and TestHubClaimVectors goes red without the check.
          round: 2
        - id: BR-3
          disposition: addressed
          note: tl-lead's owner match is deleted; sdlc PlaceClaims places only claims whose claimant is this machine's fingerprint and worktree (cmd/sdlc/internal/fleet/claims.go mine).
          round: 2
        - id: BR-4
          disposition: not-addressed
          note: Candidates now uses CommandContext with a 1-minute timeout, but no test pins it and the timeout cannot be injected; make the timeout injectable and test it with a hanging fake binary.
          round: 2
        - id: BR-5
          disposition: not-addressed
          note: The code fix is right (fleet.go:85-90, "did not run"), but there is no regression test; execResume with a nonexistent bin asserting "did not run" is cheap.
          round: 2
        - id: BR-6
          disposition: addressed
          note: renderTeams prints "no teams" for an empty hub set (tested), and team list names the scanned workspace folder.
          round: 2
      findings:
        - kind: defect
          id: BR-7
          severity: Important
          title: fleet resume reads a slot's claims from its host row only, ignoring dependency members, so a slot with in-flight work on a dependency clone is classed idle
          detail: 'This is the 2nd finding in this family. The rule: a slot observation covers the slot''s full member set, as sdlc''s AssembleSlots (slots.go:213-238) defines it, not just the host. decodeFleet (sdlc.go:169) joins only HostPath, but a claim made in a slot''s own dependency clone is recorded on that clone''s row. Fix: collect member paths, union their non-terminal claims, and set ClaimsKnown only when every member row is present or absent. Add a fixture with a claim on a dependency member.'
          family: observation-bound-to-exact-entity
          round: 2
        - kind: defect
          id: BR-8
          severity: Minor
          title: decodeFleet drops claims_error, so an unknown or stale slot prints no reason
          detail: 'This is the 2nd finding in this family. The rule: when a decoded read-quality state is not an answer, carry its error text through to the printed reason, as decodeObservation does with Card.Error. Add claims_error to the Rows struct, include it in ClaimsWhy, and sweep the other decoders in sdlc.go and team_issues.go.'
          family: error-cause-preserved
          round: 2
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 2826d585-7d9a-4d9a-b623-450100fedbc5
            repository: github.com/xianxu/ariadne
            issue_id: "000456"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/78fa6df45ade5e8f
            reviewed_head: 5dbc3fc3c36276e7c15858a9de82eac60063cc88
            slot: ariadne:4
            requested_at: "2026-10-19T21:12:14.493445Z"
          phase: finished
          pid: 88166
          started_at: "2026-10-19T21:12:14.605416Z"
          finished_at: "2026-10-19T21:13:22.136908Z"
          session_id: cdaaf6fa-a2e0-4170-8d2c-d8cb4fe632da
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 5dbc3fc3c36276e7c15858a9de82eac60063cc88
        tracker_head: b0564b2d5c96b4f5e3bc42f42477d9a8a4828eba
        branch_point: 5058591f86fce002653fdca0bf6e7f1b5c4460e6
        since: "2026-10-19T01:26:56-07:00"
        until: "2026-10-19T14:11:46-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T01:26:56-07:00"
              end: "2026-10-19T14:11:46-07:00"
              began: "2026-10-19T01:26:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 67c21bd7e09204cf3c653e200a7b35baab80e3a090c21d8e5c0b3cf3014dc031
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 88196844
              windows:
                - start: "2026-10-19T01:26:56-07:00"
                  end: "2026-10-19T14:11:46-07:00"
        queue_waits:
            - attempt_id: bab05729e5aa66707d18abc487e82c82
              slot: ariadne:4
              record_digest: 4269c85bb4cee1e5d6d512fb1ca487093fc523f8f4e0c964b593b99172217c84
              start: "2026-10-19T16:56:57.010726Z"
              end: "2026-10-19T17:00:27.968039Z"
              queue_seconds: 210.957313
              raw_seconds: 210.957313
              exact: true
            - attempt_id: f40b02c7b758c556fd24519f024975c3
              slot: ariadne:4
              record_digest: 2e5a51b6f6979f0e1b47fdafd876e549427c0fdad41895c8eadcc5909daf5884
              start: "2026-10-19T17:33:42.737584Z"
              end: "2026-10-19T17:44:43.407011Z"
              queue_seconds: 660.669427
              raw_seconds: 660.669427
              exact: true
            - attempt_id: b00d103272bccd5c7313f6eb9494dd2c
              slot: ariadne:4
              record_digest: 1f5e84b09ee13d47fb9b0bc885d70445c895f5c8c2cfdd897a6ac9e26db82257
              start: "2026-10-19T18:17:04.975887Z"
              end: "2026-10-19T18:17:31.242964Z"
              queue_seconds: 26.267077
              raw_seconds: 26.267077
              exact: true
            - attempt_id: 7d26798c7990c53f884cc3d352da7920
              slot: ariadne:4
              record_digest: c1eb38ed45e3a2afad21d66517efedd0af2886dd6be61c7b21aea05cc65fea6c
              start: "2026-10-19T20:19:38.867364Z"
              end: "2026-10-19T20:27:27.018395Z"
              queue_seconds: 468.151031
              raw_seconds: 468.151031
              exact: true
        protected:
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T16:32:27.029Z"
              end: "2026-10-19T16:56:57.010726Z"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T17:00:27.968039Z"
              end: "2026-10-19T17:12:03.292Z"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T17:12:06.457Z"
              end: "2026-10-19T17:33:42.737584Z"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T17:44:43.407011Z"
              end: "2026-10-19T17:48:45.763Z"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T17:55:31.654Z"
              end: "2026-10-19T10:57:22-07:00"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T10:57:22-07:00"
              end: "2026-10-19T11:11:34-07:00"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T11:15:17-07:00"
              end: "2026-10-19T18:16:45.598Z"
        cuts:
            - dispatched: "2026-10-19T14:02:09-07:00"
              finalized: "2026-10-19T14:05:31-07:00"
              finalizing: false
            - dispatched: "2026-10-19T14:08:55-07:00"
              finalized: "2026-10-19T14:10:47-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-19T14:13:23-07:00"
      dispatched: "2026-10-19T14:12:14-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-4
          disposition: addressed
          note: fleet.go:55-71 runs Candidates under a context with a 1-minute default limit (injectable); TestExecResumeBoundsAndCauses times out a hung script.
          round: 3
        - id: BR-5
          disposition: addressed
          note: 'fleet.go:99-103 reports a non-exit error as "did not run: <err>" and a refusal as "refused (<err>): <stderr>"; both paths are covered by the missing-couch and refuse-script cases.'
          round: 3
        - id: BR-7
          disposition: addressed
          note: sdlc.go:173-194 joins claims over every member row, with ClaimsKnown only when all are read; the ariadne:2 fixture holds a claim on a dependency clone only, and a host-only join would fail it.
          round: 3
        - id: BR-8
          disposition: addressed
          note: 'claims_error is decoded (sdlc.go:121), added to the state, and printed through ClaimsWhy; the test asserts "unknown: tracker fetch failed". Sibling decoders already carry their errors (Card.Error, owner_error).'
          round: 3
      findings:
        - kind: defect
          id: BR-9
          severity: Minor
          title: A declared but missing dependency clone makes the slot's claims unknown, so fleet resume never resumes it
          detail: 'sdlc''s AssembleSlots (slots.go:232-234) lists a MemberMissing dependency with verdict "missing" and no tree row. decodeFleet (sdlc.go:179-182) reads the missing row as an unread member and sets ClaimsKnown=false. Fix: read members[].verdict and treat a "missing" member as holding no claims. Add a fixture.'
          family: consume-producer-verdict
          round: 3
      recipe: milestone-review
      isolation: checkout
      reviewed: 5dbc3fc3c36276e7c15858a9de82eac60063cc88
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a44f6401-e953-4c9f-a029-f69b61d4a91b
            repository: github.com/xianxu/ariadne
            issue_id: "000456"
            recipe: fix-review
            round: 4
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/0368f9f381e8ec30
            reviewed_head: 38dfba58988272731fcd1df39593a6cfda0b4311
            slot: ariadne:4
            requested_at: "2026-10-19T21:14:29.4341Z"
          phase: finished
          pid: 3419
          started_at: "2026-10-19T21:14:29.593548Z"
          finished_at: "2026-10-19T21:15:23.377892Z"
          session_id: 3dc0ec2d-37c9-4d73-af4f-657cc1a6f54e
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 38dfba58988272731fcd1df39593a6cfda0b4311
        tracker_head: c6c91d469ccfadeab6f69c2d67dbf6a6c99292ec
        branch_point: 5058591f86fce002653fdca0bf6e7f1b5c4460e6
        since: "2026-10-19T01:26:56-07:00"
        until: "2026-10-19T14:13:58-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T01:26:56-07:00"
              end: "2026-10-19T14:13:58-07:00"
              began: "2026-10-19T01:26:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: ce2d6f0b2bd325fb159905017ecbbec80e190f1e666fee04db1d5dc2239c0028
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 89302770
              windows:
                - start: "2026-10-19T01:26:56-07:00"
                  end: "2026-10-19T14:13:58-07:00"
        queue_waits:
            - attempt_id: bab05729e5aa66707d18abc487e82c82
              slot: ariadne:4
              record_digest: 4269c85bb4cee1e5d6d512fb1ca487093fc523f8f4e0c964b593b99172217c84
              start: "2026-10-19T16:56:57.010726Z"
              end: "2026-10-19T17:00:27.968039Z"
              queue_seconds: 210.957313
              raw_seconds: 210.957313
              exact: true
            - attempt_id: f40b02c7b758c556fd24519f024975c3
              slot: ariadne:4
              record_digest: 2e5a51b6f6979f0e1b47fdafd876e549427c0fdad41895c8eadcc5909daf5884
              start: "2026-10-19T17:33:42.737584Z"
              end: "2026-10-19T17:44:43.407011Z"
              queue_seconds: 660.669427
              raw_seconds: 660.669427
              exact: true
            - attempt_id: b00d103272bccd5c7313f6eb9494dd2c
              slot: ariadne:4
              record_digest: 1f5e84b09ee13d47fb9b0bc885d70445c895f5c8c2cfdd897a6ac9e26db82257
              start: "2026-10-19T18:17:04.975887Z"
              end: "2026-10-19T18:17:31.242964Z"
              queue_seconds: 26.267077
              raw_seconds: 26.267077
              exact: true
            - attempt_id: 7d26798c7990c53f884cc3d352da7920
              slot: ariadne:4
              record_digest: c1eb38ed45e3a2afad21d66517efedd0af2886dd6be61c7b21aea05cc65fea6c
              start: "2026-10-19T20:19:38.867364Z"
              end: "2026-10-19T20:27:27.018395Z"
              queue_seconds: 468.151031
              raw_seconds: 468.151031
              exact: true
        protected:
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T16:32:27.029Z"
              end: "2026-10-19T16:56:57.010726Z"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T17:00:27.968039Z"
              end: "2026-10-19T17:12:03.292Z"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T17:12:06.457Z"
              end: "2026-10-19T17:33:42.737584Z"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T17:44:43.407011Z"
              end: "2026-10-19T17:48:45.763Z"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T17:55:31.654Z"
              end: "2026-10-19T10:57:22-07:00"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T10:57:22-07:00"
              end: "2026-10-19T11:11:34-07:00"
            - issue: "468"
              slot: ariadne:4
              start: "2026-10-19T11:15:17-07:00"
              end: "2026-10-19T18:16:45.598Z"
        cuts:
            - dispatched: "2026-10-19T14:02:09-07:00"
              finalized: "2026-10-19T14:05:31-07:00"
              finalizing: false
            - dispatched: "2026-10-19T14:08:55-07:00"
              finalized: "2026-10-19T14:10:47-07:00"
              finalizing: false
            - dispatched: "2026-10-19T14:12:14-07:00"
              finalized: "2026-10-19T14:13:23-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-19T14:15:25-07:00"
      dispatched: "2026-10-19T14:14:29-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-9
          disposition: addressed
          note: sdlc.go:182 exempts members whose producer verdict is "missing" (slots.go:23,52,232-234); fixture ariadne:5 in fleet_test.go fails without it; go test passes.
          round: 4
      recipe: fix-review
      isolation: checkout
      reviewed: 38dfba58988272731fcd1df39593a6cfda0b4311
      blocked: false
---

# Gate ledger — ariadne#456 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T14:05:31-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `failed-read-is-never-absence` (kind: defect) fleet resume's inFlight treats an owner_error or worktree-only owner as no work, classing the slot idle
  cmd/tl-lead/fleet.go:116-129 matches only is.Owner == addr. A non-terminal issue with OwnerErr, or an owner whose workspace is empty (decodeRepoIssues falls back to the worktree path), never matches, so the slot is silently classed idle. Plan-quality round 1 says an unmatched owner is listed as unknown, never matched, and the Done-when says a failed read is unknown. Class such slots unknown and add a fixture with an OwnerErr row.
- **BR-2** [Important] `mirrored-contract-drift` (kind: defect) tl-lead's .hub grammar copy omits pair's UTF-8 check and its 0xff test vector
  pair couchcore.ParseHubClaims refuses non-UTF-8 input, and TestHubClaimsParse includes string([]byte{0xff}). cmd/tl-lead/team.go parseHub has no utf8.Valid check, and team_test.go dropped that vector while claiming the vectors are pair's own. Add the check and restore the vector.
- **BR-3** [Important] `observation-bound-to-exact-entity` (kind: defect) fleet resume matches issue owners to slots by workspace address only, ignoring the owner's machine
  sdlc state owner rows carry machine_name, but inFlight compares owner.workspace alone (fleet.go:125). A claim on another machine's pair:3 makes this machine's down pair:3 count as work in flight, and --apply resumes it when a local conversation is proven. Filter to this machine, or document and pin that sdlc reports only local owners.
- **BR-4** [Minor] `bounded-external-calls` (kind: defect) execResume.Candidates has no timeout, so a hung couch hangs fleet resume
  fleet.go:42 uses exec.Command without a context, unlike Resume (3 min) and the SDLC seam (60 s).
- **BR-5** [Minor] `error-cause-preserved` (kind: defect) execResume.Resume reports an exec failure that is not an exit error as "refused: (none)"
  fleet.go:75-77 drops err when stderr is empty (for example, couch not on PATH). Include err in the message.
- **BR-6** [Minor] `empty-result-says-why` (kind: defect) team list prints nothing when run from a checkout whose parent folder is not the workspace
  From a worktree slot, readHubs scans the slot folder, finds no .hub, and the command prints nothing. It should say no teams were found in that folder.

## Round 2 — 2026-10-19T14:10:47-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Claims now come from inventory rows; a claims_state other than present or absent, or a missing row, is unknown. TestDecodeFleetClaims and TestFleetResumeDryRun go red when the fix is reverted.
- BR-2 — addressed — utf8.Valid added in team.go parseHub; the 0xff vector is restored and TestHubClaimVectors goes red without the check.
- BR-3 — addressed — tl-lead's owner match is deleted; sdlc PlaceClaims places only claims whose claimant is this machine's fingerprint and worktree (cmd/sdlc/internal/fleet/claims.go mine).
- BR-4 — not-addressed — Candidates now uses CommandContext with a 1-minute timeout, but no test pins it and the timeout cannot be injected; make the timeout injectable and test it with a hanging fake binary.
- BR-5 — not-addressed — The code fix is right (fleet.go:85-90, "did not run"), but there is no regression test; execResume with a nonexistent bin asserting "did not run" is cheap.
- BR-6 — addressed — renderTeams prints "no teams" for an empty hub set (tested), and team list names the scanned workspace folder.

### Raised

- **BR-7** [Important] `observation-bound-to-exact-entity` (kind: defect) fleet resume reads a slot's claims from its host row only, ignoring dependency members, so a slot with in-flight work on a dependency clone is classed idle
  This is the 2nd finding in this family. The rule: a slot observation covers the slot's full member set, as sdlc's AssembleSlots (slots.go:213-238) defines it, not just the host. decodeFleet (sdlc.go:169) joins only HostPath, but a claim made in a slot's own dependency clone is recorded on that clone's row. Fix: collect member paths, union their non-terminal claims, and set ClaimsKnown only when every member row is present or absent. Add a fixture with a claim on a dependency member.
- **BR-8** [Minor] `error-cause-preserved` (kind: defect) decodeFleet drops claims_error, so an unknown or stale slot prints no reason
  This is the 2nd finding in this family. The rule: when a decoded read-quality state is not an answer, carry its error text through to the printed reason, as decodeObservation does with Card.Error. Add claims_error to the Rows struct, include it in ClaimsWhy, and sweep the other decoders in sdlc.go and team_issues.go.

## Round 3 — 2026-10-19T14:13:23-07:00 (claude) — passed

### Disposed

- BR-4 — addressed — fleet.go:55-71 runs Candidates under a context with a 1-minute default limit (injectable); TestExecResumeBoundsAndCauses times out a hung script.
- BR-5 — addressed — fleet.go:99-103 reports a non-exit error as "did not run: <err>" and a refusal as "refused (<err>): <stderr>"; both paths are covered by the missing-couch and refuse-script cases.
- BR-7 — addressed — sdlc.go:173-194 joins claims over every member row, with ClaimsKnown only when all are read; the ariadne:2 fixture holds a claim on a dependency clone only, and a host-only join would fail it.
- BR-8 — addressed — claims_error is decoded (sdlc.go:121), added to the state, and printed through ClaimsWhy; the test asserts "unknown: tracker fetch failed". Sibling decoders already carry their errors (Card.Error, owner_error).

### Raised

- **BR-9** [Minor] `consume-producer-verdict` (kind: defect) A declared but missing dependency clone makes the slot's claims unknown, so fleet resume never resumes it
  sdlc's AssembleSlots (slots.go:232-234) lists a MemberMissing dependency with verdict "missing" and no tree row. decodeFleet (sdlc.go:179-182) reads the missing row as an unread member and sets ClaimsKnown=false. Fix: read members[].verdict and treat a "missing" member as holding no claims. Add a fixture.

## Round 4 — 2026-10-19T14:15:25-07:00 (claude) — passed

### Disposed

- BR-9 — addressed — sdlc.go:182 exempts members whose producer verdict is "missing" (slots.go:23,52,232-234); fixture ariadne:5 in fleet_test.go fails without it; go test passes.

## Open findings

(none — every finding has been disposed)
