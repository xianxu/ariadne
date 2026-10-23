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
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 7ff57d0e-7ef2-4b9e-86b6-5124415082f0
            repository: github.com/xianxu/ariadne
            issue_id: "000575"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/e8a01914f50112e4
            reviewed_head: f1e4752bd9b7920606ada1739383c73ffc1575f6
            slot: ariadne:2
            requested_at: "2026-10-23T19:16:09.565549Z"
          phase: finished
          pid: 67873
          started_at: "2026-10-23T19:16:09.774686Z"
          finished_at: "2026-10-23T19:18:30.916027Z"
          session_id: 2135aa92-3250-486d-8fb1-ce9cbfadad1d
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: f1e4752bd9b7920606ada1739383c73ffc1575f6
        tracker_head: 7d6eb4f6643322154a6495910fb09412c6bb893d
        branch_point: 588a0b0af09a8965d472f7acc216581db3afa815
        since: "2026-10-23T11:52:48-07:00"
        until: "2026-10-23T12:15:18-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T11:52:48-07:00"
              end: "2026-10-23T12:15:18-07:00"
              began: "2026-10-23T11:52:48-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 999083a4040e194897332f4151003d54c068f3b1819474f13dc9bff2c8aa40f3
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 9477615
              windows:
                - start: "2026-10-23T11:52:48-07:00"
                  end: "2026-10-23T12:15:18-07:00"
        cuts:
            - dispatched: "2026-10-23T12:07:55-07:00"
              finalized: "2026-10-23T12:11:33-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-23T12:18:32-07:00"
      dispatched: "2026-10-23T12:16:09-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Spec line 107 and the chat.go:53 doc now say ATX only; a chat_test.go row pins a setext-adjacent tag as misplaced.
          round: 2
        - id: BR-2
          disposition: addressed
          note: tagLine applies the CodeLiteralMask; reverting the mask check fails the indented-code and fenced test rows.
          round: 2
        - id: BR-3
          disposition: addressed
          note: checkChat asserts the oracle notices verbatim; reverting legacy-chat-comment/case.md fails TestExamples; spec section 7 states the line-number basis.
          round: 2
        - id: BR-4
          disposition: addressed
          note: The Core concepts row now says pkg/sidedoc/chat.go; Revisions lines 189-190 record the move and the guard's actual scope.
          round: 2
        - id: BR-5
          disposition: addressed
          note: Map default notice and CRLF-prefix notice added, with test rows; the remaining unread-frontmatter routes are raised below as the family rule.
          round: 2
        - id: BR-6
          disposition: addressed
          note: file-edit.md:140 reworded; a grep sweep of construct/local/universal-review, the atlas and README finds no remaining mode-era phrasing.
          round: 2
      findings:
        - kind: defect
          id: BR-7
          severity: Minor
          title: Frontmatter that frontmatter() refuses or YAML cannot parse drops legacy keys with no notice
          detail: 'This is the 2nd finding in family legacy-key-notice-coverage, so fix the rule, not this instance. Rule: when the doc opens with a --- fence and its keys were not read for any reason (frontmatter() refused, any CR in the doc, a ... or % line, or yaml.Unmarshal error), ChatSections emits one notice. Probes in which a legacy chat scalar vanished silently: LF frontmatter plus a CR in the body, a malformed flow sequence, a ... marker. Replace the CRLF-prefix check at chat.go:67 and the silent return at chat.go:173 with that one rule, and add rows for each route.'
          family: legacy-key-notice-coverage
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: f1e4752bd9b7920606ada1739383c73ffc1575f6
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: aee73bfb-0a0a-4f5b-b204-58fe9f882387
            repository: github.com/xianxu/ariadne
            issue_id: "000575"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/5d4601b41e81e3d7
            reviewed_head: 213f0fa2d58f44975260da50b823ea9f57760f41
            slot: ariadne:2
            requested_at: "2026-10-23T19:20:14.177636Z"
          phase: finished
          pid: 90290
          started_at: "2026-10-23T19:20:14.366087Z"
          finished_at: "2026-10-23T19:22:01.134438Z"
          session_id: 72d0912c-a9a2-4b2f-8b31-f1be1dd2713a
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 213f0fa2d58f44975260da50b823ea9f57760f41
        tracker_head: 9931424325a44d83b3cd7b30b0677abd32c7bce6
        branch_point: 588a0b0af09a8965d472f7acc216581db3afa815
        since: "2026-10-23T11:52:48-07:00"
        until: "2026-10-23T12:19:28-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T11:52:48-07:00"
              end: "2026-10-23T12:19:28-07:00"
              began: "2026-10-23T11:52:48-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: e36384d8f632afac8ae2b4e2b6314aa983799bce6a4f60216f8cc96208ed2f02
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 9627576
              windows:
                - start: "2026-10-23T11:52:48-07:00"
                  end: "2026-10-23T12:19:28-07:00"
        cuts:
            - dispatched: "2026-10-23T12:07:55-07:00"
              finalized: "2026-10-23T12:11:33-07:00"
              finalizing: false
            - dispatched: "2026-10-23T12:16:09-07:00"
              finalized: "2026-10-23T12:18:32-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-23T12:22:02-07:00"
      dispatched: "2026-10-23T12:20:14-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-7
          disposition: addressed
          note: chat.go:73 and chat.go:178 apply one rule for every unread fence; chat_test.go:67-86 rows go red when the pre-fix chat.go is restored (checked in this checkout).
          round: 3
      recipe: milestone-review
      isolation: fenced
      reviewed: 213f0fa2d58f44975260da50b823ea9f57760f41
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 46859808-70bf-4e66-88d0-9904c459b57d
            repository: github.com/xianxu/ariadne
            issue_id: "000575"
            recipe: milestone-review
            round: 4
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/14574eb7300ebb4e
            reviewed_head: 006f6dd3b2a2d0125a7594d0b5fe485b0642e544
            slot: ariadne:2
            requested_at: "2026-10-23T19:54:48.207961Z"
          phase: finished
          pid: 45740
          started_at: "2026-10-23T19:54:48.352559Z"
          finished_at: "2026-10-23T20:07:15.421098Z"
          session_id: cfbe9945-fc17-47e3-bba5-24f932f47cce
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 006f6dd3b2a2d0125a7594d0b5fe485b0642e544
        tracker_head: 1430081a35e8e4845974beb97f0749d057f45059
        branch_point: 588a0b0af09a8965d472f7acc216581db3afa815
        since: "2026-10-23T11:52:48-07:00"
        until: "2026-10-23T12:54:14-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T11:52:48-07:00"
              end: "2026-10-23T12:54:14-07:00"
              began: "2026-10-23T11:52:48-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 0853b4fbca811d7d88337b3d7738baf898175f92a9db813f755c29b5078f29d4
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 9812493
              windows:
                - start: "2026-10-23T11:52:48-07:00"
                  end: "2026-10-23T12:54:14-07:00"
        queue_waits:
            - attempt_id: 53212cf34ff54e94bbc5c828c645f562
              slot: ariadne:2
              record_digest: d3b5f0006809f76bf79820900fc8e6623a1b499588961e4af8d71ef82b0790c0
              start: "2026-10-23T19:19:36.247598Z"
              end: "2026-10-23T19:33:03.83328Z"
              queue_seconds: 807.585682
              raw_seconds: 807.585682
              exact: true
        cuts:
            - dispatched: "2026-10-23T12:07:55-07:00"
              finalized: "2026-10-23T12:11:33-07:00"
              finalizing: false
            - dispatched: "2026-10-23T12:16:09-07:00"
              finalized: "2026-10-23T12:18:32-07:00"
              finalizing: true
            - dispatched: "2026-10-23T12:20:14-07:00"
              finalized: "2026-10-23T12:22:02-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-23T13:07:16-07:00"
      dispatched: "2026-10-23T12:54:48-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-8
          severity: Minor
          title: A document opening with a thematic break and no closing fence gets a false "frontmatter could not be read" notice
          detail: 'This is the 3rd finding in family legacy-key-notice-coverage. chat.go:73 fires the notice whenever the document starts with a fence line, so "---\nJust a rule above prose.\n" reports unread frontmatter it does not have, and the spec requires replies to repeat that notice word for word. The rule to fix: notice only when a frontmatter block exists (an opening fence with a matching closing fence) and it could not be read. Use the same closing-fence detection as frontmatter(), not a prefix test.'
          family: legacy-key-notice-coverage
          round: 4
        - kind: defect
          id: BR-9
          severity: Minor
          title: 'An empty collaboration-mode value prints "collaboration-mode:  is retired and ignored"'
          detail: chat.go:189 formats an empty scalar as a retired value, with a double space. §7 classes an empty value as empty or malformed, so the notice should say that instead.
          family: notice-text-names-value
          round: 4
      recipe: milestone-review
      isolation: fenced
      reviewed: 006f6dd3b2a2d0125a7594d0b5fe485b0642e544
      blocked: false
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

