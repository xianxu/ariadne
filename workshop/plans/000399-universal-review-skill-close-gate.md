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
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 438c7434-a49d-483b-b65d-c6bcc448e293
            repository: github.com/xianxu/ariadne
            issue_id: "000399"
            recipe: fix-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/da2ac20b4554708e
            reviewed_head: e40d16b5fc230bf36f13c69415f21905f96d0698
            slot: ariadne:2
            requested_at: "2026-10-16T00:05:13.757063Z"
          phase: finished
          pid: 36302
          started_at: "2026-10-16T00:05:14.072286Z"
          finished_at: "2026-10-16T00:06:07.030111Z"
          session_id: c387247f-a9f9-448b-a918-d3bbea4ff4e7
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: e40d16b5fc230bf36f13c69415f21905f96d0698
        tracker_head: 09080a42ed142ce08345755f4f9bb6b0d473cfe3
        branch_point: f3bc999f925ab5137d1de82c6b12c3511d746f83
        since: "2026-10-15T16:24:58-07:00"
        until: "2026-10-15T17:04:26-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T16:24:58-07:00"
              end: "2026-10-15T17:04:26-07:00"
              began: "2026-10-15T16:24:58-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: cf555fb5711e937547ebe86e137e0c5ae850dd332e9734d05bb313e776572dab
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 9847902
              windows:
                - start: "2026-10-15T16:24:58-07:00"
                  end: "2026-10-15T17:04:26-07:00"
        cuts:
            - dispatched: "2026-10-15T16:57:15-07:00"
              finalized: "2026-10-15T17:03:29-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-15T17:06:17-07:00"
      dispatched: "2026-10-15T17:05:13-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: round.md:33 now says "list the known modes (the files in `modes/`, by `order`)"; the hand-kept list is removed.
          round: 2
        - id: BR-2
          disposition: addressed
          note: file-edit.md step 4 now says "Then delete BODYFILE, whether the write succeeded or was refused."
          round: 2
        - id: BR-3
          disposition: addressed
          note: export.go:86 handles the empty block. Tests crlf, quoted-value-literal and empty-frontmatter pin the edges, and the doc comment records the CRLF and literal-value behavior. The empty-frontmatter test fails without the fix.
          round: 2
        - id: BR-4
          disposition: addressed
          note: A Revisions entry records TL decision f73c1f3a, moving the live Alt+Return check to parley.nvim#332's combined acceptance.
          round: 2
      findings:
        - kind: defect
          id: BR-5
          severity: Important
          title: 'Scope changed after close: run a full `sdlc close`'
          detail: The interdiff adds operator requirement 5ef50ef6, which no prior finding asked for. It adds a second Revisions entry, the base-copy diff workflow in adapters/file-edit.md (the Untracked section), and a new "never your own edits" contract in round.md section 1. This is new behavior after close, outside the narrow fix review.
          family: scope-change-after-close
          round: 2
        - kind: defect
          id: BR-6
          severity: Minor
          title: The untracked adapter's last review-base copy is never deleted
          detail: 'file-edit.md:70-72 deletes each review-base.XXXXXX copy only when the next one replaces it, so the last copy of every conversation stays in TMPDIR. This is the 2nd finding in this family. Rule: every temp file an adapter creates must state when it is deleted, including the end of the conversation or session. Apply that rule to the whole adapter, not just this site.'
          family: artifact-without-removal-path
          round: 2
      recipe: fix-review
      isolation: fenced
      blocked: true
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

## Round 2 — 2026-10-15T17:06:17-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — round.md:33 now says "list the known modes (the files in `modes/`, by `order`)"; the hand-kept list is removed.
- BR-2 — addressed — file-edit.md step 4 now says "Then delete BODYFILE, whether the write succeeded or was refused."
- BR-3 — addressed — export.go:86 handles the empty block. Tests crlf, quoted-value-literal and empty-frontmatter pin the edges, and the doc comment records the CRLF and literal-value behavior. The empty-frontmatter test fails without the fix.
- BR-4 — addressed — A Revisions entry records TL decision f73c1f3a, moving the live Alt+Return check to parley.nvim#332's combined acceptance.

### Raised

- **BR-5** [Important] `scope-change-after-close` (kind: defect) Scope changed after close: run a full `sdlc close`
  The interdiff adds operator requirement 5ef50ef6, which no prior finding asked for. It adds a second Revisions entry, the base-copy diff workflow in adapters/file-edit.md (the Untracked section), and a new "never your own edits" contract in round.md section 1. This is new behavior after close, outside the narrow fix review.
- **BR-6** [Minor] `artifact-without-removal-path` (kind: defect) The untracked adapter's last review-base copy is never deleted
  file-edit.md:70-72 deletes each review-base.XXXXXX copy only when the next one replaces it, so the last copy of every conversation stays in TMPDIR. This is the 2nd finding in this family. Rule: every temp file an adapter creates must state when it is deleted, including the end of the conversation or session. Apply that rule to the whole adapter, not just this site.

## Open findings

- **BR-5** [Important] `scope-change-after-close` (kind: defect) Scope changed after close: run a full `sdlc close`
- **BR-6** [Minor] `artifact-without-removal-path` (kind: defect) The untracked adapter's last review-base copy is never deleted
