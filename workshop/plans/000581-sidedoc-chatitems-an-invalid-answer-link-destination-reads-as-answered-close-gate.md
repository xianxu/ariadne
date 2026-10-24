---
gate: boundary-review
issue: 581
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: ad396c90-c1fb-4eef-b9cf-7188e34bd8e9
            repository: github.com/xianxu/ariadne
            issue_id: "000581"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/1166d916b6ecc6ca
            reviewed_head: faa2aa966807907656a7ffa050c389b76eabedda
            slot: ariadne:2
            requested_at: "2026-10-24T02:35:21.869455Z"
          phase: finished
          pid: 21027
          started_at: "2026-10-24T02:35:22.015753Z"
          finished_at: "2026-10-24T02:37:48.870495Z"
          session_id: 35e96f63-19cd-4efb-90ad-fc8f06742871
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: faa2aa966807907656a7ffa050c389b76eabedda
        tracker_head: e7c3fa5f302ada8a5521cde7685b929abd807126
        branch_point: 84b68e2fc5c0773ff38f13771507d93aaa08efd4
        since: "2026-10-23T18:02:32-07:00"
        until: "2026-10-23T19:34:45-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T18:02:32-07:00"
              end: "2026-10-23T19:34:45-07:00"
              began: "2026-10-23T18:02:32-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: da088648b6fee36af95c5c8a197c547be2564be69c99b9144b09e0bc6b44e061
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 15878573
              windows:
                - start: "2026-10-23T18:02:32-07:00"
                  end: "2026-10-23T19:34:45-07:00"
      "n": 1
      timestamp: "2026-10-23T19:37:50-07:00"
      dispatched: "2026-10-23T19:35:21-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: examples checker drops every Entries error in side-link-free documents, not only unread headers
          detail: |-
            pkg/reviewexport/examples_test.go:168 gates on len(SideLinks(text)) > 0, so a readable header with a
            malformed discussions: index (indexBlock/agreesWithYAML error) and no side link now passes silently,
            but the comment claims only the unread-header case is excused. Fix: excuse only the case where FrontBlock
            found a closed block and frontmatter() refused it. This is the only instance in the window.
          family: guard-broader-than-stated-case
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: TestChatSections unread-header rows never assert "no heading" (or Tags directly)
          detail: |-
            pkg/sidedoc/chat_test.go:31-43: the Done-when clause requires no section, no tag and no heading from inside
            the header, but the table checks only sections, misplaced tags and notices, so a heading leaking into
            scan.Headings goes undetected. Add a headings/tags assertion to both rows (spaced fence, CR endings).
          family: done-when-clause-unasserted
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: unread-header body trim removes only one CR before the newline
          detail: |-
            chat.go:147: a closing fence line like a triple dash followed by two CRs and a newline leaves the body
            starting with a stray line, so line numbers and code-mask offsets are off by one. Use
            strings.Cut(rest, "\n") to take the remainder.
          family: body-offset-derivation
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: migrate.go answerLinkRE restates a looser answer-target shape
          detail: |-
            cmd/tl-lead/migrate.go:111 uses a regex that accepts empty targets and an opening paren. It only strips
            text, so it is harmless, but it is another restatement of the rule ValidAnswerTarget now owns.
          family: answer-target-single-source
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 1b2b5ec8-7bba-4ad9-94f4-893585c86765
            repository: github.com/xianxu/ariadne
            issue_id: "000581"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/76e9c98973e0aeee
            reviewed_head: b2c0639bc2f2d058a8f0387790e23f6a908bd9d1
            slot: ariadne:2
            requested_at: "2026-10-24T02:41:36.004749Z"
          phase: finished
          pid: 91088
          started_at: "2026-10-24T02:41:36.1109Z"
          finished_at: "2026-10-24T02:44:48.992613Z"
          session_id: 5de31a41-c26a-4249-bf19-9a862f0a89c4
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: b2c0639bc2f2d058a8f0387790e23f6a908bd9d1
        tracker_head: e7c3fa5f302ada8a5521cde7685b929abd807126
        branch_point: 84b68e2fc5c0773ff38f13771507d93aaa08efd4
        since: "2026-10-23T18:02:32-07:00"
        until: "2026-10-23T19:40:59-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T18:02:32-07:00"
              end: "2026-10-23T19:40:59-07:00"
              began: "2026-10-23T18:02:32-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: ddbc9fd410c508ea92654e17d87fe233db01f7148942871c5e16cd999eb3068d
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 16050464
              windows:
                - start: "2026-10-23T18:02:32-07:00"
                  end: "2026-10-23T19:40:59-07:00"
        cuts:
            - dispatched: "2026-10-23T19:35:21-07:00"
              finalized: "2026-10-23T19:37:50-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-23T19:44:50-07:00"
      dispatched: "2026-10-23T19:41:36-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Guard excuses only closed && !readable && no side link; TestCheckCaseExcusesOnlyAnUnreadHeader goes red (readable malformed index) when reverted to the SideLinks-only gate.
          round: 2
        - id: BR-2
          disposition: addressed
          note: TestChatSectionsIgnoreAnUnreadHeader asserts Headings, Tags, Misplaced, Sections and Notices for spaced fence, CR and two-CR rows; removing the header skip fails all of them.
          round: 2
        - id: BR-3
          disposition: addressed
          note: chat.go:147 uses strings.Cut(rest, "\n"); reverting to a single-CR trim fails the two-CRs subtest (heading at line 8, tag misplaced).
          round: 2
        - id: BR-4
          disposition: addressed
          note: stripAnswerLinks defers to sidedoc.ValidAnswerTarget; TestStripAnswerLinks goes red when it strips every link. No other restatement of the rule remains in non-test Go code.
          round: 2
      findings:
        - kind: defect
          id: BR-5
          severity: Minor
          title: Plan item 3 says frontmatter-unread-tag's body marker opens a side doc; the fixture edits instead
          detail: The ticked plan item says "the body marker opens a side doc", but examples/frontmatter-unread-tag/after/note.md edits the paragraph (scope s1) as a command, since an unread header can hold no index. Only the Log records the change. Add a Revisions entry so the plan describes the shipped fixture. This is the only instance in the window.
          family: plan-claim-matches-delivery
          round: 2
      recipe: small-diff-review
      isolation: fenced
      reviewed: b2c0639bc2f2d058a8f0387790e23f6a908bd9d1
      blocked: false
