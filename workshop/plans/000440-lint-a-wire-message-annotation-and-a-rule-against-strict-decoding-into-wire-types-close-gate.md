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
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: e8d57f8c-47c9-4b65-9b4b-56c4cc350a03
            repository: github.com/xianxu/ariadne
            issue_id: "000440"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/8e9f946c7f1f5851
            reviewed_head: 42793ebdf419c790073bcc343420225349588850
            slot: ariadne:1
            requested_at: "2026-10-18T13:49:21.848231Z"
          phase: finished
          pid: 62860
          started_at: "2026-10-18T13:49:21.942247Z"
          finished_at: "2026-10-18T13:51:13.914089Z"
          session_id: 01a14f46-2f8a-7193-bd3c-d7d7e7b8e026
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 42793ebdf419c790073bcc343420225349588850
        tracker_head: 1a31cd21076641eae870767c5dec2d09a37559ca
        branch_point: 3ba12c88814b92d93a6e30300fc8061940382d14
        since: "2026-10-18T05:43:21-07:00"
        until: "2026-10-18T06:48:51-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T05:43:21-07:00"
              end: "2026-10-18T06:48:51-07:00"
              began: "2026-10-18T05:43:21-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 20598f837d7faea895c17c27513826b6de6946048af34e03d5ee0b5c4be800f7
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 105949517
              windows:
                - start: "2026-10-18T05:43:21-07:00"
                  end: "2026-10-18T06:48:51-07:00"
        queue_waits:
            - attempt_id: 97e9c79d9d1d102df2d40c32801f3864
              slot: ariadne:1
              record_digest: 302427f7fcd59fc6f235b929456fe28e0d9dee2ebfd0c5ec057a740997e6047c
              start: "2026-10-18T13:00:46.040063Z"
              end: "2026-10-18T13:09:43.849079Z"
              queue_seconds: 537.809016
              raw_seconds: 537.809016
              exact: true
            - attempt_id: 8c5a5e18aa58707db347c709c810f563
              slot: ariadne:1
              record_digest: e3d4f017eb327e3efee5b4d6035c7cfc077e114ad6e05dcb010f951cd3ff0b17
              start: "2026-10-18T13:27:17.095135Z"
              end: "2026-10-18T13:38:49.768584Z"
              queue_seconds: 692.673449
              raw_seconds: 692.673449
              exact: true
        cuts:
            - dispatched: "2026-10-18T06:21:43-07:00"
              finalized: "2026-10-18T06:24:38-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-18T06:51:14-07:00"
      dispatched: "2026-10-18T06:49:21-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: The explicit-type guard and assignment invalidation cover the reported interface and target-reassignment cases. All five corresponding regression cases pass at head and fail with the pre-fix matcher.
          round: 2
        - id: BR-2
          disposition: not-addressed
          note: 'pkg/golint/wire.go:173-174 and 208-219 invalidate only the written identifier, and the branch walker recognizes only ast.AssignStmt. Three alias-write probes and one range-reassignment probe still produce false wire-strict-decode findings. For example, after d.DisallowUnknownFields(), alias := d; *alias = *json.NewDecoder(r) leaves d marked strict. Likewise, for _, d = range []*json.Decoder{json.NewDecoder(r)} {} does not invalidate d. Runtime controls confirm the overwritten decoders accept unknown fields. ARCH-PURPOSE: the correction must cover the overwrite class. This is the 2nd finding in family `wire-decoder-strictness-provenance`. Do NOT fix only these instances: state and enforce a rule that invalidates strictness whenever an unmodeled alias can overwrite the decoder, and enumerate all assignment forms, including range bindings. Add regression fixtures for that enumeration.'
          round: 2
        - id: BR-3
          disposition: addressed
          note: README.md:11-17 now documents the wire annotation syntax, gives a peer-message example, distinguishes version-pinned stores, and links to atlas/workflow/sdlc-binary.md, whose convention section documents coverage and reasoned exceptions.
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 88d7039d-6900-402d-96a5-4583527aa1dc
            repository: github.com/xianxu/ariadne
            issue_id: "000440"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/828ff8ebde57f8f1
            reviewed_head: ebc57fade3ae567df3aec3f1268a9782b73611c6
            slot: ariadne:1
            requested_at: "2026-10-18T14:26:15.537793Z"
          phase: finished
          pid: 29657
          started_at: "2026-10-18T14:26:15.766551Z"
          finished_at: "2026-10-18T14:27:54.686121Z"
          session_id: 01a14f67-f86c-7b51-a0bd-1010b9ca566d
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: ebc57fade3ae567df3aec3f1268a9782b73611c6
        tracker_head: 1a31cd21076641eae870767c5dec2d09a37559ca
        branch_point: 3ba12c88814b92d93a6e30300fc8061940382d14
        since: "2026-10-18T05:43:21-07:00"
        until: "2026-10-18T07:25:27-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T05:43:21-07:00"
              end: "2026-10-18T07:25:27-07:00"
              began: "2026-10-18T05:43:21-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 3910cf321d3f330aff1a26a69072f62109d43d900fcf2825b256cd4fc295c391
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 106621296
              windows:
                - start: "2026-10-18T05:43:21-07:00"
                  end: "2026-10-18T07:25:27-07:00"
        queue_waits:
            - attempt_id: 97e9c79d9d1d102df2d40c32801f3864
              slot: ariadne:1
              record_digest: 302427f7fcd59fc6f235b929456fe28e0d9dee2ebfd0c5ec057a740997e6047c
              start: "2026-10-18T13:00:46.040063Z"
              end: "2026-10-18T13:09:43.849079Z"
              queue_seconds: 537.809016
              raw_seconds: 537.809016
              exact: true
            - attempt_id: 8c5a5e18aa58707db347c709c810f563
              slot: ariadne:1
              record_digest: e3d4f017eb327e3efee5b4d6035c7cfc077e114ad6e05dcb010f951cd3ff0b17
              start: "2026-10-18T13:27:17.095135Z"
              end: "2026-10-18T13:38:49.768584Z"
              queue_seconds: 692.673449
              raw_seconds: 692.673449
              exact: true
            - attempt_id: 90320ab29b17a67bea80a2671c62c12a
              slot: ariadne:1
              record_digest: 2d18cf65a08f926a2e69fb28f4d9985e7505b65b6ed8c309fe046dc6d1471a51
              start: "2026-10-18T13:54:20.761552Z"
              end: "2026-10-18T14:15:05.235935Z"
              queue_seconds: 1244.474383
              raw_seconds: 1244.474383
              exact: true
        cuts:
            - dispatched: "2026-10-18T06:21:43-07:00"
              finalized: "2026-10-18T06:24:38-07:00"
              finalizing: false
            - dispatched: "2026-10-18T06:49:21-07:00"
              finalized: "2026-10-18T06:51:14-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-18T07:27:56-07:00"
      dispatched: "2026-10-18T07:26:15-07:00"
      agent: codex
      dispose:
        - id: BR-2
          disposition: not-addressed
          note: 'pkg/golint/wire.go:185-186 restores decoder proof after reassignment without remembering that the variable''s address escaped. Reproduced at pinned head with: d := json.NewDecoder(r); ref := &d; d = json.NewDecoder(r); d.DisallowUnknownFields(); *ref = json.NewDecoder(r); _ = d.Decode(&Peer{}). The matcher reports wire-strict-decode, while a runtime control decodes an unknown field successfully. Existing escape regressions fail when discardEscapedDecoders is removed, but omit this escape-before-reestablishment sequence. ARCH-PURPOSE / ARCH-ORDER: retain escape information for each lexical binding, including before its first tracked constructor; distinguish aliases to an old decoder value from aliases or closures retaining the variable. Sweep address-taking and binding-capture paths before allowing proof restoration. This repeats family wire-decoder-strictness-provenance; fix the governing rule rather than adding another overwrite special case.'
          round: 3
        - id: BR-1
          disposition: addressed
          note: 'Retained: wire.go:159 excludes initializer inference for explicitly typed interface destinations; assignment and range invalidation clear target identity. The corresponding interface and reassignment fixtures pass.'
          round: 3
        - id: BR-3
          disposition: addressed
          note: 'Retained: README.md:11-17 documents the annotation syntax, producer rationale, local-store exclusion, and atlas link. These agree with markedWireTypes and the atlas convention.'
          round: 3
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

