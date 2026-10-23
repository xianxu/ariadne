---
gate: boundary-review
issue: 575
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 985e0dd7-a2f5-48ad-a9a9-c0669fd0e544
            repository: github.com/xianxu/ariadne
            issue_id: "000575"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/c4e5cd8ddb547d5b
            reviewed_head: 10bd1d202c13536678cf5e194b733b002c5fcae5
            slot: ariadne:2
            requested_at: "2026-10-23T19:07:55.921707Z"
          phase: finished
          pid: 48460
          started_at: "2026-10-23T19:07:56.080757Z"
          finished_at: "2026-10-23T19:11:32.750156Z"
          session_id: e1cb1cda-78ca-4e26-8f83-555fd09b50a1
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 10bd1d202c13536678cf5e194b733b002c5fcae5
        tracker_head: 7d6eb4f6643322154a6495910fb09412c6bb893d
        branch_point: 588a0b0af09a8965d472f7acc216581db3afa815
        since: "2026-10-23T11:52:48-07:00"
        until: "2026-10-23T12:07:16-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T11:52:48-07:00"
              end: "2026-10-23T12:07:16-07:00"
              began: "2026-10-23T11:52:48-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c4d935f3c64c19106d3c488b1829259e2ac7098ccffebab1770bd984c41a31ea
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 9229028
              windows:
                - start: "2026-10-23T11:52:48-07:00"
                  end: "2026-10-23T12:07:16-07:00"
      "n": 1
      timestamp: "2026-10-23T12:11:33-07:00"
      dispatched: "2026-10-23T12:07:55-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Spec and ChatSections doc claim setext headings, but a tag above a setext heading is always misplaced
          detail: collaboration-modes.md:107 and chat.go:53 say "ATX or setext"; CommonMark folds the tag line into the setext heading, so ChatSections reports it misplaced. Plan said ATX only. Fix the prose to ATX and pin setext as misplaced in chat_test.go.
          family: chat-grammar-oracle-parity
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Indented-code @@chat@@ directly above a heading makes the section chat
          detail: chat.go:120 checks the previous line's trimmed text without the CodeLiteralMask used at chat.go:108; a probe of "para, blank, 4-space-indented tag, then a heading" returns Source tag. Check the mask there and add a test row.
          family: chat-grammar-oracle-parity
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: Reply notices are not cross-checked with ChatSections and the exported fixtures already diverge
          detail: Spec section 7 says replies word notices as ChatSections does, but checkChat ignores scan.Notices. The legacy-chat-comment reply cites line 11 while the oracle reports 12, and the scanner emits several notices where the spec says one line. Assert notices in checkChat, fix the fixture, and state which document the line numbers refer to.
          family: fixture-oracle-crosscheck
          round: 1
        - kind: defect
          id: BR-4
          severity: Important
          title: 'Plan Core concepts and Task 5 no longer match the code: ChatSections location and guard scope'
          detail: The table lists pkg/reviewexport/chat.go, but the code is pkg/sidedoc/chat.go. Task 5 says the guard covers the adapter and README, but TestNoSpecWritesModeKeys excludes both. The Log records both; add Revisions entries.
          family: plan-table-drift
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: Map default of a retired or unknown value is not named, and CRLF frontmatter legacy chat is silently unread
          detail: readFrontmatter emits only the generic map notice for default discussion or debate. frontmatter() refuses CRLF, so a legacy chat scalar is dropped with no notice.
          family: legacy-key-notice-coverage
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: adapters/file-edit.md:140 still says "A round in discussion or chat mode"
          family: stale-mode-prose
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#575 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T12:11:33-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `chat-grammar-oracle-parity` (kind: defect) Spec and ChatSections doc claim setext headings, but a tag above a setext heading is always misplaced
  collaboration-modes.md:107 and chat.go:53 say "ATX or setext"; CommonMark folds the tag line into the setext heading, so ChatSections reports it misplaced. Plan said ATX only. Fix the prose to ATX and pin setext as misplaced in chat_test.go.
- **BR-2** [Important] `chat-grammar-oracle-parity` (kind: defect) Indented-code @@chat@@ directly above a heading makes the section chat
  chat.go:120 checks the previous line's trimmed text without the CodeLiteralMask used at chat.go:108; a probe of "para, blank, 4-space-indented tag, then a heading" returns Source tag. Check the mask there and add a test row.
- **BR-3** [Important] `fixture-oracle-crosscheck` (kind: defect) Reply notices are not cross-checked with ChatSections and the exported fixtures already diverge
  Spec section 7 says replies word notices as ChatSections does, but checkChat ignores scan.Notices. The legacy-chat-comment reply cites line 11 while the oracle reports 12, and the scanner emits several notices where the spec says one line. Assert notices in checkChat, fix the fixture, and state which document the line numbers refer to.
- **BR-4** [Important] `plan-table-drift` (kind: defect) Plan Core concepts and Task 5 no longer match the code: ChatSections location and guard scope
  The table lists pkg/reviewexport/chat.go, but the code is pkg/sidedoc/chat.go. Task 5 says the guard covers the adapter and README, but TestNoSpecWritesModeKeys excludes both. The Log records both; add Revisions entries.
- **BR-5** [Minor] `legacy-key-notice-coverage` (kind: defect) Map default of a retired or unknown value is not named, and CRLF frontmatter legacy chat is silently unread
  readFrontmatter emits only the generic map notice for default discussion or debate. frontmatter() refuses CRLF, so a legacy chat scalar is dropped with no notice.
- **BR-6** [Minor] `stale-mode-prose` (kind: defect) adapters/file-edit.md:140 still says "A round in discussion or chat mode"

## Open findings

- **BR-1** [Important] `chat-grammar-oracle-parity` (kind: defect) Spec and ChatSections doc claim setext headings, but a tag above a setext heading is always misplaced
- **BR-2** [Important] `chat-grammar-oracle-parity` (kind: defect) Indented-code @@chat@@ directly above a heading makes the section chat
- **BR-3** [Important] `fixture-oracle-crosscheck` (kind: defect) Reply notices are not cross-checked with ChatSections and the exported fixtures already diverge
- **BR-4** [Important] `plan-table-drift` (kind: defect) Plan Core concepts and Task 5 no longer match the code: ChatSections location and guard scope
- **BR-5** [Minor] `legacy-key-notice-coverage` (kind: defect) Map default of a retired or unknown value is not named, and CRLF frontmatter legacy chat is silently unread
- **BR-6** [Minor] `stale-mode-prose` (kind: defect) adapters/file-edit.md:140 still says "A round in discussion or chat mode"
