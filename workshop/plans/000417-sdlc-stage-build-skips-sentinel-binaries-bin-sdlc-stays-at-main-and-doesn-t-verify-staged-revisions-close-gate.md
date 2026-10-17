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

## Open findings

- **BR-1** [Critical] `fail-closed-build-verification` (kind: defect) Missing or corrupt build evidence can pass verification
- **BR-2** [Important] `target-declaration-validation` (kind: defect) Sentinel declarations accept Make options and assignments
- **BR-3** [Important] `external-dependency-test-seam` (kind: defect) Owner-target execution lacks an injected stateful fake and conformance check
- **BR-4** [Important] `user-facing-docs-gate` (kind: defect) README update is missing for changed stage behavior and sentinel syntax
