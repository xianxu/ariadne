---
gate: boundary-review
issue: 399
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 4e42f121-e49f-47ce-bd4f-3f8e940ddb1a
            repository: github.com/xianxu/ariadne
            issue_id: "000399"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/81476eaa8f017e33
            reviewed_head: 087c8af70c4b394008aadbdbeaf7aba5e79b1169
            slot: ariadne:2
            requested_at: "2026-10-15T23:57:15.441874Z"
          phase: finished
          pid: 70098
          started_at: "2026-10-15T23:57:15.62149Z"
          finished_at: "2026-10-16T00:03:08.491485Z"
          session_id: 313ec06c-a55e-4137-9337-cbed3fffea00
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 087c8af70c4b394008aadbdbeaf7aba5e79b1169
        tracker_head: 39b63d22001712af25d258cc154af0e78048cd74
        branch_point: f3bc999f925ab5137d1de82c6b12c3511d746f83
        since: "2026-10-15T16:24:58-07:00"
        until: "2026-10-15T16:56:39-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T16:24:58-07:00"
              end: "2026-10-15T16:56:39-07:00"
              began: "2026-10-15T16:24:58-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: b6962a1b559f6db4b8ea1e5c2e3a66ef7c2ec9955d3a03a3496a21e8354b79cd
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 9742425
              windows:
                - start: "2026-10-15T16:24:58-07:00"
                  end: "2026-10-15T16:56:39-07:00"
      "n": 1
      timestamp: "2026-10-15T17:03:29-07:00"
      dispatched: "2026-10-15T16:57:15-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: round.md hard-codes the known-mode list, which repeats modes/ with no test guarding it
          detail: round.md:32-33 lists developmental, line-editing, copy-editing, proofreading, fact-check and free-form by hand (ARCH-DRY). Either tell the agent to list modes/, or assert in TestExportCurrent that the list equals the manifest's mode members.
          family: hand-maintained-restatement-of-source
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: The file-edit adapter creates a mktemp body file on every tracked round and never removes it
          detail: adapters/file-edit.md step 4 creates review.XXXXXX under TMPDIR for each round, and nothing removes it (ARCH-FUNERAL). Add an instruction to delete BODYFILE after the write, whether it succeeds or is refused.
          family: artifact-without-removal-path
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: ParseMode's parity with parley's parser is only tested on happy-path frontmatter
          detail: pkg/reviewexport/export.go ParseMode doesn't handle CRLF line endings or quoted values, and it reports an empty frontmatter block as unterminated. Pin these against parley's behavior, or document them as out of scope.
          family: parser-parity-untested-edges
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: The Done-when live operator check is still claimed, though the TL moved it to parley.nvim#332
          detail: Add a Revisions entry recording TL decision f73c1f3a, so the Plan and Done-when no longer claim this issue delivers the live Alt+Return check.
          family: plan-claims-undelivered-item
          round: 1
      recipe: milestone-review
      isolation: fenced
      reviewed: 087c8af70c4b394008aadbdbeaf7aba5e79b1169
      blocked: false
---

# Gate ledger — ariadne#399 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-15T17:03:29-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `hand-maintained-restatement-of-source` (kind: defect) round.md hard-codes the known-mode list, which repeats modes/ with no test guarding it
  round.md:32-33 lists developmental, line-editing, copy-editing, proofreading, fact-check and free-form by hand (ARCH-DRY). Either tell the agent to list modes/, or assert in TestExportCurrent that the list equals the manifest's mode members.
- **BR-2** [Minor] `artifact-without-removal-path` (kind: defect) The file-edit adapter creates a mktemp body file on every tracked round and never removes it
  adapters/file-edit.md step 4 creates review.XXXXXX under TMPDIR for each round, and nothing removes it (ARCH-FUNERAL). Add an instruction to delete BODYFILE after the write, whether it succeeds or is refused.
- **BR-3** [Minor] `parser-parity-untested-edges` (kind: defect) ParseMode's parity with parley's parser is only tested on happy-path frontmatter
  pkg/reviewexport/export.go ParseMode doesn't handle CRLF line endings or quoted values, and it reports an empty frontmatter block as unterminated. Pin these against parley's behavior, or document them as out of scope.
- **BR-4** [Minor] `plan-claims-undelivered-item` (kind: defect) The Done-when live operator check is still claimed, though the TL moved it to parley.nvim#332
  Add a Revisions entry recording TL decision f73c1f3a, so the Plan and Done-when no longer claim this issue delivers the live Alt+Return check.

## Open findings

- **BR-1** [Minor] `hand-maintained-restatement-of-source` (kind: defect) round.md hard-codes the known-mode list, which repeats modes/ with no test guarding it
- **BR-2** [Minor] `artifact-without-removal-path` (kind: defect) The file-edit adapter creates a mktemp body file on every tracked round and never removes it
- **BR-3** [Minor] `parser-parity-untested-edges` (kind: defect) ParseMode's parity with parley's parser is only tested on happy-path frontmatter
- **BR-4** [Minor] `plan-claims-undelivered-item` (kind: defect) The Done-when live operator check is still claimed, though the TL moved it to parley.nvim#332
