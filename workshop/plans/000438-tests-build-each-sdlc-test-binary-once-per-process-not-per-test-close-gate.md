---
gate: boundary-review
issue: 438
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: ae8f83c9-b1b9-421a-838b-7a23e79470eb
            repository: github.com/xianxu/ariadne
            issue_id: "000438"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/fb6ca8d3b568555d
            reviewed_head: dacdc2a2fcf584094ce86e48cb5ea7a14fe192fa
            slot: ariadne:2
            requested_at: "2026-10-18T10:40:03.811796Z"
          phase: finished
          pid: 54094
          started_at: "2026-10-18T10:40:03.902522Z"
          finished_at: "2026-10-18T10:41:49.176327Z"
          session_id: 72f26c6f-cf0d-43a7-aa33-ae74e09327a7
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: dacdc2a2fcf584094ce86e48cb5ea7a14fe192fa
        tracker_head: 829b7947f22644849f7dc8bbd03c61a827b19b4a
        branch_point: 81acc7b58973a0612d17c9f8e4bcdba565410e11
        since: "2026-10-18T02:40:05-07:00"
        until: "2026-10-18T02:57:19-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T02:40:05-07:00"
              end: "2026-10-18T02:57:19-07:00"
              began: "2026-10-18T02:40:05-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 9b7eea0f522b8001ccfbdaf3d43440245791fa68ac9cdaa9666cf51b029a4406
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 136540275
              windows:
                - start: "2026-10-18T02:40:05-07:00"
                  end: "2026-10-18T02:57:19-07:00"
      "n": 1
      timestamp: "2026-10-18T03:41:49-07:00"
      dispatched: "2026-10-18T03:40:03-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: atlas update appears missing for testfix.BuiltBinary / RemoveBuiltBinaries
          detail: atlas/workflow/sdlc-binary.md documents each testfix test convention (PreferRealGit, QuietBackgroundGit, Quarantine, Watch, AwaitExit, DeadlineOnCue) but not the new build-once helper. Add a paragraph saying a test needing a built binary calls testfix.BuiltBinary (one build per process into sdlc-testbin-*, 24h stale sweep) and that the package TestMain calls RemoveBuiltBinaries after m.Run.
          family: docs-gate-new-surface
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: test-go-build lint message should name testfix.BuiltBinary as the fix
          detail: pkg/golint/rules.go:30 says build it once per package (TestMain); the helper now exists and is the intended remedy.
          family: lint-message-names-fix
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: settings_native weave build lost its 5-minute CommandContext deadline
          detail: buildCommand uses exec.Command with no timeout, so a hung go build now runs until the test binary's -timeout instead of failing at 5 minutes.
          family: subprocess-bounded-deadline
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: memoized build runs under the first caller's environment
          detail: Whichever test calls BuiltBinary first decides the HOME, GOFLAGS and PATH the shared build sees. Nothing breaks today, but the BuiltBinary doc comment should state it.
          family: memoized-env-capture
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 4d1bd3ab-ecb3-4b38-956c-aef080165a15
            repository: github.com/xianxu/ariadne
            issue_id: "000438"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/40c1b49688537acd
            reviewed_head: 81c832d5e0008578d7a03783a0d4055cc6358000
            slot: ariadne:2
            requested_at: "2026-10-18T11:24:20.666817Z"
          phase: finished
          pid: 14438
          started_at: "2026-10-18T11:24:20.833827Z"
          finished_at: "2026-10-18T11:26:16.871267Z"
          session_id: a16aa917-701b-4f6b-8336-06bec7db4acd
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 81c832d5e0008578d7a03783a0d4055cc6358000
        tracker_head: 829b7947f22644849f7dc8bbd03c61a827b19b4a
        branch_point: 81acc7b58973a0612d17c9f8e4bcdba565410e11
        since: "2026-10-18T02:40:05-07:00"
        until: "2026-10-18T03:42:41-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T02:40:05-07:00"
              end: "2026-10-18T03:42:41-07:00"
              began: "2026-10-18T02:40:05-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 73d8f838809460046ceaf63b784ec5553826af98430f9fbcd519aefa3a91c5d3
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 137718536
              windows:
                - start: "2026-10-18T02:40:05-07:00"
                  end: "2026-10-18T03:42:41-07:00"
        queue_waits:
            - attempt_id: bc83a37c06dffbab2c718667941a1877
              slot: ariadne:2
              record_digest: 27b2025da663544145cd7c1240134b00e665fb57c5ebc34cfb0733755b143a77
              start: "2026-10-18T09:57:22.453377Z"
              end: "2026-10-18T10:31:02.179937Z"
              queue_seconds: 2019.72656
              raw_seconds: 2019.72656
              exact: true
        cuts:
            - dispatched: "2026-10-18T03:40:03-07:00"
              finalized: "2026-10-18T03:41:49-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-18T04:26:18-07:00"
      dispatched: "2026-10-18T04:24:20-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: atlas/workflow/sdlc-binary.md:1896 adds a paragraph next to the other testfix conventions covering BuiltBinary, the sdlc-testbin-* directory, RemoveBuiltBinaries after m.Run, and the 24 h sweep.
          round: 2
        - id: BR-2
          disposition: addressed
          note: pkg/golint/rules.go:30 now names testfix.BuiltBinary as the remedy.
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: 'build.go:38-54 adds a 5-minute AfterFunc kill, but no test exercises it, and killing only the go process may not bound a hung compile/link child holding the output pipe (Wait waits for the copy goroutines). Rule for the subprocess-bounded-deadline family: every bounded subprocess sets cmd.WaitDelay along with its kill, and a fake hanging build shows the bound holds.'
          round: 2
        - id: BR-4
          disposition: addressed
          note: The BuiltBinary doc comment (build.go:61-62) now says the build runs once under the first caller's environment (HOME, GOFLAGS, PATH).
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: 81c832d5e0008578d7a03783a0d4055cc6358000
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 62fcab04-6d4a-44ae-b10a-b2e50fa7fd7e
            repository: github.com/xianxu/ariadne
            issue_id: "000438"
            recipe: fix-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/6150bcda6c7799b7
            reviewed_head: a1bf416f615f715f56953d949d3c0d5c3c66eed1
            slot: ariadne:2
            requested_at: "2026-10-18T11:27:58.58051Z"
          phase: finished
          pid: 88026
          started_at: "2026-10-18T11:27:58.735592Z"
          finished_at: "2026-10-18T11:29:52.164621Z"
          session_id: 3906baf3-4068-4960-8852-01e0fdc9300e
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: a1bf416f615f715f56953d949d3c0d5c3c66eed1
        tracker_head: 829b7947f22644849f7dc8bbd03c61a827b19b4a
        branch_point: 81acc7b58973a0612d17c9f8e4bcdba565410e11
        since: "2026-10-18T02:40:05-07:00"
        until: "2026-10-18T04:27:33-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T02:40:05-07:00"
              end: "2026-10-18T04:27:33-07:00"
              began: "2026-10-18T02:40:05-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6558ff3660328a60784671256055c6ea5c9511825f6f9ec90cf4ab8e5311b795
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 138201603
              windows:
                - start: "2026-10-18T02:40:05-07:00"
                  end: "2026-10-18T04:27:33-07:00"
        queue_waits:
            - attempt_id: bc83a37c06dffbab2c718667941a1877
              slot: ariadne:2
              record_digest: 27b2025da663544145cd7c1240134b00e665fb57c5ebc34cfb0733755b143a77
              start: "2026-10-18T09:57:22.453377Z"
              end: "2026-10-18T10:31:02.179937Z"
              queue_seconds: 2019.72656
              raw_seconds: 2019.72656
              exact: true
            - attempt_id: 04b5567c65a5b85720405001694a9c4a
              slot: ariadne:2
              record_digest: 03e1dcefaaa5870f083909d5727a6251159cc00b1215d61a4958a374fce4b6e0
              start: "2026-10-18T10:42:44.577937Z"
              end: "2026-10-18T11:15:06.042596Z"
              queue_seconds: 1941.464659
              raw_seconds: 1941.464659
              exact: true
        cuts:
            - dispatched: "2026-10-18T03:40:03-07:00"
              finalized: "2026-10-18T03:41:49-07:00"
              finalizing: false
            - dispatched: "2026-10-18T04:24:20-07:00"
              finalized: "2026-10-18T04:26:18-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-18T04:29:53-07:00"
      dispatched: "2026-10-18T04:27:58-07:00"
      agent: claude
      dispose:
        - id: BR-3
          disposition: not-addressed
          note: 'runBounded (build.go:52-72) adds WaitDelay plus a process-group kill, and the weave build reaches it through buildCommand, but the regression test does not fail without the fix: with WaitDelay and ownGroup removed, TestRunBoundedKillsTheGroupAtItsLimit still passes, in 60.16s, because the fixture sleeps 60s, which is under HangCeiling (2m). Make the fixture outlive the ceiling (sleep 600) or assert took < limit+WaitDelay+slack. Also wrap the error as exceeded only when Wait''s err is non-nil (build.go:69-71).'
          round: 3
      recipe: fix-review
      isolation: fenced
      reviewed: a1bf416f615f715f56953d949d3c0d5c3c66eed1
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 700d0f7d-95cf-42b6-bb6b-8c6582d35f90
            repository: github.com/xianxu/ariadne
            issue_id: "000438"
            recipe: fix-review
            round: 4
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/6a862d0468693995
            reviewed_head: f2f3d2ebb4f7cfab41bfbcaf229af02b3f786c11
            slot: ariadne:2
            requested_at: "2026-10-18T11:32:20.130214Z"
          phase: finished
          pid: 87798
          started_at: "2026-10-18T11:32:20.227424Z"
          finished_at: "2026-10-18T11:43:02.531999Z"
          session_id: 6bc12a8a-41dd-4746-80ed-06f1ef58827c
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: f2f3d2ebb4f7cfab41bfbcaf229af02b3f786c11
        tracker_head: 829b7947f22644849f7dc8bbd03c61a827b19b4a
        branch_point: 81acc7b58973a0612d17c9f8e4bcdba565410e11
        since: "2026-10-18T02:40:05-07:00"
        until: "2026-10-18T04:32:02-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T02:40:05-07:00"
              end: "2026-10-18T04:32:02-07:00"
              began: "2026-10-18T02:40:05-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 609319ebbba743ff2450d18c0026f878d8f3c04c5461afee3970dab1e37dd901
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 139111837
              windows:
                - start: "2026-10-18T02:40:05-07:00"
                  end: "2026-10-18T04:32:02-07:00"
        queue_waits:
            - attempt_id: bc83a37c06dffbab2c718667941a1877
              slot: ariadne:2
              record_digest: 27b2025da663544145cd7c1240134b00e665fb57c5ebc34cfb0733755b143a77
              start: "2026-10-18T09:57:22.453377Z"
              end: "2026-10-18T10:31:02.179937Z"
              queue_seconds: 2019.72656
              raw_seconds: 2019.72656
              exact: true
            - attempt_id: 04b5567c65a5b85720405001694a9c4a
              slot: ariadne:2
              record_digest: 03e1dcefaaa5870f083909d5727a6251159cc00b1215d61a4958a374fce4b6e0
              start: "2026-10-18T10:42:44.577937Z"
              end: "2026-10-18T11:15:06.042596Z"
              queue_seconds: 1941.464659
              raw_seconds: 1941.464659
              exact: true
        cuts:
            - dispatched: "2026-10-18T03:40:03-07:00"
              finalized: "2026-10-18T03:41:49-07:00"
              finalizing: false
            - dispatched: "2026-10-18T04:24:20-07:00"
              finalized: "2026-10-18T04:26:18-07:00"
              finalizing: true
            - dispatched: "2026-10-18T04:27:58-07:00"
              finalized: "2026-10-18T04:29:53-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-18T04:43:25-07:00"
      dispatched: "2026-10-18T04:32:20-07:00"
      agent: claude
      dispose:
        - id: BR-3
          disposition: addressed
          note: build_unix_test.go now uses sleep 600 and asserts took within limit+1s+5s; removing WaitDelay and ownGroup from build.go makes TestRunBoundedKillsTheGroupAtItsLimit hang to the 600s go test timeout (FAIL), and with the fix it passes under -race in 1.5s. build.go:68 now wraps exceeded only when Wait returned an error.
          round: 4
      recipe: fix-review
      isolation: fenced
      reviewed: f2f3d2ebb4f7cfab41bfbcaf229af02b3f786c11
      blocked: false
