---
reviewer_runs:
    - version: 1
      intent:
        id: ed02d1bd-f71d-4820-a93d-ae72b5d0c924
        repository: github.com/xianxu/ariadne
        issue_id: "000455"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: claude
        root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        reviewed_head: 0a7c2c86ba208f322fefababe2e69279a36b8331
        slot: ariadne:4
        requested_at: "2026-10-19T04:25:16.34228Z"
      phase: finished
      pid: 2111
      started_at: "2026-10-19T04:25:16.702451Z"
      finished_at: "2026-10-19T04:26:07.807369Z"
      session_id: 3b2498aa-31d7-48a3-972b-3450f04dfcf6
      outcome: returned
gate: plan-quality
issue: 455
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 9531329b-98b6-40e6-8da5-bdc8f2ccae7c
            repository: github.com/xianxu/ariadne
            issue_id: "000455"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: 57dacf8acbc9e60a6a533102770789d03092a491
            slot: ariadne:4
            requested_at: "2026-10-19T04:20:04.958766Z"
          phase: finished
          pid: 7870
          started_at: "2026-10-19T04:20:05.191105Z"
          finished_at: "2026-10-19T04:22:26.802964Z"
          session_id: 84ae15ce-0163-488d-8b48-b13dd1c4a2ac
          outcome: returned
      "n": 1
      timestamp: "2026-10-18T21:22:27-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: PQ-1
          severity: Important
          title: Migration walk would move Couch lock files and harness dirs; survey missed them
          detail: 'ops qa/xian-xu holds 2026-10-18-22-load-and-echo.md.lock (a live operator lock) and .claude/.cc-writes. Decision 7 moves every non-md file, and lock conflicts are detected only via registrations. Classify first: *.md.lock is always a lock conflict and never moved; dot-entries are excluded and listed. Then revise the "qa ends empty" claim.'
          family: migration-input-classification
          round: 1
        - kind: defect
          id: PQ-2
          severity: Important
          title: Rerun after files moved cannot rebuild session/outside link edits; moved() must be a pure path map
          detail: Files only at their destination are "done (not listed)", so a snapshot-derived moved() map loses them, and the session-file links then read as dangling and are left as written. That contradicts the state table's "rerun rebuilds edits". Derive moved() from the path (qa/<op>/rel to <stem>.side/rel via MoveMember), confirmed by the destination existing. Also pin which inputs the planned bytes depend on, so a bullet deleted between runs (marker rule 2) does not turn an already-created destination into a false bytes-differ conflict.
          family: resume-plan-derivation
          round: 1
        - kind: defect
          id: PQ-3
          severity: Minor
          title: Tasks 1-4 list test cases in prose; replace with strategy lines, and fuzz RebaseLinks
          detail: 'RebaseLinks runs over arbitrary Markdown: fuzz it seeded with fences, code spans, CRLF and parens, asserting identity when nothing moved and that every rewritten target resolves. FuzzReconcileIndex is the precedent.'
          family: test-prose-enumeration
          round: 1
        - kind: defect
          id: PQ-4
          severity: Minor
          title: Reuse sidedoc.MoveMember and the SideLinks fence scanner instead of re-implementing them
          detail: MoveMember (pkg/sidedoc/tree.go:17) already maps a side-tree member across a move; links.go:28-45 owns fence detection.
          family: reuse-sidedoc-helpers
          round: 1
        - kind: defect
          id: PQ-5
          severity: Minor
          title: help.md says answer docs carry no process-instruction; migrated ones keep it (decision 3)
          detail: The Task 6 help edit should state this exception, and answerTarget must not reject docs that carry the key.
          family: help-contract-drift
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: 254073e1-7dc5-4fc3-8f54-497b82c5d987
            repository: github.com/xianxu/ariadne
            issue_id: "000455"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: 0a7c2c86ba208f322fefababe2e69279a36b8331
            slot: ariadne:4
            requested_at: "2026-10-19T04:23:17.577102Z"
          phase: finished
          pid: 27757
          started_at: "2026-10-19T04:23:17.855549Z"
          finished_at: "2026-10-19T04:23:49.661105Z"
          session_id: 40102923-8652-4e3a-8bc3-f4e04f75585e
          outcome: returned
      "n": 2
      timestamp: "2026-10-18T21:23:50-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: 'classifyEntry table: dot-segments foreign, *.md.lock and *.md.clobbered conflicts, symlinks conflicts; decision 7 now conditional.'
          round: 2
        - id: PQ-2
          disposition: addressed
          note: moved() is a pure qa-root prefix map; resume reads marker from destination frontmatter; crash test deletes a bullet between runs.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: FuzzRebaseLinks added with fence, code span, CRLF and paren seeds; prose cases demoted to strategy.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Fence scan extracted from SideLinks (links.go:28-45) and shared; MoveMember shape (tree.go:17) generalised.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: help.md exception stated; answerTarget accepts process-instruction, with a test.
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: 19db5b42-9349-4232-9dcc-9ba067bd02fe
            repository: github.com/xianxu/ariadne
            issue_id: "000455"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: 0a7c2c86ba208f322fefababe2e69279a36b8331
            slot: ariadne:4
            requested_at: "2026-10-19T04:24:54.845031Z"
          phase: finished
          pid: 90308
          started_at: "2026-10-19T04:24:55.114719Z"
          finished_at: "2026-10-19T04:25:12.546834Z"
          session_id: 3b8f552c-eeb9-4227-91af-2c9a3a636fd2
          outcome: returned
      "n": 3
      timestamp: "2026-10-18T21:25:13-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      blocked: false
      protocol_error: no valid findings block
