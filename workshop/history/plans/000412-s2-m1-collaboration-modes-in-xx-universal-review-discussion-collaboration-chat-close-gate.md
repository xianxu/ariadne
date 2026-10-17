---
gate: boundary-review
issue: 412
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 18bcd2c0-3703-4b84-8353-59b03ee47297
            repository: github.com/xianxu/ariadne
            issue_id: "000412"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/a271c0b6d57e2f38
            reviewed_head: 031ea5c641fd81964c0e613178f1bd93dced516c
            slot: ariadne:2
            requested_at: "2026-10-16T23:55:50.847012Z"
          phase: finished
          pid: 21500
          started_at: "2026-10-16T23:55:51.782743Z"
          finished_at: "2026-10-16T23:57:40.343763Z"
          session_id: ca937b3c-788e-48c2-a03c-5828a970022d
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 031ea5c641fd81964c0e613178f1bd93dced516c
        tracker_head: 27a568ffc9e65d519a006f41cc07fbe5af3f0e35
        branch_point: 35ba24e6cad39344c347af981de7ac8fcc1d1c31
        since: "2026-10-16T15:42:55-07:00"
        until: "2026-10-16T16:54:25-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T15:42:55-07:00"
              end: "2026-10-16T16:54:25-07:00"
              began: "2026-10-16T15:42:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 6681a7b6-a1e9-4bf6-8743-64ae382a6945
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: a01d2d2e30cdee9e2ab0f79bd4c475b0767f8fb71bde51c321f0d54701105ed7
              started: "2026-10-16T22:49:11.289Z"
              spawned: false
              bytes: 542081
              windows:
                - start: "2026-10-16T22:49:11.289Z"
                  end: "2026-10-16T16:54:25-07:00"
            - harness: claude
              id: a3c26c68-b277-41fb-8571-456fa2883740
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5320c188df9f1d072a2c1e3788ff32b5cea4f732e8e180b077700b714d95ba8f
              started: "2026-10-16T22:55:51.076Z"
              spawned: false
              bytes: 509821
              windows:
                - start: "2026-10-16T22:55:51.076Z"
                  end: "2026-10-16T16:54:25-07:00"
            - harness: claude
              id: c16f9f99-f7dc-493c-80e6-9950883fe1cf
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 56c17381f2e60fc56f2ae5acb83b94366afd3a2dea497b7c768fec2238d6894b
              started: "2026-10-16T22:53:46.176Z"
              spawned: false
              bytes: 499558
              windows:
                - start: "2026-10-16T22:53:46.176Z"
                  end: "2026-10-16T16:54:25-07:00"
            - harness: claude
              id: e1a45820-de66-478e-bdce-95f833a64410
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 84b46a5b1ffbb97690e35548a8500ac62e056483b444935a24569d2a8335bc77
              started: "2026-10-16T22:56:14.41Z"
              spawned: false
              bytes: 394689
              windows:
                - start: "2026-10-16T22:56:14.41Z"
                  end: "2026-10-16T16:54:25-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 98fd170565074499c786ba50f996f5a1538335550a74a2171021994b22a3f7a2
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 17062094
              windows:
                - start: "2026-10-16T15:42:55-07:00"
                  end: "2026-10-16T16:54:25-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 6.5026
      "n": 1
      timestamp: "2026-10-16T16:57:45-07:00"
      dispatched: "2026-10-16T16:55:50-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Side-doc reuse by (parent, marker) alone can link a new ask to an unrelated, already-linked discussion
          detail: "collaboration-modes.md §4 step 1 reuses any side doc whose parent and marker match, even one already linked from another section, so a second identical marker (e.g. \U0001F916[why?] elsewhere) silently gets the old discussion; same-round identical markers are unspecified. Restrict reuse to unlinked docs (and same section for discussion) and add a case or check."
          family: reuse-identity-too-coarse
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: tl-lead's instructionBullets defines an item differently from the chat mode tl-lead --help now cites as the rule
          detail: cmd/tl-lead/instructions.go only starts items at unindented `N. ` lines, treats `→ done:` as answered, and absorbs `-`/`*` lines into the previous bullet, while collaboration-modes.md §3 counts -, * and N. items at the shallowest indent; a `- foo` bullet is invisible to list and gets the previous bullet's link. Align the parser (with a table test) or narrow the help.
          family: cited-rule-restated-divergently
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: Done-when 2 (session file runs on collaboration-mode) depends on the TL's ops write; record it as owed in the landing report
          family: done-when-deferred
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: Examples checker does not assert the (folded <date>) link mark for fold cases (rows 7 and 10)
          family: checker-coverage-gap
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 13685de5-96b5-4f53-b099-6306401ad4b3
            repository: github.com/xianxu/ariadne
            issue_id: "000412"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/4da3b878f24d1b3b
            reviewed_head: 255314b4ec7bd08fa70b031805fab03b12f29a06
            slot: ariadne:2
            requested_at: "2026-10-17T00:01:22.89563Z"
          phase: finished
          pid: 91621
          started_at: "2026-10-17T00:01:23.124489Z"
          finished_at: "2026-10-17T00:02:47.849521Z"
          session_id: 8f9eacdf-f967-44c3-9986-b3b01587aee4
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 255314b4ec7bd08fa70b031805fab03b12f29a06
        tracker_head: 27a568ffc9e65d519a006f41cc07fbe5af3f0e35
        branch_point: 35ba24e6cad39344c347af981de7ac8fcc1d1c31
        since: "2026-10-16T15:42:55-07:00"
        until: "2026-10-16T17:00:38-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T15:42:55-07:00"
              end: "2026-10-16T17:00:38-07:00"
              began: "2026-10-16T15:42:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 6681a7b6-a1e9-4bf6-8743-64ae382a6945
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: a01d2d2e30cdee9e2ab0f79bd4c475b0767f8fb71bde51c321f0d54701105ed7
              started: "2026-10-16T22:49:11.289Z"
              spawned: false
              bytes: 542081
              windows:
                - start: "2026-10-16T22:49:11.289Z"
                  end: "2026-10-16T17:00:38-07:00"
            - harness: claude
              id: a3c26c68-b277-41fb-8571-456fa2883740
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5320c188df9f1d072a2c1e3788ff32b5cea4f732e8e180b077700b714d95ba8f
              started: "2026-10-16T22:55:51.076Z"
              spawned: false
              bytes: 509821
              windows:
                - start: "2026-10-16T22:55:51.076Z"
                  end: "2026-10-16T17:00:38-07:00"
            - harness: claude
              id: c16f9f99-f7dc-493c-80e6-9950883fe1cf
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 56c17381f2e60fc56f2ae5acb83b94366afd3a2dea497b7c768fec2238d6894b
              started: "2026-10-16T22:53:46.176Z"
              spawned: false
              bytes: 499558
              windows:
                - start: "2026-10-16T22:53:46.176Z"
                  end: "2026-10-16T17:00:38-07:00"
            - harness: claude
              id: e1a45820-de66-478e-bdce-95f833a64410
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 84b46a5b1ffbb97690e35548a8500ac62e056483b444935a24569d2a8335bc77
              started: "2026-10-16T22:56:14.41Z"
              spawned: false
              bytes: 394689
              windows:
                - start: "2026-10-16T22:56:14.41Z"
                  end: "2026-10-16T17:00:38-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 613de9bf66a6b721b188ca36b5fbaa0b8d3a6c6f253961de92244a8dec846321
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 17302042
              windows:
                - start: "2026-10-16T15:42:55-07:00"
                  end: "2026-10-16T17:00:38-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 6.5026
        cuts:
            - dispatched: "2026-10-16T16:55:50-07:00"
              finalized: "2026-10-16T16:57:45-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-16T17:02:49-07:00"
      dispatched: "2026-10-16T17:01:22-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Spec §4 step 1 and row 19 now require the same section and an unlinked doc; case discussion-same-marker added. Checker enforcement gap raised separately as checker-coverage-gap.
          round: 2
        - id: BR-2
          disposition: addressed
          note: instructionBullets now follows the §3 item rule; TestInstructionBulletsAreChatItems fails on the old N.-only parser.
          round: 2
        - id: BR-3
          disposition: addressed
          note: Plan revisions (plan line ~264) record Done-when 2 as owed by the TL in the landing report.
          round: 2
        - id: BR-4
          disposition: addressed
          note: foldedRE check added, with a test that breaks a fixture's fold mark and expects the checker to fail.
          round: 2
      findings:
        - kind: defect
          id: BR-5
          severity: Important
          title: 'Example checker does not enforce row 19: a side doc that is already linked can be reused with no failure'
          detail: 'I edited discussion-same-marker/after so both asks link to 01-why-per-file.md and deleted 02, which is the BR-1 defect, and TestExamples still passed. This is the 2nd finding in the family. The rule covering the class: in a parent, each side doc is the target of at most one link. Enforce it in checkTree and add a test that links two asks to one doc and expects a failure.'
          family: checker-coverage-gap
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 2aaa082b-6e39-4a43-967a-2e206e39407b
            repository: github.com/xianxu/ariadne
            issue_id: "000412"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/1676c00243d2f2c8
            reviewed_head: e123c3ca7429ac6e50381f13a646c024a08ea631
            slot: ariadne:2
            requested_at: "2026-10-17T00:04:39.693039Z"
          phase: finished
          pid: 26230
          started_at: "2026-10-17T00:04:39.893494Z"
          finished_at: "2026-10-17T00:06:56.755001Z"
          session_id: e99fcf6d-5336-48db-94dd-b43f505f70d6
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: e123c3ca7429ac6e50381f13a646c024a08ea631
        tracker_head: 27a568ffc9e65d519a006f41cc07fbe5af3f0e35
        branch_point: 35ba24e6cad39344c347af981de7ac8fcc1d1c31
        since: "2026-10-16T15:42:55-07:00"
        until: "2026-10-16T17:04:01-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T15:42:55-07:00"
              end: "2026-10-16T17:04:01-07:00"
              began: "2026-10-16T15:42:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 6681a7b6-a1e9-4bf6-8743-64ae382a6945
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: a01d2d2e30cdee9e2ab0f79bd4c475b0767f8fb71bde51c321f0d54701105ed7
              started: "2026-10-16T22:49:11.289Z"
              spawned: false
              bytes: 542081
              windows:
                - start: "2026-10-16T22:49:11.289Z"
                  end: "2026-10-16T17:04:01-07:00"
            - harness: claude
              id: a3c26c68-b277-41fb-8571-456fa2883740
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5320c188df9f1d072a2c1e3788ff32b5cea4f732e8e180b077700b714d95ba8f
              started: "2026-10-16T22:55:51.076Z"
              spawned: false
              bytes: 509821
              windows:
                - start: "2026-10-16T22:55:51.076Z"
                  end: "2026-10-16T17:04:01-07:00"
            - harness: claude
              id: c16f9f99-f7dc-493c-80e6-9950883fe1cf
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 56c17381f2e60fc56f2ae5acb83b94366afd3a2dea497b7c768fec2238d6894b
              started: "2026-10-16T22:53:46.176Z"
              spawned: false
              bytes: 499558
              windows:
                - start: "2026-10-16T22:53:46.176Z"
                  end: "2026-10-16T17:04:01-07:00"
            - harness: claude
              id: e1a45820-de66-478e-bdce-95f833a64410
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 84b46a5b1ffbb97690e35548a8500ac62e056483b444935a24569d2a8335bc77
              started: "2026-10-16T22:56:14.41Z"
              spawned: false
              bytes: 394689
              windows:
                - start: "2026-10-16T22:56:14.41Z"
                  end: "2026-10-16T17:04:01-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: e158affd1686ccbf9ef37ac3f0cd60e180e3db935ce5a4b2f14b15f9413d53c3
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 17415023
              windows:
                - start: "2026-10-16T15:42:55-07:00"
                  end: "2026-10-16T17:04:01-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 6.5026
        cuts:
            - dispatched: "2026-10-16T16:55:50-07:00"
              finalized: "2026-10-16T16:57:45-07:00"
              finalizing: false
            - dispatched: "2026-10-16T17:01:22-07:00"
              finalized: "2026-10-16T17:02:49-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-16T17:06:58-07:00"
      dispatched: "2026-10-16T17:04:39-07:00"
      agent: claude
      dispose:
        - id: BR-5
          disposition: addressed
          note: Reviewer's regression (both asks -> 01, 02 deleted) now fails TestExamples; disabling the targets check turns "a linked doc reused" red.
          round: 3
      findings:
        - kind: defect
          id: BR-6
          severity: Critical
          title: '"Linked = answer link ending the item''s last line" makes a row-14 follow-up item read as unlinked; tl-lead double-links it'
          detail: 'e123c3ca changed collaboration-modes.md:115-116 and instructions.go answeredRE (line 21, used at 96) so only a link on the last line counts. examples/chat item 3 ("3. ... → [answer](...)" then "   - and by repo?") is then a new ask by the spec''s own rule, which contradicts row 14 and the example. Probe: instructionBullets gives answered=false, and linkBullet appends a second → [answer] to the operator''s sub-item. Fix: redefine linked so lines added after the link don''t unlink it (in both the skill and the regex), and add a tl-lead test that parses the chat example''s Operator-instructions section.'
          family: spec-contradicts-own-example
          round: 3
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: b5b9ed1f-8e13-41cd-a135-19b40c710dfd
            repository: github.com/xianxu/ariadne
            issue_id: "000412"
            recipe: milestone-review
            round: 4
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/ef3323cd3bb5078b
            reviewed_head: 6fec11e686a4fa6b3ac37980884247ea46258ec8
            slot: ariadne:2
            requested_at: "2026-10-17T00:10:19.464592Z"
          phase: finished
          pid: 4003
          started_at: "2026-10-17T00:10:19.761703Z"
          finished_at: "2026-10-17T00:11:30.2375Z"
          session_id: 6da505a9-d2e3-43ce-a7aa-58fc98f299b4
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 6fec11e686a4fa6b3ac37980884247ea46258ec8
        tracker_head: 27a568ffc9e65d519a006f41cc07fbe5af3f0e35
        branch_point: 35ba24e6cad39344c347af981de7ac8fcc1d1c31
        since: "2026-10-16T15:42:55-07:00"
        until: "2026-10-16T17:09:07-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T15:42:55-07:00"
              end: "2026-10-16T17:09:07-07:00"
              began: "2026-10-16T15:42:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 6681a7b6-a1e9-4bf6-8743-64ae382a6945
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: a01d2d2e30cdee9e2ab0f79bd4c475b0767f8fb71bde51c321f0d54701105ed7
              started: "2026-10-16T22:49:11.289Z"
              spawned: false
              bytes: 542081
              windows:
                - start: "2026-10-16T22:49:11.289Z"
                  end: "2026-10-16T17:09:07-07:00"
            - harness: claude
              id: a3c26c68-b277-41fb-8571-456fa2883740
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5320c188df9f1d072a2c1e3788ff32b5cea4f732e8e180b077700b714d95ba8f
              started: "2026-10-16T22:55:51.076Z"
              spawned: false
              bytes: 509821
              windows:
                - start: "2026-10-16T22:55:51.076Z"
                  end: "2026-10-16T17:09:07-07:00"
            - harness: claude
              id: c16f9f99-f7dc-493c-80e6-9950883fe1cf
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 56c17381f2e60fc56f2ae5acb83b94366afd3a2dea497b7c768fec2238d6894b
              started: "2026-10-16T22:53:46.176Z"
              spawned: false
              bytes: 499558
              windows:
                - start: "2026-10-16T22:53:46.176Z"
                  end: "2026-10-16T17:09:07-07:00"
            - harness: claude
              id: e1a45820-de66-478e-bdce-95f833a64410
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 84b46a5b1ffbb97690e35548a8500ac62e056483b444935a24569d2a8335bc77
              started: "2026-10-16T22:56:14.41Z"
              spawned: false
              bytes: 394689
              windows:
                - start: "2026-10-16T22:56:14.41Z"
                  end: "2026-10-16T17:09:07-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: e73fe7ce2246f4301e773084f4f888b1972cfbdcf3da44083e9cc01a07036738
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 17667127
              windows:
                - start: "2026-10-16T15:42:55-07:00"
                  end: "2026-10-16T17:09:07-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 6.5026
        cuts:
            - dispatched: "2026-10-16T16:55:50-07:00"
              finalized: "2026-10-16T16:57:45-07:00"
              finalizing: false
            - dispatched: "2026-10-16T17:01:22-07:00"
              finalized: "2026-10-16T17:02:49-07:00"
              finalizing: false
            - dispatched: "2026-10-16T17:04:39-07:00"
              finalized: "2026-10-16T17:06:58-07:00"
              finalizing: false
      "n": 4
      timestamp: "2026-10-16T17:11:58-07:00"
      dispatched: "2026-10-16T17:10:19-07:00"
      agent: claude
      dispose:
        - id: BR-6
          disposition: addressed
          note: instructions.go:94-98 + collaboration-modes.md:124-126 now link on any non-fenced line; reverting the parser makes TestInstructionBulletsAreChatItems and TestInstructionBulletsReadTheChatExample fail (open 1-3 / after item 3 open); restored, suite green.
          round: 4
      findings:
        - kind: defect
          id: BR-7
          severity: Minor
          title: Stale test comment still states the old last-line rule
          detail: cmd/tl-lead/instructions_test.go:84-85 still says only a link ending the bullet's last line answers it, directly above the new comment saying any line does. Delete the stale two lines.
          family: cited-rule-restated-divergently
          round: 4
      recipe: milestone-review
      isolation: fenced
      reviewed: 6fec11e686a4fa6b3ac37980884247ea46258ec8
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: bdb93b75-3737-460d-b77f-51b89ddf2c61
            repository: github.com/xianxu/ariadne
            issue_id: "000412"
            recipe: fix-review
            round: 5
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/61cec1278371b358
            reviewed_head: d82914a9f3bf3ef2dc7b5d4147d4f17b3f10b130
            slot: ariadne:2
            requested_at: "2026-10-17T00:13:13.46521Z"
          phase: finished
          pid: 60498
          started_at: "2026-10-17T00:13:13.656123Z"
          finished_at: "2026-10-17T00:13:55.922418Z"
          session_id: 8d3592af-4019-46c2-8025-442b966a1c78
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: d82914a9f3bf3ef2dc7b5d4147d4f17b3f10b130
        tracker_head: a6e93c5ed0b4e9451cd9a5e83e30508f2a5b9484
        branch_point: 35ba24e6cad39344c347af981de7ac8fcc1d1c31
        since: "2026-10-16T15:42:55-07:00"
        until: "2026-10-16T17:12:29-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T15:42:55-07:00"
              end: "2026-10-16T17:12:29-07:00"
              began: "2026-10-16T15:42:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 6681a7b6-a1e9-4bf6-8743-64ae382a6945
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: a01d2d2e30cdee9e2ab0f79bd4c475b0767f8fb71bde51c321f0d54701105ed7
              started: "2026-10-16T22:49:11.289Z"
              spawned: false
              bytes: 542081
              windows:
                - start: "2026-10-16T22:49:11.289Z"
                  end: "2026-10-16T17:12:29-07:00"
            - harness: claude
              id: a3c26c68-b277-41fb-8571-456fa2883740
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5320c188df9f1d072a2c1e3788ff32b5cea4f732e8e180b077700b714d95ba8f
              started: "2026-10-16T22:55:51.076Z"
              spawned: false
              bytes: 509821
              windows:
                - start: "2026-10-16T22:55:51.076Z"
                  end: "2026-10-16T17:12:29-07:00"
            - harness: claude
              id: c16f9f99-f7dc-493c-80e6-9950883fe1cf
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 56c17381f2e60fc56f2ae5acb83b94366afd3a2dea497b7c768fec2238d6894b
              started: "2026-10-16T22:53:46.176Z"
              spawned: false
              bytes: 499558
              windows:
                - start: "2026-10-16T22:53:46.176Z"
                  end: "2026-10-16T17:12:29-07:00"
            - harness: claude
              id: e1a45820-de66-478e-bdce-95f833a64410
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 84b46a5b1ffbb97690e35548a8500ac62e056483b444935a24569d2a8335bc77
              started: "2026-10-16T22:56:14.41Z"
              spawned: false
              bytes: 394689
              windows:
                - start: "2026-10-16T22:56:14.41Z"
                  end: "2026-10-16T17:12:29-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 48923e512facd0482c5cc7964fd7ba8297c54da4525dcb7d0736692564fb4bd8
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 17756812
              windows:
                - start: "2026-10-16T15:42:55-07:00"
                  end: "2026-10-16T17:12:29-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 6.5026
        cuts:
            - dispatched: "2026-10-16T16:55:50-07:00"
              finalized: "2026-10-16T16:57:45-07:00"
              finalizing: false
            - dispatched: "2026-10-16T17:01:22-07:00"
              finalized: "2026-10-16T17:02:49-07:00"
              finalizing: false
            - dispatched: "2026-10-16T17:04:39-07:00"
              finalized: "2026-10-16T17:06:58-07:00"
              finalizing: false
            - dispatched: "2026-10-16T17:10:19-07:00"
              finalized: "2026-10-16T17:11:58-07:00"
              finalizing: true
      "n": 5
      timestamp: "2026-10-16T17:14:00-07:00"
      dispatched: "2026-10-16T17:13:13-07:00"
      agent: claude
      dispose:
        - id: BR-7
          disposition: addressed
          note: instructions_test.go:84-86 now holds one comment stating the any-line rule and the fence exclusion; the stale last-line lines are deleted (comment-only, no test needed).
          round: 5
      findings:
        - kind: defect
          id: BR-8
          severity: Important
          title: scope changed after close — run a full `sdlc close`
          detail: d82914a9 changes collaboration-modes.md:108-109 (exported skill prompt) so a bare acknowledgement also gets "no reply prose", and rewrites the discussion-decision example reply to match. No prior finding asked for this; the source is a post-close TL confirmation (2c4faa9e). The plan's Revisions (plan:286-287) still says the "one line" source wording was followed over the TL's "no reply at all", so the plan now contradicts the code. Run a full close with a Revisions entry for 2c4faa9e, or revert this part and ship the BR-7 fix alone.
          family: scope-changed-after-close
          round: 5
      recipe: fix-review
      isolation: fenced
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: aa34c238-5b2a-4830-9a91-8fadc3159ec7
            repository: github.com/xianxu/ariadne
            issue_id: "000412"
            recipe: milestone-review
            round: 6
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/e13c377cc2b07922
            reviewed_head: 5fdc9f3521d89f638a5ac7bcca822c808f286500
            slot: ariadne:2
            requested_at: "2026-10-17T00:14:31.405046Z"
          phase: finished
          pid: 82064
          started_at: "2026-10-17T00:14:31.502633Z"
          finished_at: "2026-10-17T00:15:21.094604Z"
          session_id: 277c36a8-6b08-496c-9684-777e9996bc99
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 5fdc9f3521d89f638a5ac7bcca822c808f286500
        tracker_head: 15644a3923c6fa87ab6eb5e4a439b5372d703f9a
        branch_point: 35ba24e6cad39344c347af981de7ac8fcc1d1c31
        since: "2026-10-16T15:42:55-07:00"
        until: "2026-10-16T17:14:08-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T15:42:55-07:00"
              end: "2026-10-16T17:14:08-07:00"
              began: "2026-10-16T15:42:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 6681a7b6-a1e9-4bf6-8743-64ae382a6945
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: a01d2d2e30cdee9e2ab0f79bd4c475b0767f8fb71bde51c321f0d54701105ed7
              started: "2026-10-16T22:49:11.289Z"
              spawned: false
              bytes: 542081
              windows:
                - start: "2026-10-16T22:49:11.289Z"
                  end: "2026-10-16T17:14:08-07:00"
            - harness: claude
              id: a3c26c68-b277-41fb-8571-456fa2883740
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5320c188df9f1d072a2c1e3788ff32b5cea4f732e8e180b077700b714d95ba8f
              started: "2026-10-16T22:55:51.076Z"
              spawned: false
              bytes: 509821
              windows:
                - start: "2026-10-16T22:55:51.076Z"
                  end: "2026-10-16T17:14:08-07:00"
            - harness: claude
              id: c16f9f99-f7dc-493c-80e6-9950883fe1cf
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 56c17381f2e60fc56f2ae5acb83b94366afd3a2dea497b7c768fec2238d6894b
              started: "2026-10-16T22:53:46.176Z"
              spawned: false
              bytes: 499558
              windows:
                - start: "2026-10-16T22:53:46.176Z"
                  end: "2026-10-16T17:14:08-07:00"
            - harness: claude
              id: e1a45820-de66-478e-bdce-95f833a64410
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 84b46a5b1ffbb97690e35548a8500ac62e056483b444935a24569d2a8335bc77
              started: "2026-10-16T22:56:14.41Z"
              spawned: false
              bytes: 394689
              windows:
                - start: "2026-10-16T22:56:14.41Z"
                  end: "2026-10-16T17:14:08-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5b524e2c28ee18464a146f8ce3694e9ad590406a7ff2a74c24e87bbabdfe7d59
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 17778627
              windows:
                - start: "2026-10-16T15:42:55-07:00"
                  end: "2026-10-16T17:14:08-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 6.5026
        cuts:
            - dispatched: "2026-10-16T16:55:50-07:00"
              finalized: "2026-10-16T16:57:45-07:00"
              finalizing: false
            - dispatched: "2026-10-16T17:01:22-07:00"
              finalized: "2026-10-16T17:02:49-07:00"
              finalizing: false
            - dispatched: "2026-10-16T17:04:39-07:00"
              finalized: "2026-10-16T17:06:58-07:00"
              finalizing: false
            - dispatched: "2026-10-16T17:10:19-07:00"
              finalized: "2026-10-16T17:11:58-07:00"
              finalizing: true
            - dispatched: "2026-10-16T17:13:13-07:00"
              finalized: "2026-10-16T17:14:00-07:00"
              finalizing: false
      "n": 6
      timestamp: "2026-10-16T17:15:21-07:00"
      dispatched: "2026-10-16T17:14:31-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: addressed
          note: 5fdc9f35 adds the Revisions entry for 2c4faa9e (plan:292-297), superseding the earlier "approved source is followed" entry; collaboration-modes.md:108-109 and the discussion-decision example match it; this review covers the full window as the full close.
          round: 6
      recipe: milestone-review
      isolation: fenced
      reviewed: 5fdc9f3521d89f638a5ac7bcca822c808f286500
      blocked: false
