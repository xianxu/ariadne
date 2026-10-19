---
gate: boundary-review
issue: 465
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: af0ef016-bea9-4bcf-a294-6a71259082fb
            repository: github.com/xianxu/ariadne
            issue_id: "000465"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/65c64d79c4c83df2
            reviewed_head: 4e8fedec83276f8fb2c014b07d2b25ba9dad3834
            slot: ariadne:3
            requested_at: "2026-10-19T15:22:48.307457Z"
          phase: finished
          pid: 97669
          started_at: "2026-10-19T15:22:48.494841Z"
          finished_at: "2026-10-19T15:24:59.038878Z"
          session_id: 01a154c2-18ff-7342-921b-226179de5ba1
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 4e8fedec83276f8fb2c014b07d2b25ba9dad3834
        tracker_head: 3adcfba4745aa1327f4d85e699615d15b16d7019
        branch_point: 45a96385ad4d24bda920822c60d7879d643a2a99
        since: "2026-10-19T07:56:05-07:00"
        until: "2026-10-19T08:22:13-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T07:56:05-07:00"
              end: "2026-10-19T08:22:13-07:00"
              began: "2026-10-19T07:56:05-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 3b9a73a0c7246691be1bf0b80dfdfc4cd054b3ce71848d30afccc3f053f946bf
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 38331993
              windows:
                - start: "2026-10-19T07:56:05-07:00"
                  end: "2026-10-19T08:22:13-07:00"
            - harness: codex
              id: 01a154a9-c7b3-73a1-9284-3ee124ec3272
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 0bfec05f99b356a7a34127c3a034e9089ba5afe4d74daa537db66e54ca93036c
              started: "2026-10-19T14:56:15.338Z"
              spawned: true
              bytes: 1003925
              windows:
                - start: "2026-10-19T14:56:15.338Z"
                  end: "2026-10-19T08:22:13-07:00"
      "n": 1
      timestamp: "2026-10-19T08:25:00-07:00"
      dispatched: "2026-10-19T08:22:48-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Future-dated edits no longer test build-start timestamp preservation
          detail: 'construct/scripts/test/dev-aliases.test.sh:443-444 sets the mid-build edit to 2030, so it remains newer than the completed binary even if timestamp restoration breaks. An in-memory mutation removing touch -r from both generated skipd definitions still passed 82/82 assertions, including lines 491 and 493. These are all instances of this weakened timestamp fixture in the window. Use ordered start/edit/output timestamps and verify the regression fails without restoration. ARCH-PURPOSE: retaining assertion counts must preserve their behavioral coverage.'
          family: preserve-regression-sensitivity
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Done when does not reflect the revised acceptance contract
          detail: workshop/issues/000465-dev-aliases-script-exceeds-its-per-script-budget-under-suite-load.md:35-36 retains the original wall-time criteria, while Revisions at lines 53-55 adds production path conformance and N/2N zero-path-utility requirements and accepts counted reductions plus a loaded recorded-suite pass. This issue's Done when section is the sole instance in the window. Update it to name those required checks and reconcile the performance evidence, preserving the recorded-suite requirement.
          family: acceptance-contract-synchronization
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: b70b1f4e-f683-4548-a58a-a7a1da42fe4e
            repository: github.com/xianxu/ariadne
            issue_id: "000465"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/43d958124e9bd312
            reviewed_head: 7b64fba896b2b67f054c10e6ab6efd9b31d84566
            slot: ariadne:3
            requested_at: "2026-10-19T15:28:38.159182Z"
          phase: finished
          pid: 67808
          started_at: "2026-10-19T15:28:38.340237Z"
          finished_at: "2026-10-19T15:31:17.622585Z"
          session_id: 01a154c7-6f68-7d42-af66-f7acbd01040d
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 7b64fba896b2b67f054c10e6ab6efd9b31d84566
        tracker_head: 3adcfba4745aa1327f4d85e699615d15b16d7019
        branch_point: 45a96385ad4d24bda920822c60d7879d643a2a99
        since: "2026-10-19T07:56:05-07:00"
        until: "2026-10-19T08:27:53-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T07:56:05-07:00"
              end: "2026-10-19T08:27:53-07:00"
              began: "2026-10-19T07:56:05-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: d99b9a87559ec9b72f49191b7b6152699992e9bb506aa8938f05774e91662f09
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 38610170
              windows:
                - start: "2026-10-19T07:56:05-07:00"
                  end: "2026-10-19T08:27:53-07:00"
            - harness: codex
              id: 01a154a9-c7b3-73a1-9284-3ee124ec3272
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 0bfec05f99b356a7a34127c3a034e9089ba5afe4d74daa537db66e54ca93036c
              started: "2026-10-19T14:56:15.338Z"
              spawned: true
              bytes: 1003925
              windows:
                - start: "2026-10-19T14:56:15.338Z"
                  end: "2026-10-19T08:27:53-07:00"
        cuts:
            - dispatched: "2026-10-19T08:22:48-07:00"
              finalized: "2026-10-19T08:25:00-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-19T08:31:18-07:00"
      dispatched: "2026-10-19T08:28:38-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: construct/scripts/test/dev-aliases.test.sh:446-454 establishes start=2000, edit=2001, output=2002. Normal execution passes 82/82; replacing timestamp restoration with a no-op makes the mid-build-edit assertion at line 499 fail (79 passed, three failures including two shape checks).
          round: 2
        - id: BR-2
          disposition: addressed
          note: The pinned before/after diff updates the issue's Done when at lines 35-37 to require path conformance, N/2N zero-launch checks, counted reductions, repeated targeted checks, and a loaded recorded-suite pass, matching Revisions at lines 54-56 while retaining the two-minute ceiling.
          round: 2
      findings:
        - kind: defect
          id: BR-3
          severity: Important
          title: Strict-list comparison no longer observes ordinary strict output
          detail: |-
            construct/scripts/test/dev-aliases.test.sh:35-44 skips ordinary --strict --list execution, substitutes a supplied status, and compares verbose strict output against a non-strict baseline. All instances are the same-origin and distinct-origin strict calls at lines 280 and 301. A production mutation suppressing ordinary strict-list output only on collisions passes 82/82 at head; the identical mutation fails the original test (81 passed, one failure).
            This is the 2nd finding in family preserve-regression-sensitivity. Earlier rounds fixed instances. Do not fix only the demonstrated collision case: enforce the rule that fixture reuse must preserve observed output and status for every originally tested mode. Capture or reuse actual strict results for both enumerated cases and verify the mutation fails. ARCH-PURPOSE requires preserving behavioral coverage, not merely assertion counts.
          family: preserve-regression-sensitivity
          round: 2
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: e3684d9e-2063-4047-8a2f-561e7bbd96ee
            repository: github.com/xianxu/ariadne
            issue_id: "000465"
            recipe: small-diff-review
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/94e7533771bd0b4a
            reviewed_head: f874054852a15da9f76d9d2fc23482a05f99a148
            slot: ariadne:3
            requested_at: "2026-10-19T15:34:54.975909Z"
          phase: finished
          pid: 60090
          started_at: "2026-10-19T15:34:55.087171Z"
          finished_at: "2026-10-19T15:37:40.0536Z"
          session_id: 01a154cd-2e71-7271-a4e8-794f335b2981
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: f874054852a15da9f76d9d2fc23482a05f99a148
        tracker_head: 3adcfba4745aa1327f4d85e699615d15b16d7019
        branch_point: 45a96385ad4d24bda920822c60d7879d643a2a99
        since: "2026-10-19T07:56:05-07:00"
        until: "2026-10-19T08:34:23-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T07:56:05-07:00"
              end: "2026-10-19T08:34:23-07:00"
              began: "2026-10-19T07:56:05-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: db0d621c78a47bf64c65019b2a91da4b9fbf38b0ed5c4194787724b721547c7d
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 38891695
              windows:
                - start: "2026-10-19T07:56:05-07:00"
                  end: "2026-10-19T08:34:23-07:00"
            - harness: codex
              id: 01a154a9-c7b3-73a1-9284-3ee124ec3272
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: b53664e4ccaf180885a584990b32745991c54a24832ab3c09d2960ef9ae940fe
              started: "2026-10-19T14:56:15.338Z"
              spawned: true
              bytes: 1034723
              windows:
                - start: "2026-10-19T14:56:15.338Z"
                  end: "2026-10-19T08:34:23-07:00"
        cuts:
            - dispatched: "2026-10-19T08:22:48-07:00"
              finalized: "2026-10-19T08:25:00-07:00"
              finalizing: false
            - dispatched: "2026-10-19T08:28:38-07:00"
              finalized: "2026-10-19T08:31:18-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-19T08:37:51-07:00"
      dispatched: "2026-10-19T08:34:54-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: construct/scripts/test/dev-aliases.test.sh:440-448 establishes ordered start/edit/output timestamps. Disabling timestamp restoration makes the intended mid-build-edit assertion at line 493 fail; the mutated run reports 79 passed, three failed.
          round: 3
        - id: BR-2
          disposition: addressed
          note: The pinned before/after diff updates the issue's Done when at lines 35-37 to include path conformance, N/2N zero-launch checks, counted reductions and loaded-suite evidence, matching its Revisions while retaining the two-minute ceiling.
          round: 3
        - id: BR-3
          disposition: addressed
          note: construct/scripts/test/dev-aliases.test.sh:35-44 captures actual output and status for all four ordinary/verbose comparisons, including strict cases at lines 274 and 295. Suppressing ordinary strict-list output on collisions makes both repaired comparisons fail (79 passed, three failed). Fresh gen_def and nous_def generation also restores observations after fixture changes; retained dupe.defs reuse observes the same mode on unchanged inputs.
          round: 3
      recipe: small-diff-review
      isolation: fenced
      reviewed: f874054852a15da9f76d9d2fc23482a05f99a148
      blocked: false
