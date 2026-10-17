---
gate: boundary-review
issue: 424
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: d0bbd2a3-1565-4700-8724-42785c89b21d
            repository: github.com/xianxu/ariadne
            issue_id: "000424"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/819ca98d954e956b
            reviewed_head: 8517f816bd85d33877493b1c257dfe24218f2df4
            slot: ariadne:2
            requested_at: "2026-10-17T10:25:46.221408Z"
          phase: finished
          pid: 70769
          started_at: "2026-10-17T10:25:46.438189Z"
          finished_at: "2026-10-17T10:31:28.62108Z"
          session_id: 1a1f82ce-8ae9-405f-b233-8ade14ec25bd
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 8517f816bd85d33877493b1c257dfe24218f2df4
        tracker_head: 04bc61d838ac50914d1f6c4401cd0df6d7a4d93f
        branch_point: 409e789d83910de366d11b8a536c2c6d8dc9900e
        since: "2026-10-17T03:02:14-07:00"
        until: "2026-10-17T03:24:49-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T03:02:14-07:00"
              end: "2026-10-17T03:24:49-07:00"
              began: "2026-10-17T03:02:14-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: b3278d84c6194c2166f27489a2503907cda05f887f9fef1ea73cf1ccb68a635f
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 34298111
              windows:
                - start: "2026-10-17T03:02:14-07:00"
                  end: "2026-10-17T03:24:49-07:00"
      "n": 1
      timestamp: "2026-10-17T03:31:29-07:00"
      dispatched: "2026-10-17T03:25:46-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: 'ReconcileIndex stops parsing discussions: at a blank line, so later entries are never pruned and adds splice mid-list'
          detail: 'pkg/sidedoc/index.go:75-79 breaks on an empty line before the t=="" continue. Reproduced with "discussions:\n  - doc: d.side/01-a.md\n\n  - doc: d.side/09-gone.md" (no link to 09): 09 is kept, and the add lands between 01-a and 09-gone, which YAML still reads as one list. Fix: treat blank lines as continuation or refuse; add a table case.'
          family: hand-edited-index-parse-refuses
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: A frontmatter ReconcileIndex can't split (CRLF, unclosed) is treated as absent and a second one is prepended
          detail: 'index.go:33-48 returns ok=false for "---\r\n..." and for an unclosed "---\n..."; ReconcileIndex (:164) then prepends a fresh frontmatter, and the original becomes body text. Reproduced for both. Fix: a document starting with --- whose frontmatter can''t be split returns an error and the input unchanged, as the plan''s Trust section promises.'
          family: hand-edited-index-parse-refuses
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: The examples checker does not enforce "a fold adds no index entry"
          detail: 'checkIndexTransitions (examples_test.go:384-388) only flags an add when the side doc is unchanged and was already linked; a fold always adds the closing note, so it passes. Confirmed: adding the 01-naming entry to fold-from-parent/after/design.md leaves TestExamples green. Fix: flag a new entry whose side doc newly matches foldNoteRE, and add a mutation for it.'
          family: contract-rule-unenforced-by-checker
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: move-rename/case.md describes a notes.md link that the fixture does not contain
          detail: The case is exported to parley as a fixture; drop the sentence or add the file.
          family: fixture-prose-matches-tree
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: FuzzReconcileIndex reads its output back with the same parser it is testing
          detail: Self-consistent misparses (the blank-line case) pass the property; use an independent YAML decode of the frontmatter as the oracle.
          family: independent-test-oracle
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: Plan still says rows 21 to 31, a links_test.go, and .agents/skills/xx-universal-review paths
          detail: The delivered rows are 21 to 29 (explained in the log, but no Revisions entry), TestSideLinks is in index_test.go, and the authored source is construct/local/universal-review/. Add a Revisions entry.
          family: plan-matches-code
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: ce460e10-dce7-44d7-a171-b96272dff9c5
            repository: github.com/xianxu/ariadne
            issue_id: "000424"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/95c83b13d48626aa
            reviewed_head: a085177edd84ce83ea91aaccf9203e1b2159f9ff
            slot: ariadne:2
            requested_at: "2026-10-17T10:40:40.154656Z"
          phase: finished
          pid: 95532
          started_at: "2026-10-17T10:40:40.43711Z"
          finished_at: "2026-10-17T10:52:01.648846Z"
          session_id: 64c810b9-15c0-4ef2-b73b-4207b2ff712a
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: a085177edd84ce83ea91aaccf9203e1b2159f9ff
        tracker_head: 04bc61d838ac50914d1f6c4401cd0df6d7a4d93f
        branch_point: 409e789d83910de366d11b8a536c2c6d8dc9900e
        since: "2026-10-17T03:02:14-07:00"
        until: "2026-10-17T03:40:05-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T03:02:14-07:00"
              end: "2026-10-17T03:40:05-07:00"
              began: "2026-10-17T03:02:14-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 2b6df5d5cf0fbd9ea60fcb473b76d052f32985d0a6c26f4de0d49dd4fe32bd00
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 37375115
              windows:
                - start: "2026-10-17T03:02:14-07:00"
                  end: "2026-10-17T03:40:05-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-17T03:25:46-07:00"
              finalized: "2026-10-17T03:31:29-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-17T03:52:02-07:00"
      dispatched: "2026-10-17T03:40:40-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: index.go:87-96 continues through blank lines; table case fails on revert to 8517f816.
          round: 2
        - id: BR-2
          disposition: addressed
          note: errFrontmatter refuses CRLF and unclosed blocks unchanged; TestReconcileIndexRefuses fails on revert.
          round: 2
        - id: BR-3
          disposition: addressed
          note: examples_test.go:389-393 flags fold-added entries; mutation "an entry added by a fold" fails with the check disabled.
          round: 2
        - id: BR-4
          disposition: addressed
          note: notes.md sentence removed from move-rename/case.md; export.json hash updated; swept all new case.md file references, none missing.
          round: 2
        - id: BR-5
          disposition: addressed
          note: Fuzz oracle is now a yaml.v3 decode of the reconciled frontmatter (index_test.go:131-154).
          round: 2
        - id: BR-6
          disposition: addressed
          note: Revisions entry records rows 21-29, the construct/local source path, and TestSideLinks location.
          round: 2
      findings:
        - id: BR-7
          severity: Important
          title: 'A column-0 comment inside discussions: ends the block, so later entries are never pruned and adds splice mid-list (3rd in family)'
          detail: 'index.go:93 breaks at any unindented non-item line. For "discussions:\n  - doc: d.side/01-a.md\n# c\n  - doc: d.side/09-gone.md\n" with add 02-b, ReconcileIndex returns nil error, reports Entries [01-a 02-b], and writes 01-a, 02-b, then the comment, then 09-gone, which YAML reads as three entries. Tab-indented items (invalid YAML) are also accepted and rewritten. This is the 3rd instance in this family. Rather than patching this site, enforce the rule: the reconcile decodes the frontmatter with yaml.v3 and refuses when the decoded discussions[].doc differs from what it parsed. Add table cases for the comment and tab inputs, and correct the plan Revisions claim "Anything else is refused".'
          family: hand-edited-index-parse-refuses
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 7ebaa550-16d6-4725-9f61-f9fdd9ed83c2
            repository: github.com/xianxu/ariadne
            issue_id: "000424"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/956b70ec7253af13
            reviewed_head: b4b62c42ffb718dc720f79d7908aa8ae56bbd901
            slot: ariadne:2
            requested_at: "2026-10-17T10:56:13.187162Z"
          phase: finished
          pid: 14792
          started_at: "2026-10-17T10:56:13.279868Z"
          finished_at: "2026-10-17T10:57:30.522148Z"
          session_id: 90a8d77e-c5a7-4ab7-aaf4-2a9b2e91154b
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: b4b62c42ffb718dc720f79d7908aa8ae56bbd901
        tracker_head: beffdc90c4c28c72051d300c3570a00d31384efd
        branch_point: 409e789d83910de366d11b8a536c2c6d8dc9900e
        since: "2026-10-17T03:02:14-07:00"
        until: "2026-10-17T03:55:52-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T03:02:14-07:00"
              end: "2026-10-17T03:55:52-07:00"
              began: "2026-10-17T03:02:14-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: bf30eeeaf84d2c6a95fd5403ed78f13a07152fad72c9c5385bb847798bfdd935
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 39810839
              windows:
                - start: "2026-10-17T03:02:14-07:00"
                  end: "2026-10-17T03:55:52-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-17T03:25:46-07:00"
              finalized: "2026-10-17T03:31:29-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-17T03:40:40-07:00"
              finalized: "2026-10-17T03:52:02-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-17T03:57:31-07:00"
      dispatched: "2026-10-17T03:56:13-07:00"
      agent: claude
      dispose:
        - id: BR-7
          disposition: addressed
          note: agreesWithYAML decodes the frontmatter with yaml.v3 before and after the rewrite (index.go:198,228). With it reverted, TestReconcileIndexRefuses fails on the comment and tab cases; restored, the tests and 20s of fuzzing pass. The plan Revisions entry is corrected.
          round: 3
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      reviewed: b4b62c42ffb718dc720f79d7908aa8ae56bbd901
      blocked: false
