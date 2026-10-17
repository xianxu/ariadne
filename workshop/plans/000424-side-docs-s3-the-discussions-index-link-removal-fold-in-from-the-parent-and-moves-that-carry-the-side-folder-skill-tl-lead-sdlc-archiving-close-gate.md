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
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 109a1e37-ebda-422d-8589-66f9d474da4e
            repository: github.com/xianxu/ariadne
            issue_id: "000424"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/2ecfa701c2823147
            reviewed_head: 4567a9e0931ed7c4bed5886d738bdfc73aab5cb8
            slot: ariadne:2
            requested_at: "2026-10-17T11:04:05.49281Z"
          phase: finished
          pid: 28404
          started_at: "2026-10-17T11:04:05.723048Z"
          finished_at: "2026-10-17T11:07:11.769654Z"
          session_id: 704720e8-70fb-40a0-8a02-3c6d66d19937
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 4567a9e0931ed7c4bed5886d738bdfc73aab5cb8
        tracker_head: beffdc90c4c28c72051d300c3570a00d31384efd
        branch_point: 409e789d83910de366d11b8a536c2c6d8dc9900e
        since: "2026-10-17T03:02:14-07:00"
        until: "2026-10-17T04:03:36-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T03:02:14-07:00"
              end: "2026-10-17T04:03:36-07:00"
              began: "2026-10-17T03:02:14-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: ff00f4896b4d10d2ec148d1a0aeee8e8d92cefe616ece2c26924eedf92b4fa6d
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 46219902
              windows:
                - start: "2026-10-17T03:02:14-07:00"
                  end: "2026-10-17T04:03:36-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-17T03:25:46-07:00"
              finalized: "2026-10-17T03:31:29-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-17T03:40:40-07:00"
              finalized: "2026-10-17T03:52:02-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-17T03:56:13-07:00"
              finalized: "2026-10-17T03:57:31-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-17T04:07:12-07:00"
      dispatched: "2026-10-17T04:04:05-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-8
          severity: Critical
          title: The index's read --all and write --all mark every operator edit as seen, so link, index and submit empty the "changes since your last read" feed
          detail: 'cmd/tl-lead/index.go:25 and :34 call ReadAll and WriteAll, which run couch''s read --all and write --all. In pair, sessionsync/sync.go:243-255 says reading the whole file refreshes the shadow so diff goes empty, and shadowAfter (sync.go:361) makes the whole new document the shadow on a whole-file write. tl-lead''s own MarkAllSeen comment (sessionsync.go:22) documents the same verb. Result: an instructions link, a no-op status index, or a submit (for edits landing after its Diff) silently consumes unread operator edits, which help.md:183-184 says arrive as changes since the last read. Fix: either a couch frontmatter-scoped read/write that moves the baseline only there (a pair change), or syncIndex prints d.ss.Diff before ReadAll and a no-op reconcile avoids the baseline reset. Add a regression test against a fake that models the shadow.'
          family: seam-side-effect-unmodeled
          round: 4
        - kind: defect
          id: BR-9
          severity: Important
          title: fakeSessionSync does not model couch's diff baseline, so no test can see the seen side effect of ReadAll and WriteAll (ARCH-MOCK)
          detail: 'sessionsync_fake_test.go:97 returns a fixed diff string, and ReadAll/WriteAll (lines 131-171) never touch it. Give the fake a shadow: ReadAll and WriteAll set it to the whole text, WriteSection updates only its section, Diff compares the file with the shadow. Add a sequence test (operator edits Needs you, TL links, the edit is still reported), and have TestSessionSyncConformance check diff after read --all against the live couch.'
          family: seam-side-effect-unmodeled
          round: 4
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 2ca6af37-7856-49dd-a7fd-be8fd016e958
            repository: github.com/xianxu/ariadne
            issue_id: "000424"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/3a4c3d609c6ec262
            reviewed_head: efbd03435ea08cc407535ea55c8ddec94fa031b7
            slot: ariadne:2
            requested_at: "2026-10-17T11:10:24.577893Z"
          phase: finished
          pid: 52469
          started_at: "2026-10-17T11:10:24.699057Z"
          finished_at: "2026-10-17T11:12:23.664892Z"
          session_id: 58a54723-7a0f-4212-84d5-39e617c121e0
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: efbd03435ea08cc407535ea55c8ddec94fa031b7
        tracker_head: beffdc90c4c28c72051d300c3570a00d31384efd
        branch_point: 409e789d83910de366d11b8a536c2c6d8dc9900e
        since: "2026-10-17T03:02:14-07:00"
        until: "2026-10-17T04:09:35-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T03:02:14-07:00"
              end: "2026-10-17T04:09:35-07:00"
              began: "2026-10-17T03:02:14-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 74424c8dad02ad36f9e52e530052ff5a42e361c1b8c3e99f07c7921e88b507fb
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 51162034
              windows:
                - start: "2026-10-17T03:02:14-07:00"
                  end: "2026-10-17T04:09:35-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-17T03:25:46-07:00"
              finalized: "2026-10-17T03:31:29-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-17T03:40:40-07:00"
              finalized: "2026-10-17T03:52:02-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-17T03:56:13-07:00"
              finalized: "2026-10-17T03:57:31-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-17T04:04:05-07:00"
              finalized: "2026-10-17T04:07:12-07:00"
              finalizing: false
      "n": 5
      timestamp: "2026-10-17T04:12:53-07:00"
      dispatched: "2026-10-17T04:10:24-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: not-addressed
          note: 'The no-op path is fixed (Peek) and the first attempt prints Diff, but on ErrStale syncIndex (index.go:47-67) loops into ReadAll again without re-printing Diff, so the edit that caused the stale refusal is marked seen and never shown. A probe against the fake (f.seen(), editAllBefore{1: Needs you edit}, status index --add) printed only "index: added" and Diff was "" afterwards. Fix: Diff and print before every ReadAll in the loop, plus a regression test with a baseline.'
          round: 5
        - id: BR-9
          disposition: addressed
          note: The fake has a shadow (sessionsync_fake_test.go:517-566), the sequence test TestIndexKeepsUnreadEdits exists and went red when I reverted the Peek and print fix, and conformance_test.go:32-46 pins diff-after-read-all and Peek equal to read --all against the live couch.
          round: 5
      findings:
        - kind: defect
          id: BR-10
          severity: Important
          title: The plan omits the new Peek seam method and still says the index holds no state between events, though couch's seen-baseline is state the fix depends on
          detail: 'This is the 2nd finding in family plan-matches-code. The rule for the class: any seam method or state assumption introduced while fixing a review finding gets a Revisions entry and an Integration points row in the same fix commit. Here, add SessionSync.Peek (cmd/tl-lead/sessionsync.go:466, os.ReadFile outside couch) to the table, and revise the ARCH-ORDER tl-lead line to name couch''s seen-baseline, the print-before-every-ReadAll rule, and the remaining window between Diff and ReadAll.'
          family: plan-matches-code
          round: 5
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 1d42e8d9-30d0-46fe-a741-9033094c2540
            repository: github.com/xianxu/ariadne
            issue_id: "000424"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/8ec7d4ba145a72ff
            reviewed_head: 1b2e41e1e7ecffbf7a7759fd2e4287e917ff8087
            slot: ariadne:2
            requested_at: "2026-10-17T11:14:40.607693Z"
          phase: finished
          pid: 50037
          started_at: "2026-10-17T11:14:40.97461Z"
          finished_at: "2026-10-17T11:16:01.175086Z"
          session_id: d6d30cbf-790d-408b-9203-0549fb3a15c7
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 1b2e41e1e7ecffbf7a7759fd2e4287e917ff8087
        tracker_head: beffdc90c4c28c72051d300c3570a00d31384efd
        branch_point: 409e789d83910de366d11b8a536c2c6d8dc9900e
        since: "2026-10-17T03:02:14-07:00"
        until: "2026-10-17T04:14:03-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T03:02:14-07:00"
              end: "2026-10-17T04:14:03-07:00"
              began: "2026-10-17T03:02:14-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 00bbeb48a1e6638c7444fd460d157bae4d1d350158841ab6a229d80645ec577f
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 53490670
              windows:
                - start: "2026-10-17T03:02:14-07:00"
                  end: "2026-10-17T04:14:03-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-17T03:25:46-07:00"
              finalized: "2026-10-17T03:31:29-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-17T03:40:40-07:00"
              finalized: "2026-10-17T03:52:02-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-17T03:56:13-07:00"
              finalized: "2026-10-17T03:57:31-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-17T04:04:05-07:00"
              finalized: "2026-10-17T04:07:12-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-17T04:10:24-07:00"
              finalized: "2026-10-17T04:12:53-07:00"
              finalizing: false
      "n": 6
      timestamp: "2026-10-17T04:16:02-07:00"
      dispatched: "2026-10-17T04:14:40-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: addressed
          note: Peek no-op path plus readAllShowingUnread on every read and retry; each half verified red when reverted (TestIndexKeepsUnreadEdits, TestIndexShowsEditsOnEveryRead); live conformance pins the read --all baseline reset.
          round: 6
        - id: BR-10
          disposition: addressed
          note: The plan adds a SessionSync.Peek Integration points row, revises the ARCH-ORDER tl-lead line (seen-baseline, print-before-read rule, remaining couch-internal window, pair 643) and adds a Revisions entry stating the class rule.
          round: 6
      findings:
        - kind: defect
          id: BR-11
          severity: Minor
          title: lineChanges in cmd/tl-lead/index.go re-implements the package's existing lineDiff (ARCH-DRY)
          detail: cmd/tl-lead/linediff.go:8 already renders the -/+ line diff of two status-file texts; lineChanges (index.go:92) is a second multiset variant. Use lineDiff and drop its shared-line output instead.
          family: reuse-existing-helper
          round: 6
        - kind: defect
          id: BR-12
          severity: Minor
          title: submit prints the unread diff twice when the index reconcile needs a write
          detail: status.go:185-189 prints Diff, then syncIndex's readAllShowingUnread prints the same unread changes again under its own header. The output is redundant but nothing is lost; the second print could be skipped when submit has just shown the same diff.
          family: duplicate-output
          round: 6
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      reviewed: 1b2e41e1e7ecffbf7a7759fd2e4287e917ff8087
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 1665dc7a-b4c4-434b-ad7c-3d1fc73d0542
            repository: github.com/xianxu/ariadne
            issue_id: "000424"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/4518a87842d7857a
            reviewed_head: e43b38b3d644683b23be83ae14df7b5aa4fa2b41
            slot: ariadne:2
            requested_at: "2026-10-17T12:30:55.777778Z"
          phase: finished
          pid: 99199
          started_at: "2026-10-17T12:30:55.873821Z"
          finished_at: "2026-10-17T12:45:25.689999Z"
          session_id: 8a578e7c-5ff2-422a-a084-fc7b4176d6b7
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: e43b38b3d644683b23be83ae14df7b5aa4fa2b41
        tracker_head: beffdc90c4c28c72051d300c3570a00d31384efd
        branch_point: 409e789d83910de366d11b8a536c2c6d8dc9900e
        since: "2026-10-17T03:02:14-07:00"
        until: "2026-10-17T05:05:49-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T03:02:14-07:00"
              end: "2026-10-17T05:05:49-07:00"
              began: "2026-10-17T03:02:14-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 7b1ccc5ada53cb48249c94d6b6f6f9c5ec790985cf54f42f01ae8e4167f59b2e
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 67464411
              windows:
                - start: "2026-10-17T03:02:14-07:00"
                  end: "2026-10-17T05:05:49-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-17T03:25:46-07:00"
              finalized: "2026-10-17T03:31:29-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-17T03:40:40-07:00"
              finalized: "2026-10-17T03:52:02-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-17T03:56:13-07:00"
              finalized: "2026-10-17T03:57:31-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-17T04:04:05-07:00"
              finalized: "2026-10-17T04:07:12-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-17T04:10:24-07:00"
              finalized: "2026-10-17T04:12:53-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-17T04:14:40-07:00"
              finalized: "2026-10-17T04:16:02-07:00"
              finalizing: true
      "n": 7
      timestamp: "2026-10-17T05:45:27-07:00"
      dispatched: "2026-10-17T05:30:55-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-13
          severity: Important
          title: 'verifyArchivedSideMoves judges operator-written parent: lines only after every move, stranding push/merge mid-archive'
          detail: 'An archive keeps the base name and moves the folder whole, so a parent: resolves after the move exactly when it resolved before. The verify (archiveside.go:166, called at push.go:216, merge.go:740 and in recovery) can only fail on a parent: already broken by hand, and it fails after every file has moved; recovery repeats the same die. Probe: parent: ../000143-old.md gives archiveDoneIssues err=nil, the file already in history, then a verify error. Check parent: in prepareSideArchive before anything moves (or warn), and keep the post-move check to "each prepared member exists at its destination". ARCH-SECURE / ARCH-ORDER.'
          family: refuse-before-effects
          round: 7
        - kind: defect
          id: BR-14
          severity: Important
          title: trunk successor episode replaces only the issue's side folder; the kept branch's plan side folders are dropped
          detail: 'trunkSideReplace (archiveside.go:310) maps only the live issue folder, so the plan side files in a.Sides (abandon.go:301-315) never match and are skipped silently, while archivemain.go:113 rewrites the archived plans from the kept branch. The plans'' links then point at the old episode''s folder. trunkSideReplace has no test. The retry residue check (archivemain.go:78) likewise covers only the issue''s folder, though plan (e) says "the issue or its plans". Rule (Revisions'' own): every move of an issue or plan document carries its folder. Map each a.Plans doc too, extend the residue check, add a successor test with a plan side doc. ARCH-PURPOSE.'
          family: move-carries-side-folder
          round: 7
        - kind: defect
          id: BR-15
          severity: Important
          title: sdlc help text not updated for side folders moving on archive/abandon/reopen
          detail: Plan Task 7 lists cmd/sdlc/helptext/ and Step 4 says "atlas and help text". The atlas was updated, but push.md:48, merge.md:45 and abandon.md:13,25 still describe only details, assets and plans moving.
          family: docs-update-for-surface
          round: 7
        - kind: defect
          id: BR-16
          severity: Minor
          title: sideFolderOf duplicates sidedoc.InSideDir; sideParent is a third ad-hoc frontmatter reader
          detail: 'This is the 2nd finding in family reuse-existing-helper. Class rule: side-doc path and frontmatter facts (folder of a path, parent: of a side doc) live in pkg/sidedoc and every cmd/ and checker consumer calls them. Sites: landingarchive.go:373 sideFolderOf vs pkg/sidedoc/tree.go InSideDir; archiveside.go:189 sideParent vs pkg/sidedoc/index.go:40 frontmatter and pkg/reviewexport/examples_test.go:63. sideParent also returns "" on CRLF frontmatter.'
          family: reuse-existing-helper
          round: 7
        - kind: defect
          id: BR-17
          severity: Minor
          title: prepareSideArchive ignores readDirNames errors; recovery ignores ReadFile errors
          detail: 'archiveside.go:67: a destination .side that exists as a file is not refused at prepare, so it fails inside the finisher after the issue has moved. archiveside.go:130-131: an unreadable pair is reported as "duplicate" instead of as an IO error.'
          family: silent-error-swallow
          round: 7
        - kind: defect
          id: BR-18
          severity: Minor
          title: reopen git-rm's a differing live side doc without comparing bytes
          detail: abandon.go:646-670 compares bytes only for assets. A live side doc at a reopen destination is replaced by the archived copy without a check, unlike the forward archive's never-replace rule.
          family: move-carries-side-folder
          round: 7
      boundary: M3
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 741192ce-6cd5-4430-94e8-e7ef12726ebf
            repository: github.com/xianxu/ariadne
            issue_id: "000424"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/f1fbd20b085e3f78
            reviewed_head: aef20f75848fbb1eec0a69dd0f9578cf2941085d
            slot: ariadne:2
            requested_at: "2026-10-17T13:20:07.450478Z"
          phase: finished
          pid: 8843
          started_at: "2026-10-17T13:20:07.62553Z"
          finished_at: "2026-10-17T13:33:24.568103Z"
          session_id: ad6a30a1-f646-44ae-81a0-0e79149341ad
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: aef20f75848fbb1eec0a69dd0f9578cf2941085d
        tracker_head: 1cc8358108942a062db4c1b5675bd2968a0b6394
        branch_point: 409e789d83910de366d11b8a536c2c6d8dc9900e
        since: "2026-10-17T03:02:14-07:00"
        until: "2026-10-17T06:19:41-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T03:02:14-07:00"
              end: "2026-10-17T06:19:41-07:00"
              began: "2026-10-17T03:02:14-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 89e4ff4aad86a1e6be268d44849182f6c88c705053fec24a8d3068190e8a2e7b
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 71050112
              windows:
                - start: "2026-10-17T03:02:14-07:00"
                  end: "2026-10-17T06:19:41-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-17T03:25:46-07:00"
              finalized: "2026-10-17T03:31:29-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-17T03:40:40-07:00"
              finalized: "2026-10-17T03:52:02-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-17T03:56:13-07:00"
              finalized: "2026-10-17T03:57:31-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-17T04:04:05-07:00"
              finalized: "2026-10-17T04:07:12-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-17T04:10:24-07:00"
              finalized: "2026-10-17T04:12:53-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-17T04:14:40-07:00"
              finalized: "2026-10-17T04:16:02-07:00"
              finalizing: true
            - boundary: M3
              dispatched: "2026-10-17T05:30:55-07:00"
              finalized: "2026-10-17T05:45:27-07:00"
              finalizing: false
      "n": 8
      timestamp: "2026-10-17T06:33:25-07:00"
      dispatched: "2026-10-17T06:20:07-07:00"
      agent: claude
      dispose:
        - id: BR-13
          disposition: addressed
          note: Verify now checks arrival only (archiveside.go:188); TestVerifyArchivedSideMoves goes red with e43b38b3's verify restored.
          round: 8
        - id: BR-14
          disposition: addressed
          note: trunkSideReplace maps every .md plan and trunkSideResidue covers plan folders; the successor and residue tests go red when reverted.
          round: 8
        - id: BR-15
          disposition: addressed
          note: push.md:46-50, merge.md:44-46 and abandon.md:13-16 and 26 now name side folders moving; matches atlas/workflow/issue-assets.md.
          round: 8
        - id: BR-16
          disposition: addressed
          note: sideFolderOf replaced by sidedoc.Folder; sideParent deleted; examples_test uses sidedoc.Fields; one frontmatter reader remains.
          round: 8
        - id: BR-17
          disposition: not-addressed
          note: Code fix present (archiveside.go:69-82, 141-148), but no test covers a destination that is a file or a ReadFile error.
          round: 8
        - id: BR-18
          disposition: not-addressed
          note: Reopen now compares side-doc bytes (abandon.go:665), but no test reopens over a differing live side doc.
          round: 8
      boundary: M3
      recipe: milestone-review
      isolation: fenced
      reviewed: aef20f75848fbb1eec0a69dd0f9578cf2941085d
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a346d263-89c0-4b8a-997b-33364c3098e1
            repository: github.com/xianxu/ariadne
            issue_id: "000424"
            recipe: milestone-review
            round: 9
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/92c7a101ea4e6a3a
            reviewed_head: 5990bf04a5d7cebb58dcf279b19e142ae3f2fe5f
            slot: ariadne:2
            requested_at: "2026-10-17T14:14:56.438235Z"
          phase: finished
          pid: 38154
          started_at: "2026-10-17T14:14:56.694041Z"
          finished_at: "2026-10-17T14:29:11.581166Z"
          session_id: febcb445-8b80-4f9a-af60-2a3e616ef19d
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 5990bf04a5d7cebb58dcf279b19e142ae3f2fe5f
        tracker_head: 1cc8358108942a062db4c1b5675bd2968a0b6394
        branch_point: 146796d884ccf9a9d589a8766d7627762b4571e6
        since: "2026-10-17T03:02:14-07:00"
        until: "2026-10-17T06:36:01-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T03:02:14-07:00"
              end: "2026-10-17T06:36:01-07:00"
              began: "2026-10-17T03:02:14-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 4604b4628718f17303bf6fb9175e791da07919259c1872d453110d2035c632b3
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 74275415
              windows:
                - start: "2026-10-17T03:02:14-07:00"
                  end: "2026-10-17T06:36:01-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-17T03:25:46-07:00"
              finalized: "2026-10-17T03:31:29-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-17T03:40:40-07:00"
              finalized: "2026-10-17T03:52:02-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-17T03:56:13-07:00"
              finalized: "2026-10-17T03:57:31-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-17T04:04:05-07:00"
              finalized: "2026-10-17T04:07:12-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-17T04:10:24-07:00"
              finalized: "2026-10-17T04:12:53-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-17T04:14:40-07:00"
              finalized: "2026-10-17T04:16:02-07:00"
              finalizing: true
            - boundary: M3
              dispatched: "2026-10-17T05:30:55-07:00"
              finalized: "2026-10-17T05:45:27-07:00"
              finalizing: false
            - boundary: M3
              dispatched: "2026-10-17T06:20:07-07:00"
              finalized: "2026-10-17T06:33:25-07:00"
              finalizing: true
      "n": 9
      timestamp: "2026-10-17T07:29:12-07:00"
      dispatched: "2026-10-17T07:14:56-07:00"
      agent: claude
      dispose:
        - id: BR-11
          disposition: addressed
          note: index.go windowChanges now filters lineDiff output; lineChanges removed (f83186dd).
          round: 9
        - id: BR-12
          disposition: addressed
          note: submit passes its printed diff to syncIndexAfter, which skips an identical reprint; TestSubmitPrintsUnreadOnce pins it.
          round: 9
        - id: BR-17
          disposition: addressed
          note: archiveside.go refuses a non-folder destination and a readDirNames error at prepare; recovery returns ReadFile errors; TestPrepareSideArchiveRefusesFileDestination passes.
          round: 9
        - id: BR-18
          disposition: addressed
          note: reopen byte-compares side docs like assets; reverting the InSideDir clause makes TestReopenDoneSideFolder fail (verified).
          round: 9
      recipe: milestone-review
      isolation: checkout
      reviewed: 5990bf04a5d7cebb58dcf279b19e142ae3f2fe5f
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