---

# Gate ledger — ariadne#412 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-16T16:57:45-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `reuse-identity-too-coarse` (kind: defect) Side-doc reuse by (parent, marker) alone can link a new ask to an unrelated, already-linked discussion
  collaboration-modes.md §4 step 1 reuses any side doc whose parent and marker match, even one already linked from another section, so a second identical marker (e.g. 🤖[why?] elsewhere) silently gets the old discussion; same-round identical markers are unspecified. Restrict reuse to unlinked docs (and same section for discussion) and add a case or check.
- **BR-2** [Important] `cited-rule-restated-divergently` (kind: defect) tl-lead's instructionBullets defines an item differently from the chat mode tl-lead --help now cites as the rule
  cmd/tl-lead/instructions.go only starts items at unindented `N. ` lines, treats `→ done:` as answered, and absorbs `-`/`*` lines into the previous bullet, while collaboration-modes.md §3 counts -, * and N. items at the shallowest indent; a `- foo` bullet is invisible to list and gets the previous bullet's link. Align the parser (with a table test) or narrow the help.
- **BR-3** [Minor] `done-when-deferred` (kind: defect) Done-when 2 (session file runs on collaboration-mode) depends on the TL's ops write; record it as owed in the landing report
- **BR-4** [Minor] `checker-coverage-gap` (kind: defect) Examples checker does not assert the (folded <date>) link mark for fold cases (rows 7 and 10)