---

# Gate ledger — ariadne#424 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T03:31:29-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `hand-edited-index-parse-refuses` (kind: defect) ReconcileIndex stops parsing discussions: at a blank line, so later entries are never pruned and adds splice mid-list
  pkg/sidedoc/index.go:75-79 breaks on an empty line before the t=="" continue. Reproduced with "discussions:\n  - doc: d.side/01-a.md\n\n  - doc: d.side/09-gone.md" (no link to 09): 09 is kept, and the add lands between 01-a and 09-gone, which YAML still reads as one list. Fix: treat blank lines as continuation or refuse; add a table case.
- **BR-2** [Important] `hand-edited-index-parse-refuses` (kind: defect) A frontmatter ReconcileIndex can't split (CRLF, unclosed) is treated as absent and a second one is prepended
  index.go:33-48 returns ok=false for "---\r\n..." and for an unclosed "---\n..."; ReconcileIndex (:164) then prepends a fresh frontmatter, and the original becomes body text. Reproduced for both. Fix: a document starting with --- whose frontmatter can't be split returns an error and the input unchanged, as the plan's Trust section promises.
- **BR-3** [Important] `contract-rule-unenforced-by-checker` (kind: defect) The examples checker does not enforce "a fold adds no index entry"
  checkIndexTransitions (examples_test.go:384-388) only flags an add when the side doc is unchanged and was already linked; a fold always adds the closing note, so it passes. Confirmed: adding the 01-naming entry to fold-from-parent/after/design.md leaves TestExamples green. Fix: flag a new entry whose side doc newly matches foldNoteRE, and add a mutation for it.
