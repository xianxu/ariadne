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
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: d7770d349b0e78ed4edc690c375179b330202473
        tracker_head: 2eac65b33290466457f3d0c8d17192c7dab2b1fe
        branch_point: 18915f6c17d9b4aeb5e618d3f736705ec355d2b3
        since: "2026-10-14T20:40:09-07:00"
        until: "2026-10-14T23:08:10-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-14T20:40:09-07:00"
              end: "2026-10-14T23:08:10-07:00"
              began: "2026-10-14T20:40:09-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 5e738ef624b4d04beb9fb0a6dd7ef219e765e428c21a4ed87733a0f98ad11e88
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 29811153
              windows:
                - start: "2026-10-14T20:40:09-07:00"
                  end: "2026-10-14T23:08:10-07:00"
            - harness: codex
              id: 01a13db1-fa52-7b83-bef0-9e0cbb5fed94
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: c0835be6d127033fc4026a726d32df65eaca885e9c30d55e73bbd3e4e0338f45
              started: "2026-10-15T03:53:57.198Z"
              spawned: false
              bytes: 540208
              windows:
                - start: "2026-10-15T03:53:57.198Z"
                  end: "2026-10-14T23:08:10-07:00"
            - harness: codex
              id: 01a13db4-79e3-73e3-b1b7-22d2f5b5b8cb
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 722aea63397d119b0d35afbd59ef2b50f11c19572fad617d61b01bbd36d5e3ec
              started: "2026-10-15T03:56:40.87Z"
              spawned: false
              bytes: 455430
              windows:
                - start: "2026-10-15T03:56:40.87Z"
                  end: "2026-10-14T23:08:10-07:00"
            - harness: codex
              id: 01a13db6-157f-7013-8b8e-ee975fe00bbd
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 959dbf7a68dc268e0a4b1e520466c83d9a89997094e2ffc6602b21ed0ed76ec4
              started: "2026-10-15T03:58:26.38Z"
              spawned: false
              bytes: 416762
              windows:
                - start: "2026-10-15T03:58:26.38Z"
                  end: "2026-10-14T23:08:10-07:00"
            - harness: codex
              id: 01a13db6-8751-73b1-ac45-2cfd823cd708
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: d7d23c22a882b3c12e70d04c8dd72742f868eba3238cda10a48a3f1744fa6e56
              started: "2026-10-15T03:58:55.257Z"
              spawned: false
              bytes: 406150
              windows:
                - start: "2026-10-15T03:58:55.257Z"
                  end: "2026-10-14T23:08:10-07:00"
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
                  end: "2026-10-14T23:08:10-07:00"
            - harness: codex
              id: 01a13e03-2033-7943-b391-58d87df3dfcd
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              digest: 388bf3c948748488691cac68dd3828c5bd200653a0ecc47039df3375208dfc58
              started: "2026-10-15T05:22:34.734Z"
              spawned: true
              bytes: 523047
              windows:
                - start: "2026-10-15T05:22:34.734Z"
                  end: "2026-10-14T23:08:10-07:00"
        cuts:
            - dispatched: "2026-10-14T22:19:15-07:00"
              finalized: "2026-10-14T22:22:10-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-14T23:12:22-07:00"
      dispatched: "2026-10-14T23:08:48-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: JSON pass/fail/malformed fixtures pass on head. Removing JSON event handling makes TestSuiteLogFactsGoTestJSON and TestSuiteLogFactsMalformedGoTestJSONWarnsWithoutInventingData fail.
          round: 2
        - id: BR-2
          disposition: addressed
          note: TestSuiteStatsCommandSmokeRecordDoesNotHideLegacySuite passes on head and fails with the prior unconditional SHA deduplication restored.
          round: 2
        - id: BR-3
          disposition: addressed
          note: The before/after plan diff corrects SuiteAttempt to cmd/sdlc/teststats.go, matching its definition at line 11. The inventory explicitly identifies PURE entities and INTEGRATION persistence, and the close-review Revisions entry records the correction.
          round: 2
        - id: BR-4
          disposition: addressed
          note: Publication temps now carry owner IDs. TestPruneSuiteAttemptsReclaimsOnlyDeadOwnerPublicationTemps protects active writers and reclaims dead-owner residue; restoring the prior store implementation makes it fail.
          round: 2
      findings:
        - kind: defect
          id: BR-5
          severity: Critical
          title: Conflicting JSON field casing can turn a failed attempt into success
          detail: 'cmd/sdlc/teststats_store.go:75 checks duplicate keys case-sensitively, but the struct decode at line 102 accepts case-insensitive field matches. A valid failed record containing "exit":2 followed by "Exit":0 is accepted with Exit=0, despite DisallowUnknownFields. TestBoundary376CaseAlias reproduced this on head. ARCH-SECURE: this admitted record can become false successful-run and flake-confirmation evidence. Reject noncanonical field names before decoding, and test aliases and conflicting keys across outcome/provenance fields.'
          family: persisted-evidence-canonical-decoding
          round: 2
        - kind: defect
          id: BR-6
          severity: Important
          title: Plain-text failed-package summaries lose their measured duration
          detail: 'cmd/sdlc/teststats_parse.go:86 extracts package timing only from ok summaries; line 97 recognizes FAIL without extracting its timing. TestBoundary376FailedPackageTiming reproduced "FAIL example.com/slow 90.0s" yielding no units and no warning, excluding slow failing runs from package p90. ARCH-PURPOSE: this is the 2nd finding in family contracted-log-format-coverage. Earlier work fixed an instance. State and enforce the rule that every supported format preserves available package durations for both passing and failing outcomes; enumerate text, JSON, and shard outcome fixtures and sweep that matrix in this round.'
          family: contracted-log-format-coverage
          round: 2
        - kind: defect
          id: BR-7
          severity: Important
          title: The planned 1 MiB log-line limit is not enforced
          detail: 'cmd/sdlc/teststats_parse.go:47-52 passes the entire log to failingTests and splits it without checking line lengths. The plan at workshop/plans/000376-test-stats-plan.md:37 requires a 1 MiB line limit and unavailable facts plus a warning for over-limit logs. TestBoundary376LineLimit supplied an oversized line and received a package measurement without warnings. ARCH-CONSTRAINTS: enforce the bound before either parser consumes the input, using a shared path for recorder and legacy reads, and add boundary-size regressions.'
          family: declared-parser-resource-bounds
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: c234e995793ba319832bb3cc02c01f06c50042fa
        tracker_head: 7163eb22588bb72115d4ce92c922e38dd184111b
        branch_point: 18915f6c17d9b4aeb5e618d3f736705ec355d2b3
        since: "2026-10-14T20:40:09-07:00"
        until: "2026-10-14T23:50:02-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-14T20:40:09-07:00"
              end: "2026-10-14T23:50:02-07:00"
              began: "2026-10-14T20:40:09-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: f4bbe03c4397cfcc3a21bfa4e595e7717bed38702b63011b91afcc790be0bf99
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 30520880
              windows:
                - start: "2026-10-14T20:40:09-07:00"
                  end: "2026-10-14T23:50:02-07:00"
            - harness: codex
              id: 01a13db1-fa52-7b83-bef0-9e0cbb5fed94
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: c0835be6d127033fc4026a726d32df65eaca885e9c30d55e73bbd3e4e0338f45
              started: "2026-10-15T03:53:57.198Z"
              spawned: false
              bytes: 540208
              windows:
                - start: "2026-10-15T03:53:57.198Z"
                  end: "2026-10-14T23:50:02-07:00"
            - harness: codex
              id: 01a13db4-79e3-73e3-b1b7-22d2f5b5b8cb
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 722aea63397d119b0d35afbd59ef2b50f11c19572fad617d61b01bbd36d5e3ec
              started: "2026-10-15T03:56:40.87Z"
              spawned: false
              bytes: 455430
              windows:
                - start: "2026-10-15T03:56:40.87Z"
                  end: "2026-10-14T23:50:02-07:00"
            - harness: codex
              id: 01a13db6-157f-7013-8b8e-ee975fe00bbd
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 959dbf7a68dc268e0a4b1e520466c83d9a89997094e2ffc6602b21ed0ed76ec4
              started: "2026-10-15T03:58:26.38Z"
              spawned: false
              bytes: 416762
              windows:
                - start: "2026-10-15T03:58:26.38Z"
                  end: "2026-10-14T23:50:02-07:00"
            - harness: codex
              id: 01a13db6-8751-73b1-ac45-2cfd823cd708
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: d7d23c22a882b3c12e70d04c8dd72742f868eba3238cda10a48a3f1744fa6e56
              started: "2026-10-15T03:58:55.257Z"
              spawned: false
              bytes: 406150
              windows:
                - start: "2026-10-15T03:58:55.257Z"
                  end: "2026-10-14T23:50:02-07:00"
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
                  end: "2026-10-14T23:50:02-07:00"
            - harness: codex
              id: 01a13e03-2033-7943-b391-58d87df3dfcd
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              digest: ff664449fd00dc265405978063a82080dc930c582d10211680080c425ebf1dcb
              started: "2026-10-15T05:22:34.734Z"
              spawned: true
              bytes: 626706
              windows:
                - start: "2026-10-15T05:22:34.734Z"
                  end: "2026-10-14T23:50:02-07:00"
        cuts:
            - dispatched: "2026-10-14T22:19:15-07:00"
              finalized: "2026-10-14T22:22:10-07:00"
              finalizing: false
            - dispatched: "2026-10-14T23:08:48-07:00"
              finalized: "2026-10-14T23:12:22-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-14T23:53:58-07:00"
      dispatched: "2026-10-14T23:50:41-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: Existing JSON package/test event and malformed-event regressions pass; explicit JSON parsing remains present.
          round: 3
        - id: BR-2
          disposition: addressed
          note: Suite-only legacy suppression remains present; TestSuiteStatsCommandSmokeRecordDoesNotHideLegacySuite passes.
          round: 3
        - id: BR-3
          disposition: addressed
          note: Plan line 15 places SuiteAttempt in cmd/sdlc/teststats.go and identifies persistence separately; the definition exists there and the round-1 Revisions passage records the correction.
          round: 3
        - id: BR-4
          disposition: addressed
          note: Owner-named publication temps and dead-owner pruning remain implemented; TestPruneSuiteAttemptsReclaimsOnlyDeadOwnerPublicationTemps passes.
          round: 3
        - id: BR-5
          disposition: addressed
          note: Canonical-name guards reject top-level and nested aliases. Removing those guards in a disposable copy makes the exit-alias and unit-alias regressions fail.
          round: 3
        - id: BR-6
          disposition: addressed
          note: Text, JSON, and shard pass/fail timing fixtures pass. Removing failed-package timing extraction makes TestSuiteLogFactsOutcomeFormats/plain_fail fail.
          round: 3
        - id: BR-7
          disposition: addressed
          note: The shared parser checks line lengths before either extractor. Exact-limit and over-limit tests pass; removing the bound makes the over-limit regression fail.
          round: 3
      findings:
        - kind: defect
          id: BR-8
          severity: Important
          title: Null nested durations become fabricated zero-second measurements
          detail: 'cmd/sdlc/teststats_store.go:194-197 accepts a unit''s "seconds":null; decoding converts it to zero, validation at line 162 accepts it, and cmd/sdlc/teststats.go:127-130 includes it in samples and p90. TestBoundary376NullUnitTiming reproduced this on the pinned head. ARCH-SECURE: this violates the plan''s unknown-value and malformed-record contracts. This is the 2nd finding in family persisted-evidence-canonical-decoding. State and enforce the schema-wide rule that required scalar fields at every depth reject null before zero-value decoding; sweep the field inventory and add regressions proving rejected records generate coverage diagnostics and never enter aggregates.'
          family: persisted-evidence-canonical-decoding
          round: 3
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