---

# Gate ledger — ariadne#465 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T08:25:00-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `preserve-regression-sensitivity` (kind: defect) Future-dated edits no longer test build-start timestamp preservation
  construct/scripts/test/dev-aliases.test.sh:443-444 sets the mid-build edit to 2030, so it remains newer than the completed binary even if timestamp restoration breaks. An in-memory mutation removing touch -r from both generated skipd definitions still passed 82/82 assertions, including lines 491 and 493. These are all instances of this weakened timestamp fixture in the window. Use ordered start/edit/output timestamps and verify the regression fails without restoration. ARCH-PURPOSE: retaining assertion counts must preserve their behavioral coverage.
- **BR-2** [Important] `acceptance-contract-synchronization` (kind: defect) Done when does not reflect the revised acceptance contract
  workshop/issues/000465-dev-aliases-script-exceeds-its-per-script-budget-under-suite-load.md:35-36 retains the original wall-time criteria, while Revisions at lines 53-55 adds production path conformance and N/2N zero-path-utility requirements and accepts counted reductions plus a loaded recorded-suite pass. This issue's Done when section is the sole instance in the window. Update it to name those required checks and reconcile the performance evidence, preserving the recorded-suite requirement.

## Round 2 — 2026-10-19T08:31:18-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — construct/scripts/test/dev-aliases.test.sh:446-454 establishes start=2000, edit=2001, output=2002. Normal execution passes 82/82; replacing timestamp restoration with a no-op makes the mid-build-edit assertion at line 499 fail (79 passed, three failures including two shape checks).
- BR-2 — addressed — The pinned before/after diff updates the issue's Done when at lines 35-37 to require path conformance, N/2N zero-launch checks, counted reductions, repeated targeted checks, and a loaded recorded-suite pass, matching Revisions at lines 54-56 while retaining the two-minute ceiling.

