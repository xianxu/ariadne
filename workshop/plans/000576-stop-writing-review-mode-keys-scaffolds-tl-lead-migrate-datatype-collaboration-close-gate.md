---
gate: boundary-review
issue: 576
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 393a7478-9e0e-474a-bb19-c380e18fe64a
            repository: github.com/xianxu/ariadne
            issue_id: "000576"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/e94d3ba9daee6b13
            reviewed_head: 4096b2e43bf00d217e1c4afdfb8384eed34bd4b1
            slot: ariadne:2
            requested_at: "2026-10-23T21:48:36.873152Z"
          phase: finished
          pid: 11238
          started_at: "2026-10-23T21:48:36.983063Z"
          finished_at: "2026-10-23T21:52:37.732996Z"
          session_id: 96bfb0ed-07f1-4de9-b80c-2fb9ee8570c3
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 4096b2e43bf00d217e1c4afdfb8384eed34bd4b1
        tracker_head: 989a9a71af55dadc67f3abd40b4fa7fb17dc3be4
        branch_point: c39d1899c04e1176d5bdce47f9f06848dd17ce2c
        since: "2026-10-23T14:24:15-07:00"
        until: "2026-10-23T14:46:28-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T14:24:15-07:00"
              end: "2026-10-23T14:46:28-07:00"
              began: "2026-10-23T14:24:15-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c23fcd720026234e94694ac8e7f2efe03120a66185d7dd7df5e7aeaeb5541284
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 11800419
              windows:
                - start: "2026-10-23T14:24:15-07:00"
                  end: "2026-10-23T14:46:28-07:00"
      "n": 1
      timestamp: "2026-10-23T14:52:38-07:00"
      dispatched: "2026-10-23T14:48:36-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: tl-lead link recomputes the item grammar from a fresh Peek instead of reusing the grammar list printed, so a tag or frontmatter-map change outside the hashed section re-numbers bullets and link writes to the wrong one
          detail: 'cmd/tl-lead/instructions.go:127 calls legacyGrammar(ss.Peek(file)) at link time. The section hash covers only the section body; the @@chat@@ tag line (above the heading) and a frontmatter collaboration-mode map do not affect it. Example: section "- note\n1. a\n2. b". list (legacy) shows position 2 = "1. a". The operator adds @@chat@@ as help.md now prescribes, and link 2 under the numbered grammar appends to "2. b". The plan''s ARCH-ORDER note promised that link reuses list''s grammar. Fix: have list print the grammar and link take it as a flag (or refuse on mismatch), plus a fake-backed list -> add tag -> link sequence test.'
          family: list-link-grammar-pinning
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Task 1's fake-backed list/link tests for a heading-comment session file and a @@chat@@ session file are missing
          detail: Only the pure helper is tested (TestLegacyGrammar, instructions_test.go:306). No test drives sectionBullets, link's Peek path, or the list output on a tagged file through the session-sync fake, so the new IO wiring has no regression coverage.
          family: promised-test-not-delivered
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: The legacy round-trip test covers SetField only; the card publish named in Task 2 and Done-when is not tested
          detail: 'nomodekeys_test.go TestLegacyModeKeySurvivesSetField checks SetField alone. Task 2 (checked) promises that a legacy collaboration-mode line survives "SetField and a card publish" byte for byte. Add a SplitCard/compose round trip on an issue that carries collaboration-mode: collaboration.'
          family: promised-test-not-delivered
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: The plan still says legacy=false for an undeclared section, uses frontmatter+section, and keeps workingdoc.md's key; the code does otherwise
          detail: 'legacyGrammar returns true for an undeclared file and reads Peek(file); workingdoc.md''s own key was removed. Add ## Revisions entries.'
          family: plan-code-drift
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: TestNoSpecWritesModeKeys skips any creator line containing the word "legacy", a broad hole in the guard
          detail: A line instructing a creator to write a key would pass if it also mentions "legacy". Consider skipping by explicit allow-list lines or by a marker phrase instead.
          family: guard-skip-too-broad
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: legacyGrammar finds the section by an exact "## Operator instructions" line and ignores enclosing sections
          detail: A tagged parent heading, or a trailing heading annotation, falls back to the legacy grammar; ChatScan.Innermost would resolve it.
          family: heading-lookup-exact-match
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 860ee805-a595-4e6b-8b5c-c18e274eed06
            repository: github.com/xianxu/ariadne
            issue_id: "000576"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/b6b051cb513eeb35
            reviewed_head: 72459336c6d684be44f6a4cb4de15e09e71886a8
            slot: ariadne:2
            requested_at: "2026-10-23T21:57:44.642944Z"
          phase: finished
          pid: 2349
          started_at: "2026-10-23T21:57:44.78014Z"
          finished_at: "2026-10-23T21:59:09.984941Z"
          session_id: 78202f8f-3a2f-4dfe-b161-5d283a2cbda5
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 72459336c6d684be44f6a4cb4de15e09e71886a8
        tracker_head: 989a9a71af55dadc67f3abd40b4fa7fb17dc3be4
        branch_point: c39d1899c04e1176d5bdce47f9f06848dd17ce2c
        since: "2026-10-23T14:24:15-07:00"
        until: "2026-10-23T14:56:53-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T14:24:15-07:00"
              end: "2026-10-23T14:56:53-07:00"
              began: "2026-10-23T14:24:15-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 4f2b7c196cd1b560abe59522830c56b0449bb1f4d66920cbd7542c758d1b445e
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 12052087
              windows:
                - start: "2026-10-23T14:24:15-07:00"
                  end: "2026-10-23T14:56:53-07:00"
        cuts:
            - dispatched: "2026-10-23T14:48:36-07:00"
              finalized: "2026-10-23T14:52:38-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-23T14:59:10-07:00"
      dispatched: "2026-10-23T14:57:44-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: The token is now hash:grammar and link refuses with exit 6 on a change. Disabling the check makes TestInstructionsGrammarThroughTheFake fail (exit 0, writes 1).
          round: 2
        - id: BR-2
          disposition: addressed
          note: TestInstructionsGrammarThroughTheFake runs list and link through the fake for a tagged session file and a heading-comment one.
          round: 2
        - id: BR-3
          disposition: addressed
          note: TestLegacyModeKeysSurviveCardPublish checks that the details frontmatter keeps a byte prefix and all three legacy keys, and that the card has none.
          round: 2
        - id: BR-4
          disposition: addressed
          note: Plan Revisions entries record Peek(file), legacy=true for an undeclared section, and workingdoc.md's key removal; each matches instructions.go:39-57.
          round: 2
        - id: BR-5
          disposition: addressed
          note: The guard now excuses only the "legacy `key`" span (examples_test.go:816,855); the rest of the line is still checked.
          round: 2
        - id: BR-6
          disposition: addressed
          note: An enclosing chat section now governs through Innermost, and a revert check confirms the test goes red. A trailing annotation on the heading itself still uses the legacy grammar, a known low-value gap.
          round: 2
      findings:
        - kind: defect
          id: BR-7
          severity: Minor
          title: help.md still says list prints "the section's hash"; it now prints a HASH:GRAMMAR token
          detail: 'cmd/tl-lead/help.md:433-436 should say the token also pins the item grammar, and that exit 6 can mean a tag or legacy declaration changed. This is the 2nd finding in family plan-code-drift. The rule: when a fix changes a printed or accepted token, the same commit updates every prose surface that describes it (help.md, the plan, atlas). Copying H verbatim still works, so this is cosmetic.'
          family: plan-code-drift
          round: 2
      recipe: milestone-review
      isolation: checkout
      reviewed: 72459336c6d684be44f6a4cb4de15e09e71886a8
      blocked: false