- **BR-4** [Minor] `fixture-prose-matches-tree` (kind: defect) move-rename/case.md describes a notes.md link that the fixture does not contain
  The case is exported to parley as a fixture; drop the sentence or add the file.
- **BR-5** [Minor] `independent-test-oracle` (kind: defect) FuzzReconcileIndex reads its output back with the same parser it is testing
  Self-consistent misparses (the blank-line case) pass the property; use an independent YAML decode of the frontmatter as the oracle.
- **BR-6** [Minor] `plan-matches-code` (kind: defect) Plan still says rows 21 to 31, a links_test.go, and .agents/skills/xx-universal-review paths
  The delivered rows are 21 to 29 (explained in the log, but no Revisions entry), TestSideLinks is in index_test.go, and the authored source is construct/local/universal-review/. Add a Revisions entry.

## Round 2 — 2026-10-17T03:52:02-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — index.go:87-96 continues through blank lines; table case fails on revert to 8517f816.
- BR-2 — addressed — errFrontmatter refuses CRLF and unclosed blocks unchanged; TestReconcileIndexRefuses fails on revert.
- BR-3 — addressed — examples_test.go:389-393 flags fold-added entries; mutation "an entry added by a fold" fails with the check disabled.
- BR-4 — addressed — notes.md sentence removed from move-rename/case.md; export.json hash updated; swept all new case.md file references, none missing.
- BR-5 — addressed — Fuzz oracle is now a yaml.v3 decode of the reconciled frontmatter (index_test.go:131-154).
- BR-6 — addressed — Revisions entry records rows 21-29, the construct/local source path, and TestSideLinks location.

### Raised

- **BR-7** [Important] `hand-edited-index-parse-refuses` A column-0 comment inside discussions: ends the block, so later entries are never pruned and adds splice mid-list (3rd in family)
  index.go:93 breaks at any unindented non-item line. For "discussions:\n  - doc: d.side/01-a.md\n# c\n  - doc: d.side/09-gone.md\n" with add 02-b, ReconcileIndex returns nil error, reports Entries [01-a 02-b], and writes 01-a, 02-b, then the comment, then 09-gone, which YAML reads as three entries. Tab-indented items (invalid YAML) are also accepted and rewritten. This is the 3rd instance in this family. Rather than patching this site, enforce the rule: the reconcile decodes the frontmatter with yaml.v3 and refuses when the decoded discussions[].doc differs from what it parsed. Add table cases for the comment and tab inputs, and correct the plan Revisions claim "Anything else is refused".

## Round 3 — 2026-10-17T03:57:31-07:00 (claude) — passed

### Disposed

- BR-7 — addressed — agreesWithYAML decodes the frontmatter with yaml.v3 before and after the rewrite (index.go:198,228). With it reverted, TestReconcileIndexRefuses fails on the comment and tab cases; restored, the tests and 20s of fuzzing pass. The plan Revisions entry is corrected.

## Open findings

(none — every finding has been disposed)
