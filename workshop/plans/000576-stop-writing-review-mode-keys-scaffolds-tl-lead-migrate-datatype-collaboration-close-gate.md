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

## Open findings

- **BR-1** [Important] `list-link-grammar-pinning` (kind: defect) tl-lead link recomputes the item grammar from a fresh Peek instead of reusing the grammar list printed, so a tag or frontmatter-map change outside the hashed section re-numbers bullets and link writes to the wrong one
- **BR-2** [Important] `promised-test-not-delivered` (kind: defect) Task 1's fake-backed list/link tests for a heading-comment session file and a @@chat@@ session file are missing
- **BR-3** [Important] `promised-test-not-delivered` (kind: defect) The legacy round-trip test covers SetField only; the card publish named in Task 2 and Done-when is not tested
- **BR-4** [Minor] `plan-code-drift` (kind: defect) The plan still says legacy=false for an undeclared section, uses frontmatter+section, and keeps workingdoc.md's key; the code does otherwise
- **BR-5** [Minor] `guard-skip-too-broad` (kind: defect) TestNoSpecWritesModeKeys skips any creator line containing the word "legacy", a broad hole in the guard
- **BR-6** [Minor] `heading-lookup-exact-match` (kind: defect) legacyGrammar finds the section by an exact "## Operator instructions" line and ignores enclosing sections
