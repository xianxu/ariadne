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

## Open findings

- **BR-1** [Critical] `core-concepts-table-drift` (kind: defect) Plan Core concepts and integration claims contradict the code (rule.go/suppress.go absent, Match signature, git-based walk)
- **BR-2** [Important] `undeclared-scope-reduction` (kind: defect) Spec seed rule for wall-clock duration assertions dropped without a declared deviation or follow-up
- **BR-3** [Important] `test-conversion-drops-coverage` (kind: defect) gitx default-deadline test no longer exercises the gitCommandTimeout ceiling of the default runner
- **BR-4** [Minor] `lint-rule-matcher-precision` (kind: defect) test-shrunk-deadline matches locals (plan says package-level) and misses selector/field targets
- **BR-5** [Minor] `lint-rule-matcher-precision` (kind: defect) test-go-build and tmp-literal match by literal and identifier name only (variable go binary or aliased import evades)
- **BR-6** [Minor] `undeclared-scope-reduction` (kind: defect) Spec rule-hygiene clause (review a rule that never fires in a quarter) has no mechanism and is not listed as deferred
