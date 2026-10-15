---
gate: boundary-review
issue: 376
id_prefix: BR
rounds:
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 44c39478add6be7e5f8c56ace079b462926dea24
        tracker_head: f2ce7c5302e7e055d5746fd427ce5a898614c9dc
        branch_point: 18915f6c17d9b4aeb5e618d3f736705ec355d2b3
        since: "2026-10-14T20:40:09-07:00"
        until: "2026-10-14T22:18:14-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-14T20:40:09-07:00"
              end: "2026-10-14T22:18:14-07:00"
              began: "2026-10-14T20:40:09-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: fd13ed455b43e3ad7b9bbf4961eaafbd0c2758427a8def26ef6cfc92a90f8ed9
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 28671020
              windows:
                - start: "2026-10-14T20:40:09-07:00"
                  end: "2026-10-14T22:18:14-07:00"
            - harness: codex
              id: 01a13db1-fa52-7b83-bef0-9e0cbb5fed94
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: c0835be6d127033fc4026a726d32df65eaca885e9c30d55e73bbd3e4e0338f45
              started: "2026-10-15T03:53:57.198Z"
              spawned: false
              bytes: 540208
              windows:
                - start: "2026-10-15T03:53:57.198Z"
                  end: "2026-10-14T22:18:14-07:00"
            - harness: codex
              id: 01a13db4-79e3-73e3-b1b7-22d2f5b5b8cb
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 722aea63397d119b0d35afbd59ef2b50f11c19572fad617d61b01bbd36d5e3ec
              started: "2026-10-15T03:56:40.87Z"
              spawned: false
              bytes: 455430
              windows:
                - start: "2026-10-15T03:56:40.87Z"
                  end: "2026-10-14T22:18:14-07:00"
            - harness: codex
              id: 01a13db6-157f-7013-8b8e-ee975fe00bbd
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 959dbf7a68dc268e0a4b1e520466c83d9a89997094e2ffc6602b21ed0ed76ec4
              started: "2026-10-15T03:58:26.38Z"
              spawned: false
              bytes: 416762
              windows:
                - start: "2026-10-15T03:58:26.38Z"
                  end: "2026-10-14T22:18:14-07:00"
            - harness: codex
              id: 01a13db6-8751-73b1-ac45-2cfd823cd708
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: d7d23c22a882b3c12e70d04c8dd72742f868eba3238cda10a48a3f1744fa6e56
              started: "2026-10-15T03:58:55.257Z"
              spawned: false
              bytes: 406150
              windows:
                - start: "2026-10-15T03:58:55.257Z"
                  end: "2026-10-14T22:18:14-07:00"
            - harness: codex
              id: 01a13db7-1c66-7172-bde5-115ba37d9e78
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              digest: f3d7fa22ea010167005b6e3c587966845e13959f95bdbc1ea9a62f72eb30c076
              started: "2026-10-15T03:59:33.078Z"
              spawned: true
              bytes: 955056
              windows:
                - start: "2026-10-15T03:59:33.078Z"
                  end: "2026-10-14T22:18:14-07:00"
      "n": 1
      timestamp: "2026-10-14T22:22:10-07:00"
      dispatched: "2026-10-14T22:19:15-07:00"
      agent: codex
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: The parser silently ignores the Spec's go test -json format
          detail: 'cmd/sdlc/teststats_parse.go:34 recognizes only plain-text summaries; JSON package/test events produce empty failures and units without warnings. TestBoundary376JSON reproduced this on the pinned head using valid test and package fail events. ARCH-PURPOSE: support the contracted JSON format, including package timings and package-qualified failure identities, with regression fixtures for pass, fail, and malformed events.'
          family: contracted-log-format-coverage
          round: 1
        - kind: defect
          id: BR-2
          severity: Critical
          title: A non-suite attempt hides the legacy suite on the same SHA
          detail: 'cmd/sdlc/teststats_cmd.go:138 marks every attempt SHA as known before filtering by check name. Since runLocalCheck records non-suite checks too, a smoke record suppresses an existing suite.log; aggregation then discards the smoke record. TestBoundary376Legacy reproduced an empty repository report with no warning despite an available suite log. ARCH-PURPOSE: deduplicate by the covered check and SHA, and add a mixed-check regression test.'
          family: evidence-deduplication-scope
          round: 1
        - kind: defect
          id: BR-3
          severity: Critical
          title: The Core concepts table names the wrong location for SuiteAttempt
          detail: workshop/plans/000376-test-stats-plan.md:15 places SuiteAttempt in teststats_parse.go, but its definition is cmd/sdlc/teststats.go:11. This contradicts the required Core concepts cross-check. Correct the location and append a Revisions entry; make PURE and INTEGRATION classifications explicit in the inventory. This is a documentation correction and needs no wording-presence test.
          family: core-concept-inventory-accuracy
          round: 1
        - kind: defect
          id: BR-4
          severity: Important
          title: Crash-left attempt publication files have no removal path
          detail: 'cmd/sdlc/teststats_store.go:225 creates attempts/.attempt-* files whose cleanup depends on a defer. Process death bypasses that defer, while pruneSuiteAttempts skips every non-.json file at line 352 and only otherwise sweeps inflight directories. ARCH-FUNERAL: associate publication temporaries with attempt ownership and reclaim aged dead-owner residue, with a crash fixture that also proves active writers remain protected.'
          family: crash-artifact-lifecycle
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#376 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T22:22:10-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `contracted-log-format-coverage` (kind: defect) The parser silently ignores the Spec's go test -json format
  cmd/sdlc/teststats_parse.go:34 recognizes only plain-text summaries; JSON package/test events produce empty failures and units without warnings. TestBoundary376JSON reproduced this on the pinned head using valid test and package fail events. ARCH-PURPOSE: support the contracted JSON format, including package timings and package-qualified failure identities, with regression fixtures for pass, fail, and malformed events.