---

# Gate ledger — ariadne#438 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T03:41:49-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `docs-gate-new-surface` (kind: defect) atlas update appears missing for testfix.BuiltBinary / RemoveBuiltBinaries
  atlas/workflow/sdlc-binary.md documents each testfix test convention (PreferRealGit, QuietBackgroundGit, Quarantine, Watch, AwaitExit, DeadlineOnCue) but not the new build-once helper. Add a paragraph saying a test needing a built binary calls testfix.BuiltBinary (one build per process into sdlc-testbin-*, 24h stale sweep) and that the package TestMain calls RemoveBuiltBinaries after m.Run.
- **BR-2** [Minor] `lint-message-names-fix` (kind: defect) test-go-build lint message should name testfix.BuiltBinary as the fix
  pkg/golint/rules.go:30 says build it once per package (TestMain); the helper now exists and is the intended remedy.
- **BR-3** [Minor] `subprocess-bounded-deadline` (kind: defect) settings_native weave build lost its 5-minute CommandContext deadline
  buildCommand uses exec.Command with no timeout, so a hung go build now runs until the test binary's -timeout instead of failing at 5 minutes.
- **BR-4** [Minor] `memoized-env-capture` (kind: defect) memoized build runs under the first caller's environment
  Whichever test calls BuiltBinary first decides the HOME, GOFLAGS and PATH the shared build sees. Nothing breaks today, but the BuiltBinary doc comment should state it.

