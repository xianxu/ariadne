---
gate: boundary-review
issue: 444
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: eaeafdcc-5e85-4baa-9beb-f422f4cc07c6
            repository: github.com/xianxu/ariadne
            issue_id: "000444"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/0152b5d1bc9a9d28
            reviewed_head: 89d3effe978c89d08f872c9b78425ad53f4d1d40
            slot: ariadne:1
            requested_at: "2026-10-18T22:18:56.338672Z"
          phase: finished
          pid: 83495
          started_at: "2026-10-18T22:18:56.432008Z"
          finished_at: "2026-10-18T22:20:32.166272Z"
          session_id: 01a15118-b725-7131-9a8c-4c6a1525342b
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 89d3effe978c89d08f872c9b78425ad53f4d1d40
        tracker_head: a7f85cd75ad1cd30b02cc3eb5ec735939fea8f62
        branch_point: cf565ef7914b53b0485a587586f4b094643cf13f
        since: "2026-10-18T14:37:53-07:00"
        until: "2026-10-18T15:18:20-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T14:37:53-07:00"
              end: "2026-10-18T15:18:20-07:00"
              began: "2026-10-18T14:37:53-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: ba60d072121d94dd0183248b3ab828354e04b5c03faa9cbc842f487defcb469e
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 122229295
              windows:
                - start: "2026-10-18T14:37:53-07:00"
                  end: "2026-10-18T15:18:20-07:00"
        queue_waits:
            - attempt_id: 3590fe8eed4602bfab53764dda88e258
              slot: ariadne:1
              record_digest: 93e77f2027052a9bf6da90fee0a07c334c6958f416065a6d439bf49d21af2bf1
              start: "2026-10-18T21:46:29.803654Z"
              end: "2026-10-18T22:08:43.529371Z"
              queue_seconds: 1333.725717
              raw_seconds: 1333.725717
              exact: true
      "n": 1
      timestamp: "2026-10-18T15:20:32-07:00"
      dispatched: "2026-10-18T15:18:56-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: The new fixture guard rejects unchanged legacy chains after a file move
          detail: "pkg/reviewexport/examples_test.go:260–264 compares each after-file only with the identical before-path. A valid move from design.md to spec.md containing an unchanged \U0001F916[ask]{reply} therefore reports “agent wrote a second turn” because beforeText[\"spec.md\"] is empty. The complete affected family is the renamed parent document, side docs moved with its folder, and destination orphan side docs displaced by the move contract. Validate move cases through their existing exact projectMove projection before applying the emission guard, or compare against a projected baseline. Add positive move regressions alongside the existing append-rejection tests."
          family: legacy-preservation-across-moves
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: README update appears missing for the changed marker convention
          detail: construct/local/fix/review-convention.md:45–55 changes the syntax users write and the clarification workflow, but README.md is unchanged in the pinned range. The affected surface comprises the one-commentary-block limit, continued exchanges opening linked and indexed side docs, and legacy-chain compatibility. Add a short explanation beside README.md:23–28’s existing universal-review usage paragraph, with a link to the canonical convention.
          family: user-facing-convention-documentation
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: cf75b681-91d0-43af-b5f3-ec87efa88208
            repository: github.com/xianxu/ariadne
            issue_id: "000444"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/fdd348b5f46af8b0
            reviewed_head: def348ccd85a7abe8b9255a048023e45369550a0
            slot: ariadne:1
            requested_at: "2026-10-18T22:50:32.341002Z"
          phase: finished
          pid: 88058
          started_at: "2026-10-18T22:50:32.491154Z"
          finished_at: "2026-10-18T22:52:22.985217Z"
          session_id: 01a15135-a62c-7d13-88bd-b52ac9fc9ac6
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: def348ccd85a7abe8b9255a048023e45369550a0
        tracker_head: 823a55ac9822042d836229a66d1f45c2c3526a04
        branch_point: cf565ef7914b53b0485a587586f4b094643cf13f
        since: "2026-10-18T14:37:53-07:00"
        until: "2026-10-18T15:22:00-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T14:37:53-07:00"
              end: "2026-10-18T15:22:00-07:00"
              began: "2026-10-18T14:37:53-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 9bebe7733a10dd4dcb1f3b0e7bbacc3eca403a87d3cbdf0fdad374b466eb30ec
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 122863272
              windows:
                - start: "2026-10-18T14:37:53-07:00"
                  end: "2026-10-18T15:22:00-07:00"
        queue_waits:
            - attempt_id: 3590fe8eed4602bfab53764dda88e258
              slot: ariadne:1
              record_digest: 93e77f2027052a9bf6da90fee0a07c334c6958f416065a6d439bf49d21af2bf1
              start: "2026-10-18T21:46:29.803654Z"
              end: "2026-10-18T22:08:43.529371Z"
              queue_seconds: 1333.725717
              raw_seconds: 1333.725717
              exact: true
        cuts:
            - dispatched: "2026-10-18T15:18:56-07:00"
              finalized: "2026-10-18T15:20:32-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-18T15:52:24-07:00"
      dispatched: "2026-10-18T15:50:32-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: examples_test.go:257–292 checks moves against projectMove before applying the emission guard. TestMovePreservesLegacyChains covers the renamed parent, both moved side docs, and displaced orphan, plus rejection of an appended turn. Moving the guard back before projection made this regression fail on all four files; restoring the fix passed.
          round: 2
        - id: BR-2
          disposition: addressed
          note: README.md:29–32 now documents “at most one” commentary turn, indexed discussion side docs, and legacy compatibility, linking the canonical convention. These statements match review-convention.md:45–55 and round.md:99–119.
          round: 2
      recipe: small-diff-review
      isolation: fenced
      reviewed: def348ccd85a7abe8b9255a048023e45369550a0
      blocked: false
---

# Gate ledger — ariadne#444 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T15:20:32-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `legacy-preservation-across-moves` (kind: defect) The new fixture guard rejects unchanged legacy chains after a file move
  pkg/reviewexport/examples_test.go:260–264 compares each after-file only with the identical before-path. A valid move from design.md to spec.md containing an unchanged 🤖[ask]{reply} therefore reports “agent wrote a second turn” because beforeText["spec.md"] is empty. The complete affected family is the renamed parent document, side docs moved with its folder, and destination orphan side docs displaced by the move contract. Validate move cases through their existing exact projectMove projection before applying the emission guard, or compare against a projected baseline. Add positive move regressions alongside the existing append-rejection tests.
- **BR-2** [Important] `user-facing-convention-documentation` (kind: defect) README update appears missing for the changed marker convention
  construct/local/fix/review-convention.md:45–55 changes the syntax users write and the clarification workflow, but README.md is unchanged in the pinned range. The affected surface comprises the one-commentary-block limit, continued exchanges opening linked and indexed side docs, and legacy-chain compatibility. Add a short explanation beside README.md:23–28’s existing universal-review usage paragraph, with a link to the canonical convention.

## Round 2 — 2026-10-18T15:52:24-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — examples_test.go:257–292 checks moves against projectMove before applying the emission guard. TestMovePreservesLegacyChains covers the renamed parent, both moved side docs, and displaced orphan, plus rejection of an appended turn. Moving the guard back before projection made this regression fail on all four files; restoring the fix passed.
- BR-2 — addressed — README.md:29–32 now documents “at most one” commentary turn, indexed discussion side docs, and legacy compatibility, linking the canonical convention. These statements match review-convention.md:45–55 and round.md:99–119.

## Open findings

(none — every finding has been disposed)
