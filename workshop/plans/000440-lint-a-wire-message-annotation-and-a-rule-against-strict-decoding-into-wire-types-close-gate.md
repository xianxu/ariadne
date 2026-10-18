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
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a07c7163-0fa0-4934-b10f-86cc1b19e050
            repository: github.com/xianxu/ariadne
            issue_id: "000440"
            recipe: milestone-review
            round: 4
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/f643b9c0c33e06b1
            reviewed_head: a60a6e6028c110b556781920e1b6cc2ed6a5b8cf
            slot: ariadne:1
            requested_at: "2026-10-18T15:09:38.35845Z"
          phase: finished
          pid: 77485
          started_at: "2026-10-18T15:09:38.473035Z"
          finished_at: "2026-10-18T15:11:48.628032Z"
          session_id: 01a14f8f-ae14-78c2-8a4c-52dddc1a0cbe
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: a60a6e6028c110b556781920e1b6cc2ed6a5b8cf
        tracker_head: ddb13de8f3ffaba6c8e3b79b55fa0cf525bc98bd
        branch_point: 3ba12c88814b92d93a6e30300fc8061940382d14
        since: "2026-10-18T05:43:21-07:00"
        until: "2026-10-18T08:09:07-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T05:43:21-07:00"
              end: "2026-10-18T08:09:07-07:00"
              began: "2026-10-18T05:43:21-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: ca1506e205a382519f8aaac07ebf2e104ef11cfde338df965816df1f98f96853
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 107386444
              windows:
                - start: "2026-10-18T05:43:21-07:00"
                  end: "2026-10-18T08:09:07-07:00"
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
            - attempt_id: 4a60e90cc408df07a4638b2c53e283ac
              slot: ariadne:1
              record_digest: b4058fa81683863384dde7421a488aa896deb34a4d79ace8aeaa02202c45ca6e
              start: "2026-10-18T14:32:04.166177Z"
              end: "2026-10-18T14:59:49.477614Z"
              queue_seconds: 1665.311437
              raw_seconds: 1665.311437
              exact: true
        cuts:
            - dispatched: "2026-10-18T06:21:43-07:00"
              finalized: "2026-10-18T06:24:38-07:00"
              finalizing: false
            - dispatched: "2026-10-18T06:49:21-07:00"
              finalized: "2026-10-18T06:51:14-07:00"
              finalizing: false
            - dispatched: "2026-10-18T07:26:15-07:00"
              finalized: "2026-10-18T07:27:56-07:00"
              finalizing: false
      "n": 4
      timestamp: "2026-10-18T08:11:49-07:00"
      dispatched: "2026-10-18T08:09:38-07:00"
      agent: codex
      dispose:
        - id: BR-2
          disposition: addressed
          note: Whole-function eligibility rejects pointer writes, aliases, escapes, range assignments, and repeated assignments. The pinned-head regression matrix passes; substituting wire.go from 826f9776 makes the original indirect-overwrite fixtures fail, and substituting 848565fa makes the later binding-escape fixtures fail.
          round: 4
        - id: BR-1
          disposition: addressed
          note: Target reassignment and explicit-interface negative fixtures pass at the pinned head; these fixtures fail with the original matcher from 826f9776.
          round: 4
        - id: BR-3
          disposition: addressed
          note: README.md:11 documents the wire annotation, independently evolving producers, and local-store exclusion, and links the supporting convention in atlas/workflow/sdlc-binary.md.
          round: 4
      findings:
        - kind: defect
          id: BR-4
          severity: Critical
          title: Imported marked types at the module root bypass the rule
          detail: pkg/golint/wire.go:384 accepts only imports beginning with Module+"/", excluding an import equal to Module. A source-map probe with annotated Message in message.go and a direct strict decode into peer.Message from reader/read.go importing peer "example.test/module" returns no finding; moving the declaration and import to the wire subpackage produces the expected finding. This violates the Spec's module-wide imported-type coverage (ARCH-PURPOSE). This is the 2nd finding in family `wire-target-provenance`. Earlier rounds fixed instances; state and enforce the rule that every supported target identity resolves consistently. Sweep same-file, cross-file, module-root and subpackage imports, explicit/default aliases, and shadowed identities. Normalize an exact module import to the index's "." directory key and add regression coverage; that correction made the root/subpackage control probe pass.
          family: wire-target-provenance
          round: 4
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: d98742e5-d605-4956-97b9-4d3776435e53
            repository: github.com/xianxu/ariadne
            issue_id: "000440"
            recipe: milestone-review
            round: 5
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/07b900c5f9e838f4
            reviewed_head: ae9feba152a8b629675f674bd5a754a9fef0bfac
            slot: ariadne:1
            requested_at: "2026-10-18T15:48:25.369234Z"
          phase: finished
          pid: 43031
          started_at: "2026-10-18T15:48:25.456178Z"
          finished_at: "2026-10-18T15:50:11.056506Z"
          session_id: 01a14fb3-2fd3-7522-a7e3-fe2005df4d10
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: ae9feba152a8b629675f674bd5a754a9fef0bfac
        tracker_head: fbe5c040cb6ebc956bb311c368442779cb27fa5c
        branch_point: 3ba12c88814b92d93a6e30300fc8061940382d14
        since: "2026-10-18T05:43:21-07:00"
        until: "2026-10-18T08:47:52-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T05:43:21-07:00"
              end: "2026-10-18T08:47:52-07:00"
              began: "2026-10-18T05:43:21-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: dc0736cba4c7f9271d270f15771f759dba01024ae061c9f162bc8969255d2ccd
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 108018579
              windows:
                - start: "2026-10-18T05:43:21-07:00"
                  end: "2026-10-18T08:47:52-07:00"
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
            - attempt_id: 4a60e90cc408df07a4638b2c53e283ac
              slot: ariadne:1
              record_digest: b4058fa81683863384dde7421a488aa896deb34a4d79ace8aeaa02202c45ca6e
              start: "2026-10-18T14:32:04.166177Z"
              end: "2026-10-18T14:59:49.477614Z"
              queue_seconds: 1665.311437
              raw_seconds: 1665.311437
              exact: true
            - attempt_id: 28f483f4799341fd56dad2c463677506
              slot: ariadne:1
              record_digest: 1f189210f8493e3900895b7da9b87104d5b38c60a7c2af139e525891aeeb29dc
              start: "2026-10-18T15:15:06.25148Z"
              end: "2026-10-18T15:39:30.12085Z"
              queue_seconds: 1463.86937
              raw_seconds: 1463.86937
              exact: true
        cuts:
            - dispatched: "2026-10-18T06:21:43-07:00"
              finalized: "2026-10-18T06:24:38-07:00"
              finalizing: false
            - dispatched: "2026-10-18T06:49:21-07:00"
              finalized: "2026-10-18T06:51:14-07:00"
              finalizing: false
            - dispatched: "2026-10-18T07:26:15-07:00"
              finalized: "2026-10-18T07:27:56-07:00"
              finalizing: false
            - dispatched: "2026-10-18T08:09:38-07:00"
              finalized: "2026-10-18T08:11:49-07:00"
              finalizing: false
      "n": 5
      timestamp: "2026-10-18T08:50:11-07:00"
      dispatched: "2026-10-18T08:48:25-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-4
          disposition: addressed
          note: Exact-module imports map to "." at pkg/golint/wire.go:392. TestWireTargetIdentityMatrix covers default and explicit root aliases; both cases fail when that correction is removed in a temporary package copy.
          round: 5
        - id: BR-1
          disposition: addressed
          note: 'Retained disposition: target writes invalidate provenance, explicit interfaces do not inherit initializer identity, and the corresponding regression cases pass.'
          round: 5
        - id: BR-2
          disposition: not-addressed
          note: 'pkg/golint/wire.go:264-268 and 301-303 accept a package-global d assigned once inside the inspected function. Reproduced sequence: d = json.NewDecoder(r); d.DisallowUnknownFields(); reset(r); var m Peer; return d.Decode(&m), where reset replaces global d with json.NewDecoder(r). Check reports wire-strict-decode although the runtime decode accepts an unknown field with no error. This repeats family wire-decoder-strictness-provenance (ARCH-PURPOSE): enforce the documented function-local binding requirement before applying the single-owner proof, sweep same-file and cross-file globals alongside local declarations and shadowing, and add regression coverage.'
          round: 5
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 746c61e2-145a-4c3c-9cb0-96808e3bb524
            repository: github.com/xianxu/ariadne
            issue_id: "000440"
            recipe: milestone-review
            round: 6
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/8d48ef835946e31e
            reviewed_head: 2e2af4e0e8d9645aaa8009e78608abdd8ff5aded
            slot: ariadne:1
            requested_at: "2026-10-18T16:27:49.371898Z"
          phase: finished
          pid: 29786
          started_at: "2026-10-18T16:27:49.461672Z"
          finished_at: "2026-10-18T16:29:49.172195Z"
          session_id: 01a14fd7-4248-7123-bbfe-7f8db28dd49b
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 2e2af4e0e8d9645aaa8009e78608abdd8ff5aded
        tracker_head: e655ff87803ccaad9f0054eadd97f53eb740d0f9
        branch_point: 3ba12c88814b92d93a6e30300fc8061940382d14
        since: "2026-10-18T05:43:21-07:00"
        until: "2026-10-18T09:27:15-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T05:43:21-07:00"
              end: "2026-10-18T09:27:15-07:00"
              began: "2026-10-18T05:43:21-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: c7c07b2718ba83c93571d3bc634fcaa5d24691d201223b14b5379d49f24d3180
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 109237234
              windows:
                - start: "2026-10-18T05:43:21-07:00"
                  end: "2026-10-18T09:27:15-07:00"
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
            - attempt_id: 4a60e90cc408df07a4638b2c53e283ac
              slot: ariadne:1
              record_digest: b4058fa81683863384dde7421a488aa896deb34a4d79ace8aeaa02202c45ca6e
              start: "2026-10-18T14:32:04.166177Z"
              end: "2026-10-18T14:59:49.477614Z"
              queue_seconds: 1665.311437
              raw_seconds: 1665.311437
              exact: true
            - attempt_id: 28f483f4799341fd56dad2c463677506
              slot: ariadne:1
              record_digest: 1f189210f8493e3900895b7da9b87104d5b38c60a7c2af139e525891aeeb29dc
              start: "2026-10-18T15:15:06.25148Z"
              end: "2026-10-18T15:39:30.12085Z"
              queue_seconds: 1463.86937
              raw_seconds: 1463.86937
              exact: true
            - attempt_id: 77d0c88769993b116c9a81f17f6899e5
              slot: ariadne:1
              record_digest: 691c68f7041c15337939e4f7fd16375e059cabd17226d2133de3bd4015692592
              start: "2026-10-18T15:53:13.613306Z"
              end: "2026-10-18T16:18:19.150532Z"
              queue_seconds: 1505.537226
              raw_seconds: 1505.537226
              exact: true
        cuts:
            - dispatched: "2026-10-18T06:21:43-07:00"
              finalized: "2026-10-18T06:24:38-07:00"
              finalizing: false
            - dispatched: "2026-10-18T06:49:21-07:00"
              finalized: "2026-10-18T06:51:14-07:00"
              finalizing: false
            - dispatched: "2026-10-18T07:26:15-07:00"
              finalized: "2026-10-18T07:27:56-07:00"
              finalizing: false
            - dispatched: "2026-10-18T08:09:38-07:00"
              finalized: "2026-10-18T08:11:49-07:00"
              finalizing: false
            - dispatched: "2026-10-18T08:48:25-07:00"
              finalized: "2026-10-18T08:50:11-07:00"
              finalizing: false
      "n": 6
      timestamp: "2026-10-18T09:29:49-07:00"
      dispatched: "2026-10-18T09:27:49-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-2
          disposition: addressed
          note: Pointer-overwrite and global-reset regressions pass at HEAD and fail when the corresponding protections are removed in disposable copies.
          round: 6
        - id: BR-1
          disposition: addressed
          note: Target-rebinding and interface fixtures pass; removing target invalidation makes direct, multiple, conditional, and range-rebinding regressions fail.
          round: 6
        - id: BR-3
          disposition: addressed
          note: README.md:11 now documents the wire annotation syntax, independently evolving producers, local-store exclusion, and atlas link, consistent with the matcher and atlas.
          round: 6
        - id: BR-4
          disposition: addressed
          note: Module-root default-name and alias fixtures pass and both fail when exact-module import resolution is removed.
          round: 6
      findings:
        - kind: defect
          id: BR-5
          severity: Critical
          title: Direct Decode calls in variable initializers bypass the rule
          detail: 'pkg/golint/wire.go:143-161 processes declaration bindings without checking initializer calls. With an annotated Peer, `d := json.NewDecoder(r); d.DisallowUnknownFields(); var m Peer; var err = d.Decode(&m); return err` produces no finding, while changing only `var err =` to `err :=` produces one. Typed and grouped var declarations also bypass detection. ARCH-PURPOSE: enumerate supported unconditional expression positions, check direct Decode calls consistently before applying binding changes, and add regression fixtures for all declaration forms.'
          family: wire-direct-call-coverage
          round: 6
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 3174a9bd-91c4-4520-94b8-6f321625d2c4
            repository: github.com/xianxu/ariadne
            issue_id: "000440"
            recipe: milestone-review
            round: 7
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/c794b93980d18ebe
            reviewed_head: dd1e08a0ef40df3b15ee27d5ba716ba6ed04485b
            slot: ariadne:1
            requested_at: "2026-10-18T16:48:18.615887Z"
          phase: finished
          pid: 17954
          started_at: "2026-10-18T16:48:18.738582Z"
          finished_at: "2026-10-18T16:49:59.396248Z"
          session_id: 01a14fea-043b-7001-9335-fa8de676f891
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: dd1e08a0ef40df3b15ee27d5ba716ba6ed04485b
        tracker_head: b529ce35026932a9477ef23c38f1372e147336a7
        branch_point: 3ba12c88814b92d93a6e30300fc8061940382d14
        since: "2026-10-18T05:43:21-07:00"
        until: "2026-10-18T09:47:50-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T05:43:21-07:00"
              end: "2026-10-18T09:47:50-07:00"
              began: "2026-10-18T05:43:21-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 1156d92ed4693479b6ceceba82d71158d63149388b44d0856194a2e3a0de3e83
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 109679893
              windows:
                - start: "2026-10-18T05:43:21-07:00"
                  end: "2026-10-18T09:47:50-07:00"
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
            - attempt_id: 4a60e90cc408df07a4638b2c53e283ac
              slot: ariadne:1
              record_digest: b4058fa81683863384dde7421a488aa896deb34a4d79ace8aeaa02202c45ca6e
              start: "2026-10-18T14:32:04.166177Z"
              end: "2026-10-18T14:59:49.477614Z"
              queue_seconds: 1665.311437
              raw_seconds: 1665.311437
              exact: true
            - attempt_id: 28f483f4799341fd56dad2c463677506
              slot: ariadne:1
              record_digest: 1f189210f8493e3900895b7da9b87104d5b38c60a7c2af139e525891aeeb29dc
              start: "2026-10-18T15:15:06.25148Z"
              end: "2026-10-18T15:39:30.12085Z"
              queue_seconds: 1463.86937
              raw_seconds: 1463.86937
              exact: true
            - attempt_id: 77d0c88769993b116c9a81f17f6899e5
              slot: ariadne:1
              record_digest: 691c68f7041c15337939e4f7fd16375e059cabd17226d2133de3bd4015692592
              start: "2026-10-18T15:53:13.613306Z"
              end: "2026-10-18T16:18:19.150532Z"
              queue_seconds: 1505.537226
              raw_seconds: 1505.537226
              exact: true
            - attempt_id: 632f45e915f1038d4da0444be6b2d4c6
              slot: ariadne:1
              record_digest: 8a092f4b63f32315462d1158109232274b3edd8788c472a9da1038c1571854a3
              start: "2026-10-18T16:31:45.679488Z"
              end: "2026-10-18T16:38:32.998977Z"
              queue_seconds: 407.319489
              raw_seconds: 407.319489
              exact: true
        cuts:
            - dispatched: "2026-10-18T06:21:43-07:00"
              finalized: "2026-10-18T06:24:38-07:00"
              finalizing: false
            - dispatched: "2026-10-18T06:49:21-07:00"
              finalized: "2026-10-18T06:51:14-07:00"
              finalizing: false
            - dispatched: "2026-10-18T07:26:15-07:00"
              finalized: "2026-10-18T07:27:56-07:00"
              finalizing: false
            - dispatched: "2026-10-18T08:09:38-07:00"
              finalized: "2026-10-18T08:11:49-07:00"
              finalizing: false
            - dispatched: "2026-10-18T08:48:25-07:00"
              finalized: "2026-10-18T08:50:11-07:00"
              finalizing: false
            - dispatched: "2026-10-18T09:27:49-07:00"
              finalized: "2026-10-18T09:29:49-07:00"
              finalizing: false
      "n": 7
      timestamp: "2026-10-18T09:50:00-07:00"
      dispatched: "2026-10-18T09:48:18-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Retain the accepted disposition; target reassignment, interface, and shadowing fixtures pass.
          round: 7
        - id: BR-2
          disposition: addressed
          note: Retain the accepted disposition; whole-function ownership checks and alias, escape, reassignment, and global-binding fixtures pass.
          round: 7
        - id: BR-3
          disposition: addressed
          note: README.md documents the annotation syntax and links the atlas convention, consistent with annotation parsing in wire.go.
          round: 7
        - id: BR-4
          disposition: addressed
          note: Exact-module imports resolve to "."; the root/subpackage identity matrix passes.
          round: 7
        - id: BR-5
          disposition: addressed
          note: TestWireStrictDecodeDirectCallPositions passes at head. Substituting the pre-fix matcher in a temporary copy makes inferred, typed, grouped var, and multiple-RHS cases fail; the copy was removed.
          round: 7
      findings:
        - kind: defect
          id: BR-6
          severity: Critical
          title: Direct expression calls in if initializers bypass the rule
          detail: 'pkg/golint/wire.go:200 only checks an if initializer when it is an AssignStmt. With an eligible strict decoder and marked destination, `if d.Decode(&m); true {}` produces no finding. Likewise, `if d.DisallowUnknownFields(); true {}; return d.Decode(&m)` loses the unconditional strictness call. Both were reproduced against the pinned head without parse errors; the assignment-initializer control reports correctly. ARCH-PURPOSE: the plan and atlas explicitly include direct calls in if initializers. This is the 2nd finding in family `wire-direct-call-coverage`. Do NOT fix only these instances: state the rule for every supported direct-call position, route supported initializer statements through shared handling, and test the expression/assignment forms with strictness-order controls.'
          family: wire-direct-call-coverage
          round: 7
      recipe: milestone-review
      isolation: fenced
      blocked: true
