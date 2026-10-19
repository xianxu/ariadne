---
gate: boundary-review
issue: 455
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 7f48fecb-f530-4274-b5ae-97342aaf75ac
            repository: github.com/xianxu/ariadne
            issue_id: "000455"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/7e10d7726a9e5cbb
            reviewed_head: 023d85a9241a99aac8751d5482937cd4afa7a08c
            slot: ariadne:4
            requested_at: "2026-10-19T05:46:13.496936Z"
          phase: finished
          pid: 78929
          started_at: "2026-10-19T05:46:13.726555Z"
          finished_at: "2026-10-19T05:50:00.125643Z"
          session_id: 7a4254c0-279e-41c4-bdc1-a455a8ff7a0a
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 023d85a9241a99aac8751d5482937cd4afa7a08c
        tracker_head: f32e2d07962bfcc0e909f548482f1eaa196c5045
        branch_point: 5e0e91bc73e248068275f5887d7cf1d5b0102801
        since: "2026-10-18T21:13:34-07:00"
        until: "2026-10-18T21:42:22-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T21:13:34-07:00"
              end: "2026-10-18T21:42:22-07:00"
              began: "2026-10-18T21:13:34-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 40f6363aac480e60168947987fc27f7ad5b5ff8a00e171196edefa504116ef67
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 47847010
              windows:
                - start: "2026-10-18T21:13:34-07:00"
                  end: "2026-10-18T21:42:22-07:00"
            - harness: claude
              id: 3b2498aa-31d7-48a3-972b-3450f04dfcf6
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 8f935031314e7c596eb7b321dfee662730c23f7261b9d60ae5147de87ad57a59
              started: "2026-10-19T04:25:19.07Z"
              spawned: false
              bytes: 392518
              windows:
                - start: "2026-10-19T04:25:19.07Z"
                  end: "2026-10-18T21:42:22-07:00"
            - harness: claude
              id: 3b8f552c-eeb9-4227-91af-2c9a3a636fd2
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 0702aeaf2740aad32b8a9c4fa71eb8ec86ee61904cd0134a13f743c2c536f5f2
              started: "2026-10-19T04:24:56.919Z"
              spawned: false
              bytes: 496753
              windows:
                - start: "2026-10-19T04:24:56.919Z"
                  end: "2026-10-18T21:42:22-07:00"
            - harness: claude
              id: 40102923-8652-4e3a-8bc3-f4e04f75585e
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 09d0dc950d55773755577b07d332f0a96d21ebb166efc0c91a4c969da81e789f
              started: "2026-10-19T04:23:20.775Z"
              spawned: false
              bytes: 511552
              windows:
                - start: "2026-10-19T04:23:20.775Z"
                  end: "2026-10-18T21:42:22-07:00"
      "n": 1
      timestamp: "2026-10-18T22:50:02-07:00"
      dispatched: "2026-10-18T22:46:13-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: A rerun after a stop between Rename and MarkAllSeen never marks the moved doc seen or checks its lock
          detail: 'fileMove.Registered is s.Registered[old] only (migrate.go:150), and couch already registers the new path after Rename, so the rerun skips Rename and MarkAllSeen (migrate_apply.go:185-195). Couch then shows the migration''s edits to the TL as unseen operator edits. s.Locked is also never computed for destination paths (migrate_apply.go:140). A throwaway probe failed at stops 4 and 9. Fix: Registered = s.Registered[old] || s.Registered[new] when atDest, check locks on destinations too, and add a seenFiles assertion to TestMigrateAnswersResumesFromEveryStep.'
          family: resume-reads-state-at-one-path
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: The gitignore check promised by the Spec and plan is not implemented
          detail: The Spec says "Gitignore follows wherever qa/ was ignored", and the plan says the command checks it and warns if the side folder isn't ignored where qa/ was. Nothing in cmd/tl-lead checks gitignore, and no plan Revision drops it. Either add the check with a warning in renderPlan, or add a Revision saying why it is unneeded.
          family: spec-clause-undelivered
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: rebaseIndex maps outside files' discussions entries as root-relative, though they are file-relative
          detail: For outside files (migrate.go:204), entries such as ../../qa/... are never rebased, and a subdirectory file's qa/op/x would be mis-mapped. Limit rebaseIndex to the session file, or resolve each entry against its own file.
          family: path-base-mismatch
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: The fake's MarkAllSeen ignores its path, and conformance never pins occurrence > 1
          detail: Any MarkAllSeen call marks the session file seen in the fake. lineRecords can emit occurrence > 1, but live conformance only checks occurrence 1.
          family: fake-fidelity
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: --from can name a folder outside the team directory
          detail: The plan says every path stays under the team directory, but readMigration accepts --from ../x.
          family: path-confinement
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: Temp files from createExclusive and replaceIfUnchanged survive a process death
          detail: A .tl-lead-migrate-* file left in the side folder after a crash has no removal path.
          family: crash-residue
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a4f4d284-f2e7-4b9e-8c26-69900c1f7ce0
            repository: github.com/xianxu/ariadne
            issue_id: "000455"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/a07222c75c04f0de
            reviewed_head: 7b3b597d0ce9de1f6cd252dfc2118aea4fb823fc
            slot: ariadne:4
            requested_at: "2026-10-19T06:57:30.829121Z"
          phase: finished
          pid: 2157
          started_at: "2026-10-19T06:57:31.030313Z"
          finished_at: "2026-10-19T06:59:19.011734Z"
          session_id: 29515d65-af61-49f0-b62c-eeb4e6362927
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 7b3b597d0ce9de1f6cd252dfc2118aea4fb823fc
        tracker_head: 3cd96f3b16a34ed71e6b84cff413fa5325cc5aca
        branch_point: 5e0e91bc73e248068275f5887d7cf1d5b0102801
        since: "2026-10-18T21:13:34-07:00"
        until: "2026-10-18T22:52:10-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T21:13:34-07:00"
              end: "2026-10-18T22:52:10-07:00"
              began: "2026-10-18T21:13:34-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: be67d3aadfae50063467a0b142312750e61def682f4a9e33ecf2bc54716efceb
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 50768528
              windows:
                - start: "2026-10-18T21:13:34-07:00"
                  end: "2026-10-18T22:52:10-07:00"
            - harness: claude
              id: 3b2498aa-31d7-48a3-972b-3450f04dfcf6
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 8f935031314e7c596eb7b321dfee662730c23f7261b9d60ae5147de87ad57a59
              started: "2026-10-19T04:25:19.07Z"
              spawned: false
              bytes: 392518
              windows:
                - start: "2026-10-19T04:25:19.07Z"
                  end: "2026-10-18T22:52:10-07:00"
            - harness: claude
              id: 3b8f552c-eeb9-4227-91af-2c9a3a636fd2
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 0702aeaf2740aad32b8a9c4fa71eb8ec86ee61904cd0134a13f743c2c536f5f2
              started: "2026-10-19T04:24:56.919Z"
              spawned: false
              bytes: 496753
              windows:
                - start: "2026-10-19T04:24:56.919Z"
                  end: "2026-10-18T22:52:10-07:00"
            - harness: claude
              id: 40102923-8652-4e3a-8bc3-f4e04f75585e
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 09d0dc950d55773755577b07d332f0a96d21ebb166efc0c91a4c969da81e789f
              started: "2026-10-19T04:23:20.775Z"
              spawned: false
              bytes: 511552
              windows:
                - start: "2026-10-19T04:23:20.775Z"
                  end: "2026-10-18T22:52:10-07:00"
        cuts:
            - dispatched: "2026-10-18T22:46:13-07:00"
              finalized: "2026-10-18T22:50:02-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-18T23:59:19-07:00"
      dispatched: "2026-10-18T23:57:30-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Registered and Locked are read at both paths (migrate.go:156-157). Reverting it makes ResumesFromEveryStep fail at stop 4 and LockAtDestination fail.
          round: 2
        - id: BR-2
          disposition: addressed
          note: gitignoreGap (migrate_apply.go:164) warns in renderPlan. TestMigrateAnswersGitignoreFollows covers it with a real git repo.
          round: 2
        - id: BR-3
          disposition: addressed
          note: rebaseIndex resolves each entry against path.Dir(file). TestRebaseIndexIsFileRelative covers the ../../qa case.
          round: 2
        - id: BR-4
          disposition: addressed
          note: The fake's MarkAllSeen is now scoped to its own file. Live conformance applies an occurrence-2 record and checks the exact text. The fake-only change has no failing-without test, which is acceptable for test code.
          round: 2
        - id: BR-5
          disposition: addressed
          note: readMigration refuses absolute paths, ".", ".." and "../" prefixes. TestMigrateAnswersFromStaysInside covers all four.
          round: 2
        - id: BR-6
          disposition: addressed
          note: sweepTemps runs at the start of --apply. Reverting it makes TestMigrateAnswersSweepsItsTemps fail.
          round: 2
      findings:
        - kind: defect
          id: BR-7
          severity: Minor
          title: No test pins the lock check on side-folder destinations; the fixture reaches it through the outside-file clause
          detail: This is the 2nd finding in family resume-reads-state-at-one-path. The rule is that every piece of per-file state the resume reads (registration, lock, seen baseline) is read at both old and new path, and each read is pinned by a fixture that reaches it only through that clause. Removing the SideDir prefix clause at migrate_apply.go:147 leaves TestMigrateAnswersLockAtDestination green, because the moved s4 doc mentions qa/op/log.jsonl. Add a destination doc that never mentions qa.
          family: resume-reads-state-at-one-path
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: 7b3b597d0ce9de1f6cd252dfc2118aea4fb823fc
      blocked: false