---

# Gate ledger — ariadne#576 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T14:52:38-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `list-link-grammar-pinning` (kind: defect) tl-lead link recomputes the item grammar from a fresh Peek instead of reusing the grammar list printed, so a tag or frontmatter-map change outside the hashed section re-numbers bullets and link writes to the wrong one
  cmd/tl-lead/instructions.go:127 calls legacyGrammar(ss.Peek(file)) at link time. The section hash covers only the section body; the @@chat@@ tag line (above the heading) and a frontmatter collaboration-mode map do not affect it. Example: section "- note\n1. a\n2. b". list (legacy) shows position 2 = "1. a". The operator adds @@chat@@ as help.md now prescribes, and link 2 under the numbered grammar appends to "2. b". The plan's ARCH-ORDER note promised that link reuses list's grammar. Fix: have list print the grammar and link take it as a flag (or refuse on mismatch), plus a fake-backed list -> add tag -> link sequence test.
- **BR-2** [Important] `promised-test-not-delivered` (kind: defect) Task 1's fake-backed list/link tests for a heading-comment session file and a @@chat@@ session file are missing
  Only the pure helper is tested (TestLegacyGrammar, instructions_test.go:306). No test drives sectionBullets, link's Peek path, or the list output on a tagged file through the session-sync fake, so the new IO wiring has no regression coverage.
