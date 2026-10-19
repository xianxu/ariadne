---
gate: boundary-review
issue: 448
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a9d58fc7-6b85-409e-97b8-f2cf48c40bfc
            repository: github.com/xianxu/ariadne
            issue_id: "000448"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/11ca7eb0cadd83e8
            reviewed_head: 3df9c09a2d08a9de0d01c24715758ee512e89a78
            slot: ariadne:3
            requested_at: "2026-10-19T03:03:41.795207Z"
          phase: finished
          pid: 52599
          started_at: "2026-10-19T03:03:41.916844Z"
          finished_at: "2026-10-19T03:05:29.79983Z"
          session_id: 01a1521d-6b8a-7fc2-9a4b-3e8e0a45e1a6
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 3df9c09a2d08a9de0d01c24715758ee512e89a78
        tracker_head: 33006f7d3ad36ed0075430b1583a99a545ce1f9e
        branch_point: 724addd9dcb79045ff15ecf934ee07daa86c52ba
        since: "2026-10-18T19:54:59-07:00"
        until: "2026-10-18T20:03:25-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T19:54:59-07:00"
              end: "2026-10-18T20:03:25-07:00"
              began: "2026-10-18T19:54:59-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: cd21e31670a54575fe873b849d074370fda8a2ef840d2605ca0bf1acd42d8a40
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 22105282
              windows:
                - start: "2026-10-18T19:54:59-07:00"
                  end: "2026-10-18T20:03:25-07:00"
            - harness: codex
              id: 01a15216-7ad0-7643-8c58-f3dfffe7c2c7
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: cfbdd7b6c76cfb3b0fc72ac2153f516196be666443824c8cefe449eca7e4fee4
              started: "2026-10-19T02:56:07.477Z"
              spawned: true
              bytes: 473010
              windows:
                - start: "2026-10-19T02:56:07.477Z"
                  end: "2026-10-18T20:03:25-07:00"
      "n": 1
      timestamp: "2026-10-18T20:05:30-07:00"
      dispatched: "2026-10-18T20:03:41-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Independent script budgets lack a regression test
          detail: scripts/test/scripts_test.go:49-68 implements the central fix, but construct_runner_test.go:50-89 exercises only the shell runner. No test distinguishes fresh per-script contexts from a context shared across the batch; the real scripts can pass under either implementation. The complete affected family is budget independence across selected scripts in plain, no_color_ci, and forced_color_terminal. Add a deterministic regression through an injected command/context seam that fails when the context is shared, while retaining the production two-minute ceiling.
          family: behavioral-contract-regression-coverage
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: README update appears missing for construct runner --list
          detail: scripts/test/construct-scripts.test.sh:15-21 introduces the --list flag, but README.md is unchanged in the pinned range and contains no usage for it. This is the sole new command-line flag in this window. Document its NUL-delimited, non-executing output and preserve the distinction from default batch execution; atlas/workflow/sdlc-binary.md already documents the architectural flow.
          family: user-facing-surface-documentation
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: cace75ec-953b-45f7-8b04-ef4b26eb7a74
            repository: github.com/xianxu/ariadne
            issue_id: "000448"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/984533fb2d66b289
            reviewed_head: 25bf575191a753a127daf2bd1833cd5b1e27a6ec
            slot: ariadne:3
            requested_at: "2026-10-19T03:09:40.239746Z"
          phase: finished
          pid: 25703
          started_at: "2026-10-19T03:09:40.415693Z"
          finished_at: "2026-10-19T03:11:26.419898Z"
          session_id: 01a15222-e413-7bd1-959e-f23bc8e18e9a
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 25bf575191a753a127daf2bd1833cd5b1e27a6ec
        tracker_head: 33006f7d3ad36ed0075430b1583a99a545ce1f9e
        branch_point: 724addd9dcb79045ff15ecf934ee07daa86c52ba
        since: "2026-10-18T19:54:59-07:00"
        until: "2026-10-18T20:09:12-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T19:54:59-07:00"
              end: "2026-10-18T20:09:12-07:00"
              began: "2026-10-18T19:54:59-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 475b6918ead14cd5068abc8075d2413b364150bca9db65ce4f85e5552c8df24c
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 22463092
              windows:
                - start: "2026-10-18T19:54:59-07:00"
                  end: "2026-10-18T20:09:12-07:00"
            - harness: codex
              id: 01a15216-7ad0-7643-8c58-f3dfffe7c2c7
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: cfbdd7b6c76cfb3b0fc72ac2153f516196be666443824c8cefe449eca7e4fee4
              started: "2026-10-19T02:56:07.477Z"
              spawned: true
              bytes: 473010
              windows:
                - start: "2026-10-19T02:56:07.477Z"
                  end: "2026-10-18T20:09:12-07:00"
        cuts:
            - dispatched: "2026-10-18T20:03:41-07:00"
              finalized: "2026-10-18T20:05:30-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-18T20:11:29-07:00"
      dispatched: "2026-10-18T20:09:40-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'TestConstructScriptBudgetsIndependent exercises the production loop across all three environments. A temporary Go overlay moving context ownership to the batch made it fail in every environment: three budgets instead of nine, with prior budgets still live. The pinned implementation passes.'
          round: 2
        - id: BR-2
          disposition: addressed
          note: README.md:370-374 now documents --list as non-executing NUL-delimited selection and distinguishes default batch execution. This matches construct-scripts.test.sh:18-41 and its real-shell fixture tests.
          round: 2
      recipe: small-diff-review
      isolation: fenced
      reviewed: 25bf575191a753a127daf2bd1833cd5b1e27a6ec
      blocked: false
---

# Gate ledger — ariadne#448 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T20:05:30-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `behavioral-contract-regression-coverage` (kind: defect) Independent script budgets lack a regression test
  scripts/test/scripts_test.go:49-68 implements the central fix, but construct_runner_test.go:50-89 exercises only the shell runner. No test distinguishes fresh per-script contexts from a context shared across the batch; the real scripts can pass under either implementation. The complete affected family is budget independence across selected scripts in plain, no_color_ci, and forced_color_terminal. Add a deterministic regression through an injected command/context seam that fails when the context is shared, while retaining the production two-minute ceiling.
- **BR-2** [Important] `user-facing-surface-documentation` (kind: defect) README update appears missing for construct runner --list
  scripts/test/construct-scripts.test.sh:15-21 introduces the --list flag, but README.md is unchanged in the pinned range and contains no usage for it. This is the sole new command-line flag in this window. Document its NUL-delimited, non-executing output and preserve the distinction from default batch execution; atlas/workflow/sdlc-binary.md already documents the architectural flow.

## Round 2 — 2026-10-18T20:11:29-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — TestConstructScriptBudgetsIndependent exercises the production loop across all three environments. A temporary Go overlay moving context ownership to the batch made it fail in every environment: three budgets instead of nine, with prior budgets still live. The pinned implementation passes.
- BR-2 — addressed — README.md:370-374 now documents --list as non-executing NUL-delimited selection and distinguishes default batch execution. This matches construct-scripts.test.sh:18-41 and its real-shell fixture tests.

## Open findings

(none — every finding has been disposed)