### Raised

- **BR-3** [Important] `preserve-regression-sensitivity` (kind: defect) Strict-list comparison no longer observes ordinary strict output
  construct/scripts/test/dev-aliases.test.sh:35-44 skips ordinary --strict --list execution, substitutes a supplied status, and compares verbose strict output against a non-strict baseline. All instances are the same-origin and distinct-origin strict calls at lines 280 and 301. A production mutation suppressing ordinary strict-list output only on collisions passes 82/82 at head; the identical mutation fails the original test (81 passed, one failure).
  This is the 2nd finding in family preserve-regression-sensitivity. Earlier rounds fixed instances. Do not fix only the demonstrated collision case: enforce the rule that fixture reuse must preserve observed output and status for every originally tested mode. Capture or reuse actual strict results for both enumerated cases and verify the mutation fails. ARCH-PURPOSE requires preserving behavioral coverage, not merely assertion counts.

## Round 3 — 2026-10-19T08:37:51-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — construct/scripts/test/dev-aliases.test.sh:440-448 establishes ordered start/edit/output timestamps. Disabling timestamp restoration makes the intended mid-build-edit assertion at line 493 fail; the mutated run reports 79 passed, three failed.
- BR-2 — addressed — The pinned before/after diff updates the issue's Done when at lines 35-37 to include path conformance, N/2N zero-launch checks, counted reductions and loaded-suite evidence, matching its Revisions while retaining the two-minute ceiling.
- BR-3 — addressed — construct/scripts/test/dev-aliases.test.sh:35-44 captures actual output and status for all four ordinary/verbose comparisons, including strict cases at lines 274 and 295. Suppressing ordinary strict-list output on collisions makes both repaired comparisons fail (79 passed, three failed). Fresh gen_def and nous_def generation also restores observations after fixture changes; retained dupe.defs reuse observes the same mode on unchanged inputs.

## Open findings

(none — every finding has been disposed)