## Round 2 — 2026-10-23T12:18:32-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Spec line 107 and the chat.go:53 doc now say ATX only; a chat_test.go row pins a setext-adjacent tag as misplaced.
- BR-2 — addressed — tagLine applies the CodeLiteralMask; reverting the mask check fails the indented-code and fenced test rows.
- BR-3 — addressed — checkChat asserts the oracle notices verbatim; reverting legacy-chat-comment/case.md fails TestExamples; spec section 7 states the line-number basis.
- BR-4 — addressed — The Core concepts row now says pkg/sidedoc/chat.go; Revisions lines 189-190 record the move and the guard's actual scope.
- BR-5 — addressed — Map default notice and CRLF-prefix notice added, with test rows; the remaining unread-frontmatter routes are raised below as the family rule.
- BR-6 — addressed — file-edit.md:140 reworded; a grep sweep of construct/local/universal-review, the atlas and README finds no remaining mode-era phrasing.

### Raised

- **BR-7** [Minor] `legacy-key-notice-coverage` (kind: defect) Frontmatter that frontmatter() refuses or YAML cannot parse drops legacy keys with no notice
  This is the 2nd finding in family legacy-key-notice-coverage, so fix the rule, not this instance. Rule: when the doc opens with a --- fence and its keys were not read for any reason (frontmatter() refused, any CR in the doc, a ... or % line, or yaml.Unmarshal error), ChatSections emits one notice. Probes in which a legacy chat scalar vanished silently: LF frontmatter plus a CR in the body, a malformed flow sequence, a ... marker. Replace the CRLF-prefix check at chat.go:67 and the silent return at chat.go:173 with that one rule, and add rows for each route.