---

# Gate ledger — ariadne#455 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T22:50:02-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `resume-reads-state-at-one-path` (kind: defect) A rerun after a stop between Rename and MarkAllSeen never marks the moved doc seen or checks its lock
  fileMove.Registered is s.Registered[old] only (migrate.go:150), and couch already registers the new path after Rename, so the rerun skips Rename and MarkAllSeen (migrate_apply.go:185-195). Couch then shows the migration's edits to the TL as unseen operator edits. s.Locked is also never computed for destination paths (migrate_apply.go:140). A throwaway probe failed at stops 4 and 9. Fix: Registered = s.Registered[old] || s.Registered[new] when atDest, check locks on destinations too, and add a seenFiles assertion to TestMigrateAnswersResumesFromEveryStep.
- **BR-2** [Important] `spec-clause-undelivered` (kind: defect) The gitignore check promised by the Spec and plan is not implemented
  The Spec says "Gitignore follows wherever qa/ was ignored", and the plan says the command checks it and warns if the side folder isn't ignored where qa/ was. Nothing in cmd/tl-lead checks gitignore, and no plan Revision drops it. Either add the check with a warning in renderPlan, or add a Revision saying why it is unneeded.
- **BR-3** [Minor] `path-base-mismatch` (kind: defect) rebaseIndex maps outside files' discussions entries as root-relative, though they are file-relative
  For outside files (migrate.go:204), entries such as ../../qa/... are never rebased, and a subdirectory file's qa/op/x would be mis-mapped. Limit rebaseIndex to the session file, or resolve each entry against its own file.
