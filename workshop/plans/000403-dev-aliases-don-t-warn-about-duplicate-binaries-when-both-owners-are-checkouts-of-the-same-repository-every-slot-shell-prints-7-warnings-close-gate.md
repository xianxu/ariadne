---
gate: boundary-review
issue: 403
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 4ee188b1-4e57-4072-a142-6a5a83652279
            repository: github.com/xianxu/ariadne
            issue_id: "000403"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/c79d45ee4a30e25c
            reviewed_head: c64466ea43415a2281cd773e9aa50be67d1c6451
            slot: ariadne:1
            requested_at: "2026-10-17T22:59:17.015214Z"
          phase: finished
          pid: 22621
          started_at: "2026-10-17T22:59:17.109274Z"
          finished_at: "2026-10-17T23:00:47.037027Z"
          session_id: 01a14c17-4adc-7a93-976a-a15088310af0
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: c64466ea43415a2281cd773e9aa50be67d1c6451
        tracker_head: 5b6dcd4f78c4b9c3d10bff111bdc0a66f78fd38e
        branch_point: 75b21892013595a0101c044ec83446353b789415
        since: "2026-10-17T15:20:14-07:00"
        until: "2026-10-17T15:29:29-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T15:20:14-07:00"
              end: "2026-10-17T15:29:29-07:00"
              began: "2026-10-17T15:20:14-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: c7815c267460e1945d5f793ae5f0fec379b0f244565c5f4f0f1d7640d2fd9a49
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 72834792
              windows:
                - start: "2026-10-17T15:20:14-07:00"
                  end: "2026-10-17T15:29:29-07:00"
      "n": 1
      timestamp: "2026-10-17T16:00:47-07:00"
      dispatched: "2026-10-17T15:59:17-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Important
          title: README update appears missing for dev-aliases --verbose
          detail: construct/dev-aliases.sh:62 introduces --verbose, but README.md is unchanged. This is the only new CLI flag in the window. Add its purpose and a diagnostic invocation near README.md:342; explain the revised --strict collision semantics there as well.
          family: user-surface-documentation
          round: 1
        - id: BR-2
          severity: Important
          title: Acceptance criteria and fixtures omit the revised origin policy
          detail: Issue 000403:44-45 retains the original Done-when despite the identity-policy revision at :67-69. construct/scripts/test/dev-aliases.test.sh:198-244 covers GitHub HTTPS/scp-style equivalence, distinct GitHub origins, and absent origins, but not identical non-GitHub URLs, non-GitHub transport differences, non-GitHub path-case differences, or unrecognized origins. Enumerate these cases in Done-when and add fixtures asserting stderr and strict status; the GitHub ssh:// branch at construct/dev-aliases.sh:213 also lacks a fixture.
          family: acceptance-contract-coverage
          round: 1
        - id: BR-3
          severity: Minor
          title: 'ARCH-DRY: warning and strict status independently classify collisions'
          detail: construct/dev-aliases.sh:267 counts identities using sort -u, while :274 independently groups them in awk for strict status. These are both instances of the same collision decision. Compute classification once and derive diagnostics and exit status from it.
          family: single-source-collision-policy
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 4777a6a3-40b3-4791-aa0f-ffabbb0ae88d
            repository: github.com/xianxu/ariadne
            issue_id: "000403"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/6b9584961e75a262
            reviewed_head: 95b82627c2109ab7a8f9dc0d50dddf08eb2ef959
            slot: ariadne:1
            requested_at: "2026-10-17T23:43:43.691335Z"
          phase: finished
          pid: 43205
          started_at: "2026-10-17T23:43:43.788667Z"
          finished_at: "2026-10-17T23:45:04.504798Z"
          session_id: 01a14c3f-fb9c-7dd2-9259-9f7bdee76bcf
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 95b82627c2109ab7a8f9dc0d50dddf08eb2ef959
        tracker_head: 5b6dcd4f78c4b9c3d10bff111bdc0a66f78fd38e
        branch_point: 75b21892013595a0101c044ec83446353b789415
        since: "2026-10-17T15:20:14-07:00"
        until: "2026-10-17T16:01:54-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T15:20:14-07:00"
              end: "2026-10-17T16:01:54-07:00"
              began: "2026-10-17T15:20:14-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: e79c60636a27786ff92019e506b1df9daa6eca6cd8c3cddc360b1d6a45e8898c
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 73423475
              windows:
                - start: "2026-10-17T15:20:14-07:00"
                  end: "2026-10-17T16:01:54-07:00"
        queue_waits:
            - attempt_id: 567468f762223462f31b84c32cd1ae06
              slot: ariadne:1
              record_digest: 8b485e6b47524914fe85a03c73dd335032eb1c7e1af7c09cc01c4c3847292dd1
              start: "2026-10-17T22:29:43.114362Z"
              end: "2026-10-17T22:50:34.522268Z"
              queue_seconds: 1251.407906
              raw_seconds: 1251.407906
              exact: true
        cuts:
            - dispatched: "2026-10-17T15:59:17-07:00"
              finalized: "2026-10-17T16:00:47-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-17T16:45:05-07:00"
      dispatched: "2026-10-17T16:43:43-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: README.md:345-349 now documents the diagnostic invocation and revised strict semantics, matching construct/dev-aliases.sh:267-274 and :358-360.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Issue Done when and Revisions enumerate the requested origin cases; tests at construct/scripts/test/dev-aliases.test.sh:218-273 exercise them. Removing ssh:// recognition or non-GitHub exact equality independently makes the suite fail.
          round: 2
        - id: BR-3
          disposition: addressed
          note: construct/dev-aliases.sh:262-278 computes classification once; the same identity-count branch controls warnings and had_dup, which determines strict status.
          round: 2
      findings:
        - id: BR-4
          severity: Important
          title: Verbose stdout preservation is accepted but untested
          detail: 'construct/scripts/test/dev-aliases.test.sh:216 discards stdout from the only verbose invocation, so it cannot verify issue Done when :46''s promise that --verbose preserves --list output. Adding unexpected stdout in the verbose branch still produces 75 passed, 0 failed. This is the 2nd finding in family acceptance-contract-coverage. Apply the rule that diagnostic flags preserve machine-readable stdout while reporting on stderr: compare complete list output with and without --verbose, both normally and under --strict, and assert diagnostics separately. The sole affected fixture is :216-217.'
          family: acceptance-contract-coverage
          round: 2
      recipe: small-diff-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#403 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T16:00:47-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `user-surface-documentation` README update appears missing for dev-aliases --verbose
  construct/dev-aliases.sh:62 introduces --verbose, but README.md is unchanged. This is the only new CLI flag in the window. Add its purpose and a diagnostic invocation near README.md:342; explain the revised --strict collision semantics there as well.