## Round 4 — 2026-10-17T04:07:12-07:00 (claude) — BLOCKED

### Raised

- **BR-8** [Critical] `seam-side-effect-unmodeled` (kind: defect) The index's read --all and write --all mark every operator edit as seen, so link, index and submit empty the "changes since your last read" feed
  cmd/tl-lead/index.go:25 and :34 call ReadAll and WriteAll, which run couch's read --all and write --all. In pair, sessionsync/sync.go:243-255 says reading the whole file refreshes the shadow so diff goes empty, and shadowAfter (sync.go:361) makes the whole new document the shadow on a whole-file write. tl-lead's own MarkAllSeen comment (sessionsync.go:22) documents the same verb. Result: an instructions link, a no-op status index, or a submit (for edits landing after its Diff) silently consumes unread operator edits, which help.md:183-184 says arrive as changes since the last read. Fix: either a couch frontmatter-scoped read/write that moves the baseline only there (a pair change), or syncIndex prints d.ss.Diff before ReadAll and a no-op reconcile avoids the baseline reset. Add a regression test against a fake that models the shadow.
- **BR-9** [Important] `seam-side-effect-unmodeled` (kind: defect) fakeSessionSync does not model couch's diff baseline, so no test can see the seen side effect of ReadAll and WriteAll (ARCH-MOCK)
  sessionsync_fake_test.go:97 returns a fixed diff string, and ReadAll/WriteAll (lines 131-171) never touch it. Give the fake a shadow: ReadAll and WriteAll set it to the whole text, WriteSection updates only its section, Diff compares the file with the shadow. Add a sequence test (operator edits Needs you, TL links, the edit is still reported), and have TestSessionSyncConformance check diff after read --all against the live couch.

