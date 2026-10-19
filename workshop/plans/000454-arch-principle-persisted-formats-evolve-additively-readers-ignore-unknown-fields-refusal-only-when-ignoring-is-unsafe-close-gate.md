---
gate: boundary-review
issue: 454
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 0a68e972-6727-4bfd-806f-ce1d748b597f
            repository: github.com/xianxu/ariadne
            issue_id: "000454"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/efbc0e7c14d1c3e5
            reviewed_head: 27fb3a3534794bd5ca441ae3070c0fb3cd268c61
            slot: ariadne:3
            requested_at: "2026-10-19T04:12:37.041855Z"
          phase: finished
          pid: 14425
          started_at: "2026-10-19T04:12:37.260429Z"
          finished_at: "2026-10-19T04:14:24.431924Z"
          session_id: 01a1525c-865f-7c30-8a40-fe5927825cae
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 27fb3a3534794bd5ca441ae3070c0fb3cd268c61
        tracker_head: 9e289f240a89fddb24b5298fd2295b0e1006cdff
        branch_point: 08fcf21d27f34f22d50ad384b21a77aff5f4b356
        since: "2026-10-18T21:03:58-07:00"
        until: "2026-10-18T21:11:43-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T21:03:58-07:00"
              end: "2026-10-18T21:11:43-07:00"
              began: "2026-10-18T21:03:58-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 3dd96cae2419bcebdb9a21d2232d62b5c0cfc001825b9496544a5b78d6e18ce6
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 24446878
              windows:
                - start: "2026-10-18T21:03:58-07:00"
                  end: "2026-10-18T21:11:43-07:00"
            - harness: codex
              id: 01a15256-9af5-78a1-9881-6f853fa95454
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: c4ab8def9616c073a8785e0b9daf3a4c32daad6993e805c865494c61c53af150
              started: "2026-10-19T04:06:09.998Z"
              spawned: true
              bytes: 514385
              windows:
                - start: "2026-10-19T04:06:09.998Z"
                  end: "2026-10-18T21:11:43-07:00"
      "n": 1
      timestamp: "2026-10-18T21:14:26-07:00"
      dispatched: "2026-10-18T21:12:37-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Two flow tests retain the old registry count and now fail
          detail: cmd/sdlc/closeflow_test.go:109 and cmd/sdlc/flow_e2e_test.go:91 both require “each of the 8 entries”; the rendered full registry now contains nine. Independently running TestCloseQuickCrossingShellUpgrades and TestQuickAndFullFlowsSameVerbSequence reproduced both failures. These are all remaining Go assertions of the old count. Under ARCH-DRY, derive the expected header from the registry or shared renderer while retaining the full-versus-quick selection checks, then rerun both tests.
          family: registry-consumers-derive-from-source
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: Existing atlas descriptions still show the previous registry
          detail: atlas/workflow/architecture-principles.md:9–10 omits EVOLVE from the current quick-flow list; atlas/workflow/sdlc-binary.md:1625–1629 claims exactly one test restates the marker set despite the remaining count assertions and quick-set tripwire. Generated prompt blocks in atlas/process-manual.md:15, :538, :1000, :1362 and :1850 also omit EVOLVE. Update the descriptions and regenerate the manual; these enumerate the stale registry descriptions found in the reviewed tree.
          family: registry-documentation-stays-current
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#454 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T21:14:26-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `registry-consumers-derive-from-source` (kind: defect) Two flow tests retain the old registry count and now fail
  cmd/sdlc/closeflow_test.go:109 and cmd/sdlc/flow_e2e_test.go:91 both require “each of the 8 entries”; the rendered full registry now contains nine. Independently running TestCloseQuickCrossingShellUpgrades and TestQuickAndFullFlowsSameVerbSequence reproduced both failures. These are all remaining Go assertions of the old count. Under ARCH-DRY, derive the expected header from the registry or shared renderer while retaining the full-versus-quick selection checks, then rerun both tests.
- **BR-2** [Minor] `registry-documentation-stays-current` (kind: defect) Existing atlas descriptions still show the previous registry
  atlas/workflow/architecture-principles.md:9–10 omits EVOLVE from the current quick-flow list; atlas/workflow/sdlc-binary.md:1625–1629 claims exactly one test restates the marker set despite the remaining count assertions and quick-set tripwire. Generated prompt blocks in atlas/process-manual.md:15, :538, :1000, :1362 and :1850 also omit EVOLVE. Update the descriptions and regenerate the manual; these enumerate the stale registry descriptions found in the reviewed tree.

## Open findings

- **BR-1** [Important] `registry-consumers-derive-from-source` (kind: defect) Two flow tests retain the old registry count and now fail
- **BR-2** [Minor] `registry-documentation-stays-current` (kind: defect) Existing atlas descriptions still show the previous registry