disputes:
    - id: BR-6
      rationale: 'The missing expression-form if initializer is a true false negative, but the hub capped coverage widening in #440 after round 6. #440 now explicitly supports assignment-form if initializers; expression-form initializers are filed as unclaimed ariadne#445 with red-fixture acceptance. Keeping a conservative syntax-only rule at this boundary avoids another unplanned coverage expansion; false positives, crashes and misparses remain blockers.'
      evidence: 'go run /tmp/ariadne-440-if-init-probe.go on the current matcher printed assignment-form if err := d.Decode(&m): 1 findings and expression-form if d.Decode(&m): 0 findings. The former is a passing supported control; the latter is the exact #445 follow-up. go test ./pkg/golint -count=1 passed, and the recorded suite on 01d74b7238484230f10740292abf606575f52c52 passed all 12 shards and 68 other packages.'
      after: 7
      timestamp: "2026-10-18T16:52:29Z"
      by: Xian Xu on Xian’s MacBook Pro at ariadne:1 (/Users/xianxu/workspace/worktree/ariadne-slot1/ariadne)
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

## Round 4 — 2026-10-18T08:11:49-07:00 (codex) — BLOCKED

### Disposed

- BR-2 — addressed — Whole-function eligibility rejects pointer writes, aliases, escapes, range assignments, and repeated assignments. The pinned-head regression matrix passes; substituting wire.go from 826f9776 makes the original indirect-overwrite fixtures fail, and substituting 848565fa makes the later binding-escape fixtures fail.
- BR-1 — addressed — Target reassignment and explicit-interface negative fixtures pass at the pinned head; these fixtures fail with the original matcher from 826f9776.
- BR-3 — addressed — README.md:11 documents the wire annotation, independently evolving producers, and local-store exclusion, and links the supporting convention in atlas/workflow/sdlc-binary.md.