## Round 2 — 2026-10-14T23:12:22-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — JSON pass/fail/malformed fixtures pass on head. Removing JSON event handling makes TestSuiteLogFactsGoTestJSON and TestSuiteLogFactsMalformedGoTestJSONWarnsWithoutInventingData fail.
- BR-2 — addressed — TestSuiteStatsCommandSmokeRecordDoesNotHideLegacySuite passes on head and fails with the prior unconditional SHA deduplication restored.
- BR-3 — addressed — The before/after plan diff corrects SuiteAttempt to cmd/sdlc/teststats.go, matching its definition at line 11. The inventory explicitly identifies PURE entities and INTEGRATION persistence, and the close-review Revisions entry records the correction.
- BR-4 — addressed — Publication temps now carry owner IDs. TestPruneSuiteAttemptsReclaimsOnlyDeadOwnerPublicationTemps protects active writers and reclaims dead-owner residue; restoring the prior store implementation makes it fail.

### Raised

- **BR-5** [Critical] `persisted-evidence-canonical-decoding` (kind: defect) Conflicting JSON field casing can turn a failed attempt into success
  cmd/sdlc/teststats_store.go:75 checks duplicate keys case-sensitively, but the struct decode at line 102 accepts case-insensitive field matches. A valid failed record containing "exit":2 followed by "Exit":0 is accepted with Exit=0, despite DisallowUnknownFields. TestBoundary376CaseAlias reproduced this on head. ARCH-SECURE: this admitted record can become false successful-run and flake-confirmation evidence. Reject noncanonical field names before decoding, and test aliases and conflicting keys across outcome/provenance fields.