- **BR-2** [Critical] `evidence-deduplication-scope` (kind: defect) A non-suite attempt hides the legacy suite on the same SHA
  cmd/sdlc/teststats_cmd.go:138 marks every attempt SHA as known before filtering by check name. Since runLocalCheck records non-suite checks too, a smoke record suppresses an existing suite.log; aggregation then discards the smoke record. TestBoundary376Legacy reproduced an empty repository report with no warning despite an available suite log. ARCH-PURPOSE: deduplicate by the covered check and SHA, and add a mixed-check regression test.
- **BR-3** [Critical] `core-concept-inventory-accuracy` (kind: defect) The Core concepts table names the wrong location for SuiteAttempt
  workshop/plans/000376-test-stats-plan.md:15 places SuiteAttempt in teststats_parse.go, but its definition is cmd/sdlc/teststats.go:11. This contradicts the required Core concepts cross-check. Correct the location and append a Revisions entry; make PURE and INTEGRATION classifications explicit in the inventory. This is a documentation correction and needs no wording-presence test.
- **BR-4** [Important] `crash-artifact-lifecycle` (kind: defect) Crash-left attempt publication files have no removal path
  cmd/sdlc/teststats_store.go:225 creates attempts/.attempt-* files whose cleanup depends on a defer. Process death bypasses that defer, while pruneSuiteAttempts skips every non-.json file at line 352 and only otherwise sweeps inflight directories. ARCH-FUNERAL: associate publication temporaries with attempt ownership and reclaim aged dead-owner residue, with a crash fixture that also proves active writers remain protected.

## Open findings

- **BR-1** [Critical] `contracted-log-format-coverage` (kind: defect) The parser silently ignores the Spec's go test -json format
- **BR-2** [Critical] `evidence-deduplication-scope` (kind: defect) A non-suite attempt hides the legacy suite on the same SHA
- **BR-3** [Critical] `core-concept-inventory-accuracy` (kind: defect) The Core concepts table names the wrong location for SuiteAttempt
- **BR-4** [Important] `crash-artifact-lifecycle` (kind: defect) Crash-left attempt publication files have no removal path