content_hash: 49026973af7b5db465b8b23055911f22e117d39883270c59179f78b82790851a
---

# Gate ledger — ariadne#455 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T21:22:27-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `migration-input-classification` (kind: defect) Migration walk would move Couch lock files and harness dirs; survey missed them
  ops qa/xian-xu holds 2026-10-18-22-load-and-echo.md.lock (a live operator lock) and .claude/.cc-writes. Decision 7 moves every non-md file, and lock conflicts are detected only via registrations. Classify first: *.md.lock is always a lock conflict and never moved; dot-entries are excluded and listed. Then revise the "qa ends empty" claim.
- **PQ-2** [Important] `resume-plan-derivation` (kind: defect) Rerun after files moved cannot rebuild session/outside link edits; moved() must be a pure path map
  Files only at their destination are "done (not listed)", so a snapshot-derived moved() map loses them, and the session-file links then read as dangling and are left as written. That contradicts the state table's "rerun rebuilds edits". Derive moved() from the path (qa/<op>/rel to <stem>.side/rel via MoveMember), confirmed by the destination existing. Also pin which inputs the planned bytes depend on, so a bullet deleted between runs (marker rule 2) does not turn an already-created destination into a false bytes-differ conflict.
- **PQ-3** [Minor] `test-prose-enumeration` (kind: defect) Tasks 1-4 list test cases in prose; replace with strategy lines, and fuzz RebaseLinks
  RebaseLinks runs over arbitrary Markdown: fuzz it seeded with fences, code spans, CRLF and parens, asserting identity when nothing moved and that every rewritten target resolves. FuzzReconcileIndex is the precedent.
- **PQ-4** [Minor] `reuse-sidedoc-helpers` (kind: defect) Reuse sidedoc.MoveMember and the SideLinks fence scanner instead of re-implementing them
  MoveMember (pkg/sidedoc/tree.go:17) already maps a side-tree member across a move; links.go:28-45 owns fence detection.
- **PQ-5** [Minor] `help-contract-drift` (kind: defect) help.md says answer docs carry no process-instruction; migrated ones keep it (decision 3)
  The Task 6 help edit should state this exception, and answerTarget must not reject docs that carry the key.

## Round 2 — 2026-10-18T21:23:50-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — classifyEntry table: dot-segments foreign, *.md.lock and *.md.clobbered conflicts, symlinks conflicts; decision 7 now conditional.
- PQ-2 — addressed — moved() is a pure qa-root prefix map; resume reads marker from destination frontmatter; crash test deletes a bullet between runs.
- PQ-3 — addressed — FuzzRebaseLinks added with fence, code span, CRLF and paren seeds; prose cases demoted to strategy.
- PQ-4 — addressed — Fence scan extracted from SideLinks (links.go:28-45) and shared; MoveMember shape (tree.go:17) generalised.
- PQ-5 — addressed — help.md exception stated; answerTarget accepts process-instruction, with a test.

## Round 3 — 2026-10-18T21:25:13-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Open findings

(none — every finding has been disposed)
