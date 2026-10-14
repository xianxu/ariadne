---
gate: boundary-review
issue: 366
id_prefix: BR
rounds:
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 5d13f3bc5d5a3d703dbb6b2a6edb2794daf22f60
        tracker_head: 7b1e911628ebfc46694388bf87ee155ce914c023
        branch_point: 407e48a39057503fab7cdecaf6c69ce338ce7631
        since: "2026-10-14T09:00:28-07:00"
        until: "2026-10-14T10:10:10-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-14T09:00:28-07:00"
              end: "2026-10-14T10:10:10-07:00"
              began: "2026-10-14T09:00:28-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a136db-6cd1-7a31-87f3-271411613545
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: b3ac2df815d6c8a3498d4efd19057da9a8b9ec62a09afcbab1d474b1c0e871d4
              started: "2026-10-13T20:02:12.897Z"
              spawned: false
              bytes: 17187302
              windows:
                - start: "2026-10-14T09:00:28-07:00"
                  end: "2026-10-14T10:10:10-07:00"
            - harness: codex
              id: 01a13b25-21bf-7a91-8b3f-73636540a43a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              parent_id: 01a136db-6cd1-7a31-87f3-271411613545
              digest: acb96ea4c234b4f42f1c7602acc00e5536c923ac09bf0e148a0c2ded69fef6d6
              started: "2026-10-14T16:00:51.801Z"
              spawned: true
              bytes: 1263697
              windows:
                - start: "2026-10-14T16:00:51.801Z"
                  end: "2026-10-14T10:10:10-07:00"
            - harness: codex
              id: 01a13b25-4c24-79a0-8aef-22ef38dcf027
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              parent_id: 01a136db-6cd1-7a31-87f3-271411613545
              digest: 5e82a6824cceb5010eea057d278edf086cb3db4cb9232cf8fc2ff5b6f5525165
              started: "2026-10-14T16:01:02.667Z"
              spawned: true
              bytes: 535353
              windows:
                - start: "2026-10-14T16:01:02.667Z"
                  end: "2026-10-14T10:10:10-07:00"
      "n": 1
      timestamp: "2026-10-14T10:15:57-07:00"
      dispatched: "2026-10-14T10:10:40-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Important
          title: README update is missing for weave refresh --substrate
          detail: cmd/weave/refresh.go:67 introduces a user-facing flag, but README.md:214 still documents only ordinary refresh and --rebase. Add substrate-only usage, branch preservation, incompatible flags, and the source-owned-helper exception to satisfy the docs update gate.
          family: user-facing-surface-documentation
          round: 1
        - kind: unverifiable
          id: BR-2
          severity: Important
          title: Existing refresh retry regression cannot initialize under the review Git ceiling
          detail: 'Running go test ./cmd/weave ./cmd/weave/internal/refresh -run ''^(TestRefresh|TestSubstrate)'' -count=1 -timeout 3m -v failed at cmd/weave/refresh_test.go:44: TestRefreshCLICompilesAndRetries calls compile before initializing its temporary repositories, and rev-parse --show-toplevel returns fatal: not a git repository under GIT_CEILING_DIRECTORIES. Provide passing host evidence for this regression and the required all-weave run at the pinned head, or make the fixture independent of enclosing repository discovery.'
          family: required-regression-verification
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#366 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T10:15:57-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `user-facing-surface-documentation` README update is missing for weave refresh --substrate
  cmd/weave/refresh.go:67 introduces a user-facing flag, but README.md:214 still documents only ordinary refresh and --rebase. Add substrate-only usage, branch preservation, incompatible flags, and the source-owned-helper exception to satisfy the docs update gate.
- **BR-2** [Important] `required-regression-verification` (kind: unverifiable) Existing refresh retry regression cannot initialize under the review Git ceiling
  Running go test ./cmd/weave ./cmd/weave/internal/refresh -run '^(TestRefresh|TestSubstrate)' -count=1 -timeout 3m -v failed at cmd/weave/refresh_test.go:44: TestRefreshCLICompilesAndRetries calls compile before initializing its temporary repositories, and rev-parse --show-toplevel returns fatal: not a git repository under GIT_CEILING_DIRECTORIES. Provide passing host evidence for this regression and the required all-weave run at the pinned head, or make the fixture independent of enclosing repository discovery.

## Open findings

- **BR-1** [Important] `user-facing-surface-documentation` README update is missing for weave refresh --substrate
- **BR-2** [Important] `required-regression-verification` (kind: unverifiable) Existing refresh retry regression cannot initialize under the review Git ceiling