## Round 5 — 2026-10-17T04:12:53-07:00 (claude) — BLOCKED

### Disposed

- BR-8 — not-addressed — The no-op path is fixed (Peek) and the first attempt prints Diff, but on ErrStale syncIndex (index.go:47-67) loops into ReadAll again without re-printing Diff, so the edit that caused the stale refusal is marked seen and never shown. A probe against the fake (f.seen(), editAllBefore{1: Needs you edit}, status index --add) printed only "index: added" and Diff was "" afterwards. Fix: Diff and print before every ReadAll in the loop, plus a regression test with a baseline.
- BR-9 — addressed — The fake has a shadow (sessionsync_fake_test.go:517-566), the sequence test TestIndexKeepsUnreadEdits exists and went red when I reverted the Peek and print fix, and conformance_test.go:32-46 pins diff-after-read-all and Peek equal to read --all against the live couch.

### Raised

- **BR-10** [Important] `plan-matches-code` (kind: defect) The plan omits the new Peek seam method and still says the index holds no state between events, though couch's seen-baseline is state the fix depends on
  This is the 2nd finding in family plan-matches-code. The rule for the class: any seam method or state assumption introduced while fixing a review finding gets a Revisions entry and an Integration points row in the same fix commit. Here, add SessionSync.Peek (cmd/tl-lead/sessionsync.go:466, os.ReadFile outside couch) to the table, and revise the ARCH-ORDER tl-lead line to name couch's seen-baseline, the print-before-every-ReadAll rule, and the remaining window between Diff and ReadAll.

