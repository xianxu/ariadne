---
gate: boundary-review
issue: 440
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: abba2962-8c81-4e86-9de0-13a41af27890
            repository: github.com/xianxu/ariadne
            issue_id: "000440"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/a2c8b5e09ac57163
            reviewed_head: 3b3545a9cf592714228e6e89cf16395a756cbd04
            slot: ariadne:1
            requested_at: "2026-10-18T13:21:43.829254Z"
          phase: finished
          pid: 71529
          started_at: "2026-10-18T13:21:43.927145Z"
          finished_at: "2026-10-18T13:24:37.910164Z"
          session_id: 01a14f2c-e2e4-7fd3-a08e-129b561c4583
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 3b3545a9cf592714228e6e89cf16395a756cbd04
        tracker_head: 1a31cd21076641eae870767c5dec2d09a37559ca
        branch_point: 3ba12c88814b92d93a6e30300fc8061940382d14
        since: "2026-10-18T05:43:21-07:00"
        until: "2026-10-18T06:21:11-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T05:43:21-07:00"
              end: "2026-10-18T06:21:11-07:00"
              began: "2026-10-18T05:43:21-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 8ddef8b50646ad153c397e2c5e8133141b80719b90c406e1f10f0e2bc7fb4114
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 105154654
              windows:
                - start: "2026-10-18T05:43:21-07:00"
                  end: "2026-10-18T06:21:11-07:00"
        queue_waits:
            - attempt_id: 97e9c79d9d1d102df2d40c32801f3864
              slot: ariadne:1
              record_digest: 302427f7fcd59fc6f235b929456fe28e0d9dee2ebfd0c5ec057a740997e6047c
              start: "2026-10-18T13:00:46.040063Z"
              end: "2026-10-18T13:09:43.849079Z"
              queue_seconds: 537.809016
              raw_seconds: 537.809016
              exact: true
      "n": 1
      timestamp: "2026-10-18T06:24:38-07:00"
      dispatched: "2026-10-18T06:21:43-07:00"
      agent: codex
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Target tracking retains wire identity after reassignment and guesses through interfaces
          detail: 'pkg/golint/wire.go:152-169 sets variables[name] from an initializer but never invalidates target bindings; branch invalidation at line 188 only handles decoders. A source probe with `var m any = &Peer{}; m = &Local{}; d.Decode(m)` incorrectly reports wire-strict-decode, as does conditional reassignment. Even `var m any = Peer{}; d.Decode(&m)` reports despite encoding/json decoding that interface into a map without rejecting unknown fields (confirmed at runtime). ARCH-PURPOSE: honor the conservative destination contract by distinguishing declared types from initializer guesses and invalidating target provenance across direct, multiple, and conditional assignments. Add regression fixtures for this enumeration.'
          family: wire-target-provenance
          round: 1
        - kind: defect
          id: BR-2
          severity: Critical
          title: Overwriting a decoder leaves stale strictness in the matcher
          detail: 'pkg/golint/wire.go:163-165 skips non-identifier assignment targets, and invalidateAssigned at lines 205-208 has the same limitation. Consequently `d := json.NewDecoder(r); d.DisallowUnknownFields(); *d = *json.NewDecoder(r); d.Decode(&Peer{})` incorrectly reports a strict decode. A source probe reproduced the finding; a runtime control confirmed the overwritten decoder accepts unknown fields. ARCH-PURPOSE: invalidate strictness when assignments overwrite a tracked decoder, including conditional writes, and add negative regression fixtures.'
          family: wire-decoder-strictness-provenance
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: README update appears missing for the wire annotation convention
          detail: 'atlas/workflow/sdlc-binary.md:1918 introduces the contributor-facing `// wire: <reason>` syntax and line 1929 documents its rule-specific suppression, but README.md is unchanged in the pinned range. The review’s README gate covers new surface a reader types. Add a concise annotation example and link to the atlas coverage and exception guidance.'
          family: readme-user-surface
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#440 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T06:24:38-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `wire-target-provenance` (kind: defect) Target tracking retains wire identity after reassignment and guesses through interfaces
  pkg/golint/wire.go:152-169 sets variables[name] from an initializer but never invalidates target bindings; branch invalidation at line 188 only handles decoders. A source probe with `var m any = &Peer{}; m = &Local{}; d.Decode(m)` incorrectly reports wire-strict-decode, as does conditional reassignment. Even `var m any = Peer{}; d.Decode(&m)` reports despite encoding/json decoding that interface into a map without rejecting unknown fields (confirmed at runtime). ARCH-PURPOSE: honor the conservative destination contract by distinguishing declared types from initializer guesses and invalidating target provenance across direct, multiple, and conditional assignments. Add regression fixtures for this enumeration.
- **BR-2** [Critical] `wire-decoder-strictness-provenance` (kind: defect) Overwriting a decoder leaves stale strictness in the matcher
  pkg/golint/wire.go:163-165 skips non-identifier assignment targets, and invalidateAssigned at lines 205-208 has the same limitation. Consequently `d := json.NewDecoder(r); d.DisallowUnknownFields(); *d = *json.NewDecoder(r); d.Decode(&Peer{})` incorrectly reports a strict decode. A source probe reproduced the finding; a runtime control confirmed the overwritten decoder accepts unknown fields. ARCH-PURPOSE: invalidate strictness when assignments overwrite a tracked decoder, including conditional writes, and add negative regression fixtures.
- **BR-3** [Important] `readme-user-surface` (kind: defect) README update appears missing for the wire annotation convention
  atlas/workflow/sdlc-binary.md:1918 introduces the contributor-facing `// wire: <reason>` syntax and line 1929 documents its rule-specific suppression, but README.md is unchanged in the pinned range. The review’s README gate covers new surface a reader types. Add a concise annotation example and link to the atlas coverage and exception guidance.

## Open findings

- **BR-1** [Critical] `wire-target-provenance` (kind: defect) Target tracking retains wire identity after reassignment and guesses through interfaces
- **BR-2** [Critical] `wire-decoder-strictness-provenance` (kind: defect) Overwriting a decoder leaves stale strictness in the matcher
- **BR-3** [Important] `readme-user-surface` (kind: defect) README update appears missing for the wire annotation convention