### Raised

- **BR-4** [Critical] `wire-target-provenance` (kind: defect) Imported marked types at the module root bypass the rule
  pkg/golint/wire.go:384 accepts only imports beginning with Module+"/", excluding an import equal to Module. A source-map probe with annotated Message in message.go and a direct strict decode into peer.Message from reader/read.go importing peer "example.test/module" returns no finding; moving the declaration and import to the wire subpackage produces the expected finding. This violates the Spec's module-wide imported-type coverage (ARCH-PURPOSE). This is the 2nd finding in family `wire-target-provenance`. Earlier rounds fixed instances; state and enforce the rule that every supported target identity resolves consistently. Sweep same-file, cross-file, module-root and subpackage imports, explicit/default aliases, and shadowed identities. Normalize an exact module import to the index's "." directory key and add regression coverage; that correction made the root/subpackage control probe pass.

## Round 5 — 2026-10-18T08:50:11-07:00 (codex) — BLOCKED

### Disposed

- BR-4 — addressed — Exact-module imports map to "." at pkg/golint/wire.go:392. TestWireTargetIdentityMatrix covers default and explicit root aliases; both cases fail when that correction is removed in a temporary package copy.
- BR-1 — addressed — Retained disposition: target writes invalidate provenance, explicit interfaces do not inherit initializer identity, and the corresponding regression cases pass.
- BR-2 — not-addressed — pkg/golint/wire.go:264-268 and 301-303 accept a package-global d assigned once inside the inspected function. Reproduced sequence: d = json.NewDecoder(r); d.DisallowUnknownFields(); reset(r); var m Peer; return d.Decode(&m), where reset replaces global d with json.NewDecoder(r). Check reports wire-strict-decode although the runtime decode accepts an unknown field with no error. This repeats family wire-decoder-strictness-provenance (ARCH-PURPOSE): enforce the documented function-local binding requirement before applying the single-owner proof, sweep same-file and cross-file globals alongside local declarations and shadowing, and add regression coverage.