## Round 6 — 2026-10-17T04:16:02-07:00 (claude) — passed

### Disposed

- BR-8 — addressed — Peek no-op path plus readAllShowingUnread on every read and retry; each half verified red when reverted (TestIndexKeepsUnreadEdits, TestIndexShowsEditsOnEveryRead); live conformance pins the read --all baseline reset.
- BR-10 — addressed — The plan adds a SessionSync.Peek Integration points row, revises the ARCH-ORDER tl-lead line (seen-baseline, print-before-read rule, remaining couch-internal window, pair 643) and adds a Revisions entry stating the class rule.

### Raised

- **BR-11** [Minor] `reuse-existing-helper` (kind: defect) lineChanges in cmd/tl-lead/index.go re-implements the package's existing lineDiff (ARCH-DRY)
  cmd/tl-lead/linediff.go:8 already renders the -/+ line diff of two status-file texts; lineChanges (index.go:92) is a second multiset variant. Use lineDiff and drop its shared-line output instead.
- **BR-12** [Minor] `duplicate-output` (kind: defect) submit prints the unread diff twice when the index reconcile needs a write
  status.go:185-189 prints Diff, then syncIndex's readAllShowingUnread prints the same unread changes again under its own header. The output is redundant but nothing is lost; the second print could be skipped when submit has just shown the same diff.

