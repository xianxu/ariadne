---
gate: boundary-review
issue: 417
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 871fe3d9-74ce-4300-80dc-9877c0106cba
            repository: github.com/xianxu/ariadne
            issue_id: "000417"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/aa0ceefdb84dd3c7
            reviewed_head: 8ea8f99e16ae3d9f3ae4117f715de9909e1da5fb
            slot: ariadne:1
            requested_at: "2026-10-17T03:52:36.775556Z"
          phase: finished
          pid: 68765
          started_at: "2026-10-17T03:52:37.023177Z"
          finished_at: "2026-10-17T03:54:17.00681Z"
          session_id: 01a147fd-7de4-7b51-8b0e-4aba4e831333
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 8ea8f99e16ae3d9f3ae4117f715de9909e1da5fb
        tracker_head: 07925b6a5d6bbc8bf3fb7321a58cb0a888d3ef7a
        branch_point: f5927dfedbfb2e6dbc7d8c91ac56085732181b7b
        since: "2026-10-16T19:43:25-07:00"
        until: "2026-10-16T20:52:03-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T19:43:25-07:00"
              end: "2026-10-16T20:52:03-07:00"
              began: "2026-10-16T19:43:25-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: ded09f30d0645983f9d38e081150e3a44bbf369116199caf992e0a670a3d3c89
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 41532929
              windows:
                - start: "2026-10-16T19:43:25-07:00"
                  end: "2026-10-16T20:52:03-07:00"
      "n": 1
      timestamp: "2026-10-16T20:54:18-07:00"
      dispatched: "2026-10-16T20:52:36-07:00"
      agent: codex
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Missing or corrupt build evidence can pass verification
          detail: cmd/sdlc/stagebuild.go:64-75 skips errors classified by notGoBinary, but Go's buildinfo reader also returns "not a Go executable" for corrupt build-info headers. A corrupt non-sentinel Go binary can therefore disappear from verification and stage succeeds. Also, buildRevision defaults missing or invalid vcs.modified values to false, accepting an unproven clean stamp despite the plan requiring vcs.modified == false. Distinguish recognizable Go candidates with unreadable metadata from genuine non-Go files, require an explicit valid clean stamp, and add regressions for both cases.
          family: fail-closed-build-verification
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Sentinel declarations accept Make options and assignments
          detail: cmd/sdlc/stagebuild.go:125-129 accepts --version, -n, and NAME=value as single tokens, then runStageMake passes them directly to make. For example, --version exits successfully without running an owner target; existing matching output can let verification pass afterward. Reject option and assignment syntax, define the supported target grammar, and test malformed and duplicate declarations (ARCH-SECURE).
          family: target-declaration-validation
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: Owner-target execution lacks an injected stateful fake and conformance check
          detail: cmd/sdlc/stagebuild.go:91-97 invokes real make directly, and stagebuild_test.go exercises real Makefiles while replacing only readBuildInfo. No injectable Make runner, stateful fake, or live conformance check covers the expanded generic-build → owner-build → verification interaction. Add the shared production/test boundary and test execution order, failure propagation, and retries against its persisted output state (ARCH-MOCK).
          family: external-dependency-test-seam
          round: 1
        - kind: defect
          id: BR-4
          severity: Important
          title: README update is missing for changed stage behavior and sentinel syntax
          detail: 'The range changes stage --build to refuse unverifiable revisions, makes stage --clear rebuild and verify binaries, and introduces # stage-build: <target>. Atlas and help text explain these changes, but README.md is unchanged. Add concise usage and recovery guidance, linking to the updated atlas documentation.'
          family: user-facing-docs-gate
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 1b59809e-d3e2-4e7f-8c01-97b9561e5bf3
            repository: github.com/xianxu/ariadne
            issue_id: "000417"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/53d1649bb259f34b
            reviewed_head: d46d47bfcf07ad9faffc142091bf1e01582b37b7
            slot: ariadne:1
            requested_at: "2026-10-17T03:58:47.184689Z"
          phase: finished
          pid: 53705
          started_at: "2026-10-17T03:58:47.407823Z"
          finished_at: "2026-10-17T04:01:00.463466Z"
          session_id: 01a14803-2473-7312-aef6-9e5783335943
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: d46d47bfcf07ad9faffc142091bf1e01582b37b7
        tracker_head: 07925b6a5d6bbc8bf3fb7321a58cb0a888d3ef7a
        branch_point: f5927dfedbfb2e6dbc7d8c91ac56085732181b7b
        since: "2026-10-16T19:43:25-07:00"
        until: "2026-10-16T20:57:26-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T19:43:25-07:00"
              end: "2026-10-16T20:57:26-07:00"
              began: "2026-10-16T19:43:25-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 032ca5bf75139f7fad798aca3fc62b5d1f59bcf11ec613ef1ac58a9767f24a2c
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 42115430
              windows:
                - start: "2026-10-16T19:43:25-07:00"
                  end: "2026-10-16T20:57:26-07:00"
        cuts:
            - dispatched: "2026-10-16T20:52:36-07:00"
              finalized: "2026-10-16T20:54:18-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-16T21:01:01-07:00"
      dispatched: "2026-10-16T20:58:47-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: not-addressed
          note: 'Clean-stamp validation is fixed, and removing either new guard makes its regression fail. However, cmd/sdlc/stagebuild.go:114-116 omits XCOFF and Plan 9 formats recognized by debug/buildinfo. Independently cross-built an AIX/ppc64 Go executable, corrupted its build-info magic, and exercised production verification through a temporary test overlay: readBuildInfo returned "not a Go executable", nativeExecutable returned false, and verification succeeded. Enumerate every supported executable format and its variants, fail closed across that enumeration, and add regressions (ARCH-PURPOSE, ARCH-SECURE).'
          round: 2
        - id: BR-2
          disposition: addressed
          note: stagebuild.go:154 and :162 reject options and assignments through an explicit target grammar. The malformed and duplicate declaration tests pass; removing grammar validation makes the option, dry-run and assignment regressions fail.
          round: 2
        - id: BR-3
          disposition: addressed
          note: Production and tests share stageBuildWithMake. The fake persists stale output through owner failure and verifies retry order; the real-Make fixture exercises owner execution and renamed outputs. Removing the owner-target call makes the failure/retry regression fail.
          round: 2
        - id: BR-4
          disposition: addressed
          note: README.md:17-24 now documents sentinel targets, verification refusal, clear rebuilding, and retry recovery, linking to the atlas. These statements match stage.go:135-176 and stagebuild.go:30-90.
          round: 2
      findings:
        - kind: defect
          id: BR-5
          severity: Critical
          title: Sentinel parser contradicts the plan's PURE classification
          detail: 'The plan''s concepts table at workshop/plans/000417-sdlc-stage-build-skips-sentinel-binaries-bin-sdlc-stays-at-main-and-doesn-t-verify-staged-revisions-plan.md:13 labels the sentinel parser PURE, but sentinelBuildTarget at cmd/sdlc/stagebuild.go:142-145 reads the filesystem, and its tests create temporary files. Extract parsing into a deterministic text-to-target function with direct tests and retain file reading in the INTEGRATION shell, or classify the existing entity INTEGRATION. Add a ## Revisions entry correcting the contract (ARCH-PURE).'
          family: pure-core-io-separation
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: cc596950-4998-4c2e-a750-66520c7caf77
            repository: github.com/xianxu/ariadne
            issue_id: "000417"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/a45d4bf521a8f985
            reviewed_head: a7f66f308e6a2e5348d6d55929844f39fe935043
            slot: ariadne:1
            requested_at: "2026-10-17T04:03:29.865435Z"
          phase: finished
          pid: 36910
          started_at: "2026-10-17T04:03:29.966868Z"
          finished_at: "2026-10-17T04:05:22.660275Z"
          session_id: 01a14807-7312-7ab2-8500-9c63e74c981e
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: a7f66f308e6a2e5348d6d55929844f39fe935043
        tracker_head: 07925b6a5d6bbc8bf3fb7321a58cb0a888d3ef7a
        branch_point: f5927dfedbfb2e6dbc7d8c91ac56085732181b7b
        since: "2026-10-16T19:43:25-07:00"
        until: "2026-10-16T21:02:58-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T19:43:25-07:00"
              end: "2026-10-16T21:02:58-07:00"
              began: "2026-10-16T19:43:25-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: b2c95369467d4379c248c9e4ac9728cbbe3e401384e75e3e8c7cd49cafbab0f2
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 42578740
              windows:
                - start: "2026-10-16T19:43:25-07:00"
                  end: "2026-10-16T21:02:58-07:00"
        cuts:
            - dispatched: "2026-10-16T20:52:36-07:00"
              finalized: "2026-10-16T20:54:18-07:00"
              finalizing: false
            - dispatched: "2026-10-16T20:58:47-07:00"
              finalized: "2026-10-16T21:01:01-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-16T21:05:23-07:00"
      dispatched: "2026-10-16T21:03:29-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: stagebuild.go:76 rejects unreadable native executables, and :85 requires an explicit vcs.modified=false. Targeted tests pass. In a disposable pinned-head tree, restoring the previous format guard made all four new cross-format cases fail; removing the explicit-clean requirement made all three missing/invalid-stamp cases fail. The disposable tree was removed.
          round: 3
        - id: BR-5
          disposition: addressed
          note: The pinned plan changes the sentinel reader/parser classification from PURE to INTEGRATION at line 13 and records the correction under Revisions at line 55. This accurately describes sentinelBuildTarget's filesystem read at stagebuild.go:156; no executable behavior changed.
          round: 3
        - id: BR-2
          disposition: addressed
          note: The target grammar at stagebuild.go:175 rejects options and assignments; the existing regression cases passed.
          round: 3
        - id: BR-3
          disposition: addressed
          note: Production and tests share stageBuildWithMake. The stateful owner-failure/retry fixture and real-Make sequencing fixtures passed.
          round: 3
        - id: BR-4
          disposition: addressed
          note: README.md:17-24 documents owner targets, verification refusals, clear rebuilding, and retry. These statements match the implementation; atlas/workflow/workspace-branching.md:100-109 describes the same flow.
          round: 3
      recipe: milestone-review
      isolation: fenced
      reviewed: a7f66f308e6a2e5348d6d55929844f39fe935043
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: ad42ca81-b758-4cfc-84e1-796e7cc4561a
            repository: github.com/xianxu/ariadne
            issue_id: "000417"
            recipe: milestone-review
            round: 4
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/e40fcdbb105325c6
            reviewed_head: 448595c97a127ed6d1bcd5d6e9b6906ea3018f55
            slot: ariadne:1
            requested_at: "2026-10-17T05:34:55.16199Z"
          phase: finished
          pid: 42411
          started_at: "2026-10-17T05:34:55.259661Z"
          finished_at: "2026-10-17T05:36:38.315907Z"
          session_id: 01a1485b-2625-7201-a178-da38d531d10a
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 448595c97a127ed6d1bcd5d6e9b6906ea3018f55
        tracker_head: 63ac306f943076f2d747e11e8eeaca6fe1b94c64
        branch_point: f5927dfedbfb2e6dbc7d8c91ac56085732181b7b
        since: "2026-10-16T19:43:25-07:00"
        until: "2026-10-16T22:34:04-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T19:43:25-07:00"
              end: "2026-10-16T22:34:04-07:00"
              began: "2026-10-16T19:43:25-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: d7276a5c00140900572dab98d9a65a12c607cca562c24ed2bfd93b4f3bd56d0f
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 43892169
              windows:
                - start: "2026-10-16T19:43:25-07:00"
                  end: "2026-10-16T22:34:04-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.036066666666666664
        cuts:
            - dispatched: "2026-10-16T20:52:36-07:00"
              finalized: "2026-10-16T20:54:18-07:00"
              finalizing: false
            - dispatched: "2026-10-16T20:58:47-07:00"
              finalized: "2026-10-16T21:01:01-07:00"
              finalizing: false
            - dispatched: "2026-10-16T21:03:29-07:00"
              finalized: "2026-10-16T21:05:23-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-16T22:36:46-07:00"
      dispatched: "2026-10-16T22:34:55-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: Verification requires an explicit clean stamp and rejects unreadable native executables. Removing these guards made the clean-stamp, ELF, XCOFF, Plan 9 and Mach-O regressions fail.
          round: 4
        - id: BR-2
          disposition: addressed
          note: sentinelBuildTarget restricts declarations to target names. Removing grammar validation made the option, dry-run and assignment regression cases fail.
          round: 4
        - id: BR-3
          disposition: addressed
          note: Production and tests share stageBuildWithMake. Stateful owner-failure/retry and real-Make fixtures passed; skipping owner execution made the retry regression fail.
          round: 4
        - id: BR-4
          disposition: addressed
          note: README.md:17-25 documents sentinel targets, verification refusal, clear rebuilding and retry. These statements match stage.go:130-182 and stagebuild.go:32-105; atlas documents the same flow.
          round: 4
        - id: BR-5
          disposition: addressed
          note: The plan's Concepts table classifies the sentinel reader/parser as INTEGRATION and records the correction under Revisions. This matches sentinelBuildTarget's filesystem read at stagebuild.go:172; the correction is prose-only.
          round: 4
      recipe: milestone-review
      isolation: fenced
      reviewed: 448595c97a127ed6d1bcd5d6e9b6906ea3018f55
      blocked: false