## Round 2 — 2026-10-18T06:51:14-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — The explicit-type guard and assignment invalidation cover the reported interface and target-reassignment cases. All five corresponding regression cases pass at head and fail with the pre-fix matcher.
- BR-2 — not-addressed — pkg/golint/wire.go:173-174 and 208-219 invalidate only the written identifier, and the branch walker recognizes only ast.AssignStmt. Three alias-write probes and one range-reassignment probe still produce false wire-strict-decode findings. For example, after d.DisallowUnknownFields(), alias := d; *alias = *json.NewDecoder(r) leaves d marked strict. Likewise, for _, d = range []*json.Decoder{json.NewDecoder(r)} {} does not invalidate d. Runtime controls confirm the overwritten decoders accept unknown fields. ARCH-PURPOSE: the correction must cover the overwrite class. This is the 2nd finding in family `wire-decoder-strictness-provenance`. Do NOT fix only these instances: state and enforce a rule that invalidates strictness whenever an unmodeled alias can overwrite the decoder, and enumerate all assignment forms, including range bindings. Add regression fixtures for that enumeration.
- BR-3 — addressed — README.md:11-17 now documents the wire annotation syntax, gives a peer-message example, distinguishes version-pinned stores, and links to atlas/workflow/sdlc-binary.md, whose convention section documents coverage and reasoned exceptions.

## Round 3 — 2026-10-18T07:27:56-07:00 (codex) — BLOCKED

### Disposed

- BR-2 — not-addressed — pkg/golint/wire.go:185-186 restores decoder proof after reassignment without remembering that the variable's address escaped. Reproduced at pinned head with: d := json.NewDecoder(r); ref := &d; d = json.NewDecoder(r); d.DisallowUnknownFields(); *ref = json.NewDecoder(r); _ = d.Decode(&Peer{}). The matcher reports wire-strict-decode, while a runtime control decodes an unknown field successfully. Existing escape regressions fail when discardEscapedDecoders is removed, but omit this escape-before-reestablishment sequence. ARCH-PURPOSE / ARCH-ORDER: retain escape information for each lexical binding, including before its first tracked constructor; distinguish aliases to an old decoder value from aliases or closures retaining the variable. Sweep address-taking and binding-capture paths before allowing proof restoration. This repeats family wire-decoder-strictness-provenance; fix the governing rule rather than adding another overwrite special case.
- BR-1 — addressed — Retained: wire.go:159 excludes initializer inference for explicitly typed interface destinations; assignment and range invalidation clear target identity. The corresponding interface and reassignment fixtures pass.
- BR-3 — addressed — Retained: README.md:11-17 documents the annotation syntax, producer rationale, local-store exclusion, and atlas link. These agree with markedWireTypes and the atlas convention.

## Open findings

- **BR-2** [Critical] `wire-decoder-strictness-provenance` (kind: defect) Overwriting a decoder leaves stale strictness in the matcher