## Round 6 — 2026-10-18T09:29:49-07:00 (codex) — BLOCKED

### Disposed

- BR-2 — addressed — Pointer-overwrite and global-reset regressions pass at HEAD and fail when the corresponding protections are removed in disposable copies.
- BR-1 — addressed — Target-rebinding and interface fixtures pass; removing target invalidation makes direct, multiple, conditional, and range-rebinding regressions fail.
- BR-3 — addressed — README.md:11 now documents the wire annotation syntax, independently evolving producers, local-store exclusion, and atlas link, consistent with the matcher and atlas.
- BR-4 — addressed — Module-root default-name and alias fixtures pass and both fail when exact-module import resolution is removed.

### Raised

- **BR-5** [Critical] `wire-direct-call-coverage` (kind: defect) Direct Decode calls in variable initializers bypass the rule
  pkg/golint/wire.go:143-161 processes declaration bindings without checking initializer calls. With an annotated Peer, `d := json.NewDecoder(r); d.DisallowUnknownFields(); var m Peer; var err = d.Decode(&m); return err` produces no finding, while changing only `var err =` to `err :=` produces one. Typed and grouped var declarations also bypass detection. ARCH-PURPOSE: enumerate supported unconditional expression positions, check direct Decode calls consistently before applying binding changes, and add regression fixtures for all declaration forms.