## Round 2 — 2026-10-16T17:02:49-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Spec §4 step 1 and row 19 now require the same section and an unlinked doc; case discussion-same-marker added. Checker enforcement gap raised separately as checker-coverage-gap.
- BR-2 — addressed — instructionBullets now follows the §3 item rule; TestInstructionBulletsAreChatItems fails on the old N.-only parser.
- BR-3 — addressed — Plan revisions (plan line ~264) record Done-when 2 as owed by the TL in the landing report.
- BR-4 — addressed — foldedRE check added, with a test that breaks a fixture's fold mark and expects the checker to fail.

### Raised

- **BR-5** [Important] `checker-coverage-gap` (kind: defect) Example checker does not enforce row 19: a side doc that is already linked can be reused with no failure
  I edited discussion-same-marker/after so both asks link to 01-why-per-file.md and deleted 02, which is the BR-1 defect, and TestExamples still passed. This is the 2nd finding in the family. The rule covering the class: in a parent, each side doc is the target of at most one link. Enforce it in checkTree and add a test that links two asks to one doc and expects a failure.

## Round 3 — 2026-10-16T17:06:58-07:00 (claude) — BLOCKED

### Disposed

- BR-5 — addressed — Reviewer's regression (both asks -> 01, 02 deleted) now fails TestExamples; disabling the targets check turns "a linked doc reused" red.