- **BR-3** [Important] `promised-test-not-delivered` (kind: defect) The legacy round-trip test covers SetField only; the card publish named in Task 2 and Done-when is not tested
  nomodekeys_test.go TestLegacyModeKeySurvivesSetField checks SetField alone. Task 2 (checked) promises that a legacy collaboration-mode line survives "SetField and a card publish" byte for byte. Add a SplitCard/compose round trip on an issue that carries collaboration-mode: collaboration.
- **BR-4** [Minor] `plan-code-drift` (kind: defect) The plan still says legacy=false for an undeclared section, uses frontmatter+section, and keeps workingdoc.md's key; the code does otherwise
  legacyGrammar returns true for an undeclared file and reads Peek(file); workingdoc.md's own key was removed. Add ## Revisions entries.
- **BR-5** [Minor] `guard-skip-too-broad` (kind: defect) TestNoSpecWritesModeKeys skips any creator line containing the word "legacy", a broad hole in the guard
  A line instructing a creator to write a key would pass if it also mentions "legacy". Consider skipping by explicit allow-list lines or by a marker phrase instead.
- **BR-6** [Minor] `heading-lookup-exact-match` (kind: defect) legacyGrammar finds the section by an exact "## Operator instructions" line and ignores enclosing sections
  A tagged parent heading, or a trailing heading annotation, falls back to the legacy grammar; ChatScan.Innermost would resolve it.

## Round 2 — 2026-10-23T14:59:10-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — The token is now hash:grammar and link refuses with exit 6 on a change. Disabling the check makes TestInstructionsGrammarThroughTheFake fail (exit 0, writes 1).
- BR-2 — addressed — TestInstructionsGrammarThroughTheFake runs list and link through the fake for a tagged session file and a heading-comment one.
- BR-3 — addressed — TestLegacyModeKeysSurviveCardPublish checks that the details frontmatter keeps a byte prefix and all three legacy keys, and that the card has none.
- BR-4 — addressed — Plan Revisions entries record Peek(file), legacy=true for an undeclared section, and workingdoc.md's key removal; each matches instructions.go:39-57.
- BR-5 — addressed — The guard now excuses only the "legacy `key`" span (examples_test.go:816,855); the rest of the line is still checked.
- BR-6 — addressed — An enclosing chat section now governs through Innermost, and a revert check confirms the test goes red. A trailing annotation on the heading itself still uses the legacy grammar, a known low-value gap.

### Raised

- **BR-7** [Minor] `plan-code-drift` (kind: defect) help.md still says list prints "the section's hash"; it now prints a HASH:GRAMMAR token
  cmd/tl-lead/help.md:433-436 should say the token also pins the item grammar, and that exit 6 can mean a tag or legacy declaration changed. This is the 2nd finding in family plan-code-drift. The rule: when a fix changes a printed or accepted token, the same commit updates every prose surface that describes it (help.md, the plan, atlas). Copying H verbatim still works, so this is cosmetic.

## Open findings

- **BR-7** [Minor] `plan-code-drift` (kind: defect) help.md still says list prints "the section's hash"; it now prints a HASH:GRAMMAR token