- **BR-4** [Minor] `fake-fidelity` (kind: defect) The fake's MarkAllSeen ignores its path, and conformance never pins occurrence > 1
  Any MarkAllSeen call marks the session file seen in the fake. lineRecords can emit occurrence > 1, but live conformance only checks occurrence 1.
- **BR-5** [Minor] `path-confinement` (kind: defect) --from can name a folder outside the team directory
  The plan says every path stays under the team directory, but readMigration accepts --from ../x.
- **BR-6** [Minor] `crash-residue` (kind: defect) Temp files from createExclusive and replaceIfUnchanged survive a process death
  A .tl-lead-migrate-* file left in the side folder after a crash has no removal path.

## Round 2 — 2026-10-18T23:59:19-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Registered and Locked are read at both paths (migrate.go:156-157). Reverting it makes ResumesFromEveryStep fail at stop 4 and LockAtDestination fail.
- BR-2 — addressed — gitignoreGap (migrate_apply.go:164) warns in renderPlan. TestMigrateAnswersGitignoreFollows covers it with a real git repo.
- BR-3 — addressed — rebaseIndex resolves each entry against path.Dir(file). TestRebaseIndexIsFileRelative covers the ../../qa case.
- BR-4 — addressed — The fake's MarkAllSeen is now scoped to its own file. Live conformance applies an occurrence-2 record and checks the exact text. The fake-only change has no failing-without test, which is acceptable for test code.
- BR-5 — addressed — readMigration refuses absolute paths, ".", ".." and "../" prefixes. TestMigrateAnswersFromStaysInside covers all four.
- BR-6 — addressed — sweepTemps runs at the start of --apply. Reverting it makes TestMigrateAnswersSweepsItsTemps fail.

### Raised

- **BR-7** [Minor] `resume-reads-state-at-one-path` (kind: defect) No test pins the lock check on side-folder destinations; the fixture reaches it through the outside-file clause
  This is the 2nd finding in family resume-reads-state-at-one-path. The rule is that every piece of per-file state the resume reads (registration, lock, seen baseline) is read at both old and new path, and each read is pinned by a fixture that reaches it only through that clause. Removing the SideDir prefix clause at migrate_apply.go:147 leaves TestMigrateAnswersLockAtDestination green, because the moved s4 doc mentions qa/op/log.jsonl. Add a destination doc that never mentions qa.

## Open findings

- **BR-7** [Minor] `resume-reads-state-at-one-path` (kind: defect) No test pins the lock check on side-folder destinations; the fixture reaches it through the outside-file clause
