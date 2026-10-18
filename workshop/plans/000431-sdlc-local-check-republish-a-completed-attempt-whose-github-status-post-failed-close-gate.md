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

## Open findings

- **BR-1** [Important] `plan-claimed-test-missing` (kind: defect) The republish verb's own code (runLocalCheckRepublish) has no test, despite plan item 2 claiming a fake-GH test of the verb
- **BR-2** [Minor] `operating-envelope-claim` (kind: defect) The status-post backoff (about 2.5 min) holds the landing turn on the merge paths, contradicting the spec's claim that it holds nothing scarce
- **BR-3** [Minor] `untested-error-path` (kind: defect) No test covers merge's republish post failing (stop with the recovery line instead of rerunning)
- **BR-4** [Minor] `duplicated-layout-knowledge` (kind: defect) savedAttemptLog hard-codes the reverse of publishSuiteAttempt's log layout; the PR lookup copies the run verb's
- **BR-5** [Minor] `attempt-selection-semantics` (kind: defect) "Newest that verifies" can choose an older pass over a newer failure whose log was pruned, and --attempt can overwrite a posted failure