---

# Gate ledger — ariadne#417 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-16T20:54:18-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `fail-closed-build-verification` (kind: defect) Missing or corrupt build evidence can pass verification
  cmd/sdlc/stagebuild.go:64-75 skips errors classified by notGoBinary, but Go's buildinfo reader also returns "not a Go executable" for corrupt build-info headers. A corrupt non-sentinel Go binary can therefore disappear from verification and stage succeeds. Also, buildRevision defaults missing or invalid vcs.modified values to false, accepting an unproven clean stamp despite the plan requiring vcs.modified == false. Distinguish recognizable Go candidates with unreadable metadata from genuine non-Go files, require an explicit valid clean stamp, and add regressions for both cases.
- **BR-2** [Important] `target-declaration-validation` (kind: defect) Sentinel declarations accept Make options and assignments
  cmd/sdlc/stagebuild.go:125-129 accepts --version, -n, and NAME=value as single tokens, then runStageMake passes them directly to make. For example, --version exits successfully without running an owner target; existing matching output can let verification pass afterward. Reject option and assignment syntax, define the supported target grammar, and test malformed and duplicate declarations (ARCH-SECURE).
- **BR-3** [Important] `external-dependency-test-seam` (kind: defect) Owner-target execution lacks an injected stateful fake and conformance check
  cmd/sdlc/stagebuild.go:91-97 invokes real make directly, and stagebuild_test.go exercises real Makefiles while replacing only readBuildInfo. No injectable Make runner, stateful fake, or live conformance check covers the expanded generic-build → owner-build → verification interaction. Add the shared production/test boundary and test execution order, failure propagation, and retries against its persisted output state (ARCH-MOCK).