- **BR-6** [Important] `contracted-log-format-coverage` (kind: defect) Plain-text failed-package summaries lose their measured duration
  cmd/sdlc/teststats_parse.go:86 extracts package timing only from ok summaries; line 97 recognizes FAIL without extracting its timing. TestBoundary376FailedPackageTiming reproduced "FAIL example.com/slow 90.0s" yielding no units and no warning, excluding slow failing runs from package p90. ARCH-PURPOSE: this is the 2nd finding in family contracted-log-format-coverage. Earlier work fixed an instance. State and enforce the rule that every supported format preserves available package durations for both passing and failing outcomes; enumerate text, JSON, and shard outcome fixtures and sweep that matrix in this round.
- **BR-7** [Important] `declared-parser-resource-bounds` (kind: defect) The planned 1 MiB log-line limit is not enforced
  cmd/sdlc/teststats_parse.go:47-52 passes the entire log to failingTests and splits it without checking line lengths. The plan at workshop/plans/000376-test-stats-plan.md:37 requires a 1 MiB line limit and unavailable facts plus a warning for over-limit logs. TestBoundary376LineLimit supplied an oversized line and received a package measurement without warnings. ARCH-CONSTRAINTS: enforce the bound before either parser consumes the input, using a shared path for recorder and legacy reads, and add boundary-size regressions.

