---
gate: boundary-review
issue: 458
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 8d765e44-9a65-4e75-91b9-905d30a95b5e
            repository: github.com/xianxu/ariadne
            issue_id: "000458"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/28a4f2053ee07160
            reviewed_head: 1b56f757d2b7852f06c162fc2b39a97c0078eb14
            slot: ariadne:1
            requested_at: "2026-10-19T12:25:12.330439Z"
          phase: finished
          pid: 47557
          started_at: "2026-10-19T12:25:12.427098Z"
          finished_at: "2026-10-19T12:27:07.102181Z"
          session_id: 01a1541f-7f0e-7d72-a761-660c9890a1fd
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 1b56f757d2b7852f06c162fc2b39a97c0078eb14
        tracker_head: d45848d988ab04490cf0d9bfce88a1eaed8da451
        branch_point: a11504c417e9c8789327362895bc88c9fad042df
        since: "2026-10-19T04:32:38-07:00"
        until: "2026-10-19T04:55:16-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T04:32:38-07:00"
              end: "2026-10-19T04:55:16-07:00"
              began: "2026-10-19T04:32:38-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: ee47cbf5e7b1d5e81968b86e16a1827fbcd2c412de47c79251443eb97bb5cf6a
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 141839773
              windows:
                - start: "2026-10-19T04:32:38-07:00"
                  end: "2026-10-19T04:55:16-07:00"
      "n": 1
      timestamp: "2026-10-19T05:27:07-07:00"
      dispatched: "2026-10-19T05:25:12-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: BR-1
          severity: Critical
          title: 'ARCH-PURE: the declared pure decision performs injected Git IO'
          detail: cmd/sdlc/flake_fixed.go:219–223 invokes contains, wired to real Git in production and in TestFixedFlakeAgainstRealGitMergeHistory. The plan’s Core concepts table at lines 19–22 instead declares a pure FixedFlakeDecision taking ancestry answers; neither that symbol nor FixedFlakeIssue exists under the declared names. Move ancestry execution into the integration shell and pass explicit results into a pure decision, or revise the table to name the actual entities and classify the orchestration as INTEGRATION.
          family: pure-core-contract
          round: 1
        - id: BR-2
          severity: Important
          title: README update appears missing for flake list -v
          detail: cmd/sdlc/flake_cmd.go:261 introduces --verbose/-v and changes default list behavior, but README.md:525 still describes only the previous daily view. Update that passage with default suppression, verbose attribution, and recurrence behavior; atlas coverage does not replace README usage.
          family: user-surface-documentation
          round: 1
        - id: BR-3
          severity: Important
          title: Command tests do not pin default suppression or recurrence rendering
          detail: cmd/sdlc/flake_fixed_test.go:83 tests the classifier against real Git, while lines 124–147 exercise only verbose fixed-row rendering with synthetic ancestry answers. No test asserts that the default command hides a proven fixed row or that a recurrence stays visible with its annotation. Add command-level tests covering both modes and conservative failure cases, using the real-history seam required by Done when.
          family: observable-behavior-regression-coverage
          round: 1
        - id: BR-4
          severity: Important
          title: Per-row evidence failures discard the promised coverage reason
          detail: cmd/sdlc/flake_fixed.go:203–225 collapses unreadable details, ambiguous ownership, and ancestry errors into flakeUnknown without a reason. cmd/sdlc/flake_cmd.go:487 prints those rows unchanged; its coverage message handles only catalogue-wide FixedErr. Preserve a reason in the decision result and render it so these cases satisfy the plan’s line-34 promise that unknown rows remain visible with a coverage reason.
          family: uncertainty-remains-observable
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 09f5f5d6-fdf5-4217-9b31-b3337ecf0308
            repository: github.com/xianxu/ariadne
            issue_id: "000458"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/feb76cf032713ca5
            reviewed_head: e4e8590d273723c84d7cc000c381abc16d622cdd
            slot: ariadne:1
            requested_at: "2026-10-19T12:59:54.84211Z"
          phase: finished
          pid: 84333
          started_at: "2026-10-19T12:59:54.936514Z"
          finished_at: "2026-10-19T13:02:41.515882Z"
          session_id: 01a1543f-45bc-7b93-b21d-fabbef7cb8e7
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: e4e8590d273723c84d7cc000c381abc16d622cdd
        tracker_head: d45848d988ab04490cf0d9bfce88a1eaed8da451
        branch_point: a11504c417e9c8789327362895bc88c9fad042df
        since: "2026-10-19T04:32:38-07:00"
        until: "2026-10-19T05:32:44-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T04:32:38-07:00"
              end: "2026-10-19T05:32:44-07:00"
              began: "2026-10-19T04:32:38-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 850973e687bb57065e56875c285c9ea04f9d928b2bf2d8347071c2aed9033dd6
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 142709310
              windows:
                - start: "2026-10-19T04:32:38-07:00"
                  end: "2026-10-19T05:32:44-07:00"
        queue_waits:
            - attempt_id: 47932ec61a1603c497832d736497e4f0
              slot: ariadne:1
              record_digest: 26e688dc36fb19ea4a4f2f0ba6b5a0c1a7640426e201e1a37b7e660686969336
              start: "2026-10-19T11:55:43.408394Z"
              end: "2026-10-19T12:12:29.80791Z"
              queue_seconds: 1006.399516
              raw_seconds: 1006.399516
              exact: true
        cuts:
            - dispatched: "2026-10-19T05:25:12-07:00"
              finalized: "2026-10-19T05:27:07-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-19T06:02:42-07:00"
      dispatched: "2026-10-19T05:59:54-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Plan lines 19–23 now name the implemented entities and classify classifyFixedFlake as INTEGRATION, matching its injected Git calls. fixedIssueCovers remains independently testable without IO.
          round: 2
        - id: BR-2
          disposition: addressed
          note: README.md:527–533 documents default suppression, flake list -v, recurrence visibility, and uncertain evidence; these descriptions match flake_cmd.go:474–486.
          round: 2
        - id: BR-3
          disposition: addressed
          note: TestFlakeListRendersFixedRecurrenceAndCoverageFromRealGit exercises default and verbose output, recurrence, and ambiguous coverage using real merge history. Temporary source overlays disabling suppression or recurrence annotation each made it fail.
          round: 2
        - id: BR-4
          disposition: addressed
          note: classifyFixedFlake now returns reasons for unreadable details, ambiguous ownership, missing merges, and ancestry failures. Decision tests assert these reasons; removing coverage-reason rendering made the real-history command test fail.
          round: 2
      findings:
        - kind: defect
          id: BR-5
          severity: Critical
          title: 'ARCH-PURPOSE: package-basename matching can hide an unrelated test'
          detail: cmd/sdlc/flake_fixed.go:171 reduces package identity to path.Base(pkg). For a failure named github.com/xianxu/ariadne/cmd/internal/launcher.TestX, details reading “TestX fails in pkg/launcher/run_test.go” incorrectly establish ownership. A temporary regression test confirmed that classifyFixedFlake returns flakeFixed when the issue’s merge is on main but absent from the failing head, causing default rendering to suppress the unrelated failure. bareCollision does not protect this case when only one package occurs in retained attempts. Require unambiguous package identity for textual ownership, retain uncertain rows, and cover same-basename packages across substantive sections and dedicated Log lines.
          family: exact-test-ownership
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 794ce4b7-8cb3-4e10-aecd-cfbbc364fb12
            repository: github.com/xianxu/ariadne
            issue_id: "000458"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/db5c271d3bf4c2fc
            reviewed_head: f2cbfe70e455169334d78e71e6c820e288089c4a
            slot: ariadne:1
            requested_at: "2026-10-19T13:37:27.892996Z"
          phase: finished
          pid: 73349
          started_at: "2026-10-19T13:37:27.987542Z"
          finished_at: "2026-10-19T13:40:38.649745Z"
          session_id: 01a15461-a6b2-76c2-9b14-55a228670acf
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: f2cbfe70e455169334d78e71e6c820e288089c4a
        tracker_head: d45848d988ab04490cf0d9bfce88a1eaed8da451
        branch_point: a11504c417e9c8789327362895bc88c9fad042df
        since: "2026-10-19T04:32:38-07:00"
        until: "2026-10-19T06:15:37-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T04:32:38-07:00"
              end: "2026-10-19T06:15:37-07:00"
              began: "2026-10-19T04:32:38-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 4113d18aef13062faa3aa1a5e93c73adfb9ee2feec4a5ffc4b4280cd6440e74d
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 143784527
              windows:
                - start: "2026-10-19T04:32:38-07:00"
                  end: "2026-10-19T06:15:37-07:00"
        queue_waits:
            - attempt_id: 47932ec61a1603c497832d736497e4f0
              slot: ariadne:1
              record_digest: 26e688dc36fb19ea4a4f2f0ba6b5a0c1a7640426e201e1a37b7e660686969336
              start: "2026-10-19T11:55:43.408394Z"
              end: "2026-10-19T12:12:29.80791Z"
              queue_seconds: 1006.399516
              raw_seconds: 1006.399516
              exact: true
            - attempt_id: 17efbf21e381b47311820e1181929f73
              slot: ariadne:1
              record_digest: 1c614cbe42ae0eab068ee411b5eff6fd8127b867521fd34d3adbf40255e1abd8
              start: "2026-10-19T12:33:45.540326Z"
              end: "2026-10-19T12:49:43.182913Z"
              queue_seconds: 957.642587
              raw_seconds: 957.642587
              exact: true
        cuts:
            - dispatched: "2026-10-19T05:25:12-07:00"
              finalized: "2026-10-19T05:27:07-07:00"
              finalizing: false
            - dispatched: "2026-10-19T05:59:54-07:00"
              finalized: "2026-10-19T06:02:42-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-19T06:40:39-07:00"
      dispatched: "2026-10-19T06:37:27-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: The Core concepts table correctly declares classifyFixedFlake as INTEGRATION; fixedIssueCovers is tested without IO.
          round: 3
        - id: BR-2
          disposition: addressed
          note: README.md:527–533 documents default suppression, verbose attribution, recurrence, and uncertainty, matching the command implementation.
          round: 3
        - id: BR-3
          disposition: addressed
          note: TestFlakeListRendersFixedRecurrenceAndCoverageFromRealGit verifies default suppression, verbose output, recurrence, and ambiguous coverage through real merge history.
          round: 3
        - id: BR-4
          disposition: addressed
          note: The classifier returns coverage reasons, decision tests assert them, and the command test verifies their rendering.
          round: 3
        - id: BR-5
          disposition: not-addressed
          note: 'ARCH-PURPOSE / ARCH-SECURE: cmd/sdlc/flake_fixed.go:239–243 searches the entire remaining line for ".go", crossing path-token boundaries. For github.com/xianxu/ariadne/cmd/internal/launcher.TestX, "TestX fails in cmd/internal/launcher/subpkg (see run_test.go)." incorrectly establishes parent-package ownership. Temporary regression probes failed in both Problem and dedicated Log sections; a command-level probe confirmed default output silently omitted the unrelated failure. The exact-test-ownership family remains unresolved. Fix the rule across all ownership contexts: recognize one contiguous path token and accept only the exact package directory or a Go source file directly inside it.'
          round: 3
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 84f29edf-e270-41d0-bd76-ac4808ebc6eb
            repository: github.com/xianxu/ariadne
            issue_id: "000458"
            recipe: milestone-review
            round: 4
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/2ef94d604cbe5e10
            reviewed_head: b64abf3422274c046cc283b613ca4a38836d2185
            slot: ariadne:1
            requested_at: "2026-10-19T14:21:34.818874Z"
          phase: finished
          pid: 31485
          started_at: "2026-10-19T14:21:34.92592Z"
          finished_at: "2026-10-19T14:24:08.706413Z"
          session_id: 01a1548a-0a51-7c50-b6e9-44a3278f436d
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: b64abf3422274c046cc283b613ca4a38836d2185
        tracker_head: c859465254df59903a291f0f6ed200ca9dcdafe1
        branch_point: a11504c417e9c8789327362895bc88c9fad042df
        since: "2026-10-19T04:32:38-07:00"
        until: "2026-10-19T06:51:41-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T04:32:38-07:00"
              end: "2026-10-19T06:51:41-07:00"
              began: "2026-10-19T04:32:38-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 32de734f0473a8b87a5ce1e5c6765376a6960cd665db57c80f75238ad6bc57bc
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 145510423
              windows:
                - start: "2026-10-19T04:32:38-07:00"
                  end: "2026-10-19T06:51:41-07:00"
        queue_waits:
            - attempt_id: 47932ec61a1603c497832d736497e4f0
              slot: ariadne:1
              record_digest: 26e688dc36fb19ea4a4f2f0ba6b5a0c1a7640426e201e1a37b7e660686969336
              start: "2026-10-19T11:55:43.408394Z"
              end: "2026-10-19T12:12:29.80791Z"
              queue_seconds: 1006.399516
              raw_seconds: 1006.399516
              exact: true
            - attempt_id: 17efbf21e381b47311820e1181929f73
              slot: ariadne:1
              record_digest: 1c614cbe42ae0eab068ee411b5eff6fd8127b867521fd34d3adbf40255e1abd8
              start: "2026-10-19T12:33:45.540326Z"
              end: "2026-10-19T12:49:43.182913Z"
              queue_seconds: 957.642587
              raw_seconds: 957.642587
              exact: true
            - attempt_id: 4cf8897332b679b50a8739bdc501b2d0
              slot: ariadne:1
              record_digest: e9ebf530ab8d199c4abc3a799d3961adb3b0a6040d2ec4d8522939c29ac3b74e
              start: "2026-10-19T13:15:52.502481Z"
              end: "2026-10-19T13:27:04.382509Z"
              queue_seconds: 671.880028
              raw_seconds: 671.880028
              exact: true
        cuts:
            - dispatched: "2026-10-19T05:25:12-07:00"
              finalized: "2026-10-19T05:27:07-07:00"
              finalizing: false
            - dispatched: "2026-10-19T05:59:54-07:00"
              finalized: "2026-10-19T06:02:42-07:00"
              finalizing: false
            - dispatched: "2026-10-19T06:37:27-07:00"
              finalized: "2026-10-19T06:40:39-07:00"
              finalizing: false
      "n": 4
      timestamp: "2026-10-19T07:24:10-07:00"
      dispatched: "2026-10-19T07:21:34-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: The plan correctly classifies classifyFixedFlake as INTEGRATION; fixedIssueCovers remains independently tested without IO.
          round: 4
        - id: BR-2
          disposition: addressed
          note: README.md:527 documents default suppression, verbose attribution, recurrence and uncertainty, matching flake_cmd.go.
          round: 4
        - id: BR-3
          disposition: addressed
          note: TestFlakeListRendersFixedRecurrenceAndCoverageFromRealGit verifies default suppression, verbose attribution, recurrence and ambiguous coverage using real Git history.
          round: 4
        - id: BR-4
          disposition: addressed
          note: The classifier preserves uncertainty reasons; decision tests assert them and the command regression verifies their rendering.
          round: 4
        - id: BR-5
          disposition: addressed
          note: Package-path heuristics are removed. Restoring the original matcher through a temporary source overlay makes both same-basename regressions and the nested-subpackage regression fail; these pass at the pinned head.
          round: 4
      findings:
        - kind: defect
          id: BR-6
          severity: Critical
          title: 'ARCH-PURPOSE / ARCH-SECURE: Markdown samples can still hide unrelated flakes'
          detail: 'cmd/sdlc/flake_fixed.go:149 uses UnterminatedIsProse, so an unclosed backtick or tilde fence in Log exposes a sample "- Test: <fully qualified name>" as ownership. Lines 162–167 also accept indented code and HTML-comment contents. Temporary matcher and command probes confirmed silent default suppression for all four forms; closed backtick and tilde samples correctly stayed visible. This is the 2nd finding in family exact-test-ownership. Earlier rounds fixed instances. State and enforce the rule across actual Log entries, closed/unclosed fences, indented code and comments: only an actual structured ownership entry may claim a test; ambiguous or truncated markup must preserve visibility. Add regression tests for that enumeration, including command output.'
          family: exact-test-ownership
          round: 4
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 6e80fea2-0e0b-4b7e-8677-43e8f13cd64f
            repository: github.com/xianxu/ariadne
            issue_id: "000458"
            recipe: milestone-review
            round: 5
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/5dc8aeb36b86856e
            reviewed_head: da8c62d2e111ef7fe86cc0c351e00d6512a5a87a
            slot: ariadne:1
            requested_at: "2026-10-19T15:27:31.331217Z"
          phase: finished
          pid: 4216
          started_at: "2026-10-19T15:27:31.501412Z"
          finished_at: "2026-10-19T15:30:58.177571Z"
          session_id: 01a154c6-6a26-7b62-bd87-13eb20e58073
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: da8c62d2e111ef7fe86cc0c351e00d6512a5a87a
        tracker_head: 3adcfba4745aa1327f4d85e699615d15b16d7019
        branch_point: a11504c417e9c8789327362895bc88c9fad042df
        since: "2026-10-19T04:32:38-07:00"
        until: "2026-10-19T07:26:37-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T04:32:38-07:00"
              end: "2026-10-19T07:26:37-07:00"
              began: "2026-10-19T04:32:38-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: c47b106663c4a90270281978683de1fd4a8482aa54ab84e5d8148f2ad72c9379
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 146491952
              windows:
                - start: "2026-10-19T04:32:38-07:00"
                  end: "2026-10-19T07:26:37-07:00"
        queue_waits:
            - attempt_id: 47932ec61a1603c497832d736497e4f0
              slot: ariadne:1
              record_digest: 26e688dc36fb19ea4a4f2f0ba6b5a0c1a7640426e201e1a37b7e660686969336
              start: "2026-10-19T11:55:43.408394Z"
              end: "2026-10-19T12:12:29.80791Z"
              queue_seconds: 1006.399516
              raw_seconds: 1006.399516
              exact: true
            - attempt_id: 17efbf21e381b47311820e1181929f73
              slot: ariadne:1
              record_digest: 1c614cbe42ae0eab068ee411b5eff6fd8127b867521fd34d3adbf40255e1abd8
              start: "2026-10-19T12:33:45.540326Z"
              end: "2026-10-19T12:49:43.182913Z"
              queue_seconds: 957.642587
              raw_seconds: 957.642587
              exact: true
            - attempt_id: 4cf8897332b679b50a8739bdc501b2d0
              slot: ariadne:1
              record_digest: e9ebf530ab8d199c4abc3a799d3961adb3b0a6040d2ec4d8522939c29ac3b74e
              start: "2026-10-19T13:15:52.502481Z"
              end: "2026-10-19T13:27:04.382509Z"
              queue_seconds: 671.880028
              raw_seconds: 671.880028
              exact: true
            - attempt_id: 055d69102c5a0f9423aa60d99ab8eca7
              slot: ariadne:1
              record_digest: ea2911473d9cecab206de465a25f9cf514eaa748b147c89bb1662dcfef8b29d5
              start: "2026-10-19T13:52:00.770977Z"
              end: "2026-10-19T14:09:56.4847Z"
              queue_seconds: 1075.713723
              raw_seconds: 1075.713723
              exact: true
        cuts:
            - dispatched: "2026-10-19T05:25:12-07:00"
              finalized: "2026-10-19T05:27:07-07:00"
              finalizing: false
            - dispatched: "2026-10-19T05:59:54-07:00"
              finalized: "2026-10-19T06:02:42-07:00"
              finalizing: false
            - dispatched: "2026-10-19T06:37:27-07:00"
              finalized: "2026-10-19T06:40:39-07:00"
              finalizing: false
            - dispatched: "2026-10-19T07:21:34-07:00"
              finalized: "2026-10-19T07:24:10-07:00"
              finalizing: false
      "n": 5
      timestamp: "2026-10-19T08:30:58-07:00"
      dispatched: "2026-10-19T08:27:31-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: The Core concepts table correctly classifies classifyFixedFlake as INTEGRATION and fixedIssueCovers as PURE; the matcher tests perform no IO.
          round: 5
        - id: BR-2
          disposition: addressed
          note: README.md:527 documents default suppression, structured ownership, verbose output, recurrence, and uncertainty, matching the implemented command.
          round: 5
        - id: BR-3
          disposition: addressed
          note: TestFlakeListRendersFixedRecurrenceAndCoverageFromRealGit exercises default suppression, verbose fixed-row rendering, and visible recurrence through a real merge history; it passes at the pinned head.
          round: 5
        - id: BR-4
          disposition: addressed
          note: classifyFixedFlake returns specific uncertainty reasons, runFlakeListMode renders them, and the decision matrix plus command-level ambiguous-coverage assertion pass.
          round: 5
        - id: BR-5
          disposition: addressed
          note: Ownership no longer uses package-basename or free-prose inference. Exact card values and dedicated Log entries are tested against wrong-package, nested-package, and substring cases.
          round: 5
        - id: BR-6
          disposition: not-addressed
          note: 'Critical, ARCH-PURPOSE / ARCH-SECURE, family exact-test-ownership: cmd/sdlc/flake_fixed.go:153-191 still accepts a sample inside "## Log\n<pre>\n- Test: <fully qualified name>\n</pre>\n", including an unclosed <pre>. A disposable command regression using the existing real-Git fixture confirmed that this sample silently removes the pre-fix row. Line 191 also accepts an unmatched opening backtick because strings.Trim removes arbitrary backticks. The reported fence, indentation, and comment fixes have genuine red/green regression evidence, but the rule that only actual, unambiguous structured entries establish ownership remains unenforced. Define and enforce the complete accepted Markdown entry grammar, rejecting raw HTML sample contexts and malformed delimiters, and cover it through matcher and command tests rather than adding another isolated exclusion.'
          round: 5
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 5fbd40cd-9746-4015-844f-5afd9aa85965
            repository: github.com/xianxu/ariadne
            issue_id: "000458"
            recipe: milestone-review
            round: 6
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/37e011fbbb6207e8
            reviewed_head: c22cc76feaaa584819a8d38ceac243ae6cd2d4f5
            slot: ariadne:1
            requested_at: "2026-10-19T15:43:48.500583Z"
          phase: finished
          pid: 47530
          started_at: "2026-10-19T15:43:48.754105Z"
          finished_at: "2026-10-19T15:47:00.407961Z"
          session_id: 01a154d5-543d-7840-87da-2224b5abaff4
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: c22cc76feaaa584819a8d38ceac243ae6cd2d4f5
        tracker_head: e739b70377206e8b595c687e2723402ed135e043
        branch_point: a11504c417e9c8789327362895bc88c9fad042df
        since: "2026-10-19T04:32:38-07:00"
        until: "2026-10-19T08:37:18-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T04:32:38-07:00"
              end: "2026-10-19T08:37:18-07:00"
              began: "2026-10-19T04:32:38-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 56e793742bddf2fce7ab312a47fe6208ab2b346ea5279bd92ad7fe2dfe6ab48f
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 147702402
              windows:
                - start: "2026-10-19T04:32:38-07:00"
                  end: "2026-10-19T08:37:18-07:00"
        queue_waits:
            - attempt_id: 47932ec61a1603c497832d736497e4f0
              slot: ariadne:1
              record_digest: 26e688dc36fb19ea4a4f2f0ba6b5a0c1a7640426e201e1a37b7e660686969336
              start: "2026-10-19T11:55:43.408394Z"
              end: "2026-10-19T12:12:29.80791Z"
              queue_seconds: 1006.399516
              raw_seconds: 1006.399516
              exact: true
            - attempt_id: 17efbf21e381b47311820e1181929f73
              slot: ariadne:1
              record_digest: 1c614cbe42ae0eab068ee411b5eff6fd8127b867521fd34d3adbf40255e1abd8
              start: "2026-10-19T12:33:45.540326Z"
              end: "2026-10-19T12:49:43.182913Z"
              queue_seconds: 957.642587
              raw_seconds: 957.642587
              exact: true
            - attempt_id: 4cf8897332b679b50a8739bdc501b2d0
              slot: ariadne:1
              record_digest: e9ebf530ab8d199c4abc3a799d3961adb3b0a6040d2ec4d8522939c29ac3b74e
              start: "2026-10-19T13:15:52.502481Z"
              end: "2026-10-19T13:27:04.382509Z"
              queue_seconds: 671.880028
              raw_seconds: 671.880028
              exact: true
            - attempt_id: 055d69102c5a0f9423aa60d99ab8eca7
              slot: ariadne:1
              record_digest: ea2911473d9cecab206de465a25f9cf514eaa748b147c89bb1662dcfef8b29d5
              start: "2026-10-19T13:52:00.770977Z"
              end: "2026-10-19T14:09:56.4847Z"
              queue_seconds: 1075.713723
              raw_seconds: 1075.713723
              exact: true
            - attempt_id: 4c96305468ac871ff567625384059cc8
              slot: ariadne:1
              record_digest: 6a38354afbf41a9d5cd335fdcec1bd2cfc0d5f890e65a6777bef7d55e5c0119d
              start: "2026-10-19T14:26:54.353487Z"
              end: "2026-10-19T14:39:52.93447Z"
              queue_seconds: 778.580983
              raw_seconds: 778.580983
              exact: true
            - attempt_id: 435cea42b0ba20eba112f18ae3d6d2c1
              slot: ariadne:1
              record_digest: 6d89a987475a366c700e6abc49ae304be85c1b9b78c7bb98b337d8c02f0851a9
              start: "2026-10-19T14:57:12.251845Z"
              end: "2026-10-19T15:14:44.103366Z"
              queue_seconds: 1051.851521
              raw_seconds: 1051.851521
              exact: true
        cuts:
            - dispatched: "2026-10-19T05:25:12-07:00"
              finalized: "2026-10-19T05:27:07-07:00"
              finalizing: false
            - dispatched: "2026-10-19T05:59:54-07:00"
              finalized: "2026-10-19T06:02:42-07:00"
              finalizing: false
            - dispatched: "2026-10-19T06:37:27-07:00"
              finalized: "2026-10-19T06:40:39-07:00"
              finalizing: false
            - dispatched: "2026-10-19T07:21:34-07:00"
              finalized: "2026-10-19T07:24:10-07:00"
              finalizing: false
            - dispatched: "2026-10-19T08:27:31-07:00"
              finalized: "2026-10-19T08:30:58-07:00"
              finalizing: false
      "n": 6
      timestamp: "2026-10-19T08:47:01-07:00"
      dispatched: "2026-10-19T08:43:48-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: The Core concepts table correctly declares classifyFixedFlake INTEGRATION; fixedIssueCovers performs only string comparisons without IO.
          round: 6
        - id: BR-2
          disposition: addressed
          note: README.md:527 documents default suppression, flake list -v and recurrence, consistent with cmd/sdlc/flake_cmd.go rendering.
          round: 6
        - id: BR-3
          disposition: addressed
          note: TestFlakeListRendersFixedRecurrenceAndCoverageFromRealGit passes and asserts default suppression, verbose fixed rows, recurrence and ambiguous coverage.
          round: 6
        - id: BR-4
          disposition: addressed
          note: Classifier uncertainty reasons reach command output; passing decision and rendering tests verify missing merges, Git uncertainty and ambiguous ownership.
          round: 6
        - id: BR-5
          disposition: addressed
          note: Ownership now uses exact card-list equality; package-basename and textual-path inference have been removed.
          round: 6
        - id: BR-6
          disposition: addressed
          note: 'Published details no longer enter ownership evidence. TestFlakeListIgnoresPublishedLogOwnership passes on head; restoring the prior implementation through a temporary Go overlay makes all three cases fail: plain Log, raw HTML and unmatched backticks.'
          round: 6
      recipe: milestone-review
      isolation: fenced
      reviewed: c22cc76feaaa584819a8d38ceac243ae6cd2d4f5
      blocked: false
