---
gate: boundary-review
issue: 562
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: dd728520-2827-437b-8bdb-1948128a8376
            repository: github.com/xianxu/ariadne
            issue_id: "000562"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/35c1e181a342a6ee
            reviewed_head: fef3c6b67b0878d6d61e5d6eff545bc2cf2d94fa
            slot: ariadne:5
            requested_at: "2026-10-23T06:14:38.216534Z"
          phase: finished
          pid: 54799
          started_at: "2026-10-23T06:14:38.313337Z"
          finished_at: "2026-10-23T06:17:39.619899Z"
          session_id: 01a16765-ac1e-7412-86bc-f3eef0948a93
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: fef3c6b67b0878d6d61e5d6eff545bc2cf2d94fa
        tracker_head: f600ad336b0980b057ee8dca5955d800b70c61ba
        branch_point: ada6c156d263d1ad3376e44147dfdacfe2c4d90e
        since: "2026-10-22T22:48:36-07:00"
        until: "2026-10-22T23:14:20-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T22:48:36-07:00"
              end: "2026-10-22T23:14:20-07:00"
              began: "2026-10-22T22:48:36-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: c0398bdb0dc9a85a106cdb189e3d903d5cd4eedf6ab7530d2eab0aad3e5b4cf6
              started: "2026-10-20T15:58:37.432Z"
              spawned: false
              bytes: 62928683
              windows:
                - start: "2026-10-22T22:48:36-07:00"
                  end: "2026-10-22T23:14:20-07:00"
            - harness: codex
              id: 01a16753-7aad-7100-bc61-d5aff38d5ed5
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              parent_id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              digest: 92e73301ebf57c2844eec0efb0233b19aab84e4c2759e281419f169d6edb5e68
              started: "2026-10-23T05:54:46.787Z"
              spawned: true
              bytes: 259781
              windows:
                - start: "2026-10-23T05:54:46.787Z"
                  end: "2026-10-22T23:14:20-07:00"
            - harness: codex
              id: 01a16755-5662-76f1-9b02-bb5e27c53deb
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              parent_id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              digest: 880b493d6bc94c41c6f4a1c15ca272bfaf968fdac9dc477ad79c3a6601967bf5
              started: "2026-10-23T05:56:48.474Z"
              spawned: true
              bytes: 441837
              windows:
                - start: "2026-10-23T05:56:48.474Z"
                  end: "2026-10-22T23:14:20-07:00"
            - harness: codex
              id: 01a16760-20e9-7421-a24e-09629f5d9ebf
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              parent_id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              digest: e724d4cf22df68eb1767ed3c2d45763e6f112791fe2d10bc40f60e24f8a157c4
              started: "2026-10-23T06:08:35.748Z"
              spawned: true
              bytes: 290575
              windows:
                - start: "2026-10-23T06:08:35.748Z"
                  end: "2026-10-22T23:14:20-07:00"
      "n": 1
      timestamp: "2026-10-22T23:17:40-07:00"
      dispatched: "2026-10-22T23:14:38-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Active consumer instructions still prescribe the retired creation contract
          detail: 'ARCH-DRY / ARCH-PURPOSE: collaboration-modes.md:241 requires every new side doc to use YYYY-MM/YYYY-MM-DD-NN-slug.md, but its routing rule at :146 and state-table rows at :401 and :410 still explicitly create flat paths. The remaining active prompt instances are construct/local/fix/SKILL.md:87 (flat path and same-flat-folder instruction), construct/local/fix/review-convention.md:55 (same flat folder), and cmd/tl-lead/help.md:436-442 (flat answer paths and incomplete written defaults); the TL session-creation instruction at :340 also still emits the retired collaboration-mode map. These are executable instructions, not the explicitly grandfathered examples. The stale grammar is exported alongside the new contract. Sweep these consumers into agreement, preferably by referencing the shared contract, regenerate the export, and exercise creation through the complete prompt composition. Coordinate the TL naming instructions with cmd/tl-lead/answers.go:24, which currently rejects month-directory targets.'
          family: shared-contract-consumer-consistency
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Fill conformance examples do not exercise preservation of existing collaboration choices or instructions
          detail: 'pkg/reviewexport/editorial_fill_test.go:17-19 covers empty, partially canonical, and legacy-only frontmatter, but :36 hardcodes collaboration-mode: discussion for every result. Across all three fixture families, no input has an existing collaboration-mode, and no input has an existing canonical processing-instruction. Consequently the preservation loop at :43 never tests those two fields, although the contract says existing values remain untouched. Add cases preserving a nondefault collaboration-mode and existing processing-instruction while missing mode keys are filled, and exercise them in the contract scenarios. These are the two uncovered field instances in the new preservation fixtures.'
          family: preservation-conformance-all-fields
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#562 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-22T23:17:40-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `shared-contract-consumer-consistency` (kind: defect) Active consumer instructions still prescribe the retired creation contract
  ARCH-DRY / ARCH-PURPOSE: collaboration-modes.md:241 requires every new side doc to use YYYY-MM/YYYY-MM-DD-NN-slug.md, but its routing rule at :146 and state-table rows at :401 and :410 still explicitly create flat paths. The remaining active prompt instances are construct/local/fix/SKILL.md:87 (flat path and same-flat-folder instruction), construct/local/fix/review-convention.md:55 (same flat folder), and cmd/tl-lead/help.md:436-442 (flat answer paths and incomplete written defaults); the TL session-creation instruction at :340 also still emits the retired collaboration-mode map. These are executable instructions, not the explicitly grandfathered examples. The stale grammar is exported alongside the new contract. Sweep these consumers into agreement, preferably by referencing the shared contract, regenerate the export, and exercise creation through the complete prompt composition. Coordinate the TL naming instructions with cmd/tl-lead/answers.go:24, which currently rejects month-directory targets.
- **BR-2** [Important] `preservation-conformance-all-fields` (kind: defect) Fill conformance examples do not exercise preservation of existing collaboration choices or instructions
  pkg/reviewexport/editorial_fill_test.go:17-19 covers empty, partially canonical, and legacy-only frontmatter, but :36 hardcodes collaboration-mode: discussion for every result. Across all three fixture families, no input has an existing collaboration-mode, and no input has an existing canonical processing-instruction. Consequently the preservation loop at :43 never tests those two fields, although the contract says existing values remain untouched. Add cases preserving a nondefault collaboration-mode and existing processing-instruction while missing mode keys are filled, and exercise them in the contract scenarios. These are the two uncovered field instances in the new preservation fixtures.

## Open findings

- **BR-1** [Critical] `shared-contract-consumer-consistency` (kind: defect) Active consumer instructions still prescribe the retired creation contract
- **BR-2** [Important] `preservation-conformance-all-fields` (kind: defect) Fill conformance examples do not exercise preservation of existing collaboration choices or instructions