## Round 3 — 2026-10-23T12:22:02-07:00 (claude) — passed

### Disposed

- BR-7 — addressed — chat.go:73 and chat.go:178 apply one rule for every unread fence; chat_test.go:67-86 rows go red when the pre-fix chat.go is restored (checked in this checkout).

## Round 4 — 2026-10-23T13:07:16-07:00 (claude) — passed

### Raised

- **BR-8** [Minor] `legacy-key-notice-coverage` (kind: defect) A document opening with a thematic break and no closing fence gets a false "frontmatter could not be read" notice
  This is the 3rd finding in family legacy-key-notice-coverage. chat.go:73 fires the notice whenever the document starts with a fence line, so "---\nJust a rule above prose.\n" reports unread frontmatter it does not have, and the spec requires replies to repeat that notice word for word. The rule to fix: notice only when a frontmatter block exists (an opening fence with a matching closing fence) and it could not be read. Use the same closing-fence detection as frontmatter(), not a prefix test.
- **BR-9** [Minor] `notice-text-names-value` (kind: defect) An empty collaboration-mode value prints "collaboration-mode:  is retired and ignored"
  chat.go:189 formats an empty scalar as a retired value, with a double space. §7 classes an empty value as empty or malformed, so the notice should say that instead.

## Open findings

- **BR-8** [Minor] `legacy-key-notice-coverage` (kind: defect) A document opening with a thematic break and no closing fence gets a false "frontmatter could not be read" notice
- **BR-9** [Minor] `notice-text-names-value` (kind: defect) An empty collaboration-mode value prints "collaboration-mode:  is retired and ignored"