### Raised

- **BR-6** [Critical] `spec-contradicts-own-example` (kind: defect) "Linked = answer link ending the item's last line" makes a row-14 follow-up item read as unlinked; tl-lead double-links it
  e123c3ca changed collaboration-modes.md:115-116 and instructions.go answeredRE (line 21, used at 96) so only a link on the last line counts. examples/chat item 3 ("3. ... → [answer](...)" then "   - and by repo?") is then a new ask by the spec's own rule, which contradicts row 14 and the example. Probe: instructionBullets gives answered=false, and linkBullet appends a second → [answer] to the operator's sub-item. Fix: redefine linked so lines added after the link don't unlink it (in both the skill and the regex), and add a tl-lead test that parses the chat example's Operator-instructions section.

## Round 4 — 2026-10-16T17:11:58-07:00 (claude) — passed

### Disposed

- BR-6 — addressed — instructions.go:94-98 + collaboration-modes.md:124-126 now link on any non-fenced line; reverting the parser makes TestInstructionBulletsAreChatItems and TestInstructionBulletsReadTheChatExample fail (open 1-3 / after item 3 open); restored, suite green.

### Raised

- **BR-7** [Minor] `cited-rule-restated-divergently` (kind: defect) Stale test comment still states the old last-line rule
  cmd/tl-lead/instructions_test.go:84-85 still says only a link ending the bullet's last line answers it, directly above the new comment saying any line does. Delete the stale two lines.