---

# Gate ledger — ariadne#458 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T05:27:07-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `pure-core-contract` ARCH-PURE: the declared pure decision performs injected Git IO
  cmd/sdlc/flake_fixed.go:219–223 invokes contains, wired to real Git in production and in TestFixedFlakeAgainstRealGitMergeHistory. The plan’s Core concepts table at lines 19–22 instead declares a pure FixedFlakeDecision taking ancestry answers; neither that symbol nor FixedFlakeIssue exists under the declared names. Move ancestry execution into the integration shell and pass explicit results into a pure decision, or revise the table to name the actual entities and classify the orchestration as INTEGRATION.
- **BR-2** [Important] `user-surface-documentation` README update appears missing for flake list -v
  cmd/sdlc/flake_cmd.go:261 introduces --verbose/-v and changes default list behavior, but README.md:525 still describes only the previous daily view. Update that passage with default suppression, verbose attribution, and recurrence behavior; atlas coverage does not replace README usage.
- **BR-3** [Important] `observable-behavior-regression-coverage` Command tests do not pin default suppression or recurrence rendering
  cmd/sdlc/flake_fixed_test.go:83 tests the classifier against real Git, while lines 124–147 exercise only verbose fixed-row rendering with synthetic ancestry answers. No test asserts that the default command hides a proven fixed row or that a recurrence stays visible with its annotation. Add command-level tests covering both modes and conservative failure cases, using the real-history seam required by Done when.