- **BR-2** [Important] `acceptance-contract-coverage` Acceptance criteria and fixtures omit the revised origin policy
  Issue 000403:44-45 retains the original Done-when despite the identity-policy revision at :67-69. construct/scripts/test/dev-aliases.test.sh:198-244 covers GitHub HTTPS/scp-style equivalence, distinct GitHub origins, and absent origins, but not identical non-GitHub URLs, non-GitHub transport differences, non-GitHub path-case differences, or unrecognized origins. Enumerate these cases in Done-when and add fixtures asserting stderr and strict status; the GitHub ssh:// branch at construct/dev-aliases.sh:213 also lacks a fixture.
- **BR-3** [Minor] `single-source-collision-policy` ARCH-DRY: warning and strict status independently classify collisions
  construct/dev-aliases.sh:267 counts identities using sort -u, while :274 independently groups them in awk for strict status. These are both instances of the same collision decision. Compute classification once and derive diagnostics and exit status from it.

## Round 2 — 2026-10-17T16:45:05-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — README.md:345-349 now documents the diagnostic invocation and revised strict semantics, matching construct/dev-aliases.sh:267-274 and :358-360.
- BR-2 — addressed — Issue Done when and Revisions enumerate the requested origin cases; tests at construct/scripts/test/dev-aliases.test.sh:218-273 exercise them. Removing ssh:// recognition or non-GitHub exact equality independently makes the suite fail.
- BR-3 — addressed — construct/dev-aliases.sh:262-278 computes classification once; the same identity-count branch controls warnings and had_dup, which determines strict status.

### Raised

- **BR-4** [Important] `acceptance-contract-coverage` Verbose stdout preservation is accepted but untested
  construct/scripts/test/dev-aliases.test.sh:216 discards stdout from the only verbose invocation, so it cannot verify issue Done when :46's promise that --verbose preserves --list output. Adding unexpected stdout in the verbose branch still produces 75 passed, 0 failed. This is the 2nd finding in family acceptance-contract-coverage. Apply the rule that diagnostic flags preserve machine-readable stdout while reporting on stderr: compare complete list output with and without --verbose, both normally and under --strict, and assert diagnostics separately. The sole affected fixture is :216-217.

## Open findings

- **BR-4** [Important] `acceptance-contract-coverage` Verbose stdout preservation is accepted but untested