## Round 7 — 2026-10-17T05:45:27-07:00 (claude) — BLOCKED

### Raised

- **BR-13** [Important] `refuse-before-effects` (kind: defect) verifyArchivedSideMoves judges operator-written parent: lines only after every move, stranding push/merge mid-archive
  An archive keeps the base name and moves the folder whole, so a parent: resolves after the move exactly when it resolved before. The verify (archiveside.go:166, called at push.go:216, merge.go:740 and in recovery) can only fail on a parent: already broken by hand, and it fails after every file has moved; recovery repeats the same die. Probe: parent: ../000143-old.md gives archiveDoneIssues err=nil, the file already in history, then a verify error. Check parent: in prepareSideArchive before anything moves (or warn), and keep the post-move check to "each prepared member exists at its destination". ARCH-SECURE / ARCH-ORDER.
- **BR-14** [Important] `move-carries-side-folder` (kind: defect) trunk successor episode replaces only the issue's side folder; the kept branch's plan side folders are dropped
  trunkSideReplace (archiveside.go:310) maps only the live issue folder, so the plan side files in a.Sides (abandon.go:301-315) never match and are skipped silently, while archivemain.go:113 rewrites the archived plans from the kept branch. The plans' links then point at the old episode's folder. trunkSideReplace has no test. The retry residue check (archivemain.go:78) likewise covers only the issue's folder, though plan (e) says "the issue or its plans". Rule (Revisions' own): every move of an issue or plan document carries its folder. Map each a.Plans doc too, extend the residue check, add a successor test with a plan side doc. ARCH-PURPOSE.
