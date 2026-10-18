---
gate: boundary-review
issue: 431
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: c73562b8-3aa6-4689-b9c1-a2b6d004e780
            repository: github.com/xianxu/ariadne
            issue_id: "000431"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/514707126d234259
            reviewed_head: 98bb4a04bde03b3e74e1736ee4b3917101154416
            slot: ariadne:2
            requested_at: "2026-10-18T02:35:17.936734Z"
          phase: finished
          pid: 41616
          started_at: "2026-10-18T02:35:18.060769Z"
          finished_at: "2026-10-18T02:43:03.853961Z"
          session_id: ecc61eb5-39f4-4793-b767-4d02c929af1c
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 98bb4a04bde03b3e74e1736ee4b3917101154416
        tracker_head: c20d4d351d24020da3f80c7e5b197e339991920d
        branch_point: 32fb3f7d056d59bc46d88e2b2d81cb035c4f2d51
        since: "2026-10-17T18:23:15-07:00"
        until: "2026-10-17T19:34:58-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T18:23:15-07:00"
              end: "2026-10-17T19:34:58-07:00"
              began: "2026-10-17T18:23:15-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 8db5d7fb4182d3079059165d1f3997992bcd6ef03ae97cddeffb0bd2560fb966
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 88902369
              windows:
                - start: "2026-10-17T18:23:15-07:00"
                  end: "2026-10-17T19:34:58-07:00"
        queue_waits:
            - attempt_id: 5c416e294bba78f2ce49c64c087a31fb
              slot: ariadne:2
              record_digest: 9aeda882ff0ad3c99a95d32773e4464fc46e4056cd5b12e37f29a62a56227219
              start: "2026-10-18T02:08:21.946098Z"
              end: "2026-10-18T02:26:06.484583Z"
              queue_seconds: 1064.538485
              raw_seconds: 1064.538485
              exact: true
        excluded_brain_sessions: 1
        excluded_brain_minutes: 3.338233333333333
      "n": 1
      timestamp: "2026-10-17T19:43:06-07:00"
      dispatched: "2026-10-17T19:35:17-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: The republish verb's own code (runLocalCheckRepublish) has no test, despite plan item 2 claiming a fake-GH test of the verb
          detail: 'The tests call loadSavedAttempt and postLocalCheckRecord directly, so --rev resolution, the store path from localCheckLogDir, the PR lookup for the failure comment, the missing --name refusal and the error wrapping in localcheck_cmd.go:183-222 are never run. Add a test that swaps ghClient for ciFakeGH and calls runLocalCheckRepublish: a pass gets posted, another --rev is refused, a failure posts a failure status plus a PR comment.'
          family: plan-claimed-test-missing
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: The status-post backoff (about 2.5 min) holds the landing turn on the merge paths, contradicting the spec's claim that it holds nothing scarce
          detail: newLocalCheckRunner, called from landing.go:517 and merge.go:623, now waits up to about 2.5 minutes inside waitCIGate while GitHub is unreachable; before this change it failed at once. Correct the claim or use a shorter backoff for merge.
          family: operating-envelope-claim
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: No test covers merge's republish post failing (stop with the recovery line instead of rerunning)
          detail: localcheck_republish.go:165 returns ok=true with an error so the gate stops; no fixture makes that post fail.
          family: untested-error-path
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: savedAttemptLog hard-codes the reverse of publishSuiteAttempt's log layout; the PR lookup copies the run verb's
          detail: Put a log-path resolver next to teststats_store.go so the layout has one source, and share the branch-to-PR lookup.
          family: duplicated-layout-knowledge
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: '"Newest that verifies" can choose an older pass over a newer failure whose log was pruned, and --attempt can overwrite a posted failure'
          detail: It is the same tree, so this only matters for flaky tests; say so in the help text, or have merge refuse to reuse when a newer attempt exists but does not verify.
          family: attempt-selection-semantics
          round: 1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: e8b024d2-76e2-4d90-9a7c-307013a4e8bb
            repository: github.com/xianxu/ariadne
            issue_id: "000431"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/16a783c437ec35a6
            reviewed_head: 906790a75819bfb7cc96924294914a61955e6e2b
            slot: ariadne:2
            requested_at: "2026-10-18T03:12:28.228565Z"
          phase: finished
          pid: 22613
          started_at: "2026-10-18T03:12:28.321564Z"
          finished_at: "2026-10-18T03:14:17.853227Z"
          session_id: 98f8c616-8a4a-4839-927e-cccf505ce969
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 906790a75819bfb7cc96924294914a61955e6e2b
        tracker_head: 2bdfd1f303db53769964601dd6acbbcc97286846
        branch_point: 32fb3f7d056d59bc46d88e2b2d81cb035c4f2d51
        since: "2026-10-17T18:23:15-07:00"
        until: "2026-10-17T19:45:41-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T18:23:15-07:00"
              end: "2026-10-17T19:45:41-07:00"
              began: "2026-10-17T18:23:15-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 8ad2f999565753c6a7de9d3557d557e73cf0c5ca02bf3b3e209dad3a5a5abd02
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 90494892
              windows:
                - start: "2026-10-17T18:23:15-07:00"
                  end: "2026-10-17T19:45:41-07:00"
        queue_waits:
            - attempt_id: 5c416e294bba78f2ce49c64c087a31fb
              slot: ariadne:2
              record_digest: 9aeda882ff0ad3c99a95d32773e4464fc46e4056cd5b12e37f29a62a56227219
              start: "2026-10-18T02:08:21.946098Z"
              end: "2026-10-18T02:26:06.484583Z"
              queue_seconds: 1064.538485
              raw_seconds: 1064.538485
              exact: true
        excluded_brain_sessions: 1
        excluded_brain_minutes: 23.38693333333333
        cuts:
            - dispatched: "2026-10-17T19:35:17-07:00"
              finalized: "2026-10-17T19:43:06-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-17T20:14:18-07:00"
      dispatched: "2026-10-17T20:12:28-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'TestLocalCheckRepublishCmd calls runLocalCheckRepublish with a swapped ghClient/detectRepo: missing --name, failure posts failure plus PR comment, other commit and bad --rev refuse; passes at head.'
          round: 2
        - id: BR-2
          disposition: addressed
          note: Spec, help text (local-check.md REPUBLISH) and the localCheckPostBackoff comment now state the wait holds the landing turn under merge; prose-only correction verified against cigate.go newLocalCheckRunner.
          round: 2
        - id: BR-3
          disposition: addressed
          note: TestLocalCheckRunnerSavedPassEdges subtest 1 makes the republish post fail and asserts the recovery line and no rerun.
          round: 2
        - id: BR-4
          disposition: addressed
          note: suiteAttemptLogFile now sits beside publishSuiteAttempt in teststats_store.go; currentBranchPR is shared by run and republish.
          round: 2
        - id: BR-5
          disposition: addressed
          note: Merge requires newest (loadSavedAttempt reports it); help text documents the verb's newest-that-verifies caveat; removing the !newest guard turns TestLocalCheckRunnerSavedPassEdges red.
          round: 2
      findings:
        - kind: defect
          id: BR-6
          severity: Important
          title: README and atlas ci-merge-check.md not updated for local-check republish and merge's saved-pass reuse
          detail: README.md:471-478 documents local-check run and cancel but not the new republish subcommand; atlas/workflow/ci-merge-check.md:160 still says merge runs a missing check itself, though newLocalCheckRunner now first posts a saved verified newest pass. Add a README sentence for the verb and update the ci-merge-check Merge bullet.
          family: docs-surface-not-updated
          round: 2
        - kind: defect
          id: BR-7
          severity: Minor
          title: The newest-attempt edge test passes only because the true and false checks write identical empty logs
          detail: 'This is the 2nd finding in family untested-error-path. The rule: every guard branch needs a fixture that reaches it on purpose, not by chance. Here a failed attempt overwrites the shared local-checks/<sha>/<check>.log, so the earlier pass verifies only because both logs are empty. Use distinct output and assert the not-newest branch explicitly (localcheck_republish_test.go:257-271).'
          family: untested-error-path
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: ef69055e-db06-484e-991c-2042eb84bfb8
            repository: github.com/xianxu/ariadne
            issue_id: "000431"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/a6d15915ed9f7f19
            reviewed_head: 5b2d792df73a42e43e6d71d6e1b22121a54e5462
            slot: ariadne:2
            requested_at: "2026-10-18T03:40:34.054194Z"
          phase: finished
          pid: 60751
          started_at: "2026-10-18T03:40:34.21089Z"
          finished_at: "2026-10-18T03:47:01.468032Z"
          session_id: fc157129-5341-4ccf-8cca-1c2dfbf0c6c8
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 5b2d792df73a42e43e6d71d6e1b22121a54e5462
        tracker_head: 948dd11dd59e2d2412bbe183781b3194004170e2
        branch_point: 32fb3f7d056d59bc46d88e2b2d81cb035c4f2d51
        since: "2026-10-17T18:23:15-07:00"
        until: "2026-10-17T20:15:18-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T18:23:15-07:00"
              end: "2026-10-17T20:15:18-07:00"
              began: "2026-10-17T18:23:15-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f1481145e1337b7bdd98337559e39fcef334b109fc58d0bd6abf237b4386e2a7
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 91515816
              windows:
                - start: "2026-10-17T18:23:15-07:00"
                  end: "2026-10-17T20:15:18-07:00"
        queue_waits:
            - attempt_id: 5c416e294bba78f2ce49c64c087a31fb
              slot: ariadne:2
              record_digest: 9aeda882ff0ad3c99a95d32773e4464fc46e4056cd5b12e37f29a62a56227219
              start: "2026-10-18T02:08:21.946098Z"
              end: "2026-10-18T02:26:06.484583Z"
              queue_seconds: 1064.538485
              raw_seconds: 1064.538485
              exact: true
            - attempt_id: 04f0a8ee91a1b328c28b0d687ea87055
              slot: ariadne:2
              record_digest: dc16a7b09a9fc2f2adb30ca37ae1b4d33006c8086770ae036abd8f28fe5052eb
              start: "2026-10-18T02:45:49.932166Z"
              end: "2026-10-18T03:03:15.535511Z"
              queue_seconds: 1045.603345
              raw_seconds: 1045.603345
              exact: true
        excluded_brain_sessions: 1
        excluded_brain_minutes: 45.42231666666667
        cuts:
            - dispatched: "2026-10-17T19:35:17-07:00"
              finalized: "2026-10-17T19:43:06-07:00"
              finalizing: false
            - dispatched: "2026-10-17T20:12:28-07:00"
              finalized: "2026-10-17T20:14:18-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-17T20:47:03-07:00"
      dispatched: "2026-10-17T20:40:34-07:00"
      agent: claude
      dispose:
        - id: BR-6
          disposition: addressed
          note: README.md:472-481 names the republish verb and merge's saved-pass-first rule; atlas ci-merge-check.md:160-162 and help HOW MERGE USES IT updated likewise.
          round: 3
        - id: BR-7
          disposition: addressed
          note: Test uses distinct pass/fail output, restores pass bytes, prunes the failure log, asserts the guard message; reverting the newest guard turns it red (localcheck_republish_test.go:293).
          round: 3
      findings:
        - kind: defect
          id: BR-8
          severity: Minor
          title: loadSavedAttempt computes "newest" over a possibly truncated store read and discards the reader's warnings
          detail: '2nd in family attempt-selection-semantics. Rule: when the store read is incomplete, the newest attempt is unknown, not absent. readSuiteAttempts stops at testStatsRecordLimit or the byte budget, reading in directory order, and returns warnings that loadSavedAttempt (localcheck_republish.go:277) drops, so merge could treat an older pass as newest. Fix: report newest=false when warnings are present.'
          family: attempt-selection-semantics
          round: 3
      recipe: milestone-review
      isolation: fenced
      reviewed: 5b2d792df73a42e43e6d71d6e1b22121a54e5462
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 15c82fa1-3078-4147-9a01-6b42811c30a6
            repository: github.com/xianxu/ariadne
            issue_id: "000431"
            recipe: fix-review
            round: 4
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/2a187b74d4b551c4
            reviewed_head: 391c17e163125b715265c3952ba0c1e1c14722c4
            slot: ariadne:2
            requested_at: "2026-10-18T03:49:06.310078Z"
          phase: finished
          pid: 81027
          started_at: "2026-10-18T03:49:06.504335Z"
          finished_at: "2026-10-18T03:50:03.437823Z"
          session_id: 815fe8de-8afc-4d34-8ec8-5507dbda03c0
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 391c17e163125b715265c3952ba0c1e1c14722c4
        tracker_head: cf0da49e86f8663accedae7646a7e26ce72cfaa6
        branch_point: b6a22b0bff4b6f37deb3046f3c2faf92552784dc
        since: "2026-10-17T18:23:15-07:00"
        until: "2026-10-17T20:47:46-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T18:23:15-07:00"
              end: "2026-10-17T20:47:46-07:00"
              began: "2026-10-17T18:23:15-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6e84f528f85d3399c0e189e98d06a219140b024bb43c932ea66865f1154c8abd
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 92597778
              windows:
                - start: "2026-10-17T18:23:15-07:00"
                  end: "2026-10-17T20:47:46-07:00"
        queue_waits:
            - attempt_id: 5c416e294bba78f2ce49c64c087a31fb
              slot: ariadne:2
              record_digest: 9aeda882ff0ad3c99a95d32773e4464fc46e4056cd5b12e37f29a62a56227219
              start: "2026-10-18T02:08:21.946098Z"
              end: "2026-10-18T02:26:06.484583Z"
              queue_seconds: 1064.538485
              raw_seconds: 1064.538485
              exact: true
            - attempt_id: 04f0a8ee91a1b328c28b0d687ea87055
              slot: ariadne:2
              record_digest: dc16a7b09a9fc2f2adb30ca37ae1b4d33006c8086770ae036abd8f28fe5052eb
              start: "2026-10-18T02:45:49.932166Z"
              end: "2026-10-18T03:03:15.535511Z"
              queue_seconds: 1045.603345
              raw_seconds: 1045.603345
              exact: true
            - attempt_id: d0a0275e7866123956e54049f9aa54fb
              slot: ariadne:2
              record_digest: b5a6ac0a3bc843ddc198a77ba6a3fbff70c72ab345a57530b336bc4229ea0e07
              start: "2026-10-18T03:15:21.136995Z"
              end: "2026-10-18T03:31:11.65677Z"
              queue_seconds: 950.519775
              raw_seconds: 950.519775
              exact: true
        excluded_brain_sessions: 1
        excluded_brain_minutes: 45.42231666666667
        cuts:
            - dispatched: "2026-10-17T19:35:17-07:00"
              finalized: "2026-10-17T19:43:06-07:00"
              finalizing: false
            - dispatched: "2026-10-17T20:12:28-07:00"
              finalized: "2026-10-17T20:14:18-07:00"
              finalizing: false
            - dispatched: "2026-10-17T20:40:34-07:00"
              finalized: "2026-10-17T20:47:03-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-17T20:50:04-07:00"
      dispatched: "2026-10-17T20:49:06-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: addressed
          note: localcheck_republish.go:65,97 sets newest=false whenever readSuiteAttempts warns (any limit, budget, open, read or decode warning); the new subtest at localcheck_republish_test.go:299 goes red with newest=true when the guard is reverted.
          round: 4
      recipe: fix-review
      isolation: fenced
      reviewed: 391c17e163125b715265c3952ba0c1e1c14722c4
      blocked: false