## Round 7 — 2026-10-18T09:50:00-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — Retain the accepted disposition; target reassignment, interface, and shadowing fixtures pass.
- BR-2 — addressed — Retain the accepted disposition; whole-function ownership checks and alias, escape, reassignment, and global-binding fixtures pass.
- BR-3 — addressed — README.md documents the annotation syntax and links the atlas convention, consistent with annotation parsing in wire.go.
- BR-4 — addressed — Exact-module imports resolve to "."; the root/subpackage identity matrix passes.
- BR-5 — addressed — TestWireStrictDecodeDirectCallPositions passes at head. Substituting the pre-fix matcher in a temporary copy makes inferred, typed, grouped var, and multiple-RHS cases fail; the copy was removed.

### Raised

- **BR-6** [Critical] `wire-direct-call-coverage` (kind: defect) Direct expression calls in if initializers bypass the rule
  pkg/golint/wire.go:200 only checks an if initializer when it is an AssignStmt. With an eligible strict decoder and marked destination, `if d.Decode(&m); true {}` produces no finding. Likewise, `if d.DisallowUnknownFields(); true {}; return d.Decode(&m)` loses the unconditional strictness call. Both were reproduced against the pinned head without parse errors; the assignment-initializer control reports correctly. ARCH-PURPOSE: the plan and atlas explicitly include direct calls in if initializers. This is the 2nd finding in family `wire-direct-call-coverage`. Do NOT fix only these instances: state the rule for every supported direct-call position, route supported initializer statements through shared handling, and test the expression/assignment forms with strictness-order controls.

## Disputes

Findings the implementer declined (`sdlc finding dispute`), each answered by the
next review round.

- **BR-6** — after round 7 (2026-10-18T16:52:29Z, Xian Xu on Xian’s MacBook Pro at ariadne:1 (/Users/xianxu/workspace/worktree/ariadne-slot1/ariadne))
  rationale: The missing expression-form if initializer is a true false negative, but the hub capped coverage widening in #440 after round 6. #440 now explicitly supports assignment-form if initializers; expression-form initializers are filed as unclaimed ariadne#445 with red-fixture acceptance. Keeping a conservative syntax-only rule at this boundary avoids another unplanned coverage expansion; false positives, crashes and misparses remain blockers.
  evidence: go run /tmp/ariadne-440-if-init-probe.go on the current matcher printed assignment-form if err := d.Decode(&m): 1 findings and expression-form if d.Decode(&m): 0 findings. The former is a passing supported control; the latter is the exact #445 follow-up. go test ./pkg/golint -count=1 passed, and the recorded suite on 01d74b7238484230f10740292abf606575f52c52 passed all 12 shards and 68 other packages.

## Open findings

- **BR-6** [Critical] `wire-direct-call-coverage` (kind: defect) Direct expression calls in if initializers bypass the rule