- **BR-4** [Important] `uncertainty-remains-observable` Per-row evidence failures discard the promised coverage reason
  cmd/sdlc/flake_fixed.go:203–225 collapses unreadable details, ambiguous ownership, and ancestry errors into flakeUnknown without a reason. cmd/sdlc/flake_cmd.go:487 prints those rows unchanged; its coverage message handles only catalogue-wide FixedErr. Preserve a reason in the decision result and render it so these cases satisfy the plan’s line-34 promise that unknown rows remain visible with a coverage reason.

## Round 2 — 2026-10-19T06:02:42-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — Plan lines 19–23 now name the implemented entities and classify classifyFixedFlake as INTEGRATION, matching its injected Git calls. fixedIssueCovers remains independently testable without IO.
- BR-2 — addressed — README.md:527–533 documents default suppression, flake list -v, recurrence visibility, and uncertain evidence; these descriptions match flake_cmd.go:474–486.
- BR-3 — addressed — TestFlakeListRendersFixedRecurrenceAndCoverageFromRealGit exercises default and verbose output, recurrence, and ambiguous coverage using real merge history. Temporary source overlays disabling suppression or recurrence annotation each made it fail.
- BR-4 — addressed — classifyFixedFlake now returns reasons for unreadable details, ambiguous ownership, missing merges, and ancestry failures. Decision tests assert these reasons; removing coverage-reason rendering made the real-history command test fail.

