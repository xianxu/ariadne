---
gate: boundary-review
issue: 407
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 19c0be95-9d9e-4532-9946-2c2512e4431e
            repository: github.com/xianxu/ariadne
            issue_id: "000407"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/44cf1ce3178f19f8
            reviewed_head: 21453b203f07329e7153db3fa97a171fde0f6be2
            slot: ariadne:2
            requested_at: "2026-10-18T07:57:06.375852Z"
          phase: finished
          pid: 93630
          started_at: "2026-10-18T07:57:06.475688Z"
          finished_at: "2026-10-18T07:59:40.540887Z"
          session_id: dcb316aa-c569-4dab-9ae1-28e6b6a381e7
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 21453b203f07329e7153db3fa97a171fde0f6be2
        tracker_head: 19faeb8cee284d6467c7e03c7cf5f5d51fb11a3a
        branch_point: e5f48d8326eab4ce46e7c8fae2d89ae96498f431
        since: "2026-10-18T00:13:46-07:00"
        until: "2026-10-18T00:21:43-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T00:13:46-07:00"
              end: "2026-10-18T00:21:43-07:00"
              began: "2026-10-18T00:13:46-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 78ff834e90574dc48033d274ef830e5bea737993992cb3d931a6a3933fe4bd6b
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 122996526
              windows:
                - start: "2026-10-18T00:13:46-07:00"
                  end: "2026-10-18T00:21:43-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.17408333333333334
      "n": 1
      timestamp: "2026-10-18T00:59:41-07:00"
      dispatched: "2026-10-18T00:57:06-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Plan Core concepts and integration claims contradict the code (rule.go/suppress.go absent, Match signature, git-based walk)
          detail: Plan places Rule/Finding in pkg/golint/rule.go and suppressions in suppress.go; both live in check.go. Plan Match is func(*ast.File,*token.FileSet) []token.Pos; code is func(*ast.File) []ast.Node. Plan says repo test walks cmd/ and pkg/, skips generated files, and calls no external binary; repo_test.go:17 shells out to git ls-files over the whole repo and skips no generated files. Fix with a Revisions entry (or skip generated files).
          family: core-concepts-table-drift
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Spec seed rule for wall-clock duration assertions dropped without a declared deviation or follow-up
          detail: The Spec seed list and issue title include wall-clock asserts; the plan's three deviations (engine, pair, wire decode) do not mention it, and 14 if-time.Since assertions exist in cmd test files. Add the rule or declare it as a fourth deviation with a filed follow-up (ARCH-PURPOSE).
          family: undeclared-scope-reduction
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: gitx default-deadline test no longer exercises the gitCommandTimeout ceiling of the default runner
          detail: The default-deadline case now calls boundedGitOutput with a cued context instead of run() with a shrunk gitCommandTimeout, so removing context.WithTimeout(ctx, gitCommandTimeout) in lifetime.go:25 fails no test. Add a race-free check, e.g. that boundedGitCommand(nil, ...) returns a context with a deadline within gitCommandTimeout.
          family: test-conversion-drops-coverage
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: test-shrunk-deadline matches locals (plan says package-level) and misses selector/field targets
          detail: shrunkDeadlines accepts any *ast.Ident LHS with =, and skips cfg.Timeout = 10*time.Millisecond style field overrides.
          family: lint-rule-matcher-precision
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: test-go-build and tmp-literal match by literal and identifier name only (variable go binary or aliased import evades)
          family: lint-rule-matcher-precision
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: Spec rule-hygiene clause (review a rule that never fires in a quarter) has no mechanism and is not listed as deferred
          family: undeclared-scope-reduction
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 107842cb-6e35-4373-b5dd-a982af787bb3
            repository: github.com/xianxu/ariadne
            issue_id: "000407"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/217836079e861855
            reviewed_head: e8a9c84116b1901d7f6e56d741816c74143d4f25
            slot: ariadne:2
            requested_at: "2026-10-18T08:42:49.155215Z"
          phase: finished
          pid: 57520
          started_at: "2026-10-18T08:42:49.272634Z"
          finished_at: "2026-10-18T08:45:55.995525Z"
          session_id: ea26b0b0-ee62-43a8-937a-8989ea780f00
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: e8a9c84116b1901d7f6e56d741816c74143d4f25
        tracker_head: 58c1e735985ff44f811dcfe8068d0bd852dfb033
        branch_point: e5f48d8326eab4ce46e7c8fae2d89ae96498f431
        since: "2026-10-18T00:13:46-07:00"
        until: "2026-10-18T01:01:32-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T00:13:46-07:00"
              end: "2026-10-18T01:01:32-07:00"
              began: "2026-10-18T00:13:46-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: b25024420b3e13a5c6454c1c21efeecf3dc5379c68b61750b376e7caf2ba1a92
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 125689663
              windows:
                - start: "2026-10-18T00:13:46-07:00"
                  end: "2026-10-18T01:01:32-07:00"
        queue_waits:
            - attempt_id: ac42ce47fe4f5970d1a44fdf709e4a73
              slot: ariadne:2
              record_digest: a2a4033a4c5479a9b6df0736854b581dc1fab2924995aded3f43fd38970796e6
              start: "2026-10-18T07:21:47.810576Z"
              end: "2026-10-18T07:43:54.809558Z"
              queue_seconds: 1326.998982
              raw_seconds: 1326.998982
              exact: true
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.17408333333333334
        cuts:
            - dispatched: "2026-10-18T00:57:06-07:00"
              finalized: "2026-10-18T00:59:41-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-18T01:45:56-07:00"
      dispatched: "2026-10-18T01:42:49-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Plan Revisions entry (close review round 1) now matches check.go/rules.go, Match signature, git ls-files walk; ast.IsGenerated skip added in check.go with TestGeneratedFilesAreSkipped.
          round: 2
        - id: BR-2
          disposition: addressed
          note: test-tight-wallclock added in rules.go with flag/no-flag cases in rules_test.go; two existing hits carry reasons.
          round: 2
        - id: BR-3
          disposition: addressed
          note: 'Verified by revert: replacing WithTimeout(ctx, gitCommandTimeout) with WithCancel in lifetime.go makes TestBoundedGitCommandAppliesDefaultCeiling fail; restored, gitx passes.'
          round: 2
        - id: BR-4
          disposition: addressed
          note: SelectorExpr field targets now matched (cfg.ShutdownGrace test case); local-vs-package limitation declared in Revisions.
          round: 2
        - id: BR-5
          disposition: addressed
          note: Import-alias resolution via localName with TestRulesFollowImportNames; variable go binary limitation declared in Revisions.
          round: 2
        - id: BR-6
          disposition: addressed
          note: 'Plan Revisions declares the never-fires review clause deferred to #439 with the reason (needs a recorded firing count).'
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: e8a9c84116b1901d7f6e56d741816c74143d4f25
      blocked: false