- **BR-4** [Important] `user-facing-docs-gate` (kind: defect) README update is missing for changed stage behavior and sentinel syntax
  The range changes stage --build to refuse unverifiable revisions, makes stage --clear rebuild and verify binaries, and introduces # stage-build: <target>. Atlas and help text explain these changes, but README.md is unchanged. Add concise usage and recovery guidance, linking to the updated atlas documentation.

## Round 2 — 2026-10-16T21:01:01-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — not-addressed — Clean-stamp validation is fixed, and removing either new guard makes its regression fail. However, cmd/sdlc/stagebuild.go:114-116 omits XCOFF and Plan 9 formats recognized by debug/buildinfo. Independently cross-built an AIX/ppc64 Go executable, corrupted its build-info magic, and exercised production verification through a temporary test overlay: readBuildInfo returned "not a Go executable", nativeExecutable returned false, and verification succeeded. Enumerate every supported executable format and its variants, fail closed across that enumeration, and add regressions (ARCH-PURPOSE, ARCH-SECURE).
- BR-2 — addressed — stagebuild.go:154 and :162 reject options and assignments through an explicit target grammar. The malformed and duplicate declaration tests pass; removing grammar validation makes the option, dry-run and assignment regressions fail.
- BR-3 — addressed — Production and tests share stageBuildWithMake. The fake persists stale output through owner failure and verifies retry order; the real-Make fixture exercises owner execution and renamed outputs. Removing the owner-target call makes the failure/retry regression fail.
- BR-4 — addressed — README.md:17-24 now documents sentinel targets, verification refusal, clear rebuilding, and retry recovery, linking to the atlas. These statements match stage.go:135-176 and stagebuild.go:30-90.