### Raised

- **BR-5** [Critical] `exact-test-ownership` (kind: defect) ARCH-PURPOSE: package-basename matching can hide an unrelated test
  cmd/sdlc/flake_fixed.go:171 reduces package identity to path.Base(pkg). For a failure named github.com/xianxu/ariadne/cmd/internal/launcher.TestX, details reading “TestX fails in pkg/launcher/run_test.go” incorrectly establish ownership. A temporary regression test confirmed that classifyFixedFlake returns flakeFixed when the issue’s merge is on main but absent from the failing head, causing default rendering to suppress the unrelated failure. bareCollision does not protect this case when only one package occurs in retained attempts. Require unambiguous package identity for textual ownership, retain uncertain rows, and cover same-basename packages across substantive sections and dedicated Log lines.

## Round 3 — 2026-10-19T06:40:39-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — The Core concepts table correctly declares classifyFixedFlake as INTEGRATION; fixedIssueCovers is tested without IO.
- BR-2 — addressed — README.md:527–533 documents default suppression, verbose attribution, recurrence, and uncertainty, matching the command implementation.
- BR-3 — addressed — TestFlakeListRendersFixedRecurrenceAndCoverageFromRealGit verifies default suppression, verbose output, recurrence, and ambiguous coverage through real merge history.
- BR-4 — addressed — The classifier returns coverage reasons, decision tests assert them, and the command test verifies their rendering.
- BR-5 — not-addressed — ARCH-PURPOSE / ARCH-SECURE: cmd/sdlc/flake_fixed.go:239–243 searches the entire remaining line for ".go", crossing path-token boundaries. For github.com/xianxu/ariadne/cmd/internal/launcher.TestX, "TestX fails in cmd/internal/launcher/subpkg (see run_test.go)." incorrectly establishes parent-package ownership. Temporary regression probes failed in both Problem and dedicated Log sections; a command-level probe confirmed default output silently omitted the unrelated failure. The exact-test-ownership family remains unresolved. Fix the rule across all ownership contexts: recognize one contiguous path token and accept only the exact package directory or a Go source file directly inside it.