- **BR-15** [Important] `docs-update-for-surface` (kind: defect) sdlc help text not updated for side folders moving on archive/abandon/reopen
  Plan Task 7 lists cmd/sdlc/helptext/ and Step 4 says "atlas and help text". The atlas was updated, but push.md:48, merge.md:45 and abandon.md:13,25 still describe only details, assets and plans moving.
- **BR-16** [Minor] `reuse-existing-helper` (kind: defect) sideFolderOf duplicates sidedoc.InSideDir; sideParent is a third ad-hoc frontmatter reader
  This is the 2nd finding in family reuse-existing-helper. Class rule: side-doc path and frontmatter facts (folder of a path, parent: of a side doc) live in pkg/sidedoc and every cmd/ and checker consumer calls them. Sites: landingarchive.go:373 sideFolderOf vs pkg/sidedoc/tree.go InSideDir; archiveside.go:189 sideParent vs pkg/sidedoc/index.go:40 frontmatter and pkg/reviewexport/examples_test.go:63. sideParent also returns "" on CRLF frontmatter.
- **BR-17** [Minor] `silent-error-swallow` (kind: defect) prepareSideArchive ignores readDirNames errors; recovery ignores ReadFile errors
  archiveside.go:67: a destination .side that exists as a file is not refused at prepare, so it fails inside the finisher after the issue has moved. archiveside.go:130-131: an unreadable pair is reported as "duplicate" instead of as an IO error.