---

# Gate ledger — ariadne#431 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T19:43:06-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `plan-claimed-test-missing` (kind: defect) The republish verb's own code (runLocalCheckRepublish) has no test, despite plan item 2 claiming a fake-GH test of the verb
  The tests call loadSavedAttempt and postLocalCheckRecord directly, so --rev resolution, the store path from localCheckLogDir, the PR lookup for the failure comment, the missing --name refusal and the error wrapping in localcheck_cmd.go:183-222 are never run. Add a test that swaps ghClient for ciFakeGH and calls runLocalCheckRepublish: a pass gets posted, another --rev is refused, a failure posts a failure status plus a PR comment.
- **BR-2** [Minor] `operating-envelope-claim` (kind: defect) The status-post backoff (about 2.5 min) holds the landing turn on the merge paths, contradicting the spec's claim that it holds nothing scarce
  newLocalCheckRunner, called from landing.go:517 and merge.go:623, now waits up to about 2.5 minutes inside waitCIGate while GitHub is unreachable; before this change it failed at once. Correct the claim or use a shorter backoff for merge.
- **BR-3** [Minor] `untested-error-path` (kind: defect) No test covers merge's republish post failing (stop with the recovery line instead of rerunning)
  localcheck_republish.go:165 returns ok=true with an error so the gate stops; no fixture makes that post fail.