## Round 4 — 2026-10-19T07:24:10-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — The plan correctly classifies classifyFixedFlake as INTEGRATION; fixedIssueCovers remains independently tested without IO.
- BR-2 — addressed — README.md:527 documents default suppression, verbose attribution, recurrence and uncertainty, matching flake_cmd.go.
- BR-3 — addressed — TestFlakeListRendersFixedRecurrenceAndCoverageFromRealGit verifies default suppression, verbose attribution, recurrence and ambiguous coverage using real Git history.
- BR-4 — addressed — The classifier preserves uncertainty reasons; decision tests assert them and the command regression verifies their rendering.
- BR-5 — addressed — Package-path heuristics are removed. Restoring the original matcher through a temporary source overlay makes both same-basename regressions and the nested-subpackage regression fail; these pass at the pinned head.

### Raised

- **BR-6** [Critical] `exact-test-ownership` (kind: defect) ARCH-PURPOSE / ARCH-SECURE: Markdown samples can still hide unrelated flakes
  cmd/sdlc/flake_fixed.go:149 uses UnterminatedIsProse, so an unclosed backtick or tilde fence in Log exposes a sample "- Test: <fully qualified name>" as ownership. Lines 162–167 also accept indented code and HTML-comment contents. Temporary matcher and command probes confirmed silent default suppression for all four forms; closed backtick and tilde samples correctly stayed visible. This is the 2nd finding in family exact-test-ownership. Earlier rounds fixed instances. State and enforce the rule across actual Log entries, closed/unclosed fences, indented code and comments: only an actual structured ownership entry may claim a test; ambiguous or truncated markup must preserve visibility. Add regression tests for that enumeration, including command output.

