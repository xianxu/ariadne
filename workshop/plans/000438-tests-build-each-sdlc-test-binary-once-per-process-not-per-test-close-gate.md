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

## Open findings

- **BR-1** [Important] `docs-gate-new-surface` (kind: defect) atlas update appears missing for testfix.BuiltBinary / RemoveBuiltBinaries
- **BR-2** [Minor] `lint-message-names-fix` (kind: defect) test-go-build lint message should name testfix.BuiltBinary as the fix
- **BR-3** [Minor] `subprocess-bounded-deadline` (kind: defect) settings_native weave build lost its 5-minute CommandContext deadline
- **BR-4** [Minor] `memoized-env-capture` (kind: defect) memoized build runs under the first caller's environment