---

# Gate ledger — ariadne#581 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T19:37:50-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `guard-broader-than-stated-case` (kind: defect) examples checker drops every Entries error in side-link-free documents, not only unread headers
  pkg/reviewexport/examples_test.go:168 gates on len(SideLinks(text)) > 0, so a readable header with a
  malformed discussions: index (indexBlock/agreesWithYAML error) and no side link now passes silently,
  but the comment claims only the unread-header case is excused. Fix: excuse only the case where FrontBlock
  found a closed block and frontmatter() refused it. This is the only instance in the window.
- **BR-2** [Important] `done-when-clause-unasserted` (kind: defect) TestChatSections unread-header rows never assert "no heading" (or Tags directly)
  pkg/sidedoc/chat_test.go:31-43: the Done-when clause requires no section, no tag and no heading from inside
  the header, but the table checks only sections, misplaced tags and notices, so a heading leaking into
  scan.Headings goes undetected. Add a headings/tags assertion to both rows (spaced fence, CR endings).
- **BR-3** [Minor] `body-offset-derivation` (kind: defect) unread-header body trim removes only one CR before the newline
  chat.go:147: a closing fence line like a triple dash followed by two CRs and a newline leaves the body
  starting with a stray line, so line numbers and code-mask offsets are off by one. Use
  strings.Cut(rest, "\n") to take the remainder.
- **BR-4** [Minor] `answer-target-single-source` (kind: defect) migrate.go answerLinkRE restates a looser answer-target shape
  cmd/tl-lead/migrate.go:111 uses a regex that accepts empty targets and an opening paren. It only strips
  text, so it is harmless, but it is another restatement of the rule ValidAnswerTarget now owns.

## Round 2 — 2026-10-23T19:44:50-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Guard excuses only closed && !readable && no side link; TestCheckCaseExcusesOnlyAnUnreadHeader goes red (readable malformed index) when reverted to the SideLinks-only gate.
- BR-2 — addressed — TestChatSectionsIgnoreAnUnreadHeader asserts Headings, Tags, Misplaced, Sections and Notices for spaced fence, CR and two-CR rows; removing the header skip fails all of them.
- BR-3 — addressed — chat.go:147 uses strings.Cut(rest, "\n"); reverting to a single-CR trim fails the two-CRs subtest (heading at line 8, tag misplaced).
- BR-4 — addressed — stripAnswerLinks defers to sidedoc.ValidAnswerTarget; TestStripAnswerLinks goes red when it strips every link. No other restatement of the rule remains in non-test Go code.

### Raised

- **BR-5** [Minor] `plan-claim-matches-delivery` (kind: defect) Plan item 3 says frontmatter-unread-tag's body marker opens a side doc; the fixture edits instead
  The ticked plan item says "the body marker opens a side doc", but examples/frontmatter-unread-tag/after/note.md edits the paragraph (scope s1) as a command, since an unread header can hold no index. Only the Log records the change. Add a Revisions entry so the plan describes the shipped fixture. This is the only instance in the window.

## Open findings

- **BR-5** [Minor] `plan-claim-matches-delivery` (kind: defect) Plan item 3 says frontmatter-unread-tag's body marker opens a side doc; the fixture edits instead