- **BR-18** [Minor] `move-carries-side-folder` (kind: defect) reopen git-rm's a differing live side doc without comparing bytes
  abandon.go:646-670 compares bytes only for assets. A live side doc at a reopen destination is replaced by the archived copy without a check, unlike the forward archive's never-replace rule.

## Round 8 — 2026-10-17T06:33:25-07:00 (claude) — passed

### Disposed

- BR-13 — addressed — Verify now checks arrival only (archiveside.go:188); TestVerifyArchivedSideMoves goes red with e43b38b3's verify restored.
- BR-14 — addressed — trunkSideReplace maps every .md plan and trunkSideResidue covers plan folders; the successor and residue tests go red when reverted.
- BR-15 — addressed — push.md:46-50, merge.md:44-46 and abandon.md:13-16 and 26 now name side folders moving; matches atlas/workflow/issue-assets.md.
- BR-16 — addressed — sideFolderOf replaced by sidedoc.Folder; sideParent deleted; examples_test uses sidedoc.Fields; one frontmatter reader remains.
- BR-17 — not-addressed — Code fix present (archiveside.go:69-82, 141-148), but no test covers a destination that is a file or a ReadFile error.
- BR-18 — not-addressed — Reopen now compares side-doc bytes (abandon.go:665), but no test reopens over a differing live side doc.

## Round 9 — 2026-10-17T07:29:12-07:00 (claude) — passed

### Disposed

- BR-11 — addressed — index.go windowChanges now filters lineDiff output; lineChanges removed (f83186dd).
- BR-12 — addressed — submit passes its printed diff to syncIndexAfter, which skips an identical reprint; TestSubmitPrintsUnreadOnce pins it.
- BR-17 — addressed — archiveside.go refuses a non-folder destination and a readDirNames error at prepare; recovery returns ReadFile errors; TestPrepareSideArchiveRefusesFileDestination passes.
- BR-18 — addressed — reopen byte-compares side docs like assets; reverting the InSideDir clause makes TestReopenDoneSideFolder fail (verified).

## Open findings

(none — every finding has been disposed)