- **BR-4** [Minor] `duplicated-layout-knowledge` (kind: defect) savedAttemptLog hard-codes the reverse of publishSuiteAttempt's log layout; the PR lookup copies the run verb's
  Put a log-path resolver next to teststats_store.go so the layout has one source, and share the branch-to-PR lookup.
- **BR-5** [Minor] `attempt-selection-semantics` (kind: defect) "Newest that verifies" can choose an older pass over a newer failure whose log was pruned, and --attempt can overwrite a posted failure
  It is the same tree, so this only matters for flaky tests; say so in the help text, or have merge refuse to reuse when a newer attempt exists but does not verify.

## Round 2 — 2026-10-17T20:14:18-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — TestLocalCheckRepublishCmd calls runLocalCheckRepublish with a swapped ghClient/detectRepo: missing --name, failure posts failure plus PR comment, other commit and bad --rev refuse; passes at head.
- BR-2 — addressed — Spec, help text (local-check.md REPUBLISH) and the localCheckPostBackoff comment now state the wait holds the landing turn under merge; prose-only correction verified against cigate.go newLocalCheckRunner.
- BR-3 — addressed — TestLocalCheckRunnerSavedPassEdges subtest 1 makes the republish post fail and asserts the recovery line and no rerun.
- BR-4 — addressed — suiteAttemptLogFile now sits beside publishSuiteAttempt in teststats_store.go; currentBranchPR is shared by run and republish.
- BR-5 — addressed — Merge requires newest (loadSavedAttempt reports it); help text documents the verb's newest-that-verifies caveat; removing the !newest guard turns TestLocalCheckRunnerSavedPassEdges red.