### Raised

- **BR-5** [Critical] `pure-core-io-separation` (kind: defect) Sentinel parser contradicts the plan's PURE classification
  The plan's concepts table at workshop/plans/000417-sdlc-stage-build-skips-sentinel-binaries-bin-sdlc-stays-at-main-and-doesn-t-verify-staged-revisions-plan.md:13 labels the sentinel parser PURE, but sentinelBuildTarget at cmd/sdlc/stagebuild.go:142-145 reads the filesystem, and its tests create temporary files. Extract parsing into a deterministic text-to-target function with direct tests and retain file reading in the INTEGRATION shell, or classify the existing entity INTEGRATION. Add a ## Revisions entry correcting the contract (ARCH-PURE).

## Round 3 — 2026-10-16T21:05:23-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — stagebuild.go:76 rejects unreadable native executables, and :85 requires an explicit vcs.modified=false. Targeted tests pass. In a disposable pinned-head tree, restoring the previous format guard made all four new cross-format cases fail; removing the explicit-clean requirement made all three missing/invalid-stamp cases fail. The disposable tree was removed.
- BR-5 — addressed — The pinned plan changes the sentinel reader/parser classification from PURE to INTEGRATION at line 13 and records the correction under Revisions at line 55. This accurately describes sentinelBuildTarget's filesystem read at stagebuild.go:156; no executable behavior changed.
- BR-2 — addressed — The target grammar at stagebuild.go:175 rejects options and assignments; the existing regression cases passed.
- BR-3 — addressed — Production and tests share stageBuildWithMake. The stateful owner-failure/retry fixture and real-Make sequencing fixtures passed.
- BR-4 — addressed — README.md:17-24 documents owner targets, verification refusals, clear rebuilding, and retry. These statements match the implementation; atlas/workflow/workspace-branching.md:100-109 describes the same flow.

## Round 4 — 2026-10-16T22:36:46-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — Verification requires an explicit clean stamp and rejects unreadable native executables. Removing these guards made the clean-stamp, ELF, XCOFF, Plan 9 and Mach-O regressions fail.
- BR-2 — addressed — sentinelBuildTarget restricts declarations to target names. Removing grammar validation made the option, dry-run and assignment regression cases fail.
- BR-3 — addressed — Production and tests share stageBuildWithMake. Stateful owner-failure/retry and real-Make fixtures passed; skipping owner execution made the retry regression fail.
- BR-4 — addressed — README.md:17-25 documents sentinel targets, verification refusal, clear rebuilding and retry. These statements match stage.go:130-182 and stagebuild.go:32-105; atlas documents the same flow.
- BR-5 — addressed — The plan's Concepts table classifies the sentinel reader/parser as INTEGRATION and records the correction under Revisions. This matches sentinelBuildTarget's filesystem read at stagebuild.go:172; the correction is prose-only.

## Open findings

(none — every finding has been disposed)