## Round 3 — 2026-10-14T23:53:58-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — Existing JSON package/test event and malformed-event regressions pass; explicit JSON parsing remains present.
- BR-2 — addressed — Suite-only legacy suppression remains present; TestSuiteStatsCommandSmokeRecordDoesNotHideLegacySuite passes.
- BR-3 — addressed — Plan line 15 places SuiteAttempt in cmd/sdlc/teststats.go and identifies persistence separately; the definition exists there and the round-1 Revisions passage records the correction.
- BR-4 — addressed — Owner-named publication temps and dead-owner pruning remain implemented; TestPruneSuiteAttemptsReclaimsOnlyDeadOwnerPublicationTemps passes.
- BR-5 — addressed — Canonical-name guards reject top-level and nested aliases. Removing those guards in a disposable copy makes the exit-alias and unit-alias regressions fail.
- BR-6 — addressed — Text, JSON, and shard pass/fail timing fixtures pass. Removing failed-package timing extraction makes TestSuiteLogFactsOutcomeFormats/plain_fail fail.
- BR-7 — addressed — The shared parser checks line lengths before either extractor. Exact-limit and over-limit tests pass; removing the bound makes the over-limit regression fail.

### Raised

- **BR-8** [Important] `persisted-evidence-canonical-decoding` (kind: defect) Null nested durations become fabricated zero-second measurements
  cmd/sdlc/teststats_store.go:194-197 accepts a unit's "seconds":null; decoding converts it to zero, validation at line 162 accepts it, and cmd/sdlc/teststats.go:127-130 includes it in samples and p90. TestBoundary376NullUnitTiming reproduced this on the pinned head. ARCH-SECURE: this violates the plan's unknown-value and malformed-record contracts. This is the 2nd finding in family persisted-evidence-canonical-decoding. State and enforce the schema-wide rule that required scalar fields at every depth reject null before zero-value decoding; sweep the field inventory and add regressions proving rejected records generate coverage diagnostics and never enter aggregates.

## Open findings

- **BR-8** [Important] `persisted-evidence-canonical-decoding` (kind: defect) Null nested durations become fabricated zero-second measurements