## Round 5 — 2026-10-19T08:30:58-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — The Core concepts table correctly classifies classifyFixedFlake as INTEGRATION and fixedIssueCovers as PURE; the matcher tests perform no IO.
- BR-2 — addressed — README.md:527 documents default suppression, structured ownership, verbose output, recurrence, and uncertainty, matching the implemented command.
- BR-3 — addressed — TestFlakeListRendersFixedRecurrenceAndCoverageFromRealGit exercises default suppression, verbose fixed-row rendering, and visible recurrence through a real merge history; it passes at the pinned head.
- BR-4 — addressed — classifyFixedFlake returns specific uncertainty reasons, runFlakeListMode renders them, and the decision matrix plus command-level ambiguous-coverage assertion pass.
- BR-5 — addressed — Ownership no longer uses package-basename or free-prose inference. Exact card values and dedicated Log entries are tested against wrong-package, nested-package, and substring cases.
- BR-6 — not-addressed — Critical, ARCH-PURPOSE / ARCH-SECURE, family exact-test-ownership: cmd/sdlc/flake_fixed.go:153-191 still accepts a sample inside "## Log\n<pre>\n- Test: <fully qualified name>\n</pre>\n", including an unclosed <pre>. A disposable command regression using the existing real-Git fixture confirmed that this sample silently removes the pre-fix row. Line 191 also accepts an unmatched opening backtick because strings.Trim removes arbitrary backticks. The reported fence, indentation, and comment fixes have genuine red/green regression evidence, but the rule that only actual, unambiguous structured entries establish ownership remains unenforced. Define and enforce the complete accepted Markdown entry grammar, rejecting raw HTML sample contexts and malformed delimiters, and cover it through matcher and command tests rather than adding another isolated exclusion.

## Round 6 — 2026-10-19T08:47:01-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — The Core concepts table correctly declares classifyFixedFlake INTEGRATION; fixedIssueCovers performs only string comparisons without IO.
- BR-2 — addressed — README.md:527 documents default suppression, flake list -v and recurrence, consistent with cmd/sdlc/flake_cmd.go rendering.
- BR-3 — addressed — TestFlakeListRendersFixedRecurrenceAndCoverageFromRealGit passes and asserts default suppression, verbose fixed rows, recurrence and ambiguous coverage.
- BR-4 — addressed — Classifier uncertainty reasons reach command output; passing decision and rendering tests verify missing merges, Git uncertainty and ambiguous ownership.
- BR-5 — addressed — Ownership now uses exact card-list equality; package-basename and textual-path inference have been removed.
- BR-6 — addressed — Published details no longer enter ownership evidence. TestFlakeListIgnoresPublishedLogOwnership passes on head; restoring the prior implementation through a temporary Go overlay makes all three cases fail: plain Log, raw HTML and unmatched backticks.

## Open findings

(none — every finding has been disposed)