---

# Gate ledger — ariadne#407 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T00:59:41-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Critical] `core-concepts-table-drift` (kind: defect) Plan Core concepts and integration claims contradict the code (rule.go/suppress.go absent, Match signature, git-based walk)
  Plan places Rule/Finding in pkg/golint/rule.go and suppressions in suppress.go; both live in check.go. Plan Match is func(*ast.File,*token.FileSet) []token.Pos; code is func(*ast.File) []ast.Node. Plan says repo test walks cmd/ and pkg/, skips generated files, and calls no external binary; repo_test.go:17 shells out to git ls-files over the whole repo and skips no generated files. Fix with a Revisions entry (or skip generated files).
- **BR-2** [Important] `undeclared-scope-reduction` (kind: defect) Spec seed rule for wall-clock duration assertions dropped without a declared deviation or follow-up
  The Spec seed list and issue title include wall-clock asserts; the plan's three deviations (engine, pair, wire decode) do not mention it, and 14 if-time.Since assertions exist in cmd test files. Add the rule or declare it as a fourth deviation with a filed follow-up (ARCH-PURPOSE).
- **BR-3** [Important] `test-conversion-drops-coverage` (kind: defect) gitx default-deadline test no longer exercises the gitCommandTimeout ceiling of the default runner
  The default-deadline case now calls boundedGitOutput with a cued context instead of run() with a shrunk gitCommandTimeout, so removing context.WithTimeout(ctx, gitCommandTimeout) in lifetime.go:25 fails no test. Add a race-free check, e.g. that boundedGitCommand(nil, ...) returns a context with a deadline within gitCommandTimeout.
- **BR-4** [Minor] `lint-rule-matcher-precision` (kind: defect) test-shrunk-deadline matches locals (plan says package-level) and misses selector/field targets
  shrunkDeadlines accepts any *ast.Ident LHS with =, and skips cfg.Timeout = 10*time.Millisecond style field overrides.
- **BR-5** [Minor] `lint-rule-matcher-precision` (kind: defect) test-go-build and tmp-literal match by literal and identifier name only (variable go binary or aliased import evades)
- **BR-6** [Minor] `undeclared-scope-reduction` (kind: defect) Spec rule-hygiene clause (review a rule that never fires in a quarter) has no mechanism and is not listed as deferred

## Round 2 — 2026-10-18T01:45:56-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Plan Revisions entry (close review round 1) now matches check.go/rules.go, Match signature, git ls-files walk; ast.IsGenerated skip added in check.go with TestGeneratedFilesAreSkipped.
- BR-2 — addressed — test-tight-wallclock added in rules.go with flag/no-flag cases in rules_test.go; two existing hits carry reasons.
- BR-3 — addressed — Verified by revert: replacing WithTimeout(ctx, gitCommandTimeout) with WithCancel in lifetime.go makes TestBoundedGitCommandAppliesDefaultCeiling fail; restored, gitx passes.
- BR-4 — addressed — SelectorExpr field targets now matched (cfg.ShutdownGrace test case); local-vs-package limitation declared in Revisions.
- BR-5 — addressed — Import-alias resolution via localName with TestRulesFollowImportNames; variable go binary limitation declared in Revisions.
- BR-6 — addressed — Plan Revisions declares the never-fires review clause deferred to #439 with the reason (needs a recorded firing count).

## Open findings

(none — every finding has been disposed)