## Round 2 — 2026-10-18T04:26:18-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — atlas/workflow/sdlc-binary.md:1896 adds a paragraph next to the other testfix conventions covering BuiltBinary, the sdlc-testbin-* directory, RemoveBuiltBinaries after m.Run, and the 24 h sweep.
- BR-2 — addressed — pkg/golint/rules.go:30 now names testfix.BuiltBinary as the remedy.
- BR-3 — not-addressed — build.go:38-54 adds a 5-minute AfterFunc kill, but no test exercises it, and killing only the go process may not bound a hung compile/link child holding the output pipe (Wait waits for the copy goroutines). Rule for the subprocess-bounded-deadline family: every bounded subprocess sets cmd.WaitDelay along with its kill, and a fake hanging build shows the bound holds.
- BR-4 — addressed — The BuiltBinary doc comment (build.go:61-62) now says the build runs once under the first caller's environment (HOME, GOFLAGS, PATH).

## Round 3 — 2026-10-18T04:29:53-07:00 (claude) — passed

### Disposed

- BR-3 — not-addressed — runBounded (build.go:52-72) adds WaitDelay plus a process-group kill, and the weave build reaches it through buildCommand, but the regression test does not fail without the fix: with WaitDelay and ownGroup removed, TestRunBoundedKillsTheGroupAtItsLimit still passes, in 60.16s, because the fixture sleeps 60s, which is under HangCeiling (2m). Make the fixture outlive the ceiling (sleep 600) or assert took < limit+WaitDelay+slack. Also wrap the error as exceeded only when Wait's err is non-nil (build.go:69-71).

## Round 4 — 2026-10-18T04:43:25-07:00 (claude) — passed

### Disposed

- BR-3 — addressed — build_unix_test.go now uses sleep 600 and asserts took within limit+1s+5s; removing WaitDelay and ownGroup from build.go makes TestRunBoundedKillsTheGroupAtItsLimit hang to the 600s go test timeout (FAIL), and with the fix it passes under -race in 1.5s. build.go:68 now wraps exceeded only when Wait returned an error.

## Open findings

(none — every finding has been disposed)
