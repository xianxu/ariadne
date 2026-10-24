---
gate: boundary-review
issue: 578
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: d6fd26fd-b9ad-4328-8d36-b6f23418964d
            repository: github.com/xianxu/ariadne
            issue_id: "000578"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/d326ff6e41a680dd
            reviewed_head: b8a7d874c2e96c494274c37f5f31c34e08c4b295
            slot: ariadne:2
            requested_at: "2026-10-24T00:48:18.968148Z"
          phase: finished
          pid: 52229
          started_at: "2026-10-24T00:48:19.085396Z"
          finished_at: "2026-10-24T00:50:35.988796Z"
          session_id: 3453467a-8963-403a-a3df-66ee7e7430cf
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: b8a7d874c2e96c494274c37f5f31c34e08c4b295
        tracker_head: 89d0ed56ffc150c725cc3c6aa8c8e4f94c18026d
        branch_point: f6b23cb7a8eb866b080df65c97f92334624b17bd
        since: "2026-10-23T17:24:51-07:00"
        until: "2026-10-23T17:47:50-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T17:24:51-07:00"
              end: "2026-10-23T17:47:50-07:00"
              began: "2026-10-23T17:24:51-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 3f9ae280462613f758523ca38b71d95a21f5b51528387399da62c5345ea308e5
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 14487315
              windows:
                - start: "2026-10-23T17:24:51-07:00"
                  end: "2026-10-23T17:47:50-07:00"
      "n": 1
      timestamp: "2026-10-23T17:50:36-07:00"
      dispatched: "2026-10-23T17:48:18-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: README and atlas still tell agents to run --name suite before close
          detail: README.md:1008 and atlas/workflow/ci-merge-check.md:282 still name `sdlc local-check run --name suite` as the pre-close run, and the same "pre-close suite" wording remains at README.md:654 and atlas/workflow/sdlc-binary.md:432-434 and 2814. These restate the rule instead of naming `sdlc local-check run --pre-close`, so pair workers following the README still run the full suite twice (ARCH-PURPOSE shadow-sweep). Update all five sites together.
          family: pre-close-guidance-single-source
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: README update appears missing for the new local-check run --pre-close flag
          detail: The diff adds the --pre-close flag (cmd/sdlc/localcheck_cmd.go:47), which excludes --name, --all and --boarding, but README.md does not mention it in its local-check section (around lines 779-830) or in the affected paragraph (line 887). The sdlc-binary.md local-check run row (line 101) also omits it.
          family: readme-new-surface
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: --pre-close on a head with no change against main suggests --name suite
          detail: In a train repo, the affected path refuses an empty change with "run --name suite" (localcheck_affected.go:61), which is an odd next step for a --pre-close caller. It rarely happens.
          family: next-action-error-messages
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 26224410-8a10-415f-86a9-1672785289fa
            repository: github.com/xianxu/ariadne
            issue_id: "000578"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/cf269f8cc08d579b
            reviewed_head: 097344071d5a13e9d7544d1b103b1a55e0e77fd3
            slot: ariadne:2
            requested_at: "2026-10-24T00:59:06.87524Z"
          phase: finished
          pid: 1762
          started_at: "2026-10-24T00:59:07.120784Z"
          finished_at: "2026-10-24T01:01:51.332463Z"
          session_id: a9908857-b5ea-421e-a683-918c04994652
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 097344071d5a13e9d7544d1b103b1a55e0e77fd3
        tracker_head: 89d0ed56ffc150c725cc3c6aa8c8e4f94c18026d
        branch_point: f6b23cb7a8eb866b080df65c97f92334624b17bd
        since: "2026-10-23T17:24:51-07:00"
        until: "2026-10-23T17:57:55-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T17:24:51-07:00"
              end: "2026-10-23T17:57:55-07:00"
              began: "2026-10-23T17:24:51-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: aed00bd1128c9d6507e005b31c8f96da99b711ec16f45af61d5ee86e47809220
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 14793692
              windows:
                - start: "2026-10-23T17:24:51-07:00"
                  end: "2026-10-23T17:57:55-07:00"
        cuts:
            - dispatched: "2026-10-23T17:48:18-07:00"
              finalized: "2026-10-23T17:50:36-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-23T18:01:53-07:00"
      dispatched: "2026-10-23T17:59:06-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'All five sites now name --pre-close: README.md:654 and 1013, ci-merge-check.md:281-282, sdlc-binary.md:432-434 and 2814. The remaining --name suite hits are the boarding alternative, bootstrap, resume, republish and the legacy merge path, none of them pre-close.'
          round: 2
        - id: BR-2
          disposition: addressed
          note: README.md:900-903 documents --pre-close (resolution, unreadable fallback, exclusivity) and the sdlc-binary.md:101 local-check run row names it; embedded help documents it too.
          round: 2
        - id: BR-3
          disposition: withdrawn
          note: With no change against main there is nothing affected to vouch for, so the suite is a correct next step for a --pre-close caller too (rationale in the issue Log).
          round: 2
      recipe: small-diff-review
      isolation: fenced
      reviewed: 097344071d5a13e9d7544d1b103b1a55e0e77fd3
      blocked: false
---

# Gate ledger — ariadne#578 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T17:50:36-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `pre-close-guidance-single-source` (kind: defect) README and atlas still tell agents to run --name suite before close
  README.md:1008 and atlas/workflow/ci-merge-check.md:282 still name `sdlc local-check run --name suite` as the pre-close run, and the same "pre-close suite" wording remains at README.md:654 and atlas/workflow/sdlc-binary.md:432-434 and 2814. These restate the rule instead of naming `sdlc local-check run --pre-close`, so pair workers following the README still run the full suite twice (ARCH-PURPOSE shadow-sweep). Update all five sites together.
- **BR-2** [Important] `readme-new-surface` (kind: defect) README update appears missing for the new local-check run --pre-close flag
  The diff adds the --pre-close flag (cmd/sdlc/localcheck_cmd.go:47), which excludes --name, --all and --boarding, but README.md does not mention it in its local-check section (around lines 779-830) or in the affected paragraph (line 887). The sdlc-binary.md local-check run row (line 101) also omits it.
- **BR-3** [Minor] `next-action-error-messages` (kind: defect) --pre-close on a head with no change against main suggests --name suite
  In a train repo, the affected path refuses an empty change with "run --name suite" (localcheck_affected.go:61), which is an odd next step for a --pre-close caller. It rarely happens.

## Round 2 — 2026-10-23T18:01:53-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — All five sites now name --pre-close: README.md:654 and 1013, ci-merge-check.md:281-282, sdlc-binary.md:432-434 and 2814. The remaining --name suite hits are the boarding alternative, bootstrap, resume, republish and the legacy merge path, none of them pre-close.
- BR-2 — addressed — README.md:900-903 documents --pre-close (resolution, unreadable fallback, exclusivity) and the sdlc-binary.md:101 local-check run row names it; embedded help documents it too.
- BR-3 — withdrawn — With no change against main there is nothing affected to vouch for, so the suite is a correct next step for a --pre-close caller too (rationale in the issue Log).

## Open findings

(none — every finding has been disposed)