## Round 5 — 2026-10-16T17:14:00-07:00 (claude) — passed

### Disposed

- BR-7 — addressed — instructions_test.go:84-86 now holds one comment stating the any-line rule and the fence exclusion; the stale last-line lines are deleted (comment-only, no test needed).

### Raised

- **BR-8** [Important] `scope-changed-after-close` (kind: defect) scope changed after close — run a full `sdlc close`
  d82914a9 changes collaboration-modes.md:108-109 (exported skill prompt) so a bare acknowledgement also gets "no reply prose", and rewrites the discussion-decision example reply to match. No prior finding asked for this; the source is a post-close TL confirmation (2c4faa9e). The plan's Revisions (plan:286-287) still says the "one line" source wording was followed over the TL's "no reply at all", so the plan now contradicts the code. Run a full close with a Revisions entry for 2c4faa9e, or revert this part and ship the BR-7 fix alone.

## Round 6 — 2026-10-16T17:15:21-07:00 (claude) — passed

### Disposed

- BR-8 — addressed — 5fdc9f35 adds the Revisions entry for 2c4faa9e (plan:292-297), superseding the earlier "approved source is followed" entry; collaboration-modes.md:108-109 and the discussion-decision example match it; this review covers the full window as the full close.

## Open findings

(none — every finding has been disposed)