### Raised

- **BR-6** [Important] `docs-surface-not-updated` (kind: defect) README and atlas ci-merge-check.md not updated for local-check republish and merge's saved-pass reuse
  README.md:471-478 documents local-check run and cancel but not the new republish subcommand; atlas/workflow/ci-merge-check.md:160 still says merge runs a missing check itself, though newLocalCheckRunner now first posts a saved verified newest pass. Add a README sentence for the verb and update the ci-merge-check Merge bullet.
- **BR-7** [Minor] `untested-error-path` (kind: defect) The newest-attempt edge test passes only because the true and false checks write identical empty logs
  This is the 2nd finding in family untested-error-path. The rule: every guard branch needs a fixture that reaches it on purpose, not by chance. Here a failed attempt overwrites the shared local-checks/<sha>/<check>.log, so the earlier pass verifies only because both logs are empty. Use distinct output and assert the not-newest branch explicitly (localcheck_republish_test.go:257-271).

## Round 3 — 2026-10-17T20:47:03-07:00 (claude) — passed

### Disposed

- BR-6 — addressed — README.md:472-481 names the republish verb and merge's saved-pass-first rule; atlas ci-merge-check.md:160-162 and help HOW MERGE USES IT updated likewise.
- BR-7 — addressed — Test uses distinct pass/fail output, restores pass bytes, prunes the failure log, asserts the guard message; reverting the newest guard turns it red (localcheck_republish_test.go:293).

### Raised

- **BR-8** [Minor] `attempt-selection-semantics` (kind: defect) loadSavedAttempt computes "newest" over a possibly truncated store read and discards the reader's warnings
  2nd in family attempt-selection-semantics. Rule: when the store read is incomplete, the newest attempt is unknown, not absent. readSuiteAttempts stops at testStatsRecordLimit or the byte budget, reading in directory order, and returns warnings that loadSavedAttempt (localcheck_republish.go:277) drops, so merge could treat an older pass as newest. Fix: report newest=false when warnings are present.

## Round 4 — 2026-10-17T20:50:04-07:00 (claude) — passed

### Disposed

- BR-8 — addressed — localcheck_republish.go:65,97 sets newest=false whenever readSuiteAttempts warns (any limit, budget, open, read or decode warning); the new subtest at localcheck_republish_test.go:299 goes red with newest=true when the guard is reverted.

## Open findings

(none — every finding has been disposed)
